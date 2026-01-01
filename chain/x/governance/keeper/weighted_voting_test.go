//go:build governance_extra
// +build governance_extra

// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"math/big"
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

func setupWeightedVotingKeeper(t *testing.T) (*Keeper, sdk.Context) {
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

// testAddrWeighted generates a valid bech32 address for testing
func testAddrWeighted(name string) string {
	padded := name + "________________"
	return sdk.AccAddress(padded[:20]).String()
}

// TestCastWeightedVote tests the basic weighted voting functionality
func TestCastWeightedVote(t *testing.T) {
	keeper, ctx := setupWeightedVotingKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	endTs, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))

	// Create proposal in voting period
	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Weighted Vote Proposal",
		Description:   "Test Description",
		Proposer:      testAddrWeighted("proposer1"),
		Status:        types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		VotingEndTime: endTs,
	}
	keeper.SetProposal(ctx, proposal)

	voter := testAddrWeighted("voter1")

	// Set voting power for voter
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)
	mockStaking.SetDelegatorBonded(voter, sdkmath.NewInt(1000000))

	// Create weighted vote options (50% Yes, 30% No, 20% Abstain)
	options := []*types.WeightedVoteOption{
		{Option: types.VoteOption_VOTE_OPTION_YES, Weight: "5000"},
		{Option: types.VoteOption_VOTE_OPTION_NO, Weight: "3000"},
		{Option: types.VoteOption_VOTE_OPTION_ABSTAIN, Weight: "2000"},
	}

	// Cast weighted vote
	err := keeper.CastWeightedVote(ctx, 1, voter, options)
	require.NoError(t, err)

	// Verify vote was stored
	vote, err := keeper.GetVote(ctx, 1, voter)
	require.NoError(t, err)
	require.NotNil(t, vote)
	require.Equal(t, uint64(1), vote.ProposalId)
	require.Equal(t, voter, vote.Voter)
	require.Equal(t, "1000000", vote.VotingPower)
	require.Equal(t, types.VoteOption_VOTE_OPTION_UNSPECIFIED, vote.Option) // Weighted votes use UNSPECIFIED

	// Verify event was emitted
	events := ctx.EventManager().Events()
	foundEvent := false
	for _, event := range events {
		if event.Type == "weighted_vote_cast" {
			foundEvent = true
			break
		}
	}
	require.True(t, foundEvent, "weighted_vote_cast event should be emitted")
}

// TestCastWeightedVote_InvalidStatus tests voting on non-voting period proposals
func TestCastWeightedVote_InvalidStatus(t *testing.T) {
	keeper, ctx := setupWeightedVotingKeeper(t)

	ts, _ := gogotypes.TimestampProto(time.Now())

	// Create proposal in deposit period (not voting)
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddrWeighted("proposer1"),
		Status:      types.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal)

	voter := testAddrWeighted("voter1")
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)
	mockStaking.SetDelegatorBonded(voter, sdkmath.NewInt(1000000))

	options := []*types.WeightedVoteOption{
		{Option: types.VoteOption_VOTE_OPTION_YES, Weight: "10000"},
	}

	// Should fail - proposal not in voting period
	err := keeper.CastWeightedVote(ctx, 1, voter, options)
	require.Error(t, err)
	require.Equal(t, types.ErrInvalidProposalStatus, err)
}

// TestCastWeightedVote_ProposalNotFound tests voting on non-existent proposal
func TestCastWeightedVote_ProposalNotFound(t *testing.T) {
	keeper, ctx := setupWeightedVotingKeeper(t)

	voter := testAddrWeighted("voter1")
	options := []*types.WeightedVoteOption{
		{Option: types.VoteOption_VOTE_OPTION_YES, Weight: "10000"},
	}

	// Should fail - proposal doesn't exist
	err := keeper.CastWeightedVote(ctx, 999, voter, options)
	require.Error(t, err)
}

