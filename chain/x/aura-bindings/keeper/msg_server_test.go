package keeper_test

import (
	"github.com/aequitas/aura/chain/x/aura-bindings/keeper"
	"github.com/aequitas/aura/chain/x/aura-bindings/types"
)

func (suite *KeeperTestSuite) TestMsgServer() {
	msgServer := keeper.NewMsgServerImpl(suite.keeper)
	require := suite.Require()
	require.NotNil(msgServer)
}

func (suite *KeeperTestSuite) TestEmptyMethod() {
	msgServer := keeper.NewMsgServerImpl(suite.keeper)
	require := suite.Require()

	// Use the SDK context from the test suite (now implements context.Context directly)
	req := &types.EmptyRequest{}

	resp, err := msgServer.EmptyMethod(suite.ctx, req)
	require.NoError(err)
	require.NotNil(resp)
}

func (suite *KeeperTestSuite) TestMsgServerImplementsInterface() {
	msgServer := keeper.NewMsgServerImpl(suite.keeper)
	require := suite.Require()

	// Verify that msgServer implements the MsgServer interface
	var _ types.MsgServer = msgServer
	require.NotNil(msgServer)
}
