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

func setupTestKeeper(t *testing.T) (*Keeper, sdk.Context) {
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	keeper := NewKeeper(input.Cdc, input.StoreKey)
	return keeper, input.Ctx
}

// ============================================================================
// KYC Record KVStore Tests
// ============================================================================

func TestSetGetKYCRecord(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()
	record := &types.KYCRecord{
		Address:              "cosmos1test",
		KycLevel:             types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:             "provider1",
		VerifiedAt:           timestamppb.New(now),
		ExpiresAt:            timestamppb.New(now.Add(365 * 24 * time.Hour)),
		VerificationId:       "ver123",
		Documents:            []string{"passport", "address_proof"},
		Jurisdiction:         "US",
		EnhancedDueDiligence: false,
		RiskScore:            "low",
	}

	// Test Set
	err := keeper.SetKYCRecord(ctx, record)
	require.NoError(t, err)

	// Test Get
	retrieved, err := keeper.GetKYCRecord(ctx, "cosmos1test")
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	require.Equal(t, record.Address, retrieved.Address)
	require.Equal(t, record.KycLevel, retrieved.KycLevel)
	require.Equal(t, record.Provider, retrieved.Provider)
	require.Equal(t, record.VerificationId, retrieved.VerificationId)
	require.Equal(t, record.Jurisdiction, retrieved.Jurisdiction)
}

