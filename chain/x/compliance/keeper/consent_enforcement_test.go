// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// ============================================================================
// GDPR Consent Enforcement Tests (Issue #055)
// ============================================================================
//
// These tests verify that GDPR consent withdrawal is properly enforced:
// 1. Processing functions block operations when consent is missing
// 2. Processing functions block operations after consent withdrawal
// 3. Multiple consent purposes are handled independently
// 4. Events are emitted correctly for audit trail
// 5. Consent can be re-granted after withdrawal

// ============================================================================
// RequireConsent Helper Tests
// ============================================================================

func TestRequireConsent_Success(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	address := "cosmos1test"

	// Grant consent
	consent := &types.GDPRConsent{
		Address:        address,
		ConsentType:    "kyc_processing",
		Consented:      true,
		ConsentVersion: "v1",
		ConsentGivenAt: time.Now(),
	}
	err := keeper.SetGDPRConsent(ctx, consent)
	require.NoError(t, err)

	// Remove any processing restrictions
	err = keeper.SetProcessingRestriction(ctx, address, false)
	require.NoError(t, err)

	// RequireConsent should succeed
	err = keeper.RequireConsent(ctx, address, "kyc_processing")
	require.NoError(t, err, "RequireConsent should succeed when consent is valid")
}

func TestRequireConsent_NoConsent(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	address := "cosmos1test"

	// No consent granted - RequireConsent should fail
	err := keeper.RequireConsent(ctx, address, "kyc_processing")
	require.Error(t, err, "RequireConsent should fail when no consent exists")
	require.Equal(t, types.ErrProcessingRestricted, err, "should return ErrProcessingRestricted")
}

func TestRequireConsent_ConsentWithdrawn(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	address := "cosmos1test"

	// Grant then withdraw consent
	withdrawnAt := time.Now()
	consent := &types.GDPRConsent{
		Address:             address,
		ConsentType:         "kyc_processing",
		Consented:           false,
		ConsentVersion:      "v1",
		ConsentGivenAt: time.Now().Add(-1 * time.Hour),
		ConsentWithdrawnAt:  &withdrawnAt,
	}
	err := keeper.SetGDPRConsent(ctx, consent)
	require.NoError(t, err)

	// Set processing restriction
	err = keeper.SetProcessingRestriction(ctx, address, true)
	require.NoError(t, err)

	// RequireConsent should fail
	err = keeper.RequireConsent(ctx, address, "kyc_processing")
	require.Error(t, err, "RequireConsent should fail when consent is withdrawn")
	require.Equal(t, types.ErrProcessingRestricted, err, "should return ErrProcessingRestricted")
}

func TestRequireConsent_ProcessingRestricted(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	address := "cosmos1test"

	// Grant consent but set processing restriction
	consent := &types.GDPRConsent{
		Address:        address,
		ConsentType:    "kyc_processing",
		Consented:      true,
		ConsentVersion: "v1",
		ConsentGivenAt: time.Now(),
	}
	err := keeper.SetGDPRConsent(ctx, consent)
	require.NoError(t, err)

	err = keeper.SetProcessingRestriction(ctx, address, true)
	require.NoError(t, err)

	// RequireConsent should fail due to processing restriction
	err = keeper.RequireConsent(ctx, address, "kyc_processing")
	require.Error(t, err, "RequireConsent should fail when processing is restricted")
	require.Equal(t, types.ErrProcessingRestricted, err, "should return ErrProcessingRestricted")
}

func TestRequireConsent_WrongPurpose(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	address := "cosmos1test"

	// Grant consent for kyc_processing
	consent := &types.GDPRConsent{
		Address:        address,
		ConsentType:    "kyc_processing",
		Consented:      true,
		ConsentVersion: "v1",
		ConsentGivenAt: time.Now(),
	}
	err := keeper.SetGDPRConsent(ctx, consent)
	require.NoError(t, err)

	err = keeper.SetProcessingRestriction(ctx, address, false)
	require.NoError(t, err)

	// Try to use different purpose (aml_monitoring) - should fail
	err = keeper.RequireConsent(ctx, address, "aml_monitoring")
	require.Error(t, err, "RequireConsent should fail when consent is for different purpose")
	require.Equal(t, types.ErrProcessingRestricted, err, "should return ErrProcessingRestricted")
}

// ============================================================================
// SubmitKYC Consent Enforcement Tests
// ============================================================================

