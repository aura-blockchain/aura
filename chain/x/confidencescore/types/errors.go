package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Confidence score module error codes
var (
	// IR completion errors (1-19)
	ErrIRNotFound         = errorsmod.Register(ModuleName, 1, "inclusion routine not found")
	ErrIRAlreadyCompleted = errorsmod.Register(ModuleName, 2, "ir already completed by this user")
	ErrIRNotActive        = errorsmod.Register(ModuleName, 3, "ir is not in active status")
	ErrIRSuspended        = errorsmod.Register(ModuleName, 4, "ir is suspended")

	// Anchor errors (20-29)
	ErrAnchorNotCompleted = errorsmod.Register(ModuleName, 20, "anchor ir-000 not completed")
	ErrAnchorRequired     = errorsmod.Register(ModuleName, 21, "anchor completion required before other irs")
	ErrInvalidAnchor      = errorsmod.Register(ModuleName, 22, "invalid anchor completion")

	// Prerequisite errors (30-39)
	ErrPrerequisitesNotMet = errorsmod.Register(ModuleName, 30, "ir prerequisites not met")
	ErrPrerequisiteMissing = errorsmod.Register(ModuleName, 31, "required prerequisite ir not completed")

	// Rate limit errors (40-49)
	ErrRateLimitExceeded   = errorsmod.Register(ModuleName, 40, "rate limit exceeded")
	ErrHourlyLimitExceeded = errorsmod.Register(ModuleName, 41, "hourly rate limit exceeded")
	ErrDailyLimitExceeded  = errorsmod.Register(ModuleName, 42, "daily rate limit exceeded")

	// Validation errors (50-69)
	ErrInvalidWalletAddress = errorsmod.Register(ModuleName, 50, "invalid wallet address")
	ErrInvalidIRID          = errorsmod.Register(ModuleName, 51, "invalid ir id")
	ErrInvalidProofHash     = errorsmod.Register(ModuleName, 52, "invalid proof hash: must be 32 bytes")
	ErrInvalidVerifierHash  = errorsmod.Register(ModuleName, 53, "invalid verifier hash: must be 32 bytes")
	ErrInvalidAssistant     = errorsmod.Register(ModuleName, 54, "invalid or inactive assistant")
	ErrInvalidTimestamp     = errorsmod.Register(ModuleName, 55, "invalid timestamp")
	ErrStaleAttestation     = errorsmod.Register(ModuleName, 56, "attestation is too old")

	// Replay protection errors (70-79)
	ErrReplayDetected      = errorsmod.Register(ModuleName, 70, "replay attack detected: proof hash already used")
	ErrDuplicateCompletion = errorsmod.Register(ModuleName, 71, "duplicate completion detected")

	// Score errors (80-89)
	ErrInvalidScore           = errorsmod.Register(ModuleName, 80, "invalid score value")
	ErrScoreCalculationFailed = errorsmod.Register(ModuleName, 81, "score calculation failed")
	ErrScoreOverflow          = errorsmod.Register(ModuleName, 82, "score value overflow")

	// Slash errors (90-109)
	ErrSlashNotFound         = errorsmod.Register(ModuleName, 90, "slash record not found")
	ErrInvalidSlashAmount    = errorsmod.Register(ModuleName, 91, "invalid slash amount")
	ErrSlashAlreadyAppealed  = errorsmod.Register(ModuleName, 92, "slash already appealed")
	ErrAppealExpired         = errorsmod.Register(ModuleName, 93, "appeal deadline has passed")
	ErrAppealAlreadyResolved = errorsmod.Register(ModuleName, 94, "appeal already resolved")
	ErrInsufficientDeposit   = errorsmod.Register(ModuleName, 95, "insufficient appeal deposit")

	// Authorization errors (110-119)
	ErrUnauthorized     = errorsmod.Register(ModuleName, 110, "unauthorized")
	ErrInvalidAuthority = errorsmod.Register(ModuleName, 111, "invalid authority address")

	// State errors (120-139)
	ErrUserRecordNotFound = errorsmod.Register(ModuleName, 120, "user confidence record not found")
	ErrCompletionNotFound = errorsmod.Register(ModuleName, 121, "ir completion not found")
	ErrHistoryNotFound    = errorsmod.Register(ModuleName, 122, "score history not found")

	// Verification errors (140-149)
	ErrNotVerified         = errorsmod.Register(ModuleName, 140, "user is not verified")
	ErrVerificationRevoked = errorsmod.Register(ModuleName, 141, "verification status revoked")
	ErrAlreadyVerified     = errorsmod.Register(ModuleName, 142, "user already verified")

	// Arena errors (150-159)
	ErrInvalidArena  = errorsmod.Register(ModuleName, 150, "invalid arena type")
	ErrArenaNotFound = errorsmod.Register(ModuleName, 151, "arena not found in user record")

	// Parameter errors (160-169)
	ErrInvalidParams     = errorsmod.Register(ModuleName, 160, "invalid module parameters")
	ErrInvalidThreshold  = errorsmod.Register(ModuleName, 161, "invalid threshold value")
	ErrInvalidMultiplier = errorsmod.Register(ModuleName, 162, "invalid multiplier value")

	// Request errors (900-909)
	ErrInvalidRequest = errorsmod.Register(ModuleName, 900, "invalid request")
)
