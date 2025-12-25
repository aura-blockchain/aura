// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

func TestNewKYCSubmittedEvent(t *testing.T) {
	addr := "cosmos1test"
	verificationID := "ver123"
	documentType := "passport"
	blockHeight := int64(100)
	blockTime := "2024-01-01T00:00:00Z"

	attrs := types.NewKYCSubmittedEvent(addr, verificationID, documentType, blockHeight, blockTime)

	require.NotNil(t, attrs)
	require.Equal(t, addr, attrs[types.AttributeKeyAddress])
	require.Equal(t, verificationID, attrs[types.AttributeKeyVerificationID])
	require.Equal(t, documentType, attrs[types.AttributeKeyDocumentType])
	require.Equal(t, "pending", attrs[types.AttributeKeyKYCStatus])
}

func TestNewKYCApprovedEvent(t *testing.T) {
	addr := "cosmos1test"
	verificationID := "ver123"
	kycLevel := "basic"
	approvedBy := "provider1"
	riskScore := 0.5
	blockHeight := int64(100)
	blockTime := "2024-01-01T00:00:00Z"

	attrs := types.NewKYCApprovedEvent(addr, verificationID, kycLevel, approvedBy, riskScore, blockHeight, blockTime)

	require.NotNil(t, attrs)
	require.Equal(t, addr, attrs[types.AttributeKeyAddress])
	require.Equal(t, verificationID, attrs[types.AttributeKeyVerificationID])
	require.Equal(t, kycLevel, attrs[types.AttributeKeyKYCLevel])
	require.Equal(t, "approved", attrs[types.AttributeKeyKYCStatus])
}

func TestNewKYCRejectedEvent(t *testing.T) {
	addr := "cosmos1test"
	verificationID := "ver123"
	reason := "Insufficient documentation"
	blockHeight := int64(100)
	blockTime := "2024-01-01T00:00:00Z"

	attrs := types.NewKYCRejectedEvent(addr, verificationID, reason, blockHeight, blockTime)

	require.NotNil(t, attrs)
	require.Equal(t, addr, attrs[types.AttributeKeyAddress])
	require.Equal(t, reason, attrs[types.AttributeKeyRejectionReason])
	require.Equal(t, "rejected", attrs[types.AttributeKeyKYCStatus])
}

func TestNewSanctionsScreeningEvent(t *testing.T) {
	addr := "cosmos1test"
	flagged := false
	matchCount := 0
	screeningResult := "clear"
	blockHeight := int64(100)
	blockTime := "2024-01-01T00:00:00Z"

	attrs := types.NewSanctionsScreeningEvent(addr, flagged, matchCount, screeningResult, blockHeight, blockTime)

	require.NotNil(t, attrs)
	require.Equal(t, addr, attrs[types.AttributeKeyAddress])
	require.Equal(t, screeningResult, attrs[types.AttributeKeyScreeningResult])
}

func TestNewSanctionsFlaggedEvent(t *testing.T) {
	addr := "cosmos1sanctioned"
	matchCount := 1
	sanctionsList := "OFAC SDN"
	blockHeight := int64(100)
	blockTime := "2024-01-01T00:00:00Z"

	attrs := types.NewSanctionsFlaggedEvent(addr, matchCount, sanctionsList, blockHeight, blockTime)

	require.NotNil(t, attrs)
	require.Equal(t, addr, attrs[types.AttributeKeyAddress])
	require.Equal(t, sanctionsList, attrs[types.AttributeKeySanctionsList])
}

func TestNewGDPRRequestEvent(t *testing.T) {
	requester := "cosmos1test"
	requestID := "req123"
	requestType := "access"
	blockHeight := int64(100)
	blockTime := "2024-01-01T00:00:00Z"

	attrs := types.NewGDPRRequestEvent(requester, requestID, requestType, blockHeight, blockTime)

	require.NotNil(t, attrs)
	require.Equal(t, requester, attrs[types.AttributeKeyRequester])
	require.Equal(t, requestID, attrs[types.AttributeKeyRequestID])
	require.Equal(t, requestType, attrs[types.AttributeKeyGDPRRequestType])
}

func TestNewGDPRCompletedEvent(t *testing.T) {
	requester := "cosmos1test"
	requestID := "req123"
	requestType := "erasure"
	dataExported := true
	dataDeleted := true
	blockHeight := int64(100)
	blockTime := "2024-01-01T00:00:00Z"

	attrs := types.NewGDPRCompletedEvent(requester, requestID, requestType, dataExported, dataDeleted, blockHeight, blockTime)

	require.NotNil(t, attrs)
	require.Equal(t, requestID, attrs[types.AttributeKeyRequestID])
	require.Equal(t, requester, attrs[types.AttributeKeyRequester])
	require.Equal(t, "true", attrs[types.AttributeKeyDataExported])
	require.Equal(t, "true", attrs[types.AttributeKeyDataDeleted])
}

func TestNewTaxReportGeneratedEvent(t *testing.T) {
	addr := "cosmos1test"
	taxYear := "2024"
	jurisdiction := "US"
	reportType := "annual"
	transactionCount := 100
	blockHeight := int64(100)
	blockTime := "2024-01-01T00:00:00Z"

	attrs := types.NewTaxReportGeneratedEvent(addr, taxYear, jurisdiction, reportType, transactionCount, blockHeight, blockTime)

	require.NotNil(t, attrs)
	require.Equal(t, addr, attrs[types.AttributeKeyAddress])
	require.Equal(t, taxYear, attrs[types.AttributeKeyTaxYear])
	require.Equal(t, jurisdiction, attrs[types.AttributeKeyJurisdiction])
	require.Equal(t, reportType, attrs[types.AttributeKeyReportType])
}

