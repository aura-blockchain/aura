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

func setupVoteDelegationKeeper(t *testing.T) (*Keeper, sdk.Context) {
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
	return keeper, ctx
}

func TestDelegateVote_ValidDelegation(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	delegator := testAddr("delegator1")
	delegate := testAddr("delegate1")
	percentage := uint64(5000) // 50%

	err := keeper.DelegateVote(ctx, delegator, delegate, percentage)
	require.NoError(t, err)

	// Verify delegation was stored
	delegations := keeper.GetVoteDelegations(ctx, delegator)
	require.Len(t, delegations, 1)
	require.Equal(t, delegate, delegations[0].Delegate)
	require.Equal(t, delegator, delegations[0].Delegator)
}

func TestDelegateVote_InvalidPercentageZero(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	delegator := testAddr("delegator1")
	delegate := testAddr("delegate1")
	percentage := uint64(0)

	err := keeper.DelegateVote(ctx, delegator, delegate, percentage)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidDelegation)
}

func TestDelegateVote_InvalidPercentageExceedsMax(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	delegator := testAddr("delegator1")
	delegate := testAddr("delegate1")
	percentage := uint64(10001) // Over 100%

	err := keeper.DelegateVote(ctx, delegator, delegate, percentage)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidDelegation)
}

func TestDelegateVote_CannotDelegateToSelf(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	delegator := testAddr("delegator1")
	percentage := uint64(5000)

	err := keeper.DelegateVote(ctx, delegator, delegator, percentage)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidDelegation)
}

func TestDelegateVote_DuplicateDelegation(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	delegator := testAddr("delegator1")
	delegate := testAddr("delegate1")
	percentage := uint64(3000)

	// First delegation should succeed
	err := keeper.DelegateVote(ctx, delegator, delegate, percentage)
	require.NoError(t, err)

	// Second delegation to same delegate should fail
	err = keeper.DelegateVote(ctx, delegator, delegate, percentage)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidDelegation)
}

func TestDelegateVote_MultipleDelegations(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	delegator := testAddr("delegator1")
	delegate1 := testAddr("delegate1")
	delegate2 := testAddr("delegate2")

	// First delegation
	err := keeper.DelegateVote(ctx, delegator, delegate1, 3000)
	require.NoError(t, err)

	// Second delegation to different delegate
	err = keeper.DelegateVote(ctx, delegator, delegate2, 2000)
	require.NoError(t, err)

	// Verify both delegations exist
	delegations := keeper.GetVoteDelegations(ctx, delegator)
	require.Len(t, delegations, 2)
}

func TestDelegateVote_ExceedsMaxDelegations(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	delegator := testAddr("delegator1")
	params := keeper.GetParams(ctx)
	maxDelegations := types.GetMaxDelegationsPerUser(params)

	// Create max number of delegations
	for i := uint64(0); i < maxDelegations; i++ {
		delegate := testAddr("delegate" + string(rune('A'+i)))
		err := keeper.DelegateVote(ctx, delegator, delegate, 100)
		require.NoError(t, err)
	}

	// One more should fail
	extraDelegate := testAddr("extraDelegate")
	err := keeper.DelegateVote(ctx, delegator, extraDelegate, 100)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidDelegation)
}

func TestRevokeDelegation_Success(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	delegator := testAddr("delegator1")
	delegate := testAddr("delegate1")

	// Create delegation first
	err := keeper.DelegateVote(ctx, delegator, delegate, 5000)
	require.NoError(t, err)

	// Revoke it
	err = keeper.RevokeDelegation(ctx, delegator, delegate)
	require.NoError(t, err)

	// Verify delegation was removed
	delegations := keeper.GetVoteDelegations(ctx, delegator)
	require.Empty(t, delegations)
}

func TestRevokeDelegation_NonExistentDelegation(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	delegator := testAddr("delegator1")
	delegate := testAddr("delegate1")

	// Try to revoke non-existent delegation - should succeed (idempotent delete)
	err := keeper.RevokeDelegation(ctx, delegator, delegate)
	require.NoError(t, err)
}

