package ml

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// AnomalyDetector implements machine learning-based anomaly detection
type AnomalyDetector struct {
	// Model parameters
	modelVersion    string
	threshold       float64
	trainingData    []DataPoint
	statistics      *Statistics
	mu              sync.RWMutex
	lastTraining    time.Time
	trainingInterval time.Duration
	accuracy        float64
}

// DataPoint represents a training data point
type DataPoint struct {
	Features  map[string]float64
	Timestamp time.Time
	Label     bool // true if anomaly, false if normal
}

// Statistics holds statistical measures for anomaly detection
type Statistics struct {
	Mean              map[string]float64
	StdDev            map[string]float64
	Min               map[string]float64
	Max               map[string]float64
	FeatureCount      int
	SampleCount       int
	AnomalyRate       float64
}

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector(threshold float64, trainingInterval time.Duration) *AnomalyDetector {
	return &AnomalyDetector{
		modelVersion:     "v1.0.0",
		threshold:        threshold,
		trainingData:     make([]DataPoint, 0),
		statistics:       newStatistics(),
		trainingInterval: trainingInterval,
		accuracy:         0.0,
	}
}

// newStatistics creates a new Statistics object
func newStatistics() *Statistics {
	return &Statistics{
		Mean:   make(map[string]float64),
		StdDev: make(map[string]float64),
		Min:    make(map[string]float64),
		Max:    make(map[string]float64),
	}
}

// DetectTransactionAnomaly detects anomalies in transaction data
func (ad *AnomalyDetector) DetectTransactionAnomaly(tx *types.TransactionMonitorData) (*types.AnomalyDetection, error) {
	if tx == nil {
		return nil, types.ErrInvalidTransaction
	}

	// Extract features from transaction
	features := ad.extractTransactionFeatures(tx)

	// Calculate anomaly score
	score := ad.calculateAnomalyScore(features)

	detection := &types.AnomalyDetection{
		ID:           generateDetectionID(),
		Type:         types.AnomalyTypeTransaction,
		Score:        score,
		Threshold:    ad.threshold,
		IsAnomaly:    score >= ad.threshold,
		Features:     features,
		Metadata: map[string]interface{}{
			"tx_hash":      tx.TxHash,
			"sender":       tx.Sender,
			"receiver":     tx.Receiver,
			"amount":       tx.Amount,
			"block_height": tx.BlockHeight,
		},
		Timestamp:    time.Now(),
		ModelVersion: ad.modelVersion,
	}

	// Add to training data
	ad.addTrainingData(features, detection.IsAnomaly)

	// Retrain if needed
	if time.Since(ad.lastTraining) > ad.trainingInterval {
		go ad.retrain()
	}

	return detection, nil
}

// extractTransactionFeatures extracts numerical features from a transaction
func (ad *AnomalyDetector) extractTransactionFeatures(tx *types.TransactionMonitorData) map[string]float64 {
	features := make(map[string]float64)

	// Amount-based features
	features["amount"] = float64(tx.Amount)
	features["log_amount"] = math.Log1p(float64(tx.Amount))

	// Gas-based features
	features["gas_used"] = float64(tx.GasUsed)
	features["gas_price"] = float64(tx.GasPrice)
	features["total_gas_cost"] = float64(tx.GasUsed * tx.GasPrice)

	// Time-based features
	features["hour_of_day"] = float64(tx.Timestamp.Hour())
	features["day_of_week"] = float64(tx.Timestamp.Weekday())

	// Block-based features
	features["block_height"] = float64(tx.BlockHeight)

	return features
}

// calculateAnomalyScore calculates an anomaly score using statistical methods
func (ad *AnomalyDetector) calculateAnomalyScore(features map[string]float64) float64 {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	if ad.statistics.SampleCount < 10 {
		// Not enough data, use simple heuristics
		return ad.simpleAnomalyScore(features)
	}

	// Use Z-score based anomaly detection
	var totalZScore float64
	var featureCount int

	for featureName, value := range features {
		if mean, exists := ad.statistics.Mean[featureName]; exists {
			if stdDev, stdExists := ad.statistics.StdDev[featureName]; stdExists && stdDev > 0 {
				zScore := math.Abs((value - mean) / stdDev)
				totalZScore += zScore
				featureCount++
			}
		}
	}

	if featureCount == 0 {
		return 0.0
	}

	// Average Z-score, normalized to 0-1 range
	avgZScore := totalZScore / float64(featureCount)
	normalizedScore := 1.0 - math.Exp(-avgZScore/3.0)

	return math.Min(normalizedScore, 1.0)
}

// simpleAnomalyScore provides a simple heuristic-based score when insufficient training data
func (ad *AnomalyDetector) simpleAnomalyScore(features map[string]float64) float64 {
	var anomalyIndicators float64

	// Check for unusually high amounts
	if amount, exists := features["amount"]; exists && amount > 1000000 {
		anomalyIndicators += 0.3
	}

	// Check for unusual gas usage
	if gasUsed, exists := features["gas_used"]; exists && gasUsed > 1000000 {
		anomalyIndicators += 0.2
	}

	// Check for unusual time patterns (late night/early morning)
	if hour, exists := features["hour_of_day"]; exists && (hour < 4 || hour > 22) {
		anomalyIndicators += 0.1
	}

	return math.Min(anomalyIndicators, 1.0)
}