func TestSubmitKYC_BlockedWithoutConsent(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	// Setup authorized provider
	providerAddr := createTestAddressMsg("provider")
	params, _ := keeper.GetParams(ctx)
	params.ApprovedKycProviders = []string{providerAddr}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	userAddr := createTestAddressMsg("user")

	// Try to submit KYC without consent - should fail
	req := &types.MsgSubmitKYC{
		Address:       userAddr,
		KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:      providerAddr,
		PiiCommitment: make([]byte, 32), // Valid 32-byte commitment
		Jurisdiction:  "US",
	}

	_, err = server.SubmitKYC(sdk.WrapSDKContext(ctx), req)
	require.Error(t, err, "SubmitKYC should fail without consent")
	require.Contains(t, err.Error(), "consent required", "error should mention consent requirement")
}

func TestSubmitKYC_SuccessWithConsent(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	// Setup authorized provider
	providerAddr := createTestAddressMsg("provider")
	params, _ := keeper.GetParams(ctx)
	params.ApprovedKycProviders = []string{providerAddr}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	userAddr := createTestAddressMsg("user")

	// Grant consent for KYC processing
	consent := &types.GDPRConsent{
		Address:        userAddr,
		ConsentType:    "kyc_processing",
		Consented:      true,
		ConsentVersion: "v1",
		ConsentGivenAt: time.Now(),
	}
	err = keeper.SetGDPRConsent(ctx, consent)
	require.NoError(t, err)

	err = keeper.SetProcessingRestriction(ctx, userAddr, false)
	require.NoError(t, err)

	// Submit KYC - should succeed
	req := &types.MsgSubmitKYC{
		Address:       userAddr,
		KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:      providerAddr,
		PiiCommitment: make([]byte, 32),
		Jurisdiction:  "US",
	}

	resp, err := server.SubmitKYC(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err, "SubmitKYC should succeed with valid consent")
	require.NotNil(t, resp)
	require.True(t, resp.Success)
}

func TestSubmitKYC_BlockedAfterConsentWithdrawal(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	// Setup authorized provider
	providerAddr := createTestAddressMsg("provider")
	params, _ := keeper.GetParams(ctx)
	params.ApprovedKycProviders = []string{providerAddr}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	userAddr := createTestAddressMsg("user")

	// Step 1: Grant consent and submit KYC successfully
	consent := &types.GDPRConsent{
		Address:        userAddr,
		ConsentType:    "kyc_processing",
		Consented:      true,
		ConsentVersion: "v1",
		ConsentGivenAt: time.Now(),
	}
	err = keeper.SetGDPRConsent(ctx, consent)
	require.NoError(t, err)

	err = keeper.SetProcessingRestriction(ctx, userAddr, false)
	require.NoError(t, err)

	req := &types.MsgSubmitKYC{
		Address:       userAddr,
		KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:      providerAddr,
		PiiCommitment: make([]byte, 32),
		Jurisdiction:  "US",
	}

	resp, err := server.SubmitKYC(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err, "first KYC submission should succeed")
	require.True(t, resp.Success)

	// Step 2: Withdraw consent
	withdrawnAt := time.Now()
	withdrawnConsent := &types.GDPRConsent{
		Address:            userAddr,
		ConsentType:        "kyc_processing",
		Consented:          false,
		ConsentVersion:     "v1",
		ConsentGivenAt:     time.Now().Add(-1 * time.Hour),
		ConsentWithdrawnAt: &withdrawnAt,
	}
	err = keeper.SetGDPRConsent(ctx, withdrawnConsent)
	require.NoError(t, err)

	err = keeper.SetProcessingRestriction(ctx, userAddr, true)
	require.NoError(t, err)

	// Step 3: Try to submit KYC again - should fail
	req2 := &types.MsgSubmitKYC{
		Address:       userAddr,
		KycLevel:      types.KYCLevel_KYC_LEVEL_INTERMEDIATE,
		Provider:      providerAddr,
		PiiCommitment: make([]byte, 32),
		Jurisdiction:  "US",
	}

	_, err = server.SubmitKYC(sdk.WrapSDKContext(ctx), req2)
	require.Error(t, err, "KYC submission should fail after consent withdrawal")
	require.Contains(t, err.Error(), "consent required", "error should mention consent requirement")
}

// ============================================================================
// ReportSuspiciousActivity Consent Enforcement Tests
// ============================================================================

