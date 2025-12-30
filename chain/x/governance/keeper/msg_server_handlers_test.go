// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/governance/types"
	govpb "github.com/aequitas/aura/proto/aura/governance/v1beta1"
)

func setupMsgServerHandlersTest(t *testing.T) (*Keeper, govpb.MsgServer, sdk.Context) {
	input := keepertest.CreateTestInputWithKeys(t, "governance")
	mockStaking := NewMockStakingKeeper()
	mockBank := &MockBankKeeper{
		balances:       make(map[string]sdk.Coins),
		moduleBalances: make(map[string]sdk.Coins),
	}
	mockSecurity := &MockSecurityKeeper{}
	keeper := NewKeeper(input.Cdc, input.StoreKey, mockStaking, mockBank, mockSecurity)
	ctx := input.Ctx.WithKVGasConfig(storetypes.GasConfig{})
	keeper.SetParams(ctx, types.DefaultParams())
	msgServer := NewMsgServerImpl(keeper)
	return keeper, msgServer, ctx
}

// ============================
// RevealSecretVote Tests
// ============================

func TestMsgServer_RevealSecretVote_Success(t *testing.T) {
	keeper, msgServer, ctx := setupMsgServerHandlersTest(t)

	voter := testAddr("voter1")

	// Create a proposal in voting period
	ts, _ := gogotypes.TimestampProto(time.Now())
	endTs, _ := gogotypes.TimestampProto(time.Now().Add(7 * 24 * time.Hour))
	proposal := &types.Proposal{
		Id:              1,
		Title:           "Test Proposal",
		Description:     "Test Description",
		Proposer:        testAddr("proposer1"),
		Status:          types.StatusVotingPeriod,
		Category:        types.CategoryText,
		SubmitTime:      ts,
		VotingStartTime: ts,
		VotingEndTime:   endTs,
	}
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	// Create a secret vote
	secretVote := &types.Vote{
		ProposalId:     1,
		Voter:          voter,
		Option:         types.OptionUnspecified,
		IsSecret:       true,
		VoteCommitment: "abc123commitment",
		Timestamp:      ts,
		VotingPower:    "1000",
	}
	keeper.SetVote(ctx, secretVote)

	// Reveal the vote
	msg := &govpb.MsgRevealSecretVote{
		ProposalId: 1,
		Voter:      voter,
		Option:     types.VoteOptionYes,
	}

	resp, err := msgServer.RevealSecretVote(testGoCtx(ctx), msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify vote was revealed
	revealedVote, err := keeper.GetVote(ctx, 1, voter)
	require.NoError(t, err)
	require.Equal(t, types.VoteOptionYes, revealedVote.Option)
	require.False(t, revealedVote.IsSecret)

	// Verify event was emitted
	events := ctx.EventManager().Events()
	foundEvent := false
	for _, event := range events {
		if event.Type == types.EventTypeRevealVote {
			foundEvent = true
			break
		}
	}
	require.True(t, foundEvent)
}

func TestMsgServer_RevealSecretVote_NilMessage(t *testing.T) {
	_, msgServer, ctx := setupMsgServerHandlersTest(t)

	resp, err := msgServer.RevealSecretVote(testGoCtx(ctx), nil)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "empty request")
}

func TestMsgServer_RevealSecretVote_VoteNotFound(t *testing.T) {
	_, msgServer, ctx := setupMsgServerHandlersTest(t)

	voter := testAddr("voter1")

	msg := &govpb.MsgRevealSecretVote{
		ProposalId: 999,
		Voter:      voter,
		Option:     types.VoteOptionYes,
	}

	resp, err := msgServer.RevealSecretVote(testGoCtx(ctx), msg)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "not found")
}

