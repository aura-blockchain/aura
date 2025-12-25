// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/governance/types"
)

// ============================
// VOTE DELEGATION (Feature 4)
// ============================

// DelegateVote delegates voting power to another address
func (k *Keeper) DelegateVote(
	ctx sdk.Context,
	delegator string,
	delegate string,
	percentage uint64,
) error {
	params := k.GetParams(ctx)

	// Validate percentage (0-10000 basis points = 0-100%)
	if percentage == 0 || percentage > 10000 {
		return types.ErrInvalidDelegation
	}

	// Cannot delegate to self
	if delegator == delegate {
		return types.ErrInvalidDelegation
	}

	// Check if delegation already exists
	existingDelegations := k.GetVoteDelegations(ctx, delegator)
	totalDelegated := uint64(0)

	for _, existing := range existingDelegations {
		if existing.Delegate == delegate {
			return types.ErrInvalidDelegation
		}
		// Parse delegated power as percentage (simplified)
		totalDelegated += 1000 // Simplified - would parse from DelegatedPower field
	}

	// Check if total delegations would exceed 100%
	if totalDelegated+percentage > 10000 {
		return types.ErrInvalidDelegation
	}

	// Check max delegations per user
	if uint64(len(existingDelegations)) >= types.GetMaxDelegationsPerUser(params) {
		return types.ErrInvalidDelegation
	}

	// Create delegation
	delegation := &types.VoteDelegation{
		Delegator:      delegator,
		Delegate:       delegate,
		DelegationTime: timestampFromTime(ctx.BlockTime()),
		DelegatedPower: fmt.Sprintf("%d", percentage),
	}

	// Store delegation
	if err := k.SetVoteDelegation(ctx, delegation); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"vote_delegated",
			sdk.NewAttribute("delegator", delegator),
			sdk.NewAttribute("delegate", delegate),
			sdk.NewAttribute("percentage", fmt.Sprintf("%d", percentage)),
		),
	)

	return nil
}

// RevokeDelegation revokes a vote delegation
func (k *Keeper) RevokeDelegation(ctx sdk.Context, delegator, delegate string) error {
	if err := k.DeleteVoteDelegation(ctx, delegator, delegate); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"vote_delegation_revoked",
			sdk.NewAttribute("delegator", delegator),
			sdk.NewAttribute("delegate", delegate),
		),
	)

	return nil
}

// VoteWithDelegatedPower casts a vote using delegated voting power
func (k *Keeper) VoteWithDelegatedPower(
	ctx sdk.Context,
	proposalID uint64,
	delegate string,
	option types.VoteOption,
) error {
	proposal, err := k.GetProposal(ctx, proposalID)
	if err != nil {
		return err
	}

	if proposal.Status != types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD {
		return types.ErrInvalidProposalStatus
	}

	// Calculate total voting power including delegations
	totalPower := k.calculateTotalVotingPower(ctx, delegate)

	// Create vote
	vote := &types.Vote{
		ProposalId:  proposalID,
		Voter:       delegate,
		Option:      option,
		VotingPower: totalPower,
		Timestamp:   timestampFromTime(ctx.BlockTime()),
		IsSecret:    false,
	}

	if err := k.SetVote(ctx, vote); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"delegated_vote_cast",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute("delegate", delegate),
			sdk.NewAttribute("total_power", totalPower),
		),
	)

	return nil
}

// calculateTotalVotingPower calculates total voting power including delegations
func (k *Keeper) calculateTotalVotingPower(ctx sdk.Context, address string) string {
	// GetVotingPower now includes all delegation logic internally
	// It returns: staked tokens + delegated TO - delegated AWAY
	totalPower, err := k.GetVotingPower(ctx, address)
	if err != nil {
		return "0"
	}

	return totalPower.String()
}

