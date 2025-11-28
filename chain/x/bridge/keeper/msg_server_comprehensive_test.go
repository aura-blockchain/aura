package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

type MsgServerComprehensiveTestSuite struct {
	KeeperTestSuite
	msgServer bridgepb.MsgServer
}

func TestMsgServerComprehensiveTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerComprehensiveTestSuite))
}

func (suite *MsgServerComprehensiveTestSuite) SetupTest() {
	suite.KeeperTestSuite.SetupTest()
	suite.msgServer = NewMsgServerImpl(suite.Keeper)
}

func (suite *MsgServerComprehensiveTestSuite) TestLockTokensNilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	_, err := suite.msgServer.LockTokens(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Contains(err.Error(), "nil")
}

func (suite *MsgServerComprehensiveTestSuite) TestLockTokensZeroAmount() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	msg := &bridgepb.MsgLockTokens{
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "recipient",
		Amount:      &bridgepb.Coin{Denom: "uaura", Amount: sdkmath.ZeroInt()},
		TargetChain: "ethereum",
	}

	_, err := suite.msgServer.LockTokens(ctx, msg)
	suite.Error(err, "should reject zero amount")
	suite.Contains(err.Error(), "positive")
}

func (suite *MsgServerComprehensiveTestSuite) TestLockTokensInvalidSender() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	msg := &bridgepb.MsgLockTokens{
		Sender:      "invalid-address",
		Recipient:   "recipient",
		Amount:      &bridgepb.Coin{Denom: "uaura", Amount: sdkmath.NewInt(1000)},
		TargetChain: "ethereum",
	}

	_, err := suite.msgServer.LockTokens(ctx, msg)
	suite.Error(err, "should reject invalid sender address")
}

func (suite *MsgServerComprehensiveTestSuite) TestLockTokensEmptyTargetChain() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	msg := &bridgepb.MsgLockTokens{
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "recipient",
		Amount:      &bridgepb.Coin{Denom: "uaura", Amount: sdkmath.NewInt(1000)},
		TargetChain: "",
	}

	_, err := suite.msgServer.LockTokens(ctx, msg)
	suite.Error(err, "should reject empty target chain")
	suite.Contains(err.Error(), "target chain")
}

func (suite *MsgServerComprehensiveTestSuite) TestMintTokensNilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	_, err := suite.msgServer.MintTokens(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Contains(err.Error(), "nil")
}

func (suite *MsgServerComprehensiveTestSuite) TestMintTokensEmptyValidator() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	msg := &bridgepb.MsgMintTokens{
		Validator:     "",
		Recipient:     sdk.AccAddress("recipient_________").String(),
		Amount:        "1000",
		Denom:         "uaura",
		SourceChain:   "ethereum",
		SourceTxHash:  "0x123",
	}

	_, err := suite.msgServer.MintTokens(ctx, msg)
	suite.Error(err, "should reject empty validator")
	suite.Contains(err.Error(), "validator")
}

func (suite *MsgServerComprehensiveTestSuite) TestMintTokensInvalidAmount() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	msg := &bridgepb.MsgMintTokens{
		Validator:     sdk.ValAddress("validator_________").String(),
		Recipient:     sdk.AccAddress("recipient_________").String(),
		Amount:        "invalid",
		Denom:         "uaura",
		SourceChain:   "ethereum",
		SourceTxHash:  "0x123",
	}

	_, err := suite.msgServer.MintTokens(ctx, msg)
	suite.Error(err, "should reject invalid amount")
	suite.Contains(err.Error(), "amount")
}

func (suite *MsgServerComprehensiveTestSuite) TestMintTokensZeroAmount() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	msg := &bridgepb.MsgMintTokens{
		Validator:     sdk.ValAddress("validator_________").String(),
		Recipient:     sdk.AccAddress("recipient_________").String(),
		Amount:        "0",
		Denom:         "uaura",
		SourceChain:   "ethereum",
		SourceTxHash:  "0x123",
	}

	_, err := suite.msgServer.MintTokens(ctx, msg)
	suite.Error(err, "should reject zero amount")
}

func (suite *MsgServerComprehensiveTestSuite) TestBurnTokensNilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	_, err := suite.msgServer.BurnTokens(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Contains(err.Error(), "nil")
}

func (suite *MsgServerComprehensiveTestSuite) TestUnlockTokensNilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	_, err := suite.msgServer.UnlockTokens(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Contains(err.Error(), "nil")
}

func (suite *MsgServerComprehensiveTestSuite) TestSubmitMerkleProofNilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	_, err := suite.msgServer.SubmitMerkleProof(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Contains(err.Error(), "nil")
}

func (suite *MsgServerComprehensiveTestSuite) TestUpdateValidatorSetNilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	_, err := suite.msgServer.UpdateValidatorSet(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Contains(err.Error(), "nil")
}
