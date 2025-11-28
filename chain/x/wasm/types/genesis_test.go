package types_test

import (
	"testing"

	"github.com/aequitas/aura/chain/x/wasm/types"
	"github.com/stretchr/testify/require"
)

func TestDefaultGenesisState(t *testing.T) {
	genesis := types.DefaultGenesisState()
	require.NotNil(t, genesis)
	require.Equal(t, types.DefaultParams(), genesis.Params)
	require.Empty(t, genesis.AuthorizedUploaders)
	require.Empty(t, genesis.PausedContracts)
	require.Equal(t, types.SecurityStats{}, genesis.SecurityStats)
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
					MaxContractSize:         1024 * 1024,
					MaxInstantiateGas:       3_000_000,
					MaxExecuteGas:           2_000_000,
					MaxQueryGas:             200_000,
					RequireAuthorization:    false,
					EnableMigration:         true,
					MaxContractSizePerBlock: 10 * 1024 * 1024,
				},
				AuthorizedUploaders: []string{"aura1abc123", "aura1def456"},
				PausedContracts:     []string{"aura14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr"},
				SecurityStats: types.SecurityStats{
					TotalContractsUploaded:    10,
					TotalContractsInstantiated: 8,
					TotalExecutions:           100,
					TotalPausedContracts:      1,
					ReentrancyAttemptsBlocked: 2,
				},
			},
			expectErr: false,
		},
		{
			name: "invalid params",
			genesis: types.GenesisState{
				Params: types.Params{
					MaxContractSize:         0, // Invalid
					MaxInstantiateGas:       2_000_000,
					MaxExecuteGas:           1_000_000,
					MaxQueryGas:             100_000,
					RequireAuthorization:    true,
					EnableMigration:         false,
					MaxContractSizePerBlock: 5 * 1024 * 1024,
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
				Params:              types.DefaultParams(),
				AuthorizedUploaders: []string{""},
				PausedContracts:     []string{},
				SecurityStats:       types.SecurityStats{},
			},
			expectErr: true,
			errMsg:    "empty authorized uploader address",
		},
		{
			name: "empty paused contract address",
			genesis: types.GenesisState{
				Params:              types.DefaultParams(),
				AuthorizedUploaders: []string{},
				PausedContracts:     []string{""},
				SecurityStats:       types.SecurityStats{},
			},
			expectErr: true,
			errMsg:    "empty paused contract address",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.genesis.Validate()
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
		MaxContractSize:         1024 * 1024,
		MaxInstantiateGas:       3_000_000,
		MaxExecuteGas:           2_000_000,
		MaxQueryGas:             200_000,
		RequireAuthorization:    false,
		EnableMigration:         true,
		MaxContractSizePerBlock: 10 * 1024 * 1024,
	}
	uploaders := []string{"aura1abc123"}
	paused := []string{"aura14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr"}
	stats := types.SecurityStats{
		TotalContractsUploaded: 5,
		TotalExecutions:        50,
	}

	genesis := types.NewGenesisState(params, uploaders, paused, stats)
	require.NotNil(t, genesis)
	require.Equal(t, params, genesis.Params)
	require.Equal(t, uploaders, genesis.AuthorizedUploaders)
	require.Equal(t, paused, genesis.PausedContracts)
	require.Equal(t, stats, genesis.SecurityStats)
}

func TestDefaultParams(t *testing.T) {
	params := types.DefaultParams()
	require.Equal(t, uint64(600*1024), params.MaxContractSize)
	require.Equal(t, uint64(2_000_000), params.MaxInstantiateGas)
	require.Equal(t, uint64(1_000_000), params.MaxExecuteGas)
	require.Equal(t, uint64(100_000), params.MaxQueryGas)
	require.True(t, params.RequireAuthorization)
	require.False(t, params.EnableMigration)
	require.Equal(t, uint64(5*1024*1024), params.MaxContractSizePerBlock)
}
