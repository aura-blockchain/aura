// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/bridge/types"
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

// =============================================================================
// Validators Query Tests
// =============================================================================

func (suite *QueryServerUncoveredTestSuite) TestQueryValidators_NilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	resp, err := suite.queryServer.Validators(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Nil(resp)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryValidators_EmptyStore() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryValidatorsRequest{}
	resp, err := suite.queryServer.Validators(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	// Empty validators list is valid
}

func (suite *QueryServerUncoveredTestSuite) TestQueryValidators_WithChainConfigs() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Add chain config with validators
	config := types.ChainConfig{
		ChainId:          "ethereum",
		ChainName:        "Ethereum",
		AddressPrefix:    "0x",
		MinConfirmations: 12,
		Enabled:          true,
		Validators:       []string{"val1", "val2", "val3"},
	}
	suite.Keeper.setChainConfig(suite.SdkCtx, config)

	req := &bridgepb.QueryValidatorsRequest{}
	resp, err := suite.queryServer.Validators(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.GreaterOrEqual(len(resp.Validators), 3)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryValidators_WithStoredValidators() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Add stored validators
	for i := 0; i < 5; i++ {
		validator := &bridgepb.BridgeValidator{
			Address: "aura1validator" + string(rune('a'+i)),
			Power:   uint64(100 + i),
			Active:  true,
			Chains:  []string{"ethereum"},
		}
		suite.Keeper.setValidator(suite.SdkCtx, validator)
	}

	req := &bridgepb.QueryValidatorsRequest{}
	resp, err := suite.queryServer.Validators(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.GreaterOrEqual(len(resp.Validators), 5)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryValidators_WithMultipleChains() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Add chain configs
	ethConfig := types.ChainConfig{
		ChainId:          "ethereum",
		ChainName:        "Ethereum",
		AddressPrefix:    "0x",
		MinConfirmations: 12,
		Enabled:          true,
		Validators:       []string{"eth-val1", "eth-val2"},
	}
	suite.Keeper.setChainConfig(suite.SdkCtx, ethConfig)

	polyConfig := types.ChainConfig{
		ChainId:          "polygon",
		ChainName:        "Polygon",
		AddressPrefix:    "0x",
		MinConfirmations: 256,
		Enabled:          true,
		Validators:       []string{"poly-val1"},
	}
	suite.Keeper.setChainConfig(suite.SdkCtx, polyConfig)

	req := &bridgepb.QueryValidatorsRequest{}
	resp, err := suite.queryServer.Validators(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	// Should have validators from both chains
	suite.GreaterOrEqual(len(resp.Validators), 3)
}

// =============================================================================
// Transfer Query Tests
// =============================================================================

func (suite *QueryServerUncoveredTestSuite) TestQueryTransfer_NilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	resp, err := suite.queryServer.Transfer(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Nil(resp)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryTransfer_EmptyId() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryTransferRequest{
		TransferId: "",
	}
	resp, err := suite.queryServer.Transfer(ctx, req)
	suite.Error(err, "should reject empty transfer ID")
	suite.Nil(resp)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryTransfer_NotFound() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryTransferRequest{
		TransferId: "nonexistent-transfer",
	}
	resp, err := suite.queryServer.Transfer(ctx, req)
	suite.Error(err, "should return not found")
	suite.Nil(resp)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryTransfer_Found() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Create a transfer
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  "query-transfer-1",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x123",
		Amount:      sdkmath.NewInt(1000),
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_PENDING,
	}
	suite.Keeper.SetTransfer(suite.SdkCtx, transfer)

	req := &bridgepb.QueryTransferRequest{
		TransferId: "query-transfer-1",
	}
	resp, err := suite.queryServer.Transfer(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.NotNil(resp.Transfer)
	suite.Equal("query-transfer-1", resp.Transfer.TransferId)
}

// =============================================================================
// AllTransfers Query Tests
// =============================================================================

func (suite *QueryServerUncoveredTestSuite) TestQueryAllTransfers_NilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	resp, err := suite.queryServer.AllTransfers(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Nil(resp)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryAllTransfers_Empty() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryAllTransfersRequest{}
	resp, err := suite.queryServer.AllTransfers(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryAllTransfers_WithData() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Create some transfers
	for i := 0; i < 5; i++ {
		transfer := &bridgepb.CrossChainTransfer{
			TransferId:  "query-all-transfer-" + string(rune('0'+i)),
			SourceChain: "aura",
			TargetChain: "ethereum",
			Sender:      sdk.AccAddress("sender____________").String(),
			Recipient:   "0x123",
			Amount:      sdkmath.NewInt(int64(1000 + i*100)),
			Denom:       "uaura",
			Status:      bridgepb.TransferStatus_PENDING,
		}
		suite.Keeper.SetTransfer(suite.SdkCtx, transfer)
	}

	req := &bridgepb.QueryAllTransfersRequest{}
	resp, err := suite.queryServer.AllTransfers(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.GreaterOrEqual(len(resp.Transfers), 5)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryAllTransfers_WithStatusFilter() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Create transfers with different statuses
	statuses := []bridgepb.TransferStatus{
		bridgepb.TransferStatus_PENDING,
		bridgepb.TransferStatus_CONFIRMED,
		bridgepb.TransferStatus_COMPLETED,
		bridgepb.TransferStatus_FAILED,
	}

	for i, status := range statuses {
		transfer := &bridgepb.CrossChainTransfer{
			TransferId:  "query-status-filter-" + string(rune('0'+i)),
			SourceChain: "aura",
			TargetChain: "ethereum",
			Sender:      sdk.AccAddress("sender____________").String(),
			Recipient:   "0x123",
			Amount:      sdkmath.NewInt(int64(1000 + i*100)),
			Denom:       "uaura",
			Status:      status,
		}
		suite.Keeper.SetTransfer(suite.SdkCtx, transfer)
	}

	req := &bridgepb.QueryAllTransfersRequest{
		Status: "PENDING", // Status is a string filter
	}
	resp, err := suite.queryServer.AllTransfers(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
}

// =============================================================================
// UserTransfers Extended Tests
// =============================================================================

func (suite *QueryServerUncoveredTestSuite) TestQueryUserTransfers_WithData() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)
	userAddr := sdk.AccAddress("testuser__________").String()

	// Create transfers for the user
	for i := 0; i < 3; i++ {
		transfer := &bridgepb.CrossChainTransfer{
			TransferId:  "user-transfer-" + string(rune('0'+i)),
			SourceChain: "aura",
			TargetChain: "ethereum",
			Sender:      userAddr,
			Recipient:   "0x123",
			Amount:      sdkmath.NewInt(int64(1000 + i*100)),
			Denom:       "uaura",
			Status:      bridgepb.TransferStatus_PENDING,
		}
		suite.Keeper.SetTransfer(suite.SdkCtx, transfer)
		suite.Keeper.IndexUserTransfer(suite.SdkCtx, userAddr, transfer.TransferId)
	}

	req := &bridgepb.QueryUserTransfersRequest{
		Address: userAddr,
	}
	resp, err := suite.queryServer.UserTransfers(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.GreaterOrEqual(len(resp.Transfers), 3)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryUserTransfers_WithChainFilter() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)
	userAddr := sdk.AccAddress("chainfiltuser_____").String()

	// Create transfers to different chains
	chains := []string{"ethereum", "polygon", "arbitrum"}
	for i, chain := range chains {
		transfer := &bridgepb.CrossChainTransfer{
			TransferId:  "chain-filter-" + string(rune('0'+i)),
			SourceChain: "aura",
			TargetChain: chain,
			Sender:      userAddr,
			Recipient:   "0x123",
			Amount:      sdkmath.NewInt(int64(1000 + i*100)),
			Denom:       "uaura",
			Status:      bridgepb.TransferStatus_PENDING,
		}
		suite.Keeper.SetTransfer(suite.SdkCtx, transfer)
		suite.Keeper.IndexUserTransfer(suite.SdkCtx, userAddr, transfer.TransferId)
	}

	req := &bridgepb.QueryUserTransfersRequest{
		Address: userAddr,
		Chain:   "ethereum",
	}
	resp, err := suite.queryServer.UserTransfers(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
}

// =============================================================================
// WrappedToken Query Tests
// =============================================================================

func (suite *QueryServerUncoveredTestSuite) TestQueryWrappedToken_NilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	resp, err := suite.queryServer.WrappedToken(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Nil(resp)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryWrappedToken_EmptyDenom() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryWrappedTokenRequest{
		WrappedDenom: "",
	}
	resp, err := suite.queryServer.WrappedToken(ctx, req)
	suite.Error(err, "should reject empty denom")
	suite.Nil(resp)
}

func (suite *QueryServerUncoveredTestSuite) TestQueryWrappedToken_NotFound() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryWrappedTokenRequest{
		WrappedDenom: "nonexistent-token",
	}
	resp, err := suite.queryServer.WrappedToken(ctx, req)
	suite.Error(err, "should return not found")
	suite.Nil(resp)
}
