package server

import (
	"crypto/ecdsa"
	"time"

	"github.com/bsm/redislock"
	eth_crypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/go-redis/redis/v8"
	"github.com/grassrootseconomics/cic-custodial/internal/keystore"
	"github.com/grassrootseconomics/cic-custodial/internal/noncestore"
	tasker_client "github.com/grassrootseconomics/cic-custodial/internal/tasker/client"
	"github.com/grassrootseconomics/cic-go-sdk/chain"
	"github.com/hibiken/asynq"
	"github.com/zerodha/logf"
)

const (
	LockTTL = 1 * time.Second
)

type Opts struct {
	SystemPublicKey       string
	SystemPrivateKey      string
	ChainProvider         *chain.Provider
	Keystore              keystore.Keystore
	Noncestore            noncestore.Noncestore
	TaskerClient          *tasker_client.TaskerClient
	RedisDSN              string
	RedisLockDB           int
	RedisLockMinIdleConns int
	RedisLockPoolSize     int
	Concurrency           int
	Logger                logf.Logger
}

type TaskerProcessor struct {
	SystemPublicKey  string
	SystemPrivateKey *ecdsa.PrivateKey
	ChainProvider    *chain.Provider
	Noncestore       noncestore.Noncestore
	Keystore         keystore.Keystore
	LockProvider     *redislock.Client
	TaskerClient     *tasker_client.TaskerClient
}

type TaskerServer struct {
	Server *asynq.Server
	Mux    *asynq.ServeMux
}

func NewTaskerServer(o Opts) (*TaskerServer, error) {
	loadedPrivateKey, err := eth_crypto.HexToECDSA(o.SystemPrivateKey)
	if err != nil {
		return nil, err
	}

	redisLockClient := redis.NewClient(&redis.Options{
		Addr:         o.RedisDSN,
		DB:           o.RedisLockDB,
		MinIdleConns: o.RedisLockMinIdleConns,
		PoolSize:     o.RedisLockPoolSize,
	})

	taskerProcessor := &TaskerProcessor{
		SystemPublicKey:  o.SystemPublicKey,
		SystemPrivateKey: loadedPrivateKey,
		ChainProvider:    o.ChainProvider,
		Noncestore:       o.Noncestore,
		Keystore:         o.Keystore,
		TaskerClient:     o.TaskerClient,
		LockProvider:     redislock.New(redisLockClient),
	}

	asynqServer := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr: o.RedisDSN,
		},
		asynq.Config{
			Concurrency: o.Concurrency,
			Logger:      asynqCompatibleLogger(o.Logger),
			RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
				if n < 6 {
					return 1 * time.Second
				} else {
					return asynq.DefaultRetryDelayFunc(n, e, t)
				}
			},
			IsFailure: func(err error) bool {
				switch err {
				case redislock.ErrNotObtained:
					return false
				default:
					return true
				}
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(string(tasker_client.ActivateAccountTask), taskerProcessor.activateAccountProcessor)
	mux.HandleFunc(string(tasker_client.SetNewAccountNonceTask), taskerProcessor.setNewAccountNonce)
	mux.HandleFunc(string(tasker_client.GiftGasTask), taskerProcessor.giftGasProcessor)
	mux.HandleFunc(string(tasker_client.TxDispatchTask), taskerProcessor.txDispatcher)

	return &TaskerServer{
		Server: asynqServer,
		Mux:    mux,
	}, nil
}
