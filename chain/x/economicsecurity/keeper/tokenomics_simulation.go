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
// TOKENOMICS SIMULATION (Feature 9)
// ============================

// SimulationParameters contains parameters for tokenomics simulation
type SimulationParameters struct {
	DurationBlocks       uint64
	InitialSupply        string
	InflationRate        uint64
	BurnRate             uint64
	StakingRatio         uint64
	ActiveUsers          uint64
	TransactionsPerBlock uint64
	AverageGasPrice      string
}

// SimulationResult contains the results of tokenomics simulation
type SimulationResult struct {
	FinalSupply          string
	TotalMinted          string
	TotalBurned          string
	TotalStakingRewards  string
	TotalMEVGenerated    string
	TotalFeesCollected   string
	AverageInflationRate uint64
	SupplyGrowthRate     float64
	Projections          []*SimulationSnapshot
}

// SimulationSnapshot represents the state at a specific block
type SimulationSnapshot struct {
	BlockHeight     uint64
	Supply          string
	InflationRate   uint64
	BurnRate        uint64
	StakedAmount    string
	MEVAccumulated  string
	FeesAccumulated string
}

// SimulateTokenomics runs a comprehensive tokenomics simulation
func (k *Keeper) SimulateTokenomics(ctx context.Context, params SimulationParameters) (*SimulationResult, error) {
	result := &SimulationResult{
		Projections: []*SimulationSnapshot{},
	}

	// Initialize simulation state
	currentSupply := new(big.Int)
	if _, ok := currentSupply.SetString(params.InitialSupply, 10); !ok {
		return nil, types.ErrInvalidAmount
	}

	totalMinted := big.NewInt(0)
	totalBurned := big.NewInt(0)
	totalStakingRewards := big.NewInt(0)
	totalMEVGenerated := big.NewInt(0)
	totalFeesCollected := big.NewInt(0)

	// Run simulation for specified duration
	snapshotInterval := params.DurationBlocks / 100 // Take 100 snapshots
	if snapshotInterval == 0 {
		snapshotInterval = 1
	}

	for block := uint64(0); block < params.DurationBlocks; block++ {
		// 1. Calculate block inflation
		inflation := new(big.Int).Mul(currentSupply, big.NewInt(int64(params.InflationRate)))
		inflation.Div(inflation, big.NewInt(10000*365*24*60*10)) // Per block (6s blocks)

		// 2. Calculate staking rewards
		stakedAmount := new(big.Int).Mul(currentSupply, big.NewInt(int64(params.StakingRatio)))
		stakedAmount.Div(stakedAmount, big.NewInt(10000))

		stakingReward := new(big.Int).Mul(inflation, big.NewInt(70)) // 70% to staking
		stakingReward.Div(stakingReward, big.NewInt(100))

		// 3. Calculate MEV generation
		mevPerBlock := new(big.Int).SetUint64(params.TransactionsPerBlock * 1000) // Simplified MEV

		// 4. Calculate transaction fees
		gasPrice := new(big.Int)
		gasPrice.SetString(params.AverageGasPrice, 10)
		fees := new(big.Int).Mul(gasPrice, big.NewInt(int64(params.TransactionsPerBlock*21000)))

		// 5. Calculate burn
		burnAmount := new(big.Int).Mul(fees, big.NewInt(int64(params.BurnRate)))
		burnAmount.Div(burnAmount, big.NewInt(10000))

		// Update totals
		currentSupply.Add(currentSupply, inflation)
		currentSupply.Sub(currentSupply, burnAmount)
		totalMinted.Add(totalMinted, inflation)
		totalBurned.Add(totalBurned, burnAmount)
		totalStakingRewards.Add(totalStakingRewards, stakingReward)
		totalMEVGenerated.Add(totalMEVGenerated, mevPerBlock)
		totalFeesCollected.Add(totalFeesCollected, fees)

		// Take snapshot at intervals
		if block%snapshotInterval == 0 {
			snapshot := &SimulationSnapshot{
				BlockHeight:     block,
				Supply:          currentSupply.String(),
				InflationRate:   params.InflationRate,
				BurnRate:        params.BurnRate,
				StakedAmount:    stakedAmount.String(),
				MEVAccumulated:  totalMEVGenerated.String(),
				FeesAccumulated: totalFeesCollected.String(),
			}
			result.Projections = append(result.Projections, snapshot)
		}
	}

	// Calculate final metrics
	result.FinalSupply = currentSupply.String()
	result.TotalMinted = totalMinted.String()
	result.TotalBurned = totalBurned.String()
	result.TotalStakingRewards = totalStakingRewards.String()
	result.TotalMEVGenerated = totalMEVGenerated.String()
	result.TotalFeesCollected = totalFeesCollected.String()

	// Calculate average inflation rate
	initialSupply := new(big.Int)
	initialSupply.SetString(params.InitialSupply, 10)

	supplyChange := new(big.Int).Sub(currentSupply, initialSupply)
	growthRate := new(big.Float).Quo(new(big.Float).SetInt(supplyChange), new(big.Float).SetInt(initialSupply))
	growthRateFloat, _ := growthRate.Float64()
	result.SupplyGrowthRate = growthRateFloat

	result.AverageInflationRate = params.InflationRate

	return result, nil
}

