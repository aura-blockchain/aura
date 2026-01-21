// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// TestSafeMul_HappyPath tests normal multiplication cases
func TestSafeMul_HappyPath(t *testing.T) {
	tests := []struct {
		name     string
		a        sdkmath.Int
		b        sdkmath.Int
		expected sdkmath.Int
	}{
		{
			name:     "small values",
			a:        sdkmath.NewInt(100),
			b:        sdkmath.NewInt(50),
			expected: sdkmath.NewInt(5000),
		},
		{
			name:     "zero first operand",
			a:        sdkmath.ZeroInt(),
			b:        sdkmath.NewInt(999999),
			expected: sdkmath.ZeroInt(),
		},
		{
			name:     "zero second operand",
			a:        sdkmath.NewInt(999999),
			b:        sdkmath.ZeroInt(),
			expected: sdkmath.ZeroInt(),
		},
		{
			name:     "identity multiplication",
			a:        sdkmath.NewInt(12345),
			b:        sdkmath.OneInt(),
			expected: sdkmath.NewInt(12345),
		},
		{
			name:     "large but safe values",
			a:        sdkmath.NewInt(1_000_000_000),
			b:        sdkmath.NewInt(1_000_000),
			expected: sdkmath.NewInt(1_000_000_000_000_000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := types.SafeMul(tt.a, tt.b)
			require.NoError(t, err)
			require.True(t, result.Equal(tt.expected), "expected %s, got %s", tt.expected, result)
		})
	}
}

