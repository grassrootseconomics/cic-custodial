package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/grassrootseconomics/cic-custodial/internal/events"
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

		if err := cu.Noncestore.SetNewAccountNonce(ctx, payload.PublicKey); err != nil {
			return err
		}

		_, err := cu.TaskerClient.CreateTask(
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
			tasker.AccountGiftVoucherTask,
			tasker.DefaultPriority,
			&tasker.Task{
				Payload: t.Payload(),
			},
		)
		if err != nil {
			return err
		}

		eventPayload := events.EventPayload{
			TrackingId: payload.TrackingId,
		}

		if err := cu.EventEmitter.Publish(
			events.AccountNewNonce,
			payload.PublicKey,
			eventPayload,
		); err != nil {
			return err
		}

		return nil
	}
}
