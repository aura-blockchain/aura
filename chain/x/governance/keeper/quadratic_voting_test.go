//go:build governance_extra
// +build governance_extra

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

func setupQuadraticVotingKeeper(t *testing.T) (*Keeper, sdk.Context) {
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

// testAddrQuadratic generates a valid bech32 address for testing
func testAddrQuadratic(name string) string {
	padded := name + "________________"
	return sdk.AccAddress(padded[:20]).String()
}

// TestCastQuadraticVote tests basic quadratic voting functionality
func TestCastQuadraticVote(t *testing.T) {
	keeper, ctx := setupQuadraticVotingKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	endTs, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))

	// Create proposal in voting period
	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Quadratic Vote Proposal",
		Description:   "Test Description",
		Proposer:      testAddrQuadratic("proposer1"),
		Status:        types.StatusVotingPeriod,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		VotingEndTime: endTs,
	}
	keeper.SetProposal(ctx, proposal)

	voter := testAddrQuadratic("voter1")

	// Set voting power for voter (need tokens for credits)
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)
	votingPower, _ := sdkmath.NewIntFromString("10000000000000000000") // 10 tokens - enough for credits
	mockStaking.SetDelegatorBonded(voter, votingPower)

	// Cast quadratic vote with 100 credits
	voteCredits := uint64(100)
	err := keeper.CastQuadraticVote(ctx, 1, voter, types.VoteOption_VOTE_OPTION_YES, voteCredits)
	require.NoError(t, err)

	// Verify vote was stored
	vote, err := keeper.GetVote(ctx, 1, voter)
	require.NoError(t, err)
	require.NotNil(t, vote)
	require.Equal(t, uint64(1), vote.ProposalId)
	require.Equal(t, voter, vote.Voter)
	require.Equal(t, types.VoteOption_VOTE_OPTION_YES, vote.Option)

	// Verify vote commitment contains quadratic metadata
	require.NotEmpty(t, vote.VoteCommitment)
	require.Contains(t, vote.VoteCommitment, "quadratic:")

	// Verify event was emitted
	events := ctx.EventManager().Events()
	foundEvent := false
	for _, event := range events {
		if event.Type == "quadratic_vote_cast" {
			foundEvent = true
			break
		}
	}
	require.True(t, foundEvent, "quadratic_vote_cast event should be emitted")
}

// TestCastQuadraticVote_InvalidStatus tests voting on non-voting period proposals
func TestCastQuadraticVote_InvalidStatus(t *testing.T) {
	keeper, ctx := setupQuadraticVotingKeeper(t)

	ts, _ := gogotypes.TimestampProto(time.Now())

	// Create proposal in deposit period (not voting)
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddrQuadratic("proposer1"),
		Status:      types.StatusDepositPeriod,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal)

	voter := testAddrQuadratic("voter1")
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)
	votingPower, _ := sdkmath.NewIntFromString("100000000000000000")
	mockStaking.SetDelegatorBonded(voter, votingPower)

	// Should fail - proposal not in voting period
	err := keeper.CastQuadraticVote(ctx, 1, voter, types.VoteOption_VOTE_OPTION_YES, 100)
	require.Error(t, err)
	require.Equal(t, types.ErrInvalidProposalStatus, err)
}

// TestCastQuadraticVote_ProposalNotFound tests voting on non-existent proposal
func TestCastQuadraticVote_ProposalNotFound(t *testing.T) {
	keeper, ctx := setupQuadraticVotingKeeper(t)

	voter := testAddrQuadratic("voter1")
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)
	votingPower, _ := sdkmath.NewIntFromString("100000000000000000")
	mockStaking.SetDelegatorBonded(voter, votingPower)

	// Should fail - proposal doesn't exist
	err := keeper.CastQuadraticVote(ctx, 999, voter, types.VoteOption_VOTE_OPTION_YES, 100)
	require.Error(t, err)
}

// TestCastQuadraticVote_InsufficientCredits tests voting with insufficient credits
func TestCastQuadraticVote_InsufficientCredits(t *testing.T) {
	keeper, ctx := setupQuadraticVotingKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	endTs, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))

	// Create proposal
	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddrQuadratic("proposer1"),
		Status:        types.StatusVotingPeriod,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		VotingEndTime: endTs,
	}
	keeper.SetProposal(ctx, proposal)

	voter := testAddrQuadratic("voter1")

	// Set minimal voting power (not enough for requested credits)
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)
	mockStaking.SetDelegatorBonded(voter, sdkmath.NewInt(100))

	// Try to vote with more credits than available
	err := keeper.CastQuadraticVote(ctx, 1, voter, types.VoteOption_VOTE_OPTION_YES, 10000)
	require.Error(t, err)
	require.Equal(t, types.ErrInsufficientVoteCredits, err)
}

