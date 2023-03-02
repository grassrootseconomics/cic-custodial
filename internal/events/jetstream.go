package events

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/grassrootseconomics/cic-custodial/internal/store"
	"github.com/nats-io/nats.go"
	"github.com/zerodha/logf"
)

const (
	// Pub
	StreamName         string = "CUSTODIAL"
	StreamSubjects     string = "CUSTODIAL.*"
	AccountNewNonce    string = "CUSTODIAL.accountNewNonce"
	AccountRegister    string = "CUSTODIAL.accountRegister"
	AccountGiftGas     string = "CUSTODIAL.systemNewAccountGas"
	AccountGiftVoucher string = "CUSTODIAL.systemNewAccountVoucher"
	AccountRefillGas   string = "CUSTODIAL.systemRefillAccountGas"
	DispatchFail       string = "CUSTODIAL.dispatchFail"
	DispatchSuccess    string = "CUSTODIAL.dispatchSuccess"
	SignTransfer       string = "CUSTODIAL.signTransfer"

	// Sub
	durableId     = "cic-custodial"
	pullStream    = "CHAIN"
	pullSubject   = "CHAIN.*"
	actionTimeout = 5 * time.Second
)

type (
	JetStreamOpts struct {
		Logg            logf.Logger
		ServerUrl       string
		PersistDuration time.Duration
		PgStore         store.Store
		DedupDuration   time.Duration
	}

	JetStream struct {
		logg     logf.Logger
		jsCtx    nats.JetStreamContext
		pgStore  store.Store
		natsConn *nats.Conn
	}

	EventPayload struct {
		OtxId      uint   `json:"otxId"`
		TrackingId string `json:"trackingId"`
		TxHash     string `json:"txHash"`
	}
)

func NewJetStreamEventEmitter(o JetStreamOpts) (*JetStream, error) {
	natsConn, err := nats.Connect(o.ServerUrl)
	if err != nil {
		return nil, err
	}

	js, err := natsConn.JetStream()
	if err != nil {
		return nil, err
	}

	// Bootstrap stream if it doesn't exist.
	stream, _ := js.StreamInfo(StreamName)
	if stream == nil {
		_, err = js.AddStream(&nats.StreamConfig{
			Name:       StreamName,
			MaxAge:     o.PersistDuration,
			Storage:    nats.FileStorage,
			Subjects:   []string{StreamSubjects},
			Duplicates: o.DedupDuration,
		})
		if err != nil {
			return nil, err
		}
	}

	// Add a durable consumer
	_, err = js.AddConsumer(pullStream, &nats.ConsumerConfig{
		Durable:       durableId,
		AckPolicy:     nats.AckExplicitPolicy,
		FilterSubject: pullSubject,
	})
	if err != nil {
		return nil, err
	}

	return &JetStream{
		logg:     o.Logg,
		jsCtx:    js,
		pgStore:  o.PgStore,
		natsConn: natsConn,
	}, nil
}

// Close gracefully shutdowns the JetStream connection.
func (js *JetStream) Close() {
	if js.natsConn != nil {
		js.natsConn.Close()
	}
}

// Publish publishes the JSON data to the NATS stream.
func (js *JetStream) Publish(subject string, dedupId string, eventPayload interface{}) error {
	jsonData, err := json.Marshal(eventPayload)
	if err != nil {
		return err
	}

	_, err = js.jsCtx.Publish(subject, jsonData, nats.MsgId(dedupId))
	if err != nil {
		return err
	}

	return nil
}

func (js *JetStream) Subscriber() error {
	subOpts := []nats.SubOpt{
		nats.ManualAck(),
		nats.Bind(pullStream, durableId),
	}

	natsSub, err := js.jsCtx.PullSubscribe(pullSubject, durableId, subOpts...)
	if err != nil {
		return err
	}

	for {
		events, err := natsSub.Fetch(1)
		if err != nil {
			if errors.Is(err, nats.ErrTimeout) {
				continue
			} else if errors.Is(err, nats.ErrConnectionClosed) {
				return nil
			} else {
				return err
			}
		}
		if len(events) > 0 {
			var (
				chainEvent store.MinimalTxInfo

				msg = events[0]
			)

			if err := json.Unmarshal(msg.Data, &chainEvent); err != nil {
				msg.Nak()
				js.logg.Error("jetstream sub: json unmarshal fail", "error", err)
			} else {
				ctx, cancel := context.WithTimeout(context.Background(), actionTimeout)

				if err := js.pgStore.UpdateOtxStatusFromChainEvent(ctx, chainEvent); err != nil {
					msg.Nak()
					js.logg.Error("jetstream sub: otx marker failed to update state", "error", err)
				} else {
					msg.Ack()
				}
				cancel()
			}

		}
	}
}
