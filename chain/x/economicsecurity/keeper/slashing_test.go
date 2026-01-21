// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// =============================================================================
// analyzeStakingIncentives Tests
// =============================================================================

func TestAnalyzeStakingIncentives_NormalStake(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.CirculatingSupply = "1000000000" // 1 billion
	params.Tokenomics.InflationRate = 500              // 5%

	result := k.analyzeStakingIncentives(*params, "500000000") // 50% staked

	rewards, ok := new(big.Int).SetString(result.rewards, 10)
	require.True(t, ok, "rewards should parse as big.Int")
	require.True(t, rewards.Sign() >= 0, "rewards must be non-negative")

	// With 50% staked and 5% inflation: 500M * 500 / 10000 = 25M rewards
	expected := new(big.Int).SetInt64(25000000)
	require.Equal(t, expected.String(), rewards.String())
}

func TestAnalyzeStakingIncentives_ZeroStake(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.CirculatingSupply = "1000000000"
	params.Tokenomics.InflationRate = 500

	result := k.analyzeStakingIncentives(*params, "0")

	rewards, ok := new(big.Int).SetString(result.rewards, 10)
	require.True(t, ok, "rewards should parse as big.Int")
	require.Equal(t, "0", rewards.String())
}

func TestAnalyzeStakingIncentives_MaximumStake(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.CirculatingSupply = "1000000000"
	params.Tokenomics.InflationRate = 500

	// 100% staked (equal to circulating supply)
	result := k.analyzeStakingIncentives(*params, "1000000000")

	rewards, ok := new(big.Int).SetString(result.rewards, 10)
	require.True(t, ok, "rewards should parse as big.Int")
	require.True(t, rewards.Sign() >= 0, "rewards must be non-negative")

	// With 100% staked and 5% inflation: 1B * 500 / 10000 = 50M rewards
	expected := new(big.Int).SetInt64(50000000)
	require.Equal(t, expected.String(), rewards.String())
}

func TestAnalyzeStakingIncentives_LowStakingRatio(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.CirculatingSupply = "1000000000"
	params.Tokenomics.InflationRate = 500

	// 10% staked - should trigger recommendation to increase staking rewards
	result := k.analyzeStakingIncentives(*params, "100000000")

	require.Contains(t, result.recommendations, "Increase staking rewards to incentivize more staking (current ratio: <30%)")
}

func TestAnalyzeStakingIncentives_HighStakingRatio(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.CirculatingSupply = "1000000000"
	params.Tokenomics.InflationRate = 500

	// 80% staked - should trigger recommendation to reduce staking rewards
	result := k.analyzeStakingIncentives(*params, "800000000")

	require.Contains(t, result.recommendations, "Consider reducing staking rewards as ratio is high (>70%)")
}

func TestAnalyzeStakingIncentives_InflationAboveTarget(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.CirculatingSupply = "1000000000"
	params.Tokenomics.InflationRate = 1500      // 15%
	params.Tokenomics.TargetInflationRate = 500 // 5%

	result := k.analyzeStakingIncentives(*params, "500000000")

	require.Contains(t, result.recommendations, "Inflation above target - reduce staking emission rate")
}

func TestAnalyzeStakingIncentives_FractionalAmounts(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.CirculatingSupply = "1000000000000000000" // 1e18 (like ERC20 tokens)
	params.Tokenomics.InflationRate = 333                       // 3.33%

	// Fractional stake amount
	result := k.analyzeStakingIncentives(*params, "123456789012345678")

	rewards, ok := new(big.Int).SetString(result.rewards, 10)
	require.True(t, ok, "rewards should parse as big.Int")
	require.True(t, rewards.Sign() >= 0, "rewards must be non-negative")
}

func TestAnalyzeStakingIncentives_VeryLargeValues(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	// Very large supply to test for overflow protection
	params.Tokenomics.CirculatingSupply = "999999999999999999999999999999999999999" // ~1e39
	params.Tokenomics.InflationRate = 9000                                          // 90%

	result := k.analyzeStakingIncentives(*params, "500000000000000000000000000000000000000")

	rewards, ok := new(big.Int).SetString(result.rewards, 10)
	require.True(t, ok, "rewards should parse as big.Int")
	require.False(t, rewards.Sign() < 0, "rewards must be non-negative")
}

