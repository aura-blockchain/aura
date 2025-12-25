// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"testing"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// AssertEventEmitted checks if a specific event type was emitted
func AssertEventEmitted(t *testing.T, ctx sdk.Context, eventType string) {
	t.Helper()
	events := ctx.EventManager().Events()
	found := false
	for _, event := range events {
		if event.Type == eventType {
			found = true
			break
		}
	}
	require.True(t, found, "Event %s not emitted", eventType)
}

// AssertEventNotEmitted checks if a specific event type was not emitted
func AssertEventNotEmitted(t *testing.T, ctx sdk.Context, eventType string) {
	t.Helper()
	events := ctx.EventManager().Events()
	for _, event := range events {
		require.NotEqual(t, eventType, event.Type, "Event %s should not be emitted", eventType)
	}
}

// AssertEventAttribute checks if an event has a specific attribute
func AssertEventAttribute(t *testing.T, ctx sdk.Context, eventType, key, value string) {
	t.Helper()
	events := ctx.EventManager().Events()
	found := false
	for _, event := range events {
		if event.Type == eventType {
			for _, attr := range event.Attributes {
				if attr.Key == key && attr.Value == value {
					found = true
					break
				}
			}
		}
	}
	require.True(t, found, "Event %s with attribute %s=%s not found", eventType, key, value)
}

// AssertBalanceEqual checks if an account balance equals expected amount
func AssertBalanceEqual(t *testing.T, expected, actual sdk.Coins) {
	t.Helper()
	require.True(t, expected.Equal(actual), "Expected balance %s, got %s", expected, actual)
}

// AssertBalanceChanged checks if balance changed by expected amount
func AssertBalanceChanged(t *testing.T, before, after sdk.Coins, delta sdk.Coins) {
	t.Helper()
	expected := before.Add(delta...)
	require.True(t, expected.Equal(after), "Expected balance change from %s to %s, got %s", before, expected, after)
}

// AssertCoinsEqual checks if two coin amounts are equal
func AssertCoinsEqual(t *testing.T, expected, actual sdk.Coins) {
	t.Helper()
	require.True(t, expected.Equal(actual), "Expected %s, got %s", expected, actual)
}

// AssertCoinEqual checks if two coin amounts are equal
func AssertCoinEqual(t *testing.T, expected, actual sdk.Coin) {
	t.Helper()
	require.True(t, expected.Equal(actual), "Expected %s, got %s", expected, actual)
}

// AssertAddressEqual checks if two addresses are equal
func AssertAddressEqual(t *testing.T, expected, actual sdk.AccAddress) {
	t.Helper()
	require.Equal(t, expected.String(), actual.String())
}

// AssertValidatorAddressEqual checks if two validator addresses are equal
func AssertValidatorAddressEqual(t *testing.T, expected, actual sdk.ValAddress) {
	t.Helper()
	require.Equal(t, expected.String(), actual.String())
}

// AssertStoreHasKey checks if a KVStore has a specific key
func AssertStoreHasKey(t *testing.T, store storetypes.KVStore, key []byte) {
	t.Helper()
	require.True(t, store.Has(key), "Store should have key %x", key)
}

// AssertStoreNotHasKey checks if a KVStore does not have a specific key
func AssertStoreNotHasKey(t *testing.T, store storetypes.KVStore, key []byte) {
	t.Helper()
	require.False(t, store.Has(key), "Store should not have key %x", key)
}

// AssertStoreValue checks if a KVStore has a specific key-value pair
func AssertStoreValue(t *testing.T, store storetypes.KVStore, key, expectedValue []byte) {
	t.Helper()
	actualValue := store.Get(key)
	require.Equal(t, expectedValue, actualValue, "Store value mismatch for key %x", key)
}

// AssertPanic checks if a function panics
func AssertPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic but got none")
		}
	}()
	f()
}

// AssertNoPanic checks if a function does not panic
func AssertNoPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()
	f()
}

// AssertValidatorJailed checks if a validator is jailed
func AssertValidatorJailed(t *testing.T, jailed bool, expected bool) {
	t.Helper()
	require.Equal(t, expected, jailed, "Validator jail status mismatch")
}

// AssertValidatorSlashed checks validator slash amount
func AssertValidatorSlashed(t *testing.T, beforeTokens, afterTokens math.Int, minSlashPercent math.LegacyDec) {
	t.Helper()
	slashed := beforeTokens.Sub(afterTokens)
	minSlash := math.LegacyNewDecFromInt(beforeTokens).Mul(minSlashPercent).TruncateInt()
	require.True(t, slashed.GTE(minSlash), "Validator should be slashed by at least %s, got %s", minSlash, slashed)
}

// AssertErrorContains checks if error contains expected message
func AssertErrorContains(t *testing.T, err error, msg string) {
	t.Helper()
	require.Error(t, err)
	require.Contains(t, err.Error(), msg)
}

// AssertNoErrorOrLog logs error if present but doesn't fail
func AssertNoErrorOrLog(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Logf("Warning: %v", err)
	}
}
