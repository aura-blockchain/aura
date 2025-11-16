package alerting

import (
	"fmt"
	"sync"
	"time"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// AlertManager manages security and operational alerts
type AlertManager struct {
	alerts          map[string]*types.Alert
	alertChannels   map[types.AlertSeverity][]chan *types.Alert
	lastAlertTime   map[string]time.Time
	mu              sync.RWMutex
	cooldownPeriod  time.Duration
}

// NewAlertManager creates a new alert manager
func NewAlertManager(cooldownPeriod time.Duration) *AlertManager {
	return &AlertManager{
		alerts:          make(map[string]*types.Alert),
		alertChannels:   make(map[types.AlertSeverity][]chan *types.Alert),
		lastAlertTime:   make(map[string]time.Time),
		cooldownPeriod:  cooldownPeriod,
	}
}

// CreateAlert creates a new alert
func (am *AlertManager) CreateAlert(alertType types.AlertType, severity types.AlertSeverity, message string, details map[string]interface{}) (*types.Alert, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Check cooldown period for critical alerts
	if severity == types.SeverityCritical {
		key := fmt.Sprintf("%s:%s", alertType, severity)
		if lastTime, exists := am.lastAlertTime[key]; exists {
			if time.Since(lastTime) < am.cooldownPeriod {
				return nil, fmt.Errorf("alert in cooldown period")
			}
		}
		am.lastAlertTime[key] = time.Now()
	}

	alert := &types.Alert{
		ID:               generateAlertID(),
		Type:             alertType,
		Severity:         severity,
		Message:          message,
		Details:          details,
		Timestamp:        time.Now(),
		Acknowledged:     false,
		Resolved:         false,
		NotificationSent: false,
	}

	am.alerts[alert.ID] = alert

	// Send to alert channels
	if channels, exists := am.alertChannels[severity]; exists {
		for _, ch := range channels {
			select {
			case ch <- alert:
			default:
				// Channel full, skip
			}
		}
	}

	return alert, nil
}

// AcknowledgeAlert marks an alert as acknowledged
func (am *AlertManager) AcknowledgeAlert(alertID, acknowledgedBy string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	alert, exists := am.alerts[alertID]
	if !exists {
		return types.ErrAlertNotFound
	}

	now := time.Now()
	alert.Acknowledged = true
	alert.AcknowledgedBy = acknowledgedBy
	alert.AcknowledgedAt = &now

	return nil
}

// ResolveAlert marks an alert as resolved
func (am *AlertManager) ResolveAlert(alertID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	alert, exists := am.alerts[alertID]
	if !exists {
		return types.ErrAlertNotFound
	}

	now := time.Now()
	alert.Resolved = true
	alert.ResolvedAt = &now

	return nil
}

// GetAlert retrieves an alert by ID
func (am *AlertManager) GetAlert(alertID string) (*types.Alert, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alert, exists := am.alerts[alertID]
	if !exists {
		return nil, types.ErrAlertNotFound
	}

	return alert, nil
}

// GetActiveAlerts returns all unresolved alerts
func (am *AlertManager) GetActiveAlerts() []*types.Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var active []*types.Alert
	for _, alert := range am.alerts {
		if !alert.Resolved {
			active = append(active, alert)
		}
	}

	return active
}

// GetAlertsBySeverity returns alerts filtered by severity
func (am *AlertManager) GetAlertsBySeverity(severity types.AlertSeverity) []*types.Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var filtered []*types.Alert
	for _, alert := range am.alerts {
		if alert.Severity == severity {
			filtered = append(filtered, alert)
		}
	}

	return filtered
}

// GetAlertsByType returns alerts filtered by type
func (am *AlertManager) GetAlertsByType(alertType types.AlertType) []*types.Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var filtered []*types.Alert
	for _, alert := range am.alerts {
		if alert.Type == alertType {
			filtered = append(filtered, alert)
		}
	}

	return filtered
}

// SubscribeToAlerts subscribes to alerts of a specific severity
func (am *AlertManager) SubscribeToAlerts(severity types.AlertSeverity, bufferSize int) <-chan *types.Alert {
	am.mu.Lock()
	defer am.mu.Unlock()

	ch := make(chan *types.Alert, bufferSize)
	am.alertChannels[severity] = append(am.alertChannels[severity], ch)

	return ch
}

// GetAlertStats returns alert statistics
func (am *AlertManager) GetAlertStats() map[string]interface{} {
	am.mu.RLock()
	defer am.mu.RUnlock()

	stats := map[string]interface{}{
		"total":        len(am.alerts),
		"active":       0,
		"acknowledged": 0,
		"resolved":     0,
		"by_severity":  make(map[types.AlertSeverity]int),
		"by_type":      make(map[types.AlertType]int),
	}

	severityCounts := make(map[types.AlertSeverity]int)
	typeCounts := make(map[types.AlertType]int)

	for _, alert := range am.alerts {
		if !alert.Resolved {
			stats["active"] = stats["active"].(int) + 1
		}
		if alert.Acknowledged {
			stats["acknowledged"] = stats["acknowledged"].(int) + 1
		}
		if alert.Resolved {
			stats["resolved"] = stats["resolved"].(int) + 1
		}

		severityCounts[alert.Severity]++
		typeCounts[alert.Type]++
	}

	stats["by_severity"] = severityCounts
	stats["by_type"] = typeCounts

	return stats
}

// CleanupResolvedAlerts removes old resolved alerts
func (am *AlertManager) CleanupResolvedAlerts(retentionPeriod time.Duration) int {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Now()
	cleaned := 0

	for id, alert := range am.alerts {
		if alert.Resolved && alert.ResolvedAt != nil {
			if now.Sub(*alert.ResolvedAt) > retentionPeriod {
				delete(am.alerts, id)
				cleaned++
			}
		}
	}

	return cleaned
}

// generateAlertID generates a unique alert ID
func generateAlertID() string {
	return fmt.Sprintf("alert-%d", time.Now().UnixNano())
}
