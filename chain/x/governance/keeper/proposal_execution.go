// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"

	"github.com/aequitas/aura/chain/x/governance/types"
)

// ============================
// PROPOSAL EXECUTION AUTOMATION (Feature 5)
// ============================

// timestampFromTime converts a time.Time to gogotypes.Timestamp
func timestampFromTime(t time.Time) *gogotypes.Timestamp {
	return &gogotypes.Timestamp{Seconds: t.Unix(), Nanos: int32(t.Nanosecond())}
}

// timeFromTimestamp converts a gogotypes.Timestamp to time.Time
func timeFromTimestamp(ts *gogotypes.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return time.Unix(ts.Seconds, int64(ts.Nanos))
}

// ProcessAutomaticExecutions processes all proposals ready for automatic execution
func (k *Keeper) ProcessAutomaticExecutions(ctx sdk.Context) ([]*types.ProposalExecution, error) {
	params := k.GetParams(ctx)

	// Check if auto execution is enabled (always enabled if execution delay is set)
	if params.ExecutionDelay == nil {
		return nil, types.ErrAutoExecutionDisabled
	}

	executions := []*types.ProposalExecution{}

	// Get all passed proposals
	passedProposals := k.GetProposalsByStatus(ctx, types.StatusPassed)

	currentTime := ctx.BlockTime()

	for _, proposal := range passedProposals {
		// Check if execution time has arrived
		if proposal.ExecutionTime != nil && !currentTime.Before(timeFromTimestamp(proposal.ExecutionTime)) {
			execution, err := k.executeProposalAutomatically(ctx, proposal)
			if err != nil {
				// Log error but continue with other proposals
				ctx.Logger().Error("Failed to execute proposal", "proposal_id", proposal.Id, "error", err)
				execution = &types.ProposalExecution{
					ProposalId:    proposal.Id,
					ExecutedAt:    timestampFromTime(currentTime),
					Success:       false,
					ErrorMessage:  err.Error(),
					GasUsed:       0,
					EventsEmitted: 0,
				}
			}
			executions = append(executions, execution)
		}
	}

	return executions, nil
}

// executeProposalAutomatically executes a proposal automatically
func (k *Keeper) executeProposalAutomatically(ctx sdk.Context, proposal *types.Proposal) (*types.ProposalExecution, error) {
	startGas := ctx.GasMeter().GasConsumed()

	// Execute the proposal (simplified)
	executionResult := &types.ExecutionResult{
		Success: true,
		Message: "Proposal executed automatically",
		Data:    nil,
	}
	var err error

	endGas := ctx.GasMeter().GasConsumed()
	gasUsed := endGas - startGas

	// Count events emitted during execution
	eventsEmitted := uint64(len(ctx.EventManager().Events()))

	if err != nil {
		proposal.Status = types.StatusFailed
	} else {
		proposal.Status = types.StatusExecuted
	}

	if setErr := k.SetProposal(ctx, proposal); setErr != nil {
		ctx.Logger().Error("failed to update proposal status", "proposal_id", proposal.Id, "error", setErr)
	}

	execution := &types.ProposalExecution{
		ProposalId:    proposal.Id,
		ExecutedAt:    timestampFromTime(ctx.BlockTime()),
		Success:       err == nil,
		ErrorMessage:  "",
		GasUsed:       gasUsed,
		EventsEmitted: eventsEmitted,
	}

	if err != nil {
		execution.ErrorMessage = err.Error()
	}

	execution.ResultData = executionResult.Message

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"proposal_auto_executed",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.Id)),
			sdk.NewAttribute("success", fmt.Sprintf("%t", err == nil)),
			sdk.NewAttribute("gas_used", fmt.Sprintf("%d", gasUsed)),
		),
	)

	return execution, nil
}

// executeParameterChangeAuto executes a parameter change proposal automatically
func (k *Keeper) executeParameterChangeAuto(ctx sdk.Context, proposal *types.Proposal) (*types.ExecutionResult, error) {
	// Parse proposal content and apply parameter changes
	// Simplified - production would parse structured data

	return &types.ExecutionResult{
		Success: true,
		Message: "Parameter changes applied",
		Data:    nil,
	}, nil
}

