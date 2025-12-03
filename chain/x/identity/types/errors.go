package types

import (
	"cosmossdk.io/errors"
)

// Identity module error codes
const (
	// Auth-related error codes (1-99)
	CodeRoleNotFound            uint32 = 1
	CodeRoleAlreadyExists       uint32 = 2
	CodeInvalidRole             uint32 = 3
	CodeInsufficientPermissions uint32 = 4
	CodePermissionDenied        uint32 = 5
	CodeInvalidRoleAssignment   uint32 = 6

	// Account and Session error codes (100-199)
	CodeAccountNotFound       uint32 = 100
	CodeAccountAlreadyExists  uint32 = 101
	CodeSessionNotFound       uint32 = 102
	CodeSessionExpired        uint32 = 103
	CodeSessionInactive       uint32 = 104
	CodeInvalidSession        uint32 = 105
	CodeRateLimitExceeded     uint32 = 106
	CodeInvalidRateLimitConfig uint32 = 107

	// Multisig error codes (200-299)
	CodeMultisigWalletNotFound  uint32 = 200
	CodeMultisigWalletExists    uint32 = 201
	CodeInvalidMultisigWallet   uint32 = 202
	CodeNotWalletSigner         uint32 = 203
	CodeAlreadySigned           uint32 = 204
	CodeProposalNotFound        uint32 = 205
	CodeProposalExpired         uint32 = 206
	CodeProposalNotApproved     uint32 = 207
	CodeProposalAlreadyExecuted uint32 = 208
	CodeInvalidProposal         uint32 = 209

	// Time-locked action error codes (300-399)
	CodeActionNotFound        uint32 = 300
	CodeActionNotReady        uint32 = 301
	CodeActionAlreadyExecuted uint32 = 302
	CodeInvalidAction         uint32 = 303

	// Emergency admin error codes (400-499)
	CodeEmergencyAdminNotFound uint32 = 400
	CodeEmergencyAdminInactive uint32 = 401
	CodeInvalidEmergencyAdmin  uint32 = 402

	// Validator error codes (500-599)
	CodeValidatorNotFound  uint32 = 500
	CodeRotationInProgress uint32 = 501
	CodeRotationNotFound   uint32 = 502

	// Identity change error codes (600-699)
	CodeIdentityNotFound         uint32 = 600
	CodeIdentityAlreadyExists    uint32 = 601
	CodeChangeRequestNotFound    uint32 = 602
	CodeChangeRequestInvalid     uint32 = 603
	CodeChangeRequestExpired     uint32 = 604
	CodeChangeRequestPending     uint32 = 605
	CodeChangeRequestAlreadyApplied uint32 = 606
	CodeChangeRequestLimitExceeded  uint32 = 607
	CodeIdentityChangeSuspended     uint32 = 608
	CodeInvalidDID                  uint32 = 609
	CodeInvalidChangeRequest        uint32 = 610

	// GDPR-related error codes (650-669)
	CodeIdentityAlreadyErased uint32 = 650
	CodeIdentityErased        uint32 = 651
	CodeNoCommitment          uint32 = 652
	CodeInvalidCommitment     uint32 = 653
	CodeUnauthorized          uint32 = 654

	// General error codes (900-999)
	CodeInvalidAddress uint32 = 900
	CodeInvalidInput   uint32 = 901
	CodeInternal       uint32 = 999
)

// Auth-related errors
var (
	ErrRoleNotFound            = errors.Register(ModuleName, CodeRoleNotFound, "role not found")
	ErrRoleAlreadyExists       = errors.Register(ModuleName, CodeRoleAlreadyExists, "role already exists")
	ErrInvalidRole             = errors.Register(ModuleName, CodeInvalidRole, "invalid role")
	ErrInsufficientPermissions = errors.Register(ModuleName, CodeInsufficientPermissions, "insufficient permissions")
	ErrPermissionDenied        = errors.Register(ModuleName, CodePermissionDenied, "permission denied")
	ErrInvalidRoleAssignment   = errors.Register(ModuleName, CodeInvalidRoleAssignment, "invalid role assignment")
)

// Account and Session errors
var (
	ErrAccountNotFound       = errors.Register(ModuleName, CodeAccountNotFound, "account not found")
	ErrAccountAlreadyExists  = errors.Register(ModuleName, CodeAccountAlreadyExists, "account already exists")
	ErrSessionNotFound       = errors.Register(ModuleName, CodeSessionNotFound, "session not found")
	ErrSessionExpired        = errors.Register(ModuleName, CodeSessionExpired, "session has expired")
	ErrSessionInactive       = errors.Register(ModuleName, CodeSessionInactive, "session is inactive")
	ErrInvalidSession        = errors.Register(ModuleName, CodeInvalidSession, "invalid session")
	ErrRateLimitExceeded     = errors.Register(ModuleName, CodeRateLimitExceeded, "rate limit exceeded")
	ErrInvalidRateLimitConfig = errors.Register(ModuleName, CodeInvalidRateLimitConfig, "invalid rate limit config")
)

