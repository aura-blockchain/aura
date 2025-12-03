package types

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	sdkmath "cosmossdk.io/math"
)

// ValidateParams validates ComplianceParams
// Ensures all compliance parameters are within acceptable bounds and properly formatted
func ValidateParams(p ComplianceParams) error {
	// KYC validation
	if err := validateKYCParams(p); err != nil {
		return err
	}

	// Transaction monitoring validation
	if err := validateTransactionMonitoringParams(p); err != nil {
		return err
	}

	// Sanctions screening validation
	if err := validateSanctionsParams(p); err != nil {
		return err
	}

	// GDPR validation
	if err := validateGDPRParams(p); err != nil {
		return err
	}

	// Tax reporting validation
	if err := validateTaxReportingParams(p); err != nil {
		return err
	}

	// Blocked jurisdictions validation
	if err := validateBlockedJurisdictions(p.BlockedJurisdictions); err != nil {
		return err
	}

	// Approved KYC providers validation
	if err := validateKYCProviders(p.ApprovedKycProviders); err != nil {
		return err
	}

	return nil
}

// validateKYCParams validates KYC-related parameters
func validateKYCParams(p ComplianceParams) error {
	// KYC expiry days must be positive and reasonable
	if p.KycExpiryDays > 0 {
		const maxKYCExpiryDays uint64 = 365 * 10 // Max 10 years
		if p.KycExpiryDays > maxKYCExpiryDays {
			return fmt.Errorf("kyc_expiry_days too large: %d (max %d)", p.KycExpiryDays, maxKYCExpiryDays)
		}
	}
	// Note: Zero is allowed when KYC is disabled

	// Validate KYC level is within valid range
	if p.MinimumKycLevel < KYCLevel_KYC_LEVEL_UNSPECIFIED || p.MinimumKycLevel > KYCLevel_KYC_LEVEL_ADVANCED {
		return fmt.Errorf("invalid minimum_kyc_level: %d", p.MinimumKycLevel)
	}

	// If KYC is required, ensure reasonable configuration
	if p.KycRequired {
		if p.KycExpiryDays == 0 {
			return fmt.Errorf("kyc_expiry_days must be positive when kyc_required is true")
		}
		if p.MinimumKycLevel == KYCLevel_KYC_LEVEL_UNSPECIFIED || p.MinimumKycLevel == KYCLevel_KYC_LEVEL_NONE {
			return fmt.Errorf("minimum_kyc_level must be at least BASIC when kyc_required is true")
		}
	}

	return nil
}

// validateTransactionMonitoringParams validates transaction monitoring parameters
func validateTransactionMonitoringParams(p ComplianceParams) error {
	// Validate velocity limit (must be non-negative numeric string)
	if p.VelocityLimit_24H != "" {
		amount, ok := sdkmath.NewIntFromString(p.VelocityLimit_24H)
		if !ok {
			return fmt.Errorf("velocity_limit_24h is not a valid integer: %s", p.VelocityLimit_24H)
		}
		if amount.IsNegative() {
			return fmt.Errorf("velocity_limit_24h cannot be negative: %s", p.VelocityLimit_24H)
		}
	}

	// Validate single transaction limit (must be non-negative numeric string)
	if p.SingleTransactionLimit != "" {
		amount, ok := sdkmath.NewIntFromString(p.SingleTransactionLimit)
		if !ok {
			return fmt.Errorf("single_transaction_limit is not a valid integer: %s", p.SingleTransactionLimit)
		}
		if amount.IsNegative() {
			return fmt.Errorf("single_transaction_limit cannot be negative: %s", p.SingleTransactionLimit)
		}

		// Logical consistency: single transaction limit shouldn't exceed velocity limit
		if p.VelocityLimit_24H != "" {
			velocityLimit, _ := sdkmath.NewIntFromString(p.VelocityLimit_24H)
			if amount.GT(velocityLimit) {
				return fmt.Errorf("single_transaction_limit (%s) cannot exceed velocity_limit_24h (%s)",
					p.SingleTransactionLimit, p.VelocityLimit_24H)
			}
		}
	}

	// Structuring threshold count validation
	if p.TransactionMonitoringEnabled && p.StructuringThresholdCount == 0 {
		return fmt.Errorf("structuring_threshold_count must be positive when transaction_monitoring_enabled is true")
	}
	const maxStructuringThreshold uint32 = 10000 // Reasonable upper limit
	if p.StructuringThresholdCount > maxStructuringThreshold {
		return fmt.Errorf("structuring_threshold_count too large: %d (max %d)",
			p.StructuringThresholdCount, maxStructuringThreshold)
	}

	return nil
}

