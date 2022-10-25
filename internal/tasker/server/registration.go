package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grassrootseconomics/cic-custodial/internal/tasker/client"
	"github.com/hibiken/asynq"
)

func (tp *TaskerProcessor) setNewAccountNonce(ctx context.Context, t *asynq.Task) error {
	var (
		p client.RegistrationPayload
	)

	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	if err := tp.ActionsProvider.SetNewAccountNonce(ctx, p.PublicKey); err != nil {
		return err
	}

	_, err := tp.TaskerClient.CreateRegistrationTask(p, client.GiftGasTask)
	if err != nil {
		return err
	}

	return nil
}

func (tp *TaskerProcessor) giftGasProcessor(ctx context.Context, t *asynq.Task) error {
	var (
		p client.RegistrationPayload
	)

	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	lock, err := tp.LockProvider.Obtain(ctx, tp.ActionsProvider.SystemPublicKey, LockTTL, nil)
	if err != nil {
		return err
	}
	defer lock.Release(ctx)

	signedTx, err := tp.ActionsProvider.SignGiftGasTx(ctx, p.PublicKey)
	if err != nil {
		return err
	}

	_, err = tp.TaskerClient.CreateTxDispatchTask(client.TxPayload{
		Tx: signedTx,
	}, client.TxDispatchTask)
	if err != nil {
		return err
	}

	_, err = tp.TaskerClient.CreateRegistrationTask(p, client.ActivateAccountTask)
	if err != nil {
		return err
	}

	return nil
}

func (tp *TaskerProcessor) activateAccountProcessor(ctx context.Context, t *asynq.Task) error {
	var (
		p client.RegistrationPayload
	)

	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	if err := tp.ActionsProvider.ActivateCustodialAccount(ctx, p.PublicKey); err != nil {
		return err
	}

	return nil
}
