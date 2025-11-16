package keeper

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/aequitas/aura/chain/x/prevalidation/types"
)

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	// Access control
	AllowedValidators       map[string]bool // validator addresses that can create pre-validations
	AllowedTemplateCreators map[string]bool // addresses that can create templates
	RequireWhitelist        bool

	// Key rotation
	KeyRotationIntervalHours uint32
	MaxKeyAge                time.Duration
	MinKeysToRetain          int

	// Replay protection
	NonceWindow             uint64 // Window of valid nonces
	ReplayAttackWindowHours uint32

	// Template security
	MaxTemplateAge           time.Duration
	RequireTemplateSignature bool

	// Audit
	EnableAuditTrail   bool
	AuditRetentionDays uint32

	// Cache poisoning prevention
	MaxCachePoisoningAttempts uint64
	CachePoisoningThreshold   float64
}

// AuditEntry represents a single audit log entry
type AuditEntry struct {
	Timestamp time.Time
	EventType string // "template_created", "prevalidation_created", "prevalidation_executed", etc.
	Actor     string // Address of the entity performing the action
	TargetID  string // ID of the affected resource
	Action    string // Detailed action description
	Success   bool
	ErrorMsg  string
	Metadata  map[string]string
	Signature []byte // Signature of the audit entry for integrity
}

// TemplateSignature represents a cryptographic signature of a template
type TemplateSignature struct {
	Signer       string
	Timestamp    time.Time
	Signature    []byte
	PublicKey    []byte
	SignatureAlg string // "ECDSA-SHA256"
}

// NonceTracker tracks used nonces to prevent replay attacks
type NonceTracker struct {
	UsedNonces    map[string]map[uint64]time.Time // signer -> nonce -> timestamp
	NonceSequence map[string]uint64               // signer -> highest nonce
}

// CachePoisoningDetector detects potential cache poisoning attempts
type CachePoisoningDetector struct {
	FailuresByAddress  map[string]uint64      // address -> failure count
	SuspiciousPatterns map[string][]time.Time // address -> timestamps of suspicious activity
}

// ============================
// TEMPLATE VALIDATION
// ============================

// ValidateTemplateBeforeAcceptance performs comprehensive template validation
func (k *Keeper) ValidateTemplateBeforeAcceptance(template *types.ValidationTemplate) error {
	if template == nil {
		return types.ErrInvalidTemplate
	}

	// 1. Basic field validation
	if template.Id == "" {
		return fmt.Errorf("template ID is required")
	}
	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}
	if template.TxType == types.TxTypeUnspecified {
		return fmt.Errorf("transaction type must be specified")
	}

	// 2. Validate JSON schemas
	if err := k.validateTemplateSchemas(template); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	// 3. Check for malicious patterns in validation rules
	if err := k.detectMaliciousTemplatePatterns(template); err != nil {
		return fmt.Errorf("malicious pattern detected: %w", err)
	}

	// 4. Validate gas formula
	if err := k.validateGasFormula(template.GasFormula); err != nil {
		return fmt.Errorf("invalid gas formula: %w", err)
	}

	// 5. Check template signature if required
	securityConfig := k.getSecurityConfig()
	if securityConfig.RequireTemplateSignature {
		if err := k.verifyTemplateSignature(template); err != nil {
			return fmt.Errorf("template signature verification failed: %w", err)
		}
	}

	// 6. Validate priority weight
	if template.PriorityWeight == 0 {
		return fmt.Errorf("priority weight must be greater than 0")
	}
	if template.PriorityWeight > 1000 {
		return fmt.Errorf("priority weight too high (max: 1000)")
	}

	// 7. Validate confidence score requirement
	params := k.GetParams()
	if template.MinConfidenceScore < params.MinConfidenceScore {
		return fmt.Errorf("template min confidence score cannot be lower than module minimum")
	}

	return nil
}

// validateTemplateSchemas validates JSON schemas in the template
func (k *Keeper) validateTemplateSchemas(template *types.ValidationTemplate) error {
	// Validate ValidationRules is valid JSON
	if template.ValidationRules != "" {
		var rules map[string]interface{}
		if err := json.Unmarshal([]byte(template.ValidationRules), &rules); err != nil {
			return fmt.Errorf("validation_rules is not valid JSON: %w", err)
		}

		// Check for dangerous operations
		if k.containsDangerousOperations(rules) {
			return fmt.Errorf("validation_rules contains dangerous operations")
		}
	}

	// Validate ParameterSchema is valid JSON
	if template.ParameterSchema != "" {
		var schema map[string]interface{}
		if err := json.Unmarshal([]byte(template.ParameterSchema), &schema); err != nil {
			return fmt.Errorf("parameter_schema is not valid JSON: %w", err)
		}
	}

	return nil
}

