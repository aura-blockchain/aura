package keeper

import (
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/bridge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================================================================
// TIME-LOCKED WITHDRAWALS
// ============================================================================

// CreateTimeLock creates a time-lock for large transfers
func (k Keeper) CreateTimeLock(
	ctx sdk.Context,
	transferId string,
	amount sdk.Int,
	denom string,
	recipient string,
) (*types.TimeLock, error) {
	params := k.GetParams(ctx)

	// Check if amount exceeds threshold
	if amount.LT(params.TimeLockThreshold) {
		return nil, fmt.Errorf("amount below time-lock threshold")
	}

	lockId := fmt.Sprintf("timelock-%s", transferId)
	lockTime := ctx.BlockTime()
	unlockTime := lockTime.Add(params.TimeLockDuration)

	timeLock := &types.TimeLock{
		LockId:         lockId,
		TransferId:     transferId,
		Amount:         amount,
		Denom:          denom,
		Recipient:      recipient,
		LockTime:       lockTime,
		UnlockTime:     unlockTime,
		Status:         types.TimeLockStatus_TIMELOCK_LOCKED,
		ChallengeCount: 0,
	}

	k.SetTimeLock(ctx, timeLock)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"time_lock_created",
			sdk.NewAttribute("lock_id", lockId),
			sdk.NewAttribute("transfer_id", transferId),
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("unlock_time", unlockTime.Format(time.RFC3339)),
		),
	)

	return timeLock, nil
}

// ProcessTimeLocks checks and releases time-locked transfers
func (k Keeper) ProcessTimeLocks(ctx sdk.Context) {
	timeLocks := k.GetAllTimeLocks(ctx)

	for _, lock := range timeLocks {
		if lock.Status == types.TimeLockStatus_TIMELOCK_LOCKED &&
			ctx.BlockTime().After(lock.UnlockTime) {

			// Release the time lock
			lock.Status = types.TimeLockStatus_TIMELOCK_UNLOCKED
			k.SetTimeLock(ctx, lock)

			// Process the transfer
			recipientAddr, err := sdk.AccAddressFromBech32(lock.Recipient)
			if err == nil {
				coins := sdk.NewCoins(sdk.NewCoin(lock.Denom, lock.Amount))
				if err := k.bankKeeper.SendCoinsFromModuleToAccount(
					ctx,
					types.ModuleName,
					recipientAddr,
					coins,
				); err != nil {
					ctx.Logger().Error("failed to release time lock", "lock_id", lock.LockId, "error", err)
					continue
				}
			}

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"time_lock_released",
					sdk.NewAttribute("lock_id", lock.LockId),
					sdk.NewAttribute("recipient", lock.Recipient),
					sdk.NewAttribute("amount", lock.Amount.String()),
				),
			)
		}
	}
}

// ChallengeTimeLock allows users to challenge a suspicious time-locked transfer
func (k Keeper) ChallengeTimeLock(
	ctx sdk.Context,
	lockId string,
	challenger string,
	evidence []byte,
) error {
	timeLock := k.GetTimeLock(ctx, lockId)
	if timeLock == nil {
		return fmt.Errorf("time lock not found: %s", lockId)
	}

	if timeLock.Status != types.TimeLockStatus_TIMELOCK_LOCKED {
		return fmt.Errorf("time lock is not locked: %s", timeLock.Status.String())
	}

	// Mark as challenged
	timeLock.Status = types.TimeLockStatus_TIMELOCK_CHALLENGED
	timeLock.ChallengeCount++
	k.SetTimeLock(ctx, timeLock)

	// Create fraud proof for the underlying transfer
	_, err := k.SubmitFraudProof(
		ctx,
		challenger,
		timeLock.TransferId,
		types.FraudType_FRAUD_AMOUNT_MISMATCH,
		evidence,
	)

	return err
}

// ============================================================================
// DAILY WITHDRAWAL LIMITS
// ============================================================================

// CheckWithdrawalLimit checks if a user is within their daily withdrawal limit
func (k Keeper) CheckWithdrawalLimit(
	ctx sdk.Context,
	address string,
	amount sdk.Int,
) error {
	params := k.GetParams(ctx)

	// Get or create withdrawal limit for user
	limit := k.GetWithdrawalLimit(ctx, address)
	if limit == nil {
		limit = &types.WithdrawalLimit{
			Address:        address,
			DailyLimit:     params.DailyWithdrawalLimit,
			WithdrawnToday: sdk.ZeroInt(),
			LastReset:      ctx.BlockTime(),
			Tier:           0, // Default tier
		}
	}

	// Check if we need to reset (new day)
	if ctx.BlockTime().Sub(limit.LastReset) >= 24*time.Hour {
		limit.WithdrawnToday = sdk.ZeroInt()
		limit.LastReset = ctx.BlockTime()
	}

	// Check if adding this amount would exceed the limit
	newTotal := limit.WithdrawnToday.Add(amount)
	if newTotal.GT(limit.DailyLimit) {
		return fmt.Errorf(
			"daily withdrawal limit exceeded: tried to withdraw %s, limit is %s, already withdrawn %s",
			amount.String(),
			limit.DailyLimit.String(),
			limit.WithdrawnToday.String(),
		)
	}

	// Update withdrawn amount
	limit.WithdrawnToday = newTotal
	k.SetWithdrawalLimit(ctx, limit)

	return nil
}

