package keeper

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

func TestQueryServer_GetNetworkHealth(t *testing.T) {
	keeper := setupTestKeeper(t)
	defer keeper.Close()

	qs := NewQueryServer(keeper)
	require.NotNil(t, qs)

	ctx := context.Background()

	// Test with nil request
	resp, err := qs.GetNetworkHealth(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Health)

	// Test with empty request
	resp, err = qs.GetNetworkHealth(ctx, &types.QueryNetworkHealthRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Health)
}

func TestQueryServer_GetValidatorUptime(t *testing.T) {
	keeper := setupTestKeeper(t)
	defer keeper.Close()

	qs := NewQueryServer(keeper)
	require.NotNil(t, qs)

	ctx := context.Background()

	// Test with nil request
	resp, err := qs.GetValidatorUptime(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, types.ErrInvalidTransaction, err)

	// Test with empty validator address
	resp, err = qs.GetValidatorUptime(ctx, &types.QueryValidatorUptimeRequest{
		ValidatorAddress: "",
	})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, types.ErrValidatorNotFound, err)

	// Test with non-existent validator
	resp, err = qs.GetValidatorUptime(ctx, &types.QueryValidatorUptimeRequest{
		ValidatorAddress: "aura1validator123",
	})
	require.Error(t, err)
	require.Nil(t, resp)

	// Add a validator uptime record
	validatorAddr := "aura1validator123"
	err = keeper.UpdateValidatorUptime(validatorAddr, "TestValidator", 100, true)
	require.NoError(t, err)

	// Test with valid validator address
	resp, err = qs.GetValidatorUptime(ctx, &types.QueryValidatorUptimeRequest{
		ValidatorAddress: validatorAddr,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Uptime)
	require.Equal(t, validatorAddr, resp.Uptime.ValidatorAddress)
}

func TestQueryServer_GetAlerts(t *testing.T) {
	keeper := setupTestKeeper(t)
	defer keeper.Close()

	qs := NewQueryServer(keeper)
	require.NotNil(t, qs)

	ctx := context.Background()

	// Test with nil request - should return all active alerts (empty initially)
	resp, err := qs.GetAlerts(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Alerts)

	// Create some test alerts
	alert1, err := keeper.CreateAlert(
		types.AlertTypeSecurityThreat,
		types.SeverityCritical,
		"Test security alert",
		map[string]interface{}{"detail": "test"},
	)
	require.NoError(t, err)
	require.NotNil(t, alert1)

	// Small sleep to ensure different timestamps for unique IDs
	time.Sleep(time.Millisecond)

	alert2, err := keeper.CreateAlert(
		types.AlertTypeAnomaly,
		types.SeverityHigh,
		"Test anomaly alert",
		map[string]interface{}{"detail": "test"},
	)
	require.NoError(t, err)
	require.NotNil(t, alert2)

	// Test getting all alerts
	resp, err = qs.GetAlerts(ctx, &types.QueryAlertsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 2, len(resp.Alerts))

	// Test filtering by severity
	resp, err = qs.GetAlerts(ctx, &types.QueryAlertsRequest{
		Severity: string(types.SeverityCritical),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, len(resp.Alerts))
	require.Equal(t, types.SeverityCritical, resp.Alerts[0].Severity)

	// Test filtering by type
	resp, err = qs.GetAlerts(ctx, &types.QueryAlertsRequest{
		Type: string(types.AlertTypeAnomaly),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, len(resp.Alerts))
	require.Equal(t, types.AlertTypeAnomaly, resp.Alerts[0].Type)

	// Test filtering by both severity and type
	resp, err = qs.GetAlerts(ctx, &types.QueryAlertsRequest{
		Severity: string(types.SeverityCritical),
		Type:     string(types.AlertTypeSecurityThreat),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, len(resp.Alerts))
	require.Equal(t, types.SeverityCritical, resp.Alerts[0].Severity)
	require.Equal(t, types.AlertTypeSecurityThreat, resp.Alerts[0].Type)
}

func TestQueryServer_GetGasPriceTracking(t *testing.T) {
	keeper := setupTestKeeper(t)
	defer keeper.Close()

	qs := NewQueryServer(keeper)
	require.NotNil(t, qs)

	ctx := context.Background()

	// Test with nil request
	resp, err := qs.GetGasPriceTracking(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Tracking)

	// Test with empty request
	resp, err = qs.GetGasPriceTracking(ctx, &types.QueryGasPriceRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Tracking)
}

func TestQueryServer_GetTVLMonitoring(t *testing.T) {
	keeper := setupTestKeeper(t)
	defer keeper.Close()

	qs := NewQueryServer(keeper)
	require.NotNil(t, qs)

	ctx := context.Background()

	// Test with nil request
	resp, err := qs.GetTVLMonitoring(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Tvl)

	// Test with empty request
	resp, err = qs.GetTVLMonitoring(ctx, &types.QueryTVLRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Tvl)
}

func TestQueryServer_ImplementsInterface(t *testing.T) {
	keeper := setupTestKeeper(t)
	defer keeper.Close()

	qs := NewQueryServer(keeper)

	// Verify that QueryServer implements the types.QueryServer interface
	var _ types.QueryServer = qs
}
