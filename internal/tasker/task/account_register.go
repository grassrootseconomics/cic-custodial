package task

import (
	"context"
	"encoding/json"

	"github.com/bsm/redislock"
	"github.com/celo-org/celo-blockchain/common/hexutil"
	"github.com/grassrootseconomics/celoutils"
	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/grassrootseconomics/cic-custodial/internal/store"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/cic-custodial/pkg/enum"
	"github.com/hibiken/asynq"
)

type AccountPayload struct {
	PublicKey  string `json:"publicKey"`
	TrackingId string `json:"trackingId"`
}

func AccountRegisterOnChainProcessor(cu *custodial.Custodial) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var (
			err     error
			payload AccountPayload
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

		input, err := cu.Abis[custodial.Register].EncodeArgs(
			celoutils.HexToAddress(payload.PublicKey),
		)
		if err != nil {
			return err
		}

		builtTx, err := cu.CeloProvider.SignContractExecutionTx(
			cu.SystemPrivateKey,
			celoutils.ContractExecutionTxOpts{
				ContractAddress: cu.RegistryMap[celoutils.CustodialProxy],
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
			TrackingId: payload.TrackingId,
			Type:       enum.ACCOUNT_REGISTER,
			RawTx:      hexutil.Encode(rawTx),
			TxHash:     builtTx.Hash().Hex(),
			From:       cu.SystemPublicKey,
			Data:       hexutil.Encode(builtTx.Data()),
			GasPrice:   builtTx.GasPrice(),
			GasLimit:   builtTx.Gas(),
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

		if err := cu.Noncestore.SetAccountNonce(ctx, payload.PublicKey, 0); err != nil {
			return err
		}

		return nil
	}
}