// UpdateWithdrawalLimit updates a user's withdrawal limit (e.g., for VIP users)
func (k Keeper) UpdateWithdrawalLimit(
	ctx sdk.Context,
	address string,
	newLimit sdk.Int,
	tier uint64,
) error {
	limit := k.GetWithdrawalLimit(ctx, address)
	if limit == nil {
		limit = &types.WithdrawalLimit{
			Address:        address,
			DailyLimit:     newLimit,
			WithdrawnToday: sdk.ZeroInt(),
			LastReset:      ctx.BlockTime(),
			Tier:           tier,
		}
	} else {
		limit.DailyLimit = newLimit
		limit.Tier = tier
	}

	k.SetWithdrawalLimit(ctx, limit)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"withdrawal_limit_updated",
			sdk.NewAttribute("address", address),
			sdk.NewAttribute("new_limit", newLimit.String()),
			sdk.NewAttribute("tier", fmt.Sprintf("%d", tier)),
		),
	)

	return nil
}

// ============================================================================
// CIRCUIT BREAKER
// ============================================================================

// UpdateCircuitBreaker updates circuit breaker metrics and checks for anomalies
func (k Keeper) UpdateCircuitBreaker(
	ctx sdk.Context,
	transferAmount sdk.Int,
	transferFailed bool,
) error {
	params := k.GetParams(ctx)

	if !params.CircuitBreakerEnabled {
		return nil
	}

	breaker := k.GetCircuitBreaker(ctx)
	if breaker == nil {
		breaker = k.initializeCircuitBreaker(ctx, params)
	}

	// Reset hourly metrics if needed
	if breaker.LastTriggered.IsZero() || ctx.BlockTime().Sub(breaker.LastTriggered) >= time.Hour {
		breaker.CurrentHourlyVolume = sdk.ZeroInt()
		breaker.CurrentFailedCount = 0
	}

	// Update metrics
	breaker.CurrentHourlyVolume = breaker.CurrentHourlyVolume.Add(transferAmount)
	if transferFailed {
		breaker.CurrentFailedCount++
	}

	// Check thresholds
	shouldTrip := false

	// Check single transfer limit
	if transferAmount.GT(breaker.MaxSingleTransfer) {
		shouldTrip = true
		ctx.Logger().Error("circuit breaker: single transfer exceeds limit",
			"amount", transferAmount.String(),
			"limit", breaker.MaxSingleTransfer.String())
	}

	// Check hourly volume
	if breaker.CurrentHourlyVolume.GT(breaker.MaxHourlyVolume) {
		shouldTrip = true
		ctx.Logger().Error("circuit breaker: hourly volume exceeded",
			"volume", breaker.CurrentHourlyVolume.String(),
			"limit", breaker.MaxHourlyVolume.String())
	}

	// Check failed transfers
	if breaker.CurrentFailedCount > breaker.MaxFailedTransfersPerHour {
		shouldTrip = true
		ctx.Logger().Error("circuit breaker: too many failed transfers",
			"count", breaker.CurrentFailedCount,
			"limit", breaker.MaxFailedTransfersPerHour)
	}

	if shouldTrip && breaker.Status == types.CircuitBreakerStatus_CIRCUIT_CLOSED {
		return k.TripCircuitBreaker(ctx, breaker)
	}

	k.SetCircuitBreaker(ctx, breaker)
	return nil
}

// TripCircuitBreaker opens the circuit breaker to pause bridge operations
func (k Keeper) TripCircuitBreaker(ctx sdk.Context, breaker *types.CircuitBreaker) error {
	breaker.Status = types.CircuitBreakerStatus_CIRCUIT_OPEN
	breaker.LastTriggered = ctx.BlockTime()
	breaker.AutoResetTime = ctx.BlockTime().Add(1 * time.Hour) // Auto-reset after 1 hour

	k.SetCircuitBreaker(ctx, breaker)

	// Emergency pause the bridge
	params := k.GetParams(ctx)
	params.EmergencyPaused = true
	k.SetParams(ctx, params)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"circuit_breaker_tripped",
			sdk.NewAttribute("hourly_volume", breaker.CurrentHourlyVolume.String()),
			sdk.NewAttribute("failed_count", fmt.Sprintf("%d", breaker.CurrentFailedCount)),
			sdk.NewAttribute("auto_reset_time", breaker.AutoResetTime.Format(time.RFC3339)),
		),
	)

	return fmt.Errorf("circuit breaker tripped: bridge operations paused")
}

