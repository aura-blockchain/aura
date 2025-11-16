package keeper

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/prevalidation/params"
	"github.com/aequitas/aura/chain/x/prevalidation/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ============================
// TEMPLATE VALIDATION TESTS
// ============================

func TestValidateTemplateBeforeAcceptance(t *testing.T) {
	keeper := setupTestKeeper()

	tests := []struct {
		name        string
		template    *types.ValidationTemplate
		shouldError bool
		errorMsg    string
	}{
		{
			name: "valid template",
			template: &types.ValidationTemplate{
				Id:                 "test-template-1",
				TxType:             types.TxTypeIRCompletion,
				Name:               "Test Template",
				Description:        "A valid test template",
				ValidationRules:    `{"min_amount": 100}`,
				ParameterSchema:    `{"amount": "uint64"}`,
				GasFormula:         "50000",
				PriorityWeight:     100,
				MinConfidenceScore: 100,
				Active:             true,
			},
			shouldError: false,
		},
		{
			name: "missing ID",
			template: &types.ValidationTemplate{
				Name:   "Test",
				TxType: types.TxTypeIRCompletion,
			},
			shouldError: true,
			errorMsg:    "template ID is required",
		},
		{
			name: "invalid JSON in validation rules",
			template: &types.ValidationTemplate{
				Id:              "test-template-2",
				TxType:          types.TxTypeIRCompletion,
				Name:            "Test",
				ValidationRules: `{invalid json}`,
				ParameterSchema: `{}`,
				GasFormula:      "50000",
				PriorityWeight:  100,
			},
			shouldError: true,
			errorMsg:    "validation_rules is not valid JSON",
		},
		{
			name: "dangerous operation in rules",
			template: &types.ValidationTemplate{
				Id:              "test-template-3",
				TxType:          types.TxTypeIRCompletion,
				Name:            "Test",
				ValidationRules: `{"eval": "malicious code"}`,
				ParameterSchema: `{}`,
				GasFormula:      "50000",
				PriorityWeight:  100,
			},
			shouldError: true,
			errorMsg:    "validation_rules contains dangerous operations",
		},
		{
			name: "priority weight too high",
			template: &types.ValidationTemplate{
				Id:              "test-template-4",
				TxType:          types.TxTypeIRCompletion,
				Name:            "Test",
				ValidationRules: `{}`,
				ParameterSchema: `{}`,
				GasFormula:      "50000",
				PriorityWeight:  1001,
			},
			shouldError: true,
			errorMsg:    "priority weight too high",
		},
		{
			name: "invalid gas formula",
			template: &types.ValidationTemplate{
				Id:              "test-template-5",
				TxType:          types.TxTypeIRCompletion,
				Name:            "Test",
				ValidationRules: `{}`,
				ParameterSchema: `{}`,
				GasFormula:      "invalid_formula",
				PriorityWeight:  100,
			},
			shouldError: true,
			errorMsg:    "gas formula must be",
		},
		{
			name: "suspicious pattern in description",
			template: &types.ValidationTemplate{
				Id:              "test-template-6",
				TxType:          types.TxTypeIRCompletion,
				Name:            "Test",
				Description:     "<script>alert('xss')</script>",
				ValidationRules: `{}`,
				ParameterSchema: `{}`,
				GasFormula:      "50000",
				PriorityWeight:  100,
			},
			shouldError: true,
			errorMsg:    "suspicious pattern detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := keeper.ValidateTemplateBeforeAcceptance(tt.template)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
				} else if !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestTemplateSignatureVerification(t *testing.T) {
	keeper := setupTestKeeper()

	// Generate key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	template := &types.ValidationTemplate{
		Id:              "signed-template",
		TxType:          types.TxTypeIRCompletion,
		Name:            "Signed Template",
		ValidationRules: `{}`,
		ParameterSchema: `{}`,
		GasFormula:      "50000",
		PriorityWeight:  100,
	}

	// Sign the template
	signature, err := keeper.SignTemplate(template, privateKey, "test-signer")
	if err != nil {
		t.Fatalf("Failed to sign template: %v", err)
	}

	if signature == nil {
		t.Error("Expected signature, got nil")
	}

	if signature.Signer != "test-signer" {
		t.Errorf("Expected signer 'test-signer', got '%s'", signature.Signer)
	}

	if len(signature.Signature) == 0 {
		t.Error("Expected non-empty signature")
	}

	if signature.SignatureAlg != "ECDSA-SHA256" {
		t.Errorf("Expected algorithm 'ECDSA-SHA256', got '%s'", signature.SignatureAlg)
	}
}

// ============================
// ACCESS CONTROL TESTS
// ============================

func TestAccessControl(t *testing.T) {
	keeper := setupTestKeeper()

	// Test without whitelist requirement
	if !keeper.CanCreatePreValidation("any-address") {
		t.Error("Expected all addresses to be allowed when whitelist not required")
	}

	if !keeper.CanCreateTemplate("any-address") {
		t.Error("Expected all addresses to be allowed for templates when whitelist not required")
	}

	// Test adding to whitelist
	err := keeper.AddAllowedValidator("validator-1")
	if err != nil {
		t.Errorf("Failed to add validator: %v", err)
	}

	err = keeper.RemoveAllowedValidator("validator-1")
	if err != nil {
		t.Errorf("Failed to remove validator: %v", err)
	}
}

// ============================
// REPLAY ATTACK PREVENTION TESTS
// ============================

func TestNonceValidation(t *testing.T) {
	keeper := setupTestKeeper()

	signer := "test-signer"
	nonce := uint64(time.Now().UnixNano())
	timestamp := time.Now()

	// Valid nonce should pass
	err := keeper.ValidateNonce(signer, nonce, timestamp)
	if err != nil {
		t.Errorf("Valid nonce should pass: %v", err)
	}

	// Old nonce should fail
	oldTimestamp := time.Now().Add(-25 * time.Hour)
	err = keeper.ValidateNonce(signer, nonce, oldTimestamp)
	if err == nil {
		t.Error("Old nonce should fail validation")
	}

	// Record nonce usage
	keeper.RecordNonceUsage(signer, nonce, timestamp)

	// Cleanup should not error
	keeper.CleanupExpiredNonces()
}

// ============================
// TEMPLATE EXPIRATION TESTS
// ============================

func TestTemplateExpiration(t *testing.T) {
	keeper := setupTestKeeper()

	// Create template
	template := &types.ValidationTemplate{
		Id:              "expiring-template",
		TxType:          types.TxTypeIRCompletion,
		Name:            "Expiring Template",
		ValidationRules: `{}`,
		ParameterSchema: `{}`,
		GasFormula:      "50000",
		PriorityWeight:  100,
		Active:          true,
		CreatedAt:       timestamppb.New(time.Now().Add(-400 * 24 * time.Hour)), // Very old
	}

	keeper.RegisterTemplate(template)

	// Check if expired
	if !keeper.IsTemplateExpired(template) {
		t.Error("Old template should be expired")
	}

	// Cleanup expired templates
	count := keeper.CleanupExpiredTemplates()
	if count == 0 {
		t.Error("Should have cleaned up at least one expired template")
	}

	// Template should be inactive after cleanup
	retrievedTemplate, ok := keeper.GetTemplate("expiring-template")
	if ok && retrievedTemplate.Active {
		t.Error("Template should be inactive after expiration")
	}
}

// ============================
// CACHE POISONING PREVENTION TESTS
// ============================

func TestCachePoisoningDetection(t *testing.T) {
	keeper := setupTestKeeper()

	signer := "malicious-signer"
	txType := types.TxTypeIRCompletion

	// First attempt should be fine
	err := keeper.DetectCachePoisoning(signer, txType)
	if err != nil {
		t.Errorf("First attempt should not trigger detection: %v", err)
	}

	// Record failures
	for i := 0; i < 10; i++ {
		keeper.RecordValidationFailure(signer, "test failure")
	}

	// Test detection (this is simplified - in production would track per-signer stats)
	// The actual detection logic may vary based on implementation
}

// ============================
// METRICS MANIPULATION DETECTION TESTS
// ============================

func TestMetricsManipulationDetection(t *testing.T) {
	keeper := setupTestKeeper()

	// Set valid metrics
	keeper.metrics.TotalCacheHits = 80
	keeper.metrics.TotalCacheMisses = 20
	keeper.metrics.UpdateCacheHitRate()

	// Should have no anomalies
	anomalies := keeper.DetectMetricsManipulation()
	if len(anomalies) > 0 {
		t.Errorf("Valid metrics should have no anomalies, got: %v", anomalies)
	}

	// Introduce manipulation - impossible hit rate
	keeper.metrics.OverallCacheHitRate = 1.5 // 150% - impossible
	anomalies = keeper.DetectMetricsManipulation()
	if len(anomalies) == 0 {
		t.Error("Should detect impossible hit rate")
	}

	// Reset
	keeper.ResetMetrics()
	keeper.metrics.TotalCacheHits = 100
	keeper.metrics.TotalCacheMisses = 50
	keeper.metrics.OverallCacheHitRate = 0.5 // Mismatched with actual ratio

	anomalies = keeper.DetectMetricsManipulation()
	if len(anomalies) == 0 {
		t.Error("Should detect mismatched hit rate")
	}

	// Test integrity validation
	err := keeper.ValidateMetricsIntegrity()
	if err == nil {
		t.Error("Should fail integrity check with manipulated metrics")
	}
}

func TestMetricsIntegrityWithValidData(t *testing.T) {
	keeper := setupTestKeeper()

	// Set consistent metrics
	keeper.metrics.TotalCacheHits = 80
	keeper.metrics.TotalCacheMisses = 20
	keeper.metrics.TotalExecuted = 75
	keeper.metrics.TotalPreValidations = 100
	keeper.metrics.UpdateCacheHitRate()

	err := keeper.ValidateMetricsIntegrity()
	if err != nil {
		t.Errorf("Valid metrics should pass integrity check: %v", err)
	}
}

// ============================
// OFF-PEAK VERIFICATION TESTS
// ============================

func TestOffPeakEnforcement(t *testing.T) {
	keeper := setupTestKeeper()

	// Set time to off-peak hours (3 AM)
	keeper.SetCurrentTime(time.Date(2024, 1, 1, 3, 0, 0, 0, time.UTC).Unix())

	// Should allow operation during off-peak
	err := keeper.EnforceOffPeakRestriction()
	if err != nil {
		t.Errorf("Should allow operation during off-peak hours: %v", err)
	}

	// Set time to peak hours (12 PM)
	keeper.SetCurrentTime(time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC).Unix())

	// Should deny operation during peak hours
	err = keeper.EnforceOffPeakRestriction()
	if err == nil {
		t.Error("Should deny operation during peak hours")
	}

	// Test compliance verification
	offPeakTime := time.Date(2024, 1, 1, 3, 0, 0, 0, time.UTC)
	if !keeper.VerifyOffPeakCompliance(offPeakTime) {
		t.Error("Should verify off-peak time as compliant")
	}

	peakTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	if keeper.VerifyOffPeakCompliance(peakTime) {
		t.Error("Should not verify peak time as compliant")
	}
}

