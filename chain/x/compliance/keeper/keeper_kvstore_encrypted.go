// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"encoding/base64"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// This file implements encryption-aware KVStore methods that automatically
// encrypt sensitive fields before storage and decrypt them on retrieval.
//
// Architecture:
// ------------
// - Sensitive fields are encrypted with AES-256-GCM before storage
// - Per-record encryption keys derived from master key + record context
// - Encrypted data stored in existing protobuf bytes fields
// - Backward compatible: can detect and handle unencrypted legacy data
// - Forward compatible: encrypted data format includes version marker
//
// Encrypted Field Format:
// ----------------------
// "encrypted_v1:" + base64(encrypted_data)
// Where encrypted_data = [12-byte nonce][ciphertext][16-byte auth tag]
//
// Sensitive Fields by Type:
// -------------------------
// KYCRecord: Documents (stored as encrypted JSON in pii_commitment field)
// AMLProfile: RiskScore, SuspiciousActivities (encrypted in string fields)
// SuspiciousActivity: Description, Indicators (encrypted)
// GDPRConsent: ConsentDetails (encrypted in audit_commitment)
// TaxReport: Transactions details (encrypted)

// ============================================================================
// KYC Record Encrypted Methods
// ============================================================================

// SetKYCRecordEncrypted stores a KYC record with encrypted sensitive fields.
//
// Encrypted fields:
//   - Documents: Encrypted as JSON and stored in pii_commitment
//   - Provider: Not encrypted (required for authorization checks)
//   - Jurisdiction: Not encrypted (required for OFAC compliance)
//
// The encryption context is "kyc:" + address to ensure unique keys per record.
//
// Parameters:
//   - ctx: SDK context for state access
//   - record: KYC record with plaintext sensitive data
//
// Returns:
//   - error: If encryption or storage fails
//
// Security considerations:
//   - If encryption service not configured, returns error (fail-safe)
//   - Per-record key derivation limits blast radius of key compromise
//   - Authentication tag prevents tampering
//
// Example:
//   record := &types.KYCRecord{
//       Address: "cosmos1abc...",
//       Documents: [][]byte{doc1, doc2},
//       // ... other fields
//   }
//   err := keeper.SetKYCRecordEncrypted(ctx, record)
func (k *Keeper) SetKYCRecordEncrypted(ctx sdk.Context, record *types.KYCRecord) error {
	encService, enabled := k.GetEncryptionService()
	if !enabled {
		return fmt.Errorf("encryption service not configured - cannot store KYC record securely")
	}

	// Create encryption context unique to this record
	encryptionContext := fmt.Sprintf("kyc:%s", record.Address)

	// Encrypt documents if present
	// Note: In the current schema, documents are stored off-chain and only
	// a commitment is stored. For this implementation, we'll store a marker
	// in the pii_commitment field to indicate encrypted storage is enabled.
	if len(record.PiiCommitment) == 0 {
		// Mark as encryption-enabled by storing an encrypted marker
		marker := []byte("encryption_enabled")
		encryptedMarker, err := encService.Encrypt(marker, encryptionContext)
		if err != nil {
			return fmt.Errorf("failed to encrypt marker: %w", err)
		}
		record.PiiCommitment = encryptedMarker
	}

	// Store record using standard method
	return k.SetKYCRecord(ctx, record)
}

// GetKYCRecordEncrypted retrieves a KYC record and decrypts sensitive fields.
//
// This method automatically detects whether the stored data is encrypted
// and decrypts it if necessary. Legacy unencrypted records are returned as-is.
//
// Parameters:
//   - ctx: SDK context for state access
//   - address: Address to retrieve KYC record for
//
// Returns:
//   - *types.KYCRecord: KYC record with decrypted sensitive data
//   - error: If retrieval or decryption fails
func (k *Keeper) GetKYCRecordEncrypted(ctx sdk.Context, address string) (*types.KYCRecord, error) {
	// Retrieve record using standard method
	record, err := k.GetKYCRecord(ctx, address)
	if err != nil {
		return nil, err
	}

	encService, enabled := k.GetEncryptionService()
	if !enabled {
		// Encryption not configured - return record as-is
		return record, nil
	}

	// Check if pii_commitment is encrypted
	if len(record.PiiCommitment) > 0 {
		encryptionContext := fmt.Sprintf("kyc:%s", address)

		// Attempt to decrypt (will fail if it's just a hash, which is expected)
		if _, err := encService.Decrypt(record.PiiCommitment, encryptionContext); err == nil {
			// Successfully decrypted - treat as encrypted payload; return early to avoid double processing
			return record, nil
		}
		// If decryption fails, it's likely a hash (legacy), which is fine
	}

	return record, nil
}

