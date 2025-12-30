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

func setupLifecycleKeeper(t *testing.T) (*Keeper, sdk.Context) {
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

func TestCreateProposal_Success(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	proposalID, err := keeper.CreateProposal(
		ctx,
		"Test Proposal",
		"This is a test proposal description",
		testAddr("proposer1"),
		types.CategoryText,
		"proposal content",
	)

	require.NoError(t, err)
	require.Equal(t, uint64(1), proposalID)

	// Verify proposal was created
	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, "Test Proposal", proposal.Title)
	require.Equal(t, "This is a test proposal description", proposal.Description)
	require.Equal(t, testAddr("proposer1"), proposal.Proposer)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD, proposal.Status)
	require.Equal(t, types.CategoryText, proposal.Category)
	require.Equal(t, "0", proposal.TotalDeposit)
	require.NotNil(t, proposal.SubmitTime)
	require.NotNil(t, proposal.DepositEndTime)
	require.Nil(t, proposal.VotingStartTime)
	require.Nil(t, proposal.VotingEndTime)
	require.Nil(t, proposal.FinalTallyResult)
	require.Nil(t, proposal.ExecutionTime)
}

func TestCreateProposal_IncrementID(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	// Create first proposal
	id1, err := keeper.CreateProposal(ctx, "Proposal 1", "Description 1", testAddr("proposer1"), types.CategoryText, "")
	require.NoError(t, err)
	require.Equal(t, uint64(1), id1)

	// Create second proposal
	id2, err := keeper.CreateProposal(ctx, "Proposal 2", "Description 2", testAddr("proposer2"), types.CategoryParameterChange, "")
	require.NoError(t, err)
	require.Equal(t, uint64(2), id2)

	// Create third proposal
	id3, err := keeper.CreateProposal(ctx, "Proposal 3", "Description 3", testAddr("proposer3"), types.CategorySoftwareUpgrade, "")
	require.NoError(t, err)
	require.Equal(t, uint64(3), id3)
}

func TestCreateProposal_DifferentCategories(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	categories := []types.ProposalCategory{
		types.CategoryText,
		types.CategoryParameterChange,
		types.CategorySoftwareUpgrade,
		types.CategorySpending,
	}

	for i, category := range categories {
		proposalID, err := keeper.CreateProposal(
			ctx,
			"Proposal",
			"Description",
			testAddr("proposer1"),
			category,
			"content",
		)
		require.NoError(t, err)
		require.Equal(t, uint64(i+1), proposalID)

		proposal, err := keeper.GetProposal(ctx, proposalID)
		require.NoError(t, err)
		require.Equal(t, category, proposal.Category)
	}
}

func TestAdvanceProposalStatus_DepositPeriodToVoting(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	// Create proposal
	proposalID, err := keeper.CreateProposal(ctx, "Test", "Description", testAddr("proposer1"), types.CategoryText, "")
	require.NoError(t, err)

	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD, proposal.Status)

	// Add minimum deposit
	params := keeper.GetParams(ctx)
	proposal.TotalDeposit = params.MinDeposit
	err = keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Advance time past deposit period
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(8 * 24 * time.Hour))

	// Advance proposal status
	err = keeper.AdvanceProposalStatus(ctx, proposalID)
	require.NoError(t, err)

	// Verify moved to voting period
	proposal, err = keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD, proposal.Status)
	require.NotNil(t, proposal.VotingStartTime)
	require.NotNil(t, proposal.VotingEndTime)
}

func TestAdvanceProposalStatus_DepositPeriodFailed(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	// Create proposal with insufficient deposit
	proposalID, err := keeper.CreateProposal(ctx, "Test", "Description", testAddr("proposer1"), types.CategoryText, "")
	require.NoError(t, err)

	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	proposal.TotalDeposit = "100" // Less than minimum
	err = keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Advance time past deposit period
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(8 * 24 * time.Hour))

	// Advance proposal status
	err = keeper.AdvanceProposalStatus(ctx, proposalID)
	require.NoError(t, err)

	// Verify proposal failed
	proposal, err = keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_FAILED, proposal.Status)
}

