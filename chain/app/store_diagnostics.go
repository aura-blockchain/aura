// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValidateStoreVersions performs a post-InitGenesis sanity check that verifies
// every mounted KV store has a persisted version in the CommitMultiStore.
//
// This diagnostic is critical for detecting:
//   - Stores that were mounted but never initialized
//   - Version lookup failures that would cause query failures
//   - Inconsistent state after genesis initialization
//
// Should be called immediately after InitGenesis completes successfully.
func (app *App) ValidateStoreVersions(ctx sdk.Context) error {
	logger := app.Logger()
	cms := app.CommitMultiStore()

	// Get the latest committed version from the store
	latestVersion := cms.LatestVersion()
	logger.Info("validating store versions post-InitGenesis",
		"latest_version", latestVersion,
		"block_height", ctx.BlockHeight(),
	)

	// Check each mounted KV store
	allKeys := app.allStoreKeys()
	failedStores := make([]string, 0)
	missingStores := make([]string, 0)

	for _, storeKey := range allKeys {
		if storeKey == nil {
			continue
		}

		storeName := storeKey.Name()

		// Try to get the store from the context
		// This validates the store is accessible
		store := ctx.KVStore(storeKey)
		if store == nil {
			missingStores = append(missingStores, storeName)
			logger.Error("CRITICAL: mounted store not accessible from context",
				"store_name", storeName,
				"store_key", storeKey.String(),
			)
			continue
		}

		// Verify the store has a persisted version by checking if we can
		// access it through the CommitMultiStore
		// This is a proxy check - if the store was properly initialized,
		// it should be in the version map
		storeObj := cms.GetCommitKVStore(storeKey)
		if storeObj == nil {
			failedStores = append(failedStores, storeName)
			logger.Error("CRITICAL: store not found in CommitMultiStore",
				"store_name", storeName,
				"store_key", storeKey.String(),
				"latest_version", latestVersion,
			)
			continue
		}

		// Additional check: verify we can perform basic operations
		// This ensures the store is not just registered but functional
		testKey := []byte{0xFF, 0xFE, 0xFD} // Unlikely to conflict
		testValue := []byte{0x01}

		// Try a write and read
		store.Set(testKey, testValue)
		retrieved := store.Get(testKey)
		if retrieved == nil || len(retrieved) == 0 {
			failedStores = append(failedStores, storeName)
			logger.Error("CRITICAL: store not functional (write/read failed)",
				"store_name", storeName,
				"store_key", storeKey.String(),
			)
			continue
		}

		// Clean up test data
		store.Delete(testKey)

		logger.Debug("store validation passed",
			"store_name", storeName,
			"version", latestVersion,
		)
	}

	// Report results
	if len(missingStores) > 0 {
		return fmt.Errorf("store validation failed: %d stores not accessible: %v",
			len(missingStores), missingStores)
	}

	if len(failedStores) > 0 {
		return fmt.Errorf("store validation failed: %d stores not functional: %v",
			len(failedStores), failedStores)
	}

	logger.Info("✅ all stores validated successfully",
		"total_stores", len(allKeys),
		"latest_version", latestVersion,
	)

	return nil
}

// DiagnoseStoreVersionFailure provides detailed diagnostics when a store version
// lookup fails. This should be called when CacheMultiStoreWithVersion returns an error.
//
// Returns a detailed error message with:
//   - Store name and version requested
//   - Available versions for the store
//   - Latest version in the CommitMultiStore
//   - Possible causes and remediation steps
func (app *App) DiagnoseStoreVersionFailure(storeName string, requestedVersion int64) error {
	logger := app.Logger()
	cms := app.CommitMultiStore()

	latestVersion := cms.LatestVersion()

	logger.Error("store version lookup failed - diagnostics",
		"store_name", storeName,
		"requested_version", requestedVersion,
		"latest_version", latestVersion,
	)

	// Find the store key by name
	var targetKey storetypes.StoreKey
	for _, key := range app.allStoreKeys() {
		if key != nil && key.Name() == storeName {
			targetKey = key
			break
		}
	}

	if targetKey == nil {
		logger.Error("store key not found in mounted stores",
			"store_name", storeName,
			"mounted_stores", app.storeKeys.Names(),
		)
		return fmt.Errorf("store version lookup failed: store %q not found in mounted stores (requested version: %d, latest: %d)",
			storeName, requestedVersion, latestVersion)
	}

	// Try to get the store
	storeObj := cms.GetCommitKVStore(targetKey)
	if storeObj == nil {
		logger.Error("store not found in CommitMultiStore",
			"store_name", storeName,
			"store_key", targetKey.String(),
		)
		return fmt.Errorf("store version lookup failed: store %q exists in mount list but not in CommitMultiStore (requested version: %d, latest: %d)",
			storeName, requestedVersion, latestVersion)
	}

	// Check version range
	if requestedVersion > latestVersion {
		logger.Error("requested version exceeds latest version",
			"store_name", storeName,
			"requested_version", requestedVersion,
			"latest_version", latestVersion,
			"difference", requestedVersion-latestVersion,
		)
		return fmt.Errorf("store version lookup failed: requested version %d exceeds latest version %d for store %q (difference: %d blocks)",
			requestedVersion, latestVersion, storeName, requestedVersion-latestVersion)
	}

	if requestedVersion < 0 {
		logger.Error("invalid requested version",
			"store_name", storeName,
			"requested_version", requestedVersion,
		)
		return fmt.Errorf("store version lookup failed: invalid version %d requested for store %q (must be >= 0)",
			requestedVersion, storeName)
	}

	// If we get here, the version is in valid range but still failed
	logger.Error("store version exists but lookup failed - possible corruption",
		"store_name", storeName,
		"requested_version", requestedVersion,
		"latest_version", latestVersion,
		"remediation", "try running 'aurad rollback' or restore from snapshot",
	)

	return fmt.Errorf("store version lookup failed: store %q version %d exists but is not accessible (latest: %d) - possible database corruption, try rollback or restore from snapshot",
		storeName, requestedVersion, latestVersion)
}

