// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package security

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

const (
	// MaxChainIDLength is the maximum length for a chain ID
	MaxChainIDLength = 64

	// MinChainIDLength is the minimum length for a chain ID
	MinChainIDLength = 1

	// MaxMonikerLength is the maximum length for a moniker
	MaxMonikerLength = 128

	// MinMonikerLength is the minimum length for a moniker
	MinMonikerLength = 1

	// MaxAddressLength is the maximum length for an address
	MaxAddressLength = 256
)

// InputValidator validates user input
type InputValidator struct {
	logger Logger
}

// NewInputValidator creates a new input validator
func NewInputValidator(logger Logger) *InputValidator {
	return &InputValidator{
		logger: logger,
	}
}

// ValidateChainID validates a chain ID
func (iv *InputValidator) ValidateChainID(chainID string) error {
	if chainID == "" {
		return fmt.Errorf("chain ID cannot be empty")
	}

	// Check length
	if len(chainID) < MinChainIDLength || len(chainID) > MaxChainIDLength {
		iv.logger.SecurityEvent("chain_id_validation_failed", map[string]interface{}{
			"reason": "invalid_length",
			"length": len(chainID),
		})
		return fmt.Errorf("chain ID must be between %d and %d characters", MinChainIDLength, MaxChainIDLength)
	}

	// Chain ID should contain only alphanumeric characters and hyphens
	validChainID := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]$`)
	if !validChainID.MatchString(chainID) {
		iv.logger.SecurityEvent("chain_id_validation_failed", map[string]interface{}{
			"reason":   "invalid_format",
			"chain_id": chainID,
		})
		return fmt.Errorf("chain ID can only contain alphanumeric characters and hyphens, and must start/end with alphanumeric")
	}

	// Check for consecutive hyphens
	if strings.Contains(chainID, "--") {
		return fmt.Errorf("chain ID cannot contain consecutive hyphens")
	}

	return nil
}

// ValidateMoniker validates a node moniker
func (iv *InputValidator) ValidateMoniker(moniker string) error {
	if moniker == "" {
		return fmt.Errorf("moniker cannot be empty")
	}

	// Check length
	if len(moniker) < MinMonikerLength || len(moniker) > MaxMonikerLength {
		iv.logger.SecurityEvent("moniker_validation_failed", map[string]interface{}{
			"reason": "invalid_length",
			"length": len(moniker),
		})
		return fmt.Errorf("moniker must be between %d and %d characters", MinMonikerLength, MaxMonikerLength)
	}

	// Check for control characters
	for _, r := range moniker {
		if unicode.IsControl(r) {
			iv.logger.SecurityEvent("moniker_validation_failed", map[string]interface{}{
				"reason":  "control_character",
				"moniker": moniker,
			})
			return fmt.Errorf("moniker cannot contain control characters")
		}
	}

	// Check for dangerous characters
	dangerousChars := []string{
		"<", ">", "&", "'", "\"", "`", "$", "\\", "/", "|", ";", "\n", "\r",
	}
	for _, char := range dangerousChars {
		if strings.Contains(moniker, char) {
			iv.logger.SecurityEvent("moniker_validation_failed", map[string]interface{}{
				"reason":    "dangerous_character",
				"character": char,
			})
			return fmt.Errorf("moniker cannot contain character: %s", char)
		}
	}

	// Moniker should contain printable characters
	validMoniker := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9 _.-]*[a-zA-Z0-9]$`)
	if !validMoniker.MatchString(moniker) {
		iv.logger.SecurityEvent("moniker_validation_failed", map[string]interface{}{
			"reason":  "invalid_format",
			"moniker": moniker,
		})
		return fmt.Errorf("moniker can only contain alphanumeric characters, spaces, underscores, dots, and hyphens")
	}

	return nil
}

// ValidateAddress validates a blockchain address
func (iv *InputValidator) ValidateAddress(address string) error {
	if address == "" {
		return fmt.Errorf("address cannot be empty")
	}

	// Check length
	if len(address) > MaxAddressLength {
		iv.logger.SecurityEvent("address_validation_failed", map[string]interface{}{
			"reason": "too_long",
			"length": len(address),
		})
		return fmt.Errorf("address exceeds maximum length of %d", MaxAddressLength)
	}

	// Check for bech32 format (aura prefix)
	bech32Pattern := regexp.MustCompile(`^aura[a-z0-9]{39,59}$`)
	if !bech32Pattern.MatchString(address) {
		iv.logger.SecurityEvent("address_validation_failed", map[string]interface{}{
			"reason":  "invalid_bech32_format",
			"address": address,
		})
		return fmt.Errorf("address must be in bech32 format with 'aura' prefix")
	}

	return nil
}

// ValidateAmount validates a token amount
func (iv *InputValidator) ValidateAmount(amount string) error {
	if amount == "" {
		return fmt.Errorf("amount cannot be empty")
	}

	// Amount should be in format: <number><denom>
	// e.g., 1000uaura, 1.5aura
	amountPattern := regexp.MustCompile(`^[0-9]+(\.[0-9]+)?[a-z][a-z0-9]{2,15}$`)
	if !amountPattern.MatchString(amount) {
		iv.logger.SecurityEvent("amount_validation_failed", map[string]interface{}{
			"reason": "invalid_format",
			"amount": amount,
		})
		return fmt.Errorf("amount must be in format <number><denom> (e.g., 1000uaura)")
	}

	return nil
}

// ValidateNetworkAddress validates a network address (host:port)
func (iv *InputValidator) ValidateNetworkAddress(address string) error {
	if address == "" {
		return fmt.Errorf("network address cannot be empty")
	}

	// Check for localhost addresses
	if strings.HasPrefix(address, "localhost:") || strings.HasPrefix(address, "127.0.0.1:") {
		return nil
	}

	// Pattern for host:port
	// Support IPv4, IPv6, and hostnames
	patterns := []string{
		// IPv4:port
		`^([0-9]{1,3}\.){3}[0-9]{1,3}:[0-9]{1,5}$`,
		// hostname:port
		`^[a-zA-Z0-9][a-zA-Z0-9.-]*:[0-9]{1,5}$`,
		// [IPv6]:port
		`^\[[0-9a-fA-F:]+\]:[0-9]{1,5}$`,
	}

	valid := false
	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, address); matched {
			valid = true
			break
		}
	}

	if !valid {
		iv.logger.SecurityEvent("network_address_validation_failed", map[string]interface{}{
			"reason":  "invalid_format",
			"address": address,
		})
		return fmt.Errorf("network address must be in format host:port")
	}

	// Validate port range
	parts := strings.Split(address, ":")
	if len(parts) < 2 {
		return fmt.Errorf("missing port in network address")
	}

	portStr := parts[len(parts)-1]
	var port int
	if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
		return fmt.Errorf("invalid port number: %s", portStr)
	}

	if port < 1 || port > 65535 {
		iv.logger.SecurityEvent("network_address_validation_failed", map[string]interface{}{
			"reason": "invalid_port",
			"port":   port,
		})
		return fmt.Errorf("port must be between 1 and 65535")
	}

	return nil
}

// ValidateConfigKey validates a configuration key
func (iv *InputValidator) ValidateConfigKey(key string) error {
	if key == "" {
		return fmt.Errorf("config key cannot be empty")
	}

	// Check length
	if len(key) > 128 {
		return fmt.Errorf("config key too long")
	}

	// Key should contain only alphanumeric, dots, hyphens, and underscores
	validKey := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]*$`)
	if !validKey.MatchString(key) {
		iv.logger.SecurityEvent("config_key_validation_failed", map[string]interface{}{
			"reason": "invalid_format",
			"key":    key,
		})
		return fmt.Errorf("config key can only contain alphanumeric characters, dots, hyphens, and underscores")
	}

	return nil
}

