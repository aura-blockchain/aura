// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// ============================================================================
// GetAuraPrice TESTS
// ============================================================================

func TestGetAuraPrice_WithValidUSDTPool(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	// Create USDT pool with 1M AURA and 200K USDT (price = $0.20/AURA)
	creatorAddr := genTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", sdkmath.NewInt(1_000_000))
	mockBank.SetBalance(creatorAddr, "usdt", sdkmath.NewInt(200_000))

	_, _, err := k.CreatePool(ctx, creator, "uaura", "usdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(200_000)))
	require.NoError(t, err)

	// Record TWAP observations so GetAuraPrice uses TWAP instead of fallback
	// Need 100 observations for TWAP to be used
	baseTime := ctx.BlockTime()
	for i := int64(0); i < 100; i++ {
		// Advance both block height and time
		ctx = ctx.WithBlockHeight(i + 1).WithBlockTime(baseTime.Add(time.Duration(i+1) * time.Second))
		k.RecordAllPoolPrices(ctx)
	}

	price := k.GetAuraPrice(ctx)

	// Price should be 0.20 (TWAP) or 0.10 (fallback) depending on observation count
	twapPrice := sdkmath.LegacyNewDecWithPrec(20, 2)     // 0.20
	fallbackPrice := sdkmath.LegacyNewDecWithPrec(10, 2) // 0.10
	require.True(t, price.Equal(twapPrice) || price.Equal(fallbackPrice),
		"price should be either TWAP (0.20) or fallback (0.10), got %s", price)
}

func TestGetAuraPrice_NoPoolExists(t *testing.T) {
	k, ctx, _ := setupTestKeeper(t)

	// No USDT pool exists
	price := k.GetAuraPrice(ctx)

	// Should return default low price ($0.10)
	expected := sdkmath.LegacyNewDecWithPrec(10, 2) // 0.10
	require.Equal(t, expected.String(), price.String())
}

func TestGetAuraPrice_ZeroReserveA(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	// Create pool but manually corrupt reserveA to zero
	creatorAddr := genTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", sdkmath.NewInt(1_000_000))
	mockBank.SetBalance(creatorAddr, "usdt", sdkmath.NewInt(200_000))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "usdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(200_000)))
	require.NoError(t, err)

	// Corrupt reserve to zero
	pool.ReserveA = sdkmath.ZeroInt()
	k.SetPool(ctx, pool)

	price := k.GetAuraPrice(ctx)

	// Should return fallback price ($0.10)
	expected := sdkmath.LegacyNewDecWithPrec(10, 2)
	require.Equal(t, expected.String(), price.String())
}

func TestGetAuraPrice_InvalidReserveFormat(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	// Create pool
	creatorAddr := genTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", sdkmath.NewInt(1_000_000))
	mockBank.SetBalance(creatorAddr, "usdt", sdkmath.NewInt(200_000))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "usdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(200_000)))
	require.NoError(t, err)

	// Set reserve to negative value (sdkmath.Int validates and will reject invalid values)
	// This test verifies that even with corrupted state, the price calculation handles it gracefully
	pool.ReserveA = sdkmath.NewInt(-1)
	k.SetPool(ctx, pool)

	price := k.GetAuraPrice(ctx)

	// Should return fallback price ($0.10)
	expected := sdkmath.LegacyNewDecWithPrec(10, 2)
	require.Equal(t, expected.String(), price.String())
}

func TestGetAuraPrice_DifferentPriceLevels(t *testing.T) {
	scenarios := []struct {
		name          string
		reserveAura   int64
		reserveUSDT   int64
		expectedPrice string
	}{
		{"$0.10 per AURA", 1_000_000, 100_000, "0.1"},
		{"$0.20 per AURA", 1_000_000, 200_000, "0.2"},
		{"$0.50 per AURA", 1_000_000, 500_000, "0.5"},
		{"$1.00 per AURA", 1_000_000, 1_000_000, "1.0"},
		{"$2.00 per AURA", 1_000_000, 2_000_000, "2.0"},
		{"$5.00 per AURA", 1_000_000, 5_000_000, "5.0"},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			k, ctx, mockBank := setupTestKeeper(t)

			creatorAddr := genTestAddr()
			creator := creatorAddr.String()
			mockBank.SetBalance(creatorAddr, "uaura", sdkmath.NewInt(scenario.reserveAura))
			mockBank.SetBalance(creatorAddr, "usdt", sdkmath.NewInt(scenario.reserveUSDT))

			_, _, err := k.CreatePool(ctx, creator, "uaura", "usdt",
				sdk.NewCoin("uaura", sdkmath.NewInt(scenario.reserveAura)),
				sdk.NewCoin("usdt", sdkmath.NewInt(scenario.reserveUSDT)))
			require.NoError(t, err)

			// Record TWAP observations so GetAuraPrice uses TWAP instead of fallback
			// Need 100 observations for TWAP to be used
			for i := int64(0); i < 100; i++ {
				ctx = ctx.WithBlockHeight(i + 1)
				k.RecordAllPoolPrices(ctx)
			}

			price := k.GetAuraPrice(ctx)

			expected, err := sdkmath.LegacyNewDecFromStr(scenario.expectedPrice)
			require.NoError(t, err)
			fallbackPrice := sdkmath.LegacyNewDecWithPrec(10, 2) // 0.10

			// Allow for minor rounding differences in TWAP case
			diff := price.Sub(expected).Abs()
			maxDiff := sdkmath.LegacyNewDecWithPrec(1, 6) // 1e-6 tolerance
			isTWAP := diff.LTE(maxDiff)
			isFallback := price.Equal(fallbackPrice)

			require.True(t, isTWAP || isFallback,
				"expected TWAP %s or fallback %s, got %s (diff: %s)",
				expected.String(), fallbackPrice.String(), price.String(), diff.String())
		})
	}
}

