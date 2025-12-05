package keeper

import (
	"context"
	"encoding/binary"

	"cosmossdk.io/core/store"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/economics/types"
	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
)

// Keeper manages the economics module state using KV store persistence
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService
	authority    string
}

// NewKeeper creates a new Keeper instance with KV store
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:          cdc,
		storeService: storeService,
		authority:    authority,
	}
}

// GetAuthority returns the module authority
func (k Keeper) GetAuthority() string {
	return k.authority
}

// ============================
// PARAMETER OPERATIONS
// ============================

// GetParams returns the current module parameters
func (k Keeper) GetParams(ctx context.Context) (*economicspb.Params, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamsKey)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return types.DefaultParams(), nil
	}

	var params economicspb.Params
	if err := k.cdc.Unmarshal(bz, &params); err != nil {
		return nil, err
	}
	return &params, nil
}

// SetParams sets new module parameters
func (k Keeper) SetParams(ctx context.Context, params *economicspb.Params) error {
	if err := types.ValidateParams(params); err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(params)
	if err != nil {
		return err
	}
	return store.Set(types.ParamsKey, bz)
}

// ============================
// GENESIS OPERATIONS
// ============================

// InitGenesis initializes the module state from genesis
func (k Keeper) InitGenesis(ctx context.Context, gs *economicspb.GenesisState) error {
	// Set params
	if err := k.SetParams(ctx, gs.Params); err != nil {
		return err
	}

	// Initialize vesting schedules
	for _, schedule := range gs.VestingSchedules {
		if err := k.SetVestingSchedule(ctx, schedule); err != nil {
			return err
		}
		// Update user index
		if err := k.AddUserVestingSchedule(ctx, schedule.Address, schedule.Id); err != nil {
			return err
		}
	}

	// Initialize vote locks
	for _, lock := range gs.VoteLocks {
		if err := k.SetVoteLock(ctx, lock); err != nil {
			return err
		}
		// Update user index
		if err := k.AddUserVoteLock(ctx, lock.Owner, lock.Id); err != nil {
			return err
		}
	}

	// Initialize pending treasury transactions
	for _, tx := range gs.PendingTreasuryTxs {
		if err := k.SetPendingTreasuryTx(ctx, tx); err != nil {
			return err
		}
	}

	// Initialize governance state
	if err := k.SetNextProposalID(ctx, gs.NextProposalId); err != nil {
		return err
	}

	for _, proposal := range gs.Proposals {
		if err := k.SetProposal(ctx, proposal); err != nil {
			return err
		}
	}

	for _, vote := range gs.Votes {
		if err := k.SetVote(ctx, vote); err != nil {
			return err
		}
	}

	for _, deposit := range gs.Deposits {
		if err := k.SetDeposit(ctx, deposit); err != nil {
			return err
		}
	}

	for _, delegation := range gs.VoteDelegations {
		if err := k.SetVoteDelegation(ctx, delegation); err != nil {
			return err
		}
	}

	// Initialize inflation metrics if present
	if gs.InflationMetrics != nil {
		// Store inflation metrics
	}

	// Initialize MEV stats if present
	if gs.MevStats != nil {
		// Store MEV stats
	}

	// Initialize liquidity mining stats if present
	if gs.LiquidityMiningStats != nil {
		// Store liquidity mining stats
	}

	// Initialize user MEV balances
	for addr, balance := range gs.UserMevBalances {
		if err := k.SetUserMEVBalance(ctx, addr, balance); err != nil {
			return err
		}
	}

	// Initialize last large tx times
	for addr, timestamp := range gs.LastLargeTxTimes {
		if err := k.SetLastLargeTxTime(ctx, addr, timestamp); err != nil {
			return err
		}
	}

	return nil
}

