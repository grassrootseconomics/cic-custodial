package tasker

import (
	"context"
	"time"

	"github.com/grassrootseconomics/cic-custodial/pkg/logg"
	"github.com/grassrootseconomics/cic-custodial/pkg/redis"
	"github.com/hibiken/asynq"
	"github.com/zerodha/logf"
)

const (
	retryRequeueInterval = 1 * time.Second
)

type TaskerServerOpts struct {
	Concurrency      int
	ErrorHandler     asynq.ErrorHandler
	IsFailureHandler func(error) bool
	Logg             logf.Logger
	LogLevel         asynq.LogLevel
	RedisPool        *redis.RedisPool
	RetryHandler     asynq.RetryDelayFunc
}

type TaskerServer struct {
	mux    *asynq.ServeMux
	server *asynq.Server
}

func NewTaskerServer(o TaskerServerOpts) *TaskerServer {
	server := asynq.NewServer(
		o.RedisPool,
		asynq.Config{
			Concurrency:              o.Concurrency,
			DelayedTaskCheckInterval: retryRequeueInterval,
			ErrorHandler:             o.ErrorHandler,
			IsFailure:                o.IsFailureHandler,
			LogLevel:                 o.LogLevel,
			Logger:                   logg.AsynqCompatibleLogger(o.Logg),
			Queues: map[string]int{
				string(HighPriority):    7,
				string(DefaultPriority): 3,
			},
			RetryDelayFunc: o.RetryHandler,
		},
	)

	mux := asynq.NewServeMux()

	return &TaskerServer{
		mux:    mux,
		server: server,
	}
}

func (ts *TaskerServer) RegisterMiddlewareStack(middlewareStack []asynq.MiddlewareFunc) {
	ts.mux.Use(middlewareStack...)
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