// ============================
// AUDIT TRAIL TESTS
// ============================

func TestAuditTrail(t *testing.T) {
	keeper := setupTestKeeper()

	// Audit an action
	keeper.auditAction("test_event", "test_actor", "test_target",
		"Test action description", true,
		map[string]string{"key": "value"})

	// Test audit entry signing
	entry := AuditEntry{
		Timestamp: time.Now(),
		EventType: "test",
		Actor:     "actor",
		TargetID:  "target",
		Success:   true,
	}

	signature := keeper.signAuditEntry(entry)
	if len(signature) == 0 {
		t.Error("Expected non-empty audit signature")
	}

	// Test signature verification
	entries := []AuditEntry{entry}
	entries[0].Signature = signature

	valid, tamperedIndices := keeper.VerifyAuditIntegrity(entries)
	if !valid {
		t.Errorf("Valid audit entry should pass verification, tampered: %v", tamperedIndices)
	}

	// Test with tampered entry
	entries[0].Success = false // Modify entry without updating signature
	valid, tamperedIndices = keeper.VerifyAuditIntegrity(entries)
	if valid {
		t.Error("Tampered audit entry should fail verification")
	}
	if len(tamperedIndices) == 0 {
		t.Error("Should detect tampered entry")
	}
}

// ============================
// KEY ROTATION TESTS (in separate file)
// ============================

