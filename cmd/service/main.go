package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/bsm/redislock"
	celo "github.com/grassrootseconomics/cic-celo-sdk"
	"github.com/grassrootseconomics/cic-custodial/internal/keystore"
	"github.com/grassrootseconomics/cic-custodial/internal/nonce"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/knadh/koanf"
	"github.com/labstack/echo/v4"
	"github.com/zerodha/logf"
)

var (
	confFlag  string
	debugFlag bool

	lo logf.Logger
	ko *koanf.Koanf
)

type custodial struct {
	celoProvider    *celo.Provider
	keystore        keystore.Keystore
	lockProvider    *redislock.Client
	noncestore      nonce.Noncestore
	systemContainer *tasker.SystemContainer
	taskerClient    *tasker.TaskerClient
}

func init() {
	flag.StringVar(&confFlag, "config", "config.toml", "Config file location")
	flag.BoolVar(&debugFlag, "log", false, "Enable debug logging")
	flag.Parse()

	lo = initLogger(debugFlag)
	ko = initConfig(confFlag)
}

func main() {
	var (
		tasker    *tasker.TaskerServer
		apiServer *echo.Echo
		wg        sync.WaitGroup
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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

	postgresKeystore, err := initPostgresKeystore(postgresPool)
	if err != nil {
		lo.Fatal("main: critical error loading postgres keystore", "error", err)
	}

	redisNoncestore := initRedisNoncestore(redisPool, celoProvider)
	lockProvider := initLockProvider(redisPool.Client)
	taskerClient := initTaskerClient(asynqRedisPool)

	systemContainer, err := initSystemContainer(context.Background(), redisNoncestore)
	if err != nil {
		lo.Fatal("main: critical error bootstrapping system container", "error", err)
	}

	custodial := &custodial{
		celoProvider:    celoProvider,
		keystore:        postgresKeystore,
		lockProvider:    lockProvider,
		noncestore:      redisNoncestore,
		systemContainer: systemContainer,
		taskerClient:    taskerClient,
	}

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
