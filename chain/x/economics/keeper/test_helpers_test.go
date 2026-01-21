// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

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

	"github.com/aequitas/aura/chain/testing/testutil"
)

const (
	testStoreKey = "economics"
)

// setupKeeperForTest creates a keeper and context for testing
// This is an internal helper to avoid import cycles
func setupKeeperForTest(t *testing.T) (*Keeper, sdk.Context) {
	t.Helper()

	// Configure SDK with proper bech32 prefix for address validation (safe to call multiple times)
	testutil.EnsureSDKConfig()

	storeKey := storetypes.NewKVStoreKey(testStoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	logger := log.NewNopLogger()
	authority := "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr" // test governance address

	k := NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		storeKey,
		authority,
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, logger)

	// Initialize with minimal params (skip setting to avoid Coin serialization issues)
	// Tests can set params as needed
	// defaultParams := types.DefaultParams()
	// require.NoError(t, k.SetParams(ctx, defaultParams))

	return k, ctx
}
