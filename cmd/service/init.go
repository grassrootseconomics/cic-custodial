package main

import (
	"context"
	"log"
	"strings"

	"github.com/grassrootseconomics/cic-custodial/internal/keystore"
	postgres_keystore "github.com/grassrootseconomics/cic-custodial/internal/keystore/providers/postgres"
	"github.com/grassrootseconomics/cic-custodial/internal/noncestore"
	redis_noncestore "github.com/grassrootseconomics/cic-custodial/internal/noncestore/providers/redis"
	tasker_client "github.com/grassrootseconomics/cic-custodial/internal/tasker/client"
	"github.com/grassrootseconomics/cic-go-sdk/chain"
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
		log.Fatalf("could not load config file: %v", err)
	}

	if err := ko.Load(env.Provider("", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(
			strings.TrimPrefix(s, "")), "_", ".")
	}), nil); err != nil {
		log.Fatalf("could not override config from env vars: %v", err)
	}

	return ko
}

func initLogger() logf.Logger {
	loOpts := logf.Opts{
		EnableColor:  true,
		EnableCaller: true,
	}

	if ko.Bool("service.debug") {
		loOpts.Level = logf.DebugLevel
	} else {
		loOpts.Level = logf.InfoLevel
	}

	lo := logf.New(loOpts)

	return lo
}

func initChainProvider() *chain.Provider {
	provider, err := chain.NewProvider(ko.MustString("chain.endpoint"))
	if err != nil {
		lo.Fatal("initChainProvider", "error", err)
	}

	lo.Debug("successfully parsed kitabu rpc endpoint")
	return provider
}

func initTaskerClient() *tasker_client.TaskerClient {
	return tasker_client.NewTaskerClient(tasker_client.Opts{
		RedisDSN: ko.MustString("redis.dsn"),
	})
}

func initKeystore() keystore.Keystore {
	switch provider := ko.MustString("keystore.provider"); provider {
	case "postgres":
		pgKeystore, err := postgres_keystore.NewPostgresKeytore(postgres_keystore.Opts{
			PostgresDSN: ko.MustString("keystore.dsn"),
		})
		if err != nil {
			lo.Fatal("initKeystore", "error", err)
		}

		return pgKeystore
	case "vault":
		lo.Fatal("initKeystore", "error", "not implemented")
	default:
		lo.Fatal("initKeystore", "error", "no keystore provider selected")
	}
	return nil
}

func initNoncestore() noncestore.Noncestore {
	var loadedNoncestore noncestore.Noncestore

	switch provider := ko.MustString("noncestore.provider"); provider {
	case "redis":
		redisNoncestore, err := redis_noncestore.NewRedisNoncestore(redis_noncestore.Opts{
			RedisDSN:      ko.MustString("redis.dsn"),
			RedisDB:       2,
			MinIdleConns:  5,
			PoolSize:      10,
			ChainProvider: chainProvider,
		})
		if err != nil {
			lo.Fatal("initNoncestore", "error", err)
		}

		loadedNoncestore = redisNoncestore
	case "postgres":
		lo.Fatal("initNoncestore", "error", "not implemented")
	default:
		lo.Fatal("initNoncestore", "error", "no noncestore provider selected")
	}

	currentSystemNonce, err := loadedNoncestore.Peek(context.Background(), ko.MustString("admin.public"))
	lo.Debug("initNoncestore: loaded (noncestore) system nonce", "nonce", currentSystemNonce)
	if err != nil {
		nonce, err := loadedNoncestore.SyncNetworkNonce(context.Background(), ko.MustString("admin.public"))
		lo.Debug("initNoncestore: syncing system nonce", "nonce", nonce)
		if err != nil {
			lo.Fatal("initNonceStore", "error", "system account nonce sync failed")
		}
	}

	return loadedNoncestore
}
