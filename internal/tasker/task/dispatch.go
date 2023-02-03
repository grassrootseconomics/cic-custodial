package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	celo "github.com/grassrootseconomics/cic-celo-sdk"
	"github.com/grassrootseconomics/w3-celo-patch/module/eth"
	"github.com/hibiken/asynq"
)

type TxPayload struct {
	Tx *types.Transaction `json:"tx"`
}

func TxDispatch(
	celoProvider *celo.Provider,
) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var (
			p      TxPayload
			txHash common.Hash
		)

		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
		}

		// TODO: Handle all fail cases
		if err := celoProvider.Client.CallCtx(
			ctx,
			eth.SendTx(p.Tx).Returns(&txHash),
		); err != nil {
			return err
		}

		return nil
	}
}
