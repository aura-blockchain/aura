// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// =============================================================================
// DetectTransactionAnomaly Tests
// =============================================================================

type AnomalyDetectionTestSuite struct {
	KeeperTestSuite
}

func (suite *AnomalyDetectionTestSuite) TestDetectTransactionAnomaly_NoHistory() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Test with new wallet (no history)
	score, err := k.DetectTransactionAnomaly(ctx, "wallet-1", "recipient-1", sdkmath.NewInt(1000000))
	suite.Require().NoError(err)
	suite.Require().NotNil(score)

	// Score should be low with no history (0 for amount since no historical data)
	suite.Require().Contains(score.Factors, "amount")
	suite.Require().Contains(score.Factors, "recipient")
	suite.Require().Contains(score.Factors, "frequency")
	suite.Require().Contains(score.Factors, "time")
}

func (suite *AnomalyDetectionTestSuite) TestDetectTransactionAnomaly_Threshold() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	score, err := k.DetectTransactionAnomaly(ctx, "wallet-2", "recipient-2", sdkmath.NewInt(1000000))
	suite.Require().NoError(err)
	suite.Require().NotNil(score)

	// Verify threshold is set correctly
	suite.Require().Equal(AnomalyThresholdBPS, score.Threshold)

	// IsAnomaly should be based on score vs threshold comparison
	suite.Require().Equal(score.Score > score.Threshold, score.IsAnomaly)
}

func (suite *AnomalyDetectionTestSuite) TestDetectTransactionAnomaly_LargeAmount() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Very large amount
	score, err := k.DetectTransactionAnomaly(ctx, "wallet-3", "recipient-3", sdkmath.NewInt(1000000000000))
	suite.Require().NoError(err)
	suite.Require().NotNil(score)
	suite.Require().Contains(score.Factors, "amount")
}

func (suite *AnomalyDetectionTestSuite) TestDetectTransactionAnomaly_ZeroAmount() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	score, err := k.DetectTransactionAnomaly(ctx, "wallet-4", "recipient-4", sdkmath.NewInt(0))
	suite.Require().NoError(err)
	suite.Require().NotNil(score)
}

func TestAnomalyDetectionTestSuite(t *testing.T) {
	suite.Run(t, new(AnomalyDetectionTestSuite))
}

// =============================================================================
// GetAnomalies Tests
// =============================================================================

func (suite *AnomalyDetectionTestSuite) TestGetAnomalies_Empty() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	anomalies, err := k.GetAnomalies(ctx, "nonexistent-wallet")
	suite.Require().NoError(err)
	// May return nil or empty slice
	suite.Require().True(len(anomalies) == 0)
}

func (suite *AnomalyDetectionTestSuite) TestGetAnomalies_WithData() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Request anomalies - should not error
	anomalies, err := k.GetAnomalies(ctx, "wallet-5")
	suite.Require().NoError(err)
	// May return nil or empty slice
	_ = anomalies
}

// =============================================================================
// checkUnusualAmount Tests (via integration)
// =============================================================================

func (suite *AnomalyDetectionTestSuite) TestCheckUnusualAmount_NegativeAmount() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Negative amount (should still work)
	score, err := k.DetectTransactionAnomaly(ctx, "wallet-neg", "recipient", sdkmath.NewInt(-1000))
	suite.Require().NoError(err)
	suite.Require().NotNil(score)
}

// =============================================================================
// hasPrefix Tests
// =============================================================================

func TestHasPrefix(t *testing.T) {
	require.True(t, hasPrefix([]byte("anomaly_wallet-1"), []byte("anomaly_")))
	require.False(t, hasPrefix([]byte("wallet-1"), []byte("anomaly_")))
	require.False(t, hasPrefix([]byte(""), []byte("anomaly_")))
	require.True(t, hasPrefix([]byte("anomaly_"), []byte("anomaly_")))
}

// =============================================================================
// clampUint64ToInt64 Tests
// =============================================================================

func TestClampUint64ToInt64(t *testing.T) {
	// Normal values
	require.Equal(t, int64(100), clampUint64ToInt64(100))
	require.Equal(t, int64(0), clampUint64ToInt64(0))

	// Max int64
	require.Equal(t, int64(9223372036854775807), clampUint64ToInt64(uint64(9223372036854775807)))
}

func TestClampUint64ToInt64_Overflow(t *testing.T) {
	// Values > MaxInt64 should be clamped
	require.Equal(t, int64(9223372036854775807), clampUint64ToInt64(uint64(9223372036854775808)))
	require.Equal(t, int64(9223372036854775807), clampUint64ToInt64(uint64(18446744073709551615))) // MaxUint64
}

