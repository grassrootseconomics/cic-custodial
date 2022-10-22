package main

import (
	"log"
	"strings"

	"github.com/grassrootseconomics/cic-custodial/internal/keystore"
	postgres_keystore "github.com/grassrootseconomics/cic-custodial/internal/keystore/providers/postgres"
	"github.com/grassrootseconomics/cic-custodial/internal/noncestore"
	redis_noncestore "github.com/grassrootseconomics/cic-custodial/internal/noncestore/providers/redis"
	system_provider "github.com/grassrootseconomics/cic-custodial/internal/system"
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

func initSystemProvider() *system_provider.SystemProvider {
	systemProvider, err := system_provider.NewSystemProvider(system_provider.Opts{
		SystemPublicKey:  ko.MustString("admin.public"),
		SystemPrivateKey: ko.MustString("admin.key"),
		ChainProvider:    chainProvider,
	})
	if err != nil {
		lo.Fatal("initSystemProvider", "error", err)
	}

	return systemProvider
}

func initTaskerClient() *tasker_client.TaskerClient {
	return tasker_client.NewTaskerClient(tasker_client.Opts{
		RedisDSN: ko.MustString("tasker.dsn"),
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
	switch provider := ko.MustString("noncestore.provider"); provider {
	case "redis":
		redisNoncestore, err := redis_noncestore.NewRedisNoncestore(redis_noncestore.Opts{
			RedisDSN:      ko.MustString("noncestore.dsn"),
			RedisDB:       2,
			MinIdleConns:  8,
			PoolSize:      15,
			ChainProvider: chainProvider,
		})
		if err != nil {
			lo.Fatal("initNoncestore", "error", err)
		}

		return redisNoncestore
	case "postgres":
		lo.Fatal("initNoncestore", "error", "not implemented")
	default:
		lo.Fatal("initNoncestore", "error", "no noncestore provider selected")
	}
	return nil
}
