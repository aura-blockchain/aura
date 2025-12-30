// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// TestLastPriceKey tests last price key generation
func TestLastPriceKey(t *testing.T) {
	tests := []struct {
		name   string
		poolID string
	}{
		{"simple pool ID", "pool1"},
		{"longer pool ID", "aura-usdt-pool"},
		{"with numbers", "pool-123-456"},
		{"empty string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.LastPriceKey(tt.poolID)
			require.NotNil(t, key)
			require.True(t, bytes.HasPrefix(key, types.LastPricePrefix))
			require.Equal(t, append(types.LastPricePrefix, []byte(tt.poolID)...), key)
		})
	}
}

// TestSecurityKeyPrefixesUniqueness tests that all security key prefixes are unique
func TestSecurityKeyPrefixesUniqueness(t *testing.T) {
	prefixes := [][]byte{
		types.TradeBlockPrefix,
		types.TWAPPrefix,
		types.LiquidityBlockPrefix,
		types.LiquidityLockPrefix,
		types.TradeHistoryPrefix,
		types.PoolCreationPrefix,
		types.CircuitBreakerPrefix,
		types.WashTradePrefix,
		types.OrderManipulationPrefix,
		types.LastPricePrefix,
	}

	seen := make(map[string]bool)
	for _, prefix := range prefixes {
		key := string(prefix)
		require.False(t, seen[key], "duplicate security prefix found: %x", prefix)
		seen[key] = true
	}
}

// TestSecurityKeyPrefixesInRange tests that security prefixes are in expected range (0x10-0x19)
func TestSecurityKeyPrefixesInRange(t *testing.T) {
	prefixes := map[string][]byte{
		"TradeBlockPrefix":        types.TradeBlockPrefix,
		"TWAPPrefix":              types.TWAPPrefix,
		"LiquidityBlockPrefix":    types.LiquidityBlockPrefix,
		"LiquidityLockPrefix":     types.LiquidityLockPrefix,
		"TradeHistoryPrefix":      types.TradeHistoryPrefix,
		"PoolCreationPrefix":      types.PoolCreationPrefix,
		"CircuitBreakerPrefix":    types.CircuitBreakerPrefix,
		"WashTradePrefix":         types.WashTradePrefix,
		"OrderManipulationPrefix": types.OrderManipulationPrefix,
		"LastPricePrefix":         types.LastPricePrefix,
	}

	for name, prefix := range prefixes {
		require.Len(t, prefix, 1, "%s should be single byte", name)
		require.GreaterOrEqual(t, prefix[0], byte(0x10), "%s should be >= 0x10", name)
		require.LessOrEqual(t, prefix[0], byte(0x19), "%s should be <= 0x19", name)
	}
}

// TestTWAPKey_BlockHeightEdgeCases tests TWAP key with various block height values
func TestTWAPKey_BlockHeightEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		blockHeight int64
	}{
		{"zero height", 0},
		{"positive height", 1000},
		{"large height", 9999999999},
		{"negative height", -1},
		{"negative large", -9999999999},
		{"max int64", int64(^uint64(0) >> 1)},
	}

	poolID := "test-pool"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.TWAPKey(poolID, tt.blockHeight)
			require.NotNil(t, key)
			require.True(t, bytes.HasPrefix(key, types.TWAPPrefix))
		})
	}
}

// TestSecurityKeysWithSeparator tests that keys with separators work correctly
func TestSecurityKeysWithSeparator(t *testing.T) {
	address := "aura1abc123"
	poolID := "pool1"

	// Test TradeBlockKey
	tradeBlockKey := types.TradeBlockKey(address, poolID)
	require.Contains(t, string(tradeBlockKey), string([]byte{0x00}))

	// Test LiquidityBlockKey
	liquidityBlockKey := types.LiquidityBlockKey(address, poolID)
	require.Contains(t, string(liquidityBlockKey), string([]byte{0x00}))

	// Test LiquidityLockKey
	liquidityLockKey := types.LiquidityLockKey(address, poolID)
	require.Contains(t, string(liquidityLockKey), string([]byte{0x00}))

	// Test WashTradeKey
	washTradeKey := types.WashTradeKey(address, poolID)
	require.Contains(t, string(washTradeKey), string([]byte{0x00}))

	// Test OrderManipulationKey
	orderManipKey := types.OrderManipulationKey(address, poolID)
	require.Contains(t, string(orderManipKey), string([]byte{0x00}))
}

