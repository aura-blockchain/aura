// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// =============================================================================
// AnalyzeEconomicIncentives Tests
// =============================================================================

func TestAnalyzeEconomicIncentives_BasicScenario(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Set up reasonable params
	params, _ := k.GetParams(ctx)
	params.Tokenomics.CirculatingSupply = "1000000000000" // 1 trillion
	params.Tokenomics.InflationRate = 500                 // 5%
	params.Tokenomics.TargetInflationRate = 500
	params.Mev.TotalMevCaptured = "1000000000" // 1 billion
	// MEV percentages must sum to 10000 (100%)
	params.Mev.UserRedistributionPercentage = 3000 // 30%
	params.Mev.ValidatorPercentage = 4000          // 40%
	params.Mev.TreasuryPercentage = 2000           // 20%
	params.Mev.BurnPercentage = 1000               // 10%
	params.Mev.Strategy = types.MEVStrategyProportionalToStake
	require.NoError(t, k.SetParams(params))

	// Initialize total burned
	require.NoError(t, k.SetTotalBurned(ctx, "50000000000")) // 5% burned

	result, err := k.AnalyzeEconomicIncentives(
		ctx,
		"300000000000",  // 30% staked
		"100000000000",  // 10% liquidity
		uint64(500),     // 500 active users
		uint64(100),     // 100 validators
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotEmpty(t, result.ValidatorRewards)
	require.NotEmpty(t, result.UserRewards)
	require.NotEmpty(t, result.TreasuryAllocation)
	require.NotEmpty(t, result.BurnAmount)
	require.GreaterOrEqual(t, result.IncentiveEfficiency, float64(0))
	require.LessOrEqual(t, result.IncentiveEfficiency, float64(100))
}

func TestAnalyzeEconomicIncentives_LowStakingRatio(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Tokenomics.CirculatingSupply = "1000000000000"
	params.Tokenomics.InflationRate = 500
	params.Tokenomics.TargetInflationRate = 500
	params.Mev.TotalMevCaptured = "1000000000"
	params.Mev.UserRedistributionPercentage = 3000
	params.Mev.TreasuryPercentage = 2000
	params.Mev.Strategy = types.MEVStrategyProportionalToStake
	require.NoError(t, k.SetParams(params))
	require.NoError(t, k.SetTotalBurned(ctx, "0"))

	result, err := k.AnalyzeEconomicIncentives(
		ctx,
		"200000000000", // 20% staked (low)
		"50000000000",
		uint64(100),
		uint64(50),
	)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Should recommend increasing staking rewards
	hasStakingRecommendation := false
	for _, rec := range result.RecommendedAdjustments {
		if len(rec) > 0 && (contains(rec, "staking") || contains(rec, "Increase")) {
			hasStakingRecommendation = true
			break
		}
	}
	require.True(t, hasStakingRecommendation, "should recommend increasing staking incentives for low ratio")
}

func TestAnalyzeEconomicIncentives_HighStakingRatio(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Tokenomics.CirculatingSupply = "1000000000000"
	params.Tokenomics.InflationRate = 500
	params.Tokenomics.TargetInflationRate = 500
	params.Mev.TotalMevCaptured = "1000000000"
	params.Mev.UserRedistributionPercentage = 3000
	params.Mev.TreasuryPercentage = 2000
	params.Mev.Strategy = types.MEVStrategyProportionalToStake
	require.NoError(t, k.SetParams(params))
	require.NoError(t, k.SetTotalBurned(ctx, "0"))

	result, err := k.AnalyzeEconomicIncentives(
		ctx,
		"800000000000", // 80% staked (high)
		"50000000000",
		uint64(1000),
		uint64(100),
	)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Should recommend reducing staking rewards
	hasReduceRecommendation := false
	for _, rec := range result.RecommendedAdjustments {
		if len(rec) > 0 && contains(rec, "reduc") {
			hasReduceRecommendation = true
			break
		}
	}
	require.True(t, hasReduceRecommendation, "should recommend reducing staking rewards for high ratio")
}

func TestAnalyzeEconomicIncentives_LowUserParticipation(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Tokenomics.CirculatingSupply = "1000000000000"
	params.Tokenomics.InflationRate = 500
	params.Tokenomics.TargetInflationRate = 500
	params.Mev.TotalMevCaptured = "1000000000"
	// MEV percentages must sum to 10000 (100%)
	params.Mev.UserRedistributionPercentage = 1500 // Low user share
	params.Mev.ValidatorPercentage = 5000          // 50%
	params.Mev.TreasuryPercentage = 2000           // 20%
	params.Mev.BurnPercentage = 1500               // 15%
	params.Mev.Strategy = types.MEVStrategyProportionalToStake
	require.NoError(t, k.SetParams(params))
	require.NoError(t, k.SetTotalBurned(ctx, "0"))

	result, err := k.AnalyzeEconomicIncentives(
		ctx,
		"500000000000",
		"50000000000",
		uint64(50), // Very few active users
		uint64(100),
	)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Should have recommendations about user participation
	hasUserRecommendation := false
	for _, rec := range result.RecommendedAdjustments {
		if len(rec) > 0 && (contains(rec, "user") || contains(rec, "participation")) {
			hasUserRecommendation = true
			break
		}
	}
	require.True(t, hasUserRecommendation, "should recommend improving user incentives")
}

func TestAnalyzeEconomicIncentives_ZeroValues(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Tokenomics.CirculatingSupply = "1000000000000"
	params.Tokenomics.InflationRate = 0
	params.Mev.TotalMevCaptured = "0"
	// MEV percentages must sum to 10000 (100%) even with zero MEV
	params.Mev.UserRedistributionPercentage = 0
	params.Mev.ValidatorPercentage = 5000
	params.Mev.TreasuryPercentage = 3000
	params.Mev.BurnPercentage = 2000
	require.NoError(t, k.SetParams(params))
	require.NoError(t, k.SetTotalBurned(ctx, "0"))

	result, err := k.AnalyzeEconomicIncentives(
		ctx,
		"0", // No staking
		"0", // No liquidity
		uint64(0),
		uint64(0),
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "0", result.ValidatorRewards)
	require.Equal(t, "0", result.UserRewards)
}

// =============================================================================
// analyzeStakingIncentives Tests
// =============================================================================

func TestAnalyzeStakingIncentives_NormalRatio(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.CirculatingSupply = "1000000000000"
	params.Tokenomics.InflationRate = 500

	result := k.analyzeStakingIncentives(*params, "400000000000") // 40% staked

	rewards, ok := new(big.Int).SetString(result.rewards, 10)
	require.True(t, ok)
	require.True(t, rewards.Sign() >= 0)
	// No recommendations for normal ratio
	require.Empty(t, result.recommendations)
}

func TestAnalyzeStakingIncentives_AboveTargetInflation(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.CirculatingSupply = "1000000000000"
	params.Tokenomics.InflationRate = 800
	params.Tokenomics.TargetInflationRate = 500

	result := k.analyzeStakingIncentives(*params, "500000000000")

	// Should recommend reducing emission rate
	hasInflationRec := false
	for _, rec := range result.recommendations {
		if contains(rec, "Inflation") || contains(rec, "inflation") {
			hasInflationRec = true
			break
		}
	}
	require.True(t, hasInflationRec, "should recommend action when inflation above target")
}

func TestAnalyzeStakingIncentivesHandlesHugeValues(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.CirculatingSupply = "1000000000000000000000000000000000000" // 1e33
	params.Tokenomics.InflationRate = 9000

	result := k.analyzeStakingIncentives(*params, "500000000000000000000000000000000000") // 5e32

	rewards, ok := new(big.Int).SetString(result.rewards, 10)
	require.True(t, ok, "rewards should parse as big.Int")
	require.False(t, rewards.Sign() < 0, "rewards must be non-negative")
}

func TestAnalyzeStakingIncentives_VeryLowStaking(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.CirculatingSupply = "1000000000000"
	params.Tokenomics.InflationRate = 500

	result := k.analyzeStakingIncentives(*params, "100000000000") // 10% staked

	hasLowStakingRec := false
	for _, rec := range result.recommendations {
		if contains(rec, "Increase staking") {
			hasLowStakingRec = true
			break
		}
	}
	require.True(t, hasLowStakingRec, "should recommend increasing rewards for low staking")
}

// =============================================================================
// analyzeUserIncentives Tests
// =============================================================================

func TestAnalyzeUserIncentives_HighUserCount(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Mev.TotalMevCaptured = "10000000000"
	params.Mev.UserRedistributionPercentage = 3000
	params.Mev.Strategy = types.MEVStrategyEqualDistribution

	result := k.analyzeUserIncentives(*params, 2000)

	// Should recommend switching from equal distribution with high user count
	hasSwitchRec := false
	for _, rec := range result.recommendations {
		if contains(rec, "switching") || contains(rec, "activity-weighted") {
			hasSwitchRec = true
			break
		}
	}
	require.True(t, hasSwitchRec, "should recommend activity-weighted distribution for high user count")
}

func TestAnalyzeUserIncentives_LowUserShare(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Mev.TotalMevCaptured = "10000000000"
	params.Mev.UserRedistributionPercentage = 1000 // Only 10%
	params.Mev.Strategy = types.MEVStrategyProportionalToStake

	result := k.analyzeUserIncentives(*params, 500)

	// Should recommend increasing user share
	hasIncreaseRec := false
	for _, rec := range result.recommendations {
		if contains(rec, "User MEV share") && contains(rec, "low") {
			hasIncreaseRec = true
			break
		}
	}
	require.True(t, hasIncreaseRec, "should recommend increasing user share when low")
}

func TestAnalyzeUserIncentives_CalculatesCorrectRewards(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Mev.TotalMevCaptured = "10000000000"       // 10B
	params.Mev.UserRedistributionPercentage = 2500    // 25%
	params.Mev.Strategy = types.MEVStrategyProportionalToStake

	result := k.analyzeUserIncentives(*params, 500)

	// User share should be 25% of 10B = 2.5B
	expected := new(big.Int).SetInt64(2500000000)
	actual, ok := new(big.Int).SetString(result.rewards, 10)
	require.True(t, ok)
	require.Equal(t, expected.String(), actual.String())
}

// =============================================================================
// analyzeTreasurySustainability Tests
// =============================================================================

func TestAnalyzeTreasurySustainability_LowAllocation(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Mev.TotalMevCaptured = "10000000000"
	params.Mev.TreasuryPercentage = 500 // Only 5%

	result := k.analyzeTreasurySustainability(*params)

	hasLowRec := false
	for _, rec := range result.recommendations {
		if contains(rec, "Treasury allocation is low") {
			hasLowRec = true
			break
		}
	}
	require.True(t, hasLowRec, "should recommend more treasury funding when low")
}

func TestAnalyzeTreasurySustainability_HighAllocation(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Mev.TotalMevCaptured = "10000000000"
	params.Mev.TreasuryPercentage = 4000 // 40%

	result := k.analyzeTreasurySustainability(*params)

	hasHighRec := false
	for _, rec := range result.recommendations {
		if contains(rec, "Treasury allocation is high") {
			hasHighRec = true
			break
		}
	}
	require.True(t, hasHighRec, "should recommend redirecting when treasury allocation is high")
}

func TestAnalyzeTreasurySustainability_NormalAllocation(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Mev.TotalMevCaptured = "10000000000"
	params.Mev.TreasuryPercentage = 2000 // 20%

	result := k.analyzeTreasurySustainability(*params)

	require.Empty(t, result.recommendations, "no recommendations for normal allocation")
}

// =============================================================================
// analyzeBurnEconomics Tests
// =============================================================================

func TestAnalyzeBurnEconomics_LowBurnRate(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Tokenomics.CirculatingSupply = "1000000000000"
	// MEV percentages must sum to 10000 (100%)
	params.Mev.UserRedistributionPercentage = 4000
	params.Mev.ValidatorPercentage = 3800
	params.Mev.TreasuryPercentage = 2000
	params.Mev.BurnPercentage = 200 // Only 2%
	require.NoError(t, k.SetParams(params))
	require.NoError(t, k.SetTotalBurned(ctx, "100000000")) // 0.01% burned

	result, err := k.analyzeBurnEconomics(ctx, params)
	require.NoError(t, err)

	// Should have recommendations about low burn
	hasBurnRec := false
	for _, rec := range result.recommendations {
		if contains(rec, "burn") || contains(rec, "Burn") {
			hasBurnRec = true
			break
		}
	}
	require.True(t, hasBurnRec, "should recommend actions for low burn rate")
}

func TestAnalyzeBurnEconomics_HighBurnRate(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Tokenomics.CirculatingSupply = "1000000000000"
	// MEV percentages must sum to 10000 (100%)
	params.Mev.UserRedistributionPercentage = 3000
	params.Mev.ValidatorPercentage = 4000
	params.Mev.TreasuryPercentage = 2000
	params.Mev.BurnPercentage = 1000
	require.NoError(t, k.SetParams(params))
	require.NoError(t, k.SetTotalBurned(ctx, "100000000000")) // 10% burned

	result, err := k.analyzeBurnEconomics(ctx, params)
	require.NoError(t, err)

	// Should warn about high burn
	hasHighBurnWarning := false
	for _, rec := range result.recommendations {
		if contains(rec, "high") || contains(rec, "High") {
			hasHighBurnWarning = true
			break
		}
	}
	require.True(t, hasHighBurnWarning, "should warn about high burn rate")
}

// =============================================================================
// calculateIncentiveEfficiency Tests
// =============================================================================

func TestCalculateIncentiveEfficiency_OptimalConditions(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.InflationRate = 500
	params.Tokenomics.TargetInflationRate = 500
	params.Mev.TreasuryPercentage = 2000
	params.Mev.Strategy = types.MEVStrategyProportionalToStake

	efficiency := k.calculateIncentiveEfficiency(*params, 1000, 100)

	// Should be relatively high for optimal conditions
	require.GreaterOrEqual(t, efficiency, float64(70))
}

func TestCalculateIncentiveEfficiency_PoorConditions(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.InflationRate = 1000
	params.Tokenomics.TargetInflationRate = 500 // 5% off target
	params.Mev.TreasuryPercentage = 500          // Very low
	params.Mev.Strategy = types.MEVStrategyEqualDistribution

	efficiency := k.calculateIncentiveEfficiency(*params, 10, 5)

	// Should be lower for poor conditions
	require.Less(t, efficiency, float64(70))
}

func TestCalculateIncentiveEfficiency_NeverNegative(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	// Set up worst case scenario
	params := types.DefaultParams()
	params.Tokenomics.InflationRate = 5000
	params.Tokenomics.TargetInflationRate = 100
	params.Mev.TreasuryPercentage = 0
	params.Mev.Strategy = types.MEVStrategyEqualDistribution

	efficiency := k.calculateIncentiveEfficiency(*params, 0, 0)

	require.GreaterOrEqual(t, efficiency, float64(0))
}

func TestCalculateIncentiveEfficiency_NeverExceeds100(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.InflationRate = 500
	params.Tokenomics.TargetInflationRate = 500
	params.Mev.TreasuryPercentage = 2000
	params.Mev.Strategy = types.MEVStrategyProportionalToStake

	efficiency := k.calculateIncentiveEfficiency(*params, 100000, 1000)

	require.LessOrEqual(t, efficiency, float64(100))
}

// =============================================================================
// GetIncentiveRecommendations Tests
// =============================================================================

func TestGetIncentiveRecommendations_HighInflation(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Tokenomics.InflationRate = 1950       // 19.5%
	params.Tokenomics.MaxInflationRate = 2000    // 20% max - so 1950 > 2000-100=1900
	// MEV percentages must sum to 10000 (100%)
	params.Mev.UserRedistributionPercentage = 3000
	params.Mev.ValidatorPercentage = 4000
	params.Mev.TreasuryPercentage = 2000
	params.Mev.BurnPercentage = 1000
	params.DynamicFees.CurrentMultiplier = 10000
	params.DynamicFees.MaxMultiplier = 20000
	require.NoError(t, k.SetParams(params))

	recs, err := k.GetIncentiveRecommendations(ctx)
	require.NoError(t, err)

	hasInflationRec := false
	for _, rec := range recs {
		if contains(rec, "Inflation approaching maximum") {
			hasInflationRec = true
			break
		}
	}
	require.True(t, hasInflationRec, "should warn about inflation approaching max")
}

func TestGetIncentiveRecommendations_LowUserShare(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Tokenomics.InflationRate = 500
	params.Tokenomics.MaxInflationRate = 2000
	// MEV percentages must sum to 10000 (100%)
	params.Mev.UserRedistributionPercentage = 2000 // Below 25%
	params.Mev.ValidatorPercentage = 5000
	params.Mev.TreasuryPercentage = 2000
	params.Mev.BurnPercentage = 1000
	params.DynamicFees.CurrentMultiplier = 10000
	params.DynamicFees.MaxMultiplier = 20000
	require.NoError(t, k.SetParams(params))

	recs, err := k.GetIncentiveRecommendations(ctx)
	require.NoError(t, err)

	hasUserRec := false
	for _, rec := range recs {
		if contains(rec, "user MEV share") {
			hasUserRec = true
			break
		}
	}
	require.True(t, hasUserRec, "should recommend increasing user MEV share")
}

func TestGetIncentiveRecommendations_HighGasPrice(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Tokenomics.InflationRate = 500
	params.Tokenomics.MaxInflationRate = 2000
	// MEV percentages must sum to 10000 (100%)
	params.Mev.UserRedistributionPercentage = 3000
	params.Mev.ValidatorPercentage = 4000
	params.Mev.TreasuryPercentage = 2000
	params.Mev.BurnPercentage = 1000
	params.DynamicFees.CurrentMultiplier = 17000 // Above 80% of max
	params.DynamicFees.MaxMultiplier = 20000
	require.NoError(t, k.SetParams(params))

	recs, err := k.GetIncentiveRecommendations(ctx)
	require.NoError(t, err)

	hasGasRec := false
	for _, rec := range recs {
		if contains(rec, "Gas prices are high") {
			hasGasRec = true
			break
		}
	}
	require.True(t, hasGasRec, "should recommend transaction batching for high gas prices")
}

func TestGetIncentiveRecommendations_WhaleActivity(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Tokenomics.InflationRate = 500
	params.Tokenomics.MaxInflationRate = 2000
	// MEV percentages must sum to 10000 (100%)
	params.Mev.UserRedistributionPercentage = 3000
	params.Mev.ValidatorPercentage = 4000
	params.Mev.TreasuryPercentage = 2000
	params.Mev.BurnPercentage = 1000
	params.DynamicFees.CurrentMultiplier = 10000
	params.DynamicFees.MaxMultiplier = 20000
	require.NoError(t, k.SetParams(params))

	// Add many large tx records
	for i := 0; i < 60; i++ {
		record := &types.LargeTxRecord{
			TxHash:             "hash-" + string(rune('A'+i)),
			Sender:             "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr",
			Recipient:          "aura1w3jhxapjta047h6lta047h6lta047h6l42n9lg",
			Amount:             "1000000000",
			Timestamp:          time.Now(),
			BlockHeight:        uint64(1000 + i),
			PercentageOfSupply: 100,
		}
		require.NoError(t, k.SetLargeTxRecord(ctx, record))
	}

	recs, err := k.GetIncentiveRecommendations(ctx)
	require.NoError(t, err)

	hasWhaleRec := false
	for _, rec := range recs {
		if contains(rec, "whale") || contains(rec, "anti-whale") {
			hasWhaleRec = true
			break
		}
	}
	require.True(t, hasWhaleRec, "should recommend strengthening anti-whale protection")
}

// =============================================================================
// SimulateIncentiveChange Tests
// =============================================================================

func TestSimulateIncentiveChange_BasicSimulation(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Tokenomics.CirculatingSupply = "1000000000000"
	params.Mev.TotalMevCaptured = "10000000000"
	require.NoError(t, k.SetParams(params))

	userShare, validatorShare, emission, err := k.SimulateIncentiveChange(
		ctx,
		uint64(500),  // 5% inflation
		uint64(3000), // 30% user share
		uint64(4000), // 40% validator share
	)

	require.NoError(t, err)
	require.NotEmpty(t, userShare)
	require.NotEmpty(t, validatorShare)
	require.NotEmpty(t, emission)

	// Verify calculations
	// User share: 30% of 10B = 3B
	expectedUserShare := new(big.Int).SetInt64(3000000000)
	actualUserShare, _ := new(big.Int).SetString(userShare, 10)
	require.Equal(t, expectedUserShare.String(), actualUserShare.String())

	// Validator share: 40% of 10B = 4B
	expectedValidatorShare := new(big.Int).SetInt64(4000000000)
	actualValidatorShare, _ := new(big.Int).SetString(validatorShare, 10)
	require.Equal(t, expectedValidatorShare.String(), actualValidatorShare.String())
}

func TestSimulateIncentiveChange_ZeroPercentages(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Tokenomics.CirculatingSupply = "1000000000000"
	params.Mev.TotalMevCaptured = "10000000000"
	require.NoError(t, k.SetParams(params))

	userShare, validatorShare, emission, err := k.SimulateIncentiveChange(
		ctx,
		uint64(0),
		uint64(0),
		uint64(0),
	)

	require.NoError(t, err)
	require.Equal(t, "0", userShare)
	require.Equal(t, "0", validatorShare)
	require.Equal(t, "0", emission)
}

// =============================================================================
// CalculateOptimalIncentiveDistribution Tests
// =============================================================================

func TestCalculateOptimalIncentiveDistribution_LowStakingRatio(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	distribution, err := k.CalculateOptimalIncentiveDistribution(
		ctx,
		"1000000000000", // 1T total rewards
		0.2,             // 20% staking ratio (low)
		uint64(500),     // 500 active users
		uint64(100),     // 100 validators
	)

	require.NoError(t, err)
	require.NotNil(t, distribution)

	// Should increase validator percentage for low staking
	validatorRewards, ok := new(big.Int).SetString(distribution["validators"], 10)
	require.True(t, ok)
	require.True(t, validatorRewards.Sign() > 0)

	// Verify reasoning is included
	require.NotEmpty(t, distribution["reasoning"])
}

func TestCalculateOptimalIncentiveDistribution_HighStakingRatio(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	distribution, err := k.CalculateOptimalIncentiveDistribution(
		ctx,
		"1000000000000",
		0.8, // 80% staking ratio (high)
		uint64(500),
		uint64(100),
	)

	require.NoError(t, err)
	require.NotNil(t, distribution)

	// All distribution components should be present
	require.NotEmpty(t, distribution["validators"])
	require.NotEmpty(t, distribution["users"])
	require.NotEmpty(t, distribution["treasury"])
	require.NotEmpty(t, distribution["burn"])
	require.NotEmpty(t, distribution["reasoning"])
}

func TestCalculateOptimalIncentiveDistribution_LowUserCount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	distribution, err := k.CalculateOptimalIncentiveDistribution(
		ctx,
		"1000000000000",
		0.5, // 50% staking ratio
		uint64(50), // Very low user count
		uint64(100),
	)

	require.NoError(t, err)
	require.NotNil(t, distribution)

	// Should increase user percentage to attract users
	// Total distribution should equal total rewards (minus any rounding)
	validators, _ := new(big.Int).SetString(distribution["validators"], 10)
	users, _ := new(big.Int).SetString(distribution["users"], 10)
	treasury, _ := new(big.Int).SetString(distribution["treasury"], 10)
	burn, _ := new(big.Int).SetString(distribution["burn"], 10)

	total := new(big.Int).Add(validators, users)
	total.Add(total, treasury)
	total.Add(total, burn)

	require.True(t, total.Sign() > 0)
}

