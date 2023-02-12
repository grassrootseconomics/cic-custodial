package api

const (
	INTERNAL_ERROR   = "ERR_INTERNAL"
	VALIDATION_ERROR = "ERR_VALIDATE"
	DUPLICATE_ERROR  = "ERR_DUPLICATE"
)

type H map[string]any

type OkResp struct {
	Ok     bool `json:"ok"`
	Result H    `json:"result"`
}

type ErrResp struct {
	Ok      bool   `json:"ok"`
	Code    string `json:"errorCode"`
	Message string `json:"message"`
}