// TestFormatSecurityKey_MultipleVariations tests formatting security keys for logging
func TestFormatSecurityKey_MultipleVariations(t *testing.T) {
	tests := []struct {
		name   string
		prefix []byte
		parts  []string
	}{
		{"single part", types.TradeBlockPrefix, []string{"part1"}},
		{"multiple parts", types.TWAPPrefix, []string{"part1", "part2", "part3"}},
		{"empty parts", types.CircuitBreakerPrefix, []string{}},
		{"empty string part", types.WashTradePrefix, []string{""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := types.FormatSecurityKey(tt.prefix, tt.parts...)
			require.NotEmpty(t, result)
			require.Contains(t, result, "security/")
		})
	}
}

// TestSecurityKeysDeterministic tests that security key functions are deterministic
func TestSecurityKeysDeterministic(t *testing.T) {
	address := "aura1test123"
	poolID := "test-pool"
	blockHeight := int64(12345)

	// TradeBlockKey
	for i := 0; i < 5; i++ {
		key1 := types.TradeBlockKey(address, poolID)
		key2 := types.TradeBlockKey(address, poolID)
		require.Equal(t, key1, key2)
	}

	// TWAPKey
	for i := 0; i < 5; i++ {
		key1 := types.TWAPKey(poolID, blockHeight)
		key2 := types.TWAPKey(poolID, blockHeight)
		require.Equal(t, key1, key2)
	}

	// LiquidityBlockKey
	for i := 0; i < 5; i++ {
		key1 := types.LiquidityBlockKey(address, poolID)
		key2 := types.LiquidityBlockKey(address, poolID)
		require.Equal(t, key1, key2)
	}

	// WashTradeKey
	for i := 0; i < 5; i++ {
		key1 := types.WashTradeKey(address, poolID)
		key2 := types.WashTradeKey(address, poolID)
		require.Equal(t, key1, key2)
	}

	// LastPriceKey
	for i := 0; i < 5; i++ {
		key1 := types.LastPriceKey(poolID)
		key2 := types.LastPriceKey(poolID)
		require.Equal(t, key1, key2)
	}
}

// TestSecurityKeysUniqueness tests that different inputs produce different keys
func TestSecurityKeysUniqueness(t *testing.T) {
	// Different addresses should produce different keys
	key1 := types.TradeBlockKey("aura1abc", "pool1")
	key2 := types.TradeBlockKey("aura1def", "pool1")
	require.NotEqual(t, key1, key2)

	// Different pool IDs should produce different keys
	key3 := types.TradeBlockKey("aura1abc", "pool1")
	key4 := types.TradeBlockKey("aura1abc", "pool2")
	require.NotEqual(t, key3, key4)

	// Different block heights should produce different TWAP keys
	key5 := types.TWAPKey("pool1", 100)
	key6 := types.TWAPKey("pool1", 200)
	require.NotEqual(t, key5, key6)
}

// TestSecurityKeyTypeSeparation tests that same data with different key types produces different keys
func TestSecurityKeyTypeSeparation(t *testing.T) {
	address := "aura1same"
	poolID := "pool-same"

	tradeBlockKey := types.TradeBlockKey(address, poolID)
	liquidityBlockKey := types.LiquidityBlockKey(address, poolID)
	liquidityLockKey := types.LiquidityLockKey(address, poolID)
	washTradeKey := types.WashTradeKey(address, poolID)
	orderManipKey := types.OrderManipulationKey(address, poolID)

	// All should be different
	keys := [][]byte{tradeBlockKey, liquidityBlockKey, liquidityLockKey, washTradeKey, orderManipKey}
	seen := make(map[string]bool)
	for _, key := range keys {
		keyStr := string(key)
		require.False(t, seen[keyStr], "duplicate key found")
		seen[keyStr] = true
	}
}
