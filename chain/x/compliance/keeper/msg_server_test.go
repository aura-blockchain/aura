package keeper

import (
	"crypto/sha256"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// createTestAddress creates a valid Bech32 address for testing (helper reused from events test)
func createTestAddressMsg(name string) string {
	hash := sha256.Sum256([]byte(name))
	addr := sdk.AccAddress(hash[:20])
	return addr.String()
}

// grantConsent is a helper to grant GDPR consent for testing
func grantConsent(t *testing.T, keeper *Keeper, ctx sdk.Context, address string, purpose string) {
	consent := &types.GDPRConsent{
		Address:        address,
		ConsentType:    purpose,
		Consented:      true,
		ConsentVersion: "v1",
		ConsentGivenAt: time.Now(),
	}
	err := keeper.SetGDPRConsent(ctx, consent)
	require.NoError(t, err)
	err = keeper.SetProcessingRestriction(ctx, address, false)
	require.NoError(t, err)
}

func TestMsgSubmitKYCStoresRecord(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	// Setup approved provider
	providerAddr := createTestAddressMsg("provider")
	params, _ := keeper.GetParams(ctx)
	params.ApprovedKycProviders = []string{providerAddr}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	piiCommitment := make([]byte, 32) // SHA-256 hash
	copy(piiCommitment, []byte("test_commitment_hash_32_bytes"))

	userAddr := createTestAddressMsg("kyc_user")
	grantConsent(t, keeper, ctx, userAddr, "kyc_processing")

	req := &types.MsgSubmitKYC{
		Address:       userAddr,
		KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:      providerAddr,
		PiiCommitment: piiCommitment,
		Jurisdiction:  "US",
	}
	_, err = server.SubmitKYC(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err)
	record, err := keeper.GetKYCRecord(ctx, req.Address)
	require.NoError(t, err)
	require.Equal(t, req.PiiCommitment, record.PiiCommitment)
}

func TestMsgReportSuspiciousActivityPersisted(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	userAddr := createTestAddressMsg("user")
	grantConsent(t, keeper, ctx, userAddr, "aml_monitoring")

	req := &types.MsgReportSuspiciousActivity{
		Reporter:        createTestAddressMsg("reporter"),
		Address:         userAddr,
		TransactionHash: "hash",
		ActivityType:    "structuring",
		Description:     "many tx",
	}
	resp, err := server.ReportSuspiciousActivity(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err)
	require.NotEmpty(t, resp.ActivityId)
	activity, err := keeper.GetSuspiciousActivity(ctx, resp.ActivityId)
	require.NoError(t, err)
	require.Equal(t, req.ActivityType, activity.ActivityType)
}

func TestMsgScreenSanctionsStoresResult(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	userAddr := createTestAddressMsg("sanction_user")
	grantConsent(t, keeper, ctx, userAddr, "sanctions_screening")

	req := &types.MsgScreenSanctions{Address: userAddr}
	resp, err := server.ScreenSanctions(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err)
	require.Equal(t, types.SanctionsStatus_SANCTIONS_CLEAR, resp.Status)
	result, err := keeper.GetSanctionsResult(ctx, req.Address)
	require.NoError(t, err)
	require.Equal(t, req.Address, result.Address)
}

func TestMsgRecordGDPRConsentPersists(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)
	req := &types.MsgRecordGDPRConsent{
		Address:        createTestAddressMsg("gdpr_user"),
		ConsentType:    "data_processing",
		Consented:      true,
		ConsentVersion: "v1",
	}
	_, err := server.RecordGDPRConsent(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err)
	consents, err := keeper.GetGDPRConsents(ctx, req.Address)
	require.NoError(t, err)
	require.Len(t, consents, 1)
	require.Equal(t, req.ConsentType, consents[0].ConsentType)
}

func TestMsgRequestGDPRDataCreatesEntry(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)
	req := &types.MsgRequestGDPRData{
		Address:     createTestAddressMsg("gdpr_requester"),
		RequestType: "access",
	}
	resp, err := server.RequestGDPRData(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err)
	require.NotEmpty(t, resp.RequestId)
	stored, err := keeper.GetGDPRRequest(ctx, resp.RequestId)
	require.NoError(t, err)
	require.Equal(t, req.Address, stored.Address)
}

