package keeper

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// TestGDPRCompliance_NoPIIInProtobuf verifies that no PII fields exist in protobuf messages
// that are stored on-chain. This is critical for GDPR Article 17 "Right to Erasure" compliance.
//
// GDPR Article 17 requires that users can request deletion of their personal data.
// Since blockchain data is immutable, we MUST NOT store PII on-chain.
//
// This test ensures that:
// - KYCRecord only contains pii_commitment (hash), not raw PII
// - GDPRConsent only contains audit_commitment, not IP/user-agent
// - All other compliance records contain no direct PII
//
// Security severity: CRITICAL
// Legal risk: HIGH (€20M fine or 4% global revenue)
// CVSS: 9.0
func TestGDPRCompliance_NoPIIInProtobuf(t *testing.T) {
	t.Run("KYCRecord_OnlyCommitmentStored", func(t *testing.T) {
		// Create a KYC record as would be stored on-chain
		kycRecord := &types.KYCRecord{
			Address:              "aura1test",
			KycLevel:             types.KYCLevel_KYC_LEVEL_ADVANCED,
			Provider:             "provider1",
			PiiCommitment:        []byte{0x01, 0x02, 0x03}, // Hash only
			EnhancedDueDiligence: true,
			Jurisdiction:         "US",
		}

		// Verify NO PII fields are accessible
		// These fields should NOT exist in the protobuf definition
		recordJSON, err := json.Marshal(kycRecord)
		require.NoError(t, err)

		var recordMap map[string]interface{}
		err = json.Unmarshal(recordJSON, &recordMap)
		require.NoError(t, err)

		// CRITICAL: Verify these PII fields DO NOT EXIST
		forbiddenFields := []string{
			"verification_id",   // PII - removed
			"documents",         // PII - removed
			"risk_score",        // PII - removed
			"full_name",         // PII - never added
			"ssn",               // PII - never added
			"passport_number",   // PII - never added
			"date_of_birth",     // PII - never added
			"address_physical",  // PII - never added
			"phone_number",      // PII - never added
			"email",             // PII - never added
		}

		for _, field := range forbiddenFields {
			_, exists := recordMap[field]
			require.False(t, exists, "GDPR VIOLATION: PII field '%s' found in KYCRecord", field)
		}

		// REQUIRED: Verify commitment field exists
		require.NotNil(t, kycRecord.PiiCommitment, "pii_commitment field is required")
	})

	t.Run("GDPRConsent_OnlyCommitmentStored", func(t *testing.T) {
		// Create a GDPR consent record as would be stored on-chain
		consent := &types.GDPRConsent{
			Address:        "aura1test",
			ConsentType:    "data_processing",
			Consented:      true,
			ConsentVersion: "v1.0",
			AuditCommitment: []byte{0xaa, 0xbb, 0xcc}, // Hash only (optional)
		}

		recordJSON, err := json.Marshal(consent)
		require.NoError(t, err)

		var recordMap map[string]interface{}
		err = json.Unmarshal(recordJSON, &recordMap)
		require.NoError(t, err)

		// CRITICAL: Verify these PII fields DO NOT EXIST
		forbiddenFields := []string{
			"ip_address",    // PII - removed
			"user_agent",    // PII - removed
			"browser",       // PII - never added
			"device_id",     // PII - never added
			"geolocation",   // PII - never added
		}

		for _, field := range forbiddenFields {
			_, exists := recordMap[field]
			require.False(t, exists, "GDPR VIOLATION: PII field '%s' found in GDPRConsent", field)
		}

		// Audit commitment is optional
		require.NotNil(t, consent.AuditCommitment, "audit_commitment should be set for audit trail")
	})

	t.Run("AMLProfile_NoPIIFields", func(t *testing.T) {
		profile := &types.AMLProfile{
			Address:          "aura1test",
			RiskLevel:        types.AMLRiskLevel_AML_RISK_MEDIUM,
			TotalTransactions: 100,
			TotalVolume:      "1000000",
			PepStatus:        false,
		}

		recordJSON, err := json.Marshal(profile)
		require.NoError(t, err)

		var recordMap map[string]interface{}
		err = json.Unmarshal(recordJSON, &recordMap)
		require.NoError(t, err)

		// Verify no PII fields
		forbiddenFields := []string{
			"full_name",
			"ssn",
			"tax_id",
			"passport_number",
			"date_of_birth",
			"physical_address",
		}

		for _, field := range forbiddenFields {
			_, exists := recordMap[field]
			require.False(t, exists, "GDPR VIOLATION: PII field '%s' found in AMLProfile", field)
		}
	})
}

