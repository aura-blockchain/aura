package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

// AuditLog represents an audit log entry for AI operations
type AuditLog struct {
	ID             uint64
	Timestamp      time.Time
	EventType      string
	Actor          string
	Resource       string
	Action         string
	UserAddress    string
	OperationType  string
	ModelHash      string
	QueryHash      string
	InputSize      uint64
	OutputSize     uint64
	Success        bool
	ErrorMessage   string
	CostEstimate   CostEstimate
	ProcessingTime uint64 // milliseconds
	IPAddress      string // Optional
	UserAgent      string // Optional
	Metadata       map[string]string
}

// LogAuditEvent provides a high-level helper used by tests to log an audit entry.
func (k Keeper) LogAuditEvent(ctx sdk.Context, eventType, actor, resource, action string, metadata map[string]string) error {
	entry := AuditLog{
		EventType:     eventType,
		Actor:         actor,
		Resource:      resource,
		Action:        action,
		UserAddress:   actor,
		OperationType: eventType,
		Metadata:      metadata,
		Success:       true,
	}
	return k.LogAIOperation(ctx, entry)
}

// LogAIOperation logs an AI operation for audit purposes
func (k Keeper) LogAIOperation(ctx sdk.Context, log AuditLog) error {
	log.ID = k.getNextAuditLogID(ctx)
	log.Timestamp = ctx.BlockTime()

	store := ctx.KVStore(k.storeKey)
	key := types.AuditLogKey(log.ID)

	bz, err := json.Marshal(&log)
	if err != nil {
		return err
	}

	store.Set(key, bz)

	// Emit audit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"ai_operation_audit",
			sdk.NewAttribute("audit_id", fmt.Sprintf("%d", log.ID)),
			sdk.NewAttribute("user", log.UserAddress),
			sdk.NewAttribute("operation", log.OperationType),
			sdk.NewAttribute("model", log.ModelHash),
			sdk.NewAttribute("success", fmt.Sprintf("%t", log.Success)),
		),
	)

	return nil
}

// GetAuditLog retrieves an audit log by ID
func (k Keeper) GetAuditLog(ctx sdk.Context, id uint64) (AuditLog, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.AuditLogKey(id)

	bz := store.Get(key)
	if len(bz) == 0 {
		return AuditLog{}, false
	}

	var log AuditLog
	if err := json.Unmarshal(bz, &log); err != nil {
		return AuditLog{}, false
	}

	return log, true
}

// ListAuditLogs returns audit logs with optional filters
func (k Keeper) ListAuditLogs(ctx sdk.Context, filters AuditLogFilters, limit uint64) []AuditLog {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuditLogKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var logs []AuditLog
	count := uint64(0)

	for ; iterator.Valid() && count < limit; iterator.Next() {
		var log AuditLog
		if err := json.Unmarshal(iterator.Value(), &log); err != nil {
			continue
		}

		// Apply filters
		if !matchesFilters(log, filters) {
			continue
		}

		logs = append(logs, log)
		count++
	}

	return logs
}

// GetAuditLogs returns up to limit audit logs without additional filtering.
func (k Keeper) GetAuditLogs(ctx sdk.Context, limit uint64) []AuditLog {
	return k.ListAuditLogs(ctx, AuditLogFilters{}, limit)
}

// GetUserAuditLogs returns audit logs for a specific user
func (k Keeper) GetUserAuditLogs(ctx sdk.Context, userAddr string, limit uint64) []AuditLog {
	filters := AuditLogFilters{
		UserAddress: userAddr,
	}
	return k.ListAuditLogs(ctx, filters, limit)
}

// GetAuditLogsByActor aliases GetUserAuditLogs for test compatibility.
func (k Keeper) GetAuditLogsByActor(ctx sdk.Context, actor string, limit uint64) []AuditLog {
	return k.GetUserAuditLogs(ctx, actor, limit)
}

// GetAuditLogsByResource filters by resource.
func (k Keeper) GetAuditLogsByResource(ctx sdk.Context, resource string, limit uint64) []AuditLog {
	return k.ListAuditLogs(ctx, AuditLogFilters{ResourceID: resource}, limit)
}

