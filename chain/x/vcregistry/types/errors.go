package types

import (
	errorsmod "cosmossdk.io/errors"
)

// VC Registry error definitions
var (
	// VC Errors
	ErrVCNotFound       = errorsmod.Register(ModuleName, 1, "VC not found")
	ErrVCAlreadyRevoked = errorsmod.Register(ModuleName, 2, "VC already revoked")
	ErrVCExpired        = errorsmod.Register(ModuleName, 3, "VC has expired")
	ErrVCSuspended      = errorsmod.Register(ModuleName, 4, "VC is suspended")
	ErrInvalidVCID      = errorsmod.Register(ModuleName, 5, "invalid VC ID")
	ErrInvalidVCType    = errorsmod.Register(ModuleName, 6, "invalid VC type")

	// Holder Errors
	ErrInvalidHolderAddress = errorsmod.Register(ModuleName, 10, "invalid holder address")
	ErrMaxVCsExceeded       = errorsmod.Register(ModuleName, 11, "maximum VCs per user exceeded")

	// DID Errors
	ErrInvalidDID        = errorsmod.Register(ModuleName, 20, "invalid DID format")
	ErrDIDNotFound       = errorsmod.Register(ModuleName, 21, "DID not found")
	ErrDIDAlreadyExists  = errorsmod.Register(ModuleName, 22, "DID already exists")
	ErrInvalidController = errorsmod.Register(ModuleName, 23, "invalid DID controller")

	// Policy Errors
	ErrPolicyNotFound     = errorsmod.Register(ModuleName, 30, "policy not found")
	ErrPolicyInactive     = errorsmod.Register(ModuleName, 31, "policy is not active")
	ErrPolicyDeprecated   = errorsmod.Register(ModuleName, 32, "policy has been deprecated")
	ErrInvalidPolicyParam = errorsmod.Register(ModuleName, 33, "invalid policy parameter")

	// Eligibility Errors
	ErrInsufficientCS         = errorsmod.Register(ModuleName, 40, "insufficient confidence score")
	ErrMissingRequiredIR      = errorsmod.Register(ModuleName, 41, "missing required inclusion routine")
	ErrInsufficientArenaScore = errorsmod.Register(ModuleName, 42, "insufficient arena score")
	ErrAnchorNotCompleted     = errorsmod.Register(ModuleName, 43, "anchor IR not completed")

	// Rate Limiting Errors
	ErrRateLimitExceeded  = errorsmod.Register(ModuleName, 50, "minting rate limit exceeded")
	ErrHourlyLimitReached = errorsmod.Register(ModuleName, 51, "hourly minting limit reached")
	ErrDailyLimitReached  = errorsmod.Register(ModuleName, 52, "daily minting limit reached")

	// Singleton Errors
	ErrSingletonViolation = errorsmod.Register(ModuleName, 60, "singleton VC already exists for this user")

	// Authorization Errors
	ErrUnauthorized  = errorsmod.Register(ModuleName, 70, "unauthorized operation")
	ErrNotVCHolder   = errorsmod.Register(ModuleName, 71, "signer is not the VC holder")
	ErrNotGovernance = errorsmod.Register(ModuleName, 72, "operation requires governance authority")

	// Revocation Errors
	ErrInvalidRevocationReason = errorsmod.Register(ModuleName, 80, "invalid revocation reason")
	ErrCannotRevokeExpired     = errorsmod.Register(ModuleName, 81, "cannot revoke already expired VC")

	// General Errors
	ErrInvalidInput   = errorsmod.Register(ModuleName, 90, "invalid input")
	ErrInternalError  = errorsmod.Register(ModuleName, 91, "internal error")
	ErrCSKeeperNotSet = errorsmod.Register(ModuleName, 92, "confidence score keeper not set")
	ErrInvalidRequest = errorsmod.Register(ModuleName, 93, "invalid request")

	// Presentation Errors
	ErrPresentationNotFound    = errorsmod.Register(ModuleName, 100, "presentation not found")
	ErrPresentationExpired     = errorsmod.Register(ModuleName, 101, "presentation has expired")
	ErrInvalidPresentationID   = errorsmod.Register(ModuleName, 102, "invalid presentation ID")
	ErrInvalidQRCodeData       = errorsmod.Register(ModuleName, 103, "invalid QR code data")
	ErrInvalidSignature        = errorsmod.Register(ModuleName, 104, "invalid presentation signature")
	ErrNonceAlreadyUsed        = errorsmod.Register(ModuleName, 105, "nonce has already been used")
	ErrInvalidNonce            = errorsmod.Register(ModuleName, 106, "invalid nonce")
	ErrPresentationNotYetValid = errorsmod.Register(ModuleName, 107, "presentation not yet valid")
	ErrEmptyVCList             = errorsmod.Register(ModuleName, 108, "VC list cannot be empty")
	ErrInvalidExpirationTime   = errorsmod.Register(ModuleName, 109, "invalid expiration time")
)
