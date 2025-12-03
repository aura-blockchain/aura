package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// ============================================================================
// Basic Commitment Generation Tests
// ============================================================================

func TestGenerateCommitment(t *testing.T) {
	service := NewDataProtectionService()

	tests := []struct {
		name     string
		data     interface{}
		wantErr  bool
	}{
		{
			name:    "simple string",
			data:    "sensitive data",
			wantErr: false,
		},
		{
			name:    "struct data",
			data:    struct{ Name string }{"Alice"},
			wantErr: false,
		},
		{
			name:    "map data",
			data:    map[string]string{"ssn": "123-45-6789"},
			wantErr: false,
		},
		{
			name:    "nil data",
			data:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commitment, err := service.GenerateCommitment(tt.data)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, commitment)
			require.Len(t, commitment, 32, "SHA-256 should produce 32 bytes")
		})
	}
}

func TestGenerateCommitmentDeterministic(t *testing.T) {
	service := NewDataProtectionService()

	data := map[string]string{
		"name":  "Alice",
		"email": "alice@example.com",
		"ssn":   "123-45-6789",
	}

	// Generate commitment multiple times
	commitment1, err := service.GenerateCommitment(data)
	require.NoError(t, err)

	commitment2, err := service.GenerateCommitment(data)
	require.NoError(t, err)

	commitment3, err := service.GenerateCommitment(data)
	require.NoError(t, err)

	// All commitments should be identical (deterministic)
	require.Equal(t, commitment1, commitment2)
	require.Equal(t, commitment2, commitment3)
}

func TestGenerateCommitmentDifferentData(t *testing.T) {
	service := NewDataProtectionService()

	data1 := "sensitive data 1"
	data2 := "sensitive data 2"

	commitment1, err := service.GenerateCommitment(data1)
	require.NoError(t, err)

	commitment2, err := service.GenerateCommitment(data2)
	require.NoError(t, err)

	// Different data should produce different commitments
	require.NotEqual(t, commitment1, commitment2)
}

func TestGenerateCommitmentFromBytes(t *testing.T) {
	service := NewDataProtectionService()

	data := []byte("sensitive data")
	commitment := service.GenerateCommitmentFromBytes(data)

	require.NotNil(t, commitment)
	require.Len(t, commitment, 32)

	// Verify it matches expected SHA-256
	expected := sha256.Sum256(data)
	require.Equal(t, expected[:], commitment)
}

// ============================================================================
// Commitment Verification Tests
// ============================================================================

func TestVerifyCommitment_Valid(t *testing.T) {
	service := NewDataProtectionService()

	data := map[string]string{
		"name": "Alice",
		"ssn":  "123-45-6789",
	}

	// Generate commitment
	commitment, err := service.GenerateCommitment(data)
	require.NoError(t, err)

	// Verify with correct data
	valid, err := service.VerifyCommitment(data, commitment)
	require.NoError(t, err)
	require.True(t, valid, "Valid data should verify successfully")
}

func TestVerifyCommitment_Invalid(t *testing.T) {
	service := NewDataProtectionService()

	originalData := map[string]string{
		"name": "Alice",
		"ssn":  "123-45-6789",
	}

	// Generate commitment for original data
	commitment, err := service.GenerateCommitment(originalData)
	require.NoError(t, err)

	// Try to verify with different data
	tampered := map[string]string{
		"name": "Bob",
		"ssn":  "987-65-4321",
	}

	valid, err := service.VerifyCommitment(tampered, commitment)
	require.NoError(t, err)
	require.False(t, valid, "Tampered data should not verify")
}

func TestVerifyCommitment_InvalidLength(t *testing.T) {
	service := NewDataProtectionService()

	data := "test data"
	invalidCommitment := make([]byte, 16) // Wrong length

	valid, err := service.VerifyCommitment(data, invalidCommitment)
	require.Error(t, err)
	require.False(t, valid)
	require.Contains(t, err.Error(), "invalid commitment length")
}

