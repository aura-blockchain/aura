// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"
)

func TestTypeExports(t *testing.T) {
	// Test that all type exports compile correctly
	var _ VestingType
	var _ ScheduleType
	var _ MEVRedistributionStrategy
	var _ InflationAlertType
	var _ AlertSeverity
	var _ VestingSchedule
	var _ VoteLock
	var _ TokenomicsConfig
	var _ DynamicFeeConfig
	var _ TransferTaxConfig
	var _ LiquidityMiningConfig
	var _ MEVConfig
	var _ GovernanceConfig
	var _ WhaleProtection
	var _ LargeTxRecord
	var _ InflationAlert
	var _ TreasuryMultisig
	var _ PendingTreasuryTx
	var _ Params
	var _ GenesisState
}

func TestMessageTypeExports(t *testing.T) {
	var _ MsgCreateVestingSchedule
	var _ MsgRevokeVestingSchedule
	var _ MsgReleaseVestedTokens
	var _ MsgLockVotingTokens
	var _ MsgUnlockVotingTokens
	var _ MsgUpdateParams
	var _ MsgAdjustInflationRate
	var _ MsgProposeTreasurySpend
	var _ MsgSignTreasurySpend
	var _ MsgExecuteTreasurySpend
}

func TestQueryTypeExports(t *testing.T) {
	var _ QueryVestingScheduleRequest
	var _ QueryVestingScheduleResponse
	var _ QueryVoteLockRequest
	var _ QueryVoteLockResponse
	var _ QueryVotingPowerRequest
	var _ QueryVotingPowerResponse
	var _ QueryTokenomicsStatsRequest
	var _ QueryTokenomicsStatsResponse
	var _ QueryMEVStatsRequest
	var _ QueryMEVStatsResponse
	var _ QueryInflationMetricsRequest
	var _ QueryInflationMetricsResponse
	var _ QueryParamsRequest
	var _ QueryParamsResponse
}

func TestEnumConstants(t *testing.T) {
	// Test VestingType constants
	_ = VestingType_VESTING_TYPE_UNSPECIFIED
	_ = VestingType_VESTING_TYPE_LINEAR
	_ = VestingType_VESTING_TYPE_MILESTONE
	_ = VestingType_VESTING_TYPE_CLIFF_THEN_LINEAR

	// Test ScheduleType constants
	_ = ScheduleType_SCHEDULE_TYPE_UNSPECIFIED
	_ = ScheduleType_SCHEDULE_TYPE_TEAM
	_ = ScheduleType_SCHEDULE_TYPE_INVESTOR

	// Test MEVRedistributionStrategy constants
	_ = MEVRedistributionStrategy_MEV_STRATEGY_UNSPECIFIED
	_ = MEVRedistributionStrategy_MEV_STRATEGY_PROPORTIONAL_TO_STAKE

	// Test InflationAlertType constants
	_ = InflationAlertType_INFLATION_ALERT_TYPE_UNSPECIFIED
	_ = InflationAlertType_INFLATION_ALERT_TYPE_ABOVE_TARGET

	// Test AlertSeverity constants
	_ = AlertSeverity_ALERT_SEVERITY_UNSPECIFIED
	_ = AlertSeverity_ALERT_SEVERITY_INFO
	_ = AlertSeverity_ALERT_SEVERITY_WARNING
	_ = AlertSeverity_ALERT_SEVERITY_CRITICAL
}

func TestCircuitBreakerType_String(t *testing.T) {
	tests := []struct {
		input    CircuitBreakerType
		expected string
	}{
		{CircuitBreakerTypeUnspecified, "UNSPECIFIED"},
		{CircuitBreakerTypePriceVolatility, "PRICE_VOLATILITY"},
		{CircuitBreakerTypeLargeTransaction, "LARGE_TRANSACTION"},
		{CircuitBreakerTypeSupplyChange, "SUPPLY_CHANGE"},
		{CircuitBreakerTypeLiquidityCrisis, "LIQUIDITY_CRISIS"},
		{CircuitBreakerTypeGasSpike, "GAS_SPIKE"},
		{CircuitBreakerType(999), "UNSPECIFIED"}, // Unknown value
	}

	for _, tt := range tests {
		result := tt.input.String()
		if result != tt.expected {
			t.Errorf("CircuitBreakerType(%d).String() = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestAttackType_String(t *testing.T) {
	tests := []struct {
		input    AttackType
		expected string
	}{
		{AttackTypeUnspecified, "UNSPECIFIED"},
		{AttackTypePumpAndDump, "PUMP_AND_DUMP"},
		{AttackTypeFlashLoan, "FLASH_LOAN"},
		{AttackTypeSybil, "SYBIL"},
		{AttackTypeWashTrading, "WASH_TRADING"},
		{AttackTypeFrontRunning, "FRONT_RUNNING"},
		{AttackType(999), "UNSPECIFIED"}, // Unknown value
	}

	for _, tt := range tests {
		result := tt.input.String()
		if result != tt.expected {
			t.Errorf("AttackType(%d).String() = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
