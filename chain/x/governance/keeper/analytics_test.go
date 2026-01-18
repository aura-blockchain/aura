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

func setupAnalyticsKeeper(t *testing.T) (*Keeper, sdk.Context) {
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

func TestGetGovernanceAnalytics_EmptyState(t *testing.T) {
	keeper, ctx := setupAnalyticsKeeper(t)

	// Test with no proposals
	analytics := keeper.GetGovernanceAnalytics(ctx)

	require.NotNil(t, analytics)
	require.Equal(t, uint64(0), analytics.TotalProposals)
	require.Equal(t, "0", analytics.AverageVotingPower)
	require.True(t, analytics.ParticipationRate.Equal(sdkmath.LegacyZeroDec()))
	require.True(t, analytics.PassRate.Equal(sdkmath.LegacyZeroDec()))
	require.True(t, analytics.VetoRate.Equal(sdkmath.LegacyZeroDec()))
}

func TestGetGovernanceAnalytics_WithProposals(t *testing.T) {
	keeper, ctx := setupAnalyticsKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	endTs, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))
	execTs, _ := gogotypes.TimestampProto(now.Add(10 * 24 * time.Hour))

	// Create proposals with different statuses
	proposals := []struct {
		id       uint64
		status   types.ProposalStatus
		category types.ProposalCategory
		tally    *types.TallyResult
	}{
		{1, types.StatusPassed, types.CategoryText, &types.TallyResult{Yes: "1000", No: "100", Abstain: "50", NoWithVeto: "0"}},
		{2, types.StatusExecuted, types.CategoryParameterChange, &types.TallyResult{Yes: "2000", No: "200", Abstain: "100", NoWithVeto: "0"}},
		{3, types.StatusRejected, types.CategoryText, &types.TallyResult{Yes: "500", No: "1000", Abstain: "0", NoWithVeto: "500"}},
		{4, types.StatusFailed, types.CategorySpending, &types.TallyResult{Yes: "100", No: "900", Abstain: "0", NoWithVeto: "0"}},
		{5, types.StatusVotingPeriod, types.CategoryText, nil},
	}

	for _, p := range proposals {
		proposal := &types.Proposal{
			Id:               p.id,
			Title:            "Test Proposal",
			Description:      "Test Description",
			Proposer:         testAddr("proposer1"),
			Status:           p.status,
			Category:         p.category,
			SubmitTime:       ts,
			VotingEndTime:    endTs,
			FinalTallyResult: p.tally,
		}
		if p.status == types.StatusExecuted {
			proposal.ExecutionTime = execTs
		}
		keeper.SetProposal(ctx, proposal)

		// Add votes for participation rate calculation
		if p.status != types.StatusVotingPeriod {
			for i := 1; i <= 3; i++ {
				vote := &types.Vote{
					ProposalId:  p.id,
					Voter:       testAddr("voter" + string(rune('0'+i))),
					Option:      types.VoteOptionYes,
					VotingPower: "100",
					Timestamp:   ts,
				}
				keeper.SetVote(ctx, vote)
			}
		}
	}
	keeper.SetNextProposalID(ctx, 6)

	// Test analytics
	analytics := keeper.GetGovernanceAnalytics(ctx)

	require.NotNil(t, analytics)
	require.Equal(t, uint64(5), analytics.TotalProposals)
	require.NotEmpty(t, analytics.ProposalsByStatus)
	require.NotEmpty(t, analytics.ProposalsByType)
	require.True(t, analytics.PassRate.GT(sdkmath.LegacyZeroDec()))
	require.True(t, analytics.VetoRate.GT(sdkmath.LegacyZeroDec()))
}

func TestCountProposalsByStatus(t *testing.T) {
	keeper, ctx := setupAnalyticsKeeper(t)

	ts, _ := gogotypes.TimestampProto(time.Now())

	// Create proposals with different statuses
	statuses := []types.ProposalStatus{
		types.StatusVotingPeriod,
		types.StatusVotingPeriod,
		types.StatusPassed,
		types.StatusFailed,
	}

	for i, status := range statuses {
		proposal := &types.Proposal{
			Id:          uint64(i + 1),
			Title:       "Test Proposal",
			Description: "Test Description",
			Proposer:    testAddr("proposer1"),
			Status:      status,
			Category:    types.CategoryText,
			SubmitTime:  ts,
		}
		keeper.SetProposal(ctx, proposal)
	}
	keeper.SetNextProposalID(ctx, 5)

	proposals := keeper.GetAllProposals(ctx)
	counts := keeper.countProposalsByStatus(proposals)

	require.Equal(t, uint64(2), counts[types.StatusVotingPeriod.String()])
	require.Equal(t, uint64(1), counts[types.StatusPassed.String()])
	require.Equal(t, uint64(1), counts[types.StatusFailed.String()])
}

