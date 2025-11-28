package keeper

import (
	"context"
	"math/big"

	"github.com/aequitas/aura/chain/x/economics/types"
)

// ============================
// DYNAMIC FEE MANAGEMENT
// ============================

// RecordBlockUtilization records block utilization for fee calculation
func (k Keeper) RecordBlockUtilization(ctx context.Context, utilization uint64) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	if params.Fees == nil || !params.Fees.DynamicFeesEnabled {
		return nil
	}

	// Store the utilization in state
	// For simplicity, we'll use a simple moving average approach
	// In production, you might store a circular buffer of recent utilizations

	return k.AdjustDynamicFees(ctx, utilization)
}

// AdjustDynamicFees adjusts fee multiplier based on network congestion
func (k Keeper) AdjustDynamicFees(ctx context.Context, currentUtilization uint64) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	if params.Fees == nil || !params.Fees.DynamicFeesEnabled {
		return nil
	}

	// Get current fee multiplier from store
	currentMultiplierStr, err := k.GetFeeMultiplier(ctx)
	if err != nil {
		return err
	}

	// Parse current fee multiplier
	currentMultiplier := new(big.Float)
	currentMultiplier.SetString(currentMultiplierStr)

	// Calculate adjustment based on utilization vs target
	targetUtil := params.Fees.TargetBlockUtilization

	// If utilization > target, increase fees
	// If utilization < target, decrease fees
	var adjustment *big.Float
	if currentUtilization > targetUtil {
		// Calculate percentage over target
		overTarget := float64(currentUtilization - targetUtil) / float64(targetUtil)

		// Convert adjustment speed from uint64 to float
		adjSpeed := new(big.Float).SetUint64(params.Fees.FeeAdjustmentSpeed)
		// Divide by 10000 to convert from basis points
		adjSpeed.Quo(adjSpeed, big.NewFloat(10000.0))

		// Calculate increase
		increase := new(big.Float).Mul(adjSpeed, big.NewFloat(overTarget))
		adjustment = new(big.Float).Add(big.NewFloat(1.0), increase)
	} else {
		// Calculate percentage under target
		underTarget := float64(targetUtil - currentUtilization) / float64(targetUtil)

		// Convert adjustment speed from uint64 to float
		adjSpeed := new(big.Float).SetUint64(params.Fees.FeeAdjustmentSpeed)
		// Divide by 10000 to convert from basis points
		adjSpeed.Quo(adjSpeed, big.NewFloat(10000.0))

		// Calculate decrease
		decrease := new(big.Float).Mul(adjSpeed, big.NewFloat(underTarget))
		adjustment = new(big.Float).Sub(big.NewFloat(1.0), decrease)
	}

	// Apply adjustment
	newMultiplier := new(big.Float).Mul(currentMultiplier, adjustment)

	// Clamp to max fee multiplier
	maxMultiplier := new(big.Float).SetUint64(params.Fees.MaxFeeMultiplier)

	if newMultiplier.Cmp(maxMultiplier) > 0 {
		newMultiplier = maxMultiplier
	}

	// Clamp to minimum fee multiplier
	minMultiplier := new(big.Float).SetUint64(params.Fees.MinFeeMultiplier)
	if newMultiplier.Cmp(minMultiplier) < 0 {
		newMultiplier = minMultiplier
	}

	// Update multiplier in store
	return k.SetFeeMultiplier(ctx, newMultiplier.Text('f', 6))
}

// GetCurrentFeeMultiplier returns the current fee multiplier
func (k Keeper) GetCurrentFeeMultiplier(ctx context.Context) (string, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return "1.0", err
	}

	if params.Fees == nil || !params.Fees.DynamicFeesEnabled {
		return "1.0", nil
	}

	return k.GetFeeMultiplier(ctx)
}

// CalculateFee calculates the fee for a transaction based on current multiplier
func (k Keeper) CalculateFee(ctx context.Context, baseFee string) (string, error) {
	multiplier, err := k.GetCurrentFeeMultiplier(ctx)
	if err != nil {
		return baseFee, err
	}

	// Parse values
	baseFeeBig := new(big.Float)
	baseFeeBig.SetString(baseFee)

	multiplierBig := new(big.Float)
	multiplierBig.SetString(multiplier)

	// Calculate adjusted fee
	adjustedFee := new(big.Float).Mul(baseFeeBig, multiplierBig)

	return adjustedFee.Text('f', 0), nil
}

// ============================
// TRANSFER TAX MANAGEMENT
// ============================
// Note: Transfer tax is stored separately in KV store, not in params

// CalculateTransferTax calculates the transfer tax for an amount
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

	// Parse values
	amountBig := new(big.Float)
	amountBig.SetString(amount)

	taxRateBig := new(big.Float)
	taxRateBig.SetString(taxRate)

	// Calculate tax
	tax := new(big.Float).Mul(amountBig, taxRateBig)

	return tax.Text('f', 0), nil
}

// GetTransferTaxRecipient returns the transfer tax recipient address
func (k Keeper) GetTransferTaxRecipient(ctx context.Context) (string, error) {
	return k.getTransferTaxRecipient(ctx)
}

// SetTransferTaxConfig updates the transfer tax configuration
func (k Keeper) SetTransferTaxConfig(ctx context.Context, enabled bool, rate string, recipient string) error {
	// Validate rate
	rateBig := new(big.Float)
	if _, ok := rateBig.SetString(rate); !ok {
		return types.ErrInvalidTaxConfig
	}

	// Check if rate is reasonable (e.g., < 10%)
	maxRate := big.NewFloat(0.1)
	if rateBig.Cmp(maxRate) > 0 {
		return types.ErrTaxRateTooHigh
	}

	// Store config in separate KV store
	if err := k.SetTransferTaxEnabled(ctx, enabled); err != nil {
		return err
	}
	if err := k.SetTransferTaxRate(ctx, rate); err != nil {
		return err
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

	if params.Fees != nil {
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