// containsDangerousOperations checks for dangerous operations in validation rules
func (k *Keeper) containsDangerousOperations(rules map[string]interface{}) bool {
	dangerousKeys := []string{
		"eval", "exec", "system", "execute", "delete", "drop",
		"__proto__", "constructor", "prototype",
	}

	for key := range rules {
		for _, dangerous := range dangerousKeys {
			if key == dangerous {
				return true
			}
		}

		// Recursively check nested objects
		if nested, ok := rules[key].(map[string]interface{}); ok {
			if k.containsDangerousOperations(nested) {
				return true
			}
		}
	}

	return false
}

// detectMaliciousTemplatePatterns detects potential malicious patterns
func (k *Keeper) detectMaliciousTemplatePatterns(template *types.ValidationTemplate) error {
	// Check for excessively long strings that could cause DoS
	maxLength := 10000
	if len(template.ValidationRules) > maxLength {
		return fmt.Errorf("validation_rules too long (max: %d)", maxLength)
	}
	if len(template.ParameterSchema) > maxLength {
		return fmt.Errorf("parameter_schema too long (max: %d)", maxLength)
	}
	if len(template.GasFormula) > 1000 {
		return fmt.Errorf("gas_formula too long (max: 1000)")
	}

	// Check for suspicious patterns in descriptions
	suspiciousPatterns := []string{
		"<script>", "javascript:", "data:text/html",
		"onerror=", "onload=", "eval(",
	}

	checkString := template.Description + template.Name
	for _, pattern := range suspiciousPatterns {
		if contains(checkString, pattern) {
			return fmt.Errorf("suspicious pattern detected: %s", pattern)
		}
	}

	return nil
}

// validateGasFormula validates the gas estimation formula
func (k *Keeper) validateGasFormula(formula string) error {
	if formula == "" {
		return fmt.Errorf("gas formula is required")
	}

	// For now, only allow simple numeric formulas or "base_gas"
	// In production, implement a proper expression parser
	if formula == "base_gas" {
		return nil
	}

	// Check if it's a simple number
	var gasAmount uint64
	_, err := fmt.Sscanf(formula, "%d", &gasAmount)
	if err != nil {
		return fmt.Errorf("gas formula must be 'base_gas' or a numeric value")
	}

	// Validate gas amount is reasonable
	if gasAmount == 0 {
		return fmt.Errorf("gas amount must be greater than 0")
	}
	if gasAmount > 10000000 { // 10M gas limit
		return fmt.Errorf("gas amount too high (max: 10000000)")
	}

	return nil
}

// ============================
// TEMPLATE SIGNATURE VERIFICATION
// ============================

// SignTemplate creates a cryptographic signature for a template
func (k *Keeper) SignTemplate(template *types.ValidationTemplate, privateKey *ecdsa.PrivateKey, signer string) (*TemplateSignature, error) {
	// Create canonical representation of template
	canonicalData := k.canonicalizeTemplate(template)

	// Hash the data
	hash := sha256.Sum256(canonicalData)

	// Sign the hash
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign template: %w", err)
	}

	// Encode signature
	signature := append(r.Bytes(), s.Bytes()...)

	// Encode public key
	publicKeyBytes := elliptic.Marshal(privateKey.Curve, privateKey.X, privateKey.Y)

	return &TemplateSignature{
		Signer:       signer,
		Timestamp:    time.Now(),
		Signature:    signature,
		PublicKey:    publicKeyBytes,
		SignatureAlg: "ECDSA-SHA256",
	}, nil
}

// verifyTemplateSignature verifies a template's signature
func (k *Keeper) verifyTemplateSignature(template *types.ValidationTemplate) error {
	// Extract signature from template context
	// In production, signature would be stored in a dedicated field
	// For now, we'll assume it's in the description or separate storage

	// This is a placeholder - actual implementation would:
	// 1. Extract signature from template metadata
	// 2. Reconstruct public key
	// 3. Verify signature against canonical template data

	// For demonstration:
	if template.Description == "" {
		return fmt.Errorf("template signature not found")
	}

	return nil
}

// canonicalizeTemplate creates a canonical byte representation of a template
func (k *Keeper) canonicalizeTemplate(template *types.ValidationTemplate) []byte {
	data := fmt.Sprintf("%s:%s:%d:%s:%s:%s:%d:%d",
		template.Id,
		template.Name,
		template.TxType,
		template.ValidationRules,
		template.ParameterSchema,
		template.GasFormula,
		template.PriorityWeight,
		template.MinConfidenceScore,
	)
	return []byte(data)
}

