// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
	"github.com/stretchr/testify/require"
)

// Type aliases for tests
type (
	GenesisState           = pb.GenesisState
	ContractRegistryParams = pb.ContractRegistryParams
)

func TestValidateGenesis(t *testing.T) {
	tests := []struct {
		name    string
		genesis *GenesisState
		wantErr bool
	}{
		{
			name:    "default genesis",
			genesis: DefaultGenesis(),
			wantErr: false,
		},
		{
			name: "valid genesis with contracts",
			genesis: &GenesisState{
				Params: *DefaultParams(),
				Contracts: []pb.ContractInfo{
					{
						Address: "cosmos1contract",
						CodeId:  1,
						Creator: "cosmos1creator",
						Status:  pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid params",
			genesis: &GenesisState{
				Params: ContractRegistryParams{
					MaxContractsPerCreator: 20000, // exceeds limit
				},
			},
			wantErr: true,
		},
		{
			name: "missing contract address",
			genesis: &GenesisState{
				Params: *DefaultParams(),
				Contracts: []pb.ContractInfo{
					{
						Address: "", // invalid
						CodeId:  1,
						Creator: "cosmos1creator",
						Status:  pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate contract address",
			genesis: &GenesisState{
				Params: *DefaultParams(),
				Contracts: []pb.ContractInfo{
					{
						Address: "cosmos1contract",
						CodeId:  1,
						Creator: "cosmos1creator",
						Status:  pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
					},
					{
						Address: "cosmos1contract", // duplicate
						CodeId:  2,
						Creator: "cosmos1creator",
						Status:  pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid code ID",
			genesis: &GenesisState{
				Params: *DefaultParams(),
				Contracts: []pb.ContractInfo{
					{
						Address: "cosmos1contract",
						CodeId:  0, // invalid
						Creator: "cosmos1creator",
						Status:  pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			genesis: &GenesisState{
				Params: *DefaultParams(),
				Contracts: []pb.ContractInfo{
					{
						Address: "cosmos1contract",
						CodeId:  1,
						Creator: "cosmos1creator",
						Status:  pb.ContractStatus_CONTRACT_STATUS_UNSPECIFIED, // invalid
					},
				},
			},
			wantErr: true,
		},
		{
			name: "metrics without matching contract",
			genesis: &GenesisState{
				Params:    *DefaultParams(),
				Contracts: []pb.ContractInfo{},
				Metrics: []pb.ContractMetrics{
					{
						ContractAddress: "cosmos1nonexistent",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGenesis(tt.genesis)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateParams(t *testing.T) {
	tests := []struct {
		name    string
		params  *ContractRegistryParams
		wantErr bool
	}{
		{
			name:    "default params",
			params:  DefaultParams(),
			wantErr: false,
		},
		{
			name: "valid custom params",
			params: &ContractRegistryParams{
				OpenRegistration:        true,
				MaxContractsPerCreator:  50,
				RequireMetadata:         true,
				RequireSecurityPolicy:   true,
				RequireComplianceConfig: false,
				DefaultRateLimit:        100,
				DefaultMaxGas:           1000000,
			},
			wantErr: false,
		},
		{
			name: "max contracts exceeds limit",
			params: &ContractRegistryParams{
				MaxContractsPerCreator: 20000,
			},
			wantErr: true,
		},
		{
			name: "default rate limit exceeds limit",
			params: &ContractRegistryParams{
				MaxContractsPerCreator: 100,
				DefaultRateLimit:       20000,
			},
			wantErr: true,
		},
		{
			name: "default max gas exceeds limit",
			params: &ContractRegistryParams{
				MaxContractsPerCreator: 100,
				DefaultRateLimit:       100,
				DefaultMaxGas:          100000000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateParams(tt.params)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDefaultParams(t *testing.T) {
	params := *DefaultParams()

	require.True(t, params.OpenRegistration)
	require.Equal(t, uint64(100), params.MaxContractsPerCreator)
	require.False(t, params.RequireMetadata)
	require.False(t, params.RequireSecurityPolicy)
	require.False(t, params.RequireComplianceConfig)
	require.Equal(t, uint64(365), params.AuditWarningDays)
	require.Equal(t, uint64(1000), params.DefaultRateLimit)
	require.Equal(t, uint64(10000000), params.DefaultMaxGas)

	// Should pass validation
	require.NoError(t, ValidateParams(&params))
}

func TestNewGenesisState(t *testing.T) {
	params := *DefaultParams()
	contracts := []pb.ContractInfo{
		{
			Address: "cosmos1contract",
			CodeId:  1,
			Creator: "cosmos1creator",
			Status:  pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
		},
	}
	metrics := []pb.ContractMetrics{
		{
			ContractAddress: "cosmos1contract",
			TotalExecutions: 10,
		},
	}

	genesis := NewGenesisState(params, contracts, metrics)

	require.NotNil(t, genesis)
	require.Equal(t, params, genesis.Params)
	require.Len(t, genesis.Contracts, 1)
	require.Len(t, genesis.Metrics, 1)
	require.Equal(t, "cosmos1contract", genesis.Contracts[0].Address)
}
