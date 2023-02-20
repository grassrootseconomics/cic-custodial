package tasker

import (
	"context"
	"time"

	"github.com/bsm/redislock"
	"github.com/grassrootseconomics/cic-custodial/pkg/redis"
	"github.com/hibiken/asynq"
	"github.com/zerodha/logf"
)

const (
	fixedRetryCount  = 25
	fixedRetryPeriod = time.Second * 1
)

type TaskerServerOpts struct {
	Concurrency int
	Logg        logf.Logger
	LogLevel    asynq.LogLevel
	RedisPool   *redis.RedisPool
}

type TaskerServer struct {
	mux    *asynq.ServeMux
	server *asynq.Server
}

func NewTaskerServer(o TaskerServerOpts) *TaskerServer {
	server := asynq.NewServer(
		o.RedisPool,
		asynq.Config{
			Concurrency: o.Concurrency,
			IsFailure:   expectedFailures,
			LogLevel:    o.LogLevel,
			Queues: map[string]int{
				string(HighPriority):    5,
				string(DefaultPriority): 2,
			},
			RetryDelayFunc: retryDelay,
		},
	)

	mux := asynq.NewServeMux()

	return &TaskerServer{
		mux:    mux,
		server: server,
	}
}

func (ts *TaskerServer) RegisterHandlers(taskName TaskName, taskHandler func(context.Context, *asynq.Task) error) {
	ts.mux.HandleFunc(string(taskName), taskHandler)
}

func (ts *TaskerServer) Start() error {
	if err := ts.server.Start(ts.mux); err != nil {
		return err
	}

	return nil
}

func (ts *TaskerServer) Stop() {
	ts.server.Stop()
	ts.server.Shutdown()
}

func expectedFailures(err error) bool {
	switch err {
	// Ignore lock contention errors; retry until lock obtain.
	case redislock.ErrNotObtained:
		return false
	default:
		return true
	}
}

// Immidiatel
func retryDelay(count int, err error, task *asynq.Task) time.Duration {
	if count < fixedRetryCount {
		return fixedRetryPeriod
	} else {
		return asynq.DefaultRetryDelayFunc(count, err, task)
	}
}
