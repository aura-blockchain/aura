// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
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
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

// KeeperTestSuite is the test suite for keeper tests in the keeper package
type KeeperTestSuite struct {
	suite.Suite

	SdkCtx sdk.Context
	Keeper Keeper
	Cdc    codec.Codec
}

// SetupTest initializes the test suite
func (suite *KeeperTestSuite) SetupTest() {
	// Create store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(suite.T(), stateStore.LoadLatestVersion())

	// Create codec
	registry := codectypes.NewInterfaceRegistry()
	suite.Cdc = codec.NewProtoCodec(registry)

	// Create context
	suite.SdkCtx = sdk.NewContext(stateStore, cmtproto.Header{Time: time.Now()}, false, log.NewNopLogger())

	// Create keeper
	suite.Keeper = NewKeeper(
		suite.Cdc,
		runtime.NewKVStoreService(storeKey),
		"authority",
		log.NewNopLogger(),
	)

	// Initialize with default params
	require.NoError(suite.T(), suite.Keeper.SetParams(suite.SdkCtx, *types.DefaultParams()))
}

// NewTestKeeperWithContext creates a new test keeper and context
// This is a helper function for table-driven tests
func NewTestKeeperWithContext(t *testing.T) (Keeper, sdk.Context) {
	// Create store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	// Create codec
	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	// Create context
	ctx := sdk.NewContext(stateStore, cmtproto.Header{Time: time.Now()}, false, log.NewNopLogger())

	// Create keeper
	keeper := NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		"authority",
		log.NewNopLogger(),
	)

	// Initialize with default params
	require.NoError(t, keeper.SetParams(ctx, *types.DefaultParams()))

	return keeper, ctx
}
