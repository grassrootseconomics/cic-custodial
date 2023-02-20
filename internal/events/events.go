package events

type EventEmitter interface {
	Close()
	Publish(subject string, dedupId string, eventPayload interface{}) error
}

type (
	EventPayload struct {
		OtxId      uint   `json:"otxId"`
		TrackingId string `json:"trackingId"`
		TxHash     string `json:"txHash"`
	}
)
