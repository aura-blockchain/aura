package types

// Event types for the economics module
const (
	// Vesting events
	EventTypeCreateVestingSchedule = "create_vesting_schedule"
	EventTypeReleaseVestedTokens   = "release_vested_tokens"
	EventTypeRevokeVestingSchedule = "revoke_vesting_schedule"

	// Governance proposal events
	EventTypeSubmitProposal  = "submit_proposal"
	EventTypeDeposit         = "proposal_deposit"
	EventTypeVote            = "vote"
	EventTypeDelegateVote    = "delegate_vote"
	EventTypeUndelegateVote  = "undelegate_vote"
	EventTypeExecuteProposal = "execute_proposal"
	EventTypeRevealVote      = "reveal_secret_vote"

	// Vote lock events
	EventTypeLockVotingTokens   = "lock_voting_tokens"
	EventTypeUnlockVotingTokens = "unlock_voting_tokens"

	// Treasury events
	EventTypeProposeTreasurySpend = "propose_treasury_spend"
	EventTypeSignTreasurySpend    = "sign_treasury_spend"
	EventTypeExecuteTreasurySpend = "execute_treasury_spend"

	// Admin events
	EventTypeUpdateParams          = "update_params"
	EventTypeAdjustInflationRate   = "adjust_inflation_rate"
)

// Event attribute keys
const (
	// Common attributes
	AttributeKeyAmount      = "amount"
	AttributeKeyOwner       = "owner"
	AttributeKeyAuthority   = "authority"
	AttributeKeyReason      = "reason"

	// Vesting attributes
	AttributeKeyScheduleID  = "schedule_id"
	AttributeKeyCreator     = "creator"
	AttributeKeyBeneficiary = "beneficiary"
	AttributeKeyRevoker     = "revoker"

	// Governance attributes
	AttributeKeyProposalID    = "proposal_id"
	AttributeKeyProposer      = "proposer"
	AttributeKeyProposalTitle = "proposal_title"
	AttributeKeyDepositor     = "depositor"
	AttributeKeyVoter         = "voter"
	AttributeKeyOption        = "option"
	AttributeKeyDelegator     = "delegator"
	AttributeKeyDelegate      = "delegate"
	AttributeKeyExecutor      = "executor"

	// Vote lock attributes
	AttributeKeyLockID      = "lock_id"
	AttributeKeyVotingPower = "voting_power"

	// Treasury attributes
	AttributeKeyTxID       = "tx_id"
	AttributeKeyRecipient  = "recipient"
	AttributeKeySigner     = "signer"
	AttributeKeySignatures = "signatures"
	AttributeKeySuccess    = "success"

	// Inflation attributes
	AttributeKeyOldRate = "old_rate"
	AttributeKeyNewRate = "new_rate"
)
