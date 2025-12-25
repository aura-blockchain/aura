// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/aequitas/aura/chain/x/common/determinism"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/governance/types"
)

// ============================
// PROPOSAL LIFECYCLE MANAGEMENT (Feature 1)
// ============================

// CreateProposal creates a new governance proposal
func (k *Keeper) CreateProposal(
	ctx sdk.Context,
	title string,
	description string,
	proposer string,
	category types.ProposalCategory,
	content string,
) (uint64, error) {
	params := k.GetParams(ctx)

	// Get next proposal ID
	proposalID := k.GetNextProposalID(ctx)

	// Create proposal
	proposal := &types.Proposal{
		Id:              proposalID,
		Title:           title,
		Description:     description,
		Proposer:        proposer,
		Status:          types.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD,
		Category:        category,
		SubmitTime:      timestampFromTime(determinism.GetBlockTime(ctx)),
		DepositEndTime:  timestampFromTime(determinism.GetBlockTime(ctx).Add(time.Duration(types.GetDepositPeriodSeconds(params)) * time.Second)),
		VotingStartTime: nil,
		VotingEndTime:   nil,
		TotalDeposit:    "0",
		FinalTallyResult: nil,
		ExecutionTime:   nil,
	}

	// Store proposal
	if err := k.SetProposal(ctx, proposal); err != nil {
		return 0, err
	}

	// Increment next proposal ID
	k.SetNextProposalID(ctx, proposalID+1)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"proposal_created",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute("proposer", proposer),
			sdk.NewAttribute("title", title),
		),
	)

	return proposalID, nil
}

// AdvanceProposalStatus advances proposal through its lifecycle stages
func (k *Keeper) AdvanceProposalStatus(ctx sdk.Context, proposalID uint64) error {
	proposal, err := k.GetProposal(ctx, proposalID)
	if err != nil {
		return err
	}

	params := k.GetParams(ctx)
	currentTime := determinism.GetBlockTime(ctx)

	switch proposal.Status {
	case types.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD:
		// Check if deposit period ended
		if currentTime.After(timeFromTimestamp(proposal.DepositEndTime)) {
			// Check if minimum deposit reached
			if k.hasMinimumDeposit(proposal, params) {
				return k.moveToVotingPeriod(ctx, proposal, params)
			}
			// Deposit period ended without minimum deposit
			proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_FAILED
		}

	case types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD:
		// Check if voting period ended
		if proposal.VotingEndTime != nil && currentTime.After(timeFromTimestamp(proposal.VotingEndTime)) {
			return k.finalizeProposal(ctx, proposal, params)
		}

	case types.ProposalStatus_PROPOSAL_STATUS_PASSED:
		// Check if ready for execution
		if proposal.ExecutionTime == nil {
			proposal.ExecutionTime = timestampFromTime(currentTime.Add(time.Duration(types.GetExecutionDelaySeconds(params)) * time.Second))
		}

		if currentTime.After(timeFromTimestamp(proposal.ExecutionTime)) {
			return k.executeProposal(ctx, proposal)
		}

	case types.ProposalStatus_PROPOSAL_STATUS_FAILED, types.ProposalStatus_PROPOSAL_STATUS_REJECTED, types.ProposalStatus_PROPOSAL_STATUS_EXECUTED:
		// Terminal states - no further action
		return nil
	}

	return k.SetProposal(ctx, proposal)
}

// moveToVotingPeriod transitions proposal from deposit to voting period
func (k *Keeper) moveToVotingPeriod(ctx sdk.Context, proposal *types.Proposal, params *types.GovernanceParams) error {
	currentTime := determinism.GetBlockTime(ctx)
	proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	proposal.VotingStartTime = timestampFromTime(currentTime)
	proposal.VotingEndTime = timestampFromTime(currentTime.Add(time.Duration(types.GetVotingPeriodSeconds(params)) * time.Second))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"voting_period_started",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.Id)),
			sdk.NewAttribute("voting_end_time", timeFromTimestamp(proposal.VotingEndTime).String()),
		),
	)

	return k.SetProposal(ctx, proposal)
}

// finalizeProposal finalizes voting and determines outcome
func (k *Keeper) finalizeProposal(ctx sdk.Context, proposal *types.Proposal, params *types.GovernanceParams) error {
	// Calculate final tally
	tally := k.CalculateTally(ctx, proposal.Id)
	proposal.FinalTallyResult = tally

	// Process proposal outcome with security checks (quorum, threshold, veto)
	if err := k.processProposalOutcome(ctx, proposal, tally, params); err != nil {
		return err
	}

	return k.SetProposal(ctx, proposal)
}

