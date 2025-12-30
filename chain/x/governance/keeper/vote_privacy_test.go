// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/governance/types"
)

func setupVotePrivacyKeeper(t *testing.T) (*Keeper, sdk.Context) {
	input := keepertest.CreateTestInputWithKeys(t, "governance")
	mockStaking := NewMockStakingKeeper()
	mockBank := &MockBankKeeper{
		balances:       make(map[string]sdk.Coins),
		moduleBalances: make(map[string]sdk.Coins),
	}
	mockSecurity := &MockSecurityKeeper{}
	keeper := NewKeeper(input.Cdc, input.StoreKey, mockStaking, mockBank, mockSecurity)
	ctx := input.Ctx.WithKVGasConfig(storetypes.GasConfig{})

	// Enable secret ballot by default
	params := types.DefaultParams()
	params.SecretBallotEnabled = true
	keeper.SetParams(ctx, params)

	return keeper, ctx
}

func createPrivacyVotingProposal(t *testing.T, keeper *Keeper, ctx sdk.Context, proposalID uint64) {
	now := time.Now()
	submitTime, _ := gogotypes.TimestampProto(now)
	votingEnd, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))

	proposal := &types.Proposal{
		Id:            proposalID,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddr("proposer1"),
		Status:        types.StatusVotingPeriod,
		Category:      types.CategoryText,
		SubmitTime:    submitTime,
		VotingEndTime: votingEnd,
	}
	err := keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)
}

func TestCommitVote_Success(t *testing.T) {
	keeper, ctx := setupVotePrivacyKeeper(t)

	proposalID := uint64(1)
	voter := testAddr("voter1")
	voteHash := "abcdef1234567890"

	// Create proposal in voting period
	createPrivacyVotingProposal(t, keeper, ctx, proposalID)

	err := keeper.CommitVote(ctx, proposalID, voter, voteHash)
	require.NoError(t, err)

	// Verify commitment was stored
	commitment, err := keeper.getVoteCommitment(ctx, proposalID, voter)
	require.NoError(t, err)
	require.Equal(t, proposalID, commitment.ProposalId)
	require.Equal(t, voter, commitment.Voter)
	require.Equal(t, voteHash, commitment.VoteHash)
	require.False(t, commitment.Revealed)
	require.NotNil(t, commitment.CommittedAt)
}

func TestCommitVote_SecretBallotDisabled(t *testing.T) {
	keeper, ctx := setupVotePrivacyKeeper(t)

	// Disable secret ballot
	params := keeper.GetParams(ctx)
	params.SecretBallotEnabled = false
	keeper.SetParams(ctx, params)

	proposalID := uint64(1)
	voter := testAddr("voter1")
	voteHash := "abcdef1234567890"

	createPrivacyVotingProposal(t, keeper, ctx, proposalID)

	err := keeper.CommitVote(ctx, proposalID, voter, voteHash)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrSecretBallotDisabled)
}

func TestCommitVote_ProposalNotFound(t *testing.T) {
	keeper, ctx := setupVotePrivacyKeeper(t)

	proposalID := uint64(999)
	voter := testAddr("voter1")
	voteHash := "abcdef1234567890"

	err := keeper.CommitVote(ctx, proposalID, voter, voteHash)
	require.Error(t, err)
}

func TestCommitVote_InvalidProposalStatus(t *testing.T) {
	keeper, ctx := setupVotePrivacyKeeper(t)

	proposalID := uint64(1)
	voter := testAddr("voter1")
	voteHash := "abcdef1234567890"

	// Create proposal in passed status (not voting)
	now := time.Now()
	submitTime, _ := gogotypes.TimestampProto(now)

	proposal := &types.Proposal{
		Id:          proposalID,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddr("proposer1"),
		Status:      types.StatusPassed,
		Category:    types.CategoryText,
		SubmitTime:  submitTime,
	}
	err := keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	err = keeper.CommitVote(ctx, proposalID, voter, voteHash)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidProposalStatus)
}

