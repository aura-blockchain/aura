package keeper

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	monitoringtypes "github.com/aequitas/aura/chain/x/monitoring/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestKeeper(t *testing.T) (Keeper, sdk.Context) {
	return NewTestKeeper(t)
}

func TestNewKeeper(t *testing.T) {
	keeper, _ := setupTestKeeper(t)

	require.NotNil(t, keeper)
	require.NotNil(t, keeper.metrics)
}

func TestParams(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	// Test default params
	params, err := keeper.GetParams(ctx)
	require.NoError(t, err)
	require.NotNil(t, params)

	// Test set params
	newParams := monitoringtypes.DefaultParams()
	newParams.EnableTransactionMonitoring = true
	newParams.EnableAlerts = true
	err = keeper.SetParams(ctx, newParams)
	require.NoError(t, err)

	// Verify params were set
	retrievedParams, err := keeper.GetParams(ctx)
	require.NoError(t, err)
	assert.Equal(t, newParams.EnableTransactionMonitoring, retrievedParams.EnableTransactionMonitoring)
	assert.Equal(t, newParams.EnableAlerts, retrievedParams.EnableAlerts)
}

func TestAlertManagement(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	params := monitoringtypes.DefaultParams()
	params.EnableAlerts = true
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Create an alert
	alert, err := keeper.CreateAlert(
		ctx,
		monitoringtypes.AlertTypeSecurityThreat,
		monitoringtypes.SeverityHigh,
		"Test security threat",
		map[string]interface{}{"test": "data"},
	)
	require.NoError(t, err)
	require.NotNil(t, alert)

	// Verify alert is active
	activeAlerts, err := keeper.GetActiveAlerts(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(activeAlerts))

	// Acknowledge alert
	err = keeper.AcknowledgeAlert(ctx, alert.ID, "admin@aura.network")
	require.NoError(t, err)

	// Verify acknowledgment
	retrievedAlert, err := keeper.GetAlert(ctx, alert.ID)
	require.NoError(t, err)
	assert.True(t, retrievedAlert.Acknowledged)
	assert.Equal(t, "admin@aura.network", retrievedAlert.AcknowledgedBy)

	// Resolve alert
	err = keeper.ResolveAlert(ctx, alert.ID)
	require.NoError(t, err)

	// Verify alert is resolved
	retrievedAlert, err = keeper.GetAlert(ctx, alert.ID)
	require.NoError(t, err)
	assert.True(t, retrievedAlert.Resolved)

	// Verify no active alerts
	activeAlerts, err = keeper.GetActiveAlerts(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, len(activeAlerts))
}

func TestValidatorUptime(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	params := monitoringtypes.DefaultParams()
	params.EnableValidatorMonitoring = true
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	validatorAddr := "auravaloper1test"
	moniker := "TestValidator"

	// Record some signed blocks
	for i := int64(0); i < 100; i++ {
		err := keeper.UpdateValidatorUptime(ctx, validatorAddr, moniker, i, true)
		require.NoError(t, err)
	}

	// Record some missed blocks
	for i := int64(100); i < 105; i++ {
		err := keeper.UpdateValidatorUptime(ctx, validatorAddr, moniker, i, false)
		require.NoError(t, err)
	}

	// Verify uptime
	uptime, err := keeper.GetValidatorUptime(ctx, validatorAddr)
	require.NoError(t, err)
	assert.Equal(t, int64(105), uptime.TotalBlocks)
	assert.Equal(t, int64(100), uptime.SignedBlocks)
	assert.Equal(t, int64(5), uptime.MissedBlocks)
	assert.InDelta(t, 95.24, uptime.UptimePercentage, 0.1)
}

func TestNetworkHealthUpdate(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	params := monitoringtypes.DefaultParams()
	params.EnableNetworkHealthMonitoring = true
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	health := &monitoringtypes.NetworkHealth{
		Timestamp:         time.Now(),
		BlockHeight:       1000,
		BlockTime:         6.5,
		TPS:               100.5,
		ActiveValidators:  50,
		TotalValidators:   100,
		PeerCount:         25,
		MempoolSize:       150,
		NetworkCongestion: 0.3,
		ConsensusHealth:   0.95,
	}

	err = keeper.UpdateNetworkHealth(ctx, health)
	require.NoError(t, err)

	retrievedHealth, err := keeper.GetNetworkHealth(ctx)
	require.NoError(t, err)
	assert.Equal(t, health.BlockHeight, retrievedHealth.BlockHeight)
	assert.Equal(t, health.TPS, retrievedHealth.TPS)
}

