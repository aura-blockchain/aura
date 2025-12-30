// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"math"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// =============================================================================
// Proto Helpers Test Suite
// =============================================================================

type ProtoHelpersTestSuite struct {
	KeeperTestSuite
}

func TestProtoHelpersTestSuite(t *testing.T) {
	suite.Run(t, new(ProtoHelpersTestSuite))
}

// =============================================================================
// timestampToGogo Tests
// =============================================================================

func TestTimestampToGogo_NilInput(t *testing.T) {
	result := timestampToGogo(nil)
	require.Nil(t, result, "nil input should return nil")
}

func TestTimestampToGogo_ValidTimestamp(t *testing.T) {
	ts := &timestamppb.Timestamp{
		Seconds: 1704067200,
		Nanos:   500000000,
	}

	result := timestampToGogo(ts)

	require.NotNil(t, result)
	require.Equal(t, int64(1704067200), result.Seconds)
	require.Equal(t, int32(500000000), result.Nanos)
}

func TestTimestampToGogo_ZeroTimestamp(t *testing.T) {
	ts := &timestamppb.Timestamp{
		Seconds: 0,
		Nanos:   0,
	}

	result := timestampToGogo(ts)

	require.NotNil(t, result)
	require.Equal(t, int64(0), result.Seconds)
	require.Equal(t, int32(0), result.Nanos)
}

func TestTimestampToGogo_NegativeTimestamp(t *testing.T) {
	// Negative timestamps represent times before Unix epoch
	ts := &timestamppb.Timestamp{
		Seconds: -86400,
		Nanos:   123456789,
	}

	result := timestampToGogo(ts)

	require.NotNil(t, result)
	require.Equal(t, int64(-86400), result.Seconds)
	require.Equal(t, int32(123456789), result.Nanos)
}

func TestTimestampToGogo_MaxValues(t *testing.T) {
	ts := &timestamppb.Timestamp{
		Seconds: math.MaxInt64,
		Nanos:   999999999,
	}

	result := timestampToGogo(ts)

	require.NotNil(t, result)
	require.Equal(t, int64(math.MaxInt64), result.Seconds)
	require.Equal(t, int32(999999999), result.Nanos)
}

// =============================================================================
// gogoTimestampNow Tests
// =============================================================================

func TestGogoTimestampNow_ReturnsNonNil(t *testing.T) {
	result := gogoTimestampNow()

	require.NotNil(t, result)
	// Should be a reasonable recent timestamp (after 2020)
	require.Greater(t, result.Seconds, int64(1577836800))
}

func TestGogoTimestampNow_NanosValid(t *testing.T) {
	result := gogoTimestampNow()

	require.GreaterOrEqual(t, result.Nanos, int32(0))
	require.Less(t, result.Nanos, int32(1000000000))
}

// =============================================================================
// timeToGogoTimestamp Tests
// =============================================================================

func TestTimeToGogoTimestamp_UnixEpoch(t *testing.T) {
	epoch := time.Unix(0, 0).UTC()
	result := timeToGogoTimestamp(epoch)

	require.NotNil(t, result)
	require.Equal(t, int64(0), result.Seconds)
	require.Equal(t, int32(0), result.Nanos)
}

func TestTimeToGogoTimestamp_SpecificTime(t *testing.T) {
	// January 1, 2024, 12:00:00.500 UTC
	tm := time.Date(2024, 1, 1, 12, 0, 0, 500000000, time.UTC)
	result := timeToGogoTimestamp(tm)

	require.NotNil(t, result)
	require.Equal(t, tm.Unix(), result.Seconds)
	require.Equal(t, int32(500000000), result.Nanos)
}

func TestTimeToGogoTimestamp_ZeroTime(t *testing.T) {
	var zeroTime time.Time
	result := timeToGogoTimestamp(zeroTime)

	require.NotNil(t, result)
	// Go's zero time is "0001-01-01 00:00:00 +0000 UTC"
	// which is before Unix epoch, so Seconds will be negative
	require.Less(t, result.Seconds, int64(0))
}

