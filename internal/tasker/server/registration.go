package server

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/grassrootseconomics/cic-custodial/internal/tasker/client"
	"github.com/grassrootseconomics/cic-go-sdk/chain"
	"github.com/hibiken/asynq"
	"github.com/lmittmann/w3"
)

const (
	initialGiftGasValue = 1000000
)

func (tp *TaskerProcessor) setNewAccountNonce(ctx context.Context, t *asynq.Task) error {
	var (
		p client.RegistrationPayload
	)

	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	if err := tp.Noncestore.SetNewAccountNonce(ctx, p.PublicKey); err != nil {
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

	lock, err := tp.LockProvider.Obtain(ctx, tp.SystemPublicKey, LockTTL, nil)
	if err != nil {
		return err
	}
	defer lock.Release(ctx)

	nonce, err := tp.Noncestore.Acquire(ctx, tp.SystemPublicKey)
	if err != nil {
		return err
	}

	builtTx, err := tp.ChainProvider.BuildGasTransferTx(tp.SystemPrivateKey, chain.TransactionData{
		To:    w3.A(p.PublicKey),
		Nonce: nonce,
	}, big.NewInt(initialGiftGasValue))
	if err != nil {
		if err := tp.Noncestore.Return(ctx, p.PublicKey); err != nil {
			return err
		}
		return err
	}

	_, err = tp.TaskerClient.CreateTxDispatchTask(client.TxPayload{
		Tx: builtTx,
	}, client.TxDispatchTask)
	if err != nil {
		if err := tp.Noncestore.Return(ctx, p.PublicKey); err != nil {
			return err
		}
		return err
	}

	_, err = tp.TaskerClient.CreateRegistrationTask(p, client.ActivateAccountTask)
	if err != nil {
		if err := tp.Noncestore.Return(ctx, p.PublicKey); err != nil {
			return err
		}
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

	if err := tp.Keystore.ActivateAccount(ctx, p.PublicKey); err != nil {
		return err
	}

	return nil
}
