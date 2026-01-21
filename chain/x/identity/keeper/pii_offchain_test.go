// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/identity/types"
)

// mustGenerateSalt is a test helper that generates salt and fails test on error
func mustGenerateSalt(t testing.TB) []byte {
	t.Helper()
	salt, err := types.GenerateCommitmentSalt()
	if err != nil {
		t.Fatalf("GenerateCommitmentSalt failed: %v", err)
	}
	return salt
}

// TestPIIOffChain_OnlyCommitmentsStored verifies that only commitments are stored on-chain
func TestPIIOffChain_OnlyCommitmentsStored(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Simulate user providing PII
	did := "did:aura:user123"
	address := "aura1user"
	piiData := map[string]string{
		"name":           "Alice Smith",
		"email":          "alice.smith@example.com",
		"phone":          "+1-555-0123",
		"date_of_birth":  "1990-01-15",
		"ssn":            "123-45-6789",
		"biometric_hash": "sha256:deadbeef...",
		"address":        "123 Main St, City, State 12345",
	}

	// Generate commitment
	salt := mustGenerateSalt(t)
	commitment := types.ComputePIICommitment(piiData, salt)

	// Store identity with only commitment (PII stored off-chain)
	record := &types.IdentityRecord{
		Did:              did,
		Address:          address,
		Status:           types.IdentityStatusActive,
		CreatedAt:        ctx.BlockTime(),
		UpdatedAt:        func() *time.Time { t := ctx.BlockTime(); return &t }(),
		PiiCommitment:    commitment,
		CommitmentSalt:   salt,
		OffChainDataRef:  "ipfs://QmTestHash123456789",
		OffChainDataType: "ipfs",
		Erased:           false,
	}

	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Retrieve record
	stored, err := keeper.GetIdentityRecord(ctx, did)
	require.NoError(t, err)

	// CRITICAL: Verify NO raw PII is stored on-chain
	require.NotContains(t, string(keeper.cdc.MustMarshal(stored)), "Alice Smith", "raw name should not be in on-chain data")
	require.NotContains(t, string(keeper.cdc.MustMarshal(stored)), "alice.smith@example.com", "email should not be in on-chain data")
	require.NotContains(t, string(keeper.cdc.MustMarshal(stored)), "+1-555-0123", "phone should not be in on-chain data")
	require.NotContains(t, string(keeper.cdc.MustMarshal(stored)), "123-45-6789", "SSN should not be in on-chain data")
	require.NotContains(t, string(keeper.cdc.MustMarshal(stored)), "123 Main St", "address should not be in on-chain data")

	// Verify only safe data is stored
	require.Equal(t, did, stored.Did, "DID should be stored")
	require.Equal(t, address, stored.Address, "blockchain address should be stored")
	require.NotEmpty(t, stored.PiiCommitment, "commitment should be stored")
	require.NotEmpty(t, stored.CommitmentSalt, "salt should be stored")
	require.Equal(t, "ipfs://QmTestHash123456789", stored.OffChainDataRef, "off-chain reference should be stored")
}

// TestPIIOffChain_VerificationWorksWithCommitments tests PII verification using commitments
func TestPIIOffChain_VerificationWorksWithCommitments(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	did := "did:aura:user456"
	address := "aura1user456"

	// Original PII (would be stored off-chain)
	piiData := map[string]string{
		"name":  "Bob Johnson",
		"email": "bob@example.com",
		"dob":   "1985-03-20",
	}

	// Create commitment
	salt := mustGenerateSalt(t)
	commitment := types.ComputePIICommitment(piiData, salt)

	// Store identity with commitment only
	record := &types.IdentityRecord{
		Did:              did,
		Address:          address,
		Status:           types.IdentityStatusActive,
		CreatedAt:        ctx.BlockTime(),
		PiiCommitment:    commitment,
		CommitmentSalt:   salt,
		OffChainDataRef:  "ipfs://QmBobData",
		OffChainDataType: "ipfs",
	}

	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// User presents PII for verification (retrieved from off-chain storage)
	presentedPII := map[string]string{
		"name":  "Bob Johnson",
		"email": "bob@example.com",
		"dob":   "1985-03-20",
	}

	// Verify PII matches commitment
	valid, err := keeper.VerifyPIICommitment(ctx, did, presentedPII)
	require.NoError(t, err)
	require.True(t, valid, "valid PII should verify against commitment")

	// Verify wrong data fails
	wrongPII := map[string]string{
		"name":  "Bob Johnson",
		"email": "bob@example.com",
		"dob":   "1985-03-21", // Wrong date
	}

	valid, err = keeper.VerifyPIICommitment(ctx, did, wrongPII)
	require.NoError(t, err)
	require.False(t, valid, "invalid PII should fail verification")
}

