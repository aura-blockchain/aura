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

type PauseComprehensiveTestSuite struct {
	KeeperTestSuite
}

func TestPauseComprehensiveTestSuite(t *testing.T) {
	suite.Run(t, new(PauseComprehensiveTestSuite))
}

func (suite *PauseComprehensiveTestSuite) SetupTest() {
	suite.KeeperTestSuite.SetupTest()
}

// =============================================================================
// EmergencyPause Tests (using direct method call on msgServer)
// =============================================================================

func (suite *PauseComprehensiveTestSuite) TestEmergencyPause_NilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)
	ms := msgServer{Keeper: suite.Keeper}

	resp, err := ms.EmergencyPause(ctx, nil)
	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "nil")
}

func (suite *PauseComprehensiveTestSuite) TestEmergencyPause_EmptySigner() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)
	ms := msgServer{Keeper: suite.Keeper}

	req := &bridgepb.MsgEmergencyPause{
		Signer: "",
		Reason: "security issue",
	}
	resp, err := ms.EmergencyPause(ctx, req)
	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "signer")
}

func (suite *PauseComprehensiveTestSuite) TestEmergencyPause_EmptyReason() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)
	ms := msgServer{Keeper: suite.Keeper}

	req := &bridgepb.MsgEmergencyPause{
		Signer: sdk.AccAddress("signer____________").String(),
		Reason: "",
	}
	resp, err := ms.EmergencyPause(ctx, req)
	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "reason")
}

func (suite *PauseComprehensiveTestSuite) TestEmergencyPause_UnauthorizedSigner() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)
	ms := msgServer{Keeper: suite.Keeper}

	req := &bridgepb.MsgEmergencyPause{
		Signer: sdk.AccAddress("unauthorized______").String(),
		Reason: "testing unauthorized",
	}
	resp, err := ms.EmergencyPause(ctx, req)
	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "not authorized")
}

func (suite *PauseComprehensiveTestSuite) TestEmergencyPause_GlobalPause() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)
	ms := msgServer{Keeper: suite.Keeper}

	// Set authorized pause address
	params := suite.Keeper.GetParams(suite.SdkCtx)
	authorizedAddr := sdk.AccAddress("authorized________").String()
	params.EmergencyPauseAddresses = []string{authorizedAddr}
	suite.Keeper.SetParams(suite.SdkCtx, params)

	req := &bridgepb.MsgEmergencyPause{
		Signer: authorizedAddr,
		Reason: "critical security issue",
		Chains: []string{}, // empty = global pause
	}
	resp, err := ms.EmergencyPause(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.True(resp.Success)

	// Verify bridge is paused
	params = suite.Keeper.GetParams(suite.SdkCtx)
	suite.True(params.Paused)
}

func (suite *PauseComprehensiveTestSuite) TestEmergencyPause_SpecificChains() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)
	ms := msgServer{Keeper: suite.Keeper}

	// Set authorized pause address
	params := suite.Keeper.GetParams(suite.SdkCtx)
	authorizedAddr := sdk.AccAddress("authorized________").String()
	params.EmergencyPauseAddresses = []string{authorizedAddr}
	suite.Keeper.SetParams(suite.SdkCtx, params)

	req := &bridgepb.MsgEmergencyPause{
		Signer: authorizedAddr,
		Reason: "ethereum chain issue",
		Chains: []string{"ethereum", "polygon"},
	}
	resp, err := ms.EmergencyPause(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.True(resp.Success)

	// Verify specific chains are paused
	params = suite.Keeper.GetParams(suite.SdkCtx)
	suite.Contains(params.PausedChains, "ethereum")
	suite.Contains(params.PausedChains, "polygon")
}

func (suite *PauseComprehensiveTestSuite) TestEmergencyPause_ChainAlreadyPaused() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)
	ms := msgServer{Keeper: suite.Keeper}

	// Set authorized pause address and pre-pause a chain
	params := suite.Keeper.GetParams(suite.SdkCtx)
	authorizedAddr := sdk.AccAddress("authorized________").String()
	params.EmergencyPauseAddresses = []string{authorizedAddr}
	params.PausedChains = []string{"ethereum"}
	suite.Keeper.SetParams(suite.SdkCtx, params)

	req := &bridgepb.MsgEmergencyPause{
		Signer: authorizedAddr,
		Reason: "double pause test",
		Chains: []string{"ethereum"}, // already paused
	}
	resp, err := ms.EmergencyPause(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.True(resp.Success)

	// Ethereum should still only appear once
	params = suite.Keeper.GetParams(suite.SdkCtx)
	count := 0
	for _, c := range params.PausedChains {
		if c == "ethereum" {
			count++
		}
	}
	suite.Equal(1, count, "ethereum should only appear once in paused chains")
}

// =============================================================================
// Unpause Tests
// =============================================================================

func (suite *PauseComprehensiveTestSuite) TestUnpause_NilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)
	ms := msgServer{Keeper: suite.Keeper}

	resp, err := ms.Unpause(ctx, nil)
	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "nil")
}

