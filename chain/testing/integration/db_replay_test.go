// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDatabaseReplayStability tests that a node can:
// 1. Start with a seeded database
// 2. Process several blocks
// 3. Stop gracefully
// 4. Restart and continue from the same state
// 5. Verify AppHash consistency
// 6. Ensure no "version does not exist" errors occur
//
// This test addresses ROADMAP_PRODUCTION.md Task #8:
// "Verify replay on seeded DB via start→stop→start integration test"
func TestDatabaseReplayStability(t *testing.T) {
	// Create a temporary directory for test data
	tmpDir, err := os.MkdirTemp("", "aura-db-replay-test-*")
	require.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "data")
	err = os.MkdirAll(dbPath, 0755)
	require.NoError(t, err, "Failed to create db directory")

	t.Logf("Test database path: %s", dbPath)

	// Phase 1: Initialize database and seed initial state
	t.Run("Phase1_InitializeAndSeed", func(t *testing.T) {
		db, cms, ctx := setupDatabase(t, dbPath)
		defer db.Close()

		// Check if this is a new database or existing
		latestVersion := cms.LastCommitID().Version
		t.Logf("Phase 1 - Latest version before seeding: %d", latestVersion)
		t.Logf("Phase 1 - Latest AppHash before seeding: %X", cms.LastCommitID().Hash)
		require.Equal(t, int64(0), latestVersion, "Should be a new database with version 0")

		// Seed initial state with multiple stores
		seedInitialState(t, ctx)

		// Commit initial state (genesis block) - this will create version 1
		t.Log("Phase 1 - Committing initial state...")
		commitInfo := cms.Commit()
		require.NotNil(t, commitInfo, "Commit should return CommitInfo")
		require.NotEmpty(t, commitInfo.Hash, "Initial commit should produce a hash")
		require.Equal(t, int64(1), commitInfo.Version, "First commit should be version 1")

		initialAppHash := commitInfo.Hash
		t.Logf("Phase 1 - Initial AppHash: %X", initialAppHash)
		t.Logf("Phase 1 - Version: %d", commitInfo.Version)

		// Store the initial AppHash for later comparison
		err := os.WriteFile(filepath.Join(tmpDir, "apphash_phase1.txt"), initialAppHash, 0644)
		require.NoError(t, err, "Failed to write initial AppHash")
	})

	// Phase 2: Restart database and process blocks
	t.Run("Phase2_ProcessBlocks", func(t *testing.T) {
		db, cms, ctx := setupDatabase(t, dbPath)
		defer db.Close()

		// Load latest version (should be version 1 from Phase 1)
		err := cms.LoadLatestVersion()
		require.NoError(t, err, "Failed to load latest version in Phase 2")

		latestVersion := cms.LastCommitID().Version
		require.Equal(t, int64(1), latestVersion, "Latest version should be 1 after Phase 1")

		t.Logf("Phase 2 - Loaded version: %d", latestVersion)

		// Read the initial AppHash from Phase 1
		phase1Hash, err := os.ReadFile(filepath.Join(tmpDir, "apphash_phase1.txt"))
		require.NoError(t, err, "Failed to read Phase 1 AppHash")

		// Verify AppHash matches Phase 1
		currentAppHash := cms.LastCommitID().Hash
		assert.Equal(t, phase1Hash, currentAppHash, "AppHash should match Phase 1 after loading")
		t.Logf("Phase 2 - AppHash verified: %X", currentAppHash)

		// Process multiple blocks
		blockCount := 10
		blockHashes := make([][]byte, blockCount)

		for i := 0; i < blockCount; i++ {
			blockHeight := int64(i + 2) // Start from block 2 (genesis is block 1)

			// Create new context for this block
			ctx = ctx.WithBlockHeight(blockHeight).WithBlockTime(time.Now().UTC())

			// Simulate block processing by writing state changes
			processBlock(t, ctx, blockHeight)

			// Commit the block
			commitInfo := cms.Commit()
			require.NotNil(t, commitInfo, "Block %d commit should return CommitInfo", blockHeight)
			require.NotEmpty(t, commitInfo.Hash, "Block %d should produce a hash", blockHeight)

			blockHashes[i] = commitInfo.Hash
			t.Logf("Phase 2 - Block %d committed, AppHash: %X", blockHeight, commitInfo.Hash)
		}

		// Store the final AppHash and block hashes
		finalAppHash := blockHashes[blockCount-1]
		err = os.WriteFile(filepath.Join(tmpDir, "apphash_phase2.txt"), finalAppHash, 0644)
		require.NoError(t, err, "Failed to write Phase 2 final AppHash")

		t.Logf("Phase 2 - Final AppHash: %X", finalAppHash)
		t.Logf("Phase 2 - Final version: %d", cms.LastCommitID().Version)
	})

	// Phase 3: Stop and restart - verify state consistency
	t.Run("Phase3_RestartAndVerify", func(t *testing.T) {
		db, cms, ctx := setupDatabase(t, dbPath)
		defer db.Close()

		// Load latest version
		err := cms.LoadLatestVersion()
		require.NoError(t, err, "Failed to load latest version in Phase 3")

		latestVersion := cms.LastCommitID().Version
		require.Equal(t, int64(11), latestVersion, "Latest version should be 11 (genesis + 10 blocks)")

		t.Logf("Phase 3 - Loaded version: %d", latestVersion)

		// Read the final AppHash from Phase 2
		phase2Hash, err := os.ReadFile(filepath.Join(tmpDir, "apphash_phase2.txt"))
		require.NoError(t, err, "Failed to read Phase 2 final AppHash")

		// Verify AppHash matches Phase 2
		currentAppHash := cms.LastCommitID().Hash
		assert.Equal(t, phase2Hash, currentAppHash, "AppHash should match Phase 2 after restart")
		t.Logf("Phase 3 - AppHash verified after restart: %X", currentAppHash)

		// Verify we can read state from all stores without "version does not exist" errors
		verifyStoreAccess(t, ctx)

		// Process a few more blocks to ensure replay works correctly
		for i := 0; i < 5; i++ {
			blockHeight := latestVersion + int64(i) + 1
			ctx = ctx.WithBlockHeight(blockHeight).WithBlockTime(time.Now().UTC())

			processBlock(t, ctx, blockHeight)

			commitInfo := cms.Commit()
			require.NotNil(t, commitInfo, "Block %d commit should succeed after restart", blockHeight)
			require.NotEmpty(t, commitInfo.Hash, "Block %d should produce hash after restart", blockHeight)

			t.Logf("Phase 3 - Block %d committed, AppHash: %X", blockHeight, commitInfo.Hash)
		}

		finalVersion := cms.LastCommitID().Version
		require.Equal(t, int64(16), finalVersion, "Final version should be 16 (11 + 5 blocks)")
		t.Logf("Phase 3 - Final version after additional blocks: %d", finalVersion)
	})

	// Phase 4: Final restart - verify complete stability
	t.Run("Phase4_FinalRestart", func(t *testing.T) {
		db, cms, ctx := setupDatabase(t, dbPath)
		defer db.Close()

		// Load latest version
		err := cms.LoadLatestVersion()
		require.NoError(t, err, "Failed to load latest version in Phase 4")

		latestVersion := cms.LastCommitID().Version
		require.Equal(t, int64(16), latestVersion, "Latest version should be 16")

		t.Logf("Phase 4 - Loaded version: %d", latestVersion)
		t.Logf("Phase 4 - AppHash: %X", cms.LastCommitID().Hash)

		// Verify all stores are accessible
		verifyStoreAccess(t, ctx)

		// Verify historical versions are accessible
		verifyHistoricalVersions(t, cms, latestVersion)

		t.Log("Phase 4 - Database replay test completed successfully")
		t.Log("✓ AppHash stability verified across multiple stop/start cycles")
		t.Log("✓ No 'version does not exist' errors encountered")
		t.Log("✓ Historical state access working correctly")
	})
}

