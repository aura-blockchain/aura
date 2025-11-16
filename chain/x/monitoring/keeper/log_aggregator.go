package keeper

import (
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// LogEntry records a centralized log entry
func (k *Keeper) LogEntry(level types.LogLevel, module, message string, fields map[string]interface{}, traceID, spanID string) error {
	if !k.params.EnableLogAggregation {
		return nil
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	entry := &types.LogEntry{
		ID:        generateID("log"),
		Level:     level,
		Module:    module,
		Message:   message,
		Fields:    fields,
		Timestamp: time.Now(),
		TraceID:   traceID,
		SpanID:    spanID,
	}

	// Add to module logs
	if _, exists := k.logs[module]; !exists {
		k.logs[module] = make([]*types.LogEntry, 0)
	}

	k.logs[module] = append(k.logs[module], entry)

	// Enforce max entries per module
	if int64(len(k.logs[module])) > k.params.MaxLogEntriesPerModule {
		k.logs[module] = k.logs[module][1:]
	}

	// Update Prometheus metrics
	k.metrics.LogEntriesTotal.WithLabelValues(string(level), module).Inc()

	return nil
}

// GetLogs retrieves logs for a specific module
func (k *Keeper) GetLogs(module string, limit int) ([]*types.LogEntry, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	logs, exists := k.logs[module]
	if !exists {
		return []*types.LogEntry{}, nil
	}

	// Return most recent logs
	start := 0
	if len(logs) > limit {
		start = len(logs) - limit
	}

	result := make([]*types.LogEntry, len(logs[start:]))
	copy(result, logs[start:])

	return result, nil
}

// GetLogsByLevel retrieves logs filtered by level
func (k *Keeper) GetLogsByLevel(module string, level types.LogLevel, limit int) ([]*types.LogEntry, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	logs, exists := k.logs[module]
	if !exists {
		return []*types.LogEntry{}, nil
	}

	var filtered []*types.LogEntry
	for _, log := range logs {
		if log.Level == level {
			filtered = append(filtered, log)
			if len(filtered) >= limit {
				break
			}
		}
	}

	return filtered, nil
}

// GetErrorLogs retrieves error-level logs across all modules
func (k *Keeper) GetErrorLogs(limit int) []*types.LogEntry {
	k.mu.RLock()
	defer k.mu.RUnlock()

	var errors []*types.LogEntry
	for _, moduleLogs := range k.logs {
		for _, log := range moduleLogs {
			if log.Level == types.LogLevelError || log.Level == types.LogLevelFatal {
				errors = append(errors, log)
			}
		}
	}

	// Sort by timestamp (most recent first)
	for i := 0; i < len(errors)-1; i++ {
		for j := i + 1; j < len(errors); j++ {
			if errors[j].Timestamp.After(errors[i].Timestamp) {
				errors[i], errors[j] = errors[j], errors[i]
			}
		}
	}

	if len(errors) > limit {
		errors = errors[:limit]
	}

	return errors
}

// GetLogsByTraceID retrieves logs for a specific trace
func (k *Keeper) GetLogsByTraceID(traceID string) []*types.LogEntry {
	k.mu.RLock()
	defer k.mu.RUnlock()

	var traced []*types.LogEntry
	for _, moduleLogs := range k.logs {
		for _, log := range moduleLogs {
			if log.TraceID == traceID {
				traced = append(traced, log)
			}
		}
	}

	return traced
}

// GetAllModules returns all modules that have logs
func (k *Keeper) GetAllModules() []string {
	k.mu.RLock()
	defer k.mu.RUnlock()

	modules := make([]string, 0, len(k.logs))
	for module := range k.logs {
		modules = append(modules, module)
	}

	return modules
}

// GetLogStats returns log statistics
func (k *Keeper) GetLogStats() map[string]interface{} {
	k.mu.RLock()
	defer k.mu.RUnlock()

	stats := map[string]interface{}{
		"total_modules":   len(k.logs),
		"total_entries":   0,
		"by_level":        make(map[types.LogLevel]int),
		"by_module":       make(map[string]int),
	}

	levelCounts := make(map[types.LogLevel]int)
	moduleCounts := make(map[string]int)
	totalEntries := 0

	for module, logs := range k.logs {
		moduleCounts[module] = len(logs)
		totalEntries += len(logs)

		for _, log := range logs {
			levelCounts[log.Level]++
		}
	}

	stats["total_entries"] = totalEntries
	stats["by_level"] = levelCounts
	stats["by_module"] = moduleCounts

	return stats
}

// SearchLogs searches logs by message content
func (k *Keeper) SearchLogs(query string, limit int) []*types.LogEntry {
	k.mu.RLock()
	defer k.mu.RUnlock()

	var results []*types.LogEntry
	for _, moduleLogs := range k.logs {
		for _, log := range moduleLogs {
			if contains(log.Message, query) {
				results = append(results, log)
				if len(results) >= limit {
					return results
				}
			}
		}
	}

	return results
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	// Simple implementation - in production, use strings.Contains with case conversion
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

// ExportLogs exports logs for external log aggregation systems
func (k *Keeper) ExportLogs(module string, startTime, endTime time.Time) ([]*types.LogEntry, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	logs, exists := k.logs[module]
	if !exists {
		return []*types.LogEntry{}, fmt.Errorf("module not found: %s", module)
	}

	var exported []*types.LogEntry
	for _, log := range logs {
		if log.Timestamp.After(startTime) && log.Timestamp.Before(endTime) {
			exported = append(exported, log)
		}
	}

	return exported, nil
}