// SimulateSupplyScenarios simulates multiple supply scenarios
func (k *Keeper) SimulateSupplyScenarios(ctx context.Context) map[string]*SimulationResult {
	currentParams, _ := k.GetParams(ctx)

	scenarios := make(map[string]*SimulationResult)

	// Scenario 1: Current parameters
	scenarios["current"], _ = k.SimulateTokenomics(ctx, SimulationParameters{
		DurationBlocks:       525600, // ~1 year
		InitialSupply:        currentParams.Tokenomics.CirculatingSupply,
		InflationRate:        currentParams.Tokenomics.InflationRate,
		BurnRate:             500,  // 5%
		StakingRatio:         5000, // 50%
		ActiveUsers:          1000,
		TransactionsPerBlock: 100,
		AverageGasPrice:      "1000000",
	})

	// Scenario 2: High inflation
	scenarios["high_inflation"], _ = k.SimulateTokenomics(ctx, SimulationParameters{
		DurationBlocks:       525600,
		InitialSupply:        currentParams.Tokenomics.CirculatingSupply,
		InflationRate:        currentParams.Tokenomics.MaxInflationRate,
		BurnRate:             500,
		StakingRatio:         5000,
		ActiveUsers:          1000,
		TransactionsPerBlock: 100,
		AverageGasPrice:      "1000000",
	})

	// Scenario 3: Low inflation
	scenarios["low_inflation"], _ = k.SimulateTokenomics(ctx, SimulationParameters{
		DurationBlocks:       525600,
		InitialSupply:        currentParams.Tokenomics.CirculatingSupply,
		InflationRate:        currentParams.Tokenomics.MinInflationRate,
		BurnRate:             500,
		StakingRatio:         5000,
		ActiveUsers:          1000,
		TransactionsPerBlock: 100,
		AverageGasPrice:      "1000000",
	})

	// Scenario 4: High burn rate
	scenarios["high_burn"], _ = k.SimulateTokenomics(ctx, SimulationParameters{
		DurationBlocks:       525600,
		InitialSupply:        currentParams.Tokenomics.CirculatingSupply,
		InflationRate:        currentParams.Tokenomics.InflationRate,
		BurnRate:             2000, // 20%
		StakingRatio:         5000,
		ActiveUsers:          1000,
		TransactionsPerBlock: 100,
		AverageGasPrice:      "1000000",
	})

	// Scenario 5: High adoption
	scenarios["high_adoption"], _ = k.SimulateTokenomics(ctx, SimulationParameters{
		DurationBlocks:       525600,
		InitialSupply:        currentParams.Tokenomics.CirculatingSupply,
		InflationRate:        currentParams.Tokenomics.InflationRate,
		BurnRate:             500,
		StakingRatio:         5000,
		ActiveUsers:          10000,
		TransactionsPerBlock: 1000,
		AverageGasPrice:      "2000000",
	})

	return scenarios
}

// ProjectSupplyGrowth projects supply growth over time periods
func (k *Keeper) ProjectSupplyGrowth(ctx context.Context, years uint64) (map[string]string, error) {
	currentParams, _ := k.GetParams(ctx)
	projections := make(map[string]string)

	currentSupply := new(big.Int)
	currentSupply.SetString(currentParams.Tokenomics.CirculatingSupply, 10)

	blocksPerYear := uint64(525600) // Assuming 6s blocks

	for year := uint64(1); year <= years; year++ {
		params := SimulationParameters{
			DurationBlocks:       blocksPerYear,
			InitialSupply:        currentSupply.String(),
			InflationRate:        currentParams.Tokenomics.InflationRate,
			BurnRate:             500,
			StakingRatio:         5000,
			ActiveUsers:          1000,
			TransactionsPerBlock: 100,
			AverageGasPrice:      "1000000",
		}

		result, err := k.SimulateTokenomics(ctx, params)
		if err != nil {
			break
		}

		projections[fmt.Sprintf("year_%d", year)] = result.FinalSupply

		// Update for next iteration
		currentSupply.SetString(result.FinalSupply, 10)
	}

	return projections, nil
}

