package main

import (
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker/task"
	"github.com/hibiken/asynq"
)

func initTasker() *tasker.TaskerServer {
	lo.Debug("Bootstrapping tasker")

	taskerServerOpts := tasker.TaskerServerOpts{
		Concurrency:     ko.MustInt("asynq.concurrency"),
		Logg:            lo,
		RedisPool:       asynqRedisPool,
		SystemContainer: nil,
		TaskerClient:    taskerClient,
	}

	if debugFlag {
		taskerServerOpts.LogLevel = asynq.DebugLevel
	}

	taskerServer := tasker.NewTaskerServer(taskerServerOpts)

	taskerServer.RegisterHandlers(tasker.PrepareAccountTask, task.PrepareAccount(
		redisNoncestore,
		taskerClient,
	))
	taskerServer.RegisterHandlers(tasker.GiftGasTask, task.GiftGasProcessor(
		celoProvider,
		redisNoncestore,
		lockProvider,
		system,
		taskerClient,
	))
	taskerServer.RegisterHandlers(tasker.GiftTokenTask, task.GiftTokenProcessor(
		celoProvider,
		redisNoncestore,
		lockProvider,
		system,
		taskerClient,
	))
	taskerServer.RegisterHandlers(tasker.RefillGasTask, task.RefillGasProcessor(
		celoProvider,
		redisNoncestore,
		lockProvider,
		system,
		taskerClient,
	))
	taskerServer.RegisterHandlers(tasker.TransferTokenTask, task.TransferToken(
		celoProvider,
		redisNoncestore,
		postgresKeystore,
		lockProvider,
		system,
		taskerClient,
	))
	taskerServer.RegisterHandlers(tasker.TxDispatchTask, task.TxDispatch(
		celoProvider,
	))

	lo.Debug("Registered all tasker handlers")
	return taskerServer
}
