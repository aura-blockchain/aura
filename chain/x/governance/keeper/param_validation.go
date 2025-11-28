package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/governance/types"
)

// ============================
// GOVERNANCE PARAMETER VALIDATION (Feature 6)
// ============================

// ValidateGovernanceParams validates governance parameters
func (k *Keeper) ValidateGovernanceParams(params *types.GovernanceParams) error {
	// Use the helper function to validate
	return types.ValidateGovernanceParams(params)
}

// UpdateGovernanceParams updates governance parameters with validation
func (k *Keeper) UpdateGovernanceParams(ctx sdk.Context, newParams *types.GovernanceParams) error {
	// Validate new parameters
	if err := k.ValidateGovernanceParams(newParams); err != nil {
		return err
	}

	// Store old params for event
	oldParams := k.GetParams(ctx)

	// Set new parameters
	k.SetParams(ctx, newParams)

	// Emit parameter change event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"governance_params_updated",
			sdk.NewAttribute("voting_period_old", fmt.Sprintf("%d", types.GetVotingPeriodSeconds(oldParams))),
			sdk.NewAttribute("voting_period_new", fmt.Sprintf("%d", types.GetVotingPeriodSeconds(newParams))),
			sdk.NewAttribute("min_deposit_old", oldParams.MinDeposit),
			sdk.NewAttribute("min_deposit_new", newParams.MinDeposit),
		),
	)

	return nil
}

// ValidateProposalContent validates proposal content
func (k *Keeper) ValidateProposalContent(title, description string, category types.ProposalCategory) error {
	// Title validation
	if len(title) == 0 {
		return fmt.Errorf("proposal title cannot be empty")
	}
	if len(title) > 200 {
		return fmt.Errorf("proposal title too long: %d characters (maximum: 200)", len(title))
	}

	// Description validation
	if len(description) == 0 {
		return fmt.Errorf("proposal description cannot be empty")
	}
	if len(description) > 10000 {
		return fmt.Errorf("proposal description too long: %d characters (maximum: 10000)", len(description))
	}

	// Content validation based on category
	switch category {
	case types.CategoryParameterChange:
		// Parameter change proposals should have structured content

	case types.CategorySoftwareUpgrade:
		// Software upgrade proposals should specify upgrade plan

	case types.CategorySpending:
		// Community spend proposals should specify recipient and amount

	case types.CategoryText:
		// Text proposals are informational only

	default:
		// Accept unknown categories
	}

	return nil
}

// ValidateVote validates a vote
func (k *Keeper) ValidateVote(ctx sdk.Context, proposalID uint64, voter string, option types.VoteOption) error {
	// Check if proposal exists
	proposal, err := k.GetProposal(ctx, proposalID)
	if err != nil {
		return err
	}

	// Check if proposal is in voting period
	if proposal.Status != types.StatusVotingPeriod {
		return types.ErrInvalidProposalStatus
	}

	// Check if vote option is valid
	if !k.isValidVoteOption(option) {
		return types.ErrInvalidVoteOption
	}

	// Check if voter has voting power
	votingPower := k.GetVotingPower(ctx, voter)
	if votingPower == "0" {
		return types.ErrNoVotingPower
	}

	return nil
}

// isValidVoteOption checks if vote option is valid
func (k *Keeper) isValidVoteOption(option types.VoteOption) bool {
	switch option {
	case types.OptionYes, types.OptionNo, types.OptionAbstain, types.OptionNoWithVeto:
		return true
	default:
		return false
	}
}

// ValidateDeposit validates a deposit
func (k *Keeper) ValidateDeposit(ctx sdk.Context, proposalID uint64, depositor string, amount string) error {
	// Check if proposal exists
	proposal, err := k.GetProposal(ctx, proposalID)
	if err != nil {
		return err
	}

	// Check if proposal is in deposit period
	if proposal.Status != types.StatusDepositPeriod {
		return types.ErrInvalidProposalStatus
	}

	// Check if amount is valid
	if amount == "" || amount == "0" {
		return fmt.Errorf("deposit amount must be positive")
	}

	// No max deposit validation for now
	return nil
}

// GetParameterValidationRules returns current parameter validation rules
func (k *Keeper) GetParameterValidationRules() *types.ParameterValidationRules {
	return &types.ParameterValidationRules{
		MinVotingPeriod:        3600,
		MaxVotingPeriod:        2592000,
		MinDepositPeriod:       3600,
		MaxDepositPeriod:       604800,
		MinQuorum:              0,
		MaxQuorum:              10000,
		MinThreshold:           0,
		MaxThreshold:           10000,
		MinVetoThreshold:       0,
		MaxVetoThreshold:       10000,
		MaxExecutionDelay:      604800,
		MaxDelegationsPerUser:  100,
		MinVoteCreditsPerToken: 1,
		MaxProposalTitleLength: 200,
		MaxProposalDescLength:  10000,
	}
}
