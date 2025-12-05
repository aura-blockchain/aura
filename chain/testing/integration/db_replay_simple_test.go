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
	dbm "github.com/cosmos/cosmos-db"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDatabaseReplaySimple tests database replay with proper IAVL version handling
// This test addresses ROADMAP_PRODUCTION.md Task #8:
// "Verify replay on seeded DB via start→stop→start integration test"
func TestDatabaseReplaySimple(t *testing.T) {
	// Create a temporary directory for test data
	tmpDir, err := os.MkdirTemp("", "aura-replay-simple-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "testdb")

	// Store keys (create fresh for each test run)
	mainKey := storetypes.NewKVStoreKey("main")
	bankKey := storetypes.NewKVStoreKey("bank")
	stakingKey := storetypes.NewKVStoreKey("staking")

	var finalAppHash []byte
	var finalVersion int64

	// Phase 1: Initialize and commit several blocks
	t.Run("Phase1_InitAndCommit", func(t *testing.T) {
		db, cms, ctx := setupTestDB(t, dbPath, mainKey, bankKey, stakingKey)
		defer db.Close()

		// Verify starting state
		require.Equal(t, int64(0), cms.LastCommitID().Version)

		// Update context to next block for first commit
		ctx = ctx.WithBlockHeight(1)

		// Seed initial data
		mainStore := ctx.KVStore(mainKey)
		mainStore.Set([]byte("genesis"), []byte("value"))

		// Commit genesis (version 1)
		commitInfo := cms.Commit()
		require.Equal(t, int64(1), commitInfo.Version)
		t.Logf("Phase 1 - Genesis committed: version=%d, hash=%X", commitInfo.Version, commitInfo.Hash)

		// Process several blocks
		for i := 0; i < 5; i++ {
			blockHeight := int64(i + 2)
			ctx = ctx.WithBlockHeight(blockHeight)

			// Write some data
			mainStore.Set([]byte(fmt.Sprintf("block_%d", blockHeight)), []byte(fmt.Sprintf("data_%d", blockHeight)))
			bankStore := ctx.KVStore(bankKey)
			bankStore.Set([]byte("balance"), []byte(math.NewInt(blockHeight*1000).String()))

			// Commit
			commitInfo = cms.Commit()
			t.Logf("Phase 1 - Block %d committed: version=%d, hash=%X", blockHeight, commitInfo.Version, commitInfo.Hash)
		}

		finalAppHash = cms.LastCommitID().Hash
		finalVersion = cms.LastCommitID().Version
		require.Equal(t, int64(6), finalVersion) // genesis + 5 blocks

		t.Logf("Phase 1 - Final state: version=%d, hash=%X", finalVersion, finalAppHash)
	})

	// Phase 2: Reopen database and verify state
	t.Run("Phase2_ReopenAndVerify", func(t *testing.T) {
		db, cms, ctx := setupTestDB(t, dbPath, mainKey, bankKey, stakingKey)
		defer db.Close()

		// Verify version was persisted
		loadedVersion := cms.LastCommitID().Version
		loadedHash := cms.LastCommitID().Hash

		require.Equal(t, finalVersion, loadedVersion, "Version should persist across restarts")
		assert.Equal(t, finalAppHash, loadedHash, "AppHash should persist across restarts")

		t.Logf("Phase 2 - Loaded state: version=%d, hash=%X", loadedVersion, loadedHash)

		// Verify we can read the data
		mainStore := ctx.KVStore(mainKey)
		genesisValue := mainStore.Get([]byte("genesis"))
		require.Equal(t, []byte("value"), genesisValue, "Genesis data should be accessible")

		block2Value := mainStore.Get([]byte("block_2"))
		require.NotNil(t, block2Value, "Block 2 data should be accessible")

		t.Log("Phase 2 - All stored data verified successfully")
	})

	// Phase 3: Continue from loaded state
	t.Run("Phase3_ContinueFromLoaded", func(t *testing.T) {
		db, cms, ctx := setupTestDB(t, dbPath, mainKey, bankKey, stakingKey)
		defer db.Close()

		startVersion := cms.LastCommitID().Version
		require.Equal(t, finalVersion, startVersion)

		// Process additional blocks
		for i := 0; i < 3; i++ {
			blockHeight := startVersion + int64(i) + 1
			ctx = ctx.WithBlockHeight(blockHeight)

			mainStore := ctx.KVStore(mainKey)
			mainStore.Set([]byte(fmt.Sprintf("block_%d", blockHeight)), []byte(fmt.Sprintf("data_%d", blockHeight)))

			commitInfo := cms.Commit()
			t.Logf("Phase 3 - Block %d committed: version=%d, hash=%X", blockHeight, commitInfo.Version, commitInfo.Hash)
		}

		newVersion := cms.LastCommitID().Version
		require.Equal(t, finalVersion+3, newVersion, "Should have 3 more versions")

		t.Logf("Phase 3 - Continued to version=%d", newVersion)
	})

	// Phase 4: Final verification
	t.Run("Phase4_FinalVerification", func(t *testing.T) {
		db, cms, ctx := setupTestDB(t, dbPath, mainKey, bankKey, stakingKey)
		defer db.Close()

		finalVers := cms.LastCommitID().Version
		require.Equal(t, int64(9), finalVers, "Should have 9 total versions (genesis + 8 blocks)")

		// Verify historical data
		mainStore := ctx.KVStore(mainKey)

		// Check genesis data
		genesisValue := mainStore.Get([]byte("genesis"))
		require.Equal(t, []byte("value"), genesisValue)

		// Check all block data
		for i := int64(2); i <= 9; i++ {
			key := []byte(fmt.Sprintf("block_%d", i))
			value := mainStore.Get(key)
			require.NotNil(t, value, "Block %d data should exist", i)
			t.Logf("Phase 4 - Verified block %d data exists", i)
		}

		t.Log("✓ Database replay test PASSED")
		t.Log("✓ AppHash stability verified")
		t.Log("✓ State persistence verified across multiple restarts")
		t.Log("✓ No 'version does not exist' errors encountered")
	})
}

// setupTestDB creates a new database connection and commit multi-store
// This function properly handles both initial creation and reopening
func setupTestDB(t *testing.T, dbPath string, storeKeys ...*storetypes.KVStoreKey) (dbm.DB, store.CommitMultiStore, sdk.Context) {
	t.Helper()

	// Ensure directory exists
	err := os.MkdirAll(dbPath, 0755)
	require.NoError(t, err)

	// Open database
	db, err := dbm.NewGoLevelDB("test", dbPath, nil)
	require.NoError(t, err)

	// Create commit multi-store
	logger := log.NewNopLogger()
	cms := store.NewCommitMultiStore(db, logger, metrics.NewNoOpMetrics())

	// Mount all stores
	for _, key := range storeKeys {
		cms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	}

	// Load latest version
	err = cms.LoadLatestVersion()
	require.NoError(t, err)

	// Get current version
	currentVersion := cms.LastCommitID().Version
	t.Logf("setupTestDB: loaded version %d from %s", currentVersion, dbPath)

	// Create context at appropriate height
	// Start at height 0 for a new database, otherwise use currentVersion
	blockHeight := currentVersion

	header := tmproto.Header{
		Height: blockHeight,
		Time:   time.Now().UTC(),
	}
	ctx := sdk.NewContext(cms, header, false, logger)

	return db, cms, ctx
}

// TestStoreVersionStability specifically tests IAVL version handling
func TestStoreVersionStability(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aura-version-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "testdb")
	key := storetypes.NewKVStoreKey("test")

	// Phase 1: Create and commit
	func() {
		db, cms, ctx := setupTestDB(t, dbPath, key)
		defer db.Close()

		store := ctx.KVStore(key)
		store.Set([]byte("key1"), []byte("value1"))

		commitInfo := cms.Commit()
		require.Equal(t, int64(1), commitInfo.Version)
		t.Logf("Committed version 1, hash: %X", commitInfo.Hash)
	}()

	// Phase 2: Reopen and verify
	func() {
		db, cms, ctx := setupTestDB(t, dbPath, key)
		defer db.Close()

		require.Equal(t, int64(1), cms.LastCommitID().Version)

		store := ctx.KVStore(key)
		value := store.Get([]byte("key1"))
		require.Equal(t, []byte("value1"), value)

		t.Log("✓ Version stability verified")
	}()
}

// TestNoVersionExistsError verifies we don't get "version does not exist" errors
func TestNoVersionExistsError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "aura-no-error-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "testdb")
	key := storetypes.NewKVStoreKey("test")

	// Create initial state
	func() {
		db, cms, ctx := setupTestDB(t, dbPath, key)
		defer db.Close()

		for i := 0; i < 10; i++ {
			store := ctx.KVStore(key)
			store.Set([]byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i)))
			cms.Commit()
		}

		t.Logf("Created 10 versions")
	}()

	// Reopen and access all versions
	func() {
		db, cms, _ := setupTestDB(t, dbPath, key)
		defer db.Close()

		currentVersion := cms.LastCommitID().Version
		require.Equal(t, int64(10), currentVersion)

		// Try to access historical versions
		for v := int64(1); v <= currentVersion; v++ {
			cacheMS, err := cms.CacheMultiStoreWithVersion(v)
			require.NoError(t, err, "Should be able to load version %d without error", v)
			require.NotNil(t, cacheMS)
		}

		t.Log("✓ No 'version does not exist' errors - all historical versions accessible")
	}()
}

// createTestCodec creates a basic codec for testing
func createTestCodecSimple() codec.Codec {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	return codec.NewProtoCodec(interfaceRegistry)
}
