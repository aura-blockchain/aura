// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// ============================
// ECONOMIC INCENTIVE ANALYSIS (Feature 8)
// ============================

// IncentiveAnalysisResult contains the results of incentive analysis
type IncentiveAnalysisResult struct {
	ValidatorRewards   string
	UserRewards        string
	TreasuryAllocation string
	BurnAmount         string
	// IncentiveEfficiencyBps stores the incentive efficiency score in basis points (0-10000 = 0-100%)
	// DETERMINISM: Using integer basis points instead of float64 ensures cross-platform
	// consistency. Float operations can produce different results on different CPU architectures,
	// causing consensus failures in blockchain state machines.
	IncentiveEfficiencyBps uint64
	RecommendedAdjustments []string
}

// AnalyzeEconomicIncentives performs comprehensive economic incentive analysis
func (k *Keeper) AnalyzeEconomicIncentives(
	ctx context.Context,
	totalStaked string,
	totalLiquidity string,
	activeUsers uint64,
	validators uint64,
) (*IncentiveAnalysisResult, error) {
	params, _ := k.GetParams(ctx)

	result := &IncentiveAnalysisResult{
		RecommendedAdjustments: []string{},
	}

	// 1. Analyze staking incentives
	stakingAnalysis := k.analyzeStakingIncentives(params, totalStaked)
	result.ValidatorRewards = stakingAnalysis.rewards
	result.RecommendedAdjustments = append(result.RecommendedAdjustments, stakingAnalysis.recommendations...)

	// 2. Analyze user participation incentives
	userAnalysis := k.analyzeUserIncentives(params, activeUsers)
	result.UserRewards = userAnalysis.rewards
	result.RecommendedAdjustments = append(result.RecommendedAdjustments, userAnalysis.recommendations...)

	// 3. Analyze treasury sustainability
	treasuryAnalysis := k.analyzeTreasurySustainability(params)
	result.TreasuryAllocation = treasuryAnalysis.allocation
	result.RecommendedAdjustments = append(result.RecommendedAdjustments, treasuryAnalysis.recommendations...)

	// 4. Analyze token burn economics
	burnAnalysis, err := k.analyzeBurnEconomics(ctx, params)
	if err != nil {
		return nil, err
	}
	result.BurnAmount = burnAnalysis.amount
	result.RecommendedAdjustments = append(result.RecommendedAdjustments, burnAnalysis.recommendations...)

	// 5. Calculate overall incentive efficiency
	result.IncentiveEfficiencyBps = k.calculateIncentiveEfficiencyBps(params, activeUsers, validators)

	return result, nil
}

// analyzeStakingIncentives analyzes staking reward incentives
func (k *Keeper) analyzeStakingIncentives(params types.Params, totalStaked string) struct {
	rewards         string
	recommendations []string
} {
	staked := new(big.Int)
	staked.SetString(totalStaked, 10)

	totalSupply := new(big.Int)
	totalSupply.SetString(params.Tokenomics.CirculatingSupply, 10)

	// Handle zero supply gracefully
	if totalSupply.Sign() == 0 {
		return struct {
			rewards         string
			recommendations []string
		}{
			rewards:         "0",
			recommendations: []string{"Circulating supply is zero - cannot calculate staking incentives"},
		}
	}

	// Calculate staking ratio
	stakingRatio := new(big.Int).Mul(staked, big.NewInt(10000))
	stakingRatio.Div(stakingRatio, totalSupply)

	// Calculate annual staking rewards (simplified)
	annualRewards := new(big.Int).Mul(staked, big.NewInt(int64(params.Tokenomics.InflationRate)))
	annualRewards.Div(annualRewards, big.NewInt(10000))

	recommendations := []string{}

	// Analyze staking ratio
	stakingRatioBps := stakingRatio.Uint64()
	if stakingRatioBps < 3000 { // Less than 30% staked
		recommendations = append(recommendations, "Increase staking rewards to incentivize more staking (current ratio: <30%)")
	} else if stakingRatioBps > 7000 { // More than 70% staked
		recommendations = append(recommendations, "Consider reducing staking rewards as ratio is high (>70%)")
	}

	// Check inflation rate alignment
	if params.Tokenomics.InflationRate > params.Tokenomics.TargetInflationRate {
		recommendations = append(recommendations, "Inflation above target - reduce staking emission rate")
	}

	return struct {
		rewards         string
		recommendations []string
	}{
		rewards:         annualRewards.String(),
		recommendations: recommendations,
	}
}

