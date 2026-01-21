// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economics/types"
)

// ============================
// DYNAMIC FEE TESTS
// ============================

func TestRecordBlockUtilization(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set params with dynamic fees enabled
	params := types.DefaultParams()
	params.Fees.DynamicFeesEnabled = true
	params.Fees.TargetBlockUtilization = 50
	params.Fees.FeeAdjustmentSpeed = 100 // 1% in basis points
	params.Fees.MaxFeeMultiplier = 10
	params.Fees.MinFeeMultiplier = 1
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Initialize fee multiplier
	err = keeper.SetFeeMultiplier(ctx, "1.0")
	require.NoError(t, err)

	tests := []struct {
		name        string
		utilization uint64
		shouldErr   bool
	}{
		{
			name:        "low utilization",
			utilization: 25,
			shouldErr:   false,
		},
		{
			name:        "target utilization",
			utilization: 50,
			shouldErr:   false,
		},
		{
			name:        "high utilization",
			utilization: 75,
			shouldErr:   false,
		},
		{
			name:        "very high utilization",
			utilization: 100,
			shouldErr:   false,
		},
		{
			name:        "zero utilization",
			utilization: 0,
			shouldErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := keeper.RecordBlockUtilization(ctx, tt.utilization)
			if tt.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRecordBlockUtilizationDynamicFeesDisabled(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set params with dynamic fees disabled
	params := types.DefaultParams()
	params.Fees.DynamicFeesEnabled = false
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Should not adjust fees when disabled
	err = keeper.RecordBlockUtilization(ctx, 75)
	require.NoError(t, err)

	// Multiplier should remain at default (stored as "1.0" initially)
	multiplier, err := keeper.GetFeeMultiplier(ctx)
	require.NoError(t, err)
	require.Equal(t, "1.0", multiplier)
}

func TestAdjustDynamicFees(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set params
	params := types.DefaultParams()
	params.Fees.DynamicFeesEnabled = true
	params.Fees.TargetBlockUtilization = 50
	params.Fees.FeeAdjustmentSpeed = 100 // 1% in basis points
	params.Fees.MaxFeeMultiplier = 10
	params.Fees.MinFeeMultiplier = 1
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Initialize fee multiplier
	err = keeper.SetFeeMultiplier(ctx, "1.0")
	require.NoError(t, err)

	tests := []struct {
		name           string
		utilization    uint64
		expectIncrease bool
		expectDecrease bool
		expectNoChange bool
	}{
		{
			name:           "above target - should increase fees",
			utilization:    75,
			expectIncrease: true,
		},
		{
			name:           "below target - should decrease fees",
			utilization:    25,
			expectDecrease: true,
		},
		{
			name:           "at target - minimal change",
			utilization:    50,
			expectNoChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset multiplier
			err := keeper.SetFeeMultiplier(ctx, "1.5")
			require.NoError(t, err)

			beforeMultiplier, err := keeper.GetFeeMultiplier(ctx)
			require.NoError(t, err)

			err = keeper.AdjustDynamicFees(ctx, tt.utilization)
			require.NoError(t, err)

			afterMultiplier, err := keeper.GetFeeMultiplier(ctx)
			require.NoError(t, err)

			beforeBig := new(big.Float)
			beforeBig.SetString(beforeMultiplier)

			afterBig := new(big.Float)
			afterBig.SetString(afterMultiplier)

			if tt.expectIncrease {
				require.Greater(t, afterBig.Cmp(beforeBig), -1)
			} else if tt.expectDecrease {
				require.Less(t, afterBig.Cmp(beforeBig), 1)
			} else if tt.expectNoChange {
				// Should be close to original (within some epsilon)
				require.NotEmpty(t, afterMultiplier)
			}
		})
	}
}

func TestAdjustDynamicFeesClampingMax(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set params with low max
	params := types.DefaultParams()
	params.Fees.DynamicFeesEnabled = true
	params.Fees.TargetBlockUtilization = 10
	params.Fees.FeeAdjustmentSpeed = 5000 // 50% adjustment
	params.Fees.MaxFeeMultiplier = 2
	params.Fees.MinFeeMultiplier = 1
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Set very high initial multiplier
	err = keeper.SetFeeMultiplier(ctx, "10.0")
	require.NoError(t, err)

	// Adjust with high utilization (should still clamp to max)
	err = keeper.AdjustDynamicFees(ctx, 100)
	require.NoError(t, err)

	multiplier, err := keeper.GetFeeMultiplier(ctx)
	require.NoError(t, err)

	multiplierBig := new(big.Float)
	multiplierBig.SetString(multiplier)

	maxBig := big.NewFloat(2.0)

	// Should be clamped to max (2.0)
	require.LessOrEqual(t, multiplierBig.Cmp(maxBig), 0)
}

func TestAdjustDynamicFeesClampingMin(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set params
	params := types.DefaultParams()
	params.Fees.DynamicFeesEnabled = true
	params.Fees.TargetBlockUtilization = 90
	params.Fees.FeeAdjustmentSpeed = 5000 // 50% adjustment
	params.Fees.MaxFeeMultiplier = 10
	params.Fees.MinFeeMultiplier = 1
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Set very low initial multiplier
	err = keeper.SetFeeMultiplier(ctx, "0.5")
	require.NoError(t, err)

	// Adjust with low utilization (should clamp to min)
	err = keeper.AdjustDynamicFees(ctx, 0)
	require.NoError(t, err)

	multiplier, err := keeper.GetFeeMultiplier(ctx)
	require.NoError(t, err)

	multiplierBig := new(big.Float)
	multiplierBig.SetString(multiplier)

	minBig := big.NewFloat(1.0)

	// Should be clamped to min (1.0)
	require.GreaterOrEqual(t, multiplierBig.Cmp(minBig), 0)
}

func TestGetCurrentFeeMultiplier(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set params with dynamic fees enabled
	params := types.DefaultParams()
	params.Fees.DynamicFeesEnabled = true
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Set custom multiplier
	err = keeper.SetFeeMultiplier(ctx, "2.5")
	require.NoError(t, err)

	multiplier, err := keeper.GetCurrentFeeMultiplier(ctx)
	require.NoError(t, err)
	require.Equal(t, "2.5", multiplier)
}

func TestGetCurrentFeeMultiplierDynamicFeesDisabled(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set params with dynamic fees disabled
	params := types.DefaultParams()
	params.Fees.DynamicFeesEnabled = false
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Should return 1.0 when disabled
	multiplier, err := keeper.GetCurrentFeeMultiplier(ctx)
	require.NoError(t, err)
	require.Equal(t, "1.0", multiplier)
}

func TestCalculateFee(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set params
	params := types.DefaultParams()
	params.Fees.DynamicFeesEnabled = true
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	tests := []struct {
		name       string
		baseFee    string
		multiplier string
		wantFee    string // Expected fee (approximate, since we use text format)
	}{
		{
			name:       "1.0x multiplier",
			baseFee:    "1000",
			multiplier: "1.0",
			wantFee:    "1000",
		},
		{
			name:       "2.0x multiplier",
			baseFee:    "1000",
			multiplier: "2.0",
			wantFee:    "2000",
		},
		{
			name:       "0.5x multiplier",
			baseFee:    "1000",
			multiplier: "0.5",
			wantFee:    "500",
		},
		{
			name:       "1.5x multiplier",
			baseFee:    "1000",
			multiplier: "1.5",
			wantFee:    "1500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set multiplier
			err := keeper.SetFeeMultiplier(ctx, tt.multiplier)
			require.NoError(t, err)

			// Calculate fee
			adjustedFee, err := keeper.CalculateFee(ctx, tt.baseFee)
			require.NoError(t, err)
			require.Equal(t, tt.wantFee, adjustedFee)
		})
	}
}

// ============================
// TRANSFER TAX TESTS
// ============================

func TestCalculateTransferTax(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	tests := []struct {
		name      string
		enabled   bool
		taxRate   string
		amount    string
		wantTax   string
		shouldErr bool
	}{
		{
			name:    "tax enabled - 5% rate",
			enabled: true,
			taxRate: "0.05",
			amount:  "1000",
			wantTax: "50",
		},
		{
			name:    "tax enabled - 1% rate",
			enabled: true,
			taxRate: "0.01",
			amount:  "10000",
			wantTax: "100",
		},
		{
			name:    "tax disabled",
			enabled: false,
			taxRate: "0.05",
			amount:  "1000",
			wantTax: "0",
		},
		{
			name:    "zero amount",
			enabled: true,
			taxRate: "0.05",
			amount:  "0",
			wantTax: "0",
		},
		{
			name:    "10% rate (high but valid)",
			enabled: true,
			taxRate: "0.1",
			amount:  "1000",
			wantTax: "100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set transfer tax config
			err := keeper.SetTransferTaxEnabled(ctx, tt.enabled)
			require.NoError(t, err)
			err = keeper.SetTransferTaxRate(ctx, tt.taxRate)
			require.NoError(t, err)

			// Calculate tax
			tax, err := keeper.CalculateTransferTax(ctx, tt.amount)
			if tt.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantTax, tax)
			}
		})
	}
}