func TestAdvanceProposalStatus_VotingPeriodToFinalized(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	// Create proposal in voting period
	proposalID, err := keeper.CreateProposal(ctx, "Test", "Description", testAddr("proposer1"), types.CategoryText, "")
	require.NoError(t, err)

	// Move to voting period
	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	params := keeper.GetParams(ctx)
	proposal.TotalDeposit = params.MinDeposit
	proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	now := ctx.BlockTime()
	votingStart, _ := gogotypes.TimestampProto(now)
	votingEnd, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))
	proposal.VotingStartTime = votingStart
	proposal.VotingEndTime = votingEnd
	err = keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Cast some votes
	voter1 := testAddr("voter1")
	vote := &types.Vote{
		ProposalId:  proposalID,
		Voter:       voter1,
		Option:      types.VoteOptionYes,
		VotingPower: "1000000",
		Timestamp:   votingStart,
	}
	err = keeper.SetVote(ctx, vote)
	require.NoError(t, err)

	// Setup staking mock with sufficient bonded tokens
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)
	mockStaking.SetDelegatorBonded(voter1, sdkmath.NewInt(1000000))

	// Advance time past voting period
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(8 * 24 * time.Hour))

	// Advance proposal status
	err = keeper.AdvanceProposalStatus(ctx, proposalID)
	require.NoError(t, err)

	// Verify proposal was finalized (status should be passed or rejected based on tally)
	proposal, err = keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.NotNil(t, proposal.FinalTallyResult)
	// Status should be passed, rejected, or vetoed
	require.Contains(t, []types.ProposalStatus{
		types.ProposalStatus_PROPOSAL_STATUS_PASSED,
		types.ProposalStatus_PROPOSAL_STATUS_REJECTED,
		types.ProposalStatus_PROPOSAL_STATUS_VETOED,
	}, proposal.Status)
}

func TestAdvanceProposalStatus_PassedToExecuted(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	// Create proposal in passed state
	proposalID, err := keeper.CreateProposal(ctx, "Test", "Description", testAddr("proposer1"), types.CategoryText, "")
	require.NoError(t, err)

	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_PASSED
	now := ctx.BlockTime()
	execTime, _ := gogotypes.TimestampProto(now.Add(-1 * time.Hour)) // Past execution time
	proposal.ExecutionTime = execTime
	err = keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Advance proposal status
	err = keeper.AdvanceProposalStatus(ctx, proposalID)
	require.NoError(t, err)

	// Verify proposal was executed
	proposal, err = keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_EXECUTED, proposal.Status)
}

func TestAdvanceProposalStatus_PassedSetExecutionTime(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	// Create proposal in passed state without execution time
	proposalID, err := keeper.CreateProposal(ctx, "Test", "Description", testAddr("proposer1"), types.CategoryText, "")
	require.NoError(t, err)

	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_PASSED
	proposal.ExecutionTime = nil
	err = keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Advance proposal status
	err = keeper.AdvanceProposalStatus(ctx, proposalID)
	require.NoError(t, err)

	// Verify execution time was set
	proposal, err = keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.NotNil(t, proposal.ExecutionTime)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status)
}

func TestAdvanceProposalStatus_TerminalStates(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	terminalStates := []types.ProposalStatus{
		types.ProposalStatus_PROPOSAL_STATUS_FAILED,
		types.ProposalStatus_PROPOSAL_STATUS_REJECTED,
		types.ProposalStatus_PROPOSAL_STATUS_EXECUTED,
	}

	for _, status := range terminalStates {
		proposalID, err := keeper.CreateProposal(ctx, "Test", "Description", testAddr("proposer1"), types.CategoryText, "")
		require.NoError(t, err)

		proposal, err := keeper.GetProposal(ctx, proposalID)
		require.NoError(t, err)
		proposal.Status = status
		err = keeper.SetProposal(ctx, proposal)
		require.NoError(t, err)

		// Advance proposal status
		err = keeper.AdvanceProposalStatus(ctx, proposalID)
		require.NoError(t, err)

		// Verify status unchanged
		proposal, err = keeper.GetProposal(ctx, proposalID)
		require.NoError(t, err)
		require.Equal(t, status, proposal.Status)
	}
}

