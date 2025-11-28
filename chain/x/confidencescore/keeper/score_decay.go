package keeper

import (
        storetypes "cosmossdk.io/store/types"
	"fmt"

	"github.com/aequitas/aura/chain/x/confidencescore/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================
// SCORE DECAY MECHANISM
// Implements time-based score depreciation to encourage continued participation
// ============================

// DecayConfig defines decay parameters
type DecayConfig struct {
	Enabled            bool
	DecayRatePercent   uint64 // Decay percentage per period (e.g., 5 = 5%)
	DecayPeriodDays    uint64 // Days between decay applications
	MinimumScore       uint64 // Minimum score after decay (never decays below this)
	ExemptionThreshold uint64 // Scores above this are exempt from decay
	InactivityDays     uint64 // Days of inactivity before decay starts
}

// GetDecayConfig returns current decay configuration from params
func (k *Keeper) GetDecayConfig() DecayConfig {
	params := k.GetParams()

	// Note: Decay params not yet in proto - using defaults
	// TODO: Add decay params to confidencescore.proto Params message
	return DecayConfig{
		Enabled:            params.StalenessEnabled, // Reuse staleness flag for decay
		DecayRatePercent:   5,                       // Default 5% per period
		DecayPeriodDays:    30,                      // Default 30 days
		MinimumScore:       1000,                    // Default minimum score
		ExemptionThreshold: 20000,                   // Scores above 20k exempt
		InactivityDays:     90,                      // 90 days inactivity
	}
}

// ApplyScoreDecay applies time-based decay to a user's score
// Returns: (oldScore, newScore, decayAmount, error)
func (k *Keeper) ApplyScoreDecay(ctx sdk.Context, walletAddr string) (uint64, uint64, uint64, error) {
	config := k.GetDecayConfig()

	if !config.Enabled {
		return 0, 0, 0, fmt.Errorf("score decay is disabled")
	}

	// Get user record
	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok {
		return 0, 0, 0, types.ErrUserRecordNotFound
	}

	oldScore := record.TotalScore

	// Check if exempt from decay
	if oldScore >= config.ExemptionThreshold {
		return oldScore, oldScore, 0, nil
	}

	// Check if minimum score reached
	if oldScore <= config.MinimumScore {
		return oldScore, oldScore, 0, nil
	}

	// Check inactivity period
	if record.LastUpdated != nil {
		lastActivity := record.LastUpdated.AsTime()
		daysSinceActivity := ctx.BlockTime().Sub(lastActivity).Hours() / 24

		if daysSinceActivity < float64(config.InactivityDays) {
			return oldScore, oldScore, 0, nil // No decay - user is active
		}
	}

	// Calculate decay amount
	decayAmount := (oldScore * config.DecayRatePercent) / 100

	// Apply decay
	newScore := oldScore
	if oldScore > decayAmount {
		newScore = oldScore - decayAmount
	} else {
		newScore = 0
	}

	// Enforce minimum
	if newScore < config.MinimumScore {
		newScore = config.MinimumScore
	}

	actualDecay := oldScore - newScore

	// Update record
	record.TotalScore = newScore
	params := k.GetParams()

	// Check verification status
	if newScore < params.VerificationThreshold && record.Status == types.VerificationStatusVerified {
		record.Status = types.VerificationStatusUnverified
		record.VerificationAchievedHeight = 0
		record.VerificationAchievedAt = nil
	}

	if err := k.SetUserRecord(ctx, record); err != nil {
		return 0, 0, 0, err
	}

	// Record decay in history
	k.AddScoreChange(ctx, types.ScoreChange{
		ScoreDelta:    -int64(actualDecay),
		NewTotal:      newScore,
		Reason:        types.ChangeReasonGovernanceAdjustment, // Use governance for decay
		RelatedIrId:   "score_decay",
		TxHash:        fmt.Sprintf("decay-%s-%d", walletAddr, ctx.BlockHeight()),
		PreviousScore: oldScore,
	})

	return oldScore, newScore, actualDecay, nil
}

// ApplyBatchDecay applies decay to all eligible users
// Used in EndBlocker to process decay periodically
func (k *Keeper) ApplyBatchDecay(ctx sdk.Context) (int, uint64, error) {
	config := k.GetDecayConfig()

	if !config.Enabled {
		return 0, 0, nil
	}

	// Iterate all users from KV store
	store := k.storeService.OpenKVStore(ctx)
	prefix := []byte(types.UserRecordStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return 0, 0, err
	}
	defer iterator.Close()

	processedCount := 0
	totalDecayed := uint64(0)

	for ; iterator.Valid(); iterator.Next() {
		var record types.UserConfidenceRecord
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			continue
		}

		_, _, decayAmount, err := k.ApplyScoreDecay(ctx, record.WalletAddress)
		if err == nil && decayAmount > 0 {
			processedCount++
			totalDecayed += decayAmount
		}
	}

	// Emit batch decay event
	if processedCount > 0 {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"batch_score_decay",
				sdk.NewAttribute("users_processed", fmt.Sprintf("%d", processedCount)),
				sdk.NewAttribute("total_decayed", fmt.Sprintf("%d", totalDecayed)),
				sdk.NewAttribute("block_height", fmt.Sprintf("%d", ctx.BlockHeight())),
			),
		)
	}

	return processedCount, totalDecayed, nil
}

