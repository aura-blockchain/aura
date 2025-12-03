package types

import (
	identitypb "github.com/aequitas/aura/proto/aura/identity/v1beta1"
)

// ============================================================================
// Re-export Proto Types
// ============================================================================

// Core types from proto
type (
	Params               = identitypb.Params
	AuthParams           = identitypb.AuthParams
	IdentityChangeParams = identitypb.IdentityChangeParams
	Role                 = identitypb.Role
	RoleAssignment       = identitypb.RoleAssignment
	AuditLog             = identitypb.AuditLog
	Session              = identitypb.Session
	RateLimitConfig      = identitypb.RateLimitConfig
	MultisigWallet       = identitypb.MultisigWallet
	MultisigProposal     = identitypb.MultisigProposal
	TimeLockedAction     = identitypb.TimeLockedAction
	EmergencyAdmin       = identitypb.EmergencyAdmin
	ValidatorKeyRotation = identitypb.ValidatorKeyRotation
	DIDKeyRotation       = identitypb.DIDKeyRotation
	DIDKeyHistory        = identitypb.DIDKeyHistory
	DIDKeyHistoryEntry   = identitypb.DIDKeyHistoryEntry
	CredentialRevocation = identitypb.CredentialRevocation
	IdentityRecord       = identitypb.IdentityRecord
	ChangeRequest        = identitypb.ChangeRequest
	ChangeHistory        = identitypb.ChangeHistory
	GenesisState         = identitypb.GenesisState
)

// Enums from proto
type (
	AuditResult          = identitypb.AuditResult
	IdentityStatus       = identitypb.IdentityStatus
	ChangeType           = identitypb.ChangeType
	ChangeStatus         = identitypb.ChangeStatus
	ProposalStatus       = identitypb.ProposalStatus
	ActionStatus         = identitypb.ActionStatus
	RotationStatus       = identitypb.RotationStatus
	DIDKeyRotationStatus = identitypb.DIDKeyRotationStatus
	WalletType           = identitypb.WalletType
)

// Enum constants - AuditResult
const (
	AuditResultUnspecified = identitypb.AuditResult_AUDIT_RESULT_UNSPECIFIED
	AuditResultSuccess     = identitypb.AuditResult_AUDIT_RESULT_SUCCESS
	AuditResultFailure     = identitypb.AuditResult_AUDIT_RESULT_FAILURE
	AuditResultDenied      = identitypb.AuditResult_AUDIT_RESULT_DENIED
)

// Enum constants - IdentityStatus
const (
	IdentityStatusUnspecified         = identitypb.IdentityStatus_IDENTITY_STATUS_UNSPECIFIED
	IdentityStatusActive              = identitypb.IdentityStatus_IDENTITY_STATUS_ACTIVE
	IdentityStatusSuspended           = identitypb.IdentityStatus_IDENTITY_STATUS_SUSPENDED
	IdentityStatusRevoked             = identitypb.IdentityStatus_IDENTITY_STATUS_REVOKED
	IdentityStatusIdle                = identitypb.IdentityStatus_IDENTITY_STATUS_IDLE
	IdentityStatusPendingVerification = identitypb.IdentityStatus_IDENTITY_STATUS_PENDING_VERIFICATION
	IdentityStatusErased              = identitypb.IdentityStatus_IDENTITY_STATUS_ERASED
)

// Enum constants - ChangeType
const (
	ChangeTypeUnspecified              = identitypb.ChangeType_CHANGE_TYPE_UNSPECIFIED
	ChangeTypeUpdateAttributes         = identitypb.ChangeType_CHANGE_TYPE_UPDATE_ATTRIBUTES
	ChangeTypeAddVerificationMethod    = identitypb.ChangeType_CHANGE_TYPE_ADD_VERIFICATION_METHOD
	ChangeTypeRevokeVerificationMethod = identitypb.ChangeType_CHANGE_TYPE_REVOKE_VERIFICATION_METHOD
	ChangeTypeTransferControl          = identitypb.ChangeType_CHANGE_TYPE_TRANSFER_CONTROL
	ChangeTypeUpdateMetadata           = identitypb.ChangeType_CHANGE_TYPE_UPDATE_METADATA
)

// Enum constants - ChangeStatus
const (
	ChangeStatusUnspecified = identitypb.ChangeStatus_CHANGE_STATUS_UNSPECIFIED
	ChangeStatusPending     = identitypb.ChangeStatus_CHANGE_STATUS_PENDING
	ChangeStatusApproved    = identitypb.ChangeStatus_CHANGE_STATUS_APPROVED
	ChangeStatusRejected    = identitypb.ChangeStatus_CHANGE_STATUS_REJECTED
	ChangeStatusExecuted    = identitypb.ChangeStatus_CHANGE_STATUS_EXECUTED
)

