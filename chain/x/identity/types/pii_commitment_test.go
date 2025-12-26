// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"bytes"
	"testing"
)

// mustGenerateSalt is a test helper that generates salt and fails test on error
func mustGenerateSalt(t testing.TB) []byte {
	t.Helper()
	salt, err := GenerateCommitmentSalt()
	if err != nil {
		t.Fatalf("GenerateCommitmentSalt failed: %v", err)
	}
	return salt
}

// TestGenerateCommitmentSalt tests salt generation
func TestGenerateCommitmentSalt(t *testing.T) {
	// Generate multiple salts
	salt1, err1 := GenerateCommitmentSalt()
	salt2, err2 := GenerateCommitmentSalt()
	salt3, err3 := GenerateCommitmentSalt()

	// Verify no errors
	if err1 != nil {
		t.Fatalf("salt1 generation failed: %v", err1)
	}
	if err2 != nil {
		t.Fatalf("salt2 generation failed: %v", err2)
	}
	if err3 != nil {
		t.Fatalf("salt3 generation failed: %v", err3)
	}

	// Verify length
	if len(salt1) != 32 {
		t.Errorf("expected salt length 32, got %d", len(salt1))
	}
	if len(salt2) != 32 {
		t.Errorf("expected salt length 32, got %d", len(salt2))
	}
	if len(salt3) != 32 {
		t.Errorf("expected salt length 32, got %d", len(salt3))
	}

	// Verify uniqueness (salts should be different with overwhelming probability)
	if bytes.Equal(salt1, salt2) {
		t.Error("salt1 and salt2 should not be equal")
	}
	if bytes.Equal(salt2, salt3) {
		t.Error("salt2 and salt3 should not be equal")
	}
	if bytes.Equal(salt1, salt3) {
		t.Error("salt1 and salt3 should not be equal")
	}

	// Verify not all zeros
	allZeros := true
	for _, b := range salt1 {
		if b != 0 {
			allZeros = false
			break
		}
	}
	if allZeros {
		t.Error("salt should not be all zeros")
	}
}

// TestComputePIICommitment_Deterministic tests that same input produces same output
func TestComputePIICommitment_Deterministic(t *testing.T) {
	piiData := map[string]string{
		"name":  "Alice Smith",
		"email": "alice@example.com",
		"dob":   "1990-01-01",
	}

	salt := []byte("test-salt-32-bytes-long-enough!")

	commitment1 := ComputePIICommitment(piiData, salt)
	commitment2 := ComputePIICommitment(piiData, salt)

	if !bytes.Equal(commitment1, commitment2) {
		t.Error("same input should produce same commitment")
	}
}

// TestComputePIICommitment_DifferentData tests that different data produces different commitments
func TestComputePIICommitment_DifferentData(t *testing.T) {
	salt := []byte("test-salt-32-bytes-long-enough!")

	data1 := map[string]string{
		"name":  "Alice Smith",
		"email": "alice@example.com",
	}

	data2 := map[string]string{
		"name":  "Bob Jones",
		"email": "bob@example.com",
	}

	commitment1 := ComputePIICommitment(data1, salt)
	commitment2 := ComputePIICommitment(data2, salt)

	if bytes.Equal(commitment1, commitment2) {
		t.Error("different data should produce different commitments")
	}
}

// TestComputePIICommitment_DifferentSalt tests that different salts produce different commitments
func TestComputePIICommitment_DifferentSalt(t *testing.T) {
	piiData := map[string]string{
		"name":  "Alice Smith",
		"email": "alice@example.com",
	}

	salt1 := []byte("salt1-32-bytes-long-enough!!!!!")
	salt2 := []byte("salt2-32-bytes-long-enough!!!!!")

	commitment1 := ComputePIICommitment(piiData, salt1)
	commitment2 := ComputePIICommitment(piiData, salt2)

	if bytes.Equal(commitment1, commitment2) {
		t.Error("different salts should produce different commitments")
	}
}

// TestComputePIICommitment_KeyOrder tests that attribute order doesn't matter
func TestComputePIICommitment_KeyOrder(t *testing.T) {
	salt := []byte("test-salt-32-bytes-long-enough!")

	// Add attributes in different orders
	data1 := map[string]string{
		"name":  "Alice",
		"email": "alice@example.com",
		"age":   "30",
	}

	data2 := map[string]string{
		"email": "alice@example.com",
		"age":   "30",
		"name":  "Alice",
	}

	data3 := map[string]string{
		"age":   "30",
		"name":  "Alice",
		"email": "alice@example.com",
	}

	commitment1 := ComputePIICommitment(data1, salt)
	commitment2 := ComputePIICommitment(data2, salt)
	commitment3 := ComputePIICommitment(data3, salt)

	if !bytes.Equal(commitment1, commitment2) {
		t.Error("attribute order should not affect commitment")
	}
	if !bytes.Equal(commitment2, commitment3) {
		t.Error("attribute order should not affect commitment")
	}
}

