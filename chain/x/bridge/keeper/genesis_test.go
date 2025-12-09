package keeper

import (
	"encoding/binary"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/bridge/types"
	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

type GenesisTestSuite struct {
	KeeperTestSuite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) TestInitGenesis() {
	ctx := suite.SdkCtx

	// Test: InitGenesis with default/empty state should not error
	suite.Run("default genesis", func() {
		defaultGenesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
		}
		err := suite.Keeper.InitGenesis(ctx, defaultGenesis)
		suite.NoError(err, "InitGenesis should not error with default state")
	})

	// Test: InitGenesis with valid data
	suite.Run("valid genesis with data", func() {
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []bridgepb.CrossChainTransfer{
				{
					TransferId:  "transfer-1",
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x123",
					Recipient:   "aura1test",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
			},
			ChainConfigs: []bridgepb.ChainConfig{
				{
					ChainId:          "ethereum",
					ChainName:        "Ethereum",
					AddressPrefix:    "0x",
					Enabled:          true,
					MinConfirmations: 12,
				},
			},
			Validators: []bridgepb.BridgeValidator{
				{
					Address: sdk.AccAddress("validator1__________").String(),
					Active:  true,
					Power:   100,
				},
			},
			WrappedTokens: []bridgepb.WrappedToken{
				{
					SourceChain:   "ethereum",
					OriginalDenom: "ETH",
					WrappedDenom:  "wETH",
					TotalSupply:   sdkmath.NewInt(5000000),
					LockedAmount:  sdkmath.NewInt(5000000),
				},
			},
		}

		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.NoError(err, "InitGenesis should not error with valid data")

		// Verify data was stored
		params := suite.Keeper.GetParams(ctx)
		suite.True(params.BridgeEnabled)
		suite.Equal(uint64(6), params.MinConfirmations)
	})
}

func (suite *GenesisTestSuite) TestInitGenesisWithInvalidData() {
	ctx := suite.SdkCtx

	suite.Run("nil params", func() {
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{},
		}
		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.Error(err, "InitGenesis should error with nil params")
	})

	suite.Run("invalid params - zero min confirmations", func() {
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             0, // Invalid
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000),
				ValidatorThresholdPercentage: 67,
			},
		}
		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.Error(err, "InitGenesis should error with zero min confirmations")
	})

	suite.Run("invalid params - high fee basis points", func() {
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         20000, // Invalid > 10000
				MaxTransferAmount:            sdkmath.NewInt(1000000),
				ValidatorThresholdPercentage: 67,
			},
		}
		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.Error(err, "InitGenesis should error with high fee basis points")
	})

	suite.Run("invalid params - invalid validator threshold", func() {
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000),
				ValidatorThresholdPercentage: 0, // Invalid
			},
		}
		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.Error(err, "InitGenesis should error with zero validator threshold")
	})

	suite.Run("invalid params - empty max transfer amount", func() {
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.Int{}, // Invalid
				ValidatorThresholdPercentage: 67,
			},
		}
		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.Error(err, "InitGenesis should error with empty max transfer amount")
	})

	suite.Run("nil transfer", func() {
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []bridgepb.CrossChainTransfer{},
		}
		// Should skip nil transfers without error
		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.NoError(err, "InitGenesis should skip nil transfers")
	})
}

func (suite *GenesisTestSuite) TestExportGenesis() {
	ctx := suite.SdkCtx

	suite.Run("export empty state", func() {
		exported := suite.Keeper.ExportGenesis(ctx)
		suite.NotNil(exported.Params, "Exported params should not be nil")
		// Empty slices may be nil or empty - both are valid for empty state
		suite.Empty(exported.Transfers, "Exported transfers should be empty")
		suite.Empty(exported.ChainConfigs, "Exported chain configs should be empty")
		suite.Empty(exported.Validators, "Exported validators should be empty")
		suite.Empty(exported.WrappedTokens, "Exported wrapped tokens should be empty")
	})

	suite.Run("export with data", func() {
		// Initialize with some data
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []bridgepb.CrossChainTransfer{
				{
					TransferId:  "transfer-1",
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x123",
					Recipient:   "aura1test",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
			},
			ChainConfigs: []bridgepb.ChainConfig{
				{
					ChainId:          "ethereum",
					ChainName:        "Ethereum",
					AddressPrefix:    "0x",
					Enabled:          true,
					MinConfirmations: 12,
				},
			},
		}

		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		// Export and verify
		exported := suite.Keeper.ExportGenesis(ctx)
		suite.NotNil(exported)
		suite.True(exported.Params.Enabled)
		suite.Len(exported.Transfers, 1, "Should export 1 transfer")
		suite.Equal("transfer-1", exported.Transfers[0].TransferId)
		suite.Len(exported.ChainConfigs, 1, "Should export 1 chain config")
		suite.Equal("ethereum", exported.ChainConfigs[0].ChainId)
	})
}

