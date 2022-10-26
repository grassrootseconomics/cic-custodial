package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker/client"
	"github.com/hibiken/asynq"
	"github.com/lmittmann/w3/module/eth"
)

func (tp *TaskerProcessor) txDispatcher(ctx context.Context, t *asynq.Task) error {
	var (
		p      client.TxPayload
		txHash common.Hash
	)

	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	if err := tp.ChainProvider.EthClient.CallCtx(
		ctx,
		eth.SendTx(p.Tx).Returns(&txHash),
	); err != nil {
		return err
	}

	return nil
}