// ExportGenesis exports the module state to genesis
func (k Keeper) ExportGenesis(ctx context.Context) (*economicspb.GenesisState, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	gs := &economicspb.GenesisState{
		Params:             params,
		VestingSchedules:   []*economicspb.VestingSchedule{},
		VoteLocks:          []*economicspb.VoteLock{},
		PendingTreasuryTxs: []*economicspb.PendingTreasuryTx{},
		Proposals:          []*economicspb.Proposal{},
		Votes:              []*economicspb.Vote{},
		Deposits:           []*economicspb.Deposit{},
		VoteDelegations:    []*economicspb.VoteDelegation{},
	}

	// Export vesting schedules
	if err := k.IterateVestingSchedules(ctx, func(schedule *economicspb.VestingSchedule) bool {
		gs.VestingSchedules = append(gs.VestingSchedules, schedule)
		return false
	}); err != nil {
		return nil, err
	}

	// Export vote locks
	if err := k.IterateVoteLocks(ctx, func(lock *economicspb.VoteLock) bool {
		gs.VoteLocks = append(gs.VoteLocks, lock)
		return false
	}); err != nil {
		return nil, err
	}

	// Export pending treasury transactions
	if err := k.IteratePendingTreasuryTxs(ctx, func(tx *economicspb.PendingTreasuryTx) bool {
		gs.PendingTreasuryTxs = append(gs.PendingTreasuryTxs, tx)
		return false
	}); err != nil {
		return nil, err
	}

	// Export governance state
	nextProposalID, _ := k.GetNextProposalID(ctx)
	gs.NextProposalId = nextProposalID

	if err := k.IterateProposals(ctx, func(proposal *economicspb.Proposal) bool {
		gs.Proposals = append(gs.Proposals, proposal)
		return false
	}); err != nil {
		return nil, err
	}

	return gs, nil
}

// ============================
// CURRENT STATE (Height & Time)
// ============================

// SetCurrentHeight sets the current block height
func (k Keeper) SetCurrentHeight(ctx context.Context, height uint64) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, height)
	return store.Set(types.CurrentHeightKey, bz)
}

// GetCurrentHeight gets the current block height
func (k Keeper) GetCurrentHeight(ctx context.Context) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.CurrentHeightKey)
	if err != nil {
		return 0, err
	}
	if bz == nil {
		return 0, nil
	}
	return binary.BigEndian.Uint64(bz), nil
}

// SetCurrentTime sets the current block time
func (k Keeper) SetCurrentTime(ctx context.Context, t int64) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(t))
	return store.Set(types.CurrentTimeKey, bz)
}

// GetCurrentTime gets the current block time
func (k Keeper) GetCurrentTime(ctx context.Context) (int64, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.CurrentTimeKey)
	if err != nil {
		return 0, err
	}
	if bz == nil {
		return 0, nil
	}
	return int64(binary.BigEndian.Uint64(bz)), nil
}

// ============================
// ITERATOR HELPERS
// ============================

// storeprefixend returns the end key for a given prefix for iteration
func storeprefixend(prefix []byte) []byte {
	end := make([]byte, len(prefix))
	copy(end, prefix)
	for i := len(end) - 1; i >= 0; i-- {
		end[i]++
		if end[i] != 0 {
			return end
		}
	}
	return nil
}

// ============================
// FEE MULTIPLIER OPERATIONS
// ============================

// SetFeeMultiplier stores the current fee multiplier
func (k Keeper) SetFeeMultiplier(ctx context.Context, multiplier string) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(types.FeeMultiplierKey, []byte(multiplier))
}

// GetFeeMultiplier retrieves the current fee multiplier
func (k Keeper) GetFeeMultiplier(ctx context.Context) (string, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.FeeMultiplierKey)
	if err != nil {
		return "1.0", err
	}
	if bz == nil {
		return "1.0", nil
	}
	return string(bz), nil
}

// ============================
// TRANSFER TAX OPERATIONS
// ============================