func TestNewComplianceViolationEvent(t *testing.T) {
	addr := "cosmos1violator"
	violationType := "missing_kyc"
	severity := "high"
	blockHeight := int64(100)
	blockTime := "2024-01-01T00:00:00Z"

	attrs := types.NewComplianceViolationEvent(addr, violationType, severity, blockHeight, blockTime)

	require.NotNil(t, attrs)
	require.Equal(t, addr, attrs[types.AttributeKeyAddress])
	require.Equal(t, violationType, attrs[types.AttributeKeyViolationType])
	require.Equal(t, severity, attrs[types.AttributeKeyViolationSeverity])
}

func TestNewSARReportedEvent(t *testing.T) {
	activityID := "activity123"
	addr := "cosmos1suspicious"
	reporter := "cosmos1reporter"
	activityType := "structuring"
	transactionHash := "hash123"
	blockHeight := int64(100)
	blockTime := "2024-01-01T00:00:00Z"

	attrs := types.NewSARReportedEvent(activityID, addr, reporter, activityType, transactionHash, blockHeight, blockTime)

	require.NotNil(t, attrs)
	require.Equal(t, activityID, attrs[types.AttributeKeyActivityID])
	require.Equal(t, addr, attrs[types.AttributeKeyAddress])
	require.Equal(t, reporter, attrs[types.AttributeKeyReporter])
	require.Equal(t, activityType, attrs[types.AttributeKeyActivityType])
}

func TestNewGDPRConsentRecordedEvent(t *testing.T) {
	addr := "cosmos1test"
	consentType := "data_processing"
	consented := true
	consentVersion := "v1.0"
	blockHeight := int64(100)
	blockTime := "2024-01-01T00:00:00Z"

	attrs := types.NewGDPRConsentRecordedEvent(addr, consentType, consented, consentVersion, blockHeight, blockTime)

	require.NotNil(t, attrs)
	require.Equal(t, addr, attrs[types.AttributeKeyAddress])
	require.Equal(t, consentType, attrs[types.AttributeKeyConsentType])
	require.Equal(t, consentVersion, attrs[types.AttributeKeyConsentVersion])
}

func TestNewGDPRDataRequestedEvent(t *testing.T) {
	requestID := "req456"
	addr := "cosmos1test"
	requestType := "portability"
	blockHeight := int64(100)
	blockTime := "2024-01-01T00:00:00Z"

	attrs := types.NewGDPRDataRequestedEvent(requestID, addr, requestType, blockHeight, blockTime)

	require.NotNil(t, attrs)
	require.Equal(t, requestID, attrs[types.AttributeKeyRequestID])
	require.Equal(t, addr, attrs[types.AttributeKeyAddress])
	require.Equal(t, requestType, attrs[types.AttributeKeyGDPRRequestType])
}

func TestNewGDPRDataErasedEvent(t *testing.T) {
	addr := "cosmos1test"
	erasureEventID := "erasure1"
	erasureReason := "user_request"
	erasureTime := "2024-01-01T00:00:00Z"
	blockHeight := int64(100)
	blockTime := "2024-01-01T00:00:00Z"

	attrs := types.NewGDPRDataErasedEvent(addr, erasureEventID, erasureReason, erasureTime, blockHeight, blockTime)

	require.NotNil(t, attrs)
	require.Equal(t, addr, attrs[types.AttributeKeyAddress])
	require.Equal(t, erasureEventID, attrs[types.AttributeKeyErasureEventID])
	require.Equal(t, erasureReason, attrs[types.AttributeKeyErasureReason])
	require.Equal(t, erasureTime, attrs[types.AttributeKeyErasureTime])
}

func TestEventConstants(t *testing.T) {
	// Test that all event type constants are defined
	require.NotEmpty(t, types.EventTypeKYCSubmitted)
	require.NotEmpty(t, types.EventTypeKYCApproved)
	require.NotEmpty(t, types.EventTypeKYCRejected)
	require.NotEmpty(t, types.EventTypeSanctionsScreening)
	require.NotEmpty(t, types.EventTypeSanctionsFlagged)
	require.NotEmpty(t, types.EventTypeGDPRRequest)
	require.NotEmpty(t, types.EventTypeGDPRCompleted)
	require.NotEmpty(t, types.EventTypeTaxReportGenerated)
	require.NotEmpty(t, types.EventTypeComplianceViolation)
	require.NotEmpty(t, types.EventTypeSARReported)
}

func TestAttributeKeyConstants(t *testing.T) {
	// Test that all attribute key constants are defined
	require.NotEmpty(t, types.AttributeKeyAddress)
	require.NotEmpty(t, types.AttributeKeyVerificationID)
	require.NotEmpty(t, types.AttributeKeyKYCStatus)
	require.NotEmpty(t, types.AttributeKeyKYCLevel)
	require.NotEmpty(t, types.AttributeKeyRequestID)
	require.NotEmpty(t, types.AttributeKeyConsentType)
	require.NotEmpty(t, types.AttributeKeyJurisdiction)
}