func TestMsgServer_RevealSecretVote_NotSecretVote(t *testing.T) {
	keeper, msgServer, ctx := setupMsgServerHandlersTest(t)

	voter := testAddr("voter1")

	// Create a proposal in voting period
	ts, _ := gogotypes.TimestampProto(time.Now())
	endTs, _ := gogotypes.TimestampProto(time.Now().Add(7 * 24 * time.Hour))
	proposal := &types.Proposal{
		Id:              1,
		Title:           "Test Proposal",
		Description:     "Test Description",
		Proposer:        testAddr("proposer1"),
		Status:          types.StatusVotingPeriod,
		Category:        types.CategoryText,
		SubmitTime:      ts,
		VotingStartTime: ts,
		VotingEndTime:   endTs,
	}
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	// Create a regular (non-secret) vote
	regularVote := &types.Vote{
		ProposalId:  1,
		Voter:       voter,
		Option:      types.VoteOptionYes,
		IsSecret:    false,
		Timestamp:   ts,
		VotingPower: "1000",
	}
	keeper.SetVote(ctx, regularVote)

	// Try to reveal a non-secret vote
	msg := &govpb.MsgRevealSecretVote{
		ProposalId: 1,
		Voter:      voter,
		Option:     types.VoteOptionYes,
	}

	resp, err := msgServer.RevealSecretVote(testGoCtx(ctx), msg)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "not secret")
}

func TestMsgServer_RevealSecretVote_InvalidAddress(t *testing.T) {
	_, msgServer, ctx := setupMsgServerHandlersTest(t)

	msg := &govpb.MsgRevealSecretVote{
		ProposalId: 1,
		Voter:      "invalid_address",
		Option:     types.VoteOptionYes,
	}

	resp, err := msgServer.RevealSecretVote(testGoCtx(ctx), msg)
	require.Error(t, err)
	require.Nil(t, resp)
}

// ============================
// CosignVeto Tests
// ============================

func TestMsgServer_CosignVeto_Success(t *testing.T) {
	keeper, msgServer, ctx := setupMsgServerHandlersTest(t)

	// Create a proposal
	ts, _ := gogotypes.TimestampProto(time.Now())
	endTs, _ := gogotypes.TimestampProto(time.Now().Add(7 * 24 * time.Hour))
	proposal := &types.Proposal{
		Id:              1,
		Title:           "Test Proposal",
		Description:     "Test Description",
		Proposer:        testAddr("proposer1"),
		Status:          types.StatusVotingPeriod,
		Category:        types.CategoryText,
		SubmitTime:      ts,
		VotingStartTime: ts,
		VotingEndTime:   endTs,
	}
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	// Create an existing veto request
	veto := &types.VetoRequest{
		ProposalId: 1,
		Vetoer:     testAddr("vetoer1"),
		Reason:     "Security concern",
		Timestamp:  ts,
		Cosigners:  []string{testAddr("vetoer1")},
	}
	keeper.SetVetoRequest(ctx, veto)

	// Cosign the veto
	cosigner := testAddr("vetoer2")
	msg := &govpb.MsgCosignVeto{
		ProposalId: 1,
		Cosigner:   cosigner,
	}

	resp, err := msgServer.CosignVeto(testGoCtx(ctx), msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify cosigner was added
	updatedVeto, err := keeper.GetVetoRequest(ctx, 1)
	require.NoError(t, err)
	require.Contains(t, updatedVeto.Cosigners, cosigner)

	// Verify event was emitted
	events := ctx.EventManager().Events()
	foundEvent := false
	for _, event := range events {
		if event.Type == types.EventTypeVeto {
			foundEvent = true
			break
		}
	}
	require.True(t, foundEvent)
}

func TestMsgServer_CosignVeto_NilMessage(t *testing.T) {
	_, msgServer, ctx := setupMsgServerHandlersTest(t)

	resp, err := msgServer.CosignVeto(testGoCtx(ctx), nil)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "empty request")
}

func TestMsgServer_CosignVeto_VetoNotFound(t *testing.T) {
	_, msgServer, ctx := setupMsgServerHandlersTest(t)

	cosigner := testAddr("cosigner1")
	msg := &govpb.MsgCosignVeto{
		ProposalId: 999,
		Cosigner:   cosigner,
	}

	resp, err := msgServer.CosignVeto(testGoCtx(ctx), msg)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "not found")
}

