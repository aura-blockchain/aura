package types

import pb "github.com/aequitas/aura/proto/aura/economicsecurity/v1beta1"

// Re-export all proto types
type (
	// Enums
	VestingType               = pb.VestingType
	ScheduleType              = pb.ScheduleType
	MEVRedistributionStrategy = pb.MEVRedistributionStrategy
	InflationAlertType        = pb.InflationAlertType
	AlertSeverity             = pb.AlertSeverity

	// Core types
	VestingSchedule        = pb.VestingSchedule
	VoteLock               = pb.VoteLock
	TokenomicsConfig       = pb.TokenomicsConfig
	DynamicFeeConfig       = pb.DynamicFeeConfig
	TransferTaxConfig      = pb.TransferTaxConfig
	LiquidityMiningConfig  = pb.LiquidityMiningConfig
	MEVConfig              = pb.MEVConfig
	GovernanceConfig       = pb.GovernanceConfig
	WhaleProtection        = pb.WhaleProtection
	LargeTxRecord          = pb.LargeTxRecord
	InflationAlert         = pb.InflationAlert
	TreasuryMultisig       = pb.TreasuryMultisig
	PendingTreasuryTx      = pb.PendingTreasuryTx
	Params                 = pb.Params

	// Message types
	MsgCreateVestingSchedule         = pb.MsgCreateVestingSchedule
	MsgCreateVestingScheduleResponse = pb.MsgCreateVestingScheduleResponse
	MsgRevokeVestingSchedule         = pb.MsgRevokeVestingSchedule
	MsgRevokeVestingScheduleResponse = pb.MsgRevokeVestingScheduleResponse
	MsgReleaseVestedTokens           = pb.MsgReleaseVestedTokens
	MsgReleaseVestedTokensResponse   = pb.MsgReleaseVestedTokensResponse
	MsgLockVotingTokens              = pb.MsgLockVotingTokens
	MsgLockVotingTokensResponse      = pb.MsgLockVotingTokensResponse
	MsgUnlockVotingTokens            = pb.MsgUnlockVotingTokens
	MsgUnlockVotingTokensResponse    = pb.MsgUnlockVotingTokensResponse
	MsgUpdateParams                  = pb.MsgUpdateParams
	MsgUpdateParamsResponse          = pb.MsgUpdateParamsResponse
	MsgAdjustInflationRate           = pb.MsgAdjustInflationRate
	MsgAdjustInflationRateResponse   = pb.MsgAdjustInflationRateResponse
	MsgProposeTreasurySpend          = pb.MsgProposeTreasurySpend
	MsgProposeTreasurySpendResponse  = pb.MsgProposeTreasurySpendResponse
	MsgSignTreasurySpend             = pb.MsgSignTreasurySpend
	MsgSignTreasurySpendResponse     = pb.MsgSignTreasurySpendResponse
	MsgExecuteTreasurySpend          = pb.MsgExecuteTreasurySpend
	MsgExecuteTreasurySpendResponse  = pb.MsgExecuteTreasurySpendResponse

	// Query types
	QueryVestingScheduleRequest                   = pb.QueryVestingScheduleRequest
	QueryVestingScheduleResponse                  = pb.QueryVestingScheduleResponse
	QueryVestingSchedulesByBeneficiaryRequest     = pb.QueryVestingSchedulesByBeneficiaryRequest
	QueryVestingSchedulesByBeneficiaryResponse    = pb.QueryVestingSchedulesByBeneficiaryResponse
	QueryVoteLockRequest                          = pb.QueryVoteLockRequest
	QueryVoteLockResponse                         = pb.QueryVoteLockResponse
	QueryVoteLocksByOwnerRequest                  = pb.QueryVoteLocksByOwnerRequest
	QueryVoteLocksByOwnerResponse                 = pb.QueryVoteLocksByOwnerResponse
	QueryVotingPowerRequest                       = pb.QueryVotingPowerRequest
	QueryVotingPowerResponse                      = pb.QueryVotingPowerResponse
	QueryTokenomicsStatsRequest                   = pb.QueryTokenomicsStatsRequest
	QueryTokenomicsStatsResponse                  = pb.QueryTokenomicsStatsResponse
	QueryLiquidityMiningStatsRequest              = pb.QueryLiquidityMiningStatsRequest
	QueryLiquidityMiningStatsResponse             = pb.QueryLiquidityMiningStatsResponse
	QueryMEVStatsRequest                          = pb.QueryMEVStatsRequest
	QueryMEVStatsResponse                         = pb.QueryMEVStatsResponse
	QueryUserMEVBalanceRequest                    = pb.QueryUserMEVBalanceRequest
	QueryUserMEVBalanceResponse                   = pb.QueryUserMEVBalanceResponse
	QueryInflationMetricsRequest                  = pb.QueryInflationMetricsRequest
	QueryInflationMetricsResponse                 = pb.QueryInflationMetricsResponse
	QueryInflationAlertsRequest                   = pb.QueryInflationAlertsRequest
	QueryInflationAlertsResponse                  = pb.QueryInflationAlertsResponse
	QueryPendingTreasuryTxRequest                 = pb.QueryPendingTreasuryTxRequest
	QueryPendingTreasuryTxResponse                = pb.QueryPendingTreasuryTxResponse
	QueryPendingTreasuryTxsRequest                = pb.QueryPendingTreasuryTxsRequest
	QueryPendingTreasuryTxsResponse               = pb.QueryPendingTreasuryTxsResponse
	QueryParamsRequest                            = pb.QueryParamsRequest
	QueryParamsResponse                           = pb.QueryParamsResponse

	// Genesis types
	GenesisState = pb.GenesisState
)

