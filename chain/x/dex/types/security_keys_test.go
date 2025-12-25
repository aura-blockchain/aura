// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"bytes"
	"testing"
)

func TestTradeBlockKey(t *testing.T) {
	addr := "aura1test"
	poolID := "pool123"

	key := TradeBlockKey(addr, poolID)

	if len(key) == 0 {
		t.Error("TradeBlockKey should not return empty byte slice")
	}

	if !bytes.Contains(key, []byte(addr)) {
		t.Error("TradeBlockKey should contain address")
	}
}

func TestTWAPKey(t *testing.T) {
	poolID := "pool123"
	blockHeight := int64(1000)

	key := TWAPKey(poolID, blockHeight)

	if len(key) == 0 {
		t.Error("TWAPKey should not return empty byte slice")
	}

	if !bytes.Contains(key, []byte(poolID)) {
		t.Error("TWAPKey should contain poolID")
	}
}

func TestTWAPPrefixKey(t *testing.T) {
	poolID := "pool123"

	prefix := TWAPPrefixKey(poolID)

	if len(prefix) == 0 {
		t.Error("TWAPPrefixKey should not return empty byte slice")
	}

	// Test that TWAP key starts with prefix
	twapKey := TWAPKey(poolID, 1000)
	if !bytes.HasPrefix(twapKey, prefix) {
		t.Error("TWAPKey should start with TWAPPrefixKey")
	}
}

func TestLiquidityBlockKey(t *testing.T) {
	addr := "aura1test"
	poolID := "pool123"

	key := LiquidityBlockKey(addr, poolID)

	if len(key) == 0 {
		t.Error("LiquidityBlockKey should not return empty byte slice")
	}

	if !bytes.Contains(key, []byte(addr)) || !bytes.Contains(key, []byte(poolID)) {
		t.Error("LiquidityBlockKey should contain both address and poolID")
	}
}

func TestLiquidityLockKey(t *testing.T) {
	addr := "aura1test"
	poolID := "pool123"

	key := LiquidityLockKey(addr, poolID)

	if len(key) == 0 {
		t.Error("LiquidityLockKey should not return empty byte slice")
	}

	if !bytes.Contains(key, []byte(addr)) || !bytes.Contains(key, []byte(poolID)) {
		t.Error("LiquidityLockKey should contain both address and poolID")
	}
}

func TestTradeHistoryKey(t *testing.T) {
	addr := "aura1test"

	key := TradeHistoryKey(addr)

	if len(key) == 0 {
		t.Error("TradeHistoryKey should not return empty byte slice")
	}

	if !bytes.Contains(key, []byte(addr)) {
		t.Error("TradeHistoryKey should contain address")
	}
}

func TestPoolCreationKey(t *testing.T) {
	creator := "aura1creator"

	key := PoolCreationKey(creator)

	if len(key) == 0 {
		t.Error("PoolCreationKey should not return empty byte slice")
	}

	if !bytes.Contains(key, []byte(creator)) {
		t.Error("PoolCreationKey should contain creator")
	}
}

func TestCircuitBreakerKey(t *testing.T) {
	key := CircuitBreakerKey()

	if len(key) == 0 {
		t.Error("CircuitBreakerKey should not return empty byte slice")
	}

	// Verify that the key equals the circuit breaker prefix
	if !bytes.Equal(key, CircuitBreakerPrefix) {
		t.Error("CircuitBreakerKey should equal CircuitBreakerPrefix")
	}
}

func TestWashTradeKey(t *testing.T) {
	addr := "aura1test"
	poolID := "pool123"

	key := WashTradeKey(addr, poolID)

	if len(key) == 0 {
		t.Error("WashTradeKey should not return empty byte slice")
	}

	if !bytes.Contains(key, []byte(addr)) || !bytes.Contains(key, []byte(poolID)) {
		t.Error("WashTradeKey should contain both address and poolID")
	}
}

func TestOrderManipulationKey(t *testing.T) {
	addr := "aura1test"
	poolID := "pool123"

	key := OrderManipulationKey(addr, poolID)

	if len(key) == 0 {
		t.Error("OrderManipulationKey should not return empty byte slice")
	}

	if !bytes.Contains(key, []byte(addr)) || !bytes.Contains(key, []byte(poolID)) {
		t.Error("OrderManipulationKey should contain both address and poolID")
	}
}

func TestFormatSecurityKey(t *testing.T) {
	prefix := []byte{0x01}
	component1 := "comp1"
	component2 := "comp2"

	// FormatSecurityKey returns a string, not []byte
	keyStr := FormatSecurityKey(prefix, component1, component2)
	keyBytes := []byte(keyStr)

	if len(keyStr) == 0 {
		t.Error("FormatSecurityKey should not return empty string")
	}

	// Check if string contains the hex representation of the prefix
	if !bytes.Contains(keyBytes, []byte("01")) {
		t.Error("FormatSecurityKey should contain hex representation of prefix")
	}

	if !bytes.Contains(keyBytes, []byte(component1)) {
		t.Error("FormatSecurityKey should contain component1")
	}

	if !bytes.Contains(keyBytes, []byte(component2)) {
		t.Error("FormatSecurityKey should contain component2")
	}
}

func TestSecurityKeyUniqueness(t *testing.T) {
	// Test that different inputs produce different keys
	key1 := TradeHistoryKey("addr1")
	key2 := TradeHistoryKey("addr2")

	if bytes.Equal(key1, key2) {
		t.Error("Different addresses should produce different TradeHistory keys")
	}

	// CircuitBreakerKey takes no arguments, so test WashTradeKey for uniqueness instead
	cb1 := WashTradeKey("addr1", "pool1")
	cb2 := WashTradeKey("addr1", "pool2")

	if bytes.Equal(cb1, cb2) {
		t.Error("Different pool IDs should produce different keys")
	}
}

func TestSecurityKeyConsistency(t *testing.T) {
	// Test that same input produces same output
	addr := "aura1test"
	poolID := "pool123"

	key1 := WashTradeKey(addr, poolID)
	key2 := WashTradeKey(addr, poolID)

	if !bytes.Equal(key1, key2) {
		t.Error("Same inputs should produce consistent keys")
	}
}
