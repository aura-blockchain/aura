package keeper

import (
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/confidencescore/types"
)

// SlashScore slashes a user's score for fraud or policy violations
func (k *Keeper) SlashScore(
	walletAddr string,
	irID string,
	slashAmount uint64,
	reason types.SlashReason,
	authority string,
	evidence string,
) (uint64, uint64, bool, string, error) {
	// Validate inputs
	if walletAddr == "" {
		return 0, 0, false, "", types.ErrInvalidWalletAddress
	}

	if slashAmount == 0 {
		return 0, 0, false, "", types.ErrInvalidSlashAmount
	}

	// Get user record
	record, ok := k.GetUserRecord(walletAddr)
	if !ok {
		return 0, 0, false, "", types.ErrUserRecordNotFound
	}

	previousScore := record.TotalScore
	params := k.GetParams()

	// Apply slash (ensuring it doesn't go below 0)
	var newScore uint64
	if record.TotalScore < slashAmount {
		newScore = 0
	} else {
		newScore = record.TotalScore - slashAmount
	}

	wasVerified := record.Status == types.VerificationStatusVerified
	verificationRevoked := false

	// Update verification status if dropped below threshold
	if newScore < params.VerificationThreshold && wasVerified {
		record.Status = types.VerificationStatusUnverified
		record.VerificationAchievedHeight = 0
		record.VerificationAchievedAt = 0
		verificationRevoked = true
	}

	record.TotalScore = newScore

	// Save updated record
	if err := k.SetUserRecord(record); err != nil {
		return 0, 0, false, "", err
	}

	// Create slash record
	slashTxHash := fmt.Sprintf("slash-%s-%d", walletAddr, k.currentHeight)
	appealDeadline := k.currentTime + (14 * 24 * 3600) // 14 days from now

	slashRecord := types.SlashRecord{
		WalletAddress:  walletAddr,
		SlashAmount:    slashAmount,
		Reason:         reason,
		SlashHeight:    k.currentHeight,
		SlashTime:      k.currentTime,
		RelatedIRID:    irID,
		SlashTxHash:    slashTxHash,
		AppealDeadline: appealDeadline,
		Appealed:       false,
		Resolved:       false,
		Authority:      authority,
		Evidence:       evidence,
	}

	if err := k.AddSlashRecord(slashRecord); err != nil {
		return 0, 0, false, "", err
	}

	// Add to score history
	k.AddScoreChange(types.ScoreChange{
		BlockHeight:   k.currentHeight,
		ScoreDelta:    -int64(slashAmount),
		NewTotal:      newScore,
		Reason:        types.ChangeReasonFraudSlash,
		RelatedIRID:   irID,
		TxHash:        slashTxHash,
		Timestamp:     k.currentTime,
		PreviousScore: previousScore,
	})

	return previousScore, newScore, verificationRevoked, slashTxHash, nil
}

// AppealSlash allows a user to appeal a slash decision
func (k *Keeper) AppealSlash(
	walletAddr string,
	slashTxHash string,
	evidence string,
	deposit string,
) (bool, int64, error) {
	// Get slash record
	slashRecord, ok := k.GetSlashRecord(walletAddr, slashTxHash)
	if !ok {
		return false, 0, types.ErrSlashNotFound
	}

	// Check if already appealed
	if slashRecord.Appealed {
		return false, 0, types.ErrSlashAlreadyAppealed
	}

	// Check if appeal deadline has passed
	if k.currentTime > slashRecord.AppealDeadline {
		return false, 0, types.ErrAppealExpired
	}

	// Check if already resolved
	if slashRecord.Resolved {
		return false, 0, types.ErrAppealAlreadyResolved
	}

	// Validate deposit (in real implementation, would verify actual token deposit)
	params := k.GetParams()
	if deposit != params.AppealDeposit {
		return false, 0, types.ErrInsufficientDeposit
	}

	// Mark as appealed
	slashRecord.Appealed = true
	slashRecord.Evidence = evidence

	if err := k.UpdateSlashRecord(walletAddr, slashRecord); err != nil {
		return false, 0, err
	}

	// Calculate review deadline (14 days from appeal)
	reviewDeadline := k.currentTime + (14 * 24 * 3600)

	return true, reviewDeadline, nil
}