func TestExecuteProposal_TextProposal(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	proposalID, err := keeper.CreateProposal(ctx, "Text Proposal", "Description", testAddr("proposer1"), types.CategoryText, "")
	require.NoError(t, err)

	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)

	// Execute text proposal
	err = keeper.executeProposal(ctx, proposal)
	require.NoError(t, err)

	// Verify status is executed
	proposal, err = keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_EXECUTED, proposal.Status)
}

func TestExecuteProposal_ParameterChange(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	proposalID, err := keeper.CreateProposal(ctx, "Param Change", "Description", testAddr("proposer1"), types.CategoryParameterChange, "param_content")
	require.NoError(t, err)

	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)

	// Execute parameter change proposal
	err = keeper.executeProposal(ctx, proposal)
	require.NoError(t, err)

	// Verify status is executed
	proposal, err = keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_EXECUTED, proposal.Status)
}

func TestExecuteProposal_SoftwareUpgrade(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	proposalID, err := keeper.CreateProposal(ctx, "Upgrade", "Description", testAddr("proposer1"), types.CategorySoftwareUpgrade, "upgrade_content")
	require.NoError(t, err)

	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)

	// Execute software upgrade proposal
	err = keeper.executeProposal(ctx, proposal)
	require.NoError(t, err)

	// Verify status is executed
	proposal, err = keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_EXECUTED, proposal.Status)
}

func TestExecuteProposal_CommunitySpend(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	proposalID, err := keeper.CreateProposal(ctx, "Community Spend", "Description", testAddr("proposer1"), types.CategorySpending, "spend_content")
	require.NoError(t, err)

	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)

	// Execute community spend proposal
	err = keeper.executeProposal(ctx, proposal)
	require.NoError(t, err)

	// Verify status is executed
	proposal, err = keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_EXECUTED, proposal.Status)
}

func TestCancelProposal_Success(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	proposer := testAddr("proposer1")
	proposalID, err := keeper.CreateProposal(ctx, "Test", "Description", proposer, types.CategoryText, "")
	require.NoError(t, err)

	// Cancel proposal
	err = keeper.CancelProposal(ctx, proposalID, proposer)
	require.NoError(t, err)

	// Verify proposal is failed
	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_FAILED, proposal.Status)
}

func TestCancelProposal_WrongProposer(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	proposer := testAddr("proposer1")
	proposalID, err := keeper.CreateProposal(ctx, "Test", "Description", proposer, types.CategoryText, "")
	require.NoError(t, err)

	// Try to cancel with different address
	err = keeper.CancelProposal(ctx, proposalID, testAddr("other"))
	require.Error(t, err)
	require.Equal(t, types.ErrInvalidProposal, err)
}

func TestCancelProposal_WrongStatus(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	proposer := testAddr("proposer1")
	proposalID, err := keeper.CreateProposal(ctx, "Test", "Description", proposer, types.CategoryText, "")
	require.NoError(t, err)

	// Move to voting period
	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	err = keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Try to cancel in voting period
	err = keeper.CancelProposal(ctx, proposalID, proposer)
	require.Error(t, err)
	require.Equal(t, types.ErrInvalidProposalStatus, err)
}

func TestCancelProposal_NonExistentProposal(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	// Try to cancel non-existent proposal
	err := keeper.CancelProposal(ctx, 999, testAddr("proposer1"))
	require.Error(t, err)
}

