package task

import (
	"context"
	"encoding/json"

	"github.com/celo-org/celo-blockchain/common/hexutil"
	"github.com/grassrootseconomics/celoutils"
	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/grassrootseconomics/cic-custodial/internal/events"
	"github.com/grassrootseconomics/cic-custodial/internal/store"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/w3-celo-patch"
	"github.com/hibiken/asynq"
)

func GiftVoucherProcessor(cu *custodial.Custodial) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var (
			payload AccountPayload
		)

		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return err
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

		input, err := cu.SystemContainer.Abis["mintTo"].EncodeArgs(
			w3.A(payload.PublicKey),
			cu.SystemContainer.GiftableTokenValue,
		)
		if err != nil {
			return err
		}

		builtTx, err := cu.CeloProvider.SignContractExecutionTx(
			cu.SystemContainer.PrivateKey,
			celoutils.ContractExecutionTxOpts{
				ContractAddress: cu.SystemContainer.GiftableToken,
				InputData:       input,
				GasFeeCap:       celoutils.SafeGasFeeCap,
				GasTipCap:       celoutils.SafeGasTipCap,
				GasLimit:        cu.SystemContainer.TokenTransferGasLimit,
				Nonce:           nonce,
			},
		)
		if err != nil {
			return err
		}

		rawTx, err := builtTx.MarshalBinary()
		if err != nil {
			return err
		}

		id, err := cu.PgStore.CreateOTX(ctx, store.OTX{
			TrackingId: payload.TrackingId,
			Type:       "GIFT_VOUCHER",
			RawTx:      hexutil.Encode(rawTx),
			TxHash:     builtTx.Hash().Hex(),
			From:       cu.SystemContainer.PublicKey,
			Data:       hexutil.Encode(builtTx.Data()),
			GasPrice:   builtTx.GasPrice().Uint64(),
			Nonce:      builtTx.Nonce(),
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
			events.AccountGiftVoucher,
			builtTx.Hash().Hex(),
			eventPayload,
		); err != nil {
			return err
		}

		return nil
	}
}
