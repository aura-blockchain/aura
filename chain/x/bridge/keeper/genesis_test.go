package keeper_test

import (
	"testing"

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
		defaultGenesis := types.DefaultGenesisState()
		err := suite.keeper.InitGenesis(ctx, *defaultGenesis)
		suite.NoError(err, "InitGenesis should not error with default state")
	})

	// Test: InitGenesis with valid data
	suite.Run("valid genesis with data", func() {
		genesis := types.GenesisState{
			Params: &bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            "1000000000000",
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []*bridgepb.Transfer{
				{
					TransferId:        "transfer-1",
					SourceChain:       "ethereum",
					DestinationChain:  "aura",
					Sender:            "0x123",
					Recipient:         "aura1test",
					Amount:            "1000000",
					Status:            "pending",
					SubmittedHeight:   100,
				},
			},
			ChainConfigs: []*bridgepb.ChainConfig{
				{
					ChainId:          "ethereum",
					ChainName:        "Ethereum",
					Enabled:          true,
					MinConfirmations: 12,
				},
			},
			Validators: []*bridgepb.BridgeValidator{
				{
					Address: "aura1validator",
					Active:  true,
				},
			},
			WrappedTokens: []*bridgepb.WrappedToken{
				{
					SourceChain:  "ethereum",
					SourceDenom:  "ETH",
					WrappedDenom: "wETH",
					TotalSupply:  "5000000",
				},
			},
		}

		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err, "InitGenesis should not error with valid data")

		// Verify data was stored
		params := suite.keeper.GetParams(ctx)
		suite.True(params.BridgeEnabled)
		suite.Equal(uint32(6), params.MinConfirmations)
	})
}

func (suite *GenesisTestSuite) TestInitGenesisWithInvalidData() {
	ctx := suite.SdkCtx

	suite.Run("nil params", func() {
		genesis := types.GenesisState{
			Params: nil,
		}
		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.Error(err, "InitGenesis should error with nil params")
	})

	suite.Run("invalid params - zero min confirmations", func() {
		genesis := types.GenesisState{
			Params: &bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             0, // Invalid
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            "1000000",
				ValidatorThresholdPercentage: 67,
			},
		}
		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.Error(err, "InitGenesis should error with zero min confirmations")
	})

	suite.Run("invalid params - high fee basis points", func() {
		genesis := types.GenesisState{
			Params: &bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         20000, // Invalid > 10000
				MaxTransferAmount:            "1000000",
				ValidatorThresholdPercentage: 67,
			},
		}
		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.Error(err, "InitGenesis should error with high fee basis points")
	})

	suite.Run("invalid params - invalid validator threshold", func() {
		genesis := types.GenesisState{
			Params: &bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            "1000000",
				ValidatorThresholdPercentage: 0, // Invalid
			},
		}
		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.Error(err, "InitGenesis should error with zero validator threshold")
	})

	suite.Run("invalid params - empty max transfer amount", func() {
		genesis := types.GenesisState{
			Params: &bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            "", // Invalid
				ValidatorThresholdPercentage: 67,
			},
		}
		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.Error(err, "InitGenesis should error with empty max transfer amount")
	})

	suite.Run("nil transfer", func() {
		genesis := types.GenesisState{
			Params: &bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            "1000000",
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []*bridgepb.Transfer{nil},
		}
		// Should skip nil transfers without error
		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err, "InitGenesis should skip nil transfers")
	})
}

