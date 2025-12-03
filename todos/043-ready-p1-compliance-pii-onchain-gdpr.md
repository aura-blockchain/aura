---
id: "043"
title: "Compliance PII Stored On-Chain - GDPR Violation"
status: ready
priority: p1
category: compliance
module: compliance
severity: CRITICAL
cvss: 9.0
legal_risk: HIGH
source: compliance-audit
---

# Compliance PII Stored On-Chain - GDPR Violation

## Problem

Personally Identifiable Information (PII) is stored directly on the immutable blockchain, violating GDPR Article 17 "Right to Erasure".

## Affected Files

- `proto/aura/compliance/v1beta1/compliance.proto:44-56`
- `chain/x/compliance/keeper/keeper_kvstore.go`

## PII Currently Stored On-Chain

```protobuf
message KYCRecord {
    string address = 1;
    KYCLevel kyc_level = 2;
    string provider = 3;
    google.protobuf.Timestamp verified_at = 4;
    google.protobuf.Timestamp expires_at = 5;
    string verification_id = 6;       // PII
    repeated string documents = 7;     // PII
    string jurisdiction = 8;           // PII
    bool enhanced_due_diligence = 9;
    string risk_score = 10;           // PII
}

message GDPRConsent {
    string ip_address = 7;            // PII
    string user_agent = 8;            // PII
}
```

## Legal Impact

- **GDPR Article 17 Violation**: Data cannot be erased from blockchain
- **Potential Fines**: Up to €20M or 4% of global annual revenue
- **Cannot Operate in EU**: Without GDPR compliance

## Required Fix - Hybrid On-Chain/Off-Chain Architecture

```protobuf
// FIXED: Only store hashes and non-PII data on-chain
message KYCRecord {
    string address = 1;
    KYCLevel kyc_level = 2;
    google.protobuf.Timestamp verified_at = 3;
    google.protobuf.Timestamp expires_at = 4;
    bytes pii_commitment = 5;  // Hash of off-chain PII
    bool enhanced_due_diligence = 6;
    // NO: verification_id, documents, jurisdiction, risk_score
}

message GDPRConsent {
    string address = 1;
    string consent_type = 2;
    bool consented = 3;
    google.protobuf.Timestamp consent_given_at = 4;
    google.protobuf.Timestamp consent_withdrawn_at = 5;
    string consent_version = 6;
    // NO: ip_address, user_agent
    bytes audit_commitment = 7;  // Hash for audit purposes
}
```

```go
// Off-chain encrypted storage for PII
type OffChainKYCData struct {
    VerificationID string
    Documents      []string
    Jurisdiction   string
    RiskScore      string
    IPAddress      string  // For GDPRConsent audit
    UserAgent      string
}

// Keeper stores commitment on-chain, data off-chain
func (k *Keeper) SetKYCRecord(ctx sdk.Context, record *types.KYCRecord, piiData *OffChainKYCData) error {
    // 1. Hash the PII for on-chain commitment
    piiBytes, _ := json.Marshal(piiData)
    commitment := sha256.Sum256(piiBytes)
    record.PiiCommitment = commitment[:]

    // 2. Store commitment on-chain
    store := ctx.KVStore(k.storeKey)
    bz, _ := k.cdc.Marshal(record)
    store.Set(KYCRecordsKey(record.Address), bz)

    // 3. Store encrypted PII off-chain
    encryptedPII, _ := k.encryptionService.Encrypt(piiBytes)
    k.offChainStorage.Store(record.Address, encryptedPII)

    return nil
}

// GDPR Right to Erasure implementation
func (k *Keeper) EraseUserData(ctx sdk.Context, address string) error {
    // 1. Delete off-chain PII (can be deleted)
    k.offChainStorage.Delete(address)

    // 2. On-chain record remains but commitment is now orphaned
    // Cannot verify PII anymore, but on-chain record shows KYC level

    // 3. Emit event for audit
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            "gdpr_data_erased",
            sdk.NewAttribute("address", address),
            sdk.NewAttribute("erasure_time", ctx.BlockTime().String()),
        ),
    )

    return nil
}
```

## Acceptance Criteria

- [ ] All PII moved off-chain
- [ ] Only commitments/hashes stored on-chain
- [ ] Encrypted off-chain storage implemented
- [ ] GDPR erasure endpoint implemented
- [ ] Migration script for existing data
- [ ] Legal review of new architecture
- [ ] Tests for commitment verification
- [ ] Tests for data erasure
