package sub

import (
	"context"
	"errors"
	"time"

	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/nats-io/nats.go"
	"github.com/zerodha/logf"
)

const (
	durableId     = "cic-custodial"
	pullStream    = "CHAIN"
	pullSubject   = "CHAIN.*"
	actionTimeout = 5 * time.Second
	waitDelay     = 1 * time.Second
)

type (
	SubOpts struct {
		CustodialContainer *custodial.Custodial
		JsCtx              nats.JetStreamContext
		Logg               logf.Logger
		NatsConn           *nats.Conn
	}

	Sub struct {
		cu       *custodial.Custodial
		jsCtx    nats.JetStreamContext
		logg     logf.Logger
		natsConn *nats.Conn
	}
)

func NewSub(o SubOpts) (*Sub, error) {
	_, err := o.JsCtx.AddConsumer(pullStream, &nats.ConsumerConfig{
		Durable:       durableId,
		AckPolicy:     nats.AckExplicitPolicy,
		FilterSubject: pullSubject,
	})
	if err != nil {
		return nil, err
	}

	return &Sub{
		cu:       o.CustodialContainer,
		jsCtx:    o.JsCtx,
		logg:     o.Logg,
		natsConn: o.NatsConn,
	}, nil
}

func (s *Sub) Process() error {
	subOpts := []nats.SubOpt{
		nats.ManualAck(),
		nats.Bind(pullStream, durableId),
	}

	natsSub, err := s.jsCtx.PullSubscribe(pullSubject, durableId, subOpts...)
	if err != nil {
		return err
	}

	for {
		events, err := natsSub.Fetch(1)
		if err != nil {
			if errors.Is(err, nats.ErrTimeout) {
				s.logg.Debug("sub: no msg to pull")
				continue
			} else if errors.Is(err, nats.ErrConnectionClosed) {
				return nil
			} else {
				return err
			}
		}

		if len(events) > 0 {
			msg := events[0]
			ctx, cancel := context.WithTimeout(context.Background(), actionTimeout)

			if err := s.handler(ctx, msg); err != nil {
				s.logg.Error("sub: handler error", "error", err)
				msg.Nak()
			}

			s.logg.Debug("sub: processed msg", "subject", msg.Subject)
			msg.Ack()
			cancel()
		}
	}
}

func (s *Sub) Close() {
	if s.natsConn != nil {
		s.natsConn.Close()
	}
}
