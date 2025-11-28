package types

import "fmt"

// Event types for the compliance module
const (
	EventTypeKYCSubmitted        = "kyc_submitted"
	EventTypeKYCApproved         = "kyc_approved"
	EventTypeKYCRejected         = "kyc_rejected"
	EventTypeKYCExpired          = "kyc_expired"
	EventTypeSanctionsScreening  = "sanctions_screening"
	EventTypeSanctionsFlagged    = "sanctions_flagged"
	EventTypeGDPRRequest         = "gdpr_request"
	EventTypeGDPRCompleted       = "gdpr_completed"
	EventTypeTaxReportGenerated  = "tax_report_generated"
	EventTypeComplianceViolation = "compliance_violation"
	EventTypeParamsUpdated       = "params_updated"
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
	AttributeKeyScreeningResult   = "screening_result"
	AttributeKeyFlagged           = "flagged"
	AttributeKeyMatchCount        = "match_count"
	AttributeKeySanctionsList     = "sanctions_list"
	AttributeKeyGDPRRequestType   = "gdpr_request_type"
	AttributeKeyRequester         = "requester"
	AttributeKeyRequestID         = "request_id"
	AttributeKeyDataExported      = "data_exported"
	AttributeKeyDataDeleted       = "data_deleted"
	AttributeKeyTaxYear           = "tax_year"
	AttributeKeyJurisdiction      = "jurisdiction"
	AttributeKeyReportType        = "report_type"
	AttributeKeyTransactionCount  = "transaction_count"
	AttributeKeyViolationType     = "violation_type"
	AttributeKeyViolationSeverity = "violation_severity"
	AttributeKeyBlockHeight       = "block_height"
	AttributeKeyBlockTime         = "block_time"
	AttributeKeyTimestamp         = "timestamp"
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