func TestTimeToGogoTimestamp_WithNanoseconds(t *testing.T) {
	tm := time.Unix(1704067200, 123456789)
	result := timeToGogoTimestamp(tm)

	require.NotNil(t, result)
	require.Equal(t, int64(1704067200), result.Seconds)
	require.Equal(t, int32(123456789), result.Nanos)
}

func TestTimeToGogoTimestamp_MaxTime(t *testing.T) {
	// Test with a far future time
	tm := time.Date(2200, 12, 31, 23, 59, 59, 999999999, time.UTC)
	result := timeToGogoTimestamp(tm)

	require.NotNil(t, result)
	require.Equal(t, tm.Unix(), result.Seconds)
	require.Equal(t, int32(999999999), result.Nanos)
}

// =============================================================================
// blockTimeToGogoTimestamp Tests (requires context)
// =============================================================================

func (suite *ProtoHelpersTestSuite) TestBlockTimeToGogoTimestamp_Valid() {
	ctx := suite.GetContext()
	result := blockTimeToGogoTimestamp(ctx)

	suite.Require().NotNil(result)
	// Should be a valid timestamp
	suite.Require().GreaterOrEqual(result.Nanos, int32(0))
	suite.Require().Less(result.Nanos, int32(1000000000))
}

// =============================================================================
// blockTimeWithOffsetToGogoTimestamp Tests (requires context)
// =============================================================================

func (suite *ProtoHelpersTestSuite) TestBlockTimeWithOffsetToGogoTimestamp_PositiveOffset() {
	ctx := suite.GetContext()
	baseTime := blockTimeToGogoTimestamp(ctx)
	offsetResult := blockTimeWithOffsetToGogoTimestamp(ctx, 1*time.Hour)

	suite.Require().NotNil(offsetResult)
	// Offset should be 1 hour (3600 seconds) later
	suite.Require().Equal(baseTime.Seconds+3600, offsetResult.Seconds)
}

func (suite *ProtoHelpersTestSuite) TestBlockTimeWithOffsetToGogoTimestamp_NegativeOffset() {
	ctx := suite.GetContext()
	baseTime := blockTimeToGogoTimestamp(ctx)
	offsetResult := blockTimeWithOffsetToGogoTimestamp(ctx, -30*time.Minute)

	suite.Require().NotNil(offsetResult)
	// Offset should be 30 minutes (1800 seconds) earlier
	suite.Require().Equal(baseTime.Seconds-1800, offsetResult.Seconds)
}

func (suite *ProtoHelpersTestSuite) TestBlockTimeWithOffsetToGogoTimestamp_ZeroOffset() {
	ctx := suite.GetContext()
	baseTime := blockTimeToGogoTimestamp(ctx)
	offsetResult := blockTimeWithOffsetToGogoTimestamp(ctx, 0)

	suite.Require().NotNil(offsetResult)
	suite.Require().Equal(baseTime.Seconds, offsetResult.Seconds)
}

func (suite *ProtoHelpersTestSuite) TestBlockTimeWithOffsetToGogoTimestamp_LargeOffset() {
	ctx := suite.GetContext()
	baseTime := blockTimeToGogoTimestamp(ctx)
	// 365 days offset
	offsetResult := blockTimeWithOffsetToGogoTimestamp(ctx, 365*24*time.Hour)

	suite.Require().NotNil(offsetResult)
	// Should be approximately 365 days later
	expectedSeconds := baseTime.Seconds + (365 * 24 * 60 * 60)
	suite.Require().Equal(expectedSeconds, offsetResult.Seconds)
}

// =============================================================================
// gogoTimestampToTime Tests
// =============================================================================

func TestGogoTimestampToTime_NilInput(t *testing.T) {
	result := gogoTimestampToTime(nil)

	require.True(t, result.IsZero(), "nil input should return zero time")
}

func TestGogoTimestampToTime_ValidTimestamp(t *testing.T) {
	ts := &gogotypes.Timestamp{
		Seconds: 1704067200,
		Nanos:   500000000,
	}

	result := gogoTimestampToTime(ts)

	require.Equal(t, int64(1704067200), result.Unix())
	require.Equal(t, 500000000, result.Nanosecond())
}

