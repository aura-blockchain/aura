package types

import pb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"

// Re-export all proto types
type (
	// Enums
	VCType               = pb.VCType
	VCStatus             = pb.VCStatus
	RevocationReason     = pb.RevocationReason
	VCPolicyStatus       = pb.VCPolicyStatus
	AttributeType        = pb.AttributeType
	DisclosurePolicyMode = pb.DisclosurePolicyMode

	// Core types - VC
	VCRecord             = pb.VCRecord
	VCPolicy             = pb.VCPolicy
	VCPresentation       = pb.VCPresentation
	VCStatusInfo         = pb.VCStatusInfo
	VCVerificationDetail = pb.VCVerificationDetail
	VerificationResult   = pb.VerificationResult
	PresentationContext  = pb.PresentationContext
	RevocationList       = pb.RevocationList
	RevocationRecord     = pb.RevocationRecord

	// Core types - DID
	DIDDocument        = pb.DIDDocument
	VerificationMethod = pb.VerificationMethod

	// Core types - Attributes
	AttributeVC             = pb.AttributeVC
	DiscloseableAttributes  = pb.DiscloseableAttributes
	AttributeDisclosure     = pb.AttributeDisclosure
	AttributeDisclosureRule = pb.AttributeDisclosureRule
	DisclosurePolicy        = pb.DisclosurePolicy
	DisclosureRequest       = pb.DisclosureRequest
	DisclosureResponse      = pb.DisclosureResponse
	VoiceCommand            = pb.VoiceCommand
	Params                  = pb.Params

	// Message types - VC
	MsgMintVC                     = pb.MsgMintVC
	MsgMintVCResponse             = pb.MsgMintVCResponse
	MsgRevokeVC                   = pb.MsgRevokeVC
	MsgRevokeVCResponse           = pb.MsgRevokeVCResponse
	MsgSuspendVC                  = pb.MsgSuspendVC
	MsgSuspendVCResponse          = pb.MsgSuspendVCResponse
	MsgReactivateVC               = pb.MsgReactivateVC
	MsgReactivateVCResponse       = pb.MsgReactivateVCResponse
	MsgAdminRevokeVC              = pb.MsgAdminRevokeVC
	MsgAdminRevokeVCResponse      = pb.MsgAdminRevokeVCResponse
	MsgCreateVCPolicy             = pb.MsgCreateVCPolicy
	MsgCreateVCPolicyResponse     = pb.MsgCreateVCPolicyResponse
	MsgUpdateVCPolicy             = pb.MsgUpdateVCPolicy
	MsgUpdateVCPolicyResponse     = pb.MsgUpdateVCPolicyResponse
	MsgDeprecateVCPolicy          = pb.MsgDeprecateVCPolicy
	MsgDeprecateVCPolicyResponse  = pb.MsgDeprecateVCPolicyResponse
	MsgCreatePresentation         = pb.MsgCreatePresentation
	MsgCreatePresentationResponse = pb.MsgCreatePresentationResponse

	// Message types - DID
	MsgRegisterDID               = pb.MsgRegisterDID
	MsgRegisterDIDResponse       = pb.MsgRegisterDIDResponse
	MsgUpdateDIDDocument         = pb.MsgUpdateDIDDocument
	MsgUpdateDIDDocumentResponse = pb.MsgUpdateDIDDocumentResponse

	// Message types - Attributes
	MsgCreateAttributeVC                  = pb.MsgCreateAttributeVC
	MsgCreateAttributeVCResponse          = pb.MsgCreateAttributeVCResponse
	MsgRevokeAttributeVC                  = pb.MsgRevokeAttributeVC
	MsgRevokeAttributeVCResponse          = pb.MsgRevokeAttributeVCResponse
	MsgUpdateDisclosurePolicy             = pb.MsgUpdateDisclosurePolicy
	MsgUpdateDisclosurePolicyResponse     = pb.MsgUpdateDisclosurePolicyResponse
	MsgCreateDisclosureRequest            = pb.MsgCreateDisclosureRequest
	MsgCreateDisclosureRequestResponse    = pb.MsgCreateDisclosureRequestResponse
	MsgRespondToDisclosureRequest         = pb.MsgRespondToDisclosureRequest
	MsgRespondToDisclosureRequestResponse = pb.MsgRespondToDisclosureRequestResponse

	// Query types - VC
	QueryGetVCRequest                    = pb.QueryGetVCRequest
	QueryGetVCResponse                   = pb.QueryGetVCResponse
	QueryListUserVCsRequest              = pb.QueryListUserVCsRequest
	QueryListUserVCsResponse             = pb.QueryListUserVCsResponse
	QueryCheckVCStatusRequest            = pb.QueryCheckVCStatusRequest
	QueryCheckVCStatusResponse           = pb.QueryCheckVCStatusResponse
	QueryBatchVCStatusRequest            = pb.QueryBatchVCStatusRequest
	QueryBatchVCStatusResponse           = pb.QueryBatchVCStatusResponse
	QueryCheckRevocationRequest          = pb.QueryCheckRevocationRequest
	QueryCheckRevocationResponse         = pb.QueryCheckRevocationResponse
	QueryGetRevocationListRequest        = pb.QueryGetRevocationListRequest
	QueryGetRevocationListResponse       = pb.QueryGetRevocationListResponse
	QueryGetVCPolicyRequest              = pb.QueryGetVCPolicyRequest
	QueryGetVCPolicyResponse             = pb.QueryGetVCPolicyResponse
	QueryListVCPoliciesRequest           = pb.QueryListVCPoliciesRequest
	QueryListVCPoliciesResponse          = pb.QueryListVCPoliciesResponse
	QueryValidateMintEligibilityRequest  = pb.QueryValidateMintEligibilityRequest
	QueryValidateMintEligibilityResponse = pb.QueryValidateMintEligibilityResponse
	QueryVerifyPresentationRequest       = pb.QueryVerifyPresentationRequest
	QueryVerifyPresentationResponse      = pb.QueryVerifyPresentationResponse
	QueryStatsRequest                    = pb.QueryStatsRequest
	QueryStatsResponse                   = pb.QueryStatsResponse

	// Query types - DID
	QueryResolveDIDRequest       = pb.QueryResolveDIDRequest
	QueryResolveDIDResponse      = pb.QueryResolveDIDResponse
	QueryGetDIDByAddressRequest  = pb.QueryGetDIDByAddressRequest
	QueryGetDIDByAddressResponse = pb.QueryGetDIDByAddressResponse

	// Query types - Attributes
	QueryGetAttributeVCRequest             = pb.QueryGetAttributeVCRequest
	QueryGetAttributeVCResponse            = pb.QueryGetAttributeVCResponse
	QueryAttributeVCsRequest               = pb.QueryAttributeVCsRequest
	QueryAttributeVCsResponse              = pb.QueryAttributeVCsResponse
	QueryDisclosurePolicyRequest           = pb.QueryDisclosurePolicyRequest
	QueryDisclosurePolicyResponse          = pb.QueryDisclosurePolicyResponse
	QueryDisclosureRequestRequest          = pb.QueryDisclosureRequestRequest
	QueryDisclosureRequestResponse         = pb.QueryDisclosureRequestResponse
	QueryPendingDisclosureRequestsRequest  = pb.QueryPendingDisclosureRequestsRequest
	QueryPendingDisclosureRequestsResponse = pb.QueryPendingDisclosureRequestsResponse
	QueryParseVoiceCommandRequest          = pb.QueryParseVoiceCommandRequest
	QueryParseVoiceCommandResponse         = pb.QueryParseVoiceCommandResponse
	QueryParamsRequest                     = pb.QueryParamsRequest
	QueryParamsResponse                    = pb.QueryParamsResponse

	// Genesis types
	GenesisState = pb.GenesisState

	// Event types
	EventVCMinted                  = pb.EventVCMinted
	EventVCRevoked                 = pb.EventVCRevoked
	EventVCSuspended               = pb.EventVCSuspended
	EventVCReactivated             = pb.EventVCReactivated
	EventVCExpired                 = pb.EventVCExpired
	EventVCPolicyCreated           = pb.EventVCPolicyCreated
	EventVCPolicyUpdated           = pb.EventVCPolicyUpdated
	EventVCPolicyDeprecated        = pb.EventVCPolicyDeprecated
	EventPresentationCreated       = pb.EventPresentationCreated
	EventPresentationVerified      = pb.EventPresentationVerified
	EventDIDRegistered             = pb.EventDIDRegistered
	EventDIDUpdated                = pb.EventDIDUpdated
	EventAttributeVCCreated        = pb.EventAttributeVCCreated
	EventAttributeVCRevoked        = pb.EventAttributeVCRevoked
	EventDisclosurePolicyUpdated   = pb.EventDisclosurePolicyUpdated
	EventDisclosureRequestCreated  = pb.EventDisclosureRequestCreated
	EventDisclosureResponseCreated = pb.EventDisclosureResponseCreated
	EventMerkleRootUpdated         = pb.EventMerkleRootUpdated

	// Wrapper types for genesis map values
	PresentationIds = pb.PresentationIds
	AttributeVcIds  = pb.AttributeVcIds
	RequestIds      = pb.RequestIds
)

