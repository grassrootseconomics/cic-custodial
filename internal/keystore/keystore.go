package keystore

import (
	"context"
	"crypto/ecdsa"

	"github.com/grassrootseconomics/cic-custodial/internal/ethereum"
)

// Keystore represents a persistent distributed keystore
type Keystore interface {
	WriteKeyPair(context.Context, ethereum.Key) error
	LoadPrivateKey(context.Context, string) (*ecdsa.PrivateKey, error)
	ActivateAccount(context.Context, string) error
}