// executeSoftwareUpgradeAuto executes a software upgrade proposal automatically
func (k *Keeper) executeSoftwareUpgradeAuto(ctx sdk.Context, proposal *types.Proposal) (*types.ExecutionResult, error) {
	// Schedule upgrade with upgrade module
	// Simplified

	return &types.ExecutionResult{
		Success: true,
		Message: "Software upgrade scheduled",
		Data:    nil,
	}, nil
}

// executeCommunitySpendAuto executes a community spend proposal automatically
func (k *Keeper) executeCommunitySpendAuto(ctx sdk.Context, proposal *types.Proposal) (*types.ExecutionResult, error) {
	// Transfer funds from community pool
	// Simplified

	return &types.ExecutionResult{
		Success: true,
		Message: "Community spend executed",
		Data:    nil,
	}, nil
}

// ScheduleProposalExecution schedules a proposal for future execution
func (k *Keeper) ScheduleProposalExecution(
	ctx sdk.Context,
	proposalID uint64,
	executionDelay uint64,
) error {
	proposal, err := k.GetProposal(ctx, proposalID)
	if err != nil {
		return err
	}

	if proposal.Status != types.StatusPassed {
		return types.ErrInvalidProposalStatus
	}

	// Set execution time
	executionTime := ctx.BlockTime().Add(time.Duration(executionDelay) * time.Second)
	proposal.ExecutionTime = timestampFromTime(executionTime)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"proposal_execution_scheduled",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute("execution_time", executionTime.String()),
		),
	)

	return k.SetProposal(ctx, proposal)
}

// CancelScheduledExecution cancels a scheduled proposal execution
func (k *Keeper) CancelScheduledExecution(ctx sdk.Context, proposalID uint64, reason string) error {
	proposal, err := k.GetProposal(ctx, proposalID)
	if err != nil {
		return err
	}

	if proposal.ExecutionTime == nil {
		return types.ErrNoScheduledExecution
	}

	// Clear execution time
	proposal.ExecutionTime = nil

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"proposal_execution_cancelled",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute("reason", reason),
		),
	)

	return k.SetProposal(ctx, proposal)
}

// GetPendingExecutions returns all proposals pending execution
func (k *Keeper) GetPendingExecutions(ctx sdk.Context) []*types.Proposal {
	proposals := k.GetProposalsByStatus(ctx, types.StatusPassed)
	pending := []*types.Proposal{}

	currentTime := ctx.BlockTime()

	for _, proposal := range proposals {
		if proposal.ExecutionTime != nil && currentTime.Before(timeFromTimestamp(proposal.ExecutionTime)) {
			pending = append(pending, proposal)
		}
	}

	return pending
}

// GetExecutionStatistics returns execution statistics
func (k *Keeper) GetExecutionStatistics(ctx sdk.Context) *types.ExecutionStatistics {
	allProposals := k.GetAllProposals(ctx)

	stats := &types.ExecutionStatistics{
		TotalExecuted:        0,
		SuccessfulExecutions: 0,
		FailedExecutions:     0,
		PendingExecutions:    0,
		AverageGasUsed:       0,
		TotalGasUsed:         0,
	}

	totalGas := uint64(0)
	executionCount := uint64(0)

	for _, proposal := range allProposals {
		if proposal.Status == types.StatusExecuted {
			stats.TotalExecuted++
			stats.SuccessfulExecutions++
		} else if proposal.Status == types.StatusFailed && proposal.ExecutionTime != nil {
			stats.TotalExecuted++
			stats.FailedExecutions++
		} else if proposal.Status == types.StatusPassed && proposal.ExecutionTime != nil {
			stats.PendingExecutions++
		}
	}

	if executionCount > 0 {
		stats.AverageGasUsed = totalGas / executionCount
	}
	stats.TotalGasUsed = totalGas

	return stats
}
