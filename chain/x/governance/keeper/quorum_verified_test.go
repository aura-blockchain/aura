package keeper

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/governance/types"
)

// setupQuorumTest creates a keeper and context for quorum testing
func setupQuorumTest(t *testing.T) (*Keeper, sdk.Context, *MockStakingKeeper) {
	input := keepertest.CreateTestInputWithKeys(t, "governance")

	// Create mock staking keeper with initialized map
	mockStaking := NewMockStakingKeeper()

	// Create mock bank keeper with initialized maps
	mockBank := &MockBankKeeper{
		balances:       make(map[string]sdk.Coins),
		moduleBalances: make(map[string]sdk.Coins),
		sendErrors:     make(map[string]error),
	}

	mockSecurity := &MockSecurityKeeper{}

	// Create keeper with our mocks
	keeper := NewKeeper(input.Cdc, input.StoreKey, mockStaking, mockBank, mockSecurity)
	ctx := input.Ctx.WithKVGasConfig(storetypes.GasConfig{})

	// Set default params
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)
	keeper.SetNextProposalID(ctx, 1)

	return keeper, ctx, mockStaking
}

// TestQuorum_InsufficientParticipation tests that proposals fail when quorum is not reached
func TestQuorum_InsufficientParticipation(t *testing.T) {
	k, ctx, mockStaking := setupQuorumTest(t)

	// Create proper SDK addresses
	voter1 := sdk.AccAddress([]byte("voter1______________"))
	voter2 := sdk.AccAddress([]byte("voter2______________"))
	voter3 := sdk.AccAddress([]byte("voter3______________"))
	voter4 := sdk.AccAddress([]byte("voter4______________"))

	// Total bonded tokens in network: 1,000,000
	// Setup staking power for voters
	mockStaking.SetDelegatorBonded(voter1.String(), sdkmath.NewInt(100_000)) // 10%
	mockStaking.SetDelegatorBonded(voter2.String(), sdkmath.NewInt(200_000)) // 20%
	mockStaking.SetDelegatorBonded(voter3.String(), sdkmath.NewInt(300_000)) // 30%
	mockStaking.SetDelegatorBonded(voter4.String(), sdkmath.NewInt(400_000)) // 40%

	// Create proposal
	proposalID, err := k.CreateProposal(
		ctx,
		"Test Quorum",
		"Testing quorum enforcement",
		"proposer1",
		types.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
		"content",
	)
	require.NoError(t, err)

	// Move to voting period
	proposal, err := k.GetProposal(ctx, proposalID)
	require.NoError(t, err)

	proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	proposal.VotingStartTime = timestamppb.New(ctx.BlockTime())
	proposal.VotingEndTime = timestamppb.New(ctx.BlockTime().Add(48 * time.Hour))
	proposal.TotalDeposit = "10000000000" // Sufficient deposit
	require.NoError(t, k.SetProposal(ctx, proposal))

	// Cast votes from only 2 voters = 300,000 voting power (30% of total)
	// This is BELOW the 33.4% quorum requirement
	vote1 := &types.Vote{
		ProposalId:  proposalID,
		Voter:       voter1.String(),
		Option:      types.VoteOption_VOTE_OPTION_YES,
		VotingPower: "100000", // Note: CalculateTally will recompute from stakingKeeper
		Timestamp:   timestamppb.Now(),
	}
	require.NoError(t, k.SetVote(ctx, vote1))

	vote2 := &types.Vote{
		ProposalId:  proposalID,
		Voter:       voter2.String(),
		Option:      types.VoteOption_VOTE_OPTION_YES,
		VotingPower: "200000", // Note: CalculateTally will recompute from stakingKeeper
		Timestamp:   timestamppb.Now(),
	}
	require.NoError(t, k.SetVote(ctx, vote2))

	// Calculate tally
	tally := k.CalculateTally(ctx, proposalID)
	require.NotNil(t, tally)

	// Verify vote counts
	require.Equal(t, "300000", tally.Yes)
	require.Equal(t, "0", tally.No)
	require.Equal(t, "0", tally.Abstain)
	require.Equal(t, "0", tally.NoWithVeto)

	// Move to finalization (voting period ended)
	proposal, _ = k.GetProposal(ctx, proposalID)
	ctx = ctx.WithBlockTime(proposal.VotingEndTime.AsTime().Add(1 * time.Second))

	// Advance proposal status to trigger finalization
	err = k.AdvanceProposalStatus(ctx, proposalID)
	require.NoError(t, err)

	// Verify proposal was REJECTED due to insufficient quorum
	proposal, err = k.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_REJECTED, proposal.Status,
		"Proposal should be rejected when quorum not reached (300k < 334k)")

	// Verify the rejection was recorded in tally
	require.NotNil(t, proposal.FinalTallyResult)
}