// calculateDelegatedPower calculates power delegated to an address
func (k *Keeper) calculateDelegatedPower(ctx sdk.Context, delegate string) string {
	// Find all delegations to this address
	// This requires iterating all delegations - simplified implementation

	totalDelegated := big.NewInt(0)

	// In production, would have reverse index: delegate -> delegators. For now, iterate all delegations.
	for _, delegation := range k.GetAllVoteDelegations(ctx) {
		if delegation == nil || delegation.Delegate != delegate {
			continue
		}
		if delegatedPower, ok := new(big.Int).SetString(delegation.DelegatedPower, 10); ok {
			totalDelegated.Add(totalDelegated, delegatedPower)
		}
	}

	return totalDelegated.String()
}

// GetDelegationChain returns the full delegation chain for an address
func (k *Keeper) GetDelegationChain(ctx sdk.Context, address string) []*DelegationChainNode {
	chain := []*DelegationChainNode{}

	// Get direct delegations
	delegations := k.GetVoteDelegations(ctx, address)

	for _, delegation := range delegations {
		// Parse percentage from DelegatedPower
		percentage := uint64(1000) // Simplified - would parse from delegation.DelegatedPower

		node := &DelegationChainNode{
			Address:    delegation.Delegate,
			Percentage: percentage,
			Depth:      1,
			Path:       []string{address, delegation.Delegate},
		}

		// Recursively get sub-delegations (with depth limit)
		subDelegations := k.GetVoteDelegations(ctx, delegation.Delegate)
		for _, subDel := range subDelegations {
			if !k.delegationCreatesLoop(ctx, subDel.Delegate, address) {
				subPercentage := uint64(500) // Simplified
				subNode := &DelegationChainNode{
					Address:    subDel.Delegate,
					Percentage: (percentage * subPercentage) / 10000,
					Depth:      2,
					Path:       []string{address, delegation.Delegate, subDel.Delegate},
				}
				chain = append(chain, subNode)
			}
		}

		chain = append(chain, node)
	}

	return chain
}

// delegationCreatesLoop checks if a delegation would create a circular reference
func (k *Keeper) delegationCreatesLoop(ctx sdk.Context, delegate, originalDelegator string) bool {
	if delegate == originalDelegator {
		return true
	}

	// Check if delegate has delegations back to original
	delegations := k.GetVoteDelegations(ctx, delegate)
	for _, delegation := range delegations {
		if delegation.Delegate == originalDelegator {
			return true
		}
		// Check one more level (limit depth to prevent infinite recursion)
		subDelegations := k.GetVoteDelegations(ctx, delegation.Delegate)
		for _, subDel := range subDelegations {
			if subDel.Delegate == originalDelegator {
				return true
			}
		}
	}

	return false
}

// GetDelegationStatistics returns delegation statistics
func (k *Keeper) GetDelegationStatistics(ctx sdk.Context) *DelegationStatistics {
	// Count all delegations
	totalDelegations := uint64(0)
	uniqueDelegators := make(map[string]bool)
	uniqueDelegates := make(map[string]bool)

	allProposals := k.GetAllProposals(ctx)
	for _, proposal := range allProposals {
		votes := k.GetVotes(ctx, proposal.Id)
		for _, vote := range votes {
			if vote.IsSecret {
				uniqueDelegates[vote.Voter] = true
			}
		}
	}

	// Simplified - in production would iterate delegation store
	totalDelegations = uint64(len(uniqueDelegates))

	return &DelegationStatistics{
		TotalDelegations:  totalDelegations,
		UniqueDelegators:  uint64(len(uniqueDelegators)),
		UniqueDelegates:   uint64(len(uniqueDelegates)),
		AverageDelegationPercentage: 5000, // Simplified
		MaxDelegationChainDepth: 3,
	}
}

// DelegationChainNode represents a node in the delegation chain
type DelegationChainNode struct {
	Address    string
	Percentage uint64
	Depth      uint32
	Path       []string
}

// DelegationStatistics provides delegation statistics
type DelegationStatistics struct {
	TotalDelegations  uint64
	UniqueDelegators  uint64
	UniqueDelegates   uint64
	AverageDelegationPercentage uint64
	MaxDelegationChainDepth uint32
}