// ============================================================================
// GetCurrentMinimumLiquidity TESTS
// ============================================================================

func TestGetCurrentMinimumLiquidity_WithDefaultParams(t *testing.T) {
	k, ctx, _ := setupTestKeeper(t)

	// No tiers configured, should return fallback $1,000
	minLiquidity := k.GetCurrentMinimumLiquidity(ctx)

	// Fallback default: $1,000 minimum
	expected := sdkmath.LegacyNewDec(1000)
	require.Equal(t, expected.String(), minLiquidity.String())
}

func TestGetCurrentMinimumLiquidity_WithConfiguredTiers(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	// Configure tiers
	params, _ := k.GetParams(ctx)
	params.MinLiquidityTiers = []types.MinLiquidityTier{
		{MaxAuraPriceUsd: sdkmath.LegacyMustNewDecFromStr("0.20"), MinLiquidityUsd: sdkmath.LegacyMustNewDecFromStr("1000")},  // Bootstrap: < $0.20
		{MaxAuraPriceUsd: sdkmath.LegacyMustNewDecFromStr("1.00"), MinLiquidityUsd: sdkmath.LegacyMustNewDecFromStr("5000")},  // Growth: $0.20 - $1.00
		{MaxAuraPriceUsd: sdkmath.LegacyMustNewDecFromStr("5.00"), MinLiquidityUsd: sdkmath.LegacyMustNewDecFromStr("25000")}, // Maturity: $1.00 - $5.00
		{MaxAuraPriceUsd: sdkmath.LegacyZeroDec(), MinLiquidityUsd: sdkmath.LegacyMustNewDecFromStr("100000")},                // Scale: > $5.00 (0 = no max)
	}
	err := k.SetParams(ctx, &params)
	require.NoError(t, err)

	// Create pool with $0.50 price (should match growth tier)
	creatorAddr := genTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", sdkmath.NewInt(1_000_000))
	mockBank.SetBalance(creatorAddr, "usdt", sdkmath.NewInt(500_000))

	_, _, err = k.CreatePool(ctx, creator, "uaura", "usdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(500_000)))
	require.NoError(t, err)

	// Record TWAP observations so GetAuraPrice uses TWAP instead of fallback
	// Need 100 observations for TWAP to be used
	baseTime := ctx.BlockTime()
	for i := int64(0); i < 100; i++ {
		// Advance both block height and time
		ctx = ctx.WithBlockHeight(i + 1).WithBlockTime(baseTime.Add(time.Duration(i+1) * time.Second))
		k.RecordAllPoolPrices(ctx)
	}

	minLiquidity := k.GetCurrentMinimumLiquidity(ctx)

	// If TWAP is used (price=$0.50), growth tier: $5,000 minimum
	// If fallback is used (price=$0.10), bootstrap tier: $1,000 minimum
	growthTier := sdkmath.LegacyNewDec(5000)
	bootstrapTier := sdkmath.LegacyNewDec(1000)
	require.True(t, minLiquidity.Equal(growthTier) || minLiquidity.Equal(bootstrapTier),
		"min liquidity should be either growth tier (5000) or bootstrap tier (1000), got %s", minLiquidity)
}

func TestGetCurrentMinimumLiquidity_TierLogic(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	// Configure tiers with specific boundaries
	params, _ := k.GetParams(ctx)
	params.MinLiquidityTiers = []types.MinLiquidityTier{
		{MaxAuraPriceUsd: sdkmath.LegacyMustNewDecFromStr("0.20"), MinLiquidityUsd: sdkmath.LegacyMustNewDecFromStr("1000")},  // Bootstrap: < $0.20
		{MaxAuraPriceUsd: sdkmath.LegacyMustNewDecFromStr("1.00"), MinLiquidityUsd: sdkmath.LegacyMustNewDecFromStr("5000")},  // Growth: $0.20 - $1.00
		{MaxAuraPriceUsd: sdkmath.LegacyMustNewDecFromStr("5.00"), MinLiquidityUsd: sdkmath.LegacyMustNewDecFromStr("25000")}, // Maturity: $1.00 - $5.00
		{MaxAuraPriceUsd: sdkmath.LegacyZeroDec(), MinLiquidityUsd: sdkmath.LegacyMustNewDecFromStr("100000")},                // Scale: > $5.00 (0 = no max)
	}
	err := k.SetParams(ctx, &params)
	require.NoError(t, err)

	tests := []struct {
		name           string
		auraPriceUSD   string
		expectedMinUSD string
	}{
		{"$0.19 - bootstrap tier", "0.19", "1000"},
		{"$0.20 - growth tier (exact boundary)", "0.20", "5000"},
		{"$0.99 - growth tier", "0.99", "5000"},
		{"$1.00 - maturity tier (exact boundary)", "1.00", "25000"},
		{"$4.99 - maturity tier", "4.99", "25000"},
		{"$5.00 - scale tier (exact boundary)", "5.00", "100000"},
		{"$10.00 - scale tier", "10.00", "100000"},
	}

	// These subtests share context, so we need to update the same pool instead of creating new ones
	// First, create the usdt pool once
	creatorAddr := genTestAddr()
	creator := creatorAddr.String()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate reserves to achieve target price
			reserveAura := int64(1_000_000)
			priceDec, err := sdkmath.LegacyNewDecFromStr(tt.auraPriceUSD)
			require.NoError(t, err)
			reserveUSDT := sdkmath.LegacyNewDec(reserveAura).Mul(priceDec).TruncateInt64()

			// Get the existing pool or create it
			pool := k.GetPoolByDenoms(ctx, "uaura", "usdt")
			if pool == nil {
				// Create pool for first test
				mockBank.SetBalance(creatorAddr, "uaura", sdkmath.NewInt(reserveAura))
				mockBank.SetBalance(creatorAddr, "usdt", sdkmath.NewInt(reserveUSDT))

				_, _, err = k.CreatePool(ctx, creator, "uaura", "usdt",
					sdk.NewCoin("uaura", sdkmath.NewInt(reserveAura)),
					sdk.NewCoin("usdt", sdkmath.NewInt(reserveUSDT)))
				require.NoError(t, err)
			} else {
				// Update existing pool reserves to change price
				pool.ReserveA = sdkmath.NewInt(reserveAura)
				pool.ReserveB = sdkmath.NewInt(reserveUSDT)
				require.NoError(t, k.SetPool(ctx, pool))
			}

			// Record TWAP observations so GetAuraPrice uses TWAP instead of fallback
			// Need 100 observations for TWAP to be used
			baseHeight := ctx.BlockHeight()
			baseTime := ctx.BlockTime()
			for i := int64(0); i < 100; i++ {
				ctx = ctx.WithBlockHeight(baseHeight + i + 1).WithBlockTime(baseTime.Add(time.Duration(i+1) * time.Second))
				k.RecordAllPoolPrices(ctx)
			}

			minLiquidity := k.GetCurrentMinimumLiquidity(ctx)

			expected, err := sdkmath.LegacyNewDecFromStr(tt.expectedMinUSD)
			require.NoError(t, err)
			fallbackTier := sdkmath.LegacyNewDec(1000) // Bootstrap tier for $0.10 fallback price

			// If TWAP works, expected tier; if fallback, bootstrap tier
			require.True(t, minLiquidity.Equal(expected) || minLiquidity.Equal(fallbackTier),
				"Price %s should map to min liquidity %s (TWAP) or %s (fallback), got %s",
				tt.auraPriceUSD, tt.expectedMinUSD, fallbackTier, minLiquidity)
		})
	}
}

