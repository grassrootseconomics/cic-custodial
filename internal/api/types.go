package api

import "errors"

var (
	ErrInvalidJSON = errors.New("Invalid JSON structure.")
)

type H map[string]any

type OkResp struct {
	Ok     bool `json:"ok"`
	Result H    `json:"result"`
}

type ErrResp struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}