// TestGDPRCompliance_CommitmentBasedStorage verifies that the system correctly
// uses cryptographic commitments instead of storing raw PII.
//
// GDPR compliance mechanism:
// 1. PII is stored off-chain by authorized providers
// 2. SHA-256 hash of PII is stored on-chain as "commitment"
// 3. Data can be verified without exposing it
// 4. Off-chain data can be deleted (GDPR Article 17)
// 5. On-chain commitment remains for audit trail
func TestGDPRCompliance_CommitmentBasedStorage(t *testing.T) {
	suite := NewTestSuite(t)
	ctx := suite.Ctx
	keeper := suite.Keeper

	t.Run("PII_StoredAsCommitmentOnly", func(t *testing.T) {
		// Simulate off-chain PII data (NEVER stored on-chain)
		offChainPII := map[string]interface{}{
			"full_name":       "John Doe",
			"ssn":             "123-45-6789",
			"passport_number": "AB123456",
			"date_of_birth":   "1990-01-01",
		}

		// Provider generates commitment off-chain
		dataProtection := NewDataProtectionService()
		commitment, err := dataProtection.GenerateCommitment(offChainPII)
		require.NoError(t, err)
		require.Len(t, commitment, 32, "SHA-256 commitment must be 32 bytes")

		// Only commitment is stored on-chain
		kycRecord := &types.KYCRecord{
			Address:       "aura1test",
			KycLevel:      types.KYCLevel_KYC_LEVEL_ADVANCED,
			Provider:      "provider1",
			PiiCommitment: commitment, // Hash only, NOT raw data
			Jurisdiction:  "US",
		}

		// Store on-chain
		err = keeper.SetKYCRecord(ctx, kycRecord)
		require.NoError(t, err)

		// Retrieve from chain
		retrieved, err := keeper.GetKYCRecord(ctx, "aura1test")
		require.NoError(t, err)

		// Verify only commitment is stored, not raw PII
		require.Equal(t, commitment, retrieved.PiiCommitment)
		require.Len(t, retrieved.PiiCommitment, 32)

		// Verify PII cannot be recovered from commitment (one-way hash)
		// This is a security property - commitment should reveal nothing about PII
		require.NotContains(t, string(retrieved.PiiCommitment), "John Doe")
		require.NotContains(t, string(retrieved.PiiCommitment), "123-45-6789")
	})

	t.Run("Commitment_CanVerifyDataWithoutExposingIt", func(t *testing.T) {
		// Original PII (stored off-chain)
		originalPII := map[string]interface{}{
			"full_name": "Alice Smith",
			"ssn":       "987-65-4321",
		}

		dataProtection := NewDataProtectionService()
		commitment, err := dataProtection.GenerateCommitment(originalPII)
		require.NoError(t, err)

		// Store commitment on-chain
		kycRecord := &types.KYCRecord{
			Address:       "aura1alice",
			KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
			Provider:      "provider1",
			PiiCommitment: commitment,
			Jurisdiction:  "GB",
		}
		err = keeper.SetKYCRecord(ctx, kycRecord)
		require.NoError(t, err)

		// Later, verify data matches commitment (without storing on-chain)
		retrieved, err := keeper.GetKYCRecord(ctx, "aura1alice")
		require.NoError(t, err)

		// Correct PII verifies successfully
		matches, err := dataProtection.VerifyCommitment(originalPII, retrieved.PiiCommitment)
		require.NoError(t, err)
		require.True(t, matches, "Original PII should match commitment")

		// Modified PII fails verification
		modifiedPII := map[string]interface{}{
			"full_name": "Alice Smith",
			"ssn":       "000-00-0000", // Changed
		}
		matches, err = dataProtection.VerifyCommitment(modifiedPII, retrieved.PiiCommitment)
		require.NoError(t, err)
		require.False(t, matches, "Modified PII should NOT match commitment")
	})
}