func TestAnalyzeStakingIncentives_ZeroInflation(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.CirculatingSupply = "1000000000"
	params.Tokenomics.InflationRate = 0

	result := k.analyzeStakingIncentives(*params, "500000000")

	rewards, ok := new(big.Int).SetString(result.rewards, 10)
	require.True(t, ok, "rewards should parse as big.Int")
	require.Equal(t, "0", rewards.String())
}

func TestAnalyzeStakingIncentives_HighInflation(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.CirculatingSupply = "1000000000"
	params.Tokenomics.InflationRate = 10000 // 100%

	result := k.analyzeStakingIncentives(*params, "500000000")

	rewards, ok := new(big.Int).SetString(result.rewards, 10)
	require.True(t, ok, "rewards should parse as big.Int")

	// 100% inflation on 500M staked = 500M rewards
	expected := new(big.Int).SetInt64(500000000)
	require.Equal(t, expected.String(), rewards.String())
}

// =============================================================================
// AnalyzeEconomicIncentives Tests
// =============================================================================

func TestAnalyzeEconomicIncentives_Success(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Set up params with known values
	params, _ := k.GetParams(ctx)
	params.Tokenomics.CirculatingSupply = "1000000000"
	params.Tokenomics.InflationRate = 500
	params.Mev.TotalMevCaptured = "10000000"
	params.Mev.UserRedistributionPercentage = 4000
	params.Mev.TreasuryPercentage = 1000
	params.Mev.BurnPercentage = 0
	params.Mev.Strategy = 0
	require.NoError(t, k.SetParams(params))

	result, err := k.AnalyzeEconomicIncentives(ctx, "500000000", "100000000", 500, 50)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify validator rewards are computed
	validatorRewards, ok := new(big.Int).SetString(result.ValidatorRewards, 10)
	require.True(t, ok)
	require.True(t, validatorRewards.Sign() >= 0)

	// Verify user rewards are computed (40% of MEV)
	require.NotEmpty(t, result.UserRewards)

	// Verify treasury allocation
	require.NotEmpty(t, result.TreasuryAllocation)

	// Verify incentive efficiency score (now in basis points: 0-10000)
	require.LessOrEqual(t, result.IncentiveEfficiencyBps, uint64(10000))
}

func TestAnalyzeEconomicIncentives_LowUserCount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Tokenomics.CirculatingSupply = "1000000000"
	params.Tokenomics.InflationRate = 500
	params.Mev.TotalMevCaptured = "10000000"
	params.Mev.Strategy = 0
	require.NoError(t, k.SetParams(params))

	// Very low user count - should affect efficiency score
	result, err := k.AnalyzeEconomicIncentives(ctx, "500000000", "100000000", 10, 50)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should have recommendation about low user participation
	hasLowUserRecommendation := false
	for _, rec := range result.RecommendedAdjustments {
		if rec == "Low user participation - increase user reward percentage" {
			hasLowUserRecommendation = true
			break
		}
	}
	require.True(t, hasLowUserRecommendation)
}

func TestAnalyzeEconomicIncentives_LowValidatorCount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Tokenomics.CirculatingSupply = "1000000000"
	params.Tokenomics.InflationRate = 500
	params.Mev.TotalMevCaptured = "10000000"
	params.Mev.Strategy = 0
	require.NoError(t, k.SetParams(params))

	// Very low validator count - should affect efficiency score
	result, err := k.AnalyzeEconomicIncentives(ctx, "500000000", "100000000", 500, 5)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Efficiency should be lower with few validators (less than 10000 bps = 100%)
	require.Less(t, result.IncentiveEfficiencyBps, uint64(10000))
}

