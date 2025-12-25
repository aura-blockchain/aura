// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package alerting

import (
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/monitoring/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAlertManager(t *testing.T) {
	am := NewAlertManager(5 * time.Minute)
	require.NotNil(t, am)
	assert.Equal(t, 5*time.Minute, am.cooldownPeriod)
}

func TestCreateAlert(t *testing.T) {
	am := NewAlertManager(5 * time.Minute)

	alert, err := am.CreateAlert(
		types.AlertTypeSecurityThreat,
		types.SeverityHigh,
		"Test alert",
		map[string]interface{}{"key": "value"},
	)

	require.NoError(t, err)
	require.NotNil(t, alert)
	assert.Equal(t, types.AlertTypeSecurityThreat, alert.Type)
	assert.Equal(t, types.SeverityHigh, alert.Severity)
	assert.False(t, alert.Acknowledged)
	assert.False(t, alert.Resolved)
}

func TestAlertCooldown(t *testing.T) {
	am := NewAlertManager(1 * time.Second)

	// Create first critical alert
	alert1, err := am.CreateAlert(
		types.AlertTypeSecurityThreat,
		types.SeverityCritical,
		"Critical alert 1",
		nil,
	)
	require.NoError(t, err)
	require.NotNil(t, alert1)

	// Try to create another immediately (should fail due to cooldown)
	alert2, err := am.CreateAlert(
		types.AlertTypeSecurityThreat,
		types.SeverityCritical,
		"Critical alert 2",
		nil,
	)
	assert.Error(t, err)
	assert.Nil(t, alert2)

	// Wait for cooldown period
	time.Sleep(1100 * time.Millisecond)

	// Now should succeed
	alert3, err := am.CreateAlert(
		types.AlertTypeSecurityThreat,
		types.SeverityCritical,
		"Critical alert 3",
		nil,
	)
	require.NoError(t, err)
	require.NotNil(t, alert3)
}

func TestAcknowledgeAlert(t *testing.T) {
	am := NewAlertManager(5 * time.Minute)

	alert, err := am.CreateAlert(
		types.AlertTypeAnomaly,
		types.SeverityMedium,
		"Test alert",
		nil,
	)
	require.NoError(t, err)

	err = am.AcknowledgeAlert(alert.ID, "operator@aura.network")
	require.NoError(t, err)

	retrievedAlert, _ := am.GetAlert(alert.ID)
	assert.True(t, retrievedAlert.Acknowledged)
	assert.Equal(t, "operator@aura.network", retrievedAlert.AcknowledgedBy)
	assert.NotNil(t, retrievedAlert.AcknowledgedAt)
}

func TestResolveAlert(t *testing.T) {
	am := NewAlertManager(5 * time.Minute)

	alert, err := am.CreateAlert(
		types.AlertTypeValidatorDown,
		types.SeverityHigh,
		"Test alert",
		nil,
	)
	require.NoError(t, err)

	// Resolve alert
	err = am.ResolveAlert(alert.ID)
	require.NoError(t, err)

	retrievedAlert, _ := am.GetAlert(alert.ID)
	assert.True(t, retrievedAlert.Resolved)
	assert.NotNil(t, retrievedAlert.ResolvedAt)
}

func TestGetActiveAlerts(t *testing.T) {
	am := NewAlertManager(5 * time.Minute)

	// Create some alerts
	_, err := am.CreateAlert(types.AlertTypeAnomaly, types.SeverityLow, "Alert 1", nil)
	require.NoError(t, err)
	_, err = am.CreateAlert(types.AlertTypeSecurityThreat, types.SeverityHigh, "Alert 2", nil)
	require.NoError(t, err)
	alert3, err := am.CreateAlert(types.AlertTypeNetworkCongestion, types.SeverityMedium, "Alert 3", nil)
	require.NoError(t, err)

	// Resolve one
	require.NoError(t, am.ResolveAlert(alert3.ID))

	// Get active alerts
	active := am.GetActiveAlerts()
	assert.Equal(t, 2, len(active))
}

func TestGetAlertsBySeverity(t *testing.T) {
	am := NewAlertManager(5 * time.Minute)

	_, err := am.CreateAlert(types.AlertTypeAnomaly, types.SeverityLow, "Alert 1", nil)
	require.NoError(t, err)
	_, err = am.CreateAlert(types.AlertTypeSecurityThreat, types.SeverityHigh, "Alert 2", nil)
	require.NoError(t, err)
	_, err = am.CreateAlert(types.AlertTypeNetworkCongestion, types.SeverityHigh, "Alert 3", nil)
	require.NoError(t, err)

	highSeverity := am.GetAlertsBySeverity(types.SeverityHigh)
	assert.Equal(t, 2, len(highSeverity))

	lowSeverity := am.GetAlertsBySeverity(types.SeverityLow)
	assert.Equal(t, 1, len(lowSeverity))
}

func TestGetAlertsByType(t *testing.T) {
	am := NewAlertManager(5 * time.Minute)

	_, err := am.CreateAlert(types.AlertTypeAnomaly, types.SeverityLow, "Alert 1", nil)
	require.NoError(t, err)
	_, err = am.CreateAlert(types.AlertTypeAnomaly, types.SeverityHigh, "Alert 2", nil)
	require.NoError(t, err)
	_, err = am.CreateAlert(types.AlertTypeSecurityThreat, types.SeverityMedium, "Alert 3", nil)
	require.NoError(t, err)

	anomalyAlerts := am.GetAlertsByType(types.AlertTypeAnomaly)
	assert.Equal(t, 2, len(anomalyAlerts))
}

func TestCleanupResolvedAlerts(t *testing.T) {
	am := NewAlertManager(5 * time.Minute)

	alert1, err := am.CreateAlert(types.AlertTypeAnomaly, types.SeverityLow, "Alert 1", nil)
	require.NoError(t, err)
	alert2, err := am.CreateAlert(types.AlertTypeSecurityThreat, types.SeverityHigh, "Alert 2", nil)
	require.NoError(t, err)

	// Resolve alerts
	require.NoError(t, am.ResolveAlert(alert1.ID))
	require.NoError(t, am.ResolveAlert(alert2.ID))

	// Initial count
	assert.Equal(t, 2, len(am.GetAlertsByType(types.AlertTypeAnomaly))+
		len(am.GetAlertsByType(types.AlertTypeSecurityThreat)))

	// Cleanup with very short retention (alerts should be cleaned)
	time.Sleep(100 * time.Millisecond)
	cleaned := am.CleanupResolvedAlerts(50 * time.Millisecond)
	assert.Equal(t, 2, cleaned)
}

func TestGetAlertStats(t *testing.T) {
	am := NewAlertManager(5 * time.Minute)

	_, err := am.CreateAlert(types.AlertTypeAnomaly, types.SeverityLow, "Alert 1", nil)
	require.NoError(t, err)
	alert2, err := am.CreateAlert(types.AlertTypeSecurityThreat, types.SeverityHigh, "Alert 2", nil)
	require.NoError(t, err)
	alert3, err := am.CreateAlert(types.AlertTypeNetworkCongestion, types.SeverityMedium, "Alert 3", nil)
	require.NoError(t, err)

	require.NoError(t, am.AcknowledgeAlert(alert2.ID, "operator"))
	require.NoError(t, am.ResolveAlert(alert3.ID))

	stats := am.GetAlertStats()
	assert.Equal(t, 3, stats["total"])
	assert.Equal(t, 2, stats["active"])
	assert.Equal(t, 1, stats["acknowledged"])
	assert.Equal(t, 1, stats["resolved"])
}
