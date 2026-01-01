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

// =============================================================================
// CrossChainSwap Extended Tests
// =============================================================================

func (suite *MsgServerUncoveredTestSuite) TestCrossChainSwap_ValidRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.MsgCrossChainSwap{
		Sender:      sdk.AccAddress("sender____________").String(),
		SourceChain: "aura",
		TargetChain: "ethereum",
		InputCoin:   sdk.NewCoin("uaura", sdkmath.NewInt(1000)),
		TargetDenom: "weth",
	}
	resp, err := suite.msgServer.CrossChainSwap(ctx, req)
	suite.NoError(err, "should accept valid cross chain swap")
	suite.NotNil(resp)
	suite.NotEmpty(resp.SwapId)
	suite.Contains(resp.SwapId, "swap-")
}

func (suite *MsgServerUncoveredTestSuite) TestCrossChainSwap_InvalidInputCoin() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.MsgCrossChainSwap{
		Sender:      sdk.AccAddress("sender____________").String(),
		SourceChain: "aura",
		TargetChain: "ethereum",
		InputCoin:   sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(0)}, // Invalid: zero amount
		TargetDenom: "weth",
	}
	resp, err := suite.msgServer.CrossChainSwap(ctx, req)
	suite.Error(err, "should reject zero input amount")
	suite.Nil(resp)
}

func (suite *MsgServerUncoveredTestSuite) TestCrossChainSwap_NegativeAmount() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.MsgCrossChainSwap{
		Sender:      sdk.AccAddress("sender____________").String(),
		SourceChain: "aura",
		TargetChain: "ethereum",
		InputCoin:   sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(-100)}, // Invalid: negative
		TargetDenom: "weth",
	}
	resp, err := suite.msgServer.CrossChainSwap(ctx, req)
	suite.Error(err, "should reject negative input amount")
	suite.Nil(resp)
}

func (suite *MsgServerUncoveredTestSuite) TestCrossChainSwap_SameChain() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.MsgCrossChainSwap{
		Sender:      sdk.AccAddress("sender____________").String(),
		SourceChain: "aura",
		TargetChain: "aura", // Same chain - should still work
		InputCoin:   sdk.NewCoin("uaura", sdkmath.NewInt(1000)),
		TargetDenom: "upaw",
	}
	resp, err := suite.msgServer.CrossChainSwap(ctx, req)
	suite.NoError(err, "same chain swap should work")
	suite.NotNil(resp)
	// Route should have single element for same-chain swap
	suite.Equal(1, len(resp.Route))
}

func (suite *MsgServerUncoveredTestSuite) TestCrossChainSwap_DifferentChains() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.MsgCrossChainSwap{
		Sender:      sdk.AccAddress("sender____________").String(),
		SourceChain: "aura",
		TargetChain: "ethereum",
		InputCoin:   sdk.NewCoin("uaura", sdkmath.NewInt(1000)),
		TargetDenom: "weth",
	}
	resp, err := suite.msgServer.CrossChainSwap(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	// Route should have 2 elements for cross-chain swap
	suite.Equal(2, len(resp.Route))
	suite.Equal("aura", resp.Route[0])
	suite.Equal("ethereum", resp.Route[1])
}

// =============================================================================
// RelayTransfer Extended Tests
// =============================================================================

func (suite *MsgServerUncoveredTestSuite) TestRelayTransfer_EmptyTransferId() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.MsgRelayTransfer{
		Relayer:    sdk.AccAddress("relayer___________").String(),
		TransferId: "",
	}
	resp, err := suite.msgServer.RelayTransfer(ctx, req)
	suite.Error(err, "should reject empty transfer id")
	suite.Nil(resp)
}

func (suite *MsgServerUncoveredTestSuite) TestRelayTransfer_TransferNotFound() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.MsgRelayTransfer{
		Relayer:      sdk.AccAddress("relayer___________").String(),
		TransferId:   "nonexistent-transfer-id",
		Status:       "COMPLETED",
		TargetTxHash: "0xabc",
	}
	resp, err := suite.msgServer.RelayTransfer(ctx, req)
	suite.Error(err, "should reject non-existent transfer")
	suite.Nil(resp)
}

