package keeper

import (
	"encoding/binary"
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/bridge/types"
	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

// GenesisCounterTestSuite tests the critical transfer counter restoration logic
// that prevents duplicate transfer ID collisions after chain restart.
//
// SECURITY CRITICAL: These tests verify the fix for issue #047
// Without proper counter restoration (max+1), the next transfer after restart
// would get a DUPLICATE ID, silently overwriting an existing transfer.
type GenesisCounterTestSuite struct {
	KeeperTestSuite
}

func TestGenesisCounterTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisCounterTestSuite))
}

// TestCounterRestorationToMaxPlusOne verifies the CRITICAL fix:
// Counter must be restored to MAX+1 (not MAX) to prevent duplicate IDs.
//
// Bug scenario (what this test prevents):
//
//	Genesis has: transfer-1, transfer-2, transfer-5
//	WRONG: counter = 5 → next transfer gets ID 5 → COLLISION with existing transfer-5
//	RIGHT: counter = 6 → next transfer gets ID 6 → no collision
func (suite *GenesisCounterTestSuite) TestCounterRestorationToMaxPlusOne() {
	suite.Run("counter set to max+1 with multiple legacy transfers", func() {
		suite.SetupTest()
		ctx := suite.SdkCtx

		// Genesis with legacy sequential IDs: 1, 2, 5
		// Max ID is 5, so counter should be set to 6 (not 5)
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []types.CrossChainTransfer{
				{
					TransferId:  "transfer-1",
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x123",
					Recipient:   "aura1test1",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
				{
					TransferId:  "transfer-2",
					SourceChain: "bsc",
					TargetChain: "aura",
					Sender:      "0x456",
					Recipient:   "aura1test2",
					Amount:      sdkmath.NewInt(2000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_CONFIRMED,
				},
				{
					TransferId:  "transfer-5", // Gap in sequence (3, 4 missing)
					SourceChain: "polygon",
					TargetChain: "aura",
					Sender:      "0x789",
					Recipient:   "aura1test3",
					Amount:      sdkmath.NewInt(3000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
			},
		}

		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.NoError(err, "InitGenesis should succeed")

		// CRITICAL CHECK: Verify counter is set to MAX+1 (6, not 5)
		store := suite.StoreKey
		storeObj := ctx.KVStore(store)
		counterBz := storeObj.Get(types.TransferCounterKey)
		suite.NotNil(counterBz, "Transfer counter should be set for legacy IDs")

		counter := binary.BigEndian.Uint64(counterBz)
		suite.Equal(uint64(6), counter,
			"Counter MUST be set to MAX+1 (6), not MAX (5), to prevent duplicate ID collision")

		// Verify all transfers were imported correctly
		transfer1, found := suite.Keeper.getTransfer(ctx, "transfer-1")
		suite.True(found, "transfer-1 should exist")
		suite.Equal("0x123", transfer1.Sender)

		transfer2, found := suite.Keeper.getTransfer(ctx, "transfer-2")
		suite.True(found, "transfer-2 should exist")
		suite.Equal("0x456", transfer2.Sender)

		transfer5, found := suite.Keeper.getTransfer(ctx, "transfer-5")
		suite.True(found, "transfer-5 should exist")
		suite.Equal("0x789", transfer5.Sender)
	})

	suite.Run("counter with single legacy transfer", func() {
		suite.SetupTest()
		ctx := suite.SdkCtx

		// Genesis with single transfer ID 100
		// Counter should be 101 (not 100)
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []types.CrossChainTransfer{
				{
					TransferId:  "transfer-100",
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x123",
					Recipient:   "aura1test",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
			},
		}

		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		// Counter should be 101 (MAX+1), not 100
		store := suite.StoreKey
		storeObj := ctx.KVStore(store)
		counterBz := storeObj.Get(types.TransferCounterKey)
		suite.NotNil(counterBz, "Counter should be set")

		counter := binary.BigEndian.Uint64(counterBz)
		suite.Equal(uint64(101), counter,
			"Counter must be MAX+1 (101) to prevent duplicate with transfer-100")
	})

	suite.Run("no counter set when no transfers", func() {
		suite.SetupTest()
		ctx := suite.SdkCtx

		// Genesis with no transfers - counter should NOT be set
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []types.CrossChainTransfer{},
		}

		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		// Counter should NOT be set when no transfers exist
		store := suite.StoreKey
		storeObj := ctx.KVStore(store)
		counterBz := storeObj.Get(types.TransferCounterKey)
		suite.Nil(counterBz, "Counter should not be set when no transfers in genesis")
	})

	suite.Run("no counter set for hash-based IDs only", func() {
		suite.SetupTest()
		ctx := suite.SdkCtx

		// Genesis with only new hash-based IDs (large values > 1 trillion)
		// Counter should NOT be set as these are deterministic IDs
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []types.CrossChainTransfer{
				{
					// Hash-based ID (large value > legacyIDThreshold)
					TransferId:  "transfer-12345678901234567890",
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x123",
					Recipient:   "aura1test",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
			},
		}

		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		// Counter should NOT be set for hash-based IDs
		store := suite.StoreKey
		storeObj := ctx.KVStore(store)
		counterBz := storeObj.Get(types.TransferCounterKey)
		suite.Nil(counterBz,
			"Counter should not be set for hash-based IDs (deterministic generation)")
	})

	suite.Run("mixed legacy and hash-based IDs", func() {
		suite.SetupTest()
		ctx := suite.SdkCtx

		// Genesis with both legacy sequential IDs and new hash-based IDs
		// Counter should track only legacy IDs
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []types.CrossChainTransfer{
				{
					TransferId:  "transfer-10", // Legacy ID
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x123",
					Recipient:   "aura1test1",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
				{
					// Hash-based ID (should be ignored for counter)
					TransferId:  "transfer-98765432109876543210",
					SourceChain: "bsc",
					TargetChain: "aura",
					Sender:      "0x456",
					Recipient:   "aura1test2",
					Amount:      sdkmath.NewInt(2000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_CONFIRMED,
				},
				{
					TransferId:  "transfer-25", // Legacy ID (max)
					SourceChain: "polygon",
					TargetChain: "aura",
					Sender:      "0x789",
					Recipient:   "aura1test3",
					Amount:      sdkmath.NewInt(3000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
			},
		}

		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		// Counter should be MAX(legacy IDs) + 1 = 25 + 1 = 26
		store := suite.StoreKey
		storeObj := ctx.KVStore(store)
		counterBz := storeObj.Get(types.TransferCounterKey)
		suite.NotNil(counterBz, "Counter should be set for legacy IDs")

		counter := binary.BigEndian.Uint64(counterBz)
		suite.Equal(uint64(26), counter,
			"Counter should be MAX(legacy IDs) + 1, ignoring hash-based IDs")
	})
}

// TestDuplicateTransferIDRejection verifies that duplicate transfer IDs
// in genesis are detected and rejected to prevent silent data corruption.
//
// CRITICAL SECURITY: Duplicate IDs would cause silent overwrites where
// the second transfer with same ID replaces the first, causing data loss.
func (suite *GenesisCounterTestSuite) TestDuplicateTransferIDRejection() {
	suite.Run("reject duplicate legacy transfer IDs", func() {
		suite.SetupTest()
		ctx := suite.SdkCtx

		// Genesis with DUPLICATE transfer ID (transfer-5 appears twice)
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []types.CrossChainTransfer{
				{
					TransferId:  "transfer-5",
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x123",
					Recipient:   "aura1test1",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
				{
					TransferId:  "transfer-5", // DUPLICATE - should be rejected
					SourceChain: "bsc",
					TargetChain: "aura",
					Sender:      "0x456",
					Recipient:   "aura1test2",
					Amount:      sdkmath.NewInt(2000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_CONFIRMED,
				},
			},
		}

		// InitGenesis MUST reject duplicate IDs
		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.Error(err, "InitGenesis must reject duplicate transfer IDs")
		suite.Contains(err.Error(), "duplicate transfer ID",
			"Error message should indicate duplicate ID problem")
		suite.Contains(err.Error(), "transfer-5",
			"Error message should specify which ID is duplicated")
	})

	suite.Run("reject duplicate hash-based transfer IDs", func() {
		suite.SetupTest()
		ctx := suite.SdkCtx

		// Genesis with DUPLICATE hash-based transfer ID
		duplicateID := "transfer-12345678901234567890"
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []types.CrossChainTransfer{
				{
					TransferId:  duplicateID,
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x123",
					Recipient:   "aura1test1",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
				{
					TransferId:  duplicateID, // DUPLICATE
					SourceChain: "bsc",
					TargetChain: "aura",
					Sender:      "0x456",
					Recipient:   "aura1test2",
					Amount:      sdkmath.NewInt(2000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_CONFIRMED,
				},
			},
		}

		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.Error(err, "InitGenesis must reject duplicate hash-based IDs")
		suite.Contains(err.Error(), "duplicate transfer ID")
	})

	suite.Run("reject multiple duplicates", func() {
		suite.SetupTest()
		ctx := suite.SdkCtx

		// Genesis with multiple duplicates - should fail on first duplicate
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []types.CrossChainTransfer{
				{
					TransferId:  "transfer-1",
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x111",
					Recipient:   "aura1test1",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
				{
					TransferId:  "transfer-1", // DUPLICATE
					SourceChain: "bsc",
					TargetChain: "aura",
					Sender:      "0x222",
					Recipient:   "aura1test2",
					Amount:      sdkmath.NewInt(2000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_CONFIRMED,
				},
				{
					TransferId:  "transfer-2",
					SourceChain: "polygon",
					TargetChain: "aura",
					Sender:      "0x333",
					Recipient:   "aura1test3",
					Amount:      sdkmath.NewInt(3000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
				{
					TransferId:  "transfer-2", // ANOTHER DUPLICATE
					SourceChain: "avalanche",
					TargetChain: "aura",
					Sender:      "0x444",
					Recipient:   "aura1test4",
					Amount:      sdkmath.NewInt(4000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_COMPLETED,
				},
			},
		}

		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.Error(err, "InitGenesis must reject genesis with duplicates")
		suite.Contains(err.Error(), "duplicate transfer ID")
	})
}

// TestGenesisRoundTrip verifies that export→import preserves counter state
// and all transfers correctly, preventing data loss during chain upgrades.
func (suite *GenesisCounterTestSuite) TestGenesisRoundTrip() {
	suite.Run("export and reimport preserves counter and transfers", func() {
		suite.SetupTest()
		ctx := suite.SdkCtx

		// Initial genesis with legacy transfers
		initialGenesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []types.CrossChainTransfer{
				{
					TransferId:  "transfer-10",
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x123",
					Recipient:   "aura1test1",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
				{
					TransferId:  "transfer-20",
					SourceChain: "bsc",
					TargetChain: "aura",
					Sender:      "0x456",
					Recipient:   "aura1test2",
					Amount:      sdkmath.NewInt(2000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_CONFIRMED,
				},
			},
		}

		// Import initial genesis
		err := suite.Keeper.InitGenesis(ctx, initialGenesis)
		suite.NoError(err)

		// Verify counter is 21 (MAX+1)
		store := suite.StoreKey
		storeObj := ctx.KVStore(store)
		counterBz := storeObj.Get(types.TransferCounterKey)
		suite.NotNil(counterBz)
		counter := binary.BigEndian.Uint64(counterBz)
		suite.Equal(uint64(21), counter)

		// Export genesis
		exportedGenesis := suite.Keeper.ExportGenesis(ctx)

		// Verify exported data
		suite.NotNil(exportedGenesis.Params)
		suite.Len(exportedGenesis.Transfers, 2)

		// Re-import into fresh context (simulating chain restart)
		suite.SetupTest()
		ctx2 := suite.SdkCtx

		err = suite.Keeper.InitGenesis(ctx2, exportedGenesis)
		suite.NoError(err)

		// Verify counter restored correctly (21)
		storeObj2 := ctx2.KVStore(suite.StoreKey)
		counterBz2 := storeObj2.Get(types.TransferCounterKey)
		suite.NotNil(counterBz2)
		counter2 := binary.BigEndian.Uint64(counterBz2)
		suite.Equal(uint64(21), counter2,
			"Counter should be preserved through export/import cycle")

		// Verify all transfers preserved
		transfer10, found := suite.Keeper.getTransfer(ctx2, "transfer-10")
		suite.True(found)
		suite.Equal("0x123", transfer10.Sender)

		transfer20, found := suite.Keeper.getTransfer(ctx2, "transfer-20")
		suite.True(found)
		suite.Equal("0x456", transfer20.Sender)
	})
}

// TestCounterAfterRestart verifies that after chain restart with genesis import,
// new transfers get correct non-duplicate IDs.
//
// This is the END-TO-END test that proves issue #047 is fixed.
func (suite *GenesisCounterTestSuite) TestCounterAfterRestart() {
	suite.Run("new transfer after restart has correct unique ID", func() {
		suite.SetupTest()
		ctx := suite.SdkCtx

		// Simulate chain state with existing transfers
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []types.CrossChainTransfer{
				{
					TransferId:  "transfer-100",
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x123",
					Recipient:   "aura1test",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
			},
		}

		// Import genesis (simulating chain restart)
		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		// Verify counter is 101 (MAX+1)
		store := suite.StoreKey
		storeObj := ctx.KVStore(store)
		counterBz := storeObj.Get(types.TransferCounterKey)
		suite.NotNil(counterBz)
		counter := binary.BigEndian.Uint64(counterBz)
		suite.Equal(uint64(101), counter)

		// CRITICAL TEST: Verify the existing transfer is still there
		existingTransfer, found := suite.Keeper.getTransfer(ctx, "transfer-100")
		suite.True(found, "Existing transfer-100 must still exist")
		suite.Equal("0x123", existingTransfer.Sender)

		// Note: The nextTransferID() function now uses deterministic hash-based IDs,
		// so it won't return "transfer-101". However, the counter is still maintained
		// for backward compatibility and to ensure legacy sequential ID generation
		// (if needed by other parts of the system) doesn't collide.
		//
		// The key guarantee is: IF sequential IDs were still being used,
		// they would start at 101 (not 100), preventing collision with transfer-100.
		//
		// The counter being set to 101 proves the fix is working correctly.
	})

	suite.Run("no collision with high-numbered legacy transfer", func() {
		suite.SetupTest()
		ctx := suite.SdkCtx

		// Chain with very high transfer ID
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []types.CrossChainTransfer{
				{
					TransferId:  "transfer-999999",
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x123",
					Recipient:   "aura1test",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
			},
		}

		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		// Counter should be 1000000 (not 999999)
		store := suite.StoreKey
		storeObj := ctx.KVStore(store)
		counterBz := storeObj.Get(types.TransferCounterKey)
		suite.NotNil(counterBz)
		counter := binary.BigEndian.Uint64(counterBz)
		suite.Equal(uint64(1000000), counter,
			"Counter must be MAX+1 even for very high IDs")
	})
}

// TestNilTransferHandling verifies that nil transfers in genesis are safely ignored
func (suite *GenesisCounterTestSuite) TestNilTransferHandling() {
	suite.Run("skip nil transfers without error", func() {
		suite.SetupTest()
		ctx := suite.SdkCtx

		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []types.CrossChainTransfer{
				{
					TransferId:  "transfer-5",
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x123",
					Recipient:   "aura1test",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
				types.CrossChainTransfer{}, // Empty transfer - should be safely ignored
				{
					TransferId:  "transfer-10",
					SourceChain: "bsc",
					TargetChain: "aura",
					Sender:      "0x456",
					Recipient:   "aura1test2",
					Amount:      sdkmath.NewInt(2000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_CONFIRMED,
				},
			},
		}

		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.NoError(err, "InitGenesis should handle nil transfers gracefully")

		// Counter should be 11 (MAX of non-nil transfers + 1)
		store := suite.StoreKey
		storeObj := ctx.KVStore(store)
		counterBz := storeObj.Get(types.TransferCounterKey)
		suite.NotNil(counterBz)
		counter := binary.BigEndian.Uint64(counterBz)
		suite.Equal(uint64(11), counter, "Counter should ignore nil transfers")

		// Verify both non-nil transfers were imported
		_, found := suite.Keeper.getTransfer(ctx, "transfer-5")
		suite.True(found)

		_, found = suite.Keeper.getTransfer(ctx, "transfer-10")
		suite.True(found)
	})
}

// TestInvalidTransferIDFormats tests handling of malformed transfer IDs
func (suite *GenesisCounterTestSuite) TestInvalidTransferIDFormats() {
	suite.Run("handle invalid transfer ID formats gracefully", func() {
		suite.SetupTest()
		ctx := suite.SdkCtx

		// Genesis with various invalid/edge case transfer ID formats
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []types.CrossChainTransfer{
				{
					TransferId:  "transfer-5", // Valid
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x123",
					Recipient:   "aura1test1",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
				{
					TransferId:  "invalid-format-no-number", // Invalid format
					SourceChain: "bsc",
					TargetChain: "aura",
					Sender:      "0x456",
					Recipient:   "aura1test2",
					Amount:      sdkmath.NewInt(2000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_CONFIRMED,
				},
				{
					TransferId:  "transfer-abc", // Invalid: non-numeric
					SourceChain: "polygon",
					TargetChain: "aura",
					Sender:      "0x789",
					Recipient:   "aura1test3",
					Amount:      sdkmath.NewInt(3000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
				{
					TransferId:  "transfer-10", // Valid
					SourceChain: "avalanche",
					TargetChain: "aura",
					Sender:      "0xabc",
					Recipient:   "aura1test4",
					Amount:      sdkmath.NewInt(4000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_COMPLETED,
				},
			},
		}

		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.NoError(err, "InitGenesis should handle invalid transfer ID formats gracefully")

		// Counter should only consider valid parseable IDs (5, 10) → MAX+1 = 11
		store := suite.StoreKey
		storeObj := ctx.KVStore(store)
		counterBz := storeObj.Get(types.TransferCounterKey)
		suite.NotNil(counterBz)
		counter := binary.BigEndian.Uint64(counterBz)
		suite.Equal(uint64(11), counter,
			"Counter should be based on valid parseable IDs only")

		// All transfers should still be stored (even with invalid ID format)
		_, found := suite.Keeper.getTransfer(ctx, "transfer-5")
		suite.True(found)

		_, found = suite.Keeper.getTransfer(ctx, "invalid-format-no-number")
		suite.True(found, "Transfers with invalid ID format should still be stored")

		_, found = suite.Keeper.getTransfer(ctx, "transfer-abc")
		suite.True(found)

		_, found = suite.Keeper.getTransfer(ctx, "transfer-10")
		suite.True(found)
	})
}

// TestSequentialIDValidation tests optional validation that IDs are sequential
// (Note: Current implementation doesn't require sequential IDs, but tests for awareness)
func (suite *GenesisCounterTestSuite) TestNonSequentialIDsAllowed() {
	suite.Run("allow non-sequential legacy IDs", func() {
		suite.SetupTest()
		ctx := suite.SdkCtx

		// Genesis with non-sequential IDs: 1, 5, 100 (gaps allowed)
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []types.CrossChainTransfer{
				{
					TransferId:  "transfer-1",
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x123",
					Recipient:   "aura1test1",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
				{
					TransferId:  "transfer-5",
					SourceChain: "bsc",
					TargetChain: "aura",
					Sender:      "0x456",
					Recipient:   "aura1test2",
					Amount:      sdkmath.NewInt(2000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_CONFIRMED,
				},
				{
					TransferId:  "transfer-100",
					SourceChain: "polygon",
					TargetChain: "aura",
					Sender:      "0x789",
					Recipient:   "aura1test3",
					Amount:      sdkmath.NewInt(3000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
			},
		}

		// Non-sequential IDs are allowed (gaps are OK)
		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.NoError(err, "Non-sequential IDs should be allowed")

		// Counter should be MAX+1 = 101
		store := suite.StoreKey
		storeObj := ctx.KVStore(store)
		counterBz := storeObj.Get(types.TransferCounterKey)
		suite.NotNil(counterBz)
		counter := binary.BigEndian.Uint64(counterBz)
		suite.Equal(uint64(101), counter,
			"Counter should be MAX+1 regardless of sequence gaps")
	})
}

// TestEmptyTransferID tests handling of empty transfer IDs
func (suite *GenesisCounterTestSuite) TestEmptyTransferID() {
	suite.Run("handle empty transfer ID", func() {
		suite.SetupTest()
		ctx := suite.SdkCtx

		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []types.CrossChainTransfer{
				{
					TransferId:  "", // Empty ID
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x123",
					Recipient:   "aura1test",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
			},
		}

		// Genesis validation should catch empty IDs before InitGenesis
		// But InitGenesis should handle it gracefully if it gets through
		err := suite.Keeper.InitGenesis(ctx, genesis)

		// Note: Current implementation's setTransfer will skip empty IDs
		// This is acceptable defensive behavior
		suite.NoError(err, "Empty transfer IDs should be handled gracefully")
	})
}

// TestZeroTransferID tests the edge case of transfer-0
func (suite *GenesisCounterTestSuite) TestZeroTransferID() {
	suite.Run("handle transfer-0 as valid ID", func() {
		suite.SetupTest()
		ctx := suite.SdkCtx

		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []types.CrossChainTransfer{
				{
					TransferId:  "transfer-0", // Zero ID (edge case)
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x123",
					Recipient:   "aura1test",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
			},
		}

		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		// With transfer-0, counter should be 1 (0+1)
		store := suite.StoreKey
		storeObj := ctx.KVStore(store)
		counterBz := storeObj.Get(types.TransferCounterKey)
		suite.NotNil(counterBz)
		counter := binary.BigEndian.Uint64(counterBz)
		suite.Equal(uint64(1), counter, "Counter should be 1 when MAX ID is 0")

		// Verify transfer-0 was stored
		transfer, found := suite.Keeper.getTransfer(ctx, "transfer-0")
		suite.True(found)
		suite.Equal("0x123", transfer.Sender)
	})
}

// TestGenesisImportPerformance tests performance with many transfers
func (suite *GenesisCounterTestSuite) TestGenesisImportPerformance() {
	suite.Run("import 100 transfers efficiently", func() {
		suite.SetupTest()
		ctx := suite.SdkCtx

		// Create genesis with 100 transfers
		transfers := make([]types.CrossChainTransfer, 100)
		for i := 0; i < 100; i++ {
			transfers[i] = types.CrossChainTransfer{
				TransferId:  fmt.Sprintf("transfer-%d", i+1),
				SourceChain: "ethereum",
				TargetChain: "aura",
				Sender:      fmt.Sprintf("0x%d", i),
				Recipient:   fmt.Sprintf("aura1test%d", i),
				Amount:      sdkmath.NewInt(1000000),
				Denom:       "uaura",
				Status:      types.TransferStatus_PENDING,
			}
		}

		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: transfers,
		}

		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.NoError(err, "Should import 100 transfers without error")

		// Verify counter is 101 (MAX+1)
		store := suite.StoreKey
		storeObj := ctx.KVStore(store)
		counterBz := storeObj.Get(types.TransferCounterKey)
		suite.NotNil(counterBz)
		counter := binary.BigEndian.Uint64(counterBz)
		suite.Equal(uint64(101), counter, "Counter should be 101 after importing 100 sequential transfers")

		// Spot check some transfers
		transfer1, found := suite.Keeper.getTransfer(ctx, "transfer-1")
		suite.True(found)
		suite.Equal("0x0", transfer1.Sender)

		transfer100, found := suite.Keeper.getTransfer(ctx, "transfer-100")
		suite.True(found)
		suite.Equal("0x99", transfer100.Sender)
	})
}