// ValidateConfigValue validates a configuration value
func (iv *InputValidator) ValidateConfigValue(value string) error {
	// Check length
	if len(value) > 4096 {
		return fmt.Errorf("config value too long")
	}

	// Check for null bytes
	if strings.Contains(value, "\x00") {
		iv.logger.SecurityEvent("config_value_validation_failed", map[string]interface{}{
			"reason": "null_byte",
		})
		return fmt.Errorf("config value cannot contain null bytes")
	}

	// Check for control characters (except newlines which may be valid in some configs)
	for _, r := range value {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			iv.logger.SecurityEvent("config_value_validation_failed", map[string]interface{}{
				"reason": "control_character",
			})
			return fmt.Errorf("config value cannot contain control characters")
		}
	}

	return nil
}

// ValidateURL validates a URL
func (iv *InputValidator) ValidateURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Check length
	if len(url) > 2048 {
		return fmt.Errorf("URL too long")
	}

	// Check for valid HTTP/HTTPS URL
	urlPattern := regexp.MustCompile(`^https?://[a-zA-Z0-9][a-zA-Z0-9.-]*(:[0-9]{1,5})?(/.*)?$`)
	if !urlPattern.MatchString(url) {
		iv.logger.SecurityEvent("url_validation_failed", map[string]interface{}{
			"reason": "invalid_format",
			"url":    url,
		})
		return fmt.Errorf("URL must start with http:// or https://")
	}

	return nil
}

// SanitizeString removes potentially dangerous characters from a string
func (iv *InputValidator) SanitizeString(input string) string {
	// Remove control characters
	var result strings.Builder
	for _, r := range input {
		if !unicode.IsControl(r) || r == '\n' || r == '\r' || r == '\t' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ValidateKeyName validates a key name for the keyring
func (iv *InputValidator) ValidateKeyName(name string) error {
	if name == "" {
		return fmt.Errorf("key name cannot be empty")
	}

	// Check length
	if len(name) < 1 || len(name) > 64 {
		return fmt.Errorf("key name must be between 1 and 64 characters")
	}

	// Key name should contain only alphanumeric, hyphens, and underscores
	validKeyName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validKeyName.MatchString(name) {
		iv.logger.SecurityEvent("key_name_validation_failed", map[string]interface{}{
			"reason": "invalid_format",
			"name":   name,
		})
		return fmt.Errorf("key name can only contain alphanumeric characters, hyphens, and underscores")
	}

	return nil
}
