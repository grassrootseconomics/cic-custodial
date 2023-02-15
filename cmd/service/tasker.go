package main

import (
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker/task"
	"github.com/grassrootseconomics/cic-custodial/pkg/redis"
	"github.com/hibiken/asynq"
)

// Load tasker handlers, injecting any necessary handler dependencies from the system container.
func initTasker(custodialContainer *custodial, redisPool *redis.RedisPool) *tasker.TaskerServer {
	lo.Debug("Bootstrapping tasker")
	js, err := initJetStream()
	if err != nil {
		lo.Fatal("filters: critical error loading jetstream", "error", err)
	}

	taskerServerOpts := tasker.TaskerServerOpts{
		Concurrency:     ko.MustInt("asynq.worker_count"),
		Logg:            lo,
		LogLevel:        asynq.ErrorLevel,
		RedisPool:       redisPool,
		SystemContainer: custodialContainer.systemContainer,
		TaskerClient:    custodialContainer.taskerClient,
	}

	if debugFlag {
		taskerServerOpts.LogLevel = asynq.DebugLevel
	}

	taskerServer := tasker.NewTaskerServer(taskerServerOpts)

	taskerServer.RegisterHandlers(tasker.PrepareAccountTask, task.PrepareAccount(
		custodialContainer.noncestore,
		custodialContainer.taskerClient,
		js,
	))
	taskerServer.RegisterHandlers(tasker.RegisterAccountOnChain, task.RegisterAccountOnChainProcessor(
		custodialContainer.celoProvider,
		custodialContainer.lockProvider,
		custodialContainer.noncestore,
		custodialContainer.pgStore,
		custodialContainer.systemContainer,
		custodialContainer.taskerClient,
		js,
	))
	taskerServer.RegisterHandlers(tasker.GiftGasTask, task.GiftGasProcessor(
		custodialContainer.celoProvider,
		custodialContainer.lockProvider,
		custodialContainer.noncestore,
		custodialContainer.pgStore,
		custodialContainer.systemContainer,
		custodialContainer.taskerClient,
		js,
	))
	taskerServer.RegisterHandlers(tasker.GiftTokenTask, task.GiftTokenProcessor(
		custodialContainer.celoProvider,
		custodialContainer.lockProvider,
		custodialContainer.noncestore,
		custodialContainer.pgStore,
		custodialContainer.systemContainer,
		custodialContainer.taskerClient,
		js,
	))
	taskerServer.RegisterHandlers(tasker.SignTransferTask, task.SignTransfer(
		custodialContainer.celoProvider,
		custodialContainer.keystore,
		custodialContainer.lockProvider,
		custodialContainer.noncestore,
		custodialContainer.pgStore,
		custodialContainer.systemContainer,
		custodialContainer.taskerClient,
		js,
	))
	taskerServer.RegisterHandlers(tasker.TxDispatchTask, task.TxDispatch(
		custodialContainer.celoProvider,
		custodialContainer.pgStore,
		js,
	))

	return taskerServer
}
