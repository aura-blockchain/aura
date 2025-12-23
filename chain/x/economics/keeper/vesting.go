package keeper

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/economics/types"
	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
)

// ============================
// SCHEDULE ID COUNTER
// ============================

// GetNextScheduleID returns the next schedule ID and increments the counter
func (k Keeper) GetNextScheduleID(ctx context.Context) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.NextScheduleIDKey)
	if err != nil {
		return 0, err
	}

	var id uint64
	if bz == nil {
		id = 1
	} else {
		id = binary.BigEndian.Uint64(bz)
	}

	// Increment for next time
	nextID := id + 1
	nextBz := make([]byte, 8)
	binary.BigEndian.PutUint64(nextBz, nextID)
	if err := store.Set(types.NextScheduleIDKey, nextBz); err != nil {
		return 0, err
	}

	return id, nil
}

// ============================
// VESTING SCHEDULE OPERATIONS
// ============================

// SetVestingSchedule stores a vesting schedule
func (k Keeper) SetVestingSchedule(ctx context.Context, schedule *economicspb.VestingSchedule) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetVestingScheduleKey(schedule.Id)
	bz, err := k.cdc.Marshal(schedule)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	return store.Set(key, bz)
}

// GetVestingSchedule retrieves a vesting schedule
func (k Keeper) GetVestingSchedule(ctx context.Context, scheduleID string) (*economicspb.VestingSchedule, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetVestingScheduleKey(scheduleID)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrVestingScheduleNotFound
	}

	var schedule economicspb.VestingSchedule
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

// IterateVestingSchedules iterates over all vesting schedules
func (k Keeper) IterateVestingSchedules(ctx context.Context, cb func(schedule *economicspb.VestingSchedule) bool) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.VestingSchedulePrefix, storeprefixend(types.VestingSchedulePrefix))
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var schedule economicspb.VestingSchedule
		if err := k.cdc.Unmarshal(iterator.Value(), &schedule); err != nil {
			return fmt.Errorf("failed to create iterator for Valid: %w", err)
		}
		if cb(&schedule) {
			break
		}
	}
	return nil
}

// ============================
// USER VESTING INDEX OPERATIONS
// ============================

// SetUserVestingIndex stores the list of vesting schedule IDs for a user
func (k Keeper) SetUserVestingIndex(ctx context.Context, userAddress string, scheduleIDs []string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetUserVestingIndexKey(userAddress)

	// Simple encoding: join with null bytes
	var bz []byte
	for i, id := range scheduleIDs {
		if i > 0 {
			bz = append(bz, 0)
		}
		bz = append(bz, []byte(id)...)
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
	if bz == nil || len(bz) == 0 {
		return []string{}, nil
	}

	// Simple decoding: split by null bytes
	scheduleIDs := make([]string, 0, 64)
	start := 0
	for i, b := range bz {
		if b == 0 {
			scheduleIDs = append(scheduleIDs, string(bz[start:i]))
			start = i + 1
		}
	}
	if start < len(bz) {
		scheduleIDs = append(scheduleIDs, string(bz[start:]))
	}
	return scheduleIDs, nil
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
// VESTING CALCULATIONS
// ============================

// CalculateVestedAmount calculates the vested amount at a given time
func (k Keeper) CalculateVestedAmount(schedule *economicspb.VestingSchedule, currentTime time.Time) (sdkmath.Int, error) {
	// Check if schedule is revoked
	if schedule.Revoked {
		return schedule.VestedAmount.Amount, nil
	}

	// Get start and end times (already time.Time due to stdtime annotation)
	startTime := schedule.StartTime
	endTime := schedule.EndTime

	// Check if cliff period has passed
	cliffEnd := startTime.Add(time.Duration(schedule.CliffDuration) * time.Second)
	if currentTime.Before(cliffEnd) {
		return sdkmath.ZeroInt(), nil
	}

	// Check if vesting is complete
	if currentTime.After(endTime) || currentTime.Equal(endTime) {
		return schedule.OriginalAmount.Amount, nil
	}

	// Get total amount
	totalAmount := schedule.OriginalAmount.Amount

	switch schedule.VestingType {
	case economicspb.VestingType_VESTING_TYPE_LINEAR:
		// Calculate linear vesting
		vestingDuration := endTime.Sub(startTime).Seconds()
		elapsedTime := currentTime.Sub(startTime).Seconds()

		// Calculate vested amount: (totalAmount * elapsedTime) / vestingDuration
		elapsed := big.NewFloat(elapsedTime)
		duration := big.NewFloat(vestingDuration)

		totalAmountFloat := new(big.Float).SetInt(totalAmount.BigInt())
		vestedAmount := new(big.Float).Mul(totalAmountFloat, elapsed)
		vestedAmount.Quo(vestedAmount, duration)

		result, _ := vestedAmount.Int(nil)
		return sdkmath.NewIntFromBigInt(result), nil

	case economicspb.VestingType_VESTING_TYPE_MILESTONE:
		// For milestone vesting, would need additional milestone data
		// For now, return 0 (would be implemented with milestone tracking)
		return sdkmath.ZeroInt(), nil

	default:
		return sdkmath.ZeroInt(), fmt.Errorf("unknown vesting type: %v", schedule.VestingType)
	}
}

// ReleaseVestedTokensInternal releases vested tokens to the beneficiary (internal method)
func (k Keeper) ReleaseVestedTokensInternal(ctx context.Context, scheduleID string) (string, error) {
	schedule, err := k.GetVestingSchedule(ctx, scheduleID)
	if err != nil {
		return "0", err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentTime := sdkCtx.BlockTime()

	vestedAmount, err := k.CalculateVestedAmount(schedule, currentTime)
	if err != nil {
		return "0", err
	}

	// Calculate releasable amount
	totalVested := new(big.Int)
	totalVested.SetString(vestedAmount.String(), 10)

	alreadyReleased := schedule.VestedAmount.Amount.BigInt()

	releasable := new(big.Int).Sub(totalVested, alreadyReleased)

	if releasable.Sign() <= 0 {
		return "0", types.ErrNoVestedTokens
	}

	// Update schedule - create new Int from string and add to the coin
	releasableInt, ok := sdkmath.NewIntFromString(releasable.String())
	if !ok {
		return "0", fmt.Errorf("failed to convert releasable amount")
	}
	newVestedCoin := schedule.VestedAmount.AddAmount(releasableInt)
	schedule.VestedAmount = newVestedCoin
	if err := k.SetVestingSchedule(ctx, schedule); err != nil {
		return "0", err
	}

	return releasable.String(), nil
}

// RevokeVestingScheduleInternal revokes a vesting schedule (internal method)
func (k Keeper) RevokeVestingScheduleInternal(ctx context.Context, scheduleID string) error {
	schedule, err := k.GetVestingSchedule(ctx, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	if schedule.Revoked {
		return types.ErrVestingAlreadyRevoked
	}

	schedule.Revoked = true
	return k.SetVestingSchedule(ctx, schedule)
}

// ============================
// VOTE LOCK OPERATIONS
// ============================

// SetVoteLock stores a vote lock
func (k Keeper) SetVoteLock(ctx context.Context, lock *economicspb.VoteLock) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetVoteLockKey(lock.Id)
	bz, err := k.cdc.Marshal(lock)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	return store.Set(key, bz)
}

// GetVoteLock retrieves a vote lock
func (k Keeper) GetVoteLock(ctx context.Context, lockID string) (*economicspb.VoteLock, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetVoteLockKey(lockID)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrVoteLockNotFound
	}

	var lock economicspb.VoteLock
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

// IterateVoteLocks iterates over all vote locks
func (k Keeper) IterateVoteLocks(ctx context.Context, cb func(lock *economicspb.VoteLock) bool) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.VoteLockPrefix, storeprefixend(types.VoteLockPrefix))
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var lock economicspb.VoteLock
		if err := k.cdc.Unmarshal(iterator.Value(), &lock); err != nil {
			return fmt.Errorf("failed to create iterator for Valid: %w", err)
		}
		if cb(&lock) {
			break
		}
	}
	return nil
}

