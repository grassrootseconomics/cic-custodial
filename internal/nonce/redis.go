package nonce

import (
	"context"

	redispool "github.com/grassrootseconomics/cic-custodial/pkg/redis"
)

type (
	Opts struct {
		RedisPool *redispool.RedisPool
	}

	// RedisNoncestore implements `Noncestore`
	RedisNoncestore struct {
		redis *redispool.RedisPool
	}
)

func NewRedisNoncestore(o Opts) Noncestore {
	return &RedisNoncestore{
		redis: o.RedisPool,
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

func (n *RedisNoncestore) SetAccountNonce(ctx context.Context, publicKey string, nonce uint64) error {
	if err := n.redis.Client.Set(ctx, publicKey, nonce, 0).Err(); err != nil {
		return err
	}

	return nil
}