// validateSanctionsParams validates sanctions screening parameters
func validateSanctionsParams(p ComplianceParams) error {
	// Screening cache hours validation (cannot be negative)
	const maxCacheHours uint64 = 24 * 30 // Max 30 days cache
	if p.ScreeningCacheHours > maxCacheHours {
		return fmt.Errorf("screening_cache_hours too large: %d (max %d hours = 30 days)",
			p.ScreeningCacheHours, maxCacheHours)
	}

	// If sanctions screening is enabled, validate configuration
	if p.SanctionsScreeningEnabled {
		if len(p.SanctionsLists) == 0 {
			return fmt.Errorf("sanctions_lists cannot be empty when sanctions_screening_enabled is true")
		}

		// Validate sanctions list names (basic format check)
		for _, listName := range p.SanctionsLists {
			if strings.TrimSpace(listName) == "" {
				return fmt.Errorf("sanctions_lists contains empty or whitespace-only entry")
			}
			if len(listName) > 100 {
				return fmt.Errorf("sanctions list name too long: %s (max 100 chars)", listName)
			}
		}
	}

	return nil
}

// validateGDPRParams validates GDPR-related parameters
func validateGDPRParams(p ComplianceParams) error {
	// Data retention days validation
	const maxRetentionDays uint64 = 365 * 20 // Max 20 years
	if p.DataRetentionDays > maxRetentionDays {
		return fmt.Errorf("data_retention_days too large: %d (max %d = 20 years)",
			p.DataRetentionDays, maxRetentionDays)
	}

	// If GDPR is enabled, ensure proper configuration
	if p.GdprEnabled {
		if p.DataRetentionDays == 0 {
			return fmt.Errorf("data_retention_days must be positive when gdpr_enabled is true")
		}
		if len(p.ProcessingPurposes) == 0 {
			return fmt.Errorf("processing_purposes cannot be empty when gdpr_enabled is true")
		}

		// Validate processing purposes format
		for _, purpose := range p.ProcessingPurposes {
			if strings.TrimSpace(purpose) == "" {
				return fmt.Errorf("processing_purposes contains empty or whitespace-only entry")
			}
			if len(purpose) > 200 {
				return fmt.Errorf("processing purpose too long: %s (max 200 chars)", purpose)
			}
		}
	}

	return nil
}

// validateTaxReportingParams validates tax reporting parameters
func validateTaxReportingParams(p ComplianceParams) error {
	// Validate tax year end format (MM-DD)
	if p.TaxYearEnd != "" {
		if err := validateTaxYearEndFormat(p.TaxYearEnd); err != nil {
			return err
		}
	}

	// If tax reporting is enabled, ensure proper configuration
	if p.TaxReportingEnabled {
		if len(p.TaxJurisdictions) == 0 {
			return fmt.Errorf("tax_jurisdictions cannot be empty when tax_reporting_enabled is true")
		}
		if p.TaxYearEnd == "" {
			return fmt.Errorf("tax_year_end must be set when tax_reporting_enabled is true")
		}

		// Validate tax jurisdiction codes
		for _, jurisdiction := range p.TaxJurisdictions {
			if err := ValidateJurisdictionCode(jurisdiction); err != nil {
				return fmt.Errorf("invalid tax jurisdiction: %w", err)
			}
		}
	}

	return nil
}

// validateBlockedJurisdictions validates jurisdiction code format
func validateBlockedJurisdictions(jurisdictions []string) error {
	for _, j := range jurisdictions {
		if err := ValidateJurisdictionCode(j); err != nil {
			return fmt.Errorf("invalid blocked jurisdiction: %w", err)
		}
	}
	return nil
}

// ValidateJurisdictionCode validates ISO 3166-1 alpha-2 country codes
// Accepts 2-letter codes and optionally extended formats like "US-NY" for subdivisions
func ValidateJurisdictionCode(code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return fmt.Errorf("jurisdiction code cannot be empty")
	}

	// Allow ISO 3166-1 alpha-2 (2 letters) or extended formats like "US-NY"
	// Pattern: 2 uppercase letters, optionally followed by dash and 2+ alphanumeric chars
	validPattern := regexp.MustCompile(`^[A-Z]{2}(-[A-Z0-9]{2,})?$`)
	if !validPattern.MatchString(code) {
		return fmt.Errorf("invalid jurisdiction code %s: must be 2-letter ISO code (e.g., 'US') or extended format (e.g., 'US-NY')", code)
	}

	return nil
}

// validateTaxYearEndFormat validates tax year end format (MM-DD)
func validateTaxYearEndFormat(yearEnd string) error {
	parts := strings.Split(yearEnd, "-")
	if len(parts) != 2 {
		return fmt.Errorf("invalid tax_year_end format %s: must be MM-DD", yearEnd)
	}

	month, err := strconv.Atoi(parts[0])
	if err != nil || month < 1 || month > 12 {
		return fmt.Errorf("invalid month in tax_year_end %s: must be 01-12", yearEnd)
	}

	day, err := strconv.Atoi(parts[1])
	if err != nil || day < 1 || day > 31 {
		return fmt.Errorf("invalid day in tax_year_end %s: must be 01-31", yearEnd)
	}

	// Additional validation for specific month-day combinations
	if month == 2 && day > 29 {
		return fmt.Errorf("invalid day for February in tax_year_end %s", yearEnd)
	}
	if (month == 4 || month == 6 || month == 9 || month == 11) && day > 30 {
		return fmt.Errorf("invalid day for month %d in tax_year_end %s", month, yearEnd)
	}

	return nil
}

