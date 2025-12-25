// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"errors"
	"testing"
)

func TestErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"ErrUnauthorized", ErrUnauthorized, "unauthorized"},
		{"ErrInvalidAmount", ErrInvalidAmount, "invalid amount"},
		{"ErrInvalidAddress", ErrInvalidAddress, "invalid address"},
		{"ErrInvalidDuration", ErrInvalidDuration, "invalid duration"},
		{"ErrInvalidScheduleID", ErrInvalidScheduleID, "invalid schedule ID"},
		{"ErrMaxSupplyExceeded", ErrMaxSupplyExceeded, "maximum supply cap exceeded"},
		{"ErrSupplyCapAlreadySet", ErrSupplyCapAlreadySet, "supply cap already set and is immutable"},
		{"ErrInvalidSupplyCap", ErrInvalidSupplyCap, "invalid supply cap"},
		{"ErrInflationRateTooHigh", ErrInflationRateTooHigh, "inflation rate exceeds maximum"},
		{"ErrInflationRateTooLow", ErrInflationRateTooLow, "inflation rate below minimum"},
		{"ErrInvalidInflationRate", ErrInvalidInflationRate, "invalid inflation rate"},
		{"ErrVestingScheduleNotFound", ErrVestingScheduleNotFound, "vesting schedule not found"},
		{"ErrVestingAlreadyRevoked", ErrVestingAlreadyRevoked, "vesting schedule already revoked"},
		{"ErrNoVestedTokens", ErrNoVestedTokens, "no tokens available to vest"},
		{"ErrCliffNotReached", ErrCliffNotReached, "cliff period not yet reached"},
		{"ErrInvalidBeneficiary", ErrInvalidBeneficiary, "invalid beneficiary address"},
		{"ErrInsufficientVestedAmount", ErrInsufficientVestedAmount, "insufficient vested amount"},
		{"ErrWhaleHoldingLimitExceeded", ErrWhaleHoldingLimitExceeded, "whale holding limit exceeded"},
		{"ErrWhaleTxLimitExceeded", ErrWhaleTxLimitExceeded, "whale transaction limit exceeded"},
		{"ErrLargeTxCooldownActive", ErrLargeTxCooldownActive, "large transaction cooldown period active"},
		{"ErrInvalidWhaleConfig", ErrInvalidWhaleConfig, "invalid whale protection configuration"},
		{"ErrInvalidTaxConfig", ErrInvalidTaxConfig, "invalid transfer tax configuration"},
		{"ErrTaxRateTooHigh", ErrTaxRateTooHigh, "tax rate exceeds maximum"},
		{"ErrInvalidTaxRecipient", ErrInvalidTaxRecipient, "invalid tax recipient address"},
		{"ErrLiquidityRewardCapExceeded", ErrLiquidityRewardCapExceeded, "liquidity mining reward cap exceeded"},
		{"ErrInvalidEpoch", ErrInvalidEpoch, "invalid epoch"},
		{"ErrInsufficientRewards", ErrInsufficientRewards, "insufficient rewards available"},
		{"ErrLiquidityMiningDisabled", ErrLiquidityMiningDisabled, "liquidity mining disabled"},
		{"ErrInsufficientStake", ErrInsufficientStake, "insufficient stake for governance proposal"},
		{"ErrInvalidProposalDeposit", ErrInvalidProposalDeposit, "invalid proposal deposit"},
		{"ErrInvalidQuorum", ErrInvalidQuorum, "invalid quorum percentage"},
		{"ErrInvalidThreshold", ErrInvalidThreshold, "invalid threshold percentage"},
		{"ErrVoteLockNotFound", ErrVoteLockNotFound, "vote lock not found"},
		{"ErrVoteLockNotExpired", ErrVoteLockNotExpired, "vote lock has not expired yet"},
		{"ErrInvalidLockDuration", ErrInvalidLockDuration, "invalid lock duration"},
		{"ErrLockDurationTooShort", ErrLockDurationTooShort, "lock duration below minimum"},
		{"ErrLockDurationTooLong", ErrLockDurationTooLong, "lock duration exceeds maximum"},
		{"ErrVoteLockAlreadyWithdrawn", ErrVoteLockAlreadyWithdrawn, "vote lock already withdrawn"},
		{"ErrInvalidTreasuryAddress", ErrInvalidTreasuryAddress, "invalid treasury address"},
		{"ErrInvalidThresholdValue", ErrInvalidThresholdValue, "invalid threshold value"},
		{"ErrInsufficientSignatures", ErrInsufficientSignatures, "insufficient signatures"},
		{"ErrTxNotFound", ErrTxNotFound, "treasury transaction not found"},
		{"ErrTxAlreadyExecuted", ErrTxAlreadyExecuted, "transaction already executed"},
		{"ErrTxAlreadyRejected", ErrTxAlreadyRejected, "transaction already rejected"},
		{"ErrTimelockNotExpired", ErrTimelockNotExpired, "timelock period not expired"},
		{"ErrInvalidSigner", ErrInvalidSigner, "invalid signer"},
		{"ErrAlreadySigned", ErrAlreadySigned, "already signed by this address"},
		{"ErrInsufficientTreasuryBalance", ErrInsufficientTreasuryBalance, "insufficient treasury balance"},
		{"ErrInvalidFeeMultiplier", ErrInvalidFeeMultiplier, "invalid fee multiplier"},
		{"ErrInvalidTargetUtilization", ErrInvalidTargetUtilization, "invalid target utilization"},
		{"ErrInvalidAdjustmentSpeed", ErrInvalidAdjustmentSpeed, "invalid adjustment speed"},
		{"ErrMEVRedistributionDisabled", ErrMEVRedistributionDisabled, "MEV redistribution disabled"},
		{"ErrInvalidMEVConfig", ErrInvalidMEVConfig, "invalid MEV configuration"},
		{"ErrInvalidRedistributionStrategy", ErrInvalidRedistributionStrategy, "invalid redistribution strategy"},
		{"ErrInsufficientMEVBalance", ErrInsufficientMEVBalance, "insufficient MEV balance"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("expected error message %q, got %q", tt.expected, tt.err.Error())
			}
		})
	}
}

func TestErrorsAreErrors(t *testing.T) {
	// Test that all defined error variables implement the error interface
	var _ error = ErrUnauthorized
	var _ error = ErrInvalidAmount
	var _ error = ErrInvalidAddress
	var _ error = ErrMaxSupplyExceeded
	var _ error = ErrInflationRateTooHigh
	var _ error = ErrVestingScheduleNotFound
	var _ error = ErrWhaleHoldingLimitExceeded
	var _ error = ErrInvalidTaxConfig
	var _ error = ErrLiquidityRewardCapExceeded
	var _ error = ErrInsufficientStake
	var _ error = ErrVoteLockNotFound
	var _ error = ErrInvalidTreasuryAddress
	var _ error = ErrInvalidFeeMultiplier
	var _ error = ErrMEVRedistributionDisabled
}

func TestErrorComparison(t *testing.T) {
	// Test that errors can be compared using errors.Is
	err := ErrInvalidAmount
	if !errors.Is(err, ErrInvalidAmount) {
		t.Error("errors.Is should return true for same error")
	}

	if errors.Is(err, ErrInvalidAddress) {
		t.Error("errors.Is should return false for different error")
	}
}
