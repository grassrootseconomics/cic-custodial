package keystore

import (
	"context"
	"crypto/ecdsa"

	eth_crypto "github.com/celo-org/celo-blockchain/crypto"
	"github.com/grassrootseconomics/cic-custodial/internal/queries"
	"github.com/grassrootseconomics/cic-custodial/pkg/keypair"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	Opts struct {
		PostgresPool *pgxpool.Pool
		Queries      *queries.Queries
	}

	PostgresKeystore struct {
		db      *pgxpool.Pool
		queries *queries.Queries
	}
)

func NewPostgresKeytore(o Opts) Keystore {
	return &PostgresKeystore{
		db:      o.PostgresPool,
		queries: o.Queries,
	}
}

// WriteKeyPair inserts a keypair into the db and returns the linked id.
func (ks *PostgresKeystore) WriteKeyPair(ctx context.Context, keypair keypair.Key) (uint, error) {
	var (
		id uint
	)

	if err := ks.db.QueryRow(ctx, ks.queries.WriteKeyPair, keypair.Public, keypair.Private).Scan(&id); err != nil {
		return id, err
	}

	return id, nil
}

// LoadPrivateKey loads a private key as a crypto primitive for direct use. An id is used to search for the private key.
func (ks *PostgresKeystore) LoadPrivateKey(ctx context.Context, publicKey string) (*ecdsa.PrivateKey, error) {
	var (
		privateKeyString string
	)

	if err := ks.db.QueryRow(ctx, ks.queries.LoadKeyPair, publicKey).Scan(&privateKeyString); err != nil {
		return nil, err
	}

	privateKey, err := eth_crypto.HexToECDSA(privateKeyString)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}
