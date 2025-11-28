package keeper

import (
	"context"
	"encoding/binary"

	"cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"

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
