package main

import (
	"context"
	"time"

	"github.com/bsm/redislock"
	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker/task"
	"github.com/grassrootseconomics/cic-custodial/pkg/redis"
	"github.com/hibiken/asynq"
)

const (
	fixedRetryCount  = 25
	fixedRetryPeriod = time.Second * 1
)

// Load tasker handlers, injecting any necessary handler dependencies from the system container.
func initTasker(custodialContainer *custodial.Custodial, redisPool *redis.RedisPool) *tasker.TaskerServer {
	taskerServerOpts := tasker.TaskerServerOpts{
		Concurrency:      ko.MustInt("asynq.worker_count"),
		IsFailureHandler: isFailureHandler,
		Logg:             lo,
		LogLevel:         asynq.InfoLevel,
		RedisPool:        redisPool,
		RetryHandler:     retryHandler,
	}

	taskerServer := tasker.NewTaskerServer(taskerServerOpts)

	taskerServer.RegisterMiddlewareStack([]asynq.MiddlewareFunc{
		observibilityMiddleware(),
	})

	taskerServer.RegisterHandlers(tasker.AccountPrepareTask, task.AccountPrepare(custodialContainer))
	taskerServer.RegisterHandlers(tasker.AccountRegisterTask, task.AccountRegisterOnChainProcessor(custodialContainer))
	taskerServer.RegisterHandlers(tasker.AccountGiftGasTask, task.AccountGiftGasProcessor(custodialContainer))
	taskerServer.RegisterHandlers(tasker.AccountGiftVoucherTask, task.GiftVoucherProcessor(custodialContainer))
	taskerServer.RegisterHandlers(tasker.AccountActivateTask, task.AccountActivateProcessor(custodialContainer))
	taskerServer.RegisterHandlers(tasker.AccountRefillGasTask, task.AccountRefillGasProcessor(custodialContainer))
	taskerServer.RegisterHandlers(tasker.SignTransferTask, task.SignTransfer(custodialContainer))
	taskerServer.RegisterHandlers(tasker.DispatchTxTask, task.DispatchTx(custodialContainer))

	return taskerServer
}

func isFailureHandler(err error) bool {
	switch err {
	// Ignore lock contention errors; retry until lock obtain.
	case redislock.ErrNotObtained:
		return false
	default:
		return true
	}
}

func retryHandler(count int, err error, task *asynq.Task) time.Duration {
	if count < fixedRetryCount {
		return fixedRetryPeriod
	} else {
		return asynq.DefaultRetryDelayFunc(count, err, task)
	}
}

func observibilityMiddleware() asynq.MiddlewareFunc {
	return func(handler asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
			taskId, _ := asynq.GetTaskID(ctx)

			err := handler.ProcessTask(ctx, task)
			if err != nil && isFailureHandler(err) {
				lo.Error("tasker: handler error", "err", err, "task_type", task.Type(), "task_id", taskId)
			} else if asynq.IsPanicError(err) {
				lo.Error("tasker: handler panic", "err", err, "task_type", task.Type(), "task_id", taskId)
			} else {
				lo.Info("tasker: process task", "task_type", task.Type(), "task_id", taskId)
			}

			return err
		})
	}
}