func TestGogoTimestampToTime_UnixEpoch(t *testing.T) {
	ts := &gogotypes.Timestamp{
		Seconds: 0,
		Nanos:   0,
	}

	result := gogoTimestampToTime(ts)

	require.Equal(t, int64(0), result.Unix())
	require.Equal(t, 0, result.Nanosecond())
}

func TestGogoTimestampToTime_NegativeSeconds(t *testing.T) {
	ts := &gogotypes.Timestamp{
		Seconds: -86400, // 1 day before Unix epoch
		Nanos:   0,
	}

	result := gogoTimestampToTime(ts)

	require.Equal(t, int64(-86400), result.Unix())
}

func TestGogoTimestampToTime_RoundTrip(t *testing.T) {
	originalTime := time.Date(2024, 6, 15, 10, 30, 45, 123456789, time.UTC)
	gogoTs := timeToGogoTimestamp(originalTime)
	resultTime := gogoTimestampToTime(gogoTs)

	require.Equal(t, originalTime.Unix(), resultTime.Unix())
	require.Equal(t, originalTime.Nanosecond(), resultTime.Nanosecond())
}

// =============================================================================
// gogoDurationToTime Tests
// =============================================================================

func TestGogoDurationToTime_NilInput(t *testing.T) {
	result := gogoDurationToTime(nil)

	require.Equal(t, time.Duration(0), result)
}

func TestGogoDurationToTime_ZeroDuration(t *testing.T) {
	d := &gogotypes.Duration{
		Seconds: 0,
		Nanos:   0,
	}

	result := gogoDurationToTime(d)

	require.Equal(t, time.Duration(0), result)
}

func TestGogoDurationToTime_SecondsOnly(t *testing.T) {
	d := &gogotypes.Duration{
		Seconds: 3600, // 1 hour
		Nanos:   0,
	}

	result := gogoDurationToTime(d)

	require.Equal(t, 1*time.Hour, result)
}

func TestGogoDurationToTime_NanosOnly(t *testing.T) {
	d := &gogotypes.Duration{
		Seconds: 0,
		Nanos:   500000000, // 500ms
	}

	result := gogoDurationToTime(d)

	require.Equal(t, 500*time.Millisecond, result)
}

func TestGogoDurationToTime_SecondsAndNanos(t *testing.T) {
	d := &gogotypes.Duration{
		Seconds: 60,
		Nanos:   500000000, // 500ms
	}

	result := gogoDurationToTime(d)

	expected := 60*time.Second + 500*time.Millisecond
	require.Equal(t, expected, result)
}

func TestGogoDurationToTime_LargeDuration(t *testing.T) {
	d := &gogotypes.Duration{
		Seconds: 86400, // 24 hours
		Nanos:   999999999,
	}

	result := gogoDurationToTime(d)

	expected := 24*time.Hour + 999999999*time.Nanosecond
	require.Equal(t, expected, result)
}

func TestGogoDurationToTime_NegativeDuration(t *testing.T) {
	d := &gogotypes.Duration{
		Seconds: -60,
		Nanos:   0,
	}

	result := gogoDurationToTime(d)

	require.Equal(t, -60*time.Second, result)
}

// =============================================================================
// recordAnomaly Tests (requires keeper context)
// =============================================================================

func (suite *ProtoHelpersTestSuite) TestRecordAnomaly_Success() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	score := &AnomalyScore{
		Score:     8000, // 80% (above 70% threshold)
		Threshold: AnomalyThresholdBPS,
		IsAnomaly: true,
		Factors: map[string]uint64{
			"amount":    3000,
			"recipient": 5000,
			"frequency": 0,
			"time":      0,
		},
	}

	err := k.recordAnomaly(ctx, "wallet-test-anomaly", score)
	suite.Require().NoError(err)

	// Verify anomaly was stored
	anomalies, err := k.GetAnomalies(ctx, "wallet-test-anomaly")
	suite.Require().NoError(err)
	suite.Require().Len(anomalies, 1)
	suite.Require().Equal("wallet-test-anomaly", anomalies[0].WalletId)
	suite.Require().False(anomalies[0].Resolved)
}

