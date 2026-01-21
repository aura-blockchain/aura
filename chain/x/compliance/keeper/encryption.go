// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// DataProtectionService provides cryptographic commitments for sensitive data
// that must not be stored in plaintext on-chain.
//
// Architecture Decision:
// ----------------------
// Blockchain data is immutable and publicly readable. Even encrypted data
// stored on-chain is vulnerable to future cryptanalysis and quantum attacks.
//
// Therefore, this module uses SHA-256 commitments instead of encryption:
// - Sensitive data is hashed with SHA-256 before storage
// - Original data is stored off-chain by authorized providers
// - On-chain commitments allow verification without exposing data
// - Complies with GDPR Article 32 (appropriate technical measures)
//
// Usage Pattern:
// --------------
// 1. Provider collects sensitive PII off-chain
// 2. Provider generates commitment: SHA-256(canonical_json(data))
// 3. Only commitment is stored on-chain
// 4. Provider stores original data in secure off-chain system
// 5. User can verify data integrity by comparing commitments
//
// Security Properties:
// -------------------
// - Pre-image resistance: Cannot derive original data from commitment
// - Second pre-image resistance: Cannot forge data matching commitment
// - Collision resistance: Infeasible to find two inputs with same commitment
// - Deterministic: Same input always produces same commitment
type DataProtectionService struct{}

// NewDataProtectionService creates a new data protection service
func NewDataProtectionService() *DataProtectionService {
	return &DataProtectionService{}
}

// GenerateCommitment creates a SHA-256 commitment for arbitrary data
//
// The input is marshaled to canonical JSON to ensure deterministic hashing:
// - Keys are sorted alphabetically
// - No whitespace
// - UTF-8 encoding
//
// Returns: 32-byte SHA-256 hash as byte slice
func (s *DataProtectionService) GenerateCommitment(data interface{}) ([]byte, error) {
	// Marshal to canonical JSON for deterministic hashing
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data for commitment: %w", err)
	}

	// Generate SHA-256 hash
	hash := sha256.Sum256(jsonData)
	return hash[:], nil
}

// GenerateCommitmentFromBytes creates a SHA-256 commitment from raw bytes
//
// Use this when data is already in byte form (e.g., serialized protobuf)
func (s *DataProtectionService) GenerateCommitmentFromBytes(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// VerifyCommitment verifies that data matches a commitment
//
// Returns true if SHA-256(data) matches the stored commitment
func (s *DataProtectionService) VerifyCommitment(data interface{}, commitment []byte) (bool, error) {
	if len(commitment) != 32 {
		return false, fmt.Errorf("invalid commitment length: expected 32 bytes, got %d", len(commitment))
	}

	// Generate commitment from provided data
	computed, err := s.GenerateCommitment(data)
	if err != nil {
		return false, err
	}

	// Compare commitments (constant-time comparison)
	return constantTimeCompare(computed, commitment), nil
}

// VerifyCommitmentFromBytes verifies raw bytes against a commitment
func (s *DataProtectionService) VerifyCommitmentFromBytes(data []byte, commitment []byte) bool {
	if len(commitment) != 32 {
		return false
	}

	computed := s.GenerateCommitmentFromBytes(data)
	return constantTimeCompare(computed, commitment)
}

// CommitmentToHex converts a commitment to hex string for display
func (s *DataProtectionService) CommitmentToHex(commitment []byte) string {
	return hex.EncodeToString(commitment)
}

// HexToCommitment converts hex string back to commitment bytes
func (s *DataProtectionService) HexToCommitment(hexStr string) ([]byte, error) {
	commitment, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hex string: %w", err)
	}
	if len(commitment) != 32 {
		return nil, fmt.Errorf("invalid commitment length: expected 32 bytes, got %d", len(commitment))
	}
	return commitment, nil
}

// constantTimeCompare performs constant-time comparison of two byte slices
// to prevent timing attacks
func constantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}

// SensitiveFieldCommitment represents a commitment for a single sensitive field
type SensitiveFieldCommitment struct {
	FieldName  string `json:"field_name"`
	Commitment []byte `json:"commitment"`
}

// GenerateFieldCommitments creates commitments for multiple sensitive fields
//
// Input: map[string]interface{} where keys are field names
// Output: map[string][]byte where values are SHA-256 commitments
//
// Example:
//
//	fields := map[string]interface{}{
//	    "ssn": "123-45-6789",
//	    "passport": "AB123456",
//	    "address": "123 Main St",
//	}
//	commitments, err := service.GenerateFieldCommitments(fields)
//	// commitments["ssn"] = SHA-256("123-45-6789")
func (s *DataProtectionService) GenerateFieldCommitments(fields map[string]interface{}) (map[string][]byte, error) {
	commitments := make(map[string][]byte)

	for fieldName, value := range fields {
		commitment, err := s.GenerateCommitment(value)
		if err != nil {
			return nil, fmt.Errorf("failed to generate commitment for field %s: %w", fieldName, err)
		}
		commitments[fieldName] = commitment
	}

	return commitments, nil
}

