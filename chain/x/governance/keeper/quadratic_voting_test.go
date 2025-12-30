// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/governance/types"
)

func setupQuadraticVotingKeeper(t *testing.T) (*Keeper, sdk.Context) {
	input := keepertest.CreateTestInputWithKeys(t, "governance")
	mockStaking := NewMockStakingKeeper()
	mockBank := &MockBankKeeper{
		balances:       make(map[string]sdk.Coins),
		moduleBalances: make(map[string]sdk.Coins),
	}
	mockSecurity := &MockSecurityKeeper{}
	keeper := NewKeeper(input.Cdc, input.StoreKey, mockStaking, mockBank, mockSecurity)
	ctx := input.Ctx.WithKVGasConfig(storetypes.GasConfig{})
	keeper.SetParams(ctx, types.DefaultParams())
	return keeper, ctx
}

func testAddrQuadratic(name string) string {
	padded := name + "________________"
	return sdk.AccAddress(padded[:20]).String()
}

// TestCalculateQuadraticVotingPower tests the quadratic formula
func TestCalculateQuadraticVotingPower(t *testing.T) {
	keeper, _ := setupQuadraticVotingKeeper(t)

	tests := []struct {
		name     string
		credits  uint64
		expected uint64
	}{
		{
			name:     "zero credits",
			credits:  0,
			expected: 0,
		},
		{
			name:     "one credit",
			credits:  1,
			expected: 10000,
		},
		{
			name:     "four credits",
			credits:  4,
			expected: 20000,
		},
		{
			name:     "nine credits",
			credits:  9,
			expected: 30000,
		},
		{
			name:     "sixteen credits",
			credits:  16,
			expected: 40000,
		},
		{
			name:     "hundred credits",
			credits:  100,
			expected: 100000,
		},
		{
			name:     "ten thousand credits",
			credits:  10000,
			expected: 1000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			power := keeper.calculateQuadraticVotingPower(tt.credits)
			require.Equal(t, tt.expected, power)
		})
	}
}

// TestIntSqrt tests the integer square root implementation
func TestIntSqrt(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected int64
	}{
		{
			name:     "sqrt of 0",
			input:    0,
			expected: 0,
		},
		{
			name:     "sqrt of 1",
			input:    1,
			expected: 1,
		},
		{
			name:     "sqrt of 4",
			input:    4,
			expected: 2,
		},
		{
			name:     "sqrt of 9",
			input:    9,
			expected: 3,
		},
		{
			name:     "sqrt of 16",
			input:    16,
			expected: 4,
		},
		{
			name:     "sqrt of 100",
			input:    100,
			expected: 10,
		},
		{
			name:     "sqrt of 144",
			input:    144,
			expected: 12,
		},
		{
			name:     "sqrt of 10000",
			input:    10000,
			expected: 100,
		},
		{
			name:     "sqrt of non-perfect square (15)",
			input:    15,
			expected: 3,
		},
		{
			name:     "sqrt of non-perfect square (99)",
			input:    99,
			expected: 9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := sdkmath.NewInt(tt.input)
			result := intSqrt(n)
			require.Equal(t, tt.expected, result.Int64())
		})
	}
}

// TestIntSqrt_Negative tests sqrt of negative numbers
func TestIntSqrt_Negative(t *testing.T) {
	n := sdkmath.NewInt(-100)
	result := intSqrt(n)
	require.Equal(t, int64(0), result.Int64())
}

// TestGetVoteCreditsPerToken tests the credits per token helper
func TestGetVoteCreditsPerToken(t *testing.T) {
	keeper, _ := setupQuadraticVotingKeeper(t)

	creditsPerToken := keeper.getVoteCreditsPerToken()
	require.Equal(t, uint64(100), creditsPerToken)
}

// TestIsQuadraticVotingEnabled tests the quadratic voting enabled check
func TestIsQuadraticVotingEnabled(t *testing.T) {
	keeper, ctx := setupQuadraticVotingKeeper(t)

	enabled := keeper.isQuadraticVotingEnabled(ctx)
	require.True(t, enabled)
}
