package main

import (
	"context"
	"flag"
	"strings"
	"sync"

	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/grassrootseconomics/cic-custodial/internal/sub"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/knadh/koanf/v2"
	"github.com/labstack/echo/v4"
	"github.com/zerodha/logf"
)

type internalServicesContainer struct {
	apiService    *echo.Echo
	jetstreamSub  *sub.Sub
	taskerService *tasker.TaskerServer
}

var (
	build string

	confFlag             string
	debugFlag            bool
	migrationsFolderFlag string
	queriesFlag          string

	lo logf.Logger
	ko *koanf.Koanf
)

func init() {
	flag.StringVar(&confFlag, "config", "config.toml", "Config file location")
	flag.BoolVar(&debugFlag, "debug", false, "Enable debug logging")
	flag.StringVar(&migrationsFolderFlag, "migrations", "migrations/", "Migrations folder location")
	flag.StringVar(&queriesFlag, "queries", "queries.sql", "Queries file location")
	flag.Parse()

	lo = initLogger()
	ko = initConfig()
}

func main() {
	lo.Info("main: starting cic-custodial", "build", build)

	celoProvider := initCeloProvider()
	asynqRedisPool := initAsynqRedisPool()
	redisPool := initCommonRedisPool()

	store := initPgStore()
	redisNoncestore := initRedisNoncestore(redisPool, celoProvider, store)
	lockProvider := initLockProvider(redisPool.Client)
	taskerClient := initTaskerClient(asynqRedisPool)

	natsConn, jsCtx := initJetStream()

	custodial, err := custodial.NewCustodial(custodial.Opts{
		CeloProvider:     celoProvider,
		LockProvider:     lockProvider,
		Logg:             lo,
		Noncestore:       redisNoncestore,
		Store:            store,
		RedisClient:      redisPool.Client,
		RegistryAddress:  ko.MustString("chain.registry_address"),
		SystemPrivateKey: ko.MustString("system.private_key"),
		SystemPublicKey:  ko.MustString("system.public_key"),
		TaskerClient:     taskerClient,
	})
	if err != nil {
		lo.Fatal("main: crtical error loading custodial container", "error", err)
	}

	internalServices := &internalServicesContainer{}
	wg := &sync.WaitGroup{}

	signalCh, closeCh := createSigChannel()
	defer closeCh()

	internalServices.apiService = initApiServer(custodial)
	wg.Add(1)
	go func() {
		defer wg.Done()
		host := ko.MustString("service.address")
		lo.Info("main: starting API server", "host", host)
		if err := internalServices.apiService.Start(host); err != nil {
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
		lo.Info("main: starting tasker")
		if err := internalServices.taskerService.Start(); err != nil {
			lo.Fatal("main: could not start task server", "err", err)
		}
	}()

	internalServices.jetstreamSub = initSub(natsConn, jsCtx, custodial)
	wg.Add(1)
	go func() {
		defer wg.Done()
		lo.Info("main: starting jetstream sub")
		if err := internalServices.jetstreamSub.Process(); err != nil {
			lo.Fatal("main: error running jetstream sub", "err", err)
		}
	}()

	lo.Info("main: graceful shutdown triggered", "signal", <-signalCh)
	startGracefulShutdown(context.Background(), internalServices)

	wg.Wait()
}
