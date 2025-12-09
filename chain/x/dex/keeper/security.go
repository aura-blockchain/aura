package keeper

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// ============================================================================
// 1. FRONT-RUNNING PROTECTION
// ============================================================================

// CheckFrontRunningProtection ensures minimum block delay between swaps
func (k Keeper) CheckFrontRunningProtection(
	ctx sdk.Context,
	address string,
	poolID string,
) error {
	params := k.GetSecurityParams(ctx)
	if params.MinBlockDelay == 0 {
		return nil // Protection disabled
	}

	// Get last trade for this address/pool
	lastTrade := k.GetLastTradeBlock(ctx, address, poolID)
	currentBlock := uint64(ctx.BlockHeight())

	// If no previous trade (lastTrade == 0), allow the first trade
	if lastTrade > 0 && currentBlock-lastTrade < params.MinBlockDelay {
		return fmt.Errorf(
			"must wait %d blocks between trades (last: %d, current: %d): %w",
			params.MinBlockDelay,
			lastTrade,
			currentBlock,
			types.ErrFrontRunningDetected,
		)
	}

	return nil
}

// RecordTradeBlock records the block height of a trade
func (k Keeper) RecordTradeBlock(ctx sdk.Context, address string, poolID string) {
	store := ctx.KVStore(k.storeKey)
	key := types.TradeBlockKey(address, poolID)
	height := uint64(ctx.BlockHeight())

	store.Set(key, sdk.Uint64ToBigEndian(height))

	// Update trade history
	k.UpdateTradeHistory(ctx, address, poolID)
}

// GetLastTradeBlock retrieves last trade block for address/pool
func (k Keeper) GetLastTradeBlock(ctx sdk.Context, address string, poolID string) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := types.TradeBlockKey(address, poolID)

	bz := store.Get(key)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// ============================================================================
// 2. TIME-WEIGHTED AVERAGE PRICE (TWAP) ORACLE
// ============================================================================

// RecordTWAPObservation records price observation for TWAP calculation
func (k Keeper) RecordTWAPObservation(ctx sdk.Context, poolID string) error {
	pool := k.GetPool(ctx, poolID)
	if pool == nil {
		return types.ErrPoolNotFound
	}

	// Parse reserves from strings
	reserveA, err := k.parseReserve(pool.ReserveA)
	if err != nil {
		return err
	}
	reserveB, err := k.parseReserve(pool.ReserveB)
	if err != nil {
		return err
	}

	// Calculate spot price
	spotPrice := reserveB.ToLegacyDec().Quo(reserveA.ToLegacyDec())

	// Get previous observation to calculate time delta
	prevObservation := k.GetLatestTWAPObservation(ctx, poolID)

	var cumulativePrice sdkmath.LegacyDec
	if prevObservation != nil {
		// prevObservation.CumulativePrice is already LegacyDec
		prevCumulativePrice := prevObservation.CumulativePrice

		// Time elapsed since last observation
		timeElapsed := ctx.BlockTime().Unix() - prevObservation.Timestamp.Unix()

		// Cumulative price = previous cumulative + (spot price * time elapsed)
		cumulativePrice = prevCumulativePrice.Add(
			spotPrice.MulInt64(timeElapsed),
		)
	} else {
		// First observation
		cumulativePrice = sdkmath.LegacyZeroDec()
	}

	observation := &types.TWAPPrice{
		PoolId:          poolID,
		CumulativePrice: cumulativePrice,
		BlockHeight:     ctx.BlockHeight(),
		Timestamp:       ctx.BlockTime(),
		ReserveA:        pool.ReserveA,
		ReserveB:        pool.ReserveB,
		SpotPrice:       spotPrice,
	}

	if err := k.SetTWAPObservation(ctx, observation); err != nil {
		ctx.Logger().Error("failed to store TWAP observation", "pool_id", poolID, "error", err)
	}

	// Prune old observations outside TWAP window
	k.PruneTWAPObservations(ctx, poolID)

	return nil
}

// GetTWAPPrice calculates TWAP over the specified window
func (k Keeper) GetTWAPPrice(ctx sdk.Context, poolID string, windowBlocks uint64) (sdkmath.LegacyDec, error) {
	params := k.GetSecurityParams(ctx)
	if windowBlocks == 0 {
		windowBlocks = params.TwapWindowBlocks
	}

	observations := k.GetTWAPObservations(ctx, poolID, windowBlocks)
	if len(observations) < 2 {
		// Need at least 2 observations for TWAP
		pool := k.GetPool(ctx, poolID)
		if pool == nil {
			return sdkmath.LegacyZeroDec(), types.ErrPoolNotFound
		}
		// Parse reserves and return spot price if insufficient data
		reserveA, err := k.parseReserve(pool.ReserveA)
		if err != nil {
			return sdkmath.LegacyZeroDec(), err
		}
		if reserveA.IsZero() {
			return sdkmath.LegacyZeroDec(), fmt.Errorf("pool reserve is zero")
		}
		reserveB, err := k.parseReserve(pool.ReserveB)
		if err != nil {
			return sdkmath.LegacyZeroDec(), err
		}
		return reserveB.ToLegacyDec().Quo(reserveA.ToLegacyDec()), nil
	}

	// TWAP = (cumulative_price_end - cumulative_price_start) / time_elapsed
	latest := observations[0]
	oldest := observations[len(observations)-1]

	timeElapsed := latest.Timestamp.Unix() - oldest.Timestamp.Unix()
	if timeElapsed == 0 {
		return latest.SpotPrice, nil
	}

	// Cumulative prices are already LegacyDec
	latestCumulativePrice := latest.CumulativePrice
	oldestCumulativePrice := oldest.CumulativePrice

	twap := latestCumulativePrice.Sub(oldestCumulativePrice).QuoInt64(timeElapsed)

	return twap, nil
}

