package keystore

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	eth_crypto "github.com/celo-org/celo-blockchain/crypto"
	"github.com/grassrootseconomics/cic-custodial/pkg/keypair"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zerodha/logf"
)

type Opts struct {
	PostgresPool *pgxpool.Pool
	Logg         logf.Logger
}

type PostgresKeystore struct {
	db *pgxpool.Pool
}

func NewPostgresKeytore(o Opts) (Keystore, error) {
	if err := applyMigration(o.PostgresPool); err != nil {
		return nil, fmt.Errorf("keystore migration failed %v", err)
	}
	o.Logg.Info("Successfully ran keystore migrations")

	return &PostgresKeystore{
		db: o.PostgresPool,
	}, nil
}

func (ks *PostgresKeystore) WriteKeyPair(ctx context.Context, keypair keypair.Key) error {
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