// TestQuorum_ExactlyAtThreshold tests the boundary case where votes exactly meet quorum
func TestQuorum_ExactlyAtThreshold(t *testing.T) {
	k, ctx, mockStaking := setupQuorumTest(t)

	// Create proper SDK addresses
	voter1 := sdk.AccAddress([]byte("voter1______________"))
	voter2 := sdk.AccAddress([]byte("voter2______________"))

	// Total bonded: 1,000,000 tokens
	// Quorum requirement: 33.4% = 334,000 tokens
	mockStaking.SetDelegatorBonded(voter1.String(), sdkmath.NewInt(334_000)) // Exactly at quorum
	mockStaking.SetDelegatorBonded(voter2.String(), sdkmath.NewInt(666_000)) // Rest of tokens

	// Create and prepare proposal
	proposalID, err := k.CreateProposal(
		ctx,
		"Test Proposal",
		"Testing exact quorum",
		"proposer1",
		types.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
		"content",
	)
	require.NoError(t, err)

	// Move to voting period
	proposal, _ := k.GetProposal(ctx, proposalID)
	proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	proposal.VotingStartTime = timestamppb.New(ctx.BlockTime())
	proposal.VotingEndTime = timestamppb.New(ctx.BlockTime().Add(48 * time.Hour))
	proposal.TotalDeposit = "10000000000"
	require.NoError(t, k.SetProposal(ctx, proposal))

	// Cast vote with exactly quorum amount (334,000)
	vote1 := &types.Vote{
		ProposalId:  proposalID,
		Voter:       voter1.String(),
		Option:      types.VoteOption_VOTE_OPTION_YES,
		VotingPower: "334000", // Exactly at quorum
		Timestamp:   timestamppb.Now(),
	}
	require.NoError(t, k.SetVote(ctx, vote1))

	// Calculate and finalize
	tally := k.CalculateTally(ctx, proposalID)
	require.Equal(t, "334000", tally.Yes)

	// Finalize proposal
	proposal, _ = k.GetProposal(ctx, proposalID)
	ctx = ctx.WithBlockTime(proposal.VotingEndTime.AsTime().Add(1 * time.Second))
	err = k.AdvanceProposalStatus(ctx, proposalID)
	require.NoError(t, err)

	// Should PASS - quorum met (334k >= 334k) and threshold met (100% yes)
	proposal, err = k.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status,
		"Proposal should pass when exactly at quorum with sufficient yes votes")
}

// TestQuorum_OneBelowThreshold tests that one vote below quorum fails
func TestQuorum_OneBelowThreshold(t *testing.T) {
	k, ctx, mockStaking := setupQuorumTest(t)

	mockStaking.SetDelegatorBonded("voter1", sdkmath.NewInt(333_999)) // One below quorum
	mockStaking.SetDelegatorBonded("voter2", sdkmath.NewInt(666_001))

	proposalID, err := k.CreateProposal(ctx, "Test", "One below quorum",
		"proposer1", types.ProposalCategory_PROPOSAL_CATEGORY_TEXT, "content")
	require.NoError(t, err)

	// Move to voting
	proposal, _ := k.GetProposal(ctx, proposalID)
	proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	proposal.VotingStartTime = timestamppb.New(ctx.BlockTime())
	proposal.VotingEndTime = timestamppb.New(ctx.BlockTime().Add(48 * time.Hour))
	proposal.TotalDeposit = "10000000000"
	require.NoError(t, k.SetProposal(ctx, proposal))

	vote := &types.Vote{
		ProposalId:  proposalID,
		Voter:       "voter1",
		Option:      types.VoteOption_VOTE_OPTION_YES,
		VotingPower: "333999", // One below quorum requirement
		Timestamp:   timestamppb.Now(),
	}
	require.NoError(t, k.SetVote(ctx, vote))

	// Finalize
	proposal, _ = k.GetProposal(ctx, proposalID)
	ctx = ctx.WithBlockTime(proposal.VotingEndTime.AsTime().Add(1 * time.Second))
	err = k.AdvanceProposalStatus(ctx, proposalID)
	require.NoError(t, err)

	// Should FAIL quorum
	proposal, err = k.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_REJECTED, proposal.Status,
		"Proposal should fail when one vote below quorum (333,999 < 334,000)")
}

