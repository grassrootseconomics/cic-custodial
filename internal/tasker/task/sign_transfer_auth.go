package task

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/bsm/redislock"
	"github.com/celo-org/celo-blockchain/accounts/abi"
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
	AuthorizeFor      string `json:"authorizeFor"`
	AuthorizedAddress string `json:"authorizedAddress"`
	Revoke            bool   `json:"revoke"`
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
			lockPrefix+payload.AuthorizeFor,
			lockTimeout,
			&redislock.Options{
				RetryStrategy: lockRetry(),
			},
		)
		if err != nil {
			return err
		}
		defer lock.Release(ctx)

		key, err := cu.Store.LoadPrivateKey(ctx, payload.AuthorizeFor)
		if err != nil {
			return err
		}

		nonce, err := cu.Noncestore.Acquire(ctx, payload.AuthorizeFor)
		if err != nil {
			return err
		}
		defer func() {
			if err != nil {
				if nErr := cu.Noncestore.Return(ctx, payload.AuthorizeFor); nErr != nil {
					err = nErr
				}
			}
		}()

		authorizeAmount := big.NewInt(0).Sub(abi.MaxUint256, big.NewInt(1))
		if payload.Revoke {
			authorizeAmount = big.NewInt(0)
		}

		input, err := cu.Abis[custodial.Approve].EncodeArgs(
			celoutils.HexToAddress(payload.AuthorizedAddress),
			authorizeAmount,
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
			eth.Balance(celoutils.HexToAddress(payload.AuthorizeFor), nil).Returns(&networkBalance),
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

		gasRefillPayload, err := json.Marshal(AccountPayload{
			PublicKey:  payload.AuthorizeFor,
			TrackingId: payload.TrackingId,
		})
		if err != nil {
			return err
		}

		if !balanceCheck(networkBalance) {
			if err := cu.Store.GasLock(ctx, payload.AuthorizeFor); err != nil {
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
