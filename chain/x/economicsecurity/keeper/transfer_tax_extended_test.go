// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// =============================================================================
// RemoveTaxExemption Tests
// =============================================================================

func TestRemoveTaxExemption_Success(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = &types.TransferTaxConfig{
		Enabled:            true,
		BaseTaxRate:        100,
		BurnPercentage:     5000,
		TreasuryPercentage: 5000,
		ExemptedAddresses:  []string{"aura1exempted1", "aura1exempted2", "aura1exempted3"},
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	// Remove middle address
	err := k.RemoveTaxExemption(ctx, "aura1exempted2")
	require.NoError(t, err)

	// Verify removal
	updatedParams, _ := k.GetParams(ctx)
	require.Len(t, updatedParams.TransferTax.ExemptedAddresses, 2)
	require.Contains(t, updatedParams.TransferTax.ExemptedAddresses, "aura1exempted1")
	require.Contains(t, updatedParams.TransferTax.ExemptedAddresses, "aura1exempted3")
	require.NotContains(t, updatedParams.TransferTax.ExemptedAddresses, "aura1exempted2")
}

func TestRemoveTaxExemption_AddressNotExempted(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = &types.TransferTaxConfig{
		Enabled:           true,
		BaseTaxRate:       100,
		ExemptedAddresses: []string{"aura1exempted1"},
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	// Removing non-exempted address should not error
	err := k.RemoveTaxExemption(ctx, "aura1notexempted")
	require.NoError(t, err)

	// Verify nothing changed
	updatedParams, _ := k.GetParams(ctx)
	require.Len(t, updatedParams.TransferTax.ExemptedAddresses, 1)
}

func TestRemoveTaxExemption_NilTransferTax(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = nil

	k, ctx := setupKeeperWithCustomParams(t, params)

	err := k.RemoveTaxExemption(ctx, "aura1anyaddress")
	require.ErrorIs(t, err, types.ErrInvalidTaxConfig)
}

func TestRemoveTaxExemption_EmptyList(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = &types.TransferTaxConfig{
		Enabled:           true,
		ExemptedAddresses: []string{},
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	err := k.RemoveTaxExemption(ctx, "aura1anyaddress")
	require.NoError(t, err)
}

// =============================================================================
// GetTaxDistribution Tests
// =============================================================================

func TestGetTaxDistribution_Success(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = &types.TransferTaxConfig{
		Enabled:                true,
		BaseTaxRate:            100,
		BurnPercentage:         4000, // 40%
		TreasuryPercentage:     3000, // 30%
		RedistributePercentage: 3000, // 30%
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	// Tax of 10000
	burnAmount, treasuryAmount, redistributeAmount, err := k.GetTaxDistribution(ctx, "10000")
	require.NoError(t, err)
	require.Equal(t, "4000", burnAmount)     // 40% of 10000
	require.Equal(t, "3000", treasuryAmount) // 30% of 10000
	require.Equal(t, "3000", redistributeAmount) // 30% of 10000
}

func TestGetTaxDistribution_NilTransferTax(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = nil

	k, ctx := setupKeeperWithCustomParams(t, params)

	burnAmount, treasuryAmount, redistributeAmount, err := k.GetTaxDistribution(ctx, "10000")
	require.ErrorIs(t, err, types.ErrInvalidTaxConfig)
	require.Equal(t, "0", burnAmount)
	require.Equal(t, "0", treasuryAmount)
	require.Equal(t, "0", redistributeAmount)
}

func TestGetTaxDistribution_InvalidAmount(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = &types.TransferTaxConfig{
		Enabled:     true,
		BaseTaxRate: 100,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	_, _, _, err := k.GetTaxDistribution(ctx, "not-a-number")
	require.ErrorIs(t, err, types.ErrInvalidAmount)
}

func TestGetTaxDistribution_ZeroTax(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = &types.TransferTaxConfig{
		Enabled:            true,
		BurnPercentage:     5000,
		TreasuryPercentage: 5000,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	burnAmount, treasuryAmount, redistributeAmount, err := k.GetTaxDistribution(ctx, "0")
	require.NoError(t, err)
	require.Equal(t, "0", burnAmount)
	require.Equal(t, "0", treasuryAmount)
	require.Equal(t, "0", redistributeAmount)
}

func TestGetTaxDistribution_LargeAmount(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = &types.TransferTaxConfig{
		Enabled:                true,
		BurnPercentage:         5000,
		TreasuryPercentage:     3000,
		RedistributePercentage: 2000,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	// Large tax amount
	burnAmount, treasuryAmount, redistributeAmount, err := k.GetTaxDistribution(ctx, "1000000000000")
	require.NoError(t, err)
	require.Equal(t, "500000000000", burnAmount)     // 50%
	require.Equal(t, "300000000000", treasuryAmount) // 30%
	require.Equal(t, "200000000000", redistributeAmount) // 20%
}

// =============================================================================
// ValidateTransferTaxConfig Tests
// =============================================================================

func TestValidateTransferTaxConfig_NilConfig(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	err := k.ValidateTransferTaxConfig(nil)
	require.NoError(t, err)
}

func TestValidateTransferTaxConfig_ValidConfig(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	config := &types.TransferTaxConfig{
		Enabled:                true,
		BaseTaxRate:            100, // 1%
		BurnPercentage:         4000,
		TreasuryPercentage:     3000,
		RedistributePercentage: 3000,
	}

	err := k.ValidateTransferTaxConfig(config)
	require.NoError(t, err)
}

func TestValidateTransferTaxConfig_TaxRateTooHigh(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	config := &types.TransferTaxConfig{
		Enabled:     true,
		BaseTaxRate: 1001, // 10.01% - exceeds 10% max
	}

	err := k.ValidateTransferTaxConfig(config)
	require.ErrorIs(t, err, types.ErrTaxRateTooHigh)
}

func TestValidateTransferTaxConfig_PercentagesSumNot100(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	config := &types.TransferTaxConfig{
		Enabled:                true,
		BaseTaxRate:            100,
		BurnPercentage:         4000,
		TreasuryPercentage:     3000,
		RedistributePercentage: 2000, // Sum = 9000, not 10000
	}

	err := k.ValidateTransferTaxConfig(config)
	require.ErrorIs(t, err, types.ErrInvalidTaxConfig)
}

func TestValidateTransferTaxConfig_PercentageExceeds100(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	config := &types.TransferTaxConfig{
		Enabled:                true,
		BaseTaxRate:            100,
		BurnPercentage:         15000, // 150% - invalid
		TreasuryPercentage:     0,
		RedistributePercentage: 0,
	}

	err := k.ValidateTransferTaxConfig(config)
	require.ErrorIs(t, err, types.ErrInvalidTaxConfig)
}

func TestValidateTransferTaxConfig_AllBurn(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	config := &types.TransferTaxConfig{
		Enabled:                true,
		BaseTaxRate:            100,
		BurnPercentage:         10000, // 100% burn
		TreasuryPercentage:     0,
		RedistributePercentage: 0,
	}

	err := k.ValidateTransferTaxConfig(config)
	require.NoError(t, err)
}

func TestValidateTransferTaxConfig_MaxValidRate(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	config := &types.TransferTaxConfig{
		Enabled:                true,
		BaseTaxRate:            1000, // 10% - exactly max
		BurnPercentage:         10000,
		TreasuryPercentage:     0,
		RedistributePercentage: 0,
	}

	err := k.ValidateTransferTaxConfig(config)
	require.NoError(t, err)
}

// =============================================================================
// GetEffectiveTransferAmount Tests
// =============================================================================

func TestGetEffectiveTransferAmount_Success(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = &types.TransferTaxConfig{
		Enabled:            true,
		BaseTaxRate:        100, // 1%
		BurnPercentage:     5000,
		TreasuryPercentage: 5000,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	// 1000000 - 1% tax = 990000
	effective, err := k.GetEffectiveTransferAmount(ctx, "aura1sender", "1000000")
	require.NoError(t, err)
	require.Equal(t, "990000", effective)
}

func TestGetEffectiveTransferAmount_TaxDisabled(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = &types.TransferTaxConfig{
		Enabled:     false,
		BaseTaxRate: 100,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	// No tax, full amount
	effective, err := k.GetEffectiveTransferAmount(ctx, "aura1sender", "1000000")
	require.NoError(t, err)
	require.Equal(t, "1000000", effective)
}

func TestGetEffectiveTransferAmount_ExemptedSender(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = &types.TransferTaxConfig{
		Enabled:           true,
		BaseTaxRate:       100,
		ExemptedAddresses: []string{"aura1exempted"},
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	// Exempted sender pays no tax
	effective, err := k.GetEffectiveTransferAmount(ctx, "aura1exempted", "1000000")
	require.NoError(t, err)
	require.Equal(t, "1000000", effective)
}

func TestGetEffectiveTransferAmount_InvalidAmount(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = &types.TransferTaxConfig{
		Enabled:     true,
		BaseTaxRate: 100,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	_, err := k.GetEffectiveTransferAmount(ctx, "aura1sender", "invalid")
	require.Error(t, err)
}

func TestGetEffectiveTransferAmount_ZeroAmount(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = &types.TransferTaxConfig{
		Enabled:     true,
		BaseTaxRate: 100,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	effective, err := k.GetEffectiveTransferAmount(ctx, "aura1sender", "0")
	require.NoError(t, err)
	require.Equal(t, "0", effective)
}

func TestGetEffectiveTransferAmount_LargeAmount(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = &types.TransferTaxConfig{
		Enabled:            true,
		BaseTaxRate:        200, // 2%
		BurnPercentage:     5000,
		TreasuryPercentage: 5000,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	// 1 trillion - 2% = 980 billion
	effective, err := k.GetEffectiveTransferAmount(ctx, "aura1sender", "1000000000000")
	require.NoError(t, err)
	require.Equal(t, "980000000000", effective)
}

// =============================================================================
// AddTaxExemption Tests (for completeness)
// =============================================================================

func TestAddTaxExemption_Success(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = &types.TransferTaxConfig{
		Enabled:                true,
		BaseTaxRate:            100,
		BurnPercentage:         5000,
		TreasuryPercentage:     5000,
		RedistributePercentage: 0,
		ExemptedAddresses:      []string{"aura1existing"},
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	err := k.AddTaxExemption(ctx, "aura1newexempt")
	require.NoError(t, err)

	updatedParams, _ := k.GetParams(ctx)
	require.Len(t, updatedParams.TransferTax.ExemptedAddresses, 2)
	require.Contains(t, updatedParams.TransferTax.ExemptedAddresses, "aura1newexempt")
}

func TestAddTaxExemption_NilTransferTax(t *testing.T) {
	params := types.DefaultParams()
	params.TransferTax = nil

	k, ctx := setupKeeperWithCustomParams(t, params)

	err := k.AddTaxExemption(ctx, "aura1newexempt")
	require.ErrorIs(t, err, types.ErrInvalidTaxConfig)
}
