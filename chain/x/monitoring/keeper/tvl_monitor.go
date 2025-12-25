// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// GetTVLMonitoring retrieves TVL monitoring data from the KV store
func (k Keeper) GetTVLMonitoring(ctx context.Context) (*types.TVLMonitoring, error) {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(TVLMonitoringKey)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		// Return empty TVL monitoring data
		return &types.TVLMonitoring{
			TVLByModule:  make(map[string]uint64),
			TVLHistory:   []types.TVLPoint{},
			LargestPools: []types.PoolTVL{},
		}, nil
	}

	var monitoring types.TVLMonitoring
	if err := json.Unmarshal(bz, &monitoring); err != nil {
		return nil, err
	}

	return &monitoring, nil
}

// UpdateTVL updates the Total Value Locked metrics for a specific module
func (k Keeper) UpdateTVL(ctx context.Context, moduleName string, tvl uint64) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	if !params.EnableTVLMonitoring {
		return nil
	}

	// Get current TVL monitoring data
	monitoring, err := k.GetTVLMonitoring(ctx)
	if err != nil {
		return err
	}

	if monitoring.TVLByModule == nil {
		monitoring.TVLByModule = make(map[string]uint64)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockTime := sdkCtx.BlockTime()

	// Update module TVL
	monitoring.TVLByModule[moduleName] = tvl

	// Calculate total TVL
	var totalTVL uint64
	for _, modTVL := range monitoring.TVLByModule {
		totalTVL += modTVL
	}

	oldTotalTVL := monitoring.TotalTVL
	monitoring.TotalTVL = totalTVL
	monitoring.Timestamp = blockTime

	// Add to history
	tvlPoint := types.TVLPoint{
		Timestamp: blockTime,
		TVL:       totalTVL,
	}
	monitoring.TVLHistory = append(monitoring.TVLHistory, tvlPoint)

	// Keep only recent history
	if len(monitoring.TVLHistory) > params.TVLHistorySize {
		monitoring.TVLHistory = monitoring.TVLHistory[len(monitoring.TVLHistory)-params.TVLHistorySize:]
	}

	// Calculate TVL changes
	k.calculateTVLChanges(monitoring)

	// Update largest pools
	k.updateLargestPools(monitoring)

	// Persist to KV store
	if err := k.SetTVLMonitoring(ctx, monitoring); err != nil {
		return err
	}

	// Update Prometheus metrics (non-consensus, observability only)
	if k.metrics != nil {
		k.metrics.TotalTVL.Set(float64(totalTVL))
		k.metrics.TVLByModule.WithLabelValues(moduleName).Set(float64(tvl))
		k.metrics.TVLChange24h.Set(monitoring.TVLChange24h)
		k.metrics.TVLChange7d.Set(monitoring.TVLChange7d)
	}

	// Check for significant TVL changes
	if oldTotalTVL > 0 {
		change := math.Abs(float64(totalTVL)-float64(oldTotalTVL)) / float64(oldTotalTVL)
		if change >= params.TVLChangeAlertThreshold {
			if params.EnableAlerts {
				if err := k.createTVLChangeAlert(ctx, oldTotalTVL, totalTVL, change); err != nil {
					// Log error but don't fail the update
					return nil
				}
			}
		}
	}

	return nil
}

// calculateTVLChanges calculates TVL changes over time
func (k Keeper) calculateTVLChanges(monitoring *types.TVLMonitoring) {
	historyLen := len(monitoring.TVLHistory)
	if historyLen < 2 {
		return
	}

	currentTVL := monitoring.TVLHistory[historyLen-1].TVL
	currentTime := monitoring.TVLHistory[historyLen-1].Timestamp

	// Calculate 24h change
	for i := historyLen - 1; i >= 0; i-- {
		timeDiff := currentTime.Sub(monitoring.TVLHistory[i].Timestamp)
		if timeDiff.Hours() >= 24 {
			oldTVL := monitoring.TVLHistory[i].TVL
			if oldTVL > 0 {
				monitoring.TVLChange24h = (float64(currentTVL) - float64(oldTVL)) / float64(oldTVL) * 100
			}
			break
		}
	}

	// Calculate 7d change
	for i := historyLen - 1; i >= 0; i-- {
		timeDiff := currentTime.Sub(monitoring.TVLHistory[i].Timestamp)
		if timeDiff.Hours() >= 168 { // 7 days = 168 hours
			oldTVL := monitoring.TVLHistory[i].TVL
			if oldTVL > 0 {
				monitoring.TVLChange7d = (float64(currentTVL) - float64(oldTVL)) / float64(oldTVL) * 100
			}
			break
		}
	}
}

// updateLargestPools updates the list of largest pools by TVL
func (k Keeper) updateLargestPools(monitoring *types.TVLMonitoring) {
	pools := make([]types.PoolTVL, 0, len(monitoring.TVLByModule))

	for moduleName, tvl := range monitoring.TVLByModule {
		percentage := 0.0
		if monitoring.TotalTVL > 0 {
			percentage = float64(tvl) / float64(monitoring.TotalTVL) * 100
		}

		pools = append(pools, types.PoolTVL{
			PoolID:     moduleName,
			PoolName:   moduleName,
			TVL:        tvl,
			Percentage: percentage,
		})
	}

	// Sort by TVL descending (bubble sort for simplicity and determinism)
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

	monitoring.LargestPools = pools
}

// GetTVLByModule returns TVL for a specific module
func (k Keeper) GetTVLByModule(ctx context.Context, moduleName string) (uint64, error) {
	monitoring, err := k.GetTVLMonitoring(ctx)
	if err != nil {
		return 0, err
	}

	tvl, exists := monitoring.TVLByModule[moduleName]
	if !exists {
		return 0, fmt.Errorf("module not found: %s", moduleName)
	}

	return tvl, nil
}

// createTVLChangeAlert creates an alert for significant TVL changes
func (k Keeper) createTVLChangeAlert(ctx context.Context, oldTVL, newTVL uint64, changePercent float64) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockTime := sdkCtx.BlockTime()

	direction := "increase"
	if newTVL < oldTVL {
		direction = "decrease"
	}

	alert := &types.Alert{
		ID:       k.generateID(ctx, "alert-tvl-change"),
		Type:     types.AlertTypeTVLChange,
		Severity: k.determineTVLChangeSeverity(changePercent),
		Message:  fmt.Sprintf("Significant TVL %s detected: %.2f%%", direction, changePercent*100),
		Details: map[string]interface{}{
			"old_tvl":        oldTVL,
			"new_tvl":        newTVL,
			"change_percent": changePercent * 100,
			"direction":      direction,
			"threshold":      params.TVLChangeAlertThreshold * 100,
		},
		Timestamp:        blockTime,
		Acknowledged:     false,
		Resolved:         false,
		NotificationSent: false,
	}

	// Store alert in KV store
	if err := k.SetAlert(ctx, alert); err != nil {
		return err
	}

	// Update metrics
	if k.metrics != nil {
		k.metrics.TotalAlerts.WithLabelValues(string(alert.Type), string(alert.Severity)).Inc()
		k.metrics.ActiveAlerts.WithLabelValues(string(alert.Type), string(alert.Severity)).Inc()
	}

	return nil
}

// determineTVLChangeSeverity determines severity based on TVL change
func (k Keeper) determineTVLChangeSeverity(changePercent float64) types.AlertSeverity {
	if changePercent >= 0.5 {
		return types.SeverityCritical
	} else if changePercent >= 0.3 {
		return types.SeverityHigh
	} else if changePercent >= 0.2 {
		return types.SeverityMedium
	}
	return types.SeverityLow
}
