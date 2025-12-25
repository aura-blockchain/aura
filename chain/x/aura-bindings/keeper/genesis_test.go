// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"github.com/aequitas/aura/chain/x/aura-bindings/types"
)

func (suite *KeeperTestSuite) TestInitGenesis() {
	require := suite.Require()

	// Create genesis state with stats
	genesisState := types.GenesisState{
		QueryStats: map[string]uint64{
			"query1": 10,
			"query2": 20,
		},
		MessageStats: map[string]uint64{
			"msg1": 5,
			"msg2": 15,
		},
	}

	// Initialize genesis
	err := suite.keeper.InitGenesis(suite.ctx, genesisState)
	require.NoError(err)

	// Verify stats were loaded
	queryStats := suite.keeper.GetQueryStats()
	require.Equal(uint64(10), queryStats["query1"])
	require.Equal(uint64(20), queryStats["query2"])

	msgStats := suite.keeper.GetMessageStats()
	require.Equal(uint64(5), msgStats["msg1"])
	require.Equal(uint64(15), msgStats["msg2"])
}

func (suite *KeeperTestSuite) TestInitGenesisEmpty() {
	require := suite.Require()

	// Create empty genesis state
	genesisState := types.DefaultGenesisState()

	// Initialize genesis
	err := suite.keeper.InitGenesis(suite.ctx, *genesisState)
	require.NoError(err)

	// Verify empty stats
	queryStats := suite.keeper.GetQueryStats()
	require.Len(queryStats, 0)

	msgStats := suite.keeper.GetMessageStats()
	require.Len(msgStats, 0)
}

func (suite *KeeperTestSuite) TestInitGenesisInvalid() {
	require := suite.Require()

	// Create invalid genesis state with nil maps
	genesisState := types.GenesisState{
		QueryStats:   nil,
		MessageStats: nil,
	}

	// Initialize genesis should fail validation
	err := suite.keeper.InitGenesis(suite.ctx, genesisState)
	require.Error(err)
}

func (suite *KeeperTestSuite) TestExportGenesis() {
	require := suite.Require()

	// Add some stats
	suite.keeper.IncrementQueryStat("export_query_1")
	suite.keeper.IncrementQueryStat("export_query_1")
	suite.keeper.IncrementQueryStat("export_query_2")
	suite.keeper.IncrementMessageStat("export_msg_1")

	// Export genesis
	exported := suite.keeper.ExportGenesis(suite.ctx)

	// Verify exported data
	require.Equal(uint64(2), exported.QueryStats["export_query_1"])
	require.Equal(uint64(1), exported.QueryStats["export_query_2"])
	require.Equal(uint64(1), exported.MessageStats["export_msg_1"])
}

func (suite *KeeperTestSuite) TestGenesisRoundTrip() {
	require := suite.Require()

	// Create genesis state with stats
	originalGenesis := types.GenesisState{
		QueryStats: map[string]uint64{
			"roundtrip_query": 100,
		},
		MessageStats: map[string]uint64{
			"roundtrip_msg": 50,
		},
	}

	// Initialize genesis
	err := suite.keeper.InitGenesis(suite.ctx, originalGenesis)
	require.NoError(err)

	// Export genesis
	exported := suite.keeper.ExportGenesis(suite.ctx)

	// Verify round-trip
	require.Equal(originalGenesis.QueryStats["roundtrip_query"], exported.QueryStats["roundtrip_query"])
	require.Equal(originalGenesis.MessageStats["roundtrip_msg"], exported.MessageStats["roundtrip_msg"])
}

func (suite *KeeperTestSuite) TestInitGenesisResetsRateLimits() {
	require := suite.Require()

	// Fill up rate limits
	address := "aura1test"
	for i := 0; i < types.MaxQueriesPerBlock; i++ {
		_ = suite.keeper.CheckQueryRateLimit(suite.ctx, address)
	}

	// Rate limit should be maxed out
	err := suite.keeper.CheckQueryRateLimit(suite.ctx, address)
	require.Error(err)

	// Initialize genesis
	genesisState := types.DefaultGenesisState()
	err = suite.keeper.InitGenesis(suite.ctx, *genesisState)
	require.NoError(err)

	// Rate limits should be reset
	err = suite.keeper.CheckQueryRateLimit(suite.ctx, address)
	require.NoError(err)
}

func (suite *KeeperTestSuite) TestExportGenesisAfterModifications() {
	require := suite.Require()

	// Start with some genesis data
	initialGenesis := types.GenesisState{
		QueryStats: map[string]uint64{
			"initial_query": 5,
		},
		MessageStats: map[string]uint64{
			"initial_msg": 3,
		},
	}

	err := suite.keeper.InitGenesis(suite.ctx, initialGenesis)
	require.NoError(err)

	// Modify the stats
	suite.keeper.IncrementQueryStat("initial_query")
	suite.keeper.IncrementQueryStat("new_query")
	suite.keeper.IncrementMessageStat("initial_msg")

	// Export and verify modifications
	exported := suite.keeper.ExportGenesis(suite.ctx)
	require.Equal(uint64(6), exported.QueryStats["initial_query"])
	require.Equal(uint64(1), exported.QueryStats["new_query"])
	require.Equal(uint64(4), exported.MessageStats["initial_msg"])
}
