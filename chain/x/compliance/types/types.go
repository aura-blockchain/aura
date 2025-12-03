package types

import pb "github.com/aequitas/aura/proto/aura/compliance/v1beta1"

// Re-export all proto types
type (
	// Enums
	KYCLevel             = pb.KYCLevel
	AMLRiskLevel         = pb.AMLRiskLevel
	TransactionRiskLevel = pb.TransactionRiskLevel
	SanctionsStatus      = pb.SanctionsStatus

	// Core types
	KYCRecord                 = pb.KYCRecord
	AMLProfile                = pb.AMLProfile
	SuspiciousActivity        = pb.SuspiciousActivity
	TransactionMonitoringRule = pb.TransactionMonitoringRule
	TransactionAlert          = pb.TransactionAlert
	TransactionAlertList      = pb.TransactionAlertList
	SanctionsScreeningResult  = pb.SanctionsScreeningResult
	SanctionsMatch            = pb.SanctionsMatch
	GDPRConsent               = pb.GDPRConsent
	GDPRConsentList           = pb.GDPRConsentList
	GDPRDataRequest           = pb.GDPRDataRequest
	TaxReport                 = pb.TaxReport
	TaxReportList             = pb.TaxReportList
	TaxTransaction            = pb.TaxTransaction
	ComplianceParams          = pb.ComplianceParams
	RateLimitEntry            = pb.RateLimitEntry

	// Query types
	QueryKYCRecordRequest           = pb.QueryKYCRecordRequest
	QueryKYCRecordResponse          = pb.QueryKYCRecordResponse
	QueryAMLProfileRequest          = pb.QueryAMLProfileRequest
	QueryAMLProfileResponse         = pb.QueryAMLProfileResponse
	QuerySanctionsScreeningRequest  = pb.QuerySanctionsScreeningRequest
	QuerySanctionsScreeningResponse = pb.QuerySanctionsScreeningResponse
	QueryTransactionAlertsRequest   = pb.QueryTransactionAlertsRequest
	QueryTransactionAlertsResponse  = pb.QueryTransactionAlertsResponse
	QueryTaxReportRequest           = pb.QueryTaxReportRequest
	QueryTaxReportResponse          = pb.QueryTaxReportResponse

	// Message types
	MsgSubmitKYC                        = pb.MsgSubmitKYC
	MsgSubmitKYCResponse                = pb.MsgSubmitKYCResponse
	MsgReportSuspiciousActivity         = pb.MsgReportSuspiciousActivity
	MsgReportSuspiciousActivityResponse = pb.MsgReportSuspiciousActivityResponse
	MsgScreenSanctions                  = pb.MsgScreenSanctions
	MsgScreenSanctionsResponse          = pb.MsgScreenSanctionsResponse
	MsgRecordGDPRConsent                = pb.MsgRecordGDPRConsent
	MsgRecordGDPRConsentResponse        = pb.MsgRecordGDPRConsentResponse
	MsgRequestGDPRData                  = pb.MsgRequestGDPRData
	MsgRequestGDPRDataResponse          = pb.MsgRequestGDPRDataResponse
	MsgEraseGDPRData                    = pb.MsgEraseGDPRData
	MsgEraseGDPRDataResponse            = pb.MsgEraseGDPRDataResponse
	MsgGenerateTaxReport                = pb.MsgGenerateTaxReport
	MsgGenerateTaxReportResponse        = pb.MsgGenerateTaxReportResponse

	// Service interfaces
	MsgServer                = pb.MsgServer
	QueryServer              = pb.QueryServer
	UnimplementedMsgServer   = pb.UnimplementedMsgServer
	UnimplementedQueryServer = pb.UnimplementedQueryServer
)

// Re-export enum values for KYCLevel
const (
	KYCLevel_KYC_LEVEL_UNSPECIFIED  = pb.KYCLevel_KYC_LEVEL_UNSPECIFIED
	KYCLevel_KYC_LEVEL_NONE         = pb.KYCLevel_KYC_LEVEL_NONE
	KYCLevel_KYC_LEVEL_BASIC        = pb.KYCLevel_KYC_LEVEL_BASIC
	KYCLevel_KYC_LEVEL_INTERMEDIATE = pb.KYCLevel_KYC_LEVEL_INTERMEDIATE
	KYCLevel_KYC_LEVEL_ADVANCED     = pb.KYCLevel_KYC_LEVEL_ADVANCED
)

// Re-export enum values for AMLRiskLevel
const (
	AMLRiskLevel_AML_RISK_UNSPECIFIED = pb.AMLRiskLevel_AML_RISK_UNSPECIFIED
	AMLRiskLevel_AML_RISK_LOW         = pb.AMLRiskLevel_AML_RISK_LOW
	AMLRiskLevel_AML_RISK_MEDIUM      = pb.AMLRiskLevel_AML_RISK_MEDIUM
	AMLRiskLevel_AML_RISK_HIGH        = pb.AMLRiskLevel_AML_RISK_HIGH
	AMLRiskLevel_AML_RISK_SEVERE      = pb.AMLRiskLevel_AML_RISK_SEVERE
)

// Re-export enum values for TransactionRiskLevel
const (
	TransactionRiskLevel_TX_RISK_UNSPECIFIED = pb.TransactionRiskLevel_TX_RISK_UNSPECIFIED
	TransactionRiskLevel_TX_RISK_LOW         = pb.TransactionRiskLevel_TX_RISK_LOW
	TransactionRiskLevel_TX_RISK_MEDIUM      = pb.TransactionRiskLevel_TX_RISK_MEDIUM
	TransactionRiskLevel_TX_RISK_HIGH        = pb.TransactionRiskLevel_TX_RISK_HIGH
	TransactionRiskLevel_TX_RISK_CRITICAL    = pb.TransactionRiskLevel_TX_RISK_CRITICAL
)

// Re-export enum values for SanctionsStatus
const (
	SanctionsStatus_SANCTIONS_UNSPECIFIED    = pb.SanctionsStatus_SANCTIONS_UNSPECIFIED
	SanctionsStatus_SANCTIONS_CLEAR          = pb.SanctionsStatus_SANCTIONS_CLEAR
	SanctionsStatus_SANCTIONS_MATCH          = pb.SanctionsStatus_SANCTIONS_MATCH
	SanctionsStatus_SANCTIONS_CONFIRMED      = pb.SanctionsStatus_SANCTIONS_CONFIRMED
	SanctionsStatus_SANCTIONS_PENDING_REVIEW = pb.SanctionsStatus_SANCTIONS_PENDING_REVIEW
)
