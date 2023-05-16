package nonce

import (
	"context"
	"errors"

	"github.com/grassrootseconomics/celoutils"
	"github.com/grassrootseconomics/cic-custodial/internal/store"
	redispool "github.com/grassrootseconomics/cic-custodial/pkg/redis"
	"github.com/grassrootseconomics/w3-celo-patch/module/eth"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type (
	Opts struct {
		ChainProvider *celoutils.Provider
		RedisPool     *redispool.RedisPool
		Store         store.Store
	}

	// RedisNoncestore implements `Noncestore`
	RedisNoncestore struct {
		chainProvider *celoutils.Provider
		redis         *redispool.RedisPool
		store         store.Store
	}
)

func NewRedisNoncestore(o Opts) Noncestore {
	return &RedisNoncestore{
		chainProvider: o.ChainProvider,
		redis:         o.RedisPool,
		store:         o.Store,
	}
}

func (n *RedisNoncestore) Peek(ctx context.Context, publicKey string) (uint64, error) {
	nonce, err := n.redis.Client.Get(ctx, publicKey).Uint64()
	if err != nil {
		if err == redis.Nil {
			nonce, err = n.bootstrap(ctx, publicKey)
			if err != nil {
				return 0, err
			}

			return nonce, nil
		} else {
			return 0, err
		}
	}

	return nonce, nil
}

func (n *RedisNoncestore) Acquire(ctx context.Context, publicKey string) (uint64, error) {
	var (
		nonce uint64
	)

	nonce, err := n.redis.Client.Get(ctx, publicKey).Uint64()
	if err != nil {
		if err == redis.Nil {
			nonce, err = n.bootstrap(ctx, publicKey)
			if err != nil {
				return 0, err
			}
		} else {
			return 0, err
		}
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

// bootstrap can be used to restore a destroyed redis nonce cache automatically.
// It first uses the otx_sign table as a source of nonce values.
// If the otx_sign table is corrupted, it can fallback to the network nonce.
// Ideally, the redis nonce cache should never be lost.
func (n *RedisNoncestore) bootstrap(ctx context.Context, publicKey string) (uint64, error) {
	lastDbNonce, err := n.store.GetNextNonce(ctx, publicKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err := n.chainProvider.Client.CallCtx(
				ctx,
				eth.Nonce(celoutils.HexToAddress(publicKey), nil).Returns(&lastDbNonce),
			)
			if err != nil {
				return 0, err
			}
		} else {
			return 0, err
		}
	}

	if err := n.SetAccountNonce(ctx, publicKey, lastDbNonce); err != nil {
		return 0, err
	}

	return lastDbNonce, nil
}
