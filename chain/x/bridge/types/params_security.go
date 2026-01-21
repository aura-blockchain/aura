// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Additional parameter store keys for security features
var (
	KeyEmergencyPaused              = []byte("EmergencyPaused")
	KeyMinTransferAmount            = []byte("MinTransferAmount")
	KeyMaxTransferAmount            = []byte("MaxTransferAmount")
	KeyTimeLockDuration             = []byte("TimeLockDuration")
	KeyTimeLockThreshold            = []byte("TimeLockThreshold")
	KeyDailyWithdrawalLimit         = []byte("DailyWithdrawalLimit")
	KeyCircuitBreakerEnabled        = []byte("CircuitBreakerEnabled")
	KeyMaxHourlyVolume              = []byte("MaxHourlyVolume")
	KeyMaxFailedTransfersPerHour    = []byte("MaxFailedTransfersPerHour")
	KeyMinValidatorSignatures       = []byte("MinValidatorSignatures")
	KeyValidatorRotationPeriod      = []byte("ValidatorRotationPeriod")
	KeySlashFractionInvalidProof    = []byte("SlashFractionInvalidProof")
	KeySlashFractionDoubleSign      = []byte("SlashFractionDoubleSign")
	KeySlashFractionDowntime        = []byte("SlashFractionDowntime")
	KeyFraudProofReward             = []byte("FraudProofReward")
	KeyFraudProofWindowDuration     = []byte("FraudProofWindowDuration")
	KeyFixedTransferFee             = []byte("FixedTransferFee")
	KeyPercentageFeeBPS             = []byte("PercentageFeeBPS")
	KeyInsuranceFundContributionBPS = []byte("InsuranceFundContributionBPS")
	KeyWhitelistEnabled             = []byte("WhitelistEnabled")
)

// SecurityParams defines the security parameters for the Bridge module
type SecurityParams struct {
	// Emergency controls
	EmergencyPaused   bool     `json:"emergency_paused"`
	MinTransferAmount math.Int `json:"min_transfer_amount"`
	MaxTransferAmount math.Int `json:"max_transfer_amount"`

	// Time-lock parameters
	TimeLockDuration  time.Duration `json:"time_lock_duration"`
	TimeLockThreshold math.Int      `json:"time_lock_threshold"`

	// Withdrawal limits
	DailyWithdrawalLimit math.Int `json:"daily_withdrawal_limit"`

	// Circuit breaker parameters
	CircuitBreakerEnabled     bool     `json:"circuit_breaker_enabled"`
	MaxHourlyVolume           math.Int `json:"max_hourly_volume"`
	MaxFailedTransfersPerHour uint64   `json:"max_failed_transfers_per_hour"`

	// Validator parameters
	MinValidatorSignatures  uint64        `json:"min_validator_signatures"`
	ValidatorRotationPeriod time.Duration `json:"validator_rotation_period"`

	// Slashing parameters
	SlashFractionInvalidProof math.LegacyDec `json:"slash_fraction_invalid_proof"`
	SlashFractionDoubleSign   math.LegacyDec `json:"slash_fraction_double_sign"`
	SlashFractionDowntime     math.LegacyDec `json:"slash_fraction_downtime"`

	// Fraud proof parameters
	FraudProofReward         math.Int      `json:"fraud_proof_reward"`
	FraudProofWindowDuration time.Duration `json:"fraud_proof_window_duration"`

	// Fee parameters
	FixedTransferFee math.Int `json:"fixed_transfer_fee"`
	PercentageFeeBPS uint64   `json:"percentage_fee_bps"`

	// Insurance fund parameters
	InsuranceFundContributionBPS uint64 `json:"insurance_fund_contribution_bps"`

	// Security parameters
	WhitelistEnabled bool `json:"whitelist_enabled"`
}

