package tasker

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/grassrootseconomics/cic-custodial/pkg/redis"
	"github.com/hibiken/asynq"
)

const (
	taskTimeout   = 60 * time.Second
	taskRetention = 48 * time.Hour
)

type TaskerClientOpts struct {
	RedisPool *redis.RedisPool
}

type TaskerClient struct {
	Client *asynq.Client
}

func NewTaskerClient(o TaskerClientOpts) *TaskerClient {
	return &TaskerClient{
		Client: asynq.NewClient(o.RedisPool),
	}
}

func (c *TaskerClient) CreateTask(ctx context.Context, taskName TaskName, queueName QueueName, task *Task, extraOpts ...asynq.Option) (*asynq.TaskInfo, error) {
	if task.Id == "" {
		task.Id = uuid.NewString()
	}

	defaultOpts := []asynq.Option{
		asynq.Queue(string(queueName)),
		asynq.TaskID(task.Id),
		asynq.Retention(taskRetention),
		asynq.Timeout(taskTimeout),
	}
	defaultOpts = append(defaultOpts, extraOpts...)

	qTask := asynq.NewTask(
		string(taskName),
		task.Payload,
		defaultOpts...,
	)

	taskInfo, err := c.Client.EnqueueContext(ctx, qTask)
	if err != nil {
		return nil, err
	}

	return taskInfo, nil
}