// analyzeUserIncentives analyzes user participation incentives
func (k *Keeper) analyzeUserIncentives(params types.Params, activeUsers uint64) struct {
	rewards         string
	recommendations []string
} {
	// Calculate MEV redistribution to users
	totalMEVCaptured := new(big.Int)
	totalMEVCaptured.SetString(params.Mev.TotalMevCaptured, 10)

	userShare := new(big.Int).Mul(totalMEVCaptured, big.NewInt(int64(params.Mev.UserRedistributionPercentage)))
	userShare.Div(userShare, big.NewInt(10000))

	recommendations := []string{}

	// Analyze user participation
	if activeUsers < 100 {
		recommendations = append(recommendations, "Low user participation - increase user reward percentage")
	}

	// Analyze MEV distribution strategy
	if params.Mev.Strategy == types.MEVStrategyEqualDistribution && activeUsers > 1000 {
		recommendations = append(recommendations, "High user count - consider switching to activity-weighted MEV distribution")
	}

	// Check user reward percentage
	if params.Mev.UserRedistributionPercentage < 2000 { // Less than 20%
		recommendations = append(recommendations, "User MEV share is low (<20%) - consider increasing to improve participation")
	}

	return struct {
		rewards         string
		recommendations []string
	}{
		rewards:         userShare.String(),
		recommendations: recommendations,
	}
}

// analyzeTreasurySustainability analyzes treasury sustainability
func (k *Keeper) analyzeTreasurySustainability(params types.Params) struct {
	allocation      string
	recommendations []string
} {
	totalMEVCaptured := new(big.Int)
	totalMEVCaptured.SetString(params.Mev.TotalMevCaptured, 10)

	treasuryShare := new(big.Int).Mul(totalMEVCaptured, big.NewInt(int64(params.Mev.TreasuryPercentage)))
	treasuryShare.Div(treasuryShare, big.NewInt(10000))

	recommendations := []string{}

	// Analyze treasury allocation
	if params.Mev.TreasuryPercentage < 1000 { // Less than 10%
		recommendations = append(recommendations, "Treasury allocation is low (<10%) - may need more funding for development")
	} else if params.Mev.TreasuryPercentage > 3000 { // More than 30%
		recommendations = append(recommendations, "Treasury allocation is high (>30%) - consider redirecting more to users/validators")
	}

	return struct {
		allocation      string
		recommendations []string
	}{
		allocation:      treasuryShare.String(),
		recommendations: recommendations,
	}
}

// analyzeBurnEconomics analyzes token burn economics
func (k *Keeper) analyzeBurnEconomics(ctx context.Context, params types.Params) (struct {
	amount          string
	recommendations []string
}, error) {
	burnAmount, err := k.GetTotalBurned(ctx)
	if err != nil {
		return struct {
			amount          string
			recommendations []string
		}{"0", []string{}}, err
	}

	recommendations := []string{}

	totalSupply := new(big.Int)
	totalSupply.SetString(params.Tokenomics.CirculatingSupply, 10)

	burned := new(big.Int)
	burned.SetString(burnAmount, 10)

	// Calculate burn ratio
	burnRatio := new(big.Int).Mul(burned, big.NewInt(10000))
	burnRatio.Div(burnRatio, totalSupply)
	burnRatioBps := burnRatio.Uint64()

	// Analyze burn rate
	if burnRatioBps < 10 { // Less than 0.1% burned
		recommendations = append(recommendations, "Burn rate is very low - consider increasing burn percentage")
	} else if burnRatioBps > 500 { // More than 5% burned
		recommendations = append(recommendations, "Burn rate is high (>5%) - ensure deflationary pressure is intentional")
	}

	// Check MEV burn percentage
	if params.Mev.BurnPercentage < 500 { // Less than 5%
		recommendations = append(recommendations, "MEV burn percentage is low - consider increasing to create deflationary pressure")
	}

	return struct {
		amount          string
		recommendations []string
	}{
		amount:          burnAmount,
		recommendations: recommendations,
	}, nil
}

