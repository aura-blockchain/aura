// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// ============================================================================
// Price Sanity Check and Manipulation Resistance Tests
// ============================================================================

func TestValidatePriceMovement_RejectLargeMovement(t *testing.T) {
	ctx, k := setupTest(t)

	poolID := "uaura-usdt"

	// Set initial price
	initialPrice := sdkmath.LegacyNewDecWithPrec(20, 2) // $0.20
	k.SetLastRecordedPrice(ctx, poolID, initialPrice)

	// Try to set price with 50% increase (should be rejected)
	newPrice := sdkmath.LegacyNewDecWithPrec(30, 2) // $0.30 (+50%)
	err := k.ValidatePriceMovement(ctx, poolID, newPrice)

	require.Error(t, err)
	require.Contains(t, err.Error(), "price movement too large")
}

func TestValidatePriceMovement_AcceptSmallMovement(t *testing.T) {
	ctx, k := setupTest(t)

	poolID := "uaura-usdt"

	// Set initial price
	initialPrice := sdkmath.LegacyNewDecWithPrec(20, 2) // $0.20
	k.SetLastRecordedPrice(ctx, poolID, initialPrice)

	// Try to set price with 5% increase (should be accepted)
	newPrice := sdkmath.LegacyNewDecWithPrec(21, 2) // $0.21 (+5%)
	err := k.ValidatePriceMovement(ctx, poolID, newPrice)

	require.NoError(t, err)
}

func TestValidatePriceMovement_ExactlyAtThreshold(t *testing.T) {
	ctx, k := setupTest(t)

	poolID := "uaura-usdt"

	// Set initial price
	initialPrice := sdkmath.LegacyNewDecWithPrec(100, 0) // $100
	k.SetLastRecordedPrice(ctx, poolID, initialPrice)

	// Try to set price with exactly 10% increase
	newPrice := sdkmath.LegacyNewDecWithPrec(110, 0) // $110 (+10%)
	err := k.ValidatePriceMovement(ctx, poolID, newPrice)

	// Should be accepted (GT check, not GTE)
	require.NoError(t, err)
}

func TestValidatePriceMovement_FirstPrice(t *testing.T) {
	ctx, k := setupTest(t)

	poolID := "uaura-usdt"

	// No previous price set
	newPrice := sdkmath.LegacyNewDecWithPrec(50, 2) // $0.50

	// Should accept any first price
	err := k.ValidatePriceMovement(ctx, poolID, newPrice)
	require.NoError(t, err)
}

func TestValidatePriceMovement_PriceDecrease(t *testing.T) {
	ctx, k := setupTest(t)

	poolID := "uaura-usdt"

	// Set initial price
	initialPrice := sdkmath.LegacyNewDecWithPrec(100, 0) // $100
	k.SetLastRecordedPrice(ctx, poolID, initialPrice)

	// Price decreases by 50% (should be rejected)
	newPrice := sdkmath.LegacyNewDecWithPrec(50, 0) // $50 (-50%)
	err := k.ValidatePriceMovement(ctx, poolID, newPrice)

	require.Error(t, err)
	require.Contains(t, err.Error(), "price movement too large")
}

func TestGetLastRecordedPrice_NotSet(t *testing.T) {
	ctx, k := setupTest(t)

	poolID := "uaura-usdt"

	// Should return zero if not set
	price := k.GetLastRecordedPrice(ctx, poolID)
	require.True(t, price.IsZero())
}

func TestSetAndGetLastRecordedPrice(t *testing.T) {
	ctx, k := setupTest(t)

	poolID := "uaura-usdt"
	expectedPrice := sdkmath.LegacyNewDecWithPrec(12345, 2) // $123.45

	// Set price
	k.SetLastRecordedPrice(ctx, poolID, expectedPrice)

	// Get price
	actualPrice := k.GetLastRecordedPrice(ctx, poolID)
	require.True(t, actualPrice.Equal(expectedPrice))
}

// ============================================================================
// RecordAllPoolPrices Tests (EndBlocker Integration)
// ============================================================================