// Re-export enum values for VestingType
const (
	VestingType_VESTING_TYPE_UNSPECIFIED       = pb.VestingType_VESTING_TYPE_UNSPECIFIED
	VestingType_VESTING_TYPE_LINEAR            = pb.VestingType_VESTING_TYPE_LINEAR
	VestingType_VESTING_TYPE_MILESTONE         = pb.VestingType_VESTING_TYPE_MILESTONE
	VestingType_VESTING_TYPE_CLIFF_THEN_LINEAR = pb.VestingType_VESTING_TYPE_CLIFF_THEN_LINEAR
)

// Re-export enum values for ScheduleType
const (
	ScheduleType_SCHEDULE_TYPE_UNSPECIFIED = pb.ScheduleType_SCHEDULE_TYPE_UNSPECIFIED
	ScheduleType_SCHEDULE_TYPE_TEAM        = pb.ScheduleType_SCHEDULE_TYPE_TEAM
	ScheduleType_SCHEDULE_TYPE_INVESTOR    = pb.ScheduleType_SCHEDULE_TYPE_INVESTOR
	ScheduleType_SCHEDULE_TYPE_ADVISOR     = pb.ScheduleType_SCHEDULE_TYPE_ADVISOR
	ScheduleType_SCHEDULE_TYPE_ECOSYSTEM   = pb.ScheduleType_SCHEDULE_TYPE_ECOSYSTEM
	ScheduleType_SCHEDULE_TYPE_COMMUNITY   = pb.ScheduleType_SCHEDULE_TYPE_COMMUNITY
)

// Re-export enum values for MEVRedistributionStrategy
const (
	MEVRedistributionStrategy_MEV_STRATEGY_UNSPECIFIED              = pb.MEVRedistributionStrategy_MEV_STRATEGY_UNSPECIFIED
	MEVRedistributionStrategy_MEV_STRATEGY_PROPORTIONAL_TO_STAKE    = pb.MEVRedistributionStrategy_MEV_STRATEGY_PROPORTIONAL_TO_STAKE
	MEVRedistributionStrategy_MEV_STRATEGY_PROPORTIONAL_TO_ACTIVITY = pb.MEVRedistributionStrategy_MEV_STRATEGY_PROPORTIONAL_TO_ACTIVITY
	MEVRedistributionStrategy_MEV_STRATEGY_EQUAL_DISTRIBUTION       = pb.MEVRedistributionStrategy_MEV_STRATEGY_EQUAL_DISTRIBUTION
	MEVRedistributionStrategy_MEV_STRATEGY_IR_WEIGHTED              = pb.MEVRedistributionStrategy_MEV_STRATEGY_IR_WEIGHTED
)

// Re-export enum values for InflationAlertType
const (
	InflationAlertType_INFLATION_ALERT_TYPE_UNSPECIFIED  = pb.InflationAlertType_INFLATION_ALERT_TYPE_UNSPECIFIED
	InflationAlertType_INFLATION_ALERT_TYPE_ABOVE_TARGET = pb.InflationAlertType_INFLATION_ALERT_TYPE_ABOVE_TARGET
	InflationAlertType_INFLATION_ALERT_TYPE_BELOW_TARGET = pb.InflationAlertType_INFLATION_ALERT_TYPE_BELOW_TARGET
	InflationAlertType_INFLATION_ALERT_TYPE_ABOVE_MAX    = pb.InflationAlertType_INFLATION_ALERT_TYPE_ABOVE_MAX
	InflationAlertType_INFLATION_ALERT_TYPE_BELOW_MIN    = pb.InflationAlertType_INFLATION_ALERT_TYPE_BELOW_MIN
	InflationAlertType_INFLATION_ALERT_TYPE_RAPID_CHANGE = pb.InflationAlertType_INFLATION_ALERT_TYPE_RAPID_CHANGE
)

