package client

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const (
	ActivateAccountTask    JobType = "registration:activate"
	GiftGasTask            JobType = "registration:gift_gas"
	SetNewAccountNonceTask JobType = "registration:set_new_account_nonce"
)

type RegistrationPayload struct {
	PublicKey string
}

func (tc *TaskerClient) CreateRegistrationTask(taskPayload RegistrationPayload, jobType JobType) (*asynq.TaskInfo, error) {
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
