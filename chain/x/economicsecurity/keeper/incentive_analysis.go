// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"
	"math/big"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// ============================
// ECONOMIC INCENTIVE ANALYSIS (Feature 8)
// ============================

// IncentiveAnalysisResult contains the results of incentive analysis
type IncentiveAnalysisResult struct {
	ValidatorRewards       string
	UserRewards            string
	TreasuryAllocation     string
	BurnAmount             string
	IncentiveEfficiency    float64
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
	result.IncentiveEfficiency = k.calculateIncentiveEfficiency(params, activeUsers, validators)

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

// calculateIncentiveEfficiency calculates overall incentive efficiency score (0-100)
func (k *Keeper) calculateIncentiveEfficiency(params types.Params, activeUsers uint64, validators uint64) float64 {
	score := float64(100)

	// Deduct points for inefficiencies

	// 1. Check inflation alignment (max -20 points)
	inflationDelta := int64(params.Tokenomics.InflationRate) - int64(params.Tokenomics.TargetInflationRate)
	if inflationDelta < 0 {
		inflationDelta = -inflationDelta
	}
	if inflationDelta > 100 { // More than 1% off target
		score -= float64(inflationDelta) / 10
		if score < 80 {
			score = 80
		}
	}

	// 2. Check user participation (max -20 points)
	if activeUsers < 100 {
		score -= 20
	} else if activeUsers < 500 {
		score -= 10
	}

	// 3. Check MEV distribution efficiency (max -20 points)
	if params.Mev.Strategy == types.MEVStrategyEqualDistribution {
		score -= 10 // Equal distribution is less efficient
	}

	// 4. Check treasury balance (max -20 points)
	if params.Mev.TreasuryPercentage < 1000 {
		score -= 15
	} else if params.Mev.TreasuryPercentage > 3000 {
		score -= 10
	}

	// 5. Check validator count (max -20 points)
	if validators < 10 {
		score -= 20 // Very few validators
	} else if validators < 50 {
		score -= 10
	}

	if score < 0 {
		score = 0
	}

	return score
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
func (k *Keeper) CalculateOptimalIncentiveDistribution(
	ctx context.Context,
	totalRewards string,
	stakingRatio float64,
	activeUserCount uint64,
	validatorCount uint64,
) (map[string]string, error) {
	rewards := new(big.Int)
	if _, ok := rewards.SetString(totalRewards, 10); !ok {
		return nil, types.ErrInvalidAmount
	}

	distribution := make(map[string]string)

	// Optimal distribution based on network health
	// Base allocation: 40% validators, 30% users, 20% treasury, 10% burn

	// Adjust based on staking ratio
	validatorPercentage := float64(40)
	if stakingRatio < 0.3 {
		validatorPercentage = 50 // Increase to incentivize staking
	} else if stakingRatio > 0.7 {
		validatorPercentage = 30 // Decrease as already high
	}

	// Adjust based on user activity
	userPercentage := float64(30)
	if activeUserCount < 100 {
		userPercentage = 40 // Increase to attract users
	}

	// Treasury gets remainder to balance
	treasuryPercentage := 100 - validatorPercentage - userPercentage - 10

	// Calculate amounts
	validatorRewards := new(big.Int).Mul(rewards, big.NewInt(int64(validatorPercentage*100)))
	validatorRewards.Div(validatorRewards, big.NewInt(10000))
	distribution["validators"] = validatorRewards.String()

	userRewards := new(big.Int).Mul(rewards, big.NewInt(int64(userPercentage*100)))
	userRewards.Div(userRewards, big.NewInt(10000))
	distribution["users"] = userRewards.String()

	treasuryRewards := new(big.Int).Mul(rewards, big.NewInt(int64(treasuryPercentage*100)))
	treasuryRewards.Div(treasuryRewards, big.NewInt(10000))
	distribution["treasury"] = treasuryRewards.String()

	burnAmount := new(big.Int).Mul(rewards, big.NewInt(1000)) // 10%
	burnAmount.Div(burnAmount, big.NewInt(10000))
	distribution["burn"] = burnAmount.String()

	distribution["reasoning"] = fmt.Sprintf(
		"Optimal distribution: Validators %.1f%%, Users %.1f%%, Treasury %.1f%%, Burn 10%% (based on staking ratio %.2f and %d active users)",
		validatorPercentage, userPercentage, treasuryPercentage, stakingRatio, activeUserCount,
	)

	return distribution, nil
}
