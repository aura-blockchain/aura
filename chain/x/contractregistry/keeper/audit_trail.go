// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	storetypes "cosmossdk.io/store/types"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/contractregistry/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// RecordAuditEvent records an audit event for a contract
// This is the primary method for recording audit events with full validation and indexing
func (k Keeper) RecordAuditEvent(ctx sdk.Context, contractID, eventType, actor string, metadata map[string]string) error {
	// Input validation
	if contractID == "" {
		return fmt.Errorf("contract ID cannot be empty")
	}
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}
	if actor == "" {
		return fmt.Errorf("actor cannot be empty")
	}

	store := ctx.KVStore(k.storeKey)

	// Get next sequence number for this contract
	seqKey := types.AuditSequenceKey(contractID)
	seqBz := store.Get(seqKey)
	var seq uint64
	if seqBz != nil {
		seq = binary.BigEndian.Uint64(seqBz)
	}
	seq++

	// Get next global sequence number
	globalSeqBz := store.Get(types.AuditGlobalSequenceKey)
	var globalSeq uint64
	if globalSeqBz != nil {
		globalSeq = binary.BigEndian.Uint64(globalSeqBz)
	}
	globalSeq++

	// Store new sequence numbers
	newSeqBz := make([]byte, 8)
	binary.BigEndian.PutUint64(newSeqBz, seq)
	store.Set(seqKey, newSeqBz)

	globalSeqBz = make([]byte, 8)
	binary.BigEndian.PutUint64(globalSeqBz, globalSeq)
	store.Set(types.AuditGlobalSequenceKey, globalSeqBz)

	// Create a copy of metadata to ensure immutability
	metadataCopy := make(map[string]string)
	for k, v := range metadata {
		metadataCopy[k] = v
	}

	// Create audit entry
	entry := &types.AuditEntry{
		Id:              seq,
		ContractAddress: contractID,
		Timestamp:       timestamppb.Now(),
		Action:          eventType, // Use eventType for both Action and EventType for compatibility
		EventType:       eventType,
		Actor:           actor,
		Details:         fmt.Sprintf("Event: %s by %s", eventType, actor),
		Metadata:        metadataCopy,
		Success:         true, // Default to true; caller can use AddAuditEntry for custom Success values
	}

	// Store the audit entry
	key := types.AuditEntryKey(contractID, seq)
	bz, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit entry: %w", err)
	}
	store.Set(key, bz)

	// Create secondary indexes
	// Index by actor
	actorIndexKey := types.AuditActorIndexKey(actor, globalSeq)
	store.Set(actorIndexKey, []byte(contractID))

	// Index by event type
	typeIndexKey := types.AuditTypeIndexKey(eventType, globalSeq)
	store.Set(typeIndexKey, []byte(contractID))

	// Emit audit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"contract_audit_event",
			sdk.NewAttribute("contract_address", contractID),
			sdk.NewAttribute("event_type", eventType),
			sdk.NewAttribute("actor", actor),
			sdk.NewAttribute("sequence", fmt.Sprintf("%d", seq)),
		),
	)

	return nil
}

// GetAuditTrail retrieves audit entries for a contract with pagination
func (k Keeper) GetAuditTrail(ctx sdk.Context, contractID string, limit uint64) []*types.AuditEntry {
	if contractID == "" {
		return []*types.AuditEntry{}
	}

	store := ctx.KVStore(k.storeKey)
	prefix := types.AuditEntriesPrefix(contractID)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	entries := []*types.AuditEntry{}
	count := uint64(0)

	for ; iterator.Valid(); iterator.Next() {
		if limit > 0 && count >= limit {
			break
		}

		var entry types.AuditEntry
		if err := json.Unmarshal(iterator.Value(), &entry); err != nil {
			continue
		}
		entries = append(entries, &entry)
		count++
	}

	return entries
}

// GetAuditEventsByActor retrieves audit events by actor across all contracts
func (k Keeper) GetAuditEventsByActor(ctx sdk.Context, actor string, limit uint64) []*types.AuditEntry {
	if actor == "" {
		return []*types.AuditEntry{}
	}

	store := ctx.KVStore(k.storeKey)
	prefix := types.AuditActorPrefix(actor)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	entries := []*types.AuditEntry{}
	count := uint64(0)

	// Map to track unique entries by global sequence
	seen := make(map[string]bool)

	for ; iterator.Valid(); iterator.Next() {
		if limit > 0 && count >= limit {
			break
		}

		// Get contract ID from index value
		contractID := string(iterator.Value())

		// Get all entries for this contract and filter by actor
		contractEntries := k.GetAuditTrail(ctx, contractID, 0)
		for _, entry := range contractEntries {
			if entry.Actor == actor {
				// Use a unique key to avoid duplicates
				uniqueKey := fmt.Sprintf("%s-%d", entry.ContractAddress, entry.Id)
				if !seen[uniqueKey] {
					seen[uniqueKey] = true
					entries = append(entries, entry)
					count++
					if limit > 0 && count >= limit {
						break
					}
				}
			}
		}

		if limit > 0 && count >= limit {
			break
		}
	}

	return entries
}

