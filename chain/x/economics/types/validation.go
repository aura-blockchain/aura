// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"fmt"

	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
)

// ValidateParams validates all economics module parameters
func ValidateParams(params *economicspb.Params) error {
	if params == nil {
		return fmt.Errorf("params cannot be nil")
	}

	if err := ValidateFeeParams(&params.Fees); err != nil {
		return fmt.Errorf("invalid fee params: %w", err)
	}

	if err := ValidateVestingParams(&params.Vesting); err != nil {
		return fmt.Errorf("invalid vesting params: %w", err)
	}

	if err := ValidateTreasuryParams(&params.Treasury); err != nil {
		return fmt.Errorf("invalid treasury params: %w", err)
	}

	if err := ValidateGovernanceParams(&params.Governance); err != nil {
		return fmt.Errorf("invalid governance params: %w", err)
	}

	if err := ValidateMEVParams(&params.Mev); err != nil {
		return fmt.Errorf("invalid MEV params: %w", err)
	}

	if err := ValidateWhaleProtectionParams(&params.WhaleProtection); err != nil {
		return fmt.Errorf("invalid whale protection params: %w", err)
	}

	if err := ValidateLiquidityMiningParams(&params.LiquidityMining); err != nil {
		return fmt.Errorf("invalid liquidity mining params: %w", err)
	}

	if err := ValidateTokenomicsParams(&params.Tokenomics); err != nil {
		return fmt.Errorf("invalid tokenomics params: %w", err)
	}

	return nil
}

// ValidateFeeParams validates fee parameters
func ValidateFeeParams(params *economicspb.FeeParams) error {
	if params == nil {
		return nil // Optional params
	}

	if params.MinFeeMultiplier > params.MaxFeeMultiplier {
		return fmt.Errorf("min fee multiplier cannot be greater than max fee multiplier")
	}

	if params.TargetBlockUtilization > 10000 {
		return fmt.Errorf("target block utilization cannot exceed 100%%")
	}

	return nil
}

// ValidateVestingParams validates vesting parameters
func ValidateVestingParams(params *economicspb.VestingParams) error {
	if params == nil {
		return nil // Optional params
	}

	if params.MaxVestingDuration < params.MinVestingDuration {
		return fmt.Errorf("max vesting duration must be greater than or equal to min vesting duration")
	}

	if params.EarlyUnlockPenalty > 10000 {
		return fmt.Errorf("early unlock penalty cannot exceed 100%%")
	}

	return nil
}

// ValidateTreasuryParams validates treasury parameters
func ValidateTreasuryParams(params *economicspb.TreasuryParams) error {
	if params == nil {
		return nil // Optional params
	}

	if params.CommunityPoolPercentage > 10000 {
		return fmt.Errorf("community pool percentage cannot exceed 100%%")
	}

	if params.BurnPercentage > 10000 {
		return fmt.Errorf("burn percentage cannot exceed 100%%")
	}

	if params.MultisigThreshold == 0 {
		return fmt.Errorf("multisig threshold must be at least 1")
	}

	return nil
}

// ValidateGovernanceParams validates governance parameters
func ValidateGovernanceParams(params *economicspb.GovernanceParams) error {
	if params == nil {
		return nil // Optional params
	}

	if params.Quorum > 10000 {
		return fmt.Errorf("quorum cannot exceed 100%%")
	}

	if params.Threshold > 10000 {
		return fmt.Errorf("threshold cannot exceed 100%%")
	}

	if params.VetoThreshold > 10000 {
		return fmt.Errorf("veto threshold cannot exceed 100%%")
	}

	return nil
}

// ValidateMEVParams validates MEV parameters
func ValidateMEVParams(params *economicspb.MEVParams) error {
	if params == nil {
		return nil // Optional params
	}

	total := params.UserRedistributionPercentage + params.ValidatorPercentage +
		params.TreasuryPercentage + params.BurnPercentage

	if total != 10000 {
		return fmt.Errorf("MEV redistribution percentages must sum to 100%%, got %d", total)
	}

	return nil
}

// ValidateWhaleProtectionParams validates whale protection parameters
func ValidateWhaleProtectionParams(params *economicspb.WhaleProtectionParams) error {
	if params == nil {
		return nil // Optional params
	}

	if params.MaxHoldingPercentage > 10000 {
		return fmt.Errorf("max holding percentage cannot exceed 100%%")
	}

	if params.LargeTxThreshold > 10000 {
		return fmt.Errorf("large tx threshold cannot exceed 100%%")
	}

	return nil
}

// ValidateLiquidityMiningParams validates liquidity mining parameters
func ValidateLiquidityMiningParams(params *economicspb.LiquidityMiningParams) error {
	if params == nil {
		return nil // Optional params
	}

	if params.EpochDurationBlocks == 0 {
		return fmt.Errorf("epoch duration blocks must be positive")
	}

	return nil
}

// ValidateTokenomicsParams validates tokenomics parameters
func ValidateTokenomicsParams(params *economicspb.TokenomicsParams) error {
	if params == nil {
		return nil // Optional params
	}

	if params.MinInflationRate > params.TargetInflationRate {
		return fmt.Errorf("min inflation rate cannot exceed target inflation rate")
	}

	if params.TargetInflationRate > params.MaxInflationRate {
		return fmt.Errorf("target inflation rate cannot exceed max inflation rate")
	}

	if params.InflationCheckInterval == 0 {
		return fmt.Errorf("inflation check interval must be positive")
	}

	return nil
}
