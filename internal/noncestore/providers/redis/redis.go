package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/grassrootseconomics/cic-custodial/internal/noncestore"
	"github.com/grassrootseconomics/cic-go-sdk/chain"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/zerodha/logf"
)

// Opts represents the Redis nonce store specific params
type Opts struct {
	RedisDSN      string
	RedisDB       int
	MinIdleConns  int
	PoolSize      int
	ChainProvider *chain.Provider
	Lo            logf.Logger
}

// RedisNoncestore implements `noncestore.Noncestore`
type RedisNoncestore struct {
	chainProvider *chain.Provider
	redis         *redis.Client
	lo            logf.Logger
}

func NewRedisNoncestore(o Opts) (noncestore.Noncestore, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:         o.RedisDSN,
		DB:           o.RedisDB,
		MinIdleConns: o.MinIdleConns,
		PoolSize:     o.PoolSize,
	})

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &RedisNoncestore{
		redis:         redisClient,
		chainProvider: o.ChainProvider,
		lo:            o.Lo,
	}, nil
}

func (ns *RedisNoncestore) Peek(ctx context.Context, publicKey string) (uint64, error) {
	nonce, err := ns.redis.Get(ctx, publicKey).Uint64()
	if err != nil {
		return 0, err
	}

	return nonce, nil
}

func (ns *RedisNoncestore) Acquire(ctx context.Context, publicKey string) (uint64, error) {
	var (
		nonce uint64
	)

	nonce, err := ns.redis.Get(ctx, publicKey).Uint64()
	if err == redis.Nil {
		networkNonce, err := ns.SyncNetworkNonce(ctx, publicKey)
		if err != nil {
			return 0, err
		}

		nonce = networkNonce
	} else if err != nil {
		return 0, nil
	}

	err = ns.redis.Incr(ctx, publicKey).Err()
	if err != nil {
		return 0, err
	}

	return nonce, nil
}

func (ns *RedisNoncestore) Return(ctx context.Context, publicKey string) (uint64, error) {
	nonce, err := ns.redis.Get(ctx, publicKey).Uint64()
	if err != nil {
		return 0, err
	}

	if nonce > 0 {
		err = ns.redis.Decr(ctx, publicKey).Err()
		if err != nil {
			return 0, err
		}
	}

	return nonce, nil
}

func (ns *RedisNoncestore) SyncNetworkNonce(ctx context.Context, publicKey string) (uint64, error) {
	var (
		networkNonce uint64
	)

	err := ns.chainProvider.EthClient.CallCtx(
		ctx,
		eth.Nonce(w3.A(publicKey), nil).Returns(&networkNonce),
	)
	if err != nil {
		return 0, err
	}

	err = ns.redis.Set(ctx, publicKey, networkNonce, 0).Err()
	if err != nil {
		return 0, err
	}

	return networkNonce, nil
}

func (ns *RedisNoncestore) SetNewAccountNonce(ctx context.Context, publicKey string) error {
	err := ns.redis.Set(ctx, publicKey, 0, 0).Err()
	if err != nil {
		ns.lo.Error("noncestore", "err", err)
		return err
	}

	return nil
}
