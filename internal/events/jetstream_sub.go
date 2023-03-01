package events

import (
	"context"
	"errors"
	"time"

	"github.com/grassrootseconomics/cic-custodial/internal/store"
	"github.com/nats-io/nats.go"
)

const (
	backOffTimer = 2 * time.Second
	durableId    = "cic-custodial"
	pullStream   = "CHAIN"
	pullSubject  = "CHAIN.*"
)

func (js *JetStream) ChainSubscription(ctx context.Context, store store.Store) error {
	subOpts := []nats.SubOpt{
		nats.ManualAck(),
		nats.Bind(pullStream, durableId),
	}

	natsSub, err := js.jsCtx.PullSubscribe(pullSubject, durableId, subOpts...)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			js.logg.Info("jetstream chain sub: shutdown signal received")
			return nil
		default:
			events, err := natsSub.Fetch(1)
			if err != nil {
				if errors.Is(err, nats.ErrTimeout) {
					// Supressed retry
					js.logg.Error("jetstream chain sub: fetch NATS timeout", "error", err)
					time.Sleep(backOffTimer)
					continue
				} else {
					js.logg.Error("jetstream chain sub: fetch other error", "error", err)
				}
			}
			if len(events) > 0 {
				// TODO: Unmarshal
				// TODO: UpdateOtxStatus
			}
		}

	}
}