// GetAuditEventsByType retrieves audit events by event type across all contracts
func (k Keeper) GetAuditEventsByType(ctx sdk.Context, eventType string, limit uint64) []*types.AuditEntry {
	if eventType == "" {
		return []*types.AuditEntry{}
	}

	store := ctx.KVStore(k.storeKey)
	prefix := types.AuditTypePrefix(eventType)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	entries := []*types.AuditEntry{}
	count := uint64(0)

	// Map to track unique entries
	seen := make(map[string]bool)

	for ; iterator.Valid(); iterator.Next() {
		if limit > 0 && count >= limit {
			break
		}

		// Get contract ID from index value
		contractID := string(iterator.Value())

		// Get all entries for this contract and filter by event type
		contractEntries := k.GetAuditTrail(ctx, contractID, 0)
		for _, entry := range contractEntries {
			if entry.EventType == eventType {
				uniqueKey := fmt.Sprintf("%s-%d", entry.ContractAddress, entry.Id)
				if !seen[uniqueKey] {
					seen[uniqueKey] = true
					entries = append(entries, entry)
					count++
					if limit > 0 && count >= limit {
						break
					}
				}
			}
		}

		if limit > 0 && count >= limit {
			break
		}
	}

	return entries
}

// SearchAuditEvents searches for audit events matching the given criteria
// Supported criteria keys: "contract_id", "actor", "event_type"
func (k Keeper) SearchAuditEvents(ctx sdk.Context, criteria map[string]string, limit uint64) []*types.AuditEntry {
	if len(criteria) == 0 {
		return []*types.AuditEntry{}
	}

	var entries []*types.AuditEntry

	// Check for specific search criteria
	if contractID, ok := criteria["contract_id"]; ok && contractID != "" {
		// Search by contract ID
		entries = k.GetAuditTrail(ctx, contractID, limit)
	} else if actor, ok := criteria["actor"]; ok && actor != "" {
		// Search by actor
		entries = k.GetAuditEventsByActor(ctx, actor, limit)
	} else if eventType, ok := criteria["event_type"]; ok && eventType != "" {
		// Search by event type
		entries = k.GetAuditEventsByType(ctx, eventType, limit)
	} else {
		// No recognized criteria, return empty
		return []*types.AuditEntry{}
	}

	// Apply additional filtering if multiple criteria are specified
	if len(criteria) > 1 {
		filtered := []*types.AuditEntry{}
		for _, entry := range entries {
			matches := true

			// Check contract_id criteria
			if contractID, ok := criteria["contract_id"]; ok && contractID != "" {
				if entry.ContractAddress != contractID {
					matches = false
				}
			}

			// Check actor criteria
			if actor, ok := criteria["actor"]; ok && actor != "" {
				if entry.Actor != actor {
					matches = false
				}
			}

			// Check event_type criteria
			if eventType, ok := criteria["event_type"]; ok && eventType != "" {
				if entry.EventType != eventType {
					matches = false
				}
			}

			if matches {
				filtered = append(filtered, entry)
				if limit > 0 && uint64(len(filtered)) >= limit {
					break
				}
			}
		}
		entries = filtered
	}

	return entries
}

// GetAuditStatistics returns statistics about audit trail for a contract
func (k Keeper) GetAuditStatistics(ctx sdk.Context, contractID string) *types.AuditStatistics {
	entries := k.GetAuditTrail(ctx, contractID, 0) // Get all entries

	stats := &types.AuditStatistics{
		ContractAddress: contractID,
		TotalEntries:    uint64(len(entries)),
		TotalEvents:     uint64(len(entries)), // Alias for TotalEntries
		SuccessCount:    0,
		FailureCount:    0,
		ActionCounts:    make(map[string]uint64),
		EventTypeCounts: make(map[string]uint64),
	}

	for _, entry := range entries {
		if entry.Success {
			stats.SuccessCount++
		} else {
			stats.FailureCount++
		}

		// Count by action
		stats.ActionCounts[entry.Action]++

		// Count by event type
		if entry.EventType != "" {
			stats.EventTypeCounts[entry.EventType]++
		}
	}

	return stats
}

// SetAuditRetentionPolicy sets the audit retention policy in days
func (k Keeper) SetAuditRetentionPolicy(ctx sdk.Context, days uint64) error {
	if days == 0 {
		return fmt.Errorf("retention policy must be greater than 0 days")
	}

	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, days)
	store.Set(types.AuditRetentionPolicyKey, bz)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"audit_retention_policy_updated",
			sdk.NewAttribute("retention_days", fmt.Sprintf("%d", days)),
		),
	)

	return nil
}

