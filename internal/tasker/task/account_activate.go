package task

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/hibiken/asynq"
)

const (
	requiredQuorum = 3
)

var (
	ErrQuorumNotReached = errors.New("Account activation quorum not reached.")
)

func AccountActivateProcessor(cu *custodial.Custodial) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var (
			payload AccountPayload
		)

		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return err
		}

		quorum, err := cu.PgStore.GetAccountActivationQuorum(ctx, payload.TrackingId)
		if err != nil {
			return err
		}

		if quorum < requiredQuorum {
			return ErrQuorumNotReached
		}

		if err := cu.PgStore.ActivateAccount(ctx, payload.PublicKey); err != nil {
			return err
		}

		return nil
	}
}