// TestQuorum_WellAboveThreshold tests proposals that clearly exceed quorum
func TestQuorum_WellAboveThreshold(t *testing.T) {
	k, ctx, mockStaking := setupQuorumTest(t)

	// Create proper SDK addresses
	voter1 := sdk.AccAddress([]byte("voter1______________"))
	voter2 := sdk.AccAddress([]byte("voter2______________"))
	voter3 := sdk.AccAddress([]byte("voter3______________"))

	// 60% participation - well above 33.4% quorum
	mockStaking.SetDelegatorBonded(voter1.String(), sdkmath.NewInt(300_000))
	mockStaking.SetDelegatorBonded(voter2.String(), sdkmath.NewInt(300_000))
	mockStaking.SetDelegatorBonded(voter3.String(), sdkmath.NewInt(400_000))

	proposalID, err := k.CreateProposal(ctx, "Test", "Well above quorum",
		"proposer1", types.ProposalCategory_PROPOSAL_CATEGORY_TEXT, "content")
	require.NoError(t, err)

	// Move to voting
	proposal, _ := k.GetProposal(ctx, proposalID)
	proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	proposal.VotingStartTime = timestamppb.New(ctx.BlockTime())
	proposal.VotingEndTime = timestamppb.New(ctx.BlockTime().Add(48 * time.Hour))
	proposal.TotalDeposit = "10000000000"
	require.NoError(t, k.SetProposal(ctx, proposal))

	// Cast votes totaling 600,000 (60% of 1M)
	votes := []*types.Vote{
		{
			ProposalId:  proposalID,
			Voter:       voter1.String(),
			Option:      types.VoteOption_VOTE_OPTION_YES,
			VotingPower: "300000",
			Timestamp:   timestamppb.Now(),
		},
		{
			ProposalId:  proposalID,
			Voter:       voter2.String(),
			Option:      types.VoteOption_VOTE_OPTION_YES,
			VotingPower: "300000",
			Timestamp:   timestamppb.Now(),
		},
	}

	for _, vote := range votes {
		require.NoError(t, k.SetVote(ctx, vote))
	}

	// Verify tally
	tally := k.CalculateTally(ctx, proposalID)
	require.Equal(t, "600000", tally.Yes)

	// Finalize
	proposal, _ = k.GetProposal(ctx, proposalID)
	ctx = ctx.WithBlockTime(proposal.VotingEndTime.AsTime().Add(1 * time.Second))
	err = k.AdvanceProposalStatus(ctx, proposalID)
	require.NoError(t, err)

	// Should PASS - well above quorum and threshold
	proposal, err = k.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status,
		"Proposal should pass with 60% participation (well above 33.4% quorum)")
}

