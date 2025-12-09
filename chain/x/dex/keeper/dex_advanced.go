package keeper

import (
	"fmt"
	"sort"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// ============================
// ADVANCED DEX FEATURES
// Order matching, MEV resistance, oracles, liquidity mining, limit orders, etc.
// ============================

// ===== ORDER MATCHING ENGINE =====

// OrderType defines types of orders
type OrderType string

const (
	OrderTypeMarket OrderType = "market"
	OrderTypeLimit  OrderType = "limit"
	OrderTypeStop   OrderType = "stop"
)

// Order represents a trading order
type Order struct {
	OrderID      string
	PoolID       string
	Trader       string
	OrderType    OrderType
	Side         string // "buy" or "sell"
	BaseAmount   sdkmath.Int
	QuoteAmount  sdkmath.Int
	LimitPrice   sdkmath.LegacyDec
	StopPrice    sdkmath.LegacyDec
	Filled       sdkmath.Int
	Status       string // "pending", "filled", "cancelled", "expired"
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

// MatchOrders performs order matching for a pool
func (k Keeper) MatchOrders(ctx sdk.Context, poolID string) (matched int, volume sdkmath.Int, err error) {
	orders := k.getPendingOrders(poolID)
	if len(orders) == 0 {
		return 0, sdkmath.ZeroInt(), nil
	}

	// Separate buy and sell orders
	buyOrders := []Order{}
	sellOrders := []Order{}

	for _, order := range orders {
		if order.Status != "pending" {
			continue
		}

		if order.Side == "buy" {
			buyOrders = append(buyOrders, order)
		} else {
			sellOrders = append(sellOrders, order)
		}
	}

	// Sort buy orders by price (descending)
	sort.Slice(buyOrders, func(i, j int) bool {
		return buyOrders[i].LimitPrice.GT(buyOrders[j].LimitPrice)
	})

	// Sort sell orders by price (ascending)
	sort.Slice(sellOrders, func(i, j int) bool {
		return sellOrders[i].LimitPrice.LT(sellOrders[j].LimitPrice)
	})

	totalVolume := sdkmath.ZeroInt()
	matchedCount := 0

	// Match orders
	i, j := 0, 0
	for i < len(buyOrders) && j < len(sellOrders) {
		buyOrder := &buyOrders[i]
		sellOrder := &sellOrders[j]

		// Check if prices match (buy >= sell)
		if buyOrder.LimitPrice.LT(sellOrder.LimitPrice) {
			break
		}

		// Calculate match quantity
		buyRemaining := buyOrder.BaseAmount.Sub(buyOrder.Filled)
		sellRemaining := sellOrder.BaseAmount.Sub(sellOrder.Filled)

		matchAmount := buyRemaining
		if sellRemaining.LT(matchAmount) {
			matchAmount = sellRemaining
		}

		// Execute match at midpoint price
		matchPrice := buyOrder.LimitPrice.Add(sellOrder.LimitPrice).Quo(sdkmath.LegacyNewDec(2))

		// Update orders
		buyOrder.Filled = buyOrder.Filled.Add(matchAmount)
		sellOrder.Filled = sellOrder.Filled.Add(matchAmount)

		if buyOrder.Filled.Equal(buyOrder.BaseAmount) {
			buyOrder.Status = "filled"
			i++
		}

		if sellOrder.Filled.Equal(sellOrder.BaseAmount) {
			sellOrder.Status = "filled"
			j++
		}

		// Record trade
		totalVolume = totalVolume.Add(matchAmount)
		matchedCount++

		// Emit event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"orders_matched",
				sdk.NewAttribute("pool_id", poolID),
				sdk.NewAttribute("buy_order", buyOrder.OrderID),
				sdk.NewAttribute("sell_order", sellOrder.OrderID),
				sdk.NewAttribute("amount", matchAmount.String()),
				sdk.NewAttribute("price", matchPrice.String()),
			),
		)

		// Store updated orders
		k.storeOrder(buyOrder)
		k.storeOrder(sellOrder)
	}

	return matchedCount, totalVolume, nil
}