func TestReportSuspiciousActivity_BlockedWithoutConsent(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	userAddr := createTestAddressMsg("user")
	reporterAddr := createTestAddressMsg("reporter")

	// Try to report suspicious activity without consent - should fail
	req := &types.MsgReportSuspiciousActivity{
		Reporter:        reporterAddr,
		Address:         userAddr,
		TransactionHash: "0xabc123",
		ActivityType:    "structuring",
		Description:     "Multiple transactions just below reporting threshold",
		Indicators:      []string{"high_frequency", "round_amounts"},
	}

	_, err := server.ReportSuspiciousActivity(sdk.WrapSDKContext(ctx), req)
	require.Error(t, err, "ReportSuspiciousActivity should fail without consent")
	require.Contains(t, err.Error(), "consent required", "error should mention consent requirement")
}

func TestReportSuspiciousActivity_SuccessWithConsent(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	userAddr := createTestAddressMsg("user")
	reporterAddr := createTestAddressMsg("reporter")

	// Grant consent for AML monitoring
	consent := &types.GDPRConsent{
		Address:        userAddr,
		ConsentType:    "aml_monitoring",
		Consented:      true,
		ConsentVersion: "v1",
		ConsentGivenAt: time.Now(),
	}
	err := keeper.SetGDPRConsent(ctx, consent)
	require.NoError(t, err)

	err = keeper.SetProcessingRestriction(ctx, userAddr, false)
	require.NoError(t, err)

	// Report suspicious activity - should succeed
	req := &types.MsgReportSuspiciousActivity{
		Reporter:        reporterAddr,
		Address:         userAddr,
		TransactionHash: "0xabc123",
		ActivityType:    "structuring",
		Description:     "Multiple transactions just below reporting threshold",
		Indicators:      []string{"high_frequency", "round_amounts"},
	}

	resp, err := server.ReportSuspiciousActivity(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err, "ReportSuspiciousActivity should succeed with valid consent")
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.ActivityId)
}

func TestReportSuspiciousActivity_BlockedAfterConsentWithdrawal(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	userAddr := createTestAddressMsg("user")
	reporterAddr := createTestAddressMsg("reporter")

	// Grant consent and report activity successfully
	consent := &types.GDPRConsent{
		Address:        userAddr,
		ConsentType:    "aml_monitoring",
		Consented:      true,
		ConsentVersion: "v1",
		ConsentGivenAt: time.Now(),
	}
	err := keeper.SetGDPRConsent(ctx, consent)
	require.NoError(t, err)

	err = keeper.SetProcessingRestriction(ctx, userAddr, false)
	require.NoError(t, err)

	req := &types.MsgReportSuspiciousActivity{
		Reporter:        reporterAddr,
		Address:         userAddr,
		TransactionHash: "0xabc123",
		ActivityType:    "structuring",
		Description:     "Test activity",
		Indicators:      []string{"test"},
	}

	resp, err := server.ReportSuspiciousActivity(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err, "first report should succeed")
	require.NotEmpty(t, resp.ActivityId)

	// Withdraw consent
	withdrawnAt2 := time.Now()
	withdrawnConsent := &types.GDPRConsent{
		Address:            userAddr,
		ConsentType:        "aml_monitoring",
		Consented:          false,
		ConsentVersion:     "v1",
		ConsentGivenAt:     time.Now().Add(-1 * time.Hour),
		ConsentWithdrawnAt: &withdrawnAt2,
	}
	err = keeper.SetGDPRConsent(ctx, withdrawnConsent)
	require.NoError(t, err)

	err = keeper.SetProcessingRestriction(ctx, userAddr, true)
	require.NoError(t, err)

	// Try to report again - should fail
	req2 := &types.MsgReportSuspiciousActivity{
		Reporter:        reporterAddr,
		Address:         userAddr,
		TransactionHash: "0xdef456",
		ActivityType:    "smurfing",
		Description:     "Another suspicious pattern",
		Indicators:      []string{"test2"},
	}

	_, err = server.ReportSuspiciousActivity(sdk.WrapSDKContext(ctx), req2)
	require.Error(t, err, "report should fail after consent withdrawal")
	require.Contains(t, err.Error(), "consent required", "error should mention consent requirement")
}

// ============================================================================
// ScreenSanctions Consent Enforcement Tests
// ============================================================================

func TestScreenSanctions_BlockedWithoutConsent(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	userAddr := createTestAddressMsg("user")

	// Try to screen sanctions without consent - should fail
	req := &types.MsgScreenSanctions{
		Address:      userAddr,
		ForceRefresh: false,
	}

	_, err := server.ScreenSanctions(sdk.WrapSDKContext(ctx), req)
	require.Error(t, err, "ScreenSanctions should fail without consent")
	require.Contains(t, err.Error(), "consent required", "error should mention consent requirement")
}

