// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// =============================================================================
// PredictGasPrice Tests
// =============================================================================

func TestPredictGasPrice_Disabled(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = false
	params.DynamicFees.BaseFee = "500"
	require.NoError(t, k.SetParams(params))

	price, confidence, err := k.PredictGasPrice(ctx, 10)
	require.NoError(t, err)
	require.Equal(t, "500", price)
	require.Equal(t, uint64(0), confidence)
}

func TestPredictGasPrice_NoData(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = true
	params.DynamicFees.RecentUtilization = []uint64{}
	params.DynamicFees.BaseFee = "1000"
	require.NoError(t, k.SetParams(params))

	price, confidence, err := k.PredictGasPrice(ctx, 10)
	require.NoError(t, err)
	require.Equal(t, "1000", price)
	require.Equal(t, uint64(0), confidence)
}

func TestPredictGasPrice_StableUtilization(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Set stable utilization data (no variance)
	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = true
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 10000
	params.DynamicFees.RecentUtilization = []uint64{7000, 7000, 7000, 7000, 7000}
	require.NoError(t, k.SetParams(params))

	price, confidence, err := k.PredictGasPrice(ctx, 10)
	require.NoError(t, err)
	require.NotEmpty(t, price)
	// With stable data, should have medium-high confidence
	require.Greater(t, confidence, uint64(5000))
}

func TestPredictGasPrice_IncreasingTrend(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Set increasing utilization trend
	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = true
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 10000
	params.DynamicFees.TargetUtilization = 7000
	params.DynamicFees.RecentUtilization = []uint64{5000, 6000, 7000, 8000, 9000}
	require.NoError(t, k.SetParams(params))

	price, _, err := k.PredictGasPrice(ctx, 10)
	require.NoError(t, err)
	require.NotEmpty(t, price)
}

func TestPredictGasPrice_DecreasingTrend(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Set decreasing utilization trend
	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = true
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 10000
	params.DynamicFees.TargetUtilization = 7000
	params.DynamicFees.RecentUtilization = []uint64{9000, 8000, 7000, 6000, 5000}
	require.NoError(t, k.SetParams(params))

	price, _, err := k.PredictGasPrice(ctx, 10)
	require.NoError(t, err)
	require.NotEmpty(t, price)
}

func TestPredictGasPrice_InvalidBaseFee(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = true
	params.DynamicFees.BaseFee = "invalid"
	params.DynamicFees.RecentUtilization = []uint64{5000, 6000, 7000}
	require.NoError(t, k.SetParams(params))

	_, _, err := k.PredictGasPrice(ctx, 10)
	require.Error(t, err)
	require.Equal(t, types.ErrInvalidAmount, err)
}

// =============================================================================
// GetGasPredictionStatistics Tests
// =============================================================================

func TestGetGasPredictionStatistics(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = true
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 10000
	params.DynamicFees.RecentUtilization = []uint64{7000, 7000, 7000, 7000, 7000}
	require.NoError(t, k.SetParams(params))

	price1, price10, price100, avgConf, err := k.GetGasPredictionStatistics(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, price1)
	require.NotEmpty(t, price10)
	require.NotEmpty(t, price100)
	require.Greater(t, avgConf, uint64(0))
}

func TestGetGasPredictionStatistics_ErrorPropagation(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = true
	params.DynamicFees.BaseFee = "invalid"
	params.DynamicFees.RecentUtilization = []uint64{5000, 6000}
	require.NoError(t, k.SetParams(params))

	_, _, _, _, err := k.GetGasPredictionStatistics(ctx)
	require.Error(t, err)
}

// =============================================================================
// GetRecommendedGasPrice Tests
// =============================================================================

func TestGetRecommendedGasPrice_Low(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 10000
	require.NoError(t, k.SetParams(params))

	price, err := k.GetRecommendedGasPrice(ctx, "low")
	require.NoError(t, err)
	require.Equal(t, "900", price) // 90% of base
}