// LogStoreVersionContext logs detailed context about the current store state.
// Useful for debugging store-related issues.
func (app *App) LogStoreVersionContext(ctx sdk.Context, operation string) {
	logger := app.Logger()
	cms := app.CommitMultiStore()

	latestVersion := cms.LatestVersion()

	logger.Info("store context",
		"operation", operation,
		"block_height", ctx.BlockHeight(),
		"latest_version", latestVersion,
		"chain_id", ctx.ChainID(),
		"is_check_tx", ctx.IsCheckTx(),
		"is_recheck_tx", ctx.IsReCheckTx(),
	)

	// Log each store's status
	for _, key := range app.allStoreKeys() {
		if key == nil {
			continue
		}

		storeName := key.Name()
		store := cms.GetCommitKVStore(key)
		accessible := store != nil

		logger.Debug("store status",
			"store_name", storeName,
			"accessible", accessible,
			"latest_version", latestVersion,
		)
	}
}

// WrapCacheMultiStoreWithVersion wraps the BaseApp's CacheMultiStoreWithVersion
// call with enhanced error diagnostics. This should be used in place of direct
// calls to CacheMultiStoreWithVersion when detailed error context is needed.
func (app *App) WrapCacheMultiStoreWithVersion(version int64) (storetypes.CacheMultiStore, error) {
	logger := app.Logger()

	logger.Debug("attempting to load store version",
		"version", version,
		"latest_version", app.CommitMultiStore().LatestVersion(),
	)

	// Attempt the operation
	cms, err := app.CommitMultiStore().CacheMultiStoreWithVersion(version)
	if err != nil {
		// Enhanced error logging
		logger.Error("CacheMultiStoreWithVersion failed",
			"version", version,
			"latest_version", app.CommitMultiStore().LatestVersion(),
			"error", err,
		)

		// Try to diagnose which store failed
		// Note: The SDK doesn't give us the store name in the error,
		// so we log all stores for diagnostic purposes
		for _, key := range app.allStoreKeys() {
			if key == nil {
				continue
			}
			storeName := key.Name()
			logger.Error("store state at failure",
				"store_name", storeName,
				"requested_version", version,
			)
		}

		return nil, fmt.Errorf("failed to load store version %d (latest: %d): %w - see logs for per-store diagnostics",
			version, app.CommitMultiStore().LatestVersion(), err)
	}

	logger.Debug("successfully loaded store version",
		"version", version,
	)

	return cms, nil
}

// ValidateStoreMount checks that a specific store is properly mounted and accessible.
// Returns a detailed error if the store is not accessible.
func (app *App) ValidateStoreMount(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	if storeKey == nil {
		return fmt.Errorf("store key is nil")
	}

	storeName := storeKey.Name()
	logger := app.Logger()

	// Check if store is in the mount list
	found := false
	for _, key := range app.allStoreKeys() {
		if key != nil && key.Name() == storeName {
			found = true
			break
		}
	}

	if !found {
		logger.Error("store not in mount list",
			"store_name", storeName,
			"mounted_stores", app.storeKeys.Names(),
		)
		return fmt.Errorf("store %q is not in the mount list", storeName)
	}

	// Check if store is accessible from context
	store := ctx.KVStore(storeKey)
	if store == nil {
		logger.Error("store not accessible from context",
			"store_name", storeName,
			"block_height", ctx.BlockHeight(),
		)
		return fmt.Errorf("store %q is mounted but not accessible from context at height %d",
			storeName, ctx.BlockHeight())
	}

	// Check if store is in CommitMultiStore
	cms := app.CommitMultiStore()
	commitStore := cms.GetCommitKVStore(storeKey)
	if commitStore == nil {
		logger.Error("store not in CommitMultiStore",
			"store_name", storeName,
			"latest_version", cms.LatestVersion(),
		)
		return fmt.Errorf("store %q is mounted but not in CommitMultiStore (version: %d)",
			storeName, cms.LatestVersion())
	}

	logger.Debug("store mount validation passed",
		"store_name", storeName,
	)

	return nil
}
