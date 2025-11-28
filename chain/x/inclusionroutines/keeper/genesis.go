package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
)

// Genesis management is handled in keeper.go via InitGenesis and ExportGenesis methods
// This file provides additional genesis-related utility functions

// ValidateGenesis validates genesis state before initialization
func (k *Keeper) ValidateGenesis(genesis types.GenesisState) error {
	// Validate params if present
	// Note: Params validation would be done in types package if needed
	_ = genesis

	return nil
}

// GetGenesisState returns the current genesis state
func (k *Keeper) GetGenesisState(ctx sdk.Context) types.GenesisState {
	return k.ExportGenesis(ctx)
}
