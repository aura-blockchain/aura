package keeper

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/confidencescore/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RecordIRCompletion records an IR completion with full validation
func (k *Keeper) RecordIRCompletion(
	ctx sdk.Context,
	walletAddr string,
	irID string,
	assistantAddr string,
	proofHash []byte,
	verifierHash []byte,
	timestamp int64,
) (uint64, error) {
	// 1. Validate inputs
	if err := k.validateIRCompletionInputs(ctx, walletAddr, irID, proofHash, verifierHash, timestamp); err != nil {
		return 0, err
	}

	// 2. Check if IR already completed
	if k.HasCompletedIR(ctx, walletAddr, irID) {
		return 0, types.ErrIRAlreadyCompleted
	}

	// 3. Validate IR exists and is active
	if k.irRegistry != nil && !k.irRegistry.IsIRActive(irID) {
		return 0, types.ErrIRNotActive
	}

	// 4. Check anchor requirement (except for IR-000 itself)
	if irID != "IR-000" {
		if err := k.ValidateAnchor(ctx, walletAddr); err != nil {
			return 0, err
		}
	}

	// 5. Validate prerequisites
	if err := k.ValidateIRPrerequisites(ctx, walletAddr, irID); err != nil {
		return 0, err
	}

	// 6. Check rate limits
	if err := k.CheckRateLimit(ctx, walletAddr); err != nil {
		return 0, err
	}

	// 7. Check for replay attacks
	if err := k.CheckReplay(ctx, walletAddr, proofHash); err != nil {
		return 0, err
	}

	// 8. Calculate score with bonuses
	scoreEarned, velocityBonus, arenaBonus, jackpotBonus, err := k.CalculateScoreEarned(ctx, walletAddr, irID)
	if err != nil {
		return 0, err
	}

	// 9. Get arena type
	arena := ""
	if k.irRegistry != nil {
		arena, _ = k.irRegistry.GetIRArena(irID)
	}

	// 10. Create completion record
	completion := &types.IRCompletion{
		IrId:             irID,
		BaseScore:        scoreEarned / uint64(velocityBonus*arenaBonus*jackpotBonus), // Reverse calculate base
		FinalScore:       scoreEarned,
		CompletedAt:      timestampFromTime(ctx.BlockTime()),
		CompletedHeight:  uint64(ctx.BlockHeight()),
		AssistantAddress: assistantAddr,
		ProofHash:        proofHash,
		VerifierHash:     verifierHash,
		TxHash:           fmt.Sprintf("tx-%d", ctx.BlockHeight()),
		VelocityBonus:    velocityBonus,
		ArenaBonus:       arenaBonus,
		JackpotBonus:     jackpotBonus,
		Status:           types.IRCompletionStatusVerified,
		Arena:            arena,
	}

	// 11. Update user record
	record := k.GetOrCreateUserRecord(ctx, walletAddr)
	previousScore := record.TotalScore

	record.CompletedIrs = append(record.CompletedIrs, completion)
	record.TotalScore += scoreEarned

	// Update anchor info if this is IR-000
	if irID == "IR-000" {
		record.HasAnchor = true
		record.AnchorInfo = &types.AnchorInfo{
			Completed:          true,
			CompletedAt:        timestampFromTime(ctx.BlockTime()),
			VerifierPluginHash: verifierHash,
			BlockHeight:        uint64(ctx.BlockHeight()),
			ProofHash:          proofHash,
		}
	}

	// Update arena scores
	if arena != "" {
		arenaScore, ok := record.ArenaScores[arena]
		if !ok {
			arenaScore = &types.ArenaScore{
				ArenaType:  arena,
				TotalScore: 0,
				IrCount:    0,
			}
		}
		arenaScore.TotalScore += scoreEarned
		arenaScore.IrCount++

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
		record.VerificationAchievedHeight = uint64(ctx.BlockHeight())
		record.VerificationAchievedAt = timestampFromTime(ctx.BlockTime())
	}

	// 12. Save records
	if err := k.SetUserRecord(ctx, record); err != nil {
		return 0, err
	}

	if err := k.SetIRCompletion(ctx, walletAddr, *completion); err != nil {
		return 0, err
	}

	// 13. Record proof hash to prevent replay
	k.recordProofHash(ctx, walletAddr, proofHash)

	// 14. Update rate limit counters
	k.incrementRateLimit(ctx, walletAddr)

	// 15. Add to score history
	k.AddScoreChange(ctx, walletAddr, types.ScoreChange{
		ScoreDelta:    int64(scoreEarned),
		NewTotal:      record.TotalScore,
		Reason:        types.ChangeReasonIRCompletion,
		RelatedIrId:   irID,
		TxHash:        completion.TxHash,
		PreviousScore: previousScore,
	})

	return scoreEarned, nil
}