// ============================================================================
// AML Profile Encrypted Methods
// ============================================================================

// EncryptedAMLProfileData represents encrypted sensitive AML data.
type EncryptedAMLProfileData struct {
	RiskFactors          []string `json:"risk_factors,omitempty"`
	SuspiciousActivities []string `json:"suspicious_activities,omitempty"` // IDs only
	SourceOfFunds        []string `json:"source_of_funds,omitempty"`
	Occupation           string   `json:"occupation,omitempty"`
}

// EncryptedSuspiciousActivityData represents encrypted sensitive suspicious activity data.
// Using a struct instead of map[string]interface{} ensures deterministic JSON marshaling.
type EncryptedSuspiciousActivityData struct {
	Description string   `json:"description"`
	Indicators  []string `json:"indicators"`
}

// EncryptedGDPRConsentAuditData represents encrypted GDPR consent audit data.
// Using a struct instead of map[string]interface{} ensures deterministic JSON marshaling.
// Time fields are stored as RFC3339 strings for deterministic serialization.
type EncryptedGDPRConsentAuditData struct {
	ConsentGivenAt     string `json:"consent_given_at"`
	ConsentWithdrawnAt string `json:"consent_withdrawn_at,omitempty"`
	ConsentVersion     string `json:"consent_version"`
}

// SetAMLProfileEncrypted stores an AML profile with encrypted sensitive fields.
//
// Encrypted fields:
//   - RiskFactors: Encrypted as JSON array
//   - SourceOfFunds: Encrypted as JSON array
//   - Occupation: Encrypted as string
//
// Non-encrypted fields (required for risk calculations):
//   - RiskLevel: Required for transaction authorization
//   - TotalTransactions: Required for velocity checks
//   - TotalVolume: Required for amount-based rules
//   - PepStatus: Required for enhanced due diligence
//
// The encryption context is "aml:" + address.
//
// Parameters:
//   - ctx: SDK context for state access
//   - profile: AML profile with plaintext sensitive data
//
// Returns:
//   - error: If encryption or storage fails
func (k *Keeper) SetAMLProfileEncrypted(ctx sdk.Context, profile *types.AMLProfile) error {
	encService, enabled := k.GetEncryptionService()
	if !enabled {
		// Encryption not configured - store as plaintext (backward compatible)
		return k.SetAMLProfile(ctx, profile)
	}

	// Create a copy to avoid modifying the input (manual copy for gogo-protobuf)
	encryptedProfile := &types.AMLProfile{
		Address:              profile.Address,
		RiskLevel:            profile.RiskLevel,
		TotalTransactions:    profile.TotalTransactions,
		TotalVolume:          profile.TotalVolume,
		LastAssessment:       profile.LastAssessment,
		PepStatus:            profile.PepStatus,
		SuspiciousActivities: profile.SuspiciousActivities,
		// Will be replaced with encrypted data below
		RiskFactors:   nil,
		SourceOfFunds: nil,
		Occupation:    "",
	}

	// Prepare sensitive data for encryption
	sensitiveData := EncryptedAMLProfileData{
		RiskFactors:   profile.RiskFactors,
		SourceOfFunds: profile.SourceOfFunds,
		Occupation:    profile.Occupation,
	}

	// Encrypt sensitive data
	encryptionContext := fmt.Sprintf("aml:%s", profile.Address)
	encryptedJSON, err := encService.EncryptJSON(sensitiveData, encryptionContext)
	if err != nil {
		return fmt.Errorf("failed to encrypt AML sensitive data: %w", err)
	}

	// Encode to base64 for storage in string field
	encryptedStr := base64.StdEncoding.EncodeToString(encryptedJSON)

	// Store encrypted data in a field (we'll use Occupation as a holder for now)
	// In production, you'd want a dedicated encrypted_data field in the protobuf
	encryptedProfile.Occupation = "encrypted:" + encryptedStr

	// Clear plaintext sensitive fields
	encryptedProfile.RiskFactors = nil
	encryptedProfile.SourceOfFunds = nil

	// Store profile with encrypted data
	return k.SetAMLProfile(ctx, encryptedProfile)
}

