package keeper

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// createTestAddress creates a valid Bech32 address for testing
func createTestAddress(name string) string {
	hash := sha256.Sum256([]byte(name))
	addr := sdk.AccAddress(hash[:20])
	return addr.String()
}

// TestKYCSubmission_EventEmitted verifies that KYC submission emits proper audit event
func TestKYCSubmission_EventEmitted(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	// Setup approved provider
	providerAddr := createTestAddress("provider_address")
	params := keeper.GetParams(ctx)
	params.ApprovedKycProviders = []string{providerAddr}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	piiCommitment := make([]byte, 32)
	copy(piiCommitment, []byte("test_commitment_hash_32_bytes"))

	req := &types.MsgSubmitKYC{
		Address:       createTestAddress("kyc_user"),
		KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:      providerAddr,
		PiiCommitment: piiCommitment,
		Jurisdiction:  "US",
	}

	_, err = server.SubmitKYC(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err)

	// Verify event was emitted
	events := ctx.EventManager().Events()
	require.NotEmpty(t, events, "no events emitted")

	var kycEvent sdk.Event
	found := false
	for _, event := range events {
		if event.Type == types.EventTypeKYCApproved {
			kycEvent = event
			found = true
			break
		}
	}
	require.True(t, found, "kyc_approved event not found")

	// Verify event attributes
	attrs := abciAttributesToMap(kycEvent)
	require.Equal(t, req.Address, attrs[types.AttributeKeyAddress])
	require.Equal(t, req.Provider, attrs[types.AttributeKeyProvider])
	require.Equal(t, req.KycLevel.String(), attrs[types.AttributeKeyKYCLevel])
	require.Equal(t, fmt.Sprintf("%x", req.PiiCommitment), attrs[types.AttributeKeyPIICommitment])

	// Verify audit trail attributes are present
	require.NotEmpty(t, attrs[types.AttributeKeyBlockHeight])
	require.NotEmpty(t, attrs[types.AttributeKeyBlockTime])
	require.NotEmpty(t, attrs[types.AttributeKeyTimestamp])

	// Verify timestamp format is valid RFC3339
	_, err = time.Parse(time.RFC3339, attrs[types.AttributeKeyBlockTime])
	require.NoError(t, err, "block_time is not valid RFC3339 format")
}

// TestSARReporting_EventEmitted verifies that SAR filing emits proper audit event
func TestSARReporting_EventEmitted(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	req := &types.MsgReportSuspiciousActivity{
		Reporter:        createTestAddress("reporter"),
		Address:         createTestAddress("user"),
		TransactionHash: "hash123",
		ActivityType:    "structuring",
		Description:     "many small transactions",
	}

	resp, err := server.ReportSuspiciousActivity(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err)
	require.NotEmpty(t, resp.ActivityId)

	// Verify event was emitted
	events := ctx.EventManager().Events()
	require.NotEmpty(t, events)

	var sarEvent sdk.Event
	found := false
	for _, event := range events {
		if event.Type == types.EventTypeSARReported {
			sarEvent = event
			found = true
			break
		}
	}
	require.True(t, found, "suspicious_activity_reported event not found")

	// Verify event attributes
	attrs := abciAttributesToMap(sarEvent)
	require.Equal(t, resp.ActivityId, attrs[types.AttributeKeyActivityID])
	require.Equal(t, req.Address, attrs[types.AttributeKeyAddress])
	require.Equal(t, req.Reporter, attrs[types.AttributeKeyReporter])
	require.Equal(t, req.ActivityType, attrs[types.AttributeKeyActivityType])
	require.Equal(t, req.TransactionHash, attrs[types.AttributeKeyTransactionHash])

	// Verify audit trail attributes
	require.NotEmpty(t, attrs[types.AttributeKeyBlockHeight])
	require.NotEmpty(t, attrs[types.AttributeKeyBlockTime])
	require.NotEmpty(t, attrs[types.AttributeKeyTimestamp])
}