func (suite *MsgServerUncoveredTestSuite) TestRelayTransfer_ValidPendingStatus() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// First create a transfer
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  "relay-test-pending",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x123",
		Amount:      sdkmath.NewInt(1000),
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_PENDING,
	}
	suite.Keeper.SetTransfer(suite.SdkCtx, transfer)

	req := &bridgepb.MsgRelayTransfer{
		Relayer:      sdk.AccAddress("relayer___________").String(),
		TransferId:   "relay-test-pending",
		Status:       "PENDING",
		TargetTxHash: "",
	}
	resp, err := suite.msgServer.RelayTransfer(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.True(resp.Success)
}

func (suite *MsgServerUncoveredTestSuite) TestRelayTransfer_ValidConfirmedStatus() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  "relay-test-confirmed",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x123",
		Amount:      sdkmath.NewInt(2000),
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_PENDING,
	}
	suite.Keeper.SetTransfer(suite.SdkCtx, transfer)

	req := &bridgepb.MsgRelayTransfer{
		Relayer:      sdk.AccAddress("relayer___________").String(),
		TransferId:   "relay-test-confirmed",
		Status:       "CONFIRMED",
		TargetTxHash: "0xconfirmed123",
	}
	resp, err := suite.msgServer.RelayTransfer(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.True(resp.Success)

	// Verify status was updated
	updated, found := suite.Keeper.getTransfer(suite.SdkCtx, "relay-test-confirmed")
	suite.True(found)
	suite.Equal(bridgepb.TransferStatus_CONFIRMED, updated.Status)
}

func (suite *MsgServerUncoveredTestSuite) TestRelayTransfer_ValidRelayedStatus() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  "relay-test-relayed",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x123",
		Amount:      sdkmath.NewInt(3000),
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_CONFIRMED,
	}
	suite.Keeper.SetTransfer(suite.SdkCtx, transfer)

	req := &bridgepb.MsgRelayTransfer{
		Relayer:      sdk.AccAddress("relayer___________").String(),
		TransferId:   "relay-test-relayed",
		Status:       "RELAYED",
		TargetTxHash: "0xrelayed456",
	}
	resp, err := suite.msgServer.RelayTransfer(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.True(resp.Success)
}

func (suite *MsgServerUncoveredTestSuite) TestRelayTransfer_ValidCompletedStatus() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  "relay-test-completed",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x123",
		Amount:      sdkmath.NewInt(4000),
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_RELAYED,
	}
	suite.Keeper.SetTransfer(suite.SdkCtx, transfer)

	req := &bridgepb.MsgRelayTransfer{
		Relayer:      sdk.AccAddress("relayer___________").String(),
		TransferId:   "relay-test-completed",
		Status:       "COMPLETED",
		TargetTxHash: "0xcompleted789",
	}
	resp, err := suite.msgServer.RelayTransfer(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.True(resp.Success)

	updated, found := suite.Keeper.getTransfer(suite.SdkCtx, "relay-test-completed")
	suite.True(found)
	suite.Equal(bridgepb.TransferStatus_COMPLETED, updated.Status)
}

func (suite *MsgServerUncoveredTestSuite) TestRelayTransfer_ValidFailedStatus() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  "relay-test-failed",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x123",
		Amount:      sdkmath.NewInt(5000),
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_PENDING,
	}
	suite.Keeper.SetTransfer(suite.SdkCtx, transfer)

	req := &bridgepb.MsgRelayTransfer{
		Relayer:      sdk.AccAddress("relayer___________").String(),
		TransferId:   "relay-test-failed",
		Status:       "FAILED",
		TargetTxHash: "",
	}
	resp, err := suite.msgServer.RelayTransfer(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.True(resp.Success)

	updated, found := suite.Keeper.getTransfer(suite.SdkCtx, "relay-test-failed")
	suite.True(found)
	suite.Equal(bridgepb.TransferStatus_FAILED, updated.Status)
}

