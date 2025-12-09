package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// TestGetQuote tests the quote calculation functionality
func TestGetQuote(t *testing.T) {
	ctx, k := setupTest(t)

	// Create a test pool with proper math.Int types for reserves
	pool := &types.LiquidityPool{
		PoolId:                "uaura-usdt",
		DenomA:                "uaura",
		DenomB:                "usdt",
		ReserveA:              sdkmath.NewInt(1000000000), // 1000 AURA
		ReserveB:              sdkmath.NewInt(500000000),  // 500 USDT
		TotalLpTokens:         sdkmath.NewInt(707106781),
		FeePercentage:         sdkmath.LegacyMustNewDecFromStr("0.003"),  // 0.3%
		ProtocolFeePercentage: sdkmath.LegacyMustNewDecFromStr("0.0005"), // 0.05%
	}
	k.SetPool(ctx, pool)

	tests := []struct {
		name          string
		poolID        string
		tokenIn       string
		amountIn      sdkmath.Int
		expectedError bool
		checkOutput   bool
	}{
		{
			name:          "valid quote - swap AURA for USDT",
			poolID:        "uaura-usdt",
			tokenIn:       "uaura",
			amountIn:      sdkmath.NewInt(1000000), // 1 AURA
			expectedError: false,
			checkOutput:   true,
		},
		{
			name:          "valid quote - swap USDT for AURA",
			poolID:        "uaura-usdt",
			tokenIn:       "usdt",
			amountIn:      sdkmath.NewInt(1000000), // 1 USDT
			expectedError: false,
			checkOutput:   true,
		},
		{
			name:          "pool not found",
			poolID:        "nonexistent",
			tokenIn:       "uaura",
			amountIn:      sdkmath.NewInt(1000000),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GetQuote returns 5 values: (estimatedOutput, effectivePrice, priceImpact, feeAmount, error)
			amountOut, effectivePrice, priceImpact, feeAmount, err := k.GetQuote(ctx, tt.poolID, tt.tokenIn, tt.amountIn)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.checkOutput {
					require.True(t, amountOut.GT(sdkmath.ZeroInt()), "output amount should be positive")
					require.NotNil(t, effectivePrice)
					require.NotNil(t, priceImpact)
					require.NotNil(t, feeAmount)
					require.True(t, amountOut.LT(tt.amountIn.Mul(sdkmath.NewInt(2))), "output should be reasonable")
				}
			}
		})
	}
}

// TestRecordSwapStats tests swap statistics recording
func TestRecordSwapStats(t *testing.T) {
	ctx, k := setupTest(t)

	// Create a test pool
	pool := &types.LiquidityPool{
		PoolId:                "uaura-usdt",
		DenomA:                "uaura",
		DenomB:                "usdt",
		ReserveA:              sdkmath.NewInt(1000000000),
		ReserveB:              sdkmath.NewInt(500000000),
		TotalLpTokens:         sdkmath.NewInt(707106781),
		FeePercentage:         sdkmath.LegacyMustNewDecFromStr("0.003"),
		ProtocolFeePercentage: sdkmath.LegacyMustNewDecFromStr("0.0005"),
	}
	k.SetPool(ctx, pool)

	tests := []struct {
		name      string
		poolID    string
		amountIn  sdkmath.Int
		amountOut sdkmath.Int
	}{
		{
			name:      "record AURA to USDT swap",
			poolID:    "uaura-usdt",
			amountIn:  sdkmath.NewInt(1000000),
			amountOut: sdkmath.NewInt(490000),
		},
		{
			name:      "record USDT to AURA swap",
			poolID:    "uaura-usdt",
			amountIn:  sdkmath.NewInt(1000000),
			amountOut: sdkmath.NewInt(1980000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Record the swap - RecordSwapStats takes amountIn, amountOut, timestamp
			k.RecordSwapStats(ctx, tt.poolID, tt.amountIn, tt.amountOut, time.Now())

			// Get the pool and verify it still exists
			updatedPool := k.GetPool(ctx, tt.poolID)
			require.NotNil(t, updatedPool)

			stats, found := k.GetSwapStats(ctx, tt.poolID)
			require.True(t, found)
			require.Equal(t, tt.poolID, stats.PoolId)

			price, found := k.GetMarketPrice(ctx, "usdt")
			require.True(t, found)
			require.Equal(t, "usdt", price.Coin)
		})
	}
}

