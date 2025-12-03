package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// ============================================================================
// ValidateParams Tests
// ============================================================================

func TestValidateParams_Valid(t *testing.T) {
	params := ComplianceParams{
		KycRequired:                  true,
		MinimumKycLevel:              KYCLevel_KYC_LEVEL_BASIC,
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

	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidateParams_EmptyParams(t *testing.T) {
	params := ComplianceParams{}
	err := ValidateParams(params)
	require.NoError(t, err) // All params are optional
}

func TestValidateParams_PartialParams(t *testing.T) {
	params := ComplianceParams{
		KycRequired:     true,
		KycExpiryDays:   365,
		MinimumKycLevel: KYCLevel_KYC_LEVEL_ADVANCED,
	}

	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidateParams_AllFieldsSet(t *testing.T) {
	params := ComplianceParams{
		KycRequired:                  false,
		MinimumKycLevel:              KYCLevel_KYC_LEVEL_NONE,
		KycExpiryDays:                0,
		TransactionMonitoringEnabled: false,
		VelocityLimit_24H:            "0",
		SingleTransactionLimit:       "0",
		StructuringThresholdCount:    0,
		SanctionsScreeningEnabled:    false,
		SanctionsLists:               []string{},
		ScreeningCacheHours:          0,
		GdprEnabled:                  false,
		DataRetentionDays:            0,
		ProcessingPurposes:           []string{},
		TaxReportingEnabled:          false,
		TaxJurisdictions:             []string{},
		TaxYearEnd:                   "",
	}

	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidateParams_MaxValues(t *testing.T) {
	params := ComplianceParams{
		KycRequired:                  true,
		MinimumKycLevel:              KYCLevel_KYC_LEVEL_ADVANCED,
		KycExpiryDays:                365 * 10, // At max allowed
		TransactionMonitoringEnabled: true,
		VelocityLimit_24H:            "999999999999999",
		SingleTransactionLimit:       "999999999999999",
		StructuringThresholdCount:    10000, // At max allowed
		SanctionsScreeningEnabled:    true,
		SanctionsLists:               []string{"LIST1", "LIST2", "LIST3", "LIST4", "LIST5"},
		ScreeningCacheHours:          24 * 30, // At max allowed
		GdprEnabled:                  true,
		DataRetentionDays:            365 * 20, // At max allowed
		ProcessingPurposes:           []string{"purpose1", "purpose2", "purpose3"},
		TaxReportingEnabled:          true,
		TaxJurisdictions:             []string{"US", "EU", "UK", "CA"},
		TaxYearEnd:                   "12-31",
	}

	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidateParams_MultipleSanctionsLists(t *testing.T) {
	params := ComplianceParams{
		SanctionsScreeningEnabled: true,
		SanctionsLists:            []string{"OFAC", "EU", "UN", "UK", "CANADA"},
	}

	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidateParams_EmptySanctionsListsWhenDisabled(t *testing.T) {
	params := ComplianceParams{
		SanctionsScreeningEnabled: false,
		SanctionsLists:            []string{},
	}

	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidateParams_MultipleJurisdictions(t *testing.T) {
	params := ComplianceParams{
		TaxReportingEnabled: true,
		TaxJurisdictions:    []string{"US", "EU", "UK", "CA", "AU", "NZ"},
		TaxYearEnd:          "12-31",
	}

	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidateParams_DifferentKYCLevels(t *testing.T) {
	levels := []KYCLevel{
		KYCLevel_KYC_LEVEL_UNSPECIFIED,
		KYCLevel_KYC_LEVEL_NONE,
		KYCLevel_KYC_LEVEL_BASIC,
		KYCLevel_KYC_LEVEL_INTERMEDIATE,
		KYCLevel_KYC_LEVEL_ADVANCED,
	}

	for _, level := range levels {
		params := ComplianceParams{
			MinimumKycLevel: level,
		}
		err := ValidateParams(params)
		require.NoError(t, err, "level %v should be valid", level)
	}
}

// ============================================================================
// DefaultParams Tests
// ============================================================================

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()

	require.NotNil(t, params)
	require.False(t, params.KycRequired)
	require.Equal(t, KYCLevel(0), params.MinimumKycLevel)
	require.Equal(t, uint64(365), params.KycExpiryDays)
	require.False(t, params.TransactionMonitoringEnabled)
	require.Equal(t, "1000000", params.VelocityLimit_24H)
	require.Equal(t, "100000", params.SingleTransactionLimit)
	require.Equal(t, uint32(10), params.StructuringThresholdCount)
	require.False(t, params.SanctionsScreeningEnabled)
	require.Empty(t, params.SanctionsLists)
	require.Equal(t, uint64(24), params.ScreeningCacheHours)
}

func TestDefaultParams_IsValid(t *testing.T) {
	params := DefaultParams()
	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestDefaultParams_CanBeModified(t *testing.T) {
	params := DefaultParams()
	params.KycRequired = true
	params.MinimumKycLevel = KYCLevel_KYC_LEVEL_ADVANCED
	params.SanctionsLists = []string{"OFAC", "EU"}

	err := ValidateParams(params)
	require.NoError(t, err)
}

// ============================================================================
// Helper Types Tests
// ============================================================================

func TestTransactionAlertList(t *testing.T) {
	list := TransactionAlertList{
		Alerts: []*TransactionAlert{
			{
				Id:      "alert1",
				Address: "cosmos1test",
			},
			{
				Id:      "alert2",
				Address: "cosmos1test2",
			},
		},
	}

	require.Len(t, list.Alerts, 2)
	require.Equal(t, "alert1", list.Alerts[0].Id)
	require.Equal(t, "alert2", list.Alerts[1].Id)
}

func TestTransactionAlertList_Empty(t *testing.T) {
	list := TransactionAlertList{
		Alerts: []*TransactionAlert{},
	}

	require.NotNil(t, list.Alerts)
	require.Empty(t, list.Alerts)
}

func TestGDPRConsentList(t *testing.T) {
	list := GDPRConsentList{
		Consents: []*GDPRConsent{
			{
				Address:     "cosmos1test",
				ConsentType: "data_processing",
				Consented:   true,
			},
			{
				Address:     "cosmos1test",
				ConsentType: "marketing",
				Consented:   false,
			},
		},
	}

	require.Len(t, list.Consents, 2)
	require.True(t, list.Consents[0].Consented)
	require.False(t, list.Consents[1].Consented)
}

func TestGDPRConsentList_Empty(t *testing.T) {
	list := GDPRConsentList{
		Consents: []*GDPRConsent{},
	}

	require.NotNil(t, list.Consents)
	require.Empty(t, list.Consents)
}

func TestTaxReportList(t *testing.T) {
	list := TaxReportList{
		Reports: []*TaxReport{
			{
				Id:      "report1",
				Address: "cosmos1test",
				TaxYear: "2023",
			},
			{
				Id:      "report2",
				Address: "cosmos1test",
				TaxYear: "2024",
			},
		},
	}

	require.Len(t, list.Reports, 2)
	require.Equal(t, "2023", list.Reports[0].TaxYear)
	require.Equal(t, "2024", list.Reports[1].TaxYear)
}

func TestTaxReportList_Empty(t *testing.T) {
	list := TaxReportList{
		Reports: []*TaxReport{},
	}

	require.NotNil(t, list.Reports)
	require.Empty(t, list.Reports)
}

// ============================================================================
// Invalid Parameter Tests
// ============================================================================

func TestValidateParams_InvalidKYCExpiryDays(t *testing.T) {
	params := ComplianceParams{
		KycExpiryDays: 365*10 + 1, // Exceeds max
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "kyc_expiry_days too large")
}

func TestValidateParams_KYCRequiredWithZeroExpiry(t *testing.T) {
	params := ComplianceParams{
		KycRequired:   true,
		KycExpiryDays: 0,
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "kyc_expiry_days must be positive when kyc_required is true")
}

func TestValidateParams_KYCRequiredWithInvalidLevel(t *testing.T) {
	params := ComplianceParams{
		KycRequired:     true,
		KycExpiryDays:   365,
		MinimumKycLevel: KYCLevel_KYC_LEVEL_NONE,
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "minimum_kyc_level must be at least BASIC")
}

func TestValidateParams_NegativeVelocityLimit(t *testing.T) {
	params := ComplianceParams{
		VelocityLimit_24H: "-1000",
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "velocity_limit_24h cannot be negative")
}

func TestValidateParams_InvalidVelocityLimitFormat(t *testing.T) {
	params := ComplianceParams{
		VelocityLimit_24H: "not-a-number",
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "velocity_limit_24h is not a valid integer")
}

func TestValidateParams_NegativeSingleTransactionLimit(t *testing.T) {
	params := ComplianceParams{
		SingleTransactionLimit: "-500",
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "single_transaction_limit cannot be negative")
}

func TestValidateParams_InvalidSingleTransactionLimitFormat(t *testing.T) {
	params := ComplianceParams{
		SingleTransactionLimit: "invalid",
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "single_transaction_limit is not a valid integer")
}

func TestValidateParams_SingleTransactionExceedsVelocity(t *testing.T) {
	params := ComplianceParams{
		VelocityLimit_24H:      "1000",
		SingleTransactionLimit: "2000",
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "single_transaction_limit")
	require.Contains(t, err.Error(), "cannot exceed velocity_limit_24h")
}

func TestValidateParams_MonitoringEnabledWithZeroThreshold(t *testing.T) {
	params := ComplianceParams{
		TransactionMonitoringEnabled: true,
		StructuringThresholdCount:    0,
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "structuring_threshold_count must be positive")
}

func TestValidateParams_StructuringThresholdTooLarge(t *testing.T) {
	params := ComplianceParams{
		StructuringThresholdCount: 10001, // Exceeds max
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "structuring_threshold_count too large")
}

func TestValidateParams_ScreeningCacheHoursTooLarge(t *testing.T) {
	params := ComplianceParams{
		ScreeningCacheHours: 24*30 + 1, // Exceeds max
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "screening_cache_hours too large")
}

func TestValidateParams_SanctionsEnabledWithEmptyLists(t *testing.T) {
	params := ComplianceParams{
		SanctionsScreeningEnabled: true,
		SanctionsLists:            []string{},
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "sanctions_lists cannot be empty")
}

func TestValidateParams_SanctionsListWithEmptyEntry(t *testing.T) {
	params := ComplianceParams{
		SanctionsScreeningEnabled: true,
		SanctionsLists:            []string{"OFAC", "", "EU"},
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "sanctions_lists contains empty")
}

func TestValidateParams_SanctionsListNameTooLong(t *testing.T) {
	longName := string(make([]byte, 101))
	params := ComplianceParams{
		SanctionsScreeningEnabled: true,
		SanctionsLists:            []string{longName},
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "sanctions list name too long")
}

func TestValidateParams_DataRetentionDaysTooLarge(t *testing.T) {
	params := ComplianceParams{
		DataRetentionDays: 365*20 + 1, // Exceeds max
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "data_retention_days too large")
}

func TestValidateParams_GDPREnabledWithZeroRetention(t *testing.T) {
	params := ComplianceParams{
		GdprEnabled:       true,
		DataRetentionDays: 0,
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "data_retention_days must be positive when gdpr_enabled is true")
}

func TestValidateParams_GDPREnabledWithEmptyPurposes(t *testing.T) {
	params := ComplianceParams{
		GdprEnabled:        true,
		DataRetentionDays:  365,
		ProcessingPurposes: []string{},
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "processing_purposes cannot be empty")
}

func TestValidateParams_ProcessingPurposeWithEmptyEntry(t *testing.T) {
	params := ComplianceParams{
		GdprEnabled:        true,
		DataRetentionDays:  365,
		ProcessingPurposes: []string{"compliance", "", "analytics"},
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "processing_purposes contains empty")
}

func TestValidateParams_ProcessingPurposeTooLong(t *testing.T) {
	longPurpose := string(make([]byte, 201))
	params := ComplianceParams{
		GdprEnabled:        true,
		DataRetentionDays:  365,
		ProcessingPurposes: []string{longPurpose},
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "processing purpose too long")
}

func TestValidateParams_TaxReportingEnabledWithEmptyJurisdictions(t *testing.T) {
	params := ComplianceParams{
		TaxReportingEnabled: true,
		TaxYearEnd:          "12-31",
		TaxJurisdictions:    []string{},
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "tax_jurisdictions cannot be empty")
}

func TestValidateParams_TaxReportingEnabledWithoutYearEnd(t *testing.T) {
	params := ComplianceParams{
		TaxReportingEnabled: true,
		TaxJurisdictions:    []string{"US"},
		TaxYearEnd:          "",
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "tax_year_end must be set")
}

func TestValidateParams_InvalidTaxYearEndFormat(t *testing.T) {
	testCases := []struct {
		name    string
		yearEnd string
	}{
		{"missing day", "12"},
		{"too many parts", "12-31-2024"},
		{"invalid month", "13-31"},
		{"invalid day", "12-32"},
		{"zero month", "00-15"},
		{"zero day", "06-00"},
		{"invalid February", "02-30"},
		{"invalid April", "04-31"},
		{"non-numeric", "AA-BB"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := ComplianceParams{
				TaxYearEnd: tc.yearEnd,
			}
			err := ValidateParams(params)
			require.Error(t, err, "Expected error for year end: %s", tc.yearEnd)
			require.Contains(t, err.Error(), "tax_year_end")
		})
	}
}

func TestValidateParams_InvalidJurisdictionCode(t *testing.T) {
	testCases := []struct {
		name         string
		jurisdiction string
	}{
		{"too short", "U"},
		{"too long", "USA"},
		{"lowercase", "us"},
		{"numbers only", "12"},
		{"empty", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := ComplianceParams{
				BlockedJurisdictions: []string{tc.jurisdiction},
			}
			err := ValidateParams(params)
			require.Error(t, err, "Expected error for jurisdiction: %s", tc.jurisdiction)
			require.Contains(t, err.Error(), "invalid")
		})
	}
}

func TestValidateParams_EmptyKYCProviderAddress(t *testing.T) {
	params := ComplianceParams{
		ApprovedKycProviders: []string{"cosmos1validaddress", ""},
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "approved_kyc_providers contains empty")
}

func TestValidateParams_DuplicateKYCProvider(t *testing.T) {
	params := ComplianceParams{
		ApprovedKycProviders: []string{"cosmos1provider1", "cosmos1provider2", "cosmos1provider1"},
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate kyc provider")
}

func TestValidateParams_KYCProviderAddressTooShort(t *testing.T) {
	params := ComplianceParams{
		ApprovedKycProviders: []string{"short"},
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "kyc provider address too short")
}

func TestValidateParams_KYCProviderAddressTooLong(t *testing.T) {
	longAddress := string(make([]byte, 101))
	params := ComplianceParams{
		ApprovedKycProviders: []string{longAddress},
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "kyc provider address too long")
}

// ============================================================================
// Valid Edge Cases Tests
// ============================================================================

func TestValidateParams_ValidExtendedJurisdictionCodes(t *testing.T) {
	params := ComplianceParams{
		BlockedJurisdictions: []string{"US", "US-NY", "CA-ON", "GB-ENG"},
		TaxJurisdictions:     []string{"US", "US-CA", "EU-DE"},
	}
	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidateParams_ValidTaxYearEnds(t *testing.T) {
	testCases := []string{
		"12-31",
		"01-01",
		"06-30",
		"03-31",
		"02-29", // Leap year day
		"04-30", // 30-day month
	}

	for _, yearEnd := range testCases {
		t.Run(yearEnd, func(t *testing.T) {
			params := ComplianceParams{
				TaxYearEnd: yearEnd,
			}
			err := ValidateParams(params)
			require.NoError(t, err, "Should accept valid year end: %s", yearEnd)
		})
	}
}

func TestValidateParams_BoundaryValues(t *testing.T) {
	params := ComplianceParams{
		KycExpiryDays:             365 * 10,         // Exactly at max
		StructuringThresholdCount: 10000,           // Exactly at max
		ScreeningCacheHours:       24 * 30,         // Exactly at max
		DataRetentionDays:         365 * 20,        // Exactly at max
		VelocityLimit_24H:         "999999999999",  // Large valid number
		SingleTransactionLimit:    "999999999999",  // Large valid number
	}
	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidateParams_ConsistentLimits(t *testing.T) {
	params := ComplianceParams{
		VelocityLimit_24H:      "1000000",
		SingleTransactionLimit: "1000000", // Equal is allowed
	}
	err := ValidateParams(params)
	require.NoError(t, err)
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestValidateParams_NilSlices(t *testing.T) {
	params := ComplianceParams{
		SanctionsLists:     nil,
		ProcessingPurposes: nil,
		TaxJurisdictions:   nil,
	}

	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidateParams_LargeNumbers(t *testing.T) {
	// Test that very large numbers within allowed bounds are accepted
	params := ComplianceParams{
		KycExpiryDays:             365 * 10,     // At max allowed
		StructuringThresholdCount: 10000,       // At max allowed
		ScreeningCacheHours:       24 * 30,     // At max allowed
		DataRetentionDays:         365 * 20,    // At max allowed
	}

	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidateParams_VeryLargeStrings(t *testing.T) {
	// Test large numeric strings (should fail if not valid numbers)
	params := ComplianceParams{
		VelocityLimit_24H:      "999999999999999999999999999999",
		SingleTransactionLimit: "888888888888888888888888888888",
	}

	// These should be valid as long as they're parseable as integers
	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidateParams_SpecialCharacters(t *testing.T) {
	params := ComplianceParams{
		SanctionsLists:     []string{"OFAC-SDN", "EU_LIST", "UN.LIST", "UK/LIST"},
		ProcessingPurposes: []string{"compliance & analytics", "reporting@domain"},
		TaxJurisdictions:   []string{"US-NY", "CA-ON"}, // Valid extended format
		TaxYearEnd:         "12-31", // Must use dash format
	}

	err := ValidateParams(params)
	require.NoError(t, err)
}

// ============================================================================
// Type Alias Tests
// ============================================================================

func TestKYCLevelConstants(t *testing.T) {
	require.Equal(t, KYCLevel(0), KYCLevel_KYC_LEVEL_UNSPECIFIED)
	require.Equal(t, KYCLevel(1), KYCLevel_KYC_LEVEL_NONE)
	require.Equal(t, KYCLevel(2), KYCLevel_KYC_LEVEL_BASIC)
	require.Equal(t, KYCLevel(3), KYCLevel_KYC_LEVEL_INTERMEDIATE)
	require.Equal(t, KYCLevel(4), KYCLevel_KYC_LEVEL_ADVANCED)
}

func TestAMLRiskLevelConstants(t *testing.T) {
	require.Equal(t, AMLRiskLevel(0), AMLRiskLevel_AML_RISK_UNSPECIFIED)
	require.Equal(t, AMLRiskLevel(1), AMLRiskLevel_AML_RISK_LOW)
	require.Equal(t, AMLRiskLevel(2), AMLRiskLevel_AML_RISK_MEDIUM)
	require.Equal(t, AMLRiskLevel(3), AMLRiskLevel_AML_RISK_HIGH)
	require.Equal(t, AMLRiskLevel(4), AMLRiskLevel_AML_RISK_SEVERE)
}

func TestTransactionRiskLevelConstants(t *testing.T) {
	require.Equal(t, TransactionRiskLevel(0), TransactionRiskLevel_TX_RISK_UNSPECIFIED)
	require.Equal(t, TransactionRiskLevel(1), TransactionRiskLevel_TX_RISK_LOW)
	require.Equal(t, TransactionRiskLevel(2), TransactionRiskLevel_TX_RISK_MEDIUM)
	require.Equal(t, TransactionRiskLevel(3), TransactionRiskLevel_TX_RISK_HIGH)
	require.Equal(t, TransactionRiskLevel(4), TransactionRiskLevel_TX_RISK_CRITICAL)
}

func TestSanctionsStatusConstants(t *testing.T) {
	require.Equal(t, SanctionsStatus(0), SanctionsStatus_SANCTIONS_UNSPECIFIED)
	require.Equal(t, SanctionsStatus(1), SanctionsStatus_SANCTIONS_CLEAR)
	require.Equal(t, SanctionsStatus(2), SanctionsStatus_SANCTIONS_MATCH)
	require.Equal(t, SanctionsStatus(3), SanctionsStatus_SANCTIONS_CONFIRMED)
	require.Equal(t, SanctionsStatus(4), SanctionsStatus_SANCTIONS_PENDING_REVIEW)
}

// ============================================================================
// Comprehensive Params Combinations
// ============================================================================

func TestValidateParams_AllCombinations(t *testing.T) {
	testCases := []struct {
		name   string
		params ComplianceParams
	}{
		{
			name: "KYC only",
			params: ComplianceParams{
				KycRequired:     true,
				MinimumKycLevel: KYCLevel_KYC_LEVEL_BASIC,
				KycExpiryDays:   365,
			},
		},
		{
			name: "AML monitoring only",
			params: ComplianceParams{
				TransactionMonitoringEnabled: true,
				VelocityLimit_24H:            "1000000",
				SingleTransactionLimit:       "100000",
				StructuringThresholdCount:    10,
			},
		},
		{
			name: "Sanctions only",
			params: ComplianceParams{
				SanctionsScreeningEnabled: true,
				SanctionsLists:            []string{"OFAC"},
				ScreeningCacheHours:       24,
			},
		},
		{
			name: "GDPR only",
			params: ComplianceParams{
				GdprEnabled:        true,
				DataRetentionDays:  730,
				ProcessingPurposes: []string{"compliance"},
			},
		},
		{
			name: "Tax reporting only",
			params: ComplianceParams{
				TaxReportingEnabled: true,
				TaxJurisdictions:    []string{"US"},
				TaxYearEnd:          "12-31",
			},
		},
		{
			name: "All enabled",
			params: ComplianceParams{
				KycRequired:                  true,
				MinimumKycLevel:              KYCLevel_KYC_LEVEL_ADVANCED,
				KycExpiryDays:                365,
				TransactionMonitoringEnabled: true,
				VelocityLimit_24H:            "1000000",
				SingleTransactionLimit:       "100000",
				StructuringThresholdCount:    10,
				SanctionsScreeningEnabled:    true,
				SanctionsLists:               []string{"OFAC", "EU"},
				ScreeningCacheHours:          24,
				GdprEnabled:                  true,
				DataRetentionDays:            730,
				ProcessingPurposes:           []string{"compliance"},
				TaxReportingEnabled:          true,
				TaxJurisdictions:             []string{"US"},
				TaxYearEnd:                   "12-31",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateParams(tc.params)
			require.NoError(t, err)
		})
	}
}
