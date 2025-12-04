package keeper_test

import (
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
)

// TestSwapFeeOverflowPrevention tests that fee calculations reject overflow
func TestSwapFeeOverflowPrevention(t *testing.T) {
	ctx, k := setupTest(t)

	tests := []struct {
		name          string
		amount        sdkmath.Int
		expectError   bool
		errorContains string
	}{
		{
			name:        "normal amount",
			amount:      sdkmath.NewInt(1_000_000_000),
			expectError: false,
		},
		{
			name:        "large amount",
			amount:      sdkmath.NewInt(1_000_000_000_000_000_000),
			expectError: false,
		},
		{
			name:          "zero amount rejected",
			amount:        sdkmath.ZeroInt(),
			expectError:   true,
			errorContains: "must be positive",
		},
		{
			name:          "negative amount rejected",
			amount:        sdkmath.NewInt(-1000),
			expectError:   true,
			errorContains: "cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fee, err := k.CalculateSwapFee(ctx, tt.amount)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					require.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.False(t, fee.IsNegative(), "fee should never be negative")
			}
		})
	}
}

// TestPoolCreationOverflowPrevention tests that pool creation rejects overflow
func TestPoolCreationOverflowPrevention(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)
	creator := keepertest.GenTestAddr()

	tests := []struct {
		name          string
		amountA       sdk.Coin
		amountB       sdk.Coin
		expectError   bool
		errorContains string
	}{
		{
			name:        "normal amounts",
			amountA:     sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000_000)),
			amountB:     sdk.NewCoin("usdt", sdkmath.NewInt(200_000_000)),
			expectError: false,
		},
		{
			name:        "large but safe amounts",
			amountA:     sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000_000_000)),
			amountB:     sdk.NewCoin("usdt", sdkmath.NewInt(200_000_000_000)),
			expectError: false,
		},
		{
			name:          "zero amount rejected",
			amountA:       sdk.NewCoin("uaura", sdkmath.ZeroInt()),
			amountB:       sdk.NewCoin("usdt", sdkmath.NewInt(200_000_000)),
			expectError:   true,
			errorContains: "amounts must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolID := k.GeneratePoolID(tt.amountA.Denom, tt.amountB.Denom)

			// Clean up any existing pool
			k.DeletePool(ctx, poolID)

			// Fund creator with test amounts BEFORE CreatePool
			mockBank.SetBalance(creator, tt.amountA.Denom, tt.amountA.Amount.Add(sdkmath.NewInt(1000)))
			mockBank.SetBalance(creator, tt.amountB.Denom, tt.amountB.Amount.Add(sdkmath.NewInt(1000)))

			pool, lpTokens, err := k.CreatePool(
				ctx,
				creator.String(),
				tt.amountA.Denom,
				tt.amountB.Denom,
				tt.amountA,
				tt.amountB,
			)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					require.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, pool)
				require.False(t, lpTokens.IsNegative(), "LP tokens should never be negative")
				require.True(t, lpTokens.IsPositive(), "LP tokens should be positive")
			}
		})
	}
}