// calculateIncentiveEfficiencyBps calculates overall incentive efficiency score in basis points (0-10000)
// DETERMINISM: This function uses math.LegacyDec for all calculations to ensure cross-platform
// consistency. Float64 operations can produce different results on different CPU architectures
// (x86 vs ARM, different FPU implementations), causing app hash mismatches and consensus failures
// during chain replay or state sync. The result is returned as basis points (integer) for storage.
func (k *Keeper) calculateIncentiveEfficiencyBps(params types.Params, activeUsers uint64, validators uint64) uint64 {
	// Start with 100% score (10000 basis points)
	score := math.LegacyNewDec(10000)

	// Deduct points for inefficiencies (all deductions in basis points)

	// 1. Check inflation alignment (max -2000 bps = 20 points)
	inflationDelta := int64(params.Tokenomics.InflationRate) - int64(params.Tokenomics.TargetInflationRate)
	if inflationDelta < 0 {
		inflationDelta = -inflationDelta
	}
	if inflationDelta > 100 { // More than 1% off target
		// Deduct inflationDelta/10 points, converted to basis points (*100)
		// inflationDelta is in basis points already, so deduction = inflationDelta * 10 (to convert to score bps)
		deduction := math.LegacyNewDec(inflationDelta).QuoInt64(10).MulInt64(100)
		score = score.Sub(deduction)
		minScore := math.LegacyNewDec(8000) // Minimum 80% (8000 bps)
		if score.LT(minScore) {
			score = minScore
		}
	}

	// 2. Check user participation (max -2000 bps = 20 points)
	if activeUsers < 100 {
		score = score.Sub(math.LegacyNewDec(2000))
	} else if activeUsers < 500 {
		score = score.Sub(math.LegacyNewDec(1000))
	}

	// 3. Check MEV distribution efficiency (max -2000 bps = 20 points)
	if params.Mev.Strategy == types.MEVStrategyEqualDistribution {
		score = score.Sub(math.LegacyNewDec(1000)) // Equal distribution is less efficient
	}

	// 4. Check treasury balance (max -2000 bps = 20 points)
	if params.Mev.TreasuryPercentage < 1000 {
		score = score.Sub(math.LegacyNewDec(1500))
	} else if params.Mev.TreasuryPercentage > 3000 {
		score = score.Sub(math.LegacyNewDec(1000))
	}

	// 5. Check validator count (max -2000 bps = 20 points)
	if validators < 10 {
		score = score.Sub(math.LegacyNewDec(2000)) // Very few validators
	} else if validators < 50 {
		score = score.Sub(math.LegacyNewDec(1000))
	}

	// Ensure score doesn't go negative
	if score.IsNegative() {
		score = math.LegacyZeroDec()
	}

	// Return as uint64 (truncate any remaining decimal)
	return uint64(score.TruncateInt64())
}

// GetIncentiveRecommendations returns specific incentive recommendations
func (k *Keeper) GetIncentiveRecommendations(ctx context.Context) ([]string, error) {
	params, _ := k.GetParams(ctx)
	recommendations := []string{}

	// Analyze current state and provide recommendations
	if params.Tokenomics.InflationRate > params.Tokenomics.MaxInflationRate-100 {
		recommendations = append(recommendations, "Inflation approaching maximum - implement deflationary measures")
	}

	if params.Mev.UserRedistributionPercentage < 2500 {
		recommendations = append(recommendations, "Increase user MEV share to improve participation incentives")
	}

	if params.DynamicFees.CurrentMultiplier > params.DynamicFees.MaxMultiplier*80/100 {
		recommendations = append(recommendations, "Gas prices are high - implement transaction batching to reduce costs")
	}

	// Count large tx records
	largeTxCount := uint64(0)
	err := k.IterateLargeTxRecords(ctx, func(record *types.LargeTxRecord) bool {
		largeTxCount++
		return false // Continue iterating
	})
	if err != nil {
		return nil, err
	}

	if largeTxCount > 50 {
		recommendations = append(recommendations, "High whale activity detected - strengthen anti-whale protection")
	}

	return recommendations, nil
}

