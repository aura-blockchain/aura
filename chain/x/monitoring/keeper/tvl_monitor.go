package keeper

import (
	"fmt"
	"math"
	"time"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// tvlMonitoringWorker periodically monitors TVL
func (k *Keeper) tvlMonitoringWorker() {
	defer k.wg.Done()
	ticker := time.NewTicker(k.params.TVLCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-k.ctx.Done():
			return
		case <-ticker.C:
			k.updateTVLMetrics()
		}
	}
}

// UpdateTVL updates the Total Value Locked metrics
func (k *Keeper) UpdateTVL(moduleName string, tvl uint64) error {
	if !k.params.EnableTVLMonitoring {
		return nil
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	// Update module TVL
	k.tvlMonitoring.TVLByModule[moduleName] = tvl

	// Calculate total TVL
	var totalTVL uint64
	for _, modTVL := range k.tvlMonitoring.TVLByModule {
		totalTVL += modTVL
	}

	oldTotalTVL := k.tvlMonitoring.TotalTVL
	k.tvlMonitoring.TotalTVL = totalTVL
	k.tvlMonitoring.Timestamp = time.Now()

	// Add to history
	tvlPoint := types.TVLPoint{
		Timestamp: time.Now(),
		TVL:       totalTVL,
	}
	k.tvlMonitoring.TVLHistory = append(k.tvlMonitoring.TVLHistory, tvlPoint)

	// Keep only recent history
	if len(k.tvlMonitoring.TVLHistory) > k.params.TVLHistorySize {
		k.tvlMonitoring.TVLHistory = k.tvlMonitoring.TVLHistory[1:]
	}

	// Calculate TVL changes
	k.calculateTVLChanges()

	// Update Prometheus metrics
	k.metrics.TotalTVL.Set(float64(totalTVL))
	k.metrics.TVLByModule.WithLabelValues(moduleName).Set(float64(tvl))
	k.metrics.TVLChange24h.Set(k.tvlMonitoring.TVLChange24h)
	k.metrics.TVLChange7d.Set(k.tvlMonitoring.TVLChange7d)

	// Check for significant TVL changes
	if oldTotalTVL > 0 {
		change := math.Abs(float64(totalTVL)-float64(oldTotalTVL)) / float64(oldTotalTVL)
		if change >= k.params.TVLChangeAlertThreshold {
			if k.params.EnableAlerts {
				k.createTVLChangeAlert(oldTotalTVL, totalTVL, change)
			}
		}
	}

	// Update largest pools
	k.updateLargestPools()

	return nil
}

// calculateTVLChanges calculates TVL changes over time
func (k *Keeper) calculateTVLChanges() {
	historyLen := len(k.tvlMonitoring.TVLHistory)
	if historyLen < 2 {
		return
	}

	currentTVL := k.tvlMonitoring.TVLHistory[historyLen-1].TVL
	currentTime := k.tvlMonitoring.TVLHistory[historyLen-1].Timestamp

	// 24h change
	for i := historyLen - 1; i >= 0; i-- {
		if currentTime.Sub(k.tvlMonitoring.TVLHistory[i].Timestamp) >= 24*time.Hour {
			oldTVL := k.tvlMonitoring.TVLHistory[i].TVL
			if oldTVL > 0 {
				k.tvlMonitoring.TVLChange24h = (float64(currentTVL) - float64(oldTVL)) / float64(oldTVL) * 100
			}
			break
		}
	}

	// 7d change
	for i := historyLen - 1; i >= 0; i-- {
		if currentTime.Sub(k.tvlMonitoring.TVLHistory[i].Timestamp) >= 7*24*time.Hour {
			oldTVL := k.tvlMonitoring.TVLHistory[i].TVL
			if oldTVL > 0 {
				k.tvlMonitoring.TVLChange7d = (float64(currentTVL) - float64(oldTVL)) / float64(oldTVL) * 100
			}
			break
		}
	}
}

// updateLargestPools updates the list of largest pools by TVL
func (k *Keeper) updateLargestPools() {
	pools := make([]types.PoolTVL, 0, len(k.tvlMonitoring.TVLByModule))

	for moduleName, tvl := range k.tvlMonitoring.TVLByModule {
		percentage := 0.0
		if k.tvlMonitoring.TotalTVL > 0 {
			percentage = float64(tvl) / float64(k.tvlMonitoring.TotalTVL) * 100
		}

		pools = append(pools, types.PoolTVL{
			PoolID:     moduleName,
			PoolName:   moduleName,
			TVL:        tvl,
			Percentage: percentage,
		})
	}

	// Sort by TVL descending
	for i := 0; i < len(pools)-1; i++ {
		for j := i + 1; j < len(pools); j++ {
			if pools[j].TVL > pools[i].TVL {
				pools[i], pools[j] = pools[j], pools[i]
			}
		}
	}

	// Keep top 10
	if len(pools) > 10 {
		pools = pools[:10]
	}

	k.tvlMonitoring.LargestPools = pools
}

// GetTVLMonitoring returns current TVL monitoring data
func (k *Keeper) GetTVLMonitoring() *types.TVLMonitoring {
	k.mu.RLock()
	defer k.mu.RUnlock()

	return k.tvlMonitoring
}

// GetTVLByModule returns TVL for a specific module
func (k *Keeper) GetTVLByModule(moduleName string) (uint64, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	tvl, exists := k.tvlMonitoring.TVLByModule[moduleName]
	if !exists {
		return 0, fmt.Errorf("module not found: %s", moduleName)
	}

	return tvl, nil
}

// updateTVLMetrics updates TVL-related Prometheus metrics
func (k *Keeper) updateTVLMetrics() {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.tvlMonitoring != nil {
		k.metrics.TotalTVL.Set(float64(k.tvlMonitoring.TotalTVL))
		k.metrics.TVLChange24h.Set(k.tvlMonitoring.TVLChange24h)
		k.metrics.TVLChange7d.Set(k.tvlMonitoring.TVLChange7d)

		for moduleName, tvl := range k.tvlMonitoring.TVLByModule {
			k.metrics.TVLByModule.WithLabelValues(moduleName).Set(float64(tvl))
		}
	}
}

// createTVLChangeAlert creates an alert for significant TVL changes
func (k *Keeper) createTVLChangeAlert(oldTVL, newTVL uint64, changePercent float64) {
	direction := "increase"
	if newTVL < oldTVL {
		direction = "decrease"
	}

	alert := &types.Alert{
		ID:       generateID("alert-tvl-change"),
		Type:     types.AlertTypeTVLChange,
		Severity: k.determineTVLChangeSeverity(changePercent),
		Message:  fmt.Sprintf("Significant TVL %s detected: %.2f%%", direction, changePercent*100),
		Details: map[string]interface{}{
			"old_tvl":        oldTVL,
			"new_tvl":        newTVL,
			"change_percent": changePercent * 100,
			"direction":      direction,
			"threshold":      k.params.TVLChangeAlertThreshold * 100,
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

// determineTVLChangeSeverity determines severity based on TVL change
func (k *Keeper) determineTVLChangeSeverity(changePercent float64) types.AlertSeverity {
	if changePercent >= 0.5 {
		return types.SeverityCritical
	} else if changePercent >= 0.3 {
		return types.SeverityHigh
	} else if changePercent >= 0.2 {
		return types.SeverityMedium
	}
	return types.SeverityLow
}
