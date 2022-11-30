package logg

import (
	"context"

	"github.com/zerodha/logf"
)

type RedisLogg struct {
	logg *logf.Logger
}

func NewRedisTraceLogg(lo logf.Logger) RedisLogg {
	return RedisLogg{
		logg: &lo,
	}
}

func (l RedisLogg) Printf(ctx context.Context, format string, v ...interface{}) {
	l.logg.Debug("redis", "debug", "format", format, "data", v)
}
