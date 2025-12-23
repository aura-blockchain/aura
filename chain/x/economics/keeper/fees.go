package keeper

import (
	"fmt"

	"context"

	sdkmath "cosmossdk.io/math"

	"github.com/aequitas/aura/chain/x/economics/types"
)

// ============================
// DYNAMIC FEE MANAGEMENT
// ============================

// RecordBlockUtilization records block utilization for fee calculation
func (k Keeper) RecordBlockUtilization(ctx context.Context, utilization uint64) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	if !params.Fees.DynamicFeesEnabled {
		return nil
	}

	// Store the utilization in state
	// For simplicity, we'll use a simple moving average approach
	// In production, you might store a circular buffer of recent utilizations

	return k.AdjustDynamicFees(ctx, utilization)
}

// AdjustDynamicFees adjusts fee multiplier based on network congestion
// Uses LegacyDec for deterministic decimal arithmetic across all validators
func (k Keeper) AdjustDynamicFees(ctx context.Context, currentUtilization uint64) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return fmt.Errorf("failed to get for validators: %w", err)
	}

	if !params.Fees.DynamicFeesEnabled {
		return nil
	}

	// Get current fee multiplier from store
	currentMultiplierStr, err := k.GetFeeMultiplier(ctx)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	// Parse current fee multiplier using deterministic LegacyDec
	currentMultiplier, err := sdkmath.LegacyNewDecFromStr(currentMultiplierStr)
	if err != nil {
		currentMultiplier = sdkmath.LegacyOneDec()
	}

	// Calculate adjustment based on utilization vs target
	targetUtil := params.Fees.TargetBlockUtilization

	// Use LegacyDec for all calculations to ensure determinism
	var adjustment sdkmath.LegacyDec
	one := sdkmath.LegacyOneDec()
	basisPoints := sdkmath.LegacyNewDec(10000)

	if currentUtilization > targetUtil {
		// Calculate percentage over target using integer division
		overTarget := sdkmath.LegacyNewDec(int64(currentUtilization - targetUtil))
		targetDec := sdkmath.LegacyNewDec(int64(targetUtil))
		overTargetRatio := overTarget.Quo(targetDec)

		// Convert adjustment speed from basis points
		adjSpeed := sdkmath.LegacyNewDec(int64(params.Fees.FeeAdjustmentSpeed))
		adjSpeed = adjSpeed.Quo(basisPoints)

		// Calculate increase: 1 + (adjSpeed * overTargetRatio)
		increase := adjSpeed.Mul(overTargetRatio)
		adjustment = one.Add(increase)
	} else {
		// Calculate percentage under target using integer division
		underTarget := sdkmath.LegacyNewDec(int64(targetUtil - currentUtilization))
		targetDec := sdkmath.LegacyNewDec(int64(targetUtil))
		underTargetRatio := underTarget.Quo(targetDec)

		// Convert adjustment speed from basis points
		adjSpeed := sdkmath.LegacyNewDec(int64(params.Fees.FeeAdjustmentSpeed))
		adjSpeed = adjSpeed.Quo(basisPoints)

		// Calculate decrease: 1 - (adjSpeed * underTargetRatio)
		decrease := adjSpeed.Mul(underTargetRatio)
		adjustment = one.Sub(decrease)
	}

	// Apply adjustment
	newMultiplier := currentMultiplier.Mul(adjustment)

	// Clamp to max fee multiplier
	maxMultiplier := sdkmath.LegacyNewDec(int64(params.Fees.MaxFeeMultiplier))
	if newMultiplier.GT(maxMultiplier) {
		newMultiplier = maxMultiplier
	}

	// Clamp to minimum fee multiplier
	minMultiplier := sdkmath.LegacyNewDec(int64(params.Fees.MinFeeMultiplier))
	if newMultiplier.LT(minMultiplier) {
		newMultiplier = minMultiplier
	}

	// Update multiplier in store (LegacyDec.String() is deterministic)
	return k.SetFeeMultiplier(ctx, newMultiplier.String())
}

// GetCurrentFeeMultiplier returns the current fee multiplier
func (k Keeper) GetCurrentFeeMultiplier(ctx context.Context) (string, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return "1.0", err
	}

	if !params.Fees.DynamicFeesEnabled {
		return "1.0", nil
	}

	return k.GetFeeMultiplier(ctx)
}

