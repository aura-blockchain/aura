package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/monitoring/types"
	monitoringpb "github.com/aequitas/aura/proto/aura/monitoring/v1beta1"
)

func TestQueryServer_GetNetworkHealth(t *testing.T) {
	testKeeper := SetupTestKeeperWithContext(t)
	defer testKeeper.Close()

	qs := NewQueryServer(testKeeper.Keeper)
	require.NotNil(t, qs)

	ctx := testKeeper.Ctx

	// Test with nil request
	resp, err := qs.GetNetworkHealth(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Health)

	// Test with empty request
	resp, err = qs.GetNetworkHealth(ctx, &monitoringpb.QueryNetworkHealthRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Health)
}

func TestQueryServer_GetValidatorUptime(t *testing.T) {
	testKeeper := SetupTestKeeperWithContext(t)
	defer testKeeper.Close()

	qs := NewQueryServer(testKeeper.Keeper)
	require.NotNil(t, qs)

	ctx := testKeeper.Ctx

	// Test with nil request
	resp, err := qs.GetValidatorUptime(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, types.ErrInvalidTransaction, err)

	// Test with empty validator address
	resp, err = qs.GetValidatorUptime(ctx, &monitoringpb.QueryValidatorUptimeRequest{
		ValidatorAddress: "",
	})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, types.ErrValidatorNotFound, err)

	// Test with non-existent validator
	resp, err = qs.GetValidatorUptime(ctx, &monitoringpb.QueryValidatorUptimeRequest{
		ValidatorAddress: "aura1validator123",
	})
	require.Error(t, err)
	require.Nil(t, resp)

	// Add a validator uptime record
	validatorAddr := "aura1validator123"
	err = testKeeper.UpdateValidatorUptime(ctx, validatorAddr, "TestValidator", 100, true)
	require.NoError(t, err)

	// Test with valid validator address
	resp, err = qs.GetValidatorUptime(ctx, &monitoringpb.QueryValidatorUptimeRequest{
		ValidatorAddress: validatorAddr,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Uptime)
	require.Equal(t, validatorAddr, resp.Uptime.ValidatorAddress)
}

func TestQueryServer_GetAlerts(t *testing.T) {
	testKeeper := SetupTestKeeperWithContext(t)
	defer testKeeper.Close()

	qs := NewQueryServer(testKeeper.Keeper)
	require.NotNil(t, qs)

	ctx := testKeeper.Ctx

	// Test with nil request - should return all active alerts (empty initially)
	resp, err := qs.GetAlerts(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Alerts)

	// Create some test alerts
	alert1, err := testKeeper.CreateAlert(
		ctx,
		types.AlertTypeSecurityThreat,
		types.SeverityCritical,
		"Test security alert",
		map[string]interface{}{"detail": "test"},
	)
	require.NoError(t, err)
	require.NotNil(t, alert1)

	// Small sleep to ensure different timestamps for unique IDs
	time.Sleep(time.Millisecond)

	alert2, err := testKeeper.CreateAlert(
		ctx,
		types.AlertTypeAnomaly,
		types.SeverityHigh,
		"Test anomaly alert",
		map[string]interface{}{"detail": "test"},
	)
	require.NoError(t, err)
	require.NotNil(t, alert2)

	// Test getting all alerts
	resp, err = qs.GetAlerts(ctx, &monitoringpb.QueryAlertsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 2, len(resp.Alerts))

	// Test filtering by severity
	resp, err = qs.GetAlerts(ctx, &monitoringpb.QueryAlertsRequest{
		Severity: string(types.SeverityCritical),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, len(resp.Alerts))

	// Test filtering by type
	resp, err = qs.GetAlerts(ctx, &monitoringpb.QueryAlertsRequest{
		Type: string(types.AlertTypeAnomaly),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, len(resp.Alerts))

	// Test filtering by both severity and type
	resp, err = qs.GetAlerts(ctx, &monitoringpb.QueryAlertsRequest{
		Severity: string(types.SeverityCritical),
		Type:     string(types.AlertTypeSecurityThreat),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, len(resp.Alerts))
}

func TestQueryServer_GetGasPriceTracking(t *testing.T) {
	testKeeper := SetupTestKeeperWithContext(t)
	defer testKeeper.Close()

	qs := NewQueryServer(testKeeper.Keeper)
	require.NotNil(t, qs)

	ctx := testKeeper.Ctx

	// Test with nil request
	resp, err := qs.GetGasPriceTracking(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Tracking)

	// Test with empty request
	resp, err = qs.GetGasPriceTracking(ctx, &monitoringpb.QueryGasPriceRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Tracking)
}

func TestQueryServer_GetTVLMonitoring(t *testing.T) {
	testKeeper := SetupTestKeeperWithContext(t)
	defer testKeeper.Close()

	qs := NewQueryServer(testKeeper.Keeper)
	require.NotNil(t, qs)

	ctx := testKeeper.Ctx

	// Test with nil request
	resp, err := qs.GetTVLMonitoring(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Tvl)

	// Test with empty request
	resp, err = qs.GetTVLMonitoring(ctx, &monitoringpb.QueryTVLRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Tvl)
}

func TestQueryServer_ImplementsInterface(t *testing.T) {
	testKeeper := SetupTestKeeperWithContext(t)
	defer testKeeper.Close()

	qs := NewQueryServer(testKeeper.Keeper)

	// Verify that QueryServer implements the monitoringpb.QueryServer interface
	var _ monitoringpb.QueryServer = qs
}