func TestVerifyCommitmentFromBytes_Valid(t *testing.T) {
	service := NewDataProtectionService()

	data := []byte("sensitive data")
	commitment := service.GenerateCommitmentFromBytes(data)

	valid := service.VerifyCommitmentFromBytes(data, commitment)
	require.True(t, valid)
}

func TestVerifyCommitmentFromBytes_Invalid(t *testing.T) {
	service := NewDataProtectionService()

	data := []byte("sensitive data")
	commitment := service.GenerateCommitmentFromBytes(data)

	tamperedData := []byte("tampered data")
	valid := service.VerifyCommitmentFromBytes(tamperedData, commitment)
	require.False(t, valid)
}

// ============================================================================
// Hex Conversion Tests
// ============================================================================

func TestCommitmentToHex(t *testing.T) {
	service := NewDataProtectionService()

	commitment := []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}
	hexStr := service.CommitmentToHex(commitment)

	expected := "0123456789abcdef"
	require.Equal(t, expected, hexStr)
}

func TestHexToCommitment_Valid(t *testing.T) {
	service := NewDataProtectionService()

	// Valid 32-byte commitment as hex
	hexStr := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	commitment, err := service.HexToCommitment(hexStr)

	require.NoError(t, err)
	require.Len(t, commitment, 32)
}

func TestHexToCommitment_InvalidHex(t *testing.T) {
	service := NewDataProtectionService()

	invalidHex := "invalid hex string"
	_, err := service.HexToCommitment(invalidHex)

	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid hex string")
}

func TestHexToCommitment_InvalidLength(t *testing.T) {
	service := NewDataProtectionService()

	// Only 16 bytes instead of 32
	shortHex := "0123456789abcdef0123456789abcdef"
	_, err := service.HexToCommitment(shortHex)

	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid commitment length")
}

func TestHexRoundTrip(t *testing.T) {
	service := NewDataProtectionService()

	original := []byte{
		0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
		0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
		0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
		0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
	}

	hexStr := service.CommitmentToHex(original)
	decoded, err := service.HexToCommitment(hexStr)

	require.NoError(t, err)
	require.Equal(t, original, decoded)
}

// ============================================================================
// Field Commitments Tests
// ============================================================================

func TestGenerateFieldCommitments(t *testing.T) {
	service := NewDataProtectionService()

	fields := map[string]interface{}{
		"ssn":      "123-45-6789",
		"passport": "AB123456",
		"address":  "123 Main St",
	}

	commitments, err := service.GenerateFieldCommitments(fields)
	require.NoError(t, err)
	require.Len(t, commitments, 3)

	// Each commitment should be 32 bytes
	for fieldName, commitment := range commitments {
		require.Len(t, commitment, 32, "Field %s should have 32-byte commitment", fieldName)
	}
}

func TestVerifyFieldCommitments_Valid(t *testing.T) {
	service := NewDataProtectionService()

	fields := map[string]interface{}{
		"ssn":      "123-45-6789",
		"passport": "AB123456",
	}

	// Generate commitments
	commitments, err := service.GenerateFieldCommitments(fields)
	require.NoError(t, err)

	// Verify with correct data
	valid, err := service.VerifyFieldCommitments(fields, commitments)
	require.NoError(t, err)
	require.True(t, valid)
}

func TestVerifyFieldCommitments_Invalid(t *testing.T) {
	service := NewDataProtectionService()

	originalFields := map[string]interface{}{
		"ssn":      "123-45-6789",
		"passport": "AB123456",
	}

	// Generate commitments for original
	commitments, err := service.GenerateFieldCommitments(originalFields)
	require.NoError(t, err)

	// Try to verify with tampered data
	tamperedFields := map[string]interface{}{
		"ssn":      "987-65-4321", // Changed
		"passport": "AB123456",
	}

	valid, err := service.VerifyFieldCommitments(tamperedFields, commitments)
	require.NoError(t, err)
	require.False(t, valid)
}

