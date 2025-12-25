// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

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
// ORDER MANIPULATION DETECTION
// Extracted from security.go for better code organization
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
// WASH TRADING DETECTION
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
// POOL CREATION LIMITS AND VALIDATION
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
// POOL CREATION AUDIT TRAIL
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

	records := make([]*types.PoolCreationRecord, 0, 64)
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