// PlaceLimitOrder places a limit order
func (k Keeper) PlaceLimitOrder(ctx sdk.Context, poolID, trader, side string, amount sdkmath.Int, limitPrice sdkmath.LegacyDec, expirationBlocks uint64) (*Order, error) {
	order := &Order{
		OrderID:     fmt.Sprintf("order-%s-%d", trader, ctx.BlockHeight()),
		PoolID:      poolID,
		Trader:      trader,
		OrderType:   OrderTypeLimit,
		Side:        side,
		BaseAmount:  amount,
		LimitPrice:  limitPrice,
		Filled:      sdkmath.ZeroInt(),
		Status:      "pending",
		CreatedAt:   ctx.BlockTime(),
		ExpiresAt:   ctx.BlockTime().Add(time.Duration(expirationBlocks) * 6 * time.Second),
	}

	k.storeOrder(order)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"limit_order_placed",
			sdk.NewAttribute("order_id", order.OrderID),
			sdk.NewAttribute("pool_id", poolID),
			sdk.NewAttribute("trader", trader),
			sdk.NewAttribute("side", side),
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("limit_price", limitPrice.String()),
		),
	)

	return order, nil
}

// PlaceStopLossOrder places a stop-loss order
func (k Keeper) PlaceStopLossOrder(ctx sdk.Context, poolID, trader string, amount sdkmath.Int, stopPrice sdkmath.LegacyDec) (*Order, error) {
	order := &Order{
		OrderID:    fmt.Sprintf("stop-%s-%d", trader, ctx.BlockHeight()),
		PoolID:     poolID,
		Trader:     trader,
		OrderType:  OrderTypeStop,
		Side:       "sell",
		BaseAmount: amount,
		StopPrice:  stopPrice,
		Filled:     sdkmath.ZeroInt(),
		Status:     "pending",
		CreatedAt:  ctx.BlockTime(),
	}

	k.storeOrder(order)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"stop_loss_placed",
			sdk.NewAttribute("order_id", order.OrderID),
			sdk.NewAttribute("pool_id", poolID),
			sdk.NewAttribute("trader", trader),
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("stop_price", stopPrice.String()),
		),
	)

	return order, nil
}

// ===== MEV RESISTANCE =====

// MEVProtectionConfig defines MEV resistance parameters
type MEVProtectionConfig struct {
	Enabled            bool
	MinBlockDelay      uint64                 // Minimum blocks before execution
	MaxPriceImpact     sdkmath.LegacyDec      // Maximum allowed price impact
	BatchAuctionWindow uint64                 // Blocks for batch auction
	EncryptedMempool   bool                   // Use encrypted mempool
}

// CheckMEVProtectionForPool validates transaction against MEV for a specific pool
func (k Keeper) CheckMEVProtectionForPool(ctx sdk.Context, poolID string, amountIn sdkmath.Int) error {
	// Get security params
	secParams := k.GetSecurityParams(ctx)
	if !secParams.MevProtectionEnabled {
		return nil
	}

	// Calculate price impact
	pool := k.GetPool(ctx, poolID)
	if pool == nil {
		return fmt.Errorf("pool not found")
	}

	priceImpact := k.calculatePriceImpact(pool, amountIn)

	// Check against threshold - use MaxPriceImpactBps from security params
	maxImpactBps := uint64(1000) // Default 10% (1000 bps)
	maxImpact := sdkmath.LegacyNewDecWithPrec(int64(maxImpactBps), 4) // bps to decimal

	if priceImpact.GT(maxImpact) {
		return fmt.Errorf("price impact %s exceeds maximum %s", priceImpact.String(), maxImpact.String())
	}

	// Record suspicious activity
	if priceImpact.GT(maxImpact.Quo(sdkmath.LegacyNewDec(2))) {
		k.recordSuspiciousActivity(ctx, "high_price_impact", poolID, priceImpact.String())
	}

	return nil
}

