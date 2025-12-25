// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	"github.com/aequitas/aura/chain/x/confidencescore/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================
// SCORE BOOSTING MECHANISMS
// Implements multipliers and bonuses for special achievements
// ============================

// BoostType defines different types of score boosts
type BoostType string

const (
	BoostTypeFirstCompletion  BoostType = "first_completion"
	BoostTypeStreak           BoostType = "completion_streak"
	BoostTypeArenaSpecialist  BoostType = "arena_specialist"
	BoostTypeEarlyAdopter     BoostType = "early_adopter"
	BoostTypeReferral         BoostType = "referral"
	BoostTypeVelocity         BoostType = "velocity_bonus"
	BoostTypeSocialImpact     BoostType = "social_impact"
	BoostTypeQualityValidator BoostType = "quality_validator"
)

// BoostMultiplier defines a score boost configuration
type BoostMultiplier struct {
	Type         BoostType
	Multiplier   math.LegacyDec // 1.0 = no boost, 1.5 = 50% boost
	Description  string
	Enabled      bool
	Requirements map[string]interface{}
}

// GetAvailableBoosts returns all configured boost multipliers
func (k *Keeper) GetAvailableBoosts(ctx sdk.Context) []BoostMultiplier {
	params, _ := k.GetParams(ctx)

	return []BoostMultiplier{
		{
			Type:        BoostTypeFirstCompletion,
			Multiplier:  math.LegacyNewDecWithPrec(12, 1), // 1.2x (20% bonus)
			Description: "First IR completion bonus",
			Enabled:     params.VelocityBonusEnabled, // Use velocity flag for all boosts
		},
		{
			Type:        BoostTypeStreak,
			Multiplier:  math.LegacyNewDecWithPrec(15, 1), // 1.5x (50% bonus)
			Description: "Completion streak bonus (7+ days)",
			Enabled:     params.VelocityBonusEnabled,
		},
		{
			Type:        BoostTypeArenaSpecialist,
			Multiplier:  math.LegacyNewDecWithPrec(13, 1), // 1.3x (30% bonus)
			Description: "Arena specialist (10+ completions in same arena)",
			Enabled:     true,
		},
		{
			Type:        BoostTypeEarlyAdopter,
			Multiplier:  math.LegacyNewDecWithPrec(20, 1), // 2.0x (100% bonus)
			Description: "Early adopter (joined in first 30 days)",
			Enabled:     params.VelocityBonusEnabled,
		},
		{
			Type:        BoostTypeReferral,
			Multiplier:  math.LegacyNewDecWithPrec(11, 1), // 1.1x (10% bonus)
			Description: "Referral bonus",
			Enabled:     params.VelocityBonusEnabled,
		},
		{
			Type:        BoostTypeVelocity,
			Multiplier:  math.LegacyNewDecWithPrec(20, 1), // 2.0x (100% bonus max)
			Description: "Velocity bonus (fast completion)",
			Enabled:     params.VelocityBonusEnabled,
		},
		{
			Type:        BoostTypeSocialImpact,
			Multiplier:  math.LegacyNewDecWithPrec(14, 1), // 1.4x (40% bonus)
			Description: "Social impact bonus (community-focused IRs)",
			Enabled:     true,
		},
		{
			Type:        BoostTypeQualityValidator,
			Multiplier:  math.LegacyNewDecWithPrec(125, 2), // 1.25x (25% bonus)
			Description: "Quality validation bonus (verifying other submissions)",
			Enabled:     true,
		},
	}
}

// CalculateApplicableBoosts determines which boosts apply to a completion
func (k *Keeper) CalculateApplicableBoosts(ctx sdk.Context, walletAddr string, irID string, completionTimeSeconds int64) ([]BoostMultiplier, math.LegacyDec, error) {
	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok {
		return nil, math.LegacyOneDec(), types.ErrUserRecordNotFound
	}

	applicableBoosts := []BoostMultiplier{}
	totalMultiplier := math.LegacyOneDec()

	allBoosts := k.GetAvailableBoosts(ctx)

	for _, boost := range allBoosts {
		if !boost.Enabled {
			continue
		}

		applies, err := k.checkBoostEligibility(ctx, walletAddr, irID, boost.Type, record, completionTimeSeconds)
		if err != nil {
			continue
		}

		if applies {
			applicableBoosts = append(applicableBoosts, boost)
			// Multiply boosts (multiplicative stacking)
			totalMultiplier = totalMultiplier.Mul(boost.Multiplier)
		}
	}

	return applicableBoosts, totalMultiplier, nil
}

// checkBoostEligibility checks if a specific boost applies
func (k *Keeper) checkBoostEligibility(ctx sdk.Context, walletAddr string, irID string, boostType BoostType, record types.UserConfidenceRecord, completionTimeSeconds int64) (bool, error) {
	switch boostType {
	case BoostTypeFirstCompletion:
		return len(record.CompletedIrs) == 0, nil

	case BoostTypeStreak:
		return k.hasActiveStreak(ctx, walletAddr, 7), nil

	case BoostTypeArenaSpecialist:
		return k.isArenaSpecialist(ctx, walletAddr, irID), nil

	case BoostTypeEarlyAdopter:
		return k.isEarlyAdopter(record), nil

	case BoostTypeReferral:
		return k.hasReferralBonus(ctx, walletAddr), nil

	case BoostTypeVelocity:
		// Velocity boost is calculated separately and applied in rewards.go
		return false, nil

	case BoostTypeSocialImpact:
		return k.isSocialImpactIR(irID), nil

	case BoostTypeQualityValidator:
		return k.isQualityValidator(ctx, walletAddr), nil

	default:
		return false, nil
	}
}

