package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
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

func TestGenesisRoundTrip(t *testing.T) {
	k, ctx, _ := setupKeeper(t)

	// Create a genesis state with data using correct Assistant struct
	genesis := types.DefaultGenesis()
	genesis.Assistants = []types.Assistant{
		{
			AssistantAddress: "aura1assistant1",
			OwnerAddress:     "aura1owner1",
			Stake: types.Balance{
				Denom:  types.DefaultStakeDenom,
				Amount: sdkmath.NewInt(10000000),
			},
			SponsorshipBalance: types.Balance{
				Denom:  types.DefaultStakeDenom,
				Amount: sdkmath.ZeroInt(),
			},
			Locales:           []string{"en-US"},
			ModelHash:         "model-hash-1",
			ApiKeyFingerprint: "fingerprint-1",
			LastHeartbeat:     time.Now(),
			Status:            types.AssistantStatus_ACTIVE,
		},
		{
			AssistantAddress: "aura1assistant2",
			OwnerAddress:     "aura1owner2",
			Stake: types.Balance{
				Denom:  types.DefaultStakeDenom,
				Amount: sdkmath.NewInt(5000000),
			},
			SponsorshipBalance: types.Balance{
				Denom:  types.DefaultStakeDenom,
				Amount: sdkmath.ZeroInt(),
			},
			Locales:           []string{"es-ES"},
			ModelHash:         "model-hash-2",
			ApiKeyFingerprint: "fingerprint-2",
			LastHeartbeat:     time.Now(),
			Status:            types.AssistantStatus_JAILED,
		},
	}

	// Import genesis
	err := k.InitGenesis(ctx, *genesis)
	require.NoError(t, err)

	// Export genesis (first export)
	exported1 := k.ExportGenesis(ctx)
	require.NotNil(t, exported1)
	require.Equal(t, len(genesis.Assistants), len(exported1.Assistants))
	require.Equal(t, genesis.Params, exported1.Params)

	// Create fresh keeper for re-import
	k2, ctx2, _ := setupKeeper(t)

	// Re-import the exported genesis
	err = k2.InitGenesis(ctx2, exported1)
	require.NoError(t, err)

	// Export again (second export)
	exported2 := k2.ExportGenesis(ctx2)
	require.NotNil(t, exported2)

	// The two exports should be identical
	require.Equal(t, exported1.Params, exported2.Params)
	require.Equal(t, len(exported1.Assistants), len(exported2.Assistants))

	// Verify individual assistant records match
	for i := range exported1.Assistants {
		require.Equal(t, exported1.Assistants[i].AssistantAddress, exported2.Assistants[i].AssistantAddress)
		require.Equal(t, exported1.Assistants[i].OwnerAddress, exported2.Assistants[i].OwnerAddress)
		require.Equal(t, exported1.Assistants[i].Status, exported2.Assistants[i].Status)
	}
}