// TestCalculateQuadraticVotingPower tests the quadratic formula
func TestCalculateQuadraticVotingPower(t *testing.T) {
	keeper, _ := setupQuadraticVotingKeeper(t)

	tests := []struct {
		name     string
		credits  uint64
		expected uint64
	}{
		{
			name:     "zero credits",
			credits:  0,
			expected: 0,
		},
		{
			name:     "one credit",
			credits:  1,
			expected: 10000, // sqrt(1) * 10000 = 10000
		},
		{
			name:     "four credits",
			credits:  4,
			expected: 20000, // sqrt(4) * 10000 = 20000
		},
		{
			name:     "nine credits",
			credits:  9,
			expected: 30000, // sqrt(9) * 10000 = 30000
		},
		{
			name:     "sixteen credits",
			credits:  16,
			expected: 40000, // sqrt(16) * 10000 = 40000
		},
		{
			name:     "hundred credits",
			credits:  100,
			expected: 100000, // sqrt(100) * 10000 = 100000
		},
		{
			name:     "ten thousand credits",
			credits:  10000,
			expected: 1000000, // sqrt(10000) * 10000 = 1000000
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			power := keeper.calculateQuadraticVotingPower(tt.credits)
			require.Equal(t, tt.expected, power)
		})
	}
}

// TestIntSqrt tests the integer square root implementation
func TestIntSqrt(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected int64
	}{
		{
			name:     "sqrt of 0",
			input:    0,
			expected: 0,
		},
		{
			name:     "sqrt of 1",
			input:    1,
			expected: 1,
		},
		{
			name:     "sqrt of 4",
			input:    4,
			expected: 2,
		},
		{
			name:     "sqrt of 9",
			input:    9,
			expected: 3,
		},
		{
			name:     "sqrt of 16",
			input:    16,
			expected: 4,
		},
		{
			name:     "sqrt of 100",
			input:    100,
			expected: 10,
		},
		{
			name:     "sqrt of 144",
			input:    144,
			expected: 12,
		},
		{
			name:     "sqrt of 10000",
			input:    10000,
			expected: 100,
		},
		{
			name:     "sqrt of non-perfect square (15)",
			input:    15,
			expected: 3, // floor(sqrt(15)) = 3
		},
		{
			name:     "sqrt of non-perfect square (99)",
			input:    99,
			expected: 9, // floor(sqrt(99)) = 9
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := sdkmath.NewInt(tt.input)
			result := intSqrt(n)
			require.Equal(t, tt.expected, result.Int64())
		})
	}
}

// TestIntSqrt_Negative tests sqrt of negative numbers
func TestIntSqrt_Negative(t *testing.T) {
	n := sdkmath.NewInt(-100)
	result := intSqrt(n)
	require.Equal(t, int64(0), result.Int64())
}

// TestGetVoteCredits tests getting available vote credits for a voter
func TestGetVoteCredits(t *testing.T) {
	keeper, ctx := setupQuadraticVotingKeeper(t)

	voter := testAddrQuadratic("voter1")

	// Set voting power (1 token = 100 credits)
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)
	votingPower, _ := sdkmath.NewIntFromString("100000000000000000")
	mockStaking.SetDelegatorBonded(voter, votingPower) // 0.1 token

	// Get available credits
	credits := keeper.GetVoteCredits(ctx, voter)
	require.Greater(t, credits, uint64(0), "Should have vote credits from voting power")
}

// TestGetVoteCredits_NoVotingPower tests credits with no voting power
func TestGetVoteCredits_NoVotingPower(t *testing.T) {
	keeper, ctx := setupQuadraticVotingKeeper(t)

	voter := testAddrQuadratic("voter_no_power")

	// No voting power set - should return 0 credits
	credits := keeper.GetVoteCredits(ctx, voter)
	require.Equal(t, uint64(0), credits)
}

// TestGetSpentVoteCredits tests tracking spent credits
func TestGetSpentVoteCredits(t *testing.T) {
	keeper, ctx := setupQuadraticVotingKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	endTs, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))

	voter := testAddrQuadratic("voter1")

	// Set voting power
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)
	votingPower, _ := sdkmath.NewIntFromString("500000000000000000")
	mockStaking.SetDelegatorBonded(voter, votingPower) // 0.5 tokens

	// Create multiple proposals and vote on them
	for i := 1; i <= 3; i++ {
		proposal := &types.Proposal{
			Id:            uint64(i),
			Title:         "Test Proposal",
			Description:   "Test Description",
			Proposer:      testAddrQuadratic("proposer1"),
			Status:        types.StatusVotingPeriod,
			Category:      types.CategoryText,
			SubmitTime:    ts,
			VotingEndTime: endTs,
		}
		keeper.SetProposal(ctx, proposal)

		// Vote with 100 credits on each
		err := keeper.CastQuadraticVote(ctx, uint64(i), voter, types.VoteOption_VOTE_OPTION_YES, 100)
		require.NoError(t, err)
	}

	// Get spent credits
	spent := keeper.GetSpentVoteCredits(ctx, voter)
	require.Equal(t, uint64(300), spent, "Should have spent 300 credits (100 * 3)")
}

