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

func (c *TaskerClient) CreateTask(ctx context.Context, taskName TaskName, queueName QueueName, task *Task) (*asynq.TaskInfo, error) {
	if task.Id == "" {
		task.Id = uuid.NewString()
	}

	qTask := asynq.NewTask(
		string(taskName),
		task.Payload,
		asynq.Queue(string(queueName)),
		asynq.TaskID(task.Id),
		asynq.Retention(taskRetention),
		asynq.Timeout(taskTimeout),
	)

	taskInfo, err := c.Client.EnqueueContext(ctx, qTask)
	if err != nil {
		return nil, err
	}

	return taskInfo, nil
}
