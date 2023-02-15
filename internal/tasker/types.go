package tasker

import (
	"crypto/ecdsa"
	"encoding/json"
	"math/big"
	"time"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/w3-celo-patch"
)

type (
	QueueName string
	TaskName  string
)

type SystemContainer struct {
	Abis                  map[string]*w3.Func
	AccountIndexContract  common.Address
	GasFaucetContract     common.Address
	GasRefillThreshold    *big.Int
	GasRefillValue        *big.Int
	GiftableGasValue      *big.Int
	GiftableToken         common.Address
	GiftableTokenValue    *big.Int
	LockPrefix            string
	LockTimeout           time.Duration
	PrivateKey            *ecdsa.PrivateKey
	PublicKey             string
	TokenDecimals         int
	TokenTransferGasLimit uint64
}

type Task struct {
	Id      string          `json:"id"`
	Payload json.RawMessage `json:"payload"`
}

const (
	PrepareAccountTask     TaskName = "sys:prepare_account"
	RegisterAccountOnChain TaskName = "sys:register_account"
	GiftGasTask            TaskName = "sys:gift_gas"
	GiftTokenTask          TaskName = "sys:gift_token"
	RefillGasTask          TaskName = "admin:refill_gas"
	SweepGasTask           TaskName = "admin:sweep_gas"
	AdminTokenApprovalTask TaskName = "admin:token_approval"
	SignTransferTask       TaskName = "usr:sign_transfer"
	TxDispatchTask         TaskName = "rpc:dispatch"
)

const (
	HighPriority    QueueName = "high_priority"
	DefaultPriority QueueName = "default_priority"
)