// GetAuditRetentionPolicy retrieves the audit retention policy
func (k Keeper) GetAuditRetentionPolicy(ctx sdk.Context) (uint64, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.AuditRetentionPolicyKey)
	if bz == nil {
		return 0, false
	}

	days := binary.BigEndian.Uint64(bz)
	return days, true
}

// DeleteAuditEventsBefore deletes audit events older than the specified cutoff time
// Returns the number of events deleted
func (k Keeper) DeleteAuditEventsBefore(ctx sdk.Context, cutoff time.Time) uint64 {
	store := ctx.KVStore(k.storeKey)

	// Iterate through all audit entries
	iterator := storetypes.KVStorePrefixIterator(store, types.AuditEntriesKeyPrefix)
	defer iterator.Close()

	keysToDelete := [][]byte{}
	deleted := uint64(0)

	for ; iterator.Valid(); iterator.Next() {
		var entry types.AuditEntry
		if err := json.Unmarshal(iterator.Value(), &entry); err != nil {
			continue
		}

		// Check if entry is older than cutoff
		if entry.Timestamp != nil && entry.Timestamp.AsTime().Before(cutoff) {
			keysToDelete = append(keysToDelete, iterator.Key())

			// Also collect index keys to delete
			// We need to find the global sequence for this entry
			// For now, we'll delete the main entry; index cleanup can be done separately
		}
	}

	// Delete collected keys
	for _, key := range keysToDelete {
		store.Delete(key)
		deleted++
	}

	if deleted > 0 {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"audit_events_pruned",
				sdk.NewAttribute("count", fmt.Sprintf("%d", deleted)),
				sdk.NewAttribute("cutoff", cutoff.Format(time.RFC3339)),
			),
		)
	}

	return deleted
}

// ExportAuditTrail exports all audit entries for a contract (same as GetAuditTrail with no limit)
func (k Keeper) ExportAuditTrail(ctx sdk.Context, contractID string) []*types.AuditEntry {
	return k.GetAuditTrail(ctx, contractID, 0) // 0 means no limit
}

// =============================================================================
// Existing methods - preserved for backward compatibility
// =============================================================================

// AddAuditEntry adds an entry to the contract's audit trail
// This is the lower-level method used by other keeper functions
func (k Keeper) AddAuditEntry(ctx sdk.Context, entry *types.AuditEntry) {
	store := ctx.KVStore(k.storeKey)

	// Get next sequence number for this contract
	seqKey := types.AuditSequenceKey(entry.ContractAddress)
	seqBz := store.Get(seqKey)
	var seq uint64
	if seqBz != nil {
		seq = binary.BigEndian.Uint64(seqBz)
	}
	seq++

	// Store new sequence
	newSeqBz := make([]byte, 8)
	binary.BigEndian.PutUint64(newSeqBz, seq)
	store.Set(seqKey, newSeqBz)

	// Set entry ID and store
	entry.Id = seq

	// Ensure Timestamp is set
	if entry.Timestamp == nil {
		entry.Timestamp = timestamppb.New(ctx.BlockTime())
	}

	// Set EventType if not already set
	if entry.EventType == "" && entry.Action != "" {
		entry.EventType = entry.Action
	}

	key := types.AuditEntryKey(entry.ContractAddress, seq)
	bz, err := json.Marshal(entry)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal audit entry: %v", err))
	}
	store.Set(key, bz)

	// Create secondary indexes if we have actor and event type
	if entry.Actor != "" || entry.EventType != "" {
		// Get global sequence
		globalSeqBz := store.Get(types.AuditGlobalSequenceKey)
		var globalSeq uint64
		if globalSeqBz != nil {
			globalSeq = binary.BigEndian.Uint64(globalSeqBz)
		}
		globalSeq++

		globalSeqBz = make([]byte, 8)
		binary.BigEndian.PutUint64(globalSeqBz, globalSeq)
		store.Set(types.AuditGlobalSequenceKey, globalSeqBz)

		// Create indexes
		if entry.Actor != "" {
			actorIndexKey := types.AuditActorIndexKey(entry.Actor, globalSeq)
			store.Set(actorIndexKey, []byte(entry.ContractAddress))
		}

		if entry.EventType != "" {
			typeIndexKey := types.AuditTypeIndexKey(entry.EventType, globalSeq)
			store.Set(typeIndexKey, []byte(entry.ContractAddress))
		}
	}

	// Emit audit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"contract_audit_entry",
			sdk.NewAttribute("contract_address", entry.ContractAddress),
			sdk.NewAttribute("action", entry.Action),
			sdk.NewAttribute("actor", entry.Actor),
			sdk.NewAttribute("success", fmt.Sprintf("%t", entry.Success)),
		),
	)
}

