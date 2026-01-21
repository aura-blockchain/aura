// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"sort"
	"strings"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// LogAudit creates a new audit log entry in the KVStore.
//
// CRITICAL: This method now uses KVStore for consensus-safe storage.
// All validators will have identical audit log state.
//
// Parameters:
//   - ctx: SDK context (required for KVStore access and deterministic time)
//   - actor: The address/identifier of the entity performing the action
//   - action: The type of action being performed
//   - resource: The resource being acted upon
//   - status: The result status (e.g., "success", "failed")
//   - metadata: Additional key-value metadata
//   - errorMsg: Error message if status is "failed"
//
// Security considerations:
//   - Uses ctx.BlockTime() for determinism (NEVER time.Now())
//   - Stored in KVStore, persisted across restarts
//   - Automatically cleans up old logs when threshold exceeded
//   - Generates sequential IDs for ordering
func (k *Keeper) LogAudit(ctx sdk.Context, actor, action, resource, status string, metadata map[string]string, errorMsg string) {
	log := &authproto.AuditLog{
		Actor:        actor,
		Action:       action,
		Resource:     resource,
		Result:       status,
		Timestamp:    ctx.BlockTime(),
		Metadata:     metadata,
		ErrorMessage: errorMsg,
	}

	// Store in KVStore (SetAuditLog will generate ID)
	if err := k.SetAuditLog(ctx, log); err != nil {
		// Log error but don't fail the transaction
		// Audit logging is important but should not block operations
		ctx.Logger().Error("failed to store audit log", "error", err, "actor", actor, "action", action)
	}
}

// GetAuditLogs retrieves audit logs with optional filters
//
// All filtering is done in-memory after loading from KVStore.
// For production use with large datasets, consider pagination.
//
// Parameters:
//   - ctx: SDK context
//   - actor: Filter by actor (empty string = no filter)
//   - action: Filter by action (empty string = no filter)
//   - startTime: Unix timestamp for start of range (0 = no filter)
//   - endTime: Unix timestamp for end of range (0 = no filter)
//   - limit: Maximum number of results (0 = no limit)
//
// Returns:
//   - Sorted list of audit logs (most recent first)
func (k *Keeper) GetAuditLogs(ctx sdk.Context, actor, action string, startTime, endTime int64, limit uint64) []*authproto.AuditLog {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, AuditLogsKeyPrefix)
	defer iterator.Close()

	filtered := make([]*authproto.AuditLog, 0, 64)

	for ; iterator.Valid(); iterator.Next() {
		var log authproto.AuditLog
		if err := k.cdc.Unmarshal(iterator.Value(), &log); err != nil {
			continue // Skip malformed entries
		}

		// Apply filters
		if actor != "" && log.Actor != actor {
			continue
		}
		if action != "" && log.Action != action {
			continue
		}
		if startTime > 0 && !log.Timestamp.IsZero() && log.Timestamp.Unix() < startTime {
			continue
		}
		if endTime > 0 && !log.Timestamp.IsZero() && log.Timestamp.Unix() > endTime {
			continue
		}

		filtered = append(filtered, &log)
	}

	// Sort by timestamp (most recent first)
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Timestamp.IsZero() || filtered[j].Timestamp.IsZero() {
			return false
		}
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	// Apply limit
	if limit > 0 && uint64(len(filtered)) > limit {
		filtered = filtered[:limit]
	}

	return filtered
}

// GetAuditLogsByActor retrieves all audit logs for a specific actor
func (k *Keeper) GetAuditLogsByActor(ctx sdk.Context, actor string, limit uint64) []*authproto.AuditLog {
	return k.GetAuditLogs(ctx, actor, "", 0, 0, limit)
}

// GetAuditLogsByAction retrieves all audit logs for a specific action type
func (k *Keeper) GetAuditLogsByAction(ctx sdk.Context, action string, limit uint64) []*authproto.AuditLog {
	return k.GetAuditLogs(ctx, "", action, 0, 0, limit)
}

