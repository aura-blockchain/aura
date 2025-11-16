package keeper

import (
	"crypto/sha256"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

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

	if currentBlock-lastTrade < params.MinBlockDelay {
		return sdkerrors.Wrapf(
			types.ErrFrontRunningDetected,
			"must wait %d blocks between trades (last: %d, current: %d)",
			params.MinBlockDelay,
			lastTrade,
			currentBlock,
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

	// Calculate spot price
	spotPrice := pool.ReserveB.ToDec().Quo(pool.ReserveA.ToDec())

	// Get previous observation to calculate time delta
	prevObservation := k.GetLatestTWAPObservation(ctx, poolID)

	var cumulativePrice sdk.Dec
	if prevObservation != nil {
		// Time elapsed since last observation
		timeElapsed := ctx.BlockTime().Unix() - prevObservation.Timestamp.Unix()

		// Cumulative price = previous cumulative + (spot price * time elapsed)
		cumulativePrice = prevObservation.CumulativePrice.Add(
			spotPrice.MulInt64(timeElapsed),
		)
	} else {
		// First observation
		cumulativePrice = sdk.ZeroDec()
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

	k.SetTWAPObservation(ctx, observation)

	// Prune old observations outside TWAP window
	k.PruneTWAPObservations(ctx, poolID)

	return nil
}

// GetTWAPPrice calculates TWAP over the specified window
func (k Keeper) GetTWAPPrice(ctx sdk.Context, poolID string, windowBlocks uint64) (sdk.Dec, error) {
	params := k.GetSecurityParams(ctx)
	if windowBlocks == 0 {
		windowBlocks = params.TwapWindowBlocks
	}

	observations := k.GetTWAPObservations(ctx, poolID, windowBlocks)
	if len(observations) < 2 {
		// Need at least 2 observations for TWAP
		pool := k.GetPool(ctx, poolID)
		if pool == nil {
			return sdk.ZeroDec(), types.ErrPoolNotFound
		}
		// Return spot price if insufficient data
		return pool.ReserveB.ToDec().Quo(pool.ReserveA.ToDec()), nil
	}

	// TWAP = (cumulative_price_end - cumulative_price_start) / time_elapsed
	latest := observations[0]
	oldest := observations[len(observations)-1]

	timeElapsed := latest.Timestamp.Unix() - oldest.Timestamp.Unix()
	if timeElapsed == 0 {
		return latest.SpotPrice, nil
	}

	twap := latest.CumulativePrice.Sub(oldest.CumulativePrice).QuoInt64(timeElapsed)

	return twap, nil
}

// SetTWAPObservation stores a TWAP observation
func (k Keeper) SetTWAPObservation(ctx sdk.Context, obs *types.TWAPPrice) {
	store := ctx.KVStore(k.storeKey)
	key := types.TWAPKey(obs.PoolId, obs.BlockHeight)

	bz := k.cdc.MustMarshal(obs)
	store.Set(key, bz)
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
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var observations []*types.TWAPPrice
	for ; iterator.Valid(); iterator.Next() {
		var obs types.TWAPPrice
		k.cdc.MustUnmarshal(iterator.Value(), &obs)

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
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var keysToDelete [][]byte
	for ; iterator.Valid(); iterator.Next() {
		var obs types.TWAPPrice
		k.cdc.MustUnmarshal(iterator.Value(), &obs)

		if uint64(obs.BlockHeight) < cutoffBlock {
			keysToDelete = append(keysToDelete, iterator.Key())
		}
	}

	for _, key := range keysToDelete {
		store.Delete(key)
	}
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
		return sdkerrors.Wrapf(
			types.ErrFlashLoanDetected,
			"must wait %d blocks between liquidity operations (last: %d, current: %d)",
			params.MinLiquidityBlocks,
			lastBlock,
			currentBlock,
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

// CheckMEVProtection implements MEV mitigation
func (k Keeper) CheckMEVProtection(ctx sdk.Context, address string) error {
	params := k.GetSecurityParams(ctx)
	if !params.MevProtectionEnabled {
		return nil
	}

	// Check maximum swaps per block
	swapsInBlock := k.GetSwapsInCurrentBlock(ctx, address)
	if swapsInBlock >= params.MaxSwapsPerBlock {
		return sdkerrors.Wrapf(
			types.ErrMEVDetected,
			"maximum swaps per block exceeded: %d >= %d",
			swapsInBlock,
			params.MaxSwapsPerBlock,
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
	priceImpact sdk.Dec,
) error {
	params := k.GetSecurityParams(ctx)

	// Convert price impact percentage to decimal (5% = 5.0)
	if priceImpact.GT(params.MaxPriceImpactPercent) {
		return sdkerrors.Wrapf(
			types.ErrPriceImpactTooHigh,
			"price impact %s%% exceeds maximum %s%%",
			priceImpact.String(),
			params.MaxPriceImpactPercent.String(),
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
	amountIn sdk.Int,
) error {
	params := k.GetSecurityParams(ctx)
	pool := k.GetPool(ctx, poolID)
	if pool == nil {
		return types.ErrPoolNotFound
	}

	// Determine reserve based on input denom
	// Assuming input is denomA, check against reserveA
	maxAmount := pool.ReserveA.ToDec().Mul(params.MaxTradeSizePercent).TruncateInt()

	if amountIn.GT(maxAmount) {
		return sdkerrors.Wrapf(
			types.ErrTradeTooLarge,
			"trade size %s exceeds maximum %s (%s%% of pool)",
			amountIn.String(),
			maxAmount.String(),
			params.MaxTradeSizePercent.MulInt64(100).String(),
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
	priceImpact sdk.Dec,
) error {
	params := k.GetSecurityParams(ctx)

	if priceImpact.GT(params.MaxPriceImpactPercent) {
		return sdkerrors.Wrapf(
			types.ErrPriceImpactTooHigh,
			"price impact %s%% exceeds threshold %s%%",
			priceImpact.String(),
			params.MaxPriceImpactPercent.String(),
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
	lpTokens sdk.Int,
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

	k.SetLiquidityLock(ctx, lock)

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
	lpTokens sdk.Int,
) error {
	lock := k.GetLiquidityLock(ctx, provider, poolID)
	if lock == nil || !lock.IsActive {
		return nil // No active lock
	}

	if ctx.BlockTime().Before(lock.LockEnd) {
		return sdkerrors.Wrapf(
			types.ErrLiquidityLocked,
			"liquidity locked until %s (current: %s)",
			lock.LockEnd.String(),
			ctx.BlockTime().String(),
		)
	}

	// Lock expired, mark as inactive
	lock.IsActive = false
	k.SetLiquidityLock(ctx, lock)

	return nil
}

// SetLiquidityLock stores a liquidity lock
func (k Keeper) SetLiquidityLock(ctx sdk.Context, lock *types.LiquidityLock) {
	store := ctx.KVStore(k.storeKey)
	key := types.LiquidityLockKey(lock.Provider, lock.PoolId)

	bz := k.cdc.MustMarshal(lock)
	store.Set(key, bz)
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
	k.cdc.MustUnmarshal(bz, &lock)
	return &lock
}

// ============================================================================
// 9. ORDER BOOK MANIPULATION DETECTION
// ============================================================================

// DetectOrderManipulation checks for layering, spoofing, and other manipulation
func (k Keeper) DetectOrderManipulation(
	ctx sdk.Context,
	address string,
	poolID string,
	orderSize sdk.Int,
) error {
	params := k.GetSecurityParams(ctx)

	// Get recent orders from this address
	recentOrders := k.GetRecentOrders(ctx, address, poolID, 10)

	if len(recentOrders) < 2 {
		return nil // Insufficient data
	}

	// Check for order size variance (spoofing indicator)
	avgSize := k.CalculateAverageOrderSize(recentOrders)
	variance := orderSize.ToDec().Sub(avgSize).Quo(avgSize).Abs()

	if variance.GT(params.MaxOrderVariance) {
		// Flag for manipulation
		k.FlagOrderManipulation(ctx, address, poolID, "high_variance")

		return sdkerrors.Wrapf(
			types.ErrOrderManipulation,
			"order size variance %s%% exceeds threshold %s%%",
			variance.MulInt64(100).String(),
			params.MaxOrderVariance.MulInt64(100).String(),
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
	bz := k.cdc.MustMarshal(detection)
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
			return sdkerrors.Wrapf(
				types.ErrWashTradingDetected,
				"wash trading detected: %d suspicious trades",
				detection.SuspiciousTrades,
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
		k.cdc.MustUnmarshal(bz, &detection)
		detection.SuspiciousTrades++
		detection.LastDetection = ctx.BlockTime()
		detection.ConfidenceScore = min(100, detection.ConfidenceScore+10)

		if detection.SuspiciousTrades >= 5 {
			detection.IsFlagged = true
		}
	}

	store.Set(key, k.cdc.MustMarshal(&detection))
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
	k.cdc.MustUnmarshal(bz, &detection)
	return &detection
}

// ============================================================================
// 11. DUST ATTACK PREVENTION
// ============================================================================

// CheckDustAttack prevents dust attacks with minimum trade amounts
func (k Keeper) CheckDustAttack(ctx sdk.Context, amountIn sdk.Int) error {
	params := k.GetSecurityParams(ctx)

	if amountIn.LT(params.MinTradeAmount) {
		return sdkerrors.Wrapf(
			types.ErrDustAttack,
			"trade amount %s below minimum %s",
			amountIn.String(),
			params.MinTradeAmount.String(),
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
	initialLiquidity sdk.Int,
) error {
	params := k.GetSecurityParams(ctx)

	// Check minimum liquidity
	if initialLiquidity.LT(params.MinPoolCreationLiquidity) {
		return sdkerrors.Wrapf(
			types.ErrInsufficientPoolLiquidity,
			"initial liquidity %s below minimum %s",
			initialLiquidity.String(),
			params.MinPoolCreationLiquidity.String(),
		)
	}

	// Check pool creation cooldown
	record := k.GetPoolCreationRecord(ctx, creator)
	if record != nil {
		timeSinceLastPool := ctx.BlockTime().Unix() - record.LastCreationTime.Unix()
		if timeSinceLastPool < params.PoolCreationCooldown {
			return sdkerrors.Wrapf(
				types.ErrPoolCreationCooldown,
				"must wait %d seconds between pool creations (waited: %d)",
				params.PoolCreationCooldown,
				timeSinceLastPool,
			)
		}

		// Check maximum pools per creator
		if record.TotalPools >= params.MaxPoolsPerCreator {
			return sdkerrors.Wrapf(
				types.ErrMaxPoolsExceeded,
				"maximum pools per creator exceeded: %d >= %d",
				record.TotalPools,
				params.MaxPoolsPerCreator,
			)
		}
	}

	return nil
}

// RecordPoolCreation records pool creation by address
func (k Keeper) RecordPoolCreation(ctx sdk.Context, creator string, poolID string) {
	store := ctx.KVStore(k.storeKey)
	key := types.PoolCreationKey(creator)

	var record types.PoolCreationRecord
	bz := store.Get(key)

	if bz == nil {
		record = types.PoolCreationRecord{
			Creator:          creator,
			PoolIds:          []string{poolID},
			LastCreationTime: ctx.BlockTime(),
			TotalPools:       1,
		}
	} else {
		k.cdc.MustUnmarshal(bz, &record)
		record.PoolIds = append(record.PoolIds, poolID)
		record.LastCreationTime = ctx.BlockTime()
		record.TotalPools++
	}

	store.Set(key, k.cdc.MustMarshal(&record))
}

// GetPoolCreationRecord retrieves pool creation record
func (k Keeper) GetPoolCreationRecord(ctx sdk.Context, creator string) *types.PoolCreationRecord {
	store := ctx.KVStore(k.storeKey)
	key := types.PoolCreationKey(creator)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var record types.PoolCreationRecord
	k.cdc.MustUnmarshal(bz, &record)
	return &record
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

	k.SetCircuitBreaker(ctx, breaker)

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
		k.SetCircuitBreaker(ctx, breaker)
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
func (k Keeper) SetCircuitBreaker(ctx sdk.Context, breaker *types.CircuitBreaker) {
	store := ctx.KVStore(k.storeKey)
	key := types.CircuitBreakerKey()

	bz := k.cdc.MustMarshal(breaker)
	store.Set(key, bz)
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
	k.cdc.MustUnmarshal(bz, &breaker)
	return &breaker
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// GetSecurityParams returns security parameters
func (k Keeper) GetSecurityParams(ctx sdk.Context) types.SecurityParams {
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
			RecentVolume:     sdk.ZeroInt(),
		}
	} else {
		k.cdc.MustUnmarshal(bz, &history)

		if history.LastTradeBlock == currentBlock {
			history.TradesInBlock++
		} else {
			history.TradesInBlock = 1
		}

		history.LastTradeTime = ctx.BlockTime()
		history.LastTradeBlock = currentBlock
		history.RecentTradeCount++
	}

	store.Set(key, k.cdc.MustMarshal(&history))
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
	k.cdc.MustUnmarshal(bz, &history)
	return &history
}

// GetRecentOrders retrieves recent orders (stub for orderbook)
func (k Keeper) GetRecentOrders(ctx sdk.Context, address string, poolID string, limit int) []sdk.Int {
	// TODO: Implement orderbook tracking
	return []sdk.Int{}
}

// CalculateAverageOrderSize calculates average order size
func (k Keeper) CalculateAverageOrderSize(orders []sdk.Int) sdk.Dec {
	if len(orders) == 0 {
		return sdk.ZeroDec()
	}

	total := sdk.ZeroInt()
	for _, order := range orders {
		total = total.Add(order)
	}

	return total.ToDec().QuoInt64(int64(len(orders)))
}

// CountRapidChanges counts rapid order changes
func (k Keeper) CountRapidChanges(ctx sdk.Context, address string, poolID string) uint64 {
	// TODO: Implement order change tracking
	return 0
}

// DetectLayering detects layering manipulation
func (k Keeper) DetectLayering(ctx sdk.Context, address string, poolID string) uint64 {
	// TODO: Implement layering detection
	return 0
}

// DetectSpoofing detects spoofing manipulation
func (k Keeper) DetectSpoofing(ctx sdk.Context, address string, poolID string) uint64 {
	// TODO: Implement spoofing detection
	return 0
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