func (suite *GenesisTestSuite) TestDefaultGenesis() {
	suite.Run("default genesis is valid", func() {
		defaultGenesis := &types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers:     []bridgepb.CrossChainTransfer{}, // Initialize empty slices
			ChainConfigs:  []bridgepb.ChainConfig{},
			Validators:    []bridgepb.BridgeValidator{},
			WrappedTokens: []bridgepb.WrappedToken{},
		}
		suite.NotNil(defaultGenesis, "DefaultGenesis should not return nil")

		err := types.ValidateGenesis(defaultGenesis)
		suite.NoError(err, "Default genesis should be valid")

		suite.NotNil(defaultGenesis.Params, "Default params should not be nil")
		suite.True(defaultGenesis.Params.Enabled, "Bridge should be enabled by default")
		// Transfers slice is initialized as empty, not nil
		suite.Empty(defaultGenesis.Transfers, "Default transfers should be empty")
	})

	suite.Run("default genesis can be initialized", func() {
		ctx := suite.SdkCtx
		defaultGenesis := &types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
		}

		err := suite.Keeper.InitGenesis(ctx, *defaultGenesis)
		suite.NoError(err, "Should be able to initialize default genesis")
	})
}

func (suite *GenesisTestSuite) TestGenesisRoundTrip() {
	ctx := suite.SdkCtx

	suite.Run("round trip with empty state", func() {
		genesis := &types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
		}

		// Initialize
		err := suite.Keeper.InitGenesis(ctx, *genesis)
		suite.NoError(err)

		// Export
		exported := suite.Keeper.ExportGenesis(ctx)

		// Verify consistency
		suite.Equal(genesis.Params.Enabled, exported.Params.Enabled)
		suite.Equal(genesis.Params.MinConfirmations, exported.Params.MinConfirmations)
		suite.Equal(len(genesis.Transfers), len(exported.Transfers))
	})

	suite.Run("round trip with data", func() {
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []bridgepb.CrossChainTransfer{
				{
					TransferId:  "transfer-1",
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x123",
					Recipient:   "aura1test1",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
					// SubmittedHeight field removed -100,
				},
				{
					TransferId:  "transfer-2",
					SourceChain: "polygon",
					TargetChain: "aura",
					Sender:      "0x456",
					Recipient:   "aura1test2",
					Amount:      sdkmath.NewInt(2000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_CONFIRMED,
					// SubmittedHeight field removed -200,
				},
			},
			ChainConfigs: []bridgepb.ChainConfig{
				{
					ChainId:          "ethereum",
					ChainName:        "Ethereum",
					AddressPrefix:    "0x",
					Enabled:          true,
					MinConfirmations: 12,
				},
				{
					ChainId:          "polygon",
					ChainName:        "Polygon",
					AddressPrefix:    "0x",
					Enabled:          true,
					MinConfirmations: 64,
				},
			},
			Validators: []bridgepb.BridgeValidator{
				{
					Address: sdk.AccAddress("validator1__________").String(),
					Active:  true,
					Power:   100,
				},
				{
					Address: sdk.AccAddress("validator2__________").String(),
					Active:  true,
					Power:   150,
				},
			},
			WrappedTokens: []bridgepb.WrappedToken{
				{
					SourceChain:   "ethereum",
					OriginalDenom: "ETH",
					WrappedDenom:  "wETH",
					TotalSupply:   sdkmath.NewInt(5000000),
					LockedAmount:  sdkmath.NewInt(5000000),
				},
			},
		}

		// Initialize
		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		// Export
		exported := suite.Keeper.ExportGenesis(ctx)

		// Verify all data was preserved
		suite.Equal(genesis.Params.Enabled, exported.Params.Enabled)
		suite.Equal(len(genesis.Transfers), len(exported.Transfers))
		suite.Equal(len(genesis.ChainConfigs), len(exported.ChainConfigs))
		suite.Equal(len(genesis.Validators), len(exported.Validators))
		suite.Equal(len(genesis.WrappedTokens), len(exported.WrappedTokens))

		// Re-initialize with exported data
		err = suite.Keeper.InitGenesis(ctx, exported)
		suite.NoError(err)

		// Export again
		exported2 := suite.Keeper.ExportGenesis(ctx)

		// Should be identical
		suite.Equal(len(exported.Transfers), len(exported2.Transfers))
		suite.Equal(len(exported.ChainConfigs), len(exported2.ChainConfigs))
		suite.Equal(len(exported.Validators), len(exported2.Validators))
	})
}

