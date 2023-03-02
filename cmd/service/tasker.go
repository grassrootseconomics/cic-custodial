package main

import (
	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker/task"
	"github.com/grassrootseconomics/cic-custodial/pkg/redis"
	"github.com/hibiken/asynq"
)

// Load tasker handlers, injecting any necessary handler dependencies from the system container.
func initTasker(custodialContainer *custodial.Custodial, redisPool *redis.RedisPool) *tasker.TaskerServer {
	taskerServerOpts := tasker.TaskerServerOpts{
		Concurrency: ko.MustInt("asynq.worker_count"),
		Logg:        lo,
		LogLevel:    asynq.InfoLevel,
		RedisPool:   redisPool,
	}

	taskerServer := tasker.NewTaskerServer(taskerServerOpts)

	taskerServer.RegisterHandlers(tasker.AccountPrepareTask, task.AccountPrepare(custodialContainer))
	taskerServer.RegisterHandlers(tasker.AccountRegisterTask, task.AccountRegisterOnChainProcessor(custodialContainer))
	taskerServer.RegisterHandlers(tasker.AccountGiftGasTask, task.AccountGiftGasProcessor(custodialContainer))
	taskerServer.RegisterHandlers(tasker.AccountGiftVoucherTask, task.GiftVoucherProcessor(custodialContainer))
	taskerServer.RegisterHandlers(tasker.AccountRefillGasTask, task.AccountRefillGasProcessor(custodialContainer))
	taskerServer.RegisterHandlers(tasker.SignTransferTask, task.SignTransfer(custodialContainer))
	taskerServer.RegisterHandlers(tasker.DispatchTxTask, task.DispatchTx(custodialContainer))

	return taskerServer
}