func TestAnalyzeEconomicIncentives_LowTreasuryPercentage(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Tokenomics.CirculatingSupply = "1000000000"
	params.Tokenomics.InflationRate = 500
	params.Mev.TotalMevCaptured = "10000000"
	params.Mev.TreasuryPercentage = 500 // 5% - below 10%
	params.Mev.UserRedistributionPercentage = 4500
	params.Mev.ValidatorPercentage = 5000
	params.Mev.BurnPercentage = 0
	params.Mev.Strategy = 0
	require.NoError(t, k.SetParams(params))

	result, err := k.AnalyzeEconomicIncentives(ctx, "500000000", "100000000", 500, 50)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should have recommendation about low treasury
	hasLowTreasuryRecommendation := false
	for _, rec := range result.RecommendedAdjustments {
		if rec == "Treasury allocation is low (<10%) - may need more funding for development" {
			hasLowTreasuryRecommendation = true
			break
		}
	}
	require.True(t, hasLowTreasuryRecommendation)
}

func TestAnalyzeEconomicIncentives_HighTreasuryPercentage(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Tokenomics.CirculatingSupply = "1000000000"
	params.Tokenomics.InflationRate = 500
	params.Mev.TotalMevCaptured = "10000000"
	params.Mev.TreasuryPercentage = 4000 // 40% - above 30%
	params.Mev.UserRedistributionPercentage = 3000
	params.Mev.ValidatorPercentage = 3000
	params.Mev.BurnPercentage = 0
	params.Mev.Strategy = 0
	require.NoError(t, k.SetParams(params))

	result, err := k.AnalyzeEconomicIncentives(ctx, "500000000", "100000000", 500, 50)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should have recommendation about high treasury
	hasHighTreasuryRecommendation := false
	for _, rec := range result.RecommendedAdjustments {
		if rec == "Treasury allocation is high (>30%) - consider redirecting more to users/validators" {
			hasHighTreasuryRecommendation = true
			break
		}
	}
	require.True(t, hasHighTreasuryRecommendation)
}

// =============================================================================
// CalculateOptimalIncentiveDistribution Tests
// =============================================================================

func TestCalculateOptimalIncentiveDistribution_NormalConditions(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	distribution, err := k.CalculateOptimalIncentiveDistribution(ctx, "1000000", uint64(5000), 500, 50) // 50% in basis points
	require.NoError(t, err)
	require.NotNil(t, distribution)

	// Check all distribution keys exist
	require.Contains(t, distribution, "validators")
	require.Contains(t, distribution, "users")
	require.Contains(t, distribution, "treasury")
	require.Contains(t, distribution, "burn")
	require.Contains(t, distribution, "reasoning")

	// Verify amounts are valid big.Int strings
	validatorRewards, ok := new(big.Int).SetString(distribution["validators"], 10)
	require.True(t, ok)
	require.True(t, validatorRewards.Sign() >= 0)

	userRewards, ok := new(big.Int).SetString(distribution["users"], 10)
	require.True(t, ok)
	require.True(t, userRewards.Sign() >= 0)

	treasuryRewards, ok := new(big.Int).SetString(distribution["treasury"], 10)
	require.True(t, ok)
	require.True(t, treasuryRewards.Sign() >= 0)

	burnAmount, ok := new(big.Int).SetString(distribution["burn"], 10)
	require.True(t, ok)
	require.True(t, burnAmount.Sign() >= 0)
}

func TestCalculateOptimalIncentiveDistribution_LargeRewards(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Very large reward amount
	distribution, err := k.CalculateOptimalIncentiveDistribution(ctx, "1000000000000000000000000", uint64(5000), 500, 50) // 50% in basis points
	require.NoError(t, err)
	require.NotNil(t, distribution)

	// Verify all amounts parse correctly
	for key, value := range distribution {
		if key == "reasoning" {
			continue
		}
		_, ok := new(big.Int).SetString(value, 10)
		require.True(t, ok, "failed to parse %s: %s", key, value)
	}
}

// =============================================================================
// SimulateIncentiveChange Tests
// =============================================================================

func TestSimulateIncentiveChange_Success(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Tokenomics.CirculatingSupply = "1000000000"
	params.Mev.TotalMevCaptured = "10000000"
	require.NoError(t, k.SetParams(params))

	newUserShare, newValidatorShare, newEmission, err := k.SimulateIncentiveChange(ctx, 600, 4500, 4500)
	require.NoError(t, err)

	// Verify user share (45% of 10M = 4.5M)
	userShare, ok := new(big.Int).SetString(newUserShare, 10)
	require.True(t, ok)
	require.Equal(t, "4500000", userShare.String())

	// Verify validator share (45% of 10M = 4.5M)
	validatorShare, ok := new(big.Int).SetString(newValidatorShare, 10)
	require.True(t, ok)
	require.Equal(t, "4500000", validatorShare.String())

	// Verify emission (6% of 1B = 60M)
	emission, ok := new(big.Int).SetString(newEmission, 10)
	require.True(t, ok)
	require.Equal(t, "60000000", emission.String())
}

