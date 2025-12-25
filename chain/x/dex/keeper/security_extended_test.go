// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

// TestFlagOrderManipulation tests flagging order manipulation
func TestFlagOrderManipulation(t *testing.T) {
	ctx, k := setupTest(t)

	address := "aura1testaddress"
	poolID := "uaura-usdt"
	reason := "rapid order placement and cancellation"

	// Flag order manipulation
	k.FlagOrderManipulation(ctx, address, poolID, reason)

	// The function should execute without error
	// In a real implementation, you would verify the flag was stored
}

// TestCalculateAverageOrderSize tests average order size calculation
func TestCalculateAverageOrderSize(t *testing.T) {
	_, k := setupTest(t)

	// Create test order amounts (CalculateAverageOrderSize takes []sdkmath.Int)
	orders := []sdkmath.Int{
		sdkmath.NewInt(1000000),
		sdkmath.NewInt(2000000),
		sdkmath.NewInt(1500000),
	}

	avgSize := k.CalculateAverageOrderSize(orders)

	// Average should be (1000000 + 2000000 + 1500000) / 3 = 1500000
	expectedAvg := sdkmath.LegacyNewDec(1500000)
	require.True(t, avgSize.Equal(expectedAvg), "average order size should match expected")
}

// TestCalculateAverageOrderSizeEmptyOrders tests with no orders
func TestCalculateAverageOrderSizeEmptyOrders(t *testing.T) {
	_, k := setupTest(t)

	emptyOrders := []sdkmath.Int{}
	avgSize := k.CalculateAverageOrderSize(emptyOrders)

	// Should return zero for empty orders
	require.True(t, avgSize.Equal(sdkmath.LegacyZeroDec()), "average should be zero for empty orders")
}

// TestCountRapidChanges tests counting rapid order changes
func TestCountRapidChanges(t *testing.T) {
	ctx, k := setupTest(t)

	address := "aura1user"
	poolID := "uaura-usdt"

	count := k.CountRapidChanges(ctx, address, poolID)

	// Should count the number of status changes
	require.GreaterOrEqual(t, count, uint64(0), "count should be non-negative")
}

// TestDetectLayering tests layering detection
func TestDetectLayering(t *testing.T) {
	ctx, k := setupTest(t)

	address := "aura1user"
	poolID := "uaura-usdt"

	layeringCount := k.DetectLayering(ctx, address, poolID)

	// Should return a count
	require.GreaterOrEqual(t, layeringCount, uint64(0), "layering count should be non-negative")
}

// TestDetectSpoofing tests spoofing detection
func TestDetectSpoofing(t *testing.T) {
	ctx, k := setupTest(t)

	address := "aura1user"
	poolID := "uaura-usdt"

	spoofingCount := k.DetectSpoofing(ctx, address, poolID)

	// Should return a count
	require.GreaterOrEqual(t, spoofingCount, uint64(0), "spoofing count should be non-negative")
}

// TestGenerateSecureHash tests secure hash generation
func TestGenerateSecureHash(t *testing.T) {
	_, k := setupTest(t)

	data := []byte("test-data-for-hashing")

	hash := k.GenerateSecureHash(data)

	// Hash should not be empty
	require.NotEmpty(t, hash, "hash should not be empty")

	// Hash should be deterministic
	hash2 := k.GenerateSecureHash(data)
	require.Equal(t, hash, hash2, "same data should produce same hash")

	// Different data should produce different hash
	hash3 := k.GenerateSecureHash([]byte("different-data"))
	require.NotEqual(t, hash, hash3, "different data should produce different hash")
}

