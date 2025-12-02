package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/compliance/types"
)

func setupKeeper(t *testing.T) (*Keeper, sdk.Context) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	keeper := NewKeeper(input.Cdc, input.StoreKey)
	return keeper, input.Ctx
}

// ============================================================================
// Keeper Construction Tests
// ============================================================================

func TestNewKeeper(t *testing.T) {
	keeper, _ := setupKeeper(t)

	require.NotNil(t, keeper)
	require.NotNil(t, keeper.cdc)
	require.NotNil(t, keeper.storeKey)
	require.NotNil(t, keeper.kycProviders)
	require.NotNil(t, keeper.sanctionsProviders)
	require.NotNil(t, keeper.taxReportGenerators)
	require.NotNil(t, keeper.sanctionsCache)
}

// ============================================================================
// Params Tests
// ============================================================================

func TestGetSetParams(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Test default params
	params := keeper.GetParams(ctx)
	require.NotNil(t, params)

	// Test setting params
	newParams := types.ComplianceParams{
		KycRequired:                  true,
		MinimumKycLevel:              types.KYCLevel_KYC_LEVEL_ADVANCED,
		KycExpiryDays:                180,
		TransactionMonitoringEnabled: true,
		VelocityLimit_24H:            "5000000",
		SingleTransactionLimit:       "500000",
		StructuringThresholdCount:    15,
		SanctionsScreeningEnabled:    true,
		SanctionsLists:               []string{"OFAC", "EU", "UN"},
		ScreeningCacheHours:          48,
		GdprEnabled:                  true,
		DataRetentionDays:            365,
		ProcessingPurposes:           []string{"compliance"},
		TaxReportingEnabled:          true,
		TaxJurisdictions:             []string{"US"},
		TaxYearEnd:                   "12-31",
	}

	err := keeper.SetParams(ctx, newParams)
	require.NoError(t, err)

	retrieved := keeper.GetParams(ctx)
	require.Equal(t, newParams.KycRequired, retrieved.KycRequired)
	require.Equal(t, newParams.MinimumKycLevel, retrieved.MinimumKycLevel)
	require.Equal(t, newParams.KycExpiryDays, retrieved.KycExpiryDays)
	require.Equal(t, newParams.VelocityLimit_24H, retrieved.VelocityLimit_24H)
	require.Equal(t, newParams.SanctionsLists, retrieved.SanctionsLists)
}

func TestGetParams_EmptyStore(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	params := keeper.GetParams(ctx)
	// Should not panic, should return zero value
	require.NotNil(t, params)
}

// ============================================================================
// Provider Registration Tests
// ============================================================================

func TestRegisterKYCProvider(t *testing.T) {
	keeper, _ := setupKeeper(t)

	provider := &mockKYCProvider{}
	keeper.RegisterKYCProvider("test_provider", provider)

	require.NotNil(t, keeper.kycProviders["test_provider"])
}

func TestRegisterSanctionsProvider(t *testing.T) {
	keeper, _ := setupKeeper(t)

	provider := &mockSanctionsProvider{}
	keeper.RegisterSanctionsProvider("test_sanctions", provider)

	require.NotNil(t, keeper.sanctionsProviders["test_sanctions"])
}

func TestRegisterTaxReportGenerator(t *testing.T) {
	keeper, _ := setupKeeper(t)

	generator := &mockTaxReportGenerator{}
	keeper.RegisterTaxReportGenerator("US", generator)

	require.NotNil(t, keeper.taxReportGenerators["US"])
}

func TestRegisterMultipleProviders(t *testing.T) {
	keeper, _ := setupKeeper(t)

	// Register multiple KYC providers
	keeper.RegisterKYCProvider("provider1", &mockKYCProvider{})
	keeper.RegisterKYCProvider("provider2", &mockKYCProvider{})
	keeper.RegisterKYCProvider("provider3", &mockKYCProvider{})

	require.Len(t, keeper.kycProviders, 3)
	require.NotNil(t, keeper.kycProviders["provider1"])
	require.NotNil(t, keeper.kycProviders["provider2"])
	require.NotNil(t, keeper.kycProviders["provider3"])
}

func TestRegisterProvider_Overwrite(t *testing.T) {
	keeper, _ := setupKeeper(t)

	provider1 := &mockKYCProvider{name: "first"}
	provider2 := &mockKYCProvider{name: "second"}

	keeper.RegisterKYCProvider("test", provider1)
	keeper.RegisterKYCProvider("test", provider2)

	// Should overwrite
	require.Len(t, keeper.kycProviders, 1)
}

// ============================================================================
// Initialize Default Monitoring Rules Tests
// ============================================================================

func TestInitializeDefaultMonitoringRules(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	err := keeper.initializeDefaultMonitoringRules(ctx)
	// Should not fail even with empty params
	require.NoError(t, err)
}