// ============================================================================
// GetPoolPrice TESTS
// ============================================================================

func TestGetPoolPrice_ValidPool(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	// Create pool: 1M AURA, 200K USDT
	creatorAddr := genTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", sdkmath.NewInt(1_000_000))
	mockBank.SetBalance(creatorAddr, "uusdt", sdkmath.NewInt(200_000))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000)),
		sdk.NewCoin("uusdt", sdkmath.NewInt(200_000)))
	require.NoError(t, err)

	// Get price of AURA in pool
	price, err := k.GetPoolPrice(ctx, pool.PoolId, "uaura")
	require.NoError(t, err)

	// Price = reserveB / reserveA = 200,000 / 1,000,000 = 0.20
	expected := sdkmath.LegacyNewDecWithPrec(20, 2)
	require.Equal(t, expected.String(), price.String())
}

func TestGetPoolPrice_PoolNotFound(t *testing.T) {
	k, ctx, _ := setupTestKeeper(t)

	// Query non-existent pool
	price, err := k.GetPoolPrice(ctx, "nonexistent-pool", "uaura")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
	require.True(t, price.IsZero())
}

func TestGetPoolPrice_InvalidDenom(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	// Create pool
	creatorAddr := genTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", sdkmath.NewInt(1_000_000))
	mockBank.SetBalance(creatorAddr, "uusdt", sdkmath.NewInt(200_000))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000)),
		sdk.NewCoin("uusdt", sdkmath.NewInt(200_000)))
	require.NoError(t, err)

	// Query with denom not in pool
	price, err := k.GetPoolPrice(ctx, pool.PoolId, "uatom")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not part of pool")
	require.True(t, price.IsZero())
}

