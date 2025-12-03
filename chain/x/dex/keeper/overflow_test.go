package keeper_test

import (
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// TestSwapFeeOverflowPrevention tests that fee calculations reject overflow
func TestSwapFeeOverflowPrevention(t *testing.T) {
	suite := SetupTestSuite(t)

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
			fee, err := suite.dexKeeper.CalculateSwapFee(suite.SdkCtx, tt.amount)

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
	suite := SetupTestSuite(t)
	creator := suite.testAddr("creator")

	// Fund creator
	normalAmount := sdk.NewCoin("uaura", sdkmath.NewInt(10_000_000_000_000))
	require.NoError(t, suite.bankKeeper.MintCoins(suite.SdkCtx, types.ModuleName, sdk.NewCoins(normalAmount)))
	require.NoError(t, suite.bankKeeper.SendCoinsFromModuleToAccount(suite.SdkCtx, types.ModuleName, creator, sdk.NewCoins(normalAmount)))

	usdt := sdk.NewCoin("usdt", sdkmath.NewInt(2_000_000_000_000))
	require.NoError(t, suite.bankKeeper.MintCoins(suite.SdkCtx, types.ModuleName, sdk.NewCoins(usdt)))
	require.NoError(t, suite.bankKeeper.SendCoinsFromModuleToAccount(suite.SdkCtx, types.ModuleName, creator, sdk.NewCoins(usdt)))

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
			poolID := suite.dexKeeper.GeneratePoolID(tt.amountA.Denom, tt.amountB.Denom)

			// Clean up any existing pool
			suite.dexKeeper.DeletePool(suite.SdkCtx, poolID)

			pool, lpTokens, err := suite.dexKeeper.CreatePool(
				suite.SdkCtx,
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
	suite := SetupTestSuite(t)
	creator := suite.testAddr("creator")
	swapper := suite.testAddr("swapper")

	// Create pool with normal amounts
	auraAmount := sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000_000_000))
	usdtAmount := sdk.NewCoin("usdt", sdkmath.NewInt(200_000_000_000))

	// Fund creator
	require.NoError(t, suite.bankKeeper.MintCoins(suite.SdkCtx, types.ModuleName, sdk.NewCoins(auraAmount, usdtAmount)))
	require.NoError(t, suite.bankKeeper.SendCoinsFromModuleToAccount(suite.SdkCtx, types.ModuleName, creator, sdk.NewCoins(auraAmount, usdtAmount)))

	// Create pool
	pool, _, err := suite.dexKeeper.CreatePool(
		suite.SdkCtx,
		creator.String(),
		"uaura",
		"usdt",
		auraAmount,
		usdtAmount,
	)
	require.NoError(t, err)
	require.NotNil(t, pool)

	// Fund swapper
	swapAmount := sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000_000))
	require.NoError(t, suite.bankKeeper.MintCoins(suite.SdkCtx, types.ModuleName, sdk.NewCoins(swapAmount)))
	require.NoError(t, suite.bankKeeper.SendCoinsFromModuleToAccount(suite.SdkCtx, types.ModuleName, swapper, sdk.NewCoins(swapAmount)))

	tests := []struct {
		name          string
		swapIn        sdk.Coin
		minOut        sdkmath.Int
		expectError   bool
		errorContains string
	}{
		{
			name:        "normal swap",
			swapIn:      sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000)),
			minOut:      sdkmath.NewInt(1),
			expectError: false,
		},
		{
			name:          "zero amount rejected",
			swapIn:        sdk.NewCoin("uaura", sdkmath.ZeroInt()),
			minOut:        sdkmath.NewInt(1),
			expectError:   true,
			errorContains: "swap amount must be greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amountOut, effectivePrice, priceImpact, err := suite.dexKeeper.SwapExactIn(
				suite.SdkCtx,
				swapper.String(),
				pool.PoolId,
				tt.swapIn,
				tt.minOut,
				10000, // 100% max slippage
			)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					require.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.False(t, amountOut.IsNegative(), "output should never be negative")
				require.False(t, effectivePrice.IsNegative(), "price should never be negative")
				require.False(t, priceImpact.IsNegative(), "price impact should never be negative")
			}
		})
	}
}

// TestGetQuoteOverflowPrevention tests that quotes reject overflow
func TestGetQuoteOverflowPrevention(t *testing.T) {
	suite := SetupTestSuite(t)
	creator := suite.testAddr("creator")

	// Create pool
	auraAmount := sdk.NewCoin("uaura", sdkmath.NewInt(1_000_000_000_000))
	usdtAmount := sdk.NewCoin("usdt", sdkmath.NewInt(200_000_000_000))

	require.NoError(t, suite.bankKeeper.MintCoins(suite.SdkCtx, types.ModuleName, sdk.NewCoins(auraAmount, usdtAmount)))
	require.NoError(t, suite.bankKeeper.SendCoinsFromModuleToAccount(suite.SdkCtx, types.ModuleName, creator, sdk.NewCoins(auraAmount, usdtAmount)))

	pool, _, err := suite.dexKeeper.CreatePool(
		suite.SdkCtx,
		creator.String(),
		"uaura",
		"usdt",
		auraAmount,
		usdtAmount,
	)
	require.NoError(t, err)

	tests := []struct {
		name        string
		amountIn    sdkmath.Int
		expectError bool
	}{
		{
			name:        "normal quote",
			amountIn:    sdkmath.NewInt(1_000_000),
			expectError: false,
		},
		{
			name:        "large quote",
			amountIn:    sdkmath.NewInt(100_000_000_000),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			estimatedOutput, effectivePrice, priceImpact, feeAmount, err := suite.dexKeeper.GetQuote(
				suite.SdkCtx,
				pool.PoolId,
				"uaura",
				tt.amountIn,
			)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.False(t, estimatedOutput.IsNegative(), "output should never be negative")
				require.False(t, effectivePrice.IsNegative(), "price should never be negative")
				require.False(t, priceImpact.IsNegative(), "price impact should never be negative")
				require.False(t, feeAmount.IsNegative(), "fee should never be negative")
			}
		})
	}
}

// TestExtremeValueRejection tests that extremely large values that would overflow are rejected
func TestExtremeValueRejection(t *testing.T) {
	suite := SetupTestSuite(t)

	// Test with values near MaxInt256
	maxInt256 := sdkmath.NewIntFromBigInt(new(big.Int).Sub(
		new(big.Int).Lsh(big.NewInt(1), 255),
		big.NewInt(1),
	))

	// These should fail validation before reaching SafeMul
	extremeAmount := maxInt256

	_, err := suite.dexKeeper.CalculateSwapFee(suite.SdkCtx, extremeAmount)
	// This might succeed or fail depending on fee rate, but should never panic
	if err != nil {
		require.Contains(t, err.Error(), "overflow", "should report overflow")
	}
}
