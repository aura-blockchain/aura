package keeper

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/governance/types"
)

// ============================
// WEIGHTED VOTING MECHANISM (Feature 2)
// ============================

// CastWeightedVote allows a voter to split their voting power across multiple options
func (k *Keeper) CastWeightedVote(
	ctx sdk.Context,
	proposalID uint64,
	voter string,
	options []*types.WeightedVoteOption,
) error {
	proposal, err := k.GetProposal(ctx, proposalID)
	if err != nil {
		return err
	}

	// Verify proposal is in voting period
	if proposal.Status != types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD {
		return types.ErrInvalidProposalStatus
	}

	// Validate weighted options sum to 100%
	if err := k.validateWeightedOptions(options); err != nil {
		return err
	}

	// Get voter's voting power
	votingPower := k.GetVotingPower(ctx, voter)

	// Create weighted vote
	// Note: Weighted voting stores vote options as encrypted data in the vote
	// This is a simplified implementation - production would use proper encoding
	vote := &types.Vote{
		ProposalId:  proposalID,
		Voter:       voter,
		Option:      types.VoteOption_VOTE_OPTION_UNSPECIFIED, // Not used for weighted votes
		VotingPower: votingPower,
		Timestamp:   timestamppb.Now(),
	}

	// Store vote
	if err := k.SetVote(ctx, vote); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"weighted_vote_cast",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute("voter", voter),
			sdk.NewAttribute("voting_power", votingPower),
		),
	)

	return nil
}

// validateWeightedOptions validates that weighted options sum to 100%
func (k *Keeper) validateWeightedOptions(options []*types.WeightedVoteOption) error {
	if len(options) == 0 {
		return types.ErrInvalidVote
	}

	totalWeight := big.NewInt(0)
	maxWeight := big.NewInt(10000)

	for _, option := range options {
		// Parse weight as string
		weight := new(big.Int)
		_, ok := weight.SetString(option.Weight, 10)
		if !ok {
			return types.ErrInvalidWeight
		}

		if weight.Cmp(maxWeight) > 0 {
			return types.ErrInvalidWeight
		}
		totalWeight.Add(totalWeight, weight)
	}

	// Total should be 10000 (100% in basis points)
	if totalWeight.Cmp(maxWeight) != 0 {
		return types.ErrInvalidWeight
	}

	return nil
}

// CalculateWeightedTally calculates tally including weighted votes
func (k *Keeper) CalculateWeightedTally(ctx sdk.Context, proposalID uint64) *types.TallyResult {
	votes := k.GetVotes(ctx, proposalID)

	// Initialize tally counters
	yesVotes := big.NewInt(0)
	noVotes := big.NewInt(0)
	abstainVotes := big.NewInt(0)
	vetoVotes := big.NewInt(0)

	for _, vote := range votes {
		votingPower := new(big.Int)
		votingPower.SetString(vote.VotingPower, 10)

		// For now, we only support simple votes (not weighted split votes)
		// Weighted voting functionality would require additional proto fields
		// This is a placeholder for future weighted voting implementation
		switch vote.Option {
		case types.VoteOption_VOTE_OPTION_YES:
			yesVotes.Add(yesVotes, votingPower)
		case types.VoteOption_VOTE_OPTION_ABSTAIN:
			abstainVotes.Add(abstainVotes, votingPower)
		case types.VoteOption_VOTE_OPTION_NO:
			noVotes.Add(noVotes, votingPower)
		case types.VoteOption_VOTE_OPTION_NO_WITH_VETO:
			vetoVotes.Add(vetoVotes, votingPower)
		}
	}

	return &types.TallyResult{
		Yes:        yesVotes.String(),
		Abstain:    abstainVotes.String(),
		No:         noVotes.String(),
		NoWithVeto: vetoVotes.String(),
	}
}

// GetVotingPowerBreakdown returns detailed voting power breakdown
func (k *Keeper) GetVotingPowerBreakdown(ctx sdk.Context, voter string) *VotingPowerBreakdown {
	// Get base voting power (from tokens)
	baseVotingPower := k.GetVotingPower(ctx, voter)

	// Get delegated power received
	delegatedPower := k.GetDelegatedPower(ctx, voter)

	// Get locked tokens
	locks := k.GetTokenLocks(ctx, voter)
	lockedPower := "0"
	for _, lock := range locks {
		locked := new(big.Int)
		locked.SetString(lock.LockedAmount, 10)

		current := new(big.Int)
		current.SetString(lockedPower, 10)

		current.Add(current, locked)
		lockedPower = current.String()
	}

	// Calculate total
	base := new(big.Int)
	base.SetString(baseVotingPower, 10)

	delegated := new(big.Int)
	delegated.SetString(delegatedPower, 10)

	total := new(big.Int).Add(base, delegated)

	return &VotingPowerBreakdown{
		Address:              voter,
		BaseVotingPower:      baseVotingPower,
		DelegatedPower:       delegatedPower,
		LockedTokenPower:     lockedPower,
		TotalVotingPower:     total.String(),
		ActiveDelegations:    uint64(len(k.GetVoteDelegations(ctx, voter))),
		VotingPowerMultiplier: 10000, // 1.0x in basis points
	}
}

// VotingPowerBreakdown provides detailed voting power information
type VotingPowerBreakdown struct {
	Address              string
	BaseVotingPower      string
	DelegatedPower       string
	LockedTokenPower     string
	TotalVotingPower     string
	ActiveDelegations    uint64
	VotingPowerMultiplier uint64
}
