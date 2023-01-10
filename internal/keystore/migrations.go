package keystore

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func applyMigration(dbPool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := dbPool.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS keystore (
		id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		public_key TEXT NOT NULL,
		private_key TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`)
	if err != nil {
		return err
	}

	return nil
}
