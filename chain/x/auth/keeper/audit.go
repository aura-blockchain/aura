package keeper

import (
	"fmt"
	"sort"
	"strings"
	"time"

	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// GetAuditLogs retrieves audit logs with optional filters
func (k *Keeper) GetAuditLogs(actor, action string, startTime, endTime int64, limit uint64) []*authproto.AuditLog {
	k.mu.RLock()
	defer k.mu.RUnlock()

	filtered := make([]*authproto.AuditLog, 0)

	for _, logs := range k.auditLogs {
		for _, log := range logs {
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

			filtered = append(filtered, log)
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

// GetAuditLogsByActor retrieves all audit logs for a specific actor
func (k *Keeper) GetAuditLogsByActor(actor string, limit uint64) []*authproto.AuditLog {
	return k.GetAuditLogs(actor, "", 0, 0, limit)
}

// GetAuditLogsByAction retrieves all audit logs for a specific action type
func (k *Keeper) GetAuditLogsByAction(action string, limit uint64) []*authproto.AuditLog {
	return k.GetAuditLogs("", action, 0, 0, limit)
}

// GetAuditLogsByResource retrieves all audit logs for a specific resource
func (k *Keeper) GetAuditLogsByResource(resource string, limit uint64) []*authproto.AuditLog {
	k.mu.RLock()
	defer k.mu.RUnlock()

	filtered := make([]*authproto.AuditLog, 0)

	for _, logs := range k.auditLogs {
		for _, log := range logs {
			if log.Resource == resource {
				filtered = append(filtered, log)
			}
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
func (k *Keeper) GetRecentAuditLogs(limit uint64) []*authproto.AuditLog {
	k.mu.RLock()
	defer k.mu.RUnlock()

	all := make([]*authproto.AuditLog, 0)
	for _, logs := range k.auditLogs {
		all = append(all, logs...)
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

// LogAudit creates a new audit log entry.
// The blockTime parameter must be ctx.BlockTime() for determinism in consensus.
// NEVER use time.Now() - it causes non-determinism across validators.
// Note: The ctx parameter accepts interface{} for backwards compatibility, but
// callers with access to sdk.Context should pass ctx.BlockTime() as blockTime.
func (k *Keeper) LogAudit(ctx interface{}, actor, action, resource, status string, metadata map[string]string, errorMsg string, blockTime time.Time) {
	k.mu.Lock()
	defer k.mu.Unlock()

	log := &authproto.AuditLog{
		Id:           fmt.Sprintf("%s-%d", actor, blockTime.UnixNano()),
		Actor:        actor,
		Action:       action,
		Resource:     resource,
		Result:       status,
		Timestamp:    blockTime,
		Metadata:     metadata,
		ErrorMessage: errorMsg,
	}

	// Store by actor
	if k.auditLogs[actor] == nil {
		k.auditLogs[actor] = make([]*authproto.AuditLog, 0)
	}
	k.auditLogs[actor] = append(k.auditLogs[actor], log)
}

// GetAuditLogsByTimeRange retrieves audit logs within a time range
func (k *Keeper) GetAuditLogsByTimeRange(startTime, endTime int64, limit uint64) []*authproto.AuditLog {
	return k.GetAuditLogs("", "", startTime, endTime, limit)
}

// SearchAuditLogs searches audit logs with multiple criteria
func (k *Keeper) SearchAuditLogs(criteria map[string]string, limit uint64) []*authproto.AuditLog {
	k.mu.RLock()
	defer k.mu.RUnlock()

	filtered := make([]*authproto.AuditLog, 0)

	for _, logs := range k.auditLogs {
		for _, log := range logs {
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
				filtered = append(filtered, log)
			}
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
func (k *Keeper) CountAuditLogs() uint64 {
	k.mu.RLock()
	defer k.mu.RUnlock()

	count := uint64(0)
	for _, logs := range k.auditLogs {
		count += uint64(len(logs))
	}
	return count
}

// CountAuditLogsByActor counts audit logs for a specific actor
func (k *Keeper) CountAuditLogsByActor(actor string) uint64 {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if logs, exists := k.auditLogs[actor]; exists {
		return uint64(len(logs))
	}
	return 0
}