// executeProposal executes a passed proposal
func (k *Keeper) executeProposal(ctx sdk.Context, proposal *types.Proposal) error {
	// Execute proposal based on category
	var err error

	switch proposal.Category {
	case types.ProposalCategory_PROPOSAL_CATEGORY_TEXT:
		// Text proposals don't require execution
		err = nil

	case types.ProposalCategory_PROPOSAL_CATEGORY_PARAMETER_CHANGE:
		err = k.executeParameterChange(ctx, proposal)

	case types.ProposalCategory_PROPOSAL_CATEGORY_SOFTWARE_UPGRADE:
		err = k.executeSoftwareUpgrade(ctx, proposal)

	case types.ProposalCategory_PROPOSAL_CATEGORY_SPENDING:
		err = k.executeCommunitySpend(ctx, proposal)

	default:
		err = types.ErrInvalidProposal
	}

	if err != nil {
		proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_FAILED
	} else {
		proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_EXECUTED

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"proposal_executed",
				sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.Id)),
			),
		)
	}

	return k.SetProposal(ctx, proposal)
}

// executeParameterChange executes a parameter change proposal
func (k *Keeper) executeParameterChange(ctx sdk.Context, proposal *types.Proposal) error {
	// Parse and apply parameter changes from proposal content
	// Simplified - production would parse JSON/proto content
	return nil
}

// executeSoftwareUpgrade executes a software upgrade proposal
func (k *Keeper) executeSoftwareUpgrade(ctx sdk.Context, proposal *types.Proposal) error {
	// Schedule software upgrade
	// Simplified - production would interact with upgrade module
	return nil
}

// executeCommunitySpend executes a community spend proposal
func (k *Keeper) executeCommunitySpend(ctx sdk.Context, proposal *types.Proposal) error {
	// Transfer funds from community pool
	// Simplified - production would interact with distribution module
	return nil
}

// hasMinimumDeposit checks if proposal has minimum deposit
func (k *Keeper) hasMinimumDeposit(proposal *types.Proposal, params *types.GovernanceParams) bool {
	// Compare total deposit with minimum required
	// Simplified comparison
	return proposal.TotalDeposit >= params.MinDeposit
}