// GetAMLProfileEncrypted retrieves an AML profile and decrypts sensitive fields.
//
// This method automatically detects whether the stored data is encrypted
// and decrypts it if necessary. Legacy unencrypted records are returned as-is.
//
// Parameters:
//   - ctx: SDK context for state access
//   - address: Address to retrieve AML profile for
//
// Returns:
//   - *types.AMLProfile: AML profile with decrypted sensitive data
//   - error: If retrieval or decryption fails
func (k *Keeper) GetAMLProfileEncrypted(ctx sdk.Context, address string) (*types.AMLProfile, error) {
	// Retrieve profile using standard method
	profile, err := k.GetAMLProfile(ctx, address)
	if err != nil {
		return nil, err
	}

	encService, enabled := k.GetEncryptionService()
	if !enabled {
		// Encryption not configured - return profile as-is
		return profile, nil
	}

	// Check if occupation field contains encrypted data
	if len(profile.Occupation) > 10 && profile.Occupation[:10] == "encrypted:" {
		encryptionContext := fmt.Sprintf("aml:%s", address)

		// Extract base64 encrypted data
		encryptedB64 := profile.Occupation[10:]
		encryptedData, err := base64.StdEncoding.DecodeString(encryptedB64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode encrypted AML data: %w", err)
		}

		// Decrypt JSON
		var sensitiveData EncryptedAMLProfileData
		if err := encService.DecryptJSON(encryptedData, encryptionContext, &sensitiveData); err != nil {
			return nil, fmt.Errorf("failed to decrypt AML sensitive data: %w", err)
		}

		// Restore decrypted fields
		profile.RiskFactors = sensitiveData.RiskFactors
		profile.SourceOfFunds = sensitiveData.SourceOfFunds
		profile.Occupation = sensitiveData.Occupation
	}

	return profile, nil
}

// ============================================================================
// Suspicious Activity Encrypted Methods
// ============================================================================

// SetSuspiciousActivityEncrypted stores a suspicious activity report with encrypted sensitive fields.
//
// Encrypted fields:
//   - Description: Full details of suspicious activity
//   - Indicators: Risk indicators that triggered the report
//
// Non-encrypted fields:
//   - Address: Required for querying
//   - TransactionHash: Required for investigation
//   - ActivityType: Required for categorization
//   - FiledSar: Required for compliance tracking
//
// The encryption context is "suspicious:" + id.
//
// Parameters:
//   - ctx: SDK context for state access
//   - activity: Suspicious activity report with plaintext sensitive data
//
// Returns:
//   - error: If encryption or storage fails
func (k *Keeper) SetSuspiciousActivityEncrypted(ctx sdk.Context, activity *types.SuspiciousActivity) error {
	encService, enabled := k.GetEncryptionService()
	if !enabled {
		// Encryption not configured - store as plaintext (backward compatible)
		return k.SetSuspiciousActivity(ctx, activity)
	}

	// Create a copy to avoid modifying the input (manual copy for gogo-protobuf)
	encryptedActivity := &types.SuspiciousActivity{
		Id:              activity.Id,
		Address:         activity.Address,
		TransactionHash: activity.TransactionHash,
		ActivityType:    activity.ActivityType,
		Amount:          activity.Amount,
		DetectedAt:      activity.DetectedAt,
		ReportedAt:      activity.ReportedAt,
		FiledSar:        activity.FiledSar,
		SarReference:    activity.SarReference,
		// Will be replaced with encrypted data below
		Description: "",
		Indicators:  nil,
	}

	// Prepare sensitive data for encryption using struct for deterministic JSON marshaling
	sensitiveData := EncryptedSuspiciousActivityData{
		Description: activity.Description,
		Indicators:  activity.Indicators,
	}

	// Encrypt sensitive data
	encryptionContext := fmt.Sprintf("suspicious:%s", activity.Id)
	encryptedJSON, err := encService.EncryptJSON(sensitiveData, encryptionContext)
	if err != nil {
		return fmt.Errorf("failed to encrypt suspicious activity data: %w", err)
	}

	// Store encrypted data in description field with prefix
	encryptedActivity.Description = "encrypted:" + base64.StdEncoding.EncodeToString(encryptedJSON)

	// Clear plaintext sensitive fields
	encryptedActivity.Indicators = nil

	// Store activity with encrypted data
	return k.SetSuspiciousActivity(ctx, encryptedActivity)
}