// Re-export enum values for VCType
const (
	VCType_VC_TYPE_UNSPECIFIED          = pb.VCType_VC_TYPE_UNSPECIFIED
	VCType_VC_TYPE_VERIFIED_HUMAN       = pb.VCType_VC_TYPE_VERIFIED_HUMAN
	VCType_VC_TYPE_AGE_OVER_18          = pb.VCType_VC_TYPE_AGE_OVER_18
	VCType_VC_TYPE_AGE_OVER_21          = pb.VCType_VC_TYPE_AGE_OVER_21
	VCType_VC_TYPE_RESIDENT_OF          = pb.VCType_VC_TYPE_RESIDENT_OF
	VCType_VC_TYPE_BIOMETRIC_AUTH       = pb.VCType_VC_TYPE_BIOMETRIC_AUTH
	VCType_VC_TYPE_KYC_VERIFICATION     = pb.VCType_VC_TYPE_KYC_VERIFICATION
	VCType_VC_TYPE_NOTARY_PUBLIC        = pb.VCType_VC_TYPE_NOTARY_PUBLIC
	VCType_VC_TYPE_PROFESSIONAL_LICENSE = pb.VCType_VC_TYPE_PROFESSIONAL_LICENSE
	VCType_VC_TYPE_BIOMETRIC_FOCUS      = pb.VCType_VC_TYPE_BIOMETRIC_FOCUS
	VCType_VC_TYPE_SOCIAL_FOCUS         = pb.VCType_VC_TYPE_SOCIAL_FOCUS
	VCType_VC_TYPE_GEOLOCATION_FOCUS    = pb.VCType_VC_TYPE_GEOLOCATION_FOCUS
	VCType_VC_TYPE_HIGH_ASSURANCE_FOCUS = pb.VCType_VC_TYPE_HIGH_ASSURANCE_FOCUS
	VCType_VC_TYPE_POSSESSION_FOCUS     = pb.VCType_VC_TYPE_POSSESSION_FOCUS
	VCType_VC_TYPE_KNOWLEDGE_FOCUS      = pb.VCType_VC_TYPE_KNOWLEDGE_FOCUS
	VCType_VC_TYPE_PERSISTENCE_FOCUS    = pb.VCType_VC_TYPE_PERSISTENCE_FOCUS
	VCType_VC_TYPE_SPECIALIZED_FOCUS    = pb.VCType_VC_TYPE_SPECIALIZED_FOCUS
	VCType_VC_TYPE_CUSTOM               = pb.VCType_VC_TYPE_CUSTOM
)