// GetAuditLogsByEventType filters by event type.
func (k Keeper) GetAuditLogsByEventType(ctx sdk.Context, eventType string, limit uint64) []AuditLog {
	return k.ListAuditLogs(ctx, AuditLogFilters{OperationType: eventType}, limit)
}

// GetAuditStats returns audit statistics
func (k Keeper) GetAuditStats(ctx sdk.Context, since time.Time) AuditStats {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuditLogKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	stats := AuditStats{
		OperationCounts: make(map[string]uint64),
		ModelUsage:      make(map[string]uint64),
		UserActivity:    make(map[string]uint64),
	}

	for ; iterator.Valid(); iterator.Next() {
		var log AuditLog
		if err := json.Unmarshal(iterator.Value(), &log); err != nil {
			continue
		}

		// Filter by time
		if log.Timestamp.Before(since) {
			continue
		}

		stats.TotalOperations++
		if log.Success {
			stats.SuccessfulOperations++
		} else {
			stats.FailedOperations++
		}

		stats.OperationCounts[log.OperationType]++
		stats.ModelUsage[log.ModelHash]++
		stats.UserActivity[log.UserAddress]++

		stats.TotalInputSize += log.InputSize
		stats.TotalOutputSize += log.OutputSize
		stats.TotalProcessingTime += log.ProcessingTime
	}

	if stats.TotalOperations > 0 {
		stats.AverageProcessingTime = stats.TotalProcessingTime / stats.TotalOperations
	}

	return stats
}

// PruneAuditLogs removes old audit logs
func (k Keeper) PruneAuditLogs(ctx sdk.Context, before time.Time) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuditLogKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	count := uint64(0)
	for ; iterator.Valid(); iterator.Next() {
		var log AuditLog
		if err := json.Unmarshal(iterator.Value(), &log); err != nil {
			continue
		}

		if log.Timestamp.Before(before) {
			store.Delete(iterator.Key())
			count++
		}
	}

	return count
}

// getNextAuditLogID returns the next audit log ID
func (k Keeper) getNextAuditLogID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := types.AuditLogCounterKey

	bz := store.Get(key)
	var id uint64 = 1

	if len(bz) > 0 {
		id = sdk.BigEndianToUint64(bz) + 1
	}

	store.Set(key, sdk.Uint64ToBigEndian(id))
	return id
}

// AuditLogFilters defines filters for querying audit logs
type AuditLogFilters struct {
	UserAddress   string
	ResourceID    string
	ModelHash     string
	OperationType string
	SuccessOnly   bool
	FailuresOnly  bool
	FromTime      *time.Time
	ToTime        *time.Time
}

// matchesFilters checks if a log matches the filters
func matchesFilters(log AuditLog, filters AuditLogFilters) bool {
	if filters.UserAddress != "" && log.UserAddress != filters.UserAddress {
		return false
	}
	if filters.ModelHash != "" && log.ModelHash != filters.ModelHash {
		return false
	}
	if filters.OperationType != "" && log.OperationType != filters.OperationType {
		return false
	}
	if filters.ResourceID != "" && log.Resource != filters.ResourceID {
		return false
	}
	if filters.SuccessOnly && !log.Success {
		return false
	}
	if filters.FailuresOnly && log.Success {
		return false
	}
	if filters.FromTime != nil && log.Timestamp.Before(*filters.FromTime) {
		return false
	}
	if filters.ToTime != nil && log.Timestamp.After(*filters.ToTime) {
		return false
	}
	return true
}

// AuditStats represents audit statistics
type AuditStats struct {
	TotalOperations       uint64
	SuccessfulOperations  uint64
	FailedOperations      uint64
	OperationCounts       map[string]uint64
	ModelUsage            map[string]uint64
	UserActivity          map[string]uint64
	TotalInputSize        uint64
	TotalOutputSize       uint64
	TotalProcessingTime   uint64
	AverageProcessingTime uint64
}