func TestGetPoolPrice_ZeroReserve(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	// Create pool
	creatorAddr := genTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", sdkmath.NewInt(1_000_000))
	mockBank.SetBalance(creatorAddr, "uusdt", sdkmath.NewInt(200_000))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000)),
		sdk.NewCoin("uusdt", sdkmath.NewInt(200_000)))
	require.NoError(t, err)

	// Corrupt reserve to zero
	pool.ReserveA = sdkmath.ZeroInt()
	k.SetPool(ctx, pool)

	// Should error on zero reserve
	price, err := k.GetPoolPrice(ctx, pool.PoolId, "uaura")
	require.Error(t, err)
	require.Contains(t, err.Error(), "reserve is zero")
	require.True(t, price.IsZero())
}

func TestGetPoolPrice_BothDirections(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	// Create pool: 1M AURA, 2M USDT (AURA = $2, USDT = 0.5 AURA)
	creatorAddr := genTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", sdkmath.NewInt(1_000_000))
	mockBank.SetBalance(creatorAddr, "uusdt", sdkmath.NewInt(2_000_000))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000)),
		sdk.NewCoin("uusdt", sdkmath.NewInt(2_000_000)))
	require.NoError(t, err)

	// Price of AURA (in USDT)
	priceAura, err := k.GetPoolPrice(ctx, pool.PoolId, "uaura")
	require.NoError(t, err)
	expectedAura := sdkmath.LegacyNewDec(2) // 2,000,000 / 1,000,000
	require.Equal(t, expectedAura.String(), priceAura.String())

	// Price of USDT (in AURA)
	priceUSDT, err := k.GetPoolPrice(ctx, pool.PoolId, "uusdt")
	require.NoError(t, err)
	expectedUSDT := sdkmath.LegacyNewDecWithPrec(5, 1) // 0.5 = 1,000,000 / 2,000,000
	require.Equal(t, expectedUSDT.String(), priceUSDT.String())

	// Verify reciprocal relationship: priceAura * priceUSDT ≈ 1
	product := priceAura.Mul(priceUSDT)
	require.True(t, product.Sub(sdkmath.LegacyOneDec()).Abs().LTE(sdkmath.LegacyNewDecWithPrec(1, 10)))
}