// GetTWAPPriceWithCount calculates TWAP and returns the observation count.
// This is used to determine if there are sufficient observations (>= MinTWAPObservations)
// to trust the TWAP price, preventing oracle manipulation attacks.
//
// SECURITY: The observation count check is critical for preventing manipulation.
// With insufficient observations, attackers could manipulate early price observations
// to affect TWAP calculations. By requiring MinTWAPObservations (100), we ensure
// sufficient historical data exists before trusting TWAP.
//
// Returns:
//   - price: TWAP price if sufficient observations, zero otherwise
//   - count: number of TWAP observations available
//   - error: any error during calculation
func (k Keeper) GetTWAPPriceWithCount(ctx sdk.Context, poolID string, windowBlocks uint64) (price sdkmath.LegacyDec, count int, err error) {
	params := k.GetSecurityParams(ctx)
	if windowBlocks == 0 {
		windowBlocks = params.TwapWindowBlocks
	}

	observations := k.GetTWAPObservations(ctx, poolID, windowBlocks)
	count = len(observations)

	if count < 2 {
		// Need at least 2 observations for TWAP calculation
		return sdkmath.LegacyZeroDec(), count, fmt.Errorf("insufficient observations for TWAP: %d < 2", count)
	}

	// TWAP = (cumulative_price_end - cumulative_price_start) / time_elapsed
	latest := observations[0]
	oldest := observations[len(observations)-1]

	timeElapsed := latest.Timestamp.Unix() - oldest.Timestamp.Unix()
	if timeElapsed == 0 {
		return latest.SpotPrice, count, nil
	}

	// Cumulative prices are already LegacyDec
	latestCumulativePrice := latest.CumulativePrice
	oldestCumulativePrice := oldest.CumulativePrice

	twap := latestCumulativePrice.Sub(oldestCumulativePrice).QuoInt64(timeElapsed)

	return twap, count, nil
}

// SetTWAPObservation stores a TWAP observation
func (k Keeper) SetTWAPObservation(ctx sdk.Context, obs *types.TWAPPrice) error {
	store := ctx.KVStore(k.storeKey)
	key := types.TWAPKey(obs.PoolId, obs.BlockHeight)

	bz, err := k.cdc.Marshal(obs)
	if err != nil {
		return types.ErrMarshalFailed.Wrapf("failed to marshal TWAP observation for pool %s: %v", obs.PoolId, err)
	}
	store.Set(key, bz)
	return nil
}

// GetLatestTWAPObservation retrieves the most recent TWAP observation
func (k Keeper) GetLatestTWAPObservation(ctx sdk.Context, poolID string) *types.TWAPPrice {
	observations := k.GetTWAPObservations(ctx, poolID, 1)
	if len(observations) == 0 {
		return nil
	}
	return observations[0]
}

// GetTWAPObservations retrieves TWAP observations within window
func (k Keeper) GetTWAPObservations(ctx sdk.Context, poolID string, windowBlocks uint64) []*types.TWAPPrice {
	store := ctx.KVStore(k.storeKey)

	cutoffBlock := uint64(ctx.BlockHeight()) - windowBlocks
	prefix := types.TWAPPrefixKey(poolID)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var observations []*types.TWAPPrice
	for ; iterator.Valid(); iterator.Next() {
		var obs types.TWAPPrice
		if err := k.cdc.Unmarshal(iterator.Value(), &obs); err != nil {
			ctx.Logger().Error("failed to unmarshal TWAP observation, skipping",
				"pool_id", poolID,
				"error", err,
				"data_len", len(iterator.Value()))
			continue
		}

		if uint64(obs.BlockHeight) >= cutoffBlock {
			observations = append(observations, &obs)
		}
	}

	return observations
}

// PruneTWAPObservations removes observations outside the window
func (k Keeper) PruneTWAPObservations(ctx sdk.Context, poolID string) {
	params := k.GetSecurityParams(ctx)
	cutoffBlock := uint64(ctx.BlockHeight()) - params.TwapWindowBlocks

	store := ctx.KVStore(k.storeKey)
	prefix := types.TWAPPrefixKey(poolID)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var keysToDelete [][]byte
	for ; iterator.Valid(); iterator.Next() {
		var obs types.TWAPPrice
		if err := k.cdc.Unmarshal(iterator.Value(), &obs); err != nil {
			ctx.Logger().Error("failed to unmarshal TWAP observation during pruning, deleting corrupted entry",
				"pool_id", poolID,
				"error", err,
				"data_len", len(iterator.Value()))
			// Delete corrupted entries during pruning
			keysToDelete = append(keysToDelete, iterator.Key())
			continue
		}

		if uint64(obs.BlockHeight) < cutoffBlock {
			keysToDelete = append(keysToDelete, iterator.Key())
		}
	}

	for _, key := range keysToDelete {
		store.Delete(key)
	}
}

// ValidatePriceMovement checks if price change is within acceptable bounds.
// Rejects price movements > 10% per block to prevent flash loan attacks.
func (k Keeper) ValidatePriceMovement(ctx sdk.Context, poolID string, newPrice sdkmath.LegacyDec) error {
	const MaxPriceChangePercent = 10 // 10% maximum per block

	lastPrice := k.GetLastRecordedPrice(ctx, poolID)

	// First price observation, no validation needed
	if lastPrice.IsZero() {
		return nil
	}

	// Calculate percentage change: |new - old| / old * 100
	priceDiff := newPrice.Sub(lastPrice).Abs()
	percentChange := priceDiff.Quo(lastPrice).MulInt64(100)

	maxChange := sdkmath.LegacyNewDec(MaxPriceChangePercent)
	if percentChange.GT(maxChange) {
		return fmt.Errorf(
			"price movement too large: %.2f%% exceeds maximum of %d%%: %w",
			percentChange.MustFloat64(),
			MaxPriceChangePercent,
			types.ErrPriceManipulation,
		)
	}

	return nil
}

// SetLastRecordedPrice stores the last validated price for sanity checks
func (k Keeper) SetLastRecordedPrice(ctx sdk.Context, poolID string, price sdkmath.LegacyDec) {
	store := ctx.KVStore(k.storeKey)
	key := types.LastPriceKey(poolID)
	store.Set(key, []byte(price.String()))
}

// GetLastRecordedPrice retrieves the last recorded price
func (k Keeper) GetLastRecordedPrice(ctx sdk.Context, poolID string) sdkmath.LegacyDec {
	store := ctx.KVStore(k.storeKey)
	key := types.LastPriceKey(poolID)

	bz := store.Get(key)
	if bz == nil {
		return sdkmath.LegacyZeroDec()
	}

	price, err := sdkmath.LegacyNewDecFromStr(string(bz))
	if err != nil {
		return sdkmath.LegacyZeroDec()
	}

	return price
}