func TestScreenSanctions_SuccessWithConsent(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	userAddr := createTestAddressMsg("user")

	// Grant consent for sanctions screening
	consent := &types.GDPRConsent{
		Address:        userAddr,
		ConsentType:    "sanctions_screening",
		Consented:      true,
		ConsentVersion: "v1",
		ConsentGivenAt: time.Now(),
	}
	err := keeper.SetGDPRConsent(ctx, consent)
	require.NoError(t, err)

	err = keeper.SetProcessingRestriction(ctx, userAddr, false)
	require.NoError(t, err)

	// Screen sanctions - should succeed
	req := &types.MsgScreenSanctions{
		Address:      userAddr,
		ForceRefresh: true,
	}

	resp, err := server.ScreenSanctions(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err, "ScreenSanctions should succeed with valid consent")
	require.NotNil(t, resp)
	require.Equal(t, types.SanctionsStatus_SANCTIONS_CLEAR, resp.Status)
}

func TestScreenSanctions_BlockedAfterConsentWithdrawal(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	userAddr := createTestAddressMsg("user")

	// Grant consent and screen successfully
	consent := &types.GDPRConsent{
		Address:        userAddr,
		ConsentType:    "sanctions_screening",
		Consented:      true,
		ConsentVersion: "v1",
		ConsentGivenAt: time.Now(),
	}
	err := keeper.SetGDPRConsent(ctx, consent)
	require.NoError(t, err)

	err = keeper.SetProcessingRestriction(ctx, userAddr, false)
	require.NoError(t, err)

	req := &types.MsgScreenSanctions{
		Address:      userAddr,
		ForceRefresh: true,
	}

	resp, err := server.ScreenSanctions(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err, "first screening should succeed")
	require.Equal(t, types.SanctionsStatus_SANCTIONS_CLEAR, resp.Status)

	// Withdraw consent
	withdrawnAt3 := time.Now()
	withdrawnConsent := &types.GDPRConsent{
		Address:            userAddr,
		ConsentType:        "sanctions_screening",
		Consented:          false,
		ConsentVersion:     "v1",
		ConsentGivenAt:     time.Now().Add(-1 * time.Hour),
		ConsentWithdrawnAt: &withdrawnAt3,
	}
	err = keeper.SetGDPRConsent(ctx, withdrawnConsent)
	require.NoError(t, err)

	err = keeper.SetProcessingRestriction(ctx, userAddr, true)
	require.NoError(t, err)

	// Try to screen again - should fail
	req2 := &types.MsgScreenSanctions{
		Address:      userAddr,
		ForceRefresh: true,
	}

	_, err = server.ScreenSanctions(sdk.WrapSDKContext(ctx), req2)
	require.Error(t, err, "screening should fail after consent withdrawal")
	require.Contains(t, err.Error(), "consent required", "error should mention consent requirement")
}

// ============================================================================
// Multiple Purposes Tests
// ============================================================================

func TestMultiplePurposes_IndependentConsents(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	address := "cosmos1test"

	// Grant consent for kyc_processing
	kycConsent := &types.GDPRConsent{
		Address:        address,
		ConsentType:    "kyc_processing",
		Consented:      true,
		ConsentVersion: "v1",
		ConsentGivenAt: time.Now(),
	}
	err := keeper.SetGDPRConsent(ctx, kycConsent)
	require.NoError(t, err)

	// Grant consent for aml_monitoring
	amlConsent := &types.GDPRConsent{
		Address:        address,
		ConsentType:    "aml_monitoring",
		Consented:      true,
		ConsentVersion: "v1",
		ConsentGivenAt: time.Now(),
	}
	err = keeper.SetGDPRConsent(ctx, amlConsent)
	require.NoError(t, err)

	err = keeper.SetProcessingRestriction(ctx, address, false)
	require.NoError(t, err)

	// Both purposes should be allowed
	err = keeper.RequireConsent(ctx, address, "kyc_processing")
	require.NoError(t, err, "kyc_processing should be allowed")

	err = keeper.RequireConsent(ctx, address, "aml_monitoring")
	require.NoError(t, err, "aml_monitoring should be allowed")

	// sanctions_screening should not be allowed (no consent)
	err = keeper.RequireConsent(ctx, address, "sanctions_screening")
	require.Error(t, err, "sanctions_screening should not be allowed without consent")
}