// TestPIIOffChain_ErasureCompliance tests GDPR Right to Erasure
func TestPIIOffChain_ErasureCompliance(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	did := "did:aura:erasure123"
	address := "aura1erasure"
	piiData := map[string]string{
		"name":  "Carol Williams",
		"email": "carol@example.com",
		"phone": "+1-555-9999",
	}

	salt := mustGenerateSalt(t)
	commitment := types.ComputePIICommitment(piiData, salt)

	// Create identity
	record := &types.IdentityRecord{
		Did:              did,
		Address:          address,
		Status:           types.IdentityStatusActive,
		CreatedAt:        ctx.BlockTime(),
		PiiCommitment:    commitment,
		CommitmentSalt:   salt,
		OffChainDataRef:  "ipfs://QmCarolData",
		OffChainDataType: "ipfs",
		Erased:           false,
	}

	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// User exercises GDPR Right to Erasure
	err = keeper.EraseIdentity(ctx, did, address, "GDPR Article 17 - Right to Erasure request")
	require.NoError(t, err)

	// Verify erasure
	erased, err := keeper.GetIdentityRecord(ctx, did)
	require.NoError(t, err)

	// GDPR Compliance checks
	require.True(t, erased.Erased, "identity must be marked as erased")
	require.Equal(t, types.IdentityStatusErased, erased.Status, "status must be ERASED")
	require.NotNil(t, erased.ErasedAt, "erasure timestamp must be recorded")
	require.Empty(t, erased.OffChainDataRef, "off-chain reference must be cleared")
	require.Empty(t, erased.OffChainDataType, "off-chain type must be cleared")

	// Audit trail preservation (for regulatory compliance)
	require.Equal(t, did, erased.Did, "DID preserved for audit trail")
	require.Equal(t, address, erased.Address, "address preserved for audit trail")
	require.NotEmpty(t, erased.PiiCommitment, "commitment preserved for audit trail")
	require.NotEmpty(t, erased.CommitmentSalt, "salt preserved for audit trail")

	// CRITICAL: Commitment reveals nothing about PII
	require.NotContains(t, string(erased.PiiCommitment), "Carol", "commitment must not reveal PII")
	require.NotContains(t, string(erased.PiiCommitment), "carol@example.com", "commitment must not reveal PII")

	// Verify cannot use identity after erasure
	valid, err := keeper.VerifyPIICommitment(ctx, did, piiData)
	require.Error(t, err, "verification should fail for erased identity")
	require.ErrorIs(t, err, types.ErrIdentityErased)
	require.False(t, valid)

	// Verify cannot update after erasure
	newSalt := mustGenerateSalt(t)
	err = keeper.UpdatePIICommitment(ctx, did, address, newSalt, "ipfs://QmNew", "ipfs")
	require.Error(t, err, "update should fail for erased identity")
	require.ErrorIs(t, err, types.ErrIdentityErased)
}

