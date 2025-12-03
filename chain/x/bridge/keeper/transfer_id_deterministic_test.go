package keeper

import (
	"fmt"
	"sync"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
)

// TransferIDTestSuite tests the deterministic transfer ID generation system
type TransferIDTestSuite struct {
	KeeperTestSuite
}

// TestTransferIDTestSuite runs the test suite
func TestTransferIDTestSuite(t *testing.T) {
	suite.Run(t, new(TransferIDTestSuite))
}

// TestTransferIDDeterministic verifies that transfer IDs are deterministic
// based on block height and transaction bytes.
func (suite *TransferIDTestSuite) TestTransferIDDeterministic() {
	ctx := suite.SdkCtx

	// Test that same transaction context produces same ID
	testCtx := ctx.WithBlockHeight(100).WithTxBytes([]byte("test-tx-1"))

	// Generate ID twice with same context
	transferID1 := suite.Keeper.nextTransferID(testCtx)
	transferID2 := suite.Keeper.nextTransferID(testCtx)

	// Should be deterministic
	suite.Equal(transferID1, transferID2,
		"Same transaction context should produce same transfer ID")
	suite.Contains(transferID1, "transfer-",
		"Transfer ID should have correct prefix")
}

// TestTransferIDUniqueness verifies that different transactions produce unique transfer IDs.
func (suite *TransferIDTestSuite) TestTransferIDUniqueness() {
	ctx := suite.SdkCtx

	seenIDs := make(map[string]struct{})
	collisions := 0

	// Test various height and tx byte combinations
	for height := int64(1); height <= 50; height++ {
		for txNum := 0; txNum < 10; txNum++ {
			txBytes := []byte(fmt.Sprintf("tx-%d-%d", height, txNum))
			testCtx := ctx.WithBlockHeight(height).WithTxBytes(txBytes)
			transferID := suite.Keeper.nextTransferID(testCtx)

			// Check for uniqueness
			if _, exists := seenIDs[transferID]; exists {
				collisions++
				suite.T().Errorf("Collision detected: ID %s generated for height=%d, tx=%d",
					transferID, height, txNum)
			}
			seenIDs[transferID] = struct{}{}
		}
	}

	suite.Equal(0, collisions, "No ID collisions should occur")
	suite.Equal(500, len(seenIDs), "Should generate 500 unique IDs")
}

// TestTransferIDConcurrency tests that deterministic IDs prevent race conditions
// in parallel execution scenarios.
func (suite *TransferIDTestSuite) TestTransferIDConcurrency() {
	ctx := suite.SdkCtx

	const numGoroutines = 100
	const numTxsPerRoutine = 10

	// Each goroutine simulates processing transactions
	var wg sync.WaitGroup
	idsChan := make(chan string, numGoroutines*numTxsPerRoutine)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			defer wg.Done()

			for j := 0; j < numTxsPerRoutine; j++ {
				height := int64(routineID*10 + j)
				txBytes := []byte(fmt.Sprintf("tx-%d-%d", routineID, j))

				testCtx := ctx.WithBlockHeight(height).WithTxBytes(txBytes)
				transferID := suite.Keeper.nextTransferID(testCtx)

				idsChan <- transferID
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(idsChan)

	// Collect all IDs and check for duplicates
	seenIDs := make(map[string]bool)
	totalIDs := 0

	for id := range idsChan {
		totalIDs++
		if seenIDs[id] {
			suite.T().Errorf("Duplicate ID detected in concurrent execution: %s", id)
		}
		seenIDs[id] = true
	}

	expectedIDs := numGoroutines * numTxsPerRoutine
	suite.Equal(expectedIDs, totalIDs,
		"Should generate expected number of IDs")
	suite.Equal(expectedIDs, len(seenIDs),
		"All IDs should be unique (no duplicates)")
}

// TestExtractLegacyTransferID verifies that legacy sequential IDs
// can still be parsed for backward compatibility.
func (suite *TransferIDTestSuite) TestExtractLegacyTransferID() {
	tests := []struct {
		name          string
		transferID    string
		expectedH     int64
		expectedValid bool
	}{
		{
			name:          "legacy ID 1",
			transferID:    "transfer-1",
			expectedH:     1,
			expectedValid: true,
		},
		{
			name:          "legacy ID 100",
			transferID:    "transfer-100",
			expectedH:     100,
			expectedValid: true,
		},
		{
			name:          "legacy ID 999999",
			transferID:    "transfer-999999",
			expectedH:     999999,
			expectedValid: true,
		},
		{
			name:          "modern hash-based ID (large number)",
			transferID:    "transfer-9223372036854775807", // Max int64, beyond legacy threshold
			expectedH:     0,
			expectedValid: false,
		},
		{
			name:          "invalid format - no prefix",
			transferID:    "12345",
			expectedH:     0,
			expectedValid: false,
		},
		{
			name:          "invalid format - wrong prefix",
			transferID:    "swap-12345",
			expectedH:     0,
			expectedValid: false,
		},
		{
			name:          "invalid format - non-numeric",
			transferID:    "transfer-abc",
			expectedH:     0,
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			height, _, valid := suite.Keeper.extractBlockHeightFromTransferID(tt.transferID)

			suite.Equal(tt.expectedValid, valid, "Valid flag should match expected")
			if tt.expectedValid {
				suite.Equal(tt.expectedH, height, "Extracted value should match for legacy IDs")
			}
		})
	}
}

