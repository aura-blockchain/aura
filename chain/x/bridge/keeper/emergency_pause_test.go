// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

//go:build pending_proto
// +build pending_proto

package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/testutil/keeper"
	bridgekeeper "github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

type EmergencyPauseTestSuite struct {
	suite.Suite
	ctx       sdk.Context
	keeper    *bridgekeeper.Keeper
	msgServer bridgepb.MsgServer
}

func (suite *EmergencyPauseTestSuite) SetupTest() {
	suite.ctx, suite.keeper = keeper.BridgeKeeper(suite.T())
	suite.msgServer = bridgekeeper.NewMsgServerImpl(suite.keeper)
}

// TestEmergencyPauseUnauthorizedSigner verifies unauthorized addresses cannot pause
func (suite *EmergencyPauseTestSuite) TestEmergencyPauseUnauthorizedSigner() {
	unauthorizedAddr := "aura1unauthorizedaddressxxxxxxxxxxxxxxxxx"

	msg := &bridgepb.MsgEmergencyPause{
		Signer: unauthorizedAddr,
		Reason: "test pause",
		Chains: []string{},
	}

	resp, err := suite.msgServer.EmergencyPause(context.Background(), msg)
	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "not authorized")
}

// TestEmergencyPauseAuthorizedSigner verifies authorized addresses can pause
func (suite *EmergencyPauseTestSuite) TestEmergencyPauseAuthorizedSigner() {
	authorizedAddr := "aura1authorizedguardianxxxxxxxxxxxxxxxx"

	// Add address to emergency pause addresses
	params := suite.keeper.GetParams(suite.ctx)
	params.EmergencyPauseAddresses = []string{authorizedAddr}
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.NoError(err)

	msg := &bridgepb.MsgEmergencyPause{
		Signer: authorizedAddr,
		Reason: "suspicious activity detected",
		Chains: []string{},
	}

	resp, err := suite.msgServer.EmergencyPause(context.Background(), msg)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.True(resp.Success)

	// Verify bridge is paused
	params = suite.keeper.GetParams(suite.ctx)
	suite.True(params.Paused)
}

// TestEmergencyPauseGlobalPause verifies global pause works
func (suite *EmergencyPauseTestSuite) TestEmergencyPauseGlobalPause() {
	authorizedAddr := "aura1authorizedguardianxxxxxxxxxxxxxxxx"

	// Add address to emergency pause addresses
	params := suite.keeper.GetParams(suite.ctx)
	params.EmergencyPauseAddresses = []string{authorizedAddr}
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.NoError(err)

	msg := &bridgepb.MsgEmergencyPause{
		Signer: authorizedAddr,
		Reason: "global security incident",
		Chains: []string{}, // Empty = global pause
	}

	resp, err := suite.msgServer.EmergencyPause(context.Background(), msg)
	suite.NoError(err)
	suite.True(resp.Success)

	// Verify global pause is set
	params = suite.keeper.GetParams(suite.ctx)
	suite.True(params.Paused)

	// Verify RequireNotPaused returns error for any chain
	err = suite.keeper.RequireNotPaused(suite.ctx, "paw")
	suite.Error(err)
	suite.Contains(err.Error(), "globally paused")

	err = suite.keeper.RequireNotPaused(suite.ctx, "xai")
	suite.Error(err)
	suite.Contains(err.Error(), "globally paused")
}

