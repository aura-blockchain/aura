package types

import (
	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
)

// Re-export proto types for convenience

// Params
type (
	Params              = economicspb.Params
	FeeParams           = economicspb.FeeParams
	VestingParams       = economicspb.VestingParams
	TreasuryParams      = economicspb.TreasuryParams
	GovernanceParams    = economicspb.GovernanceParams
	MEVParams           = economicspb.MEVParams
	WhaleProtectionParams = economicspb.WhaleProtectionParams
	LiquidityMiningParams = economicspb.LiquidityMiningParams
	TokenomicsParams    = economicspb.TokenomicsParams
)

// Vesting types
type (
	VestingSchedule = economicspb.VestingSchedule
	VestingType     = economicspb.VestingType
	ScheduleType    = economicspb.ScheduleType
	VoteLock        = economicspb.VoteLock
)

// Vesting type constants
const (
	VestingTypeUnspecified     = economicspb.VestingType_VESTING_TYPE_UNSPECIFIED
	VestingTypeLinear          = economicspb.VestingType_VESTING_TYPE_LINEAR
	VestingTypeCliff           = economicspb.VestingType_VESTING_TYPE_CLIFF
	VestingTypeGraded          = economicspb.VestingType_VESTING_TYPE_GRADED
	VestingTypeMilestone       = economicspb.VestingType_VESTING_TYPE_MILESTONE
	VestingTypeCliffThenLinear = VestingTypeCliff // Backwards compatibility alias
)

// Schedule type constants
const (
	ScheduleTypeUnspecified = economicspb.ScheduleType_SCHEDULE_TYPE_UNSPECIFIED
	ScheduleTypeTeam        = economicspb.ScheduleType_SCHEDULE_TYPE_TEAM
	ScheduleTypeInvestor    = economicspb.ScheduleType_SCHEDULE_TYPE_INVESTOR
	ScheduleTypeAdvisor     = economicspb.ScheduleType_SCHEDULE_TYPE_ADVISOR
	ScheduleTypeEcosystem   = economicspb.ScheduleType_SCHEDULE_TYPE_ECOSYSTEM
	ScheduleTypeCommunity   = economicspb.ScheduleType_SCHEDULE_TYPE_COMMUNITY
)

// Governance types
type (
	Proposal         = economicspb.Proposal
	ProposalStatus   = economicspb.ProposalStatus
	ProposalCategory = economicspb.ProposalCategory
	Vote             = economicspb.Vote
	VoteOption       = economicspb.VoteOption
	TallyResult      = economicspb.TallyResult
	Deposit          = economicspb.Deposit
	VoteDelegation   = economicspb.VoteDelegation
)

// Proposal status constants
const (
	ProposalStatusUnspecified       = economicspb.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED
	ProposalStatusDepositPeriod     = economicspb.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD
	ProposalStatusVotingPeriod      = economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	ProposalStatusPassed            = economicspb.ProposalStatus_PROPOSAL_STATUS_PASSED
	ProposalStatusRejected          = economicspb.ProposalStatus_PROPOSAL_STATUS_REJECTED
	ProposalStatusFailed            = economicspb.ProposalStatus_PROPOSAL_STATUS_FAILED
	ProposalStatusVetoed            = economicspb.ProposalStatus_PROPOSAL_STATUS_VETOED
	ProposalStatusExecutionDelay    = economicspb.ProposalStatus_PROPOSAL_STATUS_EXECUTION_DELAY
	ProposalStatusReadyForExecution = ProposalStatusExecutionDelay // Backwards compatibility
	ProposalStatusExecuted          = economicspb.ProposalStatus_PROPOSAL_STATUS_EXECUTED
)

// Proposal category constants
const (
	ProposalCategoryUnspecified     = economicspb.ProposalCategory_PROPOSAL_CATEGORY_UNSPECIFIED
	ProposalCategoryText            = economicspb.ProposalCategory_PROPOSAL_CATEGORY_TEXT
	ProposalCategoryParameterChange = economicspb.ProposalCategory_PROPOSAL_CATEGORY_PARAMETER_CHANGE
	ProposalCategorySoftwareUpgrade = economicspb.ProposalCategory_PROPOSAL_CATEGORY_SOFTWARE_UPGRADE
	ProposalCategorySpending        = economicspb.ProposalCategory_PROPOSAL_CATEGORY_SPENDING
	ProposalCategoryEmergency       = economicspb.ProposalCategory_PROPOSAL_CATEGORY_EMERGENCY
	ProposalCategoryConstitution    = economicspb.ProposalCategory_PROPOSAL_CATEGORY_CONSTITUTION
)

// Vote option constants
const (
	VoteOptionUnspecified = economicspb.VoteOption_VOTE_OPTION_UNSPECIFIED
	VoteOptionYes         = economicspb.VoteOption_VOTE_OPTION_YES
	VoteOptionAbstain     = economicspb.VoteOption_VOTE_OPTION_ABSTAIN
	VoteOptionNo          = economicspb.VoteOption_VOTE_OPTION_NO
	VoteOptionNoWithVeto  = economicspb.VoteOption_VOTE_OPTION_NO_WITH_VETO
)

// Treasury types
type (
	PendingTreasuryTx = economicspb.PendingTreasuryTx
)

// Monitoring types
type (
	InflationMetrics      = economicspb.InflationMetrics
	MEVStats              = economicspb.MEVStats
	LiquidityMiningStats  = economicspb.LiquidityMiningStats
)

// MEV redistribution strategy
type MEVRedistributionStrategy = economicspb.MEVRedistributionStrategy

// MEV strategy constants
const (
	MEVStrategyUnspecified            = economicspb.MEVRedistributionStrategy_MEV_STRATEGY_UNSPECIFIED
	MEVStrategyProportionalToStake    = economicspb.MEVRedistributionStrategy_MEV_STRATEGY_PROPORTIONAL_TO_STAKE
	MEVStrategyProportionalToActivity = economicspb.MEVRedistributionStrategy_MEV_STRATEGY_PROPORTIONAL_TO_ACTIVITY
	MEVStrategyEqualDistribution      = economicspb.MEVRedistributionStrategy_MEV_STRATEGY_EQUAL_DISTRIBUTION
	MEVStrategyIRWeighted             = economicspb.MEVRedistributionStrategy_MEV_STRATEGY_IR_WEIGHTED
)

// Genesis type
type GenesisState = economicspb.GenesisState

// Helper types (still needed for indexing)
type StringList struct {
	Values []string `json:"values"`
}