// TestSanctionsScreening_EventEmitted verifies that sanctions screening emits proper audit event
func TestSanctionsScreening_EventEmitted(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	req := &types.MsgScreenSanctions{
		Address: createTestAddress("sanction_user"),
	}

	resp, err := server.ScreenSanctions(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err)
	require.Equal(t, types.SanctionsStatus_SANCTIONS_CLEAR, resp.Status)

	// Verify event was emitted
	events := ctx.EventManager().Events()
	require.NotEmpty(t, events)

	var sanctionEvent sdk.Event
	found := false
	for _, event := range events {
		if event.Type == types.EventTypeSanctionsScreening {
			sanctionEvent = event
			found = true
			break
		}
	}
	require.True(t, found, "sanctions_screening event not found")

	// Verify event attributes
	attrs := abciAttributesToMap(sanctionEvent)
	require.Equal(t, req.Address, attrs[types.AttributeKeyAddress])
	require.Equal(t, resp.Status.String(), attrs[types.AttributeKeyStatus])
	require.Equal(t, fmt.Sprintf("%t", resp.RequiresReview), attrs[types.AttributeKeyRequiresReview])
	require.NotEmpty(t, attrs[types.AttributeKeyScreeningResult])
	require.NotEmpty(t, attrs[types.AttributeKeyMatchCount])

	// Verify audit trail attributes
	require.NotEmpty(t, attrs[types.AttributeKeyBlockHeight])
	require.NotEmpty(t, attrs[types.AttributeKeyBlockTime])
	require.NotEmpty(t, attrs[types.AttributeKeyTimestamp])
}

// TestGDPRConsent_EventEmitted verifies that GDPR consent recording emits proper audit event
func TestGDPRConsent_EventEmitted(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	req := &types.MsgRecordGDPRConsent{
		Address:        createTestAddress("gdpr_user"),
		ConsentType:    "data_processing",
		Consented:      true,
		ConsentVersion: "v1.0",
	}

	_, err := server.RecordGDPRConsent(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err)

	// Verify event was emitted
	events := ctx.EventManager().Events()
	require.NotEmpty(t, events)

	var consentEvent sdk.Event
	found := false
	for _, event := range events {
		if event.Type == types.EventTypeGDPRConsentRecorded {
			consentEvent = event
			found = true
			break
		}
	}
	require.True(t, found, "gdpr_consent_recorded event not found")

	// Verify event attributes
	attrs := abciAttributesToMap(consentEvent)
	require.Equal(t, req.Address, attrs[types.AttributeKeyAddress])
	require.Equal(t, req.ConsentType, attrs[types.AttributeKeyConsentType])
	require.Equal(t, fmt.Sprintf("%t", req.Consented), attrs[types.AttributeKeyConsented])
	require.Equal(t, req.ConsentVersion, attrs[types.AttributeKeyConsentVersion])

	// Verify audit trail attributes
	require.NotEmpty(t, attrs[types.AttributeKeyBlockHeight])
	require.NotEmpty(t, attrs[types.AttributeKeyBlockTime])
	require.NotEmpty(t, attrs[types.AttributeKeyTimestamp])
}

// TestGDPRDataRequest_EventEmitted verifies that GDPR data request emits proper audit event
func TestGDPRDataRequest_EventEmitted(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	req := &types.MsgRequestGDPRData{
		Address:     createTestAddress("gdpr_requester"),
		RequestType: "access",
	}

	resp, err := server.RequestGDPRData(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err)
	require.NotEmpty(t, resp.RequestId)

	// Verify event was emitted
	events := ctx.EventManager().Events()
	require.NotEmpty(t, events)

	var requestEvent sdk.Event
	found := false
	for _, event := range events {
		if event.Type == types.EventTypeGDPRDataRequested {
			requestEvent = event
			found = true
			break
		}
	}
	require.True(t, found, "gdpr_data_requested event not found")

	// Verify event attributes
	attrs := abciAttributesToMap(requestEvent)
	require.Equal(t, resp.RequestId, attrs[types.AttributeKeyRequestID])
	require.Equal(t, req.Address, attrs[types.AttributeKeyAddress])
	require.Equal(t, req.RequestType, attrs[types.AttributeKeyGDPRRequestType])

	// Verify audit trail attributes
	require.NotEmpty(t, attrs[types.AttributeKeyBlockHeight])
	require.NotEmpty(t, attrs[types.AttributeKeyBlockTime])
	require.NotEmpty(t, attrs[types.AttributeKeyTimestamp])
}

