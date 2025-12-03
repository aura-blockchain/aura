# GDPR-Compliant PII Storage Architecture

## Overview

The Aura Identity Module implements a commitment-based storage pattern that enables **GDPR Right to Erasure** compliance while maintaining blockchain immutability and audit trail integrity.

## Problem Statement

**Challenge**: Traditional blockchain storage is immutable, but GDPR mandates the ability to erase personal data (Right to Erasure, Article 17).

**Solution**: Store only cryptographic commitments on-chain, not raw PII data.

## Architecture

### On-Chain Storage (Blockchain)

Stored in `IdentityRecord` proto message:

```protobuf
message IdentityRecord {
  string did = 1;                           // DID identifier (public)
  string address = 2;                       // Blockchain address (public)
  IdentityStatus status = 3;                // Active, Erased, etc.
  bytes pii_commitment = 12;                // SHA-256(sorted_pii || salt)
  bytes commitment_salt = 13;               // Random 32-byte salt
  bool erased = 14;                         // GDPR erasure flag
  google.protobuf.Timestamp erased_at = 15; // Erasure timestamp
  string off_chain_data_ref = 16;           // Reference to off-chain storage
  string off_chain_data_type = 17;          // Storage type (ipfs, https, etc.)
}
```

**What is NOT stored on-chain:**
- Names, emails, phone numbers
- Biometric data
- Physical addresses
- Any personally identifiable information

### Off-Chain Storage

Actual PII data stored in:
- **IPFS**: Decentralized, content-addressed storage
- **Secure Database**: Encrypted database with access controls
- **User Device**: Encrypted local storage (wallet)
- **Trusted Third Party**: Encrypted cloud storage

## Cryptographic Commitment Scheme

### 1. Commitment Generation

```go
// Algorithm
func ComputePIICommitment(piiData map[string]string, salt []byte) []byte {
    // Step 1: Sort keys alphabetically (deterministic)
    keys := sort(piiData.keys())

    // Step 2: Serialize as: key1=value1||key2=value2||...
    serialized := ""
    for key in keys {
        serialized += key + "=" + piiData[key] + "||"
    }

    // Step 3: Append salt
    data := serialized + salt

    // Step 4: Hash with SHA-256
    return SHA256(data)
}
```

### 2. Security Properties

**Hiding**: Commitment reveals nothing about PII without the salt
- Given: `commitment = SHA256(pii || salt)`
- Attacker cannot derive `pii` from `commitment` alone

**Binding**: Cannot change PII while keeping same commitment
- Different PII → Different commitment (collision resistance)

**Verifiable**: Data owner can prove they have correct PII
- Recompute commitment from PII → Compare with stored commitment

**Salt Purpose**:
- Prevents rainbow table attacks
- Makes identical PII produce different commitments across identities
- Stored on-chain (needed for verification)

## Usage Examples

### Creating an Identity

```go
// 1. User provides PII
piiData := map[string]string{
    "name":  "Alice Smith",
    "email": "alice@example.com",
    "dob":   "1990-01-01",
}

// 2. Generate salt and commitment
salt := types.GenerateCommitmentSalt()
commitment := types.ComputePIICommitment(piiData, salt)

// 3. Store PII off-chain (IPFS, database, etc.)
offChainRef := StoreToIPFS(piiData) // Returns IPFS CID

// 4. Store commitment on-chain
record := &types.IdentityRecord{
    Did:              "did:aura:alice123",
    Address:          userAddress,
    Status:           types.IdentityStatusActive,
    PiiCommitment:    commitment,
    CommitmentSalt:   salt,
    OffChainDataRef:  offChainRef,
    OffChainDataType: "ipfs",
}

keeper.SetIdentityRecord(ctx, record)
```

### Verifying Identity

```go
// 1. User provides PII to prove identity
providedPII := map[string]string{
    "name":  "Alice Smith",
    "email": "alice@example.com",
    "dob":   "1990-01-01",
}

// 2. Retrieve commitment from blockchain
record, _ := keeper.GetIdentityRecord(ctx, "did:aura:alice123")

// 3. Verify commitment
valid, _ := keeper.VerifyPIICommitment(ctx, "did:aura:alice123", providedPII)

if valid {
    // Identity verified!
}
```

### GDPR Erasure

```go
// 1. User requests erasure
err := keeper.EraseIdentity(ctx, "did:aura:alice123", requester, "GDPR request")

// 2. Off-chain system deletes PII data
DeleteFromIPFS(offChainRef)

// 3. On-chain record updated:
//    - erased = true
//    - status = ERASED
//    - off_chain_data_ref = "" (cleared)
//    - commitment preserved (audit trail)

// 4. Future verification attempts fail:
valid, err := keeper.VerifyPIICommitment(ctx, "did:aura:alice123", piiData)
// Returns: ErrIdentityErased
```

## GDPR Compliance

### Right to Erasure (Article 17)

**Before Erasure**:
- On-chain: Commitment + salt + reference
- Off-chain: Full PII data

**After Erasure**:
- On-chain: Commitment + salt (audit) + `erased=true`
- Off-chain: PII data deleted

**Why This Works**:
- Commitment alone cannot reconstruct PII
- Off-chain PII is permanently deleted
- Audit trail preserved (commitment exists but reveals nothing)

### Right to Access (Article 15)

User proves identity via commitment → Retrieves PII from off-chain storage

