package types

import (
	"time"

	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserConfidenceRecord represents the internal user confidence record
type UserConfidenceRecord struct {
	WalletAddress              string
	TotalScore                 uint64
	CompletedIRs               []IRCompletion
	HasAnchor                  bool
	LastUpdatedHeight          uint64
	ArenaScores                map[string]ArenaScore
	RewardAmount               uint64
	UserReward                 uint64
	NodeOperatorReward         uint64
	Status                     VerificationStatus
	LastUpdated                int64 // Unix timestamp
	VerificationAchievedHeight uint64
	VerificationAchievedAt     int64 // Unix timestamp
	AnchorInfo                 *AnchorInfo
}

// ArenaScore represents score details for a specific arena
type ArenaScore struct {
	ArenaType          string
	RewardAmount       uint64
	UserReward         uint64
	NodeOperatorReward uint64
	TotalScore         uint64
	IRCount            uint32
	FocusBonusActive   bool
}

// IRCompletion represents a single IR completion
type IRCompletion struct {
	IRID               string
	BaseScore          uint64
	FinalScore         uint64
	CompletedAt        int64 // Unix timestamp
	CompletedHeight    uint64
	AssistantAddress   string
	ProofHash          []byte
	VerifierHash       []byte
	TxHash             string
	VelocityBonus      float32
	ArenaBonus         float32
	JackpotBonus       float32
	Status             IRCompletionStatus
	Arena              string
	RewardAmount       uint64
	UserReward         uint64
	NodeOperatorReward uint64
}

// AnchorInfo contains metadata about IR-000 completion
type AnchorInfo struct {
	Completed          bool
	CompletedAt        int64 // Unix timestamp
	VerifierPluginHash []byte
	BlockHeight        uint64
	ProofHash          []byte
}

// ScoreChange represents a score change event
type ScoreChange struct {
	BlockHeight   uint64
	ScoreDelta    int64
	NewTotal      uint64
	Reason        ChangeReason
	RelatedIRID   string
	TxHash        string
	Timestamp     int64 // Unix timestamp
	PreviousScore uint64
}

// SlashRecord represents a slash record
type SlashRecord struct {
	WalletAddress  string
	SlashAmount    uint64
	Reason         SlashReason
	SlashHeight    uint64
	SlashTime      int64 // Unix timestamp
	RelatedIRID    string
	SlashTxHash    string
	AppealDeadline int64 // Unix timestamp
	Appealed       bool
	Resolved       bool
	Authority      string
	Evidence       string
}

// Enums (matching proto enums)
type VerificationStatus int32
type IRCompletionStatus int32
type ChangeReason int32
type SlashReason int32

const (
	VerificationStatusUnspecified VerificationStatus = 0
	VerificationStatusUnverified  VerificationStatus = 1
	VerificationStatusVerified    VerificationStatus = 2
	VerificationStatusSuspended   VerificationStatus = 3
	VerificationStatusRevoked     VerificationStatus = 4

	IRCompletionStatusUnspecified IRCompletionStatus = 0
	IRCompletionStatusPending     IRCompletionStatus = 1
	IRCompletionStatusVerified    IRCompletionStatus = 2
	IRCompletionStatusRejected    IRCompletionStatus = 3
	IRCompletionStatusAppealed    IRCompletionStatus = 4

	ChangeReasonUnspecified          ChangeReason = 0
	ChangeReasonIRCompletion         ChangeReason = 1
	ChangeReasonFraudSlash           ChangeReason = 2
	ChangeReasonGovernanceAdjustment ChangeReason = 3
	ChangeReasonAppealReversal       ChangeReason = 4

	SlashReasonUnspecified         SlashReason = 0
	SlashReasonFraudDetected       SlashReason = 1
	SlashReasonFalseAttestation    SlashReason = 2
	SlashReasonCollusion           SlashReason = 3
	SlashReasonDuplicateCompletion SlashReason = 4
)

// UserConfidenceRecordToProto converts internal type to proto
func UserConfidenceRecordToProto(record UserConfidenceRecord) *confidencescorepb.UserConfidenceRecord {
	completedIRs := make([]*confidencescorepb.IRCompletion, len(record.CompletedIRs))
	for i, ir := range record.CompletedIRs {
		completedIRs[i] = IRCompletionToProto(ir)
	}

	arenaScores := make(map[string]*confidencescorepb.ArenaScore)
	for arena, score := range record.ArenaScores {
		arenaScores[arena] = ArenaScoreToProto(score)
	}

	var anchorInfo *confidencescorepb.AnchorInfo
	if record.AnchorInfo != nil {
		anchorInfo = AnchorInfoToProto(*record.AnchorInfo)
	}

	return &confidencescorepb.UserConfidenceRecord{
		WalletAddress:              record.WalletAddress,
		TotalScore:                 record.TotalScore,
		CompletedIrs:               completedIRs,
		HasAnchor:                  record.HasAnchor,
		LastUpdatedHeight:          record.LastUpdatedHeight,
		ArenaScores:                arenaScores,
		Status:                     confidencescorepb.VerificationStatus(record.Status),
		LastUpdated:                timestamppb.New(unixToTime(record.LastUpdated)),
		VerificationAchievedHeight: record.VerificationAchievedHeight,
		VerificationAchievedAt:     timestamppb.New(unixToTime(record.VerificationAchievedAt)),
		AnchorInfo:                 anchorInfo,
	}
}

// UserConfidenceRecordFromProto converts proto to internal type
func UserConfidenceRecordFromProto(pb *confidencescorepb.UserConfidenceRecord) UserConfidenceRecord {
	if pb == nil {
		return UserConfidenceRecord{}
	}

	completedIRs := make([]IRCompletion, len(pb.CompletedIrs))
	for i, ir := range pb.CompletedIrs {
		completedIRs[i] = IRCompletionFromProto(ir)
	}

	arenaScores := make(map[string]ArenaScore)
	for arena, score := range pb.ArenaScores {
		arenaScores[arena] = ArenaScoreFromProto(score)
	}

	var anchorInfo *AnchorInfo
	if pb.AnchorInfo != nil {
		ai := AnchorInfoFromProto(pb.AnchorInfo)
		anchorInfo = &ai
	}

	return UserConfidenceRecord{
		WalletAddress:              pb.WalletAddress,
		TotalScore:                 pb.TotalScore,
		CompletedIRs:               completedIRs,
		HasAnchor:                  pb.HasAnchor,
		LastUpdatedHeight:          pb.LastUpdatedHeight,
		ArenaScores:                arenaScores,
		Status:                     VerificationStatus(pb.Status),
		LastUpdated:                timeToUnix(pb.LastUpdated),
		VerificationAchievedHeight: pb.VerificationAchievedHeight,
		VerificationAchievedAt:     timeToUnix(pb.VerificationAchievedAt),
		AnchorInfo:                 anchorInfo,
	}
}

// ArenaScoreToProto converts internal type to proto
func ArenaScoreToProto(score ArenaScore) *confidencescorepb.ArenaScore {
	return &confidencescorepb.ArenaScore{
		ArenaType:        score.ArenaType,
		TotalScore:       score.TotalScore,
		IrCount:          score.IRCount,
		FocusBonusActive: score.FocusBonusActive,
	}
}

// ArenaScoreFromProto converts proto to internal type
func ArenaScoreFromProto(pb *confidencescorepb.ArenaScore) ArenaScore {
	if pb == nil {
		return ArenaScore{}
	}
	return ArenaScore{
		ArenaType:        pb.ArenaType,
		TotalScore:       pb.TotalScore,
		IRCount:          pb.IrCount,
		FocusBonusActive: pb.FocusBonusActive,
	}
}

// IRCompletionToProto converts internal type to proto
func IRCompletionToProto(ir IRCompletion) *confidencescorepb.IRCompletion {
	return &confidencescorepb.IRCompletion{
		IrId:             ir.IRID,
		BaseScore:        ir.BaseScore,
		FinalScore:       ir.FinalScore,
		CompletedAt:      timestamppb.New(unixToTime(ir.CompletedAt)),
		CompletedHeight:  ir.CompletedHeight,
		AssistantAddress: ir.AssistantAddress,
		ProofHash:        ir.ProofHash,
		VerifierHash:     ir.VerifierHash,
		TxHash:           ir.TxHash,
		VelocityBonus:    ir.VelocityBonus,
		ArenaBonus:       ir.ArenaBonus,
		JackpotBonus:     ir.JackpotBonus,
		Status:           confidencescorepb.IRCompletionStatus(ir.Status),
		Arena:            ir.Arena,
	}
}

// IRCompletionFromProto converts proto to internal type
func IRCompletionFromProto(pb *confidencescorepb.IRCompletion) IRCompletion {
	if pb == nil {
		return IRCompletion{}
	}
	return IRCompletion{
		IRID:             pb.IrId,
		BaseScore:        pb.BaseScore,
		FinalScore:       pb.FinalScore,
		CompletedAt:      timeToUnix(pb.CompletedAt),
		CompletedHeight:  pb.CompletedHeight,
		AssistantAddress: pb.AssistantAddress,
		ProofHash:        pb.ProofHash,
		VerifierHash:     pb.VerifierHash,
		TxHash:           pb.TxHash,
		VelocityBonus:    pb.VelocityBonus,
		ArenaBonus:       pb.ArenaBonus,
		JackpotBonus:     pb.JackpotBonus,
		Status:           IRCompletionStatus(pb.Status),
		Arena:            pb.Arena,
	}
}

// AnchorInfoToProto converts internal type to proto
func AnchorInfoToProto(info AnchorInfo) *confidencescorepb.AnchorInfo {
	return &confidencescorepb.AnchorInfo{
		Completed:          info.Completed,
		CompletedAt:        timestamppb.New(unixToTime(info.CompletedAt)),
		VerifierPluginHash: info.VerifierPluginHash,
		BlockHeight:        info.BlockHeight,
		ProofHash:          info.ProofHash,
	}
}

// AnchorInfoFromProto converts proto to internal type
func AnchorInfoFromProto(pb *confidencescorepb.AnchorInfo) AnchorInfo {
	if pb == nil {
		return AnchorInfo{}
	}
	return AnchorInfo{
		Completed:          pb.Completed,
		CompletedAt:        timeToUnix(pb.CompletedAt),
		VerifierPluginHash: pb.VerifierPluginHash,
		BlockHeight:        pb.BlockHeight,
		ProofHash:          pb.ProofHash,
	}
}

// ScoreChangeToProto converts internal type to proto
func ScoreChangeToProto(change ScoreChange) *confidencescorepb.ScoreChange {
	return &confidencescorepb.ScoreChange{
		BlockHeight:   change.BlockHeight,
		ScoreDelta:    change.ScoreDelta,
		NewTotal:      change.NewTotal,
		Reason:        confidencescorepb.ChangeReason(change.Reason),
		RelatedIrId:   change.RelatedIRID,
		TxHash:        change.TxHash,
		Timestamp:     timestamppb.New(unixToTime(change.Timestamp)),
		PreviousScore: change.PreviousScore,
	}
}

// ScoreChangeFromProto converts proto to internal type
func ScoreChangeFromProto(pb *confidencescorepb.ScoreChange) ScoreChange {
	if pb == nil {
		return ScoreChange{}
	}
	return ScoreChange{
		BlockHeight:   pb.BlockHeight,
		ScoreDelta:    pb.ScoreDelta,
		NewTotal:      pb.NewTotal,
		Reason:        ChangeReason(pb.Reason),
		RelatedIRID:   pb.RelatedIrId,
		TxHash:        pb.TxHash,
		Timestamp:     timeToUnix(pb.Timestamp),
		PreviousScore: pb.PreviousScore,
	}
}

// SlashRecordToProto converts internal type to proto
func SlashRecordToProto(record SlashRecord) *confidencescorepb.SlashRecord {
	return &confidencescorepb.SlashRecord{
		WalletAddress:  record.WalletAddress,
		SlashAmount:    record.SlashAmount,
		Reason:         confidencescorepb.SlashReason(record.Reason),
		SlashHeight:    record.SlashHeight,
		SlashTime:      timestamppb.New(unixToTime(record.SlashTime)),
		RelatedIrId:    record.RelatedIRID,
		SlashTxHash:    record.SlashTxHash,
		AppealDeadline: timestamppb.New(unixToTime(record.AppealDeadline)),
		Appealed:       record.Appealed,
		Resolved:       record.Resolved,
		Authority:      record.Authority,
		Evidence:       record.Evidence,
	}
}

// SlashRecordFromProto converts proto to internal type
func SlashRecordFromProto(pb *confidencescorepb.SlashRecord) SlashRecord {
	if pb == nil {
		return SlashRecord{}
	}
	return SlashRecord{
		WalletAddress:  pb.WalletAddress,
		SlashAmount:    pb.SlashAmount,
		Reason:         SlashReason(pb.Reason),
		SlashHeight:    pb.SlashHeight,
		SlashTime:      timeToUnix(pb.SlashTime),
		RelatedIRID:    pb.RelatedIrId,
		SlashTxHash:    pb.SlashTxHash,
		AppealDeadline: timeToUnix(pb.AppealDeadline),
		Appealed:       pb.Appealed,
		Resolved:       pb.Resolved,
		Authority:      pb.Authority,
		Evidence:       pb.Evidence,
	}
}

// Helper functions for timestamp conversion
func timeToUnix(ts *timestamppb.Timestamp) int64 {
	if ts == nil {
		return 0
	}
	return ts.AsTime().Unix()
}

func unixToTime(unix int64) time.Time {
	if unix == 0 {
		return time.Time{}
	}
	return time.Unix(unix, 0)
}
