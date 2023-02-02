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

// Load config file.
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

// Load logger.
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

// Load Celo chain provider.
func initCeloProvider() (*celo.Provider, error) {
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
		return nil, err
	}

	return provider, nil
}

// Load postgres pool.
func initPostgresPool() (*pgxpool.Pool, error) {
	poolOpts := postgres.PostgresPoolOpts{
		DSN: ko.MustString("postgres.dsn"),
	}

	pool, err := postgres.NewPostgresPool(poolOpts)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

// Load separate redis connection for the tasker on a reserved db namespace.
func initAsynqRedisPool() (*redis.RedisPool, error) {
	poolOpts := redis.RedisPoolOpts{
		DSN:          ko.MustString("asynq.dsn"),
		MinIdleConns: ko.MustInt("redis.minconn"),
	}

	pool, err := redis.NewRedisPool(poolOpts)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

// Common redis connection on a different db namespace from the takser.
func initCommonRedisPool() (*redis.RedisPool, error) {
	poolOpts := redis.RedisPoolOpts{
		DSN:          ko.MustString("redis.dsn"),
		MinIdleConns: ko.MustInt("redis.minconn"),
	}

	pool, err := redis.NewRedisPool(poolOpts)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

// Load postgres based keystore
func initPostgresKeystore(postgresPool *pgxpool.Pool) (keystore.Keystore, error) {
	keystore, err := keystore.NewPostgresKeytore(keystore.Opts{
		PostgresPool: postgresPool,
		Logg:         lo,
	})
	if err != nil {
		return nil, err
	}

	return keystore, nil
}

// Load redis backed noncestore.
func initRedisNoncestore(redisPool *redis.RedisPool, celoProvider *celo.Provider) nonce.Noncestore {
	return nonce.NewRedisNoncestore(nonce.Opts{
		RedisPool:    redisPool,
		CeloProvider: celoProvider,
	})
}

// Load global lock provider.
func initLockProvider(redisPool redislock.RedisClient) *redislock.Client {
	return redislock.New(redisPool)
}

// Load tasker client.
func initTaskerClient(redisPool *redis.RedisPool) *tasker.TaskerClient {
	return tasker.NewTaskerClient(tasker.TaskerClientOpts{
		RedisPool:     redisPool,
		TaskRetention: time.Duration(ko.MustInt64("asynq.task_retention_hrs")) * time.Hour,
	})
}