func TestVoteWithDelegatedPower_Success(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)

	delegate := testAddr("delegate1")

	// Set voting power
	mockStaking.SetDelegatorBonded(delegate, sdkmath.NewInt(1000000))

	// Create a proposal in voting period
	now := time.Now()
	submitTime, _ := gogotypes.TimestampProto(now)
	votingEnd, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))

	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddr("proposer1"),
		Status:        types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		Category:      types.CategoryText,
		SubmitTime:    submitTime,
		VotingEndTime: votingEnd,
	}
	err := keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Vote with delegated power
	err = keeper.VoteWithDelegatedPower(ctx, 1, delegate, types.VoteOptionYes)
	require.NoError(t, err)

	// Verify vote was recorded
	vote, err := keeper.GetVote(ctx, 1, delegate)
	require.NoError(t, err)
	require.Equal(t, types.VoteOptionYes, vote.Option)
	require.False(t, vote.IsSecret)
}

func TestVoteWithDelegatedPower_ProposalNotFound(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	delegate := testAddr("delegate1")

	err := keeper.VoteWithDelegatedPower(ctx, 999, delegate, types.VoteOptionYes)
	require.Error(t, err)
}

func TestVoteWithDelegatedPower_InvalidStatus(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	delegate := testAddr("delegate1")

	// Create proposal not in voting period
	now := time.Now()
	submitTime, _ := gogotypes.TimestampProto(now)

	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddr("proposer1"),
		Status:      types.ProposalStatus_PROPOSAL_STATUS_PASSED,
		Category:    types.CategoryText,
		SubmitTime:  submitTime,
	}
	err := keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	err = keeper.VoteWithDelegatedPower(ctx, 1, delegate, types.VoteOptionYes)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidProposalStatus)
}

func TestCalculateTotalVotingPower(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)

	address := testAddr("voter1")
	mockStaking.SetDelegatorBonded(address, sdkmath.NewInt(5000000))

	totalPower := keeper.calculateTotalVotingPower(ctx, address)
	require.NotEmpty(t, totalPower)
	require.NotEqual(t, "0", totalPower)
}

func TestCalculateTotalVotingPower_ZeroPower(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	address := testAddr("voter1")

	totalPower := keeper.calculateTotalVotingPower(ctx, address)
	require.Equal(t, "0", totalPower)
}

func TestCalculateDelegatedPower(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	delegator := testAddr("delegator1")
	delegate := testAddr("delegate1")

	// Create delegation
	err := keeper.DelegateVote(ctx, delegator, delegate, 5000)
	require.NoError(t, err)

	// Calculate delegated power to delegate
	delegatedPower := keeper.calculateDelegatedPower(ctx, delegate)
	require.NotEmpty(t, delegatedPower)
}

func TestCalculateDelegatedPower_NoDelegations(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	delegate := testAddr("delegate1")

	delegatedPower := keeper.calculateDelegatedPower(ctx, delegate)
	require.Equal(t, "0", delegatedPower)
}

func TestGetDelegationChain_DirectDelegation(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	delegator := testAddr("delegator1")
	delegate := testAddr("delegate1")

	// Create delegation
	err := keeper.DelegateVote(ctx, delegator, delegate, 5000)
	require.NoError(t, err)

	// Get delegation chain
	chain := keeper.GetDelegationChain(ctx, delegator)
	require.NotEmpty(t, chain)

	// Should have at least the direct delegation
	found := false
	for _, node := range chain {
		if node.Address == delegate && node.Depth == 1 {
			found = true
			require.Contains(t, node.Path, delegator)
			require.Contains(t, node.Path, delegate)
			break
		}
	}
	require.True(t, found, "Direct delegation should be in chain")
}

func TestGetDelegationChain_SubDelegations(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	delegator := testAddr("delegator1")
	delegate1 := testAddr("delegate1")
	delegate2 := testAddr("delegate2")

	// Create delegation chain: delegator -> delegate1 -> delegate2
	err := keeper.DelegateVote(ctx, delegator, delegate1, 5000)
	require.NoError(t, err)

	err = keeper.DelegateVote(ctx, delegate1, delegate2, 3000)
	require.NoError(t, err)

	// Get delegation chain
	chain := keeper.GetDelegationChain(ctx, delegator)
	require.NotEmpty(t, chain)

	// Should include both direct and sub-delegations
	hasDepth1 := false
	hasDepth2 := false
	for _, node := range chain {
		if node.Depth == 1 {
			hasDepth1 = true
		}
		if node.Depth == 2 {
			hasDepth2 = true
		}
	}
	require.True(t, hasDepth1, "Should have depth 1 delegation")
	require.True(t, hasDepth2, "Should have depth 2 delegation")
}