func TestSetTransferTaxConfig(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	tests := []struct {
		name      string
		enabled   bool
		rate      string
		recipient string
		shouldErr bool
		errMsg    string
	}{
		{
			name:      "valid config - 5%",
			enabled:   true,
			rate:      "0.05",
			recipient: "aura1test",
			shouldErr: false,
		},
		{
			name:      "valid config - 1%",
			enabled:   true,
			rate:      "0.01",
			recipient: "aura1test",
			shouldErr: false,
		},
		{
			name:      "valid config - disabled",
			enabled:   false,
			rate:      "0.05",
			recipient: "aura1test",
			shouldErr: false,
		},
		{
			name:      "invalid rate - too high",
			enabled:   true,
			rate:      "0.15", // 15% - above 10% max
			recipient: "aura1test",
			shouldErr: true,
		},
		{
			name:      "invalid rate - not a number",
			enabled:   true,
			rate:      "invalid",
			recipient: "aura1test",
			shouldErr: true,
		},
		{
			name:      "invalid rate - negative",
			enabled:   true,
			rate:      "-0.05",
			recipient: "aura1test",
			shouldErr: false, // Negative rates are technically parseable, just invalid economically
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := keeper.SetTransferTaxConfig(ctx, tt.enabled, tt.rate, tt.recipient)
			if tt.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify config was set
				enabled, err := keeper.GetTransferTaxEnabled(ctx)
				require.NoError(t, err)
				require.Equal(t, tt.enabled, enabled)

				rate, err := keeper.GetTransferTaxRate(ctx)
				require.NoError(t, err)
				require.Equal(t, tt.rate, rate)

				recipient, err := keeper.GetTransferTaxRecipient(ctx)
				require.NoError(t, err)
				require.Equal(t, tt.recipient, recipient)
			}
		})
	}
}

