package actions

import (
	"context"
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/core/types"
	eth_crypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/grassrootseconomics/cic-custodial/internal/ethereum"
	"github.com/grassrootseconomics/cic-custodial/internal/keystore"
	"github.com/grassrootseconomics/cic-custodial/internal/noncestore"
	"github.com/grassrootseconomics/cic-go-sdk/chain"
	"github.com/zerodha/logf"
)

type Actions interface {
	CreateNewKeyPair(context.Context) (ethereum.Key, error)
	ActivateCustodialAccount(context.Context, string) error
	SetNewAccountNonce(context.Context, string) error

	SignGiftGasTx(context.Context, string) (*types.Transaction, error)
	SignTopUpGasTx(context.Context, string) (*types.Transaction, error)
	SignGiftVouchertx(context.Context, string) (*types.Transaction, error)

	DispatchSignedTx(context.Context, *types.Transaction) (string, error)
}

type Opts struct {
	SystemPublicKey  string
	SystemPrivateKey string
	ChainProvider    *chain.Provider
	Keystore         keystore.Keystore
	Noncestore       noncestore.Noncestore
	Logger           logf.Logger
}

type ActionsProvider struct {
	SystemPublicKey  string
	SystemPrivateKey *ecdsa.PrivateKey
	ChainProvider    *chain.Provider
	Keystore         keystore.Keystore
	Noncestore       noncestore.Noncestore
	Lo               logf.Logger
}

func NewActionsProvider(o Opts) (*ActionsProvider, error) {
	var _ Actions = (*ActionsProvider)(nil)

	loadedPrivateKey, err := eth_crypto.HexToECDSA(o.SystemPrivateKey)
	if err != nil {
		return nil, err
	}

	_, err = o.Noncestore.Peek(context.Background(), o.SystemPublicKey)
	if err != nil {
		nonce, err := o.Noncestore.SyncNetworkNonce(context.Background(), o.SystemPublicKey)
		o.Logger.Debug("actionsProvider: syncing system nonce", "nonce", nonce)
		if err != nil {
			return nil, err
		}
	}

	return &ActionsProvider{
		SystemPublicKey:  o.SystemPublicKey,
		SystemPrivateKey: loadedPrivateKey,
		ChainProvider:    o.ChainProvider,
		Keystore:         o.Keystore,
		Noncestore:       o.Noncestore,
		Lo:               o.Logger,
	}, nil
}