// SetTransferTaxEnabled stores the transfer tax enabled flag
func (k Keeper) SetTransferTaxEnabled(ctx context.Context, enabled bool) error {
	store := k.storeService.OpenKVStore(ctx)
	var bz []byte
	if enabled {
		bz = []byte{1}
	} else {
		bz = []byte{0}
	}
	return store.Set(append(types.TransferTaxConfigKey, []byte("enabled")...), bz)
}

// GetTransferTaxEnabled retrieves the transfer tax enabled flag
func (k Keeper) GetTransferTaxEnabled(ctx context.Context) (bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(append(types.TransferTaxConfigKey, []byte("enabled")...))
	if err != nil {
		return false, err
	}
	if bz == nil {
		return false, nil
	}
	return bz[0] == 1, nil
}

// SetTransferTaxRate stores the transfer tax rate
func (k Keeper) SetTransferTaxRate(ctx context.Context, rate string) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(append(types.TransferTaxConfigKey, []byte("rate")...), []byte(rate))
}

// GetTransferTaxRate retrieves the transfer tax rate
func (k Keeper) GetTransferTaxRate(ctx context.Context) (string, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(append(types.TransferTaxConfigKey, []byte("rate")...))
	if err != nil {
		return "0", err
	}
	if bz == nil {
		return "0", nil
	}
	return string(bz), nil
}

// setTransferTaxRecipient stores the transfer tax recipient address
func (k Keeper) setTransferTaxRecipient(ctx context.Context, recipient string) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(append(types.TransferTaxConfigKey, []byte("recipient")...), []byte(recipient))
}

// getTransferTaxRecipient retrieves the transfer tax recipient address
func (k Keeper) getTransferTaxRecipient(ctx context.Context) (string, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(append(types.TransferTaxConfigKey, []byte("recipient")...))
	if err != nil {
		return "", err
	}
	if bz == nil {
		return "", nil
	}
	return string(bz), nil
}

// ============================
// MISSING QUERY METHODS
// ============================

