package types

import "fmt"

// Event types for the auth module
const (
	EventTypeRoleCreated          = "role_created"
	EventTypeRoleUpdated          = "role_updated"
	EventTypeRoleDeleted          = "role_deleted"
	EventTypeRoleAssigned         = "role_assigned"
	EventTypeRoleRevoked          = "role_revoked"
	EventTypeMultisigWalletCreated = "multisig_wallet_created"
	EventTypeMultisigProposalCreated = "multisig_proposal_created"
	EventTypeMultisigSignatureAdded = "multisig_signature_added"
	EventTypeMultisigExecuted     = "multisig_executed"
	EventTypeTimeLockCreated      = "timelock_created"
	EventTypeTimeLockExecuted     = "timelock_executed"
	EventTypeTimeLockCancelled    = "timelock_cancelled"
	EventTypeEmergencyAdminAdded  = "emergency_admin_added"
	EventTypeEmergencyAdminRemoved = "emergency_admin_removed"
	EventTypeValidatorKeyRotated  = "validator_key_rotated"
	EventTypeSessionCreated       = "session_created"
	EventTypeSessionInvalidated   = "session_invalidated"
	EventTypeRateLimitHit         = "rate_limit_hit"
	EventTypeAuditLogCreated      = "audit_log_created"
	EventTypePermissionChecked    = "permission_checked"
	EventTypeAuthorizationDenied  = "authorization_denied"
)

// Event attribute keys
const (
	AttributeKeyRoleName          = "role_name"
	AttributeKeyPermissions       = "permissions"
	AttributeKeyPermissionCount   = "permission_count"
	AttributeKeyDescription       = "description"
	AttributeKeyAddress           = "address"
	AttributeKeyUserAddress       = "user_address"
	AttributeKeyAssignedBy        = "assigned_by"
	AttributeKeyRevokedBy         = "revoked_by"
	AttributeKeyWalletAddress     = "wallet_address"
	AttributeKeySigners           = "signers"
	AttributeKeySignerCount       = "signer_count"
	AttributeKeyQuorum            = "quorum"
	AttributeKeyProposalID        = "proposal_id"
	AttributeKeySigner            = "signer"
	AttributeKeySignatureCount    = "signature_count"
	AttributeKeyProposer          = "proposer"
	AttributeKeyProposalTitle     = "proposal_title"
	AttributeKeyActionID          = "action_id"
	AttributeKeyActionType        = "action_type"
	AttributeKeyCreator           = "creator"
	AttributeKeyExecutionTime     = "execution_time"
	AttributeKeyExecutor          = "executor"
	AttributeKeyCanceller         = "canceller"
	AttributeKeyAdminAddress      = "admin_address"
	AttributeKeyAddedBy           = "added_by"
	AttributeKeyRemovedBy         = "removed_by"
	AttributeKeyOldPublicKey      = "old_public_key"
	AttributeKeyNewPublicKey      = "new_public_key"
	AttributeKeyValidatorAddress  = "validator_address"
	AttributeKeySessionID         = "session_id"
	AttributeKeySessionDuration   = "session_duration_seconds"
	AttributeKeyDeviceInfo        = "device_info"
	AttributeKeyIPAddress         = "ip_address"
	AttributeKeyInvalidatedBy     = "invalidated_by"
	AttributeKeyRateLimitKey      = "rate_limit_key"
	AttributeKeyRateLimitValue    = "rate_limit_value"
	AttributeKeyRateLimitWindow   = "rate_limit_window_seconds"
	AttributeKeyRateLimitHits     = "rate_limit_hits"
	AttributeKeyActor             = "actor"
	AttributeKeyAction            = "action"
	AttributeKeyResource          = "resource"
	AttributeKeyResult            = "result"
	AttributeKeyReason            = "reason"
	AttributeKeyPermission        = "permission"
	AttributeKeyBlockHeight       = "block_height"
	AttributeKeyBlockTime         = "block_time"
	AttributeKeyTimestamp         = "timestamp"
)

// Helper functions for creating event attributes

// NewRoleCreatedEvent creates attributes for role creation
func NewRoleCreatedEvent(roleName, description string, permissions []string, creator string, blockHeight int64, blockTime string) map[string]string {
	return map[string]string{
		AttributeKeyRoleName:        roleName,
		AttributeKeyDescription:     description,
		AttributeKeyPermissions:     formatPermissions(permissions),
		AttributeKeyPermissionCount: formatInt(len(permissions)),
		AttributeKeyCreator:         creator,
		AttributeKeyBlockHeight:     formatInt64(blockHeight),
		AttributeKeyBlockTime:       blockTime,
	}
}