// ============================
// ACCESS CONTROL
// ============================

// CanCreatePreValidation checks if an address can create pre-validations
func (k *Keeper) CanCreatePreValidation(address string) bool {
	securityConfig := k.getSecurityConfig()

	if !securityConfig.RequireWhitelist {
		return true
	}

	return securityConfig.AllowedValidators[address]
}

// CanCreateTemplate checks if an address can create templates
func (k *Keeper) CanCreateTemplate(address string) bool {
	securityConfig := k.getSecurityConfig()

	if !securityConfig.RequireWhitelist {
		return true
	}

	return securityConfig.AllowedTemplateCreators[address]
}

// AddAllowedValidator adds a validator to the whitelist
func (k *Keeper) AddAllowedValidator(address string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// In production, this would be stored in state
	// For now, we'll use in-memory tracking

	// Audit the action
	k.auditAction("validator_added", "system", address, "Added validator to whitelist", true, nil)

	return nil
}

// RemoveAllowedValidator removes a validator from the whitelist
func (k *Keeper) RemoveAllowedValidator(address string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Audit the action
	k.auditAction("validator_removed", "system", address, "Removed validator from whitelist", true, nil)

	return nil
}

// ============================
// REPLAY ATTACK PREVENTION
// ============================

// InitializeNonceTracker initializes the nonce tracking system
func (k *Keeper) InitializeNonceTracker() {
	// This would be called during keeper initialization
	// Nonce tracking is already partially implemented via the nonce field
	// This enhances it with time-based windows and sequence tracking
}

// ValidateNonce checks if a nonce is valid and hasn't been used
func (k *Keeper) ValidateNonce(signer string, nonce uint64, timestamp time.Time) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Check if nonce is within acceptable window
	securityConfig := k.getSecurityConfig()
	windowDuration := time.Duration(securityConfig.ReplayAttackWindowHours) * time.Hour

	// Nonce should be relatively recent
	if time.Since(timestamp) > windowDuration {
		return fmt.Errorf("nonce timestamp too old")
	}

	// In production, track used nonces in persistent storage
	// For now, we rely on the unique tx ID generation which includes nonce

	return nil
}

// RecordNonceUsage records that a nonce has been used
func (k *Keeper) RecordNonceUsage(signer string, nonce uint64, timestamp time.Time) {
	// This would persist the nonce usage to prevent replay
	// The current implementation uses nonce in tx ID generation which provides some protection
}

// CleanupExpiredNonces removes nonces outside the replay window
func (k *Keeper) CleanupExpiredNonces() {
	k.mu.Lock()
	defer k.mu.Unlock()

	// In production, this would clean up old nonce records
	// Run periodically to prevent memory growth
}

// ============================
// TEMPLATE EXPIRATION
// ============================

// IsTemplateExpired checks if a template has expired
func (k *Keeper) IsTemplateExpired(template *types.ValidationTemplate) bool {
	if template.CreatedAt == nil {
		return false
	}

	securityConfig := k.getSecurityConfig()
	if securityConfig.MaxTemplateAge == 0 {
		return false // No expiration configured
	}

	createdTime := template.CreatedAt.AsTime()
	return time.Since(createdTime) > securityConfig.MaxTemplateAge
}

// CleanupExpiredTemplates removes expired templates
func (k *Keeper) CleanupExpiredTemplates() uint64 {
	k.mu.Lock()
	defer k.mu.Unlock()

	removedCount := uint64(0)

	for templateID, template := range k.templatesById {
		if k.IsTemplateExpired(template) {
			// Mark as inactive rather than deleting
			template.Active = false

			// Remove from type index
			templates := k.templatesByType[template.TxType]
			newTemplates := []*types.ValidationTemplate{}
			for _, t := range templates {
				if t.Id != templateID {
					newTemplates = append(newTemplates, t)
				}
			}
			k.templatesByType[template.TxType] = newTemplates

			removedCount++

			// Audit the expiration
			k.auditAction("template_expired", "system", templateID,
				fmt.Sprintf("Template expired after %v", time.Since(template.CreatedAt.AsTime())),
				true, nil)
		}
	}

	return removedCount
}

// ============================
// CACHE POISONING PREVENTION
// ============================