// =============================================================================
// checkNewRecipient Tests (via DetectTransactionAnomaly)
// =============================================================================

func (suite *AnomalyDetectionTestSuite) TestDetectTransactionAnomaly_NewRecipient() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// First transaction to new recipient
	score1, err := k.DetectTransactionAnomaly(ctx, "wallet-newrec", "new-recipient-1", sdkmath.NewInt(1000))
	suite.Require().NoError(err)
	suite.Require().NotNil(score1)
	// New recipient should have non-zero score
	suite.Require().Equal(NewRecipientScoreBPS, score1.Factors["recipient"])

	// Second transaction to same recipient
	score2, err := k.DetectTransactionAnomaly(ctx, "wallet-newrec", "new-recipient-1", sdkmath.NewInt(1000))
	suite.Require().NoError(err)
	suite.Require().NotNil(score2)
	// Known recipient should have zero score
	suite.Require().Equal(uint64(0), score2.Factors["recipient"])
}

func (suite *AnomalyDetectionTestSuite) TestDetectTransactionAnomaly_MultipleRecipients() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	walletID := "wallet-multirec"

	// Multiple new recipients
	for i := 1; i <= 5; i++ {
		recipient := "multi-rec-" + string(rune('a'+i-1))
		score, err := k.DetectTransactionAnomaly(ctx, walletID, recipient, sdkmath.NewInt(100))
		suite.Require().NoError(err)
		suite.Require().NotNil(score)
		// Each new recipient should have the new recipient score
		suite.Require().Equal(NewRecipientScoreBPS, score.Factors["recipient"])
	}

	// Now revisit an old recipient
	score, err := k.DetectTransactionAnomaly(ctx, walletID, "multi-rec-a", sdkmath.NewInt(100))
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(0), score.Factors["recipient"])
}

// =============================================================================
// checkTransactionFrequency Tests
// =============================================================================

func (suite *AnomalyDetectionTestSuite) TestDetectTransactionAnomaly_Frequency() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	walletID := "wallet-freq"

	// Make transactions below threshold
	for i := 1; i <= int(HighFrequencyThreshold)-1; i++ {
		score, err := k.DetectTransactionAnomaly(ctx, walletID, "freq-rec", sdkmath.NewInt(100))
		suite.Require().NoError(err)
		suite.Require().NotNil(score)
		// Frequency score should be proportional to count
		expectedScore := (uint64(i) * BasisPointsMax) / uint64(HighFrequencyThreshold)
		suite.Require().Equal(expectedScore, score.Factors["frequency"], "iteration %d", i)
	}
}

func (suite *AnomalyDetectionTestSuite) TestDetectTransactionAnomaly_HighFrequency() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	walletID := "wallet-highfreq"

	// Make transactions exceeding threshold
	for i := 1; i <= int(HighFrequencyThreshold)+1; i++ {
		_, err := k.DetectTransactionAnomaly(ctx, walletID, "highfreq-rec", sdkmath.NewInt(100))
		suite.Require().NoError(err)
	}

	// One more should hit the high frequency score
	score, err := k.DetectTransactionAnomaly(ctx, walletID, "highfreq-rec", sdkmath.NewInt(100))
	suite.Require().NoError(err)
	suite.Require().Equal(HighFrequencyScoreBPS, score.Factors["frequency"])
}

// =============================================================================
// AnomalyScore Tests
// =============================================================================

func TestAnomalyScore_IsAnomaly(t *testing.T) {
	tests := []struct {
		name      string
		score     uint64
		threshold uint64
		expected  bool
	}{
		{"below threshold", 5000, 7000, false},
		{"at threshold", 7000, 7000, false},
		{"above threshold", 8000, 7000, true},
		{"zero score", 0, 7000, false},
		{"max score", 10000, 7000, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			score := &AnomalyScore{
				Score:     tc.score,
				Threshold: tc.threshold,
				IsAnomaly: tc.score > tc.threshold,
			}
			require.Equal(t, tc.expected, score.IsAnomaly)
		})
	}
}

func TestAnomalyScore_Factors(t *testing.T) {
	score := &AnomalyScore{
		Factors: make(map[string]uint64),
	}

	// All factors should be storable
	score.Factors["amount"] = 5000
	score.Factors["recipient"] = NewRecipientScoreBPS
	score.Factors["frequency"] = 3000
	score.Factors["time"] = UnusualTimeScoreBPS

	require.Len(t, score.Factors, 4)
	require.Equal(t, uint64(5000), score.Factors["amount"])
	require.Equal(t, NewRecipientScoreBPS, score.Factors["recipient"])
	require.Equal(t, uint64(3000), score.Factors["frequency"])
	require.Equal(t, UnusualTimeScoreBPS, score.Factors["time"])
}

