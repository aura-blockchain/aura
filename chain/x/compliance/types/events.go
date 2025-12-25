// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import "fmt"

// Event types for the compliance module
const (
	EventTypeKYCSubmitted              = "kyc_submitted"
	EventTypeKYCApproved               = "kyc_approved"
	EventTypeKYCRejected               = "kyc_rejected"
	EventTypeKYCExpired                = "kyc_expired"
	EventTypeSARReported               = "suspicious_activity_reported"
	EventTypeSanctionsScreening        = "sanctions_screening"
	EventTypeSanctionsFlagged          = "sanctions_flagged"
	EventTypeGDPRConsentRecorded       = "gdpr_consent_recorded"
	EventTypeGDPRConsentWithdrawn      = "gdpr_consent_withdrawn"
	EventTypeGDPRDataRequested         = "gdpr_data_requested"
	EventTypeGDPRDataErased            = "gdpr_data_erased"
	EventTypeGDPRRequest               = "gdpr_request"
	EventTypeGDPRCompleted             = "gdpr_completed"
	EventTypeTaxReportGenerated        = "tax_report_generated"
	EventTypeComplianceViolation       = "compliance_violation"
	EventTypeParamsUpdated             = "params_updated"
	EventTypeTransactionAlert          = "transaction_alert"
	EventTypeRiskLevelChanged          = "risk_level_changed"
	EventTypeAMLProfileUpdated         = "aml_profile_updated"
	EventTypeRateLimitExceeded         = "rate_limit_exceeded"
)

// Event attribute keys
const (
	AttributeKeyAddress           = "address"
	AttributeKeyVerificationID    = "verification_id"
	AttributeKeyKYCStatus         = "kyc_status"
	AttributeKeyKYCLevel          = "kyc_level"
	AttributeKeyDocumentType      = "document_type"
	AttributeKeyRiskScore         = "risk_score"
	AttributeKeyApprovedBy        = "approved_by"
	AttributeKeyRejectionReason   = "rejection_reason"
	AttributeKeyProvider          = "provider"
	AttributeKeyPIICommitment     = "pii_commitment"
	AttributeKeyActivityID        = "activity_id"
	AttributeKeyActivityType      = "activity_type"
	AttributeKeyReporter          = "reporter"
	AttributeKeyTransactionHash   = "transaction_hash"
	AttributeKeyScreeningResult   = "screening_result"
	AttributeKeyStatus            = "status"
	AttributeKeyRequiresReview    = "requires_review"
	AttributeKeyFlagged           = "flagged"
	AttributeKeyMatchCount        = "match_count"
	AttributeKeySanctionsList     = "sanctions_list"
	AttributeKeyConsentType       = "consent_type"
	AttributeKeyConsented         = "consented"
	AttributeKeyConsentVersion    = "consent_version"
	AttributeKeyGDPRRequestType   = "gdpr_request_type"
	AttributeKeyRequester         = "requester"
	AttributeKeyRequestID         = "request_id"
	AttributeKeyErasureEventID    = "erasure_event_id"
	AttributeKeyErasureReason     = "erasure_reason"
	AttributeKeyErasureTime       = "erasure_time"
	AttributeKeyDataExported      = "data_exported"
	AttributeKeyDataDeleted       = "data_deleted"
	AttributeKeyReportID          = "report_id"
	AttributeKeyTaxYear           = "tax_year"
	AttributeKeyJurisdiction      = "jurisdiction"
	AttributeKeyReportType        = "report_type"
	AttributeKeyTransactionCount  = "transaction_count"
	AttributeKeyViolationType       = "violation_type"
	AttributeKeyViolationSeverity   = "violation_severity"
	AttributeKeyProcessingRestricted = "processing_restricted"
	AttributeKeyDeletionTriggered   = "deletion_triggered"
	AttributeKeyBlockHeight         = "block_height"
	AttributeKeyBlockTime           = "block_time"
	AttributeKeyTimestamp           = "timestamp"
	AttributeKeyAlertID             = "alert_id"
	AttributeKeyRuleID              = "rule_id"
	AttributeKeyRiskLevel           = "risk_level"
	AttributeKeyDescription         = "description"
	AttributeKeyPreviousRisk        = "previous_risk"
	AttributeKeyNewRisk             = "new_risk"
	AttributeKeyTotalVolume         = "total_volume"
	AttributeKeyAmount              = "amount"
	AttributeKeyOperation           = "operation"
	AttributeKeyCount               = "count"
	AttributeKeyLimit               = "limit"
	AttributeKeyWindowStart         = "window_start"
)

// Helper functions for creating event attributes

// NewKYCSubmittedEvent creates attributes for KYC submission
func NewKYCSubmittedEvent(
	address, verificationID, documentType string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyAddress:        address,
		AttributeKeyVerificationID: verificationID,
		AttributeKeyDocumentType:   documentType,
		AttributeKeyKYCStatus:      "pending",
		AttributeKeyBlockHeight:    formatInt64(blockHeight),
		AttributeKeyBlockTime:      blockTime,
	}
}

// NewKYCApprovedEvent creates attributes for KYC approval
func NewKYCApprovedEvent(
	address, verificationID, kycLevel, approvedBy string,
	riskScore float64,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyAddress:        address,
		AttributeKeyVerificationID: verificationID,
		AttributeKeyKYCLevel:       kycLevel,
		AttributeKeyKYCStatus:      "approved",
		AttributeKeyApprovedBy:     approvedBy,
		AttributeKeyRiskScore:      formatFloat64(riskScore),
		AttributeKeyBlockHeight:    formatInt64(blockHeight),
		AttributeKeyBlockTime:      blockTime,
	}
}

