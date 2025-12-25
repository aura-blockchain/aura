// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/governance/types"
)

// InitGenesis initializes governance state from genesis data.
func (k Keeper) InitGenesis(ctx sdk.Context, gen types.GenesisState) error {
	// Set params
	params := gen.Params
	if params == nil {
		params = types.DefaultParams()
	}
	k.SetParams(ctx, params)

	// Set starting proposal ID
	if gen.StartingProposalId > 0 {
		k.SetNextProposalID(ctx, gen.StartingProposalId)
	}

	// Import proposals
	for _, proposal := range gen.Proposals {
		if proposal == nil {
			ctx.Logger().Warn("skipping nil proposal in genesis")
			continue
		}
		if err := k.SetProposal(ctx, proposal); err != nil {
			return fmt.Errorf("failed to import proposal %d: %w", proposal.Id, err)
		}
		ctx.Logger().Debug("imported proposal", "id", proposal.Id, "title", proposal.Title)
	}

	// Import deposits
	for _, deposit := range gen.Deposits {
		if deposit == nil {
			ctx.Logger().Warn("skipping nil deposit in genesis")
			continue
		}
		if err := k.SetDeposit(ctx, deposit); err != nil {
			return fmt.Errorf("failed to import deposit for proposal %d from %s: %w",
				deposit.ProposalId, deposit.Depositor, err)
		}
		ctx.Logger().Debug("imported deposit", "proposal_id", deposit.ProposalId, "depositor", deposit.Depositor)
	}

	// Import votes
	for _, vote := range gen.Votes {
		if vote == nil {
			ctx.Logger().Warn("skipping nil vote in genesis")
			continue
		}
		if err := k.SetVote(ctx, vote); err != nil {
			return fmt.Errorf("failed to import vote for proposal %d from %s: %w",
				vote.ProposalId, vote.Voter, err)
		}
		ctx.Logger().Debug("imported vote", "proposal_id", vote.ProposalId, "voter", vote.Voter)
	}

	// Import vote delegations
	for _, delegation := range gen.VoteDelegations {
		if delegation == nil {
			ctx.Logger().Warn("skipping nil vote delegation in genesis")
			continue
		}
		if err := k.SetVoteDelegation(ctx, delegation); err != nil {
			return fmt.Errorf("failed to import vote delegation from %s to %s: %w",
				delegation.Delegator, delegation.Delegate, err)
		}
		ctx.Logger().Debug("imported vote delegation", "delegator", delegation.Delegator, "delegate", delegation.Delegate)
	}

	// Import token locks
	for _, lock := range gen.TokenLocks {
		if lock == nil {
			ctx.Logger().Warn("skipping nil token lock in genesis")
			continue
		}
		if err := k.SetTokenLock(ctx, lock); err != nil {
			return fmt.Errorf("failed to import token lock for owner %s proposal %d: %w",
				lock.Owner, lock.ProposalId, err)
		}
		ctx.Logger().Debug("imported token lock", "owner", lock.Owner, "proposal_id", lock.ProposalId)
	}

	// Import veto requests
	for _, veto := range gen.VetoRequests {
		if veto == nil {
			ctx.Logger().Warn("skipping nil veto request in genesis")
			continue
		}
		if err := k.SetVetoRequest(ctx, veto); err != nil {
			return fmt.Errorf("failed to import veto request for proposal %d: %w",
				veto.ProposalId, err)
		}
		ctx.Logger().Debug("imported veto request", "proposal_id", veto.ProposalId, "vetoer", veto.Vetoer)
	}

	ctx.Logger().Info("governance genesis import complete",
		"proposals", len(gen.Proposals),
		"deposits", len(gen.Deposits),
		"votes", len(gen.Votes),
		"delegations", len(gen.VoteDelegations),
		"token_locks", len(gen.TokenLocks),
		"veto_requests", len(gen.VetoRequests),
	)

	return nil
}

// ExportGenesis exports the current governance state as genesis data.
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	params := k.GetParams(ctx)

	// Export all proposals
	proposals := k.GetAllProposals(ctx)
	ctx.Logger().Debug("exporting proposals", "count", len(proposals))

	// Export all deposits
	deposits := k.GetAllDeposits(ctx)
	ctx.Logger().Debug("exporting deposits", "count", len(deposits))

	// Export all votes
	votes := k.GetAllVotes(ctx)
	ctx.Logger().Debug("exporting votes", "count", len(votes))

	// Export all vote delegations
	delegations := k.GetAllVoteDelegations(ctx)
	ctx.Logger().Debug("exporting vote delegations", "count", len(delegations))

	// Export all token locks
	tokenLocks := k.GetAllTokenLocks(ctx)
	ctx.Logger().Debug("exporting token locks", "count", len(tokenLocks))

	// Export all veto requests
	vetoRequests := k.GetAllVetoRequests(ctx)
	ctx.Logger().Debug("exporting veto requests", "count", len(vetoRequests))

	// Get the next proposal ID to preserve ID sequence
	startingProposalID := k.GetNextProposalID(ctx)

	ctx.Logger().Info("governance genesis export complete",
		"proposals", len(proposals),
		"deposits", len(deposits),
		"votes", len(votes),
		"delegations", len(delegations),
		"token_locks", len(tokenLocks),
		"veto_requests", len(vetoRequests),
		"starting_proposal_id", startingProposalID,
	)

	return types.GenesisState{
		Params:             params,
		Proposals:          proposals,
		Deposits:           deposits,
		Votes:              votes,
		VoteDelegations:    delegations,
		TokenLocks:         tokenLocks,
		VetoRequests:       vetoRequests,
		StartingProposalId: startingProposalID,
	}
}
