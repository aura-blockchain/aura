package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

const (
	// MaxStringLength is the maximum allowed length for general strings
	MaxStringLength = 10000
	// MaxDescriptionLength is the maximum length for descriptions
	MaxDescriptionLength = 5000
	// MaxNameLength is the maximum length for names
	MaxNameLength = 256
	// MaxURLLength is the maximum length for URLs
	MaxURLLength = 2048
	// MinHashLength is the minimum length for hash strings
	MinHashLength = 32
	// MaxHashLength is the maximum length for hash strings (SHA-512 hex = 128)
	MaxHashLength = 128
)

var (
	// ErrInvalidAddress is returned when an address is invalid
	ErrInvalidAddress = fmt.Errorf("invalid address")
	// ErrInvalidAmount is returned when an amount is invalid
	ErrInvalidAmount = fmt.Errorf("invalid amount")
	// ErrInvalidString is returned when a string is invalid
	ErrInvalidString = fmt.Errorf("invalid string")
	// ErrInvalidURL is returned when a URL is invalid
	ErrInvalidURL = fmt.Errorf("invalid URL")
	// ErrInvalidHash is returned when a hash is invalid
	ErrInvalidHash = fmt.Errorf("invalid hash")
	// ErrEmptyField is returned when a required field is empty
	ErrEmptyField = fmt.Errorf("required field is empty")
	// ErrOutOfBounds is returned when a value is out of acceptable bounds
	ErrOutOfBounds = fmt.Errorf("value out of bounds")

	// Regex patterns for validation
	hexPattern         = regexp.MustCompile(`^[0-9a-fA-F]+$`)
	alphanumericPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	// DNS-safe pattern for denoms and identifiers
	denomPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9./\-_]{2,127}$`)
)

// ValidateAddress validates a bech32 address
// Returns error if the address is empty, malformed, or has invalid checksum
func ValidateAddress(addr string) error {
	if strings.TrimSpace(addr) == "" {
		return fmt.Errorf("%w: address cannot be empty", ErrInvalidAddress)
	}

	// Validate bech32 format
	_, _, err := bech32.DecodeAndConvert(addr)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidAddress, err.Error())
	}

	return nil
}

// ValidateAccAddress validates an account address using SDK's AccAddressFromBech32
// This is the primary function to use for validating user addresses
func ValidateAccAddress(addr string) error {
	if strings.TrimSpace(addr) == "" {
		return fmt.Errorf("%w: address cannot be empty", ErrInvalidAddress)
	}

	_, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidAddress, err.Error())
	}

	return nil
}

// ValidatePositiveInt validates that an integer is positive (> 0)
func ValidatePositiveInt(val sdkmath.Int, fieldName string) error {
	if val.IsNil() {
		return fmt.Errorf("%w: %s cannot be nil", ErrInvalidAmount, fieldName)
	}
	if !val.IsPositive() {
		return fmt.Errorf("%w: %s must be positive, got %s", ErrInvalidAmount, fieldName, val.String())
	}
	return nil
}

// ValidateNonNegativeInt validates that an integer is non-negative (>= 0)
func ValidateNonNegativeInt(val sdkmath.Int, fieldName string) error {
	if val.IsNil() {
		return fmt.Errorf("%w: %s cannot be nil", ErrInvalidAmount, fieldName)
	}
	if val.IsNegative() {
		return fmt.Errorf("%w: %s must be non-negative, got %s", ErrInvalidAmount, fieldName, val.String())
	}
	return nil
}

// ValidateBoundedInt validates that an integer is within specified bounds [min, max]
func ValidateBoundedInt(val sdkmath.Int, min, max sdkmath.Int, fieldName string) error {
	if val.IsNil() {
		return fmt.Errorf("%w: %s cannot be nil", ErrInvalidAmount, fieldName)
	}
	if val.LT(min) {
		return fmt.Errorf("%w: %s must be >= %s, got %s", ErrOutOfBounds, fieldName, min.String(), val.String())
	}
	if val.GT(max) {
		return fmt.Errorf("%w: %s must be <= %s, got %s", ErrOutOfBounds, fieldName, max.String(), val.String())
	}
	return nil
}

// ValidatePositiveDec validates that a decimal is positive (> 0)
func ValidatePositiveDec(val sdkmath.LegacyDec, fieldName string) error {
	if val.IsNil() {
		return fmt.Errorf("%w: %s cannot be nil", ErrInvalidAmount, fieldName)
	}
	if !val.IsPositive() {
		return fmt.Errorf("%w: %s must be positive, got %s", ErrInvalidAmount, fieldName, val.String())
	}
	return nil
}

// ValidateNonNegativeDec validates that a decimal is non-negative (>= 0)
func ValidateNonNegativeDec(val sdkmath.LegacyDec, fieldName string) error {
	if val.IsNil() {
		return fmt.Errorf("%w: %s cannot be nil", ErrInvalidAmount, fieldName)
	}
	if val.IsNegative() {
		return fmt.Errorf("%w: %s must be non-negative, got %s", ErrInvalidAmount, fieldName, val.String())
	}
	return nil
}

// ValidateBoundedDec validates that a decimal is within specified bounds [min, max]
func ValidateBoundedDec(val sdkmath.LegacyDec, min, max sdkmath.LegacyDec, fieldName string) error {
	if val.IsNil() {
		return fmt.Errorf("%w: %s cannot be nil", ErrInvalidAmount, fieldName)
	}
	if val.LT(min) {
		return fmt.Errorf("%w: %s must be >= %s, got %s", ErrOutOfBounds, fieldName, min.String(), val.String())
	}
	if val.GT(max) {
		return fmt.Errorf("%w: %s must be <= %s, got %s", ErrOutOfBounds, fieldName, max.String(), val.String())
	}
	return nil
}

// ValidateNonEmptyString validates that a string is not empty after trimming
func ValidateNonEmptyString(s string, fieldName string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("%w: %s cannot be empty", ErrEmptyField, fieldName)
	}
	return nil
}

// ValidateBoundedString validates string length is within bounds
// Also checks for control characters and excessive whitespace
func ValidateBoundedString(s string, minLen, maxLen int, fieldName string) error {
	trimmed := strings.TrimSpace(s)

	if len(trimmed) < minLen {
		return fmt.Errorf("%w: %s length must be >= %d, got %d", ErrInvalidString, fieldName, minLen, len(trimmed))
	}
	if len(trimmed) > maxLen {
		return fmt.Errorf("%w: %s length must be <= %d, got %d", ErrInvalidString, fieldName, maxLen, len(trimmed))
	}

	// Check for dangerous control characters
	if containsControlCharacters(s) {
		return fmt.Errorf("%w: %s contains invalid control characters", ErrInvalidString, fieldName)
	}

	return nil
}

// ValidateURL validates a URL string
func ValidateURL(urlStr string) error {
	if strings.TrimSpace(urlStr) == "" {
		return fmt.Errorf("%w: URL cannot be empty", ErrInvalidURL)
	}

	if len(urlStr) > MaxURLLength {
		return fmt.Errorf("%w: URL length exceeds maximum of %d", ErrInvalidURL, MaxURLLength)
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidURL, err.Error())
	}

	// Require scheme (http/https)
	if parsedURL.Scheme == "" {
		return fmt.Errorf("%w: URL must include scheme (http/https)", ErrInvalidURL)
	}

	// Only allow http and https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("%w: URL scheme must be http or https, got %s", ErrInvalidURL, parsedURL.Scheme)
	}

	// Require host
	if parsedURL.Host == "" {
		return fmt.Errorf("%w: URL must include host", ErrInvalidURL)
	}

	return nil
}

// ValidateHash validates a hex-encoded hash string
// Accepts common hash lengths: SHA-256 (64 chars), SHA-512 (128 chars), etc.
func ValidateHash(hash string) error {
	trimmed := strings.TrimSpace(hash)

	if trimmed == "" {
		return fmt.Errorf("%w: hash cannot be empty", ErrInvalidHash)
	}

	if len(trimmed) < MinHashLength {
		return fmt.Errorf("%w: hash length must be >= %d, got %d", ErrInvalidHash, MinHashLength, len(trimmed))
	}

	if len(trimmed) > MaxHashLength {
		return fmt.Errorf("%w: hash length must be <= %d, got %d", ErrInvalidHash, MaxHashLength, len(trimmed))
	}

	// Must be valid hex
	if !hexPattern.MatchString(trimmed) {
		return fmt.Errorf("%w: hash must be hex-encoded", ErrInvalidHash)
	}

	return nil
}

// ValidateDenom validates a coin denomination
// Denoms must start with a letter and contain only alphanumeric, '.', '/', '-', '_'
func ValidateDenom(denom string) error {
	if strings.TrimSpace(denom) == "" {
		return fmt.Errorf("%w: denom cannot be empty", ErrInvalidString)
	}

	if !denomPattern.MatchString(denom) {
		return fmt.Errorf("%w: denom must start with letter and contain only alphanumeric, '.', '/', '-', '_' (3-128 chars)", ErrInvalidString)
	}

	return nil
}

// ValidateCoin validates a Cosmos SDK Coin
func ValidateCoin(coin sdk.Coin, fieldName string) error {
	if err := ValidateDenom(coin.Denom); err != nil {
		return fmt.Errorf("%s: %w", fieldName, err)
	}

	if !coin.Amount.IsPositive() {
		return fmt.Errorf("%w: %s amount must be positive, got %s", ErrInvalidAmount, fieldName, coin.Amount.String())
	}

	return nil
}

// ValidateCoins validates a slice of Cosmos SDK Coins
func ValidateCoins(coins sdk.Coins, fieldName string) error {
	if len(coins) == 0 {
		return fmt.Errorf("%w: %s cannot be empty", ErrEmptyField, fieldName)
	}

	if err := coins.Validate(); err != nil {
		return fmt.Errorf("%s: %w", fieldName, err)
	}

	return nil
}

// ValidateAlphanumeric validates that a string contains only alphanumeric characters, underscores, and hyphens
func ValidateAlphanumeric(s string, fieldName string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("%w: %s cannot be empty", ErrEmptyField, fieldName)
	}

	if !alphanumericPattern.MatchString(s) {
		return fmt.Errorf("%w: %s must contain only alphanumeric characters, underscores, and hyphens", ErrInvalidString, fieldName)
	}

	return nil
}

// ValidateID validates an identifier (must be non-empty alphanumeric)
func ValidateID(id string, fieldName string) error {
	if err := ValidateNonEmptyString(id, fieldName); err != nil {
		return err
	}

	if err := ValidateBoundedString(id, 1, 128, fieldName); err != nil {
		return err
	}

	// IDs should be alphanumeric with underscores and hyphens
	if !alphanumericPattern.MatchString(id) {
		return fmt.Errorf("%w: %s must be alphanumeric (can include _ and -)", ErrInvalidString, fieldName)
	}

	return nil
}

// ValidateUint32Positive validates that a uint32 is greater than 0
func ValidateUint32Positive(val uint32, fieldName string) error {
	if val == 0 {
		return fmt.Errorf("%w: %s must be greater than 0", ErrInvalidAmount, fieldName)
	}
	return nil
}

// ValidateUint64Positive validates that a uint64 is greater than 0
func ValidateUint64Positive(val uint64, fieldName string) error {
	if val == 0 {
		return fmt.Errorf("%w: %s must be greater than 0", ErrInvalidAmount, fieldName)
	}
	return nil
}

// ValidateBoundedUint32 validates that a uint32 is within bounds [min, max]
func ValidateBoundedUint32(val uint32, min, max uint32, fieldName string) error {
	if val < min {
		return fmt.Errorf("%w: %s must be >= %d, got %d", ErrOutOfBounds, fieldName, min, val)
	}
	if val > max {
		return fmt.Errorf("%w: %s must be <= %d, got %d", ErrOutOfBounds, fieldName, max, val)
	}
	return nil
}

// ValidateBoundedUint64 validates that a uint64 is within bounds [min, max]
func ValidateBoundedUint64(val uint64, min, max uint64, fieldName string) error {
	if val < min {
		return fmt.Errorf("%w: %s must be >= %d, got %d", ErrOutOfBounds, fieldName, min, val)
	}
	if val > max {
		return fmt.Errorf("%w: %s must be <= %d, got %d", ErrOutOfBounds, fieldName, max, val)
	}
	return nil
}

// ValidateSliceNonEmpty validates that a slice is not empty
func ValidateSliceNonEmpty(slice interface{}, fieldName string) error {
	// Use type assertion to check length
	switch v := slice.(type) {
	case []string:
		if len(v) == 0 {
			return fmt.Errorf("%w: %s cannot be empty", ErrEmptyField, fieldName)
		}
	case []byte:
		if len(v) == 0 {
			return fmt.Errorf("%w: %s cannot be empty", ErrEmptyField, fieldName)
		}
	default:
		// Generic check using reflection would go here
		// For now, we'll just return nil
		return nil
	}
	return nil
}

// ValidateChainID validates a chain identifier
// Chain IDs must follow specific patterns (e.g., "paw", "xai", "aura", "osmosis-1")
func ValidateChainID(chainID string) error {
	if err := ValidateNonEmptyString(chainID, "chain_id"); err != nil {
		return err
	}

	if err := ValidateBoundedString(chainID, 2, 64, "chain_id"); err != nil {
		return err
	}

	// Chain IDs should be lowercase alphanumeric with hyphens
	chainIDPattern := regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
	if !chainIDPattern.MatchString(chainID) {
		return fmt.Errorf("%w: chain_id must be lowercase alphanumeric with hyphens", ErrInvalidString)
	}

	return nil
}

// SanitizeString removes control characters and trims whitespace
// Use this for user-provided strings before storage
func SanitizeString(s string) string {
	// First trim
	s = strings.TrimSpace(s)

	// Remove control characters except newlines and tabs (which we'll convert to spaces)
	var builder strings.Builder
	builder.Grow(len(s))

	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\t' {
			builder.WriteRune(' ')
		} else if !unicode.IsControl(r) {
			builder.WriteRune(r)
		}
	}

	// Collapse multiple spaces into one
	result := builder.String()
	spacePattern := regexp.MustCompile(`\s+`)
	result = spacePattern.ReplaceAllString(result, " ")

	return strings.TrimSpace(result)
}

// containsControlCharacters checks if a string contains dangerous control characters
// Allows newlines and tabs, but rejects other control characters
func containsControlCharacters(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return true
		}
	}
	return false
}

// ValidatePercentage validates that a value is a valid percentage (0-100)
func ValidatePercentage(val uint32, fieldName string) error {
	if val > 100 {
		return fmt.Errorf("%w: %s must be between 0 and 100, got %d", ErrOutOfBounds, fieldName, val)
	}
	return nil
}

// ValidateBasisPoints validates that a value is valid basis points (0-10000)
// Basis points are 1/100th of a percent, so 10000 bps = 100%
func ValidateBasisPoints(val uint64, fieldName string) error {
	if val > 10000 {
		return fmt.Errorf("%w: %s must be between 0 and 10000 basis points, got %d", ErrOutOfBounds, fieldName, val)
	}
	return nil
}

// ValidateTimestamp validates that a timestamp is not zero/negative
func ValidateTimestamp(ts int64, fieldName string) error {
	if ts < 0 {
		return fmt.Errorf("%w: %s cannot be negative, got %d", ErrInvalidAmount, fieldName, ts)
	}
	return nil
}

// ValidatePositiveTimestamp validates that a timestamp is positive (> 0)
func ValidatePositiveTimestamp(ts int64, fieldName string) error {
	if ts <= 0 {
		return fmt.Errorf("%w: %s must be positive, got %d", ErrInvalidAmount, fieldName, ts)
	}
	return nil
}

// ValidateBytes validates that a byte slice is not empty and within size limits
func ValidateBytes(data []byte, minLen, maxLen int, fieldName string) error {
	if len(data) < minLen {
		return fmt.Errorf("%w: %s length must be >= %d, got %d", ErrInvalidString, fieldName, minLen, len(data))
	}
	if len(data) > maxLen {
		return fmt.Errorf("%w: %s length must be <= %d, got %d", ErrInvalidString, fieldName, maxLen, len(data))
	}
	return nil
}

// ValidateStringSlice validates that all strings in a slice are non-empty
func ValidateStringSlice(slice []string, fieldName string) error {
	if len(slice) == 0 {
		return fmt.Errorf("%w: %s cannot be empty", ErrEmptyField, fieldName)
	}

	for i, s := range slice {
		if err := ValidateNonEmptyString(s, fmt.Sprintf("%s[%d]", fieldName, i)); err != nil {
			return err
		}
	}

	return nil
}
