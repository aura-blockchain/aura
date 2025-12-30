// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// TestSwapStatsKey tests swap stats key generation
func TestSwapStatsKey(t *testing.T) {
	tests := []struct {
		name   string
		poolID string
	}{
		{"simple pool ID", "pool1"},
		{"longer pool ID", "aura-usdt-pool-123"},
		{"with special chars", "pool_v2-main"},
		{"empty string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.SwapStatsKey(tt.poolID)
			require.NotNil(t, key)
			require.True(t, bytes.HasPrefix(key, types.SwapStatsPrefix))
			require.Equal(t, append(types.SwapStatsPrefix, []byte(tt.poolID)...), key)
		})
	}
}

// TestMarketPriceKey tests market price key generation
func TestMarketPriceKey(t *testing.T) {
	tests := []struct {
		name   string
		poolID string
	}{
		{"simple pool ID", "pool1"},
		{"longer pool ID", "aura-eth-pool"},
		{"numeric pool ID", "12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.MarketPriceKey(tt.poolID)
			require.NotNil(t, key)
			require.True(t, bytes.HasPrefix(key, types.MarketPricePrefix))
			require.Equal(t, append(types.MarketPricePrefix, []byte(tt.poolID)...), key)
		})
	}
}

// TestOrderCommitmentKey tests order commitment key generation
func TestOrderCommitmentKey(t *testing.T) {
	tests := []struct {
		name     string
		commitID string
	}{
		{"simple commit ID", "commit1"},
		{"uuid-like commit ID", "550e8400-e29b-41d4-a716-446655440000"},
		{"hash-like commit ID", "abc123def456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.OrderCommitmentKey(tt.commitID)
			require.NotNil(t, key)
			require.True(t, bytes.HasPrefix(key, types.OrderCommitmentPrefix))
			require.Equal(t, append(types.OrderCommitmentPrefix, []byte(tt.commitID)...), key)
		})
	}
}

// TestQueuedOrderKey tests queued order key generation
func TestQueuedOrderKey(t *testing.T) {
	tests := []struct {
		name    string
		orderID string
	}{
		{"simple order ID", "order1"},
		{"uuid order ID", "order-uuid-12345"},
		{"long order ID", "very-long-order-identifier-string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.QueuedOrderKey(tt.orderID)
			require.NotNil(t, key)
			require.True(t, bytes.HasPrefix(key, types.QueuedOrderPrefix))
			require.Equal(t, append(types.QueuedOrderPrefix, []byte(tt.orderID)...), key)
		})
	}
}

// TestOrderCleanupCursorKey tests order cleanup cursor key
func TestOrderCleanupCursorKey(t *testing.T) {
	key := types.OrderCleanupCursorKey()
	require.NotNil(t, key)
	require.Equal(t, types.OrderCleanupCursorPrefix, key)
}

// TestHTLCCleanupCursorKey tests HTLC cleanup cursor key
func TestHTLCCleanupCursorKey(t *testing.T) {
	key := types.HTLCCleanupCursorKey()
	require.NotNil(t, key)
	require.Equal(t, types.HTLCCleanupCursorPrefix, key)
}

// TestTWAPCursorKey tests TWAP cursor key
func TestTWAPCursorKey(t *testing.T) {
	key := types.TWAPCursorKey()
	require.NotNil(t, key)
	require.Equal(t, types.TWAPCursorPrefix, key)
}

// TestOrderExpirationKey tests order expiration key generation
func TestOrderExpirationKey(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt int64
		orderID   string
	}{
		{"normal expiration", 1700000000, "order1"},
		{"zero expiration", 0, "order2"},
		{"large timestamp", 9999999999, "order3"},
		{"negative timestamp", -1, "order4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.OrderExpirationKey(tt.expiresAt, tt.orderID)
			require.NotNil(t, key)
			require.True(t, bytes.HasPrefix(key, types.OrderExpirationPrefix))

			// Verify timestamp is encoded in big-endian
			expectedTsBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(expectedTsBytes, uint64(tt.expiresAt))
			expectedKey := append(types.OrderExpirationPrefix, expectedTsBytes...)
			expectedKey = append(expectedKey, []byte(tt.orderID)...)
			require.Equal(t, expectedKey, key)
		})
	}
}

// TestOrderExpirationKey_Sorting tests that expiration keys sort correctly
func TestOrderExpirationKey_Sorting(t *testing.T) {
	// Keys should sort by expiration time in ascending order
	key1 := types.OrderExpirationKey(1000, "order1")
	key2 := types.OrderExpirationKey(2000, "order2")
	key3 := types.OrderExpirationKey(3000, "order3")

	// byte comparison should maintain order
	require.True(t, bytes.Compare(key1, key2) < 0, "key1 should come before key2")
	require.True(t, bytes.Compare(key2, key3) < 0, "key2 should come before key3")
	require.True(t, bytes.Compare(key1, key3) < 0, "key1 should come before key3")
}

// TestOrderExpirationTimePrefix tests order expiration time prefix
func TestOrderExpirationTimePrefix(t *testing.T) {
	tests := []struct {
		name    string
		maxTime int64
	}{
		{"normal time", 1700000000},
		{"zero time", 0},
		{"large time", 9999999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix := types.OrderExpirationTimePrefix(tt.maxTime)
			require.NotNil(t, prefix)
			require.True(t, bytes.HasPrefix(prefix, types.OrderExpirationPrefix))

			// Should be exactly prefix + 8 bytes timestamp
			require.Equal(t, len(types.OrderExpirationPrefix)+8, len(prefix))
		})
	}
}