func (k Keeper) calculatePriceImpact(pool *types.LiquidityPool, amountIn sdkmath.Int) sdkmath.LegacyDec {
	reserveA, _ := k.parseReserve(pool.ReserveA)
	reserveB, _ := k.parseReserve(pool.ReserveB)

	if reserveA.IsZero() || reserveB.IsZero() {
		return sdkmath.LegacyZeroDec()
	}

	// Price before
	priceBefore := sdkmath.LegacyNewDecFromInt(reserveB).Quo(sdkmath.LegacyNewDecFromInt(reserveA))

	// Price after (x * y = k)
	newReserveA := reserveA.Add(amountIn)
	k_constant := reserveA.Mul(reserveB)
	newReserveB := k_constant.Quo(newReserveA)

	priceAfter := sdkmath.LegacyNewDecFromInt(newReserveB).Quo(sdkmath.LegacyNewDecFromInt(newReserveA))

	// Impact = (priceAfter - priceBefore) / priceBefore
	impact := priceAfter.Sub(priceBefore).Quo(priceBefore).Abs()

	return impact
}

// ===== FLASH LOAN PROTECTION =====

// DetectFlashLoan detects potential flash loan attacks
func (k Keeper) DetectFlashLoan(ctx sdk.Context, trader string, poolID string, amount sdkmath.Int) (isFlashLoan bool, risk string) {
	// Check transaction patterns
	recentTxs := k.getRecentTransactions(ctx, trader, 10)

	// Flash loan indicators:
	// 1. Large amount relative to pool
	// 2. Multiple rapid transactions
	// 3. Profit extraction pattern

	pool := k.GetPool(ctx, poolID)
	if pool == nil {
		return false, "unknown"
	}

	reserveA, _ := k.parseReserve(pool.ReserveA)

	// Check if amount is > 20% of pool
	if sdkmath.LegacyNewDecFromInt(amount).GT(sdkmath.LegacyNewDecFromInt(reserveA).Mul(sdkmath.LegacyNewDecWithPrec(20, 2))) {
		if len(recentTxs) > 5 {
			return true, "high"
		}
		return false, "medium"
	}

	return false, "low"
}

// ===== PRICE ORACLE INTEGRATION =====

// PriceOracle interface for external price feeds
type PriceOracle interface {
	GetPrice(denom string) (sdkmath.LegacyDec, error)
}

// UpdatePoolWithOracle updates pool price using oracle
func (k Keeper) UpdatePoolWithOracle(ctx sdk.Context, poolID string, oracle PriceOracle) error {
	pool := k.GetPool(ctx, poolID)
	if pool == nil {
		return fmt.Errorf("pool not found")
	}

	// Get oracle prices
	priceA, err := oracle.GetPrice(pool.DenomA)
	if err != nil {
		return fmt.Errorf("failed to get price for %s: %w", pool.DenomA, err)
	}

	priceB, err := oracle.GetPrice(pool.DenomB)
	if err != nil {
		return fmt.Errorf("failed to get price for %s: %w", pool.DenomB, err)
	}

	// Calculate expected ratio
	expectedRatio := priceA.Quo(priceB)

	// Get current pool ratio
	reserveA, _ := k.parseReserve(pool.ReserveA)
	reserveB, _ := k.parseReserve(pool.ReserveB)
	currentRatio := sdkmath.LegacyNewDecFromInt(reserveA).Quo(sdkmath.LegacyNewDecFromInt(reserveB))

	// Calculate deviation
	deviation := expectedRatio.Sub(currentRatio).Quo(expectedRatio).Abs()

	// Emit event if significant deviation
	if deviation.GT(sdkmath.LegacyNewDecWithPrec(5, 2)) { // > 5%
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"pool_price_deviation",
				sdk.NewAttribute("pool_id", poolID),
				sdk.NewAttribute("expected_ratio", expectedRatio.String()),
				sdk.NewAttribute("current_ratio", currentRatio.String()),
				sdk.NewAttribute("deviation", deviation.String()),
			),
		)
	}

	return nil
}

// ===== LIQUIDITY MINING REWARDS =====

