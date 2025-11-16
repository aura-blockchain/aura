package types

// Event types for the confidencescore module
const (
	EventTypeIRCompleted          = "ir_completed"
	EventTypeVerificationAchieved = "verification_achieved"
	EventTypeArenaFocusAchieved   = "arena_focus_achieved"
	EventTypeScoreSlashed         = "score_slashed"
	EventTypeJackpotTriggered     = "jackpot_triggered"
	EventTypeAppealFiled          = "appeal_filed"
	EventTypeAppealResolved       = "appeal_resolved"
)

// Event attribute keys
const (
	AttributeKeyWalletAddress       = "wallet_address"
	AttributeKeyIRID                = "ir_id"
	AttributeKeyScoreEarned         = "score_earned"
	AttributeKeyNewTotalScore       = "new_total_score"
	AttributeKeyAssistantAddress    = "assistant_address"
	AttributeKeyArena               = "arena"
	AttributeKeyBlockHeight         = "block_height"
	AttributeKeyVelocityMultiplier  = "velocity_multiplier"
	AttributeKeyArenaMultiplier     = "arena_multiplier"
	AttributeKeyJackpotMultiplier   = "jackpot_multiplier"
	AttributeKeyFinalScore          = "final_score"
	AttributeKeyIRCount             = "ir_count"
	AttributeKeyDaysSinceAnchor     = "days_since_anchor"
	AttributeKeyArenaScore          = "arena_score"
	AttributeKeySlashAmount         = "slash_amount"
	AttributeKeyNewScore            = "new_score"
	AttributeKeyReason              = "reason"
	AttributeKeyVerificationRevoked = "verification_revoked"
	AttributeKeySlashTxHash         = "slash_tx_hash"
	AttributeKeyBonusScore          = "bonus_score"
	AttributeKeyDeposit             = "deposit"
	AttributeKeyReviewDeadline      = "review_deadline"
	AttributeKeyScoreRestored       = "score_restored"
	AttributeKeyRestoredAmount      = "restored_amount"
	AttributeKeyDepositReturned     = "deposit_returned"
)
