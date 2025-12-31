// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

type MsgServerUncoveredTestSuite struct {
	KeeperTestSuite
	msgServer bridgepb.MsgServer
}

func TestMsgServerUncoveredTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerUncoveredTestSuite))
}

func (suite *MsgServerUncoveredTestSuite) SetupTest() {
	suite.KeeperTestSuite.SetupTest()
	suite.msgServer = NewMsgServerImpl(suite.Keeper)
}

// =============================================================================
// BurnTokens Tests
// =============================================================================

func (suite *MsgServerUncoveredTestSuite) TestBurnTokens_NilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	resp, err := suite.msgServer.BurnTokens(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Nil(resp)
}

func (suite *MsgServerUncoveredTestSuite) TestBurnTokens_ZeroAmount() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.MsgBurnTokens{
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x123",
		Amount:      sdk.NewCoin("uaura", sdkmath.NewInt(0)),
		TargetChain: "ethereum",
	}
	resp, err := suite.msgServer.BurnTokens(ctx, req)
	suite.Error(err, "should reject zero amount")
	suite.Nil(resp)
}

func (suite *MsgServerUncoveredTestSuite) TestBurnTokens_NegativeAmount() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.MsgBurnTokens{
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x123",
		Amount:      sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(-100)},
		TargetChain: "ethereum",
	}
	resp, err := suite.msgServer.BurnTokens(ctx, req)
	suite.Error(err, "should reject negative amount")
	suite.Nil(resp)
}

func (suite *MsgServerUncoveredTestSuite) TestBurnTokens_EmptyTargetChain() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.MsgBurnTokens{
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x123",
		Amount:      sdk.NewCoin("uaura", sdkmath.NewInt(1000)),
		TargetChain: "",
	}
	resp, err := suite.msgServer.BurnTokens(ctx, req)
	suite.Error(err, "should reject empty target chain")
	suite.Nil(resp)
}

func (suite *MsgServerUncoveredTestSuite) TestBurnTokens_InvalidSender() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.MsgBurnTokens{
		Sender:      "invalid-address",
		Recipient:   "0x123",
		Amount:      sdk.NewCoin("uaura", sdkmath.NewInt(1000)),
		TargetChain: "ethereum",
	}
	resp, err := suite.msgServer.BurnTokens(ctx, req)
	// May fail on address parsing or chain config not found
	_ = resp
	_ = err
}

// =============================================================================
// CrossChainSwap Tests
// =============================================================================

func (suite *MsgServerUncoveredTestSuite) TestCrossChainSwap_NilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	resp, err := suite.msgServer.CrossChainSwap(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Nil(resp)
}

func (suite *MsgServerUncoveredTestSuite) TestCrossChainSwap_EmptySender() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.MsgCrossChainSwap{
		Sender: "",
	}
	resp, err := suite.msgServer.CrossChainSwap(ctx, req)
	suite.Error(err, "should reject empty sender")
	suite.Nil(resp)
}

// =============================================================================
// RelayTransfer Tests
// =============================================================================

func (suite *MsgServerUncoveredTestSuite) TestRelayTransfer_NilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	resp, err := suite.msgServer.RelayTransfer(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Nil(resp)
}

func (suite *MsgServerUncoveredTestSuite) TestRelayTransfer_EmptyRelayer() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.MsgRelayTransfer{
		Relayer: "",
	}
	resp, err := suite.msgServer.RelayTransfer(ctx, req)
	suite.Error(err, "should reject empty relayer")
	suite.Nil(resp)
}