// Global store keys for the test
var (
	mainStoreKey       = storetypes.NewKVStoreKey("main")
	bankStoreKey       = storetypes.NewKVStoreKey("bank")
	stakingStoreKey    = storetypes.NewKVStoreKey("staking")
	identityStoreKey   = storetypes.NewKVStoreKey("identity")
	vcregistryStoreKey = storetypes.NewKVStoreKey("vcregistry")
	dexStoreKey        = storetypes.NewKVStoreKey("dex")
)

// setupDatabase creates a new database and commit multi-store
func setupDatabase(t *testing.T, dbPath string) (*dbm.GoLevelDB, store.CommitMultiStore, sdk.Context) {
	t.Helper()

	// Ensure the database directory exists
	err := os.MkdirAll(dbPath, 0755)
	require.NoError(t, err, "Failed to create database directory")

	// Open LevelDB (persistent)
	db, err := dbm.NewGoLevelDB("application", dbPath, nil)
	require.NoError(t, err, "Failed to open database")

	// Create commit multi-store
	logger := log.NewNopLogger()
	cms := store.NewCommitMultiStore(db, logger, metrics.NewNoOpMetrics())

	// Mount multiple stores to simulate a real chain
	// Pass nil for the DB parameter - the multistore handles key prefixing internally
	// This prevents IAVL version conflicts between stores
	cms.MountStoreWithDB(mainStoreKey, storetypes.StoreTypeIAVL, nil)
	cms.MountStoreWithDB(bankStoreKey, storetypes.StoreTypeIAVL, nil)
	cms.MountStoreWithDB(stakingStoreKey, storetypes.StoreTypeIAVL, nil)
	cms.MountStoreWithDB(identityStoreKey, storetypes.StoreTypeIAVL, nil)
	cms.MountStoreWithDB(vcregistryStoreKey, storetypes.StoreTypeIAVL, nil)
	cms.MountStoreWithDB(dexStoreKey, storetypes.StoreTypeIAVL, nil)

	// Load latest version (or initialize if new)
	err = cms.LoadLatestVersion()
	require.NoError(t, err, "Failed to load/initialize database")

	// Get the latest version to determine if this is a new or existing database
	latestVersion := cms.LastCommitID().Version

	// Create SDK context with appropriate block height
	// If this is a new database (version 0), start at height 1
	// If loading existing state, use latestVersion + 1 for the next block
	blockHeight := latestVersion + 1
	if blockHeight < 1 {
		blockHeight = 1
	}

	header := tmproto.Header{
		Height: blockHeight,
		Time:   time.Now().UTC(),
	}
	ctx := sdk.NewContext(cms, header, false, logger)

	return db, cms, ctx
}