// Re-export enum values for VCStatus
const (
	VCStatus_VC_STATUS_UNSPECIFIED = pb.VCStatus_VC_STATUS_UNSPECIFIED
	VCStatus_VC_STATUS_PENDING     = pb.VCStatus_VC_STATUS_PENDING
	VCStatus_VC_STATUS_ACTIVE      = pb.VCStatus_VC_STATUS_ACTIVE
	VCStatus_VC_STATUS_REVOKED     = pb.VCStatus_VC_STATUS_REVOKED
	VCStatus_VC_STATUS_EXPIRED     = pb.VCStatus_VC_STATUS_EXPIRED
	VCStatus_VC_STATUS_SUSPENDED   = pb.VCStatus_VC_STATUS_SUSPENDED
)

// Re-export enum values for RevocationReason
const (
	RevocationReason_REVOCATION_REASON_UNSPECIFIED         = pb.RevocationReason_REVOCATION_REASON_UNSPECIFIED
	RevocationReason_REVOCATION_REASON_USER_REQUEST        = pb.RevocationReason_REVOCATION_REASON_USER_REQUEST
	RevocationReason_REVOCATION_REASON_FRAUD_DETECTED      = pb.RevocationReason_REVOCATION_REASON_FRAUD_DETECTED
	RevocationReason_REVOCATION_REASON_CS_BELOW_THRESHOLD  = pb.RevocationReason_REVOCATION_REASON_CS_BELOW_THRESHOLD
	RevocationReason_REVOCATION_REASON_IR_INVALIDATED      = pb.RevocationReason_REVOCATION_REASON_IR_INVALIDATED
	RevocationReason_REVOCATION_REASON_EXPIRED             = pb.RevocationReason_REVOCATION_REASON_EXPIRED
	RevocationReason_REVOCATION_REASON_GOVERNANCE          = pb.RevocationReason_REVOCATION_REASON_GOVERNANCE
	RevocationReason_REVOCATION_REASON_SECURITY_COMPROMISE = pb.RevocationReason_REVOCATION_REASON_SECURITY_COMPROMISE
	RevocationReason_REVOCATION_REASON_POLICY_CHANGE       = pb.RevocationReason_REVOCATION_REASON_POLICY_CHANGE
)

