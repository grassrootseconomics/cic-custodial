package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/knadh/koanf/v2"
	"github.com/labstack/echo/v4"
	"github.com/zerodha/logf"
)

var (
	confFlag             string
	debugFlag            bool
	migrationsFolderFlag string
	queriesFlag          string

	lo logf.Logger
	ko *koanf.Koanf
)

func init() {
	flag.StringVar(&confFlag, "config", "config.toml", "Config file location")
	flag.BoolVar(&debugFlag, "log", false, "Enable debug logging")
	flag.StringVar(&migrationsFolderFlag, "migrations", "migrations/", "Migrations folder location")
	flag.StringVar(&queriesFlag, "queries", "queries.sql", "Queries file location")
	flag.Parse()

	lo = initLogger(debugFlag)
	ko = initConfig(confFlag)
}

func main() {
	var (
		tasker    *tasker.TaskerServer
		apiServer *echo.Echo
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	queries, err := initQueries(queriesFlag)
	if err != nil {
		lo.Fatal("main: critical error loading SQL queries", "error", err)
	}

	celoProvider, err := initCeloProvider()
	if err != nil {
		lo.Fatal("main: critical error loading chain provider", "error", err)
	}

	postgresPool, err := initPostgresPool()
	if err != nil {
		lo.Fatal("main: critical error connecting to postgres", "error", err)
	}

	asynqRedisPool, err := initAsynqRedisPool()
	if err != nil {
		lo.Fatal("main: critical error connecting to asynq redis db", "error", err)
	}

	redisPool, err := initCommonRedisPool()
	if err != nil {
		lo.Fatal("main: critical error connecting to common redis db", "error", err)
	}

	postgresKeystore, err := initPostgresKeystore(postgresPool, queries)
	if err != nil {
		lo.Fatal("main: critical error loading keystore")
	}

	jsEventEmitter, err := initJetStream()
	if err != nil {
		lo.Fatal("main: critical error loading jetstream event emitter")
	}

	pgStore := initPostgresStore(postgresPool, queries)
	redisNoncestore := initRedisNoncestore(redisPool, celoProvider)
	lockProvider := initLockProvider(redisPool.Client)
	taskerClient := initTaskerClient(asynqRedisPool)

	systemContainer, err := initSystemContainer(context.Background(), redisNoncestore)
	if err != nil {
		lo.Fatal("main: critical error bootstrapping system container", "error", err)
	}

	custodial := &custodial.Custodial{
		CeloProvider:    celoProvider,
		EventEmitter:    jsEventEmitter,
		Keystore:        postgresKeystore,
		LockProvider:    lockProvider,
		Noncestore:      redisNoncestore,
		PgStore:         pgStore,
		SystemContainer: systemContainer,
		TaskerClient:    taskerClient,
	}

	wg := &sync.WaitGroup{}

	apiServer = initApiServer(custodial)
	wg.Add(1)
	go func() {
		defer wg.Done()
		lo.Info("main: starting API server")
		if err := apiServer.Start(ko.MustString("service.address")); err != nil {
			if strings.Contains(err.Error(), "Server closed") {
				lo.Info("main: shutting down server")
			} else {
				lo.Fatal("main: critical error shutting down server", "err", err)
			}
		}
	}()

	tasker = initTasker(custodial, asynqRedisPool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		lo.Info("Starting tasker")
		if err := tasker.Start(); err != nil {
			lo.Fatal("main: could not start task server", "err", err)
		}
	}()

	<-ctx.Done()

	lo.Info("main: stopping tasker")
	tasker.Stop()

	lo.Info("main: stopping api server")
	if err := apiServer.Shutdown(ctx); err != nil {
		lo.Error("Could not gracefully shutdown api server", "err", err)
	}

	wg.Wait()
}
