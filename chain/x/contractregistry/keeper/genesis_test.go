// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/contractregistry/types"
	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
)

type GenesisTestSuite struct {
	KeeperTestSuite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) TestInitGenesis() {
	ctx := suite.ctx

	suite.Run("default genesis", func() {
		defaultGenesis := &pb.GenesisState{
			Params:    *types.DefaultParams(),
			Contracts: []pb.ContractInfo{},
			Metrics:   []pb.ContractMetrics{},
		}
		err := suite.keeper.InitGenesis(ctx, defaultGenesis)
		suite.NoError(err, "InitGenesis should not error with default state")
	})

	suite.Run("valid genesis with data", func() {
		genesis := &pb.GenesisState{
			Params: *types.DefaultParams(),
			Contracts: []pb.ContractInfo{
				{
					Address:   "aura1contract1",
					Creator:   "aura1creator",
					CodeId:    1,
					CreatedAt: time.Now(),
					Status:    pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
					Metadata: pb.ContractMetadata{
						Name:        "Test Contract",
						Description: "A test contract",
						Version:     "1.0.0",
						Tags:        []string{"test", "demo"},
					},
				},
			},
			Metrics: []pb.ContractMetrics{
				{
					ContractAddress:      "aura1contract1",
					TotalExecutions:      100,
					TotalGasUsed:         50000,
					SuccessfulExecutions: 95,
					FailedExecutions:     5,
				},
			},
		}

		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err, "InitGenesis should not error with valid data")

		// Verify data was stored
		info, found := suite.keeper.GetContractInfo(ctx, "aura1contract1")
		suite.True(found, "Contract info should be found")
		suite.Equal("Test Contract", info.Metadata.Name)
	})
}

func (suite *GenesisTestSuite) TestInitGenesisWithInvalidData() {
	ctx := suite.ctx

	suite.Run("nil genesis", func() {
		err := suite.keeper.InitGenesis(ctx, nil)
		suite.Error(err, "InitGenesis should error with nil state")
	})

	suite.Run("invalid params", func() {
		genesis := &pb.GenesisState{
			Params: pb.ContractRegistryParams{
				MaxContractsPerCreator: 20000, // exceeds limit
			},
		}
		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.Error(err, "InitGenesis should error with invalid params")
	})

	suite.Run("invalid contract", func() {
		genesis := &pb.GenesisState{
			Params: *types.DefaultParams(),
			Contracts: []pb.ContractInfo{
				{
					Address: "", // invalid: empty address
					CodeId:  1,
					Creator: "aura1creator",
					Status:  pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
				},
			},
		}
		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.Error(err, "InitGenesis should error with invalid contract")
	})

	suite.Run("invalid metrics", func() {
		genesis := &pb.GenesisState{
			Params: *types.DefaultParams(),
			Metrics: []pb.ContractMetrics{
				{
					ContractAddress: "", // invalid: empty address
				},
			},
		}
		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.Error(err, "InitGenesis should error with invalid metrics")
	})
}

func (suite *GenesisTestSuite) TestExportGenesis() {
	ctx := suite.ctx

	suite.Run("export empty state", func() {
		exported := suite.keeper.ExportGenesis(ctx)
		suite.NotNil(exported, "ExportGenesis should not return nil")
		suite.NotNil(exported.Params, "Exported params should not be nil")
		suite.NotNil(exported.Contracts, "Exported contracts should not be nil")
		suite.NotNil(exported.Metrics, "Exported metrics should not be nil")
	})

	suite.Run("export with data", func() {
		genesis := &pb.GenesisState{
			Params: *types.DefaultParams(),
			Contracts: []pb.ContractInfo{
				{
					Address:   "aura1contract1",
					Creator:   "aura1creator",
					CodeId:    1,
					CreatedAt: time.Now(),
					Status:    pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
					Metadata: pb.ContractMetadata{
						Name:    "Test Contract",
						Version: "1.0.0",
						Tags:    []string{"test"},
					},
				},
			},
			Metrics: []pb.ContractMetrics{
				{
					ContractAddress: "aura1contract1",
					TotalExecutions: 100,
				},
			},
		}

		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		exported := suite.keeper.ExportGenesis(ctx)
		suite.NotNil(exported)
		suite.Len(exported.Contracts, 1, "Should export 1 contract")
		suite.Equal("aura1contract1", exported.Contracts[0].Address)
		suite.Len(exported.Metrics, 1, "Should export 1 metrics")
	})
}

func (suite *GenesisTestSuite) TestDefaultGenesis() {
	suite.Run("default genesis is valid", func() {
		defaultGenesis := &pb.GenesisState{
			Params:    *types.DefaultParams(),
			Contracts: []pb.ContractInfo{},
			Metrics:   []pb.ContractMetrics{},
		}
		suite.NotNil(defaultGenesis)
		suite.NotNil(defaultGenesis.Params)
		suite.Empty(defaultGenesis.Contracts)
	})

	suite.Run("default genesis can be initialized", func() {
		ctx := suite.ctx
		defaultGenesis := &pb.GenesisState{
			Params:    *types.DefaultParams(),
			Contracts: []pb.ContractInfo{},
			Metrics:   []pb.ContractMetrics{},
		}

		err := suite.keeper.InitGenesis(ctx, defaultGenesis)
		suite.NoError(err, "Should be able to initialize default genesis")
	})
}

