package types

import "errors"

// Module-specific errors for the confidencescore module
var (
	// IR completion errors
	ErrIRNotFound         = errors.New("inclusion routine not found")
	ErrIRAlreadyCompleted = errors.New("ir already completed by this user")
	ErrIRNotActive        = errors.New("ir is not in active status")
	ErrIRSuspended        = errors.New("ir is suspended")

	// Anchor errors
	ErrAnchorNotCompleted = errors.New("anchor ir-000 not completed")
	ErrAnchorRequired     = errors.New("anchor completion required before other irs")
	ErrInvalidAnchor      = errors.New("invalid anchor completion")

	// Prerequisite errors
	ErrPrerequisitesNotMet = errors.New("ir prerequisites not met")
	ErrPrerequisiteMissing = errors.New("required prerequisite ir not completed")

	// Rate limit errors
	ErrRateLimitExceeded   = errors.New("rate limit exceeded")
	ErrHourlyLimitExceeded = errors.New("hourly rate limit exceeded")
	ErrDailyLimitExceeded  = errors.New("daily rate limit exceeded")

	// Validation errors
	ErrInvalidWalletAddress = errors.New("invalid wallet address")
	ErrInvalidIRID          = errors.New("invalid ir id")
	ErrInvalidProofHash     = errors.New("invalid proof hash: must be 32 bytes")
	ErrInvalidVerifierHash  = errors.New("invalid verifier hash: must be 32 bytes")
	ErrInvalidAssistant     = errors.New("invalid or inactive assistant")
	ErrInvalidTimestamp     = errors.New("invalid timestamp")
	ErrStaleAttestation     = errors.New("attestation is too old")

	// Replay protection errors
	ErrReplayDetected      = errors.New("replay attack detected: proof hash already used")
	ErrDuplicateCompletion = errors.New("duplicate completion detected")

	// Score errors
	ErrInvalidScore           = errors.New("invalid score value")
	ErrScoreCalculationFailed = errors.New("score calculation failed")
	ErrScoreOverflow          = errors.New("score value overflow")

	// Slash errors
	ErrSlashNotFound         = errors.New("slash record not found")
	ErrInvalidSlashAmount    = errors.New("invalid slash amount")
	ErrSlashAlreadyAppealed  = errors.New("slash already appealed")
	ErrAppealExpired         = errors.New("appeal deadline has passed")
	ErrAppealAlreadyResolved = errors.New("appeal already resolved")
	ErrInsufficientDeposit   = errors.New("insufficient appeal deposit")

	// Authorization errors
	ErrUnauthorized     = errors.New("unauthorized")
	ErrInvalidAuthority = errors.New("invalid authority address")

	// State errors
	ErrUserRecordNotFound = errors.New("user confidence record not found")
	ErrCompletionNotFound = errors.New("ir completion not found")
	ErrHistoryNotFound    = errors.New("score history not found")

	// Verification errors
	ErrNotVerified         = errors.New("user is not verified")
	ErrVerificationRevoked = errors.New("verification status revoked")
	ErrAlreadyVerified     = errors.New("user already verified")

	// Arena errors
	ErrInvalidArena  = errors.New("invalid arena type")
	ErrArenaNotFound = errors.New("arena not found in user record")

	// Parameter errors
	ErrInvalidParams     = errors.New("invalid module parameters")
	ErrInvalidThreshold  = errors.New("invalid threshold value")
	ErrInvalidMultiplier = errors.New("invalid multiplier value")
)
