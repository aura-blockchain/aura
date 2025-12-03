package types

// ValidateParams validates ComplianceParams
func ValidateParams(p ComplianceParams) error {
	// Basic validation - all params are optional
	return nil
}

// DefaultBlockedJurisdictions returns OFAC-sanctioned countries
// Uses ISO 3166-1 alpha-2 country codes
// Source: OFAC Sanctions Programs and Country Information
// https://ofac.treasury.gov/sanctions-programs-and-country-information
var DefaultBlockedJurisdictions = []string{
	"KP", // North Korea (DPRK)
	"IR", // Iran
	"SY", // Syria
	"CU", // Cuba
	"RU", // Russia (comprehensive sanctions on certain sectors)
	"BY", // Belarus (sectoral sanctions)
}

// DefaultParams returns default compliance parameters
func DefaultParams() ComplianceParams {
	return ComplianceParams{
		KycRequired:                  false,
		MinimumKycLevel:              0,
		KycExpiryDays:                365,
		ApprovedKycProviders:         []string{}, // Empty by default - must be configured
		BlockedJurisdictions:         DefaultBlockedJurisdictions,
		TransactionMonitoringEnabled: false,
		VelocityLimit_24H:            "1000000",
		SingleTransactionLimit:       "100000",
		StructuringThresholdCount:    10,
		SanctionsScreeningEnabled:    false,
		SanctionsLists:               []string{},
		ScreeningCacheHours:          24,
	}
}