func (suite *PauseComprehensiveTestSuite) TestUnpause_EmptyAuthority() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)
	ms := msgServer{Keeper: suite.Keeper}

	req := &bridgepb.MsgUnpause{
		Authority: "",
	}
	resp, err := ms.Unpause(ctx, req)
	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "authority")
}

func (suite *PauseComprehensiveTestSuite) TestUnpause_GlobalUnpause() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)
	ms := msgServer{Keeper: suite.Keeper}

	// First pause the bridge globally
	params := suite.Keeper.GetParams(suite.SdkCtx)
	params.Paused = true
	params.PausedChains = []string{"ethereum", "polygon"}
	suite.Keeper.SetParams(suite.SdkCtx, params)

	// Governance address unpause (authority check is at SDK level)
	req := &bridgepb.MsgUnpause{
		Authority: "aura1governance",
		Chains:    []string{}, // empty = global unpause
	}
	resp, err := ms.Unpause(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.True(resp.Success)

	// Verify bridge is unpaused
	params = suite.Keeper.GetParams(suite.SdkCtx)
	suite.False(params.Paused)
	suite.Len(params.PausedChains, 0)
}

func (suite *PauseComprehensiveTestSuite) TestUnpause_SpecificChains() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)
	ms := msgServer{Keeper: suite.Keeper}

	// First pause specific chains
	params := suite.Keeper.GetParams(suite.SdkCtx)
	params.PausedChains = []string{"ethereum", "polygon", "avalanche"}
	suite.Keeper.SetParams(suite.SdkCtx, params)

	// Unpause only ethereum (authority check is at SDK level)
	req := &bridgepb.MsgUnpause{
		Authority: "aura1governance",
		Chains:    []string{"ethereum"},
	}
	resp, err := ms.Unpause(ctx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.True(resp.Success)

	// Verify ethereum is unpaused, others still paused
	params = suite.Keeper.GetParams(suite.SdkCtx)
	suite.NotContains(params.PausedChains, "ethereum")
	suite.Contains(params.PausedChains, "polygon")
	suite.Contains(params.PausedChains, "avalanche")
}

// =============================================================================
// BurnTokens Tests
// =============================================================================

func (suite *PauseComprehensiveTestSuite) TestBurnTokens_NilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)
	ms := msgServer{Keeper: suite.Keeper}

	resp, err := ms.BurnTokens(ctx, nil)
	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "nil")
}

func (suite *PauseComprehensiveTestSuite) TestBurnTokens_NonPositiveAmount() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)
	ms := msgServer{Keeper: suite.Keeper}

	req := &bridgepb.MsgBurnTokens{
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x1234567890",
		Amount:      sdk.NewCoin("uaura", sdkmath.ZeroInt()),
		TargetChain: "ethereum",
	}
	resp, err := ms.BurnTokens(ctx, req)
	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "positive")
}

func (suite *PauseComprehensiveTestSuite) TestBurnTokens_NegativeAmount() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)
	ms := msgServer{Keeper: suite.Keeper}

	req := &bridgepb.MsgBurnTokens{
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x1234567890",
		Amount:      sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(-100)},
		TargetChain: "ethereum",
	}
	resp, err := ms.BurnTokens(ctx, req)
	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "positive")
}

func (suite *PauseComprehensiveTestSuite) TestBurnTokens_EmptyTargetChain() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)
	ms := msgServer{Keeper: suite.Keeper}

	req := &bridgepb.MsgBurnTokens{
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x1234567890",
		Amount:      sdk.NewCoin("uaura", sdkmath.NewInt(1000)),
		TargetChain: "",
	}
	resp, err := ms.BurnTokens(ctx, req)
	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "chain")
}

func (suite *PauseComprehensiveTestSuite) TestBurnTokens_PausedChain() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)
	ms := msgServer{Keeper: suite.Keeper}

	// Pause the target chain
	params := suite.Keeper.GetParams(suite.SdkCtx)
	params.PausedChains = []string{"ethereum"}
	suite.Keeper.SetParams(suite.SdkCtx, params)

	req := &bridgepb.MsgBurnTokens{
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x1234567890",
		Amount:      sdk.NewCoin("uaura", sdkmath.NewInt(1000)),
		TargetChain: "ethereum",
	}
	resp, err := ms.BurnTokens(ctx, req)
	suite.Error(err)
	suite.Nil(resp)
}

func (suite *PauseComprehensiveTestSuite) TestBurnTokens_ChainNotFound() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)
	ms := msgServer{Keeper: suite.Keeper}

	req := &bridgepb.MsgBurnTokens{
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x1234567890",
		Amount:      sdk.NewCoin("uaura", sdkmath.NewInt(1000)),
		TargetChain: "nonexistent-chain",
	}
	resp, err := ms.BurnTokens(ctx, req)
	suite.Error(err)
	suite.Nil(resp)
}

