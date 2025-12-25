// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestPoolKey(t *testing.T) {
	poolID := "pool123"
	key := PoolKey(poolID)

	if len(key) == 0 {
		t.Error("PoolKey should not return empty byte slice")
	}

	if !bytes.Contains(key, []byte(poolID)) {
		t.Error("PoolKey should contain poolID")
	}
}

func TestOrderKey(t *testing.T) {
	orderID := "order123"
	key := OrderKey(orderID)

	if len(key) == 0 {
		t.Error("OrderKey should not return empty byte slice")
	}

	if !bytes.Contains(key, []byte(orderID)) {
		t.Error("OrderKey should contain orderID")
	}
}

func TestOrderbookKey(t *testing.T) {
	pairKey := "uaura-uatom"
	orderID := "order123"
	key := OrderbookKey(pairKey, orderID)

	if len(key) == 0 {
		t.Error("OrderbookKey should not return empty byte slice")
	}

	if !bytes.Contains(key, []byte(pairKey)) {
		t.Error("OrderbookKey should contain pairKey")
	}

	if !bytes.Contains(key, []byte(orderID)) {
		t.Error("OrderbookKey should contain orderID")
	}
}

func TestOrderbookPairPrefix(t *testing.T) {
	pairKey := "uaura-uatom"
	prefix := OrderbookPairPrefix(pairKey)

	if len(prefix) == 0 {
		t.Error("OrderbookPairPrefix should not return empty byte slice")
	}

	// Verify it can be used as a prefix
	orderID := "order123"
	fullKey := OrderbookKey(pairKey, orderID)
	if !bytes.HasPrefix(fullKey, prefix) {
		t.Error("OrderbookKey should start with OrderbookPairPrefix")
	}
}

func TestHTLCKey(t *testing.T) {
	htlcID := "htlc123"
	key := HTLCKey(htlcID)

	if len(key) == 0 {
		t.Error("HTLCKey should not return empty byte slice")
	}

	if !bytes.Contains(key, []byte(htlcID)) {
		t.Error("HTLCKey should contain htlcID")
	}
}

func TestAtomicSwapKey(t *testing.T) {
	swapID := "swap123"
	key := AtomicSwapKey(swapID)

	if len(key) == 0 {
		t.Error("AtomicSwapKey should not return empty byte slice")
	}

	if !bytes.Contains(key, []byte(swapID)) {
		t.Error("AtomicSwapKey should contain swapID")
	}
}

func TestKeyUniqueness(t *testing.T) {
	// Test that different IDs produce different keys
	pool1 := PoolKey("pool1")
	pool2 := PoolKey("pool2")

	if bytes.Equal(pool1, pool2) {
		t.Error("Different pool IDs should produce different keys")
	}

	order1 := OrderKey("order1")
	order2 := OrderKey("order2")

	if bytes.Equal(order1, order2) {
		t.Error("Different order IDs should produce different keys")
	}
}

func TestKeyConsistency(t *testing.T) {
	// Test that same input produces same output
	poolID := "testpool"
	key1 := PoolKey(poolID)
	key2 := PoolKey(poolID)

	if !bytes.Equal(key1, key2) {
		t.Error("Same pool ID should produce consistent keys")
	}
}

func TestUserOrderKey(t *testing.T) {
	addr := "aura1xyz"
	orderID := "order123"
	ts := uint64(42)

	key := UserOrderKey(addr, ts, orderID)
	prefix := UserOrderAddressPrefix(addr)

	if !bytes.HasPrefix(key, prefix) {
		t.Fatal("user order key should start with address prefix")
	}

	if string(key[len(key)-len(orderID):]) != orderID {
		t.Fatal("order id should be included in key")
	}

	encoded := binary.BigEndian.Uint64(key[len(prefix) : len(prefix)+8])
	if encoded != ^ts {
		t.Fatal("timestamp should be inverted to preserve ordering")
	}
}