// CalculateFee calculates the fee for a transaction based on current multiplier
// Uses LegacyDec for deterministic calculations
func (k Keeper) CalculateFee(ctx context.Context, baseFee string) (string, error) {
	multiplier, err := k.GetCurrentFeeMultiplier(ctx)
	if err != nil {
		return baseFee, err
	}

	// Parse values using deterministic LegacyDec
	baseFeeDec, err := sdkmath.LegacyNewDecFromStr(baseFee)
	if err != nil {
		return baseFee, err
	}

	multiplierDec, err := sdkmath.LegacyNewDecFromStr(multiplier)
	if err != nil {
		return baseFee, err
	}

	// Calculate adjusted fee (TruncateInt for integer result)
	adjustedFee := baseFeeDec.Mul(multiplierDec).TruncateInt()

	return adjustedFee.String(), nil
}

// ============================
// TRANSFER TAX MANAGEMENT
// ============================
// Note: Transfer tax is stored separately in KV store, not in params

// CalculateTransferTax calculates the transfer tax for an amount
// Uses LegacyDec for deterministic calculations
func (k Keeper) CalculateTransferTax(ctx context.Context, amount string) (string, error) {
	// Get transfer tax config from store
	enabled, err := k.GetTransferTaxEnabled(ctx)
	if err != nil {
		return "0", err
	}

	if !enabled {
		return "0", nil
	}

	taxRate, err := k.GetTransferTaxRate(ctx)
	if err != nil {
		return "0", err
	}

	// Parse values using deterministic LegacyDec
	amountDec, err := sdkmath.LegacyNewDecFromStr(amount)
	if err != nil {
		return "0", err
	}

	taxRateDec, err := sdkmath.LegacyNewDecFromStr(taxRate)
	if err != nil {
		return "0", err
	}

	// Calculate tax (TruncateInt for integer result)
	tax := amountDec.Mul(taxRateDec).TruncateInt()

	return tax.String(), nil
}

// GetTransferTaxRecipient returns the transfer tax recipient address
func (k Keeper) GetTransferTaxRecipient(ctx context.Context) (string, error) {
	return k.getTransferTaxRecipient(ctx)
}

// SetTransferTaxConfig updates the transfer tax configuration
// Uses LegacyDec for deterministic rate validation
func (k Keeper) SetTransferTaxConfig(ctx context.Context, enabled bool, rate string, recipient string) error {
	// Validate rate using deterministic LegacyDec
	rateDec, err := sdkmath.LegacyNewDecFromStr(rate)
	if err != nil {
		return types.ErrInvalidTaxConfig
	}

	// Check if rate is reasonable (e.g., < 10%)
	maxRate := sdkmath.LegacyNewDecWithPrec(1, 1) // 0.1 = 10%
	if rateDec.GT(maxRate) {
		return types.ErrTaxRateTooHigh
	}

	// Store config in separate KV store
	if err := k.SetTransferTaxEnabled(ctx, enabled); err != nil {
		return fmt.Errorf("error in SetTransferTaxConfig: %w", err)
	}
	if err := k.SetTransferTaxRate(ctx, rate); err != nil {
		return fmt.Errorf("error in SetTransferTaxConfig: %w", err)
	}
	return k.setTransferTaxRecipient(ctx, recipient)
}

// ============================
// FEE STATISTICS
// ============================

// GetFeeStatistics returns statistics about fee collection
func (k Keeper) GetFeeStatistics(ctx context.Context) (map[string]interface{}, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})

	if true {
		stats["dynamic_fees_enabled"] = params.Fees.DynamicFeesEnabled
		stats["max_multiplier"] = params.Fees.MaxFeeMultiplier
		stats["target_utilization"] = params.Fees.TargetBlockUtilization

		// Get current multiplier from store
		if multiplier, err := k.GetFeeMultiplier(ctx); err == nil {
			stats["current_multiplier"] = multiplier
		}
	}

	// Get transfer tax config from store
	if enabled, err := k.GetTransferTaxEnabled(ctx); err == nil {
		stats["transfer_tax_enabled"] = enabled
	}
	if rate, err := k.GetTransferTaxRate(ctx); err == nil {
		stats["transfer_tax_rate"] = rate
	}

	return stats, nil
}
