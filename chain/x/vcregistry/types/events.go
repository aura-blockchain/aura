package types

// Event types for the vcregistry module
const (
	EventTypeVCMinted           = "vc_minted"
	EventTypeVCRevoked          = "vc_revoked"
	EventTypeVCExpired          = "vc_expired"
	EventTypeVCSuspended        = "vc_suspended"
	EventTypeVCReactivated      = "vc_reactivated"
	EventTypeVCPolicyCreated    = "vc_policy_created"
	EventTypeVCPolicyUpdated    = "vc_policy_updated"
	EventTypeVCPolicyDeprecated = "vc_policy_deprecated"
	EventTypeDIDRegistered      = "did_registered"
	EventTypeDIDUpdated         = "did_updated"
	EventTypeMerkleRootUpdated  = "merkle_root_updated"
)

// Event attribute keys for VC Registry events
const (
	AttributeKeyVCID              = "vc_id"
	AttributeKeyVCType            = "vc_type"
	AttributeKeyVCTypeCustom      = "vc_type_custom"
	AttributeKeyHolderDID         = "holder_did"
	AttributeKeyHolderAddress     = "holder_address"
	AttributeKeyIssuedAt          = "issued_at"
	AttributeKeyExpiresAt         = "expires_at"
	AttributeKeyBlockHeight       = "block_height"
	AttributeKeyPolicyVersion     = "policy_version"
	AttributeKeyRevocationReason  = "revocation_reason"
	AttributeKeyRevoker           = "revoker"
	AttributeKeyRevokedAt         = "revoked_at"
	AttributeKeyExpiredAt         = "expired_at"
	AttributeKeySuspensionReason  = "suspension_reason"
	AttributeKeySuspendedAt       = "suspended_at"
	AttributeKeyReactivateAt      = "reactivate_at"
	AttributeKeyReactivatedAt     = "reactivated_at"
	AttributeKeyVCTypeName        = "vc_type_name"
	AttributeKeyVCTypeEnum        = "vc_type_enum"
	AttributeKeyCSThreshold       = "cs_threshold"
	AttributeKeyVersion           = "version"
	AttributeKeyOldVersion        = "old_version"
	AttributeKeyNewVersion        = "new_version"
	AttributeKeyDeprecatedAt      = "deprecated_at"
	AttributeKeyDeprecationReason = "deprecation_reason"
	AttributeKeyDID               = "did"
	AttributeKeyController        = "controller"
	AttributeKeyCreatedAt         = "created_at"
	AttributeKeyUpdatedAt         = "updated_at"
	AttributeKeyOldMerkleRoot     = "old_merkle_root"
	AttributeKeyNewMerkleRoot     = "new_merkle_root"
	AttributeKeyTotalRevocations  = "total_revocations"
)

// NewEventVCMinted creates a new EventVCMinted event with attribute values
func NewEventVCMinted(vcID string, vcType string, vcTypeCustom string, holderDID string,
	holderAddress string, issuedAt string, expiresAt string, blockHeight string, policyVersion string) map[string]string {
	return map[string]string{
		AttributeKeyVCID:          vcID,
		AttributeKeyVCType:        vcType,
		AttributeKeyVCTypeCustom:  vcTypeCustom,
		AttributeKeyHolderDID:     holderDID,
		AttributeKeyHolderAddress: holderAddress,
		AttributeKeyIssuedAt:      issuedAt,
		AttributeKeyExpiresAt:     expiresAt,
		AttributeKeyBlockHeight:   blockHeight,
		AttributeKeyPolicyVersion: policyVersion,
	}
}

// NewEventVCRevoked creates a new EventVCRevoked event with attribute values
func NewEventVCRevoked(vcID string, vcType string, revocationReason string, revoker string,
	revokedAt string, blockHeight string) map[string]string {
	return map[string]string{
		AttributeKeyVCID:             vcID,
		AttributeKeyVCType:           vcType,
		AttributeKeyRevocationReason: revocationReason,
		AttributeKeyRevoker:          revoker,
		AttributeKeyRevokedAt:        revokedAt,
		AttributeKeyBlockHeight:      blockHeight,
	}
}

