package keystore

import (
	"context"
	"crypto/ecdsa"

	"github.com/grassrootseconomics/cic-custodial/pkg/keypair"
)

// Keystore defines how keypairs should be stored and accessed from a storage backend.
type Keystore interface {
	WriteKeyPair(context.Context, keypair.Key) (uint, error)
	LoadPrivateKey(context.Context, string) (*ecdsa.PrivateKey, error)
}
