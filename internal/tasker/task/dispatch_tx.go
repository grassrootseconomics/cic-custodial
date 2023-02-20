package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/grassrootseconomics/celoutils"
	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/grassrootseconomics/cic-custodial/internal/events"
	"github.com/grassrootseconomics/cic-custodial/internal/store"
	"github.com/grassrootseconomics/cic-custodial/pkg/status"
	"github.com/grassrootseconomics/w3-celo-patch/module/eth"
	"github.com/hibiken/asynq"
)

type TxPayload struct {
	OtxId uint               `json:"otxId"`
	Tx    *types.Transaction `json:"tx"`
}

func DispatchTx(cu *custodial.Custodial) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var (
			payload        TxPayload
			dispatchStatus store.DispatchStatus
			eventPayload   events.EventPayload
			dispathchTx    common.Hash
		)

		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return fmt.Errorf("dispatch: failed %v: %w", err, asynq.SkipRetry)
		}

		txHash := payload.Tx.Hash().Hex()

		dispatchStatus.OtxId, eventPayload.OtxId = payload.OtxId, payload.OtxId
		eventPayload.TxHash = txHash

		if err := cu.CeloProvider.Client.CallCtx(
			ctx,
			eth.SendTx(payload.Tx).Returns(&dispathchTx),
		); err != nil {
			dispatchStatus.Status = status.Unknown

			switch err.Error() {
			case celoutils.ErrGasPriceLow:
				dispatchStatus.Status = status.FailGasPrice
			case celoutils.ErrInsufficientGas:
				dispatchStatus.Status = status.FailInsufficientGas
			case celoutils.ErrNonceLow:
				dispatchStatus.Status = status.FailNonce
			}

			if err := cu.PgStore.CreateDispatchStatus(ctx, dispatchStatus); err != nil {
				return fmt.Errorf("dispatch: failed %v: %w", err, asynq.SkipRetry)
			}

			if err := cu.EventEmitter.Publish(events.DispatchFail, txHash, eventPayload); err != nil {
				return fmt.Errorf("dispatch: failed %v: %w", err, asynq.SkipRetry)
			}

			return fmt.Errorf("dispatch: failed %v: %w", err, asynq.SkipRetry)
		}

		dispatchStatus.Status = status.Successful

		if err := cu.PgStore.CreateDispatchStatus(ctx, dispatchStatus); err != nil {
			return fmt.Errorf("dispatch: failed %v: %w", err, asynq.SkipRetry)
		}

		if err := cu.EventEmitter.Publish(events.DispatchSuccess, txHash, eventPayload); err != nil {
			return fmt.Errorf("dispatch: failed %v: %w", err, asynq.SkipRetry)
		}

		return nil
	}
}
