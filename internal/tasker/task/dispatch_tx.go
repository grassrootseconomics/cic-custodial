package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/grassrootseconomics/celoutils"
	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/grassrootseconomics/cic-custodial/internal/store"
	"github.com/grassrootseconomics/cic-custodial/pkg/enum"
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
			dispathchTx    common.Hash
		)

		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return err
		}

		dispatchStatus.OtxId = payload.OtxId

		if err := cu.CeloProvider.Client.CallCtx(
			ctx,
			eth.SendTx(payload.Tx).Returns(&dispathchTx),
		); err != nil {
			switch err.Error() {
			case celoutils.ErrGasPriceLow:
				dispatchStatus.Status = enum.FAIL_LOW_GAS_PRICE
			case celoutils.ErrInsufficientGas:
				dispatchStatus.Status = enum.FAIL_NO_GAS
			case celoutils.ErrNonceLow:
				dispatchStatus.Status = enum.FAIL_LOW_NONCE
			default:
				dispatchStatus.Status = enum.FAIL_UNKNOWN_RPC_ERROR
			}

			if err := cu.PgStore.CreateDispatchStatus(ctx, dispatchStatus); err != nil {
				return fmt.Errorf("dispatch: failed %v: %w", err, asynq.SkipRetry)
			}

			return fmt.Errorf("dispatch: failed %v: %w", err, asynq.SkipRetry)
		}

		dispatchStatus.Status = enum.IN_NETWORK

		if err := cu.PgStore.CreateDispatchStatus(ctx, dispatchStatus); err != nil {
			return fmt.Errorf("dispatch: failed %v: %w", err, asynq.SkipRetry)
		}

		return nil
	}
}