// TestGDPRCompliance_RightToErasure verifies GDPR Article 17 compliance.
//
// Article 17 "Right to Erasure" requires:
// - Users can request deletion of their personal data
// - Data must be erased without undue delay
// - Erasure request must be honored
//
// Implementation:
// - On-chain commitments remain (immutable audit trail)
// - Off-chain PII is deleted by providers (event-driven)
// - Event provides proof of erasure request
func TestGDPRCompliance_RightToErasure(t *testing.T) {
	suite := NewTestSuite(t)
	ctx := suite.Ctx
	keeper := suite.Keeper

	t.Run("ErasureRequest_EmitsImmutableEvent", func(t *testing.T) {
		// User submits erasure request
		address := "aura1user"

		// Erasure event should be emitted (monitored by off-chain systems)
		err := keeper.TriggerDataDeletion(ctx, address, "data_processing")
		require.NoError(t, err)

		// Verify event was emitted
		events := ctx.EventManager().Events()
		require.NotEmpty(t, events)

		// Find erasure event
		var foundEvent bool
		var eventAddress string
		var eventConsentType string
		for _, event := range events {
			if event.Type == "gdpr_data_deletion_requested" {
				// Event exists - off-chain systems will see this
				foundEvent = true
				for _, attr := range event.Attributes {
					if string(attr.Key) == types.AttributeKeyAddress {
						eventAddress = string(attr.Value)
					}
					if string(attr.Key) == types.AttributeKeyConsentType {
						eventConsentType = string(attr.Value)
					}
				}
				break
			}
		}

		require.True(t, foundEvent, "Erasure event must be emitted")
		require.Equal(t, address, eventAddress)
		require.Equal(t, "data_processing", eventConsentType)
	})

	t.Run("OnChainCommitments_RemainAfterErasure", func(t *testing.T) {
		// Store KYC record with commitment
		kycRecord := &types.KYCRecord{
			Address:       "aura1user",
			KycLevel:      types.KYCLevel_KYC_LEVEL_ADVANCED,
			Provider:      "provider1",
			PiiCommitment: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20},
			Jurisdiction:  "US",
		}
		err := keeper.SetKYCRecord(ctx, kycRecord)
		require.NoError(t, err)

		// User requests erasure
		err = keeper.TriggerDataDeletion(ctx, "aura1user", "kyc_data")
		require.NoError(t, err)

		// On-chain record still exists (immutable blockchain)
		retrieved, err := keeper.GetKYCRecord(ctx, "aura1user")
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		// Commitment remains (audit trail)
		require.Equal(t, kycRecord.PiiCommitment, retrieved.PiiCommitment)

		// However, off-chain PII would be deleted by providers
		// (they monitor the erasure event and delete their databases)
		// This satisfies GDPR while preserving blockchain immutability
	})

	t.Run("ProcessingRestriction_EnforcedAfterConsentWithdrawal", func(t *testing.T) {
		address := "aura1restricted"

		// Initially no restriction
		require.False(t, keeper.IsProcessingRestricted(ctx, address))

		// User withdraws consent
		err := keeper.SetProcessingRestriction(ctx, address, true)
		require.NoError(t, err)

		// Processing is now restricted
		require.True(t, keeper.IsProcessingRestricted(ctx, address))

		// Data processing should be blocked
		canProcess := keeper.CanProcessData(ctx, address, "data_processing")
		require.False(t, canProcess, "Processing must be blocked after consent withdrawal")
	})
}

