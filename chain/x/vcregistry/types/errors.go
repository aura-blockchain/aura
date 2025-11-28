package types

import "errors"

// VC Registry error definitions
var (
	// VC Errors
	ErrVCNotFound       = errors.New("VC not found")
	ErrVCAlreadyRevoked = errors.New("VC already revoked")
	ErrVCExpired        = errors.New("VC has expired")
	ErrVCSuspended      = errors.New("VC is suspended")
	ErrInvalidVCID      = errors.New("invalid VC ID")
	ErrInvalidVCType    = errors.New("invalid VC type")

	// Holder Errors
	ErrInvalidHolderAddress = errors.New("invalid holder address")
	ErrMaxVCsExceeded       = errors.New("maximum VCs per user exceeded")

	// DID Errors
	ErrInvalidDID        = errors.New("invalid DID format")
	ErrDIDNotFound       = errors.New("DID not found")
	ErrDIDAlreadyExists  = errors.New("DID already exists")
	ErrInvalidController = errors.New("invalid DID controller")

	// Policy Errors
	ErrPolicyNotFound     = errors.New("policy not found")
	ErrPolicyInactive     = errors.New("policy is not active")
	ErrPolicyDeprecated   = errors.New("policy has been deprecated")
	ErrInvalidPolicyParam = errors.New("invalid policy parameter")

	// Eligibility Errors
	ErrInsufficientCS         = errors.New("insufficient confidence score")
	ErrMissingRequiredIR      = errors.New("missing required inclusion routine")
	ErrInsufficientArenaScore = errors.New("insufficient arena score")
	ErrAnchorNotCompleted     = errors.New("anchor IR not completed")

	// Rate Limiting Errors
	ErrRateLimitExceeded  = errors.New("minting rate limit exceeded")
	ErrHourlyLimitReached = errors.New("hourly minting limit reached")
	ErrDailyLimitReached  = errors.New("daily minting limit reached")

	// Singleton Errors
	ErrSingletonViolation = errors.New("singleton VC already exists for this user")

	// Authorization Errors
	ErrUnauthorized  = errors.New("unauthorized operation")
	ErrNotVCHolder   = errors.New("signer is not the VC holder")
	ErrNotGovernance = errors.New("operation requires governance authority")

	// Revocation Errors
	ErrInvalidRevocationReason = errors.New("invalid revocation reason")
	ErrCannotRevokeExpired     = errors.New("cannot revoke already expired VC")

	// General Errors
	ErrInvalidInput   = errors.New("invalid input")
	ErrInternalError  = errors.New("internal error")
	ErrCSKeeperNotSet = errors.New("confidence score keeper not set")
	ErrInvalidRequest = errors.New("invalid request")

	// Presentation Errors
	ErrPresentationNotFound    = errors.New("presentation not found")
	ErrPresentationExpired     = errors.New("presentation has expired")
	ErrInvalidPresentationID   = errors.New("invalid presentation ID")
	ErrInvalidQRCodeData       = errors.New("invalid QR code data")
	ErrInvalidSignature        = errors.New("invalid presentation signature")
	ErrNonceAlreadyUsed        = errors.New("nonce has already been used")
	ErrInvalidNonce            = errors.New("invalid nonce")
	ErrPresentationNotYetValid = errors.New("presentation not yet valid")
	ErrEmptyVCList             = errors.New("VC list cannot be empty")
	ErrInvalidExpirationTime   = errors.New("invalid expiration time")
)
