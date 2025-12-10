package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/common/determinism"
	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

// Basis points constants (0-10000 representing 0.00%-100.00%)
const (
	BasisPointsMax         uint64 = 10000 // 100.00%
	AnomalyThresholdBPS    uint64 = 7000  // 70.00%
	AmountWeightBPS        uint64 = 3000  // 30.00%
	RecipientWeightBPS     uint64 = 3000  // 30.00%
	FrequencyWeightBPS     uint64 = 2000  // 20.00%
	TimeWeightBPS          uint64 = 2000  // 20.00%
	NewRecipientScoreBPS   uint64 = 5000  // 50.00%
	UnusualTimeScoreBPS    uint64 = 6000  // 60.00%
	HighFrequencyScoreBPS  uint64 = 8000  // 80.00%
	HighFrequencyThreshold int64  = 10
)

// AnomalyScore represents an anomaly detection score using basis points (0-10000)
type AnomalyScore struct {
	Score      uint64            // Basis points (0-10000)
	Threshold  uint64            // Basis points (0-10000)
	IsAnomaly  bool
	Factors    map[string]uint64 // Each factor in basis points
	Timestamp  sdk.Context
}

// DetectTransactionAnomaly detects anomalies in transaction patterns
func (k Keeper) DetectTransactionAnomaly(ctx context.Context, walletID, recipient string, amount math.Int) (*AnomalyScore, error) {
	score := &AnomalyScore{
		Threshold: AnomalyThresholdBPS,
		Factors:   make(map[string]uint64),
	}

	// Factor 1: Unusual amount
	amountScore := k.checkUnusualAmount(ctx, walletID, amount)
	score.Factors["amount"] = amountScore

	// Factor 2: New recipient
	recipientScore := k.checkNewRecipient(ctx, walletID, recipient)
	score.Factors["recipient"] = recipientScore

	// Factor 3: Transaction frequency
	frequencyScore := k.checkTransactionFrequency(ctx, walletID)
	score.Factors["frequency"] = frequencyScore

	// Factor 4: Time-based anomalies
	timeScore := k.checkUnusualTime(ctx, walletID)
	score.Factors["time"] = timeScore

	// Calculate overall score (weighted average using basis points)
	// All weights sum to 10000 (100%)
	score.Score = (amountScore*AmountWeightBPS +
		recipientScore*RecipientWeightBPS +
		frequencyScore*FrequencyWeightBPS +
		timeScore*TimeWeightBPS) / BasisPointsMax
	score.IsAnomaly = score.Score > score.Threshold

	// Store anomaly detection result
	if score.IsAnomaly {
		k.recordAnomaly(ctx, walletID, score)
	}

	return score, nil
}

func (k Keeper) checkUnusualAmount(ctx context.Context, walletID string, amount math.Int) uint64 {
	// Get historical transaction amounts for this wallet
	avg, stdDev := k.getAmountStatistics(ctx, walletID)

	if stdDev.IsZero() {
		return 0 // No historical data
	}

	// Calculate z-score using deterministic integer arithmetic
	// z-score = |amount - avg| / stdDev
	diff := amount.Sub(avg)

	// Get absolute value of diff
	absDiff := diff
	if diff.IsNegative() {
		absDiff = diff.Neg()
	}

	// Calculate z-score scaled by 10000 to avoid decimals
	// zScore = (absDiff * 10000) / stdDev
	zScoreScaled := absDiff.Mul(math.NewInt(10000)).Quo(stdDev)

	// Normalize to 0-10000 basis points range
	// Original formula: min(zScore/3.0, 1.0)
	// Convert to basis points: min((zScoreScaled * 10000) / 30000, 10000)
	anomalyScoreBPS := zScoreScaled.Mul(math.NewInt(10000)).Quo(math.NewInt(30000))

	// Cap at BasisPointsMax (10000)
	if anomalyScoreBPS.GT(math.NewInt(int64(BasisPointsMax))) {
		return BasisPointsMax
	}

	return anomalyScoreBPS.Uint64()
}

