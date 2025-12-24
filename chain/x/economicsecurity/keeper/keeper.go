package keeper

import (
	"context"
	"encoding/binary"
	"fmt"

	"cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/aequitas/aura/chain/x/economicsecurity/params"
	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// Keeper manages the economic security module state using KV store persistence
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService
	paramsStore  *params.Store
	authority    string
}

// NewKeeper creates a new Keeper instance with KV store
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	paramsStore *params.Store,
	authority string,
) *Keeper {
	if paramsStore == nil {
		paramsStore = params.NewStore(*types.DefaultParams())
	}

	return &Keeper{
		cdc:          cdc,
		storeService: storeService,
		paramsStore:  paramsStore,
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
func (k Keeper) GetParams(ctx context.Context) (types.Params, error) {
	return k.paramsStore.GetParams(), nil
}

// SetParams sets new module parameters
func (k Keeper) SetParams(params types.Params) error {
	return k.paramsStore.SetParams(params)
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
// VESTING SCHEDULE OPERATIONS
// ============================

// SetVestingSchedule stores a vesting schedule
func (k Keeper) SetVestingSchedule(ctx context.Context, schedule *types.VestingSchedule) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetVestingScheduleKey(schedule.ScheduleId)
	bz, err := k.cdc.Marshal(schedule)
	if err != nil {
		return fmt.Errorf("failed to marshal for ScheduleId: %w", err)
	}
	return store.Set(key, bz)
}

// GetVestingSchedule retrieves a vesting schedule
func (k Keeper) GetVestingSchedule(ctx context.Context, scheduleID string) (*types.VestingSchedule, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetVestingScheduleKey(scheduleID)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrVestingScheduleNotFound
	}

	var schedule types.VestingSchedule
	if err := k.cdc.Unmarshal(bz, &schedule); err != nil {
		return nil, err
	}
	return &schedule, nil
}

// DeleteVestingSchedule removes a vesting schedule
func (k Keeper) DeleteVestingSchedule(ctx context.Context, scheduleID string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetVestingScheduleKey(scheduleID)
	return store.Delete(key)
}

// ============================
// USER VESTING INDEX OPERATIONS
// ============================

// SetUserVestingIndex stores the list of vesting schedule IDs for a user
func (k Keeper) SetUserVestingIndex(ctx context.Context, userAddress string, scheduleIDs []string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetUserVestingIndexKey(userAddress)

	// Convert to StringList for protobuf marshaling
	index := &types.StringList{Values: scheduleIDs}
	bz, err := k.cdc.Marshal(index)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	return store.Set(key, bz)
}

// GetUserVestingIndex retrieves the list of vesting schedule IDs for a user
func (k Keeper) GetUserVestingIndex(ctx context.Context, userAddress string) ([]string, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetUserVestingIndexKey(userAddress)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return []string{}, nil
	}

	var index types.StringList
	if err := k.cdc.Unmarshal(bz, &index); err != nil {
		return nil, err
	}
	return index.Values, nil
}

// AddUserVestingSchedule adds a schedule ID to a user's vesting index
func (k Keeper) AddUserVestingSchedule(ctx context.Context, userAddress, scheduleID string) error {
	scheduleIDs, err := k.GetUserVestingIndex(ctx, userAddress)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}
	scheduleIDs = append(scheduleIDs, scheduleID)
	return k.SetUserVestingIndex(ctx, userAddress, scheduleIDs)
}

// ============================
// VOTE LOCK OPERATIONS
// ============================

// SetVoteLock stores a vote lock
func (k Keeper) SetVoteLock(ctx context.Context, lock *types.VoteLock) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetVoteLockKey(lock.LockId)
	bz, err := k.cdc.Marshal(lock)
	if err != nil {
		return fmt.Errorf("failed to marshal for LockId: %w", err)
	}
	return store.Set(key, bz)
}

