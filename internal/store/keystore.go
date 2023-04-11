package store

import (
	"context"
	"crypto/ecdsa"

	eth_crypto "github.com/celo-org/celo-blockchain/crypto"
	"github.com/grassrootseconomics/cic-custodial/pkg/keypair"
)

func (s *PgStore) WriteKeyPair(
	ctx context.Context,
	keypair keypair.Key,
) (uint, error) {
	var (
		id uint
	)

	if err := s.db.QueryRow(
		ctx,
		s.queries.WriteKeyPair,
		keypair.Public,
		keypair.Private,
	).Scan(&id); err != nil {
		return id, err
	}

	return id, nil
}

func (s *PgStore) LoadPrivateKey(
	ctx context.Context,
	publicKey string,
) (*ecdsa.PrivateKey, error) {
	var (
		privateKeyString string
	)

	if err := s.db.QueryRow(
		ctx,
		s.queries.LoadKeyPair,
		publicKey,
	).Scan(&privateKeyString); err != nil {
		return nil, err
	}

	privateKey, err := eth_crypto.HexToECDSA(privateKeyString)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}
