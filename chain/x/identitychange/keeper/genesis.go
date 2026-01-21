// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/identitychange/types"
)

// InitGenesis initializes the module state from genesis data
func (k *Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) error {
	// Validate genesis before importing
	if err := types.ValidateGenesisState(&data); err != nil {
		return fmt.Errorf("invalid genesis state: %w", err)
	}

	// Set params
	if data.Params != nil {
		if err := k.paramsStore.SetParams(*data.Params); err != nil {
			return fmt.Errorf("failed to set params: %w", err)
		}
	}

	// Import identity records
	for _, record := range data.Records {
		if record == nil {
			continue
		}
		if err := k.SetIdentityRecord(ctx, *record); err != nil {
			return fmt.Errorf("failed to set identity record %s: %w", record.Did, err)
		}
	}

	// Import identity change requests
	for _, request := range data.Requests {
		if request == nil {
			continue
		}
		if err := k.SetRequest(ctx, *request); err != nil {
			return fmt.Errorf("failed to set request %s: %w", request.RequestId, err)
		}
	}

	// Import history entries
	for _, historyEntry := range data.History {
		if historyEntry == nil {
			continue
		}
		if err := k.AddHistory(ctx, *historyEntry); err != nil {
			return fmt.Errorf("failed to add history entry for DID %s: %w", historyEntry.TargetDid, err)
		}
	}

	// Set suspended flag
	if err := k.SetSuspended(ctx, data.Suspended); err != nil {
		return fmt.Errorf("failed to set suspended flag: %w", err)
	}

	k.logger.Info("identity change genesis initialized",
		"records_count", len(data.Records),
		"requests_count", len(data.Requests),
		"history_count", len(data.History),
		"suspended", data.Suspended)

	return nil
}

// ExportGenesis exports the current module state to genesis
func (k *Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	// Export params
	params, _ := k.GetParams(ctx)

	// Export identity records
	records := []*types.IdentityRecord{}
	for _, record := range k.GetAllRecords(ctx) {
		recordCopy := record
		records = append(records, &recordCopy)
	}

	// Export identity change requests
	requests := []*types.IdentityChangeRequest{}
	for _, request := range k.GetAllRequests(ctx) {
		requestCopy := request
		requests = append(requests, &requestCopy)
	}

	// Export all history entries
	history := []*types.IdentityChangeHistory{}
	// We need to iterate through all DIDs to get their history
	allRecords := k.GetAllRecords(ctx)
	seenHistory := make(map[string]bool)

	for _, record := range allRecords {
		historyEntries := k.ListHistory(ctx, record.Did)
		for _, entry := range historyEntries {
			// Use a unique key to avoid duplicates
			key := fmt.Sprintf("%s-%s-%d", entry.TargetDid, entry.RequestId, entry.ChangedHeight)
			if !seenHistory[key] {
				entryCopy := entry
				history = append(history, &entryCopy)
				seenHistory[key] = true
			}
		}
	}

	k.logger.Info("identity change genesis exported",
		"records_count", len(records),
		"requests_count", len(requests),
		"history_count", len(history),
		"suspended", k.IsSuspended(ctx))

	return types.GenesisState{
		Params:    &params,
		Records:   records,
		Requests:  requests,
		History:   history,
		Suspended: k.IsSuspended(ctx),
	}
}

// ValidateGenesis validates the genesis state
func (k *Keeper) ValidateGenesis(ctx sdk.Context, data types.GenesisState) error {
	return types.ValidateGenesisState(&data)
}

// InitGenesisDefaultParams initializes genesis with default parameters only
func (k *Keeper) InitGenesisDefaultParams(ctx sdk.Context) error {
	defaultParams := types.DefaultParams()
	if err := k.paramsStore.SetParams(defaultParams); err != nil {
		return fmt.Errorf("failed to set default params: %w", err)
	}

	// Set suspended to false by default
	if err := k.SetSuspended(ctx, false); err != nil {
		return fmt.Errorf("failed to set suspended flag: %w", err)
	}

	k.logger.Info("identity change genesis initialized with defaults")
	return nil
}

// GetGenesisState returns the current genesis state
func (k *Keeper) GetGenesisState(ctx sdk.Context) types.GenesisState {
	return k.ExportGenesis(ctx)
}

// ResetGenesis resets the module to genesis state (for testing)
func (k *Keeper) ResetGenesis(ctx sdk.Context, data types.GenesisState) error {
	// Clear all existing state
	if err := k.ClearAllState(ctx); err != nil {
		return fmt.Errorf("failed to clear state: %w", err)
	}

	// Initialize with new genesis
	return k.InitGenesis(ctx, data)
}

