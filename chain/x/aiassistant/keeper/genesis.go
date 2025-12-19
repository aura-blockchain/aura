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
	if err := k.SetParams(ctx, state.Params); err != nil {
		return err
	}
	for i := range state.Assistants {
		assistant := state.Assistants[i]
		k.setAssistant(ctx, &assistant)
	}
	return nil
}

func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	params := k.GetParams(ctx)
	store := storeprefix.NewStore(ctx.KVStore(k.storeKey), types.AssistantKeyPrefix)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	assistants := make([]types.Assistant, 0)
	for ; iter.Valid(); iter.Next() {
		var assistant types.Assistant
		if err := k.cdc.Unmarshal(iter.Value(), &assistant); err != nil {
			ctx.Logger().Error("failed to unmarshal assistant during genesis export, skipping",
				"error", err,
				"data_len", len(iter.Value()))
			continue
		}
		assistants = append(assistants, assistant)
	}

	return types.GenesisState{
		Assistants: assistants,
		Params:     params,
	}
}
