package keeper

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/confidencescore/types"
)

// RecordIRCompletion records an IR completion with full validation
func (k *Keeper) RecordIRCompletion(
	walletAddr string,
	irID string,
	assistantAddr string,
	proofHash []byte,
	verifierHash []byte,
	timestamp int64,
) (uint64, error) {
	// 1. Validate inputs
	if err := k.validateIRCompletionInputs(walletAddr, irID, proofHash, verifierHash, timestamp); err != nil {
		return 0, err
	}

	// 2. Check if IR already completed
	if k.HasCompletedIR(walletAddr, irID) {
		return 0, types.ErrIRAlreadyCompleted
	}

	// 3. Validate IR exists and is active
	if k.irRegistry != nil && !k.irRegistry.IsIRActive(irID) {
		return 0, types.ErrIRNotActive
	}

	// 4. Check anchor requirement (except for IR-000 itself)
	if irID != "IR-000" {
		if err := k.ValidateAnchor(walletAddr); err != nil {
			return 0, err
		}
	}

	// 5. Validate prerequisites
	if err := k.ValidateIRPrerequisites(walletAddr, irID); err != nil {
		return 0, err
	}

	// 6. Check rate limits
	if err := k.CheckRateLimit(walletAddr); err != nil {
		return 0, err
	}

	// 7. Check for replay attacks
	if err := k.CheckReplay(walletAddr, proofHash); err != nil {
		return 0, err
	}

	// 8. Calculate score with bonuses
	scoreEarned, velocityBonus, arenaBonus, jackpotBonus, err := k.CalculateScoreEarned(walletAddr, irID)
	if err != nil {
		return 0, err
	}

	// 9. Get arena type
	arena := ""
	if k.irRegistry != nil {
		arena, _ = k.irRegistry.GetIRArena(irID)
	}

	// 10. Create completion record
	completion := types.IRCompletion{
		IRID:             irID,
		BaseScore:        scoreEarned / uint64(velocityBonus*arenaBonus*jackpotBonus), // Reverse calculate base
		FinalScore:       scoreEarned,
		CompletedAt:      k.currentTime,
		CompletedHeight:  k.currentHeight,
		AssistantAddress: assistantAddr,
		ProofHash:        proofHash,
		VerifierHash:     verifierHash,
		TxHash:           fmt.Sprintf("tx-%d", k.currentHeight),
		VelocityBonus:    velocityBonus,
		ArenaBonus:       arenaBonus,
		JackpotBonus:     jackpotBonus,
		Status:           types.IRCompletionStatusVerified,
		Arena:            arena,
	}

	// 11. Update user record
	record := k.GetOrCreateUserRecord(walletAddr)
	previousScore := record.TotalScore

	record.CompletedIRs = append(record.CompletedIRs, completion)
	record.TotalScore += scoreEarned

	// Update anchor info if this is IR-000
	if irID == "IR-000" {
		record.HasAnchor = true
		record.AnchorInfo = &types.AnchorInfo{
			Completed:          true,
			CompletedAt:        k.currentTime,
			VerifierPluginHash: verifierHash,
			BlockHeight:        k.currentHeight,
			ProofHash:          proofHash,
		}
	}

	// Update arena scores
	if arena != "" {
		arenaScore, ok := record.ArenaScores[arena]
		if !ok {
			arenaScore = types.ArenaScore{
				ArenaType:  arena,
				TotalScore: 0,
				IRCount:    0,
			}
		}
		arenaScore.TotalScore += scoreEarned
		arenaScore.IRCount++

		// Check if arena focus bonus is now active
		params := k.GetParams()
		arenaScore.FocusBonusActive = arenaScore.TotalScore >= params.ArenaFocusThreshold

		record.ArenaScores[arena] = arenaScore
	}

	// Update verification status if threshold crossed
	params := k.GetParams()
	wasVerified := record.Status == types.VerificationStatusVerified
	if record.TotalScore >= params.VerificationThreshold && !wasVerified {
		record.Status = types.VerificationStatusVerified
		record.VerificationAchievedHeight = k.currentHeight
		record.VerificationAchievedAt = k.currentTime
	}

	// 12. Save records
	if err := k.SetUserRecord(record); err != nil {
		return 0, err
	}

	if err := k.SetIRCompletion(walletAddr, completion); err != nil {
		return 0, err
	}

	// 13. Record proof hash to prevent replay
	k.recordProofHash(walletAddr, proofHash)

	// 14. Update rate limit counters
	k.incrementRateLimit(walletAddr)

	// 15. Add to score history
	k.AddScoreChange(types.ScoreChange{
		BlockHeight:   k.currentHeight,
		ScoreDelta:    int64(scoreEarned),
		NewTotal:      record.TotalScore,
		Reason:        types.ChangeReasonIRCompletion,
		RelatedIRID:   irID,
		TxHash:        completion.TxHash,
		Timestamp:     k.currentTime,
		PreviousScore: previousScore,
	})

	return scoreEarned, nil
}

