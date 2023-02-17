package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisPoolOpts struct {
	DSN          string
	MinIdleConns int
}

type RedisPool struct {
	Client *redis.Client
}

// NewRedisPool creates a reusable connection across the cic-custodial componenent.
// Note: Each db namespace requires its own connection pool.
func NewRedisPool(o RedisPoolOpts) (*RedisPool, error) {
	redisOpts, err := redis.ParseURL(o.DSN)
	if err != nil {
		return nil, err
	}

	redisOpts.MinIdleConns = o.MinIdleConns

	redisClient := redis.NewClient(redisOpts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
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