// NewRoleUpdatedEvent creates attributes for role update
func NewRoleUpdatedEvent(roleName string, oldPermissions, newPermissions []string, updatedBy string, blockHeight int64, blockTime string) map[string]string {
	return map[string]string{
		AttributeKeyRoleName:        roleName,
		"old_permissions":           formatPermissions(oldPermissions),
		"new_permissions":           formatPermissions(newPermissions),
		"old_permission_count":      formatInt(len(oldPermissions)),
		"new_permission_count":      formatInt(len(newPermissions)),
		"updated_by":                updatedBy,
		AttributeKeyBlockHeight:     formatInt64(blockHeight),
		AttributeKeyBlockTime:       blockTime,
	}
}

// NewRoleAssignedEvent creates attributes for role assignment
func NewRoleAssignedEvent(roleName, address, assignedBy string, blockHeight int64, blockTime string) map[string]string {
	return map[string]string{
		AttributeKeyRoleName:    roleName,
		AttributeKeyAddress:     address,
		AttributeKeyAssignedBy:  assignedBy,
		AttributeKeyBlockHeight: formatInt64(blockHeight),
		AttributeKeyBlockTime:   blockTime,
	}
}

// NewRoleRevokedEvent creates attributes for role revocation
func NewRoleRevokedEvent(roleName, address, revokedBy string, blockHeight int64, blockTime string) map[string]string {
	return map[string]string{
		AttributeKeyRoleName:    roleName,
		AttributeKeyAddress:     address,
		AttributeKeyRevokedBy:   revokedBy,
		AttributeKeyBlockHeight: formatInt64(blockHeight),
		AttributeKeyBlockTime:   blockTime,
	}
}

// NewMultisigWalletCreatedEvent creates attributes for multisig wallet creation
func NewMultisigWalletCreatedEvent(walletAddress string, signers []string, quorum uint32, creator string, blockHeight int64, blockTime string) map[string]string {
	return map[string]string{
		AttributeKeyWalletAddress: walletAddress,
		AttributeKeySigners:       formatStringSlice(signers),
		AttributeKeySignerCount:   formatInt(len(signers)),
		AttributeKeyQuorum:        formatUint32(quorum),
		AttributeKeyCreator:       creator,
		AttributeKeyBlockHeight:   formatInt64(blockHeight),
		AttributeKeyBlockTime:     blockTime,
	}
}

// NewMultisigProposalCreatedEvent creates attributes for multisig proposal creation
func NewMultisigProposalCreatedEvent(proposalID uint64, walletAddress, proposer, title string, quorum uint32, blockHeight int64, blockTime string) map[string]string {
	return map[string]string{
		AttributeKeyProposalID:      formatUint64(proposalID),
		AttributeKeyWalletAddress:   walletAddress,
		AttributeKeyProposer:        proposer,
		AttributeKeyProposalTitle:   title,
		AttributeKeyQuorum:          formatUint32(quorum),
		AttributeKeyBlockHeight:     formatInt64(blockHeight),
		AttributeKeyBlockTime:       blockTime,
	}
}

// NewMultisigSignatureAddedEvent creates attributes for multisig signature addition
func NewMultisigSignatureAddedEvent(proposalID uint64, signer string, signatureCount uint32, quorum uint32, blockHeight int64, blockTime string) map[string]string {
	return map[string]string{
		AttributeKeyProposalID:     formatUint64(proposalID),
		AttributeKeySigner:         signer,
		AttributeKeySignatureCount: formatUint32(signatureCount),
		AttributeKeyQuorum:         formatUint32(quorum),
		AttributeKeyBlockHeight:    formatInt64(blockHeight),
		AttributeKeyBlockTime:      blockTime,
	}
}

// NewMultisigExecutedEvent creates attributes for multisig execution
func NewMultisigExecutedEvent(proposalID uint64, walletAddress, executor string, blockHeight int64, blockTime string) map[string]string {
	return map[string]string{
		AttributeKeyProposalID:    formatUint64(proposalID),
		AttributeKeyWalletAddress: walletAddress,
		AttributeKeyExecutor:      executor,
		AttributeKeyBlockHeight:   formatInt64(blockHeight),
		AttributeKeyBlockTime:     blockTime,
	}
}