func TestCommitVote_EventEmission(t *testing.T) {
	keeper, ctx := setupVotePrivacyKeeper(t)

	proposalID := uint64(1)
	voter := testAddr("voter1")
	voteHash := "abcdef1234567890"

	createPrivacyVotingProposal(t, keeper, ctx, proposalID)

	// Clear events
	ctx = ctx.WithEventManager(sdk.NewEventManager())

	err := keeper.CommitVote(ctx, proposalID, voter, voteHash)
	require.NoError(t, err)

	// Check event was emitted
	events := ctx.EventManager().Events()
	require.NotEmpty(t, events)

	found := false
	for _, event := range events {
		if event.Type == "vote_committed" {
			found = true
		}
	}
	require.True(t, found, "vote_committed event should be emitted")
}

func TestRevealVote_Success(t *testing.T) {
	keeper, ctx := setupVotePrivacyKeeper(t)
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)

	proposalID := uint64(1)
	voter := testAddr("voter1")
	option := types.VoteOptionYes
	salt := "randomsalt123"

	// Set voting power
	mockStaking.SetDelegatorBonded(voter, sdkmath.NewInt(1000000))

	// Create proposal
	createPrivacyVotingProposal(t, keeper, ctx, proposalID)

	// Calculate hash
	voteHash := keeper.calculateVoteHash(proposalID, voter, option, salt)

	// Commit vote
	err := keeper.CommitVote(ctx, proposalID, voter, voteHash)
	require.NoError(t, err)

	// Reveal vote
	err = keeper.RevealVote(ctx, proposalID, voter, option, salt)
	require.NoError(t, err)

	// Verify vote was stored
	vote, err := keeper.GetVote(ctx, proposalID, voter)
	require.NoError(t, err)
	require.Equal(t, proposalID, vote.ProposalId)
	require.Equal(t, voter, vote.Voter)
	require.Equal(t, option, vote.Option)
	require.True(t, vote.IsSecret)

	// Verify commitment was marked as revealed
	commitment, err := keeper.getVoteCommitment(ctx, proposalID, voter)
	require.NoError(t, err)
	require.True(t, commitment.Revealed)
}

func TestRevealVote_NoCommitment(t *testing.T) {
	keeper, ctx := setupVotePrivacyKeeper(t)

	proposalID := uint64(1)
	voter := testAddr("voter1")
	option := types.VoteOptionYes
	salt := "randomsalt123"

	createPrivacyVotingProposal(t, keeper, ctx, proposalID)

	// Try to reveal without committing first
	err := keeper.RevealVote(ctx, proposalID, voter, option, salt)
	require.Error(t, err)
}

func TestRevealVote_InvalidHash(t *testing.T) {
	keeper, ctx := setupVotePrivacyKeeper(t)
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)

	proposalID := uint64(1)
	voter := testAddr("voter1")
	option := types.VoteOptionYes
	correctSalt := "correctsalt"
	wrongSalt := "wrongsalt"

	mockStaking.SetDelegatorBonded(voter, sdkmath.NewInt(1000000))

	createPrivacyVotingProposal(t, keeper, ctx, proposalID)

	// Commit with correct salt
	voteHash := keeper.calculateVoteHash(proposalID, voter, option, correctSalt)
	err := keeper.CommitVote(ctx, proposalID, voter, voteHash)
	require.NoError(t, err)

	// Try to reveal with wrong salt
	err = keeper.RevealVote(ctx, proposalID, voter, option, wrongSalt)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidVoteReveal)
}

func TestRevealVote_AlreadyRevealed(t *testing.T) {
	keeper, ctx := setupVotePrivacyKeeper(t)
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)

	proposalID := uint64(1)
	voter := testAddr("voter1")
	option := types.VoteOptionYes
	salt := "randomsalt123"

	mockStaking.SetDelegatorBonded(voter, sdkmath.NewInt(1000000))

	createPrivacyVotingProposal(t, keeper, ctx, proposalID)

	// Commit and reveal
	voteHash := keeper.calculateVoteHash(proposalID, voter, option, salt)
	err := keeper.CommitVote(ctx, proposalID, voter, voteHash)
	require.NoError(t, err)

	err = keeper.RevealVote(ctx, proposalID, voter, option, salt)
	require.NoError(t, err)

	// Try to reveal again
	err = keeper.RevealVote(ctx, proposalID, voter, option, salt)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrVoteAlreadyRevealed)
}

