package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisPoolOpts struct {
	DSN string
	// Debug        bool
	// Logg         logg.RedisLogg
	MinIdleConns int
}

type RedisPool struct {
	Client *redis.Client
}

func NewRedisPool(o RedisPoolOpts) (*RedisPool, error) {
	redisOpts, err := redis.ParseURL(o.DSN)
	if err != nil {
		return nil, err
	}

	redisOpts.MinIdleConns = o.MinIdleConns

	redisClient := redis.NewClient(redisOpts)

	// if o.Debug {
	// 	redis.SetLogger(o.Logg)
	// }

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

func (r *RedisPool) MakeRedisClient() interface{} {
	return r.Client
}
