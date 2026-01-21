// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dbm "github.com/cosmos/cosmos-db"

	"github.com/aequitas/aura/chain/config"
)

// EnsureSDKConfig configures the SDK Bech32 prefixes for AURA.
// This delegates to config.EnsureSDKConfig() to ensure a single initialization
// point across all tests, preventing "Config is sealed" panics.
func EnsureSDKConfig() {
	config.EnsureSDKConfig()
}

// TestContext provides a common test context for all modules
type TestContext struct {
	Ctx    context.Context
	SdkCtx sdk.Context
	DB     *dbm.MemDB
	CMS    store.CommitMultiStore
	Logger log.Logger
}

// SetupTestContext creates a new test context for unit tests
func SetupTestContext(t testing.TB) *TestContext {
	t.Helper()

	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), nil)

	header := tmproto.Header{
		Height: 1,
		Time:   time.Now(),
	}
	ctx := sdk.NewContext(cms, header, false, log.NewNopLogger())

	return &TestContext{
		Ctx:    ctx,
		SdkCtx: ctx,
		DB:     db,
		CMS:    cms,
		Logger: log.NewNopLogger(),
	}
}

// CreateTestCodec creates a test codec with common types registered
func CreateTestCodec() codec.Codec {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	return codec.NewProtoCodec(interfaceRegistry)
}

// MockTime returns a deterministic time for testing
func MockTime() time.Time {
	return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
}

// AssertNoError is a helper to assert no error occurred
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	require.NoError(t, err)
}

// AssertError is a helper to assert an error occurred
func AssertError(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
}

// CreateTestStore creates a KVStore for testing
func CreateTestStore(ctx sdk.Context, key storetypes.StoreKey) storetypes.KVStore {
	return ctx.KVStore(key)
}

// GenerateTestAddress generates a test SDK address
func GenerateTestAddress() sdk.AccAddress {
	return sdk.AccAddress([]byte("test_address_______"))
}

// GenerateTestAddresses generates multiple test addresses
func GenerateTestAddresses(count int) []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, count)
	for i := 0; i < count; i++ {
		addrs[i] = sdk.AccAddress([]byte("test_address_" + string(rune(i))))
	}
	return addrs
}

// SetupMultiStore creates a multi-store with the given keys
func SetupMultiStore(t testing.TB, keys ...*storetypes.KVStoreKey) store.CommitMultiStore {
	t.Helper()

	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), nil)

	for _, key := range keys {
		cms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	}

	err := cms.LoadLatestVersion()
	require.NoError(t, err)

	return cms
}
