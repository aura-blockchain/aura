package types

import (
	"cosmossdk.io/errors"
)

// Identity module error codes (100-199 range)
const (
	// Auth-related error codes (100-109)
	CodeRoleNotFound            uint32 = 100
	CodeRoleAlreadyExists       uint32 = 101
	CodeInvalidRole             uint32 = 102
	CodeInsufficientPermissions uint32 = 103
	CodePermissionDenied        uint32 = 104
	CodeInvalidRoleAssignment   uint32 = 105

	// Account and Session error codes (110-119)
	CodeAccountNotFound        uint32 = 110
	CodeAccountAlreadyExists   uint32 = 111
	CodeSessionNotFound        uint32 = 112
	CodeSessionExpired         uint32 = 113
	CodeSessionInactive        uint32 = 114
	CodeInvalidSession         uint32 = 115
	CodeRateLimitExceeded      uint32 = 116
	CodeInvalidRateLimitConfig uint32 = 117

	// Multisig error codes (120-129)
	CodeMultisigWalletNotFound  uint32 = 120
	CodeMultisigWalletExists    uint32 = 121
	CodeInvalidMultisigWallet   uint32 = 122
	CodeNotWalletSigner         uint32 = 123
	CodeAlreadySigned           uint32 = 124
	CodeProposalNotFound        uint32 = 125
	CodeProposalExpired         uint32 = 126
	CodeProposalNotApproved     uint32 = 127
	CodeProposalAlreadyExecuted uint32 = 128
	CodeInvalidProposal         uint32 = 129

	// Time-locked action error codes (130-134)
	CodeActionNotFound        uint32 = 130
	CodeActionNotReady        uint32 = 131
	CodeActionAlreadyExecuted uint32 = 132
	CodeInvalidAction         uint32 = 133

	// Emergency admin error codes (135-137)
	CodeEmergencyAdminNotFound uint32 = 135
	CodeEmergencyAdminInactive uint32 = 136
	CodeInvalidEmergencyAdmin  uint32 = 137

	// Validator error codes (138-140)
	CodeValidatorNotFound  uint32 = 138
	CodeRotationInProgress uint32 = 139
	CodeRotationNotFound   uint32 = 140

	// DID Key Rotation error codes (141-146)
	CodeDIDKeyRotationNotFound    uint32 = 141
	CodeDIDKeyRotationInProgress  uint32 = 142
	CodeInvalidVerificationMethod uint32 = 143
	CodeKeyInGracePeriod          uint32 = 144
	CodeKeyNotValid               uint32 = 145
	CodeInvalidSignature          uint32 = 146

	// Identity change error codes (147-157)
	CodeIdentityNotFound            uint32 = 147
	CodeIdentityAlreadyExists       uint32 = 148
	CodeChangeRequestNotFound       uint32 = 149
	CodeChangeRequestInvalid        uint32 = 150
	CodeChangeRequestExpired        uint32 = 151
	CodeChangeRequestPending        uint32 = 152
	CodeChangeRequestAlreadyApplied uint32 = 153
	CodeChangeRequestLimitExceeded  uint32 = 154
	CodeIdentityChangeSuspended     uint32 = 155
	CodeInvalidDID                  uint32 = 156
	CodeInvalidChangeRequest        uint32 = 157

	// GDPR-related error codes (158-162)
	CodeIdentityAlreadyErased uint32 = 158
	CodeIdentityErased        uint32 = 159
	CodeNoCommitment          uint32 = 160
	CodeInvalidCommitment     uint32 = 161
	CodeUnauthorized          uint32 = 162

	// Credential revocation error codes (163-166)
	CodeCredentialRevoked        uint32 = 163
	CodeCredentialNotFound       uint32 = 164
	CodeCredentialAlreadyRevoked uint32 = 165
	CodeInvalidCredentialID      uint32 = 166

	// Attribute Access Control error codes (167-172)
	CodeAttributeNotFound  uint32 = 167
	CodeAccessDenied       uint32 = 168
	CodeAccessExpired      uint32 = 169
	CodeInvalidPermission  uint32 = 170
	CodePermissionNotFound uint32 = 171
	CodeInvalidAccessLevel uint32 = 172

	// ZK Proof error codes (173-178)
	CodeInvalidProof              uint32 = 173
	CodeProofVerificationFailed   uint32 = 174
	CodeInvalidVerifyingKey       uint32 = 175
	CodeInvalidPublicInputs       uint32 = 176
	CodeUnsupportedProofType      uint32 = 177
	CodeProofDeserializationError uint32 = 178

	// Serialization error codes (179-180)
	CodeMarshalFailed   uint32 = 179
	CodeUnmarshalFailed uint32 = 180

	// General error codes (181-199)
	CodeInvalidAddress uint32 = 181
	CodeInvalidInput   uint32 = 182
	CodeInternal       uint32 = 199
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

// IBC errors
var (
	ErrIBCNotEnabled = errors.Register(ModuleName, 999, "IBC not enabled for identity module - cross-chain identity features will be available in v2.0")
)