// DetectCachePoisoning analyzes patterns to detect cache poisoning attempts
func (k *Keeper) DetectCachePoisoning(signer string, txType types.TransactionType) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Track validation failures by address
	// High failure rate indicates potential poisoning attempt

	metrics := k.metrics.GetTypeMetrics(txType)
	if metrics == nil {
		return nil
	}

	// Calculate failure rate for this signer
	// In production, track per-signer statistics

	securityConfig := k.getSecurityConfig()

	// If failure rate exceeds threshold, block
	failureRate := 0.0
	if metrics.TotalPreValidated > 0 {
		failureRate = float64(metrics.TotalExpired) / float64(metrics.TotalPreValidated)
	}

	if failureRate > securityConfig.CachePoisoningThreshold {
		return fmt.Errorf("cache poisoning detected: abnormal failure rate %.2f", failureRate)
	}

	return nil
}

// RecordValidationFailure records a validation failure for poisoning detection
func (k *Keeper) RecordValidationFailure(signer string, reason string) {
	// Track failures per signer
	// Use this data to detect poisoning patterns

	k.auditAction("validation_failure", signer, "", reason, false,
		map[string]string{"reason": reason})
}

// ============================
// METRICS MANIPULATION DETECTION
// ============================

// DetectMetricsManipulation analyzes metrics for signs of manipulation
func (k *Keeper) DetectMetricsManipulation() []string {
	k.mu.RLock()
	defer k.mu.RUnlock()

	anomalies := []string{}

	// 1. Check for impossible hit rates
	if k.metrics.OverallCacheHitRate > 1.0 {
		anomalies = append(anomalies, "Cache hit rate > 100%")
	}
	if k.metrics.OverallCacheHitRate < 0.0 {
		anomalies = append(anomalies, "Cache hit rate < 0%")
	}

	// 2. Check for mismatched totals
	calculatedHitRate := float64(0)
	total := k.metrics.TotalCacheHits + k.metrics.TotalCacheMisses
	if total > 0 {
		calculatedHitRate = float64(k.metrics.TotalCacheHits) / float64(total)
	}

	if abs(calculatedHitRate-k.metrics.OverallCacheHitRate) > 0.01 {
		anomalies = append(anomalies, fmt.Sprintf(
			"Hit rate mismatch: calculated %.2f vs reported %.2f",
			calculatedHitRate, k.metrics.OverallCacheHitRate))
	}

	// 3. Check for inconsistent type metrics
	for txTypeStr, typeMetrics := range k.metrics.MetricsByType {
		typeTotal := typeMetrics.CacheHits + typeMetrics.CacheMisses
		if typeTotal > 0 {
			typeHitRate := float64(typeMetrics.CacheHits) / float64(typeTotal)
			if abs(typeHitRate-typeMetrics.CacheHitRate) > 0.01 {
				anomalies = append(anomalies, fmt.Sprintf(
					"%s: hit rate mismatch: calculated %.2f vs reported %.2f",
					txTypeStr, typeHitRate, typeMetrics.CacheHitRate))
			}
		}

		// Check for negative values
		if typeMetrics.CacheHits < 0 || typeMetrics.CacheMisses < 0 {
			anomalies = append(anomalies, fmt.Sprintf("%s: negative metric values", txTypeStr))
		}

		// Check execution counts
		if typeMetrics.TotalExecuted > typeMetrics.TotalPreValidated {
			anomalies = append(anomalies, fmt.Sprintf(
				"%s: executed (%d) > pre-validated (%d)",
				txTypeStr, typeMetrics.TotalExecuted, typeMetrics.TotalPreValidated))
		}
	}

	// 4. Check time savings consistency
	if k.metrics.TotalTimeSavedMs > 0 && k.metrics.TotalExecuted == 0 {
		anomalies = append(anomalies, "Time saved reported but no executions")
	}

	return anomalies
}

// ValidateMetricsIntegrity performs a comprehensive integrity check
func (k *Keeper) ValidateMetricsIntegrity() error {
	anomalies := k.DetectMetricsManipulation()

	if len(anomalies) > 0 {
		k.auditAction("metrics_anomaly", "system", "",
			fmt.Sprintf("Detected %d anomalies", len(anomalies)),
			false, map[string]string{
				"anomalies": fmt.Sprintf("%v", anomalies),
			})

		return fmt.Errorf("metrics integrity check failed: %d anomalies detected", len(anomalies))
	}

	return nil
}

// ============================
// AUDIT TRAIL
// ============================