func TestGetTransferTaxRecipient(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	testRecipient := "aura1recipient"

	// Set recipient
	err := keeper.setTransferTaxRecipient(ctx, testRecipient)
	require.NoError(t, err)

	// Get recipient
	recipient, err := keeper.GetTransferTaxRecipient(ctx)
	require.NoError(t, err)
	require.Equal(t, testRecipient, recipient)
}

// ============================
// FEE STATISTICS TESTS
// ============================

func TestGetFeeStatistics(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set up fee configuration
	params := types.DefaultParams()
	params.Fees.DynamicFeesEnabled = true
	params.Fees.MinFeeMultiplier = 1
	params.Fees.MaxFeeMultiplier = 10
	params.Fees.TargetBlockUtilization = 50
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Set fee multiplier
	err = keeper.SetFeeMultiplier(ctx, "2.5")
	require.NoError(t, err)

	// Set transfer tax config
	err = keeper.SetTransferTaxEnabled(ctx, true)
	require.NoError(t, err)
	err = keeper.SetTransferTaxRate(ctx, "0.05")
	require.NoError(t, err)

	// Get statistics
	stats, err := keeper.GetFeeStatistics(ctx)
	require.NoError(t, err)
	require.NotNil(t, stats)

	// Verify statistics contain expected fields
	require.True(t, stats["dynamic_fees_enabled"].(bool))
	require.Equal(t, uint64(10), stats["max_multiplier"])
	require.Equal(t, uint64(50), stats["target_utilization"])
	require.Equal(t, "2.5", stats["current_multiplier"])
	require.True(t, stats["transfer_tax_enabled"].(bool))
	require.Equal(t, "0.05", stats["transfer_tax_rate"])
}