func TestGetProposalsByStatus(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	// Create proposals with different statuses
	id1, err := keeper.CreateProposal(ctx, "Proposal 1", "Description", testAddr("proposer1"), types.CategoryText, "")
	require.NoError(t, err)

	id2, err := keeper.CreateProposal(ctx, "Proposal 2", "Description", testAddr("proposer2"), types.CategoryText, "")
	require.NoError(t, err)

	id3, err := keeper.CreateProposal(ctx, "Proposal 3", "Description", testAddr("proposer3"), types.CategoryText, "")
	require.NoError(t, err)

	// Set different statuses
	proposal1, _ := keeper.GetProposal(ctx, id1)
	proposal1.Status = types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	keeper.SetProposal(ctx, proposal1)

	proposal2, _ := keeper.GetProposal(ctx, id2)
	proposal2.Status = types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	keeper.SetProposal(ctx, proposal2)

	proposal3, _ := keeper.GetProposal(ctx, id3)
	proposal3.Status = types.ProposalStatus_PROPOSAL_STATUS_PASSED
	keeper.SetProposal(ctx, proposal3)

	// Get proposals by status
	votingProposals := keeper.GetProposalsByStatus(ctx, types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD)
	require.Len(t, votingProposals, 2)

	passedProposals := keeper.GetProposalsByStatus(ctx, types.ProposalStatus_PROPOSAL_STATUS_PASSED)
	require.Len(t, passedProposals, 1)

	depositProposals := keeper.GetProposalsByStatus(ctx, types.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD)
	require.Len(t, depositProposals, 0)
}

func TestProcessProposalOutcome_QuorumNotReached(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	proposalID, err := keeper.CreateProposal(ctx, "Test", "Description", testAddr("proposer1"), types.CategoryText, "")
	require.NoError(t, err)

	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)

	// Create tally with insufficient votes for quorum
	tally := &types.TallyResult{
		Yes:        "100",
		No:         "50",
		Abstain:    "10",
		NoWithVeto: "5",
	}

	params := keeper.GetParams(ctx)

	// Setup staking mock with total bonded tokens
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)
	mockStaking.SetDelegatorBonded("validator1", sdkmath.NewInt(10000000)) // Much higher than votes

	err = keeper.processProposalOutcome(ctx, proposal, tally, params)
	require.NoError(t, err)

	// Verify proposal was rejected due to quorum
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_REJECTED, proposal.Status)
}

func TestProcessProposalOutcome_VetoThresholdExceeded(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	proposalID, err := keeper.CreateProposal(ctx, "Test", "Description", testAddr("proposer1"), types.CategoryText, "")
	require.NoError(t, err)

	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)

	// Create tally with high veto votes
	tally := &types.TallyResult{
		Yes:        "100000",
		No:         "50000",
		Abstain:    "10000",
		NoWithVeto: "200000", // >33.4% of total votes
	}

	params := keeper.GetParams(ctx)

	// Setup staking mock
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)
	mockStaking.SetDelegatorBonded("validator1", sdkmath.NewInt(1000000))

	err = keeper.processProposalOutcome(ctx, proposal, tally, params)
	require.NoError(t, err)

	// Verify proposal was vetoed
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_VETOED, proposal.Status)
}

func TestProcessProposalOutcome_OnlyAbstainVotes(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	proposalID, err := keeper.CreateProposal(ctx, "Test", "Description", testAddr("proposer1"), types.CategoryText, "")
	require.NoError(t, err)

	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)

	// Create tally with only abstain votes
	tally := &types.TallyResult{
		Yes:        "0",
		No:         "0",
		Abstain:    "1000000",
		NoWithVeto: "0",
	}

	params := keeper.GetParams(ctx)

	// Setup staking mock
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)
	mockStaking.SetDelegatorBonded("validator1", sdkmath.NewInt(2000000))

	err = keeper.processProposalOutcome(ctx, proposal, tally, params)
	require.NoError(t, err)

	// Verify proposal was rejected
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_REJECTED, proposal.Status)
}

func TestProcessProposalOutcome_PassThresholdMet(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	proposalID, err := keeper.CreateProposal(ctx, "Test", "Description", testAddr("proposer1"), types.CategoryText, "")
	require.NoError(t, err)

	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)

	// Create tally with majority yes votes
	tally := &types.TallyResult{
		Yes:        "1000000",
		No:         "100000",
		Abstain:    "50000",
		NoWithVeto: "10000",
	}

	params := keeper.GetParams(ctx)

	// Setup staking mock
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)
	mockStaking.SetDelegatorBonded("validator1", sdkmath.NewInt(2000000))

	// Setup deposits for refund
	deposit := &types.Deposit{
		ProposalId: proposalID,
		Depositor:  testAddr("depositor1"),
		Amount:     "10000000",
		Timestamp:  nil,
	}
	err = keeper.SetDeposit(ctx, deposit)
	require.NoError(t, err)

	err = keeper.processProposalOutcome(ctx, proposal, tally, params)
	require.NoError(t, err)

	// Verify proposal passed
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status)
}

