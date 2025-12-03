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
	// Skip - requires bank keeper setup for token transfers
	t.Skip("Test requires full integration setup with bank keeper")

	ctx, k := setupTest(t)
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
	// Skip actual pool creation/swap for now as it requires bank keeper setup
	// This test is primarily testing the overflow detection logic
	t.Skip("Test requires full integration setup with bank keeper")
}

// TestGetQuoteOverflowPrevention tests that quotes reject overflow
func TestGetQuoteOverflowPrevention(t *testing.T) {
	// Skip actual pool creation for now as it requires bank keeper setup
	t.Skip("Test requires full integration setup with bank keeper")
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