// TestTransferIDDifferentBlocks verifies that IDs differ across blocks
func (suite *TransferIDTestSuite) TestTransferIDDifferentBlocks() {
	ctx := suite.SdkCtx

	// Same transaction bytes but different blocks
	txBytes := []byte("identical-tx")

	ids := make([]string, 0, 10)
	for height := int64(1); height <= 10; height++ {
		testCtx := ctx.WithBlockHeight(height).WithTxBytes(txBytes)
		transferID := suite.Keeper.nextTransferID(testCtx)
		ids = append(ids, transferID)
	}

	// Verify all IDs are unique (block height affects hash)
	seenIDs := make(map[string]bool)
	for _, id := range ids {
		suite.False(seenIDs[id], "ID %s should be unique across blocks", id)
		seenIDs[id] = true
	}
}

// TestTransferIDDifferentTxBytes verifies that IDs differ for different transactions
func (suite *TransferIDTestSuite) TestTransferIDDifferentTxBytes() {
	ctx := suite.SdkCtx

	// Same block but different transaction bytes
	height := int64(100)

	ids := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		txBytes := []byte(fmt.Sprintf("tx-%d", i))
		testCtx := ctx.WithBlockHeight(height).WithTxBytes(txBytes)
		transferID := suite.Keeper.nextTransferID(testCtx)
		ids = append(ids, transferID)
	}

	// Verify all IDs are unique (tx bytes affect hash)
	seenIDs := make(map[string]bool)
	for _, id := range ids {
		suite.False(seenIDs[id], "ID %s should be unique for different tx bytes", id)
		seenIDs[id] = true
	}
}

// TestTransferIDNoRaceConditionRealWorld simulates a real-world scenario
// where multiple validators process transactions concurrently.
func (suite *TransferIDTestSuite) TestTransferIDNoRaceConditionRealWorld() {
	ctx := suite.SdkCtx

	const numValidators = 10
	const numBlocks = 50
	const txsPerBlock = 5

	// Each validator independently generates IDs for the same blocks/transactions
	allIDs := make([]map[string]bool, numValidators)
	var wg sync.WaitGroup

	wg.Add(numValidators)
	for validatorID := 0; validatorID < numValidators; validatorID++ {
		allIDs[validatorID] = make(map[string]bool)

		go func(vID int) {
			defer wg.Done()

			localIDs := make(map[string]bool)

			// Each validator processes all blocks with same transactions
			for block := int64(1); block <= numBlocks; block++ {
				for txIdx := int64(0); txIdx < txsPerBlock; txIdx++ {
					// IMPORTANT: Same tx bytes for same block+index across all validators
					txBytes := []byte(fmt.Sprintf("tx-block%d-idx%d", block, txIdx))
					testCtx := ctx.WithBlockHeight(block).WithTxBytes(txBytes)
					transferID := suite.Keeper.nextTransferID(testCtx)
					localIDs[transferID] = true
				}
			}

			allIDs[vID] = localIDs
		}(validatorID)
	}

	wg.Wait()

	// Verify all validators generated the SAME IDs (determinism)
	baselineIDs := allIDs[0]
	for vID := 1; vID < numValidators; vID++ {
		validatorIDs := allIDs[vID]

		suite.Equal(len(baselineIDs), len(validatorIDs),
			"All validators should generate same number of IDs")

		// Check that all IDs match
		for id := range baselineIDs {
			suite.True(validatorIDs[id],
				"Validator %d should generate same ID as baseline: %s", vID, id)
		}
	}

	expectedTotalIDs := numBlocks * txsPerBlock
	suite.Equal(expectedTotalIDs, len(baselineIDs),
		"Should generate expected total number of unique IDs")
}