// TestQuorum_QuorumMetButThresholdNotMet tests the critical distinction
func TestQuorum_QuorumMetButThresholdNotMet(t *testing.T) {
	k, ctx, mockStaking := setupQuorumTest(t)

	// Create proper SDK addresses
	voter1 := sdk.AccAddress([]byte("voter1______________"))
	voter2 := sdk.AccAddress([]byte("voter2______________"))

	mockStaking.SetDelegatorBonded(voter1.String(), sdkmath.NewInt(500_000))
	mockStaking.SetDelegatorBonded(voter2.String(), sdkmath.NewInt(500_000))

	proposalID, err := k.CreateProposal(ctx, "Test", "Quorum met but threshold not",
		"proposer1", types.ProposalCategory_PROPOSAL_CATEGORY_TEXT, "content")
	require.NoError(t, err)

	// Move to voting
	proposal, _ := k.GetProposal(ctx, proposalID)
	proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	proposal.VotingStartTime = timestamppb.New(ctx.BlockTime())
	proposal.VotingEndTime = timestamppb.New(ctx.BlockTime().Add(48 * time.Hour))
	proposal.TotalDeposit = "10000000000"
	require.NoError(t, k.SetProposal(ctx, proposal))

	// Cast votes totaling 400,000 (40% participation - above 33.4% quorum)
	// But with only 45% yes votes (below 50% threshold)
	votes := []*types.Vote{
		{
			ProposalId:  proposalID,
			Voter:       voter1.String(),
			Option:      types.VoteOption_VOTE_OPTION_YES,
			VotingPower: "180000", // 45% of 400k
			Timestamp:   timestamppb.Now(),
		},
		{
			ProposalId:  proposalID,
			Voter:       voter2.String(),
			Option:      types.VoteOption_VOTE_OPTION_NO,
			VotingPower: "220000", // 55% of 400k
			Timestamp:   timestamppb.Now(),
		},
	}

	for _, vote := range votes {
		require.NoError(t, k.SetVote(ctx, vote))
	}

	// Verify tally
	tally := k.CalculateTally(ctx, proposalID)
	totalVotes, _ := sdkmath.NewIntFromString(tally.Yes)
	noVotes, _ := sdkmath.NewIntFromString(tally.No)
	totalVotes = totalVotes.Add(noVotes)

	require.True(t, totalVotes.GTE(sdkmath.NewInt(334_000)),
		"Total votes should exceed quorum")

	// Finalize
	proposal, _ = k.GetProposal(ctx, proposalID)
	ctx = ctx.WithBlockTime(proposal.VotingEndTime.AsTime().Add(1 * time.Second))
	err = k.AdvanceProposalStatus(ctx, proposalID)
	require.NoError(t, err)

	// Should be REJECTED - quorum met but threshold not met
	proposal, err = k.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_REJECTED, proposal.Status,
		"Proposal should be rejected when quorum met but yes votes < 50% threshold")
}

// TestQuorum_WithAbstainVotes tests that abstain votes count toward quorum but not threshold
func TestQuorum_WithAbstainVotes(t *testing.T) {
	k, ctx, mockStaking := setupQuorumTest(t)

	// Create proper SDK addresses
	voter1 := sdk.AccAddress([]byte("voter1______________"))
	voter2 := sdk.AccAddress([]byte("voter2______________"))
	voter3 := sdk.AccAddress([]byte("voter3______________"))

	// Set bonded amounts (actual voting power comes from staking keeper, not vote.VotingPower field)
	// Total votes: 400,000 (meets quorum of 334k)
	// Abstain: 100,000 (doesn't count toward yes/no)
	// Yes: 200,000
	// No: 100,000
	// Yes percentage of non-abstain: 200k / 300k = 66.6% > 50% threshold
	mockStaking.SetDelegatorBonded(voter1.String(), sdkmath.NewInt(100_000))
	mockStaking.SetDelegatorBonded(voter2.String(), sdkmath.NewInt(200_000))
	mockStaking.SetDelegatorBonded(voter3.String(), sdkmath.NewInt(100_000))

	proposalID, err := k.CreateProposal(ctx, "Test", "With abstain votes",
		"proposer1", types.ProposalCategory_PROPOSAL_CATEGORY_TEXT, "content")
	require.NoError(t, err)

	// Move to voting
	proposal, _ := k.GetProposal(ctx, proposalID)
	proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	proposal.VotingStartTime = timestamppb.New(ctx.BlockTime())
	proposal.VotingEndTime = timestamppb.New(ctx.BlockTime().Add(48 * time.Hour))
	proposal.TotalDeposit = "10000000000"
	require.NoError(t, k.SetProposal(ctx, proposal))

	votes := []*types.Vote{
		{
			ProposalId:  proposalID,
			Voter:       voter1.String(),
			Option:      types.VoteOption_VOTE_OPTION_ABSTAIN,
			VotingPower: "100000", // Actual power comes from staking keeper
			Timestamp:   timestamppb.Now(),
		},
		{
			ProposalId:  proposalID,
			Voter:       voter2.String(),
			Option:      types.VoteOption_VOTE_OPTION_YES,
			VotingPower: "200000", // Actual power comes from staking keeper
			Timestamp:   timestamppb.Now(),
		},
		{
			ProposalId:  proposalID,
			Voter:       voter3.String(),
			Option:      types.VoteOption_VOTE_OPTION_NO,
			VotingPower: "100000", // Actual power comes from staking keeper
			Timestamp:   timestamppb.Now(),
		},
	}

	for _, vote := range votes {
		require.NoError(t, k.SetVote(ctx, vote))
	}

	// Finalize
	proposal, _ = k.GetProposal(ctx, proposalID)
	ctx = ctx.WithBlockTime(proposal.VotingEndTime.AsTime().Add(1 * time.Second))
	err = k.AdvanceProposalStatus(ctx, proposalID)
	require.NoError(t, err)

	// Should PASS - quorum met (400k > 334k) and threshold met (66.6% yes of non-abstain)
	proposal, err = k.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status,
		"Proposal should pass when abstain votes help meet quorum and yes > 50% of non-abstain")
}

