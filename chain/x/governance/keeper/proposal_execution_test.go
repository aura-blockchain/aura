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
)

func setupExecutionKeeper(t *testing.T) (*Keeper, sdk.Context) {
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

// testAddrPE generates a valid bech32 address for testing (proposal execution)
func testAddrPE(name string) string {
	// Pad name to 20 bytes for valid AccAddress
	padded := name + "________________"
	return sdk.AccAddress(padded[:20]).String()
}

func TestTimestampFromTime(t *testing.T) {
	now := time.Now()
	ts := timestampFromTime(now)

	require.NotNil(t, ts)
	require.Equal(t, now.Unix(), ts.Seconds)
	require.Equal(t, int32(now.Nanosecond()), ts.Nanos)
}

func TestTimeFromTimestamp(t *testing.T) {
	now := time.Now()
	ts := &gogotypes.Timestamp{
		Seconds: now.Unix(),
		Nanos:   int32(now.Nanosecond()),
	}

	result := timeFromTimestamp(ts)
	require.Equal(t, now.Unix(), result.Unix())
}

func TestTimeFromTimestamp_Nil(t *testing.T) {
	result := timeFromTimestamp(nil)
	require.True(t, result.IsZero())
}

func TestProcessAutomaticExecutions_Disabled(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	// Set params with no execution delay (disabled)
	params := types.DefaultParams()
	params.ExecutionDelay = nil
	keeper.SetParams(ctx, params)

	executions, err := keeper.ProcessAutomaticExecutions(ctx)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrAutoExecutionDisabled)
	require.Nil(t, executions)
}

func TestProcessAutomaticExecutions_NoProposals(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	// Set params with execution delay enabled
	params := types.DefaultParams()
	params.ExecutionDelay = gogotypes.DurationProto(24 * time.Hour)
	keeper.SetParams(ctx, params)

	executions, err := keeper.ProcessAutomaticExecutions(ctx)
	require.NoError(t, err)
	require.NotNil(t, executions)
	require.Len(t, executions, 0)
}

func TestProcessAutomaticExecutions_PendingExecution(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	// Set params with execution delay enabled
	params := types.DefaultParams()
	params.ExecutionDelay = gogotypes.DurationProto(24 * time.Hour)
	keeper.SetParams(ctx, params)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	execTime := now.Add(-1 * time.Hour) // Execution time has passed
	execTs, _ := gogotypes.TimestampProto(execTime)

	// Create a passed proposal ready for execution
	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddrPE("proposer1"),
		Status:        types.StatusPassed,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		ExecutionTime: execTs,
	}
	keeper.SetProposal(ctx, proposal)

	executions, err := keeper.ProcessAutomaticExecutions(ctx)
	require.NoError(t, err)
	require.NotNil(t, executions)
	require.Len(t, executions, 1)
	require.Equal(t, uint64(1), executions[0].ProposalId)
	require.True(t, executions[0].Success)
}

func TestProcessAutomaticExecutions_FutureExecution(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	// Set params with execution delay enabled
	params := types.DefaultParams()
	params.ExecutionDelay = gogotypes.DurationProto(24 * time.Hour)
	keeper.SetParams(ctx, params)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	execTime := now.Add(24 * time.Hour) // Future execution time
	execTs, _ := gogotypes.TimestampProto(execTime)

	// Create a passed proposal not ready for execution
	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddrPE("proposer1"),
		Status:        types.StatusPassed,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		ExecutionTime: execTs,
	}
	keeper.SetProposal(ctx, proposal)

	executions, err := keeper.ProcessAutomaticExecutions(ctx)
	require.NoError(t, err)
	require.NotNil(t, executions)
	require.Len(t, executions, 0) // Should not execute
}

func TestProcessAutomaticExecutions_NoExecutionTime(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	// Set params with execution delay enabled
	params := types.DefaultParams()
	params.ExecutionDelay = gogotypes.DurationProto(24 * time.Hour)
	keeper.SetParams(ctx, params)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)

	// Create a passed proposal without execution time
	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddrPE("proposer1"),
		Status:        types.StatusPassed,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		ExecutionTime: nil,
	}
	keeper.SetProposal(ctx, proposal)

	executions, err := keeper.ProcessAutomaticExecutions(ctx)
	require.NoError(t, err)
	require.NotNil(t, executions)
	require.Len(t, executions, 0)
}

