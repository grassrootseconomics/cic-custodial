package postgres

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	eth_crypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/grassrootseconomics/cic-custodial/internal/ethereum"
	"github.com/grassrootseconomics/cic-custodial/internal/keystore"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Opts struct {
	PostgresDSN string
}

type PostgresKeystore struct {
	db *pgxpool.Pool
}

func NewPostgresKeytore(o Opts) (keystore.Keystore, error) {
	dbPool, err := pgxpool.New(context.Background(), o.PostgresDSN)
	if err != nil {
		return nil, err
	}

	if err := dbPool.Ping(context.Background()); err != nil {
		return nil, err
	}

	if err := applyMigration(dbPool); err != nil {
		return nil, fmt.Errorf("keystore migration failed %v", err)
	}

	return &PostgresKeystore{
		db: dbPool,
	}, nil
}

func (ks *PostgresKeystore) WriteKeyPair(ctx context.Context, keypair ethereum.Key) error {
	_, err := ks.db.Exec(ctx, "INSERT INTO keystore(public_key, private_key) VALUES($1, $2)", keypair.Public, keypair.Private)
	if err != nil {
		return err
	}

	return nil
}

func (ks *PostgresKeystore) LoadPrivateKey(ctx context.Context, publicKey string) (*ecdsa.PrivateKey, error) {
	var (
		privateKeyString string
	)

	if err := ks.db.QueryRow(ctx, "SELECT private_key FROM keystore WHERE public_key=$1", publicKey).Scan(&privateKeyString); err != nil {
		return nil, err
	}

	privateKey, err := eth_crypto.HexToECDSA(privateKeyString)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func (ks *PostgresKeystore) ActivateAccount(ctx context.Context, publicKey string) error {
	_, err := ks.db.Exec(ctx, "UPDATE keystore SET activated = true WHERE public_key=$1", publicKey)
	if err != nil {
		return err
	}

	return nil
}
