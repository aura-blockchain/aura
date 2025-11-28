package keeper

import (
	storeprefix "cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, state types.GenesisState) error {
	if err := types.ValidateGenesis(state); err != nil {
		return err
	}
	if err := k.SetParams(ctx, *state.Params); err != nil {
		return err
	}
	for _, assistant := range state.Assistants {
		k.setAssistant(ctx, assistant)
	}
	return nil
}

func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	params := k.GetParams(ctx)
	store := storeprefix.NewStore(ctx.KVStore(k.storeKey), types.AssistantKeyPrefix)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	assistants := make([]*types.Assistant, 0)
	for ; iter.Valid(); iter.Next() {
		var assistant types.Assistant
		k.cdc.MustUnmarshal(iter.Value(), &assistant)
		assistants = append(assistants, &assistant)
	}

	return types.GenesisState{
		Assistants: assistants,
		Params:     &params,
	}
}
