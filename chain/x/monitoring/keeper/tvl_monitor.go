// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

const (
	// BasisPointsPrecision is the multiplier for basis points (10000 = 100%)
	BasisPointsPrecision = int64(10000)
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
	// Convert basis points to percentage for Prometheus display
	if k.metrics != nil {
		k.metrics.TotalTVL.Set(float64(totalTVL))
		k.metrics.TVLByModule.WithLabelValues(moduleName).Set(float64(tvl))
		k.metrics.TVLChange24h.Set(float64(monitoring.TVLChange24hBps) / float64(BasisPointsPrecision) * 100)
		k.metrics.TVLChange7d.Set(float64(monitoring.TVLChange7dBps) / float64(BasisPointsPrecision) * 100)
	}

	// Check for significant TVL changes using deterministic integer math
	// Calculate change in basis points: |newTVL - oldTVL| * 10000 / oldTVL
	if oldTotalTVL > 0 {
		var changeBps int64
		if totalTVL >= oldTotalTVL {
			changeBps = int64((totalTVL - oldTotalTVL) * uint64(BasisPointsPrecision) / oldTotalTVL)
		} else {
			changeBps = int64((oldTotalTVL - totalTVL) * uint64(BasisPointsPrecision) / oldTotalTVL)
		}
		// Convert threshold from decimal (0.2 = 20%) to basis points (2000)
		thresholdBps := int64(params.TVLChangeAlertThreshold * float64(BasisPointsPrecision))
		if changeBps >= thresholdBps {
			if params.EnableAlerts {
				if err := k.createTVLChangeAlert(ctx, oldTotalTVL, totalTVL, changeBps); err != nil {
					// Log error but don't fail the update
					return nil
				}
			}
		}
	}

	return nil
}

// calculateTVLChanges calculates TVL changes over time using deterministic integer math.
// Results are stored in basis points (10000 = 100%) to avoid floating point non-determinism.
func (k Keeper) calculateTVLChanges(monitoring *types.TVLMonitoring) {
	historyLen := len(monitoring.TVLHistory)
	if historyLen < 2 {
		return
	}

	currentTVL := monitoring.TVLHistory[historyLen-1].TVL
	currentTime := monitoring.TVLHistory[historyLen-1].Timestamp

	// Calculate 24h change in basis points
	// Use integer hours comparison for determinism (avoid floating point Hours())
	twentyFourHoursNanos := int64(24 * 60 * 60 * 1e9)
	for i := historyLen - 1; i >= 0; i-- {
		timeDiffNanos := currentTime.Sub(monitoring.TVLHistory[i].Timestamp).Nanoseconds()
		if timeDiffNanos >= twentyFourHoursNanos {
			oldTVL := monitoring.TVLHistory[i].TVL
			if oldTVL > 0 {
				// Calculate: (currentTVL - oldTVL) * 10000 / oldTVL
				// Result is in basis points where 10000 = 100%
				if currentTVL >= oldTVL {
					monitoring.TVLChange24hBps = int64((currentTVL - oldTVL) * uint64(BasisPointsPrecision) / oldTVL)
				} else {
					monitoring.TVLChange24hBps = -int64((oldTVL - currentTVL) * uint64(BasisPointsPrecision) / oldTVL)
				}
			}
			break
		}
	}

	// Calculate 7d change in basis points
	sevenDaysNanos := int64(7 * 24 * 60 * 60 * 1e9)
	for i := historyLen - 1; i >= 0; i-- {
		timeDiffNanos := currentTime.Sub(monitoring.TVLHistory[i].Timestamp).Nanoseconds()
		if timeDiffNanos >= sevenDaysNanos {
			oldTVL := monitoring.TVLHistory[i].TVL
			if oldTVL > 0 {
				// Calculate: (currentTVL - oldTVL) * 10000 / oldTVL
				// Result is in basis points where 10000 = 100%
				if currentTVL >= oldTVL {
					monitoring.TVLChange7dBps = int64((currentTVL - oldTVL) * uint64(BasisPointsPrecision) / oldTVL)
				} else {
					monitoring.TVLChange7dBps = -int64((oldTVL - currentTVL) * uint64(BasisPointsPrecision) / oldTVL)
				}
			}
			break
		}
	}
}

