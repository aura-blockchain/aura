package keeper

import (
	"math/big"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// ============================
// DYNAMIC FEES (Feature 11)
// ============================

// RecordBlockUtilization records block utilization for fee calculation
func (k *Keeper) RecordBlockUtilization(utilization uint64) error {
	params := k.GetParams()

	// Get existing utilization data from params (source of truth)
	blockUtilization := params.DynamicFees.RecentUtilization
	blockUtilization = append(blockUtilization, utilization)

	// Keep only the window size
	if uint64(len(blockUtilization)) > params.DynamicFees.UtilizationWindow {
		blockUtilization = blockUtilization[1:]
	}

	// Update params with recent utilization
	params.DynamicFees.RecentUtilization = blockUtilization
	return k.SetParams(params)
}

// AdjustDynamicFees adjusts fee multiplier based on network congestion
func (k *Keeper) AdjustDynamicFees() error {
	params := k.GetParams()

	if !params.DynamicFees.Enabled {
		return nil
	}

	// Use the utilization data from params (which is persisted)
	utilizationData := params.DynamicFees.RecentUtilization
	if len(utilizationData) == 0 {
		return nil
	}

	// Calculate average utilization
	totalUtilization := uint64(0)
	for _, util := range utilizationData {
		totalUtilization += util
	}
	avgUtilization := totalUtilization / uint64(len(utilizationData))

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
	if _, ok := baseFee.SetString(params.DynamicFees.BaseFee, 10); !ok {
		return params.DynamicFees.BaseFee
	}

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
	params := k.GetParams()
	utilizationData := params.DynamicFees.RecentUtilization

	if len(utilizationData) == 0 {
		return 0
	}

	total := uint64(0)
	for _, util := range utilizationData {
		total += util
	}

	return total / uint64(len(utilizationData))
}

// GetFeeMultiplierHistory returns recent fee multiplier adjustments
func (k *Keeper) GetFeeMultiplierHistory() []uint64 {
	params := k.GetParams()

	// Return recent utilization as proxy for fee history
	// In production, you might store actual multiplier history
	return params.DynamicFees.RecentUtilization
}

// PredictFee estimates the fee at a future utilization level
func (k *Keeper) PredictFee(futureUtilization uint64) (string, error) {
	params := k.GetParams()

	if !params.DynamicFees.Enabled {
		return params.DynamicFees.BaseFee, nil
	}

	targetUtilization := params.DynamicFees.TargetUtilization
	currentMultiplier := params.DynamicFees.CurrentMultiplier

	// Simulate multiplier adjustment
	var predictedMultiplier uint64

	if futureUtilization > targetUtilization {
		deviation := futureUtilization - targetUtilization
		adjustment := (deviation * params.DynamicFees.AdjustmentSpeed) / types.BasisPoints
		predictedMultiplier = currentMultiplier + adjustment
	} else {
		deviation := targetUtilization - futureUtilization
		adjustment := (deviation * params.DynamicFees.AdjustmentSpeed) / types.BasisPoints

		if adjustment > currentMultiplier {
			predictedMultiplier = params.DynamicFees.MinMultiplier
		} else {
			predictedMultiplier = currentMultiplier - adjustment
		}
	}

	// Clamp to bounds
	if predictedMultiplier < params.DynamicFees.MinMultiplier {
		predictedMultiplier = params.DynamicFees.MinMultiplier
	}
	if predictedMultiplier > params.DynamicFees.MaxMultiplier {
		predictedMultiplier = params.DynamicFees.MaxMultiplier
	}

	// Calculate predicted fee
	baseFee := new(big.Int)
	if _, ok := baseFee.SetString(params.DynamicFees.BaseFee, 10); !ok {
		return "", types.ErrInvalidAmount
	}

	multiplier := big.NewInt(int64(predictedMultiplier))
	fee := new(big.Int).Mul(baseFee, multiplier)
	fee.Div(fee, big.NewInt(types.BasisPoints))

	return fee.String(), nil
}

// SetBaseFee updates the base fee (requires governance authority)
func (k *Keeper) SetBaseFee(baseFee string) error {
	params := k.GetParams()

	// Validate base fee
	bf := new(big.Int)
	if _, ok := bf.SetString(baseFee, 10); !ok {
		return types.ErrInvalidAmount
	}

	if bf.Sign() <= 0 {
		return types.ErrInvalidAmount
	}

	params.DynamicFees.BaseFee = baseFee
	return k.SetParams(params)
}

// ResetFeeMultiplier resets the fee multiplier to base value
func (k *Keeper) ResetFeeMultiplier() error {
	params := k.GetParams()

	// Reset to base multiplier (10000 basis points = 1.0x)
	params.DynamicFees.CurrentMultiplier = types.BasisPoints
	return k.SetParams(params)
}

// GetFeeStatistics returns comprehensive fee statistics
func (k *Keeper) GetFeeStatistics() (currentFee string, multiplier uint64, avgUtilization uint64, baseFee string) {
	params := k.GetParams()

	currentFee = k.CalculateDynamicFee()
	multiplier = params.DynamicFees.CurrentMultiplier
	avgUtilization = k.GetAverageUtilization()
	baseFee = params.DynamicFees.BaseFee

	return
}
