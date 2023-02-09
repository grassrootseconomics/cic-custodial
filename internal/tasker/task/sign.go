package task

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"math/big"

	"github.com/bsm/redislock"
	celo "github.com/grassrootseconomics/cic-celo-sdk"
	"github.com/grassrootseconomics/cic-custodial/internal/keystore"
	"github.com/grassrootseconomics/cic-custodial/internal/nonce"
	"github.com/grassrootseconomics/cic-custodial/internal/store"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/w3-celo-patch"
	"github.com/hibiken/asynq"
	"github.com/nats-io/nats.go"
)

type (
	TransferPayload struct {
		TrackingId     string `json:"trackingId"`
		From           string `json:"from" `
		To             string `json:"to"`
		VoucherAddress string `json:"voucherAddress"`
		Amount         int64  `json:"amount"`
	}

	transferEventPayload struct {
		DispatchTaskId string `json:"dispatchTaskId"`
		OTXId          uint   `json:"otxId"`
		TrackingId     string `json:"trackingId"`
		TxHash         string `json:"txHash"`
	}
)

func SignTransfer(
	celoProvider *celo.Provider,
	keystore keystore.Keystore,
	lockProvider *redislock.Client,
	noncestore nonce.Noncestore,
	pg store.Store,
	system *tasker.SystemContainer,
	taskerClient *tasker.TaskerClient,
	js nats.JetStreamContext,
) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var (
			p TransferPayload
		)

		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return err
		}

		lock, err := lockProvider.Obtain(
			ctx,
			system.LockPrefix+p.From,
			system.LockTimeout,
			nil,
		)
		if err != nil {
			return err
		}
		defer lock.Release(ctx)

		key, err := keystore.LoadPrivateKey(ctx, p.From)
		if err != nil {
			return err
		}

		nonce, err := noncestore.Acquire(ctx, p.From)
		if err != nil {
			return err
		}

		input, err := system.Abis["transfer"].EncodeArgs(w3.A(p.To), big.NewInt(p.Amount))
		if err != nil {
			return err
		}

		// TODO: Review gas params.
		builtTx, err := celoProvider.SignContractExecutionTx(
			key,
			celo.ContractExecutionTxOpts{
				ContractAddress: w3.A(p.VoucherAddress),
				InputData:       input,
				GasPrice:        big.NewInt(20000000000),
				GasLimit:        system.TokenTransferGasLimit,
				Nonce:           nonce,
			},
		)
		if err != nil {
			if err := noncestore.Return(ctx, p.From); err != nil {
				return err
			}

			return err
		}

		rawTx, err := builtTx.MarshalBinary()
		if err != nil {
			if err := noncestore.Return(ctx, p.From); err != nil {
				return err
			}

			return err
		}

		id, err := pg.CreateOTX(ctx, store.OTX{
			RawTx:    hex.EncodeToString(rawTx),
			TxHash:   builtTx.Hash().Hex(),
			From:     p.From,
			Data:     string(builtTx.Data()),
			GasPrice: builtTx.GasPrice().Uint64(),
			Nonce:    builtTx.Nonce(),
		})
		if err != nil {
			if err := noncestore.Return(ctx, p.From); err != nil {
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
			if err := noncestore.Return(ctx, p.From); err != nil {
				return err
			}

			return err
		}

		dispatchTask, err := taskerClient.CreateTask(
			tasker.TxDispatchTask,
			tasker.HighPriority,
			&tasker.Task{
				Payload: disptachJobPayload,
			},
		)
		if err != nil {
			if err := noncestore.Return(ctx, p.From); err != nil {
				return err
			}

			return err
		}

		eventPayload := &transferEventPayload{
			DispatchTaskId: dispatchTask.ID,
			OTXId:          id,
			TrackingId:     p.TrackingId,
			TxHash:         builtTx.Hash().Hex(),
		}

		eventJson, err := json.Marshal(eventPayload)
		if err != nil {
			if err := noncestore.Return(ctx, p.From); err != nil {
				return err
			}

			return err
		}

		_, err = js.Publish("CUSTODIAL.transferSign", eventJson, nats.MsgId(builtTx.Hash().Hex()))
		if err != nil {
			if err := noncestore.Return(ctx, p.From); err != nil {
				return err
			}

			return err
		}

		return nil
	}
}