// updateLargestPools updates the list of largest pools by TVL.
// Uses deterministic integer basis points for percentage calculations.
func (k Keeper) updateLargestPools(monitoring *types.TVLMonitoring) {
	pools := make([]types.PoolTVL, 0, len(monitoring.TVLByModule))

	// First, collect all module names and sort them for deterministic iteration
	moduleNames := make([]string, 0, len(monitoring.TVLByModule))
	for moduleName := range monitoring.TVLByModule {
		moduleNames = append(moduleNames, moduleName)
	}
	// Sort module names for deterministic iteration order
	for i := 0; i < len(moduleNames)-1; i++ {
		for j := i + 1; j < len(moduleNames); j++ {
			if moduleNames[i] > moduleNames[j] {
				moduleNames[i], moduleNames[j] = moduleNames[j], moduleNames[i]
			}
		}
	}

	for _, moduleName := range moduleNames {
		tvl := monitoring.TVLByModule[moduleName]
		var percentageBps int64
		if monitoring.TotalTVL > 0 {
			// Calculate: tvl * 10000 / totalTVL
			// Result is in basis points where 10000 = 100%
			percentageBps = int64(tvl * uint64(BasisPointsPrecision) / monitoring.TotalTVL)
		}

		pools = append(pools, types.PoolTVL{
			PoolID:        moduleName,
			PoolName:      moduleName,
			TVL:           tvl,
			PercentageBps: percentageBps,
		})
	}

	// Sort by TVL descending, then by PoolID for determinism when TVL is equal
	for i := 0; i < len(pools)-1; i++ {
		for j := i + 1; j < len(pools); j++ {
			if pools[j].TVL > pools[i].TVL ||
				(pools[j].TVL == pools[i].TVL && pools[j].PoolID < pools[i].PoolID) {
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

// createTVLChangeAlert creates an alert for significant TVL changes.
// changeBps is the change in basis points (10000 = 100%).
func (k Keeper) createTVLChangeAlert(ctx context.Context, oldTVL, newTVL uint64, changeBps int64) error {
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

	// Convert basis points to percentage for display (changeBps / 100 = percentage)
	// e.g., 2000 bps = 20%
	changePercentDisplay := float64(changeBps) / 100.0
	thresholdPercentDisplay := params.TVLChangeAlertThreshold * 100

	alert := &types.Alert{
		ID:       k.generateID(ctx, "alert-tvl-change"),
		Type:     types.AlertTypeTVLChange,
		Severity: k.determineTVLChangeSeverity(changeBps),
		Message:  fmt.Sprintf("Significant TVL %s detected: %.2f%%", direction, changePercentDisplay),
		Details: map[string]interface{}{
			"old_tvl":           oldTVL,
			"new_tvl":           newTVL,
			"change_bps":        changeBps,
			"change_percent":    changePercentDisplay,
			"direction":         direction,
			"threshold_percent": thresholdPercentDisplay,
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

// determineTVLChangeSeverity determines severity based on TVL change in basis points.
// Thresholds: Critical >= 50% (5000 bps), High >= 30% (3000 bps), Medium >= 20% (2000 bps)
func (k Keeper) determineTVLChangeSeverity(changeBps int64) types.AlertSeverity {
	// Use absolute value for severity determination
	absChangeBps := changeBps
	if absChangeBps < 0 {
		absChangeBps = -absChangeBps
	}

	if absChangeBps >= 5000 { // 50%
		return types.SeverityCritical
	} else if absChangeBps >= 3000 { // 30%
		return types.SeverityHigh
	} else if absChangeBps >= 2000 { // 20%
		return types.SeverityMedium
	}
	return types.SeverityLow
}
