// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"testing"

	"github.com/aequitas/aura/chain/x/wasm/types"
	"github.com/stretchr/testify/require"
)

func TestDefaultGenesisState(t *testing.T) {
	genesis := types.DefaultGenesisState()
	require.NotNil(t, genesis)
	require.Equal(t, *types.DefaultParams(), genesis.Params)
	require.Empty(t, genesis.AuthorizedUploaders)
	require.Empty(t, genesis.PausedContracts)
}

func TestGenesisStateValidation(t *testing.T) {
	testCases := []struct {
		name      string
		genesis   types.GenesisState
		expectErr bool
		errMsg    string
	}{
		{
			name:      "default genesis is valid",
			genesis:   *types.DefaultGenesisState(),
			expectErr: false,
		},
		{
			name: "valid custom genesis",
			genesis: types.GenesisState{
				Params: types.Params{
					CodeUploadAccess: types.AccessConfig{
						Permission: types.AccessTypeEverybody,
					},
					InstantiateDefaultPermission: types.AccessTypeEverybody,
					MaxWasmCodeSize:              1024 * 1024,
					MaxGasWasmExecution:          10_000_000,
					SecurityAnalysisEnabled:      true,
					RequireAdminForMigrate:       true,
				},
				AuthorizedUploaders: []string{"aura1abc123", "aura1def456"},
				PausedContracts:     []string{"aura14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr"},
				SecurityStats:       types.SecurityStats{},
			},
			expectErr: false,
		},
		{
			name: "invalid params",
			genesis: types.GenesisState{
				Params: types.Params{
					CodeUploadAccess: types.AccessConfig{
						Permission: types.AccessTypeEverybody,
					},
					InstantiateDefaultPermission: types.AccessTypeEverybody,
					MaxWasmCodeSize:              0, // Invalid - too small
					MaxGasWasmExecution:          1_000_000,
					SecurityAnalysisEnabled:      true,
					RequireAdminForMigrate:       false,
				},
				AuthorizedUploaders: []string{},
				PausedContracts:     []string{},
				SecurityStats:       types.SecurityStats{},
			},
			expectErr: true,
			errMsg:    "invalid params",
		},
		{
			name: "empty authorized uploader address",
			genesis: types.GenesisState{
				Params:              *types.DefaultParams(),
				AuthorizedUploaders: []string{""},
				PausedContracts:     []string{},
				SecurityStats:       *types.DefaultSecurityStats(),
			},
			expectErr: true,
			errMsg:    "empty authorized uploader address",
		},
		{
			name: "empty paused contract address",
			genesis: types.GenesisState{
				Params:              *types.DefaultParams(),
				AuthorizedUploaders: []string{},
				PausedContracts:     []string{""},
				SecurityStats:       *types.DefaultSecurityStats(),
			},
			expectErr: true,
			errMsg:    "empty paused contract address",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateGenesis(&tc.genesis)
			if tc.expectErr {
				require.Error(t, err)
				if tc.errMsg != "" {
					require.Contains(t, err.Error(), tc.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewGenesisState(t *testing.T) {
	params := types.Params{
		CodeUploadAccess: types.AccessConfig{
			Permission: types.AccessTypeEverybody,
		},
		InstantiateDefaultPermission: types.AccessTypeEverybody,
		MaxWasmCodeSize:              1024 * 1024,
		MaxGasWasmExecution:          10_000_000,
		SecurityAnalysisEnabled:      true,
		RequireAdminForMigrate:       true,
	}
	codes := []types.Code{}
	contracts := []types.Contract{}
	sequences := []types.Sequence{}
	uploaders := []string{"aura1abc123"}
	paused := []string{"aura14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr"}
	stats := types.SecurityStats{}

	genesis := types.NewGenesisState(params, codes, contracts, sequences, uploaders, paused, stats)
	require.NotNil(t, genesis)
	require.Equal(t, params, genesis.Params)
	require.Equal(t, uploaders, genesis.AuthorizedUploaders)
	require.Equal(t, paused, genesis.PausedContracts)
	require.Equal(t, stats, genesis.SecurityStats)
}

func TestDefaultParams(t *testing.T) {
	params := types.DefaultParams()
	require.NotNil(t, params)
	require.Equal(t, types.AccessTypeEverybody, params.CodeUploadAccess.Permission)
	require.Equal(t, types.AccessTypeEverybody, params.InstantiateDefaultPermission)
	require.Equal(t, uint64(600*1024), params.MaxWasmCodeSize)
	require.Equal(t, uint64(10_000_000), params.MaxGasWasmExecution)
	require.True(t, params.SecurityAnalysisEnabled)
	require.True(t, params.RequireAdminForMigrate)
}
