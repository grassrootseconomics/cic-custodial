package api

type H map[string]any

type OkResp struct {
	Ok     bool `json:"ok"`
	Result H    `json:"result"`
}

type ErrResp struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}