// validateIRCompletionInputs validates input parameters
func (k *Keeper) validateIRCompletionInputs(walletAddr, irID string, proofHash, verifierHash []byte, timestamp int64) error {
	if walletAddr == "" {
		return types.ErrInvalidWalletAddress
	}

	if irID == "" {
		return types.ErrInvalidIRID
	}

	if len(proofHash) != 32 {
		return types.ErrInvalidProofHash
	}

	if len(verifierHash) != 32 {
		return types.ErrInvalidVerifierHash
	}

	// Check timestamp freshness (within 5 minutes)
	if time.Now().Unix()-timestamp > 300 {
		return types.ErrStaleAttestation
	}

	return nil
}

// ValidateAnchor ensures IR-000 is completed
func (k *Keeper) ValidateAnchor(walletAddr string) error {
	record, ok := k.GetUserRecord(walletAddr)
	if !ok || !record.HasAnchor {
		return types.ErrAnchorNotCompleted
	}

	if record.AnchorInfo == nil || !record.AnchorInfo.Completed {
		return types.ErrInvalidAnchor
	}

	return nil
}

// ValidateIRPrerequisites checks if all prerequisites are met
func (k *Keeper) ValidateIRPrerequisites(walletAddr, irID string) error {
	if k.irRegistry == nil {
		// Skip validation if no registry available
		return nil
	}

	prerequisites, err := k.irRegistry.GetIRPrerequisites(irID)
	if err != nil {
		return err
	}

	for _, prereqID := range prerequisites {
		if !k.HasCompletedIR(walletAddr, prereqID) {
			return fmt.Errorf("%w: %s", types.ErrPrerequisiteMissing, prereqID)
		}
	}

	return nil
}

// CheckRateLimit enforces hourly and daily rate limits
func (k *Keeper) CheckRateLimit(walletAddr string) error {
	params := k.GetParams()

	// Check hourly limit
	hourKey := k.getRateLimitKey(walletAddr, "hour")
	k.mu.RLock()
	hourCount := k.rateLimits[walletAddr][hourKey]
	k.mu.RUnlock()

	if uint64(hourCount) >= params.MaxIRsPerHour {
		return types.ErrHourlyLimitExceeded
	}

	// Check daily limit
	dayKey := k.getRateLimitKey(walletAddr, "day")
	k.mu.RLock()
	dayCount := k.rateLimits[walletAddr][dayKey]
	k.mu.RUnlock()

	if uint64(dayCount) >= params.MaxIRsPerDay {
		return types.ErrDailyLimitExceeded
	}

	return nil
}

// CheckReplay checks for proof hash reuse
func (k *Keeper) CheckReplay(walletAddr string, proofHash []byte) error {
	k.mu.RLock()
	defer k.mu.RUnlock()

	hashKey := string(proofHash)
	if hashes, ok := k.proofHashes[walletAddr]; ok {
		if hashes[hashKey] {
			return types.ErrReplayDetected
		}
	}

	return nil
}

