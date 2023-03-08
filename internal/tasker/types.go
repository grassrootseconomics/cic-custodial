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
	AccountPrepareTask     TaskName = "sys:prepare_account"
	AccountRegisterTask    TaskName = "sys:register_account"
	AccountGiftGasTask     TaskName = "sys:gift_gas"
	AccountGiftVoucherTask TaskName = "sys:gift_token"
	AccountRefillGasTask   TaskName = "sys:refill_gas"
	AccountActivateTask    TaskName = "sys:quorum_check"
	SignTransferTask       TaskName = "usr:sign_transfer"
	DispatchTxTask         TaskName = "rpc:dispatch"
)

const (
	HighPriority    QueueName = "high_priority"
	DefaultPriority QueueName = "default_priority"
)
