package actions

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/grassrootseconomics/cic-go-sdk/chain"
	"github.com/lmittmann/w3"
)

const (
	initialGiftGasValue = 1000000
	topupGiftGasValue   = 500000
)

func (ap *ActionsProvider) SignGiftGasTx(ctx context.Context, giftTo string) (*types.Transaction, error) {
	nonce, err := ap.Noncestore.Acquire(ctx, ap.SystemPublicKey)
	if err != nil {
		return nil, err
	}

	builtTx, err := ap.ChainProvider.BuildGasTransferTx(ap.SystemPrivateKey, chain.TransactionData{
		To:    w3.A(giftTo),
		Nonce: nonce,
	}, big.NewInt(initialGiftGasValue))
	if err != nil {
		return &types.Transaction{}, err
	}

	return builtTx, nil
}

func (ap *ActionsProvider) SignTopUpGasTx(ctx context.Context, giftTo string) (*types.Transaction, error) {
	nonce, err := ap.Noncestore.Acquire(ctx, ap.SystemPublicKey)
	if err != nil {
		return nil, err
	}

	builtTx, err := ap.ChainProvider.BuildGasTransferTx(ap.SystemPrivateKey, chain.TransactionData{
		To:    w3.A(giftTo),
		Nonce: nonce,
	}, big.NewInt(topupGiftGasValue))
	if err != nil {
		return &types.Transaction{}, err
	}

	return builtTx, nil
}

func (ap *ActionsProvider) SignGiftVouchertx(ctx context.Context, giftTo string) (*types.Transaction, error) {
	return &types.Transaction{}, nil
}
