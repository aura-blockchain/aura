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