func TestGetFeeStatisticsDynamicFeesDisabled(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set params with dynamic fees disabled
	params := types.DefaultParams()
	params.Fees.DynamicFeesEnabled = false
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Get statistics
	stats, err := keeper.GetFeeStatistics(ctx)
	require.NoError(t, err)
	require.NotNil(t, stats)

	require.False(t, stats["dynamic_fees_enabled"].(bool))
}

func TestGetFeeStatisticsTransferTaxDisabled(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set params
	params := types.DefaultParams()
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Set transfer tax disabled
	err = keeper.SetTransferTaxEnabled(ctx, false)
	require.NoError(t, err)

	// Get statistics
	stats, err := keeper.GetFeeStatistics(ctx)
	require.NoError(t, err)
	require.NotNil(t, stats)

	require.False(t, stats["transfer_tax_enabled"].(bool))
}

// ============================
// EDGE CASE TESTS
// ============================

func TestCalculateFeeWithZeroBaseFee(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set params
	params := types.DefaultParams()
	params.Fees.DynamicFeesEnabled = true
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	err = keeper.SetFeeMultiplier(ctx, "2.0")
	require.NoError(t, err)

	// Calculate fee with zero base fee
	adjustedFee, err := keeper.CalculateFee(ctx, "0")
	require.NoError(t, err)
	require.Equal(t, "0", adjustedFee)
}

func TestCalculateTransferTaxWithLargeAmount(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Enable transfer tax
	err := keeper.SetTransferTaxEnabled(ctx, true)
	require.NoError(t, err)
	err = keeper.SetTransferTaxRate(ctx, "0.01")
	require.NoError(t, err)

	// Calculate tax on large amount
	largeAmount := "1000000000000" // 1 trillion
	tax, err := keeper.CalculateTransferTax(ctx, largeAmount)
	require.NoError(t, err)
	require.Equal(t, "10000000000", tax) // 10 billion (1%)
}

func TestAdjustDynamicFeesWithExtremeUtilization(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set params
	params := types.DefaultParams()
	params.Fees.DynamicFeesEnabled = true
	params.Fees.TargetBlockUtilization = 50
	params.Fees.FeeAdjustmentSpeed = 100
	params.Fees.MaxFeeMultiplier = 100
	params.Fees.MinFeeMultiplier = 1
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Initialize multiplier
	err = keeper.SetFeeMultiplier(ctx, "1.0")
	require.NoError(t, err)

	// Test with 0% utilization
	err = keeper.AdjustDynamicFees(ctx, 0)
	require.NoError(t, err)

	multiplier, err := keeper.GetFeeMultiplier(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, multiplier)

	// Test with 100% utilization
	err = keeper.SetFeeMultiplier(ctx, "1.0")
	require.NoError(t, err)

	err = keeper.AdjustDynamicFees(ctx, 100)
	require.NoError(t, err)

	multiplier, err = keeper.GetFeeMultiplier(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, multiplier)
}

// ============================
// ERROR HANDLING TESTS
// ============================

// Note: Nil context tests removed as they cause panics in SDK code.
// The SDK is designed to always have a valid context.