// GetVestingSchedulesByAddress retrieves all vesting schedules for an address
func (k Keeper) GetVestingSchedulesByAddress(ctx context.Context, address sdk.AccAddress) ([]*economicspb.VestingSchedule, error) {
	scheduleIDs, err := k.GetUserVestingIndex(ctx, address.String())
	if err != nil {
		return nil, err
	}

	schedules := make([]*economicspb.VestingSchedule, 0, len(scheduleIDs))
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

// GetAllVestingSchedules retrieves all vesting schedules in the system
func (k Keeper) GetAllVestingSchedules(ctx context.Context) ([]*economicspb.VestingSchedule, error) {
	schedules := []*economicspb.VestingSchedule{}
	err := k.IterateVestingSchedules(ctx, func(schedule *economicspb.VestingSchedule) bool {
		schedules = append(schedules, schedule)
		return false
	})
	if err != nil {
		return nil, err
	}
	return schedules, nil
}

// GetAllProposals retrieves all proposals in the system
func (k Keeper) GetAllProposals(ctx context.Context) ([]*economicspb.Proposal, error) {
	proposals := []*economicspb.Proposal{}
	err := k.IterateProposals(ctx, func(proposal *economicspb.Proposal) bool {
		proposals = append(proposals, proposal)
		return false
	})
	if err != nil {
		return nil, err
	}
	return proposals, nil
}

// GetVotes retrieves all votes for a proposal
func (k Keeper) GetVotes(ctx context.Context, proposalID uint64) ([]*economicspb.Vote, error) {
	votes := []*economicspb.Vote{}
	err := k.IterateVotes(ctx, proposalID, func(vote *economicspb.Vote) bool {
		votes = append(votes, vote)
		return false
	})
	if err != nil {
		return nil, err
	}
	return votes, nil
}

// GetDeposits retrieves all deposits for a proposal
func (k Keeper) GetDeposits(ctx context.Context, proposalID uint64) ([]*economicspb.Deposit, error) {
	deposits := []*economicspb.Deposit{}
	err := k.IterateDeposits(ctx, proposalID, func(deposit *economicspb.Deposit) bool {
		deposits = append(deposits, deposit)
		return false
	})
	if err != nil {
		return nil, err
	}
	return deposits, nil
}

// GetTallyResult retrieves or calculates the tally result for a proposal
func (k Keeper) GetTallyResult(ctx context.Context, proposalID uint64) (*economicspb.TallyResult, error) {
	// First check if the proposal has a stored tally result
	proposal, err := k.GetProposal(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	// If proposal has a final tally result, return it
	if proposal.FinalTallyResult != nil {
		return proposal.FinalTallyResult, nil
	}

	// Otherwise, calculate the current tally
	return k.CalculateTally(ctx, proposalID)
}

// SetTallyResult stores a tally result for a proposal
func (k Keeper) SetTallyResult(ctx context.Context, proposalID uint64, tally *economicspb.TallyResult) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetTallyResultKey(proposalID)
	bz, err := k.cdc.Marshal(tally)
	if err != nil {
		return err
	}
	return store.Set(key, bz)
}

// GetVoteLocksByOwner retrieves all vote locks for an owner
func (k Keeper) GetVoteLocksByOwner(ctx context.Context, owner sdk.AccAddress) ([]*economicspb.VoteLock, error) {
	lockIDs, err := k.GetUserVoteLockIndex(ctx, owner.String())
	if err != nil {
		return nil, err
	}

	locks := make([]*economicspb.VoteLock, 0, len(lockIDs))
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

// GetAllVoteLocks retrieves all vote locks in the system
func (k Keeper) GetAllVoteLocks(ctx context.Context) ([]*economicspb.VoteLock, error) {
	locks := []*economicspb.VoteLock{}
	err := k.IterateVoteLocks(ctx, func(lock *economicspb.VoteLock) bool {
		locks = append(locks, lock)
		return false
	})
	if err != nil {
		return nil, err
	}
	return locks, nil
}

// GetVoteDelegations retrieves all vote delegations for a delegator
func (k Keeper) GetVoteDelegations(ctx context.Context, delegator sdk.AccAddress) ([]*economicspb.VoteDelegation, error) {
	delegations := []*economicspb.VoteDelegation{}

	store := k.storeService.OpenKVStore(ctx)
	delegatorPrefix := append(types.VoteDelegationPrefix, []byte(delegator.String())...)

	iterator, err := store.Iterator(delegatorPrefix, storeprefixend(delegatorPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var delegation economicspb.VoteDelegation
		if err := k.cdc.Unmarshal(iterator.Value(), &delegation); err != nil {
			return nil, err
		}
		delegations = append(delegations, &delegation)
	}

	return delegations, nil
}

// GetAllPendingTreasuryTxs retrieves all pending treasury transactions
func (k Keeper) GetAllPendingTreasuryTxs(ctx context.Context) ([]*economicspb.PendingTreasuryTx, error) {
	txs := []*economicspb.PendingTreasuryTx{}
	err := k.IteratePendingTreasuryTxs(ctx, func(tx *economicspb.PendingTreasuryTx) bool {
		txs = append(txs, tx)
		return false
	})
	if err != nil {
		return nil, err
	}
	return txs, nil
}

// ============================
// STATS AND METRICS OPERATIONS
// ============================

// GetInflationMetrics retrieves inflation metrics
func (k Keeper) GetInflationMetrics(ctx context.Context) (*economicspb.InflationMetrics, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.InflationMetricsKey)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		// Return default metrics if not found
		return &economicspb.InflationMetrics{
			CurrentRate:       0,
			CirculatingSupply: "0",
			TotalVested:       "0",
			TotalVesting:      "0",
		}, nil
	}

	var metrics economicspb.InflationMetrics
	if err := k.cdc.Unmarshal(bz, &metrics); err != nil {
		return nil, err
	}
	return &metrics, nil
}

// SetInflationMetrics stores inflation metrics
func (k Keeper) SetInflationMetrics(ctx context.Context, metrics *economicspb.InflationMetrics) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(metrics)
	if err != nil {
		return err
	}
	return store.Set(types.InflationMetricsKey, bz)
}