func TestMsgServer_CosignVeto_DuplicateCosigner(t *testing.T) {
	keeper, msgServer, ctx := setupMsgServerHandlersTest(t)

	// Create a proposal
	ts, _ := gogotypes.TimestampProto(time.Now())
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddr("proposer1"),
		Status:      types.StatusVotingPeriod,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	// Create veto request with existing cosigner
	existingCosigner := testAddr("cosigner1")
	veto := &types.VetoRequest{
		ProposalId: 1,
		Vetoer:     testAddr("vetoer1"),
		Reason:     "Security concern",
		Timestamp:  ts,
		Cosigners:  []string{existingCosigner},
	}
	keeper.SetVetoRequest(ctx, veto)

	// Try to cosign with the same address again
	msg := &govpb.MsgCosignVeto{
		ProposalId: 1,
		Cosigner:   existingCosigner,
	}

	resp, err := msgServer.CosignVeto(testGoCtx(ctx), msg)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), types.ErrDuplicateCosigner.Error())
}

func TestMsgServer_CosignVeto_InvalidAddress(t *testing.T) {
	_, msgServer, ctx := setupMsgServerHandlersTest(t)

	msg := &govpb.MsgCosignVeto{
		ProposalId: 1,
		Cosigner:   "invalid_address",
	}

	resp, err := msgServer.CosignVeto(testGoCtx(ctx), msg)
	require.Error(t, err)
	require.Nil(t, resp)
}

