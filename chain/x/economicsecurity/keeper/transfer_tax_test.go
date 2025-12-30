// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// =============================================================================
// Transfer Tax Tests
// =============================================================================

func TestCalculateTransferTax_Disabled(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Tax disabled by default or explicitly
	params, _ := k.GetParams(ctx)
	if params.TransferTax != nil {
		params.TransferTax.Enabled = false
		require.NoError(t, k.SetParams(params))
	}

	totalTax, burnAmount, treasuryAmount, err := k.CalculateTransferTax(ctx, "aura1sender", "1000000")
	require.NoError(t, err)
	require.Equal(t, "0", totalTax)
	require.Equal(t, "0", burnAmount)
	require.Equal(t, "0", treasuryAmount)
}

func TestCalculateTransferTax_Enabled(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = &types.TransferTaxConfig{
		Enabled:                  true,
		BaseTaxRate:              100, // 1% tax (100 basis points)
		BurnPercentage:           5000, // 50% of tax is burned
		TreasuryPercentage:       5000, // 50% of tax goes to treasury
		ExemptedAddresses:        []string{},
		DynamicAdjustmentEnabled: false,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	// Calculate tax on 1,000,000 transfer
	// 1% tax = 10,000
	// 50% burn = 5,000
	// 50% treasury = 5,000
	totalTax, burnAmount, treasuryAmount, err := k.CalculateTransferTax(ctx, "aura1sender", "1000000")
	require.NoError(t, err)
	require.Equal(t, "10000", totalTax)
	require.Equal(t, "5000", burnAmount)
	require.Equal(t, "5000", treasuryAmount)
}

func TestCalculateTransferTax_ExemptedAddress(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = &types.TransferTaxConfig{
		Enabled:           true,
		BaseTaxRate:       100, // 1%
		BurnPercentage:    5000,
		TreasuryPercentage: 5000,
		ExemptedAddresses: []string{"aura1exempted"},
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	// Exempted address should pay no tax
	totalTax, burnAmount, treasuryAmount, err := k.CalculateTransferTax(ctx, "aura1exempted", "1000000")
	require.NoError(t, err)
	require.Equal(t, "0", totalTax)
	require.Equal(t, "0", burnAmount)
	require.Equal(t, "0", treasuryAmount)
}

func TestCalculateTransferTax_InvalidAmount(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = &types.TransferTaxConfig{
		Enabled:     true,
		BaseTaxRate: 100,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	_, _, _, err := k.CalculateTransferTax(ctx, "aura1sender", "invalid-amount")
	require.ErrorIs(t, err, types.ErrInvalidAmount)
}

func TestCalculateTransferTax_ZeroAmount(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = &types.TransferTaxConfig{
		Enabled:           true,
		BaseTaxRate:       100,
		BurnPercentage:    5000,
		TreasuryPercentage: 5000,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	totalTax, burnAmount, treasuryAmount, err := k.CalculateTransferTax(ctx, "aura1sender", "0")
	require.NoError(t, err)
	require.Equal(t, "0", totalTax)
	require.Equal(t, "0", burnAmount)
	require.Equal(t, "0", treasuryAmount)
}

func TestCalculateTransferTax_LargeAmount(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = &types.TransferTaxConfig{
		Enabled:           true,
		BaseTaxRate:       200, // 2%
		BurnPercentage:    7000, // 70%
		TreasuryPercentage: 3000, // 30%
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	// Test with 1 billion tokens (very large)
	totalTax, burnAmount, treasuryAmount, err := k.CalculateTransferTax(ctx, "aura1sender", "1000000000000")
	require.NoError(t, err)
	require.Equal(t, "20000000000", totalTax) // 2% of 1T
	require.Equal(t, "14000000000", burnAmount) // 70% of tax
	require.Equal(t, "6000000000", treasuryAmount) // 30% of tax
}

func TestCalculateTransferTax_WithDynamicAdjustment(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = &types.TransferTaxConfig{
		Enabled:                  true,
		BaseTaxRate:              100,
		BurnPercentage:           5000,
		TreasuryPercentage:       5000,
		DynamicAdjustmentEnabled: true,
		MinTaxRate:               50,
		MaxTaxRate:               200,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	// With dynamic adjustment, tax rate may vary
	totalTax, burnAmount, treasuryAmount, err := k.CalculateTransferTax(ctx, "aura1sender", "1000000")
	require.NoError(t, err)
	require.NotNil(t, totalTax)
	require.NotNil(t, burnAmount)
	require.NotNil(t, treasuryAmount)
}

func TestGetTransferTaxStats(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	enabled, baseTaxRate, currentTaxRate, totalBurned, exemptedCount := k.GetTransferTaxStats(ctx)
	// Default should have tax disabled
	require.False(t, enabled)
	require.NotNil(t, baseTaxRate)
	require.NotNil(t, currentTaxRate)
	require.NotNil(t, totalBurned)
	require.NotNil(t, exemptedCount)
}

func TestGetTransferTaxStats_Enabled(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = &types.TransferTaxConfig{
		Enabled:           true,
		BaseTaxRate:       100,
		ExemptedAddresses: []string{"addr1", "addr2"},
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	enabled, baseTaxRate, _, _, exemptedCount := k.GetTransferTaxStats(ctx)
	require.True(t, enabled)
	require.Equal(t, uint64(100), baseTaxRate)
	require.Equal(t, uint64(2), exemptedCount)
}

func TestCalculateDynamicTaxRate(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = &types.TransferTaxConfig{
		Enabled:                  true,
		BaseTaxRate:              100,
		DynamicAdjustmentEnabled: true,
		MinTaxRate:               50,
		MaxTaxRate:               200,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	rate := k.calculateDynamicTaxRate(ctx, params.TransferTax)
	require.GreaterOrEqual(t, rate, params.TransferTax.MinTaxRate)
	require.LessOrEqual(t, rate, params.TransferTax.MaxTaxRate)
}