// NewKYCRejectedEvent creates attributes for KYC rejection
func NewKYCRejectedEvent(
	address, verificationID, rejectionReason string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyAddress:         address,
		AttributeKeyVerificationID:  verificationID,
		AttributeKeyKYCStatus:       "rejected",
		AttributeKeyRejectionReason: rejectionReason,
		AttributeKeyBlockHeight:     formatInt64(blockHeight),
		AttributeKeyBlockTime:       blockTime,
	}
}

// NewSanctionsScreeningEvent creates attributes for sanctions screening
func NewSanctionsScreeningEvent(
	address string,
	flagged bool,
	matchCount int,
	screeningResult string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyAddress:         address,
		AttributeKeyFlagged:         formatBool(flagged),
		AttributeKeyMatchCount:      formatInt(matchCount),
		AttributeKeyScreeningResult: screeningResult,
		AttributeKeyBlockHeight:     formatInt64(blockHeight),
		AttributeKeyBlockTime:       blockTime,
	}
}

// NewSanctionsFlaggedEvent creates attributes for sanctions flagging
func NewSanctionsFlaggedEvent(
	address string,
	matchCount int,
	sanctionsList string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyAddress:       address,
		AttributeKeyFlagged:       "true",
		AttributeKeyMatchCount:    formatInt(matchCount),
		AttributeKeySanctionsList: sanctionsList,
		AttributeKeyBlockHeight:   formatInt64(blockHeight),
		AttributeKeyBlockTime:     blockTime,
	}
}

// NewGDPRRequestEvent creates attributes for GDPR request
func NewGDPRRequestEvent(
	requester, requestID, requestType string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyRequester:       requester,
		AttributeKeyRequestID:       requestID,
		AttributeKeyGDPRRequestType: requestType,
		AttributeKeyBlockHeight:     formatInt64(blockHeight),
		AttributeKeyBlockTime:       blockTime,
	}
}

// NewGDPRCompletedEvent creates attributes for GDPR completion
func NewGDPRCompletedEvent(
	requester, requestID, requestType string,
	dataExported, dataDeleted bool,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyRequester:       requester,
		AttributeKeyRequestID:       requestID,
		AttributeKeyGDPRRequestType: requestType,
		AttributeKeyDataExported:    formatBool(dataExported),
		AttributeKeyDataDeleted:     formatBool(dataDeleted),
		AttributeKeyBlockHeight:     formatInt64(blockHeight),
		AttributeKeyBlockTime:       blockTime,
	}
}

// NewTaxReportGeneratedEvent creates attributes for tax report generation
func NewTaxReportGeneratedEvent(
	address, taxYear, jurisdiction, reportType string,
	transactionCount int,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyAddress:          address,
		AttributeKeyTaxYear:          taxYear,
		AttributeKeyJurisdiction:     jurisdiction,
		AttributeKeyReportType:       reportType,
		AttributeKeyTransactionCount: formatInt(transactionCount),
		AttributeKeyBlockHeight:      formatInt64(blockHeight),
		AttributeKeyBlockTime:        blockTime,
	}
}

// NewComplianceViolationEvent creates attributes for compliance violation
func NewComplianceViolationEvent(
	address, violationType, violationSeverity string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyAddress:           address,
		AttributeKeyViolationType:     violationType,
		AttributeKeyViolationSeverity: violationSeverity,
		AttributeKeyBlockHeight:       formatInt64(blockHeight),
		AttributeKeyBlockTime:         blockTime,
	}
}

// NewSARReportedEvent creates attributes for suspicious activity report
func NewSARReportedEvent(
	activityID, address, reporter, activityType, transactionHash string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyActivityID:      activityID,
		AttributeKeyAddress:         address,
		AttributeKeyReporter:        reporter,
		AttributeKeyActivityType:    activityType,
		AttributeKeyTransactionHash: transactionHash,
		AttributeKeyBlockHeight:     formatInt64(blockHeight),
		AttributeKeyBlockTime:       blockTime,
	}
}

// NewGDPRConsentRecordedEvent creates attributes for GDPR consent recording
func NewGDPRConsentRecordedEvent(
	address, consentType string,
	consented bool,
	consentVersion string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyAddress:        address,
		AttributeKeyConsentType:    consentType,
		AttributeKeyConsented:      formatBool(consented),
		AttributeKeyConsentVersion: consentVersion,
		AttributeKeyBlockHeight:    formatInt64(blockHeight),
		AttributeKeyBlockTime:      blockTime,
	}
}

// NewGDPRDataRequestedEvent creates attributes for GDPR data request
func NewGDPRDataRequestedEvent(
	requestID, address, requestType string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyRequestID:       requestID,
		AttributeKeyAddress:         address,
		AttributeKeyGDPRRequestType: requestType,
		AttributeKeyBlockHeight:     formatInt64(blockHeight),
		AttributeKeyBlockTime:       blockTime,
	}
}

// NewGDPRDataErasedEvent creates attributes for GDPR data erasure
func NewGDPRDataErasedEvent(
	address, erasureEventID, erasureReason, erasureTime string,
	blockHeight int64, blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyAddress:        address,
		AttributeKeyErasureEventID: erasureEventID,
		AttributeKeyErasureReason:  erasureReason,
		AttributeKeyErasureTime:    erasureTime,
		AttributeKeyBlockHeight:    formatInt64(blockHeight),
		AttributeKeyBlockTime:      blockTime,
	}
}

// Helper formatting functions

func formatInt(i int) string {
	return fmt.Sprintf("%d", i)
}

func formatInt64(i int64) string {
	return fmt.Sprintf("%d", i)
}

func formatFloat64(f float64) string {
	return fmt.Sprintf("%.6f", f)
}

func formatBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
