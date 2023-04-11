package util

import "time"

const (
	// SLATimeout is the max duration after which any network call/context should be aborted.
	SLATimeout = 5 * time.Second
)
