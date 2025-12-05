package keeper

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/economics/types"
	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
)

// ============================
// QUERY HELPER METHODS
// ============================
// These methods support query operations by providing
// convenient accessors for collections and aggregations.

// GetAllVestingSchedules retrieves all vesting schedules
func (k Keeper) GetAllVestingSchedules(ctx context.Context) ([]*economicspb.VestingSchedule, error) {
	schedules := []*economicspb.VestingSchedule{}
	err := k.IterateVestingSchedules(ctx, func(schedule *economicspb.VestingSchedule) bool {
		schedules = append(schedules, schedule)
		return false
	})
	return schedules, err
}

// GetVestingSchedulesByAddress retrieves all vesting schedules for a specific address
func (k Keeper) GetVestingSchedulesByAddress(ctx context.Context, address sdk.AccAddress) ([]*economicspb.VestingSchedule, error) {
	scheduleIDs, err := k.GetUserVestingIndex(ctx, address.String())
	if err != nil {
		return nil, err
	}

	schedules := []*economicspb.VestingSchedule{}
	for _, scheduleID := range scheduleIDs {
		schedule, err := k.GetVestingSchedule(ctx, scheduleID)
		if err != nil {
			// Skip schedules that no longer exist
			continue
		}
		schedules = append(schedules, schedule)
	}
	return schedules, nil
}

// GetAllVoteLocks retrieves all vote locks
func (k Keeper) GetAllVoteLocks(ctx context.Context) ([]*economicspb.VoteLock, error) {
	locks := []*economicspb.VoteLock{}
	err := k.IterateVoteLocks(ctx, func(lock *economicspb.VoteLock) bool {
		locks = append(locks, lock)
		return false
	})
	return locks, err
}

// GetVoteLocksByOwner retrieves all vote locks for a specific owner
func (k Keeper) GetVoteLocksByOwner(ctx context.Context, owner sdk.AccAddress) ([]*economicspb.VoteLock, error) {
	lockIDs, err := k.GetUserVoteLockIndex(ctx, owner.String())
	if err != nil {
		return nil, err
	}

	locks := []*economicspb.VoteLock{}
	for _, lockID := range lockIDs {
		lock, err := k.GetVoteLock(ctx, lockID)
		if err != nil {
			// Skip locks that no longer exist
			continue
		}
		locks = append(locks, lock)
	}
	return locks, nil
}

// GetAllPendingTreasuryTxs retrieves all pending treasury transactions
func (k Keeper) GetAllPendingTreasuryTxs(ctx context.Context) ([]*economicspb.PendingTreasuryTx, error) {
	txs := []*economicspb.PendingTreasuryTx{}
	err := k.IteratePendingTreasuryTxs(ctx, func(tx *economicspb.PendingTreasuryTx) bool {
		txs = append(txs, tx)
		return false
	})
	return txs, err
}

// GetVotes retrieves all votes for a proposal
func (k Keeper) GetVotes(ctx context.Context, proposalID uint64) ([]*economicspb.Vote, error) {
	votes := []*economicspb.Vote{}
	err := k.IterateVotes(ctx, proposalID, func(vote *economicspb.Vote) bool {
		votes = append(votes, vote)
		return false
	})
	return votes, err
}

// GetDeposits retrieves all deposits for a proposal
func (k Keeper) GetDeposits(ctx context.Context, proposalID uint64) ([]*economicspb.Deposit, error) {
	deposits := []*economicspb.Deposit{}
	err := k.IterateDeposits(ctx, proposalID, func(deposit *economicspb.Deposit) bool {
		deposits = append(deposits, deposit)
		return false
	})
	return deposits, err
}