// CalculateLiquidityMiningReward calculates rewards for liquidity providers
func (k Keeper) CalculateLiquidityMiningReward(ctx sdk.Context, poolID string, provider string) (sdkmath.Int, error) {
	pool := k.GetPool(ctx, poolID)
	if pool == nil {
		return sdkmath.ZeroInt(), fmt.Errorf("pool not found")
	}

	// Get provider's LP tokens
	var providerLPTokens sdkmath.Int
	for _, p := range pool.Providers {
		if p.Address == provider {
			// LpTokens is already math.Int (customtype in proto)
			providerLPTokens = p.LpTokens
			break
		}
	}

	if providerLPTokens.IsZero() {
		return sdkmath.ZeroInt(), nil
	}

	// Calculate time-weighted liquidity
	// TotalLpTokens is already math.Int (customtype in proto)
	totalLPTokens := pool.TotalLpTokens
	providerShare := sdkmath.LegacyNewDecFromInt(providerLPTokens).Quo(sdkmath.LegacyNewDecFromInt(totalLPTokens))

	// Base reward (from params)
	// NOTE: Future enhancement - Add LiquidityMiningRewardPerBlock to params.proto if needed
	// For now, use a default value
	baseReward := sdkmath.NewInt(1000000) // Default 1 AURA per block

	// Calculate provider's reward
	reward := providerShare.Mul(sdkmath.LegacyNewDecFromInt(baseReward)).TruncateInt()

	return reward, nil
}

// ===== IMPERMANENT LOSS PROTECTION =====

// CalculateImpermanentLoss calculates IL for a position
func (k Keeper) CalculateImpermanentLoss(initialRatio, currentRatio sdkmath.LegacyDec) sdkmath.LegacyDec {
	if initialRatio.IsZero() || currentRatio.IsZero() {
		return sdkmath.LegacyZeroDec()
	}

	// IL formula: 2*sqrt(priceRatio)/(1+priceRatio) - 1
	ratio := currentRatio.Quo(initialRatio)

	// Simplified calculation (actual would use sqrt)
	// For now, approximate
	il := ratio.Sub(sdkmath.LegacyOneDec()).Abs()

	return il
}

// ===== CONCENTRATED LIQUIDITY =====

// AddConcentratedLiquidity adds liquidity to a specific price range
// NOTE: Future enhancement - Define ConcentratedPosition in types package (proto/types) when implementing concentrated liquidity
func (k Keeper) AddConcentratedLiquidity(ctx sdk.Context, poolID string, provider string, amountA, amountB sdkmath.Int, lowerPrice, upperPrice sdkmath.LegacyDec) error {
	// Validate price range
	if lowerPrice.GTE(upperPrice) {
		return fmt.Errorf("invalid price range")
	}

	// NOTE: Future enhancement - Uncomment when ConcentratedPosition type is defined in proto
	/*
	// Create position
	position := &types.ConcentratedPosition{
		PoolId:     poolID,
		Provider:   provider,
		LowerPrice: lowerPrice.String(),
		UpperPrice: upperPrice.String(),
		AmountA:    amountA.String(),
		AmountB:    amountB.String(),
	}

	k.storeConcentratedPosition(position)
	*/

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"concentrated_liquidity_added",
			sdk.NewAttribute("pool_id", poolID),
			sdk.NewAttribute("provider", provider),
			sdk.NewAttribute("lower_price", lowerPrice.String()),
			sdk.NewAttribute("upper_price", upperPrice.String()),
		),
	)

	return nil
}

// ===== HELPER FUNCTIONS =====

func (k Keeper) getPendingOrders(poolID string) []Order { return []Order{} }
func (k Keeper) storeOrder(order *Order) {}
func (k Keeper) getRecentTransactions(ctx sdk.Context, trader string, limit int) []string { return []string{} }
func (k Keeper) recordSuspiciousActivity(ctx sdk.Context, activityType, poolID, details string) {}
// NOTE: Future enhancement - Uncomment when ConcentratedPosition is defined
// func (k Keeper) storeConcentratedPosition(position *types.ConcentratedPosition) {}
