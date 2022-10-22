package client

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hibiken/asynq"
)

const (
	TxDispatchTask JobType = "tx:dispatch"
)

type TxPayload struct {
	Tx *types.Transaction
}

func (tc *TaskerClient) CreateTxDispatchTask(taskPayload TxPayload, jobType JobType) (*asynq.TaskInfo, error) {
	payload, err := json.Marshal(taskPayload)
	if err != nil {
		return nil, err
	}

	task := asynq.NewTask(string(jobType), payload, asynq.Retention(defaultRetentionPeriod))

	t, err := tc.Client.Enqueue(task)
	if err != nil {
		return nil, err
	}

	return t, nil
}