// TestValidateWeightedOptions tests the weighted options validation
func TestValidateWeightedOptions(t *testing.T) {
	keeper, _ := setupWeightedVotingKeeper(t)

	tests := []struct {
		name      string
		options   []*types.WeightedVoteOption
		expectErr bool
	}{
		{
			name: "valid single option",
			options: []*types.WeightedVoteOption{
				{Option: types.VoteOption_VOTE_OPTION_YES, Weight: "10000"},
			},
			expectErr: false,
		},
		{
			name: "valid multiple options summing to 100%",
			options: []*types.WeightedVoteOption{
				{Option: types.VoteOption_VOTE_OPTION_YES, Weight: "6000"},
				{Option: types.VoteOption_VOTE_OPTION_NO, Weight: "3000"},
				{Option: types.VoteOption_VOTE_OPTION_ABSTAIN, Weight: "1000"},
			},
			expectErr: false,
		},
		{
			name:      "empty options",
			options:   []*types.WeightedVoteOption{},
			expectErr: true,
		},
		{
			name: "weights sum to less than 100%",
			options: []*types.WeightedVoteOption{
				{Option: types.VoteOption_VOTE_OPTION_YES, Weight: "5000"},
				{Option: types.VoteOption_VOTE_OPTION_NO, Weight: "3000"},
			},
			expectErr: true,
		},
		{
			name: "weights sum to more than 100%",
			options: []*types.WeightedVoteOption{
				{Option: types.VoteOption_VOTE_OPTION_YES, Weight: "7000"},
				{Option: types.VoteOption_VOTE_OPTION_NO, Weight: "5000"},
			},
			expectErr: true,
		},
		{
			name: "individual weight exceeds 100%",
			options: []*types.WeightedVoteOption{
				{Option: types.VoteOption_VOTE_OPTION_YES, Weight: "15000"},
			},
			expectErr: true,
		},
		{
			name: "invalid weight format",
			options: []*types.WeightedVoteOption{
				{Option: types.VoteOption_VOTE_OPTION_YES, Weight: "invalid"},
			},
			expectErr: true,
		},
		{
			name: "negative weight",
			options: []*types.WeightedVoteOption{
				{Option: types.VoteOption_VOTE_OPTION_YES, Weight: "-1000"},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := keeper.validateWeightedOptions(tt.options)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestCalculateWeightedTally tests the weighted vote tallying
func TestCalculateWeightedTally(t *testing.T) {
	keeper, ctx := setupWeightedVotingKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	endTs, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))

	// Create proposal
	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddrWeighted("proposer1"),
		Status:        types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		VotingEndTime: endTs,
	}
	keeper.SetProposal(ctx, proposal)

	// Create votes with different options and power
	votes := []struct {
		voter       string
		option      types.VoteOption
		votingPower string
	}{
		{"voter1", types.VoteOption_VOTE_OPTION_YES, "1000000"},
		{"voter2", types.VoteOption_VOTE_OPTION_YES, "2000000"},
		{"voter3", types.VoteOption_VOTE_OPTION_NO, "500000"},
		{"voter4", types.VoteOption_VOTE_OPTION_ABSTAIN, "300000"},
		{"voter5", types.VoteOption_VOTE_OPTION_NO_WITH_VETO, "100000"},
	}

	for _, v := range votes {
		vote := &types.Vote{
			ProposalId:  1,
			Voter:       testAddrWeighted(v.voter),
			Option:      v.option,
			VotingPower: v.votingPower,
			Timestamp:   ts,
		}
		keeper.SetVote(ctx, vote)
	}

	// Calculate tally
	tally := keeper.CalculateWeightedTally(ctx, 1)
	require.NotNil(t, tally)

	// Verify tally results
	yesVotes := new(big.Int)
	yesVotes.SetString(tally.Yes, 10)
	require.Equal(t, int64(3000000), yesVotes.Int64())

	noVotes := new(big.Int)
	noVotes.SetString(tally.No, 10)
	require.Equal(t, int64(500000), noVotes.Int64())

	abstainVotes := new(big.Int)
	abstainVotes.SetString(tally.Abstain, 10)
	require.Equal(t, int64(300000), abstainVotes.Int64())

	vetoVotes := new(big.Int)
	vetoVotes.SetString(tally.NoWithVeto, 10)
	require.Equal(t, int64(100000), vetoVotes.Int64())
}

// TestCalculateWeightedTally_NoVotes tests tallying with no votes
func TestCalculateWeightedTally_NoVotes(t *testing.T) {
	keeper, ctx := setupWeightedVotingKeeper(t)

	ts, _ := gogotypes.TimestampProto(time.Now())

	// Create proposal
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddrWeighted("proposer1"),
		Status:      types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal)

	// Calculate tally with no votes
	tally := keeper.CalculateWeightedTally(ctx, 1)
	require.NotNil(t, tally)

	// All tallies should be zero
	require.Equal(t, "0", tally.Yes)
	require.Equal(t, "0", tally.No)
	require.Equal(t, "0", tally.Abstain)
	require.Equal(t, "0", tally.NoWithVeto)
}