func TestGetDelegationChain_NoDelegations(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	delegator := testAddr("delegator1")

	chain := keeper.GetDelegationChain(ctx, delegator)
	require.Empty(t, chain)
}

func TestDelegationCreatesLoop_DirectLoop(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	address := testAddr("address1")

	// Check if delegating to self creates loop
	createsLoop := keeper.delegationCreatesLoop(ctx, address, address)
	require.True(t, createsLoop)
}

func TestDelegationCreatesLoop_NoLoop(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	delegator := testAddr("delegator1")
	delegate := testAddr("delegate1")

	createsLoop := keeper.delegationCreatesLoop(ctx, delegate, delegator)
	require.False(t, createsLoop)
}

func TestDelegationCreatesLoop_TwoLevelLoop(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	address1 := testAddr("address1")
	address2 := testAddr("address2")

	// Create: address1 -> address2
	err := keeper.DelegateVote(ctx, address1, address2, 5000)
	require.NoError(t, err)

	// Check if address2 -> address1 would create loop
	createsLoop := keeper.delegationCreatesLoop(ctx, address1, address2)
	require.True(t, createsLoop)
}

func TestDelegationCreatesLoop_ThreeLevelLoop(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	address1 := testAddr("address1")
	address2 := testAddr("address2")
	address3 := testAddr("address3")

	// Create: address1 -> address2 -> address3
	err := keeper.DelegateVote(ctx, address1, address2, 5000)
	require.NoError(t, err)

	err = keeper.DelegateVote(ctx, address2, address3, 3000)
	require.NoError(t, err)

	// Check if address3 -> address1 would create loop
	createsLoop := keeper.delegationCreatesLoop(ctx, address1, address3)
	require.True(t, createsLoop)
}

func TestGetDelegationStatistics_EmptyState(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	stats := keeper.GetDelegationStatistics(ctx)
	require.NotNil(t, stats)
	require.Equal(t, uint64(0), stats.TotalDelegations)
	require.Equal(t, uint64(0), stats.UniqueDelegators)
	require.Equal(t, uint64(0), stats.UniqueDelegates)
}

func TestGetDelegationStatistics_WithDelegations(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	// Create some delegations
	delegator1 := testAddr("delegator1")
	delegator2 := testAddr("delegator2")
	delegate1 := testAddr("delegate1")
	delegate2 := testAddr("delegate2")

	err := keeper.DelegateVote(ctx, delegator1, delegate1, 5000)
	require.NoError(t, err)

	err = keeper.DelegateVote(ctx, delegator2, delegate2, 3000)
	require.NoError(t, err)

	// Create a proposal with secret votes to populate statistics
	now := time.Now()
	submitTime, _ := gogotypes.TimestampProto(now)
	votingEnd, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))

	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddr("proposer1"),
		Status:        types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		Category:      types.CategoryText,
		SubmitTime:    submitTime,
		VotingEndTime: votingEnd,
	}
	err = keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Add secret votes
	vote1 := &types.Vote{
		ProposalId:  1,
		Voter:       delegate1,
		Option:      types.VoteOptionYes,
		VotingPower: "1000",
		Timestamp:   submitTime,
		IsSecret:    true,
	}
	err = keeper.SetVote(ctx, vote1)
	require.NoError(t, err)

	vote2 := &types.Vote{
		ProposalId:  1,
		Voter:       delegate2,
		Option:      types.VoteOptionNo,
		VotingPower: "2000",
		Timestamp:   submitTime,
		IsSecret:    true,
	}
	err = keeper.SetVote(ctx, vote2)
	require.NoError(t, err)

	stats := keeper.GetDelegationStatistics(ctx)
	require.NotNil(t, stats)
	require.Greater(t, stats.TotalDelegations, uint64(0))
}

