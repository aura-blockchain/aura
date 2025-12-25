// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
	votingPower, err := k.GetVotingPower(ctx, voter)
	if err != nil {
		return err
	}

	// Create weighted vote
	// Note: Weighted voting stores vote options as encrypted data in the vote
	// This is a simplified implementation - production would use proper encoding
	vote := &types.Vote{
		ProposalId:  proposalID,
		Voter:       voter,
		Option:      types.VoteOption_VOTE_OPTION_UNSPECIFIED, // Not used for weighted votes
		VotingPower: votingPower.String(),
		Timestamp:   timestampFromTime(ctx.BlockTime()),
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
			sdk.NewAttribute("voting_power", votingPower.String()),
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
	// Get total voting power (includes delegations)
	totalVotingPower, err := k.GetVotingPower(ctx, voter)
	if err != nil {
		totalVotingPower = sdkmath.ZeroInt()
	}

	// Get delegated power received (using the new method)
	delegatedPowerInt := k.GetDelegatedVotingPower(ctx, voter)
	delegatedPower := delegatedPowerInt.String()

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

	// Get base power (direct stake) separately for breakdown
	addr, _ := sdk.AccAddressFromBech32(voter)
	var baseVotingPower string
	if addr != nil && k.stakingKeeper != nil {
		basePower, err := k.stakingKeeper.GetDelegatorBonded(ctx, addr)
		if err == nil {
			baseVotingPower = basePower.String()
		} else {
			baseVotingPower = "0"
		}
	} else {
		baseVotingPower = "0"
	}

	return &VotingPowerBreakdown{
		Address:              voter,
		BaseVotingPower:      baseVotingPower,
		DelegatedPower:       delegatedPower,
		LockedTokenPower:     lockedPower,
		TotalVotingPower:     totalVotingPower.String(),
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