// TestGetVotingPowerBreakdown tests the voting power breakdown functionality
func TestGetVotingPowerBreakdown(t *testing.T) {
	keeper, ctx := setupWeightedVotingKeeper(t)

	voter := testAddrWeighted("voter1")

	// Set up staking power
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)
	mockStaking.SetDelegatorBonded(voter, sdkmath.NewInt(5000000))

	// Create a vote delegation to this voter
	delegator := testAddrWeighted("delegator1")
	delegation := &types.VoteDelegation{
		Delegator:      delegator,
		Delegate:       voter,
		DelegatedPower: "2000000",
	}
	keeper.SetVoteDelegation(ctx, delegation)

	// Create a token lock
	now := time.Now()
	lockTime, _ := gogotypes.TimestampProto(now)
	unlockTime, _ := gogotypes.TimestampProto(now.Add(24 * time.Hour))

	lock := &types.TokenLock{
		Owner:        voter,
		ProposalId:   1,
		LockedAmount: "1000000",
		LockTime:     lockTime,
		UnlockTime:   unlockTime,
	}
	keeper.SetTokenLock(ctx, lock)

	// Get breakdown
	breakdown := keeper.GetVotingPowerBreakdown(ctx, voter)
	require.NotNil(t, breakdown)
	require.Equal(t, voter, breakdown.Address)
	require.Equal(t, "5000000", breakdown.BaseVotingPower)
	require.Equal(t, "2000000", breakdown.DelegatedPower)
	require.Equal(t, "1000000", breakdown.LockedTokenPower)
	require.Equal(t, uint64(1), breakdown.ActiveDelegations)
	require.Equal(t, uint64(10000), breakdown.VotingPowerMultiplier)
}

// TestGetVotingPowerBreakdown_NoVotingPower tests breakdown with no voting power
func TestGetVotingPowerBreakdown_NoVotingPower(t *testing.T) {
	keeper, ctx := setupWeightedVotingKeeper(t)

	voter := testAddrWeighted("voter_no_power")

	// Get breakdown for voter with no power
	breakdown := keeper.GetVotingPowerBreakdown(ctx, voter)
	require.NotNil(t, breakdown)
	require.Equal(t, voter, breakdown.Address)
	require.Equal(t, "0", breakdown.BaseVotingPower)
	require.Equal(t, "0", breakdown.DelegatedPower)
	require.Equal(t, "0", breakdown.LockedTokenPower)
	require.Equal(t, "0", breakdown.TotalVotingPower)
	require.Equal(t, uint64(0), breakdown.ActiveDelegations)
}

// TestGetVotingPowerBreakdown_MultipleDelegations tests breakdown with multiple delegations
func TestGetVotingPowerBreakdown_MultipleDelegations(t *testing.T) {
	keeper, ctx := setupWeightedVotingKeeper(t)

	voter := testAddrWeighted("voter_multi")

	// Set up staking power
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)
	mockStaking.SetDelegatorBonded(voter, sdkmath.NewInt(3000000))

	// Create multiple delegations to this voter
	for i := 1; i <= 3; i++ {
		delegator := testAddrWeighted("delegator" + string(rune('0'+i)))
		delegation := &types.VoteDelegation{
			Delegator:      delegator,
			Delegate:       voter,
			DelegatedPower: "1000000",
		}
		keeper.SetVoteDelegation(ctx, delegation)
	}

	// Get breakdown
	breakdown := keeper.GetVotingPowerBreakdown(ctx, voter)
	require.NotNil(t, breakdown)
	require.Equal(t, uint64(3), breakdown.ActiveDelegations)
	require.Equal(t, "3000000", breakdown.DelegatedPower)
}