// TestComputePIICommitment_EmptyData tests empty PII data
func TestComputePIICommitment_EmptyData(t *testing.T) {
	salt := []byte("test-salt-32-bytes-long-enough!")
	emptyData := map[string]string{}

	commitment1 := ComputePIICommitment(emptyData, salt)
	commitment2 := ComputePIICommitment(emptyData, salt)

	// Should be deterministic even with empty data
	if !bytes.Equal(commitment1, commitment2) {
		t.Error("empty data should produce deterministic commitment")
	}

	// Should produce non-zero commitment
	allZeros := true
	for _, b := range commitment1 {
		if b != 0 {
			allZeros = false
			break
		}
	}
	if allZeros {
		t.Error("commitment should not be all zeros")
	}
}

// TestComputePIICommitment_SingleAttribute tests single attribute
func TestComputePIICommitment_SingleAttribute(t *testing.T) {
	salt := []byte("test-salt-32-bytes-long-enough!")
	data := map[string]string{"name": "Alice"}

	commitment := ComputePIICommitment(data, salt)

	if len(commitment) != 32 {
		t.Errorf("expected commitment length 32, got %d", len(commitment))
	}
}

// TestComputePIICommitment_SpecialCharacters tests PII with special characters
func TestComputePIICommitment_SpecialCharacters(t *testing.T) {
	salt := []byte("test-salt-32-bytes-long-enough!")

	data := map[string]string{
		"name":    "Alice O'Brien-Smith",
		"email":   "alice+test@example.com",
		"address": "123 Main St., Apt #5",
		"notes":   "Contains: ||, =, and other special chars!@#$%",
	}

	commitment1 := ComputePIICommitment(data, salt)
	commitment2 := ComputePIICommitment(data, salt)

	if !bytes.Equal(commitment1, commitment2) {
		t.Error("special characters should not break determinism")
	}
}

// TestComputePIICommitment_Unicode tests Unicode characters
func TestComputePIICommitment_Unicode(t *testing.T) {
	salt := []byte("test-salt-32-bytes-long-enough!")

	data := map[string]string{
		"name":    "Alice 王",
		"email":   "alice@例え.jp",
		"message": "Hello 世界! 🌍",
	}

	commitment := ComputePIICommitment(data, salt)

	if len(commitment) != 32 {
		t.Errorf("expected commitment length 32, got %d", len(commitment))
	}
}

// TestVerifyPIICommitment_Valid tests successful verification
func TestVerifyPIICommitment_Valid(t *testing.T) {
	piiData := map[string]string{
		"name":  "Alice Smith",
		"email": "alice@example.com",
		"dob":   "1990-01-01",
	}

	salt := mustGenerateSalt(t)
	commitment := ComputePIICommitment(piiData, salt)

	// Verify with same data
	if !VerifyPIICommitment(piiData, commitment, salt) {
		t.Error("verification should succeed with correct data")
	}
}

// TestVerifyPIICommitment_Invalid tests failed verification
func TestVerifyPIICommitment_Invalid(t *testing.T) {
	originalData := map[string]string{
		"name":  "Alice Smith",
		"email": "alice@example.com",
	}

	wrongData := map[string]string{
		"name":  "Bob Jones",
		"email": "bob@example.com",
	}

	salt := mustGenerateSalt(t)
	commitment := ComputePIICommitment(originalData, salt)

	// Verify with wrong data
	if VerifyPIICommitment(wrongData, commitment, salt) {
		t.Error("verification should fail with wrong data")
	}
}

// TestVerifyPIICommitment_WrongSalt tests verification with wrong salt
func TestVerifyPIICommitment_WrongSalt(t *testing.T) {
	piiData := map[string]string{
		"name":  "Alice Smith",
		"email": "alice@example.com",
	}

	correctSalt := mustGenerateSalt(t)
	wrongSalt := mustGenerateSalt(t)

	commitment := ComputePIICommitment(piiData, correctSalt)

	// Verify with wrong salt
	if VerifyPIICommitment(piiData, commitment, wrongSalt) {
		t.Error("verification should fail with wrong salt")
	}
}

// TestVerifyPIICommitment_ModifiedValue tests detection of modified values
func TestVerifyPIICommitment_ModifiedValue(t *testing.T) {
	originalData := map[string]string{
		"name":  "Alice Smith",
		"email": "alice@example.com",
		"dob":   "1990-01-01",
	}

	modifiedData := map[string]string{
		"name":  "Alice Smith",
		"email": "alice@example.com",
		"dob":   "1990-01-02", // Modified
	}

	salt := mustGenerateSalt(t)
	commitment := ComputePIICommitment(originalData, salt)

	// Should detect modification
	if VerifyPIICommitment(modifiedData, commitment, salt) {
		t.Error("verification should fail when value is modified")
	}
}

