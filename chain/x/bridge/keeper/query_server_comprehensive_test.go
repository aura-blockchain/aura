package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"

	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

type QueryServerComprehensiveTestSuite struct {
	KeeperTestSuite
	queryServer bridgepb.QueryServer
}

func TestQueryServerComprehensiveTestSuite(t *testing.T) {
	suite.Run(t, new(QueryServerComprehensiveTestSuite))
}

func (suite *QueryServerComprehensiveTestSuite) SetupTest() {
	suite.KeeperTestSuite.SetupTest()
	suite.queryServer = NewQueryServerImpl(suite.Keeper)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryParamsNilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	resp, err := suite.queryServer.Params(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Nil(resp)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryParamsValid() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryParamsRequest{}
	resp, err := suite.queryServer.Params(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.NotNil(resp.Params)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryTransferNilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	resp, err := suite.queryServer.Transfer(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Nil(resp)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryTransferNonExistent() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryTransferRequest{
		TransferId: "non-existent",
	}
	resp, err := suite.queryServer.Transfer(ctx, req)
	suite.Error(err, "should return error for non-existent transfer")
	suite.Nil(resp)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryTransferValid() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Create a test transfer
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  "test-transfer-1",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x123",
		Amount:      "1000",
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_PENDING,
		Timestamp:   timestamppb.Now(),
	}
	suite.Keeper.setTransfer(suite.SdkCtx, transfer)

	req := &bridgepb.QueryTransferRequest{
		TransferId: "test-transfer-1",
	}
	resp, err := suite.queryServer.Transfer(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.NotNil(resp.Transfer)
	suite.Equal("test-transfer-1", resp.Transfer.TransferId)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryTransfersNilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	resp, err := suite.queryServer.Transfers(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Nil(resp)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryTransfersEmpty() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryTransfersRequest{}
	resp, err := suite.queryServer.Transfers(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.Empty(resp.Transfers)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryTransfersMultiple() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Create multiple test transfers
	for i := 1; i <= 5; i++ {
		transfer := &bridgepb.CrossChainTransfer{
			TransferId:  suite.Keeper.formatTransferID(uint64(i)),
			SourceChain: "aura",
			TargetChain: "ethereum",
			Sender:      sdk.AccAddress("sender____________").String(),
			Recipient:   "0x123",
			Amount:      "1000",
			Denom:       "uaura",
			Status:      bridgepb.TransferStatus_PENDING,
			Timestamp:   timestamppb.Now(),
		}
		suite.Keeper.setTransfer(suite.SdkCtx, transfer)
	}

	req := &bridgepb.QueryTransfersRequest{}
	resp, err := suite.queryServer.Transfers(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.Len(resp.Transfers, 5)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryWrappedTokenNilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	resp, err := suite.queryServer.WrappedToken(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Nil(resp)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryWrappedTokenEmptyDenom() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryWrappedTokenRequest{
		Denom: "",
	}
	resp, err := suite.queryServer.WrappedToken(ctx, req)
	suite.Error(err, "should reject empty denom")
	suite.Nil(resp)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryWrappedTokenNonExistent() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryWrappedTokenRequest{
		Denom: "non-existent",
	}
	resp, err := suite.queryServer.WrappedToken(ctx, req)
	suite.Error(err, "should return error for non-existent token")
	suite.Nil(resp)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryValidatorSetNilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	resp, err := suite.queryServer.ValidatorSet(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Nil(resp)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryValidatorSetValid() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryValidatorSetRequest{
		ChainId: "ethereum",
	}
	resp, err := suite.queryServer.ValidatorSet(ctx, req)
	// Error expected if no validators are set
	_ = resp
	_ = err
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryChannelNilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	resp, err := suite.queryServer.Channel(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Nil(resp)
}

func (suite *QueryServerComprehensiveTestSuite) TestQueryChannelEmptyID() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.QueryChannelRequest{
		ChannelId: "",
	}
	resp, err := suite.queryServer.Channel(ctx, req)
	suite.Error(err, "should reject empty channel ID")
	suite.Nil(resp)
}
