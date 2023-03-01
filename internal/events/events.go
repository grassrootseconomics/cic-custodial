package events

type EventPayload struct {
	OtxId      uint   `json:"otxId"`
	TrackingId string `json:"trackingId"`
	TxHash     string `json:"txHash"`
}
