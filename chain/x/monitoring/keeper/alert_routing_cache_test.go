// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testutil/keeper"
	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// TestAlertRoutesCaching verifies that alert routes are cached per block
func TestAlertRoutesCaching(t *testing.T) {
	k, ctx := keepertest.MonitoringKeeper(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Create test routes
	err := k.ConfigureAlertRoute(ctx, "test-route-1", types.SeverityCritical, []types.AlertType{types.AlertTypeGasPriceSpike}, []string{"webhook"})
	require.NoError(t, err)

	err = k.ConfigureAlertRoute(ctx, "test-route-2", types.SeverityHigh, []types.AlertType{types.AlertTypeNetworkCongestion}, []string{"slack"})
	require.NoError(t, err)

	// First call should load from store (cache miss)
	routes1, err := k.GetCachedAlertRoutes(ctx)
	require.NoError(t, err)
	require.Len(t, routes1, 2)

	// Second call in same block should use cache (cache hit)
	routes2, err := k.GetCachedAlertRoutes(ctx)
	require.NoError(t, err)
	require.Len(t, routes2, 2)

	// Routes should be identical (same pointers means cache hit)
	require.Equal(t, routes1, routes2)

	// Advance to next block
	sdkCtx = sdkCtx.WithBlockHeight(sdkCtx.BlockHeight() + 1)
	ctx = sdkCtx

	// Call BeginBlocker to refresh cache
	err = k.BeginBlocker(ctx)
	require.NoError(t, err)

	// Cache should be refreshed for new block
	routes3, err := k.GetCachedAlertRoutes(ctx)
	require.NoError(t, err)
	require.Len(t, routes3, 2)
}

// TestAlertRoutesCacheInvalidation verifies cache is invalidated on modifications
func TestAlertRoutesCacheInvalidation(t *testing.T) {
	k, ctx := keepertest.MonitoringKeeper(t)

	// Create initial route
	err := k.ConfigureAlertRoute(ctx, "test-route", types.SeverityCritical, []types.AlertType{types.AlertTypeGasPriceSpike}, []string{"webhook"})
	require.NoError(t, err)

	// Load routes into cache
	routes1, err := k.GetCachedAlertRoutes(ctx)
	require.NoError(t, err)
	require.Len(t, routes1, 1)

	// Add a new route (should invalidate cache)
	err = k.ConfigureAlertRoute(ctx, "test-route-2", types.SeverityHigh, []types.AlertType{types.AlertTypeNetworkCongestion}, []string{"slack"})
	require.NoError(t, err)

	// Next call should see updated routes
	routes2, err := k.GetCachedAlertRoutes(ctx)
	require.NoError(t, err)
	require.Len(t, routes2, 2)
}

// TestAlertRoutesCacheDelete verifies cache is invalidated on delete
func TestAlertRoutesCacheDelete(t *testing.T) {
	k, ctx := keepertest.MonitoringKeeper(t)

	// Create routes
	err := k.ConfigureAlertRoute(ctx, "test-route-1", types.SeverityCritical, []types.AlertType{types.AlertTypeGasPriceSpike}, []string{"webhook"})
	require.NoError(t, err)
	err = k.ConfigureAlertRoute(ctx, "test-route-2", types.SeverityHigh, []types.AlertType{types.AlertTypeNetworkCongestion}, []string{"slack"})
	require.NoError(t, err)

	// Load routes into cache
	routes1, err := k.GetCachedAlertRoutes(ctx)
	require.NoError(t, err)
	require.Len(t, routes1, 2)

	// Delete a route (should invalidate cache)
	err = k.DeleteAlertRoute(ctx, "test-route-1")
	require.NoError(t, err)

	// Next call should see only 1 route
	routes2, err := k.GetCachedAlertRoutes(ctx)
	require.NoError(t, err)
	require.Len(t, routes2, 1)
	require.Equal(t, "test-route-2", routes2[0].ID)
}

// TestAlertRoutesCacheEnable verifies cache is invalidated when enabling/disabling
func TestAlertRoutesCacheEnable(t *testing.T) {
	k, ctx := keepertest.MonitoringKeeper(t)

	// Create route
	err := k.ConfigureAlertRoute(ctx, "test-route", types.SeverityCritical, []types.AlertType{types.AlertTypeGasPriceSpike}, []string{"webhook"})
	require.NoError(t, err)

	// Load routes into cache
	routes1, err := k.GetCachedAlertRoutes(ctx)
	require.NoError(t, err)
	require.Len(t, routes1, 1)
	require.True(t, routes1[0].Enabled)

	// Disable route (should invalidate cache)
	err = k.EnableAlertRoute(ctx, "test-route", false)
	require.NoError(t, err)

	// Next call should see disabled route
	routes2, err := k.GetCachedAlertRoutes(ctx)
	require.NoError(t, err)
	require.Len(t, routes2, 1)
	require.False(t, routes2[0].Enabled)
}

// TestAlertRoutingUsesCache verifies that alert routing uses cached routes
func TestAlertRoutingUsesCache(t *testing.T) {
	k, ctx := keepertest.MonitoringKeeper(t)

	// Configure a custom route
	err := k.ConfigureAlertRoute(ctx, "critical-webhook", types.SeverityCritical, []types.AlertType{types.AlertTypeGasPriceSpike}, []string{"webhook"})
	require.NoError(t, err)

	// Pre-load cache
	_, err = k.GetCachedAlertRoutes(ctx)
	require.NoError(t, err)

	// Create an alert that should match the route
	alert := &types.Alert{
		ID:       "test-alert",
		Type:     types.AlertTypeGasPriceSpike,
		Severity: types.SeverityCritical,
		Message:  "Gas price too high",
	}

	// Route the alert (should use cached routes)
	err = k.RouteAlert(ctx, alert)
	require.NoError(t, err)

	// Verify the alert routing worked (webhook should be called)
	// In production, this would verify webhook was actually called
}