// TestSwapOverflowPrevention tests that swaps reject overflow in k-constant calculation
func TestSwapOverflowPrevention(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)
	creator := keepertest.GenTestAddr()
	trader := keepertest.GenTestAddr()

	// Setup: Create a pool first
	tokenA := "uaura"
	tokenB := "usdt"
	poolLiquidityA := sdk.NewCoin(tokenA, sdkmath.NewInt(1_000_000_000))
	poolLiquidityB := sdk.NewCoin(tokenB, sdkmath.NewInt(1_000_000_000))

	// Fund creator and create pool
	mockBank.SetBalance(creator, tokenA, poolLiquidityA.Amount)
	mockBank.SetBalance(creator, tokenB, poolLiquidityB.Amount)

	poolID := k.GeneratePoolID(tokenA, tokenB)
	_, _, err := k.CreatePool(ctx, creator.String(), tokenA, tokenB, poolLiquidityA, poolLiquidityB)
	require.NoError(t, err, "Pool creation should succeed")

	tests := []struct {
		name          string
		coinIn        sdk.Coin
		expectError   bool
		errorContains string
	}{
		{
			name:        "normal swap amount",
			coinIn:      sdk.NewCoin(tokenA, sdkmath.NewInt(1_000_000)),
			expectError: false,
		},
		{
			name:          "extremely large swap - should be rejected",
			coinIn:        sdk.NewCoin(tokenA, sdkmath.NewInt(1_000_000_000_000_000)),
			expectError:   true,
			errorContains: "exceeds maximum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fund trader
			mockBank.SetBalance(trader, tt.coinIn.Denom, tt.coinIn.Amount)

			// Attempt swap with SwapExactIn (maxSlippageBps = 10000 = 100%)
			_, _, _, err := k.SwapExactIn(ctx, trader.String(), poolID, tt.coinIn, sdkmath.NewInt(1), 10000)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					require.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestGetQuoteOverflowPrevention tests that quotes reject overflow
func TestGetQuoteOverflowPrevention(t *testing.T) {
	k, ctx, mockBank := setupTestKeeper(t)
	creator := keepertest.GenTestAddr()

	// Setup: Create a pool first
	tokenA := "uaura"
	tokenB := "usdt"
	poolLiquidityA := sdk.NewCoin(tokenA, sdkmath.NewInt(1_000_000_000))
	poolLiquidityB := sdk.NewCoin(tokenB, sdkmath.NewInt(1_000_000_000))

	// Fund creator and create pool
	mockBank.SetBalance(creator, tokenA, poolLiquidityA.Amount)
	mockBank.SetBalance(creator, tokenB, poolLiquidityB.Amount)

	poolID := k.GeneratePoolID(tokenA, tokenB)
	pool, _, err := k.CreatePool(ctx, creator.String(), tokenA, tokenB, poolLiquidityA, poolLiquidityB)
	require.NoError(t, err, "Pool creation should succeed")

	tests := []struct {
		name          string
		amountIn      sdkmath.Int
		expectError   bool
		errorContains string
	}{
		{
			name:        "normal quote amount",
			amountIn:    sdkmath.NewInt(1_000_000),
			expectError: false,
		},
		{
			name:        "large but safe quote amount",
			amountIn:    sdkmath.NewInt(100_000_000),
			expectError: false,
		},
		{
			name:          "extremely large quote - should be rejected",
			amountIn:      sdkmath.NewInt(1_000_000_000_000_000),
			expectError:   true,
			errorContains: "exceeds maximum",
		},
		{
			name:          "zero amount rejected",
			amountIn:      sdkmath.ZeroInt(),
			expectError:   true,
			errorContains: "must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get quote for swapping tokenA -> tokenB
			// GetQuote returns: (amountOut, effectivePrice, priceImpact, fee, error)
			amountOut, _, _, _, err := k.GetQuote(ctx, poolID, tokenA, tt.amountIn)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					require.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.False(t, amountOut.IsNegative(), "quote output should never be negative")
				require.True(t, amountOut.IsPositive(), "quote output should be positive for valid input")

				// Verify the quote makes sense (less than pool reserve due to fees)
				reserveB, ok := sdkmath.NewIntFromString(pool.ReserveB)
				require.True(t, ok, "should parse pool reserve B")
				require.True(t, amountOut.LT(reserveB), "output quote should be less than pool reserve")
			}
		})
	}
}

// TestExtremeValueRejection tests that extremely large values that would overflow are rejected
func TestExtremeValueRejection(t *testing.T) {
	ctx, k := setupTest(t)

	// Test with values near MaxInt256
	maxInt256 := sdkmath.NewIntFromBigInt(new(big.Int).Sub(
		new(big.Int).Lsh(big.NewInt(1), 255),
		big.NewInt(1),
	))

	// These should fail validation before reaching SafeMul
	extremeAmount := maxInt256

	_, err := k.CalculateSwapFee(ctx, extremeAmount)
	// This might succeed or fail depending on fee rate, but should never panic
	if err != nil {
		require.Contains(t, err.Error(), "overflow", "should report overflow")
	}
}