// GetVoteLock retrieves a vote lock
func (k Keeper) GetVoteLock(ctx context.Context, lockID string) (*types.VoteLock, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetVoteLockKey(lockID)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrVoteLockNotFound
	}

	var lock types.VoteLock
	if err := k.cdc.Unmarshal(bz, &lock); err != nil {
		return nil, err
	}
	return &lock, nil
}

// DeleteVoteLock removes a vote lock
func (k Keeper) DeleteVoteLock(ctx context.Context, lockID string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetVoteLockKey(lockID)
	return store.Delete(key)
}

// ============================
// USER VOTE LOCK INDEX OPERATIONS
// ============================

// SetUserVoteLockIndex stores the list of vote lock IDs for a user
func (k Keeper) SetUserVoteLockIndex(ctx context.Context, userAddress string, lockIDs []string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetUserVoteLockIndexKey(userAddress)

	index := &types.StringList{Values: lockIDs}
	bz, err := k.cdc.Marshal(index)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	return store.Set(key, bz)
}

// GetUserVoteLockIndex retrieves the list of vote lock IDs for a user
func (k Keeper) GetUserVoteLockIndex(ctx context.Context, userAddress string) ([]string, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetUserVoteLockIndexKey(userAddress)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return []string{}, nil
	}

	var index types.StringList
	if err := k.cdc.Unmarshal(bz, &index); err != nil {
		return nil, err
	}
	return index.Values, nil
}

// AddUserVoteLock adds a lock ID to a user's vote lock index
func (k Keeper) AddUserVoteLock(ctx context.Context, userAddress, lockID string) error {
	lockIDs, err := k.GetUserVoteLockIndex(ctx, userAddress)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}
	lockIDs = append(lockIDs, lockID)
	return k.SetUserVoteLockIndex(ctx, userAddress, lockIDs)
}

// ============================
// PENDING TREASURY TX OPERATIONS
// ============================

// SetPendingTreasuryTx stores a pending treasury transaction
func (k Keeper) SetPendingTreasuryTx(ctx context.Context, tx *types.PendingTreasuryTx) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetPendingTreasuryTxKey(tx.TxId)
	bz, err := k.cdc.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal for TxId: %w", err)
	}
	return store.Set(key, bz)
}

// GetPendingTreasuryTx retrieves a pending treasury transaction
func (k Keeper) GetPendingTreasuryTx(ctx context.Context, txID string) (*types.PendingTreasuryTx, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetPendingTreasuryTxKey(txID)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrTxNotFound
	}

	var tx types.PendingTreasuryTx
	if err := k.cdc.Unmarshal(bz, &tx); err != nil {
		return nil, err
	}
	return &tx, nil
}

// DeletePendingTreasuryTx removes a pending treasury transaction
func (k Keeper) DeletePendingTreasuryTx(ctx context.Context, txID string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetPendingTreasuryTxKey(txID)
	return store.Delete(key)
}

// IteratePendingTreasuryTxs iterates over all pending treasury transactions
// The callback should return true to stop iteration, false to continue
func (k Keeper) IteratePendingTreasuryTxs(ctx context.Context, cb func(tx *types.PendingTreasuryTx) bool) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.PendingTreasuryTxPrefix, storeprefixend(types.PendingTreasuryTxPrefix))
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var tx types.PendingTreasuryTx
		if err := k.cdc.Unmarshal(iterator.Value(), &tx); err != nil {
			return fmt.Errorf("failed to create iterator for Valid: %w", err)
		}
		if cb(&tx) {
			break // Callback returned true = stop iteration
		}
	}
	return nil
}

// ============================
// INFLATION ALERT OPERATIONS
// ============================

// SetInflationAlert stores an inflation alert
func (k Keeper) SetInflationAlert(ctx context.Context, alert *types.InflationAlert) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetInflationAlertKey(alert.AlertId)
	bz, err := k.cdc.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal for AlertId: %w", err)
	}
	return store.Set(key, bz)
}