// RecordAllPoolPrices records TWAP observations for all active pools with price validation.
// This should be called in EndBlocker to build TWAP history and reject manipulation.
// DEPRECATED: Use RecordAllPoolPricesBatched for production to prevent consensus failure
func (k Keeper) RecordAllPoolPrices(ctx sdk.Context) {
	pools := k.GetAllPools(ctx)

	for _, pool := range pools {
		// Skip empty pools
		reserveA, err := k.parseReserve(pool.ReserveA)
		if err != nil || reserveA.IsZero() {
			continue
		}
		reserveB, err := k.parseReserve(pool.ReserveB)
		if err != nil || reserveB.IsZero() {
			continue
		}

		// Calculate spot price
		spotPrice := reserveB.ToLegacyDec().Quo(reserveA.ToLegacyDec())

		// Validate price movement before recording
		if err := k.ValidatePriceMovement(ctx, pool.PoolId, spotPrice); err != nil {
			// Price movement too large, emit event and skip recording
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"suspicious_price_movement",
					sdk.NewAttribute("pool_id", pool.PoolId),
					sdk.NewAttribute("rejected_price", spotPrice.String()),
					sdk.NewAttribute("last_price", k.GetLastRecordedPrice(ctx, pool.PoolId).String()),
					sdk.NewAttribute("reason", err.Error()),
				),
			)
			continue
		}

		// Record TWAP observation
		if err := k.RecordTWAPObservation(ctx, pool.PoolId); err != nil {
			// Log error but don't fail EndBlocker
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"twap_recording_error",
					sdk.NewAttribute("pool_id", pool.PoolId),
					sdk.NewAttribute("error", err.Error()),
				),
			)
			continue
		}

		// Update last recorded price
		k.SetLastRecordedPrice(ctx, pool.PoolId, spotPrice)
	}
}

// RecordAllPoolPricesBatched records TWAP observations for up to 'limit' pools per call.
// Uses cursor-based rotation to cycle through all pools across multiple blocks,
// ensuring all pools are eventually updated without blocking consensus.
//
// SECURITY: This function is designed to prevent consensus failure by limiting
// the number of operations per block. With many pools, unbounded TWAP recording
// causes block production to exceed timeout, halting the chain.
//
// The rotation ensures that even with 1000+ pools, each pool gets its TWAP
// updated regularly (every N blocks where N = total_pools / limit).
//
// Returns: number of pools processed in this batch
func (k Keeper) RecordAllPoolPricesBatched(ctx sdk.Context, limit int) int {
	if limit <= 0 {
		limit = types.MaxPoolsTWAPPerBlock
	}

	pools := k.GetAllPools(ctx)
	if len(pools) == 0 {
		return 0
	}

	store := ctx.KVStore(k.storeKey)

	// Get cursor (index of last processed pool)
	cursorKey := types.TWAPCursorKey()
	cursorBytes := store.Get(cursorKey)

	startIdx := 0
	if cursorBytes != nil && len(cursorBytes) >= 8 {
		// Decode cursor as uint64
		startIdx = int(sdk.BigEndianToUint64(cursorBytes))
	}

	// Wrap around if cursor is beyond pool list
	if startIdx >= len(pools) {
		startIdx = 0
	}

	processed := 0
	currentIdx := startIdx

	// Process up to 'limit' pools in round-robin fashion
	for processed < limit && processed < len(pools) {
		pool := pools[currentIdx]

		// Skip empty pools
		reserveA, err := k.parseReserve(pool.ReserveA)
		if err != nil || reserveA.IsZero() {
			// Move to next pool
			currentIdx = (currentIdx + 1) % len(pools)
			processed++
			// Prevent infinite loop if all pools are invalid
			if processed >= len(pools) {
				break
			}
			continue
		}

		reserveB, err := k.parseReserve(pool.ReserveB)
		if err != nil || reserveB.IsZero() {
			// Move to next pool
			currentIdx = (currentIdx + 1) % len(pools)
			processed++
			if processed >= len(pools) {
				break
			}
			continue
		}

		// Calculate spot price
		spotPrice := reserveB.ToLegacyDec().Quo(reserveA.ToLegacyDec())

		// Validate price movement before recording
		if err := k.ValidatePriceMovement(ctx, pool.PoolId, spotPrice); err != nil {
			// Price movement too large, emit event and skip recording
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"suspicious_price_movement",
					sdk.NewAttribute("pool_id", pool.PoolId),
					sdk.NewAttribute("rejected_price", spotPrice.String()),
					sdk.NewAttribute("last_price", k.GetLastRecordedPrice(ctx, pool.PoolId).String()),
					sdk.NewAttribute("reason", err.Error()),
				),
			)
			// Move to next pool
			currentIdx = (currentIdx + 1) % len(pools)
			processed++
			continue
		}

		// Record TWAP observation
		if err := k.RecordTWAPObservation(ctx, pool.PoolId); err != nil {
			// Log error but don't fail EndBlocker
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"twap_recording_error",
					sdk.NewAttribute("pool_id", pool.PoolId),
					sdk.NewAttribute("error", err.Error()),
				),
			)
			// Move to next pool
			currentIdx = (currentIdx + 1) % len(pools)
			processed++
			continue
		}

		// Update last recorded price
		k.SetLastRecordedPrice(ctx, pool.PoolId, spotPrice)

		// Move to next pool
		currentIdx = (currentIdx + 1) % len(pools)
		processed++
	}

	// Save cursor for next block (next pool to process)
	nextIdx := currentIdx % len(pools)
	cursorBytes = sdk.Uint64ToBigEndian(uint64(nextIdx))
	store.Set(cursorKey, cursorBytes)

	return processed
}

// ============================================================================
// 3. FLASH LOAN ATTACK PROTECTION
// ============================================================================

// CheckFlashLoanProtection prevents rapid add/remove liquidity
func (k Keeper) CheckFlashLoanProtection(
	ctx sdk.Context,
	provider string,
	poolID string,
	isAdding bool,
) error {
	params := k.GetSecurityParams(ctx)
	if params.MinLiquidityBlocks == 0 {
		return nil // Protection disabled
	}

	lastBlock := k.GetLastLiquidityBlock(ctx, provider, poolID)
	currentBlock := uint64(ctx.BlockHeight())

	if lastBlock > 0 && currentBlock-lastBlock < params.MinLiquidityBlocks {
		return fmt.Errorf(
			"must wait %d blocks between liquidity operations (last: %d, current: %d): %w",
			params.MinLiquidityBlocks,
			lastBlock,
			currentBlock,
			types.ErrFlashLoanDetected,
		)
	}

	return nil
}

// RecordLiquidityBlock records the block height of liquidity operation
func (k Keeper) RecordLiquidityBlock(ctx sdk.Context, provider string, poolID string) {
	store := ctx.KVStore(k.storeKey)
	key := types.LiquidityBlockKey(provider, poolID)
	height := uint64(ctx.BlockHeight())

	store.Set(key, sdk.Uint64ToBigEndian(height))
}

// GetLastLiquidityBlock retrieves last liquidity operation block
func (k Keeper) GetLastLiquidityBlock(ctx sdk.Context, provider string, poolID string) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := types.LiquidityBlockKey(provider, poolID)

	bz := store.Get(key)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// ============================================================================
