package keeper

import (
	"context"
	"encoding/binary"
	"fmt"

	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"

	"github.com/aequitas/aura/chain/x/economics/types"
)

// InitGenesis initializes the module state from genesis
func (k Keeper) InitGenesis(ctx context.Context, gs *economicspb.GenesisState) error {
	// Validate genesis state
	if err := types.ValidateGenesisState(gs); err != nil {
		return fmt.Errorf("invalid genesis state: %w", err)
	}

	// Set params
	if err := k.SetParams(ctx, &gs.Params); err != nil {
		return fmt.Errorf("failed to set params: %w", err)
	}

	// Initialize vesting schedules
	seenScheduleIDs := make(map[string]bool)
	for _, schedule := range gs.VestingSchedules {
		scheduleCopy := schedule

		// CRITICAL SECURITY: Detect duplicate schedule IDs during import
		if seenScheduleIDs[scheduleCopy.Id] {
			return fmt.Errorf("duplicate vesting schedule ID in genesis: %s", scheduleCopy.Id)
		}
		seenScheduleIDs[scheduleCopy.Id] = true

		if err := k.SetVestingSchedule(ctx, &scheduleCopy); err != nil {
			return fmt.Errorf("failed to set vesting schedule %s: %w", scheduleCopy.Id, err)
		}
		// Update user index
		if err := k.AddUserVestingSchedule(ctx, scheduleCopy.Address, scheduleCopy.Id); err != nil {
			return fmt.Errorf("failed to add user vesting index for %s: %w", scheduleCopy.Address, err)
		}
	}

	// Initialize vote locks
	seenLockIDs := make(map[string]bool)
	for _, lock := range gs.VoteLocks {
		lockCopy := lock

		// CRITICAL SECURITY: Detect duplicate lock IDs during import
		if seenLockIDs[lockCopy.Id] {
			return fmt.Errorf("duplicate vote lock ID in genesis: %s", lockCopy.Id)
		}
		seenLockIDs[lockCopy.Id] = true

		if err := k.SetVoteLock(ctx, &lockCopy); err != nil {
			return fmt.Errorf("failed to set vote lock %s: %w", lockCopy.Id, err)
		}
		// Update user index
		if err := k.AddUserVoteLock(ctx, lockCopy.Owner, lockCopy.Id); err != nil {
			return fmt.Errorf("failed to add user vote lock index for %s: %w", lockCopy.Owner, err)
		}
	}

	// Initialize pending treasury transactions
	seenTxIDs := make(map[string]bool)
	for _, tx := range gs.PendingTreasuryTxs {
		txCopy := tx

		// CRITICAL SECURITY: Detect duplicate transaction IDs during import
		if seenTxIDs[txCopy.TxId] {
			return fmt.Errorf("duplicate pending treasury tx ID in genesis: %s", txCopy.TxId)
		}
		seenTxIDs[txCopy.TxId] = true

		if err := k.SetPendingTreasuryTx(ctx, &txCopy); err != nil {
			return fmt.Errorf("failed to set pending treasury tx %s: %w", txCopy.TxId, err)
		}
	}

	// Initialize governance state
	if err := k.SetNextProposalID(ctx, gs.NextProposalId); err != nil {
		return fmt.Errorf("failed to set next proposal ID: %w", err)
	}

	seenProposalIDs := make(map[uint64]bool)
	for _, proposal := range gs.Proposals {
		proposalCopy := proposal

		// CRITICAL SECURITY: Detect duplicate proposal IDs during import
		if seenProposalIDs[proposalCopy.Id] {
			return fmt.Errorf("duplicate proposal ID in genesis: %d", proposalCopy.Id)
		}
		seenProposalIDs[proposalCopy.Id] = true

		if err := k.SetProposal(ctx, &proposalCopy); err != nil {
			return fmt.Errorf("failed to set proposal %d: %w", proposalCopy.Id, err)
		}
	}

	for _, vote := range gs.Votes {
		voteCopy := vote
		if err := k.SetVote(ctx, &voteCopy); err != nil {
			return fmt.Errorf("failed to set vote for proposal %d from %s: %w", voteCopy.ProposalId, voteCopy.Voter, err)
		}
	}

	for _, deposit := range gs.Deposits {
		depositCopy := deposit
		if err := k.SetDeposit(ctx, &depositCopy); err != nil {
			return fmt.Errorf("failed to set deposit for proposal %d from %s: %w", depositCopy.ProposalId, depositCopy.Depositor, err)
		}
	}

	for _, delegation := range gs.VoteDelegations {
		delegationCopy := delegation
		if err := k.SetVoteDelegation(ctx, &delegationCopy); err != nil {
			return fmt.Errorf("failed to set vote delegation from %s to %s: %w", delegationCopy.Delegator, delegationCopy.Delegate, err)
		}
	}

	// Initialize inflation metrics if present
	if gs.InflationMetrics != nil {
		if err := k.SetInflationMetrics(ctx, gs.InflationMetrics); err != nil {
			return fmt.Errorf("failed to set inflation metrics: %w", err)
		}
	}

	// Initialize MEV stats if present
	if gs.MevStats != nil {
		if err := k.SetMEVStats(ctx, gs.MevStats); err != nil {
			return fmt.Errorf("failed to set MEV stats: %w", err)
		}
	}

	// Initialize liquidity mining stats if present
	if gs.LiquidityMiningStats != nil {
		if err := k.SetLiquidityMiningStats(ctx, gs.LiquidityMiningStats); err != nil {
			return fmt.Errorf("failed to set liquidity mining stats: %w", err)
		}
	}

	// Initialize user MEV balances
	for addr, balance := range gs.UserMevBalances {
		if err := k.SetUserMEVBalance(ctx, addr, balance); err != nil {
			return fmt.Errorf("failed to set MEV balance for %s: %w", addr, err)
		}
	}

	// Initialize last large tx times
	for addr, timestamp := range gs.LastLargeTxTimes {
		if err := k.SetLastLargeTxTime(ctx, addr, timestamp); err != nil {
			return fmt.Errorf("failed to set last large tx time for %s: %w", addr, err)
		}
	}

	return nil
}

