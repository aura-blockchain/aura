package keeper

import (
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/common/determinism"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

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
		SubmitTime:      timestamppb.New(determinism.GetBlockTime(ctx)),
		DepositEndTime:  timestamppb.New(determinism.GetBlockTime(ctx).Add(time.Duration(types.GetDepositPeriodSeconds(params)) * time.Second)),
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
		if currentTime.After(proposal.DepositEndTime.AsTime()) {
			// Check if minimum deposit reached
			if k.hasMinimumDeposit(proposal, params) {
				return k.moveToVotingPeriod(ctx, proposal, params)
			}
			// Deposit period ended without minimum deposit
			proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_FAILED
		}

	case types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD:
		// Check if voting period ended
		if proposal.VotingEndTime != nil && currentTime.After(proposal.VotingEndTime.AsTime()) {
			return k.finalizeProposal(ctx, proposal, params)
		}

	case types.ProposalStatus_PROPOSAL_STATUS_PASSED:
		// Check if ready for execution
		if proposal.ExecutionTime == nil {
			proposal.ExecutionTime = timestamppb.New(currentTime.Add(time.Duration(types.GetExecutionDelaySeconds(params)) * time.Second))
		}

		if currentTime.After(proposal.ExecutionTime.AsTime()) {
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
	proposal.VotingStartTime = timestamppb.New(currentTime)
	proposal.VotingEndTime = timestamppb.New(currentTime.Add(time.Duration(types.GetVotingPeriodSeconds(params)) * time.Second))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"voting_period_started",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.Id)),
			sdk.NewAttribute("voting_end_time", proposal.VotingEndTime.AsTime().String()),
		),
	)

	return k.SetProposal(ctx, proposal)
}

// finalizeProposal finalizes voting and determines outcome
func (k *Keeper) finalizeProposal(ctx sdk.Context, proposal *types.Proposal, params *types.GovernanceParams) error {
	// Calculate final tally
	tally := k.CalculateTally(ctx, proposal.Id)
	proposal.FinalTallyResult = tally

	// Determine if proposal passed
	if k.proposalPassed(tally, params) {
		proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_PASSED

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"proposal_passed",
				sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.Id)),
			),
		)
	} else {
		// Check if vetoed
		if k.proposalVetoed(tally, params) {
			proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_REJECTED
		} else {
			proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_FAILED
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"proposal_rejected",
				sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.Id)),
				sdk.NewAttribute("status", proposal.Status.String()),
			),
		)
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

// proposalPassed checks if proposal passed based on tally
func (k *Keeper) proposalPassed(tally *types.TallyResult, params *types.GovernanceParams) bool {
	// Check quorum
	// Simplified - production would calculate percentages properly
	yesVotes := tally.Yes
	totalVotes := tally.Yes // + other votes

	return yesVotes >= params.Quorum && totalVotes >= params.Threshold
}

// proposalVetoed checks if proposal was vetoed
func (k *Keeper) proposalVetoed(tally *types.TallyResult, params *types.GovernanceParams) bool {
	// Check if NoWithVeto votes exceed veto threshold
	// Simplified
	return tally.NoWithVeto >= params.VetoThreshold
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