// validateIRCompletionInputs validates input parameters
func (k *Keeper) validateIRCompletionInputs(ctx sdk.Context, walletAddr, irID string, proofHash, verifierHash []byte, timestamp int64) error {
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
	if ctx.BlockTime().Unix()-timestamp > 300 {
		return types.ErrStaleAttestation
	}

	return nil
}

// ValidateAnchor ensures IR-000 is completed
func (k *Keeper) ValidateAnchor(ctx sdk.Context, walletAddr string) error {
	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok || !record.HasAnchor {
		return types.ErrAnchorNotCompleted
	}

	if record.AnchorInfo == nil || !record.AnchorInfo.Completed {
		return types.ErrInvalidAnchor
	}

	return nil
}

// ValidateIRPrerequisites checks if all prerequisites are met
func (k *Keeper) ValidateIRPrerequisites(ctx sdk.Context, walletAddr, irID string) error {
	if k.irRegistry == nil {
		// Skip validation if no registry available
		return nil
	}

	prerequisites, err := k.irRegistry.GetIRPrerequisites(irID)
	if err != nil {
		return err
	}

	for _, prereqID := range prerequisites {
		if !k.HasCompletedIR(ctx, walletAddr, prereqID) {
			return fmt.Errorf("%w: %s", types.ErrPrerequisiteMissing, prereqID)
		}
	}

	return nil
}

// CheckRateLimit enforces hourly and daily rate limits
func (k *Keeper) CheckRateLimit(ctx sdk.Context, walletAddr string) error {
	params := k.GetParams()

	// Check hourly limit
	hourKey := k.getRateLimitKey(ctx, walletAddr, "hour")
	hourCount := k.getRateLimitCount(ctx, walletAddr, hourKey)

	if uint64(hourCount) >= params.MaxIrsPerHour {
		return types.ErrHourlyLimitExceeded
	}

	// Check daily limit
	dayKey := k.getRateLimitKey(ctx, walletAddr, "day")
	dayCount := k.getRateLimitCount(ctx, walletAddr, dayKey)

	if uint64(dayCount) >= params.MaxIrsPerDay {
		return types.ErrDailyLimitExceeded
	}

	return nil
}

// CheckReplay checks for proof hash reuse
func (k *Keeper) CheckReplay(ctx sdk.Context, walletAddr string, proofHash []byte) error {
	if k.HasProofHash(ctx, walletAddr, proofHash) {
		return types.ErrReplayDetected
	}
	return nil
}

// recordProofHash stores a proof hash to prevent replay
func (k *Keeper) recordProofHash(ctx sdk.Context, walletAddr string, proofHash []byte) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.ProofHashStoreKey(walletAddr, proofHash)
	store.Set([]byte(key), []byte{1})
}