func TestSimulateIncentiveChange_MaxPercentages(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Tokenomics.CirculatingSupply = "1000000000"
	params.Mev.TotalMevCaptured = "10000000"
	require.NoError(t, k.SetParams(params))

	// 100% to users, 100% to validators (unrealistic but tests math)
	newUserShare, newValidatorShare, newEmission, err := k.SimulateIncentiveChange(ctx, 10000, 10000, 10000)
	require.NoError(t, err)

	// 100% of 10M = 10M
	require.Equal(t, "10000000", newUserShare)
	require.Equal(t, "10000000", newValidatorShare)
	// 100% of 1B = 1B
	require.Equal(t, "1000000000", newEmission)
}

// =============================================================================
// GetIncentiveRecommendations Tests
// =============================================================================

func TestGetIncentiveRecommendations_NoRecommendations(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Default params should be balanced
	recommendations, err := k.GetIncentiveRecommendations(ctx)
	require.NoError(t, err)
	require.NotNil(t, recommendations)
}

func TestGetIncentiveRecommendations_LowUserMEVShare(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Mev.UserRedistributionPercentage = 2000 // 20% - below 25%
	params.Mev.ValidatorPercentage = 6000
	params.Mev.TreasuryPercentage = 2000
	params.Mev.BurnPercentage = 0
	require.NoError(t, k.SetParams(params))

	recommendations, err := k.GetIncentiveRecommendations(ctx)
	require.NoError(t, err)

	hasUserMEVRecommendation := false
	for _, rec := range recommendations {
		if rec == "Increase user MEV share to improve participation incentives" {
			hasUserMEVRecommendation = true
			break
		}
	}
	require.True(t, hasUserMEVRecommendation)
}

func TestGetIncentiveRecommendations_HighGasPrices(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.DynamicFees.CurrentMultiplier = 17000 // 85% of max (20000)
	params.DynamicFees.MaxMultiplier = 20000
	require.NoError(t, k.SetParams(params))

	recommendations, err := k.GetIncentiveRecommendations(ctx)
	require.NoError(t, err)

	hasGasRecommendation := false
	for _, rec := range recommendations {
		if rec == "Gas prices are high - implement transaction batching to reduce costs" {
			hasGasRecommendation = true
			break
		}
	}
	require.True(t, hasGasRecommendation)
}

// =============================================================================
// calculateIncentiveEfficiencyBps Tests
// =============================================================================

func TestCalculateIncentiveEfficiencyBps_PerfectScore(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.InflationRate = 500
	params.Tokenomics.TargetInflationRate = 500
	params.Mev.Strategy = types.MEVStrategyProportionalToStake
	params.Mev.TreasuryPercentage = 2000

	efficiencyBps := k.calculateIncentiveEfficiencyBps(*params, 1000, 100)

	// Should be high with aligned inflation, good strategy, balanced treasury (> 5000 bps = 50%)
	require.Greater(t, efficiencyBps, uint64(5000))
}

func TestCalculateIncentiveEfficiencyBps_LowUserCount(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.InflationRate = 500
	params.Tokenomics.TargetInflationRate = 500
	params.Mev.Strategy = types.MEVStrategyProportionalToStake
	params.Mev.TreasuryPercentage = 2000

	// Very low user count
	efficiencyBps := k.calculateIncentiveEfficiencyBps(*params, 50, 100)

	// Should be lower due to low user count (< 10000 bps = 100%)
	require.Less(t, efficiencyBps, uint64(10000))
}

func TestCalculateIncentiveEfficiencyBps_LowValidatorCount(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.InflationRate = 500
	params.Tokenomics.TargetInflationRate = 500
	params.Mev.Strategy = types.MEVStrategyProportionalToStake
	params.Mev.TreasuryPercentage = 2000

	// Very low validator count
	efficiencyBps := k.calculateIncentiveEfficiencyBps(*params, 1000, 5)

	// Should be much lower due to very few validators (< 9000 bps = 90%)
	require.Less(t, efficiencyBps, uint64(9000))
}

