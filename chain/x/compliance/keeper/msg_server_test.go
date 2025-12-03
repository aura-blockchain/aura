package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

func TestMsgSubmitKYCStoresRecord(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	// Setup approved provider
	providerAddr := sdk.AccAddress([]byte("provider_address_12")).String()
	params := keeper.GetParams(ctx)
	params.ApprovedKycProviders = []string{providerAddr}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	piiCommitment := make([]byte, 32) // SHA-256 hash
	copy(piiCommitment, []byte("test_commitment_hash_32_bytes"))

	req := &types.MsgSubmitKYC{
		Address:       "aura1kyc",
		KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:      providerAddr,
		PiiCommitment: piiCommitment,
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
	req := &types.MsgReportSuspiciousActivity{
		Reporter:        "aura1reporter",
		Address:         "aura1user",
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
	req := &types.MsgScreenSanctions{Address: "aura1sanction"}
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
		Address:        "aura1gdpr",
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
		Address:     "aura1gdprreq",
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
		Address:      "aura1tax",
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
	resp, err := server.ScreenSanctions(sdk.WrapSDKContext(ctx), &types.MsgScreenSanctions{Address: "aura1provider"})
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
		ScreenedAt:           timestamppb.New(time.Now()),
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

	address := "aura1test"
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

	address := "aura1test"
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

			// Verify event attributes
			var processingRestrictedAttr, deletionTriggeredAttr sdk.Attribute
			for _, attr := range event.Attributes {
				if attr.Key == types.AttributeKeyProcessingRestricted {
					processingRestrictedAttr = attr
				}
				if attr.Key == types.AttributeKeyDeletionTriggered {
					deletionTriggeredAttr = attr
				}
			}

			require.NotNil(t, processingRestrictedAttr.Key, "processing_restricted attribute should be present")
			require.Equal(t, "true", processingRestrictedAttr.Value, "processing_restricted should be true")
			require.NotNil(t, deletionTriggeredAttr.Key, "deletion_triggered attribute should be present")
			require.Equal(t, "true", deletionTriggeredAttr.Value, "deletion_triggered should be true")
		}

		if event.Type == "gdpr_data_deletion_requested" {
			deletionEventFound = true

			// Verify deletion event attributes
			var addressAttr, consentTypeAttr sdk.Attribute
			for _, attr := range event.Attributes {
				if attr.Key == types.AttributeKeyAddress {
					addressAttr = attr
				}
				if attr.Key == types.AttributeKeyConsentType {
					consentTypeAttr = attr
				}
			}

			require.Equal(t, address, addressAttr.Value, "address should match")
			require.Equal(t, consentType, consentTypeAttr.Value, "consent type should match")
		}
	}

	require.True(t, withdrawalEventFound, "GDPR consent withdrawn event should be emitted")
	require.True(t, deletionEventFound, "Data deletion requested event should be emitted")
}

func TestMsgRecordGDPRConsent_GivingConsentRemovesRestriction(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	address := "aura1test"
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

	address := "aura1test"

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

	address := "aura1test"
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