// GetInflationAlert retrieves an inflation alert
func (k Keeper) GetInflationAlert(ctx context.Context, alertID string) (*types.InflationAlert, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetInflationAlertKey(alertID)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, fmt.Errorf("inflation alert not found: %s", alertID)
	}

	var alert types.InflationAlert
	if err := k.cdc.Unmarshal(bz, &alert); err != nil {
		return nil, err
	}
	return &alert, nil
}

// IterateInflationAlerts iterates over all inflation alerts
// The callback should return true to stop iteration, false to continue
func (k Keeper) IterateInflationAlerts(ctx context.Context, cb func(alert *types.InflationAlert) bool) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.InflationAlertPrefix, storeprefixend(types.InflationAlertPrefix))
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var alert types.InflationAlert
		if err := k.cdc.Unmarshal(iterator.Value(), &alert); err != nil {
			return fmt.Errorf("failed to create iterator for Valid: %w", err)
		}
		if cb(&alert) {
			break // Callback returned true = stop iteration
		}
	}
	return nil
}

// ============================
// LARGE TX RECORD OPERATIONS
// ============================

// SetLargeTxRecord stores a large transaction record
func (k Keeper) SetLargeTxRecord(ctx context.Context, record *types.LargeTxRecord) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetLargeTxRecordKey(record.TxHash)
	bz, err := k.cdc.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	return store.Set(key, bz)
}

// GetLargeTxRecord retrieves a large transaction record
func (k Keeper) GetLargeTxRecord(ctx context.Context, txHash string) (*types.LargeTxRecord, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetLargeTxRecordKey(txHash)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, fmt.Errorf("large tx record not found: %s", txHash)
	}

	var record types.LargeTxRecord
	if err := k.cdc.Unmarshal(bz, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

// IterateLargeTxRecords iterates over all large tx records
// The callback should return true to stop iteration, false to continue
func (k Keeper) IterateLargeTxRecords(ctx context.Context, cb func(record *types.LargeTxRecord) bool) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.LargeTxRecordPrefix, storeprefixend(types.LargeTxRecordPrefix))
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var record types.LargeTxRecord
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			return fmt.Errorf("failed to create iterator for Valid: %w", err)
		}
		if cb(&record) {
			break // Callback returned true = stop iteration
		}
	}
	return nil
}

// ============================
// LAST LARGE TX TIME OPERATIONS
// ============================

// SetLastLargeTxTime stores the last large transaction time for an address
func (k Keeper) SetLastLargeTxTime(ctx context.Context, address string, timestamp int64) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetLastLargeTxTimeKey(address)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(timestamp))
	return store.Set(key, bz)
}

// GetLastLargeTxTime retrieves the last large transaction time for an address
func (k Keeper) GetLastLargeTxTime(ctx context.Context, address string) (int64, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetLastLargeTxTimeKey(address)
	bz, err := store.Get(key)
	if err != nil {
		return 0, err
	}
	if bz == nil {
		return 0, nil
	}
	return int64(binary.BigEndian.Uint64(bz)), nil
}

// ============================
// ADDRESS HOLDING OPERATIONS
// ============================

// SetAddressHolding stores the holding amount for an address
func (k Keeper) SetAddressHolding(ctx context.Context, address string, amount string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetAddressHoldingKey(address)
	return store.Set(key, []byte(amount))
}

// GetAddressHolding retrieves the holding amount for an address
func (k Keeper) GetAddressHolding(ctx context.Context, address string) (string, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetAddressHoldingKey(address)
	bz, err := store.Get(key)
	if err != nil {
		return "0", err
	}
	if bz == nil {
		return "0", nil
	}
	return string(bz), nil
}

// ============================
// USER MEV BALANCE OPERATIONS
// ============================

// SetUserMEVBalance stores the MEV balance for a user
func (k Keeper) SetUserMEVBalance(ctx context.Context, address string, balance string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetUserMEVBalanceKey(address)
	return store.Set(key, []byte(balance))
}

