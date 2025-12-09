package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Auth module error codes
var (
	// Role errors (1-9)
	ErrRoleNotFound            = errorsmod.Register(ModuleName, 1, "role not found")
	ErrRoleAlreadyExists       = errorsmod.Register(ModuleName, 2, "role already exists")
	ErrRoleAssignmentNotFound  = errorsmod.Register(ModuleName, 3, "role assignment not found")
	ErrInsufficientPermissions = errorsmod.Register(ModuleName, 4, "insufficient permissions")
	ErrInvalidRole             = errorsmod.Register(ModuleName, 5, "invalid role")
	ErrInvalidRoleAssignment   = errorsmod.Register(ModuleName, 6, "invalid role assignment")

	// Multisig wallet errors (10-19)
	ErrMultisigWalletNotFound = errorsmod.Register(ModuleName, 10, "multisig wallet not found")
	ErrMultisigWalletExists   = errorsmod.Register(ModuleName, 11, "multisig wallet already exists")
	ErrInvalidMultisigWallet  = errorsmod.Register(ModuleName, 12, "invalid multisig wallet")
	ErrNotWalletSigner        = errorsmod.Register(ModuleName, 13, "not a wallet signer")
	ErrAlreadySigned          = errorsmod.Register(ModuleName, 14, "already signed this proposal")

	// Proposal errors (20-29)
	ErrProposalNotFound        = errorsmod.Register(ModuleName, 20, "proposal not found")
	ErrProposalExpired         = errorsmod.Register(ModuleName, 21, "proposal has expired")
	ErrProposalNotApproved     = errorsmod.Register(ModuleName, 22, "proposal not approved")
	ErrProposalAlreadyExecuted = errorsmod.Register(ModuleName, 23, "proposal already executed")
	ErrInvalidProposal         = errorsmod.Register(ModuleName, 24, "invalid proposal")

	// Time-locked action errors (30-39)
	ErrActionNotFound        = errorsmod.Register(ModuleName, 30, "time-locked action not found")
	ErrActionNotReady        = errorsmod.Register(ModuleName, 31, "action not ready for execution")
	ErrActionAlreadyExecuted = errorsmod.Register(ModuleName, 32, "action already executed")
	ErrInvalidAction         = errorsmod.Register(ModuleName, 33, "invalid action")

	// Emergency admin errors (40-49)
	ErrEmergencyAdminNotFound = errorsmod.Register(ModuleName, 40, "emergency admin not found")
	ErrEmergencyAdminInactive = errorsmod.Register(ModuleName, 41, "emergency admin is inactive")
	ErrInvalidEmergencyAdmin  = errorsmod.Register(ModuleName, 42, "invalid emergency admin")

	// Validator errors (50-59)
	ErrValidatorNotFound  = errorsmod.Register(ModuleName, 50, "validator not found")
	ErrRotationInProgress = errorsmod.Register(ModuleName, 51, "rotation already in progress")
	ErrRotationNotFound   = errorsmod.Register(ModuleName, 52, "rotation not found")

	// Session errors (60-69)
	ErrSessionNotFound = errorsmod.Register(ModuleName, 60, "session not found")
	ErrSessionExpired  = errorsmod.Register(ModuleName, 61, "session has expired")
	ErrSessionInactive = errorsmod.Register(ModuleName, 62, "session is inactive")
	ErrInvalidSession  = errorsmod.Register(ModuleName, 63, "invalid session")

	// Rate limit errors (70-79)
	ErrRateLimitExceeded      = errorsmod.Register(ModuleName, 70, "rate limit exceeded")
	ErrInvalidRateLimitConfig = errorsmod.Register(ModuleName, 71, "invalid rate limit config")

	// General errors (900-909)
	ErrInvalidAddress = errorsmod.Register(ModuleName, 900, "invalid address")
)