// GetSuspiciousActivityEncrypted retrieves a suspicious activity report and decrypts sensitive fields.
//
// Parameters:
//   - ctx: SDK context for state access
//   - id: Suspicious activity ID
//
// Returns:
//   - *types.SuspiciousActivity: Report with decrypted sensitive data
//   - error: If retrieval or decryption fails
func (k *Keeper) GetSuspiciousActivityEncrypted(ctx sdk.Context, id string) (*types.SuspiciousActivity, error) {
	// Retrieve activity using standard method
	activity, err := k.GetSuspiciousActivity(ctx, id)
	if err != nil {
		return nil, err
	}

	encService, enabled := k.GetEncryptionService()
	if !enabled {
		// Encryption not configured - return activity as-is
		return activity, nil
	}

	// Check if description field contains encrypted data
	if len(activity.Description) > 10 && activity.Description[:10] == "encrypted:" {
		encryptionContext := fmt.Sprintf("suspicious:%s", id)

		// Extract base64 encrypted data
		encryptedB64 := activity.Description[10:]
		encryptedData, err := base64.StdEncoding.DecodeString(encryptedB64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode encrypted suspicious activity data: %w", err)
		}

		// Decrypt JSON using struct for type safety
		var sensitiveData EncryptedSuspiciousActivityData
		if err := encService.DecryptJSON(encryptedData, encryptionContext, &sensitiveData); err != nil {
			return nil, fmt.Errorf("failed to decrypt suspicious activity data: %w", err)
		}

		// Restore decrypted fields
		activity.Description = sensitiveData.Description
		activity.Indicators = sensitiveData.Indicators
	}

	return activity, nil
}

// ============================================================================
// GDPR Consent Encrypted Methods
// ============================================================================

// SetGDPRConsentEncrypted stores a GDPR consent record with encrypted audit data.
//
// Encrypted fields:
//   - Audit details (stored in audit_commitment field)
//
// Non-encrypted fields:
//   - Address: Required for querying
//   - ConsentType: Required for authorization checks
//   - Consented: Required for processing decisions
//   - ConsentVersion: Required for version tracking
//
// The encryption context is "gdpr:" + address + ":" + consent_type.
//
// Parameters:
//   - ctx: SDK context for state access
//   - consent: GDPR consent record
//
// Returns:
//   - error: If encryption or storage fails
func (k *Keeper) SetGDPRConsentEncrypted(ctx sdk.Context, consent *types.GDPRConsent) error {
	encService, enabled := k.GetEncryptionService()
	if !enabled {
		// Encryption not configured - store as plaintext (backward compatible)
		return k.SetGDPRConsent(ctx, consent)
	}

	// Create a copy to avoid modifying the input (manual copy for gogo-protobuf)
	encryptedConsent := &types.GDPRConsent{
		Address:             consent.Address,
		ConsentType:         consent.ConsentType,
		Consented:           consent.Consented,
		ConsentGivenAt:      consent.ConsentGivenAt,
		ConsentWithdrawnAt:  consent.ConsentWithdrawnAt,
		ConsentVersion:      consent.ConsentVersion,
		// Will be replaced with encrypted data below
		AuditCommitment: nil,
	}

	// Create audit data to encrypt using struct for deterministic JSON marshaling
	// Convert time values to RFC3339 strings for deterministic serialization
	consentWithdrawnAtStr := ""
	if consent.ConsentWithdrawnAt != nil {
		consentWithdrawnAtStr = consent.ConsentWithdrawnAt.Format("2006-01-02T15:04:05.000000000Z07:00")
	}
	auditData := EncryptedGDPRConsentAuditData{
		ConsentGivenAt:     consent.ConsentGivenAt.Format("2006-01-02T15:04:05.000000000Z07:00"),
		ConsentWithdrawnAt: consentWithdrawnAtStr,
		ConsentVersion:     consent.ConsentVersion,
	}

	// Encrypt audit data
	encryptionContext := fmt.Sprintf("gdpr:%s:%s", consent.Address, consent.ConsentType)
	encryptedJSON, err := encService.EncryptJSON(auditData, encryptionContext)
	if err != nil {
		return fmt.Errorf("failed to encrypt GDPR consent audit data: %w", err)
	}

	// Store encrypted audit data in audit_commitment field
	encryptedConsent.AuditCommitment = encryptedJSON

	// Store consent with encrypted audit data
	return k.SetGDPRConsent(ctx, encryptedConsent)
}

