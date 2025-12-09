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
	CodeAccountNotFound        uint32 = 100
	CodeAccountAlreadyExists   uint32 = 101
	CodeSessionNotFound        uint32 = 102
	CodeSessionExpired         uint32 = 103
	CodeSessionInactive        uint32 = 104
	CodeInvalidSession         uint32 = 105
	CodeRateLimitExceeded      uint32 = 106
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

	// DID Key Rotation error codes (550-559)
	CodeDIDKeyRotationNotFound    uint32 = 550
	CodeDIDKeyRotationInProgress  uint32 = 551
	CodeInvalidVerificationMethod uint32 = 552
	CodeKeyInGracePeriod          uint32 = 553
	CodeKeyNotValid               uint32 = 554
	CodeInvalidSignature          uint32 = 555

	// Identity change error codes (600-699)
	CodeIdentityNotFound            uint32 = 600
	CodeIdentityAlreadyExists       uint32 = 601
	CodeChangeRequestNotFound       uint32 = 602
	CodeChangeRequestInvalid        uint32 = 603
	CodeChangeRequestExpired        uint32 = 604
	CodeChangeRequestPending        uint32 = 605
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

	// Credential revocation error codes (670-689)
	CodeCredentialRevoked        uint32 = 670
	CodeCredentialNotFound       uint32 = 671
	CodeCredentialAlreadyRevoked uint32 = 672
	CodeInvalidCredentialID      uint32 = 673

	// Attribute Access Control error codes (690-709)
	CodeAttributeNotFound  uint32 = 690
	CodeAccessDenied       uint32 = 691
	CodeAccessExpired      uint32 = 692
	CodeInvalidPermission  uint32 = 693
	CodePermissionNotFound uint32 = 694
	CodeInvalidAccessLevel uint32 = 695

	// ZK Proof error codes (710-729)
	CodeInvalidProof              uint32 = 710
	CodeProofVerificationFailed   uint32 = 711
	CodeInvalidVerifyingKey       uint32 = 712
	CodeInvalidPublicInputs       uint32 = 713
	CodeUnsupportedProofType      uint32 = 714
	CodeProofDeserializationError uint32 = 715

	// Serialization error codes (800-819)
	CodeMarshalFailed   uint32 = 800
	CodeUnmarshalFailed uint32 = 801

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
	ErrAccountNotFound        = errors.Register(ModuleName, CodeAccountNotFound, "account not found")
	ErrAccountAlreadyExists   = errors.Register(ModuleName, CodeAccountAlreadyExists, "account already exists")
	ErrSessionNotFound        = errors.Register(ModuleName, CodeSessionNotFound, "session not found")
	ErrSessionExpired         = errors.Register(ModuleName, CodeSessionExpired, "session has expired")
	ErrSessionInactive        = errors.Register(ModuleName, CodeSessionInactive, "session is inactive")
	ErrInvalidSession         = errors.Register(ModuleName, CodeInvalidSession, "invalid session")
	ErrRateLimitExceeded      = errors.Register(ModuleName, CodeRateLimitExceeded, "rate limit exceeded")
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

// DID Key Rotation errors
var (
	ErrDIDKeyRotationNotFound    = errors.Register(ModuleName, CodeDIDKeyRotationNotFound, "DID key rotation not found")
	ErrDIDKeyRotationInProgress  = errors.Register(ModuleName, CodeDIDKeyRotationInProgress, "DID key rotation already in progress")
	ErrInvalidVerificationMethod = errors.Register(ModuleName, CodeInvalidVerificationMethod, "invalid verification method")
	ErrKeyInGracePeriod          = errors.Register(ModuleName, CodeKeyInGracePeriod, "old key still in grace period")
	ErrKeyNotValid               = errors.Register(ModuleName, CodeKeyNotValid, "key is not valid")
	ErrInvalidSignature          = errors.Register(ModuleName, CodeInvalidSignature, "invalid signature")
)

// Identity change errors
var (
	ErrIdentityNotFound            = errors.Register(ModuleName, CodeIdentityNotFound, "identity not found")
	ErrIdentityAlreadyExists       = errors.Register(ModuleName, CodeIdentityAlreadyExists, "identity already exists")
	ErrChangeRequestNotFound       = errors.Register(ModuleName, CodeChangeRequestNotFound, "change request not found")
	ErrChangeRequestInvalid        = errors.Register(ModuleName, CodeChangeRequestInvalid, "change request is invalid")
	ErrChangeRequestExpired        = errors.Register(ModuleName, CodeChangeRequestExpired, "change request has expired")
	ErrChangeRequestPending        = errors.Register(ModuleName, CodeChangeRequestPending, "change request is pending")
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

// Credential revocation errors
var (
	ErrCredentialRevoked        = errors.Register(ModuleName, CodeCredentialRevoked, "credential has been revoked")
	ErrCredentialNotFound       = errors.Register(ModuleName, CodeCredentialNotFound, "credential not found")
	ErrCredentialAlreadyRevoked = errors.Register(ModuleName, CodeCredentialAlreadyRevoked, "credential already revoked")
	ErrInvalidCredentialID      = errors.Register(ModuleName, CodeInvalidCredentialID, "invalid credential ID")
)

// Attribute Access Control errors
var (
	ErrAttributeNotFound  = errors.Register(ModuleName, CodeAttributeNotFound, "attribute not found")
	ErrAccessDenied       = errors.Register(ModuleName, CodeAccessDenied, "access denied")
	ErrAccessExpired      = errors.Register(ModuleName, CodeAccessExpired, "access permission expired")
	ErrInvalidPermission  = errors.Register(ModuleName, CodeInvalidPermission, "invalid permission")
	ErrPermissionNotFound = errors.Register(ModuleName, CodePermissionNotFound, "permission not found")
	ErrInvalidAccessLevel = errors.Register(ModuleName, CodeInvalidAccessLevel, "invalid access level")
)

// ZK Proof errors
var (
	ErrInvalidProof              = errors.Register(ModuleName, CodeInvalidProof, "invalid zero-knowledge proof")
	ErrProofVerificationFailed   = errors.Register(ModuleName, CodeProofVerificationFailed, "proof verification failed")
	ErrInvalidVerifyingKey       = errors.Register(ModuleName, CodeInvalidVerifyingKey, "invalid verification key")
	ErrInvalidPublicInputs       = errors.Register(ModuleName, CodeInvalidPublicInputs, "invalid public inputs")
	ErrUnsupportedProofType      = errors.Register(ModuleName, CodeUnsupportedProofType, "unsupported proof type")
	ErrProofDeserializationError = errors.Register(ModuleName, CodeProofDeserializationError, "proof deserialization failed")
)

// Serialization errors
var (
	ErrMarshalFailed   = errors.Register(ModuleName, CodeMarshalFailed, "failed to marshal data")
	ErrUnmarshalFailed = errors.Register(ModuleName, CodeUnmarshalFailed, "failed to unmarshal data")
)

// General errors
var (
	ErrInvalidAddress = errors.Register(ModuleName, CodeInvalidAddress, "invalid address")
	ErrInvalidInput   = errors.Register(ModuleName, CodeInvalidInput, "invalid input")
	ErrInternal       = errors.Register(ModuleName, CodeInternal, "internal error")
)