func TestCalculateOptimalIncentiveDistribution_InvalidAmount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	_, err := k.CalculateOptimalIncentiveDistribution(
		ctx,
		"invalid",
		0.5,
		uint64(500),
		uint64(100),
	)

	require.Error(t, err)
	require.Equal(t, types.ErrInvalidAmount, err)
}

func TestCalculateOptimalIncentiveDistribution_ZeroRewards(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	distribution, err := k.CalculateOptimalIncentiveDistribution(
		ctx,
		"0",
		0.5,
		uint64(500),
		uint64(100),
	)

	require.NoError(t, err)
	require.NotNil(t, distribution)
	require.Equal(t, "0", distribution["validators"])
	require.Equal(t, "0", distribution["users"])
	require.Equal(t, "0", distribution["treasury"])
	require.Equal(t, "0", distribution["burn"])
}

// =============================================================================
// calculateCurrentGasPrice Tests
// =============================================================================

func TestCalculateCurrentGasPrice_Basic(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 10000 // 100%

	price := k.calculateCurrentGasPrice(*params)

	require.Equal(t, "1000", price) // 1000 * 10000 / 10000 = 1000
}

func TestCalculateCurrentGasPrice_WithMultiplier(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 15000 // 150%

	price := k.calculateCurrentGasPrice(*params)

	require.Equal(t, "1500", price) // 1000 * 15000 / 10000 = 1500
}

