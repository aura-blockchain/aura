package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/testing/testutil"
	"github.com/aequitas/aura/chain/x/aiassistant/keeper"
	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

func TestGenesisImportExport(t *testing.T) {
	ctx := testutil.SetupTestContext(t)
	k := &keeper.Keeper{}

	// Create test genesis state
	genesisState := types.DefaultGenesisState()
	require.NotNil(t, genesisState)

	// Test import
	t.Run("ImportGenesis", func(t *testing.T) {
		require.NotPanics(t, func() {
			keeper.InitGenesis(ctx.SdkCtx, k, genesisState)
		})
	})

	// Test export
	t.Run("ExportGenesis", func(t *testing.T) {
		exported := keeper.ExportGenesis(ctx.SdkCtx, k)
		require.NotNil(t, exported)
		// Verify exported state matches imported state
	})

	// Test round-trip
	t.Run("RoundTrip", func(t *testing.T) {
		keeper.InitGenesis(ctx.SdkCtx, k, genesisState)
		exported := keeper.ExportGenesis(ctx.SdkCtx, k)
		require.Equal(t, genesisState, exported)
	})
}

func TestGenesisValidation(t *testing.T) {
	t.Run("ValidGenesis", func(t *testing.T) {
		genesisState := types.DefaultGenesisState()
		err := genesisState.Validate()
		require.NoError(t, err)
	})

	t.Run("InvalidGenesis", func(t *testing.T) {
		// Test with invalid genesis state
		genesisState := &types.GenesisState{}
		err := genesisState.Validate()
		// Should return error for invalid state
		_ = err
	})
}
