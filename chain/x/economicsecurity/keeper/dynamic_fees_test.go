// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// =============================================================================
// RecordBlockUtilization Tests
// =============================================================================

func TestRecordBlockUtilization(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Record some utilization data
	err := k.RecordBlockUtilization(ctx, 7500)
	require.NoError(t, err)

	err = k.RecordBlockUtilization(ctx, 8000)
	require.NoError(t, err)

	// Verify utilization was recorded
	params, _ := k.GetParams(ctx)
	require.Len(t, params.DynamicFees.RecentUtilization, 2)
	require.Equal(t, uint64(7500), params.DynamicFees.RecentUtilization[0])
	require.Equal(t, uint64(8000), params.DynamicFees.RecentUtilization[1])
}

func TestRecordBlockUtilization_WindowTrimming(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Get window size from params
	params, _ := k.GetParams(ctx)
	windowSize := params.DynamicFees.UtilizationWindow

	// Record more utilization data than the window size
	for i := uint64(0); i < windowSize+5; i++ {
		err := k.RecordBlockUtilization(ctx, i*100)
		require.NoError(t, err)
	}

	// Verify window was trimmed
	updatedParams, _ := k.GetParams(ctx)
	require.Len(t, updatedParams.DynamicFees.RecentUtilization, int(windowSize))

	// First entries should be trimmed, so first value should be 500 (index 5 * 100)
	require.Equal(t, uint64(500), updatedParams.DynamicFees.RecentUtilization[0])
}

// =============================================================================
// AdjustDynamicFees Tests
// =============================================================================

func TestAdjustDynamicFees_Disabled(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Disable dynamic fees
	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = false
	require.NoError(t, k.SetParams(params))

	// Should return without error when disabled
	err := k.AdjustDynamicFees(ctx)
	require.NoError(t, err)
}

func TestAdjustDynamicFees_NoData(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Clear utilization data
	params, _ := k.GetParams(ctx)
	params.DynamicFees.RecentUtilization = []uint64{}
	require.NoError(t, k.SetParams(params))

	// Should return without error when no data
	err := k.AdjustDynamicFees(ctx)
	require.NoError(t, err)
}

func TestAdjustDynamicFees_IncreaseMultiplier(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Set utilization above target (7000) to trigger increase
	params, _ := k.GetParams(ctx)
	params.DynamicFees.RecentUtilization = []uint64{9000, 9500, 9000}
	params.DynamicFees.CurrentMultiplier = 10000
	require.NoError(t, k.SetParams(params))

	initialMultiplier := params.DynamicFees.CurrentMultiplier

	err := k.AdjustDynamicFees(ctx)
	require.NoError(t, err)

	// Verify multiplier increased
	updatedParams, _ := k.GetParams(ctx)
	require.Greater(t, updatedParams.DynamicFees.CurrentMultiplier, initialMultiplier)
}

func TestAdjustDynamicFees_DecreaseMultiplier(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Set utilization below target (7000) to trigger decrease
	params, _ := k.GetParams(ctx)
	params.DynamicFees.RecentUtilization = []uint64{3000, 2500, 3000}
	params.DynamicFees.CurrentMultiplier = 10000
	require.NoError(t, k.SetParams(params))

	initialMultiplier := params.DynamicFees.CurrentMultiplier

	err := k.AdjustDynamicFees(ctx)
	require.NoError(t, err)

	// Verify multiplier decreased
	updatedParams, _ := k.GetParams(ctx)
	require.Less(t, updatedParams.DynamicFees.CurrentMultiplier, initialMultiplier)
}

func TestAdjustDynamicFees_ClampToMin(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Set very low utilization with already low multiplier
	params, _ := k.GetParams(ctx)
	params.DynamicFees.RecentUtilization = []uint64{100, 50, 100}
	params.DynamicFees.CurrentMultiplier = 5100 // Just above min (5000)
	params.DynamicFees.AdjustmentSpeed = 10000   // High adjustment speed
	require.NoError(t, k.SetParams(params))

	err := k.AdjustDynamicFees(ctx)
	require.NoError(t, err)

	// Verify multiplier was clamped to min
	updatedParams, _ := k.GetParams(ctx)
	require.GreaterOrEqual(t, updatedParams.DynamicFees.CurrentMultiplier, params.DynamicFees.MinMultiplier)
}

func TestAdjustDynamicFees_ClampToMax(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Set very high utilization with already high multiplier
	params, _ := k.GetParams(ctx)
	params.DynamicFees.RecentUtilization = []uint64{9900, 9950, 9900}
	params.DynamicFees.CurrentMultiplier = 19500 // Just below max (20000)
	params.DynamicFees.AdjustmentSpeed = 10000   // High adjustment speed
	require.NoError(t, k.SetParams(params))

	err := k.AdjustDynamicFees(ctx)
	require.NoError(t, err)

	// Verify multiplier was clamped to max
	updatedParams, _ := k.GetParams(ctx)
	require.LessOrEqual(t, updatedParams.DynamicFees.CurrentMultiplier, params.DynamicFees.MaxMultiplier)
}

// =============================================================================
// CalculateDynamicFee Tests
// =============================================================================

func TestCalculateDynamicFee_Enabled(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Set up params with known values
	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = true
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 15000 // 1.5x
	require.NoError(t, k.SetParams(params))

	fee := k.CalculateDynamicFee(ctx)
	require.Equal(t, "1500", fee) // 1000 * 15000 / 10000 = 1500
}

func TestCalculateDynamicFee_Disabled(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Disable dynamic fees
	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = false
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 15000
	require.NoError(t, k.SetParams(params))

	fee := k.CalculateDynamicFee(ctx)
	require.Equal(t, "1000", fee) // Returns base fee when disabled
}

