package siem

import (
	"fmt"
	"sync"
	"time"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// SIEMManager manages Security Information and Event Management
type SIEMManager struct {
	events          map[string]*types.SecurityEvent
	eventChannels   []chan *types.SecurityEvent
	mu              sync.RWMutex
	threatThreshold int
}

// NewSIEMManager creates a new SIEM manager
func NewSIEMManager(threatThreshold int) *SIEMManager {
	return &SIEMManager{
		events:          make(map[string]*types.SecurityEvent),
		eventChannels:   make([]chan *types.SecurityEvent, 0),
		threatThreshold: threatThreshold,
	}
}

// RecordSecurityEvent records a security event
func (sm *SIEMManager) RecordSecurityEvent(
	eventType types.SecurityEventType,
	severity types.AlertSeverity,
	source string,
	destination string,
	description string,
	rawData map[string]interface{},
	indicators []string,
	threatLevel int,
) (*types.SecurityEvent, error) {

	sm.mu.Lock()
	defer sm.mu.Unlock()

	event := &types.SecurityEvent{
		ID:          generateEventID(),
		EventType:   eventType,
		Severity:    severity,
		Source:      source,
		Destination: destination,
		Description: description,
		RawData:     rawData,
		Timestamp:   time.Now(),
		Indicators:  indicators,
		ThreatLevel: threatLevel,
		Mitigated:   false,
	}

	sm.events[event.ID] = event

	// Send to subscribers
	for _, ch := range sm.eventChannels {
		select {
		case ch <- event:
		default:
			// Channel full, skip
		}
	}

	return event, nil
}

// MitigateSecurityEvent marks a security event as mitigated
func (sm *SIEMManager) MitigateSecurityEvent(eventID string, mitigationSteps []string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	event, exists := sm.events[eventID]
	if !exists {
		return fmt.Errorf("security event not found: %s", eventID)
	}

	event.Mitigated = true
	event.MitigationSteps = mitigationSteps

	return nil
}

// GetSecurityEvent retrieves a security event by ID
func (sm *SIEMManager) GetSecurityEvent(eventID string) (*types.SecurityEvent, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	event, exists := sm.events[eventID]
	if !exists {
		return nil, fmt.Errorf("security event not found: %s", eventID)
	}

	return event, nil
}

// GetSecurityEvents returns all security events
func (sm *SIEMManager) GetSecurityEvents() []*types.SecurityEvent {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	events := make([]*types.SecurityEvent, 0, len(sm.events))
	for _, event := range sm.events {
		events = append(events, event)
	}

	return events
}

// GetUnmitigatedEvents returns all unmitigated security events
func (sm *SIEMManager) GetUnmitigatedEvents() []*types.SecurityEvent {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var unmitigated []*types.SecurityEvent
	for _, event := range sm.events {
		if !event.Mitigated {
			unmitigated = append(unmitigated, event)
		}
	}

	return unmitigated
}

// GetHighThreatEvents returns events above the threat threshold
func (sm *SIEMManager) GetHighThreatEvents() []*types.SecurityEvent {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var highThreat []*types.SecurityEvent
	for _, event := range sm.events {
		if event.ThreatLevel >= sm.threatThreshold {
			highThreat = append(highThreat, event)
		}
	}

	return highThreat
}

// GetEventsByType returns events filtered by type
func (sm *SIEMManager) GetEventsByType(eventType types.SecurityEventType) []*types.SecurityEvent {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var filtered []*types.SecurityEvent
	for _, event := range sm.events {
		if event.EventType == eventType {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// SubscribeToEvents subscribes to security events
func (sm *SIEMManager) SubscribeToEvents(bufferSize int) <-chan *types.SecurityEvent {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	ch := make(chan *types.SecurityEvent, bufferSize)
	sm.eventChannels = append(sm.eventChannels, ch)

	return ch
}

// GetSecurityStats returns security event statistics
func (sm *SIEMManager) GetSecurityStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	stats := map[string]interface{}{
		"total_events":      len(sm.events),
		"mitigated_events":  0,
		"high_threat_events": 0,
		"by_type":           make(map[types.SecurityEventType]int),
		"by_severity":       make(map[types.AlertSeverity]int),
	}

	typeCounts := make(map[types.SecurityEventType]int)
	severityCounts := make(map[types.AlertSeverity]int)

	for _, event := range sm.events {
		if event.Mitigated {
			stats["mitigated_events"] = stats["mitigated_events"].(int) + 1
		}
		if event.ThreatLevel >= sm.threatThreshold {
			stats["high_threat_events"] = stats["high_threat_events"].(int) + 1
		}

		typeCounts[event.EventType]++
		severityCounts[event.Severity]++
	}

	stats["by_type"] = typeCounts
	stats["by_severity"] = severityCounts

	return stats
}

// CleanupOldEvents removes old security events
func (sm *SIEMManager) CleanupOldEvents(retentionPeriod time.Duration) int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	cleaned := 0

	for id, event := range sm.events {
		if event.Mitigated && now.Sub(event.Timestamp) > retentionPeriod {
			delete(sm.events, id)
			cleaned++
		}
	}

	return cleaned
}

// AnalyzeThreatTrends analyzes threat trends over time
func (sm *SIEMManager) AnalyzeThreatTrends(duration time.Duration) map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	now := time.Now()
	cutoff := now.Add(-duration)

	recentEvents := 0
	avgThreatLevel := 0.0
	threatLevelSum := 0

	for _, event := range sm.events {
		if event.Timestamp.After(cutoff) {
			recentEvents++
			threatLevelSum += event.ThreatLevel
		}
	}

	if recentEvents > 0 {
		avgThreatLevel = float64(threatLevelSum) / float64(recentEvents)
	}

	return map[string]interface{}{
		"period":             duration.String(),
		"recent_events":      recentEvents,
		"avg_threat_level":   avgThreatLevel,
		"total_threat_score": threatLevelSum,
	}
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return fmt.Sprintf("siem-event-%d", time.Now().UnixNano())
}
