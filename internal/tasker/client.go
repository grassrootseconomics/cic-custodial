package tasker

import (
	"time"

	"github.com/google/uuid"
	"github.com/grassrootseconomics/cic-custodial/pkg/redis"
	"github.com/hibiken/asynq"
)

type TaskerClientOpts struct {
	RedisPool     *redis.RedisPool
	TaskRetention time.Duration
}

type TaskerClient struct {
	Client        *asynq.Client
	taskRetention time.Duration
}

func NewTaskerClient(o TaskerClientOpts) *TaskerClient {
	return &TaskerClient{
		Client: asynq.NewClient(o.RedisPool),
	}
}

func (c *TaskerClient) CreateTask(taskName TaskName, queueName QueueName, task *Task) (*asynq.TaskInfo, error) {
	if task.Id == "" {
		task.Id = uuid.NewString()
	}

	qTask := asynq.NewTask(
		string(taskName),
		task.Payload,
		asynq.Queue(string(queueName)),
		asynq.TaskID(task.Id),
		asynq.Retention(c.taskRetention),
	)

	taskInfo, err := c.Client.Enqueue(qTask)
	if err != nil {
		return nil, err
	}

	return taskInfo, nil
}