func (suite *ProtoHelpersTestSuite) TestRecordAnomaly_SameTimestampOverwrites() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	walletID := "wallet-overwrite-anomaly"

	// Record first anomaly
	score1 := &AnomalyScore{
		Score:     8000,
		Threshold: AnomalyThresholdBPS,
		IsAnomaly: true,
		Factors:   map[string]uint64{"amount": 8000},
	}
	err := k.recordAnomaly(ctx, walletID, score1)
	suite.Require().NoError(err)

	// Record second anomaly at same block time - should overwrite
	// (because key is based on walletID + UnixNano timestamp)
	score2 := &AnomalyScore{
		Score:     9000,
		Threshold: AnomalyThresholdBPS,
		IsAnomaly: true,
		Factors:   map[string]uint64{"recipient": 9000},
	}
	err = k.recordAnomaly(ctx, walletID, score2)
	suite.Require().NoError(err)

	// Only one anomaly stored (second overwrites first due to same timestamp key)
	anomalies, err := k.GetAnomalies(ctx, walletID)
	suite.Require().NoError(err)
	suite.Require().Len(anomalies, 1)
	// Score should be from the second (overwriting) anomaly
	suite.Require().InDelta(0.90, anomalies[0].Score, 0.01)
}

func (suite *ProtoHelpersTestSuite) TestRecordAnomaly_DifferentWallets() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Record anomalies for different wallets
	score := &AnomalyScore{
		Score:     8500,
		Threshold: AnomalyThresholdBPS,
		IsAnomaly: true,
		Factors:   map[string]uint64{"amount": 8500},
	}

	err := k.recordAnomaly(ctx, "wallet-A", score)
	suite.Require().NoError(err)

	err = k.recordAnomaly(ctx, "wallet-B", score)
	suite.Require().NoError(err)

	// Each wallet should have its own anomaly
	anomaliesA, err := k.GetAnomalies(ctx, "wallet-A")
	suite.Require().NoError(err)
	suite.Require().Len(anomaliesA, 1)
	suite.Require().Equal("wallet-A", anomaliesA[0].WalletId)

	anomaliesB, err := k.GetAnomalies(ctx, "wallet-B")
	suite.Require().NoError(err)
	suite.Require().Len(anomaliesB, 1)
	suite.Require().Equal("wallet-B", anomaliesB[0].WalletId)
}

func (suite *ProtoHelpersTestSuite) TestRecordAnomaly_EmptyWalletID() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	score := &AnomalyScore{
		Score:     8000,
		Threshold: AnomalyThresholdBPS,
		IsAnomaly: true,
		Factors:   map[string]uint64{},
	}

	// Should not error with empty wallet ID (just stores with empty key component)
	err := k.recordAnomaly(ctx, "", score)
	suite.Require().NoError(err)
}

func (suite *ProtoHelpersTestSuite) TestRecordAnomaly_ScoreConversion() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Test basis points to float conversion
	score := &AnomalyScore{
		Score:     5555, // 55.55%
		Threshold: 7000, // 70.00%
		IsAnomaly: false,
		Factors:   map[string]uint64{"test": 5555},
	}

	err := k.recordAnomaly(ctx, "wallet-score-convert", score)
	suite.Require().NoError(err)

	anomalies, err := k.GetAnomalies(ctx, "wallet-score-convert")
	suite.Require().NoError(err)
	suite.Require().Len(anomalies, 1)

	// Score should be converted from 5555 basis points to 0.5555
	suite.Require().InDelta(0.5555, anomalies[0].Score, 0.0001)
	suite.Require().InDelta(0.70, anomalies[0].Threshold, 0.0001)
}

// =============================================================================
// hasPrefix Tests (utility function)
// =============================================================================

func TestHasPrefix_EmptyKeyEmptyPrefix(t *testing.T) {
	result := hasPrefix([]byte{}, []byte{})
	require.True(t, result, "empty prefix matches empty key")
}

