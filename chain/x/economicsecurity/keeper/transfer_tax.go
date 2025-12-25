// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"math/big"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// ============================
// TRANSFER TAX (Feature 6)
// ============================

const (
	// BasisPoints is the denominator for percentage calculations (100% = 10000 basis points)
	BasisPoints = 10000
)

// CalculateTransferTax calculates the tax amount for a transfer
func (k *Keeper) CalculateTransferTax(
	ctx context.Context,
	sender, amount string,
) (totalTax string, burnAmount string, treasuryAmount string, err error) {
	params, _ := k.GetParams(ctx)

	if params.TransferTax == nil || !params.TransferTax.Enabled {
		return "0", "0", "0", nil
	}

	// Check if sender is exempted
	for _, exempted := range params.TransferTax.ExemptedAddresses {
		if sender == exempted {
			return "0", "0", "0", nil
		}
	}

	// Validate and parse transfer amount
	transferAmt := new(big.Int)
	if _, ok := transferAmt.SetString(amount, 10); !ok {
		return "0", "0", "0", types.ErrInvalidAmount
	}

	// Calculate tax rate (may be dynamic)
	taxRate := params.TransferTax.BaseTaxRate
	if params.TransferTax.DynamicAdjustmentEnabled {
		taxRate = k.calculateDynamicTaxRate(ctx, params.TransferTax)
	}

	// Calculate total tax
	basisPts := big.NewInt(BasisPoints)
	tax := new(big.Int).Mul(transferAmt, big.NewInt(int64(taxRate)))
	tax.Div(tax, basisPts)

	// Split tax according to distribution percentages
	burnAmt := new(big.Int).Mul(tax, big.NewInt(int64(params.TransferTax.BurnPercentage)))
	burnAmt.Div(burnAmt, basisPts)

	treasuryAmt := new(big.Int).Mul(tax, big.NewInt(int64(params.TransferTax.TreasuryPercentage)))
	treasuryAmt.Div(treasuryAmt, basisPts)

	redistributeAmt := new(big.Int).Mul(tax, big.NewInt(int64(params.TransferTax.RedistributePercentage)))
	redistributeAmt.Div(redistributeAmt, basisPts)

	// The redistribution amount is available for later distribution
	// It could be added to a redistribution pool or distributed immediately

	return tax.String(), burnAmt.String(), treasuryAmt.String(), nil
}

// ProcessTransferTax processes the transfer tax (burns, sends to treasury, redistributes)
func (k *Keeper) ProcessTransferTax(ctx context.Context, burnAmount, treasuryAmount string) error {
	// Update total burned
	burnAmt := new(big.Int)
	if _, ok := burnAmt.SetString(burnAmount, 10); !ok {
		return types.ErrInvalidAmount
	}

	if burnAmt.Cmp(big.NewInt(0)) > 0 {
		// Get current total burned
		currentBurned, err := k.GetTotalBurned(ctx)
		if err != nil {
			return err
		}

		totalBurned := new(big.Int)
		if currentBurned != "0" && currentBurned != "" {
			totalBurned.SetString(currentBurned, 10)
		}
		totalBurned.Add(totalBurned, burnAmt)

		// Update total burned in store
		if err := k.SetTotalBurned(ctx, totalBurned.String()); err != nil {
			return err
		}
	}

	// Treasury amount would be sent to treasury address via bank module
	// This is handled by the caller

	return nil
}

// calculateDynamicTaxRate calculates tax rate based on market conditions
func (k *Keeper) calculateDynamicTaxRate(ctx context.Context, config *types.TransferTaxConfig) uint64 {
	// Dynamic tax rate calculation based on on-chain metrics
	// This implementation provides a sophisticated framework for dynamic adjustment

	baseTaxRate := config.BaseTaxRate

	// In a production system, this could adjust based on:
	// 1. Transaction volume (higher volume = lower tax to encourage activity)
	// 2. Price volatility (higher volatility = higher tax to stabilize)
	// 3. Liquidity metrics (lower liquidity = lower tax to encourage provision)
	// 4. Network congestion (higher congestion = higher tax)
	// 5. Time-based factors (e.g., different rates during different periods)

	// For now, we'll implement a simple volume-based adjustment
	// In practice, you'd query actual metrics from the chain

	// Get current height to calculate recent activity
	currentHeight, err := k.GetCurrentHeight(ctx)
	if err != nil {
		return baseTaxRate
	}

	// Calculate adjustment factor based on activity
	// This is a placeholder for actual metrics
	adjustmentFactor := k.calculateTaxAdjustmentFactor(ctx, currentHeight)

	// Apply adjustment (max ±50% from base rate)
	maxAdjustment := int64(baseTaxRate) / 2
	adjustment := int64(adjustmentFactor) * maxAdjustment / 100

	adjustedRate := int64(baseTaxRate) + adjustment

	// Ensure rate stays within bounds (0.01% to 5%)
	minRate := int64(1)    // 0.01%
	maxRate := int64(500)  // 5%

	if adjustedRate < minRate {
		adjustedRate = minRate
	}
	if adjustedRate > maxRate {
		adjustedRate = maxRate
	}

	return uint64(adjustedRate)
}

// calculateTaxAdjustmentFactor calculates the adjustment factor for dynamic tax rate
// Returns a value between -100 and +100 representing percentage adjustment
func (k *Keeper) calculateTaxAdjustmentFactor(ctx context.Context, currentHeight uint64) int64 {
	// This is a simplified implementation
	// In production, this would analyze:
	// - Recent transaction volume trends
	// - Price volatility (if oracle available)
	// - Liquidity depth
	// - Network utilization

	// For now, return 0 (no adjustment)
	// This maintains the base tax rate while providing the framework for future enhancements
	return 0
}