func TestCalculateIncentiveEfficiencyBps_MisalignedInflation(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.InflationRate = 1500      // 15%
	params.Tokenomics.TargetInflationRate = 500 // 5%
	params.Mev.Strategy = types.MEVStrategyProportionalToStake
	params.Mev.TreasuryPercentage = 2000

	efficiencyBps := k.calculateIncentiveEfficiencyBps(*params, 1000, 100)

	// Should be lower due to inflation misalignment (< 10000 bps = 100%)
	require.Less(t, efficiencyBps, uint64(10000))
}

func TestCalculateIncentiveEfficiencyBps_EqualDistributionStrategy(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.InflationRate = 500
	params.Tokenomics.TargetInflationRate = 500
	params.Mev.Strategy = types.MEVStrategyEqualDistribution // Less efficient
	params.Mev.TreasuryPercentage = 2000

	efficiencyBps := k.calculateIncentiveEfficiencyBps(*params, 1000, 100)

	// Should be somewhat lower due to equal distribution (<= 10000 bps = 100%)
	require.LessOrEqual(t, efficiencyBps, uint64(10000))
}

func TestCalculateIncentiveEfficiencyBps_LowTreasuryPercentage(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.InflationRate = 500
	params.Tokenomics.TargetInflationRate = 500
	params.Mev.Strategy = types.MEVStrategyProportionalToStake
	params.Mev.TreasuryPercentage = 500 // Very low

	efficiencyBps := k.calculateIncentiveEfficiencyBps(*params, 1000, 100)

	// Should be lower due to low treasury (< 10000 bps = 100%)
	require.Less(t, efficiencyBps, uint64(10000))
}

func TestCalculateIncentiveEfficiencyBps_HighTreasuryPercentage(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.InflationRate = 500
	params.Tokenomics.TargetInflationRate = 500
	params.Mev.Strategy = types.MEVStrategyProportionalToStake
	params.Mev.TreasuryPercentage = 4000 // Too high

	efficiencyBps := k.calculateIncentiveEfficiencyBps(*params, 1000, 100)

	// Should be lower due to high treasury (< 10000 bps = 100%)
	require.Less(t, efficiencyBps, uint64(10000))
}

func TestCalculateIncentiveEfficiencyBps_FloorAtZero(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.InflationRate = 5000 // Way off target
	params.Tokenomics.TargetInflationRate = 500
	params.Mev.Strategy = types.MEVStrategyEqualDistribution
	params.Mev.TreasuryPercentage = 500 // Low

	// Worst case scenario: very few users and validators
	efficiencyBps := k.calculateIncentiveEfficiencyBps(*params, 10, 2)

	// uint64 is always >= 0, verify it stays within bounds (<= 10000 bps)
	require.LessOrEqual(t, efficiencyBps, uint64(10000))
}

// =============================================================================
// Multiple Validator Slashing Simulation Tests
// =============================================================================

func TestStakingIncentives_MultipleValidators(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Tokenomics.CirculatingSupply = "1000000000"
	params.Tokenomics.InflationRate = 500
	require.NoError(t, k.SetParams(params))

	// Simulate different stake amounts for multiple validators
	validatorStakes := []string{
		"100000000", // 10% stake
		"200000000", // 20% stake
		"50000000",  // 5% stake
		"150000000", // 15% stake
	}

	for _, stake := range validatorStakes {
		result := k.analyzeStakingIncentives(params, stake)

		rewards, ok := new(big.Int).SetString(result.rewards, 10)
		require.True(t, ok)
		require.True(t, rewards.Sign() >= 0)

		// Verify proportionality
		stakeAmt, _ := new(big.Int).SetString(stake, 10)
		expectedRewards := new(big.Int).Mul(stakeAmt, big.NewInt(500))
		expectedRewards.Div(expectedRewards, big.NewInt(10000))

		require.Equal(t, expectedRewards.String(), rewards.String())
	}
}

// =============================================================================
// Staking History and Statistics Tests
// =============================================================================

