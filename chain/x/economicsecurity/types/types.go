package types

import (
	economicsecuritypb "github.com/aequitas/aura/proto/aura/economicsecurity/v1beta1"
)

const (
	// ModuleName defines the module name
	ModuleName = "economicsecurity"

	// StoreKey is the default store key for the module
	StoreKey = ModuleName

	// RouterKey is the message route for the module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for the module
	QuerierRoute = ModuleName

	// Basis points precision (10000 = 100%)
	BasisPoints = 10000
)

// Type aliases for proto types
type (
	Params                    = economicsecuritypb.Params
	TokenomicsConfig          = economicsecuritypb.TokenomicsConfig
	VestingSchedule           = economicsecuritypb.VestingSchedule
	VestingType               = economicsecuritypb.VestingType
	ScheduleType              = economicsecuritypb.ScheduleType
	WhaleProtection           = economicsecuritypb.WhaleProtection
	TransferTaxConfig         = economicsecuritypb.TransferTaxConfig
	LiquidityMiningConfig     = economicsecuritypb.LiquidityMiningConfig
	GovernanceConfig          = economicsecuritypb.GovernanceConfig
	VoteLock                  = economicsecuritypb.VoteLock
	TreasuryMultisig          = economicsecuritypb.TreasuryMultisig
	PendingTreasuryTx         = economicsecuritypb.PendingTreasuryTx
	DynamicFeeConfig          = economicsecuritypb.DynamicFeeConfig
	MEVConfig                 = economicsecuritypb.MEVConfig
	MEVRedistributionStrategy = economicsecuritypb.MEVRedistributionStrategy
	InflationAlert            = economicsecuritypb.InflationAlert
	InflationAlertType        = economicsecuritypb.InflationAlertType
	AlertSeverity             = economicsecuritypb.AlertSeverity
	LargeTxRecord             = economicsecuritypb.LargeTxRecord
	GenesisState              = economicsecuritypb.GenesisState
)

// Enum constants
const (
	// VestingType
	VestingTypeUnspecified     = economicsecuritypb.VestingType_VESTING_TYPE_UNSPECIFIED
	VestingTypeLinear          = economicsecuritypb.VestingType_VESTING_TYPE_LINEAR
	VestingTypeMilestone       = economicsecuritypb.VestingType_VESTING_TYPE_MILESTONE
	VestingTypeCliffThenLinear = economicsecuritypb.VestingType_VESTING_TYPE_CLIFF_THEN_LINEAR

	// ScheduleType
	ScheduleTypeUnspecified = economicsecuritypb.ScheduleType_SCHEDULE_TYPE_UNSPECIFIED
	ScheduleTypeTeam        = economicsecuritypb.ScheduleType_SCHEDULE_TYPE_TEAM
	ScheduleTypeInvestor    = economicsecuritypb.ScheduleType_SCHEDULE_TYPE_INVESTOR
	ScheduleTypeAdvisor     = economicsecuritypb.ScheduleType_SCHEDULE_TYPE_ADVISOR
	ScheduleTypeEcosystem   = economicsecuritypb.ScheduleType_SCHEDULE_TYPE_ECOSYSTEM
	ScheduleTypeCommunity   = economicsecuritypb.ScheduleType_SCHEDULE_TYPE_COMMUNITY

	// MEVRedistributionStrategy
	MEVStrategyUnspecified            = economicsecuritypb.MEVRedistributionStrategy_MEV_STRATEGY_UNSPECIFIED
	MEVStrategyProportionalToStake    = economicsecuritypb.MEVRedistributionStrategy_MEV_STRATEGY_PROPORTIONAL_TO_STAKE
	MEVStrategyProportionalToActivity = economicsecuritypb.MEVRedistributionStrategy_MEV_STRATEGY_PROPORTIONAL_TO_ACTIVITY
	MEVStrategyEqualDistribution      = economicsecuritypb.MEVRedistributionStrategy_MEV_STRATEGY_EQUAL_DISTRIBUTION
	MEVStrategyIRWeighted             = economicsecuritypb.MEVRedistributionStrategy_MEV_STRATEGY_IR_WEIGHTED

	// InflationAlertType
	InflationAlertTypeUnspecified = economicsecuritypb.InflationAlertType_INFLATION_ALERT_TYPE_UNSPECIFIED
	InflationAlertTypeAboveTarget = economicsecuritypb.InflationAlertType_INFLATION_ALERT_TYPE_ABOVE_TARGET
	InflationAlertTypeBelowTarget = economicsecuritypb.InflationAlertType_INFLATION_ALERT_TYPE_BELOW_TARGET
	InflationAlertTypeAboveMax    = economicsecuritypb.InflationAlertType_INFLATION_ALERT_TYPE_ABOVE_MAX
	InflationAlertTypeBelowMin    = economicsecuritypb.InflationAlertType_INFLATION_ALERT_TYPE_BELOW_MIN
	InflationAlertTypeRapidChange = economicsecuritypb.InflationAlertType_INFLATION_ALERT_TYPE_RAPID_CHANGE

	// AlertSeverity
	AlertSeverityUnspecified = economicsecuritypb.AlertSeverity_ALERT_SEVERITY_UNSPECIFIED
	AlertSeverityInfo        = economicsecuritypb.AlertSeverity_ALERT_SEVERITY_INFO
	AlertSeverityWarning     = economicsecuritypb.AlertSeverity_ALERT_SEVERITY_WARNING
	AlertSeverityCritical    = economicsecuritypb.AlertSeverity_ALERT_SEVERITY_CRITICAL
	AlertSeverityEmergency   = economicsecuritypb.AlertSeverity_ALERT_SEVERITY_EMERGENCY
)
