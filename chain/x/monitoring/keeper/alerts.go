// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"

	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// CreateAlert creates a new alert (KV store version)
func (k *Keeper) CreateAlert(ctx context.Context, alertType types.AlertType, severity types.AlertSeverity, message string, details map[string]interface{}) (*types.Alert, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	if !params.EnableAlerts {
		return nil, nil
	}

	alert := &types.Alert{
		ID:               k.generateID(ctx, "alert"),
		Type:             alertType,
		Severity:         severity,
		Message:          message,
		Details:          details,
		Timestamp:        sdk.UnwrapSDKContext(ctx).BlockTime(),
		Acknowledged:     false,
		Resolved:         false,
		NotificationSent: false,
	}

	if err := k.SetAlert(ctx, alert); err != nil {
		return nil, err
	}

	// Update metrics
	k.metrics.TotalAlerts.WithLabelValues(string(alertType), string(severity)).Inc()
	k.metrics.ActiveAlerts.WithLabelValues(string(alertType), string(severity)).Inc()

	return alert, nil
}

// AcknowledgeAlert marks an alert as acknowledged (KV store version)
func (k *Keeper) AcknowledgeAlert(ctx context.Context, alertID, acknowledgedBy string) error {
	alert, err := k.GetAlert(ctx, alertID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	now := sdkCtx.BlockTime()
	alert.Acknowledged = true
	alert.AcknowledgedBy = acknowledgedBy
	alert.AcknowledgedAt = &now

	return k.SetAlert(ctx, alert)
}

// ResolveAlert marks an alert as resolved (KV store version)
func (k *Keeper) ResolveAlert(ctx context.Context, alertID string) error {
	alert, err := k.GetAlert(ctx, alertID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	now := sdkCtx.BlockTime()
	alert.Resolved = true
	alert.ResolvedAt = &now

	// Update metrics
	k.metrics.ActiveAlerts.WithLabelValues(string(alert.Type), string(alert.Severity)).Dec()

	// Record resolution time
	if alert.Acknowledged && alert.AcknowledgedAt != nil {
		resolutionTime := now.Sub(*alert.AcknowledgedAt).Seconds()
		k.metrics.AlertResolutionTime.Observe(resolutionTime)
	}

	return k.SetAlert(ctx, alert)
}

// NOTE: GetAlert is implemented in keeper.go
// NOTE: GetActiveAlerts is implemented in keeper.go
// NOTE: GetAlertsBySeverity is implemented in keeper.go
// NOTE: GetAlertsByType is implemented in keeper.go