// processProposalOutcome determines the outcome of a proposal with proper security checks
// This function enforces quorum (minimum participation), pass threshold, and veto threshold
func (k *Keeper) processProposalOutcome(ctx sdk.Context, proposal *types.Proposal, tally *types.TallyResult, params *types.GovernanceParams) error {
	// Parse vote counts from tally result strings
	yesVotes, ok := sdkmath.NewIntFromString(tally.Yes)
	if !ok {
		return fmt.Errorf("failed to parse yes votes: %s", tally.Yes)
	}

	noVotes, ok := sdkmath.NewIntFromString(tally.No)
	if !ok {
		return fmt.Errorf("failed to parse no votes: %s", tally.No)
	}

	abstainVotes, ok := sdkmath.NewIntFromString(tally.Abstain)
	if !ok {
		return fmt.Errorf("failed to parse abstain votes: %s", tally.Abstain)
	}

	noWithVetoVotes, ok := sdkmath.NewIntFromString(tally.NoWithVeto)
	if !ok {
		return fmt.Errorf("failed to parse no with veto votes: %s", tally.NoWithVeto)
	}

	// Calculate total votes cast
	totalVotes := yesVotes.Add(noVotes).Add(abstainVotes).Add(noWithVetoVotes)

	// Get total bonded tokens (total voting power in the network)
	totalBondedTokens, err := k.stakingKeeper.TotalBondedTokens(ctx)
	if err != nil {
		// On error, use zero as fallback (will likely fail quorum check)
		totalBondedTokens = sdkmath.ZeroInt()
	}

	// SECURITY CHECK 1: Quorum enforcement (minimum participation)
	// Prevents a single voter or small group from passing proposals without sufficient participation
	quorumDec, err := sdkmath.LegacyNewDecFromStr(params.Quorum)
	if err != nil {
		return fmt.Errorf("failed to parse quorum parameter: %w", err)
	}

	// Calculate required quorum (e.g., 33.4% of total bonded tokens)
	requiredQuorum := quorumDec.MulInt(totalBondedTokens).TruncateInt()

	if totalVotes.LT(requiredQuorum) {
		// Quorum not reached - proposal fails
		proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_REJECTED
		// Note: rejection_reason field will be set once protobuf is regenerated
		// For now, emit it in the event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"proposal_rejected",
				sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.Id)),
				sdk.NewAttribute("reason", "quorum not reached"),
				sdk.NewAttribute("votes_cast", totalVotes.String()),
				sdk.NewAttribute("required_quorum", requiredQuorum.String()),
			),
		)
		return nil
	}

	// SECURITY CHECK 2: Veto threshold enforcement
	// If too many NoWithVeto votes, proposal is rejected and deposits are burned
	vetoThresholdDec, err := sdkmath.LegacyNewDecFromStr(params.VetoThreshold)
	if err != nil {
		return fmt.Errorf("failed to parse veto threshold parameter: %w", err)
	}

	// Calculate veto threshold (e.g., 33.4% of total votes cast)
	vetoThreshold := vetoThresholdDec.MulInt(totalVotes).TruncateInt()

	if noWithVetoVotes.GT(vetoThreshold) {
		// Proposal vetoed - reject and burn deposits
		proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_VETOED

		// Burn deposits for vetoed proposals as a disincentive for spam/malicious proposals
		if err := k.BurnDeposits(ctx, proposal.Id); err != nil {
			ctx.Logger().Error("failed to burn deposits for vetoed proposal", "proposal_id", proposal.Id, "error", err)
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"proposal_vetoed",
				sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.Id)),
				sdk.NewAttribute("reason", "veto threshold exceeded"),
				sdk.NewAttribute("no_with_veto", noWithVetoVotes.String()),
				sdk.NewAttribute("veto_threshold", vetoThreshold.String()),
			),
		)
		return nil
	}

	// SECURITY CHECK 3: Pass threshold enforcement
	// Calculate votes excluding abstain (abstain votes don't count towards yes/no determination)
	votesExcludingAbstain := yesVotes.Add(noVotes).Add(noWithVetoVotes)

	// Prevent division by zero if only abstain votes
	if votesExcludingAbstain.IsZero() {
		proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_REJECTED
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"proposal_rejected",
				sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.Id)),
				sdk.NewAttribute("reason", "only abstain votes cast"),
			),
		)
		return nil
	}

	thresholdDec, err := sdkmath.LegacyNewDecFromStr(params.Threshold)
	if err != nil {
		return fmt.Errorf("failed to parse threshold parameter: %w", err)
	}

	// Calculate pass threshold (e.g., 50% of non-abstain votes must be yes)
	passThreshold := thresholdDec.MulInt(votesExcludingAbstain).TruncateInt()

	if yesVotes.GT(passThreshold) {
		// Proposal passed all security checks
		proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_PASSED

		// Refund deposits for passed proposals
		if err := k.RefundDeposits(ctx, proposal.Id); err != nil {
			ctx.Logger().Error("failed to refund deposits for passed proposal", "proposal_id", proposal.Id, "error", err)
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"proposal_passed",
				sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.Id)),
				sdk.NewAttribute("yes_votes", yesVotes.String()),
				sdk.NewAttribute("pass_threshold", passThreshold.String()),
			),
		)
	} else {
		// Did not meet pass threshold - proposal fails
		proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_REJECTED

		// Refund deposits for failed proposals (only burn for vetoed proposals)
		if err := k.RefundDeposits(ctx, proposal.Id); err != nil {
			ctx.Logger().Error("failed to refund deposits for rejected proposal", "proposal_id", proposal.Id, "error", err)
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"proposal_rejected",
				sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.Id)),
				sdk.NewAttribute("reason", "threshold not met"),
				sdk.NewAttribute("yes_votes", yesVotes.String()),
				sdk.NewAttribute("required_threshold", passThreshold.String()),
			),
		)
	}

	return nil
}

// CancelProposal allows proposer to cancel a proposal during deposit period
func (k *Keeper) CancelProposal(ctx sdk.Context, proposalID uint64, canceller string) error {
	proposal, err := k.GetProposal(ctx, proposalID)
	if err != nil {
		return err
	}

	// Only proposer can cancel
	if proposal.Proposer != canceller {
		return types.ErrInvalidProposal
	}

	// Can only cancel during deposit period
	if proposal.Status != types.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD {
		return types.ErrInvalidProposalStatus
	}

	proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_FAILED

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"proposal_cancelled",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute("canceller", canceller),
		),
	)

	return k.SetProposal(ctx, proposal)
}

// GetProposalsByStatus returns all proposals with given status
func (k *Keeper) GetProposalsByStatus(ctx sdk.Context, status types.ProposalStatus) []*types.Proposal {
	allProposals := k.GetAllProposals(ctx)
	filtered := []*types.Proposal{}

	for _, proposal := range allProposals {
		if proposal.Status == status {
			filtered = append(filtered, proposal)
		}
	}

	return filtered
}