// 4. MEV MITIGATION STRATEGIES
// ============================================================================

// CheckMEVProtection implements MEV mitigation by address
func (k Keeper) CheckMEVProtection(ctx sdk.Context, address string) error {
	params := k.GetSecurityParams(ctx)
	if !params.MevProtectionEnabled {
		return nil
	}

	// Check maximum swaps per block
	swapsInBlock := k.GetSwapsInCurrentBlock(ctx, address)
	if swapsInBlock >= params.MaxSwapsPerBlock {
		return fmt.Errorf(
			"maximum swaps per block exceeded: %d >= %d: %w",
			swapsInBlock,
			params.MaxSwapsPerBlock,
			types.ErrMEVDetected,
		)
	}

	return nil
}

// GetSwapsInCurrentBlock counts swaps by address in current block
func (k Keeper) GetSwapsInCurrentBlock(ctx sdk.Context, address string) uint64 {
	history := k.GetTradeHistory(ctx, address)
	if history == nil {
		return 0
	}

	if history.LastTradeBlock == ctx.BlockHeight() {
		return history.TradesInBlock
	}

	return 0
}

// ============================================================================
// 5. POOL-SPECIFIC SLIPPAGE LIMITS
// ============================================================================

// CheckPoolSlippageLimit validates slippage against pool-specific limits
func (k Keeper) CheckPoolSlippageLimit(
	ctx sdk.Context,
	poolID string,
	priceImpact sdkmath.LegacyDec,
) error {
	params := k.GetSecurityParams(ctx)

	// MaxPriceImpactPercent is already LegacyDec
	maxPriceImpact := params.MaxPriceImpactPercent

	// Convert price impact percentage to decimal (5% = 5.0)
	if priceImpact.GT(maxPriceImpact) {
		return fmt.Errorf(
			"price impact %s%% exceeds maximum %s%%: %w",
			priceImpact.String(),
			maxPriceImpact.String(),
			types.ErrPriceImpactTooHigh,
		)
	}

	return nil
}

// ============================================================================
// 6. MAXIMUM TRADE SIZE CAPS
// ============================================================================

// CheckMaxTradeSize validates trade size against pool reserves
func (k Keeper) CheckMaxTradeSize(
	ctx sdk.Context,
	poolID string,
	amountIn sdkmath.Int,
) error {
	params := k.GetSecurityParams(ctx)
	pool := k.GetPool(ctx, poolID)
	if pool == nil {
		return types.ErrPoolNotFound
	}

	// Parse reserve
	reserveA, err := k.parseReserve(pool.ReserveA)
	if err != nil {
		return err
	}

	// MaxTradeSizePercent is already LegacyDec
	maxTradeSizePercent := params.MaxTradeSizePercent

	// Determine reserve based on input denom
	// Assuming input is denomA, check against reserveA
	maxAmount := reserveA.ToLegacyDec().Mul(maxTradeSizePercent).TruncateInt()

	if amountIn.GT(maxAmount) {
		return fmt.Errorf(
			"trade size %s exceeds maximum %s (%s%% of pool): %w",
			amountIn.String(),
			maxAmount.String(),
			maxTradeSizePercent.MulInt64(100).String(),
			types.ErrTradeTooLarge,
		)
	}

	return nil
}

// ============================================================================
// 7. PRICE IMPACT REJECTION THRESHOLDS
// ============================================================================

// CheckPriceImpactThreshold validates price impact
func (k Keeper) CheckPriceImpactThreshold(
	ctx sdk.Context,
	priceImpact sdkmath.LegacyDec,
) error {
	params := k.GetSecurityParams(ctx)

	// MaxPriceImpactPercent is already LegacyDec
	maxPriceImpact := params.MaxPriceImpactPercent

	if priceImpact.GT(maxPriceImpact) {
		return fmt.Errorf(
			"price impact %s%% exceeds threshold %s%%: %w",
			priceImpact.String(),
			maxPriceImpact.String(),
			types.ErrPriceImpactTooHigh,
		)
	}

	return nil
}

// ============================================================================
// 8. LIQUIDITY LOCK-UP PERIODS (PREVENTS RUG PULLS)
// ============================================================================

// CreateLiquidityLock locks LP tokens for specified period
func (k Keeper) CreateLiquidityLock(
	ctx sdk.Context,
	provider string,
	poolID string,
	lpTokens sdkmath.Int,
) error {
	params := k.GetSecurityParams(ctx)
	if params.LiquidityLockupSeconds == 0 {
		return nil // Lockup disabled
	}

	lockEnd := ctx.BlockTime().Add(time.Duration(params.LiquidityLockupSeconds) * time.Second)

	lock := &types.LiquidityLock{
		PoolId:         poolID,
		Provider:       provider,
		LockedLpTokens: lpTokens,
		LockStart:      ctx.BlockTime(),
		LockEnd:        lockEnd,
		IsActive:       true,
	}

	if err := k.SetLiquidityLock(ctx, lock); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLiquidityLocked,
			sdk.NewAttribute(types.AttributeKeyPoolID, poolID),
			sdk.NewAttribute(types.AttributeKeyProvider, provider),
			sdk.NewAttribute(types.AttributeKeyLPTokens, lpTokens.String()),
			sdk.NewAttribute("lock_end", lockEnd.String()),
		),
	)

	return nil
}

// CheckLiquidityLock verifies if liquidity can be removed
func (k Keeper) CheckLiquidityLock(
	ctx sdk.Context,
	provider string,
	poolID string,
	lpTokens sdkmath.Int,
) error {
	lock := k.GetLiquidityLock(ctx, provider, poolID)
	if lock == nil || !lock.IsActive {
		return nil // No active lock
	}

	lockEndTime := lock.LockEnd
	if ctx.BlockTime().Before(lockEndTime) {
		return fmt.Errorf(
			"liquidity locked until %s (current: %s): %w",
			lockEndTime.String(),
			ctx.BlockTime().String(),
			types.ErrLiquidityLocked,
		)
	}

	// Lock expired, mark as inactive
	lock.IsActive = false
	if err := k.SetLiquidityLock(ctx, lock); err != nil {
		return err
	}

	return nil
}

// SetLiquidityLock stores a liquidity lock
func (k Keeper) SetLiquidityLock(ctx sdk.Context, lock *types.LiquidityLock) error {
	store := ctx.KVStore(k.storeKey)
	key := types.LiquidityLockKey(lock.Provider, lock.PoolId)

	bz, err := k.cdc.Marshal(lock)
	if err != nil {
		return types.ErrMarshalFailed.Wrapf("failed to marshal liquidity lock for provider %s pool %s: %v", lock.Provider, lock.PoolId, err)
	}
	store.Set(key, bz)
	return nil
}

