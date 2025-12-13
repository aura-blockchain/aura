package security

import (
	"fmt"
	"regexp"
	"strings"
)

// ConfigValidator validates configuration keys and values
type ConfigValidator struct {
	allowedKeys map[string]ConfigKeyValidator
	logger      Logger
}

// ConfigKeyValidator validates a specific configuration key
type ConfigKeyValidator struct {
	Description string
	Validator   func(string) error
	MaxLength   int
}

// NewConfigValidator creates a new config validator
func NewConfigValidator(logger Logger) *ConfigValidator {
	cv := &ConfigValidator{
		allowedKeys: make(map[string]ConfigKeyValidator),
		logger:      logger,
	}

	// Define allowed configuration keys and their validators
	cv.registerDefaultKeys()

	return cv
}

// registerDefaultKeys registers default configuration keys
func (cv *ConfigValidator) registerDefaultKeys() {
	// Chain configuration
	cv.allowedKeys["chain-id"] = ConfigKeyValidator{
		Description: "Chain ID",
		Validator:   cv.validateChainID,
		MaxLength:   64,
	}

	cv.allowedKeys["moniker"] = ConfigKeyValidator{
		Description: "Node moniker",
		Validator:   cv.validateMoniker,
		MaxLength:   128,
	}

	// Network configuration
	cv.allowedKeys["rpc.laddr"] = ConfigKeyValidator{
		Description: "RPC listen address",
		Validator:   cv.validateNetworkAddress,
		MaxLength:   256,
	}

	cv.allowedKeys["p2p.laddr"] = ConfigKeyValidator{
		Description: "P2P listen address",
		Validator:   cv.validateNetworkAddress,
		MaxLength:   256,
	}

	cv.allowedKeys["p2p.external_address"] = ConfigKeyValidator{
		Description: "P2P external address",
		Validator:   cv.validateNetworkAddress,
		MaxLength:   256,
	}

	cv.allowedKeys["p2p.seeds"] = ConfigKeyValidator{
		Description: "Seed nodes",
		Validator:   cv.validatePeerList,
		MaxLength:   4096,
	}

	cv.allowedKeys["p2p.persistent_peers"] = ConfigKeyValidator{
		Description: "Persistent peers",
		Validator:   cv.validatePeerList,
		MaxLength:   4096,
	}

	// gRPC configuration
	cv.allowedKeys["grpc.enable"] = ConfigKeyValidator{
		Description: "Enable gRPC server",
		Validator:   cv.validateBoolean,
		MaxLength:   5,
	}

	cv.allowedKeys["grpc.address"] = ConfigKeyValidator{
		Description: "gRPC server address",
		Validator:   cv.validateNetworkAddress,
		MaxLength:   256,
	}

	// API configuration
	cv.allowedKeys["api.enable"] = ConfigKeyValidator{
		Description: "Enable API server",
		Validator:   cv.validateBoolean,
		MaxLength:   5,
	}

	cv.allowedKeys["api.address"] = ConfigKeyValidator{
		Description: "API server address",
		Validator:   cv.validateNetworkAddress,
		MaxLength:   256,
	}

	cv.allowedKeys["api.enabled-unsafe-cors"] = ConfigKeyValidator{
		Description: "Enable unsafe CORS",
		Validator:   cv.validateBoolean,
		MaxLength:   5,
	}

	// Logging configuration
	cv.allowedKeys["logging.level"] = ConfigKeyValidator{
		Description: "Logging level",
		Validator:   cv.validateLogLevel,
		MaxLength:   16,
	}

	cv.allowedKeys["logging.format"] = ConfigKeyValidator{
		Description: "Logging format",
		Validator:   cv.validateLogFormat,
		MaxLength:   16,
	}

	// Consensus configuration
	cv.allowedKeys["consensus.timeout_propose"] = ConfigKeyValidator{
		Description: "Propose timeout",
		Validator:   cv.validateDuration,
		MaxLength:   32,
	}

	cv.allowedKeys["consensus.timeout_prevote"] = ConfigKeyValidator{
		Description: "Prevote timeout",
		Validator:   cv.validateDuration,
		MaxLength:   32,
	}

	cv.allowedKeys["consensus.timeout_precommit"] = ConfigKeyValidator{
		Description: "Precommit timeout",
		Validator:   cv.validateDuration,
		MaxLength:   32,
	}

	cv.allowedKeys["consensus.timeout_commit"] = ConfigKeyValidator{
		Description: "Commit timeout",
		Validator:   cv.validateDuration,
		MaxLength:   32,
	}

	// Module configurations
	modules := []string{
		"identitychange", "inclusionroutines", "confidencescore",
		"vcregistry", "dataregistry", "governance", "dex", "bridge",
	}

	for _, module := range modules {
		key := fmt.Sprintf("modules.%s.enabled", module)
		cv.allowedKeys[key] = ConfigKeyValidator{
			Description: fmt.Sprintf("Enable %s module", module),
			Validator:   cv.validateBoolean,
			MaxLength:   5,
		}
	}
}