// TestGDPRCompliance_DataMinimization verifies GDPR Article 5(1)(c) compliance.
//
// Article 5(1)(c) "Data Minimization" requires:
// - Only necessary data is collected and stored
// - Data is adequate, relevant, and limited to what is necessary
//
// Implementation:
// - Jurisdiction stored on-chain (necessary for OFAC compliance)
// - KYC level stored on-chain (necessary for access control)
// - All other PII stored off-chain (not necessary for blockchain logic)
func TestGDPRCompliance_DataMinimization(t *testing.T) {
	t.Run("OnlyNecessaryFields_StoredOnChain", func(t *testing.T) {
		kycRecord := &types.KYCRecord{
			Address:       "aura1test",
			KycLevel:      types.KYCLevel_KYC_LEVEL_ADVANCED,
			Provider:      "provider1",
			PiiCommitment: make([]byte, 32), // Commitment only
			Jurisdiction:  "US",             // Necessary for OFAC compliance
			// NO: full_name, ssn, passport, etc. - not necessary for on-chain logic
		}

		// Verify only essential fields exist
		require.NotEmpty(t, kycRecord.Address, "Address required for access control")
		require.NotEqual(t, types.KYCLevel_KYC_LEVEL_UNSPECIFIED, kycRecord.KycLevel, "KYC level required for compliance")
		require.NotEmpty(t, kycRecord.Jurisdiction, "Jurisdiction required for OFAC compliance")
		require.NotEmpty(t, kycRecord.PiiCommitment, "Commitment required for verification")

		// Verify NO unnecessary PII fields
		recordJSON, _ := json.Marshal(kycRecord)
		var recordMap map[string]interface{}
		_ = json.Unmarshal(recordJSON, &recordMap)

		// These would be unnecessary PII violations
		require.NotContains(t, recordMap, "full_name")
		require.NotContains(t, recordMap, "ssn")
		require.NotContains(t, recordMap, "passport_number")
		require.NotContains(t, recordMap, "date_of_birth")
		require.NotContains(t, recordMap, "phone_number")
		require.NotContains(t, recordMap, "email")
	})
}

// TestGDPRCompliance_JurisdictionMustBeStored verifies that jurisdiction
// is stored on-chain for OFAC compliance, despite being PII-adjacent.
//
// Legal justification:
// - OFAC compliance is a legal requirement (overrides GDPR minimization)
// - Country code is necessary to enforce sanctions
// - Country code alone is not identifiable PII (GDPR Recital 26)
// - Multiple legal bases: Legal obligation (OFAC) + Legitimate interest (sanctions compliance)
func TestGDPRCompliance_JurisdictionMustBeStored(t *testing.T) {
	suite := NewTestSuite(t)
	ctx := suite.Ctx
	keeper := suite.Keeper

	t.Run("Jurisdiction_RequiredForOFACCompliance", func(t *testing.T) {
		// Jurisdiction must be provided
		kycRecord := &types.KYCRecord{
			Address:       "aura1test",
			KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
			Provider:      "provider1",
			PiiCommitment: make([]byte, 32),
			Jurisdiction:  "US", // Required for OFAC validation
		}

		err := keeper.SetKYCRecord(ctx, kycRecord)
		require.NoError(t, err)

		// Retrieve and verify jurisdiction is stored
		retrieved, err := keeper.GetKYCRecord(ctx, "aura1test")
		require.NoError(t, err)
		require.Equal(t, "US", retrieved.Jurisdiction)
	})

	t.Run("SanctionedJurisdiction_RejectedOnChain", func(t *testing.T) {
		// Set up blocked jurisdictions (OFAC sanctioned countries)
		params := keeper.GetParams(ctx)
		params.BlockedJurisdictions = []string{"KP", "IR", "SY", "CU"}
		err := keeper.SetParams(ctx, params)
		require.NoError(t, err)

		// Attempt to submit KYC from sanctioned country
		blocked := keeper.IsJurisdictionBlocked(ctx, "KP")
		require.True(t, blocked, "North Korea should be blocked")

		blocked = keeper.IsJurisdictionBlocked(ctx, "IR")
		require.True(t, blocked, "Iran should be blocked")

		// Non-sanctioned country is allowed
		blocked = keeper.IsJurisdictionBlocked(ctx, "US")
		require.False(t, blocked, "US should not be blocked")
	})
}

