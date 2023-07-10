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
	"github.com/grassrootseconomics/w3-celo-patch/module/eth"
	"github.com/hibiken/asynq"
)

type TransferAuthPayload struct {
	Amount            uint64 `json:"amount"`
	Authorizer        string `json:"authorizer"`
	AuthorizedAddress string `json:"authorizedAddress"`
	TrackingId        string `json:"trackingId"`
	VoucherAddress    string `json:"voucherAddress"`
}

func SignTransferAuthorizationProcessor(cu *custodial.Custodial) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var (
			err            error
			networkBalance big.Int
			payload        TransferAuthPayload
		)

		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return err
		}

		lock, err := cu.LockProvider.Obtain(
			ctx,
			lockPrefix+payload.Authorizer,
			lockTimeout,
			&redislock.Options{
				RetryStrategy: lockRetry(),
			},
		)
		if err != nil {
			return err
		}
		defer lock.Release(ctx)

		key, err := cu.Store.LoadPrivateKey(ctx, payload.Authorizer)
		if err != nil {
			return err
		}

		nonce, err := cu.Noncestore.Acquire(ctx, payload.Authorizer)
		if err != nil {
			return err
		}
		defer func() {
			if err != nil {
				if nErr := cu.Noncestore.Return(ctx, payload.Authorizer); nErr != nil {
					err = nErr
				}
			}
		}()

		input, err := cu.Abis[custodial.Approve].EncodeArgs(
			celoutils.HexToAddress(payload.AuthorizedAddress),
			new(big.Int).SetUint64(payload.Amount),
		)
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

		if err := cu.CeloProvider.Client.CallCtx(
			ctx,
			eth.Balance(celoutils.HexToAddress(payload.Authorizer), nil).Returns(&networkBalance),
		); err != nil {
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

		// Auto-revoke every session (15 min)
		// Check if already a revoke request
		if payload.Amount > 0 {
			taskPayload, err := json.Marshal(TransferAuthPayload{
				TrackingId:        payload.TrackingId,
				Amount:            0,
				Authorizer:        payload.Authorizer,
				AuthorizedAddress: payload.AuthorizedAddress,
				VoucherAddress:    payload.VoucherAddress,
			})
			if err != nil {
				return err
			}

			_, err = cu.TaskerClient.CreateTask(
				ctx,
				tasker.SignTransferTaskAuth,
				tasker.DefaultPriority,
				&tasker.Task{
					Payload: taskPayload,
				},
				asynq.ProcessIn(cu.ApprovalTimeout),
			)
			if err != nil {
				return err
			}
		}

		gasRefillPayload, err := json.Marshal(AccountPayload{
			PublicKey:  payload.Authorizer,
			TrackingId: payload.TrackingId,
		})
		if err != nil {
			return err
		}

		if !balanceCheck(networkBalance) {
			if err := cu.Store.GasLock(ctx, payload.Authorizer); err != nil {
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
		}

		return nil
	}
}