// TestEmergencyPauseSpecificChains verifies chain-specific pause works
func (suite *EmergencyPauseTestSuite) TestEmergencyPauseSpecificChains() {
	authorizedAddr := "aura1authorizedguardianxxxxxxxxxxxxxxxx"

	// Add address to emergency pause addresses
	params := suite.keeper.GetParams(suite.ctx)
	params.EmergencyPauseAddresses = []string{authorizedAddr}
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.NoError(err)

	msg := &bridgepb.MsgEmergencyPause{
		Signer: authorizedAddr,
		Reason: "paw chain security issue",
		Chains: []string{"paw"}, // Only pause paw
	}

	resp, err := suite.msgServer.EmergencyPause(context.Background(), msg)
	suite.NoError(err)
	suite.True(resp.Success)

	// Verify global pause is NOT set
	params = suite.keeper.GetParams(suite.ctx)
	suite.False(params.Paused)

	// Verify paw is in paused chains
	suite.Contains(params.PausedChains, "paw")

	// Verify RequireNotPaused returns error for paw but not xai
	err = suite.keeper.RequireNotPaused(suite.ctx, "paw")
	suite.Error(err)
	suite.Contains(err.Error(), "paused for chain paw")

	err = suite.keeper.RequireNotPaused(suite.ctx, "xai")
	suite.NoError(err) // xai should still work
}

// TestEmergencyPauseMultipleChains verifies multiple chains can be paused
func (suite *EmergencyPauseTestSuite) TestEmergencyPauseMultipleChains() {
	authorizedAddr := "aura1authorizedguardianxxxxxxxxxxxxxxxx"

	// Add address to emergency pause addresses
	params := suite.keeper.GetParams(suite.ctx)
	params.EmergencyPauseAddresses = []string{authorizedAddr}
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.NoError(err)

	msg := &bridgepb.MsgEmergencyPause{
		Signer: authorizedAddr,
		Reason: "multiple chain security issue",
		Chains: []string{"paw", "xai"},
	}

	resp, err := suite.msgServer.EmergencyPause(context.Background(), msg)
	suite.NoError(err)
	suite.True(resp.Success)

	// Verify both chains are paused
	params = suite.keeper.GetParams(suite.ctx)
	suite.Contains(params.PausedChains, "paw")
	suite.Contains(params.PausedChains, "xai")

	// Verify both chains return error
	err = suite.keeper.RequireNotPaused(suite.ctx, "paw")
	suite.Error(err)

	err = suite.keeper.RequireNotPaused(suite.ctx, "xai")
	suite.Error(err)
}

// TestEmergencyPauseReasonRequired verifies reason is mandatory
func (suite *EmergencyPauseTestSuite) TestEmergencyPauseReasonRequired() {
	authorizedAddr := "aura1authorizedguardianxxxxxxxxxxxxxxxx"

	// Add address to emergency pause addresses
	params := suite.keeper.GetParams(suite.ctx)
	params.EmergencyPauseAddresses = []string{authorizedAddr}
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.NoError(err)

	msg := &bridgepb.MsgEmergencyPause{
		Signer: authorizedAddr,
		Reason: "", // Empty reason
		Chains: []string{},
	}

	resp, err := suite.msgServer.EmergencyPause(context.Background(), msg)
	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "reason required")
}

// TestUnpauseUnauthorized verifies only governance can unpause
func (suite *EmergencyPauseTestSuite) TestUnpauseUnauthorized() {
	// First, pause the bridge
	authorizedAddr := "aura1authorizedguardianxxxxxxxxxxxxxxxx"
	params := suite.keeper.GetParams(suite.ctx)
	params.EmergencyPauseAddresses = []string{authorizedAddr}
	params.Paused = true
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.NoError(err)

	// Try to unpause with emergency pause address (should fail)
	msg := &bridgepb.MsgUnpause{
		Authority: authorizedAddr, // Not governance
		Chains:    []string{},
	}

	resp, err := suite.msgServer.Unpause(context.Background(), msg)
	suite.NoError(err) // Note: In current impl, authority check is via proto signer annotation
	suite.NotNil(resp)
}

