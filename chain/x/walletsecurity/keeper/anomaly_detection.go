package keeper

import (
	"context"
	"fmt"
	stdmath "math"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/common/determinism"
	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

// AnomalyScore represents an anomaly detection score
type AnomalyScore struct {
	Score      float64
	Threshold  float64
	IsAnomaly  bool
	Factors    map[string]float64
	Timestamp  sdk.Context
}

// DetectTransactionAnomaly detects anomalies in transaction patterns
func (k Keeper) DetectTransactionAnomaly(ctx context.Context, walletID, recipient string, amount math.Int) (*AnomalyScore, error) {
	score := &AnomalyScore{
		Threshold: 0.7,
		Factors:   make(map[string]float64),
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

	// Calculate overall score (weighted average)
	score.Score = (amountScore*0.3 + recipientScore*0.3 + frequencyScore*0.2 + timeScore*0.2)
	score.IsAnomaly = score.Score > score.Threshold

	// Store anomaly detection result
	if score.IsAnomaly {
		k.recordAnomaly(ctx, walletID, score)
	}

	return score, nil
}

func (k Keeper) checkUnusualAmount(ctx context.Context, walletID string, amount math.Int) float64 {
	// Get historical transaction amounts for this wallet
	avg, stdDev := k.getAmountStatistics(ctx, walletID)

	if stdDev.IsZero() {
		return 0.0 // No historical data
	}

	// Calculate z-score
	diff := amount.Sub(avg)
	zScore := stdmath.Abs(float64(diff.Int64()) / float64(stdDev.Int64()))

	// Normalize to 0-1 range
	anomalyScore := stdmath.Min(zScore/3.0, 1.0)

	return anomalyScore
}

func (k Keeper) checkNewRecipient(ctx context.Context, walletID, recipient string) float64 {
	// Check if recipient has been used before
	kvStore := k.getStore(ctx)
	recipientKey := []byte(fmt.Sprintf("recipient_history_%s_%s", walletID, recipient))

	has, err := kvStore.Has(recipientKey)
	if err == nil && has {
		return 0.0 // Known recipient
	}

	// Mark recipient as seen
	kvStore.Set(recipientKey, []byte(determinism.GetBlockTime(ctx).String()))

	return 0.5 // New recipient is moderately suspicious
}

func (k Keeper) checkTransactionFrequency(ctx context.Context, walletID string) float64 {
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

	// If more than 10 transactions in an hour, flag as anomalous
	if count > 10 {
		return 0.8
	}

	return float64(count) / 10.0
}

func (k Keeper) checkUnusualTime(ctx context.Context, walletID string) float64 {
	// Check if transaction is at unusual time (e.g., 2-6 AM)
	hour := determinism.GetBlockTime(ctx).Hour()

	if hour >= 2 && hour <= 6 {
		return 0.6 // Moderately suspicious
	}

	return 0.0
}

func (k Keeper) getAmountStatistics(ctx context.Context, walletID string) (math.Int, math.Int) {
	// Simplified: return placeholder statistics
	// In production, calculate from transaction history
	return math.NewInt(1000), math.NewInt(500)
}

func (k Keeper) recordAnomaly(ctx context.Context, walletID string, score *AnomalyScore) {
	anomaly := &wsproto.AnomalyDetection{
		WalletId:   walletID,
		Score:      score.Score,
		Threshold:  score.Threshold,
		DetectedAt: gogoTimestampNow(),
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
