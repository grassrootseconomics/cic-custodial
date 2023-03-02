package postgres

import (
	"context"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
)

const (
	schemaTable = "schema_version"
)

type PostgresPoolOpts struct {
	DSN                  string
	MigrationsFolderPath string
}

// NewPostgresPool creates a reusbale connection pool across the cic-custodial component.
func NewPostgresPool(ctx context.Context, o PostgresPoolOpts) (*pgxpool.Pool, error) {
	parsedConfig, err := pgxpool.ParseConfig(o.DSN)
	if err != nil {
		return nil, err
	}

	dbPool, err := pgxpool.NewWithConfig(ctx, parsedConfig)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	migrator, err := migrate.NewMigrator(ctx, conn.Conn(), schemaTable)
	if err != nil {
		return nil, err
	}

	if err := migrator.LoadMigrations(os.DirFS(o.MigrationsFolderPath)); err != nil {
		return nil, err
	}

	if err := migrator.Migrate(ctx); err != nil {
		return nil, err
	}

	return dbPool, nil
}
