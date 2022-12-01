package task

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/bsm/redislock"
	celo "github.com/grassrootseconomics/cic-celo-sdk"
	"github.com/grassrootseconomics/cic-custodial/internal/nonce"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/w3-celo-patch"
	"github.com/grassrootseconomics/w3-celo-patch/module/eth"
	"github.com/hibiken/asynq"
)

type SystemPayload struct {
	PublicKey string `json:"publicKey"`
}

func PrepareAccount(
	nonceProvider nonce.Noncestore,
	taskerClient *tasker.TaskerClient,
) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var p SystemPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
		}

		if err := nonceProvider.SetNewAccountNonce(ctx, p.PublicKey); err != nil {
			return err
		}

		_, err := taskerClient.CreateTask(
			tasker.GiftGasTask,
			tasker.DefaultPriority,
			&tasker.Task{
				Payload: t.Payload(),
			},
		)
		if err != nil {
			return err
		}

		_, err = taskerClient.CreateTask(
			tasker.GiftTokenTask,
			tasker.DefaultPriority,
			&tasker.Task{
				Payload: t.Payload(),
			},
		)
		if err != nil {
			return err
		}

		return nil
	}
}

func GiftGasProcessor(
	celoProvider *celo.Provider,
	nonceProvider nonce.Noncestore,
	lockProvider *redislock.Client,
	system *tasker.SystemContainer,
	taskerClient *tasker.TaskerClient,
) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var p SystemPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
		}

		lock, err := lockProvider.Obtain(ctx, system.LockPrefix+system.PublicKey, system.LockTimeout, nil)
		if err != nil {
			return err
		}
		defer lock.Release(ctx)

		nonce, err := nonceProvider.Acquire(ctx, system.PublicKey)
		if err != nil {
			return err
		}

		builtTx, err := celoProvider.SignGasTransferTx(
			system.PrivateKey,
			celo.GasTransferTxOpts{
				To:       w3.A(p.PublicKey),
				Nonce:    nonce,
				Value:    system.GiftableGasValue,
				GasPrice: celo.FixedMinGas,
			},
		)
		if err != nil {
			if err := nonceProvider.Return(ctx, p.PublicKey); err != nil {
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

		return nil
	}
}

func GiftTokenProcessor(
	celoProvider *celo.Provider,
	nonceProvider nonce.Noncestore,
	lockProvider *redislock.Client,
	system *tasker.SystemContainer,
	taskerClient *tasker.TaskerClient,
) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var p SystemPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
		}

		publicKey := w3.A(p.PublicKey)

		lock, err := lockProvider.Obtain(ctx, system.LockPrefix+system.PublicKey, system.LockTimeout, nil)
		if err != nil {
			return err
		}
		defer lock.Release(ctx)

		nonce, err := nonceProvider.Acquire(ctx, system.PublicKey)
		if err != nil {
			return err
		}

		input, err := system.Abis["mint"].EncodeArgs(publicKey, system.GiftableTokenValue)
		if err != nil {
			return fmt.Errorf("ABI encode failed %v: %w", err, asynq.SkipRetry)
		}

		builtTx, err := celoProvider.SignContractExecutionTx(
			system.PrivateKey,
			celo.ContractExecutionTxOpts{
				ContractAddress: system.GiftableToken,
				InputData:       input,
				GasPrice:        celo.FixedMinGas,
				GasLimit:        system.TokenTransferGasLimit,
				Nonce:           nonce,
			},
		)
		if err != nil {
			if err := nonceProvider.Return(ctx, p.PublicKey); err != nil {
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

		return nil
	}
}

func RefillGasProcessor(
	celoProvider *celo.Provider,
	nonceProvider nonce.Noncestore,
	lockProvider *redislock.Client,
	system *tasker.SystemContainer,
	taskerClient *tasker.TaskerClient,
) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var (
			p       SystemPayload
			balance big.Int
		)
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
		}

		if err := celoProvider.Client.CallCtx(
			ctx,
			eth.Balance(w3.A(p.PublicKey), nil).Returns(&balance),
		); err != nil {
			return err
		}

		if belowThreshold := balance.Cmp(system.GasRefillThreshold); belowThreshold > 0 {
			return nil
		}

		lock, err := lockProvider.Obtain(ctx, system.LockPrefix+system.PublicKey, system.LockTimeout, nil)
		if err != nil {
			return err
		}
		defer lock.Release(ctx)

		nonce, err := nonceProvider.Acquire(ctx, system.PublicKey)
		if err != nil {
			return err
		}

		builtTx, err := celoProvider.SignGasTransferTx(
			system.PrivateKey,
			celo.GasTransferTxOpts{
				To:       w3.A(p.PublicKey),
				Nonce:    nonce,
				Value:    system.GasRefillValue,
				GasPrice: celo.FixedMinGas,
			},
		)
		if err != nil {
			if err := nonceProvider.Return(ctx, p.PublicKey); err != nil {
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

		return nil
	}
}
