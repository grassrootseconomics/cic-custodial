package task

import (
	"context"
	"encoding/json"
	"math/big"
	"time"

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

const (
	gasGiveToLimit = 250000
)

func AccountRefillGasProcessor(cu *custodial.Custodial) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var (
			err     error
			payload AccountPayload

			nextTime    big.Int
			checkStatus bool
		)

		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return err
		}

		if err := cu.CeloProvider.Client.CallCtx(
			ctx,
			eth.CallFunc(
				cu.Abis[custodial.NextTime],
				cu.RegistryMap[celoutils.GasFaucet],
				celoutils.HexToAddress(payload.PublicKey),
			).Returns(&nextTime),
		); err != nil {
			return err
		}

		// The user recently requested funds, there is a cooldown applied.
		// We can schedule an attempt after the cooldown period has passed + 10 seconds.
		if nextTime.Int64() > time.Now().Unix() {
			_, err = cu.TaskerClient.CreateTask(
				ctx,
				tasker.AccountRefillGasTask,
				tasker.DefaultPriority,
				&tasker.Task{
					Payload: t.Payload(),
				},
				asynq.ProcessAt(time.Unix(nextTime.Int64()+10, 0)),
			)
			if err != nil {
				return err
			}

			return nil
		}

		if err := cu.CeloProvider.Client.CallCtx(
			ctx,
			eth.CallFunc(
				cu.Abis[custodial.Check],
				cu.RegistryMap[celoutils.GasFaucet],
				celoutils.HexToAddress(payload.PublicKey),
			).Returns(&checkStatus),
		); err != nil {
			return err
		}

		// The gas faucet backend returns a false status, a poke will fail.
		if !checkStatus {
			return nil
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

		input, err := cu.Abis[custodial.GiveTo].EncodeArgs(
			celoutils.HexToAddress(payload.PublicKey),
		)
		if err != nil {
			return err
		}

		builtTx, err := cu.CeloProvider.SignContractExecutionTx(
			cu.SystemPrivateKey,
			celoutils.ContractExecutionTxOpts{
				ContractAddress: cu.RegistryMap[celoutils.GasFaucet],
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
			Type:       enum.REFILL_GAS,
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

		return nil
	}
}
