package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	celo "github.com/grassrootseconomics/cic-celo-sdk"
	"github.com/grassrootseconomics/cic-custodial/internal/store"
	"github.com/grassrootseconomics/cic-custodial/pkg/status"
	"github.com/grassrootseconomics/w3-celo-patch/module/eth"
	"github.com/hibiken/asynq"
	"github.com/nats-io/nats.go"
)

type (
	TxPayload struct {
		OtxId uint               `json:"otxId"`
		Tx    *types.Transaction `json:"tx"`
	}

	dispatchEventPayload struct {
		OtxId          uint
		TxHash         string
		DispatchStatus status.Status
	}
)

func TxDispatch(
	celoProvider *celo.Provider,
	pg store.Store,
	js nats.JetStreamContext,

) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var (
			p      TxPayload
			txHash common.Hash
		)

		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return err
		}

		dispatchStatus := store.DispatchStatus{
			OtxId: p.OtxId,
		}

		eventPayload := &dispatchEventPayload{
			OtxId: p.OtxId,
		}

		if err := celoProvider.Client.CallCtx(
			ctx,
			eth.SendTx(p.Tx).Returns(&txHash),
		); err != nil {
			switch err.Error() {
			case celo.ErrGasPriceLow:
				dispatchStatus.Status = status.FailGasPrice
			case celo.ErrInsufficientGas:
				dispatchStatus.Status = status.FailInsufficientGas
			case celo.ErrNonceLow:
				dispatchStatus.Status = status.FailNonce
			default:
				dispatchStatus.Status = status.Unknown
			}

			_, err := pg.CreateDispatchStatus(ctx, dispatchStatus)
			if err != nil {
				return err
			}

			eventJson, err := json.Marshal(eventPayload)
			if err != nil {
				return err
			}

			_, err = js.Publish("CUSTODIAL.dispatchFail", eventJson, nats.MsgId(txHash.Hex()))
			if err != nil {
				return err
			}

			return fmt.Errorf("dispatch: failed %v: %w", err, asynq.SkipRetry)
		}

		dispatchStatus.Status = status.Successful
		_, err := pg.CreateDispatchStatus(ctx, dispatchStatus)
		if err != nil {
			return err
		}

		eventPayload.TxHash = txHash.Hex()

		eventJson, err := json.Marshal(eventPayload)
		if err != nil {
			return err
		}

		_, err = js.Publish("CUSTODIAL.dispatchSuccess", eventJson, nats.MsgId(txHash.Hex()))
		if err != nil {
			return err
		}

		return nil
	}
}