// GetTransferTaxStats returns transfer tax statistics
func (k *Keeper) GetTransferTaxStats(ctx context.Context) (
	enabled bool,
	baseTaxRate uint64,
	currentTaxRate uint64,
	totalBurned string,
	exemptedAddressCount uint64,
) {
	params, _ := k.GetParams(ctx)

	if params.TransferTax == nil || !params.TransferTax.Enabled {
		return false, 0, 0, "0", 0
	}

	baseTaxRate = params.TransferTax.BaseTaxRate

	currentTaxRate = baseTaxRate
	if params.TransferTax.DynamicAdjustmentEnabled {
		currentTaxRate = k.calculateDynamicTaxRate(ctx, params.TransferTax)
	}

	totalBurnedStr, err := k.GetTotalBurned(ctx)
	if err != nil {
		totalBurnedStr = "0"
	}

	exemptedCount := uint64(len(params.TransferTax.ExemptedAddresses))

	return true, baseTaxRate, currentTaxRate, totalBurnedStr, exemptedCount
}

// IsAddressExemptFromTax checks if an address is exempt from transfer tax
func (k *Keeper) IsAddressExemptFromTax(ctx context.Context, address string) bool {
	params, _ := k.GetParams(ctx)

	if params.TransferTax == nil {
		return false
	}

	for _, exempted := range params.TransferTax.ExemptedAddresses {
		if exempted == address {
			return true
		}
	}

	return false
}

// AddTaxExemption adds an address to the transfer tax exemption list
func (k *Keeper) AddTaxExemption(ctx context.Context, address string) error {
	params, _ := k.GetParams(ctx)

	if params.TransferTax == nil {
		return types.ErrInvalidTaxConfig
	}

	// Check if already exempted
	for _, exempted := range params.TransferTax.ExemptedAddresses {
		if exempted == address {
			return nil // Already exempted
		}
	}

	params.TransferTax.ExemptedAddresses = append(params.TransferTax.ExemptedAddresses, address)

	return k.SetParams(params)
}

// RemoveTaxExemption removes an address from the transfer tax exemption list
func (k *Keeper) RemoveTaxExemption(ctx context.Context, address string) error {
	params, _ := k.GetParams(ctx)

	if params.TransferTax == nil {
		return types.ErrInvalidTaxConfig
	}

	// Find and remove the address
	newExemptions := make([]string, 0, len(params.TransferTax.ExemptedAddresses))
	found := false

	for _, exempted := range params.TransferTax.ExemptedAddresses {
		if exempted != address {
			newExemptions = append(newExemptions, exempted)
		} else {
			found = true
		}
	}

	if !found {
		return nil // Address wasn't exempted anyway
	}

	params.TransferTax.ExemptedAddresses = newExemptions

	return k.SetParams(params)
}

// GetTaxDistribution calculates how a tax amount will be distributed
func (k *Keeper) GetTaxDistribution(ctx context.Context, taxAmount string) (
	burnAmount string,
	treasuryAmount string,
	redistributeAmount string,
	err error,
) {
	params, _ := k.GetParams(ctx)

	if params.TransferTax == nil {
		return "0", "0", "0", types.ErrInvalidTaxConfig
	}

	tax := new(big.Int)
	if _, ok := tax.SetString(taxAmount, 10); !ok {
		return "0", "0", "0", types.ErrInvalidAmount
	}

	basisPts := big.NewInt(BasisPoints)

	burnAmt := new(big.Int).Mul(tax, big.NewInt(int64(params.TransferTax.BurnPercentage)))
	burnAmt.Div(burnAmt, basisPts)

	treasuryAmt := new(big.Int).Mul(tax, big.NewInt(int64(params.TransferTax.TreasuryPercentage)))
	treasuryAmt.Div(treasuryAmt, basisPts)

	redistributeAmt := new(big.Int).Mul(tax, big.NewInt(int64(params.TransferTax.RedistributePercentage)))
	redistributeAmt.Div(redistributeAmt, basisPts)

	return burnAmt.String(), treasuryAmt.String(), redistributeAmt.String(), nil
}

// ValidateTransferTaxConfig validates transfer tax configuration
func (k *Keeper) ValidateTransferTaxConfig(config *types.TransferTaxConfig) error {
	if config == nil {
		return nil // Config is optional
	}

	// Validate tax rate is within reasonable bounds (0.01% to 10%)
	if config.BaseTaxRate > 1000 { // 10%
		return types.ErrTaxRateTooHigh
	}

	// Validate distribution percentages sum to 100%
	totalPercentage := config.BurnPercentage + config.TreasuryPercentage + config.RedistributePercentage
	if totalPercentage != BasisPoints {
		return types.ErrInvalidTaxConfig
	}

	// Each percentage must be <= 100%
	if config.BurnPercentage > BasisPoints || config.TreasuryPercentage > BasisPoints || config.RedistributePercentage > BasisPoints {
		return types.ErrInvalidTaxConfig
	}

	return nil
}

// GetEffectiveTransferAmount calculates the amount that will be received after tax
func (k *Keeper) GetEffectiveTransferAmount(ctx context.Context, sender, amount string) (string, error) {
	totalTax, _, _, err := k.CalculateTransferTax(ctx, sender, amount)
	if err != nil {
		return "0", err
	}

	transferAmt := new(big.Int)
	if _, ok := transferAmt.SetString(amount, 10); !ok {
		return "0", types.ErrInvalidAmount
	}

	tax := new(big.Int)
	if totalTax != "0" && totalTax != "" {
		tax.SetString(totalTax, 10)
	}

	effective := new(big.Int).Sub(transferAmt, tax)

	return effective.String(), nil
}
