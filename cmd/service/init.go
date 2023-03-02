package main

import (
	"context"
	"strings"
	"time"

	"github.com/bsm/redislock"
	"github.com/grassrootseconomics/celoutils"
	"github.com/grassrootseconomics/cic-custodial/internal/events"
	"github.com/grassrootseconomics/cic-custodial/internal/keystore"
	"github.com/grassrootseconomics/cic-custodial/internal/nonce"
	"github.com/grassrootseconomics/cic-custodial/internal/queries"
	"github.com/grassrootseconomics/cic-custodial/internal/store"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/cic-custodial/pkg/logg"
	"github.com/grassrootseconomics/cic-custodial/pkg/postgres"
	"github.com/grassrootseconomics/cic-custodial/pkg/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/knadh/goyesql/v2"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
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
		lo.Fatal("Could not load config file", "error", err)
	}

	if err := ko.Load(env.Provider("CUSTODIAL_", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(
			strings.TrimPrefix(s, "CUSTODIAL_")), "__", ".")
	}), nil); err != nil {
		lo.Fatal("Could not override config from env vars", "error", err)
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

// Load postgres pool.
func initPostgresPool() *pgxpool.Pool {
	poolOpts := postgres.PostgresPoolOpts{
		DSN:                  ko.MustString("postgres.dsn"),
		MigrationsFolderPath: migrationsFolderFlag,
	}

	pool, err := postgres.NewPostgresPool(context.Background(), poolOpts)
	if err != nil {
		lo.Fatal("init: critical error connecting to postgres", "error", err)
	}

	return pool
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

// Load SQL statements into struct.
func initQueries() *queries.Queries {
	parsedQueries, err := goyesql.ParseFile(queriesFlag)
	if err != nil {
		lo.Fatal("init: critical error loading SQL queries", "error", err)
	}

	loadedQueries, err := queries.LoadQueries(parsedQueries)
	if err != nil {
		lo.Fatal("init: critical error loading SQL queries", "error", err)
	}

	return loadedQueries
}

// Load postgres based keystore.
func initPostgresKeystore(postgresPool *pgxpool.Pool, queries *queries.Queries) keystore.Keystore {
	keystore := keystore.NewPostgresKeytore(keystore.Opts{
		PostgresPool: postgresPool,
		Queries:      queries,
	})

	return keystore
}

// Load redis backed noncestore.
func initRedisNoncestore(redisPool *redis.RedisPool, celoProvider *celoutils.Provider) nonce.Noncestore {
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

// Load Postgres store
func initPostgresStore(postgresPool *pgxpool.Pool, queries *queries.Queries) store.Store {
	return store.NewPostgresStore(store.Opts{
		PostgresPool: postgresPool,
		Queries:      queries,
	})
}

// Init JetStream context for tasker events.
func initJetStream(pgStore store.Store) *events.JetStream {
	jsEmitter, err := events.NewJetStreamEventEmitter(events.JetStreamOpts{
		Logg:            lo,
		PgStore:         pgStore,
		ServerUrl:       ko.MustString("jetstream.endpoint"),
		PersistDuration: time.Duration(ko.MustInt("jetstream.persist_duration_hrs")) * time.Hour,
		DedupDuration:   time.Duration(ko.MustInt("jetstream.dedup_duration_hrs")) * time.Hour,
	})

	if err != nil {
		lo.Fatal("main: critical error loading jetstream event emitter")

	}

	return jsEmitter
}
