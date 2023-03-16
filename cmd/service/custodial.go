package main

import (
	"context"
	"math/big"
	"time"

	eth_crypto "github.com/celo-org/celo-blockchain/crypto"
	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/grassrootseconomics/cic-custodial/internal/nonce"
	"github.com/grassrootseconomics/w3-celo-patch"
	"github.com/redis/go-redis/v9"
)

// Define common smart contrcat ABI's that can be injected into the system container.
// Any relevant function signature that will be used by the custodial system can be defined here.
func initAbis() map[string]*w3.Func {
	return map[string]*w3.Func{
		// Keccak hash -> 0x449a52f8
		"mintTo": w3.MustNewFunc("mintTo(address, uint256)", "bool"),
		// Keccak hash -> 0xa9059cbb
		"transfer": w3.MustNewFunc("transfer(address,uint256)", "bool"),
		// Keccak hash -> 0x23b872dd
		"transferFrom": w3.MustNewFunc("transferFrom(address, address, uint256)", "bool"),
		// Add to account index
		"add": w3.MustNewFunc("add(address)", "bool"),
		// giveTo gas refill
		"giveTo": w3.MustNewFunc("giveTo(address)", "uint256"),
	}
}

// Bootstrap the internal custodial system configs and system signer key.
// This container is passed down to individual tasker and API handlers.
func initSystemContainer(ctx context.Context, noncestore nonce.Noncestore) *custodial.SystemContainer {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Some custodial system defaults loaded from the config file.
	systemContainer := &custodial.SystemContainer{
		Abis:                  initAbis(),
		AccountIndexContract:  w3.A(ko.MustString("system.account_index_address")),
		GasFaucetContract:     w3.A(ko.MustString("system.gas_faucet_address")),
		GasRefillThreshold:    big.NewInt(ko.MustInt64("system.gas_refill_threshold")),
		GasRefillValue:        big.NewInt(ko.MustInt64("system.gas_refill_value")),
		GiftableGasValue:      big.NewInt(ko.MustInt64("system.giftable_gas_value")),
		GiftableToken:         w3.A(ko.MustString("system.giftable_token_address")),
		GiftableTokenValue:    big.NewInt(ko.MustInt64("system.giftable_token_value")),
		LockTimeout:           1 * time.Second,
		PublicKey:             ko.MustString("system.public_key"),
		TokenDecimals:         ko.MustInt("system.token_decimals"),
		TokenTransferGasLimit: uint64(ko.MustInt64("system.token_transfer_gas_limit")),
	}

	// Check if system signer account nonce is present.
	// If not (first boot), we bootstrap it from the network.
	currentSystemNonce, err := noncestore.Peek(ctx, ko.MustString("system.public_key"))
	lo.Info("custodial: loaded system nonce from noncestore", "nonce", currentSystemNonce)
	if err == redis.Nil {
		nonce, err := noncestore.SyncNetworkNonce(ctx, ko.MustString("system.public_key"))
		lo.Info("custodial: syncing system nonce from network", "nonce", nonce)
		if err != nil {
			lo.Fatal("custodial: critical error bootstrapping system container", "error", err)
		}
	}

	loadedPrivateKey, err := eth_crypto.HexToECDSA(ko.MustString("system.private_key"))
	if err != nil {
		lo.Fatal("custodial: critical error bootstrapping system container", "error", err)
	}
	systemContainer.PrivateKey = loadedPrivateKey

	return systemContainer
}