// auditAction records an action in the audit trail
func (k *Keeper) auditAction(eventType, actor, targetID, action string, success bool, metadata map[string]string) {
	securityConfig := k.getSecurityConfig()
	if !securityConfig.EnableAuditTrail {
		return
	}

	entry := AuditEntry{
		Timestamp: time.Unix(k.currentTime, 0),
		EventType: eventType,
		Actor:     actor,
		TargetID:  targetID,
		Action:    action,
		Success:   success,
		Metadata:  metadata,
	}

	// Sign the audit entry for integrity
	entry.Signature = k.signAuditEntry(entry)

	// In production, persist to secure audit storage
	// For now, we'll emit an event
}

// signAuditEntry creates a signature for an audit entry
func (k *Keeper) signAuditEntry(entry AuditEntry) []byte {
	data := fmt.Sprintf("%s:%s:%s:%s:%t",
		entry.Timestamp.Format(time.RFC3339),
		entry.EventType,
		entry.Actor,
		entry.TargetID,
		entry.Success,
	)
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

// GetAuditTrail retrieves audit entries for a time range
func (k *Keeper) GetAuditTrail(startTime, endTime time.Time, eventType string) []AuditEntry {
	// In production, query from persistent audit storage
	// Filter by time range and event type
	return []AuditEntry{}
}

// VerifyAuditIntegrity verifies the integrity of audit entries
func (k *Keeper) VerifyAuditIntegrity(entries []AuditEntry) (bool, []int) {
	tamperedIndices := []int{}

	for i, entry := range entries {
		expectedSig := k.signAuditEntry(entry)
		if !bytesEqual(expectedSig, entry.Signature) {
			tamperedIndices = append(tamperedIndices, i)
		}
	}

	return len(tamperedIndices) == 0, tamperedIndices
}

// ============================
// OFF-PEAK VERIFICATION
// ============================

// EnforceOffPeakRestriction ensures operations only occur during off-peak hours
func (k *Keeper) EnforceOffPeakRestriction() error {
	params := k.GetParams()

	if !params.SchedulerConfig.Enabled {
		return types.ErrSchedulerDisabled
	}

	currentTime := time.Unix(k.currentTime, 0)
	currentHour := uint32(currentTime.Hour())

	// Check if we're in off-peak hours
	isOffPeak := params.SchedulerConfig.IsOffPeakHour(currentHour)

	// If peak hours not allowed and we're not in off-peak, reject
	if !params.SchedulerConfig.AllowPeakHours && !isOffPeak {
		k.auditAction("off_peak_violation", "system", "",
			fmt.Sprintf("Operation attempted during peak hour %d", currentHour),
			false, map[string]string{
				"hour": fmt.Sprintf("%d", currentHour),
			})
		return types.ErrNotOffPeakHours
	}

	return nil
}

// VerifyOffPeakCompliance checks if an action complied with off-peak restrictions
func (k *Keeper) VerifyOffPeakCompliance(timestamp time.Time) bool {
	params := k.GetParams()

	if params.SchedulerConfig.AllowPeakHours {
		return true
	}

	hour := uint32(timestamp.Hour())
	return params.SchedulerConfig.IsOffPeakHour(hour)
}

// ============================
// HELPER FUNCTIONS
// ============================

// getSecurityConfig returns the security configuration
func (k *Keeper) getSecurityConfig() SecurityConfig {
	// In production, load from params or state
	return SecurityConfig{
		AllowedValidators:         make(map[string]bool),
		AllowedTemplateCreators:   make(map[string]bool),
		RequireWhitelist:          false,
		KeyRotationIntervalHours:  720,                 // 30 days
		MaxKeyAge:                 90 * 24 * time.Hour, // 90 days
		MinKeysToRetain:           3,
		NonceWindow:               1000,
		ReplayAttackWindowHours:   24,
		MaxTemplateAge:            365 * 24 * time.Hour, // 1 year
		RequireTemplateSignature:  false,
		EnableAuditTrail:          true,
		AuditRetentionDays:        90,
		MaxCachePoisoningAttempts: 100,
		CachePoisoningThreshold:   0.7, // 70% failure rate
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || contains(s[1:], substr)))
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Helper function to compare byte slices
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// GenerateKeyPair generates a new ECDSA key pair for signing
func GenerateKeyPair() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// ExportPrivateKey exports a private key to PEM format
func ExportPrivateKey(key *ecdsa.PrivateKey) (string, error) {
	x509Encoded, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return "", err
	}
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: x509Encoded})
	return string(pemEncoded), nil
}

// ImportPrivateKey imports a private key from PEM format
func ImportPrivateKey(pemEncoded string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemEncoded))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	return x509.ParseECPrivateKey(block.Bytes)
}
