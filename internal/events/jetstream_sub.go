package events

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/grassrootseconomics/cic-custodial/internal/store"
	"github.com/nats-io/nats.go"
)

const (
	durableId   = "cic-custodial"
	pullStream  = "CHAIN"
	pullSubject = "CHAIN.*"
)

func (js *JetStream) ChainSubscription(ctx context.Context, pgStore store.Store) error {
	_, err := js.jsCtx.AddConsumer(pullStream, &nats.ConsumerConfig{
		Durable:       durableId,
		AckPolicy:     nats.AckExplicitPolicy,
		FilterSubject: pullSubject,
	})
	if err != nil {
		return err
	}

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
			js.Close()
			return nil
		default:
			events, err := natsSub.Fetch(1)
			if err != nil {
				if errors.Is(err, nats.ErrTimeout) {
					continue
				} else {
					js.logg.Error("jetstream chain sub: fetch other error", "error", err)
				}
			}
			if len(events) == 0 {
				continue
			}
			var (
				chainEvent store.MinimalTxInfo
			)

			if err := json.Unmarshal(events[0].Data, &chainEvent); err != nil {
				js.logg.Error("jetstream chain sub: json unmarshal fail", "error", err)
			}

			if err := pgStore.UpdateOtxStatusFromChainEvent(context.Background(), chainEvent); err != nil {
				events[0].Nak()
				js.logg.Error("jetstream chain sub: otx marker failed to update state", "error", err)
			}
			events[0].Ack()
			js.logg.Debug("jetstream chain sub: successfully updated status", "tx", chainEvent.TxHash)
		}

	}
}