// AnalyzeTokenDistribution analyzes token distribution patterns
func (k *Keeper) AnalyzeTokenDistribution(ctx context.Context) (map[string]interface{}, error) {
	params, _ := k.GetParams(ctx)

	totalSupply := new(big.Int)
	totalSupply.SetString(params.Tokenomics.CirculatingSupply, 10)

	analysis := make(map[string]interface{})

	// Calculate distribution metrics
	analysis["total_supply"] = totalSupply.String()
	analysis["max_supply"] = params.Tokenomics.MaxSupply

	// Calculate remaining mintable supply
	maxSupply := new(big.Int)
	maxSupply.SetString(params.Tokenomics.MaxSupply, 10)
	remaining := new(big.Int).Sub(maxSupply, totalSupply)
	analysis["remaining_mintable"] = remaining.String()

	// Calculate supply ratio
	supplyRatio := new(big.Int).Mul(totalSupply, big.NewInt(10000))
	supplyRatio.Div(supplyRatio, maxSupply)
	analysis["supply_ratio_bps"] = supplyRatio.Uint64()

	// MEV distribution
	totalMEVCaptured := new(big.Int)
	totalMEVCaptured.SetString(params.Mev.TotalMevCaptured, 10)
	analysis["total_mev_captured"] = totalMEVCaptured.String()

	userMEV := new(big.Int).Mul(totalMEVCaptured, big.NewInt(int64(params.Mev.UserRedistributionPercentage)))
	userMEV.Div(userMEV, big.NewInt(10000))
	analysis["user_mev_share"] = userMEV.String()

	validatorMEV := new(big.Int).Mul(totalMEVCaptured, big.NewInt(int64(params.Mev.ValidatorPercentage)))
	validatorMEV.Div(validatorMEV, big.NewInt(10000))
	analysis["validator_mev_share"] = validatorMEV.String()

	// Burn statistics
	burnAmount, err := k.GetTotalBurned(ctx)
	if err != nil {
		return nil, err
	}

	analysis["total_burned"] = burnAmount

	burnRatio := new(big.Int)
	burnRatio.SetString(burnAmount, 10)
	burnRatio.Mul(burnRatio, big.NewInt(10000))
	burnRatio.Div(burnRatio, totalSupply)
	analysis["burn_ratio_bps"] = burnRatio.Uint64()

	return analysis, nil
}

// OptimizeTokenomicsParameters suggests optimal tokenomics parameters
func (k *Keeper) OptimizeTokenomicsParameters(ctx context.Context, targetSupplyGrowth float64, years uint64) map[string]uint64 {
	recommendations := make(map[string]uint64)

	// Run simulations with different parameters to find optimal settings
	bestInflation := uint64(500) // 5%
	bestBurn := uint64(500)      // 5%

	minDiff := float64(1000000.0)

	// Test different combinations
	for inflation := uint64(200); inflation <= 1000; inflation += 100 {
		for burn := uint64(100); burn <= 1000; burn += 100 {
			params := SimulationParameters{
				DurationBlocks:       525600 * years,
				InitialSupply:        "1000000000000000000000000", // 1M tokens
				InflationRate:        inflation,
				BurnRate:             burn,
				StakingRatio:         5000,
				ActiveUsers:          1000,
				TransactionsPerBlock: 100,
				AverageGasPrice:      "1000000",
			}

			result, err := k.SimulateTokenomics(ctx, params)
			if err != nil {
				continue
			}

			diff := result.SupplyGrowthRate - targetSupplyGrowth
			if diff < 0 {
				diff = -diff
			}

			if diff < minDiff {
				minDiff = diff
				bestInflation = inflation
				bestBurn = burn
			}
		}
	}

	recommendations["optimal_inflation_rate"] = bestInflation
	recommendations["optimal_burn_rate"] = bestBurn
	recommendations["target_staking_ratio"] = 5000 // 50%

	return recommendations
}
