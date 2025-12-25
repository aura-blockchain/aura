// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestStoreKeysCentralization verifies that the centralized store key system works correctly.
// This test ensures that:
// 1. StoreKeyNames() returns the correct list of keys
// 2. The keys match what's in the initStoreKeys() function
// 3. The AsMap() method returns all keys with correct mapping
func TestStoreKeysCentralization(t *testing.T) {
	// Get the list of store key names
	keyNames := StoreKeyNames()

	// Verify we have a reasonable number of keys (should have at least 30 modules)
	require.GreaterOrEqual(t, len(keyNames), 30, "Should have at least 30 store keys")

	// Initialize the keys
	keys := initStoreKeys()

	// Verify AsMap() returns all keys
	keyMap := keys.AsMap()
	require.Equal(t, len(keyNames), len(keyMap), "AsMap() should return same number of keys as Names()")

	// Verify all names in keyNames exist in the map
	for _, name := range keyNames {
		key, exists := keyMap[name]
		require.True(t, exists, "Key %s from Names() should exist in AsMap()", name)
		require.NotNil(t, key, "Key %s should not be nil", name)
		require.Equal(t, name, key.Name(), "Key name mismatch for %s", name)
	}

	// Verify no duplicate names
	uniqueNames := make(map[string]bool)
	for _, name := range keyNames {
		require.False(t, uniqueNames[name], "Duplicate key name: %s", name)
		uniqueNames[name] = true
	}
}

// TestAppInitializationWithCentralizedKeys verifies that the app can be created
// successfully with the centralized store key architecture.
func TestAppInitializationWithCentralizedKeys(t *testing.T) {
	app := NewAppWithOptions(nil, nil, "")
	require.NotNil(t, app, "App should be created successfully")
	require.NotNil(t, app.storeKeys, "Store keys should be initialized")

	// Verify critical keys exist
	require.NotNil(t, app.storeKeys.account, "Account key should exist")
	require.NotNil(t, app.storeKeys.bank, "Bank key should exist")
	require.NotNil(t, app.storeKeys.staking, "Staking key should exist")
	require.NotNil(t, app.storeKeys.consensus, "Consensus key should exist")

	// Verify AURA module keys exist
	require.NotNil(t, app.storeKeys.dex, "DEX key should exist")
	require.NotNil(t, app.storeKeys.bridge, "Bridge key should exist")
	require.NotNil(t, app.storeKeys.vc, "VC registry key should exist")
	require.NotNil(t, app.storeKeys.confidenceScore, "Confidence score key should exist")
	require.NotNil(t, app.storeKeys.inclusionRoutines, "Inclusion routines key should exist")

	// Verify security module keys exist
	require.NotNil(t, app.storeKeys.walletSecurity, "Wallet security key should exist")
	require.NotNil(t, app.storeKeys.validatorSecurity, "Validator security key should exist")
	require.NotNil(t, app.storeKeys.cryptography, "Cryptography key should exist")
	require.NotNil(t, app.storeKeys.networkSecurity, "Network security key should exist")
	require.NotNil(t, app.storeKeys.privacy, "Privacy key should exist")
	require.NotNil(t, app.storeKeys.security, "Security key should exist")

	// Verify identity module keys exist
	require.NotNil(t, app.storeKeys.identity, "Identity key should exist")
	require.NotNil(t, app.storeKeys.identityChange, "Identity change key should exist")

	// Verify economics module keys exist
	require.NotNil(t, app.storeKeys.economics, "Economics key should exist")
	require.NotNil(t, app.storeKeys.economicSecurity, "Economic security key should exist")
	require.NotNil(t, app.storeKeys.governance, "Governance key should exist")
}

// TestSingleSourceOfTruthConsistency verifies that the three uses of store keys
// are consistent with each other:
// 1. StoreKeyNames() function
// 2. initStoreKeys() struct fields
// 3. AsMap() method
func TestSingleSourceOfTruthConsistency(t *testing.T) {
	keyNames := StoreKeyNames()
	keys := initStoreKeys()
	keyMap := keys.AsMap()

	// All three should have the same count
	require.Equal(t, len(keyNames), len(keyMap),
		"Names() and AsMap() should have same number of keys")

	// Every name should map to a valid key
	for _, name := range keyNames {
		key, exists := keyMap[name]
		require.True(t, exists, "Key %s should exist in map", name)
		require.NotNil(t, key, "Key %s should not be nil", name)
		require.Equal(t, name, key.Name(), "Key name should match for %s", name)
	}

	// Every key in map should be in names
	nameSet := make(map[string]bool)
	for _, name := range keyNames {
		nameSet[name] = true
	}
	for mapName := range keyMap {
		require.True(t, nameSet[mapName], "Key %s in map should be in Names()", mapName)
	}
}
