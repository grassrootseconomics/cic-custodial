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
	"github.com/grassrootseconomics/cic-custodial/pkg/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/knadh/koanf"
	"github.com/labstack/echo/v4"
	"github.com/zerodha/logf"
)

var (
	confFlag       string
	debugFlag      bool
	taskerModeFlag bool
	apiModeFlag    bool

	asynqRedisPool   *redis.RedisPool
	celoProvider     *celo.Provider
	commonRedisPool  *redis.RedisPool
	lo               logf.Logger
	lockProvider     *redislock.Client
	ko               *koanf.Koanf
	postgresPool     *pgxpool.Pool
	postgresKeystore keystore.Keystore
	redisNoncestore  nonce.Noncestore
	system           *tasker.SystemContainer
	taskerClient     *tasker.TaskerClient
)

func init() {
	flag.StringVar(&confFlag, "config", "config.toml", "Config file location")
	flag.BoolVar(&debugFlag, "log", false, "Enable debug logging")
	flag.BoolVar(&taskerModeFlag, "tasker", true, "Start tasker")
	flag.BoolVar(&apiModeFlag, "api", true, "Start API server")
	flag.Parse()

	lo = initLogger(debugFlag)
	ko = initConfig(confFlag)

	celoProvider = initCeloProvider()
	postgresPool = initPostgresPool()
	postgresKeystore = initKeystore()
	asynqRedisPool = initAsynqRedisPool()
	commonRedisPool = initCommonRedisPool()
	redisNoncestore = initRedisNoncestore()
	lockProvider = initLockProvider()
	taskerClient = initTaskerClient()
	system = initSystemContainer()
}

func main() {
	var (
		tasker    *tasker.TaskerServer
		apiServer *echo.Echo
		wg        sync.WaitGroup
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if apiModeFlag {
		apiServer = initApiServer()

		wg.Add(1)
		go func() {
			defer wg.Done()

			lo.Info("Starting API server")
			if err := apiServer.Start(ko.MustString("service.address")); err != nil {
				if strings.Contains(err.Error(), "Server closed") {
					lo.Info("Shutting down server")
				} else {
					lo.Fatal("Could not start api server", "err", err)
				}
			}
		}()
	}

	if taskerModeFlag {
		tasker = initTasker()

		wg.Add(1)
		go func() {
			defer wg.Done()

			lo.Info("Starting tasker")
			if err := tasker.Start(); err != nil {
				lo.Fatal("Could not start task server", "err", err)
			}
		}()
	}

	<-ctx.Done()
	lo.Info("Graceful shutdown triggered")

	if taskerModeFlag {
		lo.Debug("Stopping tasker")
		tasker.Stop()
	}

	if apiModeFlag {
		lo.Debug("Stopping api server")
		if err := apiServer.Shutdown(ctx); err != nil {
			lo.Error("Could not gracefully shutdown api server", "err", err)
		}
	}

	wg.Wait()
}