// TestPIIOffChain_DataCannotBeRecovered tests that PII cannot be recovered from commitment
func TestPIIOffChain_DataCannotBeRecovered(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create identity with sensitive PII
	did := "did:aura:recovery123"
	address := "aura1recovery"
	sensitivePII := map[string]string{
		"name":           "David Secret",
		"ssn":            "987-65-4321",
		"credit_card":    "4111-1111-1111-1111",
		"medical_record": "Patient has condition XYZ",
		"biometric":      "fingerprint:abc123def456",
	}

	salt := mustGenerateSalt(t)
	commitment := types.ComputePIICommitment(sensitivePII, salt)

	record := &types.IdentityRecord{
		Did:            did,
		Address:        address,
		Status:         types.IdentityStatusActive,
		CreatedAt:      ctx.BlockTime(),
		PiiCommitment:  commitment,
		CommitmentSalt: salt,
	}

	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Erase identity (off-chain PII deleted)
	err = keeper.EraseIdentity(ctx, did, address, "GDPR erasure")
	require.NoError(t, err)

	// Retrieve erased record
	erased, err := keeper.GetIdentityRecord(ctx, did)
	require.NoError(t, err)

	// CRITICAL SECURITY TEST: Verify PII cannot be recovered from on-chain data
	// Even with commitment and salt, original data cannot be derived
	marshaledRecord := keeper.cdc.MustMarshal(erased)

	// No sensitive data in serialized form
	require.NotContains(t, string(marshaledRecord), "David Secret")
	require.NotContains(t, string(marshaledRecord), "987-65-4321")
	require.NotContains(t, string(marshaledRecord), "4111-1111-1111-1111")
	require.NotContains(t, string(marshaledRecord), "condition XYZ")
	require.NotContains(t, string(marshaledRecord), "abc123def456")

	// Commitment is one-way - cannot reverse
	// This property is guaranteed by SHA-256
	require.Len(t, erased.PiiCommitment, 32, "commitment is 32-byte hash")
	require.Len(t, erased.CommitmentSalt, 32, "salt is 32 bytes")

	// Even with salt, cannot recover original data (one-way function)
	// The only way to verify is to have the original data and compute commitment
}

// TestPIIOffChain_MultipleAttributeChanges tests commitment updates
func TestPIIOffChain_MultipleAttributeChanges(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	did := "did:aura:update123"
	address := "aura1update"

	// Initial PII
	pii1 := map[string]string{
		"name":  "Eve Adams",
		"email": "eve@example.com",
		"city":  "Boston",
	}
	salt1 := mustGenerateSalt(t)
	commitment1 := types.ComputePIICommitment(pii1, salt1)

	record := &types.IdentityRecord{
		Did:              did,
		Address:          address,
		Status:           types.IdentityStatusActive,
		CreatedAt:        ctx.BlockTime(),
		PiiCommitment:    commitment1,
		CommitmentSalt:   salt1,
		OffChainDataRef:  "ipfs://QmV1",
		OffChainDataType: "ipfs",
	}

	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Verify initial data
	valid, err := keeper.VerifyPIICommitment(ctx, did, pii1)
	require.NoError(t, err)
	require.True(t, valid)

	// User moves to new city and changes email (PII updated off-chain)
	pii2 := map[string]string{
		"name":  "Eve Adams",
		"email": "eve.adams@newdomain.com",
		"city":  "San Francisco",
	}
	salt2 := mustGenerateSalt(t)
	commitment2 := types.ComputePIICommitment(pii2, salt2)

	// Update commitment
	err = keeper.UpdatePIICommitment(ctx, did, address, salt2, "ipfs://QmV2", "ipfs")
	require.NoError(t, err)

	// Update the commitment field manually
	updated, err := keeper.GetIdentityRecord(ctx, did)
	require.NoError(t, err)
	updated.PiiCommitment = commitment2
	err = keeper.SetIdentityRecord(ctx, updated)
	require.NoError(t, err)

	// Verify new data works
	valid, err = keeper.VerifyPIICommitment(ctx, did, pii2)
	require.NoError(t, err)
	require.True(t, valid, "new PII should verify")

	// Verify old data no longer works
	valid, err = keeper.VerifyPIICommitment(ctx, did, pii1)
	require.NoError(t, err)
	require.False(t, valid, "old PII should not verify after update")

	// Verify new off-chain reference (reuse updated variable from above)
	require.Equal(t, "ipfs://QmV2", updated.OffChainDataRef)
}

