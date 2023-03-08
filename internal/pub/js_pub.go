package pub

import (
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	streamName         string = "CUSTODIAL"
	streamSubjects     string = "CUSTODIAL.*"
	AccountNewNonce    string = "CUSTODIAL.accountNewNonce"
	AccountRegister    string = "CUSTODIAL.accountRegister"
	AccountGiftGas     string = "CUSTODIAL.systemNewAccountGas"
	AccountGiftVoucher string = "CUSTODIAL.systemNewAccountVoucher"
	AccountRefillGas   string = "CUSTODIAL.systemRefillAccountGas"
	DispatchFail       string = "CUSTODIAL.dispatchFail"
	DispatchSuccess    string = "CUSTODIAL.dispatchSuccess"
	SignTransfer       string = "CUSTODIAL.signTransfer"
)

type (
	PubOpts struct {
		DedupDuration   time.Duration
		JsCtx           nats.JetStreamContext
		PersistDuration time.Duration
	}

	Pub struct {
		jsCtx nats.JetStreamContext
	}

	EventPayload struct {
		OtxId      uint   `json:"otxId"`
		TrackingId string `json:"trackingId"`
		TxHash     string `json:"txHash"`
	}
)

func NewPub(o PubOpts) (*Pub, error) {
	stream, _ := o.JsCtx.StreamInfo(streamName)
	if stream == nil {
		_, err := o.JsCtx.AddStream(&nats.StreamConfig{
			Name:       streamName,
			MaxAge:     o.PersistDuration,
			Storage:    nats.FileStorage,
			Subjects:   []string{streamSubjects},
			Duplicates: o.DedupDuration,
		})
		if err != nil {
			return nil, err
		}
	}

	return &Pub{
		jsCtx: o.JsCtx,
	}, nil
}

func (p *Pub) Publish(subject string, dedupId string, eventPayload interface{}) error {
	jsonData, err := json.Marshal(eventPayload)
	if err != nil {
		return err
	}

	_, err = p.jsCtx.Publish(subject, jsonData, nats.MsgId(dedupId))
	if err != nil {
		return err
	}

	return nil
}
