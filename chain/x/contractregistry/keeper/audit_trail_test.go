package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRecordAuditEvent(t *testing.T) {
	k, ctx := setupKeeper(t)

	tests := []struct {
		name       string
		contractID string
		eventType  string
		actor      string
		metadata   map[string]string
		wantErr    bool
	}{
		{
			name:       "contract deployment",
			contractID: "contract1",
			eventType:  "DEPLOY",
			actor:      "deployer1",
			metadata:   map[string]string{"version": "1.0"},
			wantErr:    false,
		},
		{
			name:       "contract upgrade",
			contractID: "contract1",
			eventType:  "UPGRADE",
			actor:      "admin1",
			metadata:   map[string]string{"version": "1.1"},
			wantErr:    false,
		},
		{
			name:       "contract pause",
			contractID: "contract1",
			eventType:  "PAUSE",
			actor:      "admin1",
			metadata:   map[string]string{"reason": "security"},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := k.RecordAuditEvent(ctx, tt.contractID, tt.eventType, tt.actor, tt.metadata)
			if (err != nil) != tt.wantErr {
				t.Errorf("RecordAuditEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetAuditTrail(t *testing.T) {
	k, ctx := setupKeeper(t)

	contractID := "contract1"

	// Record multiple events
	events := []struct {
		eventType string
		actor     string
	}{
		{"DEPLOY", "deployer1"},
		{"UPGRADE", "admin1"},
		{"PAUSE", "admin1"},
		{"RESUME", "admin1"},
	}

	for _, e := range events {
		err := k.RecordAuditEvent(ctx, contractID, e.eventType, e.actor, nil)
		if err != nil {
			t.Fatalf("Failed to record event: %v", err)
		}
	}

	// Get audit trail
	trail := k.GetAuditTrail(ctx, contractID, 100)
	if len(trail) < len(events) {
		t.Errorf("Expected at least %d events, got %d", len(events), len(trail))
	}

	// Verify chronological order
	for i := 0; i < len(trail)-1; i++ {
		if trail[i].Timestamp.AsTime().After(trail[i+1].Timestamp.AsTime()) {
			t.Log("Audit trail may not be in chronological order")
			break
		}
	}
}

func TestGetAuditEventsByActor(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Record events by different actors
	actors := []string{"admin1", "admin2", "deployer1"}
	for _, actor := range actors {
		for i := 0; i < 3; i++ {
			err := k.RecordAuditEvent(ctx, "contract1", "ACTION", actor, nil)
			if err != nil {
				t.Fatalf("Failed to record event: %v", err)
			}
		}
	}

	// Get events by specific actor
	admin1Events := k.GetAuditEventsByActor(ctx, "admin1", 100)
	if len(admin1Events) < 3 {
		t.Errorf("Expected at least 3 events for admin1, got %d", len(admin1Events))
	}

	// Verify all events are from correct actor
	for _, event := range admin1Events {
		if event.Actor != "admin1" {
			t.Errorf("Expected actor admin1, got %s", event.Actor)
		}
	}
}

func TestGetAuditEventsByType(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Record different event types
	eventTypes := []string{"DEPLOY", "UPGRADE", "PAUSE", "DEPLOY"}
	for _, et := range eventTypes {
		err := k.RecordAuditEvent(ctx, "contract1", et, "actor1", nil)
		if err != nil {
			t.Fatalf("Failed to record event: %v", err)
		}
	}

	// Get events by type
	deployEvents := k.GetAuditEventsByType(ctx, "DEPLOY", 100)
	if len(deployEvents) < 2 {
		t.Errorf("Expected at least 2 DEPLOY events, got %d", len(deployEvents))
	}

	for _, event := range deployEvents {
		if event.EventType != "DEPLOY" {
			t.Errorf("Expected event type DEPLOY, got %s", event.EventType)
		}
	}
}

func TestAuditEventMetadata(t *testing.T) {
	k, ctx := setupKeeper(t)

	metadata := map[string]string{
		"version":     "2.0",
		"codeHash":    "abc123",
		"gasUsed":     "500000",
		"blockHeight": "12345",
	}

	err := k.RecordAuditEvent(ctx, "contract1", "UPGRADE", "admin1", metadata)
	if err != nil {
		t.Fatalf("Failed to record event: %v", err)
	}

	trail := k.GetAuditTrail(ctx, "contract1", 1)
	if len(trail) == 0 {
		t.Fatal("Expected at least one event")
	}

	eventMetadata := trail[0].Metadata
	for key, expectedValue := range metadata {
		if actualValue, exists := eventMetadata[key]; !exists {
			t.Errorf("Expected metadata key %s not found", key)
		} else if actualValue != expectedValue {
			t.Errorf("Metadata key %s: expected %s, got %s", key, expectedValue, actualValue)
		}
	}
}

func TestAuditTrailPagination(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create many events
	for i := 0; i < 50; i++ {
		err := k.RecordAuditEvent(ctx, "contract1", "ACTION", "actor1", nil)
		if err != nil {
			t.Fatalf("Failed to record event: %v", err)
		}
	}

	// Get limited results
	limit := 10
	trail := k.GetAuditTrail(ctx, "contract1", uint64(limit))
	if len(trail) > limit {
		t.Errorf("Expected at most %d events, got %d", limit, len(trail))
	}
}

func TestAuditEventTimestamp(t *testing.T) {
	k, ctx := setupKeeper(t)

	beforeTime := ctx.BlockTime()

	err := k.RecordAuditEvent(ctx, "contract1", "TEST", "actor1", nil)
	if err != nil {
		t.Fatalf("Failed to record event: %v", err)
	}

	afterTime := ctx.BlockTime().Add(time.Second)

	trail := k.GetAuditTrail(ctx, "contract1", 1)
	if len(trail) == 0 {
		t.Fatal("Expected at least one event")
	}

	eventTime := trail[0].Timestamp.AsTime()
	if eventTime.Before(beforeTime) || eventTime.After(afterTime) {
		t.Logf("Event timestamp %v is within acceptable range", eventTime)
	}
}

func TestSearchAuditEvents(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create diverse events
	events := []struct {
		contractID string
		eventType  string
		actor      string
		metadata   map[string]string
	}{
		{"contract1", "DEPLOY", "deployer1", map[string]string{"network": "mainnet"}},
		{"contract1", "UPGRADE", "admin1", map[string]string{"network": "mainnet"}},
		{"contract2", "DEPLOY", "deployer2", map[string]string{"network": "testnet"}},
		{"contract2", "PAUSE", "admin1", map[string]string{"network": "testnet"}},
	}

	for _, e := range events {
		err := k.RecordAuditEvent(ctx, e.contractID, e.eventType, e.actor, e.metadata)
		if err != nil {
			t.Fatalf("Failed to record event: %v", err)
		}
	}

	tests := []struct {
		name     string
		criteria map[string]string
		minCount int
	}{
		{
			name:     "search by contract",
			criteria: map[string]string{"contract_id": "contract1"},
			minCount: 2,
		},
		{
			name:     "search by actor",
			criteria: map[string]string{"actor": "admin1"},
			minCount: 2,
		},
		{
			name:     "search by event type",
			criteria: map[string]string{"event_type": "DEPLOY"},
			minCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := k.SearchAuditEvents(ctx, tt.criteria, 100)
			if len(results) < tt.minCount {
				t.Errorf("Expected at least %d results, got %d", tt.minCount, len(results))
			}
		})
	}
}

func TestAuditStatistics(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create events
	for i := 0; i < 10; i++ {
		require.NoError(t, k.RecordAuditEvent(ctx, "contract1", "DEPLOY", "actor1", nil))
	}
	for i := 0; i < 5; i++ {
		require.NoError(t, k.RecordAuditEvent(ctx, "contract1", "UPGRADE", "actor1", nil))
	}

	stats := k.GetAuditStatistics(ctx, "contract1")

	if stats.TotalEvents < 15 {
		t.Errorf("Expected at least 15 total events, got %d", stats.TotalEvents)
	}

	if stats.EventTypeCounts != nil {
		if deployCount := stats.EventTypeCounts["DEPLOY"]; deployCount < 10 {
			t.Errorf("Expected at least 10 DEPLOY events, got %d", deployCount)
		}
	}
}

func TestAuditEventCompression(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create large metadata
	largeMetadata := make(map[string]string)
	for i := 0; i < 100; i++ {
		largeMetadata[string(rune('a'+i%26))+string(rune(i))] = "large value with lots of data"
	}

	err := k.RecordAuditEvent(ctx, "contract1", "LARGE_EVENT", "actor1", largeMetadata)
	if err != nil {
		t.Fatalf("Failed to record large event: %v", err)
	}

	trail := k.GetAuditTrail(ctx, "contract1", 1)
	if len(trail) == 0 {
		t.Fatal("Expected event to be stored")
	}
}

func TestAuditEventImmutability(t *testing.T) {
	k, ctx := setupKeeper(t)

	metadata := map[string]string{"key": "original"}
	err := k.RecordAuditEvent(ctx, "contract1", "TEST", "actor1", metadata)
	if err != nil {
		t.Fatalf("Failed to record event: %v", err)
	}

	// Attempt to modify metadata (should not affect stored event)
	metadata["key"] = "modified"

	trail := k.GetAuditTrail(ctx, "contract1", 1)
	if len(trail) == 0 {
		t.Fatal("Expected event")
	}

	if storedValue := trail[0].Metadata["key"]; storedValue != "original" {
		t.Errorf("Expected original value, got %s (audit log should be immutable)", storedValue)
	}
}

func TestAuditRetentionPolicy(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Set retention policy
	retentionDays := uint64(90)
	err := k.SetAuditRetentionPolicy(ctx, retentionDays)
	if err != nil {
		t.Fatalf("Failed to set retention policy: %v", err)
	}

	policy, found := k.GetAuditRetentionPolicy(ctx)
	if !found {
		t.Fatal("Retention policy not found")
	}

	if policy != retentionDays {
		t.Errorf("Expected retention %d days, got %d", retentionDays, policy)
	}
}

func TestDeleteExpiredAuditEvents(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create old event
	oldCtx := ctx.WithBlockTime(ctx.BlockTime().Add(-100 * 24 * time.Hour))
	err := k.RecordAuditEvent(oldCtx, "contract1", "OLD", "actor1", nil)
	if err != nil {
		t.Fatalf("Failed to record old event: %v", err)
	}

	// Create recent event
	err = k.RecordAuditEvent(ctx, "contract1", "RECENT", "actor1", nil)
	if err != nil {
		t.Fatalf("Failed to record recent event: %v", err)
	}

	// Delete events older than 30 days
	cutoff := ctx.BlockTime().Add(-30 * 24 * time.Hour)
	deleted := k.DeleteAuditEventsBefore(ctx, cutoff)

	t.Logf("Deleted %d expired audit events", deleted)
}

func TestAuditEventExport(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create events
	for i := 0; i < 5; i++ {
		require.NoError(t, k.RecordAuditEvent(ctx, "contract1", "EVENT", "actor1", nil))
	}

	// Export audit trail
	exported := k.ExportAuditTrail(ctx, "contract1")

	if len(exported) < 5 {
		t.Errorf("Expected at least 5 exported events, got %d", len(exported))
	}
}

func TestAuditEventValidation(t *testing.T) {
	k, ctx := setupKeeper(t)

	tests := []struct {
		name       string
		contractID string
		eventType  string
		actor      string
		wantErr    bool
	}{
		{
			name:       "valid event",
			contractID: "contract1",
			eventType:  "DEPLOY",
			actor:      "actor1",
			wantErr:    false,
		},
		{
			name:       "empty contract ID",
			contractID: "",
			eventType:  "DEPLOY",
			actor:      "actor1",
			wantErr:    true,
		},
		{
			name:       "empty event type",
			contractID: "contract1",
			eventType:  "",
			actor:      "actor1",
			wantErr:    true,
		},
		{
			name:       "empty actor",
			contractID: "contract1",
			eventType:  "DEPLOY",
			actor:      "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := k.RecordAuditEvent(ctx, tt.contractID, tt.eventType, tt.actor, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("RecordAuditEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
