// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package ml

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// AnomalyDetector implements machine learning-based anomaly detection
// NOTE: This detector is NON-CONSENSUS. It is used only for monitoring/alerting.
// All calculations use deterministic integer math to ensure consensus safety.
type AnomalyDetector struct {
	// Model parameters
	modelVersion     string
	thresholdBps     uint64 // Threshold in basis points (0-10000, where 10000 = 100%)
	trainingData     []DataPoint
	statistics       *Statistics
	mu               sync.RWMutex
	lastTraining     time.Time
	trainingInterval time.Duration
	accuracyBps      uint64 // Accuracy in basis points (0-10000)
}

// DataPoint represents a training data point
type DataPoint struct {
	Features  map[string]uint64 // All features stored as integer values (scaled appropriately)
	Timestamp time.Time
	Label     bool // true if anomaly, false if normal
}

// Statistics holds statistical measures for anomaly detection
// All values use deterministic integer arithmetic
type Statistics struct {
	Mean         map[string]uint64 // Mean values (scaled by 1e6 for precision)
	StdDev       map[string]uint64 // Standard deviation (scaled by 1e6)
	Min          map[string]uint64
	Max          map[string]uint64
	FeatureCount int
	SampleCount  int
	AnomalyRate  uint64 // In basis points (0-10000)
}

// Scaling constants for deterministic integer math
const (
	ScaleFactor    = 1_000_000 // 1e6 for mean/stddev precision
	BasisPointsMax = 10_000    // 100% = 10000 basis points
)

// NewAnomalyDetector creates a new anomaly detector
// threshold should be 0.0-1.0, it will be converted to basis points
func NewAnomalyDetector(threshold float64, trainingInterval time.Duration) *AnomalyDetector {
	// Convert float threshold to basis points (0-10000)
	thresholdBps := uint64(threshold * float64(BasisPointsMax))
	if thresholdBps > BasisPointsMax {
		thresholdBps = BasisPointsMax
	}

	return &AnomalyDetector{
		modelVersion:     "v2.0.0-deterministic",
		thresholdBps:     thresholdBps,
		trainingData:     make([]DataPoint, 0),
		statistics:       newStatistics(),
		trainingInterval: trainingInterval,
		accuracyBps:      0,
	}
}

// newStatistics creates a new Statistics object
func newStatistics() *Statistics {
	return &Statistics{
		Mean:   make(map[string]uint64),
		StdDev: make(map[string]uint64),
		Min:    make(map[string]uint64),
		Max:    make(map[string]uint64),
	}
}

