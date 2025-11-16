package keeper

import (
	"github.com/aequitas/aura/chain/x/monitoring/ml"
	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// Initialize anomaly detector
func (k *Keeper) initAnomalyDetector() {
	if k.params.EnableAnomalyDetection {
		k.mu.Lock()
		defer k.mu.Unlock()
		// Anomaly detector will be created on first use
	}
}

// detectTransactionAnomaly detects anomalies in a transaction
func (k *Keeper) detectTransactionAnomaly(tx *types.TransactionMonitorData) (float64, error) {
	if !k.params.EnableAnomalyDetection {
		return 0.0, nil
	}

	// Create detector if not exists
	detector := ml.NewAnomalyDetector(k.params.AnomalyThreshold, k.params.MLModelUpdateInterval)

	detection, err := detector.DetectTransactionAnomaly(tx)
	if err != nil {
		return 0.0, err
	}

	// Store anomaly detection result
	if detection.IsAnomaly {
		k.mu.Lock()
		k.anomalies[detection.ID] = detection
		k.mu.Unlock()

		k.metrics.AnomalyDetections.WithLabelValues(string(detection.Type)).Inc()
		k.metrics.MLModelVersion.Set(1.0)
		k.metrics.MLModelAccuracy.Set(detector.GetAccuracy())
	}

	k.metrics.AnomalyScore.Observe(detection.Score)

	return detection.Score, nil
}

// DetectNetworkAnomaly detects network-level anomalies
func (k *Keeper) DetectNetworkAnomaly(health *types.NetworkHealth) (*types.AnomalyDetection, error) {
	if !k.params.EnableAnomalyDetection {
		return nil, nil
	}

	detector := ml.NewAnomalyDetector(k.params.AnomalyThreshold, k.params.MLModelUpdateInterval)
	detection, err := detector.DetectNetworkAnomaly(health)
	if err != nil {
		return nil, err
	}

	if detection.IsAnomaly {
		k.mu.Lock()
		k.anomalies[detection.ID] = detection
		k.mu.Unlock()

		k.metrics.AnomalyDetections.WithLabelValues(string(detection.Type)).Inc()

		// Create alert for network anomaly
		if k.params.EnableAlerts {
			k.createNetworkAnomalyAlert(detection)
		}
	}

	return detection, nil
}

// GetAnomalies returns all detected anomalies
func (k *Keeper) GetAnomalies() []*types.AnomalyDetection {
	k.mu.RLock()
	defer k.mu.RUnlock()

	anomalies := make([]*types.AnomalyDetection, 0, len(k.anomalies))
	for _, a := range k.anomalies {
		anomalies = append(anomalies, a)
	}

	return anomalies
}

// GetAnomaliesByType returns anomalies of a specific type
func (k *Keeper) GetAnomaliesByType(anomalyType types.AnomalyType) []*types.AnomalyDetection {
	k.mu.RLock()
	defer k.mu.RUnlock()

	var filtered []*types.AnomalyDetection
	for _, a := range k.anomalies {
		if a.Type == anomalyType {
			filtered = append(filtered, a)
		}
	}

	return filtered
}

// createAnomalyAlert creates an alert for an anomaly
func (k *Keeper) createAnomalyAlert(tx *types.TransactionMonitorData, score float64) {
	alert := &types.Alert{
		ID:       generateID("alert-anomaly"),
		Type:     types.AlertTypeAnomaly,
		Severity: k.determineAnomalySeverity(score),
		Message:  "Transaction anomaly detected",
		Details: map[string]interface{}{
			"tx_hash":       tx.TxHash,
			"anomaly_score": score,
			"threshold":     k.params.AnomalyThreshold,
			"amount":        tx.Amount,
			"gas_used":      tx.GasUsed,
		},
		Timestamp:        tx.Timestamp,
		Acknowledged:     false,
		Resolved:         false,
		NotificationSent: false,
	}

	k.mu.Lock()
	k.alerts[alert.ID] = alert
	k.mu.Unlock()

	k.metrics.TotalAlerts.WithLabelValues(string(alert.Type), string(alert.Severity)).Inc()
	k.metrics.ActiveAlerts.WithLabelValues(string(alert.Type), string(alert.Severity)).Inc()
}

// createNetworkAnomalyAlert creates an alert for a network anomaly
func (k *Keeper) createNetworkAnomalyAlert(detection *types.AnomalyDetection) {
	alert := &types.Alert{
		ID:       generateID("alert-network-anomaly"),
		Type:     types.AlertTypeAnomaly,
		Severity: k.determineAnomalySeverity(detection.Score),
		Message:  "Network anomaly detected",
		Details: map[string]interface{}{
			"detection_id":  detection.ID,
			"anomaly_score": detection.Score,
			"threshold":     detection.Threshold,
			"features":      detection.Features,
		},
		Timestamp:        detection.Timestamp,
		Acknowledged:     false,
		Resolved:         false,
		NotificationSent: false,
	}

	k.mu.Lock()
	k.alerts[alert.ID] = alert
	k.mu.Unlock()

	k.metrics.TotalAlerts.WithLabelValues(string(alert.Type), string(alert.Severity)).Inc()
	k.metrics.ActiveAlerts.WithLabelValues(string(alert.Type), string(alert.Severity)).Inc()
}

// determineAnomalySeverity determines alert severity based on anomaly score
func (k *Keeper) determineAnomalySeverity(score float64) types.AlertSeverity {
	if score >= 0.95 {
		return types.SeverityCritical
	} else if score >= 0.85 {
		return types.SeverityHigh
	} else if score >= 0.75 {
		return types.SeverityMedium
	}
	return types.SeverityLow
}