func (suite *GenesisTestSuite) TestExportGenesis() {
	ctx := suite.SdkCtx

	suite.Run("export empty state", func() {
		exported := suite.keeper.ExportGenesis(ctx)
		suite.NotNil(exported.Params, "Exported params should not be nil")
		suite.NotNil(exported.Transfers, "Exported transfers should not be nil")
		suite.NotNil(exported.ChainConfigs, "Exported chain configs should not be nil")
		suite.NotNil(exported.Validators, "Exported validators should not be nil")
		suite.NotNil(exported.WrappedTokens, "Exported wrapped tokens should not be nil")
	})

	suite.Run("export with data", func() {
		// Initialize with some data
		genesis := types.GenesisState{
			Params: &bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            "1000000000000",
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []*bridgepb.Transfer{
				{
					TransferId:       "transfer-1",
					SourceChain:      "ethereum",
					DestinationChain: "aura",
					Sender:           "0x123",
					Recipient:        "aura1test",
					Amount:           "1000000",
					Status:           "pending",
					SubmittedHeight:  100,
				},
			},
			ChainConfigs: []*bridgepb.ChainConfig{
				{
					ChainId:          "ethereum",
					ChainName:        "Ethereum",
					Enabled:          true,
					MinConfirmations: 12,
				},
			},
		}

		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		// Export and verify
		exported := suite.keeper.ExportGenesis(ctx)
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
		defaultGenesis := types.DefaultGenesisState()
		suite.NotNil(defaultGenesis, "DefaultGenesis should not return nil")

		err := types.ValidateGenesis(defaultGenesis)
		suite.NoError(err, "Default genesis should be valid")

		suite.NotNil(defaultGenesis.Params, "Default params should not be nil")
		suite.True(defaultGenesis.Params.Enabled, "Bridge should be enabled by default")
		suite.NotNil(defaultGenesis.Transfers, "Default transfers should not be nil")
		suite.Empty(defaultGenesis.Transfers, "Default transfers should be empty")
	})

	suite.Run("default genesis can be initialized", func() {
		ctx := suite.SdkCtx
		defaultGenesis := types.DefaultGenesisState()

		err := suite.keeper.InitGenesis(ctx, *defaultGenesis)
		suite.NoError(err, "Should be able to initialize default genesis")
	})
}

