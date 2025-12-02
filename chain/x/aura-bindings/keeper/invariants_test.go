package keeper_test

import (
	"strings"

	"github.com/aequitas/aura/chain/x/aura-bindings/keeper"
	"github.com/aequitas/aura/chain/x/aura-bindings/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestQueryStatsNonNegativeInvariant() {
	require := suite.Require()

	// Add valid query stats
	suite.keeper.IncrementQueryStat("valid_query")

	// Run invariant
	invariant := keeper.QueryStatsNonNegativeInvariant(suite.keeper)
	msg, broken := invariant(suite.ctx)

	require.False(broken, "invariant should not be broken with valid stats")
	require.NotEmpty(msg)
}

func (suite *KeeperTestSuite) TestMessageStatsNonNegativeInvariant() {
	require := suite.Require()

	// Add valid message stats
	suite.keeper.IncrementMessageStat("valid_msg")

	// Run invariant
	invariant := keeper.MessageStatsNonNegativeInvariant(suite.keeper)
	msg, broken := invariant(suite.ctx)

	require.False(broken, "invariant should not be broken with valid stats")
	require.NotEmpty(msg)
}

func (suite *KeeperTestSuite) TestRateLimitsValidInvariant() {
	require := suite.Require()

	// Add some rate limits
	address := "aura1test"
	err := suite.keeper.CheckQueryRateLimit(suite.ctx, address)
	require.NoError(err)

	// Run invariant
	invariant := keeper.RateLimitsValidInvariant(suite.keeper)
	msg, broken := invariant(suite.ctx)

	require.False(broken, "invariant should not be broken with valid rate limits")
	require.NotEmpty(msg)
}

func (suite *KeeperTestSuite) TestRateLimitsInvariantMaxQueries() {
	require := suite.Require()

	// Fill up to max queries (but not over)
	address := "aura1test"
	for i := 0; i < types.MaxQueriesPerBlock; i++ {
		err := suite.keeper.CheckQueryRateLimit(suite.ctx, address)
		require.NoError(err)
	}

	// Run invariant - should still be valid
	invariant := keeper.RateLimitsValidInvariant(suite.keeper)
	msg, broken := invariant(suite.ctx)

	require.False(broken, "invariant should not be broken at max queries")
	require.NotEmpty(msg)
}

func (suite *KeeperTestSuite) TestStateConsistencyInvariant() {
	require := suite.Require()

	// Keeper should be initialized with consistent state
	invariant := keeper.StateConsistencyInvariant(suite.keeper)
	msg, broken := invariant(suite.ctx)

	require.False(broken, "invariant should not be broken with consistent state")
	require.NotEmpty(msg)
}

func (suite *KeeperTestSuite) TestAllInvariants() {
	require := suite.Require()

	// Add some data
	suite.keeper.IncrementQueryStat("query")
	suite.keeper.IncrementMessageStat("msg")
	_ = suite.keeper.CheckQueryRateLimit(suite.ctx, "aura1test")

	// Run all invariants
	invariant := keeper.AllInvariants(suite.keeper)
	msg, broken := invariant(suite.ctx)

	require.False(broken, "all invariants should pass")
	require.NotEmpty(msg)
}

func (suite *KeeperTestSuite) TestInvariantsWithEmptyState() {
	require := suite.Require()

	// Test all invariants with empty state
	invariants := []func(keeper.Keeper) sdk.Invariant{
		keeper.QueryStatsNonNegativeInvariant,
		keeper.MessageStatsNonNegativeInvariant,
		keeper.RateLimitsValidInvariant,
		keeper.StateConsistencyInvariant,
	}

	for _, inv := range invariants {
		invariant := inv(suite.keeper)
		msg, broken := invariant(suite.ctx)
		require.False(broken, "invariant should not be broken with empty state")
		require.NotEmpty(msg)
	}
}

func (suite *KeeperTestSuite) TestInvariantsWithMaxData() {
	require := suite.Require()

	// Add maximum allowed data
	for i := 0; i < 100; i++ {
		suite.keeper.IncrementQueryStat(string(rune('a' + i)))
		suite.keeper.IncrementMessageStat(string(rune('a' + i)))
	}

	// Add rate limits for multiple addresses
	for i := 0; i < 10; i++ {
		address := "aura1addr" + string(rune('0'+i))
		for j := 0; j < 10; j++ {
			_ = suite.keeper.CheckQueryRateLimit(suite.ctx, address)
		}
	}

	// All invariants should still pass
	invariant := keeper.AllInvariants(suite.keeper)
	msg, broken := invariant(suite.ctx)

	require.False(broken, "invariants should pass with max data")
	require.NotEmpty(msg)
}

func (suite *KeeperTestSuite) TestInvariantsAfterGenesisImport() {
	require := suite.Require()

	// Import genesis with data
	genesisState := types.GenesisState{
		QueryStats: map[string]uint64{
			"genesis_query": 100,
		},
		MessageStats: map[string]uint64{
			"genesis_msg": 50,
		},
	}

	err := suite.keeper.InitGenesis(suite.ctx, genesisState)
	require.NoError(err)

	// All invariants should pass after genesis import
	invariant := keeper.AllInvariants(suite.keeper)
	msg, broken := invariant(suite.ctx)

	require.False(broken, "invariants should pass after genesis import")
	require.NotEmpty(msg)
}

func (suite *KeeperTestSuite) TestInvariantMessages() {
	require := suite.Require()

	// Each invariant should produce a properly formatted message
	invariants := []struct {
		name      string
		invariant func(keeper.Keeper) sdk.Invariant
	}{
		{"query-stats-non-negative", keeper.QueryStatsNonNegativeInvariant},
		{"message-stats-non-negative", keeper.MessageStatsNonNegativeInvariant},
		{"rate-limits-valid", keeper.RateLimitsValidInvariant},
		{"state-consistency", keeper.StateConsistencyInvariant},
	}

	for _, tc := range invariants {
		invariant := tc.invariant(suite.keeper)
		msg, _ := invariant(suite.ctx)

		require.NotEmpty(msg, "invariant %s should produce a message", tc.name)
		require.True(strings.Contains(msg, tc.name), "message should contain invariant name")
		require.True(strings.Contains(msg, types.ModuleName), "message should contain module name")
	}
}