// TestGetQuoteEdgeCases tests edge cases for quote calculation
func TestGetQuoteEdgeCases(t *testing.T) {
	ctx, k := setupTest(t)

	// Create a pool with very low liquidity
	lowLiqPool := &types.LiquidityPool{
		PoolId:                "uaura-token",
		DenomA:                "uaura",
		DenomB:                "token",
		ReserveA:              sdkmath.NewInt(1000), // Very low
		ReserveB:              sdkmath.NewInt(1000),
		TotalLpTokens:         sdkmath.NewInt(1000),
		FeePercentage:         sdkmath.LegacyMustNewDecFromStr("0.003"),
		ProtocolFeePercentage: sdkmath.LegacyMustNewDecFromStr("0.0005"),
	}
	k.SetPool(ctx, lowLiqPool)

	// Create a pool with imbalanced reserves
	imbalancedPool := &types.LiquidityPool{
		PoolId:                "uaura-stable",
		DenomA:                "uaura",
		DenomB:                "stable",
		ReserveA:              sdkmath.NewInt(1000000000000), // 1M AURA
		ReserveB:              sdkmath.NewInt(1000),          // 0.000001 stable
		TotalLpTokens:         sdkmath.NewInt(31622776601),
		FeePercentage:         sdkmath.LegacyMustNewDecFromStr("0.003"),
		ProtocolFeePercentage: sdkmath.LegacyMustNewDecFromStr("0.0005"),
	}
	k.SetPool(ctx, imbalancedPool)

	tests := []struct {
		name          string
		poolID        string
		tokenIn       string
		amountIn      sdkmath.Int
		expectedError bool
	}{
		{
			name:          "large swap in low liquidity pool",
			poolID:        "uaura-token",
			tokenIn:       "uaura",
			amountIn:      sdkmath.NewInt(10000), // Much larger than reserves
			expectedError: false,                 // Should work but with high slippage
		},
		{
			name:          "swap in imbalanced pool",
			poolID:        "uaura-stable",
			tokenIn:       "uaura",
			amountIn:      sdkmath.NewInt(1000000),
			expectedError: false,
		},
		{
			name:          "very small swap amount",
			poolID:        "uaura-token",
			tokenIn:       "uaura",
			amountIn:      sdkmath.NewInt(1), // Minimum amount
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amountOut, _, _, _, err := k.GetQuote(ctx, tt.poolID, tt.tokenIn, tt.amountIn)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, amountOut)
			}
		})
	}
}

// TestRecordSwapStatsMultipleSwaps tests that stats accumulate correctly
func TestRecordSwapStatsMultipleSwaps(t *testing.T) {
	ctx, k := setupTest(t)

	pool := &types.LiquidityPool{
		PoolId:                "uaura-usdt",
		DenomA:                "uaura",
		DenomB:                "usdt",
		ReserveA:              sdkmath.NewInt(1000000000),
		ReserveB:              sdkmath.NewInt(500000000),
		TotalLpTokens:         sdkmath.NewInt(707106781),
		FeePercentage:         sdkmath.LegacyMustNewDecFromStr("0.003"),
		ProtocolFeePercentage: sdkmath.LegacyMustNewDecFromStr("0.0005"),
	}
	k.SetPool(ctx, pool)

	// Record multiple swaps
	for i := 0; i < 5; i++ {
		k.RecordSwapStats(
			ctx,
			"uaura-usdt",
			sdkmath.NewInt(1000000),
			sdkmath.NewInt(490000),
			time.Now(),
		)
	}

	// Verify pool still exists and is valid
	updatedPool := k.GetPool(ctx, "uaura-usdt")
	require.NotNil(t, updatedPool)
	require.Equal(t, "uaura-usdt", updatedPool.PoolId)

	price, found := k.GetMarketPrice(ctx, "usdt")
	require.True(t, found)
	require.Equal(t, uint64(5), price.SampleSize)
}