func TestCalculateDynamicFee_InvalidBaseFee(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Set invalid base fee
	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = true
	params.DynamicFees.BaseFee = "not-a-number"
	require.NoError(t, k.SetParams(params))

	fee := k.CalculateDynamicFee(ctx)
	require.Equal(t, "not-a-number", fee) // Returns base fee on parse error
}

// =============================================================================
// GetCurrentFeeMultiplier Tests
// =============================================================================

func TestGetCurrentFeeMultiplier(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.CurrentMultiplier = 12500
	require.NoError(t, k.SetParams(params))

	multiplier := k.GetCurrentFeeMultiplier(ctx)
	require.Equal(t, uint64(12500), multiplier)
}

// =============================================================================
// GetAverageUtilization Tests
// =============================================================================

func TestGetAverageUtilization_Empty(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Clear utilization data
	params, _ := k.GetParams(ctx)
	params.DynamicFees.RecentUtilization = []uint64{}
	require.NoError(t, k.SetParams(params))

	avg := k.GetAverageUtilization(ctx)
	require.Equal(t, uint64(0), avg)
}

func TestGetAverageUtilization_WithData(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.RecentUtilization = []uint64{6000, 8000, 7000, 9000}
	require.NoError(t, k.SetParams(params))

	avg := k.GetAverageUtilization(ctx)
	require.Equal(t, uint64(7500), avg) // (6000+8000+7000+9000)/4 = 7500
}

// =============================================================================
// GetFeeMultiplierHistory Tests
// =============================================================================

func TestGetFeeMultiplierHistory(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.RecentUtilization = []uint64{5000, 6000, 7000}
	require.NoError(t, k.SetParams(params))

	history := k.GetFeeMultiplierHistory(ctx)
	require.Len(t, history, 3)
	require.Equal(t, []uint64{5000, 6000, 7000}, history)
}

// =============================================================================
// PredictFee Tests
// =============================================================================

func TestPredictFee_Disabled(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = false
	params.DynamicFees.BaseFee = "500"
	require.NoError(t, k.SetParams(params))

	predictedFee, err := k.PredictFee(ctx, 8000)
	require.NoError(t, err)
	require.Equal(t, "500", predictedFee)
}

func TestPredictFee_HighUtilization(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = true
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 10000
	params.DynamicFees.TargetUtilization = 7000
	require.NoError(t, k.SetParams(params))

	// Predict fee at high utilization (9000 vs target 7000)
	predictedFee, err := k.PredictFee(ctx, 9000)
	require.NoError(t, err)
	require.NotEmpty(t, predictedFee)
}

func TestPredictFee_LowUtilization(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = true
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 10000
	params.DynamicFees.TargetUtilization = 7000
	require.NoError(t, k.SetParams(params))

	// Predict fee at low utilization
	predictedFee, err := k.PredictFee(ctx, 3000)
	require.NoError(t, err)
	require.NotEmpty(t, predictedFee)
}

func TestPredictFee_InvalidBaseFee(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = true
	params.DynamicFees.BaseFee = "invalid"
	require.NoError(t, k.SetParams(params))

	_, err := k.PredictFee(ctx, 5000)
	require.Error(t, err)
	require.Equal(t, types.ErrInvalidAmount, err)
}

// =============================================================================
// SetBaseFee Tests
// =============================================================================

func TestSetBaseFee_Valid(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	err := k.SetBaseFee(ctx, "2000")
	require.NoError(t, err)

	params, _ := k.GetParams(ctx)
	require.Equal(t, "2000", params.DynamicFees.BaseFee)
}

func TestSetBaseFee_Invalid(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	err := k.SetBaseFee(ctx, "not-a-number")
	require.Error(t, err)
	require.Equal(t, types.ErrInvalidAmount, err)
}

func TestSetBaseFee_Zero(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	err := k.SetBaseFee(ctx, "0")
	require.Error(t, err)
	require.Equal(t, types.ErrInvalidAmount, err)
}

func TestSetBaseFee_Negative(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	err := k.SetBaseFee(ctx, "-100")
	require.Error(t, err)
	require.Equal(t, types.ErrInvalidAmount, err)
}

// =============================================================================
// ResetFeeMultiplier Tests
// =============================================================================

func TestResetFeeMultiplier(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Set a non-default multiplier
	params, _ := k.GetParams(ctx)
	params.DynamicFees.CurrentMultiplier = 15000
	require.NoError(t, k.SetParams(params))

	// Reset
	err := k.ResetFeeMultiplier(ctx)
	require.NoError(t, err)

	// Verify reset to base value (10000)
	updatedParams, _ := k.GetParams(ctx)
	require.Equal(t, uint64(types.BasisPoints), updatedParams.DynamicFees.CurrentMultiplier)
}

// =============================================================================
// GetFeeStatistics Tests
// =============================================================================

func TestGetFeeStatistics(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Set up known values
	params, _ := k.GetParams(ctx)
	params.DynamicFees.Enabled = true
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 12000
	params.DynamicFees.RecentUtilization = []uint64{6000, 8000}
	require.NoError(t, k.SetParams(params))

	currentFee, multiplier, avgUtil, baseFee := k.GetFeeStatistics(ctx)

	require.Equal(t, "1200", currentFee) // 1000 * 12000 / 10000
	require.Equal(t, uint64(12000), multiplier)
	require.Equal(t, uint64(7000), avgUtil) // (6000+8000)/2
	require.Equal(t, "1000", baseFee)
}