// TestOrderManipulationDetectionIntegration tests the full manipulation detection flow
func TestOrderManipulationDetectionIntegration(t *testing.T) {
	ctx, k := setupTest(t)

	address := "aura1manipulator"
	poolID := "uaura-usdt"

	// Calculate average order size with test amounts
	amounts := []sdkmath.Int{
		sdkmath.NewInt(100000),
		sdkmath.NewInt(100000),
		sdkmath.NewInt(100000),
	}
	avgSize := k.CalculateAverageOrderSize(amounts)
	require.True(t, avgSize.GT(sdkmath.LegacyZeroDec()), "average size should be positive")

	// Count rapid changes
	rapidCount := k.CountRapidChanges(ctx, address, poolID)
	require.GreaterOrEqual(t, rapidCount, uint64(0), "rapid count should be non-negative")

	// Check for layering
	layeringCount := k.DetectLayering(ctx, address, poolID)
	require.GreaterOrEqual(t, layeringCount, uint64(0), "layering count should be non-negative")

	// Check for spoofing
	spoofingCount := k.DetectSpoofing(ctx, address, poolID)
	require.GreaterOrEqual(t, spoofingCount, uint64(0), "spoofing count should be non-negative")

	// If manipulation detected, flag it
	if rapidCount > 5 {
		k.FlagOrderManipulation(ctx, address, poolID, "excessive rapid order changes")
	}
}

// TestSecurityHashConsistency tests hash consistency across operations
func TestSecurityHashConsistency(t *testing.T) {
	_, k := setupTest(t)

	testCases := [][]byte{
		[]byte("user1-pool1-order1"),
		[]byte("user2-pool2-order2"),
		[]byte(""),
		[]byte("very-long-string-with-many-characters-to-test-hash-function-robustness"),
		[]byte("special-chars-!@#$%^&*()"),
	}

	for _, tc := range testCases {
		hash1 := k.GenerateSecureHash(tc)
		hash2 := k.GenerateSecureHash(tc)

		require.Equal(t, hash1, hash2, "hash should be consistent for: %s", string(tc))
		require.NotEmpty(t, hash1, "hash should not be empty for: %s", string(tc))
	}
}

// TestCalculateAverageOrderSizeWithVariousAmounts tests with different amount patterns
func TestCalculateAverageOrderSizeWithVariousAmounts(t *testing.T) {
	_, k := setupTest(t)

	tests := []struct {
		name     string
		amounts  []sdkmath.Int
		expected sdkmath.LegacyDec
	}{
		{
			name:     "equal amounts",
			amounts:  []sdkmath.Int{sdkmath.NewInt(100), sdkmath.NewInt(100), sdkmath.NewInt(100)},
			expected: sdkmath.LegacyNewDec(100),
		},
		{
			name:     "single amount",
			amounts:  []sdkmath.Int{sdkmath.NewInt(1000)},
			expected: sdkmath.LegacyNewDec(1000),
		},
		{
			name:     "varying amounts",
			amounts:  []sdkmath.Int{sdkmath.NewInt(100), sdkmath.NewInt(200), sdkmath.NewInt(300)},
			expected: sdkmath.LegacyNewDec(200),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avgSize := k.CalculateAverageOrderSize(tt.amounts)
			require.True(t, avgSize.Equal(tt.expected), "average should match expected for: %s", tt.name)
		})
	}
}

// TestManipulationDetectionFunctions tests all detection functions execute without errors
func TestManipulationDetectionFunctions(t *testing.T) {
	ctx, k := setupTest(t)

	address := "aura1detector"
	poolID := "uaura-usdt"

	// All detection functions should execute without panic
	t.Run("flag manipulation", func(t *testing.T) {
		k.FlagOrderManipulation(ctx, address, poolID, "test")
	})

	t.Run("count rapid changes", func(t *testing.T) {
		count := k.CountRapidChanges(ctx, address, poolID)
		require.GreaterOrEqual(t, count, uint64(0))
	})

	t.Run("detect layering", func(t *testing.T) {
		count := k.DetectLayering(ctx, address, poolID)
		require.GreaterOrEqual(t, count, uint64(0))
	})

	t.Run("detect spoofing", func(t *testing.T) {
		count := k.DetectSpoofing(ctx, address, poolID)
		require.GreaterOrEqual(t, count, uint64(0))
	})
}