func TestRevealVote_DifferentOption(t *testing.T) {
	keeper, ctx := setupVotePrivacyKeeper(t)
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)

	proposalID := uint64(1)
	voter := testAddr("voter1")
	committedOption := types.VoteOptionYes
	revealedOption := types.VoteOptionNo
	salt := "randomsalt123"

	mockStaking.SetDelegatorBonded(voter, sdkmath.NewInt(1000000))

	createPrivacyVotingProposal(t, keeper, ctx, proposalID)

	// Commit with Yes
	voteHash := keeper.calculateVoteHash(proposalID, voter, committedOption, salt)
	err := keeper.CommitVote(ctx, proposalID, voter, voteHash)
	require.NoError(t, err)

	// Try to reveal with No (should fail due to hash mismatch)
	err = keeper.RevealVote(ctx, proposalID, voter, revealedOption, salt)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidVoteReveal)
}

func TestRevealVote_EventEmission(t *testing.T) {
	keeper, ctx := setupVotePrivacyKeeper(t)
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)

	proposalID := uint64(1)
	voter := testAddr("voter1")
	option := types.VoteOptionYes
	salt := "randomsalt123"

	mockStaking.SetDelegatorBonded(voter, sdkmath.NewInt(1000000))

	createPrivacyVotingProposal(t, keeper, ctx, proposalID)

	// Commit vote
	voteHash := keeper.calculateVoteHash(proposalID, voter, option, salt)
	err := keeper.CommitVote(ctx, proposalID, voter, voteHash)
	require.NoError(t, err)

	// Clear events
	ctx = ctx.WithEventManager(sdk.NewEventManager())

	// Reveal vote
	err = keeper.RevealVote(ctx, proposalID, voter, option, salt)
	require.NoError(t, err)

	// Check event was emitted
	events := ctx.EventManager().Events()
	require.NotEmpty(t, events)

	found := false
	for _, event := range events {
		if event.Type == "vote_revealed" {
			found = true
		}
	}
	require.True(t, found, "vote_revealed event should be emitted")
}

func TestCalculateVoteHash_Consistency(t *testing.T) {
	keeper, ctx := setupVotePrivacyKeeper(t)
	_ = ctx

	proposalID := uint64(1)
	voter := testAddr("voter1")
	option := types.VoteOptionYes
	salt := "randomsalt123"

	// Calculate hash multiple times
	hash1 := keeper.calculateVoteHash(proposalID, voter, option, salt)
	hash2 := keeper.calculateVoteHash(proposalID, voter, option, salt)

	// Should be consistent
	require.Equal(t, hash1, hash2)
}

func TestCalculateVoteHash_DifferentInputs(t *testing.T) {
	keeper, ctx := setupVotePrivacyKeeper(t)
	_ = ctx

	proposalID := uint64(1)
	voter := testAddr("voter1")
	option := types.VoteOptionYes
	salt1 := "salt1"
	salt2 := "salt2"

	// Different salts should produce different hashes
	hash1 := keeper.calculateVoteHash(proposalID, voter, option, salt1)
	hash2 := keeper.calculateVoteHash(proposalID, voter, option, salt2)

	require.NotEqual(t, hash1, hash2)
}

func TestCalculateVoteHash_Format(t *testing.T) {
	keeper, _ := setupVotePrivacyKeeper(t)

	proposalID := uint64(1)
	voter := testAddr("voter1")
	option := types.VoteOptionYes
	salt := "randomsalt123"

	hash := keeper.calculateVoteHash(proposalID, voter, option, salt)

	// Hash should be hex string of appropriate length (SHA-256 = 64 hex chars)
	require.NotEmpty(t, hash)
	require.Len(t, hash, 64)
}