// GetUserMEVBalance retrieves the MEV balance for a user
func (k Keeper) GetUserMEVBalance(ctx context.Context, address string) (string, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetUserMEVBalanceKey(address)
	bz, err := store.Get(key)
	if err != nil {
		return "0", err
	}
	if bz == nil {
		return "0", nil
	}
	return string(bz), nil
}

// IterateUserMEVBalances iterates over all user MEV balances
// The callback should return true to stop iteration, false to continue
func (k Keeper) IterateUserMEVBalances(ctx context.Context, cb func(address string, balance string) bool) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.UserMEVBalancePrefix, storeprefixend(types.UserMEVBalancePrefix))
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// Extract address from key
		key := iterator.Key()
		address := string(key[len(types.UserMEVBalancePrefix):])
		balance := string(iterator.Value())

		if cb(address, balance) {
			break // Callback returned true = stop iteration
		}
	}
	return nil
}

// ============================
// TOTAL MEV PENDING OPERATIONS
// ============================

// SetTotalMEVPending stores the total pending MEV amount
func (k Keeper) SetTotalMEVPending(ctx context.Context, amount string) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(types.TotalMEVPendingKey, []byte(amount))
}

// GetTotalMEVPending retrieves the total pending MEV amount
func (k Keeper) GetTotalMEVPending(ctx context.Context) (string, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.TotalMEVPendingKey)
	if err != nil {
		return "0", err
	}
	if bz == nil {
		return "0", nil
	}
	return string(bz), nil
}

// ============================
// TOTAL BURNED OPERATIONS
// ============================

// SetTotalBurned stores the total burned amount
func (k Keeper) SetTotalBurned(ctx context.Context, amount string) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(types.TotalBurnedKey, []byte(amount))
}

// GetTotalBurned retrieves the total burned amount
func (k Keeper) GetTotalBurned(ctx context.Context) (string, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.TotalBurnedKey)
	if err != nil {
		return "0", err
	}
	if bz == nil {
		return "0", nil
	}
	return string(bz), nil
}

// ============================
// PREVIOUS INFLATION OPERATIONS
// ============================

// SetPreviousInflation stores the previous inflation rate
func (k Keeper) SetPreviousInflation(ctx context.Context, rate uint64) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, rate)
	return store.Set(types.PreviousInflationKey, bz)
}

// GetPreviousInflation retrieves the previous inflation rate
func (k Keeper) GetPreviousInflation(ctx context.Context) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.PreviousInflationKey)
	if err != nil {
		return 0, err
	}
	if bz == nil {
		return 0, nil
	}
	return binary.BigEndian.Uint64(bz), nil
}

// ============================
// ITERATOR HELPERS
// ============================

// IterateVestingSchedules iterates over all vesting schedules
// The callback should return true to stop iteration, false to continue
func (k Keeper) IterateVestingSchedules(ctx context.Context, cb func(schedule *types.VestingSchedule) bool) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.VestingSchedulePrefix, storeprefixend(types.VestingSchedulePrefix))
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var schedule types.VestingSchedule
		if err := k.cdc.Unmarshal(iterator.Value(), &schedule); err != nil {
			return fmt.Errorf("failed to create iterator for Valid: %w", err)
		}
		if cb(&schedule) {
			break // Callback returned true = stop iteration
		}
	}
	return nil
}

// IterateVoteLocks iterates over all vote locks
// The callback should return true to stop iteration, false to continue
func (k Keeper) IterateVoteLocks(ctx context.Context, cb func(lock *types.VoteLock) bool) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.VoteLockPrefix, storeprefixend(types.VoteLockPrefix))
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var lock types.VoteLock
		if err := k.cdc.Unmarshal(iterator.Value(), &lock); err != nil {
			return fmt.Errorf("failed to create iterator for Valid: %w", err)
		}
		if cb(&lock) {
			break // Callback returned true = stop iteration
		}
	}
	return nil
}

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
