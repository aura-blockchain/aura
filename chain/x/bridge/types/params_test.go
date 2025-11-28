package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/bridge/types"
)

func TestDefaultParams(t *testing.T) {
	params := types.DefaultParams()

	require.NotNil(t, params)
	require.True(t, params.BridgeEnabled)
}

func TestParamKeyTable(t *testing.T) {
	table := types.ParamKeyTable()

	require.NotNil(t, table)
	// KeyTable doesn't have HasKeyTable method, just verify it's not nil
}

func TestParamSetPairs(t *testing.T) {
	params := types.DefaultParams()
	pairs := params.ParamSetPairs()

	require.NotNil(t, pairs)
	require.NotEmpty(t, pairs)
}

func TestValidateBool_Valid(t *testing.T) {
	// Access the validation function through ParamSetPairs
	params := types.DefaultParams()
	pairs := params.ParamSetPairs()

	// Find the BridgeEnabled pair
	for _, pair := range pairs {
		if string(pair.Key) == string(types.KeyBridgeEnabled) {
			// Test valid bool
			err := pair.ValidatorFn(true)
			require.NoError(t, err)

			err = pair.ValidatorFn(false)
			require.NoError(t, err)
		}
	}
}

func TestValidateBool_Invalid(t *testing.T) {
	params := types.DefaultParams()
	pairs := params.ParamSetPairs()

	// Find the BridgeEnabled pair
	for _, pair := range pairs {
		if string(pair.Key) == string(types.KeyBridgeEnabled) {
			// Test invalid types
			err := pair.ValidatorFn("not a bool")
			require.Error(t, err)

			err = pair.ValidatorFn(123)
			require.Error(t, err)

			err = pair.ValidatorFn(nil)
			require.Error(t, err)
		}
	}
}

func TestDefaultGenesis(t *testing.T) {
	genesis := types.DefaultGenesis()

	require.NotNil(t, genesis)
	require.NotNil(t, genesis.Params)
	require.True(t, genesis.Params.Enabled)
}

func TestDefaultTimelockDuration(t *testing.T) {
	require.Equal(t, 24*time.Hour, types.DefaultTimelockDuration)
}

func TestDefaultFraudProofWindow(t *testing.T) {
	require.Equal(t, 7*24*time.Hour, types.DefaultFraudProofWindow)
}

func TestParamsStructure(t *testing.T) {
	params := types.Params{
		BridgeEnabled: true,
	}

	require.True(t, params.BridgeEnabled)

	params.BridgeEnabled = false
	require.False(t, params.BridgeEnabled)
}

func TestKeyConstants(t *testing.T) {
	require.NotNil(t, types.KeyBridgeEnabled)
	require.NotEmpty(t, types.KeyBridgeEnabled)
	require.Equal(t, "BridgeEnabled", string(types.KeyBridgeEnabled))
}
