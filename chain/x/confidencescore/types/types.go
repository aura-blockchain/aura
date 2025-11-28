package types

import (
	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
)

// Type aliases for proto-generated types
// This allows the rest of the codebase to import from chain/x/confidencescore/types
// while the actual types are defined in the proto module

type (
	// Core types
	Params                  = confidencescorepb.Params
	UserConfidenceRecord    = confidencescorepb.UserConfidenceRecord
	AnchorInfo              = confidencescorepb.AnchorInfo
	ArenaScore              = confidencescorepb.ArenaScore
	IRCompletion            = confidencescorepb.IRCompletion
	ArenaBreakdown          = confidencescorepb.ArenaBreakdown
	ConfidenceHistory       = confidencescorepb.ConfidenceHistory
	ScoreChange             = confidencescorepb.ScoreChange
	SlashRecord             = confidencescorepb.SlashRecord
	GenesisState            = confidencescorepb.GenesisState

	// Enums
	IRCompletionStatus  = confidencescorepb.IRCompletionStatus
	SlashReason         = confidencescorepb.SlashReason
	VerificationStatus  = confidencescorepb.VerificationStatus
	ChangeReason        = confidencescorepb.ChangeReason

	// Messages
	MsgRecordIRCompletion         = confidencescorepb.MsgRecordIRCompletion
	MsgRecordIRCompletionResponse = confidencescorepb.MsgRecordIRCompletionResponse
	MsgRecalculateScore           = confidencescorepb.MsgRecalculateScore
	MsgRecalculateScoreResponse   = confidencescorepb.MsgRecalculateScoreResponse
	MsgSlashScore                 = confidencescorepb.MsgSlashScore
	MsgSlashScoreResponse         = confidencescorepb.MsgSlashScoreResponse
	MsgAppealSlash                = confidencescorepb.MsgAppealSlash
	MsgAppealSlashResponse        = confidencescorepb.MsgAppealSlashResponse
	MsgResolveAppeal              = confidencescorepb.MsgResolveAppeal
	MsgResolveAppealResponse      = confidencescorepb.MsgResolveAppealResponse

	// Queries
	QueryUserScoreRequest            = confidencescorepb.QueryUserScoreRequest
	QueryUserScoreResponse           = confidencescorepb.QueryUserScoreResponse
	QueryUserCompletionsRequest      = confidencescorepb.QueryUserCompletionsRequest
	QueryUserCompletionsResponse     = confidencescorepb.QueryUserCompletionsResponse
	QueryScoreHistoryRequest         = confidencescorepb.QueryScoreHistoryRequest
	QueryScoreHistoryResponse        = confidencescorepb.QueryScoreHistoryResponse
	QueryThresholdsRequest           = confidencescorepb.QueryThresholdsRequest
	QueryThresholdsResponse          = confidencescorepb.QueryThresholdsResponse
	QueryVerifiedUsersRequest        = confidencescorepb.QueryVerifiedUsersRequest
	QueryVerifiedUsersResponse       = confidencescorepb.QueryVerifiedUsersResponse
	QueryArenaBreakdownRequest       = confidencescorepb.QueryArenaBreakdownRequest
	QueryArenaBreakdownResponse      = confidencescorepb.QueryArenaBreakdownResponse
	QuerySlashRecordRequest          = confidencescorepb.QuerySlashRecordRequest
	QuerySlashRecordResponse         = confidencescorepb.QuerySlashRecordResponse
	QueryParamsRequest               = confidencescorepb.QueryParamsRequest
	QueryParamsResponse              = confidencescorepb.QueryParamsResponse
	QueryIRCompletionRequest         = confidencescorepb.QueryIRCompletionRequest
	QueryIRCompletionResponse        = confidencescorepb.QueryIRCompletionResponse

	// Events
	EventIRCompleted           = confidencescorepb.EventIRCompleted
	EventVerificationAchieved  = confidencescorepb.EventVerificationAchieved
	EventArenaFocusAchieved    = confidencescorepb.EventArenaFocusAchieved
	EventScoreSlashed          = confidencescorepb.EventScoreSlashed
	EventJackpotTriggered      = confidencescorepb.EventJackpotTriggered
	EventAppealFiled           = confidencescorepb.EventAppealFiled
	EventAppealResolved        = confidencescorepb.EventAppealResolved
)

