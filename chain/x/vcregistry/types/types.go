package types

import (
	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
)

// Type aliases for proto types
type (
	VCRecord           = vcregistrypb.VCRecord
	VCPolicy           = vcregistrypb.VCPolicy
	DIDDocument        = vcregistrypb.DIDDocument
	RevocationRecord   = vcregistrypb.RevocationRecord
	RevocationList     = vcregistrypb.RevocationList
	VerificationMethod = vcregistrypb.VerificationMethod

	// Enums
	VCType           = vcregistrypb.VCType
	VCStatus         = vcregistrypb.VCStatus
	RevocationReason = vcregistrypb.RevocationReason
	VCPolicyStatus   = vcregistrypb.VCPolicyStatus
)

// Enum constants for convenience
const (
	// VCType constants
	VCTypeUnspecified         = vcregistrypb.VCType_VC_TYPE_UNSPECIFIED
	VCTypeVerifiedHuman       = vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN
	VCTypeAgeOver18           = vcregistrypb.VCType_VC_TYPE_AGE_OVER_18
	VCTypeAgeOver21           = vcregistrypb.VCType_VC_TYPE_AGE_OVER_21
	VCTypeResidentOf          = vcregistrypb.VCType_VC_TYPE_RESIDENT_OF
	VCTypeBiometricAuth       = vcregistrypb.VCType_VC_TYPE_BIOMETRIC_AUTH
	VCTypeKYCVerification     = vcregistrypb.VCType_VC_TYPE_KYC_VERIFICATION
	VCTypeNotaryPublic        = vcregistrypb.VCType_VC_TYPE_NOTARY_PUBLIC
	VCTypeProfessionalLicense = vcregistrypb.VCType_VC_TYPE_PROFESSIONAL_LICENSE

	// Arena focus VCType constants
	VCTypeBiometricFocus     = vcregistrypb.VCType_VC_TYPE_BIOMETRIC_FOCUS
	VCTypeSocialFocus        = vcregistrypb.VCType_VC_TYPE_SOCIAL_FOCUS
	VCTypeGeolocationFocus   = vcregistrypb.VCType_VC_TYPE_GEOLOCATION_FOCUS
	VCTypeHighAssuranceFocus = vcregistrypb.VCType_VC_TYPE_HIGH_ASSURANCE_FOCUS
	VCTypePossessionFocus    = vcregistrypb.VCType_VC_TYPE_POSSESSION_FOCUS
	VCTypeKnowledgeFocus     = vcregistrypb.VCType_VC_TYPE_KNOWLEDGE_FOCUS
	VCTypePersistenceFocus   = vcregistrypb.VCType_VC_TYPE_PERSISTENCE_FOCUS
	VCTypeSpecializedFocus   = vcregistrypb.VCType_VC_TYPE_SPECIALIZED_FOCUS

	VCTypeCustom = vcregistrypb.VCType_VC_TYPE_CUSTOM

	// VCStatus constants
	VCStatusUnspecified = vcregistrypb.VCStatus_VC_STATUS_UNSPECIFIED
	VCStatusPending     = vcregistrypb.VCStatus_VC_STATUS_PENDING
	VCStatusActive      = vcregistrypb.VCStatus_VC_STATUS_ACTIVE
	VCStatusRevoked     = vcregistrypb.VCStatus_VC_STATUS_REVOKED
	VCStatusExpired     = vcregistrypb.VCStatus_VC_STATUS_EXPIRED
	VCStatusSuspended   = vcregistrypb.VCStatus_VC_STATUS_SUSPENDED

	// RevocationReason constants
	RevocationReasonUnspecified        = vcregistrypb.RevocationReason_REVOCATION_REASON_UNSPECIFIED
	RevocationReasonUserRequest        = vcregistrypb.RevocationReason_REVOCATION_REASON_USER_REQUEST
	RevocationReasonFraudDetected      = vcregistrypb.RevocationReason_REVOCATION_REASON_FRAUD_DETECTED
	RevocationReasonCSBelowThreshold   = vcregistrypb.RevocationReason_REVOCATION_REASON_CS_BELOW_THRESHOLD
	RevocationReasonIRInvalidated      = vcregistrypb.RevocationReason_REVOCATION_REASON_IR_INVALIDATED
	RevocationReasonExpired            = vcregistrypb.RevocationReason_REVOCATION_REASON_EXPIRED
	RevocationReasonGovernance         = vcregistrypb.RevocationReason_REVOCATION_REASON_GOVERNANCE
	RevocationReasonSecurityCompromise = vcregistrypb.RevocationReason_REVOCATION_REASON_SECURITY_COMPROMISE
	RevocationReasonPolicyChange       = vcregistrypb.RevocationReason_REVOCATION_REASON_POLICY_CHANGE

	// VCPolicyStatus constants
	VCPolicyStatusUnspecified = vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_UNSPECIFIED
	VCPolicyStatusDraft       = vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_DRAFT
	VCPolicyStatusActive      = vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE
	VCPolicyStatusDeprecated  = vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_DEPRECATED
)

// Conversion functions

// VCRecordFromProto converts proto VCRecord to internal type
func VCRecordFromProto(pb *vcregistrypb.VCRecord) VCRecord {
	if pb == nil {
		return VCRecord{}
	}
	return *pb
}

// VCRecordToProto converts internal VCRecord to proto type
func VCRecordToProto(vc VCRecord) *vcregistrypb.VCRecord {
	return &vc
}

// RevocationRecordFromProto converts proto RevocationRecord to internal type
func RevocationRecordFromProto(pb *vcregistrypb.RevocationRecord) RevocationRecord {
	if pb == nil {
		return RevocationRecord{}
	}
	return *pb
}

// RevocationRecordToProto converts internal RevocationRecord to proto type
func RevocationRecordToProto(rr RevocationRecord) *vcregistrypb.RevocationRecord {
	return &rr
}

// RevocationListFromProto converts proto RevocationList to internal type
func RevocationListFromProto(pb *vcregistrypb.RevocationList) *RevocationList {
	if pb == nil {
		return &RevocationList{}
	}
	rl := *pb
	return &rl
}

// RevocationListToProto converts internal RevocationList to proto type
func RevocationListToProto(rl *RevocationList) *vcregistrypb.RevocationList {
	if rl == nil {
		return &vcregistrypb.RevocationList{}
	}
	return rl
}

// DIDDocumentFromProto converts proto DIDDocument to internal type
func DIDDocumentFromProto(pb *vcregistrypb.DIDDocument) DIDDocument {
	if pb == nil {
		return DIDDocument{}
	}
	return *pb
}

// DIDDocumentToProto converts internal DIDDocument to proto type
func DIDDocumentToProto(doc DIDDocument) *vcregistrypb.DIDDocument {
	return &doc
}

// VCPolicyFromProto converts proto VCPolicy to internal type
func VCPolicyFromProto(pb *vcregistrypb.VCPolicy) VCPolicy {
	if pb == nil {
		return VCPolicy{}
	}
	return *pb
}

// VCPolicyToProto converts internal VCPolicy to proto type
func VCPolicyToProto(policy VCPolicy) *vcregistrypb.VCPolicy {
	return &policy
}

// Params conversion functions are defined in params.go

// RegistryStats holds statistics about the VC registry (not in proto, local type only)
type RegistryStats struct {
	TotalVCs      uint64
	ActiveVCs     uint64
	RevokedVCs    uint64
	ExpiredVCs    uint64
	TotalDIDs     uint64
	TotalPolicies uint64
	VCsByType     map[string]uint64
}