func TestHasPrefix_NonEmptyKeyEmptyPrefix(t *testing.T) {
	result := hasPrefix([]byte("some_key"), []byte{})
	require.True(t, result, "empty prefix matches any key")
}

func TestHasPrefix_EmptyKeyNonEmptyPrefix(t *testing.T) {
	result := hasPrefix([]byte{}, []byte("prefix_"))
	require.False(t, result, "non-empty prefix cannot match empty key")
}

func TestHasPrefix_ExactMatch(t *testing.T) {
	result := hasPrefix([]byte("prefix_"), []byte("prefix_"))
	require.True(t, result, "exact match should return true")
}

func TestHasPrefix_ValidPrefix(t *testing.T) {
	result := hasPrefix([]byte("prefix_key_data"), []byte("prefix_"))
	require.True(t, result)
}

func TestHasPrefix_InvalidPrefix(t *testing.T) {
	result := hasPrefix([]byte("key_data"), []byte("prefix_"))
	require.False(t, result)
}

func TestHasPrefix_LongerPrefixThanKey(t *testing.T) {
	result := hasPrefix([]byte("ab"), []byte("abcdef"))
	require.False(t, result, "prefix longer than key should return false")
}

func TestHasPrefix_BinaryData(t *testing.T) {
	key := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	prefix := []byte{0x01, 0x02, 0x03}
	result := hasPrefix(key, prefix)
	require.True(t, result)
}

func TestHasPrefix_BinaryDataNoMatch(t *testing.T) {
	key := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	prefix := []byte{0x01, 0x02, 0x04} // Different middle byte
	result := hasPrefix(key, prefix)
	require.False(t, result)
}

// =============================================================================
// clampUint64ToInt64 Tests
// =============================================================================

func TestClampUint64ToInt64_Zero(t *testing.T) {
	result := clampUint64ToInt64(0)
	require.Equal(t, int64(0), result)
}

func TestClampUint64ToInt64_SmallValue(t *testing.T) {
	result := clampUint64ToInt64(1000)
	require.Equal(t, int64(1000), result)
}

func TestClampUint64ToInt64_MaxInt64(t *testing.T) {
	result := clampUint64ToInt64(uint64(math.MaxInt64))
	require.Equal(t, int64(math.MaxInt64), result)
}

func TestClampUint64ToInt64_OverflowClamp(t *testing.T) {
	// Value greater than MaxInt64 should be clamped
	result := clampUint64ToInt64(uint64(math.MaxInt64) + 1)
	require.Equal(t, int64(math.MaxInt64), result)
}

func TestClampUint64ToInt64_MaxUint64(t *testing.T) {
	result := clampUint64ToInt64(math.MaxUint64)
	require.Equal(t, int64(math.MaxInt64), result)
}

func TestClampUint64ToInt64_JustBelowMax(t *testing.T) {
	result := clampUint64ToInt64(uint64(math.MaxInt64) - 1)
	require.Equal(t, int64(math.MaxInt64)-1, result)
}

// =============================================================================
// parseAmountString Tests (utility function from keeper.go)
// =============================================================================

func TestParseAmountString_EmptyString(t *testing.T) {
	result, err := parseAmountString("")
	require.NoError(t, err)
	require.True(t, result.IsZero())
}

func TestParseAmountString_WhitespaceOnly(t *testing.T) {
	result, err := parseAmountString("   ")
	require.NoError(t, err)
	require.True(t, result.IsZero())
}

func TestParseAmountString_ValidNumber(t *testing.T) {
	result, err := parseAmountString("1000000")
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(1000000), result)
}

func TestParseAmountString_Zero(t *testing.T) {
	result, err := parseAmountString("0")
	require.NoError(t, err)
	require.True(t, result.IsZero())
}

func TestParseAmountString_NegativeNumber(t *testing.T) {
	result, err := parseAmountString("-1000")
	require.NoError(t, err)
	require.True(t, result.IsNegative())
	require.Equal(t, sdkmath.NewInt(-1000), result)
}