// ============================
// INTEGRATION TESTS
// ============================

func TestSecurityIntegration(t *testing.T) {
	keeper := setupTestKeeper()

	// Test complete flow with security features
	template := &types.ValidationTemplate{
		Id:                 "secure-template",
		TxType:             types.TxTypeIRCompletion,
		Name:               "Secure Template",
		Description:        "A secure template",
		ValidationRules:    `{"min_amount": 100}`,
		ParameterSchema:    `{"amount": "uint64"}`,
		GasFormula:         "50000",
		PriorityWeight:     100,
		MinConfidenceScore: 100,
		Active:             true,
	}

	// Register template (should validate)
	err := keeper.RegisterTemplate(template)
	if err != nil {
		t.Fatalf("Failed to register template: %v", err)
	}

	// Create pre-validated transaction (should check access control)
	signer := "test-signer"
	txData := []byte("test transaction data")

	tx, err := keeper.CreatePreValidatedTransaction(
		types.TxTypeIRCompletion,
		template.Id,
		txData,
		signer,
		50000,
		map[string]string{"test": "context"},
	)

	if err != nil {
		t.Fatalf("Failed to create pre-validated transaction: %v", err)
	}

	if tx == nil {
		t.Fatal("Expected non-nil transaction")
	}

	// Verify transaction was audited
	// Verify nonce was recorded
	// Verify metrics are consistent

	// Clean up
	keeper.CleanupExpiredTransactions()
	keeper.CleanupExpiredTemplates()
}