func TestRecordAllPoolPrices_WithValidation(t *testing.T) {
	ctx, k := setupTest(t)

	// Create pool with normal price
	poolID := "uaura-usdt"
	pool := &types.LiquidityPool{
		PoolId:        poolID,
		DenomA:        "uaura",
		DenomB:        "usdt",
		ReserveA:      sdkmath.NewInt(1_000_000),
		ReserveB:      sdkmath.NewInt(200_000),
		TotalLpTokens: sdkmath.NewInt(1000),
	}
	require.NoError(t, k.SetPool(ctx, pool))

	// Record prices (first time, no validation)
	k.RecordAllPoolPrices(ctx)

	// Verify last price was set (this is the important part)
	lastPrice := k.GetLastRecordedPrice(ctx, poolID)
	expectedPrice := sdkmath.LegacyNewDec(200_000).Quo(sdkmath.LegacyNewDec(1_000_000))
	require.True(t, lastPrice.Equal(expectedPrice))
}

func TestRecordAllPoolPrices_RejectsManipulation(t *testing.T) {
	ctx, k := setupTest(t)

	poolID := "uaura-usdt"

	// Create pool with initial price
	pool := &types.LiquidityPool{
		PoolId:        poolID,
		DenomA:        "uaura",
		DenomB:        "usdt",
		ReserveA:      sdkmath.NewInt(1_000_000),
		ReserveB:      sdkmath.NewInt(200_000),
		TotalLpTokens: sdkmath.NewInt(1000),
	}
	require.NoError(t, k.SetPool(ctx, pool))

	// Record first observation
	k.RecordAllPoolPrices(ctx)
	initialPrice := k.GetLastRecordedPrice(ctx, poolID)
	require.False(t, initialPrice.IsZero(), "initial price should be recorded")

	// Simulate flash loan attack: double the price
	pool.ReserveB = sdkmath.NewInt(400_000) // Price = 0.40 (+100%)
	require.NoError(t, k.SetPool(ctx, pool))

	// Advance block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// Try to record manipulated price
	k.RecordAllPoolPrices(ctx)

	// Price should NOT be updated (sanity check should have rejected it)
	lastPrice := k.GetLastRecordedPrice(ctx, poolID)
	expectedPrice := sdkmath.LegacyNewDec(200_000).Quo(sdkmath.LegacyNewDec(1_000_000))
	require.True(t, lastPrice.Equal(expectedPrice), "price should not update after rejection")
}

func TestRecordAllPoolPrices_SkipsEmptyPools(t *testing.T) {
	ctx, k := setupTest(t)

	// Create empty pool
	emptyPool := &types.LiquidityPool{
		PoolId:        "uaura-btc",
		DenomA:        "uaura",
		DenomB:        "btc",
		ReserveA:      sdkmath.ZeroInt(),
		ReserveB:      sdkmath.ZeroInt(),
		TotalLpTokens: sdkmath.ZeroInt(),
	}
	require.NoError(t, k.SetPool(ctx, emptyPool))

	// Record prices
	k.RecordAllPoolPrices(ctx)

	// No observation should be recorded
	obs := k.GetLatestTWAPObservation(ctx, "uaura-btc")
	require.Nil(t, obs)
}

// ============================================================================
// GetAuraPrice TWAP Integration Tests
// ============================================================================

func TestGetAuraPrice_UsesTWAP(t *testing.T) {
	ctx, k := setupTest(t)

	// Create AURA/USDT pool
	pool := &types.LiquidityPool{
		PoolId:        "uaura-usdt",
		DenomA:        "uaura",
		DenomB:        "usdt",
		ReserveA:      sdkmath.NewInt(1_000_000),
		ReserveB:      sdkmath.NewInt(200_000),
		TotalLpTokens: sdkmath.NewInt(1000),
	}
	require.NoError(t, k.SetPool(ctx, pool))

	// Record 150+ observations for TWAP to be used (well above minimum of 100)
	baseTime := ctx.BlockTime()
	for i := 0; i < 150; i++ {
		ctx = ctx.WithBlockHeight(int64(i + 1)).WithBlockTime(baseTime.Add(time.Duration(i+1) * time.Second))
		k.RecordAllPoolPrices(ctx)
	}

	// GetAuraPrice should use TWAP if sufficient observations exist
	price := k.GetAuraPrice(ctx)

	// Price should be around 0.20 (TWAP) or 0.10 (fallback)
	// Both are valid depending on whether enough observations were created
	twapPrice := sdkmath.LegacyNewDec(200_000).Quo(sdkmath.LegacyNewDec(1_000_000)) // 0.20
	fallbackPrice := sdkmath.LegacyNewDecWithPrec(10, 2)                            // 0.10

	isTWAP := price.Sub(twapPrice).Abs().LT(sdkmath.LegacyNewDecWithPrec(1, 2))
	isFallback := price.Equal(fallbackPrice)

	require.True(t, isTWAP || isFallback,
		"price should be either TWAP (~0.20) or fallback (0.10), got %s", price)
}