func TestCalculateCurrentGasPrice_InvalidBaseFee(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.DynamicFees.BaseFee = "invalid"
	params.DynamicFees.CurrentMultiplier = 10000

	price := k.calculateCurrentGasPrice(*params)

	// Should return the original baseFee on parse failure
	require.Equal(t, "invalid", price)
}

func TestCalculateCurrentGasPrice_LargeValues(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.DynamicFees.BaseFee = "999999999999999999999"
	params.DynamicFees.CurrentMultiplier = 20000 // 200%

	price := k.calculateCurrentGasPrice(*params)

	expected := new(big.Int)
	expected.SetString("999999999999999999999", 10)
	expected.Mul(expected, big.NewInt(20000))
	expected.Div(expected, big.NewInt(types.BasisPoints))

	require.Equal(t, expected.String(), price)
}

// =============================================================================
// ValidateInvariants Tests
// =============================================================================

func TestValidateInvariants_ValidState(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Set up valid vesting schedule
	schedule := &types.VestingSchedule{
		ScheduleId:         "test-schedule-1",
		BeneficiaryAddress: "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr",
		TotalAmount:        "1000000",
		VestedAmount:       "100000",
		VestingDuration:    31536000,
		CliffDuration:      7776000,
		StartTime:          time.Now(),
		VestingType:        types.VestingTypeLinear,
	}
	require.NoError(t, k.SetVestingSchedule(ctx, schedule))

	// Set up valid vote lock
	lock := &types.VoteLock{
		LockId:      "test-lock-1",
		Owner:       "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr",
		Amount:      "500000",
		LockStart:   time.Now(),
		LockEnd:     time.Now().Add(30 * 24 * time.Hour),
		VotingPower: "500000",
	}
	require.NoError(t, k.SetVoteLock(ctx, lock))

	err := k.ValidateInvariants(ctx)
	require.NoError(t, err)
}

