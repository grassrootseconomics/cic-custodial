package system

import (
	"context"
	"sync"

	"github.com/grassrootseconomics/cic-custodial/internal/noncestore"
	"github.com/grassrootseconomics/cic-go-sdk/chain"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
)

type Opts struct {
	ChainProvider  *chain.Provider
	AccountAddress string
}

type SystemNoncestore struct {
	mx         sync.Mutex
	nonceValue uint64
}

func NewSystemNoncestore(o Opts) (noncestore.SystemNoncestore, error) {
	var (
		networkNonce uint64
	)

	err := o.ChainProvider.EthClient.CallCtx(
		context.Background(),
		eth.Nonce(w3.A(o.AccountAddress), nil).Returns(&networkNonce),
	)
	if err != nil {
		return nil, err
	}

	return &SystemNoncestore{
		nonceValue: networkNonce,
	}, nil
}

func (ns *SystemNoncestore) Peek() uint64 {
	ns.mx.Lock()
	defer ns.mx.Unlock()

	return ns.nonceValue
}

func (ns *SystemNoncestore) Acquire() uint64 {
	ns.mx.Lock()
	defer ns.mx.Unlock()

	nextNonce := ns.nonceValue
	ns.nonceValue++

	return nextNonce
}

func (ns *SystemNoncestore) Return() {
	ns.mx.Lock()
	defer ns.mx.Unlock()

	if ns.nonceValue > 0 {
		ns.nonceValue--
	}
}
