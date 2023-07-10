package custodial

import (
	"context"
	"crypto/ecdsa"
	"time"

	"github.com/bsm/redislock"
	"github.com/celo-org/celo-blockchain/common"
	eth_crypto "github.com/celo-org/celo-blockchain/crypto"
	"github.com/grassrootseconomics/celoutils"
	"github.com/grassrootseconomics/cic-custodial/internal/nonce"
	"github.com/grassrootseconomics/cic-custodial/internal/store"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/cic-custodial/pkg/util"
	"github.com/grassrootseconomics/w3-celo-patch"
	"github.com/redis/go-redis/v9"
	"github.com/zerodha/logf"
)

type (
	Opts struct {
		ApprovalTimeout  time.Duration
		CeloProvider     *celoutils.Provider
		LockProvider     *redislock.Client
		Logg             logf.Logger
		Noncestore       nonce.Noncestore
		Store            store.Store
		RedisClient      *redis.Client
		RegistryAddress  string
		SystemPrivateKey string
		SystemPublicKey  string
		TaskerClient     *tasker.TaskerClient
	}

	Custodial struct {
		ApprovalTimeout  time.Duration
		Abis             map[string]*w3.Func
		CeloProvider     *celoutils.Provider
		LockProvider     *redislock.Client
		Logg             logf.Logger
		Noncestore       nonce.Noncestore
		Store            store.Store
		RedisClient      *redis.Client
		RegistryMap      map[string]common.Address
		SystemPrivateKey *ecdsa.PrivateKey
		SystemPublicKey  string
		TaskerClient     *tasker.TaskerClient
	}
)

func NewCustodial(o Opts) (*Custodial, error) {
	ctx, cancel := context.WithTimeout(context.Background(), util.SLATimeout)
	defer cancel()

	registryMap, err := o.CeloProvider.RegistryMap(ctx, celoutils.HexToAddress(o.RegistryAddress))
	if err != nil {
		o.Logg.Error("custodial: critical error loading contracts from registry: %v", err)
		return nil, err
	}

	systemNonce, err := o.Noncestore.Peek(ctx, o.SystemPublicKey)
	if err != nil {
		return nil, err
	}

	o.Logg.Info("custodial: loaded_nonce", "system_nonce", systemNonce)

	privateKey, err := eth_crypto.HexToECDSA(o.SystemPrivateKey)
	if err != nil {
		return nil, err
	}

	return &Custodial{
		ApprovalTimeout:  o.ApprovalTimeout,
		Abis:             initAbis(),
		CeloProvider:     o.CeloProvider,
		LockProvider:     o.LockProvider,
		Logg:             o.Logg,
		Noncestore:       o.Noncestore,
		Store:            o.Store,
		RedisClient:      o.RedisClient,
		RegistryMap:      registryMap,
		SystemPrivateKey: privateKey,
		SystemPublicKey:  o.SystemPublicKey,
		TaskerClient:     o.TaskerClient,
	}, nil

}