// Re-export enum values for VCPolicyStatus
const (
	VCPolicyStatus_VC_POLICY_STATUS_UNSPECIFIED = pb.VCPolicyStatus_VC_POLICY_STATUS_UNSPECIFIED
	VCPolicyStatus_VC_POLICY_STATUS_DRAFT       = pb.VCPolicyStatus_VC_POLICY_STATUS_DRAFT
	VCPolicyStatus_VC_POLICY_STATUS_ACTIVE      = pb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE
	VCPolicyStatus_VC_POLICY_STATUS_DEPRECATED  = pb.VCPolicyStatus_VC_POLICY_STATUS_DEPRECATED
)

// Re-export enum values for DisclosurePolicyMode
const (
	DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_DENY        = pb.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_DENY
	DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_ASK         = pb.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_ASK
	DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_ALLOW       = pb.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_ALLOW
	DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_CONDITIONAL = pb.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_CONDITIONAL
)

// Re-export enum values for AttributeType
const (
	AttributeType_ATTRIBUTE_TYPE_UNSPECIFIED          = pb.AttributeType_ATTRIBUTE_TYPE_UNSPECIFIED
	AttributeType_ATTRIBUTE_TYPE_FULL_NAME            = pb.AttributeType_ATTRIBUTE_TYPE_FULL_NAME
	AttributeType_ATTRIBUTE_TYPE_FIRST_NAME           = pb.AttributeType_ATTRIBUTE_TYPE_FIRST_NAME
	AttributeType_ATTRIBUTE_TYPE_LAST_NAME            = pb.AttributeType_ATTRIBUTE_TYPE_LAST_NAME
	AttributeType_ATTRIBUTE_TYPE_DATE_OF_BIRTH        = pb.AttributeType_ATTRIBUTE_TYPE_DATE_OF_BIRTH
	AttributeType_ATTRIBUTE_TYPE_AGE                  = pb.AttributeType_ATTRIBUTE_TYPE_AGE
	AttributeType_ATTRIBUTE_TYPE_GENDER               = pb.AttributeType_ATTRIBUTE_TYPE_GENDER
	AttributeType_ATTRIBUTE_TYPE_EMAIL                = pb.AttributeType_ATTRIBUTE_TYPE_EMAIL
	AttributeType_ATTRIBUTE_TYPE_PHONE                = pb.AttributeType_ATTRIBUTE_TYPE_PHONE
	AttributeType_ATTRIBUTE_TYPE_ADDRESS_FULL         = pb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_FULL
	AttributeType_ATTRIBUTE_TYPE_ADDRESS_STREET       = pb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_STREET
	AttributeType_ATTRIBUTE_TYPE_ADDRESS_CITY         = pb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_CITY
	AttributeType_ATTRIBUTE_TYPE_ADDRESS_STATE        = pb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_STATE
	AttributeType_ATTRIBUTE_TYPE_ADDRESS_ZIP          = pb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_ZIP
	AttributeType_ATTRIBUTE_TYPE_ADDRESS_COUNTRY      = pb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_COUNTRY
	AttributeType_ATTRIBUTE_TYPE_PASSPORT_NUMBER      = pb.AttributeType_ATTRIBUTE_TYPE_PASSPORT_NUMBER
	AttributeType_ATTRIBUTE_TYPE_DRIVERS_LICENSE      = pb.AttributeType_ATTRIBUTE_TYPE_DRIVERS_LICENSE
	AttributeType_ATTRIBUTE_TYPE_SSN                  = pb.AttributeType_ATTRIBUTE_TYPE_SSN
	AttributeType_ATTRIBUTE_TYPE_TAX_ID               = pb.AttributeType_ATTRIBUTE_TYPE_TAX_ID
	AttributeType_ATTRIBUTE_TYPE_HEIGHT               = pb.AttributeType_ATTRIBUTE_TYPE_HEIGHT
	AttributeType_ATTRIBUTE_TYPE_WEIGHT               = pb.AttributeType_ATTRIBUTE_TYPE_WEIGHT
	AttributeType_ATTRIBUTE_TYPE_EYE_COLOR            = pb.AttributeType_ATTRIBUTE_TYPE_EYE_COLOR
	AttributeType_ATTRIBUTE_TYPE_HAIR_COLOR           = pb.AttributeType_ATTRIBUTE_TYPE_HAIR_COLOR
	AttributeType_ATTRIBUTE_TYPE_OCCUPATION           = pb.AttributeType_ATTRIBUTE_TYPE_OCCUPATION
	AttributeType_ATTRIBUTE_TYPE_EMPLOYER             = pb.AttributeType_ATTRIBUTE_TYPE_EMPLOYER
	AttributeType_ATTRIBUTE_TYPE_PROFESSIONAL_LICENSE = pb.AttributeType_ATTRIBUTE_TYPE_PROFESSIONAL_LICENSE
	AttributeType_ATTRIBUTE_TYPE_EDUCATION_LEVEL      = pb.AttributeType_ATTRIBUTE_TYPE_EDUCATION_LEVEL
	AttributeType_ATTRIBUTE_TYPE_DEGREE               = pb.AttributeType_ATTRIBUTE_TYPE_DEGREE
	AttributeType_ATTRIBUTE_TYPE_SCUBA_CERTIFIED      = pb.AttributeType_ATTRIBUTE_TYPE_SCUBA_CERTIFIED
	AttributeType_ATTRIBUTE_TYPE_PILOTS_LICENSE       = pb.AttributeType_ATTRIBUTE_TYPE_PILOTS_LICENSE
	AttributeType_ATTRIBUTE_TYPE_SECURITY_CLEARANCE   = pb.AttributeType_ATTRIBUTE_TYPE_SECURITY_CLEARANCE
	AttributeType_ATTRIBUTE_TYPE_CUSTOM               = pb.AttributeType_ATTRIBUTE_TYPE_CUSTOM
)
