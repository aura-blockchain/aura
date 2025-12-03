package types

// ValidateParams validates ComplianceParams
func ValidateParams(p ComplianceParams) error {
	// Basic validation - all params are optional
	return nil
}

// DefaultParams returns default compliance parameters
func DefaultParams() ComplianceParams {
	return ComplianceParams{
		KycRequired:                  false,
		MinimumKycLevel:              0,
		KycExpiryDays:                365,
		ApprovedKycProviders:         []string{}, // Empty by default - must be configured
		TransactionMonitoringEnabled: false,
		VelocityLimit_24H:            "1000000",
		SingleTransactionLimit:       "100000",
		StructuringThresholdCount:    10,
		SanctionsScreeningEnabled:    false,
		SanctionsLists:               []string{},
		ScreeningCacheHours:          24,
	}
}