// GetVoteDelegations retrieves all vote delegations for a delegator
func (k Keeper) GetVoteDelegations(ctx context.Context, delegator sdk.AccAddress) ([]*economicspb.VoteDelegation, error) {
	store := k.storeService.OpenKVStore(ctx)

	// Create prefix for this delegator's delegations
	delegatorPrefix := append(types.VoteDelegationPrefix, []byte(delegator.String())...)
	delegatorPrefix = append(delegatorPrefix, 0) // null separator

	iterator, err := store.Iterator(delegatorPrefix, storeprefixend(delegatorPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	delegations := []*economicspb.VoteDelegation{}
	for ; iterator.Valid(); iterator.Next() {
		var delegation economicspb.VoteDelegation
		if err := k.cdc.Unmarshal(iterator.Value(), &delegation); err != nil {
			return nil, err
		}
		delegations = append(delegations, &delegation)
	}
	return delegations, nil
}

// HasVoted checks if a voter has voted on a proposal
func (k Keeper) HasVoted(ctx context.Context, proposalID uint64, voter sdk.AccAddress) (bool, error) {
	vote, err := k.GetVote(ctx, proposalID, voter.String())
	if err != nil {
		if err == types.ErrInvalidVote {
			return false, nil
		}
		return false, err
	}
	return vote != nil, nil
}

// HasDeposited checks if a depositor has deposited to a proposal
func (k Keeper) HasDeposited(ctx context.Context, proposalID uint64, depositor sdk.AccAddress) (bool, error) {
	deposit, err := k.GetDeposit(ctx, proposalID, depositor.String())
	if err != nil {
		if err == types.ErrInvalidDeposit {
			return false, nil
		}
		return false, err
	}
	return deposit != nil, nil
}

// ============================
// TALLY RESULT OPERATIONS
// ============================

// SetTallyResult stores the tally result for a proposal
func (k Keeper) SetTallyResult(ctx context.Context, proposalID uint64, tally *economicspb.TallyResult) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetTallyResultKey(proposalID)
	bz, err := k.cdc.Marshal(tally)
	if err != nil {
		return err
	}
	return store.Set(key, bz)
}

// GetTallyResult retrieves the tally result for a proposal
func (k Keeper) GetTallyResult(ctx context.Context, proposalID uint64) (*economicspb.TallyResult, error) {
	// First check if there's a stored tally result
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetTallyResultKey(proposalID)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}

	if bz != nil {
		var tally economicspb.TallyResult
		if err := k.cdc.Unmarshal(bz, &tally); err != nil {
			return nil, err
		}
		return &tally, nil
	}

	// If no stored tally, calculate it from votes
	proposal, err := k.GetProposal(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	tally := &economicspb.TallyResult{
		YesCount:        "0",
		NoCount:         "0",
		AbstainCount:    "0",
		NoWithVetoCount: "0",
	}

	// If proposal is still in deposit period, return zero tally
	if proposal.Status == economicspb.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD {
		return tally, nil
	}

	// Calculate tally from all votes - using string math
	yesCount := math.ZeroInt()
	noCount := math.ZeroInt()
	abstainCount := math.ZeroInt()
	vetoCount := math.ZeroInt()

	err = k.IterateVotes(ctx, proposalID, func(vote *economicspb.Vote) bool {
		// Parse voting power from string
		votePower, ok := math.NewIntFromString(vote.VotingPower)
		if !ok {
			return false // Skip invalid votes
		}

		switch vote.Option {
		case economicspb.VoteOption_VOTE_OPTION_YES:
			yesCount = yesCount.Add(votePower)
		case economicspb.VoteOption_VOTE_OPTION_NO:
			noCount = noCount.Add(votePower)
		case economicspb.VoteOption_VOTE_OPTION_ABSTAIN:
			abstainCount = abstainCount.Add(votePower)
		case economicspb.VoteOption_VOTE_OPTION_NO_WITH_VETO:
			vetoCount = vetoCount.Add(votePower)
		}
		return false
	})
	if err != nil {
		return nil, err
	}

	// Convert to strings for proto
	tally.YesCount = yesCount.String()
	tally.NoCount = noCount.String()
	tally.AbstainCount = abstainCount.String()
	tally.NoWithVetoCount = vetoCount.String()

	return tally, nil
}

// ============================
// INFLATION METRICS OPERATIONS
// ============================

// SetInflationMetrics stores inflation metrics
func (k Keeper) SetInflationMetrics(ctx context.Context, metrics *economicspb.InflationMetrics) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(metrics)
	if err != nil {
		return err
	}
	return store.Set(types.InflationMetricsKey, bz)
}

// GetInflationMetrics retrieves inflation metrics
func (k Keeper) GetInflationMetrics(ctx context.Context) (*economicspb.InflationMetrics, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.InflationMetricsKey)
	if err != nil {
		return nil, err
	}

	// Return default metrics if none exist
	if bz == nil {
		return &economicspb.InflationMetrics{
			CurrentRate: 0,
			// TODO: Add other fields when proto is finalized
		}, nil
	}

	var metrics economicspb.InflationMetrics
	if err := k.cdc.Unmarshal(bz, &metrics); err != nil {
		return nil, err
	}
	return &metrics, nil
}

// ============================
// MEV STATS OPERATIONS
// ============================

// SetMEVStats stores MEV statistics
func (k Keeper) SetMEVStats(ctx context.Context, stats *economicspb.MEVStats) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(stats)
	if err != nil {
		return err
	}
	return store.Set(types.MEVStatsKey, bz)
}

