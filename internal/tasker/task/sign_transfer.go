package task

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/celo-org/celo-blockchain/common/hexutil"
	"github.com/grassrootseconomics/celoutils"
	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/grassrootseconomics/cic-custodial/internal/pub"
	"github.com/grassrootseconomics/cic-custodial/internal/store"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/cic-custodial/pkg/enum"
	"github.com/grassrootseconomics/w3-celo-patch"
	"github.com/hibiken/asynq"
)

type (
	TransferPayload struct {
		TrackingId     string `json:"trackingId"`
		From           string `json:"from" `
		To             string `json:"to"`
		VoucherAddress string `json:"voucherAddress"`
		Amount         uint64 `json:"amount"`
	}

	transferEventPayload struct {
		DispatchTaskId string `json:"dispatchTaskId"`
		OTXId          uint   `json:"otxId"`
		TrackingId     string `json:"trackingId"`
		TxHash         string `json:"txHash"`
	}
)

func SignTransfer(cu *custodial.Custodial) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var (
			err     error
			payload TransferPayload
		)

		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return fmt.Errorf("account: failed %v: %w", err, asynq.SkipRetry)
		}

		lock, err := cu.LockProvider.Obtain(
			ctx,
			cu.SystemContainer.LockPrefix+payload.From,
			cu.SystemContainer.LockTimeout,
			nil,
		)
		if err != nil {
			return err
		}
		defer lock.Release(ctx)

		key, err := cu.Keystore.LoadPrivateKey(ctx, payload.From)
		if err != nil {
			return err
		}

		nonce, err := cu.Noncestore.Acquire(ctx, payload.From)
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

		input, err := cu.SystemContainer.Abis["transfer"].EncodeArgs(w3.A(payload.To), new(big.Int).SetUint64(payload.Amount))
		if err != nil {
			return err
		}

		// TODO: Review gas params.
		builtTx, err := cu.CeloProvider.SignContractExecutionTx(
			key,
			celoutils.ContractExecutionTxOpts{
				ContractAddress: w3.A(payload.VoucherAddress),
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

		id, err := cu.PgStore.CreateOtx(ctx, store.OTX{
			TrackingId:    payload.TrackingId,
			Type:          enum.TRANSFER_VOUCHER,
			RawTx:         hexutil.Encode(rawTx),
			TxHash:        builtTx.Hash().Hex(),
			From:          payload.From,
			Data:          hexutil.Encode(builtTx.Data()),
			GasPrice:      builtTx.GasPrice().Uint64(),
			GasLimit:      builtTx.Gas(),
			TransferValue: payload.Amount,
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
			ctx,
			tasker.DispatchTxTask,
			tasker.HighPriority,
			&tasker.Task{
				Payload: disptachJobPayload,
			},
		)
		if err != nil {
			return err
		}

		gasRefillPayload, err := json.Marshal(AccountPayload{
			PublicKey: payload.From,
		})
		if err != nil {
			return err
		}

		_, err = cu.TaskerClient.CreateTask(
			ctx,
			tasker.AccountRefillGasTask,
			tasker.DefaultPriority,
			&tasker.Task{
				Payload: gasRefillPayload,
			},
		)
		if err != nil {
			return err
		}

		eventPayload := &transferEventPayload{
			OTXId:      id,
			TrackingId: payload.TrackingId,
			TxHash:     builtTx.Hash().Hex(),
		}

		if err := cu.Pub.Publish(
			pub.SignTransfer,
			builtTx.Hash().Hex(),
			eventPayload,
		); err != nil {
			return err
		}

		return nil
	}
}
