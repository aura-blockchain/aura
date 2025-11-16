package ml

import (
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/monitoring/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAnomalyDetector(t *testing.T) {
	detector := NewAnomalyDetector(0.75, 24*time.Hour)
	require.NotNil(t, detector)
	assert.Equal(t, 0.75, detector.threshold)
}

func TestDetectTransactionAnomaly(t *testing.T) {
	detector := NewAnomalyDetector(0.75, 24*time.Hour)

	tx := &types.TransactionMonitorData{
		TxHash:      "test-hash",
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

	detection, err := detector.DetectTransactionAnomaly(tx)
	require.NoError(t, err)
	require.NotNil(t, detection)

	assert.Equal(t, types.AnomalyTypeTransaction, detection.Type)
	assert.GreaterOrEqual(t, detection.Score, 0.0)
	assert.LessOrEqual(t, detection.Score, 1.0)
}

func TestSimpleAnomalyScore(t *testing.T) {
	detector := NewAnomalyDetector(0.75, 24*time.Hour)

	// Normal transaction
	normalFeatures := map[string]float64{
		"amount":   1000,
		"gas_used": 50000,
	}
	normalScore := detector.simpleAnomalyScore(normalFeatures)
	assert.Less(t, normalScore, 0.5)

	// Large transaction
	largeFeatures := map[string]float64{
		"amount":   2000000,
		"gas_used": 50000,
	}
	largeScore := detector.simpleAnomalyScore(largeFeatures)
	assert.Greater(t, largeScore, normalScore)
}

func TestExtractTransactionFeatures(t *testing.T) {
	detector := NewAnomalyDetector(0.75, 24*time.Hour)

	tx := &types.TransactionMonitorData{
		TxHash:      "test-hash",
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

	features := detector.extractTransactionFeatures(tx)

	assert.Contains(t, features, "amount")
	assert.Contains(t, features, "gas_used")
	assert.Contains(t, features, "gas_price")
	assert.Contains(t, features, "hour_of_day")
	assert.Contains(t, features, "block_height")

	assert.Equal(t, float64(1000), features["amount"])
	assert.Equal(t, float64(50000), features["gas_used"])
}

func TestAddTrainingData(t *testing.T) {
	detector := NewAnomalyDetector(0.75, 24*time.Hour)

	features := map[string]float64{
		"amount":   1000,
		"gas_used": 50000,
	}

	detector.addTrainingData(features, false)
	assert.Equal(t, 1, len(detector.trainingData))

	detector.addTrainingData(features, true)
	assert.Equal(t, 2, len(detector.trainingData))
}

func TestRetrain(t *testing.T) {
	detector := NewAnomalyDetector(0.75, 24*time.Hour)

	// Add some training data
	for i := 0; i < 20; i++ {
		features := map[string]float64{
			"amount":   float64(1000 + i*100),
			"gas_used": float64(50000 + i*1000),
		}
		detector.addTrainingData(features, false)
	}

	// Retrain
	detector.retrain()

	// Check statistics were calculated
	assert.Greater(t, detector.statistics.SampleCount, 0)
	assert.Greater(t, detector.statistics.FeatureCount, 0)
	assert.NotZero(t, detector.statistics.Mean["amount"])
}

func TestCalculateAnomalyScore(t *testing.T) {
	detector := NewAnomalyDetector(0.75, 24*time.Hour)

	// Add training data with normal patterns
	for i := 0; i < 30; i++ {
		features := map[string]float64{
			"amount":   float64(1000),
			"gas_used": float64(50000),
		}
		detector.addTrainingData(features, false)
	}

	detector.retrain()

	// Test with normal transaction
	normalFeatures := map[string]float64{
		"amount":   1000,
		"gas_used": 50000,
	}
	normalScore := detector.calculateAnomalyScore(normalFeatures)
	assert.Less(t, normalScore, 0.5)

	// Test with anomalous transaction
	anomalousFeatures := map[string]float64{
		"amount":   100000, // Much higher
		"gas_used": 50000,
	}
	anomalousScore := detector.calculateAnomalyScore(anomalousFeatures)
	assert.Greater(t, anomalousScore, normalScore)
}

func TestDetectNetworkAnomaly(t *testing.T) {
	detector := NewAnomalyDetector(0.75, 24*time.Hour)

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

	detection, err := detector.DetectNetworkAnomaly(health)
	require.NoError(t, err)
	require.NotNil(t, detection)

	assert.Equal(t, types.AnomalyTypeNetworkPattern, detection.Type)
	assert.Contains(t, detection.Features, "block_time")
	assert.Contains(t, detection.Features, "tps")
	assert.Contains(t, detection.Features, "network_congestion")
}

func TestGetModelInfo(t *testing.T) {
	detector := NewAnomalyDetector(0.75, 24*time.Hour)

	// Add some data and retrain
	for i := 0; i < 15; i++ {
		features := map[string]float64{"amount": float64(1000 + i*100)}
		detector.addTrainingData(features, false)
	}
	detector.retrain()

	info := detector.GetModelInfo()
	assert.Contains(t, info, "version")
	assert.Contains(t, info, "threshold")
	assert.Contains(t, info, "sample_count")
	assert.Equal(t, 0.75, info["threshold"])
}

func TestGetAccuracy(t *testing.T) {
	detector := NewAnomalyDetector(0.75, 24*time.Hour)

	// Initially should be 0
	assert.Equal(t, 0.0, detector.GetAccuracy())

	// After training with normal data
	for i := 0; i < 20; i++ {
		features := map[string]float64{"amount": float64(1000)}
		detector.addTrainingData(features, false)
	}
	detector.retrain()

	// Should have high accuracy (all normal data)
	assert.Greater(t, detector.GetAccuracy(), 0.9)
}

func TestModelVersion(t *testing.T) {
	detector := NewAnomalyDetector(0.75, 24*time.Hour)
	version := detector.GetModelVersion()
	assert.NotEmpty(t, version)
	assert.Equal(t, "v1.0.0", version)
}
