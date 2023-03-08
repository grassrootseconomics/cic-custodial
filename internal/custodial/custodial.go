package custodial

import (
	"crypto/ecdsa"
	"math/big"
	"time"

	"github.com/bsm/redislock"
	"github.com/celo-org/celo-blockchain/common"
	"github.com/go-redis/redis/v8"
	"github.com/grassrootseconomics/celoutils"
	"github.com/grassrootseconomics/cic-custodial/internal/keystore"
	"github.com/grassrootseconomics/cic-custodial/internal/nonce"
	"github.com/grassrootseconomics/cic-custodial/internal/pub"
	"github.com/grassrootseconomics/cic-custodial/internal/store"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/w3-celo-patch"
)

type (
	SystemContainer struct {
		Abis                  map[string]*w3.Func
		AccountIndexContract  common.Address
		GasFaucetContract     common.Address
		GasRefillThreshold    *big.Int
		GasRefillValue        *big.Int
		GiftableGasValue      *big.Int
		GiftableToken         common.Address
		GiftableTokenValue    *big.Int
		LockTimeout           time.Duration
		PrivateKey            *ecdsa.PrivateKey
		PublicKey             string
		TokenDecimals         int
		TokenTransferGasLimit uint64
	}
	Custodial struct {
		CeloProvider    *celoutils.Provider
		Keystore        keystore.Keystore
		LockProvider    *redislock.Client
		Noncestore      nonce.Noncestore
		PgStore         store.Store
		Pub             *pub.Pub
		RedisClient     *redis.Client
		SystemContainer *SystemContainer
		TaskerClient    *tasker.TaskerClient
	}
)