func TestMsgServer_CosignVeto_MultipleSigners(t *testing.T) {
	keeper, msgServer, ctx := setupMsgServerHandlersTest(t)

	// Create a proposal
	ts, _ := gogotypes.TimestampProto(time.Now())
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddr("proposer1"),
		Status:      types.StatusVotingPeriod,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	// Create veto request
	veto := &types.VetoRequest{
		ProposalId: 1,
		Vetoer:     testAddr("vetoer1"),
		Reason:     "Security concern",
		Timestamp:  ts,
		Cosigners:  []string{},
	}
	keeper.SetVetoRequest(ctx, veto)

	// Add multiple cosigners
	cosigners := []string{
		testAddr("cosigner1"),
		testAddr("cosigner2"),
		testAddr("cosigner3"),
	}

	for _, cosigner := range cosigners {
		msg := &govpb.MsgCosignVeto{
			ProposalId: 1,
			Cosigner:   cosigner,
		}

		resp, err := msgServer.CosignVeto(testGoCtx(ctx), msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
	}

	// Verify all cosigners were added
	updatedVeto, err := keeper.GetVetoRequest(ctx, 1)
	require.NoError(t, err)
	require.Len(t, updatedVeto.Cosigners, 3)
	for _, cosigner := range cosigners {
		require.Contains(t, updatedVeto.Cosigners, cosigner)
	}
}

// ============================
// Additional Edge Case Tests
// ============================

func TestMsgServer_RevealSecretVote_MultipleReveals(t *testing.T) {
	keeper, msgServer, ctx := setupMsgServerHandlersTest(t)

	voter := testAddr("voter1")

	// Create a proposal in voting period
	ts, _ := gogotypes.TimestampProto(time.Now())
	endTs, _ := gogotypes.TimestampProto(time.Now().Add(7 * 24 * time.Hour))
	proposal := &types.Proposal{
		Id:              1,
		Title:           "Test Proposal",
		Description:     "Test Description",
		Proposer:        testAddr("proposer1"),
		Status:          types.StatusVotingPeriod,
		Category:        types.CategoryText,
		SubmitTime:      ts,
		VotingStartTime: ts,
		VotingEndTime:   endTs,
	}
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	// Create a secret vote
	secretVote := &types.Vote{
		ProposalId:     1,
		Voter:          voter,
		Option:         types.OptionUnspecified,
		IsSecret:       true,
		VoteCommitment: "abc123commitment",
		Timestamp:      ts,
		VotingPower:    "1000",
	}
	keeper.SetVote(ctx, secretVote)

	// First reveal
	msg := &govpb.MsgRevealSecretVote{
		ProposalId: 1,
		Voter:      voter,
		Option:     types.VoteOptionYes,
	}

	resp, err := msgServer.RevealSecretVote(testGoCtx(ctx), msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Try to reveal again (should fail - vote is no longer secret)
	resp, err = msgServer.RevealSecretVote(testGoCtx(ctx), msg)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "not secret")
}

func TestMsgServer_CosignVeto_OrderIndependent(t *testing.T) {
	keeper, msgServer, ctx := setupMsgServerHandlersTest(t)

	// Create a proposal
	ts, _ := gogotypes.TimestampProto(time.Now())
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddr("proposer1"),
		Status:      types.StatusVotingPeriod,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	// Create veto request
	veto := &types.VetoRequest{
		ProposalId: 1,
		Vetoer:     testAddr("vetoer1"),
		Reason:     "Security concern",
		Timestamp:  ts,
		Cosigners:  []string{},
	}
	keeper.SetVetoRequest(ctx, veto)

	// Add cosigners in specific order
	cosigner1 := testAddr("cosigner1")
	cosigner2 := testAddr("cosigner2")

	// First cosigner
	msg1 := &govpb.MsgCosignVeto{
		ProposalId: 1,
		Cosigner:   cosigner1,
	}
	resp1, err := msgServer.CosignVeto(testGoCtx(ctx), msg1)
	require.NoError(t, err)
	require.NotNil(t, resp1)

	// Second cosigner
	msg2 := &govpb.MsgCosignVeto{
		ProposalId: 1,
		Cosigner:   cosigner2,
	}
	resp2, err := msgServer.CosignVeto(testGoCtx(ctx), msg2)
	require.NoError(t, err)
	require.NotNil(t, resp2)

	// Verify both cosigners are present
	updatedVeto, err := keeper.GetVetoRequest(ctx, 1)
	require.NoError(t, err)
	require.Contains(t, updatedVeto.Cosigners, cosigner1)
	require.Contains(t, updatedVeto.Cosigners, cosigner2)
}

func TestMsgServer_RevealSecretVote_DifferentOptions(t *testing.T) {
	keeper, msgServer, ctx := setupMsgServerHandlersTest(t)

	// Create a proposal
	ts, _ := gogotypes.TimestampProto(time.Now())
	endTs, _ := gogotypes.TimestampProto(time.Now().Add(7 * 24 * time.Hour))
	proposal := &types.Proposal{
		Id:              1,
		Title:           "Test Proposal",
		Description:     "Test Description",
		Proposer:        testAddr("proposer1"),
		Status:          types.StatusVotingPeriod,
		Category:        types.CategoryText,
		SubmitTime:      ts,
		VotingStartTime: ts,
		VotingEndTime:   endTs,
	}
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	// Test revealing different vote options
	options := []types.VoteOption{
		types.VoteOptionYes,
		types.VoteOptionNo,
		types.VoteOptionAbstain,
		types.VoteOptionNoWithVeto,
	}

	for i, option := range options {
		voter := testAddr("voter" + string(rune('1'+i)))

		// Create secret vote
		secretVote := &types.Vote{
			ProposalId:     1,
			Voter:          voter,
			Option:         types.OptionUnspecified,
			IsSecret:       true,
			VoteCommitment: "commitment" + string(rune('1'+i)),
			Timestamp:      ts,
			VotingPower:    "1000",
		}
		keeper.SetVote(ctx, secretVote)

		// Reveal with specific option
		msg := &govpb.MsgRevealSecretVote{
			ProposalId: 1,
			Voter:      voter,
			Option:     option,
		}

		resp, err := msgServer.RevealSecretVote(testGoCtx(ctx), msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify revealed option is correct
		revealedVote, err := keeper.GetVote(ctx, 1, voter)
		require.NoError(t, err)
		require.Equal(t, option, revealedVote.Option)
		require.False(t, revealedVote.IsSecret)
	}
}

// Note: TestMsgServer_CosignVeto_NoSigners and TestMsgServer_RevealSecretVote_NoSigners
// were removed because with the current implementation, GetSigners() always returns
// the cosigner/voter as signer for valid addresses, and invalid addresses are rejected
// before GetSigners() is called.
