package keeper

import (
    sdk "github.com/cosmos/cosmos-sdk/types"

    "github.com/aequitas/aura/chain/x/governance/types"
)

// InitGenesis initializes governance state from genesis data.
func (k Keeper) InitGenesis(ctx sdk.Context, gen types.GenesisState) error {
    params := gen.Params
    if params == nil {
        params = types.DefaultParams()
    }
    k.SetParams(ctx, params)
    return nil
}

// ExportGenesis exports the current governance state as genesis data.
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
    params := k.GetParams(ctx)
    return types.GenesisState{Params: params}
}
