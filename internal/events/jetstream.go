package events

import (
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/zerodha/logf"
)

const (
	StreamName     string = "CUSTODIAL"
	StreamSubjects string = "CUSTODIAL.*"
	// Subjects
	AccountNewNonce    string = "CUSTODIAL.accountNewNonce"
	AccountRegister    string = "CUSTODIAL.accountRegister"
	AccountGiftGas     string = "CUSTODIAL.systemNewAccountGas"
	AccountGiftVoucher string = "CUSTODIAL.systemNewAccountVoucher"
	AccountRefillGas   string = "CUSTODIAL.systemRefillAccountGas"
	DispatchFail       string = "CUSTODIAL.dispatchFail"
	DispatchSuccess    string = "CUSTODIAL.dispatchSuccess"
	SignTransfer       string = "CUSTODIAL.signTransfer"
)

type JetStreamOpts struct {
	Logg            logf.Logger
	ServerUrl       string
	PersistDuration time.Duration
	DedupDuration   time.Duration
}

type JetStream struct {
	logg     logf.Logger
	jsCtx    nats.JetStreamContext
	natsConn *nats.Conn
}

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

	return &JetStream{
		logg:     o.Logg,
		jsCtx:    js,
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
