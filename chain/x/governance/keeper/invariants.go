package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/governance/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers all governance module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "proposal-validity", ProposalValidityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "vote-consistency", VoteConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "deposit-consistency", DepositConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "voting-power-consistency", VotingPowerConsistencyInvariant(k))
}

// AllInvariants runs all invariants of the governance module
func AllInvariants(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		invariants := []sdk.Invariant{
			ParamsInvariant(k),
			ProposalValidityInvariant(k),
			VoteConsistencyInvariant(k),
			DepositConsistencyInvariant(k),
			VotingPowerConsistencyInvariant(k),
		}

		for _, inv := range invariants {
			msg, broken := inv(ctx)
			if broken {
				return msg, broken
			}
		}

		return "", false
	}
}

// ParamsInvariant checks that module parameters are valid
func ParamsInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Basic validation of params
		params := k.GetParams(ctx)
		if params == nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				"params are nil",
			), true
		}

		// Basic field validation
		if params.MinDeposit == "" {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				"min_deposit is empty",
			), true
		}

		return "", false
	}
}

// ProposalValidityInvariant checks that all proposals have valid state
func ProposalValidityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		iterator := storetypes.KVStorePrefixIterator(store, ProposalsKeyPrefix)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var proposal types.Proposal
			if err := k.cdc.Unmarshal(iterator.Value(), &proposal); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"proposal-validity",
					fmt.Sprintf("failed to unmarshal proposal: %s", err.Error()),
				), true
			}

			// Check proposal ID is positive
			if proposal.Id == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"proposal-validity",
					"proposal has zero ID",
				), true
			}

			// Check proposer is valid address
			if _, err := sdk.AccAddressFromBech32(proposal.Proposer); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"proposal-validity",
					fmt.Sprintf("proposal %d has invalid proposer: %s", proposal.Id, proposal.Proposer),
				), true
			}

			// Check title and description are not empty
			if proposal.Title == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"proposal-validity",
					fmt.Sprintf("proposal %d has empty title", proposal.Id),
				), true
			}

			if proposal.Description == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"proposal-validity",
					fmt.Sprintf("proposal %d has empty description", proposal.Id),
				), true
			}

			// Check status is valid
			validStatuses := []types.ProposalStatus{
				types.StatusDepositPeriod,
				types.StatusVotingPeriod,
				types.StatusPassed,
				types.StatusRejected,
				types.StatusFailed,
				types.StatusVetoed,
				types.StatusExecutionDelay,
				types.StatusReadyForExecution,
				types.StatusExecuted,
			}
			statusValid := false
			for _, vs := range validStatuses {
				if proposal.Status == vs {
					statusValid = true
					break
				}
			}
			if !statusValid {
				return sdk.FormatInvariant(
					types.ModuleName,
					"proposal-validity",
					fmt.Sprintf("proposal %d has invalid status: %s", proposal.Id, proposal.Status),
				), true
			}

			// Check timestamps
			if proposal.SubmitTime == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"proposal-validity",
					fmt.Sprintf("proposal %d has nil submit_time", proposal.Id),
				), true
			}

			// Active proposals should have voting period
			if proposal.Status == types.StatusVotingPeriod {
				if proposal.VotingStartTime == nil {
					return sdk.FormatInvariant(
						types.ModuleName,
						"proposal-validity",
						fmt.Sprintf("active proposal %d has nil voting_start_time", proposal.Id),
					), true
				}
				if proposal.VotingEndTime == nil {
					return sdk.FormatInvariant(
						types.ModuleName,
						"proposal-validity",
						fmt.Sprintf("active proposal %d has nil voting_end_time", proposal.Id),
					), true
				}
			}
		}

		return "", false
	}
}

// VoteConsistencyInvariant checks vote consistency
func VoteConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		iterator := storetypes.KVStorePrefixIterator(store, VotesKeyPrefix)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var vote types.Vote
			if err := k.cdc.Unmarshal(iterator.Value(), &vote); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"vote-consistency",
					fmt.Sprintf("failed to unmarshal vote: %s", err.Error()),
				), true
			}

			// Check proposal ID is positive
			if vote.ProposalId == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"vote-consistency",
					"vote has zero proposal ID",
				), true
			}

			// Check voter is valid address
			if _, err := sdk.AccAddressFromBech32(vote.Voter); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"vote-consistency",
					fmt.Sprintf("vote for proposal %d has invalid voter: %s", vote.ProposalId, vote.Voter),
				), true
			}

			// Check vote option is valid
			validOptions := []types.VoteOption{
				types.OptionYes,
				types.OptionNo,
				types.OptionAbstain,
				types.OptionNoWithVeto,
			}
			optionValid := false
			for _, vo := range validOptions {
				if vote.Option == vo {
					optionValid = true
					break
				}
			}
			if !optionValid {
				return sdk.FormatInvariant(
					types.ModuleName,
					"vote-consistency",
					fmt.Sprintf("vote has invalid option: %s", vote.Option),
				), true
			}

			// Check voting power is positive
			power, ok := sdkmath.NewIntFromString(vote.VotingPower)
			if !ok || !power.IsPositive() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"vote-consistency",
					fmt.Sprintf("vote has invalid voting power: %s", vote.VotingPower),
				), true
			}

			// Check timestamp
			if vote.Timestamp == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"vote-consistency",
					"vote has nil timestamp",
				), true
			}
		}

		return "", false
	}
}

