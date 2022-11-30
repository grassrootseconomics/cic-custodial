package task

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/bsm/redislock"
	celo "github.com/grassrootseconomics/cic-celo-sdk"
	"github.com/grassrootseconomics/cic-custodial/internal/keystore"
	"github.com/grassrootseconomics/cic-custodial/internal/nonce"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/w3-celo-patch"
	"github.com/hibiken/asynq"
)

type TransferPayload struct {
	From           string `json:"from"`
	To             string `json:"to"`
	VoucherAddress string `json:"voucherAddress"`
	Amount         string `json:"amount"`
}

func TransferToken(
	celoProvider *celo.Provider,
	nonceProvider nonce.Noncestore,
	keystoreProvider keystore.Keystore,
	lockProvider *redislock.Client,
	system *tasker.SystemContainer,
	taskerClient *tasker.TaskerClient,
) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var p TransferPayload

		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
		}

		lock, err := lockProvider.Obtain(ctx, system.LockPrefix+p.From, system.LockTimeout, nil)
		if err != nil {
			return err
		}
		defer lock.Release(ctx)

		nonce, err := nonceProvider.Acquire(ctx, p.From)
		if err != nil {
			return err
		}

		key, err := keystoreProvider.LoadPrivateKey(ctx, p.From)
		if err != nil {
			return err
		}

		abi, err := w3.NewFunc("transfer(address,uint256)", "bool")
		if err != nil {
			return err
		}

		input, err := abi.EncodeArgs(p.To, parseTransferValue(p.Amount, system.TokenDecimals))
		if err != nil {
			return fmt.Errorf("ABI encode failed %v: %w", err, asynq.SkipRetry)
		}

		builtTx, err := celoProvider.SignContractExecutionTx(
			key,
			celo.ContractExecutionTxOpts{
				ContractAddress: system.GiftableToken,
				InputData:       input,
				GasPrice:        celo.FixedMinGas,
				GasLimit:        system.TokenTransferGasLimit,
				Nonce:           nonce,
			},
		)
		if err != nil {
			if err := nonceProvider.Return(ctx, p.From); err != nil {
				return err
			}
			return fmt.Errorf("nonce.Return failed: %v: %w", err, asynq.SkipRetry)
		}

		disptachJobPayload, err := json.Marshal(TxPayload{
			Tx: builtTx,
		})
		if err != nil {
			return fmt.Errorf("json.Marshal failed: %v: %w", err, asynq.SkipRetry)
		}

		_, err = taskerClient.CreateTask(
			tasker.TxDispatchTask,
			tasker.HighPriority,
			&tasker.Task{
				Payload: disptachJobPayload,
			},
		)
		if err != nil {
			return err
		}

		gasRefillPayload, err := json.Marshal(SystemPayload{
			PublicKey: p.From,
		})
		if err != nil {
			return fmt.Errorf("json.Marshal failed: %v: %w", err, asynq.SkipRetry)
		}

		_, err = taskerClient.CreateTask(
			tasker.RefillGasTask,
			tasker.DefaultPriority,
			&tasker.Task{
				Payload: gasRefillPayload,
			},
		)
		if err != nil {
			return err
		}

		return nil
	}
}

func parseTransferValue(value string, tokenDecimals int) *big.Int {
	floatValue, _ := strconv.ParseFloat(value, 64)

	return big.NewInt(int64(floatValue * math.Pow10(tokenDecimals)))
}
