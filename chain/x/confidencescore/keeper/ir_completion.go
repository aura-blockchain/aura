package keeper

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"

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
	// Calculate base score by reversing the multipliers
	// baseScore = finalScore * 10000^3 / (velocity * arena * jackpot)
	// To avoid overflow and match the forward calculation:
	baseScore := scoreEarned * BasisPointsBase / velocityBonus
	baseScore = baseScore * BasisPointsBase / arenaBonus
	baseScore = baseScore * BasisPointsBase / jackpotBonus

	completion := &types.IRCompletion{
		IrId:             irID,
		BaseScore:        baseScore,
		FinalScore:       scoreEarned,
		CompletedAt:      timestampFromTime(ctx.BlockTime()),
		CompletedHeight:  uint64(ctx.BlockHeight()),
		AssistantAddress: assistantAddr,
		ProofHash:        proofHash,
		VerifierHash:     verifierHash,
		TxHash:           fmt.Sprintf("tx-%d", ctx.BlockHeight()),
		VelocityBonusBps: velocityBonus,
		ArenaBonusBps:    arenaBonus,
		JackpotBonusBps:  jackpotBonus,
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
		params, _ := k.GetParams(ctx)
		arenaScore.FocusBonusActive = arenaScore.TotalScore >= params.ArenaFocusThreshold

		record.ArenaScores[arena] = arenaScore
	}

	// Update verification status if threshold crossed
	params, _ := k.GetParams(ctx)
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
	if err := k.AddScoreChange(ctx, walletAddr, types.ScoreChange{
		ScoreDelta:    int64(scoreEarned),
		NewTotal:      record.TotalScore,
		Reason:        types.ChangeReasonIRCompletion,
		RelatedIrId:   irID,
		TxHash:        completion.TxHash,
		PreviousScore: previousScore,
	}); err != nil {
		return 0, err
	}

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
		return fmt.Errorf("failed to get for ValidateIRPrerequisites: %w", err)
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
	params, _ := k.GetParams(ctx)

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
	if err := store.Set([]byte(key), []byte{1}); err != nil {
		k.logger.Error("failed to store proof hash", "wallet", walletAddr, "err", err)
	}
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
	if err := store.Set([]byte(hourStoreKey), hourCountBz); err != nil {
		k.logger.Error("failed to persist hourly rate limit", "wallet", walletAddr, "err", err)
	}

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
	if err := store.Set([]byte(dayStoreKey), dayCountBz); err != nil {
		k.logger.Error("failed to persist daily rate limit", "wallet", walletAddr, "err", err)
	}
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

// BasisPointsBase is the base for multiplier calculations (10000 = 1.0x)
const BasisPointsBase uint64 = 10000

// CalculateVelocityBonus calculates time-based completion bonus in basis points
// Returns: multiplier in basis points (10000 = 1.0x, 12500 = 1.25x, etc.)
func (k *Keeper) CalculateVelocityBonus(ctx sdk.Context, walletAddr string) uint64 {
	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok || !record.HasAnchor {
		return BasisPointsBase // 1.0x
	}

	// Don't apply velocity bonus if already verified
	params, _ := k.GetParams(ctx)
	if record.TotalScore >= params.VerificationThreshold {
		return BasisPointsBase // 1.0x
	}

	anchorTime := record.AnchorInfo.CompletedAt
	currentTime := ctx.BlockTime()

	// Calculate days elapsed since anchor completion using integer math
	var daysElapsed uint64
	if anchorTime != nil {
		anchorTimeGo := time.Unix(anchorTime.Seconds, int64(anchorTime.Nanos))
		hoursElapsed := uint64(currentTime.Sub(anchorTimeGo).Hours())
		daysElapsed = hoursElapsed / 24
	}

	// Apply velocity bonuses based on time tiers
	for i, dayThreshold := range params.VelocityBonusDays {
		if daysElapsed <= dayThreshold {
			return params.VelocityBonusMultipliersBps[i]
		}
	}

	return BasisPointsBase // 1.0x - No bonus after all tiers
}

// CheckJackpotWin checks for probabilistic jackpot bonus in basis points
// Returns: multiplier in basis points (10000 = 1.0x, 50000 = 5.0x, etc.)
func (k *Keeper) CheckJackpotWin(ctx sdk.Context, walletAddr, irID string) uint64 {
	params, _ := k.GetParams(ctx)

	// Generate deterministic but unpredictable seed
	seedData := fmt.Sprintf("%s:%d:%s", walletAddr, ctx.BlockHeight(), irID)
	hash := sha256.Sum256([]byte(seedData))
	seedInt := binary.BigEndian.Uint64(hash[:8])

	// Check jackpot odds (highest multiplier first)
	for i := len(params.JackpotOdds) - 1; i >= 0; i-- {
		odds := params.JackpotOdds[i]
		multiplierBps := params.JackpotMultipliersBps[i]

		// Use different modulo values for different tiers to avoid overlap
		checkValue := seedInt % (odds * uint64(i+1))
		if checkValue == uint64(77*(i+1)) {
			return multiplierBps
		}
	}

	return BasisPointsBase // 1.0x
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
		if err := store.Delete([]byte(key)); err != nil {
			k.logger.Error("failed to cleanup rate limit key", "key", key, "err", err)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr
}
