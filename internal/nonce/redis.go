package nonce

import (
	"context"

	celo "github.com/grassrootseconomics/cic-celo-sdk"
	redispool "github.com/grassrootseconomics/cic-custodial/pkg/redis"
	"github.com/grassrootseconomics/w3-celo-patch"
	"github.com/grassrootseconomics/w3-celo-patch/module/eth"
)

type Opts struct {
	RedisPool    *redispool.RedisPool
	CeloProvider *celo.Provider
}

// RedisNoncestore implements `Noncestore`
type RedisNoncestore struct {
	chainProvider *celo.Provider
	redis         *redispool.RedisPool
}

func NewRedisNoncestore(o Opts) Noncestore {
	return &RedisNoncestore{
		redis:         o.RedisPool,
		chainProvider: o.CeloProvider,
	}
}

func (n *RedisNoncestore) Peek(ctx context.Context, publicKey string) (uint64, error) {
	nonce, err := n.redis.Client.Get(ctx, publicKey).Uint64()
	if err != nil {
		return 0, err
	}

	return nonce, nil
}

func (n *RedisNoncestore) Acquire(ctx context.Context, publicKey string) (uint64, error) {
	var (
		nonce uint64
	)

	nonce, err := n.redis.Client.Get(ctx, publicKey).Uint64()
	if err != nil {
		return 0, nil
	}

	err = n.redis.Client.Incr(ctx, publicKey).Err()
	if err != nil {
		return 0, err
	}

	return nonce, nil
}

func (n *RedisNoncestore) Return(ctx context.Context, publicKey string) error {
	nonce, err := n.redis.Client.Get(ctx, publicKey).Uint64()
	if err != nil {
		return err
	}

	if nonce > 0 {
		err = n.redis.Client.Decr(ctx, publicKey).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *RedisNoncestore) SyncNetworkNonce(ctx context.Context, publicKey string) (uint64, error) {
	var (
		networkNonce uint64
	)

	err := n.chainProvider.Client.CallCtx(
		ctx,
		eth.Nonce(w3.A(publicKey), nil).Returns(&networkNonce),
	)
	if err != nil {
		return 0, err
	}

	err = n.redis.Client.Set(ctx, publicKey, networkNonce, 0).Err()
	if err != nil {
		return 0, err
	}

	return networkNonce, nil
}

func (n *RedisNoncestore) SetNewAccountNonce(ctx context.Context, publicKey string) error {
	err := n.redis.Client.Set(ctx, publicKey, 0, 0).Err()
	if err != nil {
		return err
	}

	return nil
}