func (suite *GenesisTestSuite) TestGenesisRoundTrip() {
	ctx := suite.SdkCtx

	suite.Run("round trip with empty state", func() {
		genesis := types.DefaultGenesisState()

		// Initialize
		err := suite.keeper.InitGenesis(ctx, *genesis)
		suite.NoError(err)

		// Export
		exported := suite.keeper.ExportGenesis(ctx)

		// Verify consistency
		suite.Equal(genesis.Params.Enabled, exported.Params.Enabled)
		suite.Equal(genesis.Params.MinConfirmations, exported.Params.MinConfirmations)
		suite.Equal(len(genesis.Transfers), len(exported.Transfers))
	})

	suite.Run("round trip with data", func() {
		genesis := types.GenesisState{
			Params: &bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            "1000000000000",
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []*bridgepb.Transfer{
				{
					TransferId:       "transfer-1",
					SourceChain:      "ethereum",
					DestinationChain: "aura",
					Sender:           "0x123",
					Recipient:        "aura1test1",
					Amount:           "1000000",
					Status:           "pending",
					SubmittedHeight:  100,
				},
				{
					TransferId:       "transfer-2",
					SourceChain:      "polygon",
					DestinationChain: "aura",
					Sender:           "0x456",
					Recipient:        "aura1test2",
					Amount:           "2000000",
					Status:           "completed",
					SubmittedHeight:  200,
				},
			},
			ChainConfigs: []*bridgepb.ChainConfig{
				{
					ChainId:          "ethereum",
					ChainName:        "Ethereum",
					Enabled:          true,
					MinConfirmations: 12,
				},
				{
					ChainId:          "polygon",
					ChainName:        "Polygon",
					Enabled:          true,
					MinConfirmations: 64,
				},
			},
			Validators: []*bridgepb.BridgeValidator{
				{
					Address: "aura1validator1",
					Active:  true,
				},
				{
					Address: "aura1validator2",
					Active:  true,
				},
			},
			WrappedTokens: []*bridgepb.WrappedToken{
				{
					SourceChain:  "ethereum",
					SourceDenom:  "ETH",
					WrappedDenom: "wETH",
					TotalSupply:  "5000000",
				},
			},
		}

		// Initialize
		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		// Export
		exported := suite.keeper.ExportGenesis(ctx)

		// Verify all data was preserved
		suite.Equal(genesis.Params.Enabled, exported.Params.Enabled)
		suite.Equal(len(genesis.Transfers), len(exported.Transfers))
		suite.Equal(len(genesis.ChainConfigs), len(exported.ChainConfigs))
		suite.Equal(len(genesis.Validators), len(exported.Validators))
		suite.Equal(len(genesis.WrappedTokens), len(exported.WrappedTokens))

		// Re-initialize with exported data
		err = suite.keeper.InitGenesis(ctx, exported)
		suite.NoError(err)

		// Export again
		exported2 := suite.keeper.ExportGenesis(ctx)

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
			Params: &bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            "1000000",
				ValidatorThresholdPercentage: 67,
			},
			Transfers:     []*bridgepb.Transfer{},
			ChainConfigs:  []*bridgepb.ChainConfig{},
			Validators:    []*bridgepb.BridgeValidator{},
			WrappedTokens: []*bridgepb.WrappedToken{},
		}

		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		exported := suite.keeper.ExportGenesis(ctx)
		suite.NotNil(exported.Transfers)
		suite.NotNil(exported.ChainConfigs)
		suite.NotNil(exported.Validators)
		suite.NotNil(exported.WrappedTokens)
	})

	suite.Run("many transfers", func() {
		genesis := types.GenesisState{
			Params: &bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            "1000000000000",
				ValidatorThresholdPercentage: 67,
			},
			Transfers: make([]*bridgepb.Transfer, 100),
		}

		for i := 0; i < 100; i++ {
			genesis.Transfers[i] = &bridgepb.Transfer{
				TransferId:       "transfer-" + string(rune(i)),
				SourceChain:      "ethereum",
				DestinationChain: "aura",
				Sender:           "0xsender",
				Recipient:        "aura1recipient",
				Amount:           "1000000",
				Status:           "pending",
				SubmittedHeight:  uint64(100 + i),
			}
		}

		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		exported := suite.keeper.ExportGenesis(ctx)
		suite.Len(exported.Transfers, 100)
	})

	suite.Run("disabled bridge", func() {
		genesis := types.GenesisState{
			Params: &bridgepb.BridgeParams{
				Enabled:                      false,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            "1000000",
				ValidatorThresholdPercentage: 67,
			},
		}

		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		params := suite.keeper.GetParams(ctx)
		suite.False(params.BridgeEnabled)
	})

	suite.Run("transfer counter preservation", func() {
		genesis := types.GenesisState{
			Params: &bridgepb.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             6,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            "1000000",
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []*bridgepb.Transfer{
				{
					TransferId:       "transfer-100",
					SourceChain:      "ethereum",
					DestinationChain: "aura",
					Sender:           "0x123",
					Recipient:        "aura1test",
					Amount:           "1000000",
					Status:           "pending",
					SubmittedHeight:  100,
				},
			},
		}

		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		exported := suite.keeper.ExportGenesis(ctx)
		suite.NotNil(exported)
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
				Params: nil,
			},
			expectErr: true,
		},
		{
			name:      "valid default genesis",
			genesis:   types.DefaultGenesisState(),
			expectErr: false,
		},
		{
			name: "valid genesis with data",
			genesis: &types.GenesisState{
				Params: &bridgepb.BridgeParams{
					Enabled:                      true,
					MinConfirmations:             6,
					BridgeFeeBasisPoints:         30,
					MaxTransferAmount:            "1000000",
					ValidatorThresholdPercentage: 67,
				},
			},
			expectErr: false,
		},
		{
			name: "invalid - zero min confirmations",
			genesis: &types.GenesisState{
				Params: &bridgepb.BridgeParams{
					Enabled:                      true,
					MinConfirmations:             0,
					BridgeFeeBasisPoints:         30,
					MaxTransferAmount:            "1000000",
					ValidatorThresholdPercentage: 67,
				},
			},
			expectErr: true,
		},
		{
			name: "invalid - high fee basis points",
			genesis: &types.GenesisState{
				Params: &bridgepb.BridgeParams{
					Enabled:                      true,
					MinConfirmations:             6,
					BridgeFeeBasisPoints:         15000,
					MaxTransferAmount:            "1000000",
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
