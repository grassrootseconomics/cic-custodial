package server

import (
	"time"

	"github.com/grassrootseconomics/cic-custodial/internal/actions"
	tasker_client "github.com/grassrootseconomics/cic-custodial/internal/tasker/client"
	"github.com/hibiken/asynq"
	"github.com/zerodha/logf"
)

type Opts struct {
	ActionsProvider *actions.ActionsProvider
	TaskerClient    *tasker_client.TaskerClient
	RedisDSN        string
	Concurrency     int
	Logger          logf.Logger
}

type TaskerProcessor struct {
	ActionsProvider *actions.ActionsProvider
	TaskerClient    *tasker_client.TaskerClient
}

type TaskerServer struct {
	Server *asynq.Server
	Mux    *asynq.ServeMux
}

func NewTaskerServer(o Opts) *TaskerServer {
	taskerProcessor := &TaskerProcessor{
		ActionsProvider: o.ActionsProvider,
		TaskerClient:    o.TaskerClient,
	}

	asynqServer := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr: o.RedisDSN,
		},
		asynq.Config{
			Concurrency: o.Concurrency,
			Logger:      asynqCompatibleLogger(o.Logger),
			RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
				if n < 6 {
					return 1 * time.Second
				} else {
					return asynq.DefaultRetryDelayFunc(n, e, t)
				}
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(string(tasker_client.ActivateAccountTask), taskerProcessor.activateAccountProcessor)
	mux.HandleFunc(string(tasker_client.SetNewAccountNonceTask), taskerProcessor.setNewAccountNonce)
	mux.HandleFunc(string(tasker_client.GiftGasTask), taskerProcessor.giftGasProcessor)
	mux.HandleFunc(string(tasker_client.TxDispatchTask), taskerProcessor.txDispatcher)

	return &TaskerServer{
		Server: asynqServer,
		Mux:    mux,
	}
}
