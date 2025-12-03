package types

// Event types for the identity module
const (
	// Role and permission events
	EventTypeRoleCreated         = "role_created"
	EventTypeRoleAssigned        = "role_assigned"
	EventTypeRoleRevoked         = "role_revoked"
	EventTypePermissionGranted   = "permission_granted"
	EventTypePermissionRevoked   = "permission_revoked"

	// Session events
	EventTypeSessionCreated = "session_created"
	EventTypeSessionEnded   = "session_ended"

	// Identity change events
	EventTypeIdentityCreated       = "identity_created"
	EventTypeIdentityUpdated       = "identity_updated"
	EventTypeChangeRequestCreated  = "change_request_created"
	EventTypeChangeRequestApproved = "change_request_approved"
	EventTypeChangeRequestRejected = "change_request_rejected"
	EventTypeChangeRequestExecuted = "change_request_executed"

	// GDPR events
	EventTypeIdentityErased = "identity_erased"

	// Credential revocation events
	EventTypeCredentialRevoked          = "credential_revoked"
	EventTypeBatchCredentialRevocation  = "batch_credential_revocation"
	EventTypeCredentialRestored         = "credential_restored"

	// Multisig events
	EventTypeMultisigWalletCreated   = "multisig_wallet_created"
	EventTypeMultisigProposalCreated = "multisig_proposal_created"
	EventTypeMultisigProposalSigned  = "multisig_proposal_signed"
	EventTypeMultisigProposalExecuted = "multisig_proposal_executed"

	// Time-locked action events
	EventTypeTimeLockedActionProposed = "timelock_action_proposed"
	EventTypeTimeLockedActionExecuted = "timelock_action_executed"
	EventTypeTimeLockedActionCancelled = "timelock_action_cancelled"

	// Emergency admin events
	EventTypeEmergencyAdminActivated   = "emergency_admin_activated"
	EventTypeEmergencyAdminDeactivated = "emergency_admin_deactivated"

	// Validator events
	EventTypeValidatorKeyRotated = "validator_key_rotated"
)

// Event attribute keys
const (
	// Common attributes
	AttributeKeyAddress   = "address"
	AttributeKeyActor     = "actor"
	AttributeKeyTimestamp = "timestamp"
	AttributeKeyReason    = "reason"

	// Role attributes
	AttributeKeyRoleName    = "role_name"
	AttributeKeyPermissions = "permissions"
	AttributeKeyAssigner    = "assigner"
	AttributeKeyRevoker     = "revoker"

	// Session attributes
	AttributeKeySessionID        = "session_id"
	AttributeKeyDeviceFingerprint = "device_fingerprint"

	// Identity attributes
	AttributeKeyDID             = "did"
	AttributeKeyRequestID       = "request_id"
	AttributeKeyChangeType      = "change_type"
	AttributeKeyRequester       = "requester"
	AttributeKeyAssistant       = "assistant"
	AttributeKeyConfidenceScore = "confidence_score"

	// Credential revocation attributes
	AttributeKeyCredentialId = "credential_id"
	AttributeKeyRevokedBy    = "revoked_by"
	AttributeKeyRestoredBy   = "restored_by"

	// Multisig attributes
	AttributeKeyWalletID   = "wallet_id"
	AttributeKeyProposalID = "proposal_id"
	AttributeKeySigner     = "signer"
	AttributeKeyThreshold  = "threshold"

	// Time-lock attributes
	AttributeKeyActionID     = "action_id"
	AttributeKeyActionType   = "action_type"
	AttributeKeyProposer     = "proposer"
	AttributeKeyExecutor     = "executor"
	AttributeKeyExecutableAt = "executable_at"

	// Validator attributes
	AttributeKeyValidatorAddress = "validator_address"
	AttributeKeyOldPubKey        = "old_pubkey"
	AttributeKeyNewPubKey        = "new_pubkey"
)
