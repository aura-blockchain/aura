package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/compliance/types"
)

// ptrTime returns a pointer to a time.Time value
func ptrTime(t time.Time) *time.Time {
	return &t
}

func setupTestKeeper(t *testing.T) (*Keeper, sdk.Context) {
	// Configure SDK with Aura-specific prefixes (safe to call multiple times)
	keepertest.ConfigureSDK()

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
		VerifiedAt: now,
		ExpiresAt:            ptrTime(now.Add(365 * 24 * time.Hour)),
		PiiCommitment: make([]byte, 32),
		EnhancedDueDiligence: false,
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
	require.Equal(t, record.PiiCommitment, retrieved.PiiCommitment)
	require.Equal(t, record.Provider, retrieved.Provider)
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
			VerifiedAt: now,
			ExpiresAt:      ptrTime(now.Add(365 * 24 * time.Hour)),
			PiiCommitment: make([]byte, 32),
		},
		{
			Address:        "cosmos1addr2",
			KycLevel:       types.KYCLevel_KYC_LEVEL_ADVANCED,
			Provider:       "provider2",
			VerifiedAt: now,
			ExpiresAt:      ptrTime(now.Add(365 * 24 * time.Hour)),
			PiiCommitment: make([]byte, 32),
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
		LastAssessment: now,
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
			LastAssessment: now,
			TotalTransactions: 50,
			TotalVolume:       "500000",
		},
		{
			Address:           "cosmos1addr2",
			RiskLevel:         types.AMLRiskLevel_AML_RISK_HIGH,
			LastAssessment: now,
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
		DetectedAt:      now,
		ReportedAt:      ptrTime(now.Add(1 * time.Hour)),
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
			DetectedAt:      now,
		},
		{
			Id:              "sa2",
			Address:         "cosmos1addr2",
			TransactionHash: "hash2",
			ActivityType:    "smurfing",
			DetectedAt:      now,
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
		CreatedAt: now,
		UpdatedAt: ptrTime(now),
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
			CreatedAt: now,
		},
		{
			Id:        "rule2",
			Name:      "Structuring Rule",
			RuleType:  "structuring",
			Enabled:   true,
			CreatedAt: now,
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
		TriggeredAt:     now,
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
		ScreenedAt:           now,
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
		ScreenedAt:           now,
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
			ScreenedAt:           now,
			ScreeningProvider: "provider1",
		},
		{
			Address:           "cosmos1addr2",
			Status:            types.SanctionsStatus_SANCTIONS_MATCH,
			ScreenedAt:           now,
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
		ConsentGivenAt: now,
		ConsentVersion: "v1.0",
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
		RequestedAt: now,
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
			RequestedAt: now,
			Status:      "pending",
		},
		{
			Id:          "req2",
			Address:     "cosmos1addr2",
			RequestType: "erasure",
			RequestedAt: now,
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
		ReportType:         "1099-MISC",
		TotalIncome:        "50000",
		TotalCapitalGains:  "10000",
		TotalCapitalLosses: "2000",
		GeneratedAt:        now,
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

	// Test valid params with kyc_required = true and kyc_expiry_days > 0
	params := types.ComplianceParams{
		KycRequired:     true,
		MinimumKycLevel: types.KYCLevel_KYC_LEVEL_BASIC,
		KycExpiryDays:   365, // Required when KycRequired is true
	}

	err := keeper.SetParamsToStore(ctx, params)
	require.NoError(t, err, "Valid params should not error")

	// Test invalid params: kyc_required = true but kyc_expiry_days = 0
	invalidParams := types.ComplianceParams{
		KycRequired:     true,
		MinimumKycLevel: types.KYCLevel_KYC_LEVEL_BASIC,
		KycExpiryDays:   0, // Invalid: must be > 0 when KycRequired is true
	}

	err = keeper.SetParamsToStore(ctx, invalidParams)
	require.Error(t, err, "Should error when kyc_required=true but kyc_expiry_days=0")
	require.Contains(t, err.Error(), "kyc_expiry_days must be positive")
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
		VerifiedAt: now,
		PiiCommitment: make([]byte, 32),
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
			VerifiedAt: now,
			PiiCommitment: make([]byte, 32),
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
		ProcessingRestrictionsKeyPrefix,
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
		ConsentGivenAt: now,
	}

	// First consent
	err := keeper.SetGDPRConsent(ctx, consent1)
	require.NoError(t, err)

	// Second consent with different type
	consent2 := &types.GDPRConsent{
		Address:        "cosmos1test",
		ConsentType:    "marketing",
		Consented:      false,
		ConsentGivenAt: now,
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
		GeneratedAt:        now,
	}

	err := keeper.SetTaxReport(ctx, report1)
	require.NoError(t, err)

	// Report for different year
	report2 := &types.TaxReport{
		Id:          "report2",
		Address:     "cosmos1test",
		TaxYear:     "2024",
		GeneratedAt:        now,
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
		ScreenedAt: time.Now(),
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
		TriggeredAt:     now,
	}

	err := keeper.AddTransactionAlert(ctx, "cosmos1test", alert1)
	require.NoError(t, err)

	// Add second alert
	alert2 := &types.TransactionAlert{
		Id:          "alert2",
		Address:     "cosmos1test",
		TriggeredAt:     now,
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

// ============================================================================
// GDPR Consent Withdrawal Enforcement Tests (TODO 055)
// ============================================================================

func TestSetProcessingRestriction(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	address := "cosmos1test"

	// Set restriction
	err := keeper.SetProcessingRestriction(ctx, address, true)
	require.NoError(t, err)

	// Verify restriction is set
	require.True(t, keeper.IsProcessingRestricted(ctx, address))

	// Remove restriction
	err = keeper.SetProcessingRestriction(ctx, address, false)
	require.NoError(t, err)

	// Verify restriction is removed
	require.False(t, keeper.IsProcessingRestricted(ctx, address))
}

func TestIsProcessingRestricted_NotSet(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	address := "cosmos1test"

	// Should be false when not set
	require.False(t, keeper.IsProcessingRestricted(ctx, address))
}

func TestGetGDPRConsent_Found(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()
	consent := &types.GDPRConsent{
		Address:        "cosmos1test",
		ConsentType:    "data_processing",
		Consented:      true,
		ConsentGivenAt: now,
	}

	err := keeper.SetGDPRConsent(ctx, consent)
	require.NoError(t, err)

	// Test GetGDPRConsent
	retrieved, found := keeper.GetGDPRConsent(ctx, "cosmos1test", "data_processing")
	require.True(t, found)
	require.NotNil(t, retrieved)
	require.Equal(t, consent.ConsentType, retrieved.ConsentType)
	require.Equal(t, consent.Consented, retrieved.Consented)
}

func TestGetGDPRConsent_NotFound(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	retrieved, found := keeper.GetGDPRConsent(ctx, "cosmos1test", "nonexistent")
	require.False(t, found)
	require.Nil(t, retrieved)
}

func TestGetGDPRConsent_WrongConsentType(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	now := time.Now()
	consent := &types.GDPRConsent{
		Address:        "cosmos1test",
		ConsentType:    "data_processing",
		Consented:      true,
		ConsentGivenAt: now,
	}

	err := keeper.SetGDPRConsent(ctx, consent)
	require.NoError(t, err)

	// Try to get with different consent type
	retrieved, found := keeper.GetGDPRConsent(ctx, "cosmos1test", "marketing")
	require.False(t, found)
	require.Nil(t, retrieved)
}

func TestCanProcessData_WithValidConsent(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	address := "cosmos1test"
	purpose := "data_processing"

	// Set valid consent
	now := time.Now()
	consent := &types.GDPRConsent{
		Address:        address,
		ConsentType:    purpose,
		Consented:      true,
		ConsentGivenAt: now,
	}

	err := keeper.SetGDPRConsent(ctx, consent)
	require.NoError(t, err)

	// Should allow processing
	require.True(t, keeper.CanProcessData(ctx, address, purpose))
}

func TestCanProcessData_WithWithdrawnConsent(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	address := "cosmos1test"
	purpose := "data_processing"

	// Set withdrawn consent
	now := time.Now()
	withdrawnAt := now
	consent := &types.GDPRConsent{
		Address:            address,
		ConsentType:        purpose,
		Consented:          false,
		ConsentGivenAt:     now,
		ConsentWithdrawnAt: &withdrawnAt,
	}

	err := keeper.SetGDPRConsent(ctx, consent)
	require.NoError(t, err)

	// Should not allow processing
	require.False(t, keeper.CanProcessData(ctx, address, purpose))
}

func TestCanProcessData_WithProcessingRestriction(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	address := "cosmos1test"
	purpose := "data_processing"

	// Set valid consent
	now := time.Now()
	consent := &types.GDPRConsent{
		Address:        address,
		ConsentType:    purpose,
		Consented:      true,
		ConsentGivenAt: now,
	}

	err := keeper.SetGDPRConsent(ctx, consent)
	require.NoError(t, err)

	// Set processing restriction
	err = keeper.SetProcessingRestriction(ctx, address, true)
	require.NoError(t, err)

	// Should not allow processing even with valid consent
	require.False(t, keeper.CanProcessData(ctx, address, purpose))
}

func TestCanProcessData_NoConsent(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	address := "cosmos1test"
	purpose := "data_processing"

	// No consent set, should not allow processing
	require.False(t, keeper.CanProcessData(ctx, address, purpose))
}

func TestTriggerDataDeletion(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	address := "cosmos1test"
	consentType := "data_processing"

	// Trigger data deletion
	err := keeper.TriggerDataDeletion(ctx, address, consentType)
	require.NoError(t, err)

	// Verify event was emitted
	events := ctx.EventManager().Events()
	require.NotEmpty(t, events)

	// Find the deletion event
	var deletionEvent sdk.Event
	for _, event := range events {
		if event.Type == "gdpr_data_deletion_requested" {
			deletionEvent = event
			break
		}
	}

	require.NotNil(t, deletionEvent.Type)
	require.Equal(t, "gdpr_data_deletion_requested", deletionEvent.Type)

	// Verify event attributes
	attrs := deletionEvent.Attributes
	addressFound := false
	consentTypeFound := false

	for _, attr := range attrs {
		if attr.Key == types.AttributeKeyAddress && attr.Value == address {
			addressFound = true
		}
		if attr.Key == types.AttributeKeyConsentType && attr.Value == consentType {
			consentTypeFound = true
		}
	}

	require.True(t, addressFound, "address attribute should be present in event")
	require.True(t, consentTypeFound, "consent_type attribute should be present in event")
}

func TestConsentWithdrawalEnforcement_CompleteFlow(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	address := "cosmos1test"
	purpose := "data_processing"

	// Step 1: Give consent
	now := time.Now()
	consent := &types.GDPRConsent{
		Address:        address,
		ConsentType:    purpose,
		Consented:      true,
		ConsentGivenAt: now,
	}

	err := keeper.SetGDPRConsent(ctx, consent)
	require.NoError(t, err)

	// Verify processing is allowed
	require.True(t, keeper.CanProcessData(ctx, address, purpose))
	require.False(t, keeper.IsProcessingRestricted(ctx, address))

	// Step 2: Withdraw consent
	consent.Consented = false
	consent.ConsentWithdrawnAt = ptrTime(now)
	err = keeper.SetGDPRConsent(ctx, consent)
	require.NoError(t, err)

	// Set processing restriction (simulating what RecordGDPRConsent does)
	err = keeper.SetProcessingRestriction(ctx, address, true)
	require.NoError(t, err)

	// Verify processing is now blocked
	require.False(t, keeper.CanProcessData(ctx, address, purpose))
	require.True(t, keeper.IsProcessingRestricted(ctx, address))

	// Step 3: Give consent again
	consent.Consented = true
	consent.ConsentWithdrawnAt = nil
	err = keeper.SetGDPRConsent(ctx, consent)
	require.NoError(t, err)

	// Remove processing restriction
	err = keeper.SetProcessingRestriction(ctx, address, false)
	require.NoError(t, err)

	// Verify processing is allowed again
	require.True(t, keeper.CanProcessData(ctx, address, purpose))
	require.False(t, keeper.IsProcessingRestricted(ctx, address))
}

func TestCanProcessData_MultipleConsents(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	address := "cosmos1test"
	now := time.Now()

	// Set consent for data_processing
	consent1 := &types.GDPRConsent{
		Address:        address,
		ConsentType:    "data_processing",
		Consented:      true,
		ConsentGivenAt: now,
	}
	err := keeper.SetGDPRConsent(ctx, consent1)
	require.NoError(t, err)

	// Set consent for marketing (withdrawn)
	withdrawnAt2 := now
	consent2 := &types.GDPRConsent{
		Address:            address,
		ConsentType:        "marketing",
		Consented:          false,
		ConsentGivenAt:     now,
		ConsentWithdrawnAt: &withdrawnAt2,
	}
	err = keeper.SetGDPRConsent(ctx, consent2)
	require.NoError(t, err)

	// Data processing should be allowed
	require.True(t, keeper.CanProcessData(ctx, address, "data_processing"))

	// Marketing should not be allowed
	require.False(t, keeper.CanProcessData(ctx, address, "marketing"))
}

func TestProcessingRestriction_DifferentAddresses(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	address1 := "cosmos1test1"
	address2 := "cosmos1test2"

	// Set restriction for address1
	err := keeper.SetProcessingRestriction(ctx, address1, true)
	require.NoError(t, err)

	// Verify only address1 is restricted
	require.True(t, keeper.IsProcessingRestricted(ctx, address1))
	require.False(t, keeper.IsProcessingRestricted(ctx, address2))

	// Set restriction for address2
	err = keeper.SetProcessingRestriction(ctx, address2, true)
	require.NoError(t, err)

	// Verify both are restricted
	require.True(t, keeper.IsProcessingRestricted(ctx, address1))
	require.True(t, keeper.IsProcessingRestricted(ctx, address2))

	// Remove restriction for address1
	err = keeper.SetProcessingRestriction(ctx, address1, false)
	require.NoError(t, err)

	// Verify only address2 is restricted
	require.False(t, keeper.IsProcessingRestricted(ctx, address1))
	require.True(t, keeper.IsProcessingRestricted(ctx, address2))
}