// GetAuditLogsByResource retrieves all audit logs for a specific resource
func (k *Keeper) GetAuditLogsByResource(ctx sdk.Context, resource string, limit uint64) []*authproto.AuditLog {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, AuditLogsKeyPrefix)
	defer iterator.Close()

	filtered := make([]*authproto.AuditLog, 0, 64)

	for ; iterator.Valid(); iterator.Next() {
		var log authproto.AuditLog
		if err := k.cdc.Unmarshal(iterator.Value(), &log); err != nil {
			continue
		}

		if log.Resource == resource {
			filtered = append(filtered, &log)
		}
	}

	// Sort by timestamp (most recent first)
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Timestamp.IsZero() || filtered[j].Timestamp.IsZero() {
			return false
		}
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	// Apply limit
	if limit > 0 && uint64(len(filtered)) > limit {
		filtered = filtered[:limit]
	}

	return filtered
}

// GetRecentAuditLogs retrieves the most recent audit logs
func (k *Keeper) GetRecentAuditLogs(ctx sdk.Context, limit uint64) []*authproto.AuditLog {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, AuditLogsKeyPrefix)
	defer iterator.Close()

	all := make([]*authproto.AuditLog, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var log authproto.AuditLog
		if err := k.cdc.Unmarshal(iterator.Value(), &log); err != nil {
			continue
		}
		all = append(all, &log)
	}

	// Sort by timestamp (most recent first)
	sort.Slice(all, func(i, j int) bool {
		if all[i].Timestamp.IsZero() || all[j].Timestamp.IsZero() {
			return false
		}
		return all[i].Timestamp.After(all[j].Timestamp)
	})

	// Apply limit
	if limit > 0 && uint64(len(all)) > limit {
		all = all[:limit]
	}

	return all
}

// GetAuditLogsByTimeRange retrieves audit logs within a time range
func (k *Keeper) GetAuditLogsByTimeRange(ctx sdk.Context, startTime, endTime int64, limit uint64) []*authproto.AuditLog {
	return k.GetAuditLogs(ctx, "", "", startTime, endTime, limit)
}

// SearchAuditLogs searches audit logs with multiple criteria
func (k *Keeper) SearchAuditLogs(ctx sdk.Context, criteria map[string]string, limit uint64) []*authproto.AuditLog {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, AuditLogsKeyPrefix)
	defer iterator.Close()

	filtered := make([]*authproto.AuditLog, 0, 64)

	for ; iterator.Valid(); iterator.Next() {
		var log authproto.AuditLog
		if err := k.cdc.Unmarshal(iterator.Value(), &log); err != nil {
			continue
		}

		match := true

		// Check each criterion
		for key, value := range criteria {
			switch key {
			case "actor":
				if !strings.Contains(log.Actor, value) {
					match = false
				}
			case "action":
				if !strings.Contains(log.Action, value) {
					match = false
				}
			case "resource":
				if !strings.Contains(log.Resource, value) {
					match = false
				}
			case "status":
				if log.Result != value {
					match = false
				}
			}
		}

		if match {
			filtered = append(filtered, &log)
		}
	}

	// Sort by timestamp (most recent first)
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Timestamp.IsZero() || filtered[j].Timestamp.IsZero() {
			return false
		}
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	// Apply limit
	if limit > 0 && uint64(len(filtered)) > limit {
		filtered = filtered[:limit]
	}

	return filtered
}

// CountAuditLogs counts total audit logs
func (k *Keeper) CountAuditLogs(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, AuditLogsKeyPrefix)
	defer iterator.Close()

	count := uint64(0)
	for ; iterator.Valid(); iterator.Next() {
		count++
	}
	return count
}

// CountAuditLogsByActor counts audit logs for a specific actor
func (k *Keeper) CountAuditLogsByActor(ctx sdk.Context, actor string) uint64 {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, AuditLogsKeyPrefix)
	defer iterator.Close()

	count := uint64(0)
	for ; iterator.Valid(); iterator.Next() {
		var log authproto.AuditLog
		if err := k.cdc.Unmarshal(iterator.Value(), &log); err != nil {
			continue
		}
		if log.Actor == actor {
			count++
		}
	}
	return count
}
