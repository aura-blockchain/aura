package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/aura-bindings/types"
)

func TestDefaultGenesisState(t *testing.T) {
	genesis := types.DefaultGenesisState()

	require.NotNil(t, genesis)
	require.NotNil(t, genesis.QueryStats)
	require.NotNil(t, genesis.MessageStats)
	require.Len(t, genesis.QueryStats, 0)
	require.Len(t, genesis.MessageStats, 0)
}

func TestGenesisStateValidate(t *testing.T) {
	tests := []struct {
		name      string
		genesis   types.GenesisState
		expectErr bool
	}{
		{
			name: "valid genesis state",
			genesis: types.GenesisState{
				QueryStats: map[string]uint64{
					"query1": 10,
					"query2": 20,
				},
				MessageStats: map[string]uint64{
					"msg1": 5,
					"msg2": 15,
				},
			},
			expectErr: false,
		},
		{
			name: "empty genesis state",
			genesis: types.GenesisState{
				QueryStats:   make(map[string]uint64),
				MessageStats: make(map[string]uint64),
			},
			expectErr: false,
		},
		{
			name: "nil query stats",
			genesis: types.GenesisState{
				QueryStats:   nil,
				MessageStats: make(map[string]uint64),
			},
			expectErr: true,
		},
		{
			name: "nil message stats",
			genesis: types.GenesisState{
				QueryStats:   make(map[string]uint64),
				MessageStats: nil,
			},
			expectErr: true,
		},
		{
			name: "empty query type",
			genesis: types.GenesisState{
				QueryStats: map[string]uint64{
					"": 10,
				},
				MessageStats: make(map[string]uint64),
			},
			expectErr: true,
		},
		{
			name: "empty message type",
			genesis: types.GenesisState{
				QueryStats: make(map[string]uint64),
				MessageStats: map[string]uint64{
					"": 10,
				},
			},
			expectErr: true,
		},
		{
			name: "large stats values",
			genesis: types.GenesisState{
				QueryStats: map[string]uint64{
					"query": 18446744073709551615, // max uint64
				},
				MessageStats: map[string]uint64{
					"msg": 18446744073709551615, // max uint64
				},
			},
			expectErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.genesis.Validate()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGenesisStateValidateMultipleEntries(t *testing.T) {
	genesis := types.GenesisState{
		QueryStats: map[string]uint64{
			"query1": 1,
			"query2": 2,
			"query3": 3,
		},
		MessageStats: map[string]uint64{
			"msg1": 10,
			"msg2": 20,
			"msg3": 30,
		},
	}

	err := genesis.Validate()
	require.NoError(t, err)
}

func TestGenesisStateValidateMixedEmpty(t *testing.T) {
	tests := []struct {
		name      string
		genesis   types.GenesisState
		expectErr bool
	}{
		{
			name: "some empty query keys",
			genesis: types.GenesisState{
				QueryStats: map[string]uint64{
					"valid": 10,
					"":      5,
				},
				MessageStats: make(map[string]uint64),
			},
			expectErr: true,
		},
		{
			name: "some empty message keys",
			genesis: types.GenesisState{
				QueryStats: make(map[string]uint64),
				MessageStats: map[string]uint64{
					"valid": 10,
					"":      5,
				},
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.genesis.Validate()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDefaultGenesisStateValidates(t *testing.T) {
	genesis := types.DefaultGenesisState()
	err := genesis.Validate()
	require.NoError(t, err)
}
