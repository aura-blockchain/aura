package types

import "errors"

var (
	// Security errors
	ErrBridgePaused           = errors.New("bridge is paused")
	ErrInvalidMerkleProof     = errors.New("invalid Merkle proof")
	ErrInvalidTSSSignature    = errors.New("invalid threshold signature")
	ErrInsufficientSignatures = errors.New("insufficient validator signatures")
	ErrValidatorNotFound      = errors.New("validator not found")
	ErrValidatorNotActive     = errors.New("validator is not active")
	ErrInvalidNonce           = errors.New("invalid nonce")
	ErrNonceAlreadyUsed       = errors.New("nonce already used")

	// Transfer limit errors
	ErrAmountBelowMinimum   = errors.New("amount below minimum transfer limit")
	ErrAmountExceedsMaximum = errors.New("amount exceeds maximum transfer limit")
	ErrDailyLimitExceeded   = errors.New("daily withdrawal limit exceeded")
	ErrTimeLockRequired     = errors.New("time-lock required for large transfer")
	ErrTimeLockNotExpired   = errors.New("time-lock has not expired")
	ErrTimeLockChallenged   = errors.New("time-lock has been challenged")

	// Circuit breaker errors
	ErrCircuitBreakerOpen     = errors.New("circuit breaker is open")
	ErrHourlyVolumeExceeded   = errors.New("hourly volume limit exceeded")
	ErrTooManyFailedTransfers = errors.New("too many failed transfers")

	// Permission errors
	ErrAddressBlacklisted    = errors.New("address is blacklisted")
	ErrAddressNotWhitelisted = errors.New("address is not whitelisted")
	ErrWhitelistEnabled      = errors.New("whitelist is enabled")

	// Fraud proof errors
	ErrFraudProofExpired         = errors.New("fraud proof has expired")
	ErrFraudProofAlreadyResolved = errors.New("fraud proof already resolved")
	ErrFraudProofPending         = errors.New("fraud proof already pending review")
	ErrFraudProofNotFound        = errors.New("fraud proof not found")
	ErrInvalidEvidence           = errors.New("invalid evidence")

	// Insurance fund errors
	ErrInsufficientInsuranceFund = errors.New("insufficient insurance fund balance")
	ErrClaimNotFound             = errors.New("insurance claim not found")
	ErrClaimAlreadyResolved      = errors.New("claim already resolved")

	// Validator errors
	ErrValidatorSlashed     = errors.New("validator has been slashed")
	ErrValidatorJailed      = errors.New("validator is jailed")
	ErrRotationNotApproved  = errors.New("validator rotation not approved")
	ErrRotationNotEffective = errors.New("validator rotation not yet effective")
)
