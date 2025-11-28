package types

import (
	"strings"
	"testing"
)

func TestIRStoreKey(t *testing.T) {
	irID := "ir123"
	key := IRStoreKey(irID)

	if key == "" {
		t.Error("IRStoreKey should not return empty string")
	}

	if !strings.Contains(key, irID) {
		t.Errorf("IRStoreKey should contain irID %s, got %s", irID, key)
	}
}

func TestPrerequisiteStoreKey(t *testing.T) {
	irID := "ir123"
	key := PrerequisiteStoreKey(irID)

	if key == "" {
		t.Error("PrerequisiteStoreKey should not return empty string")
	}

	if !strings.Contains(key, irID) {
		t.Errorf("PrerequisiteStoreKey should contain irID %s, got %s", irID, key)
	}
}

func TestRateLimitStoreKey(t *testing.T) {
	irID := "ir123"
	key := RateLimitStoreKey(irID)

	if key == "" {
		t.Error("RateLimitStoreKey should not return empty string")
	}

	if !strings.Contains(key, irID) {
		t.Errorf("RateLimitStoreKey should contain irID %s, got %s", irID, key)
	}
}

func TestRateLimitUsageKey(t *testing.T) {
	irID := "ir123"
	userAddr := "aura1test"
	timeWindow := "2023-01-01"
	key := RateLimitUsageKey(irID, userAddr, timeWindow)

	if key == "" {
		t.Error("RateLimitUsageKey should not return empty string")
	}

	if !strings.Contains(key, irID) || !strings.Contains(key, userAddr) || !strings.Contains(key, timeWindow) {
		t.Errorf("RateLimitUsageKey should contain irID, userAddr, and timeWindow, got %s", key)
	}
}

func TestKeyUniqueness(t *testing.T) {
	// Test that different IDs produce different keys
	ir1 := IRStoreKey("ir1")
	ir2 := IRStoreKey("ir2")

	if ir1 == ir2 {
		t.Error("Different IR IDs should produce different keys")
	}

	prereq1 := PrerequisiteStoreKey("ir1")
	prereq2 := PrerequisiteStoreKey("ir2")

	if prereq1 == prereq2 {
		t.Error("Different IR IDs should produce different prerequisite keys")
	}

	rateLimit1 := RateLimitStoreKey("ir1")
	rateLimit2 := RateLimitStoreKey("ir2")

	if rateLimit1 == rateLimit2 {
		t.Error("Different IR IDs should produce different rate limit keys")
	}

	usage1 := RateLimitUsageKey("ir1", "addr1", "2023-01-01")
	usage2 := RateLimitUsageKey("ir1", "addr2", "2023-01-01")

	if usage1 == usage2 {
		t.Error("Different addresses should produce different usage keys")
	}
}

func TestKeyConsistency(t *testing.T) {
	// Test that same input produces same output
	irID := "testir"
	key1 := IRStoreKey(irID)
	key2 := IRStoreKey(irID)

	if key1 != key2 {
		t.Error("Same IR ID should produce consistent keys")
	}

	userAddr := "aura1test"
	usage1 := RateLimitUsageKey(irID, userAddr, "2023-01-01")
	usage2 := RateLimitUsageKey(irID, userAddr, "2023-01-01")

	if usage1 != usage2 {
		t.Error("Same inputs should produce consistent usage keys")
	}
}

func TestKeyPrefixes(t *testing.T) {
	// Test that different key types don't collide
	irID := "ir123"

	irKey := IRStoreKey(irID)
	prereqKey := PrerequisiteStoreKey(irID)
	rateLimitKey := RateLimitStoreKey(irID)

	if irKey == prereqKey {
		t.Error("IR key and prerequisite key should be different")
	}

	if irKey == rateLimitKey {
		t.Error("IR key and rate limit key should be different")
	}

	if prereqKey == rateLimitKey {
		t.Error("Prerequisite key and rate limit key should be different")
	}
}