func TestCountProposalsByType(t *testing.T) {
	keeper, ctx := setupAnalyticsKeeper(t)

	ts, _ := gogotypes.TimestampProto(time.Now())

	// Create proposals with different categories
	categories := []types.ProposalCategory{
		types.CategoryText,
		types.CategoryText,
		types.CategoryParameterChange,
		types.CategorySpending,
	}

	for i, category := range categories {
		proposal := &types.Proposal{
			Id:          uint64(i + 1),
			Title:       "Test Proposal",
			Description: "Test Description",
			Proposer:    testAddr("proposer1"),
			Status:      types.StatusVotingPeriod,
			Category:    category,
			SubmitTime:  ts,
		}
		keeper.SetProposal(ctx, proposal)
	}
	keeper.SetNextProposalID(ctx, 5)

	proposals := keeper.GetAllProposals(ctx)
	counts := keeper.countProposalsByType(proposals)

	require.Equal(t, uint64(2), counts[types.CategoryText.String()])
	require.Equal(t, uint64(1), counts[types.CategoryParameterChange.String()])
	require.Equal(t, uint64(1), counts[types.CategorySpending.String()])
}

func TestCalculateAverageVotingPower(t *testing.T) {
	keeper, ctx := setupAnalyticsKeeper(t)

	ts, _ := gogotypes.TimestampProto(time.Now())

	// Create proposals with tally results
	proposals := []*types.Proposal{
		{
			Id:          1,
			Title:       "Test Proposal 1",
			Description: "Test Description",
			Proposer:    testAddr("proposer1"),
			Status:      types.StatusPassed,
			Category:    types.CategoryText,
			SubmitTime:  ts,
			FinalTallyResult: &types.TallyResult{
				Yes:        "1000",
				No:         "200",
				Abstain:    "100",
				NoWithVeto: "50",
			},
		},
		{
			Id:          2,
			Title:       "Test Proposal 2",
			Description: "Test Description",
			Proposer:    testAddr("proposer1"),
			Status:      types.StatusPassed,
			Category:    types.CategoryText,
			SubmitTime:  ts,
			FinalTallyResult: &types.TallyResult{
				Yes:        "2000",
				No:         "500",
				Abstain:    "200",
				NoWithVeto: "100",
			},
		},
	}

	avgPower := keeper.calculateAverageVotingPower(ctx, proposals)
	require.NotEmpty(t, avgPower)
	require.NotEqual(t, "0", avgPower)
}

func TestCalculateAverageVotingPower_NoTally(t *testing.T) {
	keeper, ctx := setupAnalyticsKeeper(t)

	ts, _ := gogotypes.TimestampProto(time.Now())

	// Create proposals without tally results
	proposals := []*types.Proposal{
		{
			Id:               1,
			Title:            "Test Proposal 1",
			Description:      "Test Description",
			Proposer:         testAddr("proposer1"),
			Status:           types.StatusVotingPeriod,
			Category:         types.CategoryText,
			SubmitTime:       ts,
			FinalTallyResult: nil,
		},
	}

	avgPower := keeper.calculateAverageVotingPower(ctx, proposals)
	require.Equal(t, "0", avgPower)
}

func TestCalculatePassRate(t *testing.T) {
	keeper, _ := setupAnalyticsKeeper(t)

	ts, _ := gogotypes.TimestampProto(time.Now())

	// Test empty proposals
	passRate := keeper.calculatePassRate([]*types.Proposal{})
	require.True(t, passRate.Equal(sdkmath.LegacyZeroDec()))

	// Test with proposals
	proposals := []*types.Proposal{
		{Id: 1, Status: types.StatusPassed, SubmitTime: ts},
		{Id: 2, Status: types.StatusExecuted, SubmitTime: ts},
		{Id: 3, Status: types.StatusFailed, SubmitTime: ts},
		{Id: 4, Status: types.StatusRejected, SubmitTime: ts},
	}

	passRate = keeper.calculatePassRate(proposals)
	// 2 passed out of 4 = 50%
	require.True(t, passRate.Equal(sdkmath.LegacyNewDec(50)))
}

func TestCalculateVetoRate(t *testing.T) {
	keeper, _ := setupAnalyticsKeeper(t)

	ts, _ := gogotypes.TimestampProto(time.Now())

	// Test empty proposals
	vetoRate := keeper.calculateVetoRate([]*types.Proposal{})
	require.True(t, vetoRate.Equal(sdkmath.LegacyZeroDec()))

	// Test with proposals
	proposals := []*types.Proposal{
		{Id: 1, Status: types.StatusPassed, SubmitTime: ts},
		{Id: 2, Status: types.StatusExecuted, SubmitTime: ts},
		{Id: 3, Status: types.StatusFailed, SubmitTime: ts},
		{Id: 4, Status: types.StatusRejected, SubmitTime: ts},
	}

	vetoRate = keeper.calculateVetoRate(proposals)
	// 1 rejected out of 4 = 25%
	require.True(t, vetoRate.Equal(sdkmath.LegacyNewDec(25)))
}