// ClearAllState clears all module state (for testing)
func (k *Keeper) ClearAllState(ctx sdk.Context) error {
	store := k.storeService.OpenKVStore(ctx)

	// Clear all records
	allRecords := k.GetAllRecords(ctx)
	for _, record := range allRecords {
		key := []byte(types.RecordStoreKey(record.Did))
		if err := store.Delete(key); err != nil {
			return fmt.Errorf("failed to delete record %s: %w", record.Did, err)
		}
	}

	// Clear all requests
	allRequests := k.GetAllRequests(ctx)
	for _, request := range allRequests {
		key := []byte(types.RequestStoreKey(request.RequestId))
		if err := store.Delete(key); err != nil {
			return fmt.Errorf("failed to delete request %s: %w", request.RequestId, err)
		}
	}

	// Clear all history
	for _, record := range allRecords {
		historyEntries := k.ListHistory(ctx, record.Did)
		for _, entry := range historyEntries {
			key := []byte(fmt.Sprintf("%s%s/%d",
				types.HistoryStoreKeyPrefix,
				entry.TargetDid,
				entry.ChangedHeight))
			if err := store.Delete(key); err != nil {
				return fmt.Errorf("failed to delete history entry: %w", err)
			}
		}
	}

	// Clear suspended flag
	if err := store.Delete([]byte(types.SuspendedKey)); err != nil {
		return fmt.Errorf("failed to delete suspended flag: %w", err)
	}

	k.logger.Info("identity change state cleared")
	return nil
}

// ImportIdentityRecords imports identity records in batch
func (k *Keeper) ImportIdentityRecords(ctx sdk.Context, records []*types.IdentityRecord) error {
	for _, record := range records {
		if record == nil {
			continue
		}
		if err := k.SetIdentityRecord(ctx, *record); err != nil {
			return fmt.Errorf("failed to import record %s: %w", record.Did, err)
		}
	}
	k.logger.Info("imported identity records", "count", len(records))
	return nil
}

// ImportIdentityChangeRequests imports requests in batch
func (k *Keeper) ImportIdentityChangeRequests(ctx sdk.Context, requests []*types.IdentityChangeRequest) error {
	for _, request := range requests {
		if request == nil {
			continue
		}
		if err := k.SetRequest(ctx, *request); err != nil {
			return fmt.Errorf("failed to import request %s: %w", request.RequestId, err)
		}
	}
	k.logger.Info("imported identity change requests", "count", len(requests))
	return nil
}

// ImportIdentityHistory imports history entries in batch
func (k *Keeper) ImportIdentityHistory(ctx sdk.Context, history []*types.IdentityChangeHistory) error {
	for _, entry := range history {
		if entry == nil {
			continue
		}
		if err := k.AddHistory(ctx, *entry); err != nil {
			return fmt.Errorf("failed to import history entry for DID %s: %w", entry.TargetDid, err)
		}
	}
	k.logger.Info("imported identity history", "count", len(history))
	return nil
}

// ExportIdentityRecords exports all identity records
func (k *Keeper) ExportIdentityRecords(ctx sdk.Context) []*types.IdentityRecord {
	records := []*types.IdentityRecord{}
	for _, record := range k.GetAllRecords(ctx) {
		recordCopy := record
		records = append(records, &recordCopy)
	}
	return records
}

// ExportIdentityChangeRequests exports all requests
func (k *Keeper) ExportIdentityChangeRequests(ctx sdk.Context) []*types.IdentityChangeRequest {
	requests := []*types.IdentityChangeRequest{}
	for _, request := range k.GetAllRequests(ctx) {
		requestCopy := request
		requests = append(requests, &requestCopy)
	}
	return requests
}

// ExportIdentityHistory exports all history entries
func (k *Keeper) ExportIdentityHistory(ctx sdk.Context) []*types.IdentityChangeHistory {
	history := []*types.IdentityChangeHistory{}
	allRecords := k.GetAllRecords(ctx)
	seenHistory := make(map[string]bool)

	for _, record := range allRecords {
		historyEntries := k.ListHistory(ctx, record.Did)
		for _, entry := range historyEntries {
			key := fmt.Sprintf("%s-%s-%d", entry.TargetDid, entry.RequestId, entry.ChangedHeight)
			if !seenHistory[key] {
				entryCopy := entry
				history = append(history, &entryCopy)
				seenHistory[key] = true
			}
		}
	}
	return history
}

// GenesisStateSize returns the size of the genesis state
func (k *Keeper) GenesisStateSize(ctx sdk.Context) (recordsCount, requestsCount, historyCount int) {
	recordsCount = len(k.GetAllRecords(ctx))
	requestsCount = len(k.GetAllRequests(ctx))

	// Count unique history entries
	allRecords := k.GetAllRecords(ctx)
	seenHistory := make(map[string]bool)
	for _, record := range allRecords {
		historyEntries := k.ListHistory(ctx, record.Did)
		for _, entry := range historyEntries {
			key := fmt.Sprintf("%s-%s-%d", entry.TargetDid, entry.RequestId, entry.ChangedHeight)
			seenHistory[key] = true
		}
	}
	historyCount = len(seenHistory)

	return
}