// TestQuorum_OnlyAbstainVotes tests the edge case where all votes are abstain
func TestQuorum_OnlyAbstainVotes(t *testing.T) {
	k, ctx, mockStaking := setupQuorumTest(t)

	mockStaking.SetDelegatorBonded("voter1", sdkmath.NewInt(400_000))
	mockStaking.SetDelegatorBonded("voter2", sdkmath.NewInt(600_000))

	proposalID, err := k.CreateProposal(ctx, "Test", "Only abstain votes",
		"proposer1", types.ProposalCategory_PROPOSAL_CATEGORY_TEXT, "content")
	require.NoError(t, err)

	// Move to voting
	proposal, _ := k.GetProposal(ctx, proposalID)
	proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	proposal.VotingStartTime = timestamppb.New(ctx.BlockTime())
	proposal.VotingEndTime = timestamppb.New(ctx.BlockTime().Add(48 * time.Hour))
	proposal.TotalDeposit = "10000000000"
	require.NoError(t, k.SetProposal(ctx, proposal))

	// Only abstain votes - meets quorum but has no yes/no votes
	vote := &types.Vote{
		ProposalId:  proposalID,
		Voter:       "voter1",
		Option:      types.VoteOption_VOTE_OPTION_ABSTAIN,
		VotingPower: "400000",
		Timestamp:   timestamppb.Now(),
	}
	require.NoError(t, k.SetVote(ctx, vote))

	// Finalize
	proposal, _ = k.GetProposal(ctx, proposalID)
	ctx = ctx.WithBlockTime(proposal.VotingEndTime.AsTime().Add(1 * time.Second))
	err = k.AdvanceProposalStatus(ctx, proposalID)
	require.NoError(t, err)

	// Should be REJECTED - only abstain votes (no yes/no to calculate threshold)
	proposal, err = k.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_REJECTED, proposal.Status,
		"Proposal should be rejected when only abstain votes (cannot calculate threshold)")
}

