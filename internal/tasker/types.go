package tasker

import (
	"encoding/json"
)

type (
	QueueName string
	TaskName  string
)

type Task struct {
	Id      string          `json:"id"`
	Payload json.RawMessage `json:"payload"`
}

const (
	AccountRegisterTask  TaskName = "sys:register_account"
	AccountRefillGasTask TaskName = "sys:refill_gas"
	SignTransferTask     TaskName = "usr:sign_transfer"
	SignTransferTaskAuth TaskName = "usr:sign_transfer_auth"
	DispatchTxTask       TaskName = "rpc:dispatch"
)

const (
	HighPriority    QueueName = "high_priority"
	DefaultPriority QueueName = "default_priority"
)
