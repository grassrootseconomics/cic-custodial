package actions

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lmittmann/w3/module/eth"
)

func (ap *ActionsProvider) DispatchSignedTx(ctx context.Context, builtTx *types.Transaction) (string, error) {
	var txHash common.Hash

	if err := ap.ChainProvider.EthClient.CallCtx(
		ctx,
		eth.SendTx(builtTx).Returns(&txHash),
	); err != nil {
		return "", err
	}

	return txHash.String(), nil
}