// Multisig errors
var (
	ErrMultisigWalletNotFound  = errors.Register(ModuleName, CodeMultisigWalletNotFound, "multisig wallet not found")
	ErrMultisigWalletExists    = errors.Register(ModuleName, CodeMultisigWalletExists, "multisig wallet already exists")
	ErrInvalidMultisigWallet   = errors.Register(ModuleName, CodeInvalidMultisigWallet, "invalid multisig wallet")
	ErrNotWalletSigner         = errors.Register(ModuleName, CodeNotWalletSigner, "not a wallet signer")
	ErrAlreadySigned           = errors.Register(ModuleName, CodeAlreadySigned, "already signed this proposal")
	ErrProposalNotFound        = errors.Register(ModuleName, CodeProposalNotFound, "proposal not found")
	ErrProposalExpired         = errors.Register(ModuleName, CodeProposalExpired, "proposal has expired")
	ErrProposalNotApproved     = errors.Register(ModuleName, CodeProposalNotApproved, "proposal not approved")
	ErrProposalAlreadyExecuted = errors.Register(ModuleName, CodeProposalAlreadyExecuted, "proposal already executed")
	ErrInvalidProposal         = errors.Register(ModuleName, CodeInvalidProposal, "invalid proposal")
)

// Time-locked action errors
var (
	ErrActionNotFound        = errors.Register(ModuleName, CodeActionNotFound, "time-locked action not found")
	ErrActionNotReady        = errors.Register(ModuleName, CodeActionNotReady, "action not ready for execution")
	ErrActionAlreadyExecuted = errors.Register(ModuleName, CodeActionAlreadyExecuted, "action already executed")
	ErrInvalidAction         = errors.Register(ModuleName, CodeInvalidAction, "invalid action")
)

// Emergency admin errors
var (
	ErrEmergencyAdminNotFound = errors.Register(ModuleName, CodeEmergencyAdminNotFound, "emergency admin not found")
	ErrEmergencyAdminInactive = errors.Register(ModuleName, CodeEmergencyAdminInactive, "emergency admin is inactive")
	ErrInvalidEmergencyAdmin  = errors.Register(ModuleName, CodeInvalidEmergencyAdmin, "invalid emergency admin")
)

// Validator errors
var (
	ErrValidatorNotFound  = errors.Register(ModuleName, CodeValidatorNotFound, "validator not found")
	ErrRotationInProgress = errors.Register(ModuleName, CodeRotationInProgress, "rotation already in progress")
	ErrRotationNotFound   = errors.Register(ModuleName, CodeRotationNotFound, "rotation not found")
)

// Identity change errors
var (
	ErrIdentityNotFound         = errors.Register(ModuleName, CodeIdentityNotFound, "identity not found")
	ErrIdentityAlreadyExists    = errors.Register(ModuleName, CodeIdentityAlreadyExists, "identity already exists")
	ErrChangeRequestNotFound    = errors.Register(ModuleName, CodeChangeRequestNotFound, "change request not found")
	ErrChangeRequestInvalid     = errors.Register(ModuleName, CodeChangeRequestInvalid, "change request is invalid")
	ErrChangeRequestExpired     = errors.Register(ModuleName, CodeChangeRequestExpired, "change request has expired")
	ErrChangeRequestPending     = errors.Register(ModuleName, CodeChangeRequestPending, "change request is pending")
	ErrChangeRequestAlreadyApplied = errors.Register(ModuleName, CodeChangeRequestAlreadyApplied, "change request already applied")
	ErrChangeRequestLimitExceeded  = errors.Register(ModuleName, CodeChangeRequestLimitExceeded, "change request limit exceeded")
	ErrIdentityChangeSuspended     = errors.Register(ModuleName, CodeIdentityChangeSuspended, "identity changes are suspended")
	ErrInvalidDID                  = errors.Register(ModuleName, CodeInvalidDID, "invalid DID")
	ErrInvalidChangeRequest        = errors.Register(ModuleName, CodeInvalidChangeRequest, "invalid change request")
)

// GDPR-related errors
var (
	ErrIdentityAlreadyErased = errors.Register(ModuleName, CodeIdentityAlreadyErased, "identity already erased")
	ErrIdentityErased        = errors.Register(ModuleName, CodeIdentityErased, "identity has been erased")
	ErrNoCommitment          = errors.Register(ModuleName, CodeNoCommitment, "no PII commitment found")
	ErrInvalidCommitment     = errors.Register(ModuleName, CodeInvalidCommitment, "invalid PII commitment")
	ErrUnauthorized          = errors.Register(ModuleName, CodeUnauthorized, "unauthorized action")
)

// General errors
var (
	ErrInvalidAddress = errors.Register(ModuleName, CodeInvalidAddress, "invalid address")
	ErrInvalidInput   = errors.Register(ModuleName, CodeInvalidInput, "invalid input")
	ErrInternal       = errors.Register(ModuleName, CodeInternal, "internal error")
)
