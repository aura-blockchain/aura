// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

type QueryServerUncoveredTestSuite struct {
	KeeperTestSuite
	queryServer bridgepb.QueryServer
}

func TestQueryServerUncoveredTestSuite(t *testing.T) {
	suite.Run(t, new(QueryServerUncoveredTestSuite))
}

func (suite *QueryServerUncoveredTestSuite) SetupTest() {
	suite.KeeperTestSuite.SetupTest()
	suite.queryServer = NewQueryServerImpl(suite.Keeper)
}

// =============================================================================
// Params Query Tests
// =============================================================================

func (suite *QueryServerUncoveredTestSuite) TestQueryParams_NilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Params accepts nil request
	resp, err := suite.queryServer.Params(ctx, nil)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.NotNil(resp.Params)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryParams_Valid() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryParamsRequest{}
	resp, err := suite.queryServer.Params(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.NotNil(resp.Params)
}

// =============================================================================
// UserTransfers Query Tests
// =============================================================================

func (suite *QueryServerUncoveredTestSuite) TestQueryUserTransfers_NilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	resp, err := suite.queryServer.UserTransfers(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Nil(resp)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryUserTransfers_EmptyAddress() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryUserTransfersRequest{
		Address: "",
	}
	resp, err := suite.queryServer.UserTransfers(ctx, req)
	suite.Error(err, "should reject empty address")
	suite.Nil(resp)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryUserTransfers_ValidNoTransfers() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryUserTransfersRequest{
		Address: "aura1user123",
	}
	resp, err := suite.queryServer.UserTransfers(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	// Empty result is valid
}

// =============================================================================
// ChainConfig Query Tests
// =============================================================================

func (suite *QueryServerUncoveredTestSuite) TestQueryChainConfig_NilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	resp, err := suite.queryServer.ChainConfig(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Nil(resp)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryChainConfig_EmptyChainId() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryChainConfigRequest{
		ChainId: "",
	}
	resp, err := suite.queryServer.ChainConfig(ctx, req)
	suite.Error(err, "should reject empty chain ID")
	suite.Nil(resp)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryChainConfig_NotFound() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryChainConfigRequest{
		ChainId: "nonexistent-chain",
	}
	resp, err := suite.queryServer.ChainConfig(ctx, req)
	suite.Error(err, "should return not found")
	suite.Nil(resp)
}

// =============================================================================
// AllChains Query Tests
// =============================================================================

func (suite *QueryServerUncoveredTestSuite) TestQueryAllChains_NilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	resp, err := suite.queryServer.AllChains(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Nil(resp)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryAllChains_Valid() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryAllChainsRequest{}
	resp, err := suite.queryServer.AllChains(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
}

// =============================================================================
// AllWrappedTokens Query Tests
// =============================================================================

func (suite *QueryServerUncoveredTestSuite) TestQueryAllWrappedTokens_NilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	resp, err := suite.queryServer.AllWrappedTokens(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Nil(resp)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryAllWrappedTokens_Valid() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryAllWrappedTokensRequest{}
	resp, err := suite.queryServer.AllWrappedTokens(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
}

// =============================================================================
// SharedIdentity Query Tests
// =============================================================================

func (suite *QueryServerUncoveredTestSuite) TestQuerySharedIdentity_NilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	resp, err := suite.queryServer.SharedIdentity(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Nil(resp)
}

func (suite *QueryServerUncoveredTestSuite) TestQuerySharedIdentity_EmptyAddress() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QuerySharedIdentityRequest{
		Address: "",
	}
	resp, err := suite.queryServer.SharedIdentity(ctx, req)
	suite.Error(err, "should reject empty address")
	suite.Nil(resp)
}

func (suite *QueryServerUncoveredTestSuite) TestQuerySharedIdentity_NotFound() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QuerySharedIdentityRequest{
		Address: "aura1unknown",
	}
	resp, err := suite.queryServer.SharedIdentity(ctx, req)
	suite.Error(err, "should return not found")
	suite.Nil(resp)
}

// =============================================================================
// CrossChainSwap Query Tests
// =============================================================================

func (suite *QueryServerUncoveredTestSuite) TestQueryCrossChainSwap_NilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	resp, err := suite.queryServer.CrossChainSwap(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Nil(resp)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryCrossChainSwap_EmptySwapId() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryCrossChainSwapRequest{
		SwapId: "",
	}
	resp, err := suite.queryServer.CrossChainSwap(ctx, req)
	suite.Error(err, "should reject empty swap ID")
	suite.Nil(resp)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryCrossChainSwap_NotFound() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryCrossChainSwapRequest{
		SwapId: "nonexistent-swap",
	}
	resp, err := suite.queryServer.CrossChainSwap(ctx, req)
	suite.Error(err, "should return not found")
	suite.Nil(resp)
}

// =============================================================================
// BridgeStats Query Tests
// =============================================================================

func (suite *QueryServerUncoveredTestSuite) TestQueryBridgeStats_NilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	resp, err := suite.queryServer.BridgeStats(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Nil(resp)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryBridgeStats_Valid() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryBridgeStatsRequest{}
	resp, err := suite.queryServer.BridgeStats(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
}

// =============================================================================
// RelayerStats Query Tests
// =============================================================================

func (suite *QueryServerUncoveredTestSuite) TestQueryRelayerStats_NilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	resp, err := suite.queryServer.RelayerStats(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Nil(resp)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryRelayerStats_EmptyAddress() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryRelayerStatsRequest{
		RelayerAddress: "",
	}
	resp, err := suite.queryServer.RelayerStats(ctx, req)
	suite.Error(err, "should reject empty address")
	suite.Nil(resp)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryRelayerStats_NotFound() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryRelayerStatsRequest{
		RelayerAddress: "aura1relayer123",
	}
	resp, err := suite.queryServer.RelayerStats(ctx, req)
	suite.Error(err, "should return not found")
	suite.Nil(resp)
}