func TestSetVoteCommitment_Success(t *testing.T) {
	keeper, ctx := setupVotePrivacyKeeper(t)

	now := time.Now()
	committedAt, _ := gogotypes.TimestampProto(now)

	commitment := &types.VoteCommitment{
		ProposalId:  1,
		Voter:       testAddr("voter1"),
		VoteHash:    "abcdef1234567890",
		CommittedAt: committedAt,
		Revealed:    false,
	}

	err := keeper.setVoteCommitment(ctx, commitment)
	require.NoError(t, err)

	// Verify it was stored
	retrieved, err := keeper.getVoteCommitment(ctx, 1, testAddr("voter1"))
	require.NoError(t, err)
	require.Equal(t, commitment.ProposalId, retrieved.ProposalId)
	require.Equal(t, commitment.Voter, retrieved.Voter)
	require.Equal(t, commitment.VoteHash, retrieved.VoteHash)
}

func TestSetVoteCommitment_UpdateExisting(t *testing.T) {
	keeper, ctx := setupVotePrivacyKeeper(t)

	now := time.Now()
	committedAt, _ := gogotypes.TimestampProto(now)

	commitment := &types.VoteCommitment{
		ProposalId:  1,
		Voter:       testAddr("voter1"),
		VoteHash:    "hash1",
		CommittedAt: committedAt,
		Revealed:    false,
	}

	err := keeper.setVoteCommitment(ctx, commitment)
	require.NoError(t, err)

	// Update to revealed
	commitment.Revealed = true
	err = keeper.setVoteCommitment(ctx, commitment)
	require.NoError(t, err)

	// Verify update
	retrieved, err := keeper.getVoteCommitment(ctx, 1, testAddr("voter1"))
	require.NoError(t, err)
	require.True(t, retrieved.Revealed)
}

func TestGetVoteCommitment_NotFound(t *testing.T) {
	keeper, ctx := setupVotePrivacyKeeper(t)

	_, err := keeper.getVoteCommitment(ctx, 999, testAddr("nonexistent"))
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrNoVoteCommitment)
}

func TestGetVoteCommitment_MultipleVoters(t *testing.T) {
	keeper, ctx := setupVotePrivacyKeeper(t)

	now := time.Now()
	committedAt, _ := gogotypes.TimestampProto(now)

	voter1 := testAddr("voter1")
	voter2 := testAddr("voter2")
	proposalID := uint64(1)

	commitment1 := &types.VoteCommitment{
		ProposalId:  proposalID,
		Voter:       voter1,
		VoteHash:    "hash1",
		CommittedAt: committedAt,
		Revealed:    false,
	}

	commitment2 := &types.VoteCommitment{
		ProposalId:  proposalID,
		Voter:       voter2,
		VoteHash:    "hash2",
		CommittedAt: committedAt,
		Revealed:    false,
	}

	err := keeper.setVoteCommitment(ctx, commitment1)
	require.NoError(t, err)

	err = keeper.setVoteCommitment(ctx, commitment2)
	require.NoError(t, err)

	// Retrieve both commitments
	retrieved1, err := keeper.getVoteCommitment(ctx, proposalID, voter1)
	require.NoError(t, err)
	require.Equal(t, "hash1", retrieved1.VoteHash)

	retrieved2, err := keeper.getVoteCommitment(ctx, proposalID, voter2)
	require.NoError(t, err)
	require.Equal(t, "hash2", retrieved2.VoteHash)
}

