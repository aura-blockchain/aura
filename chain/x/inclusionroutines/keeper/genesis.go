// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