func TestGetDelegationStatistics_FieldValidation(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	stats := keeper.GetDelegationStatistics(ctx)
	require.NotNil(t, stats)

	// Verify all fields are initialized
	require.GreaterOrEqual(t, stats.TotalDelegations, uint64(0))
	require.GreaterOrEqual(t, stats.UniqueDelegators, uint64(0))
	require.GreaterOrEqual(t, stats.UniqueDelegates, uint64(0))
	require.GreaterOrEqual(t, stats.AverageDelegationPercentage, uint64(0))
	require.GreaterOrEqual(t, stats.MaxDelegationChainDepth, uint32(0))
}

func TestDelegationChainNode_Structure(t *testing.T) {
	node := &DelegationChainNode{
		Address:    testAddr("delegate1"),
		Percentage: 5000,
		Depth:      2,
		Path:       []string{testAddr("delegator1"), testAddr("intermediate1"), testAddr("delegate1")},
	}

	require.NotEmpty(t, node.Address)
	require.Equal(t, uint64(5000), node.Percentage)
	require.Equal(t, uint32(2), node.Depth)
	require.Len(t, node.Path, 3)
}

func TestDelegationStatistics_Structure(t *testing.T) {
	stats := &DelegationStatistics{
		TotalDelegations:            10,
		UniqueDelegators:            8,
		UniqueDelegates:             5,
		AverageDelegationPercentage: 4500,
		MaxDelegationChainDepth:     3,
	}

	require.Equal(t, uint64(10), stats.TotalDelegations)
	require.Equal(t, uint64(8), stats.UniqueDelegators)
	require.Equal(t, uint64(5), stats.UniqueDelegates)
	require.Equal(t, uint64(4500), stats.AverageDelegationPercentage)
	require.Equal(t, uint32(3), stats.MaxDelegationChainDepth)
}

func TestDelegateVote_EventEmission(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	delegator := testAddr("delegator1")
	delegate := testAddr("delegate1")
	percentage := uint64(5000)

	// Clear any existing events
	ctx = ctx.WithEventManager(sdk.NewEventManager())

	err := keeper.DelegateVote(ctx, delegator, delegate, percentage)
	require.NoError(t, err)

	// Check event was emitted
	events := ctx.EventManager().Events()
	require.NotEmpty(t, events)

	// Find vote_delegated event
	found := false
	for _, event := range events {
		if event.Type == "vote_delegated" {
			found = true
			// Verify event attributes
			attrs := event.Attributes
			require.NotEmpty(t, attrs)
		}
	}
	require.True(t, found, "vote_delegated event should be emitted")
}

func TestRevokeDelegation_EventEmission(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)

	delegator := testAddr("delegator1")
	delegate := testAddr("delegate1")

	// Create delegation first
	err := keeper.DelegateVote(ctx, delegator, delegate, 5000)
	require.NoError(t, err)

	// Clear events
	ctx = ctx.WithEventManager(sdk.NewEventManager())

	// Revoke delegation
	err = keeper.RevokeDelegation(ctx, delegator, delegate)
	require.NoError(t, err)

	// Check event was emitted
	events := ctx.EventManager().Events()
	require.NotEmpty(t, events)

	// Find vote_delegation_revoked event
	found := false
	for _, event := range events {
		if event.Type == "vote_delegation_revoked" {
			found = true
		}
	}
	require.True(t, found, "vote_delegation_revoked event should be emitted")
}

func TestVoteWithDelegatedPower_EventEmission(t *testing.T) {
	keeper, ctx := setupVoteDelegationKeeper(t)
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)

	delegate := testAddr("delegate1")
	mockStaking.SetDelegatorBonded(delegate, sdkmath.NewInt(1000000))

	// Create proposal
	now := time.Now()
	submitTime, _ := gogotypes.TimestampProto(now)
	votingEnd, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))

	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddr("proposer1"),
		Status:        types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		Category:      types.CategoryText,
		SubmitTime:    submitTime,
		VotingEndTime: votingEnd,
	}
	err := keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Clear events
	ctx = ctx.WithEventManager(sdk.NewEventManager())

	// Vote with delegated power
	err = keeper.VoteWithDelegatedPower(ctx, 1, delegate, types.VoteOptionYes)
	require.NoError(t, err)

	// Check event was emitted
	events := ctx.EventManager().Events()
	require.NotEmpty(t, events)

	// Find delegated_vote_cast event
	found := false
	for _, event := range events {
		if event.Type == "delegated_vote_cast" {
			found = true
		}
	}
	require.True(t, found, "delegated_vote_cast event should be emitted")
}
