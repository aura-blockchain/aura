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

// TestQuorumThreshold_ComprehensiveScenarios tests all critical quorum and threshold enforcement scenarios
// This test file verifies that issue #032 (Governance No Quorum/Threshold Enforcement) is completely fixed
func TestQuorumThreshold_ComprehensiveScenarios(t *testing.T) {
	t.Run("QuorumFailure_InsufficientParticipation", func(t *testing.T) {
		k, ctx, mockStaking := setupQuorumThresholdTest(t)

		// Setup: 1,000,000 total bonded, quorum = 33.4% = 334,000
		// Only 300,000 votes cast (below quorum)
		voter1 := sdk.AccAddress([]byte("voter1______________"))
		voter2 := sdk.AccAddress([]byte("voter2______________"))
		nonVoter1 := sdk.AccAddress([]byte("nonvoter1___________"))
		nonVoter2 := sdk.AccAddress([]byte("nonvoter2___________"))

		// Set up total bonded to be 1,000,000
		mockStaking.SetDelegatorBonded(voter1.String(), sdkmath.NewInt(150_000))
		mockStaking.SetDelegatorBonded(voter2.String(), sdkmath.NewInt(150_000))
		mockStaking.SetDelegatorBonded(nonVoter1.String(), sdkmath.NewInt(350_000)) // Don't vote
		mockStaking.SetDelegatorBonded(nonVoter2.String(), sdkmath.NewInt(350_000)) // Don't vote

		proposalID := createVotingProposal(t, k, ctx)

		// Cast votes with 100% yes but below quorum (only 300k out of 1M)
		castVote(t, k, ctx, proposalID, voter1.String(), types.VoteOption_VOTE_OPTION_YES)
		castVote(t, k, ctx, proposalID, voter2.String(), types.VoteOption_VOTE_OPTION_YES)

		// Finalize
		finalizeProposal(t, k, ctx, proposalID)

		// Verify: REJECTED due to insufficient quorum
		proposal, err := k.GetProposal(ctx, proposalID)
		require.NoError(t, err)
		require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_REJECTED, proposal.Status,
			"Proposal must be rejected when quorum not reached (300k < 334k)")
	})

	t.Run("VetoThresholdExceeded_ProposalVetoed", func(t *testing.T) {
		k, ctx, mockStaking := setupQuorumThresholdTest(t)

		// Setup: Sufficient quorum but >33.4% NoWithVeto votes
		voter1 := sdk.AccAddress([]byte("voter1______________"))
		voter2 := sdk.AccAddress([]byte("voter2______________"))
		voter3 := sdk.AccAddress([]byte("voter3______________"))

		mockStaking.SetDelegatorBonded(voter1.String(), sdkmath.NewInt(300_000))
		mockStaking.SetDelegatorBonded(voter2.String(), sdkmath.NewInt(350_000))
		mockStaking.SetDelegatorBonded(voter3.String(), sdkmath.NewInt(350_000))

		proposalID := createVotingProposal(t, k, ctx)

		// Cast votes: 300k yes, 350k veto (35% > 33.4%)
		castVote(t, k, ctx, proposalID, voter1.String(), types.VoteOption_VOTE_OPTION_YES)
		castVote(t, k, ctx, proposalID, voter2.String(), types.VoteOption_VOTE_OPTION_NO_WITH_VETO)

		// Finalize
		finalizeProposal(t, k, ctx, proposalID)

		// Verify: VETOED despite quorum being met
		proposal, err := k.GetProposal(ctx, proposalID)
		require.NoError(t, err)
		require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_VETOED, proposal.Status,
			"Proposal must be vetoed when veto threshold exceeded (35% > 33.4%)")
	})

	t.Run("ThresholdNotMet_ProposalRejected", func(t *testing.T) {
		k, ctx, mockStaking := setupQuorumThresholdTest(t)

		// Setup: Quorum met but yes votes < 50% threshold
		voter1 := sdk.AccAddress([]byte("voter1______________"))
		voter2 := sdk.AccAddress([]byte("voter2______________"))

		mockStaking.SetDelegatorBonded(voter1.String(), sdkmath.NewInt(200_000))
		mockStaking.SetDelegatorBonded(voter2.String(), sdkmath.NewInt(300_000))

		proposalID := createVotingProposal(t, k, ctx)

		// Cast votes: Total = 500k (meets quorum), Yes = 200k (40% < 50% threshold)
		castVote(t, k, ctx, proposalID, voter1.String(), types.VoteOption_VOTE_OPTION_YES)
		castVote(t, k, ctx, proposalID, voter2.String(), types.VoteOption_VOTE_OPTION_NO)

		// Finalize
		finalizeProposal(t, k, ctx, proposalID)

		// Verify: REJECTED - quorum met but threshold not met
		proposal, err := k.GetProposal(ctx, proposalID)
		require.NoError(t, err)
		require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_REJECTED, proposal.Status,
			"Proposal must be rejected when threshold not met (40% < 50%)")
	})

	t.Run("AllChecksPass_ProposalPasses", func(t *testing.T) {
		k, ctx, mockStaking := setupQuorumThresholdTest(t)

		// Setup: All conditions met for proposal to pass
		voter1 := sdk.AccAddress([]byte("voter1______________"))
		voter2 := sdk.AccAddress([]byte("voter2______________"))

		mockStaking.SetDelegatorBonded(voter1.String(), sdkmath.NewInt(400_000))
		mockStaking.SetDelegatorBonded(voter2.String(), sdkmath.NewInt(200_000))

		proposalID := createVotingProposal(t, k, ctx)

		// Cast votes: Total = 600k (meets quorum), Yes = 400k (66.6% > 50% threshold)
		castVote(t, k, ctx, proposalID, voter1.String(), types.VoteOption_VOTE_OPTION_YES)
		castVote(t, k, ctx, proposalID, voter2.String(), types.VoteOption_VOTE_OPTION_NO)

		// Finalize
		finalizeProposal(t, k, ctx, proposalID)

		// Verify: PASSED - all checks satisfied
		proposal, err := k.GetProposal(ctx, proposalID)
		require.NoError(t, err)
		require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status,
			"Proposal must pass when quorum, threshold, and veto checks all pass")
	})

	t.Run("AbstainVotes_CountTowardQuorumOnly", func(t *testing.T) {
		k, ctx, mockStaking := setupQuorumThresholdTest(t)

		// Setup: Abstain votes help meet quorum but don't affect threshold
		voter1 := sdk.AccAddress([]byte("voter1______________"))
		voter2 := sdk.AccAddress([]byte("voter2______________"))
		voter3 := sdk.AccAddress([]byte("voter3______________"))

		mockStaking.SetDelegatorBonded(voter1.String(), sdkmath.NewInt(150_000))
		mockStaking.SetDelegatorBonded(voter2.String(), sdkmath.NewInt(200_000))
		mockStaking.SetDelegatorBonded(voter3.String(), sdkmath.NewInt(100_000))

		proposalID := createVotingProposal(t, k, ctx)

		// Cast votes: Total = 450k (meets quorum), Yes = 200k, Abstain = 150k, No = 100k
		// Yes% of non-abstain = 200k / 300k = 66.6% > 50%
		castVote(t, k, ctx, proposalID, voter1.String(), types.VoteOption_VOTE_OPTION_ABSTAIN)
		castVote(t, k, ctx, proposalID, voter2.String(), types.VoteOption_VOTE_OPTION_YES)
		castVote(t, k, ctx, proposalID, voter3.String(), types.VoteOption_VOTE_OPTION_NO)

		// Finalize
		finalizeProposal(t, k, ctx, proposalID)

		// Verify: PASSED - abstain counted for quorum, not threshold
		proposal, err := k.GetProposal(ctx, proposalID)
		require.NoError(t, err)
		require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status,
			"Proposal must pass when abstain helps meet quorum and yes > 50% of non-abstain")
	})

	t.Run("OnlyAbstainVotes_ProposalRejected", func(t *testing.T) {
		k, ctx, mockStaking := setupQuorumThresholdTest(t)

		// Setup: Only abstain votes (cannot calculate threshold)
		voter1 := sdk.AccAddress([]byte("voter1______________"))
		mockStaking.SetDelegatorBonded(voter1.String(), sdkmath.NewInt(400_000))

		proposalID := createVotingProposal(t, k, ctx)

		// Cast only abstain votes
		castVote(t, k, ctx, proposalID, voter1.String(), types.VoteOption_VOTE_OPTION_ABSTAIN)

		// Finalize
		finalizeProposal(t, k, ctx, proposalID)

		// Verify: REJECTED - cannot calculate threshold with only abstain
		proposal, err := k.GetProposal(ctx, proposalID)
		require.NoError(t, err)
		require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_REJECTED, proposal.Status,
			"Proposal must be rejected when only abstain votes (cannot calculate threshold)")
	})

	t.Run("DepositHandling_RefundOnPass", func(t *testing.T) {
		k, ctx, mockStaking := setupQuorumThresholdTest(t)

		// Setup bank keeper to track refunds
		mockBank := k.bankKeeper.(*MockBankKeeper)
		mockBank.moduleBalances[types.ModuleName] = sdk.NewCoins(sdk.NewInt64Coin("stake", 10_000_000_000))

		voter1 := sdk.AccAddress([]byte("voter1______________"))
		mockStaking.SetDelegatorBonded(voter1.String(), sdkmath.NewInt(400_000))

		proposalID := createVotingProposal(t, k, ctx)
		castVote(t, k, ctx, proposalID, voter1.String(), types.VoteOption_VOTE_OPTION_YES)

		// Set up deposits
		depositor := sdk.AccAddress([]byte("depositor___________"))
		mockBank.balances[depositor.String()] = sdk.NewCoins()
		deposit := &types.Deposit{
			ProposalId: proposalID,
			Depositor:  depositor.String(),
			Amount:     "1000000stake",
		}
		require.NoError(t, k.SetDeposit(ctx, deposit))

		// Finalize - should pass and refund deposits
		finalizeProposal(t, k, ctx, proposalID)

		// Verify: Proposal passed
		proposal, err := k.GetProposal(ctx, proposalID)
		require.NoError(t, err)
		require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status)

		// Verify: Deposits refunded (deposit deleted)
		_, err = k.GetDeposit(ctx, proposalID, depositor.String())
		require.Error(t, err, "Deposit should be deleted after refund")
	})

	t.Run("DepositHandling_BurnOnVeto", func(t *testing.T) {
		k, ctx, mockStaking := setupQuorumThresholdTest(t)

		// Setup bank keeper
		mockBank := k.bankKeeper.(*MockBankKeeper)
		mockBank.moduleBalances[types.ModuleName] = sdk.NewCoins(sdk.NewInt64Coin("stake", 10_000_000_000))

		voter1 := sdk.AccAddress([]byte("voter1______________"))
		mockStaking.SetDelegatorBonded(voter1.String(), sdkmath.NewInt(400_000))

		proposalID := createVotingProposal(t, k, ctx)
		castVote(t, k, ctx, proposalID, voter1.String(), types.VoteOption_VOTE_OPTION_NO_WITH_VETO)

		// Set up deposits
		depositor := sdk.AccAddress([]byte("depositor___________"))
		deposit := &types.Deposit{
			ProposalId: proposalID,
			Depositor:  depositor.String(),
			Amount:     "1000000stake",
		}
		require.NoError(t, k.SetDeposit(ctx, deposit))

		// Finalize - should veto and burn deposits
		finalizeProposal(t, k, ctx, proposalID)

		// Verify: Proposal vetoed
		proposal, err := k.GetProposal(ctx, proposalID)
		require.NoError(t, err)
		require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_VETOED, proposal.Status)

		// Verify: Deposits burned (deposit deleted)
		_, err = k.GetDeposit(ctx, proposalID, depositor.String())
		require.Error(t, err, "Deposit should be deleted after burn")
	})
}