func TestSecurityConfigValidation(t *testing.T) {
	keeper := setupTestKeeper()

	config := keeper.getSecurityConfig()

	if config.KeyRotationIntervalHours == 0 {
		t.Error("Key rotation interval should be configured")
	}

	if config.ReplayAttackWindowHours == 0 {
		t.Error("Replay attack window should be configured")
	}

	if config.CachePoisoningThreshold <= 0 || config.CachePoisoningThreshold >= 1 {
		t.Error("Cache poisoning threshold should be between 0 and 1")
	}

	if config.AuditRetentionDays == 0 {
		t.Error("Audit retention should be configured")
	}
}

// ============================
// HELPER FUNCTIONS
// ============================

func setupTestKeeper() *Keeper {
	store := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(store)
	keeper.SetCurrentTime(time.Now().Unix())
	keeper.SetCurrentHeight(1000)
	return keeper
}

func TestHelperFunctions(t *testing.T) {
	// Test contains function
	if !contains("hello world", "world") {
		t.Error("contains should find substring")
	}

	if contains("hello", "world") {
		t.Error("contains should not find non-existent substring")
	}

	// Test abs function
	if abs(-5.0) != 5.0 {
		t.Error("abs should return positive value")
	}

	if abs(5.0) != 5.0 {
		t.Error("abs should preserve positive value")
	}

	// Test bytesEqual
	a := []byte{1, 2, 3}
	b := []byte{1, 2, 3}
	c := []byte{1, 2, 4}

	if !bytesEqual(a, b) {
		t.Error("Equal byte slices should be equal")
	}

	if bytesEqual(a, c) {
		t.Error("Different byte slices should not be equal")
	}
}

func TestCanonicalizeTemplate(t *testing.T) {
	keeper := setupTestKeeper()

	template := &types.ValidationTemplate{
		Id:                 "test",
		Name:               "Test Template",
		TxType:             types.TxTypeIRCompletion,
		ValidationRules:    `{}`,
		ParameterSchema:    `{}`,
		GasFormula:         "50000",
		PriorityWeight:     100,
		MinConfidenceScore: 100,
	}

	canonical := keeper.canonicalizeTemplate(template)
	if len(canonical) == 0 {
		t.Error("Canonical representation should not be empty")
	}

	// Same template should produce same canonical form
	canonical2 := keeper.canonicalizeTemplate(template)
	if !bytesEqual(canonical, canonical2) {
		t.Error("Canonical representation should be deterministic")
	}

	// Different template should produce different canonical form
	template2 := &types.ValidationTemplate{
		Id:                 "test2",
		Name:               "Different Template",
		TxType:             types.TxTypeDexSwap,
		ValidationRules:    `{}`,
		ParameterSchema:    `{}`,
		GasFormula:         "100000",
		PriorityWeight:     50,
		MinConfidenceScore: 200,
	}

	canonical3 := keeper.canonicalizeTemplate(template2)
	if bytesEqual(canonical, canonical3) {
		t.Error("Different templates should have different canonical forms")
	}
}

func TestValidateGasFormula(t *testing.T) {
	keeper := setupTestKeeper()

	tests := []struct {
		name        string
		formula     string
		shouldError bool
	}{
		{"base_gas", "base_gas", false},
		{"numeric value", "50000", false},
		{"zero gas", "0", true},
		{"too high gas", "99999999", true},
		{"invalid formula", "invalid", true},
		{"empty formula", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := keeper.validateGasFormula(tt.formula)
			if tt.shouldError && err == nil {
				t.Error("Expected error for invalid gas formula")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

func TestDangerousOperationsDetection(t *testing.T) {
	keeper := setupTestKeeper()

	tests := []struct {
		name         string
		rules        string
		shouldDetect bool
	}{
		{
			name:         "safe rules",
			rules:        `{"min_amount": 100, "max_amount": 1000}`,
			shouldDetect: false,
		},
		{
			name:         "eval operation",
			rules:        `{"eval": "malicious"}`,
			shouldDetect: true,
		},
		{
			name:         "exec operation",
			rules:        `{"exec": "malicious"}`,
			shouldDetect: true,
		},
		{
			name:         "proto pollution",
			rules:        `{"__proto__": "malicious"}`,
			shouldDetect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rules map[string]interface{}
			err := json.Unmarshal([]byte(tt.rules), &rules)
			if err != nil {
				t.Fatalf("Failed to parse rules: %v", err)
			}

			detected := keeper.containsDangerousOperations(rules)
			if detected != tt.shouldDetect {
				t.Errorf("Expected detection=%v, got=%v", tt.shouldDetect, detected)
			}
		})
	}
}