func TestCommitRevealWorkflow_Complete(t *testing.T) {
	keeper, ctx := setupVotePrivacyKeeper(t)
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)

	proposalID := uint64(1)
	voter1 := testAddr("voter1")
	voter2 := testAddr("voter2")
	salt1 := "salt1"
	salt2 := "salt2"

	mockStaking.SetDelegatorBonded(voter1, sdkmath.NewInt(1000000))
	mockStaking.SetDelegatorBonded(voter2, sdkmath.NewInt(2000000))

	createPrivacyVotingProposal(t, keeper, ctx, proposalID)

	// Voter 1 commits Yes vote
	hash1 := keeper.calculateVoteHash(proposalID, voter1, types.VoteOptionYes, salt1)
	err := keeper.CommitVote(ctx, proposalID, voter1, hash1)
	require.NoError(t, err)

	// Voter 2 commits No vote
	hash2 := keeper.calculateVoteHash(proposalID, voter2, types.VoteOptionNo, salt2)
	err = keeper.CommitVote(ctx, proposalID, voter2, hash2)
	require.NoError(t, err)

	// Verify both commitments exist
	commitment1, err := keeper.getVoteCommitment(ctx, proposalID, voter1)
	require.NoError(t, err)
	require.False(t, commitment1.Revealed)

	commitment2, err := keeper.getVoteCommitment(ctx, proposalID, voter2)
	require.NoError(t, err)
	require.False(t, commitment2.Revealed)

	// Reveal both votes
	err = keeper.RevealVote(ctx, proposalID, voter1, types.VoteOptionYes, salt1)
	require.NoError(t, err)

	err = keeper.RevealVote(ctx, proposalID, voter2, types.VoteOptionNo, salt2)
	require.NoError(t, err)

	// Verify both votes were recorded
	vote1, err := keeper.GetVote(ctx, proposalID, voter1)
	require.NoError(t, err)
	require.Equal(t, types.VoteOptionYes, vote1.Option)
	require.True(t, vote1.IsSecret)

	vote2, err := keeper.GetVote(ctx, proposalID, voter2)
	require.NoError(t, err)
	require.Equal(t, types.VoteOptionNo, vote2.Option)
	require.True(t, vote2.IsSecret)

	// Verify commitments are marked as revealed
	commitment1, err = keeper.getVoteCommitment(ctx, proposalID, voter1)
	require.NoError(t, err)
	require.True(t, commitment1.Revealed)

	commitment2, err = keeper.getVoteCommitment(ctx, proposalID, voter2)
	require.NoError(t, err)
	require.True(t, commitment2.Revealed)
}

func TestSecretVote_VotingPowerAtReveal(t *testing.T) {
	keeper, ctx := setupVotePrivacyKeeper(t)
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)

	proposalID := uint64(1)
	voter := testAddr("voter1")
	option := types.VoteOptionYes
	salt := "randomsalt123"

	// Set initial voting power
	initialPower := sdkmath.NewInt(1000000)
	mockStaking.SetDelegatorBonded(voter, initialPower)

	createPrivacyVotingProposal(t, keeper, ctx, proposalID)

	// Commit vote
	voteHash := keeper.calculateVoteHash(proposalID, voter, option, salt)
	err := keeper.CommitVote(ctx, proposalID, voter, voteHash)
	require.NoError(t, err)

	// Change voting power before reveal
	newPower := sdkmath.NewInt(2000000)
	mockStaking.SetDelegatorBonded(voter, newPower)

	// Reveal vote - should use current voting power
	err = keeper.RevealVote(ctx, proposalID, voter, option, salt)
	require.NoError(t, err)

	// Verify vote has current voting power
	vote, err := keeper.GetVote(ctx, proposalID, voter)
	require.NoError(t, err)
	require.Equal(t, newPower.String(), vote.VotingPower)
}