func TestVerifyFieldCommitments_MissingField(t *testing.T) {
	service := NewDataProtectionService()

	fields := map[string]interface{}{
		"ssn":      "123-45-6789",
		"passport": "AB123456",
	}

	commitments, err := service.GenerateFieldCommitments(fields)
	require.NoError(t, err)

	// Try to verify with missing field
	incompleteFields := map[string]interface{}{
		"ssn": "123-45-6789",
		// passport missing
	}

	valid, err := service.VerifyFieldCommitments(incompleteFields, commitments)
	require.Error(t, err)
	require.False(t, valid)
	require.Contains(t, err.Error(), "field count mismatch")
}

// ============================================================================
// PII Data Tests
// ============================================================================

func TestGeneratePIICommitment(t *testing.T) {
	service := NewDataProtectionService()

	pii := &PIIData{
		FullName:       "Alice Smith",
		DateOfBirth:    "1990-01-01",
		SSN:            "123-45-6789",
		PassportNumber: "AB123456",
		Addresses:      []string{"123 Main St"},
		PhoneNumbers:   []string{"+1-555-0123"},
		EmailAddresses: []string{"alice@example.com"},
	}

	commitment, err := service.GeneratePIICommitment(pii)
	require.NoError(t, err)
	require.NotNil(t, commitment)
	require.Len(t, commitment, 32)
}