func (suite *GenesisTestSuite) TestGenesisRoundTrip() {
	ctx := suite.ctx

	suite.Run("round trip with empty state", func() {
		genesis := &pb.GenesisState{
			Params:    *types.DefaultParams(),
			Contracts: []pb.ContractInfo{},
			Metrics:   []pb.ContractMetrics{},
		}

		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		exported := suite.keeper.ExportGenesis(ctx)
		suite.Equal(len(genesis.Contracts), len(exported.Contracts))
		suite.Equal(len(genesis.Metrics), len(exported.Metrics))
	})

	suite.Run("round trip with data", func() {
		genesis := &pb.GenesisState{
			Params: *types.DefaultParams(),
			Contracts: []pb.ContractInfo{
				{
					Address:   "aura1contract1",
					Creator:   "aura1creator1",
					CodeId:    1,
					CreatedAt: time.Now(),
					Status:    pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
					Metadata: pb.ContractMetadata{
						Name:    "Contract 1",
						Version: "1.0.0",
						Tags:    []string{"tag1"},
					},
				},
				{
					Address:   "aura1contract2",
					Creator:   "aura1creator2",
					CodeId:    2,
					CreatedAt: time.Now(),
					Status:    pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
					Metadata: pb.ContractMetadata{
						Name:    "Contract 2",
						Version: "2.0.0",
						Tags:    []string{"tag2"},
					},
				},
			},
			Metrics: []pb.ContractMetrics{
				{
					ContractAddress: "aura1contract1",
					TotalExecutions: 100,
					TotalGasUsed:    50000,
				},
				{
					ContractAddress: "aura1contract2",
					TotalExecutions: 200,
					TotalGasUsed:    100000,
				},
			},
		}

		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		exported := suite.keeper.ExportGenesis(ctx)
		suite.Equal(len(genesis.Contracts), len(exported.Contracts))
		suite.Equal(len(genesis.Metrics), len(exported.Metrics))

		err = suite.keeper.InitGenesis(ctx, exported)
		suite.NoError(err)

		exported2 := suite.keeper.ExportGenesis(ctx)
		suite.Equal(len(exported.Contracts), len(exported2.Contracts))
		suite.Equal(len(exported.Metrics), len(exported2.Metrics))
	})
}

func (suite *GenesisTestSuite) TestGenesisEdgeCases() {
	suite.Run("contract with multiple tags", func() {
		// Setup fresh state for this subtest
		suite.SetupTest()
		ctx := suite.ctx
		genesis := &pb.GenesisState{
			Params: *types.DefaultParams(),
			Contracts: []pb.ContractInfo{
				{
					Address:   "aura1contract1",
					Creator:   "aura1creator",
					CodeId:    1,
					CreatedAt: time.Now(),
					Status:    pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
					Metadata: pb.ContractMetadata{
						Name:    "Multi-Tag Contract",
						Version: "1.0.0",
						Tags:    []string{"tag1", "tag2", "tag3"},
					},
				},
			},
		}

		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		exported := suite.keeper.ExportGenesis(ctx)
		suite.Len(exported.Contracts, 1)
		suite.Len(exported.Contracts[0].Metadata.Tags, 3)
	})

	suite.Run("many contracts", func() {
		// Setup fresh state for this subtest
		suite.SetupTest()
		ctx := suite.ctx

		contracts := make([]pb.ContractInfo, 50)
		for i := 0; i < 50; i++ {
			// Generate unique addresses using zero-padded hex to ensure exactly 50 contracts
			contracts[i] = pb.ContractInfo{
				Address:   fmt.Sprintf("aura1contract%02d", i),
				Creator:   "aura1creator",
				CodeId:    uint64(i + 1),
				CreatedAt: time.Now(),
				Status:    pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
			}
		}

		genesis := &pb.GenesisState{
			Params:    *types.DefaultParams(),
			Contracts: contracts,
		}

		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		exported := suite.keeper.ExportGenesis(ctx)
		suite.Len(exported.Contracts, 50)
	})

	suite.Run("paused contract", func() {
		// Setup fresh state for this subtest
		suite.SetupTest()
		ctx := suite.ctx
		genesis := &pb.GenesisState{
			Params: *types.DefaultParams(),
			Contracts: []pb.ContractInfo{
				{
					Address:   "aura1pausedcontract",
					Creator:   "aura1creator",
					CodeId:    1,
					CreatedAt: time.Now(),
					Status:    pb.ContractStatus_CONTRACT_STATUS_PAUSED,
				},
			},
		}

		err := suite.keeper.InitGenesis(ctx, genesis)
		suite.NoError(err)

		info, found := suite.keeper.GetContractInfo(ctx, "aura1pausedcontract")
		suite.True(found)
		suite.Equal(pb.ContractStatus_CONTRACT_STATUS_PAUSED, info.Status)
	})
}

func TestValidateGenesis(t *testing.T) {
	tests := []struct {
		name      string
		genesis   *pb.GenesisState
		expectErr bool
	}{
		{
			name:      "nil genesis",
			genesis:   nil,
			expectErr: true,
		},
		{
			name: "invalid params",
			genesis: &pb.GenesisState{
				Params: pb.ContractRegistryParams{
					MaxContractsPerCreator: 20000, // exceeds limit
				},
			},
			expectErr: true,
		},
		{
			name: "valid default genesis",
			genesis: &pb.GenesisState{
				Params:    *types.DefaultParams(),
				Contracts: []pb.ContractInfo{},
				Metrics:   []pb.ContractMetrics{},
			},
			expectErr: false,
		},
		{
			name: "valid genesis with data",
			genesis: &pb.GenesisState{
				Params: *types.DefaultParams(),
				Contracts: []pb.ContractInfo{
					{
						Address:   "aura1contract1",
						Creator:   "aura1creator",
						CodeId:    1,
						CreatedAt: time.Now(),
						Status:    pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
					},
				},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.genesis == nil {
				err = fmt.Errorf("genesis is nil")
			} else {
				// Validate params
				err = types.ValidateParams(&tt.genesis.Params)
			}

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