// ExportGenesis exports the module state to genesis
func (k Keeper) ExportGenesis(ctx context.Context) (*economicspb.GenesisState, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}

	gs := &economicspb.GenesisState{
		Params:             *params,
		VestingSchedules:   []economicspb.VestingSchedule{},
		VoteLocks:          []economicspb.VoteLock{},
		PendingTreasuryTxs: []economicspb.PendingTreasuryTx{},
		Proposals:          []economicspb.Proposal{},
		Votes:              []economicspb.Vote{},
		Deposits:           []economicspb.Deposit{},
		VoteDelegations:    []economicspb.VoteDelegation{},
		UserMevBalances:    make(map[string]string),
		LastLargeTxTimes:   make(map[string]int64),
	}

	// Export vesting schedules
	if err := k.IterateVestingSchedules(ctx, func(schedule *economicspb.VestingSchedule) bool {
		gs.VestingSchedules = append(gs.VestingSchedules, *schedule)
		return false
	}); err != nil {
		return nil, fmt.Errorf("failed to export vesting schedules: %w", err)
	}

	// Export vote locks
	if err := k.IterateVoteLocks(ctx, func(lock *economicspb.VoteLock) bool {
		gs.VoteLocks = append(gs.VoteLocks, *lock)
		return false
	}); err != nil {
		return nil, fmt.Errorf("failed to export vote locks: %w", err)
	}

	// Export pending treasury transactions
	if err := k.IteratePendingTreasuryTxs(ctx, func(tx *economicspb.PendingTreasuryTx) bool {
		gs.PendingTreasuryTxs = append(gs.PendingTreasuryTxs, *tx)
		return false
	}); err != nil {
		return nil, fmt.Errorf("failed to export pending treasury txs: %w", err)
	}

	// Export governance state
	nextProposalID, _ := k.GetNextProposalID(ctx)
	gs.NextProposalId = nextProposalID

	if err := k.IterateProposals(ctx, func(proposal *economicspb.Proposal) bool {
		gs.Proposals = append(gs.Proposals, *proposal)
		return false
	}); err != nil {
		return nil, fmt.Errorf("failed to export proposals: %w", err)
	}

	// Export votes for all proposals
	for _, proposal := range gs.Proposals {
		votes, err := k.GetVotes(ctx, proposal.Id)
		if err == nil {
			for _, vote := range votes {
				gs.Votes = append(gs.Votes, *vote)
			}
		}
	}

	// Export deposits for all proposals
	for _, proposal := range gs.Proposals {
		deposits, err := k.GetDeposits(ctx, proposal.Id)
		if err == nil {
			for _, deposit := range deposits {
				gs.Deposits = append(gs.Deposits, *deposit)
			}
		}
	}

	// Export vote delegations (iterate using store prefix)
	store := k.storeService.OpenKVStore(ctx)
	delegationIterator, err := store.Iterator(types.VoteDelegationPrefix, storeprefixend(types.VoteDelegationPrefix))
	if err == nil {
		defer delegationIterator.Close()
		for ; delegationIterator.Valid(); delegationIterator.Next() {
			var delegation economicspb.VoteDelegation
			if err := k.cdc.Unmarshal(delegationIterator.Value(), &delegation); err == nil {
				gs.VoteDelegations = append(gs.VoteDelegations, delegation)
			}
		}
	}

	// Export inflation metrics
	inflationMetrics, err := k.GetInflationMetrics(ctx)
	if err == nil {
		gs.InflationMetrics = inflationMetrics
	}

	// Export MEV stats
	mevStats, err := k.GetMEVStats(ctx)
	if err == nil {
		gs.MevStats = mevStats
	}

	// Export liquidity mining stats
	lmStats, err := k.GetLiquidityMiningStats(ctx)
	if err == nil {
		gs.LiquidityMiningStats = lmStats
	}

	// Export user MEV balances (iterate using store prefix)
	mevIterator, err := store.Iterator(types.UserMEVBalancePrefix, storeprefixend(types.UserMEVBalancePrefix))
	if err == nil {
		defer mevIterator.Close()
		for ; mevIterator.Valid(); mevIterator.Next() {
			addr := string(mevIterator.Key()[len(types.UserMEVBalancePrefix):])
			balance := string(mevIterator.Value())
			gs.UserMevBalances[addr] = balance
		}
	}

	// Export last large tx times (iterate using store prefix)
	txTimeIterator, err := store.Iterator(types.LastLargeTxTimePrefix, storeprefixend(types.LastLargeTxTimePrefix))
	if err == nil {
		defer txTimeIterator.Close()
		for ; txTimeIterator.Valid(); txTimeIterator.Next() {
			addr := string(txTimeIterator.Key()[len(types.LastLargeTxTimePrefix):])
			timestamp := int64(binary.BigEndian.Uint64(txTimeIterator.Value()))
			gs.LastLargeTxTimes[addr] = timestamp
		}
	}

	return gs, nil
}
