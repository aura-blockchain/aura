// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// =============================================================================
// Liquidity Mining Tests
// =============================================================================

func TestDistributeLiquidityRewards_Success(t *testing.T) {
	params := types.DefaultParams()
	params.LiquidityMining = &types.LiquidityMiningConfig{
		Enabled:                 true,
		TotalRewardsAllocated:   "10000000000",
		TotalRewardsDistributed: "0",
		MaxRewardsPerEpoch:      "1000000000",
		CurrentEpoch:            0,
		LastDistributionHeight:  0,
		EpochDurationBlocks:     100,
		IrVerifiedMultiplier:    12000, // 1.2x
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	_ = k.SetCurrentHeight(ctx, 150) // Past epoch duration

	recipients := map[string]string{
		"aura1user1": "100000",
		"aura1user2": "200000",
	}
	irVerified := map[string]bool{
		"aura1user1": true,
		"aura1user2": false,
	}

	err := k.DistributeLiquidityRewards(ctx, recipients, irVerified)
	require.NoError(t, err)
}

func TestDistributeLiquidityRewards_Disabled(t *testing.T) {
	params := types.DefaultParams()
	params.LiquidityMining = &types.LiquidityMiningConfig{
		Enabled: false,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	recipients := map[string]string{"aura1user1": "100000"}
	irVerified := map[string]bool{}

	err := k.DistributeLiquidityRewards(ctx, recipients, irVerified)
	require.ErrorIs(t, err, types.ErrLiquidityMiningDisabled)
}

func TestDistributeLiquidityRewards_EpochNotReached(t *testing.T) {
	params := types.DefaultParams()
	params.LiquidityMining = &types.LiquidityMiningConfig{
		Enabled:                true,
		TotalRewardsAllocated:  "10000000000",
		MaxRewardsPerEpoch:     "1000000000",
		LastDistributionHeight: 100,
		EpochDurationBlocks:    100,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	_ = k.SetCurrentHeight(ctx, 150) // Not yet at 200

	recipients := map[string]string{"aura1user1": "100000"}
	err := k.DistributeLiquidityRewards(ctx, recipients, map[string]bool{})
	require.ErrorIs(t, err, types.ErrInvalidEpoch)
}

func TestDistributeLiquidityRewards_ExceedsEpochCap(t *testing.T) {
	params := types.DefaultParams()
	params.LiquidityMining = &types.LiquidityMiningConfig{
		Enabled:                 true,
		TotalRewardsAllocated:   "10000000000",
		TotalRewardsDistributed: "0",
		MaxRewardsPerEpoch:      "1000", // Very low cap
		LastDistributionHeight:  0,
		EpochDurationBlocks:     100,
		IrVerifiedMultiplier:    10000,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	_ = k.SetCurrentHeight(ctx, 150)

	recipients := map[string]string{"aura1user1": "100000"} // Exceeds cap
	err := k.DistributeLiquidityRewards(ctx, recipients, map[string]bool{})
	require.ErrorIs(t, err, types.ErrLiquidityRewardCapExceeded)
}

func TestDistributeLiquidityRewards_InsufficientAllocation(t *testing.T) {
	params := types.DefaultParams()
	params.LiquidityMining = &types.LiquidityMiningConfig{
		Enabled:                 true,
		TotalRewardsAllocated:   "1000",
		TotalRewardsDistributed: "900",
		MaxRewardsPerEpoch:      "1000000000",
		LastDistributionHeight:  0,
		EpochDurationBlocks:     100,
		IrVerifiedMultiplier:    10000,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	_ = k.SetCurrentHeight(ctx, 150)

	recipients := map[string]string{"aura1user1": "200"} // Would exceed total
	err := k.DistributeLiquidityRewards(ctx, recipients, map[string]bool{})
	require.ErrorIs(t, err, types.ErrInsufficientRewards)
}

func TestGetLiquidityMiningStats_Enabled(t *testing.T) {
	params := types.DefaultParams()
	params.LiquidityMining = &types.LiquidityMiningConfig{
		Enabled:                 true,
		TotalRewardsAllocated:   "10000000",
		TotalRewardsDistributed: "3000000",
		CurrentEpoch:            5,
		LastDistributionHeight:  500,
		EpochDurationBlocks:     100,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	enabled, allocated, distributed, remaining, epoch, nextHeight := k.GetLiquidityMiningStats(ctx)
	require.True(t, enabled)
	require.Equal(t, "10000000", allocated)
	require.Equal(t, "3000000", distributed)
	require.Equal(t, "7000000", remaining)
	require.Equal(t, uint64(5), epoch)
	require.Equal(t, uint64(600), nextHeight)
}

func TestGetLiquidityMiningStats_Disabled(t *testing.T) {
	params := types.DefaultParams()
	params.LiquidityMining = &types.LiquidityMiningConfig{
		Enabled: false,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	enabled, allocated, distributed, remaining, epoch, nextHeight := k.GetLiquidityMiningStats(ctx)
	require.False(t, enabled)
	require.Equal(t, "0", allocated)
	require.Equal(t, "0", distributed)
	require.Equal(t, "0", remaining)
	require.Equal(t, uint64(0), epoch)
	require.Equal(t, uint64(0), nextHeight)
}

func TestCheckLiquidityRewardCap_Success(t *testing.T) {
	params := types.DefaultParams()
	params.LiquidityMining = &types.LiquidityMiningConfig{
		Enabled:                 true,
		TotalRewardsAllocated:   "10000000",
		TotalRewardsDistributed: "1000000",
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	err := k.CheckLiquidityRewardCap(ctx, "5000000")
	require.NoError(t, err)
}

func TestCheckLiquidityRewardCap_Disabled(t *testing.T) {
	params := types.DefaultParams()
	params.LiquidityMining = &types.LiquidityMiningConfig{
		Enabled: false,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	err := k.CheckLiquidityRewardCap(ctx, "1000")
	require.ErrorIs(t, err, types.ErrLiquidityMiningDisabled)
}

func TestCheckLiquidityRewardCap_ExceedsCap(t *testing.T) {
	params := types.DefaultParams()
	params.LiquidityMining = &types.LiquidityMiningConfig{
		Enabled:                 true,
		TotalRewardsAllocated:   "10000000",
		TotalRewardsDistributed: "9000000",
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	err := k.CheckLiquidityRewardCap(ctx, "5000000") // Would exceed
	require.ErrorIs(t, err, types.ErrLiquidityRewardCapExceeded)
}

func TestCanDistributeRewards_Success(t *testing.T) {
	params := types.DefaultParams()
	params.LiquidityMining = &types.LiquidityMiningConfig{
		Enabled:                true,
		LastDistributionHeight: 0,
		EpochDurationBlocks:    100,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	_ = k.SetCurrentHeight(ctx, 150)

	canDistribute, err := k.CanDistributeRewards(ctx)
	require.NoError(t, err)
	require.True(t, canDistribute)
}

func TestCanDistributeRewards_NotYet(t *testing.T) {
	params := types.DefaultParams()
	params.LiquidityMining = &types.LiquidityMiningConfig{
		Enabled:                true,
		LastDistributionHeight: 100,
		EpochDurationBlocks:    100,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	_ = k.SetCurrentHeight(ctx, 150) // Not yet at 200

	canDistribute, err := k.CanDistributeRewards(ctx)
	require.NoError(t, err)
	require.False(t, canDistribute)
}

func TestCanDistributeRewards_Disabled(t *testing.T) {
	params := types.DefaultParams()
	params.LiquidityMining = &types.LiquidityMiningConfig{
		Enabled: false,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	canDistribute, err := k.CanDistributeRewards(ctx)
	require.NoError(t, err)
	require.False(t, canDistribute)
}

func TestGetRemainingRewards_Success(t *testing.T) {
	params := types.DefaultParams()
	params.LiquidityMining = &types.LiquidityMiningConfig{
		TotalRewardsAllocated:   "10000000",
		TotalRewardsDistributed: "3000000",
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	remaining, err := k.GetRemainingRewards(ctx)
	require.NoError(t, err)
	require.Equal(t, "7000000", remaining)
}

func TestCalculateRewardWithMultiplier_WithMultiplier(t *testing.T) {
	params := types.DefaultParams()
	params.LiquidityMining = &types.LiquidityMiningConfig{
		IrVerifiedMultiplier: 12000, // 1.2x (12000 basis points)
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	result, err := k.CalculateRewardWithMultiplier(ctx, "1000000", true)
	require.NoError(t, err)
	require.Equal(t, "1200000", result) // 1.2x
}

func TestCalculateRewardWithMultiplier_WithoutMultiplier(t *testing.T) {
	params := types.DefaultParams()
	params.LiquidityMining = &types.LiquidityMiningConfig{
		IrVerifiedMultiplier: 12000,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	result, err := k.CalculateRewardWithMultiplier(ctx, "1000000", false)
	require.NoError(t, err)
	require.Equal(t, "1000000", result) // No multiplier
}

func TestCalculateRewardWithMultiplier_InvalidAmount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	_, err := k.CalculateRewardWithMultiplier(ctx, "not-a-number", true)
	require.ErrorIs(t, err, types.ErrInvalidAmount)
}

func TestGetEpochInfo_Success(t *testing.T) {
	params := types.DefaultParams()
	params.LiquidityMining = &types.LiquidityMiningConfig{
		CurrentEpoch:           5,
		LastDistributionHeight: 500,
		EpochDurationBlocks:    100,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	_ = k.SetCurrentHeight(ctx, 550)

	currentEpoch, lastDist, nextDist, blocksRemaining, err := k.GetEpochInfo(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(5), currentEpoch)
	require.Equal(t, uint64(500), lastDist)
	require.Equal(t, uint64(600), nextDist)
	require.Equal(t, uint64(50), blocksRemaining)
}

func TestGetEpochInfo_EpochPassed(t *testing.T) {
	params := types.DefaultParams()
	params.LiquidityMining = &types.LiquidityMiningConfig{
		CurrentEpoch:           5,
		LastDistributionHeight: 500,
		EpochDurationBlocks:    100,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	_ = k.SetCurrentHeight(ctx, 650) // Past next distribution

	_, _, _, blocksRemaining, err := k.GetEpochInfo(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(0), blocksRemaining)
}

func TestEstimateEpochRewards_Success(t *testing.T) {
	params := types.DefaultParams()
	params.LiquidityMining = &types.LiquidityMiningConfig{
		IrVerifiedMultiplier: 15000, // 1.5x
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	recipients := map[string]string{
		"aura1user1": "1000000",
		"aura1user2": "2000000",
	}
	irVerified := map[string]bool{
		"aura1user1": true,
		"aura1user2": false,
	}

	totalBase, totalWithMultipliers, err := k.EstimateEpochRewards(ctx, recipients, irVerified)
	require.NoError(t, err)
	require.Equal(t, "3000000", totalBase)
	require.Equal(t, "3500000", totalWithMultipliers) // 1.5M + 2M
}

func TestEstimateEpochRewards_InvalidAmount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	recipients := map[string]string{
		"aura1user1": "not-a-number",
	}

	_, _, err := k.EstimateEpochRewards(ctx, recipients, map[string]bool{})
	require.ErrorIs(t, err, types.ErrInvalidAmount)
}

func TestUpdateLiquidityMiningConfig_Success(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	err := k.UpdateLiquidityMiningConfig(ctx, true, 200, 15000)
	require.NoError(t, err)

	params, _ := k.GetParams(ctx)
	require.True(t, params.LiquidityMining.Enabled)
	require.Equal(t, uint64(200), params.LiquidityMining.EpochDurationBlocks)
	require.Equal(t, uint64(15000), params.LiquidityMining.IrVerifiedMultiplier)
}

func TestUpdateLiquidityMiningConfig_InvalidDuration(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	err := k.UpdateLiquidityMiningConfig(ctx, true, 0, 15000) // Invalid duration
	require.ErrorIs(t, err, types.ErrInvalidDuration)
}

func TestUpdateLiquidityMiningConfig_InvalidMultiplier(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	err := k.UpdateLiquidityMiningConfig(ctx, true, 100, 5000) // Below basis points
	require.ErrorIs(t, err, types.ErrInvalidAmount)
}

func TestIncreaseTotalRewardsAllocated_Success(t *testing.T) {
	params := types.DefaultParams()
	params.LiquidityMining = &types.LiquidityMiningConfig{
		TotalRewardsAllocated: "10000000",
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	err := k.IncreaseTotalRewardsAllocated(ctx, "5000000")
	require.NoError(t, err)

	newParams, _ := k.GetParams(ctx)
	require.Equal(t, "15000000", newParams.LiquidityMining.TotalRewardsAllocated)
}

func TestIncreaseTotalRewardsAllocated_InvalidAmount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	err := k.IncreaseTotalRewardsAllocated(ctx, "not-a-number")
	require.ErrorIs(t, err, types.ErrInvalidAmount)
}

func TestIncreaseTotalRewardsAllocated_ZeroAmount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	err := k.IncreaseTotalRewardsAllocated(ctx, "0")
	require.ErrorIs(t, err, types.ErrInvalidAmount)
}

func TestIncreaseTotalRewardsAllocated_NegativeAmount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	err := k.IncreaseTotalRewardsAllocated(ctx, "-1000")
	require.ErrorIs(t, err, types.ErrInvalidAmount)
}