// GetLiquidityLock retrieves a liquidity lock
func (k Keeper) GetLiquidityLock(ctx sdk.Context, provider string, poolID string) *types.LiquidityLock {
	store := ctx.KVStore(k.storeKey)
	key := types.LiquidityLockKey(provider, poolID)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var lock types.LiquidityLock
	if err := k.cdc.Unmarshal(bz, &lock); err != nil {
		ctx.Logger().Error("failed to unmarshal liquidity lock",
			"provider", provider,
			"pool_id", poolID,
			"error", err,
			"data_len", len(bz))
		return nil
	}
	return &lock
}

// ============================================================================
// 9. ORDER BOOK MANIPULATION DETECTION
// ============================================================================

type orderSample struct {
	amount    sdkmath.Int
	price     sdkmath.LegacyDec
	timestamp time.Time
	order     *types.SwapOrder
}

// DetectOrderManipulation checks for layering, spoofing, and other manipulation
func (k Keeper) DetectOrderManipulation(
	ctx sdk.Context,
	address string,
	poolID string,
	orderSize sdkmath.Int,
) error {
	params := k.GetSecurityParams(ctx)

	// Get recent orders from this address
	recentOrders := k.GetRecentOrders(ctx, address, poolID, 10)

	// Require a richer history before enforcing variance-based rejection.
	if len(recentOrders) < 5 {
		return nil // Insufficient data for reliable detection
	}

	// MaxOrderVariance is already LegacyDec
	maxOrderVariance := params.MaxOrderVariance

	// Check for order size variance (spoofing indicator)
	avgSize := k.CalculateAverageOrderSize(recentOrders)
	variance := orderSize.ToLegacyDec().Sub(avgSize).Quo(avgSize).Abs()

	if variance.GT(maxOrderVariance) {
		// Flag for manipulation
		k.FlagOrderManipulation(ctx, address, poolID, "high_variance")

		return fmt.Errorf(
			"order size variance %s%% exceeds threshold %s%%: %w",
			variance.MulInt64(100).String(),
			maxOrderVariance.MulInt64(100).String(),
			types.ErrOrderManipulation,
		)
	}

	return nil
}

// FlagOrderManipulation records manipulation detection
func (k Keeper) FlagOrderManipulation(ctx sdk.Context, address string, poolID string, reason string) {
	store := ctx.KVStore(k.storeKey)

	detection := &types.OrderManipulationDetection{
		Address:       address,
		PoolId:        poolID,
		DetectedAt:    ctx.BlockTime(),
		IsFlagged:     true,
		RapidChanges:  k.CountRapidChanges(ctx, address, poolID),
		LayeringCount: k.DetectLayering(ctx, address, poolID),
		SpoofingCount: k.DetectSpoofing(ctx, address, poolID),
	}

	key := types.OrderManipulationKey(address, poolID)
	bz, err := k.cdc.Marshal(detection)
	if err != nil {
		ctx.Logger().Error("failed to marshal order manipulation detection",
			"address", address,
			"pool_id", poolID,
			"error", err)
		return
	}
	store.Set(key, bz)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeManipulationDetected,
			sdk.NewAttribute("address", address),
			sdk.NewAttribute("pool_id", poolID),
			sdk.NewAttribute("reason", reason),
		),
	)
}

// GetRecentOrders returns the most recent order sizes for an address/pool.
func (k Keeper) GetRecentOrders(ctx sdk.Context, address string, poolID string, limit int) []sdkmath.Int {
	if limit <= 0 {
		return nil
	}

	samples := k.getOrderSamples(ctx, address, poolID, false)
	if len(samples) == 0 {
		return nil
	}

	if len(samples) > limit {
		samples = samples[:limit]
	}

	result := make([]sdkmath.Int, 0, len(samples))
	for _, sample := range samples {
		result = append(result, sample.amount)
	}
	return result
}

// CountRapidChanges counts order submissions occurring faster than the allowed interval.
func (k Keeper) CountRapidChanges(ctx sdk.Context, address string, poolID string) uint64 {
	samples := k.getOrderSamples(ctx, address, poolID, false)
	if len(samples) < 2 {
		return 0
	}

	params := k.GetSecurityParams(ctx)
	interval := time.Duration(params.WashTradeMinInterval) * time.Second
	if interval <= 0 {
		interval = time.Minute
	}

	var rapid uint64
	for i := 1; i < len(samples); i++ {
		if samples[i-1].timestamp.Sub(samples[i].timestamp) < interval {
			rapid++
		}
	}
	return rapid
}

// DetectLayering counts monotonic order ladders that resemble layering.
func (k Keeper) DetectLayering(ctx sdk.Context, address string, poolID string) uint64 {
	samples := k.getOrderSamples(ctx, address, poolID, true)
	if len(samples) < 3 {
		return 0
	}

	params := k.GetSecurityParams(ctx)
	window := time.Duration(params.WashTradeMinInterval) * time.Second
	if window <= 0 {
		window = 2 * time.Minute
	}

	var buys, sells []orderSample
	for _, sample := range samples {
		if sample.order.OrderType == types.SwapOrderType_BUY {
			buys = append(buys, sample)
		} else {
			sells = append(sells, sample)
		}
	}

	return layeringSequences(buys, false, window) + layeringSequences(sells, true, window)
}

// DetectSpoofing counts the number of suspiciously small pending orders.
func (k Keeper) DetectSpoofing(ctx sdk.Context, address string, poolID string) uint64 {
	samples := k.getOrderSamples(ctx, address, poolID, true)
	if len(samples) == 0 {
		return 0
	}

	params := k.GetSecurityParams(ctx)
	// MinTradeAmount is already Int
	minTrade := params.MinTradeAmount
	threshold := minTrade.MulRaw(2)

	var count uint64
	for _, sample := range samples {
		if sample.amount.LT(threshold) {
			count++
		}
	}
	return count
}

func layeringSequences(samples []orderSample, sell bool, window time.Duration) uint64 {
	if len(samples) < 3 {
		return 0
	}

	sort.Slice(samples, func(i, j int) bool {
		return samples[i].timestamp.Before(samples[j].timestamp)
	})

	var sequences uint64
	run := 1
	for i := 1; i < len(samples); i++ {
		timeGap := samples[i].timestamp.Sub(samples[i-1].timestamp)
		if timeGap > window {
			run = 1
			continue
		}

		if sell {
			if samples[i].price.GTE(samples[i-1].price) {
				run++
			} else {
				run = 1
			}
		} else {
			if samples[i].price.LTE(samples[i-1].price) {
				run++
			} else {
				run = 1
			}
		}

		if run >= 3 {
			sequences++
			run = 1
		}
	}
	return sequences
}