// recordProofHash stores a proof hash to prevent replay
func (k *Keeper) recordProofHash(walletAddr string, proofHash []byte) {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.proofHashes[walletAddr] == nil {
		k.proofHashes[walletAddr] = make(map[string]bool)
	}

	hashKey := string(proofHash)
	k.proofHashes[walletAddr][hashKey] = true
}

// incrementRateLimit increments rate limit counters
func (k *Keeper) incrementRateLimit(walletAddr string) {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.rateLimits[walletAddr] == nil {
		k.rateLimits[walletAddr] = make(map[string]int)
	}

	hourKey := k.getRateLimitKey(walletAddr, "hour")
	dayKey := k.getRateLimitKey(walletAddr, "day")

	k.rateLimits[walletAddr][hourKey]++
	k.rateLimits[walletAddr][dayKey]++
}

// getRateLimitKey generates a time-based key for rate limiting
func (k *Keeper) getRateLimitKey(walletAddr, window string) string {
	now := time.Now()

	switch window {
	case "hour":
		return fmt.Sprintf("hour_%d_%d", now.Year(), now.YearDay()*24+now.Hour())
	case "day":
		return fmt.Sprintf("day_%d_%d", now.Year(), now.YearDay())
	default:
		return ""
	}
}

// CalculateVelocityBonus calculates time-based completion bonus
func (k *Keeper) CalculateVelocityBonus(walletAddr string) float32 {
	record, ok := k.GetUserRecord(walletAddr)
	if !ok || !record.HasAnchor {
		return 1.0
	}

	// Don't apply velocity bonus if already verified
	params := k.GetParams()
	if record.TotalScore >= params.VerificationThreshold {
		return 1.0
	}

	anchorTime := record.AnchorInfo.CompletedAt
	currentTime := k.currentTime
	daysElapsed := float64(currentTime-anchorTime) / 86400.0 // seconds to days

	// Apply velocity bonuses based on time tiers
	for i, dayThreshold := range params.VelocityBonusDays {
		if daysElapsed <= float64(dayThreshold) {
			return params.VelocityBonusMultipliers[i]
		}
	}

	return 1.0 // No bonus after all tiers
}

// CheckJackpotWin checks for probabilistic jackpot bonus
func (k *Keeper) CheckJackpotWin(walletAddr, irID string) float32 {
	params := k.GetParams()

	// Generate deterministic but unpredictable seed
	seedData := fmt.Sprintf("%s:%d:%s", walletAddr, k.currentHeight, irID)
	hash := sha256.Sum256([]byte(seedData))
	seedInt := binary.BigEndian.Uint64(hash[:8])

	// Check jackpot odds (highest multiplier first)
	for i := len(params.JackpotOdds) - 1; i >= 0; i-- {
		odds := params.JackpotOdds[i]
		multiplier := params.JackpotMultipliers[i]

		// Use different modulo values for different tiers to avoid overlap
		checkValue := seedInt % (odds * uint64(i+1))
		if checkValue == uint64(77*(i+1)) {
			return multiplier
		}
	}

	return 1.0
}

// CleanupExpiredRateLimits removes old rate limit entries
func (k *Keeper) CleanupExpiredRateLimits() {
	k.mu.Lock()
	defer k.mu.Unlock()

	now := time.Now()
	currentHourKey := fmt.Sprintf("hour_%d_%d", now.Year(), now.YearDay()*24+now.Hour())
	currentDayKey := fmt.Sprintf("day_%d_%d", now.Year(), now.YearDay())

	for wallet, limits := range k.rateLimits {
		for key := range limits {
			// Remove if not current hour or day
			if key != currentHourKey && key != currentDayKey {
				delete(k.rateLimits[wallet], key)
			}
		}
	}
}
