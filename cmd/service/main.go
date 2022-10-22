package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/grassrootseconomics/cic-custodial/internal/actions"
	"github.com/grassrootseconomics/cic-custodial/internal/api"
	tasker_client "github.com/grassrootseconomics/cic-custodial/internal/tasker/client"
	tasker_server "github.com/grassrootseconomics/cic-custodial/internal/tasker/server"
	"github.com/grassrootseconomics/cic-go-sdk/chain"
	"github.com/knadh/koanf"
	"github.com/labstack/echo/v4"
	"github.com/zerodha/logf"
)

var (
	lo logf.Logger
	ko *koanf.Koanf

	chainProvider *chain.Provider
	taskerClient  *tasker_client.TaskerClient

	httpServer   *echo.Echo
	taskerServer *tasker_server.TaskerServer
)

func init() {
	ko = initConfig("config.toml")
	lo = initLogger()

	chainProvider = initChainProvider()
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	taskerClient = initTaskerClient()

	actionsProvider := actions.NewActionsProvider(actions.Opts{
		SystemProvider: initSystemProvider(),
		ChainProvider:  chainProvider,
		Keystore:       initKeystore(),
		Noncestore:     initNoncestore(),
	})

	httpServer = api.BootstrapHTTPServer(api.Opts{
		ActionsProvider: actionsProvider,
		TaskerClient:    taskerClient,
	})

	taskerServer = tasker_server.NewTaskerServer(tasker_server.Opts{
		ActionsProvider: actionsProvider,
		TaskerClient:    taskerClient,
		RedisDSN:        ko.MustString("tasker.dsn"),
		Concurrency:     15,
		Logger:          lo,
	})

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := httpServer.Start(ko.MustString("service.address")); err != nil {
			if strings.Contains(err.Error(), "Server closed") {
				lo.Info("shutting down server")
			} else {
				lo.Fatal("could not start api server", "err", err)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := taskerServer.Server.Run(taskerServer.Mux); err != nil {
			lo.Fatal("could not start task server", "err", err)
		}
	}()

	gracefulShutdown(ctx)
	wg.Wait()
}

func gracefulShutdown(ctx context.Context) {
	<-ctx.Done()
	lo.Info("graceful shutdown triggered")

	taskerServer.Server.Shutdown()
	lo.Info("job processor successfully shutdown")

	if err := httpServer.Shutdown(ctx); err != nil {
		lo.Error("could not gracefully shutdown api server", "err", err)
	}
	lo.Info("api server gracefully shutdown")
}