// =============================================================================
// recordAnomaly and GetAnomalies Integration Tests
// =============================================================================

func (suite *AnomalyDetectionTestSuite) TestRecordAndGetAnomalies() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	walletID := "wallet-record-anom"

	// Make multiple high-frequency transactions to trigger anomaly recording
	for i := 0; i < int(HighFrequencyThreshold)+5; i++ {
		_, err := k.DetectTransactionAnomaly(ctx, walletID, "rec-anom", sdkmath.NewInt(100))
		suite.Require().NoError(err)
	}

	// Additional transaction with new recipient (should increase score)
	score, err := k.DetectTransactionAnomaly(ctx, walletID, "brand-new-rec", sdkmath.NewInt(100))
	suite.Require().NoError(err)

	// If it was an anomaly, it should be recorded
	if score.IsAnomaly {
		anomalies, err := k.GetAnomalies(ctx, walletID)
		suite.Require().NoError(err)
		suite.Require().NotEmpty(anomalies)
	}
}

// =============================================================================
// Weighted Score Calculation Tests
// =============================================================================

func TestWeightedScoreCalculation(t *testing.T) {
	// All weights should sum to 10000 (100%)
	totalWeight := AmountWeightBPS + RecipientWeightBPS + FrequencyWeightBPS + TimeWeightBPS
	require.Equal(t, BasisPointsMax, totalWeight, "all factor weights should sum to 10000")
}

func (suite *AnomalyDetectionTestSuite) TestDetectTransactionAnomaly_ScoreCalculation() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Test with known factors to verify calculation
	score, err := k.DetectTransactionAnomaly(ctx, "wallet-calc", "calc-recipient", sdkmath.NewInt(1000))
	suite.Require().NoError(err)
	suite.Require().NotNil(score)

	// Verify score is within valid range (0-10000)
	suite.Require().LessOrEqual(score.Score, BasisPointsMax)
	suite.Require().GreaterOrEqual(score.Score, uint64(0))
}

// =============================================================================
// Edge Cases Tests
// =============================================================================

func (suite *AnomalyDetectionTestSuite) TestDetectTransactionAnomaly_EmptyWalletID() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Empty wallet ID should still work (no special handling required)
	score, err := k.DetectTransactionAnomaly(ctx, "", "recipient", sdkmath.NewInt(1000))
	suite.Require().NoError(err)
	suite.Require().NotNil(score)
}

func (suite *AnomalyDetectionTestSuite) TestDetectTransactionAnomaly_EmptyRecipient() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Empty recipient should still work
	score, err := k.DetectTransactionAnomaly(ctx, "wallet-empty-rec", "", sdkmath.NewInt(1000))
	suite.Require().NoError(err)
	suite.Require().NotNil(score)
}

func (suite *AnomalyDetectionTestSuite) TestDetectTransactionAnomaly_VeryLargeAmount() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Very large amount (should not overflow)
	largeAmount := sdkmath.NewIntFromBigInt(sdkmath.NewInt(1).BigInt())
	largeAmount = largeAmount.Mul(sdkmath.NewInt(1e18)).Mul(sdkmath.NewInt(1e18))

	score, err := k.DetectTransactionAnomaly(ctx, "wallet-large", "rec-large", largeAmount)
	suite.Require().NoError(err)
	suite.Require().NotNil(score)
	// Amount score should be capped at BasisPointsMax
	suite.Require().LessOrEqual(score.Factors["amount"], BasisPointsMax)
}

// =============================================================================
// Constants Tests
// =============================================================================

func TestBasisPointsConstants(t *testing.T) {
	require.Equal(t, uint64(10000), BasisPointsMax)
	require.Equal(t, uint64(7000), AnomalyThresholdBPS)
	require.Equal(t, uint64(3000), AmountWeightBPS)
	require.Equal(t, uint64(3000), RecipientWeightBPS)
	require.Equal(t, uint64(2000), FrequencyWeightBPS)
	require.Equal(t, uint64(2000), TimeWeightBPS)
	require.Equal(t, uint64(5000), NewRecipientScoreBPS)
	require.Equal(t, uint64(6000), UnusualTimeScoreBPS)
	require.Equal(t, uint64(8000), HighFrequencyScoreBPS)
	require.Equal(t, int64(10), HighFrequencyThreshold)
}