func TestCalculateParticipationRate(t *testing.T) {
	keeper, ctx := setupAnalyticsKeeper(t)

	ts, _ := gogotypes.TimestampProto(time.Now())

	// Test empty proposals
	rate := keeper.calculateParticipationRate(ctx, []*types.Proposal{})
	require.True(t, rate.Equal(sdkmath.LegacyZeroDec()))

	// Create proposal with votes
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddr("proposer1"),
		Status:      types.StatusPassed,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal)

	// Add votes
	for i := 1; i <= 5; i++ {
		vote := &types.Vote{
			ProposalId:  1,
			Voter:       testAddr("voter" + string(rune('0'+i))),
			Option:      types.VoteOptionYes,
			VotingPower: "100",
			Timestamp:   ts,
		}
		keeper.SetVote(ctx, vote)
	}

	proposals := keeper.GetAllProposals(ctx)
	rate = keeper.calculateParticipationRate(ctx, proposals)
	require.True(t, rate.GT(sdkmath.LegacyZeroDec()))
}

func TestCalculateAverageProposalDuration(t *testing.T) {
	keeper, _ := setupAnalyticsKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	endTs, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))

	// Test empty proposals
	duration := keeper.calculateAverageProposalDuration([]*types.Proposal{})
	require.Equal(t, uint64(0), duration)

	// Test with proposals
	proposals := []*types.Proposal{
		{
			Id:            1,
			SubmitTime:    ts,
			VotingEndTime: endTs,
		},
	}

	duration = keeper.calculateAverageProposalDuration(proposals)
	require.Greater(t, duration, uint64(0))
}

func TestGetProposerAnalytics(t *testing.T) {
	keeper, ctx := setupAnalyticsKeeper(t)

	proposer := testAddr("proposer1")
	ts, _ := gogotypes.TimestampProto(time.Now())

	// Create proposals for the proposer
	proposals := []struct {
		id     uint64
		status types.ProposalStatus
	}{
		{1, types.StatusPassed},
		{2, types.StatusExecuted},
		{3, types.StatusFailed},
		{4, types.StatusRejected},
	}

	for _, p := range proposals {
		proposal := &types.Proposal{
			Id:          p.id,
			Title:       "Test Proposal",
			Description: "Test Description",
			Proposer:    proposer,
			Status:      p.status,
			Category:    types.CategoryText,
			SubmitTime:  ts,
		}
		keeper.SetProposal(ctx, proposal)
	}
	keeper.SetNextProposalID(ctx, 5)

	// Get proposer analytics
	analytics := keeper.GetProposerAnalytics(ctx, proposer)

	require.NotNil(t, analytics)
	require.Equal(t, proposer, analytics.Proposer)
	require.Equal(t, uint64(4), analytics.TotalProposals)
	require.Equal(t, uint64(1), analytics.PassedProposals)
	require.Equal(t, uint64(1), analytics.ExecutedProposals)
	require.Equal(t, uint64(1), analytics.FailedProposals)
	require.Equal(t, uint64(1), analytics.RejectedProposals)
	// 2 passed out of 4 = 50%
	require.True(t, analytics.SuccessRate.Equal(sdkmath.LegacyNewDec(50)))
}

func TestGetProposerAnalytics_NoProposals(t *testing.T) {
	keeper, ctx := setupAnalyticsKeeper(t)

	proposer := testAddr("proposer1")

	// Get analytics for non-existent proposer
	analytics := keeper.GetProposerAnalytics(ctx, proposer)

	require.NotNil(t, analytics)
	require.Equal(t, uint64(0), analytics.TotalProposals)
	require.True(t, analytics.SuccessRate.Equal(sdkmath.LegacyZeroDec()))
}

func TestCountByStatus(t *testing.T) {
	keeper, _ := setupAnalyticsKeeper(t)

	ts, _ := gogotypes.TimestampProto(time.Now())

	proposals := []*types.Proposal{
		{Id: 1, Status: types.StatusPassed, SubmitTime: ts},
		{Id: 2, Status: types.StatusPassed, SubmitTime: ts},
		{Id: 3, Status: types.StatusFailed, SubmitTime: ts},
		{Id: 4, Status: types.StatusRejected, SubmitTime: ts},
	}

	passedCount := keeper.countByStatus(proposals, types.StatusPassed)
	require.Equal(t, uint64(2), passedCount)

	failedCount := keeper.countByStatus(proposals, types.StatusFailed)
	require.Equal(t, uint64(1), failedCount)

	rejectedCount := keeper.countByStatus(proposals, types.StatusRejected)
	require.Equal(t, uint64(1), rejectedCount)
}
