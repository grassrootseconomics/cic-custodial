package main

import (
	"context"
	"flag"
	"strings"
	"sync"

	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/grassrootseconomics/cic-custodial/internal/events"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/knadh/koanf/v2"
	"github.com/labstack/echo/v4"
	"github.com/zerodha/logf"
)

type (
	internalServiceContainer struct {
		apiService    *echo.Echo
		jetstreamSub  *events.JetStream
		taskerService *tasker.TaskerServer
	}
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

	lo = initLogger()
	ko = initConfig()
}

func main() {
	parsedQueries := initQueries()
	celoProvider := initCeloProvider()
	postgresPool := initPostgresPool()
	asynqRedisPool := initAsynqRedisPool()
	redisPool := initCommonRedisPool()

	postgresKeystore := initPostgresKeystore(postgresPool, parsedQueries)
	pgStore := initPostgresStore(postgresPool, parsedQueries)
	redisNoncestore := initRedisNoncestore(redisPool, celoProvider)
	lockProvider := initLockProvider(redisPool.Client)
	taskerClient := initTaskerClient(asynqRedisPool)
	systemContainer := initSystemContainer(context.Background(), redisNoncestore)

	jsEventEmitter := initJetStream(pgStore)

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

	internalServices := &internalServiceContainer{}
	wg := &sync.WaitGroup{}

	signalCh, closeCh := createSigChannel()
	defer closeCh()

	internalServices.apiService = initApiServer(custodial)
	wg.Add(1)
	go func() {
		defer wg.Done()
		lo.Info("main: starting API server")
		if err := internalServices.apiService.Start(ko.MustString("service.address")); err != nil {
			if strings.Contains(err.Error(), "Server closed") {
				lo.Info("main: shutting down server")
			} else {
				lo.Fatal("main: critical error shutting down server", "err", err)
			}
		}
	}()

	internalServices.taskerService = initTasker(custodial, asynqRedisPool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		lo.Info("Starting tasker")
		if err := internalServices.taskerService.Start(); err != nil {
			lo.Fatal("main: could not start task server", "err", err)
		}
	}()

	internalServices.jetstreamSub = jsEventEmitter
	wg.Add(1)
	go func() {
		defer wg.Done()
		lo.Info("Starting jetstream sub")
		if err := internalServices.jetstreamSub.Subscriber(); err != nil {
			lo.Fatal("main: error running jetstream sub", "err", err)
		}
	}()

	<-signalCh
	startGracefulShutdown(context.Background(), internalServices)

	wg.Wait()
}