// TestGetVotingPowerBreakdown_MultipleTokenLocks tests breakdown with multiple token locks
func TestGetVotingPowerBreakdown_MultipleTokenLocks(t *testing.T) {
	keeper, ctx := setupWeightedVotingKeeper(t)

	voter := testAddrWeighted("voter_locks")

	// Set up staking power
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)
	mockStaking.SetDelegatorBonded(voter, sdkmath.NewInt(5000000))

	// Create multiple token locks
	now := time.Now()
	for i := 1; i <= 3; i++ {
		lockTime, _ := gogotypes.TimestampProto(now)
		unlockTime, _ := gogotypes.TimestampProto(now.Add(24 * time.Hour))

		lock := &types.TokenLock{
			Owner:        voter,
			ProposalId:   uint64(i),
			LockedAmount: "500000",
			LockTime:     lockTime,
			UnlockTime:   unlockTime,
		}
		keeper.SetTokenLock(ctx, lock)
	}

	// Get breakdown
	breakdown := keeper.GetVotingPowerBreakdown(ctx, voter)
	require.NotNil(t, breakdown)
	// Total locked should be 3 * 500000 = 1500000
	require.Equal(t, "1500000", breakdown.LockedTokenPower)
}

// TestGetVotingPowerBreakdown_InvalidAddress tests breakdown with invalid address
func TestGetVotingPowerBreakdown_InvalidAddress(t *testing.T) {
	keeper, ctx := setupWeightedVotingKeeper(t)

	// Invalid address
	invalidAddr := "invalid_address"

	// Should not panic and return valid breakdown
	breakdown := keeper.GetVotingPowerBreakdown(ctx, invalidAddr)
	require.NotNil(t, breakdown)
	require.Equal(t, invalidAddr, breakdown.Address)
	require.Equal(t, "0", breakdown.BaseVotingPower)
	require.Equal(t, "0", breakdown.TotalVotingPower)
}

// TestWeightedVoting_Integration tests a complete weighted voting scenario
func TestWeightedVoting_Integration(t *testing.T) {
	keeper, ctx := setupWeightedVotingKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	endTs, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))

	// Create proposal
	proposal := &types.Proposal{
		Id:            1,
		Title:         "Integration Test Proposal",
		Description:   "Testing weighted voting integration",
		Proposer:      testAddrWeighted("proposer1"),
		Status:        types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		VotingEndTime: endTs,
	}
	keeper.SetProposal(ctx, proposal)

	// Setup voters with different voting powers
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)
	voters := []struct {
		name  string
		power int64
	}{
		{"voter1", 10000000},
		{"voter2", 5000000},
		{"voter3", 3000000},
	}

	for _, v := range voters {
		addr := testAddrWeighted(v.name)
		mockStaking.SetDelegatorBonded(addr, sdkmath.NewInt(v.power))
	}

	// Cast weighted votes
	// Voter1: 70% Yes, 30% No
	err := keeper.CastWeightedVote(ctx, 1, testAddrWeighted("voter1"), []*types.WeightedVoteOption{
		{Option: types.VoteOption_VOTE_OPTION_YES, Weight: "7000"},
		{Option: types.VoteOption_VOTE_OPTION_NO, Weight: "3000"},
	})
	require.NoError(t, err)

	// Voter2: 100% Abstain
	err = keeper.CastWeightedVote(ctx, 1, testAddrWeighted("voter2"), []*types.WeightedVoteOption{
		{Option: types.VoteOption_VOTE_OPTION_ABSTAIN, Weight: "10000"},
	})
	require.NoError(t, err)

	// Voter3: 50% Yes, 50% Veto
	err = keeper.CastWeightedVote(ctx, 1, testAddrWeighted("voter3"), []*types.WeightedVoteOption{
		{Option: types.VoteOption_VOTE_OPTION_YES, Weight: "5000"},
		{Option: types.VoteOption_VOTE_OPTION_NO_WITH_VETO, Weight: "5000"},
	})
	require.NoError(t, err)

	// Verify all votes were stored
	for _, v := range voters {
		vote, err := keeper.GetVote(ctx, 1, testAddrWeighted(v.name))
		require.NoError(t, err)
		require.NotNil(t, vote)
	}

	// Get voting power breakdown for voter1
	breakdown := keeper.GetVotingPowerBreakdown(ctx, testAddrWeighted("voter1"))
	require.NotNil(t, breakdown)
	require.Equal(t, "10000000", breakdown.BaseVotingPower)
}