// Enum value constants
const (
	// IRCompletionStatus
	IRCompletionStatus_IR_COMPLETION_STATUS_UNSPECIFIED = confidencescorepb.IRCompletionStatus_IR_COMPLETION_STATUS_UNSPECIFIED
	IRCompletionStatus_IR_COMPLETION_STATUS_PENDING     = confidencescorepb.IRCompletionStatus_IR_COMPLETION_STATUS_PENDING
	IRCompletionStatus_IR_COMPLETION_STATUS_VERIFIED    = confidencescorepb.IRCompletionStatus_IR_COMPLETION_STATUS_VERIFIED
	IRCompletionStatus_IR_COMPLETION_STATUS_REJECTED    = confidencescorepb.IRCompletionStatus_IR_COMPLETION_STATUS_REJECTED
	IRCompletionStatus_IR_COMPLETION_STATUS_APPEALED    = confidencescorepb.IRCompletionStatus_IR_COMPLETION_STATUS_APPEALED

	// SlashReason
	SlashReason_SLASH_REASON_UNSPECIFIED         = confidencescorepb.SlashReason_SLASH_REASON_UNSPECIFIED
	SlashReason_SLASH_REASON_FRAUD_DETECTED      = confidencescorepb.SlashReason_SLASH_REASON_FRAUD_DETECTED
	SlashReason_SLASH_REASON_FALSE_ATTESTATION   = confidencescorepb.SlashReason_SLASH_REASON_FALSE_ATTESTATION
	SlashReason_SLASH_REASON_COLLUSION           = confidencescorepb.SlashReason_SLASH_REASON_COLLUSION
	SlashReason_SLASH_REASON_DUPLICATE_COMPLETION = confidencescorepb.SlashReason_SLASH_REASON_DUPLICATE_COMPLETION

	// VerificationStatus
	VerificationStatus_VERIFICATION_STATUS_UNSPECIFIED = confidencescorepb.VerificationStatus_VERIFICATION_STATUS_UNSPECIFIED
	VerificationStatus_VERIFICATION_STATUS_UNVERIFIED  = confidencescorepb.VerificationStatus_VERIFICATION_STATUS_UNVERIFIED
	VerificationStatus_VERIFICATION_STATUS_VERIFIED    = confidencescorepb.VerificationStatus_VERIFICATION_STATUS_VERIFIED
	VerificationStatus_VERIFICATION_STATUS_SUSPENDED   = confidencescorepb.VerificationStatus_VERIFICATION_STATUS_SUSPENDED
	VerificationStatus_VERIFICATION_STATUS_REVOKED     = confidencescorepb.VerificationStatus_VERIFICATION_STATUS_REVOKED

	// ChangeReason
	ChangeReason_CHANGE_REASON_UNSPECIFIED          = confidencescorepb.ChangeReason_CHANGE_REASON_UNSPECIFIED
	ChangeReason_CHANGE_REASON_IR_COMPLETION        = confidencescorepb.ChangeReason_CHANGE_REASON_IR_COMPLETION
	ChangeReason_CHANGE_REASON_FRAUD_SLASH          = confidencescorepb.ChangeReason_CHANGE_REASON_FRAUD_SLASH
	ChangeReason_CHANGE_REASON_GOVERNANCE_ADJUSTMENT = confidencescorepb.ChangeReason_CHANGE_REASON_GOVERNANCE_ADJUSTMENT
	ChangeReason_CHANGE_REASON_APPEAL_REVERSAL      = confidencescorepb.ChangeReason_CHANGE_REASON_APPEAL_REVERSAL
)