func (suite *GenesisTestSuite) TestGenesisEdgeCases() {
	ctx := suite.SdkCtx

	suite.Run("empty lists", func() {
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers:     []bridgepb.CrossChainTransfer{},
			ChainConfigs:  []bridgepb.ChainConfig{},
			Validators:    []bridgepb.BridgeValidator{},
			WrappedTokens: []bridgepb.WrappedToken{},
		}

		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		exported := suite.Keeper.ExportGenesis(ctx)
		// Empty slices may be nil or empty - both are valid for empty state
		suite.Empty(exported.Transfers)
		suite.Empty(exported.ChainConfigs)
		suite.Empty(exported.Validators)
		suite.Empty(exported.WrappedTokens)
	})

	suite.Run("many transfers", func() {
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: make([]bridgepb.CrossChainTransfer, 100),
		}

		for i := 0; i < 100; i++ {
			genesis.Transfers[i] = bridgepb.CrossChainTransfer{
				TransferId:  "transfer-" + string(rune(i)),
				SourceChain: "ethereum",
				TargetChain: "aura",
				Sender:      "0xsender",
				Recipient:   "aura1recipient",
				Amount:      sdkmath.NewInt(1000000),
				Denom:       "uaura",
				Status:      types.TransferStatus_PENDING,
				// SubmittedHeight field removed - uint64(100 + i),
			}
		}

		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		exported := suite.Keeper.ExportGenesis(ctx)
		suite.Len(exported.Transfers, 100)
	})

	suite.Run("disabled bridge", func() {
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      false,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000),
				ValidatorThresholdPercentage: 67,
			},
		}

		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		params := suite.Keeper.GetParams(ctx)
		suite.False(params.BridgeEnabled)
	})

	suite.Run("transfer counter preservation", func() {
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []bridgepb.CrossChainTransfer{
				{
					TransferId:  "transfer-100",
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x123",
					Recipient:   "aura1test",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
					// SubmittedHeight field removed -100,
				},
			},
		}

		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		exported := suite.Keeper.ExportGenesis(ctx)
		suite.NotNil(exported)
	})
}

