package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// ============================================================================
// TIME-WEIGHTED AVERAGE PRICE (TWAP) ORACLE
// Extracted from security.go for better code organization
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

	observations := make([]*types.TWAPPrice, 0, 64)
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

	keysToDelete := make([][]byte, 0, 64)
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
	if len(cursorBytes) >= 8 {
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