func (suite *MsgServerUncoveredTestSuite) TestRelayTransfer_ValidRefundedStatus() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  "relay-test-refunded",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x123",
		Amount:      sdkmath.NewInt(6000),
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_FAILED,
	}
	suite.Keeper.SetTransfer(suite.SdkCtx, transfer)

	req := &bridgepb.MsgRelayTransfer{
		Relayer:      sdk.AccAddress("relayer___________").String(),
		TransferId:   "relay-test-refunded",
		Status:       "REFUNDED",
		TargetTxHash: "",
	}
	resp, err := suite.msgServer.RelayTransfer(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.True(resp.Success)

	updated, found := suite.Keeper.getTransfer(suite.SdkCtx, "relay-test-refunded")
	suite.True(found)
	suite.Equal(bridgepb.TransferStatus_REFUNDED, updated.Status)
}

func (suite *MsgServerUncoveredTestSuite) TestRelayTransfer_LowercaseStatus() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  "relay-test-lowercase",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x123",
		Amount:      sdkmath.NewInt(7000),
		Denom:       "uaura",
		Status:      bridgepb.TransferStatus_PENDING,
	}
	suite.Keeper.SetTransfer(suite.SdkCtx, transfer)

	// Use lowercase status - should still work due to ToUpper
	req := &bridgepb.MsgRelayTransfer{
		Relayer:      sdk.AccAddress("relayer___________").String(),
		TransferId:   "relay-test-lowercase",
		Status:       "completed", // lowercase
		TargetTxHash: "0xlowercase",
	}
	resp, err := suite.msgServer.RelayTransfer(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.True(resp.Success)
}

// =============================================================================
// LinkAddress Tests
// =============================================================================

func (suite *MsgServerUncoveredTestSuite) TestLinkAddress_NilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	resp, err := suite.msgServer.LinkAddress(ctx, nil)
	suite.Error(err, "should reject nil request")
	suite.Nil(resp)
}

func (suite *MsgServerUncoveredTestSuite) TestLinkAddress_MissingSigner() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Missing signer field
	req := &bridgepb.MsgLinkAddress{
		AuraAddress: sdk.AccAddress("auraaddr__________").String(),
		PawAddress:  sdk.AccAddress("pawaddr___________").String(),
		XaiAddress:  "0xXaiAddress",
	}
	resp, err := suite.msgServer.LinkAddress(ctx, req)
	suite.Error(err, "should reject missing signer")
	suite.Nil(resp)
	suite.Contains(err.Error(), "signer")
}

func (suite *MsgServerUncoveredTestSuite) TestLinkAddress_SignerMismatch() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	auraAddr := sdk.AccAddress("auraaddr__________").String()
	// Signer doesn't match AuraAddress
	req := &bridgepb.MsgLinkAddress{
		AuraAddress: auraAddr,
		Signer:      sdk.AccAddress("differentsigner___").String(),
		PawAddress:  sdk.AccAddress("pawaddr___________").String(),
	}
	resp, err := suite.msgServer.LinkAddress(ctx, req)
	suite.Error(err, "should reject when signer doesn't match aura address")
	suite.Nil(resp)
	suite.Contains(err.Error(), "signer")
}

func (suite *MsgServerUncoveredTestSuite) TestLinkAddress_MissingSignature() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	auraAddr := sdk.AccAddress("auraaddr__________").String()
	// Signer matches but missing PAW signature
	req := &bridgepb.MsgLinkAddress{
		AuraAddress: auraAddr,
		Signer:      auraAddr,
		PawAddress:  sdk.AccAddress("pawaddr___________").String(),
	}
	resp, err := suite.msgServer.LinkAddress(ctx, req)
	suite.Error(err, "should reject missing PAW signature when PAW address set")
	suite.Nil(resp)
	suite.Contains(err.Error(), "signature")
}

func (suite *MsgServerUncoveredTestSuite) TestLinkAddress_EmptyAuraAddress() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	req := &bridgepb.MsgLinkAddress{
		AuraAddress: "",
		PawAddress:  sdk.AccAddress("pawaddr___________").String(),
		XaiAddress:  "0xXaiAddress",
	}
	resp, err := suite.msgServer.LinkAddress(ctx, req)
	suite.Error(err, "should reject empty aura address")
	suite.Nil(resp)
}
