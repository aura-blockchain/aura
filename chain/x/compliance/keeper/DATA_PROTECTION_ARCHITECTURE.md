# Compliance Data Protection Architecture

## Executive Summary

This document describes the data protection architecture for the Aura blockchain compliance module, designed to meet GDPR Article 32 requirements while maintaining blockchain immutability.

**Key Principle**: Blockchain data is immutable and publicly readable. Therefore, sensitive PII is **NEVER** stored on-chain, even encrypted. Instead, we use cryptographic commitments (SHA-256 hashes) on-chain and secure off-chain storage for actual PII.

## Problem Statement

The compliance module handles highly sensitive data:
- **KYC Records**: Names, birthdates, SSNs, passport numbers, addresses
- **AML Profiles**: Financial information, source of funds, occupation
- **Suspicious Activity Reports**: Investigation details, SAR filings
- **Tax Reports**: Income, capital gains, financial account information
- **GDPR Consent**: Personal data processing records

Storing this data in plaintext on-chain violates:
- **GDPR Article 32**: Security of processing (appropriate technical measures)
- **GDPR Article 17**: Right to erasure (immutable blockchain conflict)
- **GDPR Article 5(1)(f)**: Data integrity and confidentiality

Even encryption is problematic:
- Encrypted data remains on-chain forever
- Key management is complex in decentralized systems
- Future cryptanalysis or quantum computing may break encryption
- "Right to erasure" still cannot be satisfied

## Architecture Solution

### Commitment-Based Storage

**On-Chain Storage:**
- SHA-256 cryptographic commitments (32-byte hashes)
- Non-sensitive metadata (status flags, timestamps, enum values)
- References to off-chain storage locations

**Off-Chain Storage:**
- Actual PII stored in secure, GDPR-compliant databases
- Managed by authorized compliance providers
- Supports right to erasure (Article 17)
- Supports data portability (Article 20)

### Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│ 1. KYC Provider Collects PII (Off-Chain)                    │
│    - Full name, DOB, SSN, documents, etc.                   │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. Provider Generates SHA-256 Commitment                     │
│    commitment = SHA-256(canonical_json(pii_data))            │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ├─────────────────────┬───────────────────────┐
                 ▼                     ▼                       ▼
┌──────────────────────┐  ┌──────────────────────┐  ┌─────────────────┐
│ 3a. Store Commitment │  │ 3b. Store Original   │  │ 3c. Record in   │
│     ON-CHAIN         │  │     PII OFF-CHAIN    │  │     Audit Log   │
│  (KYCRecord.pii_     │  │  (Provider's secure  │  │  (provider's    │
│   commitment field)  │  │   database)          │  │   system)       │
└──────────────────────┘  └──────────────────────┘  └─────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. Later Verification                                        │
│    - Provider retrieves PII from off-chain storage           │
│    - Recomputes commitment: SHA-256(pii_data)                │
│    - Compares with on-chain commitment                       │
│    - If match: Data integrity verified                       │
└─────────────────────────────────────────────────────────────┘
```

## Implementation

### DataProtectionService

Located in: `chain/x/compliance/keeper/encryption.go`

Core service providing:
- **Commitment Generation**: SHA-256 hashing of canonical JSON
- **Commitment Verification**: Constant-time comparison
- **Field-Level Protection**: Individual field commitments
- **PII Handling**: Structured PII data commitments
- **Hex Encoding**: Human-readable commitment display

### PIIData Struct

Comprehensive structure defining all PII fields that must be protected:

```go
type PIIData struct {
    // Identity
    FullName       string
    DateOfBirth    string
    SSN            string
    PassportNumber string
    // ... (see encryption.go for full definition)
}
```

**Important**: This struct is NEVER stored on-chain. It exists only for:
1. Off-chain provider systems
2. Commitment generation
3. Verification operations

## Usage Patterns

### Pattern 1: KYC Record Storage

```go
// Provider side: Collect PII (off-chain)
pii := &PIIData{
    FullName:       "Alice Smith",
    DateOfBirth:    "1990-01-01",
    SSN:            "123-45-6789",
    PassportNumber: "AB123456",
    Addresses:      []string{"123 Main St"},
    // ... other fields
}

// Generate commitment
service := NewDataProtectionService()
commitment, err := service.GeneratePIICommitment(pii)
if err != nil {
    return err
}

// Store commitment on-chain (actual PII stays off-chain)
record := &types.KYCRecord{
    Address:       userAddress,
    KycLevel:      types.KYCLevel_KYC_LEVEL_ADVANCED,
    Provider:      providerAddress,
    PiiCommitment: commitment, // Only commitment on-chain
    VerifiedAt:    timestamppb.Now(),
    ExpiresAt:     timestamppb.New(time.Now().Add(365*24*time.Hour)),
}