// Enum constants - ProposalStatus
const (
	ProposalStatusUnspecified = identitypb.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED
	ProposalStatusPending     = identitypb.ProposalStatus_PROPOSAL_STATUS_PENDING
	ProposalStatusApproved    = identitypb.ProposalStatus_PROPOSAL_STATUS_APPROVED
	ProposalStatusRejected    = identitypb.ProposalStatus_PROPOSAL_STATUS_REJECTED
	ProposalStatusExecuted    = identitypb.ProposalStatus_PROPOSAL_STATUS_EXECUTED
	ProposalStatusExpired     = identitypb.ProposalStatus_PROPOSAL_STATUS_EXPIRED
)

// Enum constants - ActionStatus
const (
	ActionStatusUnspecified = identitypb.ActionStatus_ACTION_STATUS_UNSPECIFIED
	ActionStatusPending     = identitypb.ActionStatus_ACTION_STATUS_PENDING
	ActionStatusReady       = identitypb.ActionStatus_ACTION_STATUS_READY
	ActionStatusExecuted    = identitypb.ActionStatus_ACTION_STATUS_EXECUTED
	ActionStatusCancelled   = identitypb.ActionStatus_ACTION_STATUS_CANCELLED
)

// Enum constants - RotationStatus
const (
	RotationStatusUnspecified = identitypb.RotationStatus_ROTATION_STATUS_UNSPECIFIED
	RotationStatusPending     = identitypb.RotationStatus_ROTATION_STATUS_PENDING
	RotationStatusCompleted   = identitypb.RotationStatus_ROTATION_STATUS_COMPLETED
	RotationStatusFailed      = identitypb.RotationStatus_ROTATION_STATUS_FAILED
)

// Enum constants - DIDKeyRotationStatus
const (
	DIDKeyRotationStatusUnspecified = identitypb.DIDKeyRotationStatus_DID_KEY_ROTATION_STATUS_UNSPECIFIED
	DIDKeyRotationStatusPending     = identitypb.DIDKeyRotationStatus_DID_KEY_ROTATION_STATUS_PENDING
	DIDKeyRotationStatusCompleted   = identitypb.DIDKeyRotationStatus_DID_KEY_ROTATION_STATUS_COMPLETED
	DIDKeyRotationStatusReverted    = identitypb.DIDKeyRotationStatus_DID_KEY_ROTATION_STATUS_REVERTED
)

// Enum constants - WalletType
const (
	WalletTypeUnspecified = identitypb.WalletType_WALLET_TYPE_UNSPECIFIED
	WalletType3Of5        = identitypb.WalletType_WALLET_TYPE_3_OF_5
	WalletType5Of7        = identitypb.WalletType_WALLET_TYPE_5_OF_7
	WalletTypeCustom      = identitypb.WalletType_WALLET_TYPE_CUSTOM
)

// ============================================================================
// Permission Constants
// ============================================================================

const (
	// Role names
	RoleAdmin     = "admin"
	RoleModerator = "moderator"
	RoleValidator = "validator"
	RoleUser      = "user"

	// Permission types
	PermissionAdmin                = "admin"
	PermissionCreateRole           = "create_role"
	PermissionAssignRole           = "assign_role"
	PermissionRevokeRole           = "revoke_role"
	PermissionManageMultisig       = "manage_multisig"
	PermissionManageTimeLock       = "manage_timelock"
	PermissionManageEmergency      = "manage_emergency"
	PermissionRotateValidatorKey   = "rotate_validator_key"
	PermissionManageSession        = "manage_session"
	PermissionViewAuditLogs        = "view_audit_logs"
	PermissionManageIdentity       = "manage_identity"
	PermissionVerifyIdentity       = "verify_identity"
	PermissionApproveChangeRequest = "approve_change_request"
)

// ============================================================================
// Helper Types for Collections
// ============================================================================

// RoleAssignmentList is a helper type for storing multiple role assignments
type RoleAssignmentList struct {
	Assignments []*RoleAssignment `json:"assignments"`
}

// SessionIDList is a helper type for storing session IDs
type SessionIDList struct {
	SessionIDs []string `json:"session_ids"`
}

// ============================================================================
// Default Params
// ============================================================================

// DefaultParams returns default module parameters
func DefaultParams() *Params {
	return &Params{
		Auth: &AuthParams{
			EnableRbac:                    true,
			MaxRolesPerAccount:            10,
			SessionTimeout:                nil, // Will be set by proto
			EnableAuditLogging:            true,
			DefaultTimelockDelaySeconds:   3600,  // 1 hour
			DefaultRequestsPerMinute:      60,
			DefaultRequestsPerHour:        3600,
			DefaultRequestsPerDay:         86400,
			MultisigProposalExpirySeconds: 604800, // 7 days
		},
		Change: &IdentityChangeParams{
			MaxRequestsPerWalletPerMonth:     10,
			MinConfidenceAfterChange:         50,
			StalenessHeightThreshold:         100000,
			AssistantSlashOnFalsePositive:    true,
			StalenessInvestigatorChain:       "",
			KeyRotationGracePeriodSeconds:    86400, // 24 hours default grace period
		},
	}
}