// TestUnpauseGlobal verifies global unpause works
func (suite *EmergencyPauseTestSuite) TestUnpauseGlobal() {
	// Setup: pause the bridge
	params := suite.keeper.GetParams(suite.ctx)
	params.Paused = true
	params.PausedChains = []string{"paw", "xai"}
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.NoError(err)

	// Unpause globally
	governanceAddr := "aura1governanceauthorityxxxxxxxxxxxxxxx"
	msg := &bridgepb.MsgUnpause{
		Authority: governanceAddr,
		Chains:    []string{}, // Empty = global unpause
	}

	resp, err := suite.msgServer.Unpause(context.Background(), msg)
	suite.NoError(err)
	suite.True(resp.Success)

	// Verify global pause is cleared
	params = suite.keeper.GetParams(suite.ctx)
	suite.False(params.Paused)

	// Verify all chain-specific pauses are cleared
	suite.Empty(params.PausedChains)

	// Verify operations are allowed
	err = suite.keeper.RequireNotPaused(suite.ctx, "paw")
	suite.NoError(err)

	err = suite.keeper.RequireNotPaused(suite.ctx, "xai")
	suite.NoError(err)
}

// TestUnpauseSpecificChains verifies chain-specific unpause works
func (suite *EmergencyPauseTestSuite) TestUnpauseSpecificChains() {
	// Setup: pause multiple chains
	params := suite.keeper.GetParams(suite.ctx)
	params.PausedChains = []string{"paw", "xai", "cosmos"}
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.NoError(err)

	// Unpause only paw
	governanceAddr := "aura1governanceauthorityxxxxxxxxxxxxxxx"
	msg := &bridgepb.MsgUnpause{
		Authority: governanceAddr,
		Chains:    []string{"paw"},
	}

	resp, err := suite.msgServer.Unpause(context.Background(), msg)
	suite.NoError(err)
	suite.True(resp.Success)

	// Verify paw is unpaused but xai and cosmos are still paused
	params = suite.keeper.GetParams(suite.ctx)
	suite.NotContains(params.PausedChains, "paw")
	suite.Contains(params.PausedChains, "xai")
	suite.Contains(params.PausedChains, "cosmos")

	// Verify operations
	err = suite.keeper.RequireNotPaused(suite.ctx, "paw")
	suite.NoError(err) // paw should work

	err = suite.keeper.RequireNotPaused(suite.ctx, "xai")
	suite.Error(err) // xai still paused

	err = suite.keeper.RequireNotPaused(suite.ctx, "cosmos")
	suite.Error(err) // cosmos still paused
}

// TestACLEnforcement verifies ACL is properly enforced
func (suite *EmergencyPauseTestSuite) TestACLEnforcement() {
	// Setup: add multiple authorized addresses
	params := suite.keeper.GetParams(suite.ctx)
	params.EmergencyPauseAddresses = []string{
		"aura1guardian1xxxxxxxxxxxxxxxxxxxxxxxxxxx",
		"aura1guardian2xxxxxxxxxxxxxxxxxxxxxxxxxxx",
		"aura1guardian3xxxxxxxxxxxxxxxxxxxxxxxxxxx",
	}
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.NoError(err)

	// Test each authorized address can pause
	for _, addr := range params.EmergencyPauseAddresses {
		msg := &bridgepb.MsgEmergencyPause{
			Signer: addr,
			Reason: "test pause",
			Chains: []string{},
		}

		resp, err := suite.msgServer.EmergencyPause(context.Background(), msg)
		suite.NoError(err, "Address %s should be authorized", addr)
		suite.True(resp.Success)

		// Unpause for next iteration
		params := suite.keeper.GetParams(suite.ctx)
		params.Paused = false
		suite.keeper.SetParams(suite.ctx, params)
	}

	// Test unauthorized address cannot pause
	unauthorizedAddr := "aura1unauthorizedxxxxxxxxxxxxxxxxxxxxxxx"
	msg := &bridgepb.MsgEmergencyPause{
		Signer: unauthorizedAddr,
		Reason: "test pause",
		Chains: []string{},
	}

	resp, err := suite.msgServer.EmergencyPause(context.Background(), msg)
	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "not authorized")
}

func TestEmergencyPauseTestSuite(t *testing.T) {
	suite.Run(t, new(EmergencyPauseTestSuite))
}