func (k Keeper) getOrderSamples(ctx sdk.Context, address string, poolID string, pendingOnly bool) []orderSample {
	orders := k.GetOrdersByUser(ctx, address)
	if len(orders) == 0 {
		return nil
	}

	normalizedPool := normalizePair(poolID)
	samples := make([]orderSample, 0, len(orders))
	for _, order := range orders {
		if pendingOnly && order.Status != types.SwapOrderStatus_PENDING {
			continue
		}
		if normalizedPool != "" && !matchesPool(order, normalizedPool) {
			continue
		}
		// AuraAmount is already Int
		amt := order.AuraAmount
		ts := order.Timestamp
		price := orderPriceDec(order)

		samples = append(samples, orderSample{
			amount:    amt,
			price:     price,
			timestamp: ts,
			order:     order,
		})
	}

	sort.Slice(samples, func(i, j int) bool {
		return samples[i].timestamp.After(samples[j].timestamp)
	})
	return samples
}

func normalizePair(pair string) string {
	if pair == "" {
		return ""
	}
	pair = strings.ReplaceAll(pair, "/", "-")
	parts := strings.Split(pair, "-")
	if len(parts) != 2 {
		return pair
	}
	a := strings.TrimSpace(parts[0])
	b := strings.TrimSpace(parts[1])
	if a > b {
		a, b = b, a
	}
	return fmt.Sprintf("%s-%s", a, b)
}

func matchesPool(order *types.SwapOrder, normalizedPool string) bool {
	if normalizedPool == "" {
		return true
	}
	pair := normalizePair(fmt.Sprintf("%s-%s", "uaura", order.OtherCoin))
	return pair == normalizedPool
}

func orderPriceDec(order *types.SwapOrder) sdkmath.LegacyDec {
	// PricePerAura is already LegacyDec
	if !order.PricePerAura.IsZero() {
		return order.PricePerAura
	}
	// AuraAmount and OtherAmount are already Int
	auraAmt := order.AuraAmount
	otherAmt := order.OtherAmount
	if auraAmt.IsZero() {
		return sdkmath.LegacyZeroDec()
	}
	return otherAmt.ToLegacyDec().Quo(auraAmt.ToLegacyDec())
}

// ============================================================================
// 10. WASH TRADING DETECTION
// ============================================================================

// DetectWashTrading identifies potential wash trading patterns
func (k Keeper) DetectWashTrading(
	ctx sdk.Context,
	address string,
	poolID string,
) error {
	params := k.GetSecurityParams(ctx)
	history := k.GetTradeHistory(ctx, address)

	if history == nil {
		return nil
	}

	// Check if trades are too frequent (wash trading indicator)
	timeSinceLastTrade := ctx.BlockTime().Unix() - history.LastTradeTime.Unix()

	if timeSinceLastTrade < params.WashTradeMinInterval {
		// Increment suspicious trade counter
		k.IncrementWashTradeDetection(ctx, address, poolID)

		detection := k.GetWashTradeDetection(ctx, address, poolID)
		if detection != nil && detection.SuspiciousTrades >= 5 {
			return fmt.Errorf(
				"wash trading detected: %d suspicious trades: %w",
				detection.SuspiciousTrades,
				types.ErrWashTradingDetected,
			)
		}
	}

	return nil
}

// IncrementWashTradeDetection increments wash trade counter
func (k Keeper) IncrementWashTradeDetection(ctx sdk.Context, address string, poolID string) {
	store := ctx.KVStore(k.storeKey)
	key := types.WashTradeKey(address, poolID)

	var detection types.WashTradeDetection
	bz := store.Get(key)

	if bz == nil {
		detection = types.WashTradeDetection{
			Address:          address,
			PoolId:           poolID,
			SuspiciousTrades: 1,
			FirstDetection:   ctx.BlockTime(),
			LastDetection:    ctx.BlockTime(),
			IsFlagged:        false,
			ConfidenceScore:  10,
		}
	} else {
		if err := k.cdc.Unmarshal(bz, &detection); err != nil {
			ctx.Logger().Error("failed to unmarshal wash trade detection, resetting",
				"address", address,
				"pool_id", poolID,
				"error", err)
			// Reset to new detection on corruption
			detection = types.WashTradeDetection{
				Address:          address,
				PoolId:           poolID,
				SuspiciousTrades: 1,
				FirstDetection:   ctx.BlockTime(),
				LastDetection:    ctx.BlockTime(),
				IsFlagged:        false,
				ConfidenceScore:  10,
			}
		} else {
			detection.SuspiciousTrades++
			detection.LastDetection = ctx.BlockTime()
			detection.ConfidenceScore = min(100, detection.ConfidenceScore+10)

			if detection.SuspiciousTrades >= 5 {
				detection.IsFlagged = true
			}
		}
	}

	bz, err := k.cdc.Marshal(&detection)
	if err != nil {
		ctx.Logger().Error("failed to marshal wash trade detection",
			"address", address,
			"pool_id", poolID,
			"error", err)
		return
	}
	store.Set(key, bz)
}

// GetWashTradeDetection retrieves wash trade detection record
func (k Keeper) GetWashTradeDetection(ctx sdk.Context, address string, poolID string) *types.WashTradeDetection {
	store := ctx.KVStore(k.storeKey)
	key := types.WashTradeKey(address, poolID)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var detection types.WashTradeDetection
	if err := k.cdc.Unmarshal(bz, &detection); err != nil {
		ctx.Logger().Error("failed to unmarshal wash trade detection",
			"address", address,
			"pool_id", poolID,
			"error", err,
			"data_len", len(bz))
		return nil
	}
	return &detection
}

// ============================================================================
// 11. DUST ATTACK PREVENTION
// ============================================================================

// CheckDustAttack prevents dust attacks with minimum trade amounts
func (k Keeper) CheckDustAttack(ctx sdk.Context, amountIn sdkmath.Int) error {
	params := k.GetSecurityParams(ctx)

	// MinTradeAmount is already Int
	minTradeAmount := params.MinTradeAmount

	if amountIn.LT(minTradeAmount) {
		return fmt.Errorf(
			"trade amount %s below minimum %s: %w",
			amountIn.String(),
			minTradeAmount.String(),
			types.ErrDustAttack,
		)
	}

	return nil
}

// ============================================================================
// 12. POOL CREATION LIMITS AND VALIDATION
// ============================================================================