// SimulateIncentiveChange simulates the impact of changing incentive parameters
func (k *Keeper) SimulateIncentiveChange(
	ctx context.Context,
	newInflationRate uint64,
	newMEVUserPercentage uint64,
	newMEVValidatorPercentage uint64,
) (string, string, string, error) {
	params, _ := k.GetParams(ctx)

	// Simulate new distribution
	totalMEV := new(big.Int)
	totalMEV.SetString(params.Mev.TotalMevCaptured, 10)

	// Calculate new user share
	newUserShare := new(big.Int).Mul(totalMEV, big.NewInt(int64(newMEVUserPercentage)))
	newUserShare.Div(newUserShare, big.NewInt(10000))

	// Calculate new validator share
	newValidatorShare := new(big.Int).Mul(totalMEV, big.NewInt(int64(newMEVValidatorPercentage)))
	newValidatorShare.Div(newValidatorShare, big.NewInt(10000))

	// Calculate new inflation impact
	totalSupply := new(big.Int)
	totalSupply.SetString(params.Tokenomics.CirculatingSupply, 10)

	newEmission := new(big.Int).Mul(totalSupply, big.NewInt(int64(newInflationRate)))
	newEmission.Div(newEmission, big.NewInt(10000))

	return newUserShare.String(), newValidatorShare.String(), newEmission.String(), nil
}

// CalculateOptimalIncentiveDistribution calculates optimal incentive distribution across stakeholders
// DETERMINISM: This function uses math.LegacyDec and integer basis points for all calculations
// to ensure cross-platform consistency. Float64 operations can produce different results on
// different CPU architectures, causing consensus failures in blockchain state machines.
// stakingRatioBps is the staking ratio in basis points (e.g., 3000 = 30%)
func (k *Keeper) CalculateOptimalIncentiveDistribution(
	ctx context.Context,
	totalRewards string,
	stakingRatioBps uint64,
	activeUserCount uint64,
	validatorCount uint64,
) (map[string]string, error) {
	rewards := new(big.Int)
	if _, ok := rewards.SetString(totalRewards, 10); !ok {
		return nil, types.ErrInvalidAmount
	}

	distribution := make(map[string]string)

	// Optimal distribution based on network health (all percentages in basis points)
	// Base allocation: 40% validators (4000 bps), 30% users (3000 bps), 20% treasury (2000 bps), 10% burn (1000 bps)

	// Adjust based on staking ratio (using basis points for comparison)
	validatorPercentageBps := uint64(4000) // 40%
	if stakingRatioBps < 3000 {            // Less than 30% staked
		validatorPercentageBps = 5000 // Increase to 50% to incentivize staking
	} else if stakingRatioBps > 7000 { // More than 70% staked
		validatorPercentageBps = 3000 // Decrease to 30% as already high
	}

	// Adjust based on user activity
	userPercentageBps := uint64(3000) // 30%
	if activeUserCount < 100 {
		userPercentageBps = 4000 // Increase to 40% to attract users
	}

	// Burn is fixed at 10% (1000 bps)
	burnPercentageBps := uint64(1000)

	// Treasury gets remainder to balance (10000 - validators - users - burn)
	treasuryPercentageBps := uint64(10000) - validatorPercentageBps - userPercentageBps - burnPercentageBps

	// Calculate amounts using math.LegacyDec for deterministic division
	rewardsDec := math.LegacyNewDecFromBigInt(rewards)

	validatorRewardsDec := rewardsDec.MulInt64(int64(validatorPercentageBps)).QuoInt64(10000)
	distribution["validators"] = validatorRewardsDec.TruncateInt().String()

	userRewardsDec := rewardsDec.MulInt64(int64(userPercentageBps)).QuoInt64(10000)
	distribution["users"] = userRewardsDec.TruncateInt().String()

	treasuryRewardsDec := rewardsDec.MulInt64(int64(treasuryPercentageBps)).QuoInt64(10000)
	distribution["treasury"] = treasuryRewardsDec.TruncateInt().String()

	burnAmountDec := rewardsDec.MulInt64(int64(burnPercentageBps)).QuoInt64(10000)
	distribution["burn"] = burnAmountDec.TruncateInt().String()

	// Convert basis points to percentages for display (integer division for determinism)
	validatorPct := validatorPercentageBps / 100
	userPct := userPercentageBps / 100
	treasuryPct := treasuryPercentageBps / 100
	stakingPct := stakingRatioBps / 100

	distribution["reasoning"] = fmt.Sprintf(
		"Optimal distribution: Validators %d%%, Users %d%%, Treasury %d%%, Burn 10%% (based on staking ratio %d%% and %d active users)",
		validatorPct, userPct, treasuryPct, stakingPct, activeUserCount,
	)

	return distribution, nil
}