func TestMsgGenerateTaxReportCreatesReport(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)
	req := &types.MsgGenerateTaxReport{
		Address:      createTestAddressMsg("tax_user"),
		TaxYear:      "2024",
		Jurisdiction: "US",
		ReportType:   "1099",
	}
	resp, err := server.GenerateTaxReport(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err)
	require.NotEmpty(t, resp.ReportId)
	reports, err := keeper.GetTaxReports(ctx, req.Address)
	require.NoError(t, err)
	require.Len(t, reports, 1)
	require.Equal(t, req.TaxYear, reports[0].TaxYear)
}

func TestMsgScreenSanctionsUsesProviderWhenAvailable(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	provider := &testSanctionsProvider{}
	keeper.RegisterSanctionsProvider("mock", provider)
	server := NewMsgServer(keeper)

	userAddr := createTestAddressMsg("provider_user")
	grantConsent(t, keeper, ctx, userAddr, "sanctions_screening")

	resp, err := server.ScreenSanctions(sdk.WrapSDKContext(ctx), &types.MsgScreenSanctions{Address: userAddr})
	require.NoError(t, err)
	require.Equal(t, types.SanctionsStatus_SANCTIONS_CLEAR, resp.Status)
}

// testSanctionsProvider mirrors the keeper tests but ensures timestamps are set.
type testSanctionsProvider struct{}

func (m *testSanctionsProvider) ScreenAddress(address string) (*types.SanctionsScreeningResult, error) {
	return &types.SanctionsScreeningResult{
		Address:              address,
		Status:               types.SanctionsStatus_SANCTIONS_CLEAR,
		Matches:              []*types.SanctionsMatch{},
		ScreenedAt: time.Now(),
		RequiresManualReview: false,
	}, nil
}

func (m *testSanctionsProvider) CheckLists(_ []string) ([]*types.SanctionsMatch, error) {
	return []*types.SanctionsMatch{}, nil
}

// ============================================================================
// GDPR Consent Withdrawal Enforcement Tests (TODO 055)
// ============================================================================

func TestMsgRecordGDPRConsent_WithdrawalEnforcesProcessingRestriction(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	address := createTestAddressMsg("test_withdrawal")
	consentType := "data_processing"

	// Step 1: Give consent
	reqGive := &types.MsgRecordGDPRConsent{
		Address:        address,
		ConsentType:    consentType,
		Consented:      true,
		ConsentVersion: "v1",
	}
	_, err := server.RecordGDPRConsent(sdk.WrapSDKContext(ctx), reqGive)
	require.NoError(t, err)

	// Verify consent is stored and processing is allowed
	consent, found := keeper.GetGDPRConsent(ctx, address, consentType)
	require.True(t, found)
	require.True(t, consent.Consented)
	require.True(t, keeper.CanProcessData(ctx, address, consentType))
	require.False(t, keeper.IsProcessingRestricted(ctx, address))

	// Step 2: Withdraw consent
	reqWithdraw := &types.MsgRecordGDPRConsent{
		Address:        address,
		ConsentType:    consentType,
		Consented:      false,
		ConsentVersion: "v1",
	}
	_, err = server.RecordGDPRConsent(sdk.WrapSDKContext(ctx), reqWithdraw)
	require.NoError(t, err)

	// Verify withdrawal enforcement
	consent, found = keeper.GetGDPRConsent(ctx, address, consentType)
	require.True(t, found)
	require.False(t, consent.Consented, "consent should be withdrawn")
	require.NotNil(t, consent.ConsentWithdrawnAt, "withdrawal timestamp should be set")

	// Critical: Verify processing is now restricted
	require.True(t, keeper.IsProcessingRestricted(ctx, address), "processing should be restricted after withdrawal")
	require.False(t, keeper.CanProcessData(ctx, address, consentType), "data processing should be blocked")
}