// Re-export enum values for AlertSeverity
const (
	AlertSeverity_ALERT_SEVERITY_UNSPECIFIED = pb.AlertSeverity_ALERT_SEVERITY_UNSPECIFIED
	AlertSeverity_ALERT_SEVERITY_INFO        = pb.AlertSeverity_ALERT_SEVERITY_INFO
	AlertSeverity_ALERT_SEVERITY_WARNING     = pb.AlertSeverity_ALERT_SEVERITY_WARNING
	AlertSeverity_ALERT_SEVERITY_CRITICAL    = pb.AlertSeverity_ALERT_SEVERITY_CRITICAL
	AlertSeverity_ALERT_SEVERITY_EMERGENCY   = pb.AlertSeverity_ALERT_SEVERITY_EMERGENCY
)

// Additional type aliases for circuit breaker and attack detection features
// These types are used for advanced economic security monitoring

// CircuitBreakerType represents different types of circuit breakers
type CircuitBreakerType int32

const (
	CircuitBreakerTypeUnspecified      CircuitBreakerType = 0
	CircuitBreakerTypePriceVolatility  CircuitBreakerType = 1
	CircuitBreakerTypeLargeTransaction CircuitBreakerType = 2
	CircuitBreakerTypeSupplyChange     CircuitBreakerType = 3
	CircuitBreakerTypeLiquidityCrisis  CircuitBreakerType = 4
	CircuitBreakerTypeGasSpike         CircuitBreakerType = 5
)

func (t CircuitBreakerType) String() string {
	switch t {
	case CircuitBreakerTypePriceVolatility:
		return "PRICE_VOLATILITY"
	case CircuitBreakerTypeLargeTransaction:
		return "LARGE_TRANSACTION"
	case CircuitBreakerTypeSupplyChange:
		return "SUPPLY_CHANGE"
	case CircuitBreakerTypeLiquidityCrisis:
		return "LIQUIDITY_CRISIS"
	case CircuitBreakerTypeGasSpike:
		return "GAS_SPIKE"
	default:
		return "UNSPECIFIED"
	}
}

// CircuitBreakerEvent represents a circuit breaker trigger event
type CircuitBreakerEvent struct {
	BreakerId     string             `json:"breaker_id"`
	BreakerType   CircuitBreakerType `json:"breaker_type"`
	Severity      AlertSeverity      `json:"severity"`
	TriggeredAt   interface{}        `json:"triggered_at"` // timestamppb.Timestamp
	Message       string             `json:"message"`
	CurrentValue  string             `json:"current_value"`
	Threshold     string             `json:"threshold"`
	AutoMitigated bool               `json:"auto_mitigated"`
	Active        bool               `json:"active"`
}

// AttackType represents different types of economic attacks
type AttackType int32

const (
	AttackTypeUnspecified  AttackType = 0
	AttackTypePumpAndDump  AttackType = 1
	AttackTypeFlashLoan    AttackType = 2
	AttackTypeSybil        AttackType = 3
	AttackTypeWashTrading  AttackType = 4
	AttackTypeFrontRunning AttackType = 5
)

func (t AttackType) String() string {
	switch t {
	case AttackTypePumpAndDump:
		return "PUMP_AND_DUMP"
	case AttackTypeFlashLoan:
		return "FLASH_LOAN"
	case AttackTypeSybil:
		return "SYBIL"
	case AttackTypeWashTrading:
		return "WASH_TRADING"
	case AttackTypeFrontRunning:
		return "FRONT_RUNNING"
	default:
		return "UNSPECIFIED"
	}
}

// AttackAlert represents an economic attack detection alert
type AttackAlert struct {
	AlertId          string        `json:"alert_id"`
	AttackType       AttackType    `json:"attack_type"`
	Severity         AlertSeverity `json:"severity"`
	Message          string        `json:"message"`
	DetectedAt       interface{}   `json:"detected_at"` // timestamppb.Timestamp
	SuspectAddress   string        `json:"suspect_address"`
	EvidenceCount    uint64        `json:"evidence_count"`
	AutoMitigated    bool          `json:"auto_mitigated"`
	MitigationAction string        `json:"mitigation_action"`
}

// CircuitBreakerConfig represents circuit breaker configuration in params
type CircuitBreakerConfig struct {
	PriceVolatilityEnabled  bool                   `json:"price_volatility_enabled"`
	LargeTransactionEnabled bool                   `json:"large_transaction_enabled"`
	SupplyChangeEnabled     bool                   `json:"supply_change_enabled"`
	LiquidityCrisisEnabled  bool                   `json:"liquidity_crisis_enabled"`
	GasSpikeEnabled         bool                   `json:"gas_spike_enabled"`
	TriggeredEvents         []*CircuitBreakerEvent `json:"triggered_events"`
	TotalTriggered          uint64                 `json:"total_triggered"`
}

// AttackDetectionConfig represents attack detection configuration in params
type AttackDetectionConfig struct {
	DetectedAttacks      []*AttackAlert `json:"detected_attacks"`
	TotalAttacksDetected uint64         `json:"total_attacks_detected"`
}
