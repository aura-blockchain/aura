package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// InitGenesis initializes the dex module state from genesis
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) error {
	// Set parameters
	if err := k.SetParams(ctx, data.Params); err != nil {
		return err
	}

	return nil
}

// ExportGenesis exports the dex module state to genesis
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	// Get parameters
	params := k.GetParams(ctx)

	return types.GenesisState{
		Params: params,
	}
}
