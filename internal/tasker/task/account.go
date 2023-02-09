package task

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/bsm/redislock"
	celo "github.com/grassrootseconomics/cic-celo-sdk"
	"github.com/grassrootseconomics/cic-custodial/internal/nonce"
	"github.com/grassrootseconomics/cic-custodial/internal/store"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/w3-celo-patch"
	"github.com/grassrootseconomics/w3-celo-patch/module/eth"
	"github.com/hibiken/asynq"
	"github.com/nats-io/nats.go"
)

type (
	AccountPayload struct {
		PublicKey  string `json:"publicKey"`
		TrackingId string `json:"trackingId"`
	}

	accountEventPayload struct {
		TrackingId string `json:"trackingId"`
	}
)

func PrepareAccount(
	js nats.JetStreamContext,
	noncestore nonce.Noncestore,
	taskerClient *tasker.TaskerClient,
) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var (
			p AccountPayload
		)

		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return err
		}

		if err := noncestore.SetNewAccountNonce(ctx, p.PublicKey); err != nil {
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

		eventPayload := &accountEventPayload{
			TrackingId: p.TrackingId,
		}

		eventJson, err := json.Marshal(eventPayload)
		if err != nil {
			return err
		}

		_, err = js.Publish("CUSTODIAL.accountNewNonce", eventJson)
		if err != nil {
			return err
		}

		return nil
	}
}

func GiftGasProcessor(
	celoProvider *celo.Provider,
	js nats.JetStreamContext,
	lockProvider *redislock.Client,
	noncestore nonce.Noncestore,
	pg store.Store,
	system *tasker.SystemContainer,
	taskerClient *tasker.TaskerClient,
) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var (
			p AccountPayload
		)

		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return err
		}

		lock, err := lockProvider.Obtain(ctx, system.LockPrefix+system.PublicKey, system.LockTimeout, nil)
		if err != nil {
			return err
		}
		defer lock.Release(ctx)

		nonce, err := noncestore.Acquire(ctx, system.PublicKey)
		if err != nil {
			return err
		}

		// TODO: Review gas params
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
			if err := noncestore.Return(ctx, system.PublicKey); err != nil {
				return err
			}

			return err
		}

		rawTx, err := builtTx.MarshalBinary()
		if err != nil {
			if err := noncestore.Return(ctx, system.PublicKey); err != nil {
				return err
			}

			return err
		}

		id, err := pg.CreateOTX(ctx, store.OTX{
			RawTx:    hex.EncodeToString(rawTx),
			TxHash:   builtTx.Hash().Hex(),
			From:     system.PublicKey,
			Data:     string(builtTx.Data()),
			GasPrice: builtTx.GasPrice().Uint64(),
			Nonce:    builtTx.Nonce(),
		})
		if err != nil {
			if err := noncestore.Return(ctx, system.PublicKey); err != nil {
				return err
			}

			return err
		}

		disptachJobPayload, err := json.Marshal(TxPayload{
			OtxId:      id,
			TrackingId: p.TrackingId,
			Tx:         builtTx,
		})
		if err != nil {
			if err := noncestore.Return(ctx, system.PublicKey); err != nil {
				return err
			}

			return err
		}

		_, err = taskerClient.CreateTask(
			tasker.TxDispatchTask,
			tasker.HighPriority,
			&tasker.Task{
				Payload: disptachJobPayload,
			},
		)
		if err != nil {
			if err := noncestore.Return(ctx, system.PublicKey); err != nil {
				return err
			}

			return err
		}

		eventPayload := &accountEventPayload{
			TrackingId: p.TrackingId,
		}

		eventJson, err := json.Marshal(eventPayload)
		if err != nil {
			return err
		}

		_, err = js.Publish("CUSTODIAL.giftNewAccountGas", eventJson)
		if err != nil {
			if err := noncestore.Return(ctx, system.PublicKey); err != nil {
				return err
			}

			return err
		}

		return nil
	}
}

func GiftTokenProcessor(
	celoProvider *celo.Provider,
	js nats.JetStreamContext,
	lockProvider *redislock.Client,
	noncestore nonce.Noncestore,
	pg store.Store,
	system *tasker.SystemContainer,
	taskerClient *tasker.TaskerClient,
) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var (
			p AccountPayload
		)

		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return err
		}

		lock, err := lockProvider.Obtain(ctx, system.LockPrefix+system.PublicKey, system.LockTimeout, nil)
		if err != nil {
			return err
		}
		defer lock.Release(ctx)

		nonce, err := noncestore.Acquire(ctx, system.PublicKey)
		if err != nil {
			return err
		}

		input, err := system.Abis["mintTo"].EncodeArgs(w3.A(p.PublicKey), system.GiftableTokenValue)
		if err != nil {
			return err
		}

		// TODO: Review gas params.
		builtTx, err := celoProvider.SignContractExecutionTx(
			system.PrivateKey,
			celo.ContractExecutionTxOpts{
				ContractAddress: system.GiftableToken,
				InputData:       input,
				GasPrice:        big.NewInt(20000000000),
				GasLimit:        system.TokenTransferGasLimit,
				Nonce:           nonce,
			},
		)
		if err != nil {
			if err := noncestore.Return(ctx, system.PublicKey); err != nil {
				return err
			}
			return err
		}

		rawTx, err := builtTx.MarshalBinary()
		if err != nil {
			if err := noncestore.Return(ctx, system.PublicKey); err != nil {
				return err
			}

			return err
		}

		id, err := pg.CreateOTX(ctx, store.OTX{
			RawTx:    hex.EncodeToString(rawTx),
			TxHash:   builtTx.Hash().Hex(),
			From:     system.PublicKey,
			Data:     string(builtTx.Data()),
			GasPrice: builtTx.GasPrice().Uint64(),
			Nonce:    builtTx.Nonce(),
		})
		if err != nil {
			if err := noncestore.Return(ctx, system.PublicKey); err != nil {
				return err
			}

			return err
		}

		disptachJobPayload, err := json.Marshal(TxPayload{
			OtxId:      id,
			TrackingId: p.TrackingId,
			Tx:         builtTx,
		})
		if err != nil {
			if err := noncestore.Return(ctx, system.PublicKey); err != nil {
				return err
			}

			return err
		}

		_, err = taskerClient.CreateTask(
			tasker.TxDispatchTask,
			tasker.HighPriority,
			&tasker.Task{
				Payload: disptachJobPayload,
			},
		)
		if err != nil {
			if err := noncestore.Return(ctx, system.PublicKey); err != nil {
				return err
			}

			return err
		}

		eventPayload := &accountEventPayload{
			TrackingId: p.TrackingId,
		}

		eventJson, err := json.Marshal(eventPayload)
		if err != nil {
			return err
		}

		_, err = js.Publish("CUSTODIAL.giftNewAccountVoucher", eventJson)
		if err != nil {
			if err := noncestore.Return(ctx, system.PublicKey); err != nil {
				return err
			}

			return err
		}

		return nil
	}
}

// TODO: https://github.com/grassrootseconomics/cic-custodial/issues/43
// TODO:
func RefillGasProcessor(
	celoProvider *celo.Provider,
	nonceProvider nonce.Noncestore,
	lockProvider *redislock.Client,
	system *tasker.SystemContainer,
	taskerClient *tasker.TaskerClient,
) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var (
			p       AccountPayload
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
