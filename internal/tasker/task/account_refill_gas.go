package task

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/celo-org/celo-blockchain/common/hexutil"
	"github.com/grassrootseconomics/celoutils"
	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/grassrootseconomics/cic-custodial/internal/events"
	"github.com/grassrootseconomics/cic-custodial/internal/store"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/cic-custodial/pkg/enum"
	"github.com/grassrootseconomics/w3-celo-patch"
	"github.com/grassrootseconomics/w3-celo-patch/module/eth"
	"github.com/hibiken/asynq"
)

func AccountRefillGasProcessor(cu *custodial.Custodial) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var (
			payload AccountPayload
			balance big.Int
		)

		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return fmt.Errorf("account: failed %v: %w", err, asynq.SkipRetry)
		}

		if err := cu.CeloProvider.Client.CallCtx(
			ctx,
			eth.Balance(w3.A(payload.PublicKey), nil).Returns(&balance),
		); err != nil {
			return err
		}

		if belowThreshold := balance.Cmp(cu.SystemContainer.GasRefillThreshold); belowThreshold > 0 {
			return nil
		}

		lock, err := cu.LockProvider.Obtain(
			ctx,
			cu.SystemContainer.LockPrefix+cu.SystemContainer.PublicKey,
			cu.SystemContainer.LockTimeout,
			nil,
		)
		if err != nil {
			return err
		}
		defer lock.Release(ctx)

		nonce, err := cu.Noncestore.Acquire(ctx, cu.SystemContainer.PublicKey)
		if err != nil {
			return err
		}
		defer func() {
			if err != nil {
				if nErr := cu.Noncestore.Return(ctx, cu.SystemContainer.PublicKey); nErr != nil {
					err = nErr
				}
			}
		}()

		// TODO: Review gas params
		builtTx, err := cu.CeloProvider.SignGasTransferTx(
			cu.SystemContainer.PrivateKey,
			celoutils.GasTransferTxOpts{
				To:        w3.A(payload.PublicKey),
				Nonce:     nonce,
				Value:     cu.SystemContainer.GiftableGasValue,
				GasFeeCap: celoutils.SafeGasFeeCap,
				GasTipCap: celoutils.SafeGasTipCap,
			},
		)
		if err != nil {
			return err
		}

		rawTx, err := builtTx.MarshalBinary()
		if err != nil {
			return err
		}

		id, err := cu.PgStore.CreateOtx(ctx, store.OTX{
			TrackingId:    payload.TrackingId,
			Type:          enum.REFILL_GAS,
			RawTx:         hexutil.Encode(rawTx),
			TxHash:        builtTx.Hash().Hex(),
			From:          cu.SystemContainer.PublicKey,
			Data:          hexutil.Encode(builtTx.Data()),
			GasPrice:      builtTx.GasPrice().Uint64(),
			GasLimit:      builtTx.Gas(),
			TransferValue: cu.SystemContainer.GiftableGasValue.Uint64(),
			Nonce:         builtTx.Nonce(),
		})
		if err != nil {
			return err
		}

		disptachJobPayload, err := json.Marshal(TxPayload{
			OtxId: id,
			Tx:    builtTx,
		})
		if err != nil {
			return err
		}

		_, err = cu.TaskerClient.CreateTask(
			tasker.DispatchTxTask,
			tasker.HighPriority,
			&tasker.Task{
				Payload: disptachJobPayload,
			},
		)
		if err != nil {
			return err
		}

		eventPayload := &events.EventPayload{
			OtxId:      id,
			TrackingId: payload.TrackingId,
			TxHash:     builtTx.Hash().Hex(),
		}

		if err := cu.EventEmitter.Publish(
			events.AccountRefillGas,
			builtTx.Hash().Hex(),
			eventPayload,
		); err != nil {
			return err
		}

		return nil
	}
}