func TestProcessAutomaticExecutions_MultipleProposals(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	// Set params with execution delay enabled
	params := types.DefaultParams()
	params.ExecutionDelay = gogotypes.DurationProto(24 * time.Hour)
	keeper.SetParams(ctx, params)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	execTime := now.Add(-1 * time.Hour)
	execTs, _ := gogotypes.TimestampProto(execTime)

	// Create multiple passed proposals ready for execution
	for i := 1; i <= 3; i++ {
		proposal := &types.Proposal{
			Id:            uint64(i),
			Title:         "Test Proposal",
			Description:   "Test Description",
			Proposer:      testAddrPE("proposer1"),
			Status:        types.StatusPassed,
			Category:      types.CategoryText,
			SubmitTime:    ts,
			ExecutionTime: execTs,
		}
		keeper.SetProposal(ctx, proposal)
	}

	executions, err := keeper.ProcessAutomaticExecutions(ctx)
	require.NoError(t, err)
	require.NotNil(t, executions)
	require.Len(t, executions, 3)
}

func TestExecuteProposalAutomatically_Success(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	execTime := now.Add(-1 * time.Hour)
	execTs, _ := gogotypes.TimestampProto(execTime)

	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddrPE("proposer1"),
		Status:        types.StatusPassed,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		ExecutionTime: execTs,
	}

	execution, err := keeper.executeProposalAutomatically(ctx, proposal)
	require.NoError(t, err)
	require.NotNil(t, execution)
	require.Equal(t, uint64(1), execution.ProposalId)
	require.True(t, execution.Success)
	require.Empty(t, execution.ErrorMessage)

	// Verify proposal status updated
	require.Equal(t, types.StatusExecuted, proposal.Status)
}

func TestExecuteProposalAutomatically_EventEmitted(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	execTime := now.Add(-1 * time.Hour)
	execTs, _ := gogotypes.TimestampProto(execTime)

	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddrPE("proposer1"),
		Status:        types.StatusPassed,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		ExecutionTime: execTs,
	}

	_, err := keeper.executeProposalAutomatically(ctx, proposal)
	require.NoError(t, err)

	// Verify event was emitted
	events := ctx.EventManager().Events()
	hasExecutionEvent := false
	for _, event := range events {
		if event.Type == "proposal_auto_executed" {
			hasExecutionEvent = true
			attrs := make(map[string]string)
			for _, attr := range event.Attributes {
				attrs[attr.Key] = attr.Value
			}
			require.Equal(t, "1", attrs["proposal_id"])
			require.Equal(t, "true", attrs["success"])
		}
	}
	require.True(t, hasExecutionEvent)
}

func TestExecuteParameterChangeAuto(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)

	proposal := &types.Proposal{
		Id:          1,
		Title:       "Parameter Change",
		Description: "Change params",
		Proposer:    testAddrPE("proposer1"),
		Status:      types.StatusPassed,
		Category:    types.CategoryParameterChange,
		SubmitTime:  ts,
	}

	result, err := keeper.executeParameterChangeAuto(ctx, proposal)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Success)
	require.Equal(t, "Parameter changes applied", result.Message)
}

func TestExecuteSoftwareUpgradeAuto(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)

	proposal := &types.Proposal{
		Id:          1,
		Title:       "Software Upgrade",
		Description: "Upgrade to v2",
		Proposer:    testAddrPE("proposer1"),
		Status:      types.StatusPassed,
		Category:    types.CategorySoftwareUpgrade,
		SubmitTime:  ts,
	}

	result, err := keeper.executeSoftwareUpgradeAuto(ctx, proposal)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Success)
	require.Equal(t, "Software upgrade scheduled", result.Message)
}

func TestExecuteCommunitySpendAuto(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)

	proposal := &types.Proposal{
		Id:          1,
		Title:       "Community Spend",
		Description: "Fund project",
		Proposer:    testAddrPE("proposer1"),
		Status:      types.StatusPassed,
		Category:    types.CategorySpending,
		SubmitTime:  ts,
	}

	result, err := keeper.executeCommunitySpendAuto(ctx, proposal)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Success)
	require.Equal(t, "Community spend executed", result.Message)
}

func TestScheduleProposalExecution_Success(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)

	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddrPE("proposer1"),
		Status:      types.StatusPassed,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal)

	executionDelay := uint64(86400) // 24 hours in seconds

	err := keeper.ScheduleProposalExecution(ctx, 1, executionDelay)
	require.NoError(t, err)

	// Verify execution time was set
	updatedProposal, err := keeper.GetProposal(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, updatedProposal.ExecutionTime)

	// Verify event was emitted
	events := ctx.EventManager().Events()
	hasScheduleEvent := false
	for _, event := range events {
		if event.Type == "proposal_execution_scheduled" {
			hasScheduleEvent = true
		}
	}
	require.True(t, hasScheduleEvent)
}

func TestScheduleProposalExecution_ProposalNotFound(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	err := keeper.ScheduleProposalExecution(ctx, 999, 86400)
	require.Error(t, err)
}