err = keeper.SetKYCRecord(ctx, record)
```

### Pattern 2: Verification

```go
// User or auditor requests verification
// Provider retrieves PII from off-chain secure storage
offChainPII := providerDB.GetPII(userAddress)

// Retrieve on-chain commitment
record, err := keeper.GetKYCRecord(ctx, userAddress)
if err != nil {
    return err
}

// Verify integrity
service := NewDataProtectionService()
valid, err := service.VerifyPIICommitment(offChainPII, record.PiiCommitment)
if err != nil {
    return err
}

if valid {
    // Data integrity confirmed
    // Provider can now share verified data off-chain
} else {
    // Data has been tampered with or commitment is invalid
    return errors.New("PII verification failed")
}
```

### Pattern 3: Field-Level Commitments

For granular protection:

```go
// Protect individual AML fields
fields := map[string]interface{}{
    "source_of_funds": []string{"employment", "investments"},
    "occupation":      "Software Engineer",
    "employer":        "Tech Corp",
    "risk_factors":    []string{"high_volume_trading"},
}

service := NewDataProtectionService()
commitments, err := service.GenerateFieldCommitments(fields)

// Store commitments (implementation depends on proto schema)
// Each field gets its own 32-byte commitment
```

### Pattern 4: Data Redaction for Logs

```go
// When logging or displaying data
data := map[string]interface{}{
    "address":    userAddress,
    "kyc_level":  "advanced",
    "ssn":        "123-45-6789",    // Sensitive
    "passport":   "AB123456",       // Sensitive
}

sensitiveFields := GetSensitiveFieldsList("kyc")
redacted := RedactSensitiveFields(data, sensitiveFields)

// redacted now has:
// {
//   "address": "cosmos1...",
//   "kyc_level": "advanced",
//   "ssn": "[REDACTED]",
//   "passport": "[REDACTED]",
// }

