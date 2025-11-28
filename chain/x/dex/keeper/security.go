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
	"google.golang.org/protobuf/types/known/timestamppb"

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
		// Parse previous cumulative price from string
		prevCumulativePrice, err := sdkmath.LegacyNewDecFromStr(prevObservation.CumulativePrice)
		if err != nil {
			return err
		}

		// Time elapsed since last observation
		timeElapsed := ctx.BlockTime().Unix() - prevObservation.Timestamp.AsTime().Unix()

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
		CumulativePrice: cumulativePrice.String(),
		BlockHeight:     ctx.BlockHeight(),
		Timestamp:       timestamppb.New(ctx.BlockTime()),
		ReserveA:        pool.ReserveA,
		ReserveB:        pool.ReserveB,
		SpotPrice:       spotPrice.String(),
	}

	k.SetTWAPObservation(ctx, observation)

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
		reserveB, err := k.parseReserve(pool.ReserveB)
		if err != nil {
			return sdkmath.LegacyZeroDec(), err
		}
		return reserveB.ToLegacyDec().Quo(reserveA.ToLegacyDec()), nil
	}

	// TWAP = (cumulative_price_end - cumulative_price_start) / time_elapsed
	latest := observations[0]
	oldest := observations[len(observations)-1]

	timeElapsed := latest.Timestamp.AsTime().Unix() - oldest.Timestamp.AsTime().Unix()
	if timeElapsed == 0 {
		spotPrice, err := sdkmath.LegacyNewDecFromStr(latest.SpotPrice)
		if err != nil {
			return sdkmath.LegacyZeroDec(), err
		}
		return spotPrice, nil
	}

	// Parse cumulative prices from strings
	latestCumulativePrice, err := sdkmath.LegacyNewDecFromStr(latest.CumulativePrice)
	if err != nil {
		return sdkmath.LegacyZeroDec(), err
	}
	oldestCumulativePrice, err := sdkmath.LegacyNewDecFromStr(oldest.CumulativePrice)
	if err != nil {
		return sdkmath.LegacyZeroDec(), err
	}

	twap := latestCumulativePrice.Sub(oldestCumulativePrice).QuoInt64(timeElapsed)

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
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
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
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
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

	// Parse max price impact from string
	maxPriceImpact, err := sdkmath.LegacyNewDecFromStr(params.MaxPriceImpactPercent)
	if err != nil {
		return err
	}

	// Convert price impact percentage to decimal (5% = 5.0)
	if priceImpact.GT(maxPriceImpact) {
		return fmt.Errorf(
			"price impact %s%% exceeds maximum %s%%: %w",
			priceImpact.String(),
			params.MaxPriceImpactPercent,
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

	// Parse reserve and max trade size percent
	reserveA, err := k.parseReserve(pool.ReserveA)
	if err != nil {
		return err
	}
	maxTradeSizePercent, err := sdkmath.LegacyNewDecFromStr(params.MaxTradeSizePercent)
	if err != nil {
		return err
	}

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

	// Parse max price impact from string
	maxPriceImpact, err := sdkmath.LegacyNewDecFromStr(params.MaxPriceImpactPercent)
	if err != nil {
		return err
	}

	if priceImpact.GT(maxPriceImpact) {
		return fmt.Errorf(
			"price impact %s%% exceeds threshold %s%%: %w",
			priceImpact.String(),
			params.MaxPriceImpactPercent,
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
		LockedLpTokens: lpTokens.String(),
		LockStart:      timestamppb.New(ctx.BlockTime()),
		LockEnd:        timestamppb.New(lockEnd),
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
	lpTokens sdkmath.Int,
) error {
	lock := k.GetLiquidityLock(ctx, provider, poolID)
	if lock == nil || !lock.IsActive {
		return nil // No active lock
	}

	lockEndTime := lock.LockEnd.AsTime()
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

	// Parse max order variance from string
	maxOrderVariance, err := sdkmath.LegacyNewDecFromStr(params.MaxOrderVariance)
	if err != nil {
		return err
	}

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
		DetectedAt:    timestamppb.New(ctx.BlockTime()),
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
	minTrade, ok := sdkmath.NewIntFromString(params.MinTradeAmount)
	if !ok {
		minTrade = sdkmath.NewInt(1_000000)
	}
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
		amt, ok := sdkmath.NewIntFromString(order.AuraAmount)
		if !ok {
			continue
		}
		ts := time.Unix(0, 0)
		if order.Timestamp != nil {
			ts = order.Timestamp.AsTime()
		}
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
	if order.PricePerAura != "" {
		if dec, err := sdkmath.LegacyNewDecFromStr(order.PricePerAura); err == nil {
			return dec
		}
	}
	auraAmt, okA := sdkmath.NewIntFromString(order.AuraAmount)
	otherAmt, okB := sdkmath.NewIntFromString(order.OtherAmount)
	if !okA || !okB || auraAmt.IsZero() {
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
	timeSinceLastTrade := ctx.BlockTime().Unix() - history.LastTradeTime.AsTime().Unix()

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
			FirstDetection:   timestamppb.New(ctx.BlockTime()),
			LastDetection:    timestamppb.New(ctx.BlockTime()),
			IsFlagged:        false,
			ConfidenceScore:  10,
		}
	} else {
		k.cdc.MustUnmarshal(bz, &detection)
		detection.SuspiciousTrades++
		detection.LastDetection = timestamppb.New(ctx.BlockTime())
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
func (k Keeper) CheckDustAttack(ctx sdk.Context, amountIn sdkmath.Int) error {
	params := k.GetSecurityParams(ctx)

	// Parse min trade amount from string
	minTradeAmount, ok := sdkmath.NewIntFromString(params.MinTradeAmount)
	if !ok {
		return fmt.Errorf("invalid min trade amount: %s", params.MinTradeAmount)
	}

	if amountIn.LT(minTradeAmount) {
		return fmt.Errorf(
			"trade amount %s below minimum %s: %w",
			amountIn.String(),
			params.MinTradeAmount,
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

	// Parse min pool creation liquidity from string
	minPoolCreationLiquidity, ok := sdkmath.NewIntFromString(params.MinPoolCreationLiquidity)
	if !ok {
		return fmt.Errorf("invalid min pool creation liquidity: %s", params.MinPoolCreationLiquidity)
	}

	// Check minimum liquidity
	if initialLiquidity.LT(minPoolCreationLiquidity) {
		return fmt.Errorf(
			"initial liquidity %s below minimum %s: %w",
			initialLiquidity.String(),
			params.MinPoolCreationLiquidity,
			types.ErrInsufficientPoolLiquidity,
		)
	}

	// Check pool creation cooldown
	record := k.GetPoolCreationRecord(ctx, creator)
	if record != nil {
		timeSinceLastPool := ctx.BlockTime().Unix() - record.LastCreationTime.AsTime().Unix()
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
			LastCreationTime: timestamppb.New(ctx.BlockTime()),
			TotalPools:       1,
		}
	} else {
		k.cdc.MustUnmarshal(bz, &record)
		record.PoolIds = append(record.PoolIds, poolID)
		record.LastCreationTime = timestamppb.New(ctx.BlockTime())
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
		PausedAt:      timestamppb.New(ctx.BlockTime()),
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
			LastTradeTime:    timestamppb.New(ctx.BlockTime()),
			LastTradeBlock:   currentBlock,
			TradesInBlock:    1,
			RecentTradeCount: 1,
			RecentVolume:     sdkmath.ZeroInt().String(),
		}
	} else {
		k.cdc.MustUnmarshal(bz, &history)

		if history.LastTradeBlock == currentBlock {
			history.TradesInBlock++
		} else {
			history.TradesInBlock = 1
		}

		history.LastTradeTime = timestamppb.New(ctx.BlockTime())
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
