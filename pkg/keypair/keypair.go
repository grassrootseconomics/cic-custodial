package keypair

import (
	"crypto/ecdsa"

	"github.com/celo-org/celo-blockchain/common/hexutil"
	"github.com/celo-org/celo-blockchain/crypto"
)

type Key struct {
	Public  string
	Private string
}

// Generate creates a new keypair from internally randomized entropy.
func Generate() (Key, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return Key{}, err
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)

	publicKey := privateKey.Public()
	publicKeyECDSA := publicKey.(*ecdsa.PublicKey)

	return Key{
		Public:  crypto.PubkeyToAddress(*publicKeyECDSA).Hex(),
		Private: hexutil.Encode(privateKeyBytes)[2:],
	}, nil
}