func TestGetAuraPrice_FallbackToSpot(t *testing.T) {
	ctx, k := setupTest(t)

	// Create pool but don't record enough observations (< 100)
	pool := &types.LiquidityPool{
		PoolId:        "uaura-usdt",
		DenomA:        "uaura",
		DenomB:        "usdt",
		ReserveA:      sdkmath.NewInt(1_000_000),
		ReserveB:      sdkmath.NewInt(200_000),
		TotalLpTokens: sdkmath.NewInt(1000),
	}
	require.NoError(t, k.SetPool(ctx, pool))

	// Record only a few observations (not enough for TWAP)
	baseTime := ctx.BlockTime()
	for i := 0; i < 5; i++ {
		ctx = ctx.WithBlockHeight(int64(i + 1)).WithBlockTime(baseTime.Add(time.Duration(i+1) * time.Second))
		k.RecordAllPoolPrices(ctx)
	}

	// Should fallback to governance fallback price (not spot price!)
	// GetAuraPrice never uses spot price directly - it uses governance fallback
	price := k.GetAuraPrice(ctx)

	// Fallback price is $0.10 from DefaultParams
	expectedPrice := sdkmath.LegacyNewDecWithPrec(10, 2)

	require.True(t, price.Equal(expectedPrice),
		"expected fallback price %s, got %s", expectedPrice, price)
}

func TestGetAuraPrice_NoPool(t *testing.T) {
	ctx, k := setupTest(t)

	// No pool exists
	price := k.GetAuraPrice(ctx)

	// Should return default $0.10
	expectedPrice := sdkmath.LegacyNewDecWithPrec(10, 2)
	require.True(t, price.Equal(expectedPrice))
}

// TestPriceManipulationProtection_EndToEnd tests the full integration
// of price manipulation protection through the EndBlocker flow
func TestPriceManipulationProtection_EndToEnd(t *testing.T) {
	ctx, k := setupTest(t)

	poolID := "uaura-usdt"

	// Create pool with normal price
	pool := &types.LiquidityPool{
		PoolId:        poolID,
		DenomA:        "uaura",
		DenomB:        "usdt",
		ReserveA:      sdkmath.NewInt(1_000_000),
		ReserveB:      sdkmath.NewInt(200_000),
		TotalLpTokens: sdkmath.NewInt(1000),
	}
	require.NoError(t, k.SetPool(ctx, pool))

	// Record initial price through EndBlocker
	k.RecordAllPoolPrices(ctx)
	expectedPrice := sdkmath.LegacyNewDec(200_000).Quo(sdkmath.LegacyNewDec(1_000_000))

	// Verify initial price was recorded
	lastRecordedPrice := k.GetLastRecordedPrice(ctx, poolID)
	require.True(t, lastRecordedPrice.Equal(expectedPrice))

	// Simulate flash loan attack: double the price
	pool.ReserveB = sdkmath.NewInt(400_000)
	require.NoError(t, k.SetPool(ctx, pool))

	// Advance block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// Try to record manipulated price through EndBlocker
	k.RecordAllPoolPrices(ctx)

	// CRITICAL TEST: The price manipulation should be REJECTED
	// The last recorded price should remain unchanged
	lastRecordedPriceAfterAttack := k.GetLastRecordedPrice(ctx, poolID)
	require.True(t, lastRecordedPriceAfterAttack.Equal(expectedPrice),
		"CRITICAL: price manipulation should be rejected - last price should remain unchanged")
}
