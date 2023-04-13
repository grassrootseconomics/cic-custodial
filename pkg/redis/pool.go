package redis

import (
	"context"

	"github.com/grassrootseconomics/cic-custodial/pkg/util"
	"github.com/redis/go-redis/v9"
)

const (
	systemGlobalLockKey = "system:global_lock"
)

type (
	RedisPoolOpts struct {
		DSN          string
		MinIdleConns int
	}

	RedisPool struct {
		Client *redis.Client
	}
)

// NewRedisPool creates a reusable connection across the cic-custodial componenent.
// Note: Each db namespace requires its own connection pool.
func NewRedisPool(ctx context.Context, o RedisPoolOpts) (*RedisPool, error) {
	redisOpts, err := redis.ParseURL(o.DSN)
	if err != nil {
		return nil, err
	}

	redisOpts.MinIdleConns = o.MinIdleConns

	redisClient := redis.NewClient(redisOpts)

	ctx, cancel := context.WithTimeout(ctx, util.SLATimeout)
	defer cancel()

	if err := redisClient.SetNX(ctx, systemGlobalLockKey, false, 0).Err(); err != nil {
		return nil, err
	}

	return &RedisPool{
		Client: redisClient,
	}, nil
}

// Interface adapter for asynq to resuse the same Redis connection pool.
func (r *RedisPool) MakeRedisClient() interface{} {
	return r.Client
}