func (k Keeper) checkNewRecipient(ctx context.Context, walletID, recipient string) uint64 {
	// Check if recipient has been used before
	kvStore := k.getStore(ctx)
	recipientKey := []byte(fmt.Sprintf("recipient_history_%s_%s", walletID, recipient))

	has, err := kvStore.Has(recipientKey)
	if err == nil && has {
		return 0 // Known recipient
	}

	// Mark recipient as seen
	kvStore.Set(recipientKey, []byte(determinism.GetBlockTime(ctx).String()))

	return NewRecipientScoreBPS // New recipient is moderately suspicious (50.00%)
}

func (k Keeper) checkTransactionFrequency(ctx context.Context, walletID string) uint64 {
	kvStore := k.getStore(ctx)

	// Get transaction count in last hour
	countKey := []byte(fmt.Sprintf("tx_count_1h_%s", walletID))
	countBytes, _ := kvStore.Get(countKey)

	var count int64
	if countBytes != nil {
		count = int64(sdk.BigEndianToUint64(countBytes))
	}

	count++
	kvStore.Set(countKey, sdk.Uint64ToBigEndian(uint64(count)))

	// If more than threshold transactions in an hour, flag as highly anomalous
	if count > HighFrequencyThreshold {
		return HighFrequencyScoreBPS // 80.00%
	}

	// Linear scale: (count / threshold) * BasisPointsMax
	// Returns basis points proportional to transaction count
	score := (uint64(count) * BasisPointsMax) / uint64(HighFrequencyThreshold)
	return score
}

func (k Keeper) checkUnusualTime(ctx context.Context, walletID string) uint64 {
	// Check if transaction is at unusual time (e.g., 2-6 AM)
	hour := determinism.GetBlockTime(ctx).Hour()

	if hour >= 2 && hour <= 6 {
		return UnusualTimeScoreBPS // Moderately suspicious (60.00%)
	}

	return 0
}

func (k Keeper) getAmountStatistics(ctx context.Context, walletID string) (math.Int, math.Int) {
	// Simplified: return placeholder statistics
	// In production, calculate from transaction history
	return math.NewInt(1000), math.NewInt(500)
}

func (k Keeper) recordAnomaly(ctx context.Context, walletID string, score *AnomalyScore) {
	// Convert basis points to float64 for proto storage (legacy compatibility)
	// Note: This conversion is safe as it happens AFTER all consensus-critical
	// calculations are complete using deterministic uint64 basis points
	scoreFloat := float64(score.Score) / float64(BasisPointsMax)
	thresholdFloat := float64(score.Threshold) / float64(BasisPointsMax)

	anomaly := &wsproto.AnomalyDetection{
		WalletId:   walletID,
		Score:      scoreFloat,
		Threshold:  thresholdFloat,
		DetectedAt: blockTimeToGogoTimestamp(ctx),
		Resolved:   false,
	}

	anomalyBytes, _ := k.cdc.Marshal(anomaly)

	kvStore := k.getStore(ctx)
	key := []byte(fmt.Sprintf("anomaly_%s_%d", walletID, determinism.GetBlockTime(ctx).UnixNano()))
	kvStore.Set(key, anomalyBytes)
}

// GetAnomalies retrieves detected anomalies for a wallet
func (k Keeper) GetAnomalies(ctx context.Context, walletID string) ([]*wsproto.AnomalyDetection, error) {
	kvStore := k.getStore(ctx)
	prefix := []byte(fmt.Sprintf("anomaly_%s_", walletID))

	iterator, err := kvStore.Iterator(prefix, nil)
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var anomalies []*wsproto.AnomalyDetection
	for ; iterator.Valid(); iterator.Next() {
		// Check if key starts with prefix
		if !hasPrefix(iterator.Key(), prefix) {
			break
		}

		var anomaly wsproto.AnomalyDetection
		if err := k.cdc.Unmarshal(iterator.Value(), &anomaly); err != nil {
			continue
		}
		anomalies = append(anomalies, &anomaly)
	}

	return anomalies, nil
}

// hasPrefix checks if a byte slice has a given prefix
func hasPrefix(key, prefix []byte) bool {
	if len(key) < len(prefix) {
		return false
	}
	for i := range prefix {
		if key[i] != prefix[i] {
			return false
		}
	}
	return true
}