// GetAuditEntries retrieves audit entries for a contract
// Alias for GetAuditTrail for backward compatibility
func (k Keeper) GetAuditEntries(ctx sdk.Context, contractAddr string, limit uint64) []*types.AuditEntry {
	return k.GetAuditTrail(ctx, contractAddr, limit)
}

// GetAuditEntry retrieves a specific audit entry
func (k Keeper) GetAuditEntry(ctx sdk.Context, contractAddr string, id uint64) (*types.AuditEntry, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.AuditEntryKey(contractAddr, id)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var entry types.AuditEntry
	if err := json.Unmarshal(bz, &entry); err != nil {
		return nil, false
	}
	return &entry, true
}

// GetAuditTrailCount returns the total number of audit entries for a contract
func (k Keeper) GetAuditTrailCount(ctx sdk.Context, contractAddr string) uint64 {
	store := ctx.KVStore(k.storeKey)
	seqKey := types.AuditSequenceKey(contractAddr)
	seqBz := store.Get(seqKey)
	if seqBz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(seqBz)
}

// RecordContractExecution records a contract execution in audit trail
func (k Keeper) RecordContractExecution(ctx sdk.Context, contractAddr, executor string, gasUsed uint64, success bool, errorMsg string) {
	details := fmt.Sprintf("Gas used: %d", gasUsed)
	if !success && errorMsg != "" {
		details = fmt.Sprintf("%s, Error: %s", details, errorMsg)
	}

	metadata := map[string]string{
		"gas_used": fmt.Sprintf("%d", gasUsed),
	}
	if errorMsg != "" {
		metadata["error"] = errorMsg
	}

	entry := &types.AuditEntry{
		ContractAddress: contractAddr,
		Timestamp:       timestamppb.New(ctx.BlockTime()),
		Action:          "EXECUTE_CONTRACT",
		EventType:       "EXECUTE",
		Actor:           executor,
		Details:         details,
		Metadata:        metadata,
		Success:         success,
	}

	k.AddAuditEntry(ctx, entry)
	k.UpdateMetricsOnExecution(ctx, contractAddr, gasUsed, success)
}

// RecordContractUpdate records a contract metadata/policy update
func (k Keeper) RecordContractUpdate(ctx sdk.Context, contractAddr, admin, updateType string, details string) {
	entry := &types.AuditEntry{
		ContractAddress: contractAddr,
		Timestamp:       timestamppb.New(ctx.BlockTime()),
		Action:          updateType,
		EventType:       updateType,
		Actor:           admin,
		Details:         details,
		Success:         true,
	}

	k.AddAuditEntry(ctx, entry)
}

// RecordContractStatusChange records a contract status change
func (k Keeper) RecordContractStatusChange(ctx sdk.Context, contractAddr, actor string, oldStatus, newStatus types.ContractStatus, reason string) {
	details := fmt.Sprintf("Status changed from %s to %s. Reason: %s",
		oldStatus.String(), newStatus.String(), reason)

	entry := &types.AuditEntry{
		ContractAddress: contractAddr,
		Timestamp:       timestamppb.New(ctx.BlockTime()),
		Action:          "STATUS_CHANGE",
		EventType:       "STATUS_CHANGE",
		Actor:           actor,
		Details:         details,
		Metadata: map[string]string{
			"old_status": oldStatus.String(),
			"new_status": newStatus.String(),
			"reason":     reason,
		},
		Success: true,
	}

	k.AddAuditEntry(ctx, entry)
}

// PruneOldAuditEntries removes audit entries older than the specified timestamp
// This helps manage storage for high-usage contracts
// Alias for DeleteAuditEventsBefore for backward compatibility
func (k Keeper) PruneOldAuditEntries(ctx sdk.Context, contractAddr string, beforeTimestamp int64) uint64 {
	// Convert Unix timestamp to time.Time
	cutoff := time.Unix(beforeTimestamp, 0)

	store := ctx.KVStore(k.storeKey)
	prefix := types.AuditEntriesPrefix(contractAddr)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	pruned := uint64(0)
	keysToDelete := [][]byte{}

	for ; iterator.Valid(); iterator.Next() {
		var entry types.AuditEntry
		if err := json.Unmarshal(iterator.Value(), &entry); err != nil {
			continue
		}

		// Check timestamp
		if entry.Timestamp != nil && entry.Timestamp.AsTime().Before(cutoff) {
			keysToDelete = append(keysToDelete, iterator.Key())
			pruned++
		}
	}

	// Delete collected keys
	for _, key := range keysToDelete {
		store.Delete(key)
	}

	return pruned
}