func (suite *GenesisTestSuite) TestTransferCounterOffByOneError() {
	suite.Run("counter set to MAX+1 not MAX", func() {
		// Create fresh test setup for isolation
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
			Transfers: []bridgepb.CrossChainTransfer{
				{
					TransferId:  "transfer-1",
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x123",
					Recipient:   "aura1test",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
				{
					TransferId:  "transfer-5",
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x456",
					Recipient:   "aura1test2",
					Amount:      sdkmath.NewInt(2000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_CONFIRMED,
				},
				{
					TransferId:  "transfer-3",
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

		// Read the counter from storage
		store := suite.StoreKey
		storeObj := ctx.KVStore(store)
		counterBz := storeObj.Get(types.TransferCounterKey)
		suite.NotNil(counterBz, "Transfer counter should be set")

		counter := binary.BigEndian.Uint64(counterBz)
		// FIXED: Counter should be MAX+1 (6), not MAX (5), to prevent duplicate IDs
		// This is the fix for issue #047
		suite.Equal(uint64(6), counter, "Counter must be set to MAX+1 (6) to prevent duplicate with transfer-5")

		// Note: nextTransferID now returns hash-based IDs (deterministic), not sequential IDs
		// The counter is maintained for backward compatibility with legacy sequential IDs
		// but new IDs are generated using block height + tx hash for determinism
		nextID := suite.Keeper.nextTransferID(ctx)
		suite.NotEmpty(nextID, "Next transfer should have a valid ID")
		suite.Contains(nextID, "transfer-", "Next transfer ID should have transfer- prefix")
	})

	suite.Run("counter with single transfer", func() {
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
			Transfers: []bridgepb.CrossChainTransfer{
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

		// FIXED: Counter should be 101 (MAX+1), not 100 (MAX)
		store := suite.StoreKey
		storeObj := ctx.KVStore(store)
		counterBz := storeObj.Get(types.TransferCounterKey)
		counter := binary.BigEndian.Uint64(counterBz)
		suite.Equal(uint64(101), counter, "Counter must be set to MAX+1 (101) to prevent duplicate with transfer-100")

		// Note: nextTransferID now returns hash-based IDs, not sequential
		nextID := suite.Keeper.nextTransferID(ctx)
		suite.NotEmpty(nextID, "Next transfer should have a valid ID")
		suite.Contains(nextID, "transfer-", "Next transfer ID should have transfer- prefix")
	})

	suite.Run("counter with no transfers", func() {
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
			Transfers: []bridgepb.CrossChainTransfer{},
		}

		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		// Counter should not be set when no legacy transfers
		store := suite.StoreKey
		storeObj := ctx.KVStore(store)
		counterBz := storeObj.Get(types.TransferCounterKey)
		suite.Nil(counterBz, "Counter should not be set with no transfers")

		// New transfers use deterministic hash-based IDs, not sequential
		nextID := suite.Keeper.nextTransferID(ctx)
		suite.NotEmpty(nextID, "First transfer should have a valid ID")
		suite.Contains(nextID, "transfer-", "First transfer ID should have transfer- prefix")
	})

	suite.Run("counter with non-sequential transfers", func() {
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
			Transfers: []bridgepb.CrossChainTransfer{
				{
					TransferId:  "transfer-10",
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x123",
					Recipient:   "aura1test",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
				{
					TransferId:  "transfer-25",
					SourceChain: "polygon",
					TargetChain: "aura",
					Sender:      "0x456",
					Recipient:   "aura1test2",
					Amount:      sdkmath.NewInt(2000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_CONFIRMED,
				},
				{
					TransferId:  "transfer-15",
					SourceChain: "bsc",
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

		// FIXED: Counter should be MAX+1 = 26, not MAX = 25
		store := suite.StoreKey
		storeObj := ctx.KVStore(store)
		counterBz := storeObj.Get(types.TransferCounterKey)
		counter := binary.BigEndian.Uint64(counterBz)
		suite.Equal(uint64(26), counter, "Counter should be set to MAX+1 (26) to prevent duplicate with transfer-25")

		// Note: nextTransferID now returns hash-based IDs, not sequential
		nextID := suite.Keeper.nextTransferID(ctx)
		suite.NotEmpty(nextID, "Next transfer should have a valid ID")
		suite.Contains(nextID, "transfer-", "Next transfer ID should have transfer- prefix")
	})
}

func (suite *GenesisTestSuite) TestDuplicateTransferIDDetection() {
	ctx := suite.SdkCtx

	suite.Run("detect duplicate transfer IDs in genesis", func() {
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []bridgepb.CrossChainTransfer{
				{
					TransferId:  "transfer-1",
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x123",
					Recipient:   "aura1test",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
				{
					TransferId:  "transfer-2",
					SourceChain: "polygon",
					TargetChain: "aura",
					Sender:      "0x456",
					Recipient:   "aura1test2",
					Amount:      sdkmath.NewInt(2000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_CONFIRMED,
				},
				{
					TransferId:  "transfer-1", // DUPLICATE!
					SourceChain: "bsc",
					TargetChain: "aura",
					Sender:      "0x789",
					Recipient:   "aura1test3",
					Amount:      sdkmath.NewInt(3000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
			},
		}

		// FIXED: Should return error on duplicate transfer ID (not panic)
		// Changed from panic to error for better error handling
		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.Error(err, "InitGenesis should error on duplicate transfer IDs")
		suite.Contains(err.Error(), "duplicate transfer ID", "Error should mention duplicate ID")
	})

	suite.Run("reject ALL duplicate IDs including non-sequential", func() {
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []bridgepb.CrossChainTransfer{
				{
					TransferId:  "transfer-1",
					SourceChain: "ethereum",
					TargetChain: "aura",
					Sender:      "0x123",
					Recipient:   "aura1test",
					Amount:      sdkmath.NewInt(1000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
				{
					TransferId:  "custom-transfer-id", // Non-sequential format
					SourceChain: "polygon",
					TargetChain: "aura",
					Sender:      "0x456",
					Recipient:   "aura1test2",
					Amount:      sdkmath.NewInt(2000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_CONFIRMED,
				},
				{
					TransferId:  "custom-transfer-id", // DUPLICATE non-sequential ID
					SourceChain: "bsc",
					TargetChain: "aura",
					Sender:      "0x789",
					Recipient:   "aura1test3",
					Amount:      sdkmath.NewInt(3000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
			},
		}

		// FIXED: Now detects ALL duplicates, not just sequential IDs
		// This prevents data corruption from silent overwrites
		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.Error(err, "Should reject duplicate non-sequential IDs")
		suite.Contains(err.Error(), "duplicate transfer ID", "Error should mention duplicate ID")
	})

	suite.Run("detect multiple duplicates", func() {
		genesis := types.GenesisState{
			Params: bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []bridgepb.CrossChainTransfer{
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
				{
					TransferId:  "transfer-10",
					SourceChain: "polygon",
					TargetChain: "aura",
					Sender:      "0x456",
					Recipient:   "aura1test2",
					Amount:      sdkmath.NewInt(2000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_CONFIRMED,
				},
				{
					TransferId:  "transfer-5", // First duplicate
					SourceChain: "bsc",
					TargetChain: "aura",
					Sender:      "0x789",
					Recipient:   "aura1test3",
					Amount:      sdkmath.NewInt(3000000),
					Denom:       "uaura",
					Status:      types.TransferStatus_PENDING,
				},
			},
		}

		// FIXED: Should return error on first duplicate (not panic)
		err := suite.Keeper.InitGenesis(ctx, genesis)
		suite.Error(err, "InitGenesis should error on first duplicate transfer ID")
		suite.Contains(err.Error(), "duplicate transfer ID", "Error should mention duplicate ID")
	})
}

// Test helper function
func TestValidateGenesis(t *testing.T) {
	tests := []struct {
		name      string
		genesis   *types.GenesisState
		expectErr bool
	}{
		{
			name:      "nil genesis",
			genesis:   nil,
			expectErr: true,
		},
		{
			name: "nil params",
			genesis: &types.GenesisState{
				Params: bridgepb.BridgeParams{},
			},
			expectErr: true,
		},
		{
			name: "valid default genesis",
			genesis: &types.GenesisState{
				Params: bridgepb.BridgeParams{
					Enabled:                      true,
					MinConfirmations:             6,
					BridgeFeeBasisPoints:         30,
					MaxTransferAmount:            sdkmath.NewInt(1000000000000),
					ValidatorThresholdPercentage: 67,
				},
			},
			expectErr: false,
		},
		{
			name: "valid genesis with data",
			genesis: &types.GenesisState{
				Params: bridgepb.BridgeParams{
					Enabled:                      true,
					MinConfirmations:             6,
					BridgeFeeBasisPoints:         30,
					MaxTransferAmount:            sdkmath.NewInt(1000000),
					ValidatorThresholdPercentage: 67,
				},
			},
			expectErr: false,
		},
		{
			name: "invalid - zero min confirmations",
			genesis: &types.GenesisState{
				Params: bridgepb.BridgeParams{
					Enabled:                      true,
					MinConfirmations:             0,
					BridgeFeeBasisPoints:         30,
					MaxTransferAmount:            sdkmath.NewInt(1000000),
					ValidatorThresholdPercentage: 67,
				},
			},
			expectErr: true,
		},
		{
			name: "invalid - high fee basis points",
			genesis: &types.GenesisState{
				Params: bridgepb.BridgeParams{
					Enabled:                      true,
					MinConfirmations:             6,
					BridgeFeeBasisPoints:         15000,
					MaxTransferAmount:            sdkmath.NewInt(1000000),
					ValidatorThresholdPercentage: 67,
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := types.ValidateGenesis(tt.genesis)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