// TestOrderStatusKey tests order status key generation
func TestOrderStatusKey(t *testing.T) {
	tests := []struct {
		name    string
		status  types.SwapOrderStatus
		orderID string
	}{
		{"pending status", types.SwapOrderStatus_PENDING, "order1"},
		{"matched status", types.SwapOrderStatus_MATCHED, "order2"},
		{"htlc created status", types.SwapOrderStatus_HTLC_CREATED, "order3"},
		{"completed status", types.SwapOrderStatus_COMPLETED, "order4"},
		{"cancelled status", types.SwapOrderStatus_CANCELLED, "order5"},
		{"expired status", types.SwapOrderStatus_EXPIRED, "order6"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.OrderStatusKey(tt.status, tt.orderID)
			require.NotNil(t, key)
			require.True(t, bytes.HasPrefix(key, types.OrderStatusPrefix))

			// Should include status byte
			expectedKey := append(types.OrderStatusPrefix, byte(tt.status))
			expectedKey = append(expectedKey, []byte(tt.orderID)...)
			require.Equal(t, expectedKey, key)
		})
	}
}

// TestOrderStatusPrefixByStatus tests order status prefix by status
func TestOrderStatusPrefixByStatus(t *testing.T) {
	statuses := []types.SwapOrderStatus{
		types.SwapOrderStatus_PENDING,
		types.SwapOrderStatus_MATCHED,
		types.SwapOrderStatus_HTLC_CREATED,
		types.SwapOrderStatus_COMPLETED,
		types.SwapOrderStatus_CANCELLED,
		types.SwapOrderStatus_EXPIRED,
	}

	for _, status := range statuses {
		t.Run(status.String(), func(t *testing.T) {
			prefix := types.OrderStatusPrefixByStatus(status)
			require.NotNil(t, prefix)
			require.True(t, bytes.HasPrefix(prefix, types.OrderStatusPrefix))
			require.Equal(t, len(types.OrderStatusPrefix)+1, len(prefix))
			require.Equal(t, byte(status), prefix[len(prefix)-1])
		})
	}
}

// TestOrderStatusKey_UniquenessAcrossStatuses tests that same order ID with different statuses produces different keys
func TestOrderStatusKey_UniquenessAcrossStatuses(t *testing.T) {
	orderID := "same-order-id"
	keys := make(map[string]bool)

	statuses := []types.SwapOrderStatus{
		types.SwapOrderStatus_PENDING,
		types.SwapOrderStatus_MATCHED,
		types.SwapOrderStatus_HTLC_CREATED,
		types.SwapOrderStatus_COMPLETED,
		types.SwapOrderStatus_CANCELLED,
		types.SwapOrderStatus_EXPIRED,
	}

	for _, status := range statuses {
		key := types.OrderStatusKey(status, orderID)
		keyStr := string(key)
		require.False(t, keys[keyStr], "key should be unique for status %v", status)
		keys[keyStr] = true
	}
}

// TestSupportedCoinKey tests supported coin key generation
func TestSupportedCoinKey(t *testing.T) {
	tests := []struct {
		name  string
		denom string
	}{
		{"uaura", "uaura"},
		{"uusdt", "uusdt"},
		{"uatom", "uatom"},
		{"ibc denom", "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"},
		{"empty denom", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := types.SupportedCoinKey(tt.denom)
			require.NotNil(t, key)
			require.True(t, bytes.HasPrefix(key, types.SupportedCoinsPrefix))
			require.Equal(t, append(types.SupportedCoinsPrefix, []byte(tt.denom)...), key)
		})
	}
}

// TestKeyPrefixesUniqueness tests that all key prefixes are unique
func TestKeyPrefixesUniqueness(t *testing.T) {
	prefixes := [][]byte{
		types.PoolPrefix,
		types.OrderPrefix,
		types.OrderbookPrefix,
		types.HTLCPrefix,
		types.AtomicSwapPrefix,
		types.UserOrderPrefix,
		types.SwapStatsPrefix,
		types.MarketPricePrefix,
		types.OrderCommitmentPrefix,
		types.QueuedOrderPrefix,
		types.OrderCleanupCursorPrefix,
		types.HTLCCleanupCursorPrefix,
		types.TWAPCursorPrefix,
		types.OrderExpirationPrefix,
		types.OrderStatusPrefix,
		types.SupportedCoinsPrefix,
	}

	seen := make(map[string]bool)
	for _, prefix := range prefixes {
		key := string(prefix)
		require.False(t, seen[key], "duplicate prefix found: %x", prefix)
		seen[key] = true
	}
}

// TestKeyFunctionsReturnConsistentResults tests that key functions are deterministic
func TestKeyFunctionsReturnConsistentResults(t *testing.T) {
	// Call same function multiple times, should get same result
	poolID := "test-pool"
	for i := 0; i < 10; i++ {
		key1 := types.SwapStatsKey(poolID)
		key2 := types.SwapStatsKey(poolID)
		require.Equal(t, key1, key2)
	}

	orderID := "test-order"
	expiresAt := int64(1700000000)
	for i := 0; i < 10; i++ {
		key1 := types.OrderExpirationKey(expiresAt, orderID)
		key2 := types.OrderExpirationKey(expiresAt, orderID)
		require.Equal(t, key1, key2)
	}
}

// TestModuleConstants tests module constant values
func TestModuleConstants(t *testing.T) {
	require.Equal(t, "dex", types.ModuleName)
	require.Equal(t, "dex", types.StoreKey)
	require.Equal(t, "dex", types.RouterKey)
	require.Equal(t, "dex", types.QuerierRoute)

	// EndBlocker limits
	require.Equal(t, 100, types.MaxOrdersCleanupPerBlock)
	require.Equal(t, 50, types.MaxHTLCsCleanupPerBlock)
	require.Equal(t, 20, types.MaxPoolsTWAPPerBlock)
	require.Equal(t, 100, types.MaxBatchExecutionSize)
}