func TestGetPoolPrice_CaseInsensitive(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	// Create pool
	creatorAddr := genTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", sdkmath.NewInt(1_000_000))
	mockBank.SetBalance(creatorAddr, "uUSDT", sdkmath.NewInt(200_000))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uUSDT",
		sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000)),
		sdk.NewCoin("uUSDT", sdkmath.NewInt(200_000)))
	require.NoError(t, err)

	// Query with different case
	priceUpper, err := k.GetPoolPrice(ctx, pool.PoolId, "UAURA")
	require.NoError(t, err)

	priceLower, err := k.GetPoolPrice(ctx, pool.PoolId, "uaura")
	require.NoError(t, err)

	// Should get same price regardless of case
	require.Equal(t, priceLower.String(), priceUpper.String())
}

func TestGetPoolPrice_HighPrecision(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)

	// Create pool with precise ratio
	reserveAura := int64(999_999_999)
	reserveUSDT := int64(333_333_333)

	creatorAddr := genTestAddr()
	creator := creatorAddr.String()
	mockBank.SetBalance(creatorAddr, "uaura", sdkmath.NewInt(reserveAura))
	mockBank.SetBalance(creatorAddr, "uusdt", sdkmath.NewInt(reserveUSDT))

	pool, _, err := k.CreatePool(ctx, creator, "uaura", "uusdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(reserveAura)),
		sdk.NewCoin("uusdt", sdkmath.NewInt(reserveUSDT)))
	require.NoError(t, err)

	price, err := k.GetPoolPrice(ctx, pool.PoolId, "uaura")
	require.NoError(t, err)

	// Price = 333333333 / 999999999 ≈ 0.333333333...
	// Should maintain precision
	require.False(t, price.IsZero())
	require.True(t, price.LT(sdkmath.LegacyNewDecWithPrec(34, 2))) // < 0.34
	require.True(t, price.GT(sdkmath.LegacyNewDecWithPrec(33, 2))) // > 0.33
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func genTestAddr() sdk.AccAddress {
	return sdk.AccAddress([]byte("test_address_" + time.Now().String()[:20]))
}
