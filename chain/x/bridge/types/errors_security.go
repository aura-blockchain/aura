package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Bridge security error codes (100-199 range)
var (
	// Security errors (100-109)
	ErrBridgePaused           = errorsmod.Register(ModuleName, 100, "bridge is paused")
	ErrInvalidMerkleProof     = errorsmod.Register(ModuleName, 101, "invalid Merkle proof")
	ErrInvalidTSSSignature    = errorsmod.Register(ModuleName, 102, "invalid threshold signature")
	ErrInsufficientSignatures = errorsmod.Register(ModuleName, 103, "insufficient validator signatures")
	ErrValidatorNotFound      = errorsmod.Register(ModuleName, 104, "validator not found")
	ErrValidatorNotActive     = errorsmod.Register(ModuleName, 105, "validator is not active")
	ErrInvalidNonce           = errorsmod.Register(ModuleName, 106, "invalid nonce")
	ErrNonceAlreadyUsed       = errorsmod.Register(ModuleName, 107, "nonce already used")

	// Transfer limit errors (110-119)
	ErrAmountBelowMinimum   = errorsmod.Register(ModuleName, 110, "amount below minimum transfer limit")
	ErrAmountExceedsMaximum = errorsmod.Register(ModuleName, 111, "amount exceeds maximum transfer limit")
	ErrDailyLimitExceeded   = errorsmod.Register(ModuleName, 112, "daily withdrawal limit exceeded")
	ErrTimeLockRequired     = errorsmod.Register(ModuleName, 113, "time-lock required for large transfer")
	ErrTimeLockNotExpired   = errorsmod.Register(ModuleName, 114, "time-lock has not expired")
	ErrTimeLockChallenged   = errorsmod.Register(ModuleName, 115, "time-lock has been challenged")

	// Circuit breaker errors (120-129)
	ErrCircuitBreakerOpen     = errorsmod.Register(ModuleName, 120, "circuit breaker is open")
	ErrHourlyVolumeExceeded   = errorsmod.Register(ModuleName, 121, "hourly volume limit exceeded")
	ErrTooManyFailedTransfers = errorsmod.Register(ModuleName, 122, "too many failed transfers")

	// Permission errors (130-139)
	ErrAddressBlacklisted    = errorsmod.Register(ModuleName, 130, "address is blacklisted")
	ErrAddressNotWhitelisted = errorsmod.Register(ModuleName, 131, "address is not whitelisted")
	ErrWhitelistEnabled      = errorsmod.Register(ModuleName, 132, "whitelist is enabled")

	// Fraud proof errors (140-149)
	ErrFraudProofExpired         = errorsmod.Register(ModuleName, 140, "fraud proof has expired")
	ErrFraudProofAlreadyResolved = errorsmod.Register(ModuleName, 141, "fraud proof already resolved")
	ErrFraudProofPending         = errorsmod.Register(ModuleName, 142, "fraud proof already pending review")
	ErrFraudProofNotFound        = errorsmod.Register(ModuleName, 143, "fraud proof not found")
	ErrInvalidEvidence           = errorsmod.Register(ModuleName, 144, "invalid evidence")

	// Insurance fund errors (150-159)
	ErrInsufficientInsuranceFund = errorsmod.Register(ModuleName, 150, "insufficient insurance fund balance")
	ErrClaimNotFound             = errorsmod.Register(ModuleName, 151, "insurance claim not found")
	ErrClaimAlreadyResolved      = errorsmod.Register(ModuleName, 152, "claim already resolved")

	// Validator errors (160-199)
	ErrValidatorSlashed        = errorsmod.Register(ModuleName, 160, "validator has been slashed")
	ErrValidatorJailed         = errorsmod.Register(ModuleName, 161, "validator is jailed")
	ErrRotationNotApproved     = errorsmod.Register(ModuleName, 162, "validator rotation not approved")
	ErrRotationNotEffective    = errorsmod.Register(ModuleName, 163, "validator rotation not yet effective")
	ErrValidatorUnauthorized   = errorsmod.Register(ModuleName, 164, "validator is not authorized")
	ErrNoActiveValidators      = errorsmod.Register(ModuleName, 165, "no active validators available")
	ErrSignatureSetAlreadyUsed = errorsmod.Register(ModuleName, 166, "signature set already used for this transfer")
)