func TestScheduleProposalExecution_InvalidStatus(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)

	// Create proposal in voting period (not passed)
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddrPE("proposer1"),
		Status:      types.StatusVotingPeriod,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal)

	err := keeper.ScheduleProposalExecution(ctx, 1, 86400)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidProposalStatus)
}

func TestCancelScheduledExecution_Success(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	execTime := now.Add(24 * time.Hour)
	execTs, _ := gogotypes.TimestampProto(execTime)

	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddrPE("proposer1"),
		Status:        types.StatusPassed,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		ExecutionTime: execTs,
	}
	keeper.SetProposal(ctx, proposal)

	err := keeper.CancelScheduledExecution(ctx, 1, "Security concern")
	require.NoError(t, err)

	// Verify execution time was cleared
	updatedProposal, err := keeper.GetProposal(ctx, 1)
	require.NoError(t, err)
	require.Nil(t, updatedProposal.ExecutionTime)

	// Verify event was emitted
	events := ctx.EventManager().Events()
	hasCancelEvent := false
	for _, event := range events {
		if event.Type == "proposal_execution_cancelled" {
			hasCancelEvent = true
			attrs := make(map[string]string)
			for _, attr := range event.Attributes {
				attrs[attr.Key] = attr.Value
			}
			require.Equal(t, "Security concern", attrs["reason"])
		}
	}
	require.True(t, hasCancelEvent)
}

func TestCancelScheduledExecution_NoScheduledExecution(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)

	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddrPE("proposer1"),
		Status:        types.StatusPassed,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		ExecutionTime: nil, // No scheduled execution
	}
	keeper.SetProposal(ctx, proposal)

	err := keeper.CancelScheduledExecution(ctx, 1, "test")
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrNoScheduledExecution)
}

func TestCancelScheduledExecution_ProposalNotFound(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	err := keeper.CancelScheduledExecution(ctx, 999, "test")
	require.Error(t, err)
}

func TestGetPendingExecutions_Empty(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	pending := keeper.GetPendingExecutions(ctx)
	require.NotNil(t, pending)
	require.Len(t, pending, 0)
}

func TestGetPendingExecutions_WithPending(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	futureTime := now.Add(24 * time.Hour)
	futureTs, _ := gogotypes.TimestampProto(futureTime)

	// Create pending proposal
	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddrPE("proposer1"),
		Status:        types.StatusPassed,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		ExecutionTime: futureTs,
	}
	keeper.SetProposal(ctx, proposal)

	pending := keeper.GetPendingExecutions(ctx)
	require.NotNil(t, pending)
	require.Len(t, pending, 1)
	require.Equal(t, uint64(1), pending[0].Id)
}

func TestGetPendingExecutions_ExcludesReady(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	pastTime := now.Add(-1 * time.Hour)
	pastTs, _ := gogotypes.TimestampProto(pastTime)

	// Create proposal ready for execution (not pending)
	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddrPE("proposer1"),
		Status:        types.StatusPassed,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		ExecutionTime: pastTs,
	}
	keeper.SetProposal(ctx, proposal)

	pending := keeper.GetPendingExecutions(ctx)
	require.NotNil(t, pending)
	require.Len(t, pending, 0) // Should not include ready proposals
}

func TestGetPendingExecutions_Mixed(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	futureTime := now.Add(24 * time.Hour)
	futureTs, _ := gogotypes.TimestampProto(futureTime)
	pastTime := now.Add(-1 * time.Hour)
	pastTs, _ := gogotypes.TimestampProto(pastTime)

	// Create pending proposal
	proposal1 := &types.Proposal{
		Id:            1,
		Title:         "Pending Proposal",
		Description:   "Test Description",
		Proposer:      testAddrPE("proposer1"),
		Status:        types.StatusPassed,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		ExecutionTime: futureTs,
	}
	keeper.SetProposal(ctx, proposal1)

	// Create ready proposal
	proposal2 := &types.Proposal{
		Id:            2,
		Title:         "Ready Proposal",
		Description:   "Test Description",
		Proposer:      testAddrPE("proposer1"),
		Status:        types.StatusPassed,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		ExecutionTime: pastTs,
	}
	keeper.SetProposal(ctx, proposal2)

	pending := keeper.GetPendingExecutions(ctx)
	require.NotNil(t, pending)
	require.Len(t, pending, 1)
	require.Equal(t, uint64(1), pending[0].Id)
}