// CheckPoolCreationLimits validates pool creation constraints
func (k Keeper) CheckPoolCreationLimits(
	ctx sdk.Context,
	creator string,
	initialLiquidity sdkmath.Int,
) error {
	params := k.GetSecurityParams(ctx)

	// MinPoolCreationLiquidity is already Int
	minPoolCreationLiquidity := params.MinPoolCreationLiquidity

	// Check minimum liquidity
	if initialLiquidity.LT(minPoolCreationLiquidity) {
		return fmt.Errorf(
			"initial liquidity %s below minimum %s: %w",
			initialLiquidity.String(),
			minPoolCreationLiquidity.String(),
			types.ErrInsufficientPoolLiquidity,
		)
	}

	// Check pool creation cooldown
	record := k.GetPoolCreationRecord(ctx, creator)
	if record != nil {
		timeSinceLastPool := ctx.BlockTime().Unix() - record.LastCreationTime.Unix()
		if timeSinceLastPool < params.PoolCreationCooldown {
			return fmt.Errorf(
				"must wait %d seconds between pool creations (waited: %d): %w",
				params.PoolCreationCooldown,
				timeSinceLastPool,
				types.ErrPoolCreationCooldown,
			)
		}

		// Check maximum pools per creator
		if record.TotalPools >= params.MaxPoolsPerCreator {
			return fmt.Errorf(
				"maximum pools per creator exceeded: %d >= %d: %w",
				record.TotalPools,
				params.MaxPoolsPerCreator,
				types.ErrMaxPoolsExceeded,
			)
		}
	}

	return nil
}

// ============================================================================
// CIRCUIT BREAKER
// ============================================================================

// ActivateCircuitBreaker pauses trading
func (k Keeper) ActivateCircuitBreaker(
	ctx sdk.Context,
	pausedBy string,
	reason string,
	affectedPools []string,
) {
	breaker := &types.CircuitBreaker{
		IsPaused:      true,
		PauseReason:   reason,
		PausedAt:      ctx.BlockTime(),
		PausedBy:      pausedBy,
		AffectedPools: affectedPools,
	}

	_ = k.SetCircuitBreaker(ctx, breaker) // Best effort, log if needed

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCircuitBreakerActivated,
			sdk.NewAttribute("paused_by", pausedBy),
			sdk.NewAttribute("reason", reason),
		),
	)
}

// DeactivateCircuitBreaker resumes trading
func (k Keeper) DeactivateCircuitBreaker(ctx sdk.Context) {
	breaker := k.GetCircuitBreaker(ctx)
	if breaker != nil {
		breaker.IsPaused = false
		_ = k.SetCircuitBreaker(ctx, breaker) // Best effort, log if needed
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCircuitBreakerDeactivated,
		),
	)
}

// IsCircuitBreakerActive checks if trading is paused
func (k Keeper) IsCircuitBreakerActive(ctx sdk.Context, poolID string) bool {
	breaker := k.GetCircuitBreaker(ctx)
	if breaker == nil || !breaker.IsPaused {
		return false
	}

	// If no specific pools, all trading is paused
	if len(breaker.AffectedPools) == 0 {
		return true
	}

	// Check if specific pool is affected
	for _, p := range breaker.AffectedPools {
		if p == poolID {
			return true
		}
	}

	return false
}

// SetCircuitBreaker stores circuit breaker state
func (k Keeper) SetCircuitBreaker(ctx sdk.Context, breaker *types.CircuitBreaker) error {
	store := ctx.KVStore(k.storeKey)
	key := types.CircuitBreakerKey()

	bz, err := k.cdc.Marshal(breaker)
	if err != nil {
		return types.ErrMarshalFailed.Wrapf("failed to marshal circuit breaker: %v", err)
	}
	store.Set(key, bz)
	return nil
}

