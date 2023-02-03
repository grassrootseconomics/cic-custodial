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

	taskerServerOpts := tasker.TaskerServerOpts{
		Concurrency:     ko.MustInt("asynq.worker_count"),
		Logg:            lo,
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
	))
	taskerServer.RegisterHandlers(tasker.GiftGasTask, task.GiftGasProcessor(
		custodialContainer.celoProvider,
		custodialContainer.noncestore,
		custodialContainer.lockProvider,
		custodialContainer.systemContainer,
		custodialContainer.taskerClient,
	))
	taskerServer.RegisterHandlers(tasker.GiftTokenTask, task.GiftTokenProcessor(
		custodialContainer.celoProvider,
		custodialContainer.noncestore,
		custodialContainer.lockProvider,
		custodialContainer.systemContainer,
		custodialContainer.taskerClient,
	))
	taskerServer.RegisterHandlers(tasker.RefillGasTask, task.RefillGasProcessor(
		custodialContainer.celoProvider,
		custodialContainer.noncestore,
		custodialContainer.lockProvider,
		custodialContainer.systemContainer,
		custodialContainer.taskerClient,
	))
	taskerServer.RegisterHandlers(tasker.TxDispatchTask, task.TxDispatch(
		custodialContainer.celoProvider,
	))

	return taskerServer
}