// NewEventVCExpired creates a new EventVCExpired event with attribute values
func NewEventVCExpired(vcID string, vcType string, holderAddress string, expiredAt string) map[string]string {
	return map[string]string{
		AttributeKeyVCID:          vcID,
		AttributeKeyVCType:        vcType,
		AttributeKeyHolderAddress: holderAddress,
		AttributeKeyExpiredAt:     expiredAt,
	}
}

// NewEventVCSuspended creates a new EventVCSuspended event with attribute values
func NewEventVCSuspended(vcID string, suspensionReason string, suspendedAt string, reactivateAt string) map[string]string {
	return map[string]string{
		AttributeKeyVCID:             vcID,
		AttributeKeySuspensionReason: suspensionReason,
		AttributeKeySuspendedAt:      suspendedAt,
		AttributeKeyReactivateAt:     reactivateAt,
	}
}

// NewEventVCReactivated creates a new EventVCReactivated event with attribute values
func NewEventVCReactivated(vcID string, reactivatedAt string) map[string]string {
	return map[string]string{
		AttributeKeyVCID:          vcID,
		AttributeKeyReactivatedAt: reactivatedAt,
	}
}

// NewEventVCPolicyCreated creates a new EventVCPolicyCreated event with attribute values
func NewEventVCPolicyCreated(vcTypeName string, vcTypeEnum string, csThreshold string, version string, blockHeight string) map[string]string {
	return map[string]string{
		AttributeKeyVCTypeName:  vcTypeName,
		AttributeKeyVCTypeEnum:  vcTypeEnum,
		AttributeKeyCSThreshold: csThreshold,
		AttributeKeyVersion:     version,
		AttributeKeyBlockHeight: blockHeight,
	}
}

// NewEventVCPolicyUpdated creates a new EventVCPolicyUpdated event with attribute values
func NewEventVCPolicyUpdated(vcTypeName string, oldVersion string, newVersion string, blockHeight string) map[string]string {
	return map[string]string{
		AttributeKeyVCTypeName:  vcTypeName,
		AttributeKeyOldVersion:  oldVersion,
		AttributeKeyNewVersion:  newVersion,
		AttributeKeyBlockHeight: blockHeight,
	}
}

// NewEventVCPolicyDeprecated creates a new EventVCPolicyDeprecated event with attribute values
func NewEventVCPolicyDeprecated(vcTypeName string, deprecationReason string, deprecatedAt string) map[string]string {
	return map[string]string{
		AttributeKeyVCTypeName:        vcTypeName,
		AttributeKeyDeprecationReason: deprecationReason,
		AttributeKeyDeprecatedAt:      deprecatedAt,
	}
}

// NewEventDIDRegistered creates a new EventDIDRegistered event with attribute values
func NewEventDIDRegistered(did string, controller string, createdAt string, blockHeight string) map[string]string {
	return map[string]string{
		AttributeKeyDID:         did,
		AttributeKeyController:  controller,
		AttributeKeyCreatedAt:   createdAt,
		AttributeKeyBlockHeight: blockHeight,
	}
}

// NewEventDIDUpdated creates a new EventDIDUpdated event with attribute values
func NewEventDIDUpdated(did string, updatedAt string, blockHeight string) map[string]string {
	return map[string]string{
		AttributeKeyDID:         did,
		AttributeKeyUpdatedAt:   updatedAt,
		AttributeKeyBlockHeight: blockHeight,
	}
}

// NewEventMerkleRootUpdated creates a new EventMerkleRootUpdated event with attribute values
func NewEventMerkleRootUpdated(oldMerkleRoot string, newMerkleRoot string, totalRevocations string, blockHeight string) map[string]string {
	return map[string]string{
		AttributeKeyOldMerkleRoot:    oldMerkleRoot,
		AttributeKeyNewMerkleRoot:    newMerkleRoot,
		AttributeKeyTotalRevocations: totalRevocations,
		AttributeKeyBlockHeight:      blockHeight,
	}
}
