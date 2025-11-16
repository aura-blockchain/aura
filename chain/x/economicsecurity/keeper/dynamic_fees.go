package keeper

import (
	"math/big"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// ============================
// DYNAMIC FEES (Feature 11)
// ============================

// RecordBlockUtilization records block utilization for fee calculation
func (k *Keeper) RecordBlockUtilization(utilization uint64) {
	k.mu.Lock()
	defer k.mu.Unlock()

	params := k.GetParams()

	k.blockUtilization = append(k.blockUtilization, utilization)

	// Keep only the window size
	if uint64(len(k.blockUtilization)) > params.DynamicFees.UtilizationWindow {
		k.blockUtilization = k.blockUtilization[1:]
	}

	// Update params with recent utilization
	params.DynamicFees.RecentUtilization = k.blockUtilization
	k.SetParams(params)
}

// AdjustDynamicFees adjusts fee multiplier based on network congestion
func (k *Keeper) AdjustDynamicFees() error {
	params := k.GetParams()

	if !params.DynamicFees.Enabled {
		return nil
	}

	if len(k.blockUtilization) == 0 {
		return nil
	}

	// Calculate average utilization
	totalUtilization := uint64(0)
	for _, util := range k.blockUtilization {
		totalUtilization += util
	}
	avgUtilization := totalUtilization / uint64(len(k.blockUtilization))

	targetUtilization := params.DynamicFees.TargetUtilization
	currentMultiplier := params.DynamicFees.CurrentMultiplier

	// Adjust multiplier based on deviation from target
	var newMultiplier uint64

	if avgUtilization > targetUtilization {
		// Increase fees if above target
		deviation := avgUtilization - targetUtilization
		adjustment := (deviation * params.DynamicFees.AdjustmentSpeed) / types.BasisPoints

		newMultiplier = currentMultiplier + adjustment
	} else {
		// Decrease fees if below target
		deviation := targetUtilization - avgUtilization
		adjustment := (deviation * params.DynamicFees.AdjustmentSpeed) / types.BasisPoints

		if adjustment > currentMultiplier {
			newMultiplier = params.DynamicFees.MinMultiplier
		} else {
			newMultiplier = currentMultiplier - adjustment
		}
	}

	// Clamp to min/max bounds
	if newMultiplier < params.DynamicFees.MinMultiplier {
		newMultiplier = params.DynamicFees.MinMultiplier
	}
	if newMultiplier > params.DynamicFees.MaxMultiplier {
		newMultiplier = params.DynamicFees.MaxMultiplier
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	params.DynamicFees.CurrentMultiplier = newMultiplier
	return k.SetParams(params)
}

// CalculateDynamicFee calculates the current fee for a transaction
func (k *Keeper) CalculateDynamicFee() string {
	params := k.GetParams()

	if !params.DynamicFees.Enabled {
		return params.DynamicFees.BaseFee
	}

	baseFee := new(big.Int)
	baseFee.SetString(params.DynamicFees.BaseFee, 10)

	multiplier := big.NewInt(int64(params.DynamicFees.CurrentMultiplier))

	fee := new(big.Int).Mul(baseFee, multiplier)
	fee.Div(fee, big.NewInt(types.BasisPoints))

	return fee.String()
}

// GetCurrentFeeMultiplier returns the current fee multiplier
func (k *Keeper) GetCurrentFeeMultiplier() uint64 {
	params := k.GetParams()
	return params.DynamicFees.CurrentMultiplier
}

// GetAverageUtilization returns the average block utilization
func (k *Keeper) GetAverageUtilization() uint64 {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if len(k.blockUtilization) == 0 {
		return 0
	}

	total := uint64(0)
	for _, util := range k.blockUtilization {
		total += util
	}

	return total / uint64(len(k.blockUtilization))
}