// GetMEVStats retrieves MEV statistics
func (k Keeper) GetMEVStats(ctx context.Context) (*economicspb.MEVStats, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.MEVStatsKey)
	if err != nil {
		return nil, err
	}

	// Return default stats if none exist
	if bz == nil {
		return &economicspb.MEVStats{
			TotalCaptured:      "0",
			TotalRedistributed: "0",
			// TODO: Add other fields when proto is finalized
		}, nil
	}

	var stats economicspb.MEVStats
	if err := k.cdc.Unmarshal(bz, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// GetUserMEVBalanceDetailed retrieves the MEV balance for a user with details
// Returns balance and lifetime received
func (k Keeper) GetUserMEVBalanceDetailed(ctx context.Context, address sdk.AccAddress) (math.Int, math.Int, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetUserMEVBalanceKey(address.String())
	bz, err := store.Get(key)
	if err != nil {
		return math.ZeroInt(), math.ZeroInt(), err
	}

	// Return zero if no balance exists
	if bz == nil {
		return math.ZeroInt(), math.ZeroInt(), nil
	}

	// Note: UserMEVBalance type would need to be defined in economicspb
	// For now, return zeros as placeholder
	return math.ZeroInt(), math.ZeroInt(), nil
}

// SetUserMEVBalanceObject stores the MEV balance object for a user
func (k Keeper) SetUserMEVBalanceObject(ctx context.Context, address string, balanceAmount string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetUserMEVBalanceKey(address)
	// Store the balance amount as string bytes
	return store.Set(key, []byte(balanceAmount))
}

// ============================
// LIQUIDITY MINING STATS OPERATIONS
// ============================

// SetLiquidityMiningStats stores liquidity mining statistics
func (k Keeper) SetLiquidityMiningStats(ctx context.Context, stats *economicspb.LiquidityMiningStats) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(stats)
	if err != nil {
		return err
	}
	return store.Set(types.LiquidityMiningStatsKey, bz)
}

// GetLiquidityMiningStats retrieves liquidity mining statistics
func (k Keeper) GetLiquidityMiningStats(ctx context.Context) (*economicspb.LiquidityMiningStats, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.LiquidityMiningStatsKey)
	if err != nil {
		return nil, err
	}

	// Return default stats if none exist
	if bz == nil {
		return &economicspb.LiquidityMiningStats{
			CurrentEpoch:           0,
			TotalDistributed:       "0",
			RemainingRewards:       "0",
			RewardsThisEpoch:       "0",
			NextDistributionHeight: 0,
		}, nil
	}

	var stats economicspb.LiquidityMiningStats
	if err := k.cdc.Unmarshal(bz, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// ============================
// VOTING POWER CALCULATION
// ============================

// CalculateVotingPower calculates the total voting power for an address
// Returns: votingPower, lockedAmount, delegatedPower, activeLocks
func (k Keeper) CalculateVotingPower(ctx context.Context, address sdk.AccAddress, proposalID uint64) (math.Int, math.Int, math.Int, uint64, error) {
	votingPower := math.ZeroInt()
	lockedAmount := math.ZeroInt()
	delegatedPower := math.ZeroInt()
	activeLocks := uint64(0)

	// Get all vote locks for the address
	locks, err := k.GetVoteLocksByOwner(ctx, address)
	if err != nil {
		return math.ZeroInt(), math.ZeroInt(), math.ZeroInt(), 0, err
	}

	// Sum up voting power from locks
	for _, lock := range locks {
		activeLocks++
		lockedAmount = lockedAmount.Add(lock.Amount.Amount)

		// Parse voting power from string (customtype field)
		lockVotingPower, ok := math.NewIntFromString(lock.VotingPower)
		if !ok {
			continue // Skip invalid voting power
		}
		votingPower = votingPower.Add(lockVotingPower)
	}

	// Get delegated voting power
	delegations, err := k.GetVoteDelegations(ctx, address)
	if err != nil {
		return votingPower, lockedAmount, delegatedPower, activeLocks, nil
	}

	for _, delegation := range delegations {
		// Parse delegated power from string (customtype field)
		delPower, ok := math.NewIntFromString(delegation.DelegatedPower)
		if !ok {
			continue // Skip invalid delegated power
		}
		delegatedPower = delegatedPower.Add(delPower)
	}

	// For snapshot voting, we would retrieve the snapshot at the proposal submission height
	// For now, use current voting power
	if proposalID > 0 {
		// Snapshot voting logic would go here
		// For now, return current voting power
	}

	return votingPower, lockedAmount, delegatedPower, activeLocks, nil
}