func TestCommitVote_MultipleProposals(t *testing.T) {
	keeper, ctx := setupVotePrivacyKeeper(t)

	voter := testAddr("voter1")

	// Create multiple proposals
	createPrivacyVotingProposal(t, keeper, ctx, 1)
	createPrivacyVotingProposal(t, keeper, ctx, 2)

	// Commit votes on both proposals
	hash1 := keeper.calculateVoteHash(1, voter, types.VoteOptionYes, "salt1")
	err := keeper.CommitVote(ctx, 1, voter, hash1)
	require.NoError(t, err)

	hash2 := keeper.calculateVoteHash(2, voter, types.VoteOptionNo, "salt2")
	err = keeper.CommitVote(ctx, 2, voter, hash2)
	require.NoError(t, err)

	// Verify both commitments exist independently
	commitment1, err := keeper.getVoteCommitment(ctx, 1, voter)
	require.NoError(t, err)
	require.Equal(t, uint64(1), commitment1.ProposalId)

	commitment2, err := keeper.getVoteCommitment(ctx, 2, voter)
	require.NoError(t, err)
	require.Equal(t, uint64(2), commitment2.ProposalId)
}

func TestVoteCommitment_TimestampRecorded(t *testing.T) {
	keeper, ctx := setupVotePrivacyKeeper(t)

	proposalID := uint64(1)
	voter := testAddr("voter1")
	voteHash := "abcdef1234567890"

	createPrivacyVotingProposal(t, keeper, ctx, proposalID)

	err := keeper.CommitVote(ctx, proposalID, voter, voteHash)
	require.NoError(t, err)

	// Verify timestamp is recorded
	commitment, err := keeper.getVoteCommitment(ctx, proposalID, voter)
	require.NoError(t, err)
	require.NotNil(t, commitment.CommittedAt, "CommittedAt should be set")

	// After JSON unmarshal, gogotypes.Timestamp becomes map[string]interface{}
	// Verify the timestamp has a valid seconds field
	switch ts := commitment.CommittedAt.(type) {
	case *gogotypes.Timestamp:
		require.NotZero(t, ts.Seconds, "Timestamp seconds should be non-zero")
	case map[string]interface{}:
		// JSON unmarshaled format (check both cases for key name)
		seconds, ok := ts["Seconds"].(float64)
		if !ok {
			seconds, ok = ts["seconds"].(float64)
		}
		require.True(t, ok, "Seconds field should be present and numeric")
		require.GreaterOrEqual(t, seconds, float64(0), "Timestamp should be non-negative")
	default:
		require.Fail(t, "CommittedAt should be a Timestamp or map, got %T", commitment.CommittedAt)
	}
}

func TestCalculateVoteHash_AllVoteOptions(t *testing.T) {
	keeper, _ := setupVotePrivacyKeeper(t)

	proposalID := uint64(1)
	voter := testAddr("voter1")
	salt := "salt"

	options := []types.VoteOption{
		types.VoteOptionYes,
		types.VoteOptionNo,
		types.VoteOptionAbstain,
		types.VoteOptionNoWithVeto,
	}

	hashes := make([]string, len(options))
	for i, option := range options {
		hashes[i] = keeper.calculateVoteHash(proposalID, voter, option, salt)
		require.NotEmpty(t, hashes[i])
	}

	// All hashes should be different
	for i := 0; i < len(hashes); i++ {
		for j := i + 1; j < len(hashes); j++ {
			require.NotEqual(t, hashes[i], hashes[j],
				"Hash for option %v should differ from option %v", options[i], options[j])
		}
	}
}

func TestRevealVote_NoVotingPower(t *testing.T) {
	keeper, ctx := setupVotePrivacyKeeper(t)

	proposalID := uint64(1)
	voter := testAddr("voter1")
	option := types.VoteOptionYes
	salt := "randomsalt123"

	// Don't set any voting power for the voter

	createPrivacyVotingProposal(t, keeper, ctx, proposalID)

	// Commit vote
	voteHash := keeper.calculateVoteHash(proposalID, voter, option, salt)
	err := keeper.CommitVote(ctx, proposalID, voter, voteHash)
	require.NoError(t, err)

	// Reveal vote - should succeed even with 0 voting power
	err = keeper.RevealVote(ctx, proposalID, voter, option, salt)
	require.NoError(t, err)

	// Verify vote was recorded with 0 power
	vote, err := keeper.GetVote(ctx, proposalID, voter)
	require.NoError(t, err)
	require.Equal(t, "0", vote.VotingPower)
}
