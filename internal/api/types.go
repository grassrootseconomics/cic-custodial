package api

const (
	INTERNAL_ERROR     = "ERR_INTERNAL"
	KEYPAIR_ERROR      = "ERR_GEN_KEYPAIR"
	JSON_MARSHAL_ERROR = "ERR_PAYLOAD_SERIALIZATION"
	TASK_CHAIN_ERROR   = "ERR_START_TASK_CHAIN"
	VALIDATION_ERROR   = "ERR_VALIDATE"
	BIND_ERROR         = "ERR_BIND"
)

type okResp struct {
	Ok   bool        `json:"ok"`
	Data interface{} `json:"data"`
}

type errResp struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}
