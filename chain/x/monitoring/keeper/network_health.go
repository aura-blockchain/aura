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

// GetNetworkHealth retrieves the network health from the KV store
func (k Keeper) GetNetworkHealth(ctx context.Context) (*types.NetworkHealth, error) {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(NetworkHealthKey)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		// Return default/empty network health
		return &types.NetworkHealth{}, nil
	}

	var health types.NetworkHealth
	if err := json.Unmarshal(bz, &health); err != nil {
		return nil, err
	}

	return &health, nil
}

// UpdateNetworkHealth updates the network health status in the KV store
func (k Keeper) UpdateNetworkHealth(ctx context.Context, health *types.NetworkHealth) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	if !params.EnableNetworkHealthMonitoring {
		return nil
	}

	if health == nil {
		return fmt.Errorf("network health data cannot be nil")
	}

	// Persist deterministically to KV store
	if err := k.SetNetworkHealth(ctx, health); err != nil {
		return err
	}

	// Deterministic Prometheus metrics (based on deterministic health)
	if k.metrics != nil {
		k.metrics.BlockTime.Set(health.BlockTime)
		k.metrics.TransactionsPerSecond.Set(health.TPS)
		k.metrics.MempoolSize.Set(float64(health.MempoolSize))
		k.metrics.PeerCount.Set(float64(health.PeerCount))
		k.metrics.NetworkCongestion.Set(health.NetworkCongestion)
		k.metrics.ConsensusHealth.Set(health.ConsensusHealth)
	}

	return nil
}

// createNetworkCongestionAlert creates an alert for network congestion
func (k Keeper) createNetworkCongestionAlert(ctx context.Context, health *types.NetworkHealth) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockTime := sdkCtx.BlockTime()

	alert := &types.Alert{
		ID:       k.generateID(ctx, "alert-congestion"),
		Type:     types.AlertTypeNetworkCongestion,
		Severity: k.determineCongestionSeverity(health.NetworkCongestion),
		Message:  fmt.Sprintf("Network congestion detected: %.2f%%", health.NetworkCongestion*100),
		Details: map[string]interface{}{
			"congestion":   health.NetworkCongestion,
			"threshold":    params.CongestionThreshold,
			"mempool_size": health.MempoolSize,
			"tps":          health.TPS,
			"block_time":   health.BlockTime,
			"block_height": health.BlockHeight,
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

// determineCongestionSeverity determines the severity based on congestion level
func (k Keeper) determineCongestionSeverity(congestion float64) types.AlertSeverity {
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
// NOTE: Currently returns only the latest entry. For full historical support,
// implement time-series storage with additional KV store keys.
func (k Keeper) GetNetworkHealthHistory(ctx context.Context) ([]*types.NetworkHealth, error) {
	health, err := k.GetNetworkHealth(ctx)
	if err != nil {
		return nil, err
	}

	if health.Timestamp.IsZero() {
		return []*types.NetworkHealth{}, nil
	}

	return []*types.NetworkHealth{health}, nil
}
