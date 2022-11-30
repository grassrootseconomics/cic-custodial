package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresPoolOpts struct {
	DSN string
	// Debug bool
	// Logg  tracelog.Logger
}

func NewPostgresPool(o PostgresPoolOpts) (*pgxpool.Pool, error) {
	parsedConfig, err := pgxpool.ParseConfig(o.DSN)
	if err != nil {
		return nil, err
	}

	// if o.Debug {
	// 	parsedConfig.ConnConfig.Tracer = &tracelog.TraceLog{
	// 		Logger:   o.Logg,
	// 		LogLevel: tracelog.LogLevelDebug,
	// 	}
	// }

	dbPool, err := pgxpool.NewWithConfig(context.Background(), parsedConfig)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := dbPool.Ping(ctx); err != nil {
		return nil, err
	}

	return dbPool, nil
}
