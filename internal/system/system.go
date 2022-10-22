package system

import (
	"crypto/ecdsa"

	eth_crypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/grassrootseconomics/cic-custodial/internal/noncestore"
	system_noncestore "github.com/grassrootseconomics/cic-custodial/internal/noncestore/providers/system"
	"github.com/grassrootseconomics/cic-go-sdk/chain"
)

type Opts struct {
	SystemPublicKey  string
	SystemPrivateKey string
	ChainProvider    *chain.Provider
}

type SystemProvider struct {
	SystemNoncestore noncestore.SystemNoncestore
	SystemPublicKey  string
	SystemPrivateKey *ecdsa.PrivateKey
}

func NewSystemProvider(o Opts) (*SystemProvider, error) {
	loadedPrivateKey, err := eth_crypto.HexToECDSA(o.SystemPrivateKey)
	if err != nil {
		return nil, err
	}

	systemNoncestore, err := system_noncestore.NewSystemNoncestore(system_noncestore.Opts{
		ChainProvider:  o.ChainProvider,
		AccountAddress: o.SystemPublicKey,
	})
	if err != nil {
		return nil, err
	}

	return &SystemProvider{
		SystemNoncestore: systemNoncestore,
		SystemPublicKey:  o.SystemPublicKey,
		SystemPrivateKey: loadedPrivateKey,
	}, nil
}
