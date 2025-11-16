package keeper

import (
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// explorerSyncWorker periodically syncs with blockchain explorer
func (k *Keeper) explorerSyncWorker() {
	defer k.wg.Done()
	ticker := time.NewTicker(k.params.ExplorerSyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-k.ctx.Done():
			return
		case <-ticker.C:
			k.syncWithExplorer()
		}
	}
}

// InitializeExplorerIntegration initializes the blockchain explorer integration
func (k *Keeper) InitializeExplorerIntegration(endpoint, apiKey string, customMetadata map[string]string) error {
	if !k.params.EnableExplorerIntegration {
		return nil
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	k.explorerIntegration = &types.ExplorerIntegration{
		Endpoint:       endpoint,
		APIKey:         apiKey,
		SyncStatus:     "initialized",
		LastSyncHeight: 0,
		LastSyncTime:   time.Now(),
		HealthStatus:   "unknown",
		CustomMetadata: customMetadata,
	}

	return nil
}

// UpdateExplorerSync updates the explorer sync status
func (k *Keeper) UpdateExplorerSync(currentHeight int64, syncStatus, healthStatus string) error {
	if !k.params.EnableExplorerIntegration {
		return nil
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	if k.explorerIntegration == nil {
		return types.ErrExplorerIntegrationFailed
	}

	k.explorerIntegration.LastSyncHeight = currentHeight
	k.explorerIntegration.LastSyncTime = time.Now()
	k.explorerIntegration.SyncStatus = syncStatus
	k.explorerIntegration.HealthStatus = healthStatus

	// Update Prometheus metrics
	k.metrics.ExplorerSyncHeight.Set(float64(currentHeight))

	return nil
}

// GetExplorerIntegration returns the current explorer integration status
func (k *Keeper) GetExplorerIntegration() (*types.ExplorerIntegration, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.explorerIntegration == nil {
		return nil, types.ErrExplorerIntegrationFailed
	}

	return k.explorerIntegration, nil
}

// syncWithExplorer performs synchronization with blockchain explorer
func (k *Keeper) syncWithExplorer() {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.explorerIntegration == nil {
		return
	}

	// In a real implementation, this would make API calls to the explorer
	// For now, we'll just update the sync time
	k.explorerIntegration.LastSyncTime = time.Now()

	// Calculate sync lag if we have network health data
	if k.networkHealth != nil && k.networkHealth.BlockHeight > 0 {
		lag := k.networkHealth.BlockHeight - k.explorerIntegration.LastSyncHeight
		k.metrics.ExplorerSyncLag.Set(float64(lag))

		// Create alert if lag is significant
		if lag > 100 && k.params.EnableAlerts {
			k.createExplorerSyncLagAlert(lag)
		}
	}
}

// RecordExplorerAPIError records an API error from the explorer
func (k *Keeper) RecordExplorerAPIError(errorMessage string) {
	k.metrics.ExplorerAPIErrors.Inc()

	// Log the error
	_ = k.LogEntry(
		types.LogLevelError,
		"explorer",
		fmt.Sprintf("Explorer API error: %s", errorMessage),
		map[string]interface{}{
			"error": errorMessage,
		},
		"",
		"",
	)
}

// GetExplorerStats returns explorer integration statistics
func (k *Keeper) GetExplorerStats() map[string]interface{} {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.explorerIntegration == nil {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	syncLag := int64(0)
	if k.networkHealth != nil && k.networkHealth.BlockHeight > 0 {
		syncLag = k.networkHealth.BlockHeight - k.explorerIntegration.LastSyncHeight
	}

	return map[string]interface{}{
		"enabled":          true,
		"endpoint":         k.explorerIntegration.Endpoint,
		"sync_status":      k.explorerIntegration.SyncStatus,
		"health_status":    k.explorerIntegration.HealthStatus,
		"last_sync_height": k.explorerIntegration.LastSyncHeight,
		"last_sync_time":   k.explorerIntegration.LastSyncTime,
		"sync_lag":         syncLag,
	}
}

// createExplorerSyncLagAlert creates an alert for explorer sync lag
func (k *Keeper) createExplorerSyncLagAlert(lag int64) {
	alert := &types.Alert{
		ID:       generateID("alert-explorer-lag"),
		Type:     types.AlertTypeSystemError,
		Severity: k.determineExplorerLagSeverity(lag),
		Message:  fmt.Sprintf("Explorer sync lag detected: %d blocks behind", lag),
		Details: map[string]interface{}{
			"sync_lag":         lag,
			"last_sync_height": k.explorerIntegration.LastSyncHeight,
			"current_height":   k.networkHealth.BlockHeight,
			"last_sync_time":   k.explorerIntegration.LastSyncTime,
		},
		Timestamp:        time.Now(),
		Acknowledged:     false,
		Resolved:         false,
		NotificationSent: false,
	}

	k.alerts[alert.ID] = alert
	k.metrics.TotalAlerts.WithLabelValues(string(alert.Type), string(alert.Severity)).Inc()
	k.metrics.ActiveAlerts.WithLabelValues(string(alert.Type), string(alert.Severity)).Inc()
}

// determineExplorerLagSeverity determines severity based on sync lag
func (k *Keeper) determineExplorerLagSeverity(lag int64) types.AlertSeverity {
	if lag >= 1000 {
		return types.SeverityCritical
	} else if lag >= 500 {
		return types.SeverityHigh
	} else if lag >= 200 {
		return types.SeverityMedium
	}
	return types.SeverityLow
}

// ProvideExplorerData provides data for blockchain explorer integration
func (k *Keeper) ProvideExplorerData() map[string]interface{} {
	k.mu.RLock()
	defer k.mu.RUnlock()

	data := map[string]interface{}{
		"network_health":    k.networkHealth,
		"validator_count":   len(k.validatorUptime),
		"active_alerts":     len(k.alerts),
		"security_events":   len(k.securityEvents),
		"recent_anomalies":  len(k.anomalies),
	}

	// Add transaction statistics
	data["transaction_stats"] = k.GetTransactionStats()

	// Add validator statistics
	data["validator_stats"] = k.GetValidatorStats()

	return data
}
