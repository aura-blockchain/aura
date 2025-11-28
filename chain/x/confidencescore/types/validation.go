package types

import (
	"fmt"

	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
)

// DefaultParams returns default confidence score parameters
func DefaultParams() Params {
	return Params{
		// Verification thresholds
		VerificationThreshold:  10000, // Default: 10,000 CS required for verification
		HighAssuranceThreshold: 15000, // Default: 15,000 CS for high assurance VCs
		ArenaFocusThreshold:    5000,  // Default: 5,000 CS for arena focus credentials

		// Velocity bonuses (time-based completion bonuses)
		VelocityBonusDays:        []uint64{7, 30},       // Complete in 7 days or 30 days
		VelocityBonusMultipliers: []float32{1.25, 1.10}, // Get 1.25x or 1.10x bonus

		// Arena multipliers (threshold -> multiplier)
		ArenaMultipliers: map[uint64]float32{
			3000: 1.1, // 3000+ CS in arena = 1.1x multiplier
			4000: 1.2, // 4000+ CS in arena = 1.2x multiplier
			5000: 1.5, // 5000+ CS in arena = 1.5x multiplier (arena focus unlocked)
		},

		// Slashing parameters
		SlashPercentage: 50,          // 50% max slash for fraud
		AppealDeposit:   "1000uaura", // 1000 AEQ to appeal a slash

		// Rate limiting
		MaxIrsPerDay:  10, // Max 10 IR completions per day
		MaxIrsPerHour: 3,  // Max 3 IR completions per hour

		// Jackpot probabilities
		JackpotOdds:        []uint64{100, 1000},  // 1 in 100, or 1 in 1000
		JackpotMultipliers: []float32{5.0, 25.0}, // Get 5x or 25x score bonus

		// Staleness (future feature - disabled by default)
		StalenessEnabled:       false, // Score staleness not enabled
		DegradationRatePerYear: 0,     // No degradation

		// PoI (Proof-of-Identity) Rewards (Whitepaper Section 12.0)
		PoiRewardsEnabled:      true, // Enable PoI rewards
		UserRewardSplitPercent: 70,   // 70% to user, 30% to node operator
		VelocityBonusEnabled:   true, // Enable velocity bonus tier
	}
}

// ValidateParams validates confidence score parameters
func ValidateParams(p *Params) error {
	if p == nil {
		return fmt.Errorf("params cannot be nil")
	}
	// Validate thresholds
	if p.VerificationThreshold == 0 {
		return fmt.Errorf("verification_threshold must be positive")
	}
	if p.HighAssuranceThreshold < p.VerificationThreshold {
		return fmt.Errorf("high_assurance_threshold (%d) must be >= verification_threshold (%d)",
			p.HighAssuranceThreshold, p.VerificationThreshold)
	}
	if p.ArenaFocusThreshold == 0 {
		return fmt.Errorf("arena_focus_threshold must be positive")
	}

	// Validate velocity bonus configuration
	if len(p.VelocityBonusDays) != len(p.VelocityBonusMultipliers) {
		return fmt.Errorf("velocity_bonus_days and velocity_bonus_multipliers must have same length")
	}
	for i, mult := range p.VelocityBonusMultipliers {
		if mult < 1.0 {
			return fmt.Errorf("velocity_bonus_multipliers[%d] must be >= 1.0, got %f", i, mult)
		}
	}

	// Validate arena multipliers
	for threshold, mult := range p.ArenaMultipliers {
		if threshold == 0 {
			return fmt.Errorf("arena_multipliers threshold must be positive")
		}
		if mult < 1.0 {
			return fmt.Errorf("arena_multipliers[%d] must be >= 1.0, got %f", threshold, mult)
		}
	}

	// Validate slashing
	if p.SlashPercentage > 100 {
		return fmt.Errorf("slash_percentage must be <= 100, got %d", p.SlashPercentage)
	}

	// Validate rate limits
	if p.MaxIrsPerDay == 0 {
		return fmt.Errorf("max_irs_per_day must be positive")
	}
	if p.MaxIrsPerHour == 0 {
		return fmt.Errorf("max_irs_per_hour must be positive")
	}
	if p.MaxIrsPerHour > p.MaxIrsPerDay {
		return fmt.Errorf("max_irs_per_hour (%d) must be <= max_irs_per_day (%d)",
			p.MaxIrsPerHour, p.MaxIrsPerDay)
	}

	// Validate jackpot configuration
	if len(p.JackpotOdds) != len(p.JackpotMultipliers) {
		return fmt.Errorf("jackpot_odds and jackpot_multipliers must have same length")
	}
	for i, mult := range p.JackpotMultipliers {
		if mult < 1.0 {
			return fmt.Errorf("jackpot_multipliers[%d] must be >= 1.0, got %f", i, mult)
		}
	}

	// Validate PoI reward split
	if p.UserRewardSplitPercent > 100 {
		return fmt.Errorf("user_reward_split_percent must be <= 100, got %d", p.UserRewardSplitPercent)
	}

	return nil
}


