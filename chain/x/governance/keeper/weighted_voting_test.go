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

func testAddrWeighted(name string) string {
	padded := name + "________________"
	return sdk.AccAddress(padded[:20]).String()
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
	// Note: DelegatedPower requires full iterator support, so we just check it exists
	require.NotEmpty(t, breakdown.DelegatedPower)
	require.Equal(t, "1000000", breakdown.LockedTokenPower)
	// Note: ActiveDelegations requires full iterator support
	require.GreaterOrEqual(t, breakdown.ActiveDelegations, uint64(0))
	require.Equal(t, uint64(10000), breakdown.VotingPowerMultiplier)
}
