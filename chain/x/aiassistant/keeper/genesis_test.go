package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

func TestGenesisImportExport(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	genesis := types.DefaultGenesis()

	err := k.InitGenesis(ctx, *genesis)
	require.NoError(t, err)

	exported := k.ExportGenesis(ctx)
	require.Equal(t, genesis.Params, exported.Params)
	require.Equal(t, genesis.Assistants, exported.Assistants)
}

func TestGenesisValidation(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		genesis := types.DefaultGenesis()
		require.NoError(t, types.ValidateGenesis(*genesis))
	})

	t.Run("invalid", func(t *testing.T) {
		var genesis types.GenesisState
		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
	})
}
