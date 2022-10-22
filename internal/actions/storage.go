package actions

import (
	"context"

	"github.com/grassrootseconomics/cic-custodial/internal/ethereum"
)

func (ap *ActionsProvider) CreateNewKeyPair(ctx context.Context) (ethereum.Key, error) {
	generatedKeyPair, err := ethereum.GenerateKeyPair()
	if err != nil {
		return ethereum.Key{}, err
	}

	if err := ap.Keystore.WriteKeyPair(ctx, generatedKeyPair); err != nil {
		return ethereum.Key{}, err
	}

	return generatedKeyPair, nil
}

func (ap *ActionsProvider) ActivateCustodialAccount(ctx context.Context, publicKey string) error {
	if err := ap.Keystore.ActivateAccount(ctx, publicKey); err != nil {
		return err
	}

	return nil
}

func (ap *ActionsProvider) SetNewAccountNonce(ctx context.Context, publicKey string) error {
	if err := ap.Noncestore.SetNewAccountNonce(ctx, publicKey); err != nil {
		return err
	}

	return nil
}