func TestGetRecommendedGasPrice_Medium(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 10000
	require.NoError(t, k.SetParams(params))

	price, err := k.GetRecommendedGasPrice(ctx, "medium")
	require.NoError(t, err)
	require.Equal(t, "1000", price) // 100% of base
}

func TestGetRecommendedGasPrice_High(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 10000
	require.NoError(t, k.SetParams(params))

	price, err := k.GetRecommendedGasPrice(ctx, "high")
	require.NoError(t, err)
	require.Equal(t, "1100", price) // 110% of base
}

func TestGetRecommendedGasPrice_Urgent(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 10000
	require.NoError(t, k.SetParams(params))

	price, err := k.GetRecommendedGasPrice(ctx, "urgent")
	require.NoError(t, err)
	require.Equal(t, "1500", price) // 150% of base
}

func TestGetRecommendedGasPrice_InvalidPriority(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	_, err := k.GetRecommendedGasPrice(ctx, "invalid")
	require.Error(t, err)
	require.Equal(t, types.ErrInvalidPriority, err)
}

func TestGetRecommendedGasPrice_ClampToMax(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 19000 // Close to max
	params.DynamicFees.MaxMultiplier = 20000
	require.NoError(t, k.SetParams(params))

	// Urgent would try to go to 150% (28500), should be clamped to max (20000)
	price, err := k.GetRecommendedGasPrice(ctx, "urgent")
	require.NoError(t, err)
	require.Equal(t, "2000", price) // Clamped to max: 1000 * 20000 / 10000
}

func TestGetRecommendedGasPrice_ClampToMin(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 5500 // Close to min
	params.DynamicFees.MinMultiplier = 5000
	require.NoError(t, k.SetParams(params))

	// Low would try to go to 90% (4950), should be clamped to min (5000)
	price, err := k.GetRecommendedGasPrice(ctx, "low")
	require.NoError(t, err)
	require.Equal(t, "500", price) // Clamped to min: 1000 * 5000 / 10000
}

func TestGetRecommendedGasPrice_InvalidBaseFee(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.BaseFee = "invalid"
	require.NoError(t, k.SetParams(params))

	_, err := k.GetRecommendedGasPrice(ctx, "medium")
	require.Error(t, err)
	require.Equal(t, types.ErrInvalidAmount, err)
}

// =============================================================================
// EstimateTransactionCost Tests
// =============================================================================

func TestEstimateTransactionCost(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = true
	params.DynamicFees.BaseFee = "100"
	params.DynamicFees.CurrentMultiplier = 10000
	params.DynamicFees.RecentUtilization = []uint64{7000, 7000, 7000, 7000, 7000}
	require.NoError(t, k.SetParams(params))

	cost, err := k.EstimateTransactionCost(ctx, 100000, 5)
	require.NoError(t, err)
	require.NotEmpty(t, cost)
}

func TestEstimateTransactionCost_LowConfidence(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Set highly variable utilization for low confidence
	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = true
	params.DynamicFees.BaseFee = "100"
	params.DynamicFees.CurrentMultiplier = 10000
	params.DynamicFees.RecentUtilization = []uint64{1000, 9000, 2000, 8000}
	require.NoError(t, k.SetParams(params))

	cost, err := k.EstimateTransactionCost(ctx, 100000, 5)
	require.NoError(t, err)
	require.NotEmpty(t, cost)
}

func TestEstimateTransactionCost_InvalidPrice(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = true
	params.DynamicFees.BaseFee = "invalid"
	params.DynamicFees.RecentUtilization = []uint64{5000, 6000}
	require.NoError(t, k.SetParams(params))

	_, err := k.EstimateTransactionCost(ctx, 100000, 5)
	require.Error(t, err)
}

// =============================================================================
// GetGasPriceTrend Tests
// =============================================================================

func TestGetGasPriceTrend_InsufficientData(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.RecentUtilization = []uint64{5000, 6000} // Only 2 points
	require.NoError(t, k.SetParams(params))

	direction, strength := k.GetGasPriceTrend(ctx)
	require.Equal(t, "stable", direction)
	require.Equal(t, uint64(0), strength)
}

