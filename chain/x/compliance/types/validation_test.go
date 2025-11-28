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
		KycExpiryDays:                999999,
		TransactionMonitoringEnabled: true,
		VelocityLimit_24H:            "999999999999999",
		SingleTransactionLimit:       "999999999999999",
		StructuringThresholdCount:    999999,
		SanctionsScreeningEnabled:    true,
		SanctionsLists:               []string{"LIST1", "LIST2", "LIST3", "LIST4", "LIST5"},
		ScreeningCacheHours:          999999,
		GdprEnabled:                  true,
		DataRetentionDays:            999999,
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

func TestValidateParams_EmptySanctionsLists(t *testing.T) {
	params := ComplianceParams{
		SanctionsScreeningEnabled: true,
		SanctionsLists:            []string{},
	}

	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidateParams_MultipleJurisdictions(t *testing.T) {
	params := ComplianceParams{
		TaxReportingEnabled: true,
		TaxJurisdictions:    []string{"US", "EU", "UK", "CA", "AU", "NZ"},
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
	params := ComplianceParams{
		KycExpiryDays:             18446744073709551615, // max uint64
		StructuringThresholdCount: 4294967295,           // max uint32
		ScreeningCacheHours:       18446744073709551615, // max uint64
		DataRetentionDays:         18446744073709551615, // max uint64
	}

	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidateParams_VeryLargeStrings(t *testing.T) {
	longString := string(make([]byte, 10000))
	params := ComplianceParams{
		VelocityLimit_24H:      longString,
		SingleTransactionLimit: longString,
		TaxYearEnd:             longString,
	}

	err := ValidateParams(params)
	require.NoError(t, err) // Currently no validation on string length
}

func TestValidateParams_SpecialCharacters(t *testing.T) {
	params := ComplianceParams{
		SanctionsLists:     []string{"OFAC-SDN", "EU_LIST", "UN.LIST", "UK/LIST"},
		ProcessingPurposes: []string{"compliance & analytics", "reporting@domain"},
		TaxJurisdictions:   []string{"US-NY", "EU/UK", "CA_ON"},
		TaxYearEnd:         "12/31",
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