// ============================
// USER VOTE LOCK INDEX OPERATIONS
// ============================

// SetUserVoteLockIndex stores the list of vote lock IDs for a user
func (k Keeper) SetUserVoteLockIndex(ctx context.Context, userAddress string, lockIDs []string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetUserVoteLockIndexKey(userAddress)

	// Simple encoding: join with null bytes
	var bz []byte
	for i, id := range lockIDs {
		if i > 0 {
			bz = append(bz, 0)
		}
		bz = append(bz, []byte(id)...)
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
	if bz == nil || len(bz) == 0 {
		return []string{}, nil
	}

	// Simple decoding: split by null bytes
	lockIDs := make([]string, 0, 64)
	start := 0
	for i, b := range bz {
		if b == 0 {
			lockIDs = append(lockIDs, string(bz[start:i]))
			start = i + 1
		}
	}
	if start < len(bz) {
		lockIDs = append(lockIDs, string(bz[start:]))
	}
	return lockIDs, nil
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

// CalculateVotingPower calculates voting power based on locked tokens
// Uses integer math (basis points) for deterministic consensus
func (k Keeper) CalculateVotingPowerFromDuration(amount string, lockDuration int64) string {
	// Parse amount
	amountBig := new(big.Int)
	amountBig.SetString(amount, 10)

	// Simple formula: voting power = amount * (1 + lockDuration / maxDuration)
	// This incentivizes longer locks
	// Using basis points (10000 = 1.0) for deterministic integer math
	const basisPointsBase int64 = 10000
	maxDuration := int64(31536000) // 1 year in seconds

	// Calculate multiplier in basis points: 10000 + (lockDuration * 10000 / maxDuration)
	multiplierBps := basisPointsBase
	if lockDuration > 0 {
		durationBonusBps := (lockDuration * basisPointsBase) / maxDuration
		if durationBonusBps > basisPointsBase {
			durationBonusBps = basisPointsBase // Cap at 2x (10000 + 10000 = 20000 bps)
		}
		multiplierBps += durationBonusBps
	}

	// Calculate voting power using integer math
	// votingPower = amount * multiplierBps / basisPointsBase
	votingPower := new(big.Int).Mul(amountBig, big.NewInt(multiplierBps))
	votingPower.Div(votingPower, big.NewInt(basisPointsBase))

	return votingPower.String()
}
