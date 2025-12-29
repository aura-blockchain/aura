//go:build integration

package main

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	auraclient "github.com/aura-chain/aura/sdk/go/client"
	"github.com/aura-chain/aura/sdk/go/helpers"
)

func getTestConfig() auraclient.Config {
	rpcEndpoint := os.Getenv("AURA_RPC_ENDPOINT")
	if rpcEndpoint == "" {
		rpcEndpoint = "http://localhost:27657"
	}

	grpcEndpoint := os.Getenv("AURA_GRPC_ENDPOINT")
	if grpcEndpoint == "" {
		grpcEndpoint = "localhost:10090"
	}

	chainID := os.Getenv("AURA_CHAIN_ID")
	if chainID == "" {
		chainID = "aura-local-4"
	}

	return auraclient.Config{
		RPCEndpoint:  rpcEndpoint,
		GRPCEndpoint: grpcEndpoint,
		ChainID:      chainID,
		Prefix:       "aura",
	}
}

func TestIntegration_ClientConnection(t *testing.T) {
	config := getTestConfig()
	client, err := auraclient.NewClient(config)
	require.NoError(t, err, "failed to create client")
	require.NotNil(t, client, "client is nil")
	defer client.Close()

	// Verify chain ID
	chainID := client.GetChainID()
	require.Equal(t, config.ChainID, chainID, "chain ID mismatch")
	t.Logf("Connected to chain: %s", chainID)
}

func TestIntegration_WalletGeneration(t *testing.T) {
	// Generate new mnemonic
	mnemonic, err := helpers.GenerateMnemonic()
	require.NoError(t, err, "failed to generate mnemonic")
	require.NotEmpty(t, mnemonic, "mnemonic is empty")

	// Count words
	words := strings.Fields(mnemonic)
	require.Equal(t, 24, len(words), "expected 24-word mnemonic")

	t.Logf("Generated valid %d-word mnemonic", len(words))
}

func TestIntegration_WalletImport(t *testing.T) {
	config := getTestConfig()
	client, err := auraclient.NewClient(config)
	require.NoError(t, err, "failed to create client")
	defer client.Close()

	// Generate and import wallet
	mnemonic, err := helpers.GenerateMnemonic()
	require.NoError(t, err, "failed to generate mnemonic")

	addr, err := client.ImportWalletFromMnemonic("integration-test-wallet", mnemonic, "")
	require.NoError(t, err, "failed to import wallet")
	require.NotNil(t, addr, "address is nil")

	// Verify address has correct prefix
	addrStr := addr.String()
	require.True(t, len(addrStr) > 4, "address too short")
	require.Equal(t, "aura", addrStr[:4], "address should start with 'aura'")

	t.Logf("Imported wallet with address: %s", addrStr)
}

func TestIntegration_QueryBalance(t *testing.T) {
	config := getTestConfig()
	client, err := auraclient.NewClient(config)
	require.NoError(t, err, "failed to create client")
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Query balance of a generated test address
	mnemonic, err := helpers.GenerateMnemonic()
	require.NoError(t, err)

	addr, err := client.ImportWalletFromMnemonic("balance-test", mnemonic, "")
	require.NoError(t, err)

	balance, err := client.GetBalance(ctx, addr.String(), "uaura")
	if err != nil {
		t.Logf("Balance query returned error (expected for unfunded account): %v", err)
	} else if balance != nil {
		t.Logf("Balance for %s: %s", addr.String(), balance.String())
	} else {
		t.Logf("Balance for %s: 0", addr.String())
	}
}

func TestIntegration_QueryAllBalances(t *testing.T) {
	config := getTestConfig()
	client, err := auraclient.NewClient(config)
	require.NoError(t, err, "failed to create client")
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Query all balances for a test address
	mnemonic, err := helpers.GenerateMnemonic()
	require.NoError(t, err)

	addr, err := client.ImportWalletFromMnemonic("all-balance-test", mnemonic, "")
	require.NoError(t, err)

	balances, err := client.GetAllBalances(ctx, addr.String())
	if err != nil {
		t.Logf("All balances query returned error: %v", err)
	} else {
		t.Logf("All balances for %s: %v", addr.String(), balances)
	}
}

func TestIntegration_QueryAccount(t *testing.T) {
	config := getTestConfig()
	client, err := auraclient.NewClient(config)
	require.NoError(t, err, "failed to create client")
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Generate test address
	mnemonic, err := helpers.GenerateMnemonic()
	require.NoError(t, err)

	addr, err := client.ImportWalletFromMnemonic("account-test", mnemonic, "")
	require.NoError(t, err)

	account, err := client.GetAccount(ctx, addr.String())
	if err != nil {
		t.Logf("Account query returned error (expected for non-existent account): %v", err)
	} else if account != nil {
		t.Logf("Account: %v", account.GetAddress())
	}
}