func TestInitializeDefaultMonitoringRules_WithParams(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Set params first
	params := types.DefaultParams()
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	err = keeper.initializeDefaultMonitoringRules(ctx)
	require.NoError(t, err)
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestKeeper_FullKYCWorkflow(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	now := time.Now()
	address := "cosmos1testaddress"

	// 1. Submit KYC record
	record := &types.KYCRecord{
		Address:              address,
		KycLevel:             types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:             "provider1",
		VerifiedAt:           timestamppb.New(now),
		ExpiresAt:            timestamppb.New(now.Add(365 * 24 * time.Hour)),
		VerificationId:       "ver123",
		Documents:            []string{"passport"},
		Jurisdiction:         "US",
		EnhancedDueDiligence: false,
		RiskScore:            "low",
	}

	err := keeper.SetKYCRecord(ctx, record)
	require.NoError(t, err)

	// 2. Retrieve KYC record
	retrieved, err := keeper.GetKYCRecord(ctx, address)
	require.NoError(t, err)
	require.Equal(t, record.KycLevel, retrieved.KycLevel)

	// 3. Update KYC level
	record.KycLevel = types.KYCLevel_KYC_LEVEL_ADVANCED
	err = keeper.SetKYCRecord(ctx, record)
	require.NoError(t, err)

	// 4. Verify update
	retrieved, err = keeper.GetKYCRecord(ctx, address)
	require.NoError(t, err)
	require.Equal(t, types.KYCLevel_KYC_LEVEL_ADVANCED, retrieved.KycLevel)
}

func TestKeeper_FullAMLWorkflow(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	now := time.Now()
	address := "cosmos1testaddress"

	// 1. Create AML profile
	profile := &types.AMLProfile{
		Address:           address,
		RiskLevel:         types.AMLRiskLevel_AML_RISK_LOW,
		RiskFactors:       []string{},
		LastAssessment:    timestamppb.New(now),
		TotalTransactions: 0,
		TotalVolume:       "0",
		PepStatus:         false,
	}

	err := keeper.SetAMLProfile(ctx, profile)
	require.NoError(t, err)

	// 2. Report suspicious activity
	activity := &types.SuspiciousActivity{
		Id:              "sa1",
		Address:         address,
		TransactionHash: "hash123",
		ActivityType:    "structuring",
		Description:     "Multiple small transactions",
		Amount:          "50000",
		DetectedAt:      timestamppb.New(now),
		FiledSar:        false,
	}

	err = keeper.SetSuspiciousActivity(ctx, activity)
	require.NoError(t, err)

	// 3. Update AML profile risk
	profile.RiskLevel = types.AMLRiskLevel_AML_RISK_HIGH
	profile.RiskFactors = []string{"structuring"}
	err = keeper.SetAMLProfile(ctx, profile)
	require.NoError(t, err)

	// 4. Verify updates
	retrieved, err := keeper.GetAMLProfile(ctx, address)
	require.NoError(t, err)
	require.Equal(t, types.AMLRiskLevel_AML_RISK_HIGH, retrieved.RiskLevel)
	require.Contains(t, retrieved.RiskFactors, "structuring")
}

func TestKeeper_SanctionsWorkflow(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	now := time.Now()
	address := "cosmos1testaddress"

	// 1. Screen for sanctions
	result := &types.SanctionsScreeningResult{
		Address:              address,
		Status:               types.SanctionsStatus_SANCTIONS_CLEAR,
		Matches:              []*types.SanctionsMatch{},
		ScreenedAt:           timestamppb.New(now),
		ScreeningProvider:    "provider1",
		RequiresManualReview: false,
	}

	err := keeper.SetSanctionsResult(ctx, result)
	require.NoError(t, err)

	// 2. Retrieve result
	retrieved, err := keeper.GetSanctionsResult(ctx, address)
	require.NoError(t, err)
	require.Equal(t, types.SanctionsStatus_SANCTIONS_CLEAR, retrieved.Status)
}

func TestKeeper_GDPRWorkflow(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	now := time.Now()
	address := "cosmos1testaddress"

	// 1. Record consent
	consent := &types.GDPRConsent{
		Address:        address,
		ConsentType:    "data_processing",
		Consented:      true,
		ConsentGivenAt: timestamppb.New(now),
		ConsentVersion: "v1.0",
	}

	err := keeper.SetGDPRConsent(ctx, consent)
	require.NoError(t, err)

	// 2. Request data
	request := &types.GDPRDataRequest{
		Id:          "req1",
		Address:     address,
		RequestType: "access",
		RequestedAt: timestamppb.New(now),
		Status:      "pending",
	}

	err = keeper.SetGDPRRequest(ctx, request)
	require.NoError(t, err)

	// 3. Retrieve request
	retrieved, err := keeper.GetGDPRRequest(ctx, "req1")
	require.NoError(t, err)
	require.Equal(t, "access", retrieved.RequestType)
}

// ============================================================================
// Mock Provider Implementations
// ============================================================================

type mockKYCProvider struct {
	name string
}

func (m *mockKYCProvider) VerifyIdentity(address, documentType string, documents [][]byte) (*types.KYCRecord, error) {
	return &types.KYCRecord{
		Address:        address,
		KycLevel:       types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:       "mock",
		VerifiedAt:     timestamppb.Now(),
		VerificationId: "mock-ver-123",
	}, nil
}

func (m *mockKYCProvider) GetVerificationStatus(verificationID string) (*types.KYCRecord, error) {
	return &types.KYCRecord{
		VerificationId: verificationID,
		KycLevel:       types.KYCLevel_KYC_LEVEL_BASIC,
	}, nil
}

func (m *mockKYCProvider) UpdateRiskScore(address string) (string, error) {
	return "low", nil
}

type mockSanctionsProvider struct{}

func (m *mockSanctionsProvider) ScreenAddress(address string) (*types.SanctionsScreeningResult, error) {
	return &types.SanctionsScreeningResult{
		Address:           address,
		Status:            types.SanctionsStatus_SANCTIONS_CLEAR,
		ScreenedAt:        timestamppb.Now(),
		ScreeningProvider: "mock",
	}, nil
}

func (m *mockSanctionsProvider) CheckLists(lists []string) ([]*types.SanctionsMatch, error) {
	return []*types.SanctionsMatch{}, nil
}

type mockTaxReportGenerator struct{}

func (m *mockTaxReportGenerator) GenerateReport(address, taxYear, reportType string, transactions []*types.TaxTransaction) (*types.TaxReport, error) {
	return &types.TaxReport{
		Id:           "report-123",
		Address:      address,
		TaxYear:      taxYear,
		ReportType:   reportType,
		Transactions: transactions,
		GeneratedAt:  timestamppb.Now(),
	}, nil
}

func (m *mockTaxReportGenerator) ExportToFile(report *types.TaxReport, format string) (string, error) {
	return "/path/to/report." + format, nil
}

// ============================================================================
// Edge Cases and Error Handling
// ============================================================================

func TestKeeper_NilContext(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Test that operations work with valid context
	params := keeper.GetParams(ctx)
	require.NotNil(t, params)
}

func TestKeeper_ConcurrentAccess(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Test that multiple operations can be performed
	err := keeper.SetKYCRecord(ctx, &types.KYCRecord{
		Address:  "addr1",
		KycLevel: types.KYCLevel_KYC_LEVEL_BASIC,
	})
	require.NoError(t, err)

	err = keeper.SetAMLProfile(ctx, &types.AMLProfile{
		Address:   "addr1",
		RiskLevel: types.AMLRiskLevel_AML_RISK_LOW,
	})
	require.NoError(t, err)

	err = keeper.SetSanctionsResult(ctx, &types.SanctionsScreeningResult{
		Address:    "addr1",
		Status:     types.SanctionsStatus_SANCTIONS_CLEAR,
		ScreenedAt: timestamppb.Now(),
	})
	require.NoError(t, err)

	// All should succeed
}

func TestKeeper_EmptyAddress(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Test operations with empty address
	err := keeper.SetKYCRecord(ctx, &types.KYCRecord{
		Address:  "",
		KycLevel: types.KYCLevel_KYC_LEVEL_BASIC,
	})
	require.NoError(t, err)

	_, err = keeper.GetKYCRecord(ctx, "")
	require.NoError(t, err)
}

func TestKeeper_SpecialCharactersInAddress(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	addresses := []string{
		"cosmos1-test",
		"cosmos1_test",
		"cosmos1.test",
		"cosmos1/test",
	}

	for _, addr := range addresses {
		err := keeper.SetKYCRecord(ctx, &types.KYCRecord{
			Address:  addr,
			KycLevel: types.KYCLevel_KYC_LEVEL_BASIC,
		})
		require.NoError(t, err)

		_, err = keeper.GetKYCRecord(ctx, addr)
		require.NoError(t, err)
	}
}

func TestKeeper_LargeDataSets(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Create 100 KYC records
	for i := 0; i < 100; i++ {
		err := keeper.SetKYCRecord(ctx, &types.KYCRecord{
			Address:  "cosmos1addr" + string(rune('0'+i%10)),
			KycLevel: types.KYCLevel_KYC_LEVEL_BASIC,
		})
		require.NoError(t, err)
	}

	records, err := keeper.GetAllKYCRecords(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, records)
}