// TestGetSpentVoteCredits_NoVotes tests spent credits with no votes
func TestGetSpentVoteCredits_NoVotes(t *testing.T) {
	keeper, ctx := setupQuadraticVotingKeeper(t)

	voter := testAddrQuadratic("voter_no_votes")

	// No votes cast - should return 0
	spent := keeper.GetSpentVoteCredits(ctx, voter)
	require.Equal(t, uint64(0), spent)
}

// TestDeductVoteCredits tests credit deduction
func TestDeductVoteCredits(t *testing.T) {
	keeper, ctx := setupQuadraticVotingKeeper(t)

	voter := testAddrQuadratic("voter1")

	// Deduct credits (currently a no-op as credits are tracked via votes)
	err := keeper.DeductVoteCredits(ctx, voter, 100)
	require.NoError(t, err)
}

// TestCalculateQuadraticTally tests quadratic vote tallying
func TestCalculateQuadraticTally(t *testing.T) {
	keeper, ctx := setupQuadraticVotingKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	endTs, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))

	// Create proposal
	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddrQuadratic("proposer1"),
		Status:        types.StatusVotingPeriod,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		VotingEndTime: endTs,
	}
	keeper.SetProposal(ctx, proposal)

	// Set voting power for multiple voters
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)
	voters := []string{"voter1", "voter2", "voter3", "voter4"}
	for _, v := range voters {
		votingPower, _ := sdkmath.NewIntFromString("500000000000000000")
		mockStaking.SetDelegatorBonded(testAddrQuadratic(v), votingPower)
	}

	// Cast quadratic votes with different credits and options
	votes := []struct {
		voter   string
		option  types.VoteOption
		credits uint64
	}{
		{"voter1", types.VoteOption_VOTE_OPTION_YES, 100},     // sqrt(100) * 10000 = 100000 power
		{"voter2", types.VoteOption_VOTE_OPTION_YES, 400},     // sqrt(400) * 10000 = 200000 power
		{"voter3", types.VoteOption_VOTE_OPTION_NO, 100},      // sqrt(100) * 10000 = 100000 power
		{"voter4", types.VoteOption_VOTE_OPTION_ABSTAIN, 225}, // sqrt(225) * 10000 = 150000 power
	}

	for _, v := range votes {
		err := keeper.CastQuadraticVote(ctx, 1, testAddrQuadratic(v.voter), v.option, v.credits)
		require.NoError(t, err)
	}

	// Calculate tally
	tally := keeper.CalculateQuadraticTally(ctx, 1)
	require.NotNil(t, tally)

	// Verify results
	require.Equal(t, uint64(300000), tally.YesPower)     // 100000 + 200000
	require.Equal(t, uint64(100000), tally.NoPower)      // 100000
	require.Equal(t, uint64(150000), tally.AbstainPower) // 150000
	require.Equal(t, uint64(0), tally.VetoPower)
	require.Equal(t, uint64(825), tally.TotalCreditsSpent) // 100 + 400 + 100 + 225
	require.Equal(t, uint64(4), tally.UniqueVoters)
}

// TestCalculateQuadraticTally_NoVotes tests tallying with no votes
func TestCalculateQuadraticTally_NoVotes(t *testing.T) {
	keeper, ctx := setupQuadraticVotingKeeper(t)

	ts, _ := gogotypes.TimestampProto(time.Now())

	// Create proposal
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddrQuadratic("proposer1"),
		Status:      types.StatusVotingPeriod,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal)

	// Calculate tally with no votes
	tally := keeper.CalculateQuadraticTally(ctx, 1)
	require.NotNil(t, tally)

	// All tallies should be zero
	require.Equal(t, uint64(0), tally.YesPower)
	require.Equal(t, uint64(0), tally.NoPower)
	require.Equal(t, uint64(0), tally.AbstainPower)
	require.Equal(t, uint64(0), tally.VetoPower)
	require.Equal(t, uint64(0), tally.TotalCreditsSpent)
	require.Equal(t, uint64(0), tally.UniqueVoters)
}