// DefaultSecurityParams returns default security parameters
func DefaultSecurityParams() SecurityParams {
	return SecurityParams{
		EmergencyPaused:              false,
		MinTransferAmount:            math.NewInt(1000000),      // 1 token (6 decimals)
		MaxTransferAmount:            math.NewInt(100000000000), // 100,000 tokens
		TimeLockDuration:             24 * time.Hour,
		TimeLockThreshold:            math.NewInt(10000000000),  // 10,000 tokens
		DailyWithdrawalLimit:         math.NewInt(500000000000), // 500,000 tokens (5x max transfer)
		CircuitBreakerEnabled:        true,
		MaxHourlyVolume:              math.NewInt(1000000000000), // 1,000,000 tokens/hour (10x max transfer)
		MaxFailedTransfersPerHour:    10,
		MinValidatorSignatures:       3,
		ValidatorRotationPeriod:      30 * 24 * time.Hour,              // 30 days
		SlashFractionInvalidProof:    math.LegacyNewDecWithPrec(5, 2),  // 5%
		SlashFractionDoubleSign:      math.LegacyNewDecWithPrec(10, 2), // 10%
		SlashFractionDowntime:        math.LegacyNewDecWithPrec(1, 2),  // 1%
		FraudProofReward:             math.NewInt(1000000000),          // 1,000 tokens
		FraudProofWindowDuration:     7 * 24 * time.Hour,
		FixedTransferFee:             math.NewInt(10000), // 0.01 token (reduced to not exceed 5% for small amounts)
		PercentageFeeBPS:             10,                 // 0.1%
		InsuranceFundContributionBPS: 2000,               // 20% of fees
		WhitelistEnabled:             false,
	}
}

// SecurityParamSetPairs implements params.ParamSet
func (p *SecurityParams) SecurityParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyEmergencyPaused, &p.EmergencyPaused, validateBool),
		paramtypes.NewParamSetPair(KeyMinTransferAmount, &p.MinTransferAmount, validateInt),
		paramtypes.NewParamSetPair(KeyMaxTransferAmount, &p.MaxTransferAmount, validateInt),
		paramtypes.NewParamSetPair(KeyTimeLockDuration, &p.TimeLockDuration, validateDuration),
		paramtypes.NewParamSetPair(KeyTimeLockThreshold, &p.TimeLockThreshold, validateInt),
		paramtypes.NewParamSetPair(KeyDailyWithdrawalLimit, &p.DailyWithdrawalLimit, validateInt),
		paramtypes.NewParamSetPair(KeyCircuitBreakerEnabled, &p.CircuitBreakerEnabled, validateBool),
		paramtypes.NewParamSetPair(KeyMaxHourlyVolume, &p.MaxHourlyVolume, validateInt),
		paramtypes.NewParamSetPair(KeyMaxFailedTransfersPerHour, &p.MaxFailedTransfersPerHour, validateUint64),
		paramtypes.NewParamSetPair(KeyMinValidatorSignatures, &p.MinValidatorSignatures, validateUint64),
		paramtypes.NewParamSetPair(KeyValidatorRotationPeriod, &p.ValidatorRotationPeriod, validateDuration),
		paramtypes.NewParamSetPair(KeySlashFractionInvalidProof, &p.SlashFractionInvalidProof, validateDec),
		paramtypes.NewParamSetPair(KeySlashFractionDoubleSign, &p.SlashFractionDoubleSign, validateDec),
		paramtypes.NewParamSetPair(KeySlashFractionDowntime, &p.SlashFractionDowntime, validateDec),
		paramtypes.NewParamSetPair(KeyFraudProofReward, &p.FraudProofReward, validateInt),
		paramtypes.NewParamSetPair(KeyFraudProofWindowDuration, &p.FraudProofWindowDuration, validateDuration),
		paramtypes.NewParamSetPair(KeyFixedTransferFee, &p.FixedTransferFee, validateInt),
		paramtypes.NewParamSetPair(KeyPercentageFeeBPS, &p.PercentageFeeBPS, validateUint64),
		paramtypes.NewParamSetPair(KeyInsuranceFundContributionBPS, &p.InsuranceFundContributionBPS, validateUint64),
		paramtypes.NewParamSetPair(KeyWhitelistEnabled, &p.WhitelistEnabled, validateBool),
	}
}

// Validation functions
func validateInt(i interface{}) error {
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("value cannot be negative: %s", v.String())
	}
	return nil
}

func validateUint64(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateDuration(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v < 0 {
		return fmt.Errorf("duration cannot be negative: %s", v.String())
	}
	return nil
}

func validateDec(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("value cannot be negative: %s", v.String())
	}
	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("value cannot be greater than 1: %s", v.String())
	}
	return nil
}
