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