// DepositConsistencyInvariant checks deposit consistency
func DepositConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		iterator := storetypes.KVStorePrefixIterator(store, DepositsKeyPrefix)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var deposit types.Deposit
			if err := k.cdc.Unmarshal(iterator.Value(), &deposit); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"deposit-consistency",
					fmt.Sprintf("failed to unmarshal deposit: %s", err.Error()),
				), true
			}

			// Check proposal ID is positive
			if deposit.ProposalId == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"deposit-consistency",
					"deposit has zero proposal ID",
				), true
			}

			// Check depositor is valid address
			if _, err := sdk.AccAddressFromBech32(deposit.Depositor); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"deposit-consistency",
					fmt.Sprintf("deposit for proposal %d has invalid depositor: %s",
						deposit.ProposalId, deposit.Depositor),
				), true
			}

			// Check amount is positive
			amount, ok := sdkmath.NewIntFromString(deposit.Amount)
			if !ok || !amount.IsPositive() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"deposit-consistency",
					fmt.Sprintf("deposit has invalid amount: %s", deposit.Amount),
				), true
			}

			// Check timestamp
			if deposit.Timestamp == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"deposit-consistency",
					"deposit has nil timestamp",
				), true
			}
		}

		return "", false
	}
}

// VotingPowerConsistencyInvariant checks voting power totals match votes
func VotingPowerConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)

		// Build map of proposal -> total voting power
		proposalPower := make(map[uint64]sdkmath.Int)
		voteIter := storetypes.KVStorePrefixIterator(store, VotesKeyPrefix)
		for ; voteIter.Valid(); voteIter.Next() {
			var vote types.Vote
			if err := k.cdc.Unmarshal(voteIter.Value(), &vote); err == nil {
				power, ok := sdkmath.NewIntFromString(vote.VotingPower)
				if ok {
					if proposalPower[vote.ProposalId].IsNil() {
						proposalPower[vote.ProposalId] = sdkmath.ZeroInt()
					}
					proposalPower[vote.ProposalId] = proposalPower[vote.ProposalId].Add(power)
				}
			}
		}
		voteIter.Close()

		// Check proposals total voting power matches
		proposalIter := storetypes.KVStorePrefixIterator(store, ProposalsKeyPrefix)
		defer proposalIter.Close()

		for ; proposalIter.Valid(); proposalIter.Next() {
			var proposal types.Proposal
			if err := k.cdc.Unmarshal(proposalIter.Value(), &proposal); err != nil {
				continue
			}

			// Only check active/completed proposals
			if proposal.Status != types.StatusVotingPeriod &&
				proposal.Status != types.StatusPassed &&
				proposal.Status != types.StatusRejected {
				continue
			}

			// Calculate total from tally result
			if proposal.FinalTallyResult == nil {
				continue
			}

			yes, ok1 := sdkmath.NewIntFromString(proposal.FinalTallyResult.Yes)
			no, ok2 := sdkmath.NewIntFromString(proposal.FinalTallyResult.No)
			abstain, ok3 := sdkmath.NewIntFromString(proposal.FinalTallyResult.Abstain)
			veto, ok4 := sdkmath.NewIntFromString(proposal.FinalTallyResult.NoWithVeto)

			if !ok1 || !ok2 || !ok3 || !ok4 {
				continue
			}

			totalPower := yes.Add(no).Add(abstain).Add(veto)
			calculatedPower := proposalPower[proposal.Id]
			if calculatedPower.IsNil() {
				calculatedPower = sdkmath.ZeroInt()
			}

			// Allow small discrepancy for rounding
			if !totalPower.Equal(calculatedPower) {
				return sdk.FormatInvariant(
					types.ModuleName,
					"voting-power-consistency",
					fmt.Sprintf("proposal %d voting power mismatch: stored=%s, calculated=%s",
						proposal.Id, totalPower.String(), calculatedPower.String()),
				), true
			}
		}

		return "", false
	}
}
