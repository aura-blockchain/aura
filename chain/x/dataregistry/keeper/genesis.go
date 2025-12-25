// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/dataregistry/types"
)

// InitGenesis initializes the module state from genesis
func (k *Keeper) InitGenesis(ctx sdk.Context, genesis types.GenesisState) error {
	// Validate genesis state
	if err := types.ValidateGenesis(&genesis); err != nil {
		return fmt.Errorf("invalid genesis state: %w", err)
	}

	// Set params
	if genesis.Params != nil {
		if err := k.SetParams(*genesis.Params); err != nil {
			return fmt.Errorf("failed to set params: %w", err)
		}
	}

	// Validate and track data item IDs to prevent duplicates
	seenDataIDs := make(map[string]bool)
	for _, item := range genesis.DataItems {
		if item == nil {
			continue
		}

		// CRITICAL SECURITY: Detect duplicate data item IDs during import
		// Duplicate IDs would cause silent overwrites of existing items
		if seenDataIDs[item.DataId] {
			return fmt.Errorf("duplicate data item ID in genesis: %s", item.DataId)
		}
		seenDataIDs[item.DataId] = true

		// Import data item
		if err := k.SetDataItem(ctx, *item); err != nil {
			return fmt.Errorf("failed to set data item %s: %w", item.DataId, err)
		}
	}

	// Set next data ID counter
	// CRITICAL FIX: Use the genesis value directly (already validated and correct)
	if err := k.SetNextDataID(ctx, genesis.NextDataId); err != nil {
		return fmt.Errorf("failed to set next data ID: %w", err)
	}

	return nil
}

// ExportGenesis exports the module state to genesis
func (k *Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	params, _ := k.GetParams(ctx)

	// Export all data items
	items := []*types.DataItem{}
	store := k.storeService.OpenKVStore(ctx)

	prefix := types.DataItemKeyPrefix
	iterator, err := store.Iterator(prefix, types.PrefixEndBytes(prefix))
	if err == nil {
		defer iterator.Close()
		for ; iterator.Valid(); iterator.Next() {
			var item types.DataItem
			if err := k.cdc.Unmarshal(iterator.Value(), &item); err == nil {
				itemCopy := item
				items = append(items, &itemCopy)
			}
		}
	}

	// Get next data ID counter
	nextDataID := k.GetNextDataID(ctx)

	return types.GenesisState{
		Params:     &params,
		DataItems:  items,
		NextDataId: nextDataID,
	}
}