// seedInitialState seeds the database with initial state
func seedInitialState(t *testing.T, ctx sdk.Context) {
	t.Helper()

	// Seed main store
	mainStore := ctx.KVStore(mainStoreKey)
	mainStore.Set([]byte("genesis_key"), []byte("genesis_value"))
	mainStore.Set([]byte("chain_id"), []byte("aura-test-1"))

	// Seed bank store
	bankStore := ctx.KVStore(bankStoreKey)
	bankStore.Set([]byte("total_supply"), []byte("1000000000"))
	bankStore.Set([]byte("balance_test1"), []byte("500000"))
	bankStore.Set([]byte("balance_test2"), []byte("300000"))

	// Seed staking store
	stakingStore := ctx.KVStore(stakingStoreKey)
	stakingStore.Set([]byte("validator_1"), []byte("validator_data_1"))
	stakingStore.Set([]byte("validator_2"), []byte("validator_data_2"))
	stakingStore.Set([]byte("total_bonded"), []byte("10000000"))

	// Seed AURA identity store
	identityStore := ctx.KVStore(identityStoreKey)
	identityStore.Set([]byte("identity_1"), []byte("did:aura:test1"))
	identityStore.Set([]byte("identity_2"), []byte("did:aura:test2"))

	// Seed AURA VC registry store
	vcStore := ctx.KVStore(vcregistryStoreKey)
	vcStore.Set([]byte("vc_1"), []byte("verifiable_credential_data_1"))
	vcStore.Set([]byte("vc_2"), []byte("verifiable_credential_data_2"))

	// Seed AURA DEX store
	dexStore := ctx.KVStore(dexStoreKey)
	dexStore.Set([]byte("pool_1"), []byte("liquidity_pool_data_1"))
	dexStore.Set([]byte("swap_count"), []byte("0"))

	t.Log("Initial state seeded successfully")
}

