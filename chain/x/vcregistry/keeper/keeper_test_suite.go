// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
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

	"github.com/aequitas/aura/chain/x/vcregistry/params"
	"github.com/aequitas/aura/chain/x/vcregistry/types"
)

// KeeperTestSuite is a comprehensive test suite for the vcregistry keeper
// It provides a complete test environment with:
// - In-memory database
// - Initialized keeper with KV store
// - SDK context with block time and height
// - Mock confidence score keeper
type KeeperTestSuite struct {
	suite.Suite

	Keeper *Keeper
	SdkCtx sdk.Context
	Cdc    codec.BinaryCodec

	// Test fixtures
	testBlockTime   time.Time
	testBlockHeight int64
}

// SetupTest initializes the test suite before each test
// This creates a fresh keeper and context for each test
func (suite *KeeperTestSuite) SetupTest() {
	// Create in-memory database
	db := dbm.NewMemDB()

	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create multi-store
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	suite.Require().NoError(stateStore.LoadLatestVersion())

	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	suite.Cdc = codec.NewProtoCodec(interfaceRegistry)

	// Create keeper with default params
	suite.Keeper = NewKeeper(
		params.NewStore(*types.DefaultParams()),
		"cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn", // authority address
	).WithStore(storeKey, suite.Cdc)

	// Set up test block metadata
	suite.testBlockTime = time.Unix(1700000000, 0) // Fixed time for deterministic tests
	suite.testBlockHeight = 100

	// Create SDK context
	suite.SdkCtx = sdk.NewContext(
		stateStore,
		tmproto.Header{
			Height: suite.testBlockHeight,
			Time:   suite.testBlockTime,
		},
		false,
		log.NewNopLogger(),
	)

	// Set deterministic time and height in keeper for testing
	suite.Keeper.SetCurrentTime(suite.testBlockTime.Unix())
	suite.Keeper.SetCurrentHeight(uint64(suite.testBlockHeight))

	// Set up mock confidence score keeper
	suite.Keeper.SetConfidenceScoreKeeper(&mockConfidenceScoreKeeper{
		userScore: 1000,
	})
}

// TearDownTest cleans up after each test
func (suite *KeeperTestSuite) TearDownTest() {
	// Cleanup is handled automatically by garbage collection
}

// AdvanceBlock advances the block height and time
// Useful for testing time-dependent logic
func (suite *KeeperTestSuite) AdvanceBlock(blocks int64, duration time.Duration) {
	suite.testBlockHeight += blocks
	suite.testBlockTime = suite.testBlockTime.Add(duration)

	suite.SdkCtx = suite.SdkCtx.
		WithBlockHeight(suite.testBlockHeight).
		WithBlockTime(suite.testBlockTime)

	suite.Keeper.SetCurrentTime(suite.testBlockTime.Unix())
	suite.Keeper.SetCurrentHeight(uint64(suite.testBlockHeight))
}

// SetBlockTime sets a specific block time
// Useful for testing time-based expiration logic
func (suite *KeeperTestSuite) SetBlockTime(t time.Time) {
	suite.testBlockTime = t
	suite.SdkCtx = suite.SdkCtx.WithBlockTime(t)
	suite.Keeper.SetCurrentTime(t.Unix())
}

// SetBlockHeight sets a specific block height
func (suite *KeeperTestSuite) SetBlockHeight(height int64) {
	suite.testBlockHeight = height
	suite.SdkCtx = suite.SdkCtx.WithBlockHeight(height)
	suite.Keeper.SetCurrentHeight(uint64(height))
}

// GetBlockTime returns the current block time
func (suite *KeeperTestSuite) GetBlockTime() time.Time {
	return suite.testBlockTime
}

// GetBlockHeight returns the current block height
func (suite *KeeperTestSuite) GetBlockHeight() int64 {
	return suite.testBlockHeight
}

// mockConfidenceScoreKeeper is a mock implementation of ConfidenceScoreKeeper
// for testing without depending on the confidencescore module
type mockConfidenceScoreKeeper struct {
	userScore uint64
}

func (m *mockConfidenceScoreKeeper) GetUserScore(walletAddr string) (uint64, bool) {
	return m.userScore, true
}

func (m *mockConfidenceScoreKeeper) HasCompletedIR(walletAddr, irID string) bool {
	return true
}

func (m *mockConfidenceScoreKeeper) GetArenaScore(walletAddr, arena string) (uint64, error) {
	return 5000, nil
}

func (m *mockConfidenceScoreKeeper) GetAnchorInfo(walletAddr string) (interface{}, bool) {
	return struct{}{}, true
}

func (m *mockConfidenceScoreKeeper) IsVerified(walletAddr string) bool {
	return true
}

// SetUserScore updates the mock user score
func (suite *KeeperTestSuite) SetUserScore(score uint64) {
	suite.Keeper.SetConfidenceScoreKeeper(&mockConfidenceScoreKeeper{
		userScore: score,
	})
}