func TestGetExecutionStatistics_Empty(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	stats := keeper.GetExecutionStatistics(ctx)
	require.NotNil(t, stats)
	require.Equal(t, uint64(0), stats.TotalExecuted)
	require.Equal(t, uint64(0), stats.SuccessfulExecutions)
	require.Equal(t, uint64(0), stats.FailedExecutions)
	require.Equal(t, uint64(0), stats.PendingExecutions)
	require.Equal(t, uint64(0), stats.AverageGasUsed)
	require.Equal(t, uint64(0), stats.TotalGasUsed)
}

func TestGetExecutionStatistics_WithExecuted(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)

	// Create executed proposal
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddrPE("proposer1"),
		Status:      types.StatusExecuted,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal)

	stats := keeper.GetExecutionStatistics(ctx)
	require.NotNil(t, stats)
	require.Equal(t, uint64(1), stats.TotalExecuted)
	require.Equal(t, uint64(1), stats.SuccessfulExecutions)
	require.Equal(t, uint64(0), stats.FailedExecutions)
}

func TestGetExecutionStatistics_WithFailed(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	execTs, _ := gogotypes.TimestampProto(now.Add(1 * time.Hour))

	// Create failed proposal with execution time
	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddrPE("proposer1"),
		Status:        types.StatusFailed,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		ExecutionTime: execTs,
	}
	keeper.SetProposal(ctx, proposal)

	stats := keeper.GetExecutionStatistics(ctx)
	require.NotNil(t, stats)
	require.Equal(t, uint64(1), stats.TotalExecuted)
	require.Equal(t, uint64(0), stats.SuccessfulExecutions)
	require.Equal(t, uint64(1), stats.FailedExecutions)
}

func TestGetExecutionStatistics_WithPending(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	futureTs, _ := gogotypes.TimestampProto(now.Add(24 * time.Hour))

	// Create pending proposal
	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddrPE("proposer1"),
		Status:        types.StatusPassed,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		ExecutionTime: futureTs,
	}
	keeper.SetProposal(ctx, proposal)

	stats := keeper.GetExecutionStatistics(ctx)
	require.NotNil(t, stats)
	require.Equal(t, uint64(0), stats.TotalExecuted)
	require.Equal(t, uint64(1), stats.PendingExecutions)
}

func TestGetExecutionStatistics_Mixed(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	futureTs, _ := gogotypes.TimestampProto(now.Add(24 * time.Hour))
	execTs, _ := gogotypes.TimestampProto(now.Add(1 * time.Hour))

	// Create executed proposal
	proposal1 := &types.Proposal{
		Id:          1,
		Title:       "Executed Proposal",
		Description: "Test Description",
		Proposer:    testAddrPE("proposer1"),
		Status:      types.StatusExecuted,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal1)

	// Create failed proposal with execution time
	proposal2 := &types.Proposal{
		Id:            2,
		Title:         "Failed Proposal",
		Description:   "Test Description",
		Proposer:      testAddrPE("proposer1"),
		Status:        types.StatusFailed,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		ExecutionTime: execTs,
	}
	keeper.SetProposal(ctx, proposal2)

	// Create pending proposal
	proposal3 := &types.Proposal{
		Id:            3,
		Title:         "Pending Proposal",
		Description:   "Test Description",
		Proposer:      testAddrPE("proposer1"),
		Status:        types.StatusPassed,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		ExecutionTime: futureTs,
	}
	keeper.SetProposal(ctx, proposal3)

	stats := keeper.GetExecutionStatistics(ctx)
	require.NotNil(t, stats)
	require.Equal(t, uint64(2), stats.TotalExecuted)
	require.Equal(t, uint64(1), stats.SuccessfulExecutions)
	require.Equal(t, uint64(1), stats.FailedExecutions)
	require.Equal(t, uint64(1), stats.PendingExecutions)
}

func TestScheduleProposalExecution_CorrectTiming(t *testing.T) {
	keeper, ctx := setupExecutionKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)

	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    testAddrPE("proposer1"),
		Status:      types.StatusPassed,
		Category:    types.CategoryText,
		SubmitTime:  ts,
	}
	keeper.SetProposal(ctx, proposal)

	executionDelay := uint64(3600) // 1 hour in seconds

	err := keeper.ScheduleProposalExecution(ctx, 1, executionDelay)
	require.NoError(t, err)

	// Verify execution time is approximately correct
	updatedProposal, err := keeper.GetProposal(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, updatedProposal.ExecutionTime)

	execTime := timeFromTimestamp(updatedProposal.ExecutionTime)
	expectedTime := ctx.BlockTime().Add(time.Duration(executionDelay) * time.Second)

	// Allow small timing difference
	timeDiff := execTime.Sub(expectedTime)
	require.Less(t, timeDiff.Abs().Seconds(), float64(2))
}
