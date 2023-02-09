package task

import (
	"context"
	"encoding/json"

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
		OtxId      uint               `json:"otxId"`
		TrackingId string             `json:"trackingId"`
		Tx         *types.Transaction `json:"tx"`
	}

	dispatchEventPayload struct {
		TrackingId string
		TxHash     string
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
			OtxId:      p.OtxId,
			TrackingId: p.TrackingId,
		}

		eventPayload := &dispatchEventPayload{
			TrackingId: p.TrackingId,
		}

		if err := celoProvider.Client.CallCtx(
			ctx,
			eth.SendTx(p.Tx).Returns(&txHash),
		); err != nil {
			// TODO: Coreect error status
			dispatchStatus.Status = status.FailGasPrice

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

			return err
		}

		dispatchStatus.TrackingId = status.Successful
		eventPayload.TxHash = txHash.Hex()

		eventJson, err := json.Marshal(eventPayload)
		if err != nil {
			return err
		}

		_, err = js.Publish("CUSTODIAL.dispatchSuccessful", eventJson, nats.MsgId(txHash.Hex()))
		if err != nil {
			return err
		}

		return nil
	}
}