func TestProcessProposalOutcome_ThresholdNotMet(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	proposalID, err := keeper.CreateProposal(ctx, "Test", "Description", testAddr("proposer1"), types.CategoryText, "")
	require.NoError(t, err)

	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)

	// Create tally with insufficient yes votes
	tally := &types.TallyResult{
		Yes:        "100000",
		No:         "1000000",
		Abstain:    "50000",
		NoWithVeto: "10000",
	}

	params := keeper.GetParams(ctx)

	// Setup staking mock
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)
	mockStaking.SetDelegatorBonded("validator1", sdkmath.NewInt(2000000))

	err = keeper.processProposalOutcome(ctx, proposal, tally, params)
	require.NoError(t, err)

	// Verify proposal was rejected
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_REJECTED, proposal.Status)
}

func TestHasMinimumDeposit(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	params := keeper.GetParams(ctx)

	// Create proposal with sufficient deposit
	proposalID, err := keeper.CreateProposal(ctx, "Test", "Description", testAddr("proposer1"), types.CategoryText, "")
	require.NoError(t, err)

	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)

	// Test insufficient deposit
	proposal.TotalDeposit = "100"
	hasMin := keeper.hasMinimumDeposit(proposal, params)
	require.False(t, hasMin)

	// Test sufficient deposit
	proposal.TotalDeposit = params.MinDeposit
	hasMin = keeper.hasMinimumDeposit(proposal, params)
	require.True(t, hasMin)

	// Test more than minimum deposit
	proposal.TotalDeposit = "99999999999999"
	hasMin = keeper.hasMinimumDeposit(proposal, params)
	require.True(t, hasMin)
}

func TestMoveToVotingPeriod(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	proposalID, err := keeper.CreateProposal(ctx, "Test", "Description", testAddr("proposer1"), types.CategoryText, "")
	require.NoError(t, err)

	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD, proposal.Status)

	params := keeper.GetParams(ctx)

	// Move to voting period
	err = keeper.moveToVotingPeriod(ctx, proposal, params)
	require.NoError(t, err)

	// Verify status changed
	require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD, proposal.Status)
	require.NotNil(t, proposal.VotingStartTime)
	require.NotNil(t, proposal.VotingEndTime)

	// Verify voting times make sense
	startTime := timeFromTimestamp(proposal.VotingStartTime)
	endTime := timeFromTimestamp(proposal.VotingEndTime)
	require.True(t, endTime.After(startTime))
}

func TestFinalizeProposal(t *testing.T) {
	keeper, ctx := setupLifecycleKeeper(t)

	proposalID, err := keeper.CreateProposal(ctx, "Test", "Description", testAddr("proposer1"), types.CategoryText, "")
	require.NoError(t, err)

	// Setup proposal in voting period with votes
	proposal, err := keeper.GetProposal(ctx, proposalID)
	require.NoError(t, err)
	proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	now := ctx.BlockTime()
	votingStart, _ := gogotypes.TimestampProto(now)
	proposal.VotingStartTime = votingStart
	err = keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// Add votes
	voter1 := testAddr("voter1")
	vote := &types.Vote{
		ProposalId:  proposalID,
		Voter:       voter1,
		Option:      types.VoteOptionYes,
		VotingPower: "1000000",
		Timestamp:   votingStart,
	}
	err = keeper.SetVote(ctx, vote)
	require.NoError(t, err)

	// Setup staking mock
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)
	mockStaking.SetDelegatorBonded(voter1, sdkmath.NewInt(2000000))

	params := keeper.GetParams(ctx)

	// Finalize proposal
	err = keeper.finalizeProposal(ctx, proposal, params)
	require.NoError(t, err)

	// Verify tally result is set
	require.NotNil(t, proposal.FinalTallyResult)
}