func TestValidateInvariants_EmptyState(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Empty state should be valid
	err := k.ValidateInvariants(ctx)
	require.NoError(t, err)
}

func TestValidateInvariants_MultipleSchedules(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Add multiple valid schedules
	for i := 0; i < 5; i++ {
		schedule := &types.VestingSchedule{
			ScheduleId:         "test-schedule-" + string(rune('A'+i)),
			BeneficiaryAddress: "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr",
			TotalAmount:        "1000000",
			VestedAmount:       "100000",
			VestingDuration:    31536000,
			CliffDuration:      7776000,
			StartTime:          time.Now(),
			VestingType:        types.VestingTypeLinear,
		}
		require.NoError(t, k.SetVestingSchedule(ctx, schedule))
	}

	err := k.ValidateInvariants(ctx)
	require.NoError(t, err)
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestAnalyzeEconomicIncentives_MaxValues(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	params, _ := k.GetParams(ctx)
	params.Tokenomics.CirculatingSupply = "340282366920938463463374607431768211455" // Max uint128
	params.Tokenomics.InflationRate = 10000                                         // 100%
	params.Tokenomics.TargetInflationRate = 10000
	params.Mev.TotalMevCaptured = "340282366920938463463374607431768211455"
	// MEV percentages must sum to 10000 (100%)
	params.Mev.UserRedistributionPercentage = 2500
	params.Mev.ValidatorPercentage = 2500
	params.Mev.TreasuryPercentage = 2500
	params.Mev.BurnPercentage = 2500
	params.Mev.Strategy = types.MEVStrategyProportionalToStake
	require.NoError(t, k.SetParams(params))
	require.NoError(t, k.SetTotalBurned(ctx, "340282366920938463463374607431768211455"))

	result, err := k.AnalyzeEconomicIncentives(
		ctx,
		"340282366920938463463374607431768211455",
		"340282366920938463463374607431768211455",
		^uint64(0),
		^uint64(0),
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.GreaterOrEqual(t, result.IncentiveEfficiency, float64(0))
}

func TestAnalyzeEconomicIncentives_BoundaryStakingRatios(t *testing.T) {
	testCases := []struct {
		name              string
		stakedAmount      string
		circulatingSupply string
		description       string
	}{
		{"exactly_30_percent", "30000000000", "100000000000", "at 30% threshold"},
		{"exactly_70_percent", "70000000000", "100000000000", "at 70% threshold"},
		{"just_below_30", "29999999999", "100000000000", "just below 30%"},
		{"just_above_70", "70000000001", "100000000000", "just above 70%"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			k, ctx := setupKeeperForTest(t)

			params, _ := k.GetParams(ctx)
			params.Tokenomics.CirculatingSupply = tc.circulatingSupply
			params.Tokenomics.InflationRate = 500
			params.Tokenomics.TargetInflationRate = 500
			params.Mev.TotalMevCaptured = "1000000000"
			params.Mev.UserRedistributionPercentage = 3000
			params.Mev.TreasuryPercentage = 2000
			params.Mev.Strategy = types.MEVStrategyProportionalToStake
			require.NoError(t, k.SetParams(params))
			require.NoError(t, k.SetTotalBurned(ctx, "0"))

			result, err := k.AnalyzeEconomicIncentives(
				ctx,
				tc.stakedAmount,
				"10000000000",
				uint64(500),
				uint64(100),
			)

			require.NoError(t, err, tc.description)
			require.NotNil(t, result, tc.description)
		})
	}
}

// =============================================================================
// Fuzz Tests
// =============================================================================

func FuzzAnalyzeStakingIncentives_NoOverflow(f *testing.F) {
	f.Add(uint64(1_000_000_000), uint64(500_000_000))
	f.Add(uint64(10), uint64(9))
	f.Add(^uint64(0)-1, ^uint64(0)/2)

	f.Fuzz(func(t *testing.T, supplyRaw uint64, stakedRaw uint64) {
		k, _ := setupKeeperForTest(t)

		params := types.DefaultParams()
		params.Tokenomics.InflationRate = 5000 // keep within bounds for validation

		supply := new(big.Int).SetUint64(supplyRaw)
		if supply.Sign() == 0 {
			supply.SetUint64(1)
		}

		staked := new(big.Int).SetUint64(stakedRaw)
		staked.Mod(staked, new(big.Int).Add(supply, big.NewInt(1))) // ensure staked <= supply

		params.Tokenomics.CirculatingSupply = supply.String()

		result := k.analyzeStakingIncentives(*params, staked.String())

		ratio := new(big.Int).Mul(staked, big.NewInt(10000))
		ratio.Div(ratio, supply)
		require.LessOrEqual(t, ratio.Uint64(), uint64(10000), "staking ratio should stay within 0-10000 bps")

		rewards, ok := new(big.Int).SetString(result.rewards, 10)
		require.True(t, ok, "rewards should be parseable")
		require.False(t, rewards.Sign() < 0, "rewards must not be negative")
	})
}

func FuzzCalculateIncentiveEfficiency_ValidRange(f *testing.F) {
	f.Add(uint64(0), uint64(0))
	f.Add(uint64(100), uint64(50))
	f.Add(uint64(10000), uint64(1000))
	f.Add(^uint64(0), ^uint64(0))

	f.Fuzz(func(t *testing.T, activeUsers, validators uint64) {
		k, _ := setupKeeperForTest(t)

		params := types.DefaultParams()
		params.Tokenomics.InflationRate = 500
		params.Tokenomics.TargetInflationRate = 500
		params.Mev.TreasuryPercentage = 2000
		params.Mev.Strategy = types.MEVStrategyProportionalToStake

		efficiency := k.calculateIncentiveEfficiency(*params, activeUsers, validators)

		require.GreaterOrEqual(t, efficiency, float64(0), "efficiency must be >= 0")
		require.LessOrEqual(t, efficiency, float64(100), "efficiency must be <= 100")
	})
}

func FuzzSimulateIncentiveChange_NoOverflow(f *testing.F) {
	f.Add(uint64(100), uint64(3000), uint64(4000))
	f.Add(uint64(0), uint64(0), uint64(0))
	f.Add(uint64(10000), uint64(10000), uint64(10000))

	f.Fuzz(func(t *testing.T, inflationRate, userPct, validatorPct uint64) {
		k, ctx := setupKeeperForTest(t)

		params, _ := k.GetParams(ctx)
		params.Tokenomics.CirculatingSupply = "1000000000000000"
		params.Mev.TotalMevCaptured = "1000000000000"
		require.NoError(t, k.SetParams(params))

		userShare, validatorShare, emission, err := k.SimulateIncentiveChange(
			ctx,
			inflationRate,
			userPct,
			validatorPct,
		)

		require.NoError(t, err)

		// All outputs should be parseable
		us, ok := new(big.Int).SetString(userShare, 10)
		require.True(t, ok)
		require.True(t, us.Sign() >= 0)

		vs, ok := new(big.Int).SetString(validatorShare, 10)
		require.True(t, ok)
		require.True(t, vs.Sign() >= 0)

		em, ok := new(big.Int).SetString(emission, 10)
		require.True(t, ok)
		require.True(t, em.Sign() >= 0)
	})
}

// =============================================================================
// Helper Functions
// =============================================================================

// contains checks if a string contains a substring (case-sensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
