package keeper_test

import (
	"context"

	"github.com/aequitas/aura/chain/x/aura-bindings/keeper"
	"github.com/aequitas/aura/chain/x/aura-bindings/types"
)

func (suite *KeeperTestSuite) TestQueryServer() {
	queryServer := keeper.NewQueryServerImpl(suite.keeper)
	require := suite.Require()
	require.NotNil(queryServer)
}

func (suite *KeeperTestSuite) TestQueryStats() {
	queryServer := keeper.NewQueryServerImpl(suite.keeper)
	require := suite.Require()

	// Add some query stats
	suite.keeper.IncrementQueryStat("test_query_1")
	suite.keeper.IncrementQueryStat("test_query_1")
	suite.keeper.IncrementQueryStat("test_query_2")

	ctx := context.Background()
	req := &types.QueryStatsRequest{}

	resp, err := queryServer.QueryStats(ctx, req)
	require.NoError(err)
	require.NotNil(resp)
	require.Equal(uint64(2), resp.QueryStats["test_query_1"])
	require.Equal(uint64(1), resp.QueryStats["test_query_2"])
}

func (suite *KeeperTestSuite) TestQueryStatsNilRequest() {
	queryServer := keeper.NewQueryServerImpl(suite.keeper)
	require := suite.Require()

	ctx := context.Background()

	resp, err := queryServer.QueryStats(ctx, nil)
	require.Error(err)
	require.Nil(resp)
	require.Equal(types.ErrInvalidParam, err)
}

func (suite *KeeperTestSuite) TestMessageStats() {
	queryServer := keeper.NewQueryServerImpl(suite.keeper)
	require := suite.Require()

	// Add some message stats
	suite.keeper.IncrementMessageStat("test_msg_1")
	suite.keeper.IncrementMessageStat("test_msg_2")
	suite.keeper.IncrementMessageStat("test_msg_2")
	suite.keeper.IncrementMessageStat("test_msg_2")

	ctx := context.Background()
	req := &types.MessageStatsRequest{}

	resp, err := queryServer.MessageStats(ctx, req)
	require.NoError(err)
	require.NotNil(resp)
	require.Equal(uint64(1), resp.MessageStats["test_msg_1"])
	require.Equal(uint64(3), resp.MessageStats["test_msg_2"])
}

func (suite *KeeperTestSuite) TestMessageStatsNilRequest() {
	queryServer := keeper.NewQueryServerImpl(suite.keeper)
	require := suite.Require()

	ctx := context.Background()

	resp, err := queryServer.MessageStats(ctx, nil)
	require.Error(err)
	require.Nil(resp)
	require.Equal(types.ErrInvalidParam, err)
}

func (suite *KeeperTestSuite) TestAllStats() {
	queryServer := keeper.NewQueryServerImpl(suite.keeper)
	require := suite.Require()

	// Add some stats
	suite.keeper.IncrementQueryStat("query_1")
	suite.keeper.IncrementQueryStat("query_2")
	suite.keeper.IncrementMessageStat("msg_1")
	suite.keeper.IncrementMessageStat("msg_1")

	ctx := context.Background()
	req := &types.AllStatsRequest{}

	resp, err := queryServer.AllStats(ctx, req)
	require.NoError(err)
	require.NotNil(resp)
	require.Equal(uint64(1), resp.QueryStats["query_1"])
	require.Equal(uint64(1), resp.QueryStats["query_2"])
	require.Equal(uint64(2), resp.MessageStats["msg_1"])
}

func (suite *KeeperTestSuite) TestAllStatsNilRequest() {
	queryServer := keeper.NewQueryServerImpl(suite.keeper)
	require := suite.Require()

	ctx := context.Background()

	resp, err := queryServer.AllStats(ctx, nil)
	require.Error(err)
	require.Nil(resp)
	require.Equal(types.ErrInvalidParam, err)
}

func (suite *KeeperTestSuite) TestAllStatsEmpty() {
	queryServer := keeper.NewQueryServerImpl(suite.keeper)
	require := suite.Require()

	ctx := context.Background()
	req := &types.AllStatsRequest{}

	resp, err := queryServer.AllStats(ctx, req)
	require.NoError(err)
	require.NotNil(resp)
	require.NotNil(resp.QueryStats)
	require.NotNil(resp.MessageStats)
	require.Len(resp.QueryStats, 0)
	require.Len(resp.MessageStats, 0)
}

func (suite *KeeperTestSuite) TestQueryServerImplementsInterface() {
	queryServer := keeper.NewQueryServerImpl(suite.keeper)
	require := suite.Require()

	// Verify that queryServer implements the QueryServer interface
	var _ types.QueryServer = queryServer
	require.NotNil(queryServer)
}
