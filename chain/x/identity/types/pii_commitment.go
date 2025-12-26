// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"crypto/rand"
	"crypto/sha256"
	"sort"
	"strings"
)

// GenerateCommitmentSalt generates a cryptographically secure random salt
// for PII commitments. Returns 32 bytes of random data.
//
// CRITICAL: This function MUST NOT be called within blockchain transaction handlers
// or any deterministic consensus code path. The crypto/rand source is non-deterministic
// and will cause different validators to produce different state roots, breaking consensus.
//
// USAGE:
//   - Client-side: Generate salt before submitting transaction, include in message
//   - Off-chain tools: Safe to use for generating commitments
//   - Tests: Safe to use for test data generation
//   - NEVER: In keeper methods, message handlers, or any on-chain code
//
// For on-chain operations, the salt MUST be provided as a parameter from the client.
//
// Returns:
//   - salt: 32-byte cryptographically random salt
//   - error: non-nil if crypto/rand fails (should never happen on modern systems)
func GenerateCommitmentSalt() ([]byte, error) {
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// ComputePIICommitment computes a cryptographic commitment to PII data
//
// DETERMINISM: This function is deterministic and safe for on-chain use.
// However, the salt MUST be generated client-side before transaction submission.
//
// Security properties:
//   - Hiding: Commitment reveals nothing about the data without the salt
//   - Binding: Cannot change data while keeping same commitment
//   - Verifiable: Data owner can prove they have correct data
//
// Algorithm:
//  1. Sort attribute keys alphabetically for deterministic ordering
//  2. Serialize as: key1=value1||key2=value2||...||keyN=valueN
//  3. Append salt: serialized_data || salt
//  4. Hash: SHA-256(data || salt)
//
// Parameters:
//   - piiData: Map of attribute name to value (e.g., {"name": "Alice", "email": "alice@example.com"})
//   - salt: Random 32-byte salt (MUST be generated client-side using crypto/rand)
//
// Returns:
//   - 32-byte SHA-256 hash as commitment
//
// Example (CLIENT-SIDE):
//
//	// Client generates salt before submitting transaction
//	salt := make([]byte, 32)
//	_, err := crypto/rand.Read(salt)
//	if err != nil { panic(err) }
//
//	data := map[string]string{"name": "Alice", "email": "alice@example.com"}
//	commitment := ComputePIICommitment(data, salt)
//	// Include both salt and commitment in transaction message
//	// Store commitment and salt on-chain, store data off-chain
func ComputePIICommitment(piiData map[string]string, salt []byte) []byte {
	// Sort keys for deterministic serialization
	keys := make([]string, 0, len(piiData))
	for k := range piiData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Serialize attributes: key1=value1||key2=value2||...
	var builder strings.Builder
	for i, key := range keys {
		if i > 0 {
			builder.WriteString("||")
		}
		builder.WriteString(key)
		builder.WriteString("=")
		builder.WriteString(piiData[key])
	}

	// Append salt
	serialized := builder.String()
	data := append([]byte(serialized), salt...)

	// Compute SHA-256 hash
	hash := sha256.Sum256(data)
	return hash[:]
}

// VerifyPIICommitment verifies that PII data matches a commitment
//
// Parameters:
//   - piiData: Map of attribute name to value to verify
//   - commitment: Stored commitment (32 bytes)
//   - salt: Salt used in commitment (32 bytes)
//
// Returns:
//   - true if data matches commitment, false otherwise
//
// Example:
//
//	data := map[string]string{"name": "Alice", "email": "alice@example.com"}
//	valid := VerifyPIICommitment(data, storedCommitment, storedSalt)
//	if valid {
//	    // Data is authentic
//	}
func VerifyPIICommitment(piiData map[string]string, commitment []byte, salt []byte) bool {
	computed := ComputePIICommitment(piiData, salt)
	return string(computed) == string(commitment)
}

// PIICommitmentScheme describes the commitment scheme for documentation
const PIICommitmentScheme = `
PII Commitment Scheme (GDPR-compliant)
======================================

Purpose: Store identity data commitments on-chain without exposing raw PII

Security Model:
- On-chain: DID, commitment (hash), salt, metadata
- Off-chain: Actual PII data (name, email, biometrics, etc.)
- Verifiable: Users can prove they have correct PII
- Erasable: Off-chain data can be deleted for GDPR compliance
- Auditable: Commitment remains for audit trail after erasure

CRITICAL - Blockchain Determinism:
- Salt MUST be generated CLIENT-SIDE using crypto/rand
- Salt is included in the transaction message
- NEVER generate salt on-chain (breaks consensus)
- On-chain code only validates and stores the salt
- All validators must see the same salt value

Data Flow:
1. User Registration (CLIENT-SIDE):
   - User provides PII (name, email, etc.)
   - CLIENT generates random salt using crypto/rand
   - CLIENT computes commitment = SHA-256(sorted_pii_attributes || salt)
   - CLIENT stores PII in off-chain storage (IPFS, secure database, etc.)
   - CLIENT submits transaction with commitment + salt + off-chain-reference
   - BLOCKCHAIN validates and stores commitment + salt + metadata

2. PII Verification:
   - User provides PII data for verification
   - System retrieves commitment and salt from blockchain
   - Computes commitment from provided PII
   - Compares: if equal, PII is authentic

3. GDPR Erasure:
   - Delete PII from off-chain storage
   - Mark identity as "erased" on blockchain
   - Commitment remains for audit (but reveals nothing)
   - Cannot reconstruct PII from commitment alone

Commitment Properties:
- Deterministic: Same data + salt always produces same commitment
- One-way: Cannot derive PII from commitment
- Collision-resistant: Computationally infeasible to find different data with same commitment
- Verifiable: Data owner can prove authenticity without revealing data

Implementation Notes:
- Salt prevents rainbow table attacks
- Sorted keys ensure deterministic serialization
- SHA-256 provides 128-bit security level
- Compatible with zero-knowledge proof systems

Example Off-Chain Storage Options:
- IPFS: Decentralized, content-addressed storage
- Secure Database: Encrypted database with access controls
- User Device: Encrypted local storage (wallet)
- Trusted Third Party: Encrypted cloud storage with privacy guarantees

GDPR Compliance:
- Right to Erasure: Delete off-chain PII, mark as erased on-chain
- Right to Access: User proves identity via commitment, retrieves PII
- Right to Portability: User exports PII from off-chain storage
- Right to Rectification: Update PII off-chain, compute new commitment
- Data Minimization: Only necessary data (commitment) stored on-chain
- Purpose Limitation: Commitment used only for identity verification
`