func TestGeneratePIICommitment_Nil(t *testing.T) {
	service := NewDataProtectionService()

	_, err := service.GeneratePIICommitment(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot be nil")
}

func TestVerifyPIICommitment_Valid(t *testing.T) {
	service := NewDataProtectionService()

	pii := &PIIData{
		FullName:    "Alice Smith",
		DateOfBirth: "1990-01-01",
		SSN:         "123-45-6789",
	}

	// Generate commitment
	commitment, err := service.GeneratePIICommitment(pii)
	require.NoError(t, err)

	// Verify with correct data
	valid, err := service.VerifyPIICommitment(pii, commitment)
	require.NoError(t, err)
	require.True(t, valid)
}

func TestVerifyPIICommitment_Invalid(t *testing.T) {
	service := NewDataProtectionService()

	originalPII := &PIIData{
		FullName:    "Alice Smith",
		DateOfBirth: "1990-01-01",
		SSN:         "123-45-6789",
	}

	// Generate commitment
	commitment, err := service.GeneratePIICommitment(originalPII)
	require.NoError(t, err)

	// Try to verify with different data
	tamperedPII := &PIIData{
		FullName:    "Bob Jones",
		DateOfBirth: "1985-05-05",
		SSN:         "987-65-4321",
	}

	valid, err := service.VerifyPIICommitment(tamperedPII, commitment)
	require.NoError(t, err)
	require.False(t, valid)
}

func TestPIICommitment_Deterministic(t *testing.T) {
	service := NewDataProtectionService()

	pii := &PIIData{
		FullName:       "Alice Smith",
		DateOfBirth:    "1990-01-01",
		SSN:            "123-45-6789",
		SourceOfFunds:  []string{"employment", "investments"},
		Occupation:     "Software Engineer",
	}

	// Generate multiple commitments
	commitment1, err := service.GeneratePIICommitment(pii)
	require.NoError(t, err)

	commitment2, err := service.GeneratePIICommitment(pii)
	require.NoError(t, err)

	commitment3, err := service.GeneratePIICommitment(pii)
	require.NoError(t, err)

	// All should be identical
	require.Equal(t, commitment1, commitment2)
	require.Equal(t, commitment2, commitment3)
}

// ============================================================================
// Redaction Tests
// ============================================================================

func TestRedactSensitiveFields(t *testing.T) {
	data := map[string]interface{}{
		"name":     "Alice",
		"email":    "alice@example.com",
		"ssn":      "123-45-6789",
		"passport": "AB123456",
		"age":      30,
	}

	sensitiveFields := []string{"ssn", "passport", "email"}

	redacted := RedactSensitiveFields(data, sensitiveFields)

	require.Equal(t, "Alice", redacted["name"])
	require.Equal(t, 30, redacted["age"])
	require.Equal(t, "[REDACTED]", redacted["ssn"])
	require.Equal(t, "[REDACTED]", redacted["passport"])
	require.Equal(t, "[REDACTED]", redacted["email"])
}

func TestRedactSensitiveFields_EmptyList(t *testing.T) {
	data := map[string]interface{}{
		"name": "Alice",
		"ssn":  "123-45-6789",
	}

	redacted := RedactSensitiveFields(data, []string{})

	// Nothing should be redacted
	require.Equal(t, data["name"], redacted["name"])
	require.Equal(t, data["ssn"], redacted["ssn"])
}

// ============================================================================
// Sensitive Fields List Tests
// ============================================================================

func TestGetSensitiveFieldsList(t *testing.T) {
	tests := []struct {
		dataType      string
		expectedCount int
		mustContain   []string
	}{
		{
			dataType:      "kyc",
			expectedCount: 11,
			mustContain:   []string{"ssn", "passport_number", "full_name"},
		},
		{
			dataType:      "aml",
			expectedCount: 8,
			mustContain:   []string{"source_of_funds", "occupation", "risk_factors"},
		},
		{
			dataType:      "suspicious_activity",
			expectedCount: 6,
			mustContain:   []string{"description", "sar_details", "indicators"},
		},
		{
			dataType:      "tax",
			expectedCount: 8,
			mustContain:   []string{"ssn", "tax_id", "income"},
		},
		{
			dataType:      "sanctions",
			expectedCount: 3,
			mustContain:   []string{"screening_details", "match_details"},
		},
		{
			dataType:      "unknown",
			expectedCount: 0,
			mustContain:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.dataType, func(t *testing.T) {
			fields := GetSensitiveFieldsList(tt.dataType)
			require.Len(t, fields, tt.expectedCount)

			for _, mustHave := range tt.mustContain {
				require.Contains(t, fields, mustHave)
			}
		})
	}
}

// ============================================================================
// Constant-Time Comparison Tests
// ============================================================================

func TestConstantTimeCompare_Equal(t *testing.T) {
	a := []byte{0x01, 0x02, 0x03, 0x04}
	b := []byte{0x01, 0x02, 0x03, 0x04}

	result := constantTimeCompare(a, b)
	require.True(t, result)
}

func TestConstantTimeCompare_NotEqual(t *testing.T) {
	a := []byte{0x01, 0x02, 0x03, 0x04}
	b := []byte{0x01, 0x02, 0x03, 0x05}

	result := constantTimeCompare(a, b)
	require.False(t, result)
}

func TestConstantTimeCompare_DifferentLength(t *testing.T) {
	a := []byte{0x01, 0x02, 0x03, 0x04}
	b := []byte{0x01, 0x02, 0x03}

	result := constantTimeCompare(a, b)
	require.False(t, result)
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestCompleteWorkflow_KYCCommitment(t *testing.T) {
	service := NewDataProtectionService()

	// Step 1: Provider collects PII off-chain
	pii := &PIIData{
		FullName:       "Alice Smith",
		DateOfBirth:    "1990-01-01",
		SSN:            "123-45-6789",
		PassportNumber: "AB123456",
		Addresses:      []string{"123 Main St, Anytown, USA"},
		PhoneNumbers:   []string{"+1-555-0123"},
		EmailAddresses: []string{"alice@example.com"},
	}

	// Step 2: Generate commitment for on-chain storage
	commitment, err := service.GeneratePIICommitment(pii)
	require.NoError(t, err)

	// Step 3: Store commitment on-chain (simulated)
	// In real code: keeper.SetKYCRecord(ctx, &types.KYCRecord{PiiCommitment: commitment})
	storedCommitment := commitment

	// Step 4: Later, verify data integrity
	// Provider retrieves original PII from off-chain database
	retrievedPII := pii // Simulated off-chain retrieval

	// Step 5: Verify commitment matches
	valid, err := service.VerifyPIICommitment(retrievedPII, storedCommitment)
	require.NoError(t, err)
	require.True(t, valid, "PII should verify against stored commitment")
}

func TestCompleteWorkflow_TamperedData(t *testing.T) {
	service := NewDataProtectionService()

	// Original PII
	originalPII := &PIIData{
		FullName:    "Alice Smith",
		SSN:         "123-45-6789",
		DateOfBirth: "1990-01-01",
	}

	// Generate and store commitment
	commitment, err := service.GeneratePIICommitment(originalPII)
	require.NoError(t, err)

	// Attacker tries to modify data
	tamperedPII := &PIIData{
		FullName:    "Alice Smith",
		SSN:         "999-99-9999", // Tampered!
		DateOfBirth: "1990-01-01",
	}

	// Verification should fail
	valid, err := service.VerifyPIICommitment(tamperedPII, commitment)
	require.NoError(t, err)
	require.False(t, valid, "Tampered data should not verify")
}

func TestCompleteWorkflow_DisplayCommitment(t *testing.T) {
	service := NewDataProtectionService()

	pii := &PIIData{
		FullName: "Alice Smith",
		SSN:      "123-45-6789",
	}

	// Generate commitment
	commitment, err := service.GeneratePIICommitment(pii)
	require.NoError(t, err)

	// Convert to hex for display in logs/UI
	hexCommitment := service.CommitmentToHex(commitment)
	require.Len(t, hexCommitment, 64) // 32 bytes = 64 hex chars

	// Later, parse hex commitment from input
	parsedCommitment, err := service.HexToCommitment(hexCommitment)
	require.NoError(t, err)
	require.Equal(t, commitment, parsedCommitment)
}

// ============================================================================
// Security Property Tests
// ============================================================================

func TestSecurityProperty_PreimageResistance(t *testing.T) {
	service := NewDataProtectionService()

	// Given a commitment, it should be computationally infeasible
	// to find the original data
	pii := &PIIData{
		SSN: "123-45-6789",
	}

	commitment, err := service.GeneratePIICommitment(pii)
	require.NoError(t, err)

	// An attacker with only the commitment cannot derive the SSN
	// This test verifies the commitment doesn't leak information
	require.NotContains(t, string(commitment), "123-45-6789")
	require.NotContains(t, hex.EncodeToString(commitment), "123456789")
}

func TestSecurityProperty_DeterministicCommitment(t *testing.T) {
	service := NewDataProtectionService()

	// Same input must always produce same output
	data := map[string]string{
		"field1": "value1",
		"field2": "value2",
	}

	commitments := make([][]byte, 100)
	for i := 0; i < 100; i++ {
		c, err := service.GenerateCommitment(data)
		require.NoError(t, err)
		commitments[i] = c
	}

	// All commitments must be identical
	for i := 1; i < 100; i++ {
		require.Equal(t, commitments[0], commitments[i])
	}
}

func TestSecurityProperty_JSONCanonicalization(t *testing.T) {
	service := NewDataProtectionService()

	// Different field ordering should produce same commitment
	// due to JSON canonicalization
	data1 := map[string]string{
		"z_field": "value_z",
		"a_field": "value_a",
		"m_field": "value_m",
	}

	data2 := map[string]string{
		"a_field": "value_a",
		"m_field": "value_m",
		"z_field": "value_z",
	}

	commitment1, err := service.GenerateCommitment(data1)
	require.NoError(t, err)

	commitment2, err := service.GenerateCommitment(data2)
	require.NoError(t, err)

	// Marshal both to verify they produce same JSON
	json1, _ := json.Marshal(data1)
	json2, _ := json.Marshal(data2)

	// JSON marshaling sorts keys, so both should be identical
	require.Equal(t, string(json1), string(json2))
	require.Equal(t, commitment1, commitment2)
}