// ValidateKey validates a configuration key
func (cv *ConfigValidator) ValidateKey(key string) error {
	if key == "" {
		return fmt.Errorf("configuration key cannot be empty")
	}

	// Check key length
	if len(key) > 256 {
		return fmt.Errorf("configuration key too long")
	}

	// Check if key is in whitelist
	if _, allowed := cv.allowedKeys[key]; !allowed {
		cv.logger.SecurityEvent("config_key_not_allowed", map[string]interface{}{
			"key": key,
		})
		return fmt.Errorf("configuration key not allowed: %s", key)
	}

	// Validate key format
	if !cv.isValidKeyFormat(key) {
		return fmt.Errorf("invalid configuration key format")
	}

	return nil
}

// ValidateValue validates a configuration value for a given key
func (cv *ConfigValidator) ValidateValue(key, value string) error {
	// Validate key first
	if err := cv.ValidateKey(key); err != nil {
		return err
	}

	validator := cv.allowedKeys[key]

	// Check value length
	if len(value) > validator.MaxLength {
		cv.logger.SecurityEvent("config_value_too_long", map[string]interface{}{
			"key":        key,
			"max_length": validator.MaxLength,
		})
		return fmt.Errorf("value for %s exceeds maximum length of %d", key, validator.MaxLength)
	}

	// Check for TOML injection
	if cv.containsTOMLInjection(value) {
		cv.logger.SecurityEvent("toml_injection_detected", map[string]interface{}{
			"key": key,
		})
		return fmt.Errorf("value contains potential TOML injection")
	}

	// Run specific validator
	if validator.Validator != nil {
		if err := validator.Validator(value); err != nil {
			return fmt.Errorf("invalid value for %s: %w", key, err)
		}
	}

	return nil
}

// isValidKeyFormat checks if the key has a valid format
func (cv *ConfigValidator) isValidKeyFormat(key string) bool {
	// Key should contain only alphanumeric, dots, hyphens, and underscores
	validKey := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]*$`)
	return validKey.MatchString(key)
}

// containsTOMLInjection checks for TOML injection patterns
func (cv *ConfigValidator) containsTOMLInjection(value string) bool {
	// Check for patterns that could inject new TOML keys
	injectionPatterns := []string{
		"\n[",      // New section
		"\n[[",     // New array section
		"]\n",      // Close section prematurely
		"]]\n",     // Close array section
		"\"\"\"\n", // Multiline string break
		"'''\n",    // Alternative multiline string
	}

	for _, pattern := range injectionPatterns {
		if strings.Contains(value, pattern) {
			return true
		}
	}

	return false
}

// Validator functions

func (cv *ConfigValidator) validateChainID(value string) error {
	if value == "" {
		return fmt.Errorf("chain ID cannot be empty")
	}
	validChainID := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]$`)
	if !validChainID.MatchString(value) {
		return fmt.Errorf("invalid chain ID format")
	}
	return nil
}