// DetectTransactionAnomaly detects anomalies in transaction data
// All calculations are deterministic (integer-based) for consensus safety
func (ad *AnomalyDetector) DetectTransactionAnomaly(ctx context.Context, tx *types.TransactionMonitorData) (*types.AnomalyDetection, error) {
	if tx == nil {
		return nil, types.ErrInvalidTransaction
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockTime := sdkCtx.BlockTime()

	// Extract features from transaction (returns uint64 values)
	featuresInt := ad.extractTransactionFeatures(tx)

	// Calculate anomaly score in basis points (0-10000)
	scoreBps := ad.calculateAnomalyScoreBps(featuresInt)

	// Convert basis points to float64 for the AnomalyDetection struct
	// This is safe because it's only for output/display, not used in consensus logic
	scoreFloat := float64(scoreBps) / float64(BasisPointsMax)
	thresholdFloat := float64(ad.thresholdBps) / float64(BasisPointsMax)

	// Convert features to float64 map for compatibility with existing types
	featuresFloat := make(map[string]float64)
	for k, v := range featuresInt {
		featuresFloat[k] = float64(v)
	}

	detection := &types.AnomalyDetection{
		ID:        generateDetectionIDWithCtx(ctx),
		Type:      types.AnomalyTypeTransaction,
		Score:     scoreFloat,
		Threshold: thresholdFloat,
		IsAnomaly: scoreBps >= ad.thresholdBps,
		Features:  featuresFloat,
		Metadata: map[string]interface{}{
			"tx_hash":      tx.TxHash,
			"sender":       tx.Sender,
			"receiver":     tx.Receiver,
			"amount":       tx.Amount,
			"block_height": tx.BlockHeight,
		},
		Timestamp:    blockTime,
		ModelVersion: ad.modelVersion,
	}

	// Add to training data (non-consensus: used only for model updates)
	ad.addTrainingData(featuresInt, detection.IsAnomaly)

	// NOTE: Automatic retraining has been removed from consensus code.
	// The previous implementation used time.Since() which calls time.Now() (non-deterministic)
	// and spawned a goroutine which causes race conditions and non-determinism.
	//
	// Model retraining should be triggered via:
	// - Off-chain monitoring services that periodically call RetrainModel()
	// - CLI commands (e.g., `aurad tx monitoring retrain-model`)
	// - Scheduled cron jobs outside the consensus path
	//
	// The lastTraining field should use block time (ctx.BlockTime()) if tracking
	// is needed for on-chain logic, but retraining itself must remain off-chain.

	return detection, nil
}

// extractTransactionFeatures extracts numerical features from a transaction
// Returns deterministic uint64 values (no floating-point operations)
func (ad *AnomalyDetector) extractTransactionFeatures(tx *types.TransactionMonitorData) map[string]uint64 {
	features := make(map[string]uint64)

	// Amount-based features (use raw uint64 values)
	features["amount"] = tx.Amount
	// Instead of log1p, use scaled log approximation or just use amount directly
	// For determinism, we'll use a simple scaled amount (amount / 1000 for larger scale features)
	if tx.Amount > 0 {
		// Simple deterministic scaling: divide by 1000 to represent magnitude
		features["amount_magnitude"] = tx.Amount / 1000
	} else {
		features["amount_magnitude"] = 0
	}

	// Gas-based features (all uint64, no floating point)
	features["gas_used"] = tx.GasUsed
	features["gas_price"] = tx.GasPrice
	features["total_gas_cost"] = tx.GasUsed * tx.GasPrice

	// Time-based features (use uint64 for hour and day)
	features["hour_of_day"] = uint64(tx.Timestamp.Hour())
	features["day_of_week"] = uint64(tx.Timestamp.Weekday())

	// Block-based features (convert int64 to uint64)
	if tx.BlockHeight >= 0 {
		features["block_height"] = uint64(tx.BlockHeight)
	} else {
		features["block_height"] = 0
	}

	return features
}

// calculateAnomalyScoreBps calculates an anomaly score using deterministic statistical methods
// Returns score in basis points (0-10000, where 10000 = 100%)
func (ad *AnomalyDetector) calculateAnomalyScoreBps(features map[string]uint64) uint64 {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	if ad.statistics.SampleCount < 10 {
		// Not enough data, use simple heuristics
		return ad.simpleAnomalyScoreBps(features)
	}

	// Use deterministic Z-score based anomaly detection
	var totalZScore uint64 // Scaled by ScaleFactor for precision
	var featureCount uint64

	// Sort feature names for deterministic iteration
	featureNames := make([]string, 0, len(features))
	for name := range features {
		featureNames = append(featureNames, name)
	}
	sort.Strings(featureNames)

	for _, featureName := range featureNames {
		value := features[featureName]
		if mean, exists := ad.statistics.Mean[featureName]; exists {
			if stdDev, stdExists := ad.statistics.StdDev[featureName]; stdExists {
				if stdDev > 0 {
					// Calculate Z-score using integer arithmetic
					// Z = |value - mean| / stdDev
					var diff uint64
					if value > mean {
						diff = value - mean
					} else {
						diff = mean - value
					}
					// zScore = (diff * ScaleFactor) / stdDev
					// Both diff and stdDev are already scaled, so we don't need additional scaling
					zScore := (diff * ScaleFactor) / stdDev
					totalZScore += zScore
					featureCount++
				} else if value != mean {
					// StdDev is 0 (all training data identical), but value differs from mean
					// This is highly anomalous - add maximum Z-score
					totalZScore += 10 * ScaleFactor // High anomaly score
					featureCount++
				}
			}
		}
	}

	if featureCount == 0 {
		return 0
	}

	// Average Z-score
	avgZScore := totalZScore / featureCount

	// Normalize to basis points (0-10000)
	// Instead of exp(-x/3), use a simple linear mapping with saturation
	// Map Z-score of 0 -> 0 bps, Z-score of 3*ScaleFactor -> 10000 bps
	thresholdVal := uint64(3 * ScaleFactor)
	var scoreBps uint64
	if avgZScore >= thresholdVal {
		scoreBps = BasisPointsMax
	} else {
		// Linear interpolation: score = (avgZScore / threshold) * BasisPointsMax
		scoreBps = (avgZScore * BasisPointsMax) / thresholdVal
	}

	return scoreBps
}

// simpleAnomalyScoreBps provides a simple heuristic-based score when insufficient training data
// Returns score in basis points (0-10000)
func (ad *AnomalyDetector) simpleAnomalyScoreBps(features map[string]uint64) uint64 {
	var anomalyIndicatorsBps uint64

	// Check for unusually high amounts (add 3000 bps = 30%)
	if amount, exists := features["amount"]; exists && amount > 1_000_000 {
		anomalyIndicatorsBps += 3000
	}

	// Check for unusual gas usage (add 2000 bps = 20%)
	if gasUsed, exists := features["gas_used"]; exists && gasUsed > 1_000_000 {
		anomalyIndicatorsBps += 2000
	}

	// Check for unusual time patterns - late night/early morning (add 1000 bps = 10%)
	if hour, exists := features["hour_of_day"]; exists && (hour < 4 || hour > 22) {
		anomalyIndicatorsBps += 1000
	}

	// Cap at 10000 bps (100%)
	if anomalyIndicatorsBps > BasisPointsMax {
		return BasisPointsMax
	}
	return anomalyIndicatorsBps
}

// addTrainingData adds a new data point to the training set
// NOTE: Training data uses wall-clock time as it's for ML training (non-consensus)
func (ad *AnomalyDetector) addTrainingData(features map[string]uint64, isAnomaly bool) {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	dataPoint := DataPoint{
		Features:  features,
		Timestamp: time.Now(), // Non-consensus: training timestamps don't affect chain state
		Label:     isAnomaly,
	}

	ad.trainingData = append(ad.trainingData, dataPoint)

	// Keep only recent data (last 10000 points)
	if len(ad.trainingData) > 10000 {
		ad.trainingData = ad.trainingData[len(ad.trainingData)-10000:]
	}
}

// RetrainModel is the public method for triggering model retraining.
// This should ONLY be called from off-chain mechanisms such as:
// - CLI commands
// - Off-chain monitoring services
// - Scheduled cron jobs
//
// DO NOT call this from consensus code (BeginBlocker, EndBlocker, message handlers)
// as it uses wall-clock time and would cause non-determinism.
func (ad *AnomalyDetector) RetrainModel() {
	ad.retrain()
}

// retrain updates the model statistics based on training data
// NOTE: This is an internal method. Use RetrainModel() for external calls.
// Training uses wall-clock time as it's for ML model updates (non-consensus)
// Uses deterministic integer arithmetic for all calculations
func (ad *AnomalyDetector) retrain() {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	if len(ad.trainingData) < 10 {
		return
	}

	// Calculate new statistics using deterministic integer math
	newStats := newStatistics()
	featureSums := make(map[string]uint64)
	featureCounts := make(map[string]int)
	anomalyCount := 0

	// First pass: calculate sums for means
	for _, dp := range ad.trainingData {
		for fname, fval := range dp.Features {
			featureSums[fname] += fval
			featureCounts[fname]++
		}
		if dp.Label {
			anomalyCount++
		}
	}

	// Calculate means (scale by ScaleFactor for precision)
	for fname, sum := range featureSums {
		count := featureCounts[fname]
		if count > 0 {
			// mean = (sum * ScaleFactor) / count
			newStats.Mean[fname] = (sum * ScaleFactor) / uint64(count)
		}
	}

	// Second pass: calculate variance and min/max
	varianceSums := make(map[string]uint64)

	for _, dp := range ad.trainingData {
		for fname, fval := range dp.Features {
			mean := newStats.Mean[fname]

			// Calculate squared difference using integer math
			// Convert fval to same scale as mean
			fvalScaled := fval * ScaleFactor
			var squaredDiff uint64
			if fvalScaled > mean {
				diff := fvalScaled - mean
				squaredDiff = (diff * diff) / ScaleFactor // Divide once to keep scale reasonable
			} else {
				diff := mean - fvalScaled
				squaredDiff = (diff * diff) / ScaleFactor
			}
			varianceSums[fname] += squaredDiff

			// Update min/max (using unscaled values)
			if min, exists := newStats.Min[fname]; !exists || fval < min {
				newStats.Min[fname] = fval
			}
			if max, exists := newStats.Max[fname]; !exists || fval > max {
				newStats.Max[fname] = fval
			}
		}
	}

	// Calculate standard deviations using integer square root approximation
	for fname, varSum := range varianceSums {
		count := featureCounts[fname]
		if count > 1 {
			// variance = varSum / (count - 1)
			variance := varSum / uint64(count-1)
			// stddev = sqrt(variance) using integer square root
			newStats.StdDev[fname] = isqrt(variance)
		}
	}

	newStats.FeatureCount = len(featureSums)
	newStats.SampleCount = len(ad.trainingData)
	// Anomaly rate in basis points
	newStats.AnomalyRate = (uint64(anomalyCount) * BasisPointsMax) / uint64(len(ad.trainingData))

	ad.statistics = newStats
	ad.lastTraining = time.Now() // Non-consensus: training timestamps don't affect chain state

	// Calculate accuracy (simplified) - accuracy = 1 - anomaly_rate
	ad.accuracyBps = BasisPointsMax - newStats.AnomalyRate
}

// isqrt calculates integer square root using Newton's method (deterministic)
func isqrt(n uint64) uint64 {
	if n == 0 {
		return 0
	}
	if n <= 1 {
		return n
	}

	// Initial guess
	x := n
	y := (x + 1) / 2

	// Newton's method: y = (x + n/x) / 2
	for y < x {
		x = y
		y = (x + n/x) / 2
	}

	return x
}

// DetectNetworkAnomaly detects network-level anomalies using deterministic math
func (ad *AnomalyDetector) DetectNetworkAnomaly(ctx context.Context, health *types.NetworkHealth) (*types.AnomalyDetection, error) {
	if health == nil {
		return nil, fmt.Errorf("network health data cannot be nil")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockTime := sdkCtx.BlockTime()

	// Convert float64 fields to uint64 (scale by 1000 for precision)
	blockTimeScaled := uint64(health.BlockTime * 1000)
	tpsScaled := uint64(health.TPS * 1000)
	networkCongestionScaled := uint64(health.NetworkCongestion * 1000)
	consensusHealthScaled := uint64(health.ConsensusHealth * 1000)

	featuresInt := map[string]uint64{
		"block_time":         blockTimeScaled,
		"tps":                tpsScaled,
		"active_validators":  uint64(health.ActiveValidators),
		"peer_count":         uint64(health.PeerCount),
		"mempool_size":       uint64(health.MempoolSize),
		"network_congestion": networkCongestionScaled,
		"consensus_health":   consensusHealthScaled,
	}

	// Calculate score in basis points
	scoreBps := ad.calculateAnomalyScoreBps(featuresInt)

	// Convert to float64 for output
	scoreFloat := float64(scoreBps) / float64(BasisPointsMax)
	thresholdFloat := float64(ad.thresholdBps) / float64(BasisPointsMax)

	// Convert features to float64 for output
	featuresFloat := make(map[string]float64)
	featuresFloat["block_time"] = health.BlockTime
	featuresFloat["tps"] = health.TPS
	featuresFloat["active_validators"] = float64(health.ActiveValidators)
	featuresFloat["peer_count"] = float64(health.PeerCount)
	featuresFloat["mempool_size"] = float64(health.MempoolSize)
	featuresFloat["network_congestion"] = health.NetworkCongestion
	featuresFloat["consensus_health"] = health.ConsensusHealth

	detection := &types.AnomalyDetection{
		ID:        generateDetectionIDWithCtx(ctx),
		Type:      types.AnomalyTypeNetworkPattern,
		Score:     scoreFloat,
		Threshold: thresholdFloat,
		IsAnomaly: scoreBps >= ad.thresholdBps,
		Features:  featuresFloat,
		Metadata: map[string]interface{}{
			"block_height": health.BlockHeight,
			"timestamp":    health.Timestamp,
		},
		Timestamp:    blockTime,
		ModelVersion: ad.modelVersion,
	}

	return detection, nil
}

// GetModelInfo returns information about the current model
func (ad *AnomalyDetector) GetModelInfo() map[string]interface{} {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	// Convert basis points to float64 for display
	thresholdFloat := float64(ad.thresholdBps) / float64(BasisPointsMax)
	anomalyRateFloat := float64(ad.statistics.AnomalyRate) / float64(BasisPointsMax)
	accuracyFloat := float64(ad.accuracyBps) / float64(BasisPointsMax)

	return map[string]interface{}{
		"version":          ad.modelVersion,
		"threshold":        thresholdFloat,
		"threshold_bps":    ad.thresholdBps,
		"sample_count":     ad.statistics.SampleCount,
		"feature_count":    ad.statistics.FeatureCount,
		"anomaly_rate":     anomalyRateFloat,
		"anomaly_rate_bps": ad.statistics.AnomalyRate,
		"last_training":    ad.lastTraining,
		"accuracy":         accuracyFloat,
		"accuracy_bps":     ad.accuracyBps,
	}
}

// GetAccuracy returns the current model accuracy as float64
func (ad *AnomalyDetector) GetAccuracy() float64 {
	ad.mu.RLock()
	defer ad.mu.RUnlock()
	return float64(ad.accuracyBps) / float64(BasisPointsMax)
}

// GetModelVersion returns the current model version
func (ad *AnomalyDetector) GetModelVersion() string {
	ad.mu.RLock()
	defer ad.mu.RUnlock()
	return ad.modelVersion
}

// generateDetectionIDWithCtx generates a unique detection ID using block time (consensus-safe)
func generateDetectionIDWithCtx(ctx context.Context) string {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return fmt.Sprintf("detection-%d", sdkCtx.BlockTime().UnixNano())
}