// TestVerifyPIICommitment_MissingAttribute tests detection of missing attributes
func TestVerifyPIICommitment_MissingAttribute(t *testing.T) {
	fullData := map[string]string{
		"name":  "Alice Smith",
		"email": "alice@example.com",
		"dob":   "1990-01-01",
	}

	partialData := map[string]string{
		"name":  "Alice Smith",
		"email": "alice@example.com",
		// "dob" missing
	}

	salt := mustGenerateSalt(t)
	commitment := ComputePIICommitment(fullData, salt)

	// Should detect missing attribute
	if VerifyPIICommitment(partialData, commitment, salt) {
		t.Error("verification should fail when attribute is missing")
	}
}

// TestVerifyPIICommitment_ExtraAttribute tests detection of extra attributes
func TestVerifyPIICommitment_ExtraAttribute(t *testing.T) {
	originalData := map[string]string{
		"name":  "Alice Smith",
		"email": "alice@example.com",
	}

	dataWithExtra := map[string]string{
		"name":  "Alice Smith",
		"email": "alice@example.com",
		"dob":   "1990-01-01", // Extra attribute
	}

	salt := mustGenerateSalt(t)
	commitment := ComputePIICommitment(originalData, salt)

	// Should detect extra attribute
	if VerifyPIICommitment(dataWithExtra, commitment, salt) {
		t.Error("verification should fail when extra attribute is present")
	}
}

// TestComputePIICommitment_LargeData tests commitment with many attributes
func TestComputePIICommitment_LargeData(t *testing.T) {
	largeData := make(map[string]string)
	for i := 0; i < 100; i++ {
		largeData[string(rune('a'+i%26))+string(rune('0'+i%10))] = "value" + string(rune('0'+i))
	}

	salt := mustGenerateSalt(t)
	commitment1 := ComputePIICommitment(largeData, salt)
	commitment2 := ComputePIICommitment(largeData, salt)

	if !bytes.Equal(commitment1, commitment2) {
		t.Error("large data should produce deterministic commitment")
	}

	if len(commitment1) != 32 {
		t.Errorf("expected commitment length 32, got %d", len(commitment1))
	}
}

// TestComputePIICommitment_LongValues tests commitment with long attribute values
func TestComputePIICommitment_LongValues(t *testing.T) {
	longValue := make([]byte, 10000)
	for i := range longValue {
		longValue[i] = byte('a' + (i % 26))
	}

	data := map[string]string{
		"name":     "Alice",
		"longdata": string(longValue),
	}

	salt := mustGenerateSalt(t)
	commitment := ComputePIICommitment(data, salt)

	if len(commitment) != 32 {
		t.Errorf("expected commitment length 32, got %d", len(commitment))
	}
}

// TestComputePIICommitment_CollisionResistance tests different data produces different commitments
func TestComputePIICommitment_CollisionResistance(t *testing.T) {
	salt := mustGenerateSalt(t)
	commitments := make(map[string]bool)

	// Generate many commitments with slightly different data
	for i := 0; i < 100; i++ {
		data := map[string]string{
			"name":  "Alice",
			"id":    string(rune('0' + i/10)),
			"index": string(rune('0' + i%10)),
		}
		commitment := ComputePIICommitment(data, salt)
		commitmentStr := string(commitment)

		if commitments[commitmentStr] {
			t.Errorf("collision detected at index %d", i)
		}
		commitments[commitmentStr] = true
	}
}

// BenchmarkComputePIICommitment benchmarks commitment computation
func BenchmarkComputePIICommitment(b *testing.B) {
	piiData := map[string]string{
		"name":    "Alice Smith",
		"email":   "alice@example.com",
		"dob":     "1990-01-01",
		"address": "123 Main St",
		"phone":   "+1234567890",
	}
	salt := mustGenerateSalt(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputePIICommitment(piiData, salt)
	}
}

// BenchmarkVerifyPIICommitment benchmarks verification
func BenchmarkVerifyPIICommitment(b *testing.B) {
	piiData := map[string]string{
		"name":    "Alice Smith",
		"email":   "alice@example.com",
		"dob":     "1990-01-01",
		"address": "123 Main St",
		"phone":   "+1234567890",
	}
	salt := mustGenerateSalt(b)
	commitment := ComputePIICommitment(piiData, salt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyPIICommitment(piiData, commitment, salt)
	}
}

// BenchmarkGenerateCommitmentSalt benchmarks salt generation
func BenchmarkGenerateCommitmentSalt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateCommitmentSalt()
	}
}
