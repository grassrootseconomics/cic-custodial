package keystore

import (
	"context"
	"crypto/ecdsa"

	"github.com/grassrootseconomics/cic-custodial/pkg/keypair"
)

type Keystore interface {
	WriteKeyPair(context.Context, keypair.Key) error
	LoadPrivateKey(context.Context, string) (*ecdsa.PrivateKey, error)
}
