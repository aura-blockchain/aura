// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"sync"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/params"
	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

const (
	testStoreKey  = "economicsecurity"
	testAuthority = "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr"
)

var (
	// Ensure SDK config is only set up once across all tests
	setupSDKConfigOnce sync.Once
)

// setupKeeperForTest creates a keeper and context for testing
// This is an internal helper to avoid import cycles
func setupKeeperForTest(t *testing.T) (*Keeper, sdk.Context) {
	t.Helper()

	// Configure SDK with proper bech32 prefix for address validation
	// This is required for invariant checks that validate addresses
	// Use sync.Once to ensure this only happens once across all tests
	setupSDKConfigOnce.Do(func() {
		config := sdk.GetConfig()
		config.SetBech32PrefixForAccount("aura", "aurapub")
		config.SetBech32PrefixForValidator("auravaloper", "auravaloper pub")
		config.SetBech32PrefixForConsensusNode("auravalcons", "auravalconspub")
		config.Seal()
	})

	storeKey := storetypes.NewKVStoreKey(testStoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	logger := log.NewNopLogger()
	authority := testAuthority // test governance address

	// Create params store with default params
	defaultParams := types.DefaultParams()
	k, ctx := buildKeeperWithParams(t, storeKey, cdc, authority, defaultParams, stateStore, logger)

	// Initialize params
	require.NoError(t, k.SetParams(*defaultParams))

	return k, ctx
}

// buildKeeperWithParams constructs a keeper using the provided params without validating them first.
func buildKeeperWithParams(
	t *testing.T,
	storeKey *storetypes.KVStoreKey,
	cdc codec.BinaryCodec,
	authority string,
	paramsVal *types.Params,
	stateStore store.CommitMultiStore,
	logger log.Logger,
) (*Keeper, sdk.Context) {
	t.Helper()

	if paramsVal == nil {
		paramsVal = types.DefaultParams()
	}

	paramsStore := params.NewStore(*paramsVal)

	k := NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		paramsStore,
		authority,
	)

	// Create context with proper block time for timestamp-dependent tests
	header := cmtproto.Header{
		Height: 1,
		Time:   time.Now(),
	}
	ctx := sdk.NewContext(stateStore, header, false, logger)
	return k, ctx
}

// setupKeeperWithCustomParams builds a keeper initialized with the supplied params.
// Params are not re-validated here so tests can construct deliberately invalid configurations.
func setupKeeperWithCustomParams(t *testing.T, paramsVal *types.Params) (*Keeper, sdk.Context) {
	t.Helper()

	setupSDKConfigOnce.Do(func() {
		config := sdk.GetConfig()
		config.SetBech32PrefixForAccount("aura", "aurapub")
		config.SetBech32PrefixForValidator("auravaloper", "auravaloper pub")
		config.SetBech32PrefixForConsensusNode("auravalcons", "auravalconspub")
		config.Seal()
	})

	storeKey := storetypes.NewKVStoreKey(testStoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	logger := log.NewNopLogger()

	return buildKeeperWithParams(t, storeKey, cdc, testAuthority, paramsVal, stateStore, logger)
}
