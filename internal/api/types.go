package api

const (
	INTERNAL_ERROR   = "ERR_INTERNAL"
	VALIDATION_ERROR = "ERR_VALIDATE"
)

type H map[string]any

type okResp struct {
	Ok     bool `json:"ok"`
	Result H    `json:"result"`
}

type errResp struct {
	Ok   bool   `json:"ok"`
	Code string `json:"errorCode"`
}
