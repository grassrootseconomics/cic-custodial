package store

import (
	"github.com/grassrootseconomics/cic-custodial/internal/queries"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	Opts struct {
		PostgresPool *pgxpool.Pool
		Queries      *queries.Queries
	}

	PostgresStore struct {
		db      *pgxpool.Pool
		queries *queries.Queries
	}
)

func NewPostgresStore(o Opts) Store {
	return &PostgresStore{
		db:      o.PostgresPool,
		queries: o.Queries,
	}
}