// incrementRateLimit increments rate limit counters
func (k *Keeper) incrementRateLimit(ctx sdk.Context, walletAddr string) {
	hourKey := k.getRateLimitKey(ctx, walletAddr, "hour")
	dayKey := k.getRateLimitKey(ctx, walletAddr, "day")

	store := k.storeService.OpenKVStore(ctx)

	// Increment hour counter
	hourStoreKey := types.RateLimitStoreKey(walletAddr, hourKey)
	hourBz, _ := store.Get([]byte(hourStoreKey))
	hourCount := 0
	if len(hourBz) > 0 {
		hourCount = int(binary.BigEndian.Uint32(hourBz))
	}
	hourCount++
	hourCountBz := make([]byte, 4)
	binary.BigEndian.PutUint32(hourCountBz, uint32(hourCount))
	store.Set([]byte(hourStoreKey), hourCountBz)

	// Increment day counter
	dayStoreKey := types.RateLimitStoreKey(walletAddr, dayKey)
	dayBz, _ := store.Get([]byte(dayStoreKey))
	dayCount := 0
	if len(dayBz) > 0 {
		dayCount = int(binary.BigEndian.Uint32(dayBz))
	}
	dayCount++
	dayCountBz := make([]byte, 4)
	binary.BigEndian.PutUint32(dayCountBz, uint32(dayCount))
	store.Set([]byte(dayStoreKey), dayCountBz)
}

// getRateLimitCount gets the current rate limit count
func (k *Keeper) getRateLimitCount(ctx sdk.Context, walletAddr, window string) int {
	store := k.storeService.OpenKVStore(ctx)
	key := types.RateLimitStoreKey(walletAddr, window)
	bz, err := store.Get([]byte(key))
	if err != nil || len(bz) == 0 {
		return 0
	}
	return int(binary.BigEndian.Uint32(bz))
}

// getRateLimitKey generates a time-based key for rate limiting
func (k *Keeper) getRateLimitKey(ctx sdk.Context, walletAddr, window string) string {
	now := ctx.BlockTime()

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
func (k *Keeper) CalculateVelocityBonus(ctx sdk.Context, walletAddr string) float32 {
	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok || !record.HasAnchor {
		return 1.0
	}

	// Don't apply velocity bonus if already verified
	params := k.GetParams()
	if record.TotalScore >= params.VerificationThreshold {
		return 1.0
	}

	anchorTime := record.AnchorInfo.CompletedAt
	currentTime := ctx.BlockTime()

	// Convert anchorTime from protobuf timestamp to time.Time
	var daysElapsed float64
	if anchorTime != nil {
		anchorUnix := anchorTime.AsTime()
		daysElapsed = currentTime.Sub(anchorUnix).Hours() / 24.0
	}

	// Apply velocity bonuses based on time tiers
	for i, dayThreshold := range params.VelocityBonusDays {
		if daysElapsed <= float64(dayThreshold) {
			return params.VelocityBonusMultipliers[i]
		}
	}

	return 1.0 // No bonus after all tiers
}

// CheckJackpotWin checks for probabilistic jackpot bonus
func (k *Keeper) CheckJackpotWin(ctx sdk.Context, walletAddr, irID string) float32 {
	params := k.GetParams()

	// Generate deterministic but unpredictable seed
	seedData := fmt.Sprintf("%s:%d:%s", walletAddr, ctx.BlockHeight(), irID)
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
func (k *Keeper) CleanupExpiredRateLimits(ctx sdk.Context) {
	now := ctx.BlockTime()
	currentHourKey := fmt.Sprintf("hour_%d_%d", now.Year(), now.YearDay()*24+now.Hour())
	currentDayKey := fmt.Sprintf("day_%d_%d", now.Year(), now.YearDay())

	store := k.storeService.OpenKVStore(ctx)
	prefix := []byte(types.RateLimitStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return
	}
	defer iterator.Close()

	keysToDelete := []string{}
	for ; iterator.Valid(); iterator.Next() {
		key := string(iterator.Key())
		// Delete if not current hour or day
		if !contains(key, currentHourKey) && !contains(key, currentDayKey) {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		store.Delete([]byte(key))
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr
}