// GetCircuitBreaker retrieves circuit breaker state
func (k Keeper) GetCircuitBreaker(ctx sdk.Context) *types.CircuitBreaker {
	store := ctx.KVStore(k.storeKey)
	key := types.CircuitBreakerKey()

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var breaker types.CircuitBreaker
	if err := k.cdc.Unmarshal(bz, &breaker); err != nil {
		ctx.Logger().Error("failed to unmarshal circuit breaker",
			"error", err,
			"data_len", len(bz))
		return nil
	}
	return &breaker
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// GetSecurityParams returns security parameters
func (k Keeper) GetSecurityParams(ctx sdk.Context) *types.SecurityParams {
	// In production, load from param store
	// For now, return default values
	return types.DefaultSecurityParams()
}

// UpdateTradeHistory updates trade history for address
func (k Keeper) UpdateTradeHistory(ctx sdk.Context, address string, poolID string) {
	store := ctx.KVStore(k.storeKey)
	key := types.TradeHistoryKey(address)

	var history types.TradeHistory
	bz := store.Get(key)

	currentBlock := ctx.BlockHeight()

	if bz == nil {
		history = types.TradeHistory{
			Address:          address,
			PoolId:           poolID,
			LastTradeTime:    ctx.BlockTime(),
			LastTradeBlock:   currentBlock,
			TradesInBlock:    1,
			RecentTradeCount: 1,
			RecentVolume:     sdkmath.ZeroInt(),
		}
	} else {
		if err := k.cdc.Unmarshal(bz, &history); err != nil {
			ctx.Logger().Error("failed to unmarshal trade history, resetting",
				"address", address,
				"error", err)
			// Reset to new history on corruption
			history = types.TradeHistory{
				Address:          address,
				PoolId:           poolID,
				LastTradeTime:    ctx.BlockTime(),
				LastTradeBlock:   currentBlock,
				TradesInBlock:    1,
				RecentTradeCount: 1,
				RecentVolume:     sdkmath.ZeroInt(),
			}
		} else {
			if history.LastTradeBlock == currentBlock {
				history.TradesInBlock++
			} else {
				history.TradesInBlock = 1
			}

			history.LastTradeTime = ctx.BlockTime()
			history.LastTradeBlock = currentBlock
			history.RecentTradeCount++
		}
	}

	bz, err := k.cdc.Marshal(&history)
	if err != nil {
		ctx.Logger().Error("failed to marshal trade history",
			"address", address,
			"pool_id", poolID,
			"error", err)
		return
	}
	store.Set(key, bz)
}

// GetTradeHistory retrieves trade history
func (k Keeper) GetTradeHistory(ctx sdk.Context, address string) *types.TradeHistory {
	store := ctx.KVStore(k.storeKey)
	key := types.TradeHistoryKey(address)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var history types.TradeHistory
	if err := k.cdc.Unmarshal(bz, &history); err != nil {
		ctx.Logger().Error("failed to unmarshal trade history",
			"address", address,
			"error", err,
			"data_len", len(bz))
		return nil
	}
	return &history
}

// CalculateAverageOrderSize calculates average order size
func (k Keeper) CalculateAverageOrderSize(orders []sdkmath.Int) sdkmath.LegacyDec {
	if len(orders) == 0 {
		return sdkmath.LegacyZeroDec()
	}

	total := sdkmath.ZeroInt()
	for _, order := range orders {
		total = total.Add(order)
	}

	return total.ToLegacyDec().QuoInt64(int64(len(orders)))
}

// GenerateSecureHash generates a secure hash for HTLC
func (k Keeper) GenerateSecureHash(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

// ============================================================================
// 8. POOL CREATION AUDIT TRAIL
// ============================================================================

// RecordPoolCreation records pool creation for audit trail and compliance.
// This creates a permanent record of who created which pools and when,
// enabling:
// - Regulatory compliance and audit trails
// - Pool creation limit enforcement
// - Creation cooldown period checks
// - Pool history reconstruction
//
// SECURITY: This function should be called immediately after successful pool creation
// to ensure all pools have proper audit records.
func (k Keeper) RecordPoolCreation(ctx sdk.Context, creator string, poolID string, tokenA string, tokenB string, amountA sdkmath.Int, amountB sdkmath.Int) {
	store := ctx.KVStore(k.storeKey)
	key := types.PoolCreationKey(creator)

	var record types.PoolCreationRecord
	bz := store.Get(key)

	if bz == nil {
		// First pool created by this address
		record = types.PoolCreationRecord{
			Creator:          creator,
			PoolIds:          []string{poolID},
			LastCreationTime: ctx.BlockTime(),
			TotalPools:       1,
		}
	} else {
		// Existing creator - append new pool
		if err := k.cdc.Unmarshal(bz, &record); err != nil {
			ctx.Logger().Error("failed to unmarshal pool creation record, resetting",
				"creator", creator,
				"error", err)
			// Reset to new record on corruption
			record = types.PoolCreationRecord{
				Creator:          creator,
				PoolIds:          []string{poolID},
				LastCreationTime: ctx.BlockTime(),
				TotalPools:       1,
			}
		} else {
			record.PoolIds = append(record.PoolIds, poolID)
			record.LastCreationTime = ctx.BlockTime()
			record.TotalPools++
		}
	}

	// Store updated record
	bz, err := k.cdc.Marshal(&record)
	if err != nil {
		ctx.Logger().Error("failed to marshal pool creation record",
			"creator", creator,
			"pool_id", poolID,
			"error", err)
		return
	}
	store.Set(key, bz)

	// Emit detailed audit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"pool_creation_recorded",
			sdk.NewAttribute("creator", creator),
			sdk.NewAttribute("pool_id", poolID),
			sdk.NewAttribute("token_a", tokenA),
			sdk.NewAttribute("token_b", tokenB),
			sdk.NewAttribute("initial_liquidity_a", amountA.String()),
			sdk.NewAttribute("initial_liquidity_b", amountB.String()),
			sdk.NewAttribute("total_pools_created", fmt.Sprintf("%d", record.TotalPools)),
			sdk.NewAttribute("timestamp", ctx.BlockTime().String()),
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)
}

// GetPoolCreationRecord retrieves the pool creation record for a creator.
// Returns nil if the creator has never created any pools.
func (k Keeper) GetPoolCreationRecord(ctx sdk.Context, creator string) *types.PoolCreationRecord {
	store := ctx.KVStore(k.storeKey)
	key := types.PoolCreationKey(creator)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var record types.PoolCreationRecord
	if err := k.cdc.Unmarshal(bz, &record); err != nil {
		ctx.Logger().Error("failed to unmarshal pool creation record",
			"creator", creator,
			"error", err,
			"data_len", len(bz))
		return nil
	}
	return &record
}

// GetAllPoolCreationRecords retrieves all pool creation records for genesis export.
// This enables full reconstruction of pool creation history from genesis state.
func (k Keeper) GetAllPoolCreationRecords(ctx sdk.Context) []*types.PoolCreationRecord {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.PoolCreationPrefix)
	defer iterator.Close()

	var records []*types.PoolCreationRecord
	for ; iterator.Valid(); iterator.Next() {
		var record types.PoolCreationRecord
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			ctx.Logger().Error("failed to unmarshal pool creation record, skipping",
				"error", err,
				"data_len", len(iterator.Value()))
			continue
		}
		records = append(records, &record)
	}

	return records
}

// CheckPoolCreationLimit validates if creator can create another pool.
// Enforces max_pools_per_creator parameter to prevent spam pool creation.
//
// SECURITY: Call this BEFORE creating a pool to reject requests from
// addresses that have exceeded their pool creation quota.
func (k Keeper) CheckPoolCreationLimit(ctx sdk.Context, creator string) error {
	params := k.GetSecurityParams(ctx)
	if params.MaxPoolsPerCreator == 0 {
		return nil // No limit enforced
	}

	record := k.GetPoolCreationRecord(ctx, creator)
	if record == nil {
		return nil // First pool, allowed
	}

	if record.TotalPools >= params.MaxPoolsPerCreator {
		return fmt.Errorf(
			"pool creation limit exceeded: creator %s has %d pools, maximum allowed is %d: %w",
			creator,
			record.TotalPools,
			params.MaxPoolsPerCreator,
			types.ErrPoolCreationLimitExceeded,
		)
	}

	return nil
}

// CheckPoolCreationCooldown validates if creator can create a pool now.
// Enforces pool_creation_cooldown parameter to prevent rapid pool spam.
//
// SECURITY: Call this BEFORE creating a pool to reject requests from
// addresses that are creating pools too rapidly.
func (k Keeper) CheckPoolCreationCooldown(ctx sdk.Context, creator string) error {
	params := k.GetSecurityParams(ctx)
	if params.PoolCreationCooldown == 0 {
		return nil // No cooldown enforced
	}

	record := k.GetPoolCreationRecord(ctx, creator)
	if record == nil {
		return nil // First pool, no cooldown applies
	}

	// Calculate time since last pool creation
	lastCreationTime := record.LastCreationTime
	currentTime := ctx.BlockTime()
	timeSinceLastCreation := currentTime.Sub(lastCreationTime)

	cooldownDuration := time.Duration(params.PoolCreationCooldown) * time.Second
	if timeSinceLastCreation < cooldownDuration {
		return fmt.Errorf(
			"pool creation cooldown active: must wait %s between pool creations, last creation was %s ago: %w",
			cooldownDuration.String(),
			timeSinceLastCreation.String(),
			types.ErrPoolCreationCooldown,
		)
	}

	return nil
}
