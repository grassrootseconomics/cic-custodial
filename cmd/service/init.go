package main

import (
	"context"
	"strings"

	"github.com/bsm/redislock"
	"github.com/grassrootseconomics/celoutils"
	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/grassrootseconomics/cic-custodial/internal/nonce"
	"github.com/grassrootseconomics/cic-custodial/internal/store"
	"github.com/grassrootseconomics/cic-custodial/internal/sub"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/cic-custodial/pkg/logg"
	"github.com/grassrootseconomics/cic-custodial/pkg/redis"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/nats-io/nats.go"
	"github.com/zerodha/logf"
)

// Load logger.
func initLogger() logf.Logger {
	loggOpts := logg.LoggOpts{}

	if debugFlag {
		loggOpts.Color = true
		loggOpts.Caller = true
		loggOpts.Debug = true
	}

	return logg.NewLogg(loggOpts)
}

// Load config file.
func initConfig() *koanf.Koanf {
	var (
		ko = koanf.New(".")
	)

	confFile := file.Provider(confFlag)
	if err := ko.Load(confFile, toml.Parser()); err != nil {
		lo.Fatal("init: could not load config file", "error", err)
	}

	if err := ko.Load(env.Provider("CUSTODIAL_", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(
			strings.TrimPrefix(s, "CUSTODIAL_")), "__", ".")
	}), nil); err != nil {
		lo.Fatal("init: could not override config from env vars", "error", err)
	}

	if debugFlag {
		ko.Print()
	}

	return ko
}

// Load Celo chain provider.
func initCeloProvider() *celoutils.Provider {
	providerOpts := celoutils.ProviderOpts{
		RpcEndpoint: ko.MustString("chain.rpc_endpoint"),
	}

	if ko.Bool("chain.testnet") {
		providerOpts.ChainId = celoutils.TestnetChainId
	} else {
		providerOpts.ChainId = celoutils.MainnetChainId
	}

	provider, err := celoutils.NewProvider(providerOpts)
	if err != nil {
		lo.Fatal("init: critical error loading chain provider", "error", err)
	}

	return provider
}

// Load separate redis connection for the tasker on a reserved db namespace.
func initAsynqRedisPool() *redis.RedisPool {
	poolOpts := redis.RedisPoolOpts{
		DSN:          ko.MustString("asynq.dsn"),
		MinIdleConns: ko.MustInt("redis.min_idle_conn"),
	}

	pool, err := redis.NewRedisPool(context.Background(), poolOpts)
	if err != nil {
		lo.Fatal("init: critical error connecting to asynq redis db", "error", err)
	}

	return pool
}

// Common redis connection on a different db namespace from the takser.
func initCommonRedisPool() *redis.RedisPool {
	poolOpts := redis.RedisPoolOpts{
		DSN:          ko.MustString("redis.dsn"),
		MinIdleConns: ko.MustInt("redis.min_idle_conn"),
	}

	pool, err := redis.NewRedisPool(context.Background(), poolOpts)
	if err != nil {
		lo.Fatal("init: critical error connecting to common redis db", "error", err)
	}

	return pool
}

// Load redis backed noncestore.
func initRedisNoncestore(redisPool *redis.RedisPool, chainProvider *celoutils.Provider, store store.Store) nonce.Noncestore {
	return nonce.NewRedisNoncestore(nonce.Opts{
		ChainProvider: chainProvider,
		RedisPool:     redisPool,
		Store:         store,
	})
}

// Load global lock provider.
func initLockProvider(redisPool redislock.RedisClient) *redislock.Client {
	return redislock.New(redisPool)
}

// Load tasker client.
func initTaskerClient(redisPool *redis.RedisPool) *tasker.TaskerClient {
	return tasker.NewTaskerClient(tasker.TaskerClientOpts{
		RedisPool: redisPool,
	})
}

// Load Postgres store.
func initPgStore() store.Store {
	store, err := store.NewPgStore(store.Opts{
		DSN:                  ko.MustString("postgres.dsn"),
		MigrationsFolderPath: migrationsFolderFlag,
		QueriesFolderPath:    queriesFlag,
	})
	if err != nil {
		lo.Fatal("init: critical error loading Postgres store", "error", err)
	}

	return store
}

// Init JetStream context for both pub/sub.
func initJetStream() (*nats.Conn, nats.JetStreamContext) {
	natsConn, err := nats.Connect(ko.MustString("jetstream.endpoint"))
	if err != nil {
		lo.Fatal("init: critical error connecting to NATS", "error", err)
	}

	js, err := natsConn.JetStream()
	if err != nil {
		lo.Fatal("init: bad JetStream opts", "error", err)

	}

	return natsConn, js
}

func initSub(natsConn *nats.Conn, jsCtx nats.JetStreamContext, cu *custodial.Custodial) *sub.Sub {
	sub, err := sub.NewSub(sub.SubOpts{
		CustodialContainer: cu,
		JsCtx:              jsCtx,
		Logg:               lo,
		NatsConn:           natsConn,
	})
	if err != nil {
		lo.Fatal("init: critical error bootstrapping sub", "error", err)
	}

	return sub
}