func TestParseAmountString_LargeNumber(t *testing.T) {
	result, err := parseAmountString("9999999999999999999999")
	require.NoError(t, err)
	expected, _ := sdkmath.NewIntFromString("9999999999999999999999")
	require.Equal(t, expected, result)
}

func TestParseAmountString_InvalidFormat(t *testing.T) {
	_, err := parseAmountString("abc")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid amount")
}

func TestParseAmountString_DecimalNotSupported(t *testing.T) {
	// SDK integers don't support decimals
	_, err := parseAmountString("100.50")
	require.Error(t, err)
}

// =============================================================================
// containsString Tests (utility function from keeper.go)
// =============================================================================

func TestContainsString_EmptySlice(t *testing.T) {
	result := containsString([]string{}, "target")
	require.False(t, result)
}

func TestContainsString_NilSlice(t *testing.T) {
	result := containsString(nil, "target")
	require.False(t, result)
}

func TestContainsString_TargetFound(t *testing.T) {
	items := []string{"alpha", "beta", "gamma"}
	result := containsString(items, "beta")
	require.True(t, result)
}

func TestContainsString_TargetNotFound(t *testing.T) {
	items := []string{"alpha", "beta", "gamma"}
	result := containsString(items, "delta")
	require.False(t, result)
}

func TestContainsString_EmptyTarget(t *testing.T) {
	items := []string{"alpha", "", "gamma"}
	result := containsString(items, "")
	require.True(t, result)
}

func TestContainsString_FirstElement(t *testing.T) {
	items := []string{"target", "other", "items"}
	result := containsString(items, "target")
	require.True(t, result)
}

func TestContainsString_LastElement(t *testing.T) {
	items := []string{"other", "items", "target"}
	result := containsString(items, "target")
	require.True(t, result)
}

func TestContainsString_CaseSensitive(t *testing.T) {
	items := []string{"Alpha", "Beta", "Gamma"}
	result := containsString(items, "alpha")
	require.False(t, result, "search should be case-sensitive")
}

// =============================================================================
// AnomalyScore Tests
// =============================================================================

func TestAnomalyScore_Constants(t *testing.T) {
	// Verify basis point constants are properly defined
	require.Equal(t, uint64(10000), BasisPointsMax)
	require.Equal(t, uint64(7000), AnomalyThresholdBPS)
	require.Equal(t, uint64(3000), AmountWeightBPS)
	require.Equal(t, uint64(3000), RecipientWeightBPS)
	require.Equal(t, uint64(2000), FrequencyWeightBPS)
	require.Equal(t, uint64(2000), TimeWeightBPS)

	// Weights should sum to 10000 (100%)
	totalWeight := AmountWeightBPS + RecipientWeightBPS + FrequencyWeightBPS + TimeWeightBPS
	require.Equal(t, uint64(10000), totalWeight, "weights should sum to 100%")
}

func TestAnomalyScore_FactorScores(t *testing.T) {
	require.Equal(t, uint64(5000), NewRecipientScoreBPS)
	require.Equal(t, uint64(6000), UnusualTimeScoreBPS)
	require.Equal(t, uint64(8000), HighFrequencyScoreBPS)
	require.Equal(t, int64(10), HighFrequencyThreshold)
}

// =============================================================================
// checkUnusualTime Tests (via integration with context)
// =============================================================================

func (suite *ProtoHelpersTestSuite) TestCheckUnusualTime_ReturnsScore() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	score := k.checkUnusualTime(ctx, "wallet-time-test")

	// Score should be 0 or UnusualTimeScoreBPS depending on hour
	suite.Require().True(score == 0 || score == UnusualTimeScoreBPS)
}

// =============================================================================
// getAmountStatistics Tests
// =============================================================================

func (suite *ProtoHelpersTestSuite) TestGetAmountStatistics_ReturnsPlaceholder() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	avg, stdDev := k.getAmountStatistics(ctx, "wallet-stats-test")

	// Current implementation returns placeholder values
	suite.Require().Equal(sdkmath.NewInt(1000), avg)
	suite.Require().Equal(sdkmath.NewInt(500), stdDev)
}