logger.Info("KYC verification complete", "data", redacted)
```

## GDPR Compliance

### Article 32 - Security of Processing

✅ **Satisfied**: SHA-256 commitments provide:
- Pseudonymization (original data not on-chain)
- Confidentiality (pre-image resistance)
- Integrity (tamper detection)
- Availability (on-chain commitment always accessible)

### Article 17 - Right to Erasure

✅ **Satisfied**:
- PII stored off-chain can be deleted
- On-chain commitment reveals no PII (pre-image resistant)
- Audit trail maintained (erasure request event on-chain)

### Article 20 - Right to Data Portability

✅ **Satisfied**:
- Off-chain PII can be exported
- User can verify integrity via on-chain commitment

### Article 5(1)(f) - Integrity and Confidentiality

✅ **Satisfied**:
- SHA-256 ensures data integrity
- Off-chain storage ensures confidentiality
- Commitment-based verification prevents tampering

## Security Properties

### Pre-image Resistance
Given commitment `C = SHA-256(PII)`, it is computationally infeasible to derive `PII` from `C`.

**Security Level**: 2^256 operations required (effectively impossible)

### Second Pre-image Resistance
Given `PII1` and `C1 = SHA-256(PII1)`, it is computationally infeasible to find `PII2` where `SHA-256(PII2) = C1`.

**Security Level**: 2^256 operations required

### Collision Resistance
It is computationally infeasible to find `PII1 ≠ PII2` where `SHA-256(PII1) = SHA-256(PII2)`.

**Security Level**: 2^128 operations required (still effectively impossible)

### Deterministic
Same PII always produces same commitment (required for verification).

**Implementation**: Canonical JSON ensures deterministic serialization

### Constant-Time Comparison
Commitment comparison uses constant-time algorithm to prevent timing attacks.

**Implementation**: XOR-based comparison, independent of byte values

## Off-Chain Provider Requirements

Authorized KYC/AML providers MUST implement:

### 1. Secure Storage
- GDPR-compliant database (encryption at rest, access controls)
- Regular security audits
- Penetration testing
- Incident response procedures

### 2. Data Retention Policies
- Honor user erasure requests (GDPR Article 17)
- Automatic data deletion after retention period
- Audit logging of all access

### 3. Commitment Generation
- Use provided `DataProtectionService`
- Canonical JSON serialization (sorted keys, no whitespace)
- Store commitment-to-data mapping

### 4. Verification API
- Provide endpoint for commitment verification
- Rate limiting to prevent abuse
- Authentication required

### 5. Audit Trail
- Log all PII access
- Log all commitment generations
- Log all verification requests
- Tamper-proof audit logs

## Migration from Plaintext Storage

If existing compliance records have plaintext PII on-chain:

### Step 1: Audit Current Data
```bash
# Identify records with plaintext PII
aurad query compliance kyc-records --all
```

### Step 2: Generate Commitments
```go
// For each existing record
for _, record := range existingRecords {
    // Extract PII fields (if still accessible)
    pii := extractPII(record)

    // Generate commitment
    commitment, _ := service.GeneratePIICommitment(pii)

    // Store in off-chain database
    providerDB.StorePII(record.Address, pii, commitment)
}
```

### Step 3: Update Proto Schema
Add commitment fields to protobuf definitions:
```protobuf
message AMLProfile {
    // ... existing fields ...
    bytes sensitive_data_commitment = 20;  // New field
}
```

### Step 4: Chain Upgrade
- Deploy new keeper code with commitment support
- Governance proposal for upgrade
- Migration function to populate commitments

### Step 5: Deprecate Plaintext Fields
```protobuf
message AMLProfile {
    // ...
    repeated string risk_factors = 3 [deprecated = true];
    string occupation = 10 [deprecated = true];
    // ...
    bytes sensitive_data_commitment = 20;  // Use this instead
}
```

## Testing

Comprehensive test suite in `encryption_test.go`:

- **Basic Commitment Tests**: Generation, verification
- **Field Commitment Tests**: Multiple field handling
- **PII Tests**: Full PII lifecycle
- **Security Tests**: Pre-image resistance, determinism
- **Integration Tests**: Complete workflows
- **Edge Cases**: Invalid inputs, tampering detection

Run tests:
```bash
cd chain
go test ./x/compliance/keeper -run Encryption -v
```

## Performance Characteristics

### Commitment Generation
- **Time**: ~1-5 μs per commitment (SHA-256 is fast)
- **Space**: 32 bytes per commitment (fixed size)
- **CPU**: O(n) where n = data size (linear in input)

### Verification
- **Time**: ~1-5 μs (same as generation + comparison)
- **Space**: O(1) (constant)

### Blockchain Impact
- **Gas Cost**: ~2,000 gas per 32-byte commitment write
- **Storage**: 32 bytes per commitment (vs. potentially KB for plaintext PII)
- **Queries**: O(1) lookup by address

## Best Practices

### DO:
✅ Use commitment-based storage for all PII
✅ Store original PII in secure off-chain systems
✅ Generate commitments using canonical JSON
✅ Verify commitments before trusting off-chain data
✅ Use constant-time comparison for verification
✅ Redact sensitive fields in logs
✅ Honor GDPR erasure requests
✅ Maintain audit trails

### DON'T:
❌ Store plaintext PII on-chain
❌ Store encrypted PII on-chain (keys can be compromised)
❌ Log sensitive data without redaction
❌ Expose PII in query responses
❌ Skip commitment verification
❌ Use non-canonical serialization (breaks determinism)
❌ Ignore off-chain data retention policies

## Monitoring and Alerting

Implement monitoring for:
- Failed commitment verifications (potential tampering)
- High volumes of verification requests (potential abuse)
- Off-chain database access patterns
- GDPR erasure request volumes
- Provider compliance with off-chain requirements

## Future Enhancements

### Zero-Knowledge Proofs
Use zk-SNARKs to prove properties about PII without revealing data:
- Prove age > 18 without revealing birthdate
- Prove jurisdiction without revealing exact address
- Prove income bracket without revealing exact amount

### Threshold Encryption
For cases where on-chain encrypted data is necessary:
- Multi-party computation for key shards
- M-of-N threshold decryption
- No single party can decrypt alone

### Verifiable Off-Chain Computation
Use optimistic rollups or zk-rollups for:
- Complex AML risk scoring
- ML-based fraud detection
- Privacy-preserving analytics

## References

- **GDPR**: [https://gdpr.eu/](https://gdpr.eu/)
- **SHA-256 Specification**: FIPS 180-4
- **Commitment Schemes**: [Cryptographic Commitments](https://en.wikipedia.org/wiki/Commitment_scheme)
- **Cosmos SDK**: [https://docs.cosmos.network/](https://docs.cosmos.network/)
- **NIST Privacy Framework**: [https://www.nist.gov/privacy-framework](https://www.nist.gov/privacy-framework)

## Contact

For questions about this architecture:
- Review the code: `chain/x/compliance/keeper/encryption.go`
- Check tests: `chain/x/compliance/keeper/encryption_test.go`
- Consult compliance module documentation
