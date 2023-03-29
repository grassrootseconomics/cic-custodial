package task

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/bsm/redislock"
	"github.com/celo-org/celo-blockchain/common/hexutil"
	"github.com/grassrootseconomics/celoutils"
	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/grassrootseconomics/cic-custodial/internal/store"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/cic-custodial/pkg/enum"
	"github.com/hibiken/asynq"
)

type TransferPayload struct {
	TrackingId     string `json:"trackingId"`
	From           string `json:"from" `
	To             string `json:"to"`
	VoucherAddress string `json:"voucherAddress"`
	Amount         uint64 `json:"amount"`
}

func SignTransfer(cu *custodial.Custodial) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var (
			err     error
			payload TransferPayload
		)

		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return err
		}

		lock, err := cu.LockProvider.Obtain(
			ctx,
			lockPrefix+payload.From,
			lockTimeout,
			&redislock.Options{
				RetryStrategy: lockRetry(),
			},
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
				if nErr := cu.Noncestore.Return(ctx, cu.SystemPublicKey); nErr != nil {
					err = nErr
				}
			}
		}()

		input, err := cu.Abis[custodial.Transfer].EncodeArgs(celoutils.HexToAddress(payload.To), new(big.Int).SetUint64(payload.Amount))
		if err != nil {
			return err
		}

		builtTx, err := cu.CeloProvider.SignContractExecutionTx(
			key,
			celoutils.ContractExecutionTxOpts{
				ContractAddress: celoutils.HexToAddress(payload.VoucherAddress),
				InputData:       input,
				GasFeeCap:       celoutils.SafeGasFeeCap,
				GasTipCap:       celoutils.SafeGasTipCap,
				GasLimit:        gasLimit,
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

		if err := cu.PgStore.DecrGasQuota(ctx, payload.From); err != nil {
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
			PublicKey:  payload.From,
			TrackingId: payload.TrackingId,
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

		return nil
	}
}
