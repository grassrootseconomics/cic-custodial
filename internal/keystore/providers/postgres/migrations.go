package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func applyMigration(dbPool *pgxpool.Pool) error {
	_, err := dbPool.Exec(context.Background(), `
	CREATE TABLE IF NOT EXISTS keystore (
		id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		public_key TEXT NOT NULL,
		private_key TEXT NOT NULL,
		activated BOOLEAN DEFAULT false,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`)
	if err != nil {
		return err
	}

	return nil
}