// TestGetQuadraticVotingStats tests the quadratic voting statistics
func TestGetQuadraticVotingStats(t *testing.T) {
	keeper, ctx := setupQuadraticVotingKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	endTs, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))

	// Create proposal
	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddrQuadratic("proposer1"),
		Status:        types.StatusVotingPeriod,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		VotingEndTime: endTs,
	}
	keeper.SetProposal(ctx, proposal)

	// Set voting power
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)
	voters := []string{"voter1", "voter2", "voter3"}
	for _, v := range voters {
		votingPower, _ := sdkmath.NewIntFromString("500000000000000000")
		mockStaking.SetDelegatorBonded(testAddrQuadratic(v), votingPower)
	}

	// Cast votes
	votes := []struct {
		voter   string
		option  types.VoteOption
		credits uint64
	}{
		{"voter1", types.VoteOption_VOTE_OPTION_YES, 100},
		{"voter2", types.VoteOption_VOTE_OPTION_YES, 400},
		{"voter3", types.VoteOption_VOTE_OPTION_NO, 225},
	}

	for _, v := range votes {
		err := keeper.CastQuadraticVote(ctx, 1, testAddrQuadratic(v.voter), v.option, v.credits)
		require.NoError(t, err)
	}

	// Get stats
	stats := keeper.GetQuadraticVotingStats(ctx, 1)
	require.NotNil(t, stats)
	require.Equal(t, uint64(1), stats.ProposalId)
	require.Equal(t, uint64(3), stats.UniqueVoters)
	require.Equal(t, uint64(725), stats.TotalCreditsSpent)      // 100 + 400 + 225
	require.Equal(t, uint64(241), stats.AverageCreditsPerVoter) // 725 / 3 = 241
	require.Greater(t, stats.TotalVotingPower, uint64(0))
	// CostEfficiency and QuadraticAdvantage are now strings (sdk.Dec) for determinism
	// They should be non-zero (not "0.000000000000000000")
	require.NotEqual(t, "0.000000000000000000", stats.CostEfficiency)
	require.NotEqual(t, "0.000000000000000000", stats.QuadraticAdvantage)
}

// TestGetQuadraticVotingStats_NoVotes tests stats with no votes
func TestGetQuadraticVotingStats_NoVotes(t *testing.T) {
	keeper, ctx := setupQuadraticVotingKeeper(t)

	ts, _ := gogotypes.TimestampProto(time.Now())

	// Create proposal
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddrQuadratic("proposer1"),
		Status:      types.StatusVotingPeriod,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal)

	// Get stats with no votes
	stats := keeper.GetQuadraticVotingStats(ctx, 1)
	require.NotNil(t, stats)
	require.Equal(t, uint64(0), stats.UniqueVoters)
	require.Equal(t, uint64(0), stats.TotalCreditsSpent)
	require.Equal(t, uint64(0), stats.AverageCreditsPerVoter)
	// CostEfficiency and QuadraticAdvantage are now strings (sdk.Dec) for determinism
	require.Equal(t, "0.000000000000000000", stats.CostEfficiency)
	require.Equal(t, "0.000000000000000000", stats.QuadraticAdvantage)
}

// TestCalculateQuadraticAdvantage tests the quadratic advantage calculation
// The function now returns a string representation of sdk.Dec for determinism
func TestCalculateQuadraticAdvantage(t *testing.T) {
	keeper, _ := setupQuadraticVotingKeeper(t)

	tests := []struct {
		name   string
		tally  *types.QuadraticTallyResult
		expect string // sdk.Dec string representation
	}{
		{
			name: "no votes",
			tally: &types.QuadraticTallyResult{
				YesPower:          0,
				NoPower:           0,
				AbstainPower:      0,
				VetoPower:         0,
				TotalCreditsSpent: 0,
				UniqueVoters:      0,
			},
			expect: "0.000000000000000000", // sdk.Dec zero
		},
		{
			name: "typical quadratic advantage",
			tally: &types.QuadraticTallyResult{
				YesPower:          100000, // 100k power
				NoPower:           0,
				AbstainPower:      0,
				VetoPower:         0,
				TotalCreditsSpent: 1000000, // 1M credits spent
				UniqueVoters:      10,
			},
			// Linear = 1000000, Quadratic = 100000
			// Reduction = (1000000 - 100000) / 1000000 * 100 = 90%
			expect: "90.000000000000000000", // 90% as sdk.Dec
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			advantage := keeper.calculateQuadraticAdvantage(tt.tally)
			require.Equal(t, tt.expect, advantage)
		})
	}
}

// TestGetVoteCreditsPerToken tests the credits per token helper
func TestGetVoteCreditsPerToken(t *testing.T) {
	keeper, _ := setupQuadraticVotingKeeper(t)

	creditsPerToken := keeper.getVoteCreditsPerToken()
	require.Equal(t, uint64(100), creditsPerToken)
}

// TestIsQuadraticVotingEnabled tests the quadratic voting enabled check
func TestIsQuadraticVotingEnabled(t *testing.T) {
	keeper, ctx := setupQuadraticVotingKeeper(t)

	enabled := keeper.isQuadraticVotingEnabled(ctx)
	require.True(t, enabled)
}
