package keeper

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/confidencescore/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CalculateScoreEarned calculates the final score for an IR completion with all multipliers
// Returns: finalScore, velocityBonusBps, arenaBonusBps, jackpotBonusBps, error
// All bonus values are in basis points where 10000 = 1.0x, 12500 = 1.25x, etc.
func (k *Keeper) CalculateScoreEarned(ctx sdk.Context, walletAddr, irID string) (uint64, uint64, uint64, uint64, error) {
	// Get base score from IR registry
	var baseScore uint64
	if k.irRegistry != nil {
		score, err := k.irRegistry.GetIRScore(irID)
		if err != nil {
			return 0, BasisPointsBase, BasisPointsBase, BasisPointsBase, err
		}
		baseScore = score
	} else {
		// Default score for testing
		baseScore = 100
	}

	// Get arena for multiplier calculation
	arena := ""
	if k.irRegistry != nil {
		arena, _ = k.irRegistry.GetIRArena(irID)
	}

	// Calculate multipliers in basis points (10000 = 1.0x)
	velocityBonusBps := k.CalculateVelocityBonus(ctx, walletAddr)
	arenaBonusBps := k.CalculateArenaMultiplier(ctx, walletAddr, arena)
	jackpotBonusBps := k.CheckJackpotWin(ctx, walletAddr, irID)

	// Calculate final score using integer math to ensure determinism
	// Apply multipliers one at a time to avoid overflow
	// finalScore = baseScore * (velocity / 10000) * (arena / 10000) * (jackpot / 10000)
	finalScore := baseScore * velocityBonusBps / BasisPointsBase
	finalScore = finalScore * arenaBonusBps / BasisPointsBase
	finalScore = finalScore * jackpotBonusBps / BasisPointsBase

	return finalScore, velocityBonusBps, arenaBonusBps, jackpotBonusBps, nil
}

// CalculateArenaMultiplier calculates the arena focus multiplier in basis points
// Returns: multiplier in basis points (10000 = 1.0x, 15000 = 1.5x, etc.)
func (k *Keeper) CalculateArenaMultiplier(ctx sdk.Context, walletAddr, arena string) uint64 {
	if arena == "" {
		return BasisPointsBase // 1.0x
	}

	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok {
		return BasisPointsBase // 1.0x
	}

	arenaScore, ok := record.ArenaScores[arena]
	if !ok {
		return BasisPointsBase // 1.0x
	}

	params, _ := k.GetParams(ctx)

	// Apply graduated multipliers based on arena score thresholds
	// Thresholds are stored in descending order, so check from highest to lowest
	var thresholds []uint64
	for threshold := range params.ArenaMultipliersBps {
		thresholds = append(thresholds, threshold)
	}

	// Sort thresholds in descending order (simple bubble sort for small slices)
	for i := 0; i < len(thresholds); i++ {
		for j := i + 1; j < len(thresholds); j++ {
			if thresholds[j] > thresholds[i] {
				thresholds[i], thresholds[j] = thresholds[j], thresholds[i]
			}
		}
	}

	// Find applicable multiplier (highest threshold that user has reached)
	for _, threshold := range thresholds {
		if arenaScore.TotalScore >= threshold {
			return params.ArenaMultipliersBps[threshold]
		}
	}

	return BasisPointsBase // 1.0x
}

// CalculateTotalScore recalculates a user's total score from completions
func (k *Keeper) CalculateTotalScore(ctx sdk.Context, walletAddr string) (uint64, error) {
	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok {
		return 0, types.ErrUserRecordNotFound
	}

	var totalScore uint64
	for _, completion := range record.CompletedIrs {
		totalScore += completion.FinalScore
	}

	return totalScore, nil
}

// CalculateArenaScore calculates the total score for a specific arena
func (k *Keeper) CalculateArenaScore(ctx sdk.Context, walletAddr, arena string) (uint64, error) {
	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok {
		return 0, types.ErrUserRecordNotFound
	}

	var arenaScore uint64
	for _, completion := range record.CompletedIrs {
		if completion.Arena == arena {
			arenaScore += completion.FinalScore
		}
	}

	return arenaScore, nil
}