// VerifyFieldCommitments verifies multiple field commitments
//
// Returns true only if ALL fields match their commitments
func (s *DataProtectionService) VerifyFieldCommitments(
	fields map[string]interface{},
	commitments map[string][]byte,
) (bool, error) {
	if len(fields) != len(commitments) {
		return false, fmt.Errorf("field count mismatch: %d fields vs %d commitments", len(fields), len(commitments))
	}

	for fieldName, value := range fields {
		commitment, exists := commitments[fieldName]
		if !exists {
			return false, fmt.Errorf("missing commitment for field: %s", fieldName)
		}

		matches, err := s.VerifyCommitment(value, commitment)
		if err != nil {
			return false, fmt.Errorf("failed to verify field %s: %w", fieldName, err)
		}
		if !matches {
			return false, nil
		}
	}

	return true, nil
}

// PIIData represents personally identifiable information stored off-chain
//
// This struct is NEVER stored on-chain. It is used by off-chain providers
// to generate commitments that are stored on-chain.
type PIIData struct {
	// Identity fields
	FullName       string   `json:"full_name,omitempty"`
	DateOfBirth    string   `json:"date_of_birth,omitempty"`
	Nationality    string   `json:"nationality,omitempty"`
	Residence      string   `json:"residence,omitempty"`
	TaxID          string   `json:"tax_id,omitempty"`
	SSN            string   `json:"ssn,omitempty"`
	PassportNumber string   `json:"passport_number,omitempty"`
	IDNumber       string   `json:"id_number,omitempty"`
	Addresses      []string `json:"addresses,omitempty"`
	PhoneNumbers   []string `json:"phone_numbers,omitempty"`
	EmailAddresses []string `json:"email_addresses,omitempty"`

	// Financial fields
	SourceOfFunds      []string `json:"source_of_funds,omitempty"`
	Occupation         string   `json:"occupation,omitempty"`
	Employer           string   `json:"employer,omitempty"`
	AnnualIncome       string   `json:"annual_income,omitempty"`
	NetWorth           string   `json:"net_worth,omitempty"`
	BankAccounts       []string `json:"bank_accounts,omitempty"`
	CreditCardNumbers  []string `json:"credit_card_numbers,omitempty"`
	InvestmentAccounts []string `json:"investment_accounts,omitempty"`

	// Transaction details
	TransactionDetails  string   `json:"transaction_details,omitempty"`
	CounterpartyDetails string   `json:"counterparty_details,omitempty"`
	PaymentInstruments  []string `json:"payment_instruments,omitempty"`

	// Compliance fields
	RiskFactors        []string `json:"risk_factors,omitempty"`
	SARDetails         string   `json:"sar_details,omitempty"`
	InvestigationNotes string   `json:"investigation_notes,omitempty"`
	ScreeningDetails   string   `json:"screening_details,omitempty"`

	// Metadata for audit
	CollectedBy        string   `json:"collected_by,omitempty"`
	CollectedAt        string   `json:"collected_at,omitempty"`
	VerificationMethod string   `json:"verification_method,omitempty"`
	DocumentHashes     []string `json:"document_hashes,omitempty"`
}

// GeneratePIICommitment creates a commitment for PII data
//
// This is the primary method for protecting personal data in compliance records.
// The PIIData struct should be stored off-chain in a secure, GDPR-compliant
// database, and only the commitment stored on-chain.
func (s *DataProtectionService) GeneratePIICommitment(pii *PIIData) ([]byte, error) {
	if pii == nil {
		return nil, fmt.Errorf("PII data cannot be nil")
	}
	return s.GenerateCommitment(pii)
}

// VerifyPIICommitment verifies PII data against a stored commitment
func (s *DataProtectionService) VerifyPIICommitment(pii *PIIData, commitment []byte) (bool, error) {
	if pii == nil {
		return false, fmt.Errorf("PII data cannot be nil")
	}
	return s.VerifyCommitment(pii, commitment)
}

// RedactSensitiveFields replaces sensitive fields with "[REDACTED]" markers
//
// Use this when displaying data in logs or responses where the actual
// sensitive values must not be exposed.
func RedactSensitiveFields(data map[string]interface{}, sensitiveFields []string) map[string]interface{} {
	redacted := make(map[string]interface{})
	sensitiveSet := make(map[string]bool)

	for _, field := range sensitiveFields {
		sensitiveSet[field] = true
	}

	for key, value := range data {
		if sensitiveSet[key] {
			redacted[key] = "[REDACTED]"
		} else {
			redacted[key] = value
		}
	}

	return redacted
}

// GetSensitiveFieldsList returns the list of fields considered sensitive
// for each compliance data type
func GetSensitiveFieldsList(dataType string) []string {
	switch dataType {
	case "kyc":
		return []string{
			"full_name", "date_of_birth", "ssn", "passport_number",
			"id_number", "addresses", "phone_numbers", "email_addresses",
			"tax_id", "nationality", "residence",
		}
	case "aml":
		return []string{
			"risk_factors", "source_of_funds", "occupation", "employer",
			"annual_income", "net_worth", "bank_accounts", "investment_accounts",
		}
	case "suspicious_activity":
		return []string{
			"description", "transaction_details", "counterparty_details",
			"investigation_notes", "sar_details", "indicators",
		}
	case "tax":
		return []string{
			"ssn", "tax_id", "income", "capital_gains", "capital_losses",
			"bank_accounts", "investment_accounts", "transaction_details",
		}
	case "sanctions":
		return []string{
			"screening_details", "match_details", "investigation_notes",
		}
	default:
		return []string{}
	}
}