func (cv *ConfigValidator) validateMoniker(value string) error {
	if value == "" {
		return fmt.Errorf("moniker cannot be empty")
	}
	validMoniker := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9 _.-]*[a-zA-Z0-9]$`)
	if !validMoniker.MatchString(value) {
		return fmt.Errorf("invalid moniker format")
	}
	return nil
}

func (cv *ConfigValidator) validateNetworkAddress(value string) error {
	if value == "" {
		return nil // Empty is allowed for optional addresses
	}

	// Support tcp:// and unix:// prefixes
	if strings.HasPrefix(value, "tcp://") {
		value = strings.TrimPrefix(value, "tcp://")
	} else if strings.HasPrefix(value, "unix://") {
		// Unix socket paths are allowed
		return nil
	}

	// Validate host:port format
	patterns := []string{
		`^([0-9]{1,3}\.){3}[0-9]{1,3}:[0-9]{1,5}$`,         // IPv4:port
		`^[a-zA-Z0-9][a-zA-Z0-9.-]*:[0-9]{1,5}$`,           // hostname:port
		`^\[[0-9a-fA-F:]+\]:[0-9]{1,5}$`,                   // [IPv6]:port
		`^(0\.0\.0\.0|localhost|127\.0\.0\.1):[0-9]{1,5}$`, // Local addresses
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, value); matched {
			return nil
		}
	}

	return fmt.Errorf("invalid network address format")
}

func (cv *ConfigValidator) validatePeerList(value string) error {
	if value == "" {
		return nil // Empty peer list is allowed
	}

	// Split by comma
	peers := strings.Split(value, ",")
	for _, peer := range peers {
		peer = strings.TrimSpace(peer)
		if peer == "" {
			continue
		}

		// Peer format: node_id@host:port
		parts := strings.Split(peer, "@")
		if len(parts) != 2 {
			return fmt.Errorf("invalid peer format: %s", peer)
		}

		// Validate node ID (hex string)
		nodeID := parts[0]
		if !regexp.MustCompile(`^[0-9a-fA-F]{40}$`).MatchString(nodeID) {
			return fmt.Errorf("invalid node ID: %s", nodeID)
		}

		// Validate address
		if err := cv.validateNetworkAddress(parts[1]); err != nil {
			return fmt.Errorf("invalid peer address: %w", err)
		}
	}

	return nil
}

func (cv *ConfigValidator) validateBoolean(value string) error {
	lower := strings.ToLower(value)
	if lower != "true" && lower != "false" {
		return fmt.Errorf("value must be 'true' or 'false'")
	}
	return nil
}

func (cv *ConfigValidator) validateLogLevel(value string) error {
	validLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	lower := strings.ToLower(value)
	for _, level := range validLevels {
		if lower == level {
			return nil
		}
	}
	return fmt.Errorf("invalid log level, must be one of: %s", strings.Join(validLevels, ", "))
}

func (cv *ConfigValidator) validateLogFormat(value string) error {
	validFormats := []string{"json", "plain"}
	lower := strings.ToLower(value)
	for _, format := range validFormats {
		if lower == format {
			return nil
		}
	}
	return fmt.Errorf("invalid log format, must be one of: %s", strings.Join(validFormats, ", "))
}

func (cv *ConfigValidator) validateDuration(value string) error {
	// Duration format: number + unit (s, ms, us, ns, m, h)
	durationPattern := regexp.MustCompile(`^[0-9]+(\.[0-9]+)?(ns|us|ms|s|m|h)$`)
	if !durationPattern.MatchString(value) {
		return fmt.Errorf("invalid duration format, examples: 1s, 500ms, 1m")
	}
	return nil
}

// GetAllowedKeys returns a list of all allowed configuration keys
func (cv *ConfigValidator) GetAllowedKeys() []string {
	keys := make([]string, 0, len(cv.allowedKeys))
	for key := range cv.allowedKeys {
		keys = append(keys, key)
	}
	return keys
}

// GetKeyDescription returns the description for a configuration key
func (cv *ConfigValidator) GetKeyDescription(key string) string {
	if validator, exists := cv.allowedKeys[key]; exists {
		return validator.Description
	}
	return ""
}