func (suite *PauseComprehensiveTestSuite) TestBurnTokens_ChainDisabled() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)
	ms := msgServer{Keeper: suite.Keeper}

	// Create a disabled chain config
	chainConfig := types.ChainConfig{
		ChainId:          "disabled-chain",
		ChainName:        "Disabled Chain",
		AddressPrefix:    "dis",
		MinConfirmations: 10,
		Enabled:          false,
	}
	suite.Keeper.setChainConfig(suite.SdkCtx, chainConfig)

	req := &bridgepb.MsgBurnTokens{
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x1234567890",
		Amount:      sdk.NewCoin("uaura", sdkmath.NewInt(1000)),
		TargetChain: "disabled-chain",
	}
	resp, err := ms.BurnTokens(ctx, req)
	suite.Error(err)
	suite.Nil(resp)
}

func (suite *PauseComprehensiveTestSuite) TestBurnTokens_InvalidSenderAddress() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)
	ms := msgServer{Keeper: suite.Keeper}

	// Create an enabled chain config
	chainConfig := types.ChainConfig{
		ChainId:          "ethereum",
		ChainName:        "Ethereum",
		AddressPrefix:    "0x",
		MinConfirmations: 10,
		Enabled:          true,
	}
	suite.Keeper.setChainConfig(suite.SdkCtx, chainConfig)

	req := &bridgepb.MsgBurnTokens{
		Sender:      "invalid-address",
		Recipient:   "0x1234567890",
		Amount:      sdk.NewCoin("uaura", sdkmath.NewInt(1000)),
		TargetChain: "ethereum",
	}
	resp, err := ms.BurnTokens(ctx, req)
	suite.Error(err)
	suite.Nil(resp)
}

func (suite *PauseComprehensiveTestSuite) TestBurnTokens_ValidRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)
	ms := msgServer{Keeper: suite.Keeper}

	// Create an enabled chain config
	chainConfig := types.ChainConfig{
		ChainId:          "ethereum",
		ChainName:        "Ethereum",
		AddressPrefix:    "0x",
		MinConfirmations: 10,
		Enabled:          true,
	}
	suite.Keeper.setChainConfig(suite.SdkCtx, chainConfig)

	req := &bridgepb.MsgBurnTokens{
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x1234567890",
		Amount:      sdk.NewCoin("uaura", sdkmath.NewInt(1000)),
		TargetChain: "ethereum",
	}
	resp, err := ms.BurnTokens(ctx, req)
	// Without bankKeeper, this should succeed
	suite.NoError(err)
	suite.NotNil(resp)
	suite.NotEmpty(resp.TransferId)
}

// =============================================================================
// payoutFraudProofReward Tests
// =============================================================================

func (suite *PauseComprehensiveTestSuite) TestPayoutFraudProofReward_ZeroReward() {
	err := suite.Keeper.payoutFraudProofReward(suite.SdkCtx, "challenger", "uaura", sdkmath.ZeroInt())
	suite.NoError(err) // Should be handled gracefully
}

func (suite *PauseComprehensiveTestSuite) TestPayoutFraudProofReward_NegativeReward() {
	err := suite.Keeper.payoutFraudProofReward(suite.SdkCtx, "challenger", "uaura", sdkmath.NewInt(-100))
	suite.NoError(err) // Should be handled gracefully
}

func (suite *PauseComprehensiveTestSuite) TestPayoutFraudProofReward_EmptyChallenger() {
	err := suite.Keeper.payoutFraudProofReward(suite.SdkCtx, "", "uaura", sdkmath.NewInt(100))
	suite.NoError(err) // Should be handled gracefully
}

func (suite *PauseComprehensiveTestSuite) TestPayoutFraudProofReward_EmptyDenom() {
	err := suite.Keeper.payoutFraudProofReward(suite.SdkCtx, "challenger", "", sdkmath.NewInt(100))
	suite.NoError(err) // Should be handled gracefully
}

func (suite *PauseComprehensiveTestSuite) TestPayoutFraudProofReward_NilBankKeeper() {
	// bankKeeper is nil in test setup
	err := suite.Keeper.payoutFraudProofReward(suite.SdkCtx, sdk.AccAddress("challenger________").String(), "uaura", sdkmath.NewInt(100))
	suite.NoError(err) // Should be handled gracefully
}

func (suite *PauseComprehensiveTestSuite) TestPayoutFraudProofReward_InvalidChallengerAddress() {
	// With nil bankKeeper, this returns early, so we just verify it doesn't panic
	err := suite.Keeper.payoutFraudProofReward(suite.SdkCtx, "invalid-address", "uaura", sdkmath.NewInt(100))
	suite.NoError(err)
}

// =============================================================================
// getFraudProofWindow Tests
// =============================================================================

func (suite *PauseComprehensiveTestSuite) TestGetFraudProofWindow() {
	window := suite.Keeper.getFraudProofWindow(suite.SdkCtx)
	suite.GreaterOrEqual(int64(window), int64(0))
}

// =============================================================================
// getFraudProofReward Tests
// =============================================================================

func (suite *PauseComprehensiveTestSuite) TestGetFraudProofReward() {
	reward := suite.Keeper.getFraudProofReward(suite.SdkCtx)
	suite.NotNil(reward)
}
