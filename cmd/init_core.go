package main

import (
	"strings"
	"time"

	"github.com/bsm/redislock"
	celo "github.com/grassrootseconomics/cic-celo-sdk"
	"github.com/grassrootseconomics/cic-custodial/internal/keystore"
	"github.com/grassrootseconomics/cic-custodial/internal/nonce"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/cic-custodial/pkg/logg"
	"github.com/grassrootseconomics/cic-custodial/pkg/postgres"
	"github.com/grassrootseconomics/cic-custodial/pkg/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/zerodha/logf"
)

func initConfig(configFilePath string) *koanf.Koanf {
	var (
		ko = koanf.New(".")
	)

	confFile := file.Provider(configFilePath)
	if err := ko.Load(confFile, toml.Parser()); err != nil {
		lo.Fatal("Could not load config file", "error", err)
	}

	if err := ko.Load(env.Provider("", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(
			strings.TrimPrefix(s, "")), "_", ".")
	}), nil); err != nil {
		lo.Fatal("Could not override config from env vars", "error", err)
	}

	return ko
}

func initLogger(debug bool) logf.Logger {
	loggOpts := logg.LoggOpts{
		Color: true,
	}

	if debug {
		loggOpts.Caller = true
		loggOpts.Debug = true
	}

	return logg.NewLogg(loggOpts)
}

func initCeloProvider() *celo.Provider {
	providerOpts := celo.ProviderOpts{
		RpcEndpoint: ko.MustString("chain.rpc_endpoint"),
	}

	if ko.Bool("chain.testnet") {
		providerOpts.ChainId = celo.TestnetChainId
	} else {
		providerOpts.ChainId = celo.MainnetChainId
	}

	provider, err := celo.NewProvider(providerOpts)
	if err != nil {
		lo.Fatal("initChainProvider", "error", err)
	}

	return provider
}

func initPostgresPool() *pgxpool.Pool {
	poolOpts := postgres.PostgresPoolOpts{
		DSN: ko.MustString("postgres.dsn"),
	}

	pool, err := postgres.NewPostgresPool(poolOpts)
	if err != nil {
		lo.Fatal("initPostgresPool", "error", err)
	}

	return pool
}

func initKeystore() keystore.Keystore {
	keystore, err := keystore.NewPostgresKeytore(keystore.Opts{
		PostgresPool: postgresPool,
		Logg:         lo,
	})
	if err != nil {
		lo.Fatal("initKeystore", "error", err)
	}

	return keystore
}

func initAsynqRedisPool() *redis.RedisPool {
	poolOpts := redis.RedisPoolOpts{
		DSN:          ko.MustString("asynq.dsn"),
		MinIdleConns: ko.MustInt("redis.minconn"),
	}

	pool, err := redis.NewRedisPool(poolOpts)
	if err != nil {
		lo.Fatal("initAsynqRedisPool", "error", err)
	}

	return pool
}

func initCommonRedisPool() *redis.RedisPool {
	poolOpts := redis.RedisPoolOpts{
		DSN:          ko.MustString("redis.dsn"),
		MinIdleConns: ko.MustInt("redis.minconn"),
	}

	pool, err := redis.NewRedisPool(poolOpts)
	if err != nil {
		lo.Fatal("initCommonRedisPool", "error", err)
	}

	return pool
}

func initRedisNoncestore() nonce.Noncestore {
	return nonce.NewRedisNoncestore(nonce.Opts{
		RedisPool:     commonRedisPool,
		ChainProvider: celoProvider,
	})
}

func initLockProvider() *redislock.Client {
	return redislock.New(commonRedisPool.Client)
}

func initTaskerClient() *tasker.TaskerClient {
	return tasker.NewTaskerClient(tasker.TaskerClientOpts{
		RedisPool:     asynqRedisPool,
		TaskRetention: time.Hour * 12,
	})
}
