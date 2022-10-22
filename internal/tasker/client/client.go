package client

import (
	"time"

	"github.com/hibiken/asynq"
)

type JobType string

var (
	defaultRetentionPeriod = 24 * 2 * time.Hour
)

type Opts struct {
	RedisDSN string
}

type TaskerClient struct {
	Client *asynq.Client
}

func NewTaskerClient(o Opts) *TaskerClient {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr: o.RedisDSN,
	})

	return &TaskerClient{
		Client: client,
	}
}
