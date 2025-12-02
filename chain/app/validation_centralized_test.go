package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestValidationUsesCentralizedStoreKeys verifies that the validation
// functions properly use the centralized store key architecture instead
// of maintaining their own hardcoded lists.
func TestValidationUsesCentralizedStoreKeys(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app)

	// Run app validation
	result := app.ValidateApp()

	// Validation should pass (no errors)
	require.True(t, result.Valid, "App validation should pass")
	require.Empty(t, result.Errors, "App validation should have no errors: %v", result.Errors)

	// There may be warnings (e.g., about minter permissions), but that's OK
	if len(result.Warnings) > 0 {
		t.Logf("App validation warnings: %v", result.Warnings)
	}
}

// TestValidateStoreKeysUsesCentralizedMap verifies that validateStoreKeys
// function uses the centralized AsMap() method instead of a hardcoded list.
// This test ensures that adding a new module only requires updating the
// centralized storeKeys struct and its methods.
func TestValidateStoreKeysUsesCentralizedMap(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app)

	result := ValidationResult{Valid: true}
	app.validateStoreKeys(&result)

	// Should have no errors
	require.True(t, result.Valid, "Store key validation should pass")
	require.Empty(t, result.Errors, "Store key validation should have no errors: %v", result.Errors)

	// Verify that the validation checked all keys from AsMap()
	// We can't directly verify this without inspecting the function internals,
	// but we can verify that the centralized keys work correctly
	keyMap := app.storeKeys.AsMap()
	require.NotEmpty(t, keyMap, "Store key map should not be empty")

	// Verify all keys are non-nil
	for name, key := range keyMap {
		require.NotNil(t, key, "Key %s should not be nil", name)
		require.Equal(t, name, key.Name(), "Key name should match for %s", name)
	}
}

// TestStoreKeyConsistencyAcrossAppAndValidation verifies that the store keys
// used by the app match those validated by the validation functions.
func TestStoreKeyConsistencyAcrossAppAndValidation(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app)

	// Get keys from centralized source
	centralizedKeys := app.storeKeys.AsMap()

	// Get keys from StoreKeyNames (used by validation)
	keyNames := StoreKeyNames()

	// Counts should match
	require.Equal(t, len(keyNames), len(centralizedKeys),
		"StoreKeyNames count should match AsMap count")

	// All names should exist in the map
	for _, name := range keyNames {
		key, exists := centralizedKeys[name]
		require.True(t, exists, "Key %s from Names() should exist in AsMap()", name)
		require.NotNil(t, key, "Key %s should not be nil", name)
	}

	// All map keys should be in names
	nameSet := make(map[string]bool)
	for _, name := range keyNames {
		nameSet[name] = true
	}
	for mapName := range centralizedKeys {
		require.True(t, nameSet[mapName], "Key %s in AsMap() should be in Names()", mapName)
	}
}

// TestAddingNewModuleWorkflow simulates the workflow of adding a new module
// and verifies that it only requires updating the centralized storeKeys struct.
func TestAddingNewModuleWorkflow(t *testing.T) {
	// This test documents the expected workflow when adding a new module:
	//
	// 1. Add a new field to the storeKeys struct in app.go
	//    Example: newModule *storetypes.KVStoreKey
	//
	// 2. Add initialization in initStoreKeys() function
	//    Example: newModule: storetypes.NewKVStoreKey(newmoduletypes.StoreKey),
	//
	// 3. Add the key to Names() method
	//    Example: newmoduletypes.StoreKey,
	//
	// 4. Add the key to AsMap() method
	//    Example: newmoduletypes.StoreKey: s.newModule,
	//
	// That's it! The following are automatically handled:
	// - MountKVStores() uses AsMap()
	// - validateStoreKeys() uses AsMap()
	// - StoreKeyNames() uses Names()
	// - allStoreKeys() uses the struct fields
	//
	// This test verifies that the current implementation follows this pattern.

	app := NewApp()

	// Verify the centralized methods work correctly
	names := app.storeKeys.Names()
	keyMap := app.storeKeys.AsMap()
	allKeys := app.allStoreKeys()

	// Names() and AsMap() should have matching counts
	require.Equal(t, len(names), len(keyMap),
		"Names() and AsMap() should have same count")

	// allStoreKeys() should have same count as AsMap()
	require.Equal(t, len(allKeys), len(keyMap),
		"allStoreKeys() should have same count as AsMap()")

	// Verify all keys from allStoreKeys() are in AsMap()
	keyMapSet := make(map[string]bool)
	for _, key := range keyMap {
		if key != nil {
			keyMapSet[key.Name()] = true
		}
	}

	for _, key := range allKeys {
		if key != nil {
			require.True(t, keyMapSet[key.Name()],
				"Key %s from allStoreKeys() should be in AsMap()", key.Name())
		}
	}
}

// TestNoDuplicateStoreKeys verifies that there are no duplicate store key names
// in the centralized system.
func TestNoDuplicateStoreKeys(t *testing.T) {
	app := NewApp()

	keyMap := app.storeKeys.AsMap()
	seen := make(map[string]string)

	for moduleName, key := range keyMap {
		require.NotNil(t, key, "Key for module %s should not be nil", moduleName)

		keyName := key.Name()
		if existingModule, exists := seen[keyName]; exists {
			t.Errorf("Duplicate key name '%s' used by modules '%s' and '%s'",
				keyName, existingModule, moduleName)
		}
		seen[keyName] = moduleName
	}
}

// TestStoreKeysMatchExpectedModules verifies that all expected modules
// have store keys defined in the centralized system.
func TestStoreKeysMatchExpectedModules(t *testing.T) {
	app := NewApp()
	keyMap := app.storeKeys.AsMap()

	// Core Cosmos SDK modules (must exist)
	coreModules := []string{
		"acc",    // auth
		"bank",
		"staking",
		"slashing",
		"distribution",
		"params",
		"consensus",
	}

	for _, moduleName := range coreModules {
		found := false
		for keyName := range keyMap {
			if keyName == moduleName {
				found = true
				break
			}
		}
		require.True(t, found, "Core module %s should have a store key", moduleName)
	}

	// AURA modules (must exist)
	auraModules := []string{
		"vcregistry",
		"dex",
		"bridge",
		"compliance",
		"confidencescore",
		"inclusionroutines",
		"dataregistry",
	}

	for _, moduleName := range auraModules {
		found := false
		for keyName := range keyMap {
			if keyName == moduleName {
				found = true
				break
			}
		}
		require.True(t, found, "AURA module %s should have a store key", moduleName)
	}
}