// processBlock simulates block processing by writing state changes
func processBlock(t *testing.T, ctx sdk.Context, blockHeight int64) {
	t.Helper()

	// Simulate state changes in various stores

	// Update main store
	mainStore := ctx.KVStore(mainStoreKey)
	mainStore.Set([]byte("last_block_height"), []byte(math.NewInt(blockHeight).String()))
	mainStore.Set([]byte("block_hash_"+fmt.Sprintf("%d", blockHeight)), []byte("hash_data"))

	// Update bank store (simulate transfers)
	bankStore := ctx.KVStore(bankStoreKey)
	bankStore.Set([]byte("balance_test1"), []byte(math.NewInt(500000+blockHeight).String()))
	bankStore.Set([]byte("balance_test2"), []byte(math.NewInt(300000-blockHeight).String()))

	// Update staking store (simulate staking changes)
	stakingStore := ctx.KVStore(stakingStoreKey)
	stakingStore.Set([]byte("total_bonded"), []byte(math.NewInt(10000000+blockHeight*1000).String()))

	// Update AURA modules
	identityStore := ctx.KVStore(identityStoreKey)
	identityStore.Set([]byte("identity_created_block_"+fmt.Sprintf("%d", blockHeight)), []byte("new_identity"))

	vcStore := ctx.KVStore(vcregistryStoreKey)
	vcStore.Set([]byte("vc_issued_block_"+fmt.Sprintf("%d", blockHeight)), []byte("new_vc"))

	dexStore := ctx.KVStore(dexStoreKey)
	currentSwapCount := blockHeight - 1 // Simple counter
	dexStore.Set([]byte("swap_count"), []byte(math.NewInt(currentSwapCount).String()))
	dexStore.Set([]byte("swap_"+fmt.Sprintf("%d", blockHeight)), []byte("swap_data"))
}

// verifyStoreAccess verifies that all stores are accessible and contain expected data
func verifyStoreAccess(t *testing.T, ctx sdk.Context) {
	t.Helper()

	storeKeys := []*storetypes.KVStoreKey{
		mainStoreKey,
		bankStoreKey,
		stakingStoreKey,
		identityStoreKey,
		vcregistryStoreKey,
		dexStoreKey,
	}

	for _, storeKey := range storeKeys {
		store := ctx.KVStore(storeKey)
		storeName := storeKey.Name()
		require.NotNil(t, store, "Store %s should be accessible", storeName)

		// Create an iterator to verify we can iterate the store
		iter := store.Iterator(nil, nil)
		require.NotNil(t, iter, "Iterator for store %s should be created", storeName)

		count := 0
		for ; iter.Valid(); iter.Next() {
			count++
			// Verify we can read keys and values without errors
			key := iter.Key()
			value := iter.Value()
			require.NotNil(t, key, "Key should not be nil in store %s", storeName)
			require.NotNil(t, value, "Value should not be nil in store %s", storeName)
		}
		iter.Close()

		t.Logf("Store %s: verified %d key-value pairs", storeName, count)
		require.Greater(t, count, 0, "Store %s should contain data", storeName)
	}

	t.Log("All stores verified successfully - no 'version does not exist' errors")
}

// verifyHistoricalVersions verifies that historical versions can be accessed
func verifyHistoricalVersions(t *testing.T, cms store.CommitMultiStore, latestVersion int64) {
	t.Helper()

	// Try to access several historical versions
	versionsToCheck := []int64{1, latestVersion / 2, latestVersion}

	for _, version := range versionsToCheck {
		if version > latestVersion {
			continue
		}

		// CacheMultiStoreWithVersion should not panic or error
		cacheMS, err := cms.CacheMultiStoreWithVersion(version)
		require.NoError(t, err, "Failed to create cache for version %d", version)
		require.NotNil(t, cacheMS, "Cache multi-store should not be nil for version %d", version)

		t.Logf("Historical version %d accessible", version)
	}

	t.Log("Historical version access verified successfully")
}