// validateKYCProviders validates KYC provider addresses
func validateKYCProviders(providers []string) error {
	seen := make(map[string]bool)
	for _, provider := range providers {
		provider = strings.TrimSpace(provider)
		if provider == "" {
			return fmt.Errorf("approved_kyc_providers contains empty or whitespace-only entry")
		}

		// Check for duplicates
		if seen[provider] {
			return fmt.Errorf("duplicate kyc provider: %s", provider)
		}
		seen[provider] = true

		// Basic format validation (Cosmos bech32 addresses typically start with prefix)
		if len(provider) < 10 {
			return fmt.Errorf("kyc provider address too short: %s", provider)
		}
		if len(provider) > 100 {
			return fmt.Errorf("kyc provider address too long: %s", provider)
		}
	}
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

// ValidateFilePath validates file paths for security against path traversal attacks
// This prevents attackers from using .. sequences, absolute paths, or suspicious characters
// to write files outside of the intended directory
//
// Security requirements:
//   - Reject empty paths
//   - Reject absolute paths (starting with /)
//   - Reject path traversal sequences (..)
//   - Reject suspicious characters that could enable attacks (<>|"'`)
//   - Ensure cleaned path doesn't escape base directory
//
// Returns an error if the path is invalid or could be used in an attack
func ValidateFilePath(path string) error {
	// Reject empty paths
	if path == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Reject paths that are too long (reasonable limit for file paths)
	const maxPathLength = 4096 // Standard Linux PATH_MAX
	if len(path) > maxPathLength {
		return fmt.Errorf("file path too long: %d bytes (max %d)", len(path), maxPathLength)
	}

	// Reject absolute paths (must be relative to base directory)
	// This prevents writing to arbitrary locations like /etc/passwd
	if strings.HasPrefix(path, "/") {
		return fmt.Errorf("absolute paths not allowed: %s", path)
	}

	// Reject Windows-style absolute paths (C:\, D:\, etc.)
	if len(path) >= 3 && path[1] == ':' && (path[2] == '\\' || path[2] == '/') {
		return fmt.Errorf("Windows-style absolute paths not allowed: %s", path)
	}

	// Reject explicit path traversal sequences
	// This prevents attacks like ../../etc/passwd
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal sequences (..) not allowed: %s", path)
	}

	// Reject suspicious characters that could be used in injection attacks
	// < > | " ' ` ; are commonly used in command injection and shell attacks
	// Semicolon (;) is particularly dangerous for command chaining
	suspiciousChars := "<>|\"'`;$&"
	for _, char := range suspiciousChars {
		if strings.ContainsRune(path, char) {
			return fmt.Errorf("invalid character '%c' in file path: %s", char, path)
		}
	}

	// Reject null bytes (used in null byte injection attacks)
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("null bytes not allowed in file path")
	}

	// Reject paths containing newlines or other control characters
	for _, char := range path {
		if char < 32 && char != 9 { // Allow tab (9), reject other control characters
			return fmt.Errorf("control characters not allowed in file path")
		}
	}

	// Clean the path and verify it doesn't escape the base directory
	// filepath.Clean normalizes the path by removing ./ and ../
	// but we already rejected .. above as defense in depth
	cleaned := strings.TrimSpace(path)
	if cleaned == "" {
		return fmt.Errorf("file path cannot be whitespace only")
	}

	// Reject paths containing consecutive slashes (path normalization evasion)
	// Check this before other validations to catch "//etc//passwd" type attacks
	if strings.Contains(path, "//") {
		return fmt.Errorf("consecutive slashes not allowed in file path: %s", path)
	}

	// Reject paths that contain hidden files (any segment starting with .)
	// This prevents access to files like .env, .ssh/id_rsa, .bashrc, reports/.secret.txt
	if strings.HasPrefix(cleaned, ".") && cleaned != "." {
		return fmt.Errorf("hidden file paths not allowed: %s", path)
	}
	// Also check for hidden files in any path segment
	for _, segment := range strings.Split(path, "/") {
		if strings.HasPrefix(segment, ".") && segment != "" && segment != "." && segment != ".." {
			return fmt.Errorf("hidden file paths not allowed: %s", path)
		}
	}

	// Additional validation: ensure the path doesn't try to escape via multiple techniques
	// This catches edge cases where the path might be valid but still suspicious
	pathSegments := strings.Split(path, "/")
	for _, segment := range pathSegments {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue // Empty segments from trailing slashes are okay
		}
		if segment == "." || segment == ".." {
			return fmt.Errorf("relative path components (. or ..) not allowed: %s", path)
		}
	}

	return nil
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
