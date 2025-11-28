package types

import "fmt"

// Error types for the auth module
var (
	ErrRoleNotFound            = fmt.Errorf("role not found")
	ErrRoleAlreadyExists       = fmt.Errorf("role already exists")
	ErrRoleAssignmentNotFound  = fmt.Errorf("role assignment not found")
	ErrInsufficientPermissions = fmt.Errorf("insufficient permissions")
	ErrInvalidRole             = fmt.Errorf("invalid role")
	ErrInvalidRoleAssignment   = fmt.Errorf("invalid role assignment")

	ErrMultisigWalletNotFound = fmt.Errorf("multisig wallet not found")
	ErrMultisigWalletExists   = fmt.Errorf("multisig wallet already exists")
	ErrInvalidMultisigWallet  = fmt.Errorf("invalid multisig wallet")
	ErrNotWalletSigner        = fmt.Errorf("not a wallet signer")
	ErrAlreadySigned          = fmt.Errorf("already signed this proposal")

	ErrProposalNotFound        = fmt.Errorf("proposal not found")
	ErrProposalExpired         = fmt.Errorf("proposal has expired")
	ErrProposalNotApproved     = fmt.Errorf("proposal not approved")
	ErrProposalAlreadyExecuted = fmt.Errorf("proposal already executed")
	ErrInvalidProposal         = fmt.Errorf("invalid proposal")

	ErrActionNotFound        = fmt.Errorf("time-locked action not found")
	ErrActionNotReady        = fmt.Errorf("action not ready for execution")
	ErrActionAlreadyExecuted = fmt.Errorf("action already executed")
	ErrInvalidAction         = fmt.Errorf("invalid action")

	ErrEmergencyAdminNotFound = fmt.Errorf("emergency admin not found")
	ErrEmergencyAdminInactive = fmt.Errorf("emergency admin is inactive")
	ErrInvalidEmergencyAdmin  = fmt.Errorf("invalid emergency admin")

	ErrValidatorNotFound  = fmt.Errorf("validator not found")
	ErrRotationInProgress = fmt.Errorf("rotation already in progress")
	ErrRotationNotFound   = fmt.Errorf("rotation not found")

	ErrSessionNotFound = fmt.Errorf("session not found")
	ErrSessionExpired  = fmt.Errorf("session has expired")
	ErrSessionInactive = fmt.Errorf("session is inactive")
	ErrInvalidSession  = fmt.Errorf("invalid session")

	ErrRateLimitExceeded      = fmt.Errorf("rate limit exceeded")
	ErrInvalidRateLimitConfig = fmt.Errorf("invalid rate limit config")

	ErrInvalidAddress = fmt.Errorf("invalid address")
)