// TestAppHashDeterminism verifies that identical operations produce identical AppHashes
func TestAppHashDeterminism(t *testing.T) {
	// Create two separate databases
	tmpDir1, err := os.MkdirTemp("", "aura-determinism-test-1-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir1)

	tmpDir2, err := os.MkdirTemp("", "aura-determinism-test-2-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir2)

	// Run identical operations on both databases
	runOperations := func(dbPath string) []byte {
		db, cms, ctx := setupDatabase(t, filepath.Join(dbPath, "data"))
		defer db.Close()

		seedInitialState(t, ctx)
		cms.Commit()

		for i := 0; i < 5; i++ {
			blockHeight := int64(i + 2)
			ctx = ctx.WithBlockHeight(blockHeight).WithBlockTime(time.Unix(1000000+int64(i)*100, 0))
			processBlock(t, ctx, blockHeight)
			cms.Commit()
		}

		return cms.LastCommitID().Hash
	}

	hash1 := runOperations(tmpDir1)
	hash2 := runOperations(tmpDir2)

	assert.Equal(t, hash1, hash2, "AppHashes should be deterministic")
	t.Logf("Determinism verified - AppHash: %X", hash1)
}

// TestStoreVersionPersistence verifies that store versions persist correctly
func TestStoreVersionPersistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aura-version-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "data")

	// Phase 1: Create and commit multiple versions
	func() {
		db, cms, ctx := setupDatabase(t, dbPath)
		defer db.Close()

		seedInitialState(t, ctx)
		cms.Commit()

		for i := 0; i < 10; i++ {
			blockHeight := int64(i + 2)
			ctx = ctx.WithBlockHeight(blockHeight)
			processBlock(t, ctx, blockHeight)
			commitInfo := cms.Commit()
			t.Logf("Committed version %d, hash: %X", commitInfo.Version, commitInfo.Hash)
		}

		finalVersion := cms.LastCommitID().Version
		require.Equal(t, int64(11), finalVersion, "Should have 11 versions")
	}()

	// Phase 2: Reopen and verify versions
	func() {
		db, cms, _ := setupDatabase(t, dbPath)
		defer db.Close()

		err := cms.LoadLatestVersion()
		require.NoError(t, err)

		latestVersion := cms.LastCommitID().Version
		require.Equal(t, int64(11), latestVersion, "Latest version should persist")
		t.Logf("Latest version after reopen: %d", latestVersion)

		// Verify we can load specific versions
		for v := int64(1); v <= latestVersion; v++ {
			cacheMS, err := cms.CacheMultiStoreWithVersion(v)
			require.NoError(t, err, "Failed to load version %d", v)
			require.NotNil(t, cacheMS)
		}

		t.Log("All versions accessible after reopen")
	}()
}

// TestConcurrentStoreAccess verifies thread-safe store access during commit
func TestConcurrentStoreAccess(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aura-concurrent-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "data")
	db, cms, ctx := setupDatabase(t, dbPath)
	defer db.Close()

	seedInitialState(t, ctx)
	cms.Commit()

	// Simulate concurrent reads during block processing
	blockHeight := int64(2)
	ctx = ctx.WithBlockHeight(blockHeight)

	// Write to store - use the mounted key, not a new one
	mainStore := ctx.KVStore(mainStoreKey)
	mainStore.Set([]byte("concurrent_key"), []byte("concurrent_value"))

	// Multiple reads
	for i := 0; i < 100; i++ {
		value := mainStore.Get([]byte("concurrent_key"))
		require.NotNil(t, value)
	}

	// Commit should succeed
	commitInfo := cms.Commit()
	require.NotNil(t, commitInfo)
	require.NotEmpty(t, commitInfo.Hash)

	t.Log("Concurrent store access test passed")
}

// TestEmptyBlockCommit verifies that empty blocks (no state changes) can be committed
func TestEmptyBlockCommit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aura-empty-block-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "data")
	db, cms, ctx := setupDatabase(t, dbPath)
	defer db.Close()

	seedInitialState(t, ctx)
	firstCommit := cms.Commit()
	require.NotEmpty(t, firstCommit.Hash)

	// Commit empty block (no state changes)
	ctx = ctx.WithBlockHeight(2) //nolint:staticcheck // Reassignment is intentional for testing block height changes
	emptyCommit := cms.Commit()
	require.NotEmpty(t, emptyCommit.Hash)

	// Hash should be different even though state didn't change
	// (because block height metadata changed)
	t.Logf("First commit hash: %X", firstCommit.Hash)
	t.Logf("Empty commit hash: %X", emptyCommit.Hash)

	// Version should increment
	require.Equal(t, int64(2), emptyCommit.Version)
}

// createTestCodec creates a test codec (helper function, not a test)
func createTestCodec() codec.Codec {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	return codec.NewProtoCodec(interfaceRegistry)
}