// =============================================================================
// Integration Tests - Full Flow
// =============================================================================

func (suite *ProtoHelpersTestSuite) TestTimestampConversion_RoundTrip() {
	ctx := suite.GetContext()

	// Get block time as gogo timestamp
	gogoTs := blockTimeToGogoTimestamp(ctx)
	suite.Require().NotNil(gogoTs)

	// Convert back to time.Time
	resultTime := gogoTimestampToTime(gogoTs)

	// Convert again to gogo
	gogoTs2 := timeToGogoTimestamp(resultTime)

	// Should match original
	suite.Require().Equal(gogoTs.Seconds, gogoTs2.Seconds)
	suite.Require().Equal(gogoTs.Nanos, gogoTs2.Nanos)
}

func (suite *ProtoHelpersTestSuite) TestDurationConversion_RoundTrip() {
	originalDuration := 2*time.Hour + 30*time.Minute + 45*time.Second + 123*time.Millisecond

	gogoDur := &gogotypes.Duration{
		Seconds: int64(originalDuration.Truncate(time.Second).Seconds()),
		Nanos:   int32(originalDuration.Nanoseconds() % 1e9),
	}

	resultDuration := gogoDurationToTime(gogoDur)

	suite.Require().Equal(originalDuration, resultDuration)
}

func (suite *ProtoHelpersTestSuite) TestAnomalyDetection_FullFlow() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	walletID := "wallet-full-flow"
	recipient := "recipient-1"
	amount := sdkmath.NewInt(100000)

	// Detect anomaly
	score, err := k.DetectTransactionAnomaly(ctx, walletID, recipient, amount)
	suite.Require().NoError(err)
	suite.Require().NotNil(score)

	// Verify all factors are present
	suite.Require().Contains(score.Factors, "amount")
	suite.Require().Contains(score.Factors, "recipient")
	suite.Require().Contains(score.Factors, "frequency")
	suite.Require().Contains(score.Factors, "time")

	// Threshold should be set
	suite.Require().Equal(AnomalyThresholdBPS, score.Threshold)
}

// =============================================================================
// Edge Cases and Boundary Tests
// =============================================================================

func TestTimeToGogoTimestamp_NanoEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		nanos    int
		expected int32
	}{
		{"Zero nanoseconds", 0, 0},
		{"One nanosecond", 1, 1},
		{"One millisecond", 1000000, 1000000},
		{"One microsecond", 1000, 1000},
		{"Max valid nanoseconds", 999999999, 999999999},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tm := time.Unix(1704067200, int64(tc.nanos))
			result := timeToGogoTimestamp(tm)
			require.Equal(t, tc.expected, result.Nanos)
		})
	}
}

func (suite *ProtoHelpersTestSuite) TestBlockTimeWithOffset_EdgeCases() {
	ctx := suite.GetContext()

	testCases := []struct {
		name     string
		offset   time.Duration
		validate func(base, result *gogotypes.Timestamp)
	}{
		{
			name:   "One second offset",
			offset: time.Second,
			validate: func(base, result *gogotypes.Timestamp) {
				suite.Require().Equal(base.Seconds+1, result.Seconds)
			},
		},
		{
			name:   "One nanosecond offset",
			offset: time.Nanosecond,
			validate: func(base, result *gogotypes.Timestamp) {
				// Either nanos increased by 1 or seconds increased with nanos wrapped
				totalNanosBase := base.Seconds*1e9 + int64(base.Nanos)
				totalNanosResult := result.Seconds*1e9 + int64(result.Nanos)
				suite.Require().Equal(totalNanosBase+1, totalNanosResult)
			},
		},
		{
			name:   "Negative one second offset",
			offset: -time.Second,
			validate: func(base, result *gogotypes.Timestamp) {
				suite.Require().Equal(base.Seconds-1, result.Seconds)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			baseTime := blockTimeToGogoTimestamp(ctx)
			offsetResult := blockTimeWithOffsetToGogoTimestamp(ctx, tc.offset)
			tc.validate(baseTime, offsetResult)
		})
	}
}
