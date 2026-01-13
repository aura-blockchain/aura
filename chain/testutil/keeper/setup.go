// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

// TestApp represents a minimal test application structure
type TestApp struct {
	Cdc           codec.Codec
	LegacyAmino   *codec.LegacyAmino
	InterfaceRegistry codectypes.InterfaceRegistry
	StoreKey      storetypes.StoreKey
	MemStoreKey   storetypes.StoreKey
	StateStore    store.CommitMultiStore
}

// SetupTestApp creates a fully initialized test app for testing
func SetupTestApp(t *testing.T) (*TestApp, sdk.Context) {
	t.Helper()

	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	authtypes.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)
	stakingtypes.RegisterInterfaces(interfaceRegistry)

	cdc := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()

	// Create store keys
	storeKey := storetypes.NewKVStoreKey("test")
	memStoreKey := storetypes.NewMemoryStoreKey("test_mem")

	// Create state store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	// Create context
	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		ChainID: "aura-mvp-1",
		Height:  1,
		Time:    time.Now().UTC(),
	}, false, log.NewNopLogger())

	app := &TestApp{
		Cdc:               cdc,
		LegacyAmino:       legacyAmino,
		InterfaceRegistry: interfaceRegistry,
		StoreKey:          storeKey,
		MemStoreKey:       memStoreKey,
		StateStore:        stateStore,
	}

	return app, ctx
}

// SetupTestAppWithValidators creates a test app with validators initialized
func SetupTestAppWithValidators(t *testing.T, numValidators int) (*TestApp, sdk.Context, []stakingtypes.Validator) {
	t.Helper()

	app, ctx := SetupTestApp(t)
	validators := CreateTestValidators(t, numValidators)

	return app, ctx, validators
}

// CreateTestValidators creates test validator set
func CreateTestValidators(t *testing.T, count int) []stakingtypes.Validator {
	t.Helper()

	validators := make([]stakingtypes.Validator, count)
	for i := 0; i < count; i++ {
		privKey := ed25519.GenPrivKey()
		pubKey := privKey.PubKey()
		valAddr := sdk.ValAddress(pubKey.Address())

		validator := stakingtypes.Validator{
			OperatorAddress: valAddr.String(),
			ConsensusPubkey: nil, // Will be set if needed
			Jailed:          false,
			Status:          stakingtypes.Bonded,
			Tokens:          math.NewInt(1000000),
			DelegatorShares: math.LegacyNewDec(1000000),
			Description: stakingtypes.Description{
				Moniker: "test-validator-" + string(rune(i)),
			},
			Commission: stakingtypes.Commission{
				CommissionRates: stakingtypes.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(1, 2), // 1%
					MaxRate:       math.LegacyNewDecWithPrec(20, 2), // 20%
					MaxChangeRate: math.LegacyNewDecWithPrec(1, 2), // 1%
				},
			},
		}

		// Set consensus pubkey
		pkAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)
		validator.ConsensusPubkey = pkAny

		validators[i] = validator
	}

	return validators
}

// CreateTestAccounts creates funded test accounts
func CreateTestAccounts(t *testing.T, count int) []sdk.AccAddress {
	t.Helper()

	accounts := make([]sdk.AccAddress, count)
	for i := 0; i < count; i++ {
		privKey := secp256k1.GenPrivKey()
		addr := sdk.AccAddress(privKey.PubKey().Address())
		accounts[i] = addr
	}

	return accounts
}

// CreateTestAccountsWithBalances creates accounts and returns them with initial balances
func CreateTestAccountsWithBalances(t *testing.T, count int, initialBalance sdk.Coins) ([]sdk.AccAddress, map[string]sdk.Coins) {
	t.Helper()

	accounts := CreateTestAccounts(t, count)
	balances := make(map[string]sdk.Coins)

	for _, addr := range accounts {
		balances[addr.String()] = initialBalance
	}

	return accounts, balances
}

// GenTestAddress generates a random test address
func GenTestAddress() sdk.AccAddress {
	privKey := secp256k1.GenPrivKey()
	return sdk.AccAddress(privKey.PubKey().Address())
}

// GenTestValidatorAddress generates a random validator address
func GenTestValidatorAddress() sdk.ValAddress {
	privKey := ed25519.GenPrivKey()
	return sdk.ValAddress(privKey.PubKey().Address())
}

// GenTestPubKey generates a random public key
func GenTestPubKey() cryptotypes.PubKey {
	privKey := secp256k1.GenPrivKey()
	return privKey.PubKey()
}

// MakeEncodingConfig creates a test encoding config
func MakeEncodingConfig() EncodingConfig {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	authtypes.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)
	stakingtypes.RegisterInterfaces(interfaceRegistry)

	marshaler := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             marshaler,
		TxConfig:          nil, // Set if needed
		Amino:             legacyAmino,
	}
}

// EncodingConfig specifies the concrete encoding types to use for a given app.
type EncodingConfig struct {
	InterfaceRegistry codectypes.InterfaceRegistry
	Codec             codec.Codec
	TxConfig          interface{} // Use client.TxConfig if importing client
	Amino             *codec.LegacyAmino
}

// NewTestContext creates a new SDK context for testing
func NewTestContext(t *testing.T) sdk.Context {
	t.Helper()

	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())

	ctx := sdk.NewContext(cms, cmtproto.Header{
		ChainID: "aura-mvp-1",
		Height:  1,
		Time:    time.Now().UTC(),
	}, false, log.NewNopLogger())

	return ctx
}

// NewTestContextWithHeight creates a context at a specific height
func NewTestContextWithHeight(t *testing.T, height int64) sdk.Context {
	t.Helper()

	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())

	ctx := sdk.NewContext(cms, cmtproto.Header{
		ChainID: "aura-mvp-1",
		Height:  height,
		Time:    time.Now().UTC(),
	}, false, log.NewNopLogger())

	return ctx
}

// CreateTestCoins creates a coin set for testing
func CreateTestCoins(amount int64, denom string) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(denom, math.NewInt(amount)))
}

// CreateMultipleTestCoins creates multiple coins for testing
func CreateMultipleTestCoins(amounts map[string]int64) sdk.Coins {
	coins := sdk.NewCoins()
	for denom, amount := range amounts {
		coins = coins.Add(sdk.NewCoin(denom, math.NewInt(amount)))
	}
	return coins
}
