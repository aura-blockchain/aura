package keeper

import (
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// networkHealthWorker periodically updates network health metrics
func (k *Keeper) networkHealthWorker() {
	defer k.wg.Done()
	ticker := time.NewTicker(k.params.NetworkHealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-k.ctx.Done():
			return
		case <-ticker.C:
			k.updateNetworkHealth()
		}
	}
}

// UpdateNetworkHealth updates the network health status
func (k *Keeper) UpdateNetworkHealth(health *types.NetworkHealth) error {
	if !k.params.EnableNetworkHealthMonitoring {
		return nil
	}

	if health == nil {
		return fmt.Errorf("network health data cannot be nil")
	}

	k.mu.Lock()
	k.networkHealth = health
	k.mu.Unlock()

	// Update Prometheus metrics
	k.metrics.BlockTime.Set(health.BlockTime)
	k.metrics.TransactionsPerSecond.Set(health.TPS)
	k.metrics.MempoolSize.Set(float64(health.MempoolSize))
	k.metrics.PeerCount.Set(float64(health.PeerCount))
	k.metrics.NetworkCongestion.Set(health.NetworkCongestion)
	k.metrics.ConsensusHealth.Set(health.ConsensusHealth)

	// Check for congestion alerts
	if health.NetworkCongestion >= k.params.CongestionThreshold {
		if k.params.EnableAlerts {
			k.createNetworkCongestionAlert(health)
		}
	}

	// Run anomaly detection
	if k.params.EnableAnomalyDetection {
		_, _ = k.DetectNetworkAnomaly(health)
	}

	return nil
}

// GetNetworkHealth returns the current network health status
func (k *Keeper) GetNetworkHealth() *types.NetworkHealth {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.networkHealth == nil {
		return &types.NetworkHealth{}
	}

	return k.networkHealth
}

// updateNetworkHealth simulates network health updates
func (k *Keeper) updateNetworkHealth() {
	// In a real implementation, this would query the actual network
	// For now, we'll just update the timestamp
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.networkHealth != nil {
		k.networkHealth.Timestamp = time.Now()
	}
}

// createNetworkCongestionAlert creates an alert for network congestion
func (k *Keeper) createNetworkCongestionAlert(health *types.NetworkHealth) {
	k.mu.Lock()
	defer k.mu.Unlock()

	alert := &types.Alert{
		ID:       generateID("alert-congestion"),
		Type:     types.AlertTypeNetworkCongestion,
		Severity: k.determineCongestionSeverity(health.NetworkCongestion),
		Message:  fmt.Sprintf("Network congestion detected: %.2f%%", health.NetworkCongestion*100),
		Details: map[string]interface{}{
			"congestion":     health.NetworkCongestion,
			"threshold":      k.params.CongestionThreshold,
			"mempool_size":   health.MempoolSize,
			"tps":            health.TPS,
			"block_time":     health.BlockTime,
			"block_height":   health.BlockHeight,
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

// determineCongestionSeverity determines the severity based on congestion level
func (k *Keeper) determineCongestionSeverity(congestion float64) types.AlertSeverity {
	if congestion >= 0.95 {
		return types.SeverityCritical
	} else if congestion >= 0.85 {
		return types.SeverityHigh
	} else if congestion >= 0.75 {
		return types.SeverityMedium
	}
	return types.SeverityLow
}

// GetNetworkHealthHistory returns historical network health data
func (k *Keeper) GetNetworkHealthHistory(duration time.Duration) []types.NetworkHealth {
	// In a real implementation, this would return historical data
	// For now, return current state
	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.networkHealth == nil {
		return []types.NetworkHealth{}
	}

	return []types.NetworkHealth{*k.networkHealth}
}
