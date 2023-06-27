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

type TransferAuthPayload struct {
	AuthorizedAddress string `json:"authorizedAddress"`
	AuthorizedAmount  uint64 `json:"authorizedAmount"`
	TrackingId        string `json:"trackingId"`
	VoucherAddress    string `json:"voucherAddress"`
}

func SignTransferAuthorizationProcessor(cu *custodial.Custodial) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var (
			err     error
			payload TransferAuthPayload
		)

		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return err
		}

		lock, err := cu.LockProvider.Obtain(
			ctx,
			lockPrefix+cu.SystemPublicKey,
			lockTimeout,
			&redislock.Options{
				RetryStrategy: lockRetry(),
			},
		)
		if err != nil {
			return err
		}
		defer lock.Release(ctx)

		nonce, err := cu.Noncestore.Acquire(ctx, cu.SystemPublicKey)
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

		input, err := cu.Abis[custodial.Approve].EncodeArgs(
			celoutils.HexToAddress(payload.AuthorizedAddress),
			new(big.Int).SetUint64(payload.AuthorizedAmount),
		)
		if err != nil {
			return err
		}

		builtTx, err := cu.CeloProvider.SignContractExecutionTx(
			cu.SystemPrivateKey,
			celoutils.ContractExecutionTxOpts{
				ContractAddress: celoutils.HexToAddress(payload.VoucherAddress),
				InputData:       input,
				GasFeeCap:       celoutils.SafeGasFeeCap,
				GasTipCap:       celoutils.SafeGasTipCap,
				GasLimit:        uint64(celoutils.SafeGasLimit),
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

		id, err := cu.Store.CreateOtx(ctx, store.Otx{
			TrackingId:    payload.TrackingId,
			Type:          enum.TRANSFER_AUTH,
			RawTx:         hexutil.Encode(rawTx),
			TxHash:        builtTx.Hash().Hex(),
			From:          cu.SystemPublicKey,
			Data:          hexutil.Encode(builtTx.Data()),
			GasPrice:      builtTx.GasPrice(),
			GasLimit:      builtTx.Gas(),
			TransferValue: 0,
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

		return nil
	}
}