func TestGetGasPriceTrend_Increasing(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Clear increase in utilization
	params, _ := k.GetParams(ctx)
	params.DynamicFees.RecentUtilization = []uint64{3000, 3000, 3000, 8000, 8000, 8000}
	require.NoError(t, k.SetParams(params))

	direction, strength := k.GetGasPriceTrend(ctx)
	require.Equal(t, "increasing", direction)
	require.Greater(t, strength, uint64(0))
}

func TestGetGasPriceTrend_Decreasing(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Clear decrease in utilization
	params, _ := k.GetParams(ctx)
	params.DynamicFees.RecentUtilization = []uint64{8000, 8000, 8000, 3000, 3000, 3000}
	require.NoError(t, k.SetParams(params))

	direction, strength := k.GetGasPriceTrend(ctx)
	require.Equal(t, "decreasing", direction)
	require.Greater(t, strength, uint64(0))
}

func TestGetGasPriceTrend_Stable(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Stable utilization
	params, _ := k.GetParams(ctx)
	params.DynamicFees.RecentUtilization = []uint64{7000, 7000, 7000, 7000, 7000, 7000}
	require.NoError(t, k.SetParams(params))

	direction, strength := k.GetGasPriceTrend(ctx)
	require.Equal(t, "stable", direction)
	require.Equal(t, uint64(0), strength)
}

// =============================================================================
// GetOptimalSubmissionTime Tests
// =============================================================================

func TestGetOptimalSubmissionTime_Default(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = true
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 10000
	params.DynamicFees.RecentUtilization = []uint64{7000, 7000, 7000}
	require.NoError(t, k.SetParams(params))

	blocks, price, err := k.GetOptimalSubmissionTime(ctx, 0)
	require.NoError(t, err)
	require.NotEmpty(t, price)
	require.GreaterOrEqual(t, blocks, uint64(0))
}

func TestGetOptimalSubmissionTime_MaxCapped(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = true
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 10000
	params.DynamicFees.RecentUtilization = []uint64{7000, 7000, 7000}
	require.NoError(t, k.SetParams(params))

	// Try with huge max, should be capped
	blocks, price, err := k.GetOptimalSubmissionTime(ctx, 5000)
	require.NoError(t, err)
	require.NotEmpty(t, price)
	require.LessOrEqual(t, blocks, uint64(1000)) // Capped at 1000
}

func TestGetOptimalSubmissionTime_SmallWindow(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = true
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 10000
	params.DynamicFees.RecentUtilization = []uint64{7000, 7000, 7000}
	require.NoError(t, k.SetParams(params))

	blocks, price, err := k.GetOptimalSubmissionTime(ctx, 5)
	require.NoError(t, err)
	require.NotEmpty(t, price)
	require.LessOrEqual(t, blocks, uint64(5))
}

// =============================================================================
// intSqrt Tests (internal helper)
// =============================================================================

func TestIntSqrt(t *testing.T) {
	testCases := []struct {
		input    int64
		expected int64
	}{
		{0, 0},
		{1, 1},
		{4, 2},
		{9, 3},
		{16, 4},
		{25, 5},
		{100, 10},
		{-5, 0}, // Negative should return 0
	}

	for _, tc := range testCases {
		result := intSqrt(tc.input)
		require.Equal(t, tc.expected, result, "intSqrt(%d) should be %d", tc.input, tc.expected)
	}
}

func TestIntSqrt_NonPerfectSquare(t *testing.T) {
	// Non-perfect squares should return floor of sqrt
	require.Equal(t, int64(2), intSqrt(5))  // sqrt(5) ≈ 2.23
	require.Equal(t, int64(3), intSqrt(10)) // sqrt(10) ≈ 3.16
	require.Equal(t, int64(4), intSqrt(20)) // sqrt(20) ≈ 4.47
}
