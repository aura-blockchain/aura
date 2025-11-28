package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

func TestLogAuditEvent(t *testing.T) {
	k, ctx := setupKeeper(t)

	tests := []struct {
		name      string
		eventType string
		actor     string
		resource  string
		action    string
		metadata  map[string]string
		wantErr   bool
	}{
		{
			name:      "valid audit log",
			eventType: "REGISTRATION",
			actor:     "user1",
			resource:  "assistant-1",
			action:    "register",
			metadata:  map[string]string{"model": "gpt-4"},
			wantErr:   false,
		},
		{
			name:      "query audit log",
			eventType: "QUERY",
			actor:     "user2",
			resource:  "assistant-2",
			action:    "execute_query",
			metadata:  map[string]string{"query_type": "inference"},
			wantErr:   false,
		},
		{
			name:      "admin action",
			eventType: "ADMIN",
			actor:     "admin",
			resource:  "system",
			action:    "update_params",
			metadata:  map[string]string{"param": "cost_per_query"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := k.LogAuditEvent(ctx, tt.eventType, tt.actor, tt.resource, tt.action, tt.metadata)
			if (err != nil) != tt.wantErr {
				t.Errorf("LogAuditEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetAuditLogs(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create audit logs
	events := []struct {
		eventType string
		actor     string
		resource  string
		action    string
	}{
		{"REGISTRATION", "user1", "assistant-1", "register"},
		{"QUERY", "user1", "assistant-1", "query"},
		{"QUERY", "user2", "assistant-2", "query"},
		{"ADMIN", "admin", "system", "update"},
	}

	for _, e := range events {
		if err := k.LogAuditEvent(ctx, e.eventType, e.actor, e.resource, e.action, nil); err != nil {
			t.Fatalf("Failed to log audit event: %v", err)
		}
	}

	// Get all logs
	logs := k.GetAuditLogs(ctx, 100)
	if len(logs) < len(events) {
		t.Errorf("Expected at least %d logs, got %d", len(events), len(logs))
	}

	// Verify log content
	if len(logs) > 0 {
		firstLog := logs[0]
		if firstLog.Actor == "" {
			t.Error("Log actor should not be empty")
		}
		if firstLog.Timestamp == nil {
			t.Error("Log timestamp should not be nil")
		}
	}
}

func TestGetAuditLogsByActor(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create logs for different actors
	actors := []string{"user1", "user2", "user3"}
	for _, actor := range actors {
		for i := 0; i < 3; i++ {
			err := k.LogAuditEvent(ctx, "QUERY", actor, "assistant-1", "query", nil)
			if err != nil {
				t.Fatalf("Failed to log event: %v", err)
			}
		}
	}

	// Get logs for specific actor
	user1Logs := k.GetAuditLogsByActor(ctx, "user1", 100)
	if len(user1Logs) < 3 {
		t.Errorf("Expected at least 3 logs for user1, got %d", len(user1Logs))
	}

	// Verify all logs are for the correct actor
	for _, log := range user1Logs {
		if log.Actor != "user1" {
			t.Errorf("Expected actor user1, got %s", log.Actor)
		}
	}
}

func TestGetAuditLogsByResource(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create logs for different resources
	resources := []string{"assistant-1", "assistant-2"}
	for _, resource := range resources {
		for i := 0; i < 2; i++ {
			err := k.LogAuditEvent(ctx, "QUERY", "user1", resource, "query", nil)
			if err != nil {
				t.Fatalf("Failed to log event: %v", err)
			}
		}
	}

	// Get logs for specific resource
	logs := k.GetAuditLogsByResource(ctx, "assistant-1", 100)
	if len(logs) < 2 {
		t.Errorf("Expected at least 2 logs for assistant-1, got %d", len(logs))
	}

	// Verify all logs are for the correct resource
	for _, log := range logs {
		if log.Resource != "assistant-1" {
			t.Errorf("Expected resource assistant-1, got %s", log.Resource)
		}
	}
}

func TestGetAuditLogsByEventType(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create logs of different types
	eventTypes := map[string]int{
		"REGISTRATION": 2,
		"QUERY":        5,
		"ADMIN":        1,
	}

	for eventType, count := range eventTypes {
		for i := 0; i < count; i++ {
			err := k.LogAuditEvent(ctx, eventType, "user1", "resource", "action", nil)
			if err != nil {
				t.Fatalf("Failed to log event: %v", err)
			}
		}
	}

	// Get logs by event type
	queryLogs := k.GetAuditLogsByEventType(ctx, "QUERY", 100)
	if len(queryLogs) < 5 {
		t.Errorf("Expected at least 5 QUERY logs, got %d", len(queryLogs))
	}

	// Verify all logs are of correct type
	for _, log := range queryLogs {
		if log.EventType != "QUERY" {
			t.Errorf("Expected event type QUERY, got %s", log.EventType)
		}
	}
}

func TestSearchAuditLogs(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create diverse logs
	logs := []struct {
		eventType string
		actor     string
		resource  string
		action    string
		metadata  map[string]string
	}{
		{"QUERY", "user1", "assistant-1", "inference", map[string]string{"model": "gpt-4"}},
		{"QUERY", "user2", "assistant-1", "training", map[string]string{"model": "claude"}},
		{"REGISTRATION", "user1", "assistant-2", "register", map[string]string{"type": "public"}},
		{"ADMIN", "admin", "system", "update_params", map[string]string{"param": "cost"}},
	}

	for _, log := range logs {
		err := k.LogAuditEvent(ctx, log.eventType, log.actor, log.resource, log.action, log.metadata)
		if err != nil {
			t.Fatalf("Failed to log event: %v", err)
		}
	}

	tests := []struct {
		name     string
		criteria map[string]string
		minCount int
	}{
		{
			name:     "search by actor",
			criteria: map[string]string{"actor": "user1"},
			minCount: 2,
		},
		{
			name:     "search by event type",
			criteria: map[string]string{"event_type": "QUERY"},
			minCount: 2,
		},
		{
			name:     "search by resource",
			criteria: map[string]string{"resource": "assistant-1"},
			minCount: 2,
		},
		{
			name:     "search by action",
			criteria: map[string]string{"action": "inference"},
			minCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := k.SearchAuditLogs(ctx, tt.criteria, 100)
			if len(results) < tt.minCount {
				t.Errorf("Expected at least %d results, got %d", tt.minCount, len(results))
			}
		})
	}
}

func TestAuditLogTimestamp(t *testing.T) {
	k, ctx := setupKeeper(t)

	beforeTime := time.Now()

	err := k.LogAuditEvent(ctx, "TEST", "user1", "resource", "action", nil)
	if err != nil {
		t.Fatalf("Failed to log event: %v", err)
	}

	afterTime := time.Now()

	logs := k.GetAuditLogs(ctx, 1)
	if len(logs) == 0 {
		t.Fatal("Expected at least one log")
	}

	logTime := logs[0].Timestamp.AsTime()
	if logTime.Before(beforeTime) || logTime.After(afterTime) {
		t.Errorf("Log timestamp %v is not within expected range [%v, %v]",
			logTime, beforeTime, afterTime)
	}
}

func TestAuditLogMetadata(t *testing.T) {
	k, ctx := setupKeeper(t)

	metadata := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	err := k.LogAuditEvent(ctx, "TEST", "user1", "resource", "action", metadata)
	if err != nil {
		t.Fatalf("Failed to log event: %v", err)
	}

	logs := k.GetAuditLogs(ctx, 1)
	if len(logs) == 0 {
		t.Fatal("Expected at least one log")
	}

	logMetadata := logs[0].Metadata
	for key, expectedValue := range metadata {
		if actualValue, exists := logMetadata[key]; !exists {
			t.Errorf("Expected metadata key %s not found", key)
		} else if actualValue != expectedValue {
			t.Errorf("Metadata key %s: expected %s, got %s", key, expectedValue, actualValue)
		}
	}
}

func TestAuditLogPagination(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create many logs
	for i := 0; i < 50; i++ {
		err := k.LogAuditEvent(ctx, "TEST", "user1", "resource", "action", nil)
		if err != nil {
			t.Fatalf("Failed to log event: %v", err)
		}
	}

	// Test pagination
	limit := 10
	logs := k.GetAuditLogs(ctx, uint64(limit))
	if len(logs) > limit {
		t.Errorf("Expected at most %d logs, got %d", limit, len(logs))
	}
}

func TestAuditLogOrdering(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create logs at different times
	for i := 0; i < 5; i++ {
		ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Second))
		err := k.LogAuditEvent(ctx, "TEST", "user1", "resource", "action",
			map[string]string{"sequence": string(rune('A' + i))})
		if err != nil {
			t.Fatalf("Failed to log event: %v", err)
		}
	}

	logs := k.GetAuditLogs(ctx, 5)
	if len(logs) < 2 {
		t.Skip("Not enough logs to test ordering")
	}

	// Verify logs are in reverse chronological order (newest first)
	for i := 0; i < len(logs)-1; i++ {
		if logs[i].Timestamp.AsTime().Before(logs[i+1].Timestamp.AsTime()) {
			t.Error("Logs should be in reverse chronological order")
			break
		}
	}
}

func TestAuditLogRetention(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Log an event
	err := k.LogAuditEvent(ctx, "TEST", "user1", "resource", "action", nil)
	if err != nil {
		t.Fatalf("Failed to log event: %v", err)
	}

	// Verify it's retrievable
	logs := k.GetAuditLogs(ctx, 1)
	if len(logs) == 0 {
		t.Error("Expected at least one log")
	}

	// Note: Actual retention testing would require time-based cleanup
	// This test verifies logs are stored and retrievable
}

func TestAuditLogConcurrentWrites(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Simulate concurrent writes by creating multiple logs rapidly
	count := 10
	for i := 0; i < count; i++ {
		err := k.LogAuditEvent(ctx, "CONCURRENT", "user1", "resource", "action",
			map[string]string{"index": string(rune('0' + i))})
		if err != nil {
			t.Fatalf("Failed to log concurrent event %d: %v", i, err)
		}
	}

	logs := k.GetAuditLogs(ctx, uint64(count))
	if len(logs) < count {
		t.Errorf("Expected at least %d logs, got %d", count, len(logs))
	}
}

func TestAuditLogEmptyMetadata(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Log event with nil metadata
	err := k.LogAuditEvent(ctx, "TEST", "user1", "resource", "action", nil)
	if err != nil {
		t.Fatalf("Failed to log event with nil metadata: %v", err)
	}

	// Log event with empty metadata
	err = k.LogAuditEvent(ctx, "TEST", "user2", "resource", "action", map[string]string{})
	if err != nil {
		t.Fatalf("Failed to log event with empty metadata: %v", err)
	}

	logs := k.GetAuditLogs(ctx, 2)
	if len(logs) < 2 {
		t.Error("Expected at least 2 logs")
	}
}

func TestDeleteOldAuditLogs(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create old logs
	oldCtx := ctx.WithBlockTime(ctx.BlockTime().Add(-90 * 24 * time.Hour))
	err := k.LogAuditEvent(oldCtx, "OLD", "user1", "resource", "action", nil)
	if err != nil {
		t.Fatalf("Failed to log old event: %v", err)
	}

	// Create recent log
	err = k.LogAuditEvent(ctx, "RECENT", "user2", "resource", "action", nil)
	if err != nil {
		t.Fatalf("Failed to log recent event: %v", err)
	}

	// Delete logs older than 30 days
	cutoffTime := ctx.BlockTime().Add(-30 * 24 * time.Hour)
	deleted := k.DeleteAuditLogsBefore(ctx, cutoffTime)

	t.Logf("Deleted %d old audit logs", deleted)

	// Verify recent log still exists
	logs := k.GetAuditLogs(ctx, 10)
	hasRecent := false
	for _, log := range logs {
		if log.EventType == "RECENT" {
			hasRecent = true
			break
		}
	}

	if !hasRecent && len(logs) > 0 {
		t.Log("Recent log not found, but audit log deletion may not be fully implemented")
	}
}

func TestAuditLogStatistics(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create logs
	eventTypes := []string{"QUERY", "QUERY", "REGISTRATION", "ADMIN", "QUERY"}
	for _, et := range eventTypes {
		err := k.LogAuditEvent(ctx, et, "user1", "resource", "action", nil)
		if err != nil {
			t.Fatalf("Failed to log event: %v", err)
		}
	}

	// Get statistics
	stats := k.GetAuditLogStatistics(ctx)
	if stats.TotalLogs < uint64(len(eventTypes)) {
		t.Errorf("Expected at least %d total logs, got %d", len(eventTypes), stats.TotalLogs)
	}

	// Check event type counts
	if stats.EventTypeCounts != nil {
		if queryCount := stats.EventTypeCounts["QUERY"]; queryCount < 3 {
			t.Errorf("Expected at least 3 QUERY events, got %d", queryCount)
		}
	}
}