func TestMsgRecordGDPRConsent_WithdrawalEmitsEnforcementEvent(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	address := createTestAddressMsg("test_event")
	consentType := "data_processing"

	// Withdraw consent
	req := &types.MsgRecordGDPRConsent{
		Address:        address,
		ConsentType:    consentType,
		Consented:      false,
		ConsentVersion: "v1",
	}
	_, err := server.RecordGDPRConsent(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err)

	// Verify withdrawal event was emitted
	events := ctx.EventManager().Events()
	withdrawalEventFound := false
	deletionEventFound := false

	for _, event := range events {
		if event.Type == types.EventTypeGDPRConsentWithdrawn {
			withdrawalEventFound = true

			// Verify event attributes - event.Attributes are ABCI types
			var processingRestrictedValue, deletionTriggeredValue string
			for _, attr := range event.Attributes {
				if attr.Key == types.AttributeKeyProcessingRestricted {
					processingRestrictedValue = attr.Value
				}
				if attr.Key == types.AttributeKeyDeletionTriggered {
					deletionTriggeredValue = attr.Value
				}
			}

			require.NotEmpty(t, processingRestrictedValue, "processing_restricted attribute should be present")
			require.Equal(t, "true", processingRestrictedValue, "processing_restricted should be true")
			require.NotEmpty(t, deletionTriggeredValue, "deletion_triggered attribute should be present")
			require.Equal(t, "true", deletionTriggeredValue, "deletion_triggered should be true")
		}

		if event.Type == "gdpr_data_deletion_requested" {
			deletionEventFound = true

			// Verify deletion event attributes - event.Attributes are ABCI types
			var addressValue, consentTypeValue string
			for _, attr := range event.Attributes {
				if attr.Key == types.AttributeKeyAddress {
					addressValue = attr.Value
				}
				if attr.Key == types.AttributeKeyConsentType {
					consentTypeValue = attr.Value
				}
			}

			require.Equal(t, address, addressValue, "address should match")
			require.Equal(t, consentType, consentTypeValue, "consent type should match")
		}
	}

	require.True(t, withdrawalEventFound, "GDPR consent withdrawn event should be emitted")
	require.True(t, deletionEventFound, "Data deletion requested event should be emitted")
}

func TestMsgRecordGDPRConsent_GivingConsentRemovesRestriction(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	address := createTestAddressMsg("test_removal")
	consentType := "data_processing"

	// First withdraw consent
	reqWithdraw := &types.MsgRecordGDPRConsent{
		Address:        address,
		ConsentType:    consentType,
		Consented:      false,
		ConsentVersion: "v1",
	}
	_, err := server.RecordGDPRConsent(sdk.WrapSDKContext(ctx), reqWithdraw)
	require.NoError(t, err)

	// Verify restriction is set
	require.True(t, keeper.IsProcessingRestricted(ctx, address))

	// Now give consent
	reqGive := &types.MsgRecordGDPRConsent{
		Address:        address,
		ConsentType:    consentType,
		Consented:      true,
		ConsentVersion: "v1",
	}
	_, err = server.RecordGDPRConsent(sdk.WrapSDKContext(ctx), reqGive)
	require.NoError(t, err)

	// Verify restriction is removed
	require.False(t, keeper.IsProcessingRestricted(ctx, address), "processing restriction should be removed when consent is given")
	require.True(t, keeper.CanProcessData(ctx, address, consentType), "data processing should be allowed")
}

func TestMsgRecordGDPRConsent_MultipleConsentTypes(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	address := createTestAddressMsg("test_multiple")

	// Give consent for data_processing
	req1 := &types.MsgRecordGDPRConsent{
		Address:        address,
		ConsentType:    "data_processing",
		Consented:      true,
		ConsentVersion: "v1",
	}
	_, err := server.RecordGDPRConsent(sdk.WrapSDKContext(ctx), req1)
	require.NoError(t, err)

	// Give consent for marketing
	req2 := &types.MsgRecordGDPRConsent{
		Address:        address,
		ConsentType:    "marketing",
		Consented:      true,
		ConsentVersion: "v1",
	}
	_, err = server.RecordGDPRConsent(sdk.WrapSDKContext(ctx), req2)
	require.NoError(t, err)

	// Verify both consents are active
	require.True(t, keeper.CanProcessData(ctx, address, "data_processing"))
	require.True(t, keeper.CanProcessData(ctx, address, "marketing"))

	// Withdraw marketing consent
	req3 := &types.MsgRecordGDPRConsent{
		Address:        address,
		ConsentType:    "marketing",
		Consented:      false,
		ConsentVersion: "v1",
	}
	_, err = server.RecordGDPRConsent(sdk.WrapSDKContext(ctx), req3)
	require.NoError(t, err)

	// Verify processing is restricted (global restriction applies)
	require.True(t, keeper.IsProcessingRestricted(ctx, address))

	// Even though data_processing consent exists, global restriction blocks it
	require.False(t, keeper.CanProcessData(ctx, address, "data_processing"))
	require.False(t, keeper.CanProcessData(ctx, address, "marketing"))
}