// GetMEVStats retrieves MEV statistics
func (k Keeper) GetMEVStats(ctx context.Context) (*economicspb.MEVStats, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.MEVStatsKey)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		// Return default stats if not found
		return &economicspb.MEVStats{
			TotalCaptured:         "0",
			TotalRedistributed:    "0",
			PendingRedistribution: "0",
		}, nil
	}

	var stats economicspb.MEVStats
	if err := k.cdc.Unmarshal(bz, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// SetMEVStats stores MEV statistics
func (k Keeper) SetMEVStats(ctx context.Context, stats *economicspb.MEVStats) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(stats)
	if err != nil {
		return err
	}
	return store.Set(types.MEVStatsKey, bz)
}

// GetLiquidityMiningStats retrieves liquidity mining statistics
func (k Keeper) GetLiquidityMiningStats(ctx context.Context) (*economicspb.LiquidityMiningStats, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.LiquidityMiningStatsKey)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		// Return default stats if not found
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

// SetLiquidityMiningStats stores liquidity mining statistics
func (k Keeper) SetLiquidityMiningStats(ctx context.Context, stats *economicspb.LiquidityMiningStats) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(stats)
	if err != nil {
		return err
	}
	return store.Set(types.LiquidityMiningStatsKey, bz)
}

// ============================
// HELPER QUERY METHODS
// ============================

// HasVoted checks if a voter has voted on a proposal
func (k Keeper) HasVoted(ctx context.Context, proposalID uint64, voter sdk.AccAddress) (bool, error) {
	_, err := k.GetVote(ctx, proposalID, voter.String())
	if err != nil {
		if err == types.ErrInvalidVote {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// HasDeposited checks if a depositor has deposited on a proposal
func (k Keeper) HasDeposited(ctx context.Context, proposalID uint64, depositor sdk.AccAddress) (bool, error) {
	_, err := k.GetDeposit(ctx, proposalID, depositor.String())
	if err != nil {
		if err == types.ErrInvalidDeposit {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ============================
// VOTING POWER CALCULATION
// ============================

// CalculateVotingPower calculates total voting power for an address including locks and delegations
func (k Keeper) CalculateVotingPower(ctx context.Context, address sdk.AccAddress, proposalID uint64) (votingPower, lockedAmount, delegatedPower sdkmath.Int, activeLocks uint64, err error) {
	votingPower = sdkmath.ZeroInt()
	lockedAmount = sdkmath.ZeroInt()
	delegatedPower = sdkmath.ZeroInt()
	activeLocks = 0

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentTime := sdkCtx.BlockTime()

	// Get all vote locks for the address
	locks, err := k.GetVoteLocksByOwner(ctx, address)
	if err != nil {
		return votingPower, lockedAmount, delegatedPower, activeLocks, err
	}

	// Sum up voting power from active locks
	for _, lock := range locks {
		// Skip withdrawn locks
		if lock.Withdrawn {
			continue
		}

		// Check if lock is still active
		if lock.LockEnd != nil && currentTime.Before(lock.LockEnd.AsTime()) {
			activeLocks++

			// Add to locked amount
			if lock.Amount != nil {
				lockedAmount = lockedAmount.Add(lock.Amount.Amount)
			}

			// Add to voting power (lock.VotingPower is a string)
			lockPower, ok := sdkmath.NewIntFromString(lock.VotingPower)
			if ok {
				votingPower = votingPower.Add(lockPower)
			}
		}
	}

	// Get delegated power (from others delegating to this address)
	// Iterate through all delegations where this address is the delegate
	// For simplicity, we'll return zero delegated power for now
	// A full implementation would need to track reverse delegation mappings

	return votingPower, lockedAmount, delegatedPower, activeLocks, nil
}