// addTrainingData adds a new data point to the training set
func (ad *AnomalyDetector) addTrainingData(features map[string]float64, isAnomaly bool) {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	dataPoint := DataPoint{
		Features:  features,
		Timestamp: time.Now(),
		Label:     isAnomaly,
	}

	ad.trainingData = append(ad.trainingData, dataPoint)

	// Keep only recent data (last 10000 points)
	if len(ad.trainingData) > 10000 {
		ad.trainingData = ad.trainingData[len(ad.trainingData)-10000:]
	}
}

// retrain updates the model statistics based on training data
func (ad *AnomalyDetector) retrain() {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	if len(ad.trainingData) < 10 {
		return
	}

	// Calculate new statistics
	newStats := newStatistics()
	featureSums := make(map[string]float64)
	featureCounts := make(map[string]int)
	anomalyCount := 0

	// First pass: calculate means
	for _, dp := range ad.trainingData {
		for fname, fval := range dp.Features {
			featureSums[fname] += fval
			featureCounts[fname]++
		}
		if dp.Label {
			anomalyCount++
		}
	}

	for fname, sum := range featureSums {
		count := featureCounts[fname]
		if count > 0 {
			newStats.Mean[fname] = sum / float64(count)
		}
	}

	// Second pass: calculate standard deviations and min/max
	varianceSums := make(map[string]float64)

	for _, dp := range ad.trainingData {
		for fname, fval := range dp.Features {
			mean := newStats.Mean[fname]
			varianceSums[fname] += math.Pow(fval-mean, 2)

			// Update min/max
			if min, exists := newStats.Min[fname]; !exists || fval < min {
				newStats.Min[fname] = fval
			}
			if max, exists := newStats.Max[fname]; !exists || fval > max {
				newStats.Max[fname] = fval
			}
		}
	}

	for fname, varSum := range varianceSums {
		count := featureCounts[fname]
		if count > 1 {
			variance := varSum / float64(count-1)
			newStats.StdDev[fname] = math.Sqrt(variance)
		}
	}

	newStats.FeatureCount = len(featureSums)
	newStats.SampleCount = len(ad.trainingData)
	newStats.AnomalyRate = float64(anomalyCount) / float64(len(ad.trainingData))

	ad.statistics = newStats
	ad.lastTraining = time.Now()

	// Calculate accuracy (simplified)
	ad.accuracy = 1.0 - newStats.AnomalyRate
}

// DetectNetworkAnomaly detects network-level anomalies
func (ad *AnomalyDetector) DetectNetworkAnomaly(health *types.NetworkHealth) (*types.AnomalyDetection, error) {
	if health == nil {
		return nil, fmt.Errorf("network health data cannot be nil")
	}

	features := map[string]float64{
		"block_time":         health.BlockTime,
		"tps":                health.TPS,
		"active_validators":  float64(health.ActiveValidators),
		"peer_count":         float64(health.PeerCount),
		"mempool_size":       float64(health.MempoolSize),
		"network_congestion": health.NetworkCongestion,
		"consensus_health":   health.ConsensusHealth,
	}

	score := ad.calculateAnomalyScore(features)

	detection := &types.AnomalyDetection{
		ID:        generateDetectionID(),
		Type:      types.AnomalyTypeNetworkPattern,
		Score:     score,
		Threshold: ad.threshold,
		IsAnomaly: score >= ad.threshold,
		Features:  features,
		Metadata: map[string]interface{}{
			"block_height": health.BlockHeight,
			"timestamp":    health.Timestamp,
		},
		Timestamp:    time.Now(),
		ModelVersion: ad.modelVersion,
	}

	return detection, nil
}

// GetModelInfo returns information about the current model
func (ad *AnomalyDetector) GetModelInfo() map[string]interface{} {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	return map[string]interface{}{
		"version":           ad.modelVersion,
		"threshold":         ad.threshold,
		"sample_count":      ad.statistics.SampleCount,
		"feature_count":     ad.statistics.FeatureCount,
		"anomaly_rate":      ad.statistics.AnomalyRate,
		"last_training":     ad.lastTraining,
		"accuracy":          ad.accuracy,
	}
}

// GetAccuracy returns the current model accuracy
func (ad *AnomalyDetector) GetAccuracy() float64 {
	ad.mu.RLock()
	defer ad.mu.RUnlock()
	return ad.accuracy
}

// GetModelVersion returns the current model version
func (ad *AnomalyDetector) GetModelVersion() string {
	ad.mu.RLock()
	defer ad.mu.RUnlock()
	return ad.modelVersion
}

// generateDetectionID generates a unique detection ID
func generateDetectionID() string {
	return fmt.Sprintf("detection-%d", time.Now().UnixNano())
}