// TestPIIOffChain_UnauthorizedAccess tests access control for PII operations
func TestPIIOffChain_UnauthorizedAccess(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	did := "did:aura:auth123"
	owner := "aura1owner"
	attacker := "aura1attacker"

	piiData := map[string]string{
		"name":  "Frank Miller",
		"email": "frank@example.com",
	}
	salt := mustGenerateSalt(t)
	commitment := types.ComputePIICommitment(piiData, salt)

	record := &types.IdentityRecord{
		Did:            did,
		Address:        owner,
		Status:         types.IdentityStatusActive,
		CreatedAt:      ctx.BlockTime(),
		PiiCommitment:  commitment,
		CommitmentSalt: salt,
	}

	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Attacker tries to update PII commitment
	attackerSalt := mustGenerateSalt(t)

	err = keeper.UpdatePIICommitment(ctx, did, attacker, attackerSalt, "ipfs://QmEvil", "ipfs")
	require.Error(t, err, "unauthorized update should fail")
	require.ErrorIs(t, err, types.ErrUnauthorized)

	// Verify original commitment unchanged
	valid, err := keeper.VerifyPIICommitment(ctx, did, piiData)
	require.NoError(t, err)
	require.True(t, valid, "original PII should still verify")

	// Attacker tries to erase identity
	err = keeper.EraseIdentity(ctx, did, attacker, "malicious erasure")
	require.Error(t, err, "unauthorized erasure should fail")
	require.ErrorIs(t, err, types.ErrUnauthorized)

	// Verify identity not erased
	stored, err := keeper.GetIdentityRecord(ctx, did)
	require.NoError(t, err)
	require.False(t, stored.Erased, "identity should not be erased by unauthorized user")
}

// TestPIIOffChain_CommitmentCollisionResistance tests cryptographic security
func TestPIIOffChain_CommitmentCollisionResistance(t *testing.T) {
	// Generate many different PII records and ensure no commitment collisions
	commitments := make(map[string]bool)
	salt := mustGenerateSalt(t)

	for i := 0; i < 10000; i++ {
		pii := map[string]string{
			"name":  "User" + string(rune(i)),
			"id":    string(rune(i % 256)),
			"index": string(rune(i / 256)),
		}
		commitment := types.ComputePIICommitment(pii, salt)
		commitmentStr := string(commitment)

		// Ensure no collision
		require.False(t, commitments[commitmentStr], "commitment collision detected at iteration %d", i)
		commitments[commitmentStr] = true
	}

	require.Len(t, commitments, 10000, "all commitments should be unique")
}

// TestPIIOffChain_ProtobufFieldsCompliance tests that protobuf has no PII fields
func TestPIIOffChain_ProtobufFieldsCompliance(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	did := "did:aura:proto123"
	address := "aura1proto"

	// Create identity with all fields populated
	record := &types.IdentityRecord{
		Did:                 did,
		Address:             address,
		Status:              types.IdentityStatusActive,
		CreatedAt:           ctx.BlockTime(),
		UpdatedAt:           func() *time.Time { t := ctx.BlockTime(); return &t }(),
		VerificationMethods: []string{"key1", "key2"},
		ConfidenceScore:     95,
		MetadataHash:        "hash123",
		LatestIrVersion:     "v1.0",
		LastChangedHeight:   ctx.BlockHeight(),
		PiiCommitment:       []byte("commitment32bytes0000000000000"),
		CommitmentSalt:      []byte("salt32bytes000000000000000000"),
		Erased:              false,
		OffChainDataRef:     "ipfs://QmTest",
		OffChainDataType:    "ipfs",
	}

	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Serialize to protobuf
	serialized := keeper.cdc.MustMarshal(record)

	// CRITICAL: Verify no PII field names or sensitive data in protobuf
	// These fields should NOT exist in the protobuf definition
	forbiddenFields := []string{
		"name", "email", "phone", "ssn",
		"date_of_birth", "dob", "biometric_hash",
		"physical_address", "credit_card",
		"passport", "drivers_license",
	}

	for _, field := range forbiddenFields {
		require.NotContains(t, string(serialized), field,
			"protobuf should not contain PII field: %s", field)
	}

	// Verify safe fields are present
	require.Contains(t, string(serialized), did, "DID should be in protobuf")
	require.Contains(t, string(serialized), "ipfs://QmTest", "off-chain ref should be in protobuf")
}

