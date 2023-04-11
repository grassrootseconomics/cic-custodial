package custodial

import (
	"context"
	"crypto/ecdsa"

	"github.com/bsm/redislock"
	"github.com/celo-org/celo-blockchain/common"
	eth_crypto "github.com/celo-org/celo-blockchain/crypto"
	"github.com/grassrootseconomics/celoutils"
	"github.com/grassrootseconomics/cic-custodial/internal/nonce"
	"github.com/grassrootseconomics/cic-custodial/internal/store"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/cic-custodial/pkg/util"
	"github.com/grassrootseconomics/w3-celo-patch"
	"github.com/grassrootseconomics/w3-celo-patch/module/eth"
	"github.com/labstack/gommon/log"
	"github.com/redis/go-redis/v9"
)

type (
	Opts struct {
		CeloProvider     *celoutils.Provider
		LockProvider     *redislock.Client
		Noncestore       nonce.Noncestore
		Store            store.Store
		RedisClient      *redis.Client
		RegistryAddress  string
		SystemPrivateKey string
		SystemPublicKey  string
		TaskerClient     *tasker.TaskerClient
	}

	Custodial struct {
		Abis             map[string]*w3.Func
		CeloProvider     *celoutils.Provider
		LockProvider     *redislock.Client
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
		log.Errorf("err: %v", err)
		return nil, err
	}

	_, err = o.Noncestore.Peek(ctx, o.SystemPublicKey)
	if err == redis.Nil {
		// TODO: Bootsrap from Postgres first
		var networkNonce uint64

		err := o.CeloProvider.Client.CallCtx(
			ctx,
			eth.Nonce(celoutils.HexToAddress(o.SystemPublicKey), nil).Returns(&networkNonce),
		)
		if err != nil {
			return nil, err
		}

		if err := o.Noncestore.SetAccountNonce(ctx, o.SystemPublicKey, networkNonce); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	privateKey, err := eth_crypto.HexToECDSA(o.SystemPrivateKey)
	if err != nil {
		return nil, err
	}

	return &Custodial{
		Abis:             initAbis(),
		CeloProvider:     o.CeloProvider,
		LockProvider:     o.LockProvider,
		Noncestore:       o.Noncestore,
		Store:            o.Store,
		RedisClient:      o.RedisClient,
		RegistryMap:      registryMap,
		SystemPrivateKey: privateKey,
		SystemPublicKey:  o.SystemPublicKey,
		TaskerClient:     o.TaskerClient,
	}, nil

}