// ApplyBoost applies boost multiplier to base score
func (k *Keeper) ApplyBoost(baseScore uint64, multiplier math.LegacyDec) uint64 {
	boostedScore := math.LegacyNewDec(int64(baseScore)).Mul(multiplier)
	return boostedScore.TruncateInt().Uint64()
}

// hasActiveStreak checks if user has completion streak of at least N days
func (k *Keeper) hasActiveStreak(ctx sdk.Context, walletAddr string, minDays int) bool {
	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok || len(record.CompletedIrs) < minDays {
		return false
	}

	now := ctx.BlockTime()
	consecutiveDays := 0

	for i := 0; i < minDays; i++ {
		dayStart := now.AddDate(0, 0, -i).Truncate(24 * 3600 * 1000000000)
		dayEnd := dayStart.Add(24 * 3600 * 1000000000)

		hasCompletionThisDay := false
		for _, completion := range record.CompletedIrs {
			if completion.CompletedAt != nil {
				completionTime := time.Unix(completion.CompletedAt.Seconds, int64(completion.CompletedAt.Nanos))
				if completionTime.After(dayStart) && completionTime.Before(dayEnd) {
					hasCompletionThisDay = true
					break
				}
			}
		}

		if hasCompletionThisDay {
			consecutiveDays++
		} else {
			break
		}
	}

	return consecutiveDays >= minDays
}

// isArenaSpecialist checks if user has 10+ completions in same arena as current IR
func (k *Keeper) isArenaSpecialist(ctx sdk.Context, walletAddr string, irID string) bool {
	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok {
		return false
	}

	// Get arena for current IR
	arena := ""
	if k.irRegistry != nil {
		arena, _ = k.irRegistry.GetIRArena(irID)
	}
	if arena == "" {
		return false
	}

	// Count completions in this arena
	arenaScore, ok := record.ArenaScores[arena]
	if !ok {
		return false
	}

	return arenaScore.IrCount >= 10
}

// isEarlyAdopter checks if user joined in first 30 days of network
func (k *Keeper) isEarlyAdopter(record types.UserConfidenceRecord) bool {
	// Early adopter if first completion was within 30 days of genesis
	// In production, would check against actual genesis time
	if len(record.CompletedIrs) == 0 {
		return false
	}

	firstCompletion := record.CompletedIrs[0]
	if firstCompletion.CompletedAt == nil {
		return false
	}

	// Placeholder: consider users with completions before height 100000 as early adopters
	return firstCompletion.CompletedHeight < 100000
}

// hasReferralBonus checks if user has active referral bonus
func (k *Keeper) hasReferralBonus(ctx sdk.Context, walletAddr string) bool {
	// In production, would check referral system state
	// Placeholder: always false for now
	return false
}

// isSocialImpactIR checks if IR is tagged as social impact
func (k *Keeper) isSocialImpactIR(irID string) bool {
	// In production, would check IR metadata
	// Placeholder: check if IR ID contains "social" or "community"
	return false
}

// isQualityValidator checks if user has validated other submissions
func (k *Keeper) isQualityValidator(ctx sdk.Context, walletAddr string) bool {
	// In production, would check validation history
	// Placeholder: always false for now
	return false
}

// ApplyBoostToCompletion applies all eligible boosts to an IR completion
func (k *Keeper) ApplyBoostToCompletion(ctx sdk.Context, walletAddr string, irID string, baseScore uint64, completionTimeSeconds int64) (finalScore uint64, boostsApplied []string, totalMultiplier math.LegacyDec, error error) {
	boosts, multiplier, err := k.CalculateApplicableBoosts(ctx, walletAddr, irID, completionTimeSeconds)
	if err != nil {
		return baseScore, nil, math.LegacyOneDec(), err
	}

	finalScore = k.ApplyBoost(baseScore, multiplier)

	boostNames := make([]string, len(boosts))
	for i, boost := range boosts {
		boostNames[i] = string(boost.Type)
	}

	return finalScore, boostNames, multiplier, nil
}

// GetBoostDetails returns detailed information about user's current boosts
func (k *Keeper) GetBoostDetails(ctx sdk.Context, walletAddr string) (map[string]interface{}, error) {
	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok {
		return nil, types.ErrUserRecordNotFound
	}

	details := make(map[string]interface{})

	// Check each boost type
	allBoosts := k.GetAvailableBoosts(ctx)
	for _, boost := range allBoosts {
		if !boost.Enabled {
			continue
		}

		eligible, _ := k.checkBoostEligibility(ctx, walletAddr, "", boost.Type, record, 0)
		details[string(boost.Type)] = map[string]interface{}{
			"eligible":    eligible,
			"multiplier":  boost.Multiplier.String(),
			"description": boost.Description,
		}
	}

	return details, nil
}

// EnableBoost allows governance to enable/disable specific boosts
func (k *Keeper) EnableBoost(ctx sdk.Context, boostType BoostType, enabled bool, authority string) error {
	if authority != k.authority {
		return types.ErrUnauthorized
	}

	params, _ := k.GetParams(ctx)

	// All boost types use the velocity_bonus_enabled flag for now
	// In future, can add individual flags to params proto
	params.VelocityBonusEnabled = enabled

	if err := k.SetParams(params); err != nil {
		return fmt.Errorf("error in EnableBoost for individual: %w", err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"boost_config_updated",
			sdk.NewAttribute("boost_type", string(boostType)),
			sdk.NewAttribute("enabled", fmt.Sprintf("%t", enabled)),
			sdk.NewAttribute("authority", authority),
		),
	)

	return nil
}