// TestGDPRDataErasure_EventEmitted verifies that GDPR data erasure emits proper audit event
func TestGDPRDataErasure_EventEmitted(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	req := &types.MsgEraseGDPRData{
		Address:       createTestAddress("erase_user"),
		ErasureReason: "user_request",
	}

	resp, err := server.EraseGDPRData(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err)
	require.True(t, resp.Success)
	require.NotEmpty(t, resp.ErasureEventId)

	// Verify event was emitted
	events := ctx.EventManager().Events()
	require.NotEmpty(t, events)

	var erasureEvent sdk.Event
	found := false
	for _, event := range events {
		if event.Type == types.EventTypeGDPRDataErased {
			erasureEvent = event
			found = true
			break
		}
	}
	require.True(t, found, "gdpr_data_erased event not found")

	// Verify event attributes
	attrs := abciAttributesToMap(erasureEvent)
	require.Equal(t, req.Address, attrs[types.AttributeKeyAddress])
	require.Equal(t, resp.ErasureEventId, attrs[types.AttributeKeyErasureEventID])
	require.Equal(t, req.ErasureReason, attrs[types.AttributeKeyErasureReason])
	require.NotEmpty(t, attrs[types.AttributeKeyErasureTime])

	// Verify audit trail attributes
	require.NotEmpty(t, attrs[types.AttributeKeyBlockHeight])
	require.NotEmpty(t, attrs[types.AttributeKeyBlockTime])
	require.NotEmpty(t, attrs[types.AttributeKeyTimestamp])

	// Verify erasure time format is valid RFC3339
	_, err = time.Parse(time.RFC3339, attrs[types.AttributeKeyErasureTime])
	require.NoError(t, err, "erasure_time is not valid RFC3339 format")
}

// TestTaxReportGeneration_EventEmitted verifies that tax report generation emits proper audit event
func TestTaxReportGeneration_EventEmitted(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	req := &types.MsgGenerateTaxReport{
		Address:      createTestAddress("tax_user"),
		TaxYear:      "2024",
		Jurisdiction: "US",
		ReportType:   "1099",
	}

	resp, err := server.GenerateTaxReport(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err)
	require.NotEmpty(t, resp.ReportId)

	// Verify event was emitted
	events := ctx.EventManager().Events()
	require.NotEmpty(t, events)

	var taxEvent sdk.Event
	found := false
	for _, event := range events {
		if event.Type == types.EventTypeTaxReportGenerated {
			taxEvent = event
			found = true
			break
		}
	}
	require.True(t, found, "tax_report_generated event not found")

	// Verify event attributes
	attrs := abciAttributesToMap(taxEvent)
	require.Equal(t, resp.ReportId, attrs[types.AttributeKeyReportID])
	require.Equal(t, req.Address, attrs[types.AttributeKeyAddress])
	require.Equal(t, req.TaxYear, attrs[types.AttributeKeyTaxYear])
	require.Equal(t, req.Jurisdiction, attrs[types.AttributeKeyJurisdiction])
	require.Equal(t, req.ReportType, attrs[types.AttributeKeyReportType])

	// Verify audit trail attributes
	require.NotEmpty(t, attrs[types.AttributeKeyBlockHeight])
	require.NotEmpty(t, attrs[types.AttributeKeyBlockTime])
	require.NotEmpty(t, attrs[types.AttributeKeyTimestamp])
}

// TestGDPRConsentWithdrawal_EventEmitted verifies withdrawal emits proper audit event
func TestGDPRConsentWithdrawal_EventEmitted(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	req := &types.MsgRecordGDPRConsent{
		Address:        createTestAddress("gdpr_withdraw"),
		ConsentType:    "data_processing",
		Consented:      false, // Withdrawal
		ConsentVersion: "v1.0",
	}

	_, err := server.RecordGDPRConsent(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err)

	// Verify event was emitted
	events := ctx.EventManager().Events()
	require.NotEmpty(t, events)

	var consentEvent sdk.Event
	found := false
	for _, event := range events {
		if event.Type == types.EventTypeGDPRConsentWithdrawn {
			consentEvent = event
			found = true
			break
		}
	}
	require.True(t, found, "gdpr_consent_withdrawn event not found")

	// Verify withdrawal is properly recorded
	attrs := abciAttributesToMap(consentEvent)
	require.Equal(t, req.ConsentType, attrs[types.AttributeKeyConsentType])
	require.Equal(t, "true", attrs[types.AttributeKeyProcessingRestricted])
}

