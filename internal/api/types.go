package api

const (
	INTERNAL      = "ERR_INTERNAL"
	KEYPAIR_ERROR = "ERR_GEN_KEYPAIR"
	JSON_MARSHAL  = "ERR_PAYLOAD_SERIALIZATION"
	TASK_CHAIN    = "ERR_START_TASK_CHAIN"
)

type okResp struct {
	Ok   bool        `json:"ok"`
	Data interface{} `json:"data"`
}

type errResp struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}
