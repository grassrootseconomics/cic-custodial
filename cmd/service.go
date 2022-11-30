package main

import (
	"context"
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

type service struct {
	taskerServer *tasker.TaskerServer
	apiServer    *echo.Echo
}

var (
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
	ko = initConfig("config.toml")
	lo = initLogger()

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
	service := &service{}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	service.taskerServer = initTasker()
	service.apiServer = initApiServer()

	startServices(service, ctx)
}

func startServices(serviceContainer *service, ctx context.Context) {
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := serviceContainer.taskerServer.Start(); err != nil {
			lo.Fatal("Could not start task server", "err", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := serviceContainer.apiServer.Start(ko.MustString("service.address")); err != nil {
			if strings.Contains(err.Error(), "Server closed") {
				lo.Info("Shutting down server")
			} else {
				lo.Fatal("Could not start api server", "err", err)
			}
		}
	}()

	gracefulShutdown(serviceContainer, ctx)
	wg.Wait()
}

func gracefulShutdown(serviceContainer *service, ctx context.Context) {
	<-ctx.Done()
	lo.Info("Graceful shutdown triggered")

	lo.Info("Stopping tasker dequeue")
	serviceContainer.taskerServer.Stop()
	lo.Info("Stopped tasker")

	if err := serviceContainer.apiServer.Shutdown(ctx); err != nil {
		lo.Error("Could not gracefully shutdown api server", "err", err)
	}
	lo.Info("Stopped API server")
}