// TestSanctionsScreening_WithMatches_EventEmitted verifies flagged screening emits proper event
func TestSanctionsScreening_WithMatches_EventEmitted(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	// Register a mock provider that returns matches
	provider := &testSanctionsProviderWithMatches{}
	keeper.RegisterSanctionsProvider("mock", provider)

	req := &types.MsgScreenSanctions{
		Address: createTestAddress("flagged_user"),
	}

	resp, err := server.ScreenSanctions(sdk.WrapSDKContext(ctx), req)
	require.NoError(t, err)
	require.True(t, resp.RequiresReview, "should require review when matches found")

	// Verify event was emitted
	events := ctx.EventManager().Events()
	require.NotEmpty(t, events)

	var sanctionEvent sdk.Event
	found := false
	for _, event := range events {
		if event.Type == types.EventTypeSanctionsScreening {
			sanctionEvent = event
			found = true
			break
		}
	}
	require.True(t, found, "sanctions_screening event not found")

	// Verify match count is greater than zero
	attrs := abciAttributesToMap(sanctionEvent)
	require.NotEqual(t, "0", attrs[types.AttributeKeyMatchCount])
}

// TestMultipleEvents_InSingleTransaction verifies multiple events in one tx
func TestMultipleEvents_InSingleTransaction(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	// Setup approved provider
	providerAddr := createTestAddress("multi_provider")
	params := keeper.GetParams(ctx)
	params.ApprovedKycProviders = []string{providerAddr}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	multiAddr := createTestAddress("multi_user")

	// Submit KYC
	piiCommitment := make([]byte, 32)
	copy(piiCommitment, []byte("test_commitment_hash_32_bytes"))

	kycReq := &types.MsgSubmitKYC{
		Address:       multiAddr,
		KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:      providerAddr,
		PiiCommitment: piiCommitment,
		Jurisdiction:  "US",
	}
	_, err = server.SubmitKYC(sdk.WrapSDKContext(ctx), kycReq)
	require.NoError(t, err)

	// Screen sanctions
	sanctionsReq := &types.MsgScreenSanctions{
		Address: multiAddr,
	}
	_, err = server.ScreenSanctions(sdk.WrapSDKContext(ctx), sanctionsReq)
	require.NoError(t, err)

	// Verify both events were emitted
	events := ctx.EventManager().Events()
	require.NotEmpty(t, events)

	kycFound := false
	sanctionsFound := false
	for _, event := range events {
		if event.Type == types.EventTypeKYCSubmitted {
			kycFound = true
		}
		if event.Type == types.EventTypeSanctionsScreening {
			sanctionsFound = true
		}
	}
	require.True(t, kycFound, "kyc_submitted event not found")
	require.True(t, sanctionsFound, "sanctions_screening event not found")
}

// Helper function to convert event attributes to a map
func attributesToMap(attrs []sdk.Attribute) map[string]string {
	result := make(map[string]string)
	for _, attr := range attrs {
		result[attr.Key] = attr.Value
	}
	return result
}

// Helper function to convert ABCI event attributes to a map
func abciAttributesToMap(event sdk.Event) map[string]string {
	result := make(map[string]string)
	for _, attr := range event.Attributes {
		result[attr.Key] = attr.Value
	}
	return result
}

// testSanctionsProviderWithMatches returns sanctions matches for testing
type testSanctionsProviderWithMatches struct{}

func (m *testSanctionsProviderWithMatches) ScreenAddress(address string) (*types.SanctionsScreeningResult, error) {
	return &types.SanctionsScreeningResult{
		Address: address,
		Status:  types.SanctionsStatus_SANCTIONS_MATCH,
		Matches: []*types.SanctionsMatch{
			{
				ListName:   "OFAC-SDN",
				MatchScore: "95.0",
			},
		},
		ScreenedAt:           nil, // Will be set by msg_server
		RequiresManualReview: true,
	}, nil
}

func (m *testSanctionsProviderWithMatches) CheckLists(_ []string) ([]*types.SanctionsMatch, error) {
	return []*types.SanctionsMatch{
		{
			ListName:   "OFAC-SDN",
			MatchScore: "95.0",
		},
	}, nil
}