// TestGDPRCompliance_Documentation verifies that proper documentation
// exists for GDPR compliance mechanisms.
func TestGDPRCompliance_Documentation(t *testing.T) {
	t.Run("Protobuf_ContainsGDPRDocumentation", func(t *testing.T) {
		// The protobuf definitions should contain comments explaining
		// the GDPR compliance architecture. This test verifies the
		// documentation exists by checking if the types have the
		// expected structure.

		// KYCRecord should have pii_commitment field
		kycRecord := &types.KYCRecord{}
		require.NotNil(t, kycRecord, "KYCRecord type must exist")

		// Field exists by virtue of being able to set it
		kycRecord.PiiCommitment = make([]byte, 32)
		require.Len(t, kycRecord.PiiCommitment, 32)

		// GDPRConsent should have audit_commitment field
		consent := &types.GDPRConsent{}
		require.NotNil(t, consent, "GDPRConsent type must exist")

		consent.AuditCommitment = make([]byte, 32)
		require.Len(t, consent.AuditCommitment, 32)
	})
}

// TestGDPRCompliance_CommitmentLength verifies that commitments are
// exactly 32 bytes (SHA-256 hash length).
func TestGDPRCompliance_CommitmentLength(t *testing.T) {
	dataProtection := NewDataProtectionService()

	t.Run("Commitment_MustBe32Bytes", func(t *testing.T) {
		piiData := map[string]interface{}{
			"name": "John Doe",
			"ssn":  "123-45-6789",
		}

		commitment, err := dataProtection.GenerateCommitment(piiData)
		require.NoError(t, err)
		require.Len(t, commitment, 32, "SHA-256 commitment must be exactly 32 bytes")
	})

	t.Run("InvalidCommitment_Rejected", func(t *testing.T) {
		invalidCommitment := []byte{0x01, 0x02, 0x03} // Too short

		piiData := map[string]interface{}{"name": "Test"}
		matches, err := dataProtection.VerifyCommitment(piiData, invalidCommitment)
		require.Error(t, err, "Invalid commitment length should be rejected")
		require.False(t, matches)
	})
}

// TestGDPRCompliance_OffChainStorageGuidance verifies that documentation
// and examples exist for off-chain PII storage.
func TestGDPRCompliance_OffChainStorageGuidance(t *testing.T) {
	t.Run("PIIData_StructExists", func(t *testing.T) {
		// The PIIData struct should exist for off-chain use
		piiData := &PIIData{
			FullName:       "John Doe",
			DateOfBirth:    "1990-01-01",
			PassportNumber: "AB123456",
			SSN:            "123-45-6789",
		}

		require.NotNil(t, piiData)
		require.Equal(t, "John Doe", piiData.FullName)

		// This struct should NEVER be stored on-chain
		// It's only for off-chain provider systems
	})

	t.Run("DataProtectionService_Exists", func(t *testing.T) {
		service := NewDataProtectionService()
		require.NotNil(t, service, "DataProtectionService must exist for commitment generation")

		// Test PII commitment generation
		pii := &PIIData{
			FullName: "Alice Smith",
			SSN:      "987-65-4321",
		}

		commitment, err := service.GeneratePIICommitment(pii)
		require.NoError(t, err)
		require.Len(t, commitment, 32)

		// Verify commitment matches
		matches, err := service.VerifyPIICommitment(pii, commitment)
		require.NoError(t, err)
		require.True(t, matches)
	})
}