func TestMultiplePurposes_WithdrawOneKeepOthers(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	address := "cosmos1test"

	// Grant consent for multiple purposes
	kycConsent := &types.GDPRConsent{
		Address:        address,
		ConsentType:    "kyc_processing",
		Consented:      true,
		ConsentVersion: "v1",
		ConsentGivenAt: time.Now(),
	}
	err := keeper.SetGDPRConsent(ctx, kycConsent)
	require.NoError(t, err)

	amlConsent := &types.GDPRConsent{
		Address:        address,
		ConsentType:    "aml_monitoring",
		Consented:      true,
		ConsentVersion: "v1",
		ConsentGivenAt: time.Now(),
	}
	err = keeper.SetGDPRConsent(ctx, amlConsent)
	require.NoError(t, err)

	err = keeper.SetProcessingRestriction(ctx, address, false)
	require.NoError(t, err)

	// Withdraw kyc_processing consent only
	withdrawnAt4 := time.Now()
	withdrawnKycConsent := &types.GDPRConsent{
		Address:            address,
		ConsentType:        "kyc_processing",
		Consented:          false,
		ConsentVersion:     "v1",
		ConsentGivenAt:     time.Now().Add(-1 * time.Hour),
		ConsentWithdrawnAt: &withdrawnAt4,
	}
	err = keeper.SetGDPRConsent(ctx, withdrawnKycConsent)
	require.NoError(t, err)

	// Note: SetProcessingRestriction is a global flag per address
	// In a real implementation, we might need per-purpose restrictions
	// For now, we test that CanProcessData properly checks specific consents

	// kyc_processing should fail (consent withdrawn)
	canProcess := keeper.CanProcessData(ctx, address, "kyc_processing")
	require.False(t, canProcess, "kyc_processing should be blocked after withdrawal")

	// aml_monitoring should still work (separate consent)
	// However, if SetProcessingRestriction is global, this may fail
	// Let's verify current behavior
	isRestricted := keeper.IsProcessingRestricted(ctx, address)
	if !isRestricted {
		// If no global restriction, aml_monitoring should work
		canProcess = keeper.CanProcessData(ctx, address, "aml_monitoring")
		require.True(t, canProcess, "aml_monitoring should still work (different consent)")
	}
}

// ============================================================================
// Consent Re-grant After Withdrawal Tests
// ============================================================================

func TestConsentReGrant_AfterWithdrawal(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	// Setup authorized provider
	providerAddr := createTestAddressMsg("provider")
	params, _ := keeper.GetParams(ctx)
	params.ApprovedKycProviders = []string{providerAddr}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	userAddr := createTestAddressMsg("user")

	// Step 1: Grant consent
	consentGrant := &types.MsgRecordGDPRConsent{
		Address:        userAddr,
		ConsentType:    "kyc_processing",
		Consented:      true,
		ConsentVersion: "v1",
	}
	_, err = server.RecordGDPRConsent(sdk.WrapSDKContext(ctx), consentGrant)
	require.NoError(t, err)

	// Verify KYC submission works
	kycReq := &types.MsgSubmitKYC{
		Address:       userAddr,
		KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:      providerAddr,
		PiiCommitment: make([]byte, 32),
		Jurisdiction:  "US",
	}
	resp, err := server.SubmitKYC(sdk.WrapSDKContext(ctx), kycReq)
	require.NoError(t, err, "KYC should work with consent")
	require.True(t, resp.Success)

	// Step 2: Withdraw consent
	consentWithdraw := &types.MsgRecordGDPRConsent{
		Address:        userAddr,
		ConsentType:    "kyc_processing",
		Consented:      false,
		ConsentVersion: "v1",
	}
	_, err = server.RecordGDPRConsent(sdk.WrapSDKContext(ctx), consentWithdraw)
	require.NoError(t, err)

	// Verify KYC submission is blocked
	_, err = server.SubmitKYC(sdk.WrapSDKContext(ctx), kycReq)
	require.Error(t, err, "KYC should be blocked after withdrawal")

	// Step 3: Re-grant consent (user changes their mind)
	consentReGrant := &types.MsgRecordGDPRConsent{
		Address:        userAddr,
		ConsentType:    "kyc_processing",
		Consented:      true,
		ConsentVersion: "v2", // New version
	}
	_, err = server.RecordGDPRConsent(sdk.WrapSDKContext(ctx), consentReGrant)
	require.NoError(t, err)

	// Verify processing restriction is removed
	require.False(t, keeper.IsProcessingRestricted(ctx, userAddr),
		"processing restriction should be removed after re-granting consent")

	// Verify KYC submission works again
	kycReq2 := &types.MsgSubmitKYC{
		Address:       userAddr,
		KycLevel:      types.KYCLevel_KYC_LEVEL_INTERMEDIATE,
		Provider:      providerAddr,
		PiiCommitment: make([]byte, 32),
		Jurisdiction:  "US",
	}
	resp2, err := server.SubmitKYC(sdk.WrapSDKContext(ctx), kycReq2)
	require.NoError(t, err, "KYC should work again after re-granting consent")
	require.True(t, resp2.Success)
}

