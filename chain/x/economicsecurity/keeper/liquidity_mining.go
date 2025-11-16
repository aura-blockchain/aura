package keeper

import (
	"math/big"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// ============================
// LIQUIDITY MINING (Feature 3)
// ============================

// DistributeLiquidityRewards distributes rewards for the current epoch
func (k *Keeper) DistributeLiquidityRewards(recipients map[string]string, irVerifiedUsers map[string]bool) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	params := k.GetParams()

	if !params.LiquidityMining.Enabled {
		return types.ErrLiquidityMiningDisabled
	}

	// Check if we've reached distribution height
	if k.currentHeight < params.LiquidityMining.LastDistributionHeight+params.LiquidityMining.EpochDurationBlocks {
		return types.ErrInvalidEpoch
	}

	totalAllocated := new(big.Int)
	totalAllocated.SetString(params.LiquidityMining.TotalRewardsAllocated, 10)

	totalDistributed := new(big.Int)
	totalDistributed.SetString(params.LiquidityMining.TotalRewardsDistributed, 10)

	maxPerEpoch := new(big.Int)
	maxPerEpoch.SetString(params.LiquidityMining.MaxRewardsPerEpoch, 10)

	// Calculate total rewards to distribute this epoch
	epochRewards := big.NewInt(0)
	for _, rewardStr := range recipients {
		reward := new(big.Int)
		if _, ok := reward.SetString(rewardStr, 10); !ok {
			return types.ErrInvalidAmount
		}

		epochRewards.Add(epochRewards, reward)
	}

	// Apply IR-verified multiplier
	adjustedRewards := big.NewInt(0)
	for addr, rewardStr := range recipients {
		reward := new(big.Int)
		reward.SetString(rewardStr, 10)

		if irVerifiedUsers[addr] {
			// Apply IR multiplier
			multiplier := big.NewInt(int64(params.LiquidityMining.IrVerifiedMultiplier))
			reward.Mul(reward, multiplier)
			reward.Div(reward, big.NewInt(types.BasisPoints))
		}

		adjustedRewards.Add(adjustedRewards, reward)
	}

	// Check cap per epoch
	if adjustedRewards.Cmp(maxPerEpoch) > 0 {
		return types.ErrLiquidityRewardCapExceeded
	}

	// Check total allocated cap
	newTotal := new(big.Int).Add(totalDistributed, adjustedRewards)
	if newTotal.Cmp(totalAllocated) > 0 {
		return types.ErrInsufficientRewards
	}

	// Update state
	params.LiquidityMining.TotalRewardsDistributed = newTotal.String()
	params.LiquidityMining.CurrentEpoch++
	params.LiquidityMining.LastDistributionHeight = k.currentHeight

	return k.SetParams(params)
}

// GetLiquidityMiningStats returns liquidity mining statistics
func (k *Keeper) GetLiquidityMiningStats() (bool, string, string, string, uint64, uint64) {
	params := k.GetParams()

	if !params.LiquidityMining.Enabled {
		return false, "0", "0", "0", 0, 0
	}

	totalAllocated := new(big.Int)
	totalAllocated.SetString(params.LiquidityMining.TotalRewardsAllocated, 10)

	totalDistributed := new(big.Int)
	totalDistributed.SetString(params.LiquidityMining.TotalRewardsDistributed, 10)

	remaining := new(big.Int).Sub(totalAllocated, totalDistributed)

	nextDistHeight := params.LiquidityMining.LastDistributionHeight + params.LiquidityMining.EpochDurationBlocks

	return true,
		params.LiquidityMining.TotalRewardsAllocated,
		params.LiquidityMining.TotalRewardsDistributed,
		remaining.String(),
		params.LiquidityMining.CurrentEpoch,
		nextDistHeight
}

// CheckLiquidityRewardCap checks if distributing rewards would exceed cap
func (k *Keeper) CheckLiquidityRewardCap(amount string) error {
	params := k.GetParams()

	if !params.LiquidityMining.Enabled {
		return types.ErrLiquidityMiningDisabled
	}

	totalAllocated := new(big.Int)
	totalAllocated.SetString(params.LiquidityMining.TotalRewardsAllocated, 10)

	totalDistributed := new(big.Int)
	totalDistributed.SetString(params.LiquidityMining.TotalRewardsDistributed, 10)

	rewardAmt := new(big.Int)
	if _, ok := rewardAmt.SetString(amount, 10); !ok {
		return types.ErrInvalidAmount
	}

	newTotal := new(big.Int).Add(totalDistributed, rewardAmt)
	if newTotal.Cmp(totalAllocated) > 0 {
		return types.ErrLiquidityRewardCapExceeded
	}

	return nil
}