// ResetCircuitBreaker manually resets the circuit breaker
func (k Keeper) ResetCircuitBreaker(ctx sdk.Context) error {
	breaker := k.GetCircuitBreaker(ctx)
	if breaker == nil {
		return fmt.Errorf("circuit breaker not initialized")
	}

	breaker.Status = types.CircuitBreakerStatus_CIRCUIT_CLOSED
	breaker.CurrentHourlyVolume = sdk.ZeroInt()
	breaker.CurrentFailedCount = 0
	breaker.LastTriggered = time.Time{}

	k.SetCircuitBreaker(ctx, breaker)

	// Unpause the bridge
	params := k.GetParams(ctx)
	params.EmergencyPaused = false
	k.SetParams(ctx, params)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"circuit_breaker_reset",
			sdk.NewAttribute("status", breaker.Status.String()),
		),
	)

	return nil
}

// CheckAutoResetCircuitBreaker checks if circuit breaker should auto-reset
func (k Keeper) CheckAutoResetCircuitBreaker(ctx sdk.Context) {
	breaker := k.GetCircuitBreaker(ctx)
	if breaker == nil {
		return
	}

	if breaker.Status == types.CircuitBreakerStatus_CIRCUIT_OPEN &&
		!breaker.AutoResetTime.IsZero() &&
		ctx.BlockTime().After(breaker.AutoResetTime) {

		breaker.Status = types.CircuitBreakerStatus_CIRCUIT_HALF_OPEN
		k.SetCircuitBreaker(ctx, breaker)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"circuit_breaker_half_open",
				sdk.NewAttribute("status", breaker.Status.String()),
			),
		)
	}
}

func (k Keeper) initializeCircuitBreaker(ctx sdk.Context, params types.Params) *types.CircuitBreaker {
	return &types.CircuitBreaker{
		BreakerId:                 "main",
		Status:                    types.CircuitBreakerStatus_CIRCUIT_CLOSED,
		MaxHourlyVolume:           params.MaxHourlyVolume,
		MaxSingleTransfer:         params.MaxTransferAmount,
		MaxFailedTransfersPerHour: params.MaxFailedTransfersPerHour,
		CurrentHourlyVolume:       sdk.ZeroInt(),
		CurrentFailedCount:        0,
		LastTriggered:             time.Time{},
		AutoResetTime:             time.Time{},
	}
}

// ============================================================================
// STORAGE
// ============================================================================

// TimeLock storage
func (k Keeper) GetTimeLock(ctx sdk.Context, lockId string) *types.TimeLock {
	store := ctx.KVStore(k.storeKey)
	key := types.TimeLockKey(lockId)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var lock types.TimeLock
	k.cdc.MustUnmarshal(bz, &lock)
	return &lock
}

func (k Keeper) SetTimeLock(ctx sdk.Context, lock *types.TimeLock) {
	store := ctx.KVStore(k.storeKey)
	key := types.TimeLockKey(lock.LockId)

	bz := k.cdc.MustMarshal(lock)
	store.Set(key, bz)
}

func (k Keeper) GetAllTimeLocks(ctx sdk.Context) []*types.TimeLock {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.TimeLockPrefix)
	defer iterator.Close()

	locks := []*types.TimeLock{}
	for ; iterator.Valid(); iterator.Next() {
		var lock types.TimeLock
		k.cdc.MustUnmarshal(iterator.Value(), &lock)
		locks = append(locks, &lock)
	}

	return locks
}

// WithdrawalLimit storage
func (k Keeper) GetWithdrawalLimit(ctx sdk.Context, address string) *types.WithdrawalLimit {
	store := ctx.KVStore(k.storeKey)
	key := types.WithdrawalLimitKey(address)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var limit types.WithdrawalLimit
	k.cdc.MustUnmarshal(bz, &limit)
	return &limit
}

func (k Keeper) SetWithdrawalLimit(ctx sdk.Context, limit *types.WithdrawalLimit) {
	store := ctx.KVStore(k.storeKey)
	key := types.WithdrawalLimitKey(limit.Address)

	bz := k.cdc.MustMarshal(limit)
	store.Set(key, bz)
}

// CircuitBreaker storage
func (k Keeper) GetCircuitBreaker(ctx sdk.Context) *types.CircuitBreaker {
	store := ctx.KVStore(k.storeKey)
	key := types.CircuitBreakerKey()

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var breaker types.CircuitBreaker
	k.cdc.MustUnmarshal(bz, &breaker)
	return &breaker
}

func (k Keeper) SetCircuitBreaker(ctx sdk.Context, breaker *types.CircuitBreaker) {
	store := ctx.KVStore(k.storeKey)
	key := types.CircuitBreakerKey()

	bz := k.cdc.MustMarshal(breaker)
	store.Set(key, bz)
}
