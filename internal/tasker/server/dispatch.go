package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grassrootseconomics/cic-custodial/internal/tasker/client"
	"github.com/hibiken/asynq"
)

func (tp *TaskerProcessor) txDispatcher(ctx context.Context, t *asynq.Task) error {
	var (
		p client.TxPayload
	)

	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	_, err := tp.ActionsProvider.DispatchSignedTx(ctx, p.Tx)
	if err != nil {
		return err
	}

	return nil
}
