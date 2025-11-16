package keeper

import (
	"time"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// CreateAlert creates a new alert
func (k *Keeper) CreateAlert(alertType types.AlertType, severity types.AlertSeverity, message string, details map[string]interface{}) (*types.Alert, error) {
	if !k.params.EnableAlerts {
		return nil, nil
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	alert := &types.Alert{
		ID:               generateID("alert"),
		Type:             alertType,
		Severity:         severity,
		Message:          message,
		Details:          details,
		Timestamp:        time.Now(),
		Acknowledged:     false,
		Resolved:         false,
		NotificationSent: false,
	}

	k.alerts[alert.ID] = alert

	// Update metrics
	k.metrics.TotalAlerts.WithLabelValues(string(alertType), string(severity)).Inc()
	k.metrics.ActiveAlerts.WithLabelValues(string(alertType), string(severity)).Inc()

	return alert, nil
}

// AcknowledgeAlert marks an alert as acknowledged
func (k *Keeper) AcknowledgeAlert(alertID, acknowledgedBy string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	alert, exists := k.alerts[alertID]
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
func (k *Keeper) ResolveAlert(alertID string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	alert, exists := k.alerts[alertID]
	if !exists {
		return types.ErrAlertNotFound
	}

	now := time.Now()
	alert.Resolved = true
	alert.ResolvedAt = &now

	// Update metrics
	k.metrics.ActiveAlerts.WithLabelValues(string(alert.Type), string(alert.Severity)).Dec()

	// Record resolution time
	if alert.Acknowledged && alert.AcknowledgedAt != nil {
		resolutionTime := now.Sub(*alert.AcknowledgedAt).Seconds()
		k.metrics.AlertResolutionTime.Observe(resolutionTime)
	}

	return nil
}

// GetAlert retrieves an alert by ID
func (k *Keeper) GetAlert(alertID string) (*types.Alert, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	alert, exists := k.alerts[alertID]
	if !exists {
		return nil, types.ErrAlertNotFound
	}

	return alert, nil
}

// GetActiveAlerts returns all unresolved alerts
func (k *Keeper) GetActiveAlerts() []*types.Alert {
	k.mu.RLock()
	defer k.mu.RUnlock()

	var active []*types.Alert
	for _, alert := range k.alerts {
		if !alert.Resolved {
			active = append(active, alert)
		}
	}

	return active
}

// GetAlertsBySeverity returns alerts filtered by severity
func (k *Keeper) GetAlertsBySeverity(severity types.AlertSeverity) []*types.Alert {
	k.mu.RLock()
	defer k.mu.RUnlock()

	var filtered []*types.Alert
	for _, alert := range k.alerts {
		if alert.Severity == severity {
			filtered = append(filtered, alert)
		}
	}

	return filtered
}

// GetAlertsByType returns alerts filtered by type
func (k *Keeper) GetAlertsByType(alertType types.AlertType) []*types.Alert {
	k.mu.RLock()
	defer k.mu.RUnlock()

	var filtered []*types.Alert
	for _, alert := range k.alerts {
		if alert.Type == alertType {
			filtered = append(filtered, alert)
		}
	}

	return filtered
}