// Constant aliases for ChangeReason enum (defined here as they're module-specific)
const (
	ChangeReasonIRCompletion         = ChangeReason_CHANGE_REASON_IR_COMPLETION
	ChangeReasonFraudSlash           = ChangeReason_CHANGE_REASON_FRAUD_SLASH
	ChangeReasonGovernanceAdjustment = ChangeReason_CHANGE_REASON_GOVERNANCE_ADJUSTMENT
	ChangeReasonAppealReversal       = ChangeReason_CHANGE_REASON_APPEAL_REVERSAL
)

// Helper functions for proto conversions from types to confidencescorepb
func ArenaScoreToProto(score *ArenaScore) *confidencescorepb.ArenaScore {
	if score == nil {
		return nil
	}
	return &confidencescorepb.ArenaScore{
		ArenaType:        score.ArenaType,
		TotalScore:       score.TotalScore,
		IrCount:          score.IrCount,
		FocusBonusActive: score.FocusBonusActive,
	}
}

func AnchorInfoToProto(info *AnchorInfo) *confidencescorepb.AnchorInfo {
	if info == nil {
		return nil
	}
	return &confidencescorepb.AnchorInfo{
		Completed:          info.Completed,
		CompletedAt:        info.CompletedAt,
		VerifierPluginHash: info.VerifierPluginHash,
		BlockHeight:        info.BlockHeight,
		ProofHash:          info.ProofHash,
	}
}

func IRCompletionToProto(completion *IRCompletion) *confidencescorepb.IRCompletion {
	if completion == nil {
		return nil
	}
	return &confidencescorepb.IRCompletion{
		IrId:             completion.IrId,
		BaseScore:        completion.BaseScore,
		FinalScore:       completion.FinalScore,
		CompletedAt:      completion.CompletedAt,
		CompletedHeight:  completion.CompletedHeight,
		AssistantAddress: completion.AssistantAddress,
		ProofHash:        completion.ProofHash,
		VerifierHash:     completion.VerifierHash,
		TxHash:           completion.TxHash,
		VelocityBonus:    completion.VelocityBonus,
		ArenaBonus:       completion.ArenaBonus,
		JackpotBonus:     completion.JackpotBonus,
		Status:           confidencescorepb.IRCompletionStatus(completion.Status),
		Arena:            completion.Arena,
	}
}

func ScoreChangeToProto(change *ScoreChange) *confidencescorepb.ScoreChange {
	if change == nil {
		return nil
	}
	return &confidencescorepb.ScoreChange{
		BlockHeight:   change.BlockHeight,
		ScoreDelta:    change.ScoreDelta,
		NewTotal:      change.NewTotal,
		Reason:        confidencescorepb.ChangeReason(change.Reason),
		RelatedIrId:   change.RelatedIrId,
		TxHash:        change.TxHash,
		Timestamp:     change.Timestamp,
		PreviousScore: change.PreviousScore,
	}
}

func SlashRecordToProto(record *SlashRecord) *confidencescorepb.SlashRecord {
	if record == nil {
		return nil
	}
	return &confidencescorepb.SlashRecord{
		WalletAddress:  record.WalletAddress,
		SlashAmount:    record.SlashAmount,
		Reason:         confidencescorepb.SlashReason(record.Reason),
		SlashHeight:    record.SlashHeight,
		SlashTime:      record.SlashTime,
		RelatedIrId:    record.RelatedIrId,
		SlashTxHash:    record.SlashTxHash,
		AppealDeadline: record.AppealDeadline,
		Appealed:       record.Appealed,
		Resolved:       record.Resolved,
		Authority:      record.Authority,
		Evidence:       record.Evidence,
	}
}

func ParamsToProto(params Params) *confidencescorepb.Params {
	return &confidencescorepb.Params{
		VerificationThreshold:    params.VerificationThreshold,
		HighAssuranceThreshold:   params.HighAssuranceThreshold,
		ArenaFocusThreshold:      params.ArenaFocusThreshold,
		VelocityBonusDays:        params.VelocityBonusDays,
		VelocityBonusMultipliers: params.VelocityBonusMultipliers,
		ArenaMultipliers:         params.ArenaMultipliers,
		SlashPercentage:          params.SlashPercentage,
		AppealDeposit:            params.AppealDeposit,
		MaxIrsPerDay:             params.MaxIrsPerDay,
		MaxIrsPerHour:            params.MaxIrsPerHour,
		JackpotOdds:              params.JackpotOdds,
		JackpotMultipliers:       params.JackpotMultipliers,
		StalenessEnabled:         params.StalenessEnabled,
		DegradationRatePerYear:   params.DegradationRatePerYear,
		PoiRewardsEnabled:        params.PoiRewardsEnabled,
		UserRewardSplitPercent:   params.UserRewardSplitPercent,
		VelocityBonusEnabled:     params.VelocityBonusEnabled,
	}
}
