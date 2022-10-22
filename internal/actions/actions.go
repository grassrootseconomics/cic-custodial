package actions

import (
	"context"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/grassrootseconomics/cic-custodial/internal/ethereum"
	"github.com/grassrootseconomics/cic-custodial/internal/keystore"
	"github.com/grassrootseconomics/cic-custodial/internal/noncestore"
	system_provider "github.com/grassrootseconomics/cic-custodial/internal/system"
	"github.com/grassrootseconomics/cic-go-sdk/chain"
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
	SystemProvider *system_provider.SystemProvider
	ChainProvider  *chain.Provider
	Keystore       keystore.Keystore
	Noncestore     noncestore.Noncestore
}

type ActionsProvider struct {
	SystemProvider *system_provider.SystemProvider
	ChainProvider  *chain.Provider
	Keystore       keystore.Keystore
	Noncestore     noncestore.Noncestore
}

func NewActionsProvider(o Opts) *ActionsProvider {
	var _ Actions = (*ActionsProvider)(nil)

	return &ActionsProvider{
		SystemProvider: o.SystemProvider,
		ChainProvider:  o.ChainProvider,
		Keystore:       o.Keystore,
		Noncestore:     o.Noncestore,
	}
}