### Right to Rectification (Article 16)

```go
// Update PII off-chain
newPII := map[string]string{
    "name":  "Alice Johnson", // Changed name
    "email": "alice.johnson@example.com",
    "dob":   "1990-01-01",
}

// Compute new commitment
keeper.UpdatePIICommitment(ctx, did, updater, newPII, newOffChainRef, "ipfs")
```

### Right to Data Portability (Article 20)

User exports PII from off-chain storage

### Data Minimization (Article 5)

Only necessary data (commitment) stored on-chain

## Security Considerations

### Attack Vectors Mitigated

**1. Rainbow Table Attack**
- **Threat**: Pre-compute hashes of common names/emails
- **Mitigation**: Random salt per identity

**2. Brute Force Attack**
- **Threat**: Try all possible PII combinations
- **Mitigation**: Large search space (names × emails × dobs × ...)

**3. Commitment Substitution**
- **Threat**: Replace commitment with attacker's commitment
- **Mitigation**: Blockchain immutability + access control

**4. Off-Chain Data Loss**
- **Threat**: Off-chain storage fails
- **Mitigation**: Redundancy, backups, distributed storage (IPFS)

### Best Practices

**Salt Generation**:
- Use cryptographically secure random number generator
- 32 bytes minimum (256 bits)
- One salt per identity (never reuse)

**Off-Chain Storage**:
- Encrypt PII at rest
- Access control (only identity owner)
- Redundant backups
- Secure deletion (not just file deletion, but secure wipe)

**Commitment Verification**:
- Rate limit verification attempts
- Log all verification attempts (audit)
- Fail securely (don't leak information on error)

## Implementation Details

### Files

**Proto Definitions**:
- `/proto/aura/identity/v1beta1/identity.proto` - IdentityRecord message

**Keeper Methods**:
- `/chain/x/identity/keeper/changes.go`:
  - `EraseIdentity()` - GDPR erasure
  - `VerifyPIICommitment()` - Verify PII
  - `UpdatePIICommitment()` - Update PII

**Utility Functions**:
- `/chain/x/identity/types/pii_commitment.go`:
  - `GenerateCommitmentSalt()` - Generate salt
  - `ComputePIICommitment()` - Compute commitment
  - `VerifyPIICommitment()` - Verify commitment

**Tests**:
- `/chain/x/identity/types/pii_commitment_test.go` - Commitment tests
- `/chain/x/identity/keeper/erasure_test.go` - Keeper erasure tests

### Error Codes

```go
ErrIdentityAlreadyErased = 650  // Identity already erased
ErrIdentityErased        = 651  // Identity has been erased
ErrNoCommitment          = 652  // No PII commitment found
ErrInvalidCommitment     = 653  // Invalid PII commitment
ErrUnauthorized          = 654  // Unauthorized action
```

## Migration from Raw PII Storage

If you have existing identities with raw PII on-chain:

```go
// 1. Read existing identity
record, _ := keeper.GetIdentityRecord(ctx, did)
oldPII := record.Attributes // Old raw PII storage

// 2. Generate commitment
salt := types.GenerateCommitmentSalt()
commitment := types.ComputePIICommitment(oldPII, salt)

// 3. Store PII off-chain
offChainRef := StoreToIPFS(oldPII)

// 4. Update record
record.PiiCommitment = commitment
record.CommitmentSalt = salt
record.OffChainDataRef = offChainRef
record.OffChainDataType = "ipfs"
record.Attributes = nil // Clear old storage

keeper.SetIdentityRecord(ctx, record)
```

## Performance Characteristics

**Commitment Generation**: ~0.1ms
- SHA-256 is fast (~300 MB/s)
- Negligible overhead for identity operations

**Verification**: ~0.1ms
- Same as generation (recompute + compare)

**Storage**:
- On-chain: 64 bytes (32-byte commitment + 32-byte salt)
- Off-chain: Arbitrary size

## Future Enhancements

### Zero-Knowledge Proofs

Enable proving properties without revealing PII:

```go
// Prove "age > 18" without revealing birthdate
zkProof := GenerateAgeProof(piiData["dob"], commitment, salt)
verified := VerifyAgeProof(zkProof, commitment, 18)
```

### Threshold Cryptography

Split salt across multiple parties (no single party can verify):

```go
// 2-of-3 threshold: need 2 parties to reconstruct salt
shares := SplitSalt(salt, 2, 3)
// Distribute shares to: identity owner, trusted party, blockchain
```

### Homomorphic Encryption

Compute on encrypted PII without decryption:

```go
// Encrypt PII with homomorphic encryption
encrypted := HomomorphicEncrypt(piiData)
// Compute commitment on encrypted data
commitment := ComputeCommitmentHomomorphic(encrypted, salt)
```

## References

- **GDPR**: [Regulation (EU) 2016/679](https://gdpr-info.eu/)
- **Commitment Schemes**: [Wikipedia](https://en.wikipedia.org/wiki/Commitment_scheme)
- **SHA-256**: [NIST FIPS 180-4](https://csrc.nist.gov/publications/detail/fips/180/4/final)
- **IPFS**: [InterPlanetary File System](https://ipfs.io/)

## Support

For questions or issues:
- Open GitHub issue: https://github.com/aequitas/aura/issues
- Security issues: security@aura.network (do not open public issues)
