package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/bridge/types"
)

// InitGenesis initializes the bridge module state from genesis
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) error {
	// Set parameters
	if err := k.SetParams(ctx, data.Params); err != nil {
		return err
	}

	return nil
}

// ExportGenesis exports the bridge module state to genesis
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	// Get parameters
	params := k.GetParams(ctx)

	return types.GenesisState{
		Params: params,
	}
}
