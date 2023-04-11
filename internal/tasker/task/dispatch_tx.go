package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/grassrootseconomics/celoutils"
	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
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
			dispathchTx    common.Hash
			dispatchStatus enum.OtxStatus = enum.IN_NETWORK
		)

		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return err
		}

		if err := cu.CeloProvider.Client.CallCtx(
			ctx,
			eth.SendTx(payload.Tx).Returns(&dispathchTx),
		); err != nil {
			switch err.Error() {
			case celoutils.ErrGasPriceLow:
				dispatchStatus = enum.FAIL_LOW_GAS_PRICE
			case celoutils.ErrInsufficientGas:
				dispatchStatus = enum.FAIL_NO_GAS
			case celoutils.ErrNonceLow:
				dispatchStatus = enum.FAIL_LOW_NONCE
			default:
				dispatchStatus = enum.FAIL_UNKNOWN_RPC_ERROR
			}

			if err := cu.Store.CreateDispatchStatus(ctx, payload.OtxId, dispatchStatus); err != nil {
				return fmt.Errorf("dispatch: failed %v: %w", err, asynq.SkipRetry)
			}

			return fmt.Errorf("dispatch: failed %v: %w", err, asynq.SkipRetry)
		}

		if err := cu.Store.CreateDispatchStatus(ctx, payload.OtxId, dispatchStatus); err != nil {
			return fmt.Errorf("dispatch: failed %v: %w", err, asynq.SkipRetry)
		}

		return nil
	}
}