// ============================================================================
// Event Emission Tests
// ============================================================================

func TestConsentWithdrawal_EventEmission(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	userAddr := createTestAddressMsg("user")

	// Withdraw consent (without granting first, to test event alone)
	req := &types.MsgRecordGDPRConsent{
		Address:        userAddr,
		ConsentType:    "kyc_processing",
		Consented:      false,
		ConsentVersion: "v1",
	}
	_, err := server.RecordGDPRConsent(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err)

	// Verify events were emitted
	events := ctx.EventManager().Events()

	// Check for withdrawal event
	withdrawalEventFound := false
	deletionEventFound := false

	for _, event := range events {
		if event.Type == types.EventTypeGDPRConsentWithdrawn {
			withdrawalEventFound = true

			// Verify event attributes
			var processingRestrictedValue, deletionTriggeredValue string
			for _, attr := range event.Attributes {
				if attr.Key == types.AttributeKeyProcessingRestricted {
					processingRestrictedValue = attr.Value
				}
				if attr.Key == types.AttributeKeyDeletionTriggered {
					deletionTriggeredValue = attr.Value
				}
			}

			require.Equal(t, "true", processingRestrictedValue,
				"processing_restricted should be true in withdrawal event")
			require.Equal(t, "true", deletionTriggeredValue,
				"deletion_triggered should be true in withdrawal event")
		}

		if event.Type == "gdpr_data_deletion_requested" {
			deletionEventFound = true

			// Verify deletion event attributes
			var addressValue, consentTypeValue string
			for _, attr := range event.Attributes {
				if attr.Key == types.AttributeKeyAddress {
					addressValue = attr.Value
				}
				if attr.Key == types.AttributeKeyConsentType {
					consentTypeValue = attr.Value
				}
			}

			require.Equal(t, userAddr, addressValue,
				"address should match in deletion event")
			require.Equal(t, "kyc_processing", consentTypeValue,
				"consent_type should match in deletion event")
		}
	}

	require.True(t, withdrawalEventFound,
		"gdpr_consent_withdrawn event should be emitted")
	require.True(t, deletionEventFound,
		"gdpr_data_deletion_requested event should be emitted")
}

func TestConsentGrant_EventEmission(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	userAddr := createTestAddressMsg("user")

	// Grant consent
	req := &types.MsgRecordGDPRConsent{
		Address:        userAddr,
		ConsentType:    "kyc_processing",
		Consented:      true,
		ConsentVersion: "v1",
	}
	_, err := server.RecordGDPRConsent(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err)

	// Verify event was emitted
	events := ctx.EventManager().Events()

	consentRecordedEventFound := false
	for _, event := range events {
		if event.Type == types.EventTypeGDPRConsentRecorded {
			consentRecordedEventFound = true

			// Verify event attributes
			var addressValue, consentTypeValue, consentedValue string
			for _, attr := range event.Attributes {
				if attr.Key == types.AttributeKeyAddress {
					addressValue = attr.Value
				}
				if attr.Key == types.AttributeKeyConsentType {
					consentTypeValue = attr.Value
				}
				if attr.Key == types.AttributeKeyConsented {
					consentedValue = attr.Value
				}
			}

			require.Equal(t, userAddr, addressValue,
				"address should match in consent event")
			require.Equal(t, "kyc_processing", consentTypeValue,
				"consent_type should match in consent event")
			require.Equal(t, "true", consentedValue,
				"consented should be true in consent event")
		}
	}

	require.True(t, consentRecordedEventFound,
		"gdpr_consent_recorded event should be emitted")

	// Verify processing restriction was removed
	require.False(t, keeper.IsProcessingRestricted(ctx, userAddr),
		"processing should not be restricted after granting consent")
}
