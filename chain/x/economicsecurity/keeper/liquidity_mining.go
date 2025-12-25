// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"math/big"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// ============================
// LIQUIDITY MINING (Feature 3)
// ============================

// DistributeLiquidityRewards distributes rewards for the current epoch
func (k *Keeper) DistributeLiquidityRewards(ctx context.Context, recipients map[string]string, irVerifiedUsers map[string]bool) error {
	params, _ := k.GetParams(ctx)

	if !params.LiquidityMining.Enabled {
		return types.ErrLiquidityMiningDisabled
	}

	// Get current height
	currentHeight, err := k.GetCurrentHeight(ctx)
	if err != nil {
		return err
	}

	// Check if we've reached distribution height
	if currentHeight < params.LiquidityMining.LastDistributionHeight+params.LiquidityMining.EpochDurationBlocks {
		return types.ErrInvalidEpoch
	}

	totalAllocated := new(big.Int)
	if _, ok := totalAllocated.SetString(params.LiquidityMining.TotalRewardsAllocated, 10); !ok {
		return types.ErrInvalidAmount
	}

	totalDistributed := new(big.Int)
	if _, ok := totalDistributed.SetString(params.LiquidityMining.TotalRewardsDistributed, 10); !ok {
		return types.ErrInvalidAmount
	}

	maxPerEpoch := new(big.Int)
	if _, ok := maxPerEpoch.SetString(params.LiquidityMining.MaxRewardsPerEpoch, 10); !ok {
		return types.ErrInvalidAmount
	}

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
		if _, ok := reward.SetString(rewardStr, 10); !ok {
			return types.ErrInvalidAmount
		}

		if irVerifiedUsers[addr] {
			// Apply IR multiplier (e.g., 12000 basis points = 1.2x)
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
	params.LiquidityMining.LastDistributionHeight = currentHeight

	return k.SetParams(params)
}

// GetLiquidityMiningStats returns liquidity mining statistics
func (k *Keeper) GetLiquidityMiningStats(ctx context.Context) (bool, string, string, string, uint64, uint64) {
	params, _ := k.GetParams(ctx)

	if !params.LiquidityMining.Enabled {
		return false, "0", "0", "0", 0, 0
	}

	totalAllocated := new(big.Int)
	if _, ok := totalAllocated.SetString(params.LiquidityMining.TotalRewardsAllocated, 10); !ok {
		totalAllocated = big.NewInt(0)
	}

	totalDistributed := new(big.Int)
	if _, ok := totalDistributed.SetString(params.LiquidityMining.TotalRewardsDistributed, 10); !ok {
		totalDistributed = big.NewInt(0)
	}

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
func (k *Keeper) CheckLiquidityRewardCap(ctx context.Context, amount string) error {
	params, _ := k.GetParams(ctx)

	if !params.LiquidityMining.Enabled {
		return types.ErrLiquidityMiningDisabled
	}

	totalAllocated := new(big.Int)
	if _, ok := totalAllocated.SetString(params.LiquidityMining.TotalRewardsAllocated, 10); !ok {
		return types.ErrInvalidAmount
	}

	totalDistributed := new(big.Int)
	if _, ok := totalDistributed.SetString(params.LiquidityMining.TotalRewardsDistributed, 10); !ok {
		return types.ErrInvalidAmount
	}

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

// CanDistributeRewards checks if rewards can be distributed in the current epoch
func (k *Keeper) CanDistributeRewards(ctx context.Context) (bool, error) {
	params, _ := k.GetParams(ctx)

	if !params.LiquidityMining.Enabled {
		return false, nil
	}

	currentHeight, err := k.GetCurrentHeight(ctx)
	if err != nil {
		return false, err
	}

	// Check if epoch duration has passed
	nextDistHeight := params.LiquidityMining.LastDistributionHeight + params.LiquidityMining.EpochDurationBlocks
	return currentHeight >= nextDistHeight, nil
}

// GetRemainingRewards returns the amount of rewards remaining to be distributed
func (k *Keeper) GetRemainingRewards(ctx context.Context) (string, error) {
	params, _ := k.GetParams(ctx)

	totalAllocated := new(big.Int)
	if _, ok := totalAllocated.SetString(params.LiquidityMining.TotalRewardsAllocated, 10); !ok {
		return "0", types.ErrInvalidAmount
	}

	totalDistributed := new(big.Int)
	if _, ok := totalDistributed.SetString(params.LiquidityMining.TotalRewardsDistributed, 10); !ok {
		return "0", types.ErrInvalidAmount
	}

	remaining := new(big.Int).Sub(totalAllocated, totalDistributed)
	return remaining.String(), nil
}

// CalculateRewardWithMultiplier calculates reward amount with IR multiplier applied
func (k *Keeper) CalculateRewardWithMultiplier(ctx context.Context, baseReward string, isIRVerified bool) (string, error) {
	params, _ := k.GetParams(ctx)

	reward := new(big.Int)
	if _, ok := reward.SetString(baseReward, 10); !ok {
		return "0", types.ErrInvalidAmount
	}

	if isIRVerified {
		multiplier := big.NewInt(int64(params.LiquidityMining.IrVerifiedMultiplier))
		reward.Mul(reward, multiplier)
		reward.Div(reward, big.NewInt(types.BasisPoints))
	}

	return reward.String(), nil
}

// GetEpochInfo returns information about the current epoch
func (k *Keeper) GetEpochInfo(ctx context.Context) (currentEpoch uint64, lastDistHeight uint64, nextDistHeight uint64, blocksRemaining uint64, err error) {
	params, _ := k.GetParams(ctx)

	currentHeight, err := k.GetCurrentHeight(ctx)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	currentEpoch = params.LiquidityMining.CurrentEpoch
	lastDistHeight = params.LiquidityMining.LastDistributionHeight
	nextDistHeight = lastDistHeight + params.LiquidityMining.EpochDurationBlocks

	if currentHeight >= nextDistHeight {
		blocksRemaining = 0
	} else {
		blocksRemaining = nextDistHeight - currentHeight
	}

	return currentEpoch, lastDistHeight, nextDistHeight, blocksRemaining, nil
}

// EstimateEpochRewards estimates total rewards for an epoch based on recipient map
func (k *Keeper) EstimateEpochRewards(ctx context.Context, recipients map[string]string, irVerifiedUsers map[string]bool) (totalBase string, totalWithMultipliers string, err error) {
	params, _ := k.GetParams(ctx)

	baseRewards := big.NewInt(0)
	adjustedRewards := big.NewInt(0)

	for addr, rewardStr := range recipients {
		reward := new(big.Int)
		if _, ok := reward.SetString(rewardStr, 10); !ok {
			return "0", "0", types.ErrInvalidAmount
		}

		baseRewards.Add(baseRewards, reward)

		if irVerifiedUsers[addr] {
			multiplier := big.NewInt(int64(params.LiquidityMining.IrVerifiedMultiplier))
			adjustedReward := new(big.Int).Mul(reward, multiplier)
			adjustedReward.Div(adjustedReward, big.NewInt(types.BasisPoints))
			adjustedRewards.Add(adjustedRewards, adjustedReward)
		} else {
			adjustedRewards.Add(adjustedRewards, reward)
		}
	}

	return baseRewards.String(), adjustedRewards.String(), nil
}

// UpdateLiquidityMiningConfig updates liquidity mining parameters
func (k *Keeper) UpdateLiquidityMiningConfig(ctx context.Context, enabled bool, epochDuration uint64, irMultiplier uint64) error {
	params, _ := k.GetParams(ctx)

	if epochDuration == 0 {
		return types.ErrInvalidDuration
	}

	if irMultiplier < types.BasisPoints {
		return types.ErrInvalidAmount
	}

	params.LiquidityMining.Enabled = enabled
	params.LiquidityMining.EpochDurationBlocks = epochDuration
	params.LiquidityMining.IrVerifiedMultiplier = irMultiplier

	return k.SetParams(params)
}

// IncreaseTotalRewardsAllocated increases the total reward allocation
func (k *Keeper) IncreaseTotalRewardsAllocated(ctx context.Context, additionalRewards string) error {
	params, _ := k.GetParams(ctx)

	currentAllocated := new(big.Int)
	if _, ok := currentAllocated.SetString(params.LiquidityMining.TotalRewardsAllocated, 10); !ok {
		return types.ErrInvalidAmount
	}

	additional := new(big.Int)
	if _, ok := additional.SetString(additionalRewards, 10); !ok {
		return types.ErrInvalidAmount
	}

	if additional.Sign() <= 0 {
		return types.ErrInvalidAmount
	}

	newAllocated := new(big.Int).Add(currentAllocated, additional)
	params.LiquidityMining.TotalRewardsAllocated = newAllocated.String()

	return k.SetParams(params)
}