func TestMsgRecordGDPRConsent_WithdrawalAuditTrail(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	address := createTestAddressMsg("test_audit")
	consentType := "data_processing"

	// Give consent
	req1 := &types.MsgRecordGDPRConsent{
		Address:        address,
		ConsentType:    consentType,
		Consented:      true,
		ConsentVersion: "v1",
	}
	_, err := server.RecordGDPRConsent(sdk.WrapSDKContext(ctx), req1)
	require.NoError(t, err)

	consent1, found := keeper.GetGDPRConsent(ctx, address, consentType)
	require.True(t, found)
	require.NotNil(t, consent1.ConsentGivenAt)
	require.Nil(t, consent1.ConsentWithdrawnAt)

	// Withdraw consent
	req2 := &types.MsgRecordGDPRConsent{
		Address:        address,
		ConsentType:    consentType,
		Consented:      false,
		ConsentVersion: "v1",
	}
	_, err = server.RecordGDPRConsent(sdk.WrapSDKContext(ctx), req2)
	require.NoError(t, err)

	consent2, found := keeper.GetGDPRConsent(ctx, address, consentType)
	require.True(t, found)
	require.NotNil(t, consent2.ConsentGivenAt, "original consent timestamp should be preserved")
	require.NotNil(t, consent2.ConsentWithdrawnAt, "withdrawal timestamp should be set")

	// Verify audit trail shows withdrawal
	require.False(t, consent2.Consented)
	require.NotNil(t, consent2.ConsentWithdrawnAt)
}

// TestGenerateTaxReportValidFilePath tests that valid file paths are accepted
func TestGenerateTaxReportValidFilePath(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	address := sdk.AccAddress([]byte("test_address_123456")).String()

	req := &types.MsgGenerateTaxReport{
		Address:      address,
		TaxYear:      "2023",
		Jurisdiction: "US",
		ReportType:   "1099-MISC",
		FilePath:     "reports/tax_2023.pdf",
	}

	resp, err := server.GenerateTaxReport(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.ReportId)
	require.Equal(t, "reports/tax_2023.pdf", resp.FilePath)
}

// TestGenerateTaxReportPathTraversalBlocked tests that path traversal attacks are blocked
func TestGenerateTaxReportPathTraversalBlocked(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	address := sdk.AccAddress([]byte("test_address_123456")).String()

	attackPaths := []string{
		"../../../etc/passwd",
		"/etc/shadow",
		"C:\\Windows\\System32\\config\\SAM",
		"reports/../../etc/passwd",
		"report<script>.pdf",
		"report|malicious.sh",
		".ssh/id_rsa",
		"report\x00.pdf",
		"reports//double//slash.txt",
	}

	for _, attackPath := range attackPaths {
		t.Run(attackPath, func(t *testing.T) {
			req := &types.MsgGenerateTaxReport{
				Address:      address,
				TaxYear:      "2023",
				Jurisdiction: "US",
				ReportType:   "1099-MISC",
				FilePath:     attackPath,
			}

			resp, err := server.GenerateTaxReport(sdk.WrapSDKContext(ctx), req)
			require.Error(t, err, "Should reject malicious path: %s", attackPath)
			require.Nil(t, resp)
			require.Contains(t, err.Error(), "invalid file path")
		})
	}
}

// TestGenerateTaxReportEmptyFilePathAllowed tests that empty file path is allowed (optional field)
func TestGenerateTaxReportEmptyFilePathAllowed(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	address := sdk.AccAddress([]byte("test_address_123456")).String()

	req := &types.MsgGenerateTaxReport{
		Address:      address,
		TaxYear:      "2023",
		Jurisdiction: "US",
		ReportType:   "1099-MISC",
		FilePath:     "", // Empty file path should be allowed
	}

	resp, err := server.GenerateTaxReport(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.ReportId)
	require.Empty(t, resp.FilePath)
}