// TestSafeMul_Overflow tests that overflow is detected and rejected
func TestSafeMul_Overflow(t *testing.T) {
	// MaxInt256 = 2^255 - 1
	maxInt256 := sdkmath.NewIntFromBigInt(new(big.Int).Sub(
		new(big.Int).Lsh(big.NewInt(1), 255),
		big.NewInt(1),
	))

	tests := []struct {
		name      string
		a         sdkmath.Int
		b         sdkmath.Int
		shouldErr bool
	}{
		{
			name:      "maxint256 * 2 overflows",
			a:         maxInt256,
			b:         sdkmath.NewInt(2),
			shouldErr: true,
		},
		{
			name:      "large * large overflows",
			a:         sdkmath.NewIntFromBigInt(new(big.Int).Lsh(big.NewInt(1), 200)),
			b:         sdkmath.NewIntFromBigInt(new(big.Int).Lsh(big.NewInt(1), 100)),
			shouldErr: true,
		},
		{
			name: "maxint256 / 2 * 3 overflows",
			a:    sdkmath.NewIntFromBigInt(new(big.Int).Div(maxInt256.BigInt(), big.NewInt(2))),
			b:    sdkmath.NewInt(3),
			// This should overflow since (MaxInt256/2) * 3 > MaxInt256
			shouldErr: true,
		},
		{
			name: "just under overflow is ok",
			a:    sdkmath.NewIntFromBigInt(new(big.Int).Div(maxInt256.BigInt(), big.NewInt(2))),
			b:    sdkmath.NewInt(2),
			// (MaxInt256/2) * 2 should be safe
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := types.SafeMul(tt.a, tt.b)
			if tt.shouldErr {
				require.Error(t, err, "expected overflow error")
				require.True(t, result.IsZero(), "result should be zero on error")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestSafeMul_NegativeInputs tests that negative inputs are rejected
func TestSafeMul_NegativeInputs(t *testing.T) {
	tests := []struct {
		name string
		a    sdkmath.Int
		b    sdkmath.Int
	}{
		{
			name: "negative first operand",
			a:    sdkmath.NewInt(-100),
			b:    sdkmath.NewInt(50),
		},
		{
			name: "negative second operand",
			a:    sdkmath.NewInt(100),
			b:    sdkmath.NewInt(-50),
		},
		{
			name: "both negative",
			a:    sdkmath.NewInt(-100),
			b:    sdkmath.NewInt(-50),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := types.SafeMul(tt.a, tt.b)
			require.Error(t, err, "expected error for negative input")
			require.True(t, result.IsZero(), "result should be zero on error")
		})
	}
}

// TestSafeAdd_HappyPath tests normal addition cases
func TestSafeAdd_HappyPath(t *testing.T) {
	tests := []struct {
		name     string
		a        sdkmath.Int
		b        sdkmath.Int
		expected sdkmath.Int
	}{
		{
			name:     "small values",
			a:        sdkmath.NewInt(100),
			b:        sdkmath.NewInt(50),
			expected: sdkmath.NewInt(150),
		},
		{
			name:     "zero first operand",
			a:        sdkmath.ZeroInt(),
			b:        sdkmath.NewInt(999),
			expected: sdkmath.NewInt(999),
		},
		{
			name:     "zero second operand",
			a:        sdkmath.NewInt(999),
			b:        sdkmath.ZeroInt(),
			expected: sdkmath.NewInt(999),
		},
		{
			name:     "large but safe values",
			a:        sdkmath.NewInt(1_000_000_000_000),
			b:        sdkmath.NewInt(2_000_000_000_000),
			expected: sdkmath.NewInt(3_000_000_000_000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := types.SafeAdd(tt.a, tt.b)
			require.NoError(t, err)
			require.True(t, result.Equal(tt.expected), "expected %s, got %s", tt.expected, result)
		})
	}
}

// TestSafeAdd_Overflow tests that overflow is detected
func TestSafeAdd_Overflow(t *testing.T) {
	maxInt256 := sdkmath.NewIntFromBigInt(new(big.Int).Sub(
		new(big.Int).Lsh(big.NewInt(1), 255),
		big.NewInt(1),
	))

	tests := []struct {
		name      string
		a         sdkmath.Int
		b         sdkmath.Int
		shouldErr bool
	}{
		{
			name:      "maxint256 + 1 overflows",
			a:         maxInt256,
			b:         sdkmath.OneInt(),
			shouldErr: true,
		},
		{
			name:      "maxint256 + maxint256 overflows",
			a:         maxInt256,
			b:         maxInt256,
			shouldErr: true,
		},
		{
			name:      "maxint256 + 0 is safe",
			a:         maxInt256,
			b:         sdkmath.ZeroInt(),
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := types.SafeAdd(tt.a, tt.b)
			if tt.shouldErr {
				require.Error(t, err, "expected overflow error")
				require.True(t, result.IsZero(), "result should be zero on error")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestSafeAdd_NegativeInputs tests that negative inputs are rejected
func TestSafeAdd_NegativeInputs(t *testing.T) {
	tests := []struct {
		name string
		a    sdkmath.Int
		b    sdkmath.Int
	}{
		{
			name: "negative first operand",
			a:    sdkmath.NewInt(-100),
			b:    sdkmath.NewInt(50),
		},
		{
			name: "negative second operand",
			a:    sdkmath.NewInt(100),
			b:    sdkmath.NewInt(-50),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := types.SafeAdd(tt.a, tt.b)
			require.Error(t, err, "expected error for negative input")
			require.True(t, result.IsZero(), "result should be zero on error")
		})
	}
}

// TestSafeMulDec_HappyPath tests normal Dec multiplication
func TestSafeMulDec_HappyPath(t *testing.T) {
	tests := []struct {
		name     string
		amount   sdkmath.Int
		rate     sdkmath.LegacyDec
		expected sdkmath.Int
	}{
		{
			name:     "0.3% fee on 1000",
			amount:   sdkmath.NewInt(1000),
			rate:     sdkmath.LegacyNewDecWithPrec(3, 3), // 0.003
			expected: sdkmath.NewInt(3),
		},
		{
			name:     "10% on 100",
			amount:   sdkmath.NewInt(100),
			rate:     sdkmath.LegacyNewDecWithPrec(1, 1), // 0.1
			expected: sdkmath.NewInt(10),
		},
		{
			name:     "zero amount",
			amount:   sdkmath.ZeroInt(),
			rate:     sdkmath.LegacyNewDecWithPrec(5, 2), // 0.05
			expected: sdkmath.ZeroInt(),
		},
		{
			name:     "zero rate",
			amount:   sdkmath.NewInt(1000),
			rate:     sdkmath.LegacyZeroDec(),
			expected: sdkmath.ZeroInt(),
		},
		{
			name:     "truncation test",
			amount:   sdkmath.NewInt(100),
			rate:     sdkmath.LegacyNewDecWithPrec(333, 5), // 0.00333
			expected: sdkmath.ZeroInt(),                    // 100 * 0.00333 = 0.333, truncates to 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := types.SafeMulDec(tt.amount, tt.rate)
			require.NoError(t, err)
			require.True(t, result.Equal(tt.expected), "expected %s, got %s", tt.expected, result)
		})
	}
}

// TestSafeMulDec_NegativeInputs tests rejection of negative inputs
func TestSafeMulDec_NegativeInputs(t *testing.T) {
	tests := []struct {
		name   string
		amount sdkmath.Int
		rate   sdkmath.LegacyDec
	}{
		{
			name:   "negative amount",
			amount: sdkmath.NewInt(-1000),
			rate:   sdkmath.LegacyNewDecWithPrec(3, 3),
		},
		{
			name:   "negative rate",
			amount: sdkmath.NewInt(1000),
			rate:   sdkmath.LegacyNewDec(-1),
		},
		{
			name:   "both negative",
			amount: sdkmath.NewInt(-1000),
			rate:   sdkmath.LegacyNewDec(-1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := types.SafeMulDec(tt.amount, tt.rate)
			require.Error(t, err, "expected error for negative input")
			require.True(t, result.IsZero(), "result should be zero on error")
		})
	}
}

// TestCheckPositive tests input validation
func TestCheckPositive(t *testing.T) {
	tests := []struct {
		name      string
		value     sdkmath.Int
		shouldErr bool
	}{
		{
			name:      "positive value",
			value:     sdkmath.NewInt(100),
			shouldErr: false,
		},
		{
			name:      "zero value",
			value:     sdkmath.ZeroInt(),
			shouldErr: true,
		},
		{
			name:      "negative value",
			value:     sdkmath.NewInt(-100),
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := types.CheckPositive(tt.value, "test_field")
			if tt.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestCheckNonNegative tests non-negative validation
func TestCheckNonNegative(t *testing.T) {
	tests := []struct {
		name      string
		value     sdkmath.Int
		shouldErr bool
	}{
		{
			name:      "positive value",
			value:     sdkmath.NewInt(100),
			shouldErr: false,
		},
		{
			name:      "zero value",
			value:     sdkmath.ZeroInt(),
			shouldErr: false,
		},
		{
			name:      "negative value",
			value:     sdkmath.NewInt(-100),
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := types.CheckNonNegative(tt.value, "test_field")
			if tt.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestSafeMul_FeeCalculationScenarios tests realistic fee calculation scenarios
func TestSafeMul_FeeCalculationScenarios(t *testing.T) {
	tests := []struct {
		name          string
		tradeAmount   sdkmath.Int
		boostBps      int64 // basis points (1 bps = 0.01%)
		expectSuccess bool
	}{
		{
			name:          "normal trade with 40% boost",
			tradeAmount:   sdkmath.NewInt(1_000_000_000), // 1B tokens
			boostBps:      4000,                          // 40%
			expectSuccess: true,
		},
		{
			name:          "huge trade amount",
			tradeAmount:   sdkmath.NewInt(1_000_000_000_000_000_000), // 1 quintillion
			boostBps:      4000,
			expectSuccess: true,
		},
		{
			name: "extremely large trade near maxint",
			// 2^240 * 4000 = 2^240 * 2^11.96... = 2^251.96 which is still < 2^255
			// This should actually succeed since 2^240 * 4000 < MaxInt256
			tradeAmount:   sdkmath.NewIntFromBigInt(new(big.Int).Lsh(big.NewInt(1), 240)),
			boostBps:      4000,
			expectSuccess: true, // Fixed: this actually doesn't overflow
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate fee boost calculation: amount * (boostBps / 10000)
			boostMultiplier := sdkmath.NewInt(tt.boostBps)
			result, err := types.SafeMul(tt.tradeAmount, boostMultiplier)

			if tt.expectSuccess {
				require.NoError(t, err, "expected successful calculation")
				// Verify result makes sense
				expected := tt.tradeAmount.Mul(boostMultiplier)
				require.True(t, result.Equal(expected))
			} else {
				require.Error(t, err, "expected overflow detection")
			}
		})
	}
}

// TestSafeMul_PoolInvariantCalculations tests k = x * y calculations
func TestSafeMul_PoolInvariantCalculations(t *testing.T) {
	tests := []struct {
		name          string
		reserveA      sdkmath.Int
		reserveB      sdkmath.Int
		expectSuccess bool
	}{
		{
			name:          "normal pool reserves",
			reserveA:      sdkmath.NewInt(1_000_000_000_000), // 1T
			reserveB:      sdkmath.NewInt(5_000_000_000_000), // 5T
			expectSuccess: true,
		},
		{
			name:          "large pool reserves",
			reserveA:      sdkmath.NewIntFromBigInt(new(big.Int).Lsh(big.NewInt(1), 100)),
			reserveB:      sdkmath.NewIntFromBigInt(new(big.Int).Lsh(big.NewInt(1), 100)),
			expectSuccess: true,
		},
		{
			name: "extremely large reserves that would overflow",
			// 2^200 * 2^200 = 2^400 which exceeds 2^255
			reserveA:      sdkmath.NewIntFromBigInt(new(big.Int).Lsh(big.NewInt(1), 200)),
			reserveB:      sdkmath.NewIntFromBigInt(new(big.Int).Lsh(big.NewInt(1), 200)),
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate constant product k = x * y
			k, err := types.SafeMul(tt.reserveA, tt.reserveB)

			if tt.expectSuccess {
				require.NoError(t, err, "expected successful k calculation")
				require.False(t, k.IsNegative(), "k should never be negative")
			} else {
				require.Error(t, err, "expected overflow detection")
			}
		})
	}
}
