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
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/prevalidation/types"
)

// KeeperTestSuite defines a test suite for the keeper
type KeeperTestSuite struct {
	suite.Suite

	Keeper *Keeper
	SdkCtx sdk.Context
	cdc    codec.BinaryCodec
}

// SetupTest sets up the test suite
func (suite *KeeperTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(types.ModuleName)
	testCtx := suite.defaultContextWithDB(key, storetypes.NewTransientStoreKey("transient_test"))

	suite.SdkCtx = testCtx.Ctx
	encCfg := MakeTestEncodingConfig()
	suite.cdc = encCfg.Codec

	suite.Keeper = NewKeeper(
		encCfg.Codec,
		key,
	)
	suite.Keeper.SetLogger(log.NewNopLogger())

	// Initialize default params
	err := suite.Keeper.SetParams(suite.SdkCtx, types.DefaultParams())
	suite.Require().NoError(err)
}

// defaultContextWithDB creates a default context with DB for testing
func (suite *KeeperTestSuite) defaultContextWithDB(key storetypes.StoreKey, tkey storetypes.StoreKey) testCtx {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	cms.MountStoreWithDB(tkey, storetypes.StoreTypeTransient, db)
	err := cms.LoadLatestVersion()
	suite.Require().NoError(err)

	ctx := sdk.NewContext(cms, cmtproto.Header{}, false, log.NewNopLogger())

	return testCtx{
		Ctx: ctx,
		DB:  db,
		CMS: cms,
	}
}

type testCtx struct {
	Ctx sdk.Context
	DB  *dbm.MemDB
	CMS storetypes.CommitMultiStore
}

// MakeTestEncodingConfig creates a test encoding config
func MakeTestEncodingConfig() EncodingConfig {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             cdc,
	}
}

// EncodingConfig defines the encoding configuration
type EncodingConfig struct {
	InterfaceRegistry codectypes.InterfaceRegistry
	Codec             codec.BinaryCodec
}

// NewTestKeeper creates a new keeper for testing
func NewTestKeeper(t *testing.T) Keeper {
	t.Helper()

	key := storetypes.NewKVStoreKey(types.ModuleName)
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	err := cms.LoadLatestVersion()
	if err != nil {
		t.Fatal(err)
	}

	encCfg := MakeTestEncodingConfig()

	keeper := NewKeeper(encCfg.Codec, key)
	keeper.SetLogger(log.NewNopLogger())

	return *keeper
}
