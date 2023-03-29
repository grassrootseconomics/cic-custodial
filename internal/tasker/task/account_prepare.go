package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/grassrootseconomics/cic-custodial/internal/pub"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/hibiken/asynq"
)

type AccountPayload struct {
	PublicKey  string `json:"publicKey"`
	TrackingId string `json:"trackingId"`
}

func AccountPrepare(cu *custodial.Custodial) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var payload AccountPayload

		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return fmt.Errorf("account: failed %v: %w", err, asynq.SkipRetry)
		}

		if err := cu.Noncestore.SetAccountNonce(ctx, payload.PublicKey, 0); err != nil {
			return err
		}

		_, err := cu.TaskerClient.CreateTask(
			ctx,
			tasker.AccountRegisterTask,
			tasker.DefaultPriority,
			&tasker.Task{
				Payload: t.Payload(),
			},
		)
		if err != nil {
			return err
		}

		_, err = cu.TaskerClient.CreateTask(
			ctx,
			tasker.AccountGiftGasTask,
			tasker.DefaultPriority,
			&tasker.Task{
				Payload: t.Payload(),
			},
		)
		if err != nil {
			return err
		}

		_, err = cu.TaskerClient.CreateTask(
			ctx,
			tasker.AccountGiftVoucherTask,
			tasker.DefaultPriority,
			&tasker.Task{
				Payload: t.Payload(),
			},
		)
		if err != nil {
			return err
		}

		eventPayload := pub.EventPayload{
			TrackingId: payload.TrackingId,
		}

		if err := cu.Pub.Publish(
			pub.AccountNewNonce,
			payload.PublicKey,
			eventPayload,
		); err != nil {
			return err
		}

		return nil
	}
}