func TestAnalyzeStakingIncentives_HistoricalComparison(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.CirculatingSupply = "1000000000"

	// Compare rewards at different inflation rates
	inflationRates := []uint64{100, 300, 500, 1000, 2000}
	var previousRewards *big.Int

	for _, rate := range inflationRates {
		params.Tokenomics.InflationRate = rate
		result := k.analyzeStakingIncentives(*params, "500000000")

		currentRewards, ok := new(big.Int).SetString(result.rewards, 10)
		require.True(t, ok)

		if previousRewards != nil {
			// Each higher inflation rate should yield more rewards
			require.True(t, currentRewards.Cmp(previousRewards) > 0,
				"higher inflation (%d) should yield more rewards", rate)
		}

		previousRewards = currentRewards
	}
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestAnalyzeStakingIncentives_EmptySupply(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.CirculatingSupply = "0"
	params.Tokenomics.InflationRate = 500

	// This should handle division by zero gracefully
	result := k.analyzeStakingIncentives(*params, "0")

	// Should not panic and return valid result
	require.NotNil(t, result)
}

func TestAnalyzeStakingIncentives_SingleUnit(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.CirculatingSupply = "1000000"
	params.Tokenomics.InflationRate = 500

	// Single unit stake
	result := k.analyzeStakingIncentives(*params, "1")

	rewards, ok := new(big.Int).SetString(result.rewards, 10)
	require.True(t, ok)
	require.True(t, rewards.Sign() >= 0)
}

func TestAnalyzeStakingIncentives_StakeExceedsSupply(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.CirculatingSupply = "1000000"
	params.Tokenomics.InflationRate = 500

	// Stake exceeds supply (edge case)
	result := k.analyzeStakingIncentives(*params, "2000000")

	rewards, ok := new(big.Int).SetString(result.rewards, 10)
	require.True(t, ok)
	// Rewards should be proportional to stake regardless of supply
	require.True(t, rewards.Sign() >= 0)
}

// =============================================================================
// Fuzz Tests
// =============================================================================

func FuzzCalculateOptimalIncentiveDistribution(f *testing.F) {
	f.Add("1000000", uint64(5000), uint64(500), uint64(50))
	f.Add("0", uint64(0), uint64(0), uint64(0))
	f.Add("999999999999999999", uint64(10000), uint64(10000), uint64(200))

	f.Fuzz(func(t *testing.T, rewards string, stakingRatioBps, users, validators uint64) {
		k, ctx := setupKeeperForTest(t)

		// stakingRatioBps is now directly in basis points, no conversion needed
		distribution, err := k.CalculateOptimalIncentiveDistribution(ctx, rewards, stakingRatioBps, users, validators)

		// Should not panic regardless of input
		if err == nil {
			require.NotNil(t, distribution)

			// All percentage keys should have valid values
			for key, value := range distribution {
				if key == "reasoning" {
					continue
				}
				_, ok := new(big.Int).SetString(value, 10)
				require.True(t, ok, "key %s has invalid value: %s", key, value)
			}
		}
	})
}

func FuzzSimulateIncentiveChange(f *testing.F) {
	f.Add(uint64(500), uint64(4000), uint64(5000))
	f.Add(uint64(0), uint64(0), uint64(0))
	f.Add(uint64(10000), uint64(10000), uint64(10000))

	f.Fuzz(func(t *testing.T, inflationRate, userPct, validatorPct uint64) {
		k, ctx := setupKeeperForTest(t)

		params, _ := k.GetParams(ctx)
		params.Tokenomics.CirculatingSupply = "1000000000"
		params.Mev.TotalMevCaptured = "10000000"
		require.NoError(t, k.SetParams(params))

		newUserShare, newValidatorShare, newEmission, err := k.SimulateIncentiveChange(ctx, inflationRate, userPct, validatorPct)

		// Should not panic
		require.NoError(t, err)

		// All results should be valid big.Int strings
		_, ok := new(big.Int).SetString(newUserShare, 10)
		require.True(t, ok)
		_, ok = new(big.Int).SetString(newValidatorShare, 10)
		require.True(t, ok)
		_, ok = new(big.Int).SetString(newEmission, 10)
		require.True(t, ok)
	})
}