// RecalculateScore performs a full recalculation of a user's score
// This is useful for admin/governance to fix discrepancies
func (k *Keeper) RecalculateScore(ctx sdk.Context, walletAddr string) (uint64, uint64, []string, error) {
	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok {
		return 0, 0, nil, types.ErrUserRecordNotFound
	}

	previousScore := record.TotalScore
	discrepancies := []string{}

	// Recalculate total from completions
	calculatedTotal, err := k.CalculateTotalScore(ctx, walletAddr)
	if err != nil {
		return 0, 0, nil, err
	}

	if calculatedTotal != previousScore {
		discrepancies = append(discrepancies,
			fmt.Sprintf("Total score mismatch: stored=%d, calculated=%d", previousScore, calculatedTotal))
	}

	// Recalculate arena scores
	arenaScores := make(map[string]*types.ArenaScore)
	for _, completion := range record.CompletedIrs {
		if completion == nil || completion.Arena == "" {
			continue
		}

		arena := completion.Arena
		score, ok := arenaScores[arena]
		if !ok {
			score = &types.ArenaScore{
				ArenaType:  arena,
				TotalScore: 0,
				IrCount:    0,
			}
		}

		score.TotalScore += completion.FinalScore
		score.IrCount++

		params, _ := k.GetParams(ctx)
		score.FocusBonusActive = score.TotalScore >= params.ArenaFocusThreshold

		arenaScores[arena] = score
	}

	// Check for arena score discrepancies
	for arena, calculatedScore := range arenaScores {
		storedScore, ok := record.ArenaScores[arena]
		if !ok {
			discrepancies = append(discrepancies,
				fmt.Sprintf("Arena %s: not found in stored record", arena))
			continue
		}

		if storedScore.TotalScore != calculatedScore.TotalScore {
			discrepancies = append(discrepancies,
				fmt.Sprintf("Arena %s: score mismatch (stored=%d, calculated=%d)",
					arena, storedScore.TotalScore, calculatedScore.TotalScore))
		}

		if storedScore.IrCount != calculatedScore.IrCount {
			discrepancies = append(discrepancies,
				fmt.Sprintf("Arena %s: IR count mismatch (stored=%d, calculated=%d)",
					arena, storedScore.IrCount, calculatedScore.IrCount))
		}
	}

	// Update record if discrepancies found
	if len(discrepancies) > 0 {
		record.TotalScore = calculatedTotal
		record.ArenaScores = arenaScores

		// Update verification status
		params, _ := k.GetParams(ctx)
		if calculatedTotal >= params.VerificationThreshold {
			if record.Status == types.VerificationStatusUnverified {
				record.Status = types.VerificationStatusVerified
				record.VerificationAchievedHeight = uint64(ctx.BlockHeight())
				record.VerificationAchievedAt = timestampFromTime(ctx.BlockTime())
			}
		} else {
			if record.Status == types.VerificationStatusVerified {
				record.Status = types.VerificationStatusUnverified
				record.VerificationAchievedHeight = 0
				record.VerificationAchievedAt = nil
			}
		}

		if err := k.SetUserRecord(ctx, record); err != nil {
			return 0, 0, discrepancies, err
		}
	}

	return previousScore, calculatedTotal, discrepancies, nil
}

// CheckVerificationStatus computes the current verification status
func (k *Keeper) CheckVerificationStatus(ctx sdk.Context, walletAddr string) (types.VerificationStatus, error) {
	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok {
		return types.VerificationStatusUnverified, nil
	}

	params, _ := k.GetParams(ctx)

	// Check for suspended or revoked status (set by governance)
	if record.Status == types.VerificationStatusSuspended ||
		record.Status == types.VerificationStatusRevoked {
		return record.Status, nil
	}

	// Check score threshold
	if record.TotalScore >= params.VerificationThreshold {
		return types.VerificationStatusVerified, nil
	}

	return types.VerificationStatusUnverified, nil
}

// ApplyArenaFocusBonus calculates if a user qualifies for arena focus bonuses
func (k *Keeper) ApplyArenaFocusBonus(ctx sdk.Context, walletAddr string) ([]string, error) {
	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok {
		return nil, types.ErrUserRecordNotFound
	}

	params, _ := k.GetParams(ctx)
	focusArenas := []string{}

	for arena, score := range record.ArenaScores {
		if score.TotalScore >= params.ArenaFocusThreshold {
			focusArenas = append(focusArenas, arena)

			// Update focus bonus active flag
			score.FocusBonusActive = true
			record.ArenaScores[arena] = score
		}
	}

	if len(focusArenas) > 0 {
		if err := k.SetUserRecord(ctx, record); err != nil {
			return nil, err
		}
	}

	return focusArenas, nil
}

// GetArenaBreakdown returns detailed arena score breakdown
func (k *Keeper) GetArenaBreakdown(ctx sdk.Context, walletAddr string) (map[string]*types.ArenaScore, []string, error) {
	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok {
		return nil, nil, types.ErrUserRecordNotFound
	}

	params, _ := k.GetParams(ctx)
	focusArenas := []string{}

	for arena, score := range record.ArenaScores {
		if score != nil && score.TotalScore >= params.ArenaFocusThreshold {
			focusArenas = append(focusArenas, arena)
		}
	}

	return record.ArenaScores, focusArenas, nil
}