// ResolveAppeal resolves an appeal (governance action)
func (k *Keeper) ResolveAppeal(
	walletAddr string,
	slashTxHash string,
	restoreScore bool,
	authority string,
	resolutionNotes string,
) (uint64, bool, error) {
	// Get slash record
	slashRecord, ok := k.GetSlashRecord(walletAddr, slashTxHash)
	if !ok {
		return 0, false, types.ErrSlashNotFound
	}

	// Check if appealed
	if !slashRecord.Appealed {
		return 0, false, fmt.Errorf("slash not appealed")
	}

	// Check if already resolved
	if slashRecord.Resolved {
		return 0, false, types.ErrAppealAlreadyResolved
	}

	var restoredScore uint64
	depositReturned := false

	if restoreScore {
		// Restore the slashed score
		record, ok := k.GetUserRecord(walletAddr)
		if !ok {
			return 0, false, types.ErrUserRecordNotFound
		}

		previousScore := record.TotalScore
		record.TotalScore += slashRecord.SlashAmount
		restoredScore = slashRecord.SlashAmount

		// Re-check verification status
		params := k.GetParams()
		if record.TotalScore >= params.VerificationThreshold {
			if record.Status == types.VerificationStatusUnverified {
				record.Status = types.VerificationStatusVerified
				record.VerificationAchievedHeight = k.currentHeight
				record.VerificationAchievedAt = k.currentTime
			}
		}

		if err := k.SetUserRecord(record); err != nil {
			return 0, false, err
		}

		// Add to score history
		k.AddScoreChange(types.ScoreChange{
			BlockHeight:   k.currentHeight,
			ScoreDelta:    int64(slashRecord.SlashAmount),
			NewTotal:      record.TotalScore,
			Reason:        types.ChangeReasonAppealReversal,
			RelatedIRID:   slashRecord.RelatedIRID,
			TxHash:        fmt.Sprintf("appeal-resolved-%d", k.currentHeight),
			Timestamp:     k.currentTime,
			PreviousScore: previousScore,
		})

		depositReturned = true
	}

	// Mark slash record as resolved
	slashRecord.Resolved = true

	if err := k.UpdateSlashRecord(walletAddr, slashRecord); err != nil {
		return 0, false, err
	}

	return restoredScore, depositReturned, nil
}

// CalculateSlashAmount calculates the slash amount based on percentage
func (k *Keeper) CalculateSlashAmount(walletAddr string, percentage uint64) (uint64, error) {
	record, ok := k.GetUserRecord(walletAddr)
	if !ok {
		return 0, types.ErrUserRecordNotFound
	}

	params := k.GetParams()

	// Ensure percentage doesn't exceed max
	if percentage > params.SlashPercentage {
		percentage = params.SlashPercentage
	}

	// Calculate slash amount
	slashAmount := (record.TotalScore * percentage) / 100

	return slashAmount, nil
}

// GetPendingAppeals returns all slash records with pending appeals
func (k *Keeper) GetPendingAppeals() []types.SlashRecord {
	k.mu.RLock()
	defer k.mu.RUnlock()

	pending := []types.SlashRecord{}

	for _, records := range k.slashRecords {
		for _, record := range records {
			if record.Appealed && !record.Resolved {
				// Check if appeal is still within review window
				currentTime := time.Now().Unix()
				if currentTime < record.AppealDeadline+(14*24*3600) {
					pending = append(pending, record)
				}
			}
		}
	}

	return pending
}

// IsSlashAppealed checks if a specific slash has been appealed
func (k *Keeper) IsSlashAppealed(walletAddr, slashTxHash string) bool {
	slashRecord, ok := k.GetSlashRecord(walletAddr, slashTxHash)
	if !ok {
		return false
	}
	return slashRecord.Appealed
}

// IsSlashResolved checks if a slash appeal has been resolved
func (k *Keeper) IsSlashResolved(walletAddr, slashTxHash string) bool {
	slashRecord, ok := k.GetSlashRecord(walletAddr, slashTxHash)
	if !ok {
		return false
	}
	return slashRecord.Resolved
}
