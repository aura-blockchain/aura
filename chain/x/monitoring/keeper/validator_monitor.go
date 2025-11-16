package keeper

import (
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// validatorMonitoringWorker monitors validator uptime
func (k *Keeper) validatorMonitoringWorker() {
	defer k.wg.Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-k.ctx.Done():
			return
		case <-ticker.C:
			k.updateValidatorMetrics()
		}
	}
}

// UpdateValidatorUptime updates uptime data for a validator
func (k *Keeper) UpdateValidatorUptime(validatorAddr, moniker string, blockHeight int64, signed bool) error {
	if !k.params.EnableValidatorMonitoring {
		return nil
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	uptime, exists := k.validatorUptime[validatorAddr]
	if !exists {
		uptime = &types.ValidatorUptime{
			ValidatorAddress: validatorAddr,
			Moniker:          moniker,
			TotalBlocks:      0,
			SignedBlocks:     0,
			MissedBlocks:     0,
			LastSeen:         time.Now(),
			Status:           "active",
			Jailed:           false,
		}
		k.validatorUptime[validatorAddr] = uptime
	}

	uptime.TotalBlocks++
	uptime.LastSeen = time.Now()

	if signed {
		uptime.SignedBlocks++
		uptime.ConsecutiveMisses = 0
	} else {
		uptime.MissedBlocks++
		uptime.ConsecutiveMisses++

		// Update metrics
		k.metrics.ValidatorMissedBlocks.WithLabelValues(validatorAddr, moniker).Inc()

		// Check if validator should be jailed
		if uptime.ConsecutiveMisses >= k.params.MaxConsecutiveMisses {
			uptime.Jailed = true
			uptime.Status = "jailed"

			// Create alert
			if k.params.EnableAlerts {
				k.createValidatorDownAlert(uptime)
			}
		}
	}

	// Calculate uptime percentage
	if uptime.TotalBlocks > 0 {
		uptime.UptimePercentage = float64(uptime.SignedBlocks) / float64(uptime.TotalBlocks) * 100
	}

	// Update Prometheus metrics
	k.metrics.ValidatorUptime.WithLabelValues(validatorAddr, moniker).Set(uptime.UptimePercentage)

	return nil
}

// GetValidatorUptime retrieves uptime data for a validator
func (k *Keeper) GetValidatorUptime(validatorAddr string) (*types.ValidatorUptime, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	uptime, exists := k.validatorUptime[validatorAddr]
	if !exists {
		return nil, types.ErrValidatorNotFound
	}

	return uptime, nil
}

// GetAllValidatorUptimes returns uptime data for all validators
func (k *Keeper) GetAllValidatorUptimes() []*types.ValidatorUptime {
	k.mu.RLock()
	defer k.mu.RUnlock()

	uptimes := make([]*types.ValidatorUptime, 0, len(k.validatorUptime))
	for _, uptime := range k.validatorUptime {
		uptimes = append(uptimes, uptime)
	}

	return uptimes
}

// GetJailedValidators returns all jailed validators
func (k *Keeper) GetJailedValidators() []*types.ValidatorUptime {
	k.mu.RLock()
	defer k.mu.RUnlock()

	var jailed []*types.ValidatorUptime
	for _, uptime := range k.validatorUptime {
		if uptime.Jailed {
			jailed = append(jailed, uptime)
		}
	}

	return jailed
}

// updateValidatorMetrics updates validator-related Prometheus metrics
func (k *Keeper) updateValidatorMetrics() {
	k.mu.RLock()
	defer k.mu.RUnlock()

	activeCount := 0
	jailedCount := 0

	for _, uptime := range k.validatorUptime {
		if uptime.Status == "active" && !uptime.Jailed {
			activeCount++
		}
		if uptime.Jailed {
			jailedCount++
		}
	}

	k.metrics.ActiveValidators.Set(float64(activeCount))
	k.metrics.JailedValidators.Set(float64(jailedCount))
}

// createValidatorDownAlert creates an alert when a validator goes down
func (k *Keeper) createValidatorDownAlert(uptime *types.ValidatorUptime) {
	alert := &types.Alert{
		ID:       generateID("alert-validator-down"),
		Type:     types.AlertTypeValidatorDown,
		Severity: types.SeverityHigh,
		Message:  fmt.Sprintf("Validator %s is down after %d consecutive misses", uptime.Moniker, uptime.ConsecutiveMisses),
		Details: map[string]interface{}{
			"validator_address":  uptime.ValidatorAddress,
			"moniker":            uptime.Moniker,
			"consecutive_misses": uptime.ConsecutiveMisses,
			"uptime_percentage":  uptime.UptimePercentage,
			"total_blocks":       uptime.TotalBlocks,
			"missed_blocks":      uptime.MissedBlocks,
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

// GetValidatorStats returns validator statistics
func (k *Keeper) GetValidatorStats() map[string]interface{} {
	k.mu.RLock()
	defer k.mu.RUnlock()

	totalValidators := len(k.validatorUptime)
	activeValidators := 0
	jailedValidators := 0
	var avgUptime float64

	for _, uptime := range k.validatorUptime {
		if uptime.Status == "active" && !uptime.Jailed {
			activeValidators++
		}
		if uptime.Jailed {
			jailedValidators++
		}
		avgUptime += uptime.UptimePercentage
	}

	if totalValidators > 0 {
		avgUptime /= float64(totalValidators)
	}

	return map[string]interface{}{
		"total_validators":  totalValidators,
		"active_validators": activeValidators,
		"jailed_validators": jailedValidators,
		"average_uptime":    avgUptime,
	}
}
