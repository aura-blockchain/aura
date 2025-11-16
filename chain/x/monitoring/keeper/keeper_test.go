package keeper

import (
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/monitoring/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewKeeper(t *testing.T) {
	params := types.DefaultParams()
	keeper := NewKeeper(&params)

	require.NotNil(t, keeper)
	require.NotNil(t, keeper.metrics)
	assert.Equal(t, params, keeper.GetParams())

	// Cleanup
	keeper.Close()
}

func TestMonitorTransaction(t *testing.T) {
	params := types.DefaultParams()
	params.EnableTransactionMonitoring = true
	keeper := NewKeeper(&params)
	defer keeper.Close()

	tx := &types.TransactionMonitorData{
		TxHash:      "test-hash-123",
		Sender:      "aura1sender",
		Receiver:    "aura1receiver",
		Amount:      1000,
		GasUsed:     50000,
		GasPrice:    100,
		Status:      "success",
		Timestamp:   time.Now(),
		BlockHeight: 1000,
		Module:      "bank",
	}

	err := keeper.MonitorTransaction(tx)
	require.NoError(t, err)

	// Verify transaction was stored
	retrievedTx, err := keeper.GetTransaction(tx.TxHash)
	require.NoError(t, err)
	assert.Equal(t, tx.TxHash, retrievedTx.TxHash)
	assert.Equal(t, tx.Amount, retrievedTx.Amount)
}

func TestLargeTransactionAlert(t *testing.T) {
	params := types.DefaultParams()
	params.EnableTransactionMonitoring = true
	params.EnableAlerts = true
	params.LargeTransactionThreshold = 500000
	keeper := NewKeeper(&params)
	defer keeper.Close()

	tx := &types.TransactionMonitorData{
		TxHash:      "large-tx-123",
		Sender:      "aura1sender",
		Receiver:    "aura1receiver",
		Amount:      1000000, // Above threshold
		GasUsed:     50000,
		GasPrice:    100,
		Status:      "success",
		Timestamp:   time.Now(),
		BlockHeight: 1000,
		Module:      "bank",
	}

	err := keeper.MonitorTransaction(tx)
	require.NoError(t, err)

	// Verify large transaction was flagged
	assert.True(t, tx.IsLargeTransfer)

	// Verify alert was created
	alerts := keeper.GetActiveAlerts()
	found := false
	for _, alert := range alerts {
		if alert.Type == types.AlertTypeLargeTransaction {
			found = true
			break
		}
	}
	assert.True(t, found, "Large transaction alert should be created")
}

func TestValidatorUptime(t *testing.T) {
	params := types.DefaultParams()
	params.EnableValidatorMonitoring = true
	keeper := NewKeeper(&params)
	defer keeper.Close()

	validatorAddr := "auravaloper1test"
	moniker := "TestValidator"

	// Record some signed blocks
	for i := int64(0); i < 100; i++ {
		err := keeper.UpdateValidatorUptime(validatorAddr, moniker, i, true)
		require.NoError(t, err)
	}

	// Record some missed blocks
	for i := int64(100); i < 105; i++ {
		err := keeper.UpdateValidatorUptime(validatorAddr, moniker, i, false)
		require.NoError(t, err)
	}

	// Verify uptime
	uptime, err := keeper.GetValidatorUptime(validatorAddr)
	require.NoError(t, err)
	assert.Equal(t, int64(105), uptime.TotalBlocks)
	assert.Equal(t, int64(100), uptime.SignedBlocks)
	assert.Equal(t, int64(5), uptime.MissedBlocks)
	assert.InDelta(t, 95.24, uptime.UptimePercentage, 0.1)
}

func TestNetworkHealthUpdate(t *testing.T) {
	params := types.DefaultParams()
	params.EnableNetworkHealthMonitoring = true
	keeper := NewKeeper(&params)
	defer keeper.Close()

	health := &types.NetworkHealth{
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

	err := keeper.UpdateNetworkHealth(health)
	require.NoError(t, err)

	retrievedHealth := keeper.GetNetworkHealth()
	assert.Equal(t, health.BlockHeight, retrievedHealth.BlockHeight)
	assert.Equal(t, health.TPS, retrievedHealth.TPS)
}

func TestGasPriceTracking(t *testing.T) {
	params := types.DefaultParams()
	params.EnableGasPriceTracking = true
	keeper := NewKeeper(&params)
	defer keeper.Close()

	prices := []uint64{100, 110, 105, 108, 112}
	for _, price := range prices {
		err := keeper.TrackGasPrice(price)
		require.NoError(t, err)
	}

	tracking := keeper.GetGasPriceTracking()
	assert.Equal(t, uint64(112), tracking.CurrentPrice)
	assert.Equal(t, len(prices), len(tracking.PriceHistory))
	assert.Greater(t, tracking.AveragePrice, uint64(0))
}

func TestTVLMonitoring(t *testing.T) {
	params := types.DefaultParams()
	params.EnableTVLMonitoring = true
	keeper := NewKeeper(&params)
	defer keeper.Close()

	// Update TVL for different modules
	err := keeper.UpdateTVL("dex", 1000000)
	require.NoError(t, err)

	err = keeper.UpdateTVL("staking", 5000000)
	require.NoError(t, err)

	tvl := keeper.GetTVLMonitoring()
	assert.Equal(t, uint64(6000000), tvl.TotalTVL)
	assert.Equal(t, uint64(1000000), tvl.TVLByModule["dex"])
	assert.Equal(t, uint64(5000000), tvl.TVLByModule["staking"])
}

func TestAlertManagement(t *testing.T) {
	params := types.DefaultParams()
	params.EnableAlerts = true
	keeper := NewKeeper(&params)
	defer keeper.Close()

	// Create an alert
	alert, err := keeper.CreateAlert(
		types.AlertTypeSecurityThreat,
		types.SeverityHigh,
		"Test security threat",
		map[string]interface{}{"test": "data"},
	)
	require.NoError(t, err)
	require.NotNil(t, alert)

	// Verify alert is active
	activeAlerts := keeper.GetActiveAlerts()
	assert.Equal(t, 1, len(activeAlerts))

	// Acknowledge alert
	err = keeper.AcknowledgeAlert(alert.ID, "admin@aura.network")
	require.NoError(t, err)

	// Verify acknowledgment
	retrievedAlert, err := keeper.GetAlert(alert.ID)
	require.NoError(t, err)
	assert.True(t, retrievedAlert.Acknowledged)
	assert.Equal(t, "admin@aura.network", retrievedAlert.AcknowledgedBy)

	// Resolve alert
	err = keeper.ResolveAlert(alert.ID)
	require.NoError(t, err)

	// Verify alert is resolved
	retrievedAlert, err = keeper.GetAlert(alert.ID)
	require.NoError(t, err)
	assert.True(t, retrievedAlert.Resolved)

	// Verify no active alerts
	activeAlerts = keeper.GetActiveAlerts()
	assert.Equal(t, 0, len(activeAlerts))
}

func TestLogAggregation(t *testing.T) {
	params := types.DefaultParams()
	params.EnableLogAggregation = true
	keeper := NewKeeper(&params)
	defer keeper.Close()

	// Log some entries
	err := keeper.LogEntry(
		types.LogLevelInfo,
		"test-module",
		"Test info message",
		map[string]interface{}{"key": "value"},
		"trace-123",
		"span-456",
	)
	require.NoError(t, err)

	err = keeper.LogEntry(
		types.LogLevelError,
		"test-module",
		"Test error message",
		map[string]interface{}{"error": "test"},
		"trace-123",
		"span-789",
	)
	require.NoError(t, err)

	// Retrieve logs
	logs, err := keeper.GetLogs("test-module", 10)
	require.NoError(t, err)
	assert.Equal(t, 2, len(logs))

	// Get error logs
	errorLogs := keeper.GetErrorLogs(10)
	assert.Equal(t, 1, len(errorLogs))
	assert.Equal(t, types.LogLevelError, errorLogs[0].Level)

	// Get logs by trace ID
	tracedLogs := keeper.GetLogsByTraceID("trace-123")
	assert.Equal(t, 2, len(tracedLogs))
}

func TestFailedTransactionPattern(t *testing.T) {
	params := types.DefaultParams()
	params.EnableFailedTxAnalysis = true
	params.FailedTxPatternThreshold = 3
	keeper := NewKeeper(&params)
	defer keeper.Close()

	// Record failed transactions with same reason
	for i := 0; i < 5; i++ {
		tx := &types.TransactionMonitorData{
			TxHash:      generateID("tx"),
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

		err := keeper.RecordFailedTransaction(tx, "insufficient_funds")
		require.NoError(t, err)
	}

	// Verify pattern was detected
	patterns := keeper.GetFailedTransactionPatterns()
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

func TestCleanupExpiredData(t *testing.T) {
	params := types.DefaultParams()
	params.EnableAlerts = true
	params.AlertRetentionPeriod = 1 * time.Second
	keeper := NewKeeper(&params)
	defer keeper.Close()

	// Create and resolve an alert
	alert, err := keeper.CreateAlert(
		types.AlertTypeSystemError,
		types.SeverityLow,
		"Test alert",
		map[string]interface{}{},
	)
	require.NoError(t, err)

	err = keeper.ResolveAlert(alert.ID)
	require.NoError(t, err)

	// Wait for retention period
	time.Sleep(2 * time.Second)

	// Run cleanup
	keeper.cleanupExpiredData()

	// Verify alert was cleaned up
	_, err = keeper.GetAlert(alert.ID)
	assert.Error(t, err)
}