// TestQuorum_VetoOverridesQuorumAndThreshold tests that veto takes precedence
func TestQuorum_VetoOverridesQuorumAndThreshold(t *testing.T) {
	k, ctx, mockStaking := setupQuorumTest(t)

	// Create proper SDK addresses
	voter1 := sdk.AccAddress([]byte("voter1______________"))
	voter2 := sdk.AccAddress([]byte("voter2______________"))
	voter3 := sdk.AccAddress([]byte("voter3______________"))

	mockStaking.SetDelegatorBonded(voter1.String(), sdkmath.NewInt(400_000))
	mockStaking.SetDelegatorBonded(voter2.String(), sdkmath.NewInt(350_000))
	mockStaking.SetDelegatorBonded(voter3.String(), sdkmath.NewInt(250_000))

	proposalID, err := k.CreateProposal(ctx, "Test", "Veto overrides",
		"proposer1", types.ProposalCategory_PROPOSAL_CATEGORY_TEXT, "content")
	require.NoError(t, err)

	// Move to voting
	proposal, _ := k.GetProposal(ctx, proposalID)
	proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	proposal.VotingStartTime = timestamppb.New(ctx.BlockTime())
	proposal.VotingEndTime = timestamppb.New(ctx.BlockTime().Add(48 * time.Hour))
	proposal.TotalDeposit = "10000000000"
	require.NoError(t, k.SetProposal(ctx, proposal))

	// Total votes: 750,000 (75% - well above quorum)
	// Yes: 400,000 (53% of non-abstain - above threshold)
	// NoWithVeto: 350,000 (46.6% of total - above 33.4% veto threshold)
	votes := []*types.Vote{
		{
			ProposalId:  proposalID,
			Voter:       voter1.String(),
			Option:      types.VoteOption_VOTE_OPTION_YES,
			VotingPower: "400000",
			Timestamp:   timestamppb.Now(),
		},
		{
			ProposalId:  proposalID,
			Voter:       voter2.String(),
			Option:      types.VoteOption_VOTE_OPTION_NO_WITH_VETO,
			VotingPower: "350000",
			Timestamp:   timestamppb.Now(),
		},
	}

	for _, vote := range votes {
		require.NoError(t, k.SetVote(ctx, vote))
	}

	// Finalize
	proposal, _ = k.GetProposal(ctx, proposalID)
	ctx = ctx.WithBlockTime(proposal.VotingEndTime.AsTime().Add(1 * time.Second))
	err = k.AdvanceProposalStatus(ctx, proposalID)
	require.NoError(t, err)

	// Should be VETOED despite meeting quorum and threshold
	proposal, err = k.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_VETOED, proposal.Status,
		"Proposal should be vetoed when veto votes > 33.4% (even if quorum and threshold met)")
}

// TestQuorum_CalculationCorrectness tests the mathematical correctness of quorum calculation
func TestQuorum_CalculationCorrectness(t *testing.T) {
	tests := []struct {
		name             string
		totalBonded      int64
		votes            int64
		expectedQuorum   int64
		shouldMeetQuorum bool
	}{
		{
			name:             "1M total - exactly at quorum",
			totalBonded:      1_000_000,
			votes:            334_000,
			expectedQuorum:   334_000,
			shouldMeetQuorum: true,
		},
		{
			name:             "1M total - one below quorum",
			totalBonded:      1_000_000,
			votes:            333_999,
			expectedQuorum:   334_000,
			shouldMeetQuorum: false,
		},
		{
			name:             "10M total - at quorum",
			totalBonded:      10_000_000,
			votes:            3_340_000,
			expectedQuorum:   3_340_000,
			shouldMeetQuorum: true,
		},
		{
			name:             "100 total - at quorum (small network)",
			totalBonded:      100,
			votes:            34, // 33.4% of 100 = 33.4, truncated to 33
			expectedQuorum:   33,
			shouldMeetQuorum: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify quorum calculation
			params := types.DefaultParams()
			quorumDec, err := sdkmath.LegacyNewDecFromStr(params.Quorum)
			require.NoError(t, err)

			totalBondedInt := sdkmath.NewInt(tt.totalBonded)
			requiredQuorum := quorumDec.MulInt(totalBondedInt).TruncateInt()

			require.Equal(t, tt.expectedQuorum, requiredQuorum.Int64(),
				"Quorum calculation should match expected value")

			votesInt := sdkmath.NewInt(tt.votes)
			meetsQuorum := !votesInt.LT(requiredQuorum)

			require.Equal(t, tt.shouldMeetQuorum, meetsQuorum,
				"Quorum check should match expected result")
		})
	}
}