// TestPIIOffChain_AuditTrailPreservation tests audit trail after erasure
func TestPIIOffChain_AuditTrailPreservation(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	did := "did:aura:audit123"
	address := "aura1audit"
	piiData := map[string]string{
		"name": "Grace Hopper",
		"role": "Computer Scientist",
	}

	salt := mustGenerateSalt(t)
	commitment := types.ComputePIICommitment(piiData, salt)
	originalCommitment := make([]byte, len(commitment))
	copy(originalCommitment, commitment)

	record := &types.IdentityRecord{
		Did:            did,
		Address:        address,
		Status:         types.IdentityStatusActive,
		CreatedAt:      ctx.BlockTime(),
		PiiCommitment:  commitment,
		CommitmentSalt: salt,
	}

	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Erase identity
	err = keeper.EraseIdentity(ctx, did, address, "User request")
	require.NoError(t, err)

	// Verify audit trail preserved
	erased, err := keeper.GetIdentityRecord(ctx, did)
	require.NoError(t, err)

	// Audit requirements
	require.NotNil(t, erased, "record must exist for audit trail")
	require.True(t, erased.Erased, "erasure flag must be set")
	require.NotNil(t, erased.ErasedAt, "erasure timestamp must be recorded")
	require.NotNil(t, erased.CreatedAt, "creation timestamp must be preserved")
	require.Equal(t, did, erased.Did, "DID must be preserved")
	require.Equal(t, address, erased.Address, "address must be preserved")
	require.Equal(t, originalCommitment, erased.PiiCommitment, "commitment must be preserved")
	require.NotEmpty(t, erased.CommitmentSalt, "salt must be preserved")

	// Commitment still verifies structure (even though data is gone)
	// This proves the record existed and was validly created
	require.Len(t, erased.PiiCommitment, 32, "commitment structure preserved")
}

// TestPIIOffChain_OffChainStorageReferences tests off-chain storage handling
func TestPIIOffChain_OffChainStorageReferences(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	testCases := []struct {
		name        string
		dataRef     string
		dataType    string
		description string
	}{
		{
			name:        "IPFS Storage",
			dataRef:     "ipfs://QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG",
			dataType:    "ipfs",
			description: "Decentralized storage via IPFS",
		},
		{
			name:        "Secure Database",
			dataRef:     "https://secure-vault.example.com/identity/user123",
			dataType:    "https",
			description: "Encrypted database with access controls",
		},
		{
			name:        "DID Document",
			dataRef:     "did:web:example.com:users:alice",
			dataType:    "did-document",
			description: "DID document with service endpoints",
		},
		{
			name:        "Encrypted Cloud Storage",
			dataRef:     "s3://bucket/encrypted/user-data-hash.enc",
			dataType:    "s3",
			description: "Encrypted cloud storage",
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			did := "did:aura:storage" + string(rune(i))
			address := "aura1storage" + string(rune(i))

			piiData := map[string]string{"name": "User" + string(rune(i))}
			salt := mustGenerateSalt(t)
			commitment := types.ComputePIICommitment(piiData, salt)

			record := &types.IdentityRecord{
				Did:              did,
				Address:          address,
				Status:           types.IdentityStatusActive,
				CreatedAt:        ctx.BlockTime(),
				PiiCommitment:    commitment,
				CommitmentSalt:   salt,
				OffChainDataRef:  tc.dataRef,
				OffChainDataType: tc.dataType,
			}

			err := keeper.SetIdentityRecord(ctx, record)
			require.NoError(t, err)

			// Verify storage reference preserved
			stored, err := keeper.GetIdentityRecord(ctx, did)
			require.NoError(t, err)
			require.Equal(t, tc.dataRef, stored.OffChainDataRef)
			require.Equal(t, tc.dataType, stored.OffChainDataType)

			// Verify erasure clears reference
			err = keeper.EraseIdentity(ctx, did, address, "GDPR request")
			require.NoError(t, err)

			erased, err := keeper.GetIdentityRecord(ctx, did)
			require.NoError(t, err)
			require.Empty(t, erased.OffChainDataRef, "ref should be cleared after erasure")
			require.Empty(t, erased.OffChainDataType, "type should be cleared after erasure")
		})
	}
}
