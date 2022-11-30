package logg

import (
	"context"

	"github.com/jackc/pgx/v5/tracelog"
	"github.com/zerodha/logf"
)

type PgxLogg struct {
	logg *logf.Logger
}

func NewPgxTraceLogg(lo logf.Logger) PgxLogg {
	return PgxLogg{
		logg: &lo,
	}
}

func (l PgxLogg) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	l.logg.Debug("pgx", "level", level, "msg", msg, "data", data)
}
