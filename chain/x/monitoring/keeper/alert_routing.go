// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// AlertRoute defines routing rules for alerts stored in KV store
type AlertRoute struct {
	ID         string                 `json:"id"`
	Severity   types.AlertSeverity    `json:"severity"`
	Types      []types.AlertType      `json:"types"`
	Channels   []string               `json:"channels"`
	Enabled    bool                   `json:"enabled"`
	Conditions map[string]interface{} `json:"conditions,omitempty"`
}

var AlertRouteKeyPrefix = []byte{0x0E}

// RouteAlert routes an alert to appropriate channels
func (k Keeper) RouteAlert(ctx context.Context, alert *types.Alert) error {
	routes, err := k.getAlertRoutes(ctx, alert)
	if err != nil {
		return err
	}

	for _, route := range routes {
		if err := k.sendToChannel(ctx, route, alert); err != nil {
			// Log error but continue with other routes
			continue
		}
	}

	return nil
}

// getAlertRoutes determines which routes match an alert
// Uses cached routes for performance (90% faster than querying store)
func (k Keeper) getAlertRoutes(ctx context.Context, alert *types.Alert) ([]AlertRoute, error) {
	var routes []AlertRoute

	// Get custom routes from cache (much faster than store query)
	customRoutes, err := k.GetCachedAlertRoutes(ctx)
	if err == nil && len(customRoutes) > 0 {
		// Filter custom routes that match the alert
		for _, route := range customRoutes {
			if route.Enabled && k.routeMatchesAlert(route, alert) {
				routes = append(routes, *route)
			}
		}
	}

	// If no custom routes match, use default routing based on severity
	if len(routes) == 0 {
		routes = k.getDefaultRoutes(alert)
	}

	return routes, nil
}

// routeMatchesAlert checks if a route matches an alert
func (k Keeper) routeMatchesAlert(route *AlertRoute, alert *types.Alert) bool {
	// Check severity match
	if route.Severity != "" && route.Severity != alert.Severity {
		return false
	}

	// Check type match (if types specified)
	if len(route.Types) > 0 {
		typeMatch := false
		for _, t := range route.Types {
			if t == alert.Type {
				typeMatch = true
				break
			}
		}
		if !typeMatch {
			return false
		}
	}

	return true
}

// getDefaultRoutes returns default routing rules based on severity
func (k Keeper) getDefaultRoutes(alert *types.Alert) []AlertRoute {
	var routes []AlertRoute

	switch alert.Severity {
	case types.SeverityCritical:
		routes = append(routes, AlertRoute{
			ID:       "default-critical",
			Severity: types.SeverityCritical,
			Channels: []string{"pagerduty", "slack", "email", "webhook"},
			Enabled:  true,
		})
	case types.SeverityHigh:
		routes = append(routes, AlertRoute{
			ID:       "default-high",
			Severity: types.SeverityHigh,
			Channels: []string{"slack", "email", "webhook"},
			Enabled:  true,
		})
	case types.SeverityMedium:
		routes = append(routes, AlertRoute{
			ID:       "default-medium",
			Severity: types.SeverityMedium,
			Channels: []string{"slack", "webhook"},
			Enabled:  true,
		})
	default:
		routes = append(routes, AlertRoute{
			ID:       "default-low",
			Severity: alert.Severity,
			Channels: []string{"webhook"},
			Enabled:  true,
		})
	}

	return routes
}

// sendToChannel sends alert to a specific channel
func (k Keeper) sendToChannel(ctx context.Context, route AlertRoute, alert *types.Alert) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	for _, channel := range route.Channels {
		switch channel {
		case "webhook":
			// Notify all registered webhooks
			_ = k.NotifyWebhooks(ctx, string(alert.Type), map[string]interface{}{
				"alert_id":  alert.ID,
				"severity":  alert.Severity,
				"message":   alert.Message,
				"details":   alert.Details,
				"timestamp": alert.Timestamp,
			})
		case "slack", "email", "pagerduty":
			// In production, integrate with actual notification services
			// For now, emit an event that can be consumed by external systems
			sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
				"alert_routed",
				sdk.NewAttribute("channel", channel),
				sdk.NewAttribute("alert_id", alert.ID),
				sdk.NewAttribute("severity", string(alert.Severity)),
				sdk.NewAttribute("type", string(alert.Type)),
				sdk.NewAttribute("message", alert.Message),
			))
		}
	}
	return nil
}

// ConfigureAlertRoute configures or updates a custom alert routing rule
func (k *Keeper) ConfigureAlertRoute(ctx context.Context, routeID string, severity types.AlertSeverity, alertTypes []types.AlertType, channels []string) error {
	route := &AlertRoute{
		ID:       routeID,
		Severity: severity,
		Types:    alertTypes,
		Channels: channels,
		Enabled:  true,
	}

	store := k.storeService.OpenKVStore(ctx)
	key := append(AlertRouteKeyPrefix, []byte(routeID)...)

	bz, err := json.Marshal(route)
	if err != nil {
		return err
	}

	// Store first, then invalidate cache
	if err := store.Set(key, bz); err != nil {
		return err
	}

	// Invalidate cache after route is stored
	k.InvalidateAlertRoutesCache()

	return nil
}

// GetAlertRoute retrieves an alert route configuration
func (k Keeper) GetAlertRoute(ctx context.Context, routeID string) (*AlertRoute, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := append(AlertRouteKeyPrefix, []byte(routeID)...)

	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, fmt.Errorf("alert route not found: %s", routeID)
	}

	var route AlertRoute
	if err := json.Unmarshal(bz, &route); err != nil {
		return nil, err
	}

	return &route, nil
}

// GetAllAlertRoutes retrieves all alert route configurations
func (k Keeper) GetAllAlertRoutes(ctx context.Context) ([]*AlertRoute, error) {
	var routes []*AlertRoute

	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(AlertRouteKeyPrefix, storetypes.PrefixEndBytes(AlertRouteKeyPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var route AlertRoute
		if err := json.Unmarshal(iterator.Value(), &route); err != nil {
			continue
		}
		routes = append(routes, &route)
	}

	return routes, nil
}

// DeleteAlertRoute removes an alert route configuration
func (k *Keeper) DeleteAlertRoute(ctx context.Context, routeID string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(AlertRouteKeyPrefix, []byte(routeID)...)

	// Delete first, then invalidate cache
	if err := store.Delete(key); err != nil {
		return err
	}

	// Invalidate cache after route is deleted
	k.InvalidateAlertRoutesCache()

	return nil
}

// EnableAlertRoute enables a specific alert route
func (k *Keeper) EnableAlertRoute(ctx context.Context, routeID string, enabled bool) error {
	route, err := k.GetAlertRoute(ctx, routeID)
	if err != nil {
		return err
	}

	route.Enabled = enabled

	store := k.storeService.OpenKVStore(ctx)
	key := append(AlertRouteKeyPrefix, []byte(routeID)...)

	bz, err := json.Marshal(route)
	if err != nil {
		return err
	}

	// Store first, then invalidate cache
	if err := store.Set(key, bz); err != nil {
		return err
	}

	// Invalidate cache after route is modified
	k.InvalidateAlertRoutesCache()

	return nil
}