// ShouldApplyDecay checks if decay should be applied to a user
func (k *Keeper) ShouldApplyDecay(ctx sdk.Context, walletAddr string) (bool, string) {
	config := k.GetDecayConfig()

	if !config.Enabled {
		return false, "decay disabled"
	}

	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok {
		return false, "user not found"
	}

	if record.TotalScore >= config.ExemptionThreshold {
		return false, "score above exemption threshold"
	}

	if record.TotalScore <= config.MinimumScore {
		return false, "score at minimum"
	}

	if record.LastUpdated != nil {
		lastActivity := record.LastUpdated.AsTime()
		daysSinceActivity := ctx.BlockTime().Sub(lastActivity).Hours() / 24

		if daysSinceActivity < float64(config.InactivityDays) {
			return false, "user is active"
		}
	}

	return true, "eligible for decay"
}

// GetDecayPreview calculates decay without applying it
func (k *Keeper) GetDecayPreview(ctx sdk.Context, walletAddr string) (currentScore, projectedScore, decayAmount uint64, err error) {
	config := k.GetDecayConfig()

	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok {
		return 0, 0, 0, types.ErrUserRecordNotFound
	}

	currentScore = record.TotalScore
	projectedScore = currentScore
	decayAmount = 0

	shouldDecay, _ := k.ShouldApplyDecay(ctx, walletAddr)
	if !shouldDecay {
		return currentScore, projectedScore, 0, nil
	}

	// Calculate decay
	decayAmount = (currentScore * config.DecayRatePercent) / 100
	if currentScore > decayAmount {
		projectedScore = currentScore - decayAmount
	} else {
		projectedScore = 0
	}

	if projectedScore < config.MinimumScore {
		projectedScore = config.MinimumScore
	}

	decayAmount = currentScore - projectedScore

	return currentScore, projectedScore, decayAmount, nil
}

// PauseDecayForUser temporarily exempts a user from decay
// Useful for special circumstances (e.g., medical leave, military service)
func (k *Keeper) PauseDecayForUser(ctx sdk.Context, walletAddr string, durationDays uint64, reason string) error {
	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok {
		return types.ErrUserRecordNotFound
	}

	// Update the LastUpdated to prevent decay
	record.LastUpdated = timestampFromTime(ctx.BlockTime())

	if err := k.SetUserRecord(ctx, record); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"decay_paused",
			sdk.NewAttribute("wallet", walletAddr),
			sdk.NewAttribute("duration_days", fmt.Sprintf("%d", durationDays)),
			sdk.NewAttribute("reason", reason),
		),
	)

	return nil
}

// RestoreDecayedScore allows governance to restore score lost to decay
func (k *Keeper) RestoreDecayedScore(ctx sdk.Context, walletAddr string, amount uint64, authority string) error {
	// Validate authority
	if authority != k.authority {
		return types.ErrUnauthorized
	}

	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok {
		return types.ErrUserRecordNotFound
	}

	previousScore := record.TotalScore
	record.TotalScore += amount

	params := k.GetParams()
	if record.TotalScore >= params.VerificationThreshold {
		if record.Status == types.VerificationStatusUnverified {
			record.Status = types.VerificationStatusVerified
			record.VerificationAchievedHeight = uint64(ctx.BlockHeight())
			record.VerificationAchievedAt = timestampFromTime(ctx.BlockTime())
		}
	}

	if err := k.SetUserRecord(ctx, record); err != nil {
		return err
	}

	k.AddScoreChange(ctx, types.ScoreChange{
		ScoreDelta:    int64(amount),
		NewTotal:      record.TotalScore,
		Reason:        types.ChangeReasonGovernanceAdjustment,
		RelatedIrId:   "decay_restoration",
		TxHash:        fmt.Sprintf("restore-%s-%d", walletAddr, ctx.BlockHeight()),
		PreviousScore: previousScore,
	})

	return nil
}