// NewTimeLockCreatedEvent creates attributes for timelock creation
func NewTimeLockCreatedEvent(actionID uint64, actionType, creator string, executionTime int64, blockHeight int64, blockTime string) map[string]string {
	return map[string]string{
		AttributeKeyActionID:       formatUint64(actionID),
		AttributeKeyActionType:     actionType,
		AttributeKeyCreator:        creator,
		AttributeKeyExecutionTime:  formatInt64(executionTime),
		AttributeKeyBlockHeight:    formatInt64(blockHeight),
		AttributeKeyBlockTime:      blockTime,
	}
}

// NewTimeLockExecutedEvent creates attributes for timelock execution
func NewTimeLockExecutedEvent(actionID uint64, executor string, blockHeight int64, blockTime string) map[string]string {
	return map[string]string{
		AttributeKeyActionID:    formatUint64(actionID),
		AttributeKeyExecutor:    executor,
		AttributeKeyBlockHeight: formatInt64(blockHeight),
		AttributeKeyBlockTime:   blockTime,
	}
}

// NewSessionCreatedEvent creates attributes for session creation
func NewSessionCreatedEvent(sessionID, userAddress, deviceInfo, ipAddress string, duration int64, blockHeight int64, blockTime string) map[string]string {
	return map[string]string{
		AttributeKeySessionID:       sessionID,
		AttributeKeyUserAddress:     userAddress,
		AttributeKeyDeviceInfo:      deviceInfo,
		AttributeKeyIPAddress:       ipAddress,
		AttributeKeySessionDuration: formatInt64(duration),
		AttributeKeyBlockHeight:     formatInt64(blockHeight),
		AttributeKeyBlockTime:       blockTime,
	}
}

// NewSessionInvalidatedEvent creates attributes for session invalidation
func NewSessionInvalidatedEvent(sessionID, userAddress, invalidatedBy, reason string, blockHeight int64, blockTime string) map[string]string {
	return map[string]string{
		AttributeKeySessionID:     sessionID,
		AttributeKeyUserAddress:   userAddress,
		AttributeKeyInvalidatedBy: invalidatedBy,
		AttributeKeyReason:        reason,
		AttributeKeyBlockHeight:   formatInt64(blockHeight),
		AttributeKeyBlockTime:     blockTime,
	}
}

// NewRateLimitHitEvent creates attributes for rate limit hit
func NewRateLimitHitEvent(key, address string, limit, hits uint32, window int64, blockHeight int64, blockTime string) map[string]string {
	return map[string]string{
		AttributeKeyRateLimitKey:    key,
		AttributeKeyAddress:         address,
		AttributeKeyRateLimitValue:  formatUint32(limit),
		AttributeKeyRateLimitHits:   formatUint32(hits),
		AttributeKeyRateLimitWindow: formatInt64(window),
		AttributeKeyBlockHeight:     formatInt64(blockHeight),
		AttributeKeyBlockTime:       blockTime,
	}
}

// NewAuditLogCreatedEvent creates attributes for audit log creation
func NewAuditLogCreatedEvent(actor, action, resource, result string, blockHeight int64, blockTime string) map[string]string {
	return map[string]string{
		AttributeKeyActor:       actor,
		AttributeKeyAction:      action,
		AttributeKeyResource:    resource,
		AttributeKeyResult:      result,
		AttributeKeyBlockHeight: formatInt64(blockHeight),
		AttributeKeyBlockTime:   blockTime,
	}
}

// NewAuthorizationDeniedEvent creates attributes for authorization denial
func NewAuthorizationDeniedEvent(actor, permission, resource, reason string, blockHeight int64, blockTime string) map[string]string {
	return map[string]string{
		AttributeKeyActor:       actor,
		AttributeKeyPermission:  permission,
		AttributeKeyResource:    resource,
		AttributeKeyReason:      reason,
		AttributeKeyBlockHeight: formatInt64(blockHeight),
		AttributeKeyBlockTime:   blockTime,
	}
}

// Helper formatting functions

func formatInt(i int) string {
	return fmt.Sprintf("%d", i)
}

func formatInt64(i int64) string {
	return fmt.Sprintf("%d", i)
}

func formatUint32(u uint32) string {
	return fmt.Sprintf("%d", u)
}

func formatUint64(u uint64) string {
	return fmt.Sprintf("%d", u)
}

func formatPermissions(perms []string) string {
	return formatStringSlice(perms)
}

func formatStringSlice(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += "," + strs[i]
	}
	return result
}