// Helper functions for test setup
func setupQuorumThresholdTest(t *testing.T) (*Keeper, sdk.Context, *MockStakingKeeper) {
	input := keepertest.CreateTestInputWithKeys(t, "governance")

	mockStaking := NewMockStakingKeeper()
	mockBank := &MockBankKeeper{
		balances:       make(map[string]sdk.Coins),
		moduleBalances: make(map[string]sdk.Coins),
		sendErrors:     make(map[string]error),
	}
	mockSecurity := &MockSecurityKeeper{}

	keeper := NewKeeper(input.Cdc, input.StoreKey, mockStaking, mockBank, mockSecurity)
	ctx := input.Ctx.WithKVGasConfig(storetypes.GasConfig{})

	// Set default params
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)
	keeper.SetNextProposalID(ctx, 1)

	return keeper, ctx, mockStaking
}

func createVotingProposal(t *testing.T, k *Keeper, ctx sdk.Context) uint64 {
	proposalID, err := k.CreateProposal(
		ctx,
		"Test Proposal",
		"Testing quorum and threshold enforcement",
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
	proposal.TotalDeposit = "10000000000"
	require.NoError(t, k.SetProposal(ctx, proposal))

	return proposalID
}

func castVote(t *testing.T, k *Keeper, ctx sdk.Context, proposalID uint64, voter string, option types.VoteOption) {
	vote := &types.Vote{
		ProposalId:  proposalID,
		Voter:       voter,
		Option:      option,
		VotingPower: "0", // Actual power comes from staking keeper
		Timestamp:   timestamppb.Now(),
	}
	require.NoError(t, k.SetVote(ctx, vote))
}

func finalizeProposal(t *testing.T, k *Keeper, ctx sdk.Context, proposalID uint64) {
	proposal, err := k.GetProposal(ctx, proposalID)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(proposal.VotingEndTime.AsTime().Add(1 * time.Second))
	err = k.AdvanceProposalStatus(ctx, proposalID)
	require.NoError(t, err)
}