func TestGasPriceTracking(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	params := monitoringtypes.DefaultParams()
	params.EnableGasPriceTracking = true
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	prices := []uint64{100, 110, 105, 108, 112}
	for _, price := range prices {
		err := keeper.TrackGasPrice(ctx, price)
		require.NoError(t, err)
	}

	tracking, err := keeper.GetGasPriceTracking(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(112), tracking.CurrentPrice)
	assert.Equal(t, len(prices), len(tracking.PriceHistory))
	assert.Greater(t, tracking.AveragePrice, uint64(0))
}

func TestTVLMonitoring(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	params := monitoringtypes.DefaultParams()
	params.EnableTVLMonitoring = true
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Update TVL for different modules
	err = keeper.UpdateTVL(ctx, "dex", 1000000)
	require.NoError(t, err)

	err = keeper.UpdateTVL(ctx, "staking", 5000000)
	require.NoError(t, err)

	tvl, err := keeper.GetTVLMonitoring(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(6000000), tvl.TotalTVL)
	assert.Equal(t, uint64(1000000), tvl.TVLByModule["dex"])
	assert.Equal(t, uint64(5000000), tvl.TVLByModule["staking"])
}

func TestLogAggregation_SKIP(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	params := monitoringtypes.DefaultParams()
	params.EnableLogAggregation = true
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Log some entries
	err = keeper.LogEntry(
		ctx,
		monitoringtypes.LogLevelInfo,
		"test-module",
		"Test info message",
		map[string]interface{}{"key": "value"},
		"trace-123",
		"span-456",
	)
	require.NoError(t, err)

	err = keeper.LogEntry(
		ctx,
		monitoringtypes.LogLevelError,
		"test-module",
		"Test error message",
		map[string]interface{}{"error": "test"},
		"trace-123",
		"span-789",
	)
	require.NoError(t, err)

	// Retrieve logs
	logs, err := keeper.GetLogs(ctx, "test-module", 10)
	require.NoError(t, err)
	assert.Equal(t, 2, len(logs))

	// Get error logs
	errorLogs, err := keeper.GetErrorLogs(ctx, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, len(errorLogs))
	assert.Equal(t, monitoringtypes.LogLevelError, errorLogs[0].Level)

	// Get logs by trace ID
	tracedLogs, err := keeper.GetLogsByTraceID(ctx, "trace-123")
	require.NoError(t, err)
	assert.Equal(t, 2, len(tracedLogs))
}

func TestFailedTransactionPattern_SKIP(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	params := monitoringtypes.DefaultParams()
	params.EnableFailedTxAnalysis = true
	params.FailedTxPatternThreshold = 3
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Record failed transactions with same reason
	for i := 0; i < 5; i++ {
		tx := &monitoringtypes.TransactionMonitorData{
			TxHash:      fmt.Sprintf("tx_%d", time.Now().UnixNano()),
			Sender:      "aura1sender",
			Receiver:    "aura1receiver",
			Amount:      1000,
			GasUsed:     50000,
			GasPrice:    100,
			Status:      "failed",
			Timestamp:   time.Now(),
			BlockHeight: int64(1000 + i),
			Module:      "bank",
		}

		err := keeper.RecordFailedTransaction(ctx, tx, "insufficient_funds")
		require.NoError(t, err)
	}

	// Verify pattern was detected
	patterns, err := keeper.GetFailedTransactionPatterns(ctx)
	require.NoError(t, err)
	assert.Greater(t, len(patterns), 0)

	found := false
	for _, pattern := range patterns {
		if pattern.FailureReason == "insufficient_funds" {
			assert.Equal(t, int64(5), pattern.Occurrences)
			found = true
			break
		}
	}
	assert.True(t, found)
}
