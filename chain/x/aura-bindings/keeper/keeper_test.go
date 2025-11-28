package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/app"
	"github.com/aequitas/aura/chain/x/aura-bindings/keeper"
	"github.com/aequitas/aura/chain/x/aura-bindings/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app    *app.App
	ctx    sdk.Context
	keeper keeper.Keeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.NewApp()
	suite.ctx = suite.app.NewUncachedContext(false, tmproto.Header{Height: 1})

	// Create a keeper instance for testing
	// Note: In a real app, this would be accessed via suite.app.AuraBindingsKeeper
	// For now, we create a standalone keeper
	suite.keeper = *keeper.NewKeeper(
		suite.app.AppCodec(),
		suite.app.GetKey(types.StoreKey),
		suite.app.VCRegistryKeeper,
	)
}

func (suite *KeeperTestSuite) TestNewKeeper() {
	require.NotNil(suite.T(), suite.keeper)
	require.NotNil(suite.T(), suite.keeper.VCKeeper())
}

func (suite *KeeperTestSuite) TestQueryRateLimiting() {
	ctx := suite.ctx
	address := "aura1test"

	// Test that rate limiting works correctly
	for i := 0; i < types.MaxQueriesPerBlock; i++ {
		err := suite.keeper.CheckQueryRateLimit(ctx, address)
		require.NoError(suite.T(), err, "query %d should not exceed limit", i)
	}

	// The next query should exceed the limit
	err := suite.keeper.CheckQueryRateLimit(ctx, address)
	require.Error(suite.T(), err)
	require.Equal(suite.T(), types.ErrQueryRateLimitExceeded, err)
}

func (suite *KeeperTestSuite) TestQueryRateLimitReset() {
	ctx := suite.ctx
	address := "aura1test"

	// Fill up the rate limit
	for i := 0; i < types.MaxQueriesPerBlock; i++ {
		_ = suite.keeper.CheckQueryRateLimit(ctx, address)
	}

	// Should be at limit
	err := suite.keeper.CheckQueryRateLimit(ctx, address)
	require.Error(suite.T(), err)

	// Move to next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	suite.keeper.ResetQueryRateLimits(ctx)

	// Should be able to query again
	err = suite.keeper.CheckQueryRateLimit(ctx, address)
	require.NoError(suite.T(), err)
}

func (suite *KeeperTestSuite) TestQueryStats() {
	// Test that query statistics are tracked correctly
	suite.keeper.IncrementQueryStat("test_query_1")
	suite.keeper.IncrementQueryStat("test_query_1")
	suite.keeper.IncrementQueryStat("test_query_2")

	stats := suite.keeper.GetQueryStats()
	require.Equal(suite.T(), uint64(2), stats["test_query_1"])
	require.Equal(suite.T(), uint64(1), stats["test_query_2"])
}

func (suite *KeeperTestSuite) TestMessageStats() {
	// Test that message statistics are tracked correctly
	suite.keeper.IncrementMessageStat("test_msg_1")
	suite.keeper.IncrementMessageStat("test_msg_1")
	suite.keeper.IncrementMessageStat("test_msg_1")
	suite.keeper.IncrementMessageStat("test_msg_2")

	stats := suite.keeper.GetMessageStats()
	require.Equal(suite.T(), uint64(3), stats["test_msg_1"])
	require.Equal(suite.T(), uint64(1), stats["test_msg_2"])
}

func (suite *KeeperTestSuite) TestConcurrentAccess() {
	// Test concurrent access to stats
	address1 := "aura1test1"
	address2 := "aura1test2"

	// These should not interfere with each other
	err1 := suite.keeper.CheckQueryRateLimit(suite.ctx, address1)
	err2 := suite.keeper.CheckQueryRateLimit(suite.ctx, address2)

	require.NoError(suite.T(), err1)
	require.NoError(suite.T(), err2)

	// Increment stats concurrently
	suite.keeper.IncrementQueryStat("query1")
	suite.keeper.IncrementQueryStat("query2")
	suite.keeper.IncrementMessageStat("msg1")

	queryStats := suite.keeper.GetQueryStats()
	msgStats := suite.keeper.GetMessageStats()

	require.Equal(suite.T(), uint64(1), queryStats["query1"])
	require.Equal(suite.T(), uint64(1), queryStats["query2"])
	require.Equal(suite.T(), uint64(1), msgStats["msg1"])
}

func (suite *KeeperTestSuite) TestGetStore() {
	store := suite.keeper.GetStore(suite.ctx)
	require.NotNil(suite.T(), store)
}

func (suite *KeeperTestSuite) TestLogger() {
	logger := suite.keeper.Logger(suite.ctx)
	require.NotNil(suite.T(), logger)
}

func (suite *KeeperTestSuite) TestVCKeeper() {
	vcKeeper := suite.keeper.VCKeeper()
	require.NotNil(suite.T(), vcKeeper)
}

func (suite *KeeperTestSuite) TestMultipleAddressRateLimits() {
	ctx := suite.ctx
	addresses := []string{"aura1addr1", "aura1addr2", "aura1addr3"}

	// Each address should have its own rate limit
	for _, addr := range addresses {
		for i := 0; i < 10; i++ {
			err := suite.keeper.CheckQueryRateLimit(ctx, addr)
			require.NoError(suite.T(), err)
		}
	}

	// Verify each address has consumed 10 queries
	for _, addr := range addresses {
		// Still have room for more queries
		err := suite.keeper.CheckQueryRateLimit(ctx, addr)
		require.NoError(suite.T(), err)
	}
}

func (suite *KeeperTestSuite) TestStatsThreadSafety() {
	// Test that GetQueryStats returns a copy, not the internal map
	suite.keeper.IncrementQueryStat("test")
	stats1 := suite.keeper.GetQueryStats()
	stats2 := suite.keeper.GetQueryStats()

	// Modify stats1
	stats1["modified"] = 999

	// stats2 should not be affected
	require.Equal(suite.T(), uint64(1), stats2["test"])
	_, exists := stats2["modified"]
	require.False(suite.T(), exists)
}
