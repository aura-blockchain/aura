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
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

// KeeperTestSuite is the test suite for keeper package tests
// This is separate from keeper_test package to allow internal package access
type KeeperTestSuite struct {
	suite.Suite

	SdkCtx sdk.Context
	Keeper Keeper
	cdc    codec.Codec
}

// SetupTest initializes the test suite
func (suite *KeeperTestSuite) SetupTest() {
	// Setup store
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	// Create database and commit multi-store
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	cms.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	suite.Require().NoError(cms.LoadLatestVersion())

	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create context with proper store
	header := tmproto.Header{
		Height: 1,
		Time:   time.Now(),
	}
	ctx := sdk.NewContext(cms, header, false, log.NewNopLogger())

	// Create keeper
	k := NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		"authority",
		nil, // staking keeper mock
		nil, // slashing keeper mock
		nil, // bank keeper mock
	)

	suite.SdkCtx = ctx
	suite.Keeper = k
	suite.cdc = cdc
}

// TearDownTest cleans up after each test
func (suite *KeeperTestSuite) TearDownTest() {
	// Cleanup if needed
}

// TestKeeperTestSuite runs the keeper test suite
func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// NewTestKeeper creates a new keeper for testing, returning the keeper and SDK context
func NewTestKeeper(t *testing.T) (Keeper, sdk.Context) {
	// Setup store
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	// Create database and commit multi-store
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	cms.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	if err := cms.LoadLatestVersion(); err != nil {
		t.Fatalf("failed to load latest version: %v", err)
	}

	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create context with proper store
	header := tmproto.Header{
		Height: 1,
		Time:   time.Now(),
	}
	ctx := sdk.NewContext(cms, header, false, log.NewNopLogger())

	// Create keeper
	k := NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		"authority",
		nil, // staking keeper mock
		nil, // slashing keeper mock
		nil, // bank keeper mock
	)

	return k, ctx
}