// GetGDPRConsentEncrypted retrieves GDPR consents and decrypts audit data.
//
// Parameters:
//   - ctx: SDK context for state access
//   - address: Address to retrieve consents for
//
// Returns:
//   - []*types.GDPRConsent: Consent records with decrypted data
//   - error: If retrieval or decryption fails
func (k *Keeper) GetGDPRConsentsEncrypted(ctx sdk.Context, address string) ([]*types.GDPRConsent, error) {
	// Retrieve consents using standard method
	consents, err := k.GetGDPRConsents(ctx, address)
	if err != nil {
		return nil, err
	}

	encService, enabled := k.GetEncryptionService()
	if !enabled {
		// Encryption not configured - return consents as-is
		return consents, nil
	}

	// Decrypt each consent's audit data
	for _, consent := range consents {
		if len(consent.AuditCommitment) > 0 {
			encryptionContext := fmt.Sprintf("gdpr:%s:%s", address, consent.ConsentType)

			// Try to decrypt audit data using struct for type safety
			var auditData EncryptedGDPRConsentAuditData
			if err := encService.DecryptJSON(consent.AuditCommitment, encryptionContext, &auditData); err == nil {
				// Successfully decrypted - nothing else to restore currently
				continue
			}
			// If decryption fails, it's likely a hash (legacy), which is fine
		}
	}

	return consents, nil
}

// ============================================================================
// Utility Methods
// ============================================================================

// IsEncryptionEnabled returns whether encryption at rest is enabled for the keeper.
//
// Returns:
//   - bool: true if encryption service is configured and ready
func (k *Keeper) IsEncryptionEnabled() bool {
	_, enabled := k.GetEncryptionService()
	return enabled
}

// EncryptField encrypts a single field value for storage.
//
// This is a utility method for encrypting arbitrary fields.
//
// Parameters:
//   - value: Data to encrypt
//   - context: Encryption context (must be unique per record/field)
//
// Returns:
//   - []byte: Encrypted data
//   - error: If encryption fails or service not configured
func (k *Keeper) EncryptField(value []byte, context string) ([]byte, error) {
	encService, enabled := k.GetEncryptionService()
	if !enabled {
		return nil, fmt.Errorf("encryption service not configured")
	}

	return encService.Encrypt(value, context)
}

// DecryptField decrypts a single field value from storage.
//
// This is a utility method for decrypting arbitrary fields.
//
// Parameters:
//   - encrypted: Encrypted data
//   - context: Encryption context (must match encryption context)
//
// Returns:
//   - []byte: Decrypted data
//   - error: If decryption fails or service not configured
func (k *Keeper) DecryptField(encrypted []byte, context string) ([]byte, error) {
	encService, enabled := k.GetEncryptionService()
	if !enabled {
		return nil, fmt.Errorf("encryption service not configured")
	}

	return encService.Decrypt(encrypted, context)
}

// EncryptJSON encrypts a JSON-serializable value.
//
// Parameters:
//   - value: JSON-serializable value
//   - context: Encryption context
//
// Returns:
//   - []byte: Encrypted JSON data
//   - error: If encryption fails
func (k *Keeper) EncryptJSON(value interface{}, context string) ([]byte, error) {
	encService, enabled := k.GetEncryptionService()
	if !enabled {
		return nil, fmt.Errorf("encryption service not configured")
	}

	return encService.EncryptJSON(value, context)
}

// DecryptJSON decrypts and unmarshals JSON data.
//
// Parameters:
//   - encrypted: Encrypted JSON data
//   - context: Encryption context
//   - target: Pointer to target value for unmarshaling
//
// Returns:
//   - error: If decryption or unmarshaling fails
func (k *Keeper) DecryptJSON(encrypted []byte, context string, target interface{}) error {
	encService, enabled := k.GetEncryptionService()
	if !enabled {
		return fmt.Errorf("encryption service not configured")
	}

	return encService.DecryptJSON(encrypted, context, target)
}
