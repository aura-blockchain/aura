package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/monitoring/types"
	monitoringpb "github.com/aequitas/aura/proto/aura/monitoring/v1beta1"
)

func TestMsgServer_AcknowledgeAlert(t *testing.T) {
	testKeeper := SetupTestKeeperWithContext(t)
	defer testKeeper.Close()

	ms := NewMsgServer(testKeeper.Keeper)
	require.NotNil(t, ms)

	ctx := testKeeper.Ctx

	// Test with nil message
	resp, err := ms.AcknowledgeAlert(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, types.ErrInvalidTransaction, err)

	// Test with empty alert ID
	resp, err = ms.AcknowledgeAlert(ctx, &monitoringpb.MsgAcknowledgeAlert{
		AlertId:        "",
		AcknowledgedBy: "user123",
	})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, types.ErrAlertNotFound, err)

	// Test with empty acknowledged by
	resp, err = ms.AcknowledgeAlert(ctx, &monitoringpb.MsgAcknowledgeAlert{
		AlertId:        "alert123",
		AcknowledgedBy: "",
	})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, types.ErrInvalidTransaction, err)

	// Test with non-existent alert
	resp, err = ms.AcknowledgeAlert(ctx, &monitoringpb.MsgAcknowledgeAlert{
		AlertId:        "alert123",
		AcknowledgedBy: "user123",
	})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, types.ErrAlertNotFound, err)

	// Create a test alert
	alert, err := testKeeper.CreateAlert(
		ctx,
		types.AlertTypeSecurityThreat,
		types.SeverityCritical,
		"Test alert",
		map[string]interface{}{"detail": "test"},
	)
	require.NoError(t, err)
	require.NotNil(t, alert)
	require.False(t, alert.Acknowledged)

	// Test acknowledging the alert
	resp, err = ms.AcknowledgeAlert(ctx, &monitoringpb.MsgAcknowledgeAlert{
		AlertId:        alert.ID,
		AcknowledgedBy: "user123",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.True(t, resp.Success)

	// Verify the alert was acknowledged
	acknowledgedAlert, err := testKeeper.GetAlert(ctx, alert.ID)
	require.NoError(t, err)
	require.NotNil(t, acknowledgedAlert)
	require.True(t, acknowledgedAlert.Acknowledged)
	require.Equal(t, "user123", acknowledgedAlert.AcknowledgedBy)
	require.NotNil(t, acknowledgedAlert.AcknowledgedAt)
}

func TestMsgServer_ResolveAlert(t *testing.T) {
	testKeeper := SetupTestKeeperWithContext(t)
	defer testKeeper.Close()

	ms := NewMsgServer(testKeeper.Keeper)
	require.NotNil(t, ms)

	ctx := testKeeper.Ctx

	// Test with nil message
	resp, err := ms.ResolveAlert(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, types.ErrInvalidTransaction, err)

	// Test with empty alert ID
	resp, err = ms.ResolveAlert(ctx, &monitoringpb.MsgResolveAlert{
		AlertId: "",
	})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, types.ErrAlertNotFound, err)

	// Test with non-existent alert
	resp, err = ms.ResolveAlert(ctx, &monitoringpb.MsgResolveAlert{
		AlertId: "alert123",
	})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, types.ErrAlertNotFound, err)

	// Create a test alert
	alert, err := testKeeper.CreateAlert(
		ctx,
		types.AlertTypeSecurityThreat,
		types.SeverityCritical,
		"Test alert",
		map[string]interface{}{"detail": "test"},
	)
	require.NoError(t, err)
	require.NotNil(t, alert)
	require.False(t, alert.Resolved)

	// Test resolving the alert
	resp, err = ms.ResolveAlert(ctx, &monitoringpb.MsgResolveAlert{
		AlertId: alert.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.True(t, resp.Success)

	// Verify the alert was resolved
	resolvedAlert, err := testKeeper.GetAlert(ctx, alert.ID)
	require.NoError(t, err)
	require.NotNil(t, resolvedAlert)
	require.True(t, resolvedAlert.Resolved)
	require.NotNil(t, resolvedAlert.ResolvedAt)
}

func TestMsgServer_AcknowledgeAndResolve(t *testing.T) {
	testKeeper := SetupTestKeeperWithContext(t)
	defer testKeeper.Close()

	ms := NewMsgServer(testKeeper.Keeper)
	require.NotNil(t, ms)

	ctx := testKeeper.Ctx

	// Create a test alert
	alert, err := testKeeper.CreateAlert(
		ctx,
		types.AlertTypeAnomaly,
		types.SeverityHigh,
		"Test workflow alert",
		map[string]interface{}{"detail": "test"},
	)
	require.NoError(t, err)
	require.NotNil(t, alert)

	// Acknowledge the alert
	ackResp, err := ms.AcknowledgeAlert(ctx, &monitoringpb.MsgAcknowledgeAlert{
		AlertId:        alert.ID,
		AcknowledgedBy: "admin",
	})
	require.NoError(t, err)
	require.NotNil(t, ackResp)
	require.True(t, ackResp.Success)

	// Resolve the alert
	resolveResp, err := ms.ResolveAlert(ctx, &monitoringpb.MsgResolveAlert{
		AlertId: alert.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, resolveResp)
	require.True(t, resolveResp.Success)

	// Verify the final state
	finalAlert, err := testKeeper.GetAlert(ctx, alert.ID)
	require.NoError(t, err)
	require.NotNil(t, finalAlert)
	require.True(t, finalAlert.Acknowledged)
	require.True(t, finalAlert.Resolved)
	require.Equal(t, "admin", finalAlert.AcknowledgedBy)
	require.NotNil(t, finalAlert.AcknowledgedAt)
	require.NotNil(t, finalAlert.ResolvedAt)
}

func TestMsgServer_ImplementsInterface(t *testing.T) {
	testKeeper := SetupTestKeeperWithContext(t)
	defer testKeeper.Close()

	ms := NewMsgServer(testKeeper.Keeper)

	// Verify that MsgServer implements the monitoringpb.MsgServer interface
	var _ monitoringpb.MsgServer = ms
}