// TestTransferIDDuplicateDetection verifies that the defensive duplicate check
// works correctly when a duplicate ID is somehow generated.
func (suite *TransferIDTestSuite) TestTransferIDDuplicateDetection() {
	ctx := suite.SdkCtx

	// Create a context and generate ID
	txBytes := []byte("test-tx")
	testCtx := ctx.WithBlockHeight(100).WithTxBytes(txBytes)

	// First call - should succeed
	transferID1 := suite.Keeper.nextTransferID(testCtx)

	// Create and store a transfer with this ID
	transfer := &types.CrossChainTransfer{
		TransferId:  transferID1,
		SourceChain: "aura",
		TargetChain: "paw",
		Sender:      "sender1",
		Recipient:   "recipient1",
		Amount:      "100",
		Denom:       "uaura",
		Status:      types.TransferStatus_PENDING,
	}
	suite.Keeper.SetTransfer(testCtx, transfer)

	// Second call with same context - should return DIFFERENT ID
	// The function detects collision and adds nonce to generate unique ID
	transferID2 := suite.Keeper.nextTransferID(testCtx)

	// IDs should be DIFFERENT (collision avoidance working correctly)
	suite.NotEqual(transferID1, transferID2, "Should generate new ID when collision detected")
}

// TestTransferIDEmptyTxBytes verifies behavior with empty transaction bytes
func (suite *TransferIDTestSuite) TestTransferIDEmptyTxBytes() {
	ctx := suite.SdkCtx

	// Different blocks with empty tx bytes should produce different IDs
	// (because block height and header hash differ)
	ids := make([]string, 0, 5)
	for height := int64(1); height <= 5; height++ {
		testCtx := ctx.WithBlockHeight(height).WithTxBytes([]byte{})
		transferID := suite.Keeper.nextTransferID(testCtx)
		ids = append(ids, transferID)
	}

	// All IDs should be unique
	seenIDs := make(map[string]bool)
	for _, id := range ids {
		suite.False(seenIDs[id], "ID should be unique even with empty tx bytes")
		seenIDs[id] = true
	}
}

// TestTransferIDFormat verifies the format of generated IDs
func (suite *TransferIDTestSuite) TestTransferIDFormat() {
	ctx := suite.SdkCtx

	testCtx := ctx.WithBlockHeight(100).WithTxBytes([]byte("test-tx"))
	transferID := suite.Keeper.nextTransferID(testCtx)

	// Verify format: "transfer-{number}"
	suite.Contains(transferID, "transfer-", "ID should have transfer prefix")

	// Extract numeric part
	_, _, valid := suite.Keeper.extractBlockHeightFromTransferID(transferID)
	suite.True(valid || len(transferID) > len("transfer-"),
		"ID should have valid format")
}

// Benchmark tests
func BenchmarkTransferIDGeneration(b *testing.B) {
	suite := new(TransferIDTestSuite)
	suite.SetupTest()
	ctx := suite.SdkCtx

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		txBytes := []byte(fmt.Sprintf("tx-%d", i))
		testCtx := ctx.WithBlockHeight(int64(i/100)).WithTxBytes(txBytes)
		_ = suite.Keeper.nextTransferID(testCtx)
	}
}

func BenchmarkTransferIDConcurrent(b *testing.B) {
	suite := new(TransferIDTestSuite)
	suite.SetupTest()
	ctx := suite.SdkCtx

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			txBytes := []byte(fmt.Sprintf("tx-%d", i))
			testCtx := ctx.WithBlockHeight(int64(i/100)).WithTxBytes(txBytes)
			_ = suite.Keeper.nextTransferID(testCtx)
			i++
		}
	})
}

// TestTransferIDSimple tests transfer ID generation with simple setup
func TestTransferIDSimple(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	legacyAmino := codec.NewLegacyAmino()
	ps := paramtypes.NewSubspace(input.Cdc, legacyAmino, input.StoreKey, input.MemStoreKey, "bridge")

	keeper := NewKeeper(
		input.Cdc,
		input.StoreKey,
		&ps,
		nil, // bankKeeper
		nil, // accountKeeper
		nil, // vcKeeper
		nil, // stakingKeeper
	)

	ctx := input.Ctx.WithBlockHeight(100).WithTxBytes([]byte("test"))

	// First call generates ID
	id1 := keeper.nextTransferID(ctx)

	// Second call with SAME context (no storage in between) should generate SAME ID
	id2 := keeper.nextTransferID(ctx)

	require.Equal(t, id1, id2, "Deterministic IDs should match when context is identical")
	require.Contains(t, id1, "transfer-")
}
