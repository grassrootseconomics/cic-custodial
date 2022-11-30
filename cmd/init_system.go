package main

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"time"

	eth_crypto "github.com/celo-org/celo-blockchain/crypto"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/w3-celo-patch"
)

func initSystemContainer() *tasker.SystemContainer {
	return &tasker.SystemContainer{
		GasRefillThreshold:    big.NewInt(ko.MustInt64("system.gas_refill_threshold")),
		GasRefillValue:        big.NewInt(ko.MustInt64("system.gas_refill_value")),
		GiftableGasValue:      big.NewInt(ko.MustInt64("system.giftable_gas_value")),
		GiftableToken:         w3.A(ko.MustString("system.giftable_token_address")),
		GiftableTokenValue:    big.NewInt(ko.MustInt64("system.giftable_token_value")),
		LockPrefix:            ko.MustString("system.lock_prefix"),
		LockTimeout:           1 * time.Second,
		PrivateKey:            initSystemKey(),
		PublicKey:             ko.MustString("system.public_key"),
		TokenDecimals:         ko.MustInt("system.token_decimals"),
		TokenTransferGasLimit: uint64(ko.MustInt64("system.token_transfer_gas_limit")),
	}
}

func initSystemKey() *ecdsa.PrivateKey {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	currentSystemNonce, err := redisNoncestore.Peek(ctx, ko.MustString("system.public_key"))
	lo.Debug("initNoncestore: loaded (noncestore) system nonce", "nonce", currentSystemNonce)
	if err != nil {
		nonce, err := redisNoncestore.SyncNetworkNonce(ctx, ko.MustString("system.public_key"))
		lo.Debug("initNoncestore: syncing system nonce", "nonce", nonce)
		if err != nil {
			lo.Fatal("initNonceStore", "error", "system account nonce sync failed")
		}
	}

	loadedPrivateKey, err := eth_crypto.HexToECDSA(ko.MustString("system.private_key"))
	if err != nil {
		lo.Fatal("Failed to load system private key", "error", err)
	}

	return loadedPrivateKey
}