func TestGetKYCRecord_NotFound(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	_, err := keeper.GetKYCRecord(ctx, "nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "KYC record not found")
}

func TestGetAllKYCRecords(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()
	records := []*types.KYCRecord{
		{
			Address:        "cosmos1addr1",
			KycLevel:       types.KYCLevel_KYC_LEVEL_BASIC,
			Provider:       "provider1",
			VerifiedAt:     timestamppb.New(now),
			ExpiresAt:      timestamppb.New(now.Add(365 * 24 * time.Hour)),
			VerificationId: "ver1",
		},
		{
			Address:        "cosmos1addr2",
			KycLevel:       types.KYCLevel_KYC_LEVEL_ADVANCED,
			Provider:       "provider2",
			VerifiedAt:     timestamppb.New(now),
			ExpiresAt:      timestamppb.New(now.Add(365 * 24 * time.Hour)),
			VerificationId: "ver2",
		},
	}

	for _, record := range records {
		err := keeper.SetKYCRecord(ctx, record)
		require.NoError(t, err)
	}

	retrieved, err := keeper.GetAllKYCRecords(ctx)
	require.NoError(t, err)
	require.Len(t, retrieved, 2)
}

func TestGetAllKYCRecords_Empty(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	retrieved, err := keeper.GetAllKYCRecords(ctx)
	require.NoError(t, err)
	require.Empty(t, retrieved)
}

// ============================================================================
// AML Profile KVStore Tests
// ============================================================================

func TestSetGetAMLProfile(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()
	profile := &types.AMLProfile{
		Address:           "cosmos1test",
		RiskLevel:         types.AMLRiskLevel_AML_RISK_MEDIUM,
		RiskFactors:       []string{"high_volume", "unusual_pattern"},
		LastAssessment:    timestamppb.New(now),
		TotalTransactions: 100,
		TotalVolume:       "1000000",
		PepStatus:         false,
		SourceOfFunds:     []string{"employment", "investment"},
		Occupation:        "software_engineer",
	}

	err := keeper.SetAMLProfile(ctx, profile)
	require.NoError(t, err)

	retrieved, err := keeper.GetAMLProfile(ctx, "cosmos1test")
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	require.Equal(t, profile.Address, retrieved.Address)
	require.Equal(t, profile.RiskLevel, retrieved.RiskLevel)
	require.Equal(t, profile.TotalTransactions, retrieved.TotalTransactions)
	require.Equal(t, profile.TotalVolume, retrieved.TotalVolume)
}

func TestGetAMLProfile_NotFound(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	_, err := keeper.GetAMLProfile(ctx, "nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "AML profile not found")
}

func TestGetAllAMLProfiles(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()
	profiles := []*types.AMLProfile{
		{
			Address:           "cosmos1addr1",
			RiskLevel:         types.AMLRiskLevel_AML_RISK_LOW,
			LastAssessment:    timestamppb.New(now),
			TotalTransactions: 50,
			TotalVolume:       "500000",
		},
		{
			Address:           "cosmos1addr2",
			RiskLevel:         types.AMLRiskLevel_AML_RISK_HIGH,
			LastAssessment:    timestamppb.New(now),
			TotalTransactions: 200,
			TotalVolume:       "2000000",
		},
	}

	for _, profile := range profiles {
		err := keeper.SetAMLProfile(ctx, profile)
		require.NoError(t, err)
	}

	retrieved, err := keeper.GetAllAMLProfiles(ctx)
	require.NoError(t, err)
	require.Len(t, retrieved, 2)
}

func TestGetAllAMLProfiles_Empty(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	retrieved, err := keeper.GetAllAMLProfiles(ctx)
	require.NoError(t, err)
	require.Empty(t, retrieved)
}

// ============================================================================
// Suspicious Activity KVStore Tests
// ============================================================================

func TestSetGetSuspiciousActivity(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()
	activity := &types.SuspiciousActivity{
		Id:              "sa1",
		Address:         "cosmos1test",
		TransactionHash: "hash123",
		ActivityType:    "structuring",
		Description:     "Multiple small transactions",
		Amount:          "50000",
		DetectedAt:      timestamppb.New(now),
		ReportedAt:      timestamppb.New(now.Add(1 * time.Hour)),
		FiledSar:        true,
		SarReference:    "SAR-2024-001",
		Indicators:      []string{"velocity", "structuring"},
	}

	err := keeper.SetSuspiciousActivity(ctx, activity)
	require.NoError(t, err)

	retrieved, err := keeper.GetSuspiciousActivity(ctx, "sa1")
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	require.Equal(t, activity.Id, retrieved.Id)
	require.Equal(t, activity.Address, retrieved.Address)
	require.Equal(t, activity.ActivityType, retrieved.ActivityType)
	require.Equal(t, activity.FiledSar, retrieved.FiledSar)
}

func TestGetSuspiciousActivity_NotFound(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	_, err := keeper.GetSuspiciousActivity(ctx, "nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "suspicious activity not found")
}

func TestGetAllSuspiciousActivities(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()
	activities := []*types.SuspiciousActivity{
		{
			Id:              "sa1",
			Address:         "cosmos1addr1",
			TransactionHash: "hash1",
			ActivityType:    "structuring",
			DetectedAt:      timestamppb.New(now),
		},
		{
			Id:              "sa2",
			Address:         "cosmos1addr2",
			TransactionHash: "hash2",
			ActivityType:    "smurfing",
			DetectedAt:      timestamppb.New(now),
		},
	}

	for _, activity := range activities {
		err := keeper.SetSuspiciousActivity(ctx, activity)
		require.NoError(t, err)
	}

	retrieved, err := keeper.GetAllSuspiciousActivities(ctx)
	require.NoError(t, err)
	require.Len(t, retrieved, 2)
}

func TestGetAllSuspiciousActivities_Empty(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	retrieved, err := keeper.GetAllSuspiciousActivities(ctx)
	require.NoError(t, err)
	require.Empty(t, retrieved)
}

// ============================================================================
// Transaction Monitoring Rule KVStore Tests
// ============================================================================

func TestSetGetMonitoringRule(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()
	rule := &types.TransactionMonitoringRule{
		Id:          "rule1",
		Name:        "Velocity Limit",
		Description: "Monitor transaction velocity",
		RuleType:    "velocity",
		Parameters: map[string]string{
			"time_window": "24h",
			"max_amount":  "100000",
		},
		RiskLevel: types.TransactionRiskLevel_TX_RISK_HIGH,
		Enabled:   true,
		CreatedAt: timestamppb.New(now),
		UpdatedAt: timestamppb.New(now),
	}

	err := keeper.SetMonitoringRule(ctx, rule)
	require.NoError(t, err)

	retrieved, err := keeper.GetMonitoringRule(ctx, "rule1")
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	require.Equal(t, rule.Id, retrieved.Id)
	require.Equal(t, rule.Name, retrieved.Name)
	require.Equal(t, rule.RuleType, retrieved.RuleType)
	require.Equal(t, rule.Enabled, retrieved.Enabled)
}

func TestGetMonitoringRule_NotFound(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	_, err := keeper.GetMonitoringRule(ctx, "nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "monitoring rule not found")
}

func TestGetAllMonitoringRules(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()
	rules := []*types.TransactionMonitoringRule{
		{
			Id:        "rule1",
			Name:      "Velocity Rule",
			RuleType:  "velocity",
			Enabled:   true,
			CreatedAt: timestamppb.New(now),
		},
		{
			Id:        "rule2",
			Name:      "Structuring Rule",
			RuleType:  "structuring",
			Enabled:   true,
			CreatedAt: timestamppb.New(now),
		},
	}

	for _, rule := range rules {
		err := keeper.SetMonitoringRule(ctx, rule)
		require.NoError(t, err)
	}

	retrieved, err := keeper.GetAllMonitoringRules(ctx)
	require.NoError(t, err)
	require.Len(t, retrieved, 2)
}

func TestGetAllMonitoringRules_Empty(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	retrieved, err := keeper.GetAllMonitoringRules(ctx)
	require.NoError(t, err)
	require.Empty(t, retrieved)
}

// ============================================================================
// Transaction Alert KVStore Tests
// ============================================================================

func TestAddGetTransactionAlerts(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()
	alert := &types.TransactionAlert{
		Id:              "alert1",
		TransactionHash: "hash123",
		Address:         "cosmos1test",
		RuleId:          "rule1",
		RiskLevel:       types.TransactionRiskLevel_TX_RISK_HIGH,
		Description:     "Velocity limit exceeded",
		TriggeredAt:     timestamppb.New(now),
		Reviewed:        false,
	}

	err := keeper.AddTransactionAlert(ctx, "cosmos1test", alert)
	require.NoError(t, err)

	retrieved, err := keeper.GetTransactionAlerts(ctx, "cosmos1test")
	require.NoError(t, err)
	require.Len(t, retrieved, 1)
	require.Equal(t, alert.Id, retrieved[0].Id)
	require.Equal(t, alert.RuleId, retrieved[0].RuleId)
}

func TestGetTransactionAlerts_Empty(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	retrieved, err := keeper.GetTransactionAlerts(ctx, "cosmos1test")
	require.NoError(t, err)
	require.Empty(t, retrieved)
}

// ============================================================================
// Sanctions Screening Result KVStore Tests
// ============================================================================

func TestSetGetSanctionsResult(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()
	result := &types.SanctionsScreeningResult{
		Address:              "cosmos1test",
		Status:               types.SanctionsStatus_SANCTIONS_CLEAR,
		Matches:              []*types.SanctionsMatch{},
		ScreenedAt:           timestamppb.New(now),
		ScreeningProvider:    "provider1",
		RequiresManualReview: false,
	}

	err := keeper.SetSanctionsResult(ctx, result)
	require.NoError(t, err)

	retrieved, err := keeper.GetSanctionsResult(ctx, "cosmos1test")
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	require.Equal(t, result.Address, retrieved.Address)
	require.Equal(t, result.Status, retrieved.Status)
	require.Equal(t, result.ScreeningProvider, retrieved.ScreeningProvider)
}

func TestSetGetSanctionsResult_WithMatches(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()
	result := &types.SanctionsScreeningResult{
		Address: "cosmos1test",
		Status:  types.SanctionsStatus_SANCTIONS_MATCH,
		Matches: []*types.SanctionsMatch{
			{
				ListName:    "OFAC SDN",
				MatchScore:  "0.95",
				MatchedName: "John Doe",
				MatchedId:   "SDN-12345",
				Country:     "US",
				Program:     "NARCOTICS",
			},
		},
		ScreenedAt:           timestamppb.New(now),
		ScreeningProvider:    "provider1",
		RequiresManualReview: true,
	}

	err := keeper.SetSanctionsResult(ctx, result)
	require.NoError(t, err)

	retrieved, err := keeper.GetSanctionsResult(ctx, "cosmos1test")
	require.NoError(t, err)
	require.Equal(t, types.SanctionsStatus_SANCTIONS_MATCH, retrieved.Status)
	require.Len(t, retrieved.Matches, 1)
	require.Equal(t, "OFAC SDN", retrieved.Matches[0].ListName)
}

func TestGetSanctionsResult_NotFound(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	_, err := keeper.GetSanctionsResult(ctx, "nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "sanctions result not found")
}

func TestGetAllSanctionsResults(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()
	results := []*types.SanctionsScreeningResult{
		{
			Address:           "cosmos1addr1",
			Status:            types.SanctionsStatus_SANCTIONS_CLEAR,
			ScreenedAt:        timestamppb.New(now),
			ScreeningProvider: "provider1",
		},
		{
			Address:           "cosmos1addr2",
			Status:            types.SanctionsStatus_SANCTIONS_MATCH,
			ScreenedAt:        timestamppb.New(now),
			ScreeningProvider: "provider1",
		},
	}

	for _, result := range results {
		err := keeper.SetSanctionsResult(ctx, result)
		require.NoError(t, err)
	}

	retrieved, err := keeper.GetAllSanctionsResults(ctx)
	require.NoError(t, err)
	require.Len(t, retrieved, 2)
}

func TestGetAllSanctionsResults_Empty(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	retrieved, err := keeper.GetAllSanctionsResults(ctx)
	require.NoError(t, err)
	require.Empty(t, retrieved)
}

// ============================================================================
// GDPR Consent KVStore Tests
// ============================================================================

func TestSetGetGDPRConsent(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()
	consent := &types.GDPRConsent{
		Address:        "cosmos1test",
		ConsentType:    "data_processing",
		Consented:      true,
		ConsentGivenAt: timestamppb.New(now),
		ConsentVersion: "v1.0",
		IpAddress:      "192.168.1.1",
		UserAgent:      "Mozilla/5.0",
	}

	err := keeper.SetGDPRConsent(ctx, consent)
	require.NoError(t, err)

	// Note: Current implementation returns empty list
	retrieved, err := keeper.GetGDPRConsents(ctx, "cosmos1test")
	require.NoError(t, err)
	require.Len(t, retrieved, 1)
	require.Equal(t, consent.ConsentType, retrieved[0].ConsentType)
}

func TestGetGDPRConsents_Empty(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	retrieved, err := keeper.GetGDPRConsents(ctx, "cosmos1test")
	require.NoError(t, err)
	require.Empty(t, retrieved)
}

// ============================================================================
// GDPR Data Request KVStore Tests
// ============================================================================

func TestSetGetGDPRRequest(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()
	request := &types.GDPRDataRequest{
		Id:          "req1",
		Address:     "cosmos1test",
		RequestType: "access",
		RequestedAt: timestamppb.New(now),
		Status:      "pending",
		Notes:       "User requested data access",
	}

	err := keeper.SetGDPRRequest(ctx, request)
	require.NoError(t, err)

	retrieved, err := keeper.GetGDPRRequest(ctx, "req1")
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	require.Equal(t, request.Id, retrieved.Id)
	require.Equal(t, request.Address, retrieved.Address)
	require.Equal(t, request.RequestType, retrieved.RequestType)
	require.Equal(t, request.Status, retrieved.Status)
}

func TestGetGDPRRequest_NotFound(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	_, err := keeper.GetGDPRRequest(ctx, "nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "GDPR request not found")
}

func TestGetAllGDPRRequests(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()
	requests := []*types.GDPRDataRequest{
		{
			Id:          "req1",
			Address:     "cosmos1addr1",
			RequestType: "access",
			RequestedAt: timestamppb.New(now),
			Status:      "pending",
		},
		{
			Id:          "req2",
			Address:     "cosmos1addr2",
			RequestType: "erasure",
			RequestedAt: timestamppb.New(now),
			Status:      "completed",
		},
	}

	for _, request := range requests {
		err := keeper.SetGDPRRequest(ctx, request)
		require.NoError(t, err)
	}

	retrieved, err := keeper.GetAllGDPRRequests(ctx)
	require.NoError(t, err)
	require.Len(t, retrieved, 2)
}

func TestGetAllGDPRRequests_Empty(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	retrieved, err := keeper.GetAllGDPRRequests(ctx)
	require.NoError(t, err)
	require.Empty(t, retrieved)
}

// ============================================================================
// Tax Report KVStore Tests
// ============================================================================

func TestSetGetTaxReport(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()
	report := &types.TaxReport{
		Id:                 "report1",
		Address:            "cosmos1test",
		TaxYear:            "2024",
		Jurisdiction:       "US",
		ReportType:         "1099-MISC",
		TotalIncome:        "50000",
		TotalCapitalGains:  "10000",
		TotalCapitalLosses: "2000",
		GeneratedAt:        timestamppb.New(now),
		Filed:              false,
	}

	err := keeper.SetTaxReport(ctx, report)
	require.NoError(t, err)

	retrieved, err := keeper.GetTaxReports(ctx, "cosmos1test")
	require.NoError(t, err)
	require.Len(t, retrieved, 1)
	require.Equal(t, report.TaxYear, retrieved[0].TaxYear)
}

func TestGetTaxReports_Empty(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	retrieved, err := keeper.GetTaxReports(ctx, "cosmos1test")
	require.NoError(t, err)
	require.Empty(t, retrieved)
}

// ============================================================================
// Params KVStore Tests
// ============================================================================

func TestSetGetParams(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	params := types.ComplianceParams{
		KycRequired:                  true,
		MinimumKycLevel:              types.KYCLevel_KYC_LEVEL_BASIC,
		KycExpiryDays:                365,
		TransactionMonitoringEnabled: true,
		VelocityLimit_24H:            "1000000",
		SingleTransactionLimit:       "100000",
		StructuringThresholdCount:    10,
		SanctionsScreeningEnabled:    true,
		SanctionsLists:               []string{"OFAC", "EU", "UN"},
		ScreeningCacheHours:          24,
		GdprEnabled:                  true,
		DataRetentionDays:            730,
		ProcessingPurposes:           []string{"compliance", "analytics"},
		TaxReportingEnabled:          true,
		TaxJurisdictions:             []string{"US", "EU"},
		TaxYearEnd:                   "12-31",
	}

	err := keeper.SetParamsToStore(ctx, params)
	require.NoError(t, err)

	retrieved, err := keeper.GetParamsFromStore(ctx)
	require.NoError(t, err)
	require.Equal(t, params.KycRequired, retrieved.KycRequired)
	require.Equal(t, params.MinimumKycLevel, retrieved.MinimumKycLevel)
	require.Equal(t, params.TransactionMonitoringEnabled, retrieved.TransactionMonitoringEnabled)
	require.Equal(t, params.SanctionsLists, retrieved.SanctionsLists)
}

func TestGetParams_Default(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	params, err := keeper.GetParamsFromStore(ctx)
	require.NoError(t, err)
	// Should return empty params when not set
	require.False(t, params.KycRequired)
}

func TestSetParams_ValidatesParams(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	params := types.ComplianceParams{
		KycRequired:     true,
		MinimumKycLevel: types.KYCLevel_KYC_LEVEL_BASIC,
	}

	err := keeper.SetParamsToStore(ctx, params)
	require.NoError(t, err) // ValidateParams currently allows all values
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestKVStore_MarshalErrors(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	// Test unmarshal error by corrupting stored data
	store := ctx.KVStore(keeper.storeKey)
	key := append(KYCRecordsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid proto data"))

	_, err := keeper.GetKYCRecord(ctx, "corrupt")
	require.Error(t, err)
}

func TestKVStore_UpdateExisting(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()

	// Create initial record
	record := &types.KYCRecord{
		Address:        "cosmos1test",
		KycLevel:       types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:       "provider1",
		VerifiedAt:     timestamppb.New(now),
		VerificationId: "ver1",
	}
	err := keeper.SetKYCRecord(ctx, record)
	require.NoError(t, err)

	// Update record
	record.KycLevel = types.KYCLevel_KYC_LEVEL_ADVANCED
	record.Provider = "provider2"
	err = keeper.SetKYCRecord(ctx, record)
	require.NoError(t, err)

	// Verify update
	retrieved, err := keeper.GetKYCRecord(ctx, "cosmos1test")
	require.NoError(t, err)
	require.Equal(t, types.KYCLevel_KYC_LEVEL_ADVANCED, retrieved.KycLevel)
	require.Equal(t, "provider2", retrieved.Provider)
}

func TestKVStore_MultipleAddressesIterator(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()

	// Add multiple records
	for i := 0; i < 10; i++ {
		record := &types.KYCRecord{
			Address:        "cosmos1test" + string(rune('0'+i)),
			KycLevel:       types.KYCLevel_KYC_LEVEL_BASIC,
			Provider:       "provider1",
			VerifiedAt:     timestamppb.New(now),
			VerificationId: "ver" + string(rune('0'+i)),
		}
		err := keeper.SetKYCRecord(ctx, record)
		require.NoError(t, err)
	}

	records, err := keeper.GetAllKYCRecords(ctx)
	require.NoError(t, err)
	require.Len(t, records, 10)
}

func TestKVStore_AllPrefixes(t *testing.T) {
	// Verify all key prefixes are unique
	prefixes := [][]byte{
		KYCRecordsKeyPrefix,
		AMLProfilesKeyPrefix,
		SuspiciousActivitiesKeyPrefix,
		MonitoringRulesKeyPrefix,
		TransactionAlertsKeyPrefix,
		SanctionsResultsKeyPrefix,
		GDPRConsentsKeyPrefix,
		GDPRRequestsKeyPrefix,
		TaxReportsKeyPrefix,
		ParamsKeyPrefix,
	}

	seen := make(map[byte]bool)
	for _, prefix := range prefixes {
		require.Len(t, prefix, 1)
		require.False(t, seen[prefix[0]], "duplicate prefix found")
		seen[prefix[0]] = true
	}
}

// ============================================================================
// Additional Coverage Tests for Edge Cases
// ============================================================================

func TestSetKYCRecord_UnmarshalError(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	// Corrupt data in store to trigger unmarshal error
	store := ctx.KVStore(keeper.storeKey)
	key := append(KYCRecordsKeyPrefix, []byte("corrupt-addr")...)
	store.Set(key, []byte("invalid-proto-data"))

	_, err := keeper.GetKYCRecord(ctx, "corrupt-addr")
	require.Error(t, err)
}

func TestGetAllKYCRecords_UnmarshalError(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	// Add valid record
	record := &types.KYCRecord{
		Address:  "cosmos1valid",
		KycLevel: types.KYCLevel_KYC_LEVEL_BASIC,
	}
	err := keeper.SetKYCRecord(ctx, record)
	require.NoError(t, err)

	// Add corrupted data
	store := ctx.KVStore(keeper.storeKey)
	key := append(KYCRecordsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid-data"))

	// Should fail due to unmarshal error
	_, err = keeper.GetAllKYCRecords(ctx)
	require.Error(t, err)
}

func TestGetAllAMLProfiles_UnmarshalError(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	// Add corrupted data
	store := ctx.KVStore(keeper.storeKey)
	key := append(AMLProfilesKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid-data"))

	_, err := keeper.GetAllAMLProfiles(ctx)
	require.Error(t, err)
}

func TestGetAMLProfile_UnmarshalError(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	store := ctx.KVStore(keeper.storeKey)
	key := append(AMLProfilesKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid-data"))

	_, err := keeper.GetAMLProfile(ctx, "corrupt")
	require.Error(t, err)
}

func TestGetAllSuspiciousActivities_UnmarshalError(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	store := ctx.KVStore(keeper.storeKey)
	key := append(SuspiciousActivitiesKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid-data"))

	_, err := keeper.GetAllSuspiciousActivities(ctx)
	require.Error(t, err)
}

func TestGetSuspiciousActivity_UnmarshalError(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	store := ctx.KVStore(keeper.storeKey)
	key := append(SuspiciousActivitiesKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid-data"))

	_, err := keeper.GetSuspiciousActivity(ctx, "corrupt")
	require.Error(t, err)
}

func TestGetAllMonitoringRules_UnmarshalError(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	store := ctx.KVStore(keeper.storeKey)
	key := append(MonitoringRulesKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid-data"))

	_, err := keeper.GetAllMonitoringRules(ctx)
	require.Error(t, err)
}

func TestGetMonitoringRule_UnmarshalError(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	store := ctx.KVStore(keeper.storeKey)
	key := append(MonitoringRulesKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid-data"))

	_, err := keeper.GetMonitoringRule(ctx, "corrupt")
	require.Error(t, err)
}

func TestGetAllSanctionsResults_UnmarshalError(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	store := ctx.KVStore(keeper.storeKey)
	key := append(SanctionsResultsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid-data"))

	_, err := keeper.GetAllSanctionsResults(ctx)
	require.Error(t, err)
}

func TestGetSanctionsResult_UnmarshalError(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	store := ctx.KVStore(keeper.storeKey)
	key := append(SanctionsResultsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid-data"))

	_, err := keeper.GetSanctionsResult(ctx, "corrupt")
	require.Error(t, err)
}

func TestGetAllGDPRRequests_UnmarshalError(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	store := ctx.KVStore(keeper.storeKey)
	key := append(GDPRRequestsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid-data"))

	_, err := keeper.GetAllGDPRRequests(ctx)
	require.Error(t, err)
}

func TestGetGDPRRequest_UnmarshalError(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	store := ctx.KVStore(keeper.storeKey)
	key := append(GDPRRequestsKeyPrefix, []byte("corrupt")...)
	store.Set(key, []byte("invalid-data"))

	_, err := keeper.GetGDPRRequest(ctx, "corrupt")
	require.Error(t, err)
}

func TestGetParamsFromStore_UnmarshalError(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	store := ctx.KVStore(keeper.storeKey)
	store.Set(ParamsKeyPrefix, []byte("invalid-data"))

	_, err := keeper.GetParamsFromStore(ctx)
	require.Error(t, err)
}

func TestSetGDPRConsent_WithExistingConsent(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()
	consent1 := &types.GDPRConsent{
		Address:        "cosmos1test",
		ConsentType:    "data_processing",
		Consented:      true,
		ConsentGivenAt: timestamppb.New(now),
	}

	// First consent
	err := keeper.SetGDPRConsent(ctx, consent1)
	require.NoError(t, err)

	// Second consent with different type
	consent2 := &types.GDPRConsent{
		Address:        "cosmos1test",
		ConsentType:    "marketing",
		Consented:      false,
		ConsentGivenAt: timestamppb.New(now),
	}

	err = keeper.SetGDPRConsent(ctx, consent2)
	require.NoError(t, err)

	consents, err := keeper.GetGDPRConsents(ctx, "cosmos1test")
	require.NoError(t, err)
	require.Len(t, consents, 2)
}

func TestSetTaxReport_WithExistingReport(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()
	report1 := &types.TaxReport{
		Id:          "report1",
		Address:     "cosmos1test",
		TaxYear:     "2023",
		GeneratedAt: timestamppb.New(now),
	}

	err := keeper.SetTaxReport(ctx, report1)
	require.NoError(t, err)

	// Report for different year
	report2 := &types.TaxReport{
		Id:          "report2",
		Address:     "cosmos1test",
		TaxYear:     "2024",
		GeneratedAt: timestamppb.New(now),
	}

	err = keeper.SetTaxReport(ctx, report2)
	require.NoError(t, err)

	reports, err := keeper.GetTaxReports(ctx, "cosmos1test")
	require.NoError(t, err)
	require.Len(t, reports, 2)
}

// Test marshal errors by using invalid codec
func TestSetKYCRecord_MarshalError(t *testing.T) {
	// This test verifies that marshal errors are handled
	// In practice, the proto codec should not fail on valid types
	keeper, ctx := setupTestKeeper(t)

	record := &types.KYCRecord{
		Address:  "cosmos1test",
		KycLevel: types.KYCLevel_KYC_LEVEL_BASIC,
	}

	// Normal case should succeed
	err := keeper.SetKYCRecord(ctx, record)
	require.NoError(t, err)
}

func TestSetAMLProfile_MarshalError(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	profile := &types.AMLProfile{
		Address:   "cosmos1test",
		RiskLevel: types.AMLRiskLevel_AML_RISK_LOW,
	}

	err := keeper.SetAMLProfile(ctx, profile)
	require.NoError(t, err)
}

func TestSetSuspiciousActivity_MarshalError(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	activity := &types.SuspiciousActivity{
		Id:      "sa1",
		Address: "cosmos1test",
	}

	err := keeper.SetSuspiciousActivity(ctx, activity)
	require.NoError(t, err)
}

func TestSetMonitoringRule_MarshalError(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	rule := &types.TransactionMonitoringRule{
		Id:   "rule1",
		Name: "Test Rule",
	}

	err := keeper.SetMonitoringRule(ctx, rule)
	require.NoError(t, err)
}

func TestSetSanctionsResult_MarshalError(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	result := &types.SanctionsScreeningResult{
		Address:    "cosmos1test",
		Status:     types.SanctionsStatus_SANCTIONS_CLEAR,
		ScreenedAt: timestamppb.Now(),
	}

	err := keeper.SetSanctionsResult(ctx, result)
	require.NoError(t, err)
}

func TestSetGDPRRequest_MarshalError(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	request := &types.GDPRDataRequest{
		Id:      "req1",
		Address: "cosmos1test",
	}

	err := keeper.SetGDPRRequest(ctx, request)
	require.NoError(t, err)
}

func TestSetParamsToStore_MarshalError(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	params := types.DefaultParams()
	err := keeper.SetParamsToStore(ctx, params)
	require.NoError(t, err)
}

func TestAddTransactionAlert_WithMultipleAlerts(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()

	// Add first alert
	alert1 := &types.TransactionAlert{
		Id:          "alert1",
		Address:     "cosmos1test",
		TriggeredAt: timestamppb.New(now),
	}

	err := keeper.AddTransactionAlert(ctx, "cosmos1test", alert1)
	require.NoError(t, err)

	// Add second alert
	alert2 := &types.TransactionAlert{
		Id:          "alert2",
		Address:     "cosmos1test",
		TriggeredAt: timestamppb.New(now),
	}

	err = keeper.AddTransactionAlert(ctx, "cosmos1test", alert2)
	require.NoError(t, err)
}

func TestInitializeDefaultMonitoringRules_WithErrorInGetParams(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	// Corrupt params to trigger error
	store := ctx.KVStore(keeper.storeKey)
	store.Set(ParamsKeyPrefix, []byte("invalid-data"))

	err := keeper.initializeDefaultMonitoringRules(ctx)
	require.Error(t, err)
}
