# GDPR-Compliant Architecture for Compliance Module

## Overview

The Aura blockchain compliance module implements a **commitment-based architecture** that stores only cryptographic commitments (hashes) of Personally Identifiable Information (PII) on-chain, while the actual PII is stored off-chain. This design satisfies GDPR requirements while maintaining blockchain immutability.

## Problem Statement

Storing PII directly on an immutable blockchain violates GDPR Article 17 "Right to Erasure," which requires that users can request deletion of their personal data. Once data is written to a blockchain, it cannot be deleted.

## Solution: Commitment-Based Storage

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         User/KYC Provider                        │
└────────────────┬────────────────────────────────────────────────┘
                 │
                 │ 1. Collect PII
                 │    (verification_id, documents, jurisdiction, etc.)
                 │
                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Off-Chain PII Storage                        │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  Encrypted Database (KYC Provider Responsibility)         │  │
│  │  - verification_id                                        │  │
│  │  - documents                                              │  │
│  │  - jurisdiction                                           │  │
│  │  - risk_score                                             │  │
│  │  - Other sensitive data                                   │  │
│  └───────────────────────────────────────────────────────────┘  │
└────────────────┬────────────────────────────────────────────────┘
                 │
                 │ 2. Compute SHA-256(PII_JSON)
                 │    → 32-byte commitment hash
                 │
                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Aura Blockchain (On-Chain)                    │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  KYCRecord {                                              │  │
│  │    address: "aura1..."                                    │  │
│  │    kyc_level: KYC_LEVEL_ADVANCED                          │  │
│  │    provider: "aura1provider..."                           │  │
│  │    verified_at: 2024-01-15T10:00:00Z                      │  │
│  │    expires_at: 2025-01-15T10:00:00Z                       │  │
│  │    pii_commitment: 0xabcd1234... (32 bytes)               │  │
│  │    enhanced_due_diligence: true                           │  │
│  │  }                                                        │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### Data Flow

#### KYC Submission

1. **Off-Chain**: KYC provider collects and verifies user's PII
2. **Off-Chain**: Provider stores encrypted PII in their secure database
3. **Off-Chain**: Provider computes SHA-256 hash of PII JSON
4. **On-Chain**: Provider submits `MsgSubmitKYC` with commitment hash
5. **On-Chain**: Blockchain stores only the commitment, not the PII

#### Data Verification

1. **Off-Chain**: Verifier requests PII from KYC provider
2. **Off-Chain**: Provider returns PII (after authorization checks)
3. **Off-Chain**: Verifier computes SHA-256(PII_JSON)
4. **On-Chain**: Verifier queries blockchain for stored commitment
5. **Verification**: Compare computed hash with on-chain commitment

#### GDPR Erasure

1. **On-Chain**: User submits `MsgEraseGDPRData` transaction
2. **On-Chain**: Blockchain emits "gdpr_data_erased" event
3. **Off-Chain**: KYC provider monitors blockchain for erasure events
4. **Off-Chain**: Provider deletes PII from their database
5. **Result**: On-chain commitment remains (audit trail), but PII is gone

## Implementation Details

### Proto Definitions

#### Before (GDPR Violation)

```protobuf
message KYCRecord {
  string address = 1;
  KYCLevel kyc_level = 2;
  string provider = 3;
  google.protobuf.Timestamp verified_at = 4;
  google.protobuf.Timestamp expires_at = 5;
  string verification_id = 6;        // ❌ PII on-chain
  repeated string documents = 7;     // ❌ PII on-chain
  string jurisdiction = 8;           // ❌ PII on-chain
  string risk_score = 10;            // ❌ PII on-chain
}
```

#### After (GDPR Compliant)

```protobuf
message KYCRecord {
  string address = 1;
  KYCLevel kyc_level = 2;
  string provider = 3;
  google.protobuf.Timestamp verified_at = 4;
  google.protobuf.Timestamp expires_at = 5;
  bytes pii_commitment = 6;          // ✅ Only hash on-chain (32 bytes)
  bool enhanced_due_diligence = 7;
}
```

### Computing PII Commitment

The commitment hash must be computed deterministically:

```go
package main

import (
    "crypto/sha256"
    "encoding/json"
)

// OffChainKYCData represents PII stored off-chain
type OffChainKYCData struct {
    VerificationID string   `json:"verification_id"`
    Documents      []string `json:"documents"`
    Jurisdiction   string   `json:"jurisdiction"`
    RiskScore      string   `json:"risk_score"`
}

// ComputePIICommitment creates a SHA-256 hash of the PII
func ComputePIICommitment(data OffChainKYCData) ([]byte, error) {
    // Serialize to canonical JSON (deterministic order)
    jsonBytes, err := json.Marshal(data)
    if err != nil {
        return nil, err
    }

    // Compute SHA-256 hash
    hash := sha256.Sum256(jsonBytes)
    return hash[:], nil
}

// Example usage:
// piiData := OffChainKYCData{
//     VerificationID: "KYC-2024-001234",
//     Documents:      []string{"passport", "utility_bill"},
//     Jurisdiction:   "US-CA",
//     RiskScore:      "low",
// }
// commitment, _ := ComputePIICommitment(piiData)
// // Now submit commitment to blockchain via MsgSubmitKYC
```

### Submitting KYC with Commitment

```go
import (
    "github.com/aequitas/aura/chain/x/compliance/types"
)

// Submit KYC record with PII commitment
func SubmitKYCRecord(
    providerAddr string,
    userAddr string,
    kycLevel types.KYCLevel,
    piiCommitment []byte, // Must be 32 bytes (SHA-256)
) (*types.MsgSubmitKYC, error) {

    if len(piiCommitment) != 32 {
        return nil, fmt.Errorf("commitment must be 32 bytes")
    }

    msg := &types.MsgSubmitKYC{
        Address:       userAddr,
        KycLevel:      kycLevel,
        Provider:      providerAddr,
        PiiCommitment: piiCommitment,
    }

    return msg, nil
}
```

### GDPR Erasure Request

```go
// User requests data erasure (GDPR Article 17)
func RequestDataErasure(userAddr string, reason string) *types.MsgEraseGDPRData {
    return &types.MsgEraseGDPRData{
        Address:       userAddr,
        ErasureReason: reason,
    }
}

// Off-chain system monitors for erasure events
func MonitorErasureEvents(client *grpc.ClientConn) {
    // Subscribe to blockchain events
    events := subscribeToEvents(client, "gdpr_data_erased")

    for event := range events {
        address := event.GetAttribute("address")
        erasureEventID := event.GetAttribute("erasure_event_id")

        // Delete PII from off-chain database
        err := deletePIIFromDatabase(address)
        if err != nil {
            log.Errorf("Failed to delete PII for %s: %v", address, err)
            continue
        }

        // Log erasure in audit trail
        logErasure(address, erasureEventID)

        // Notify user (off-chain)
        notifyUserErasureComplete(address)
    }
}
```

## Off-Chain Storage Requirements

### KYC Provider Responsibilities

KYC providers MUST implement the following:

1. **Secure Storage**
   - Encrypt PII at rest (AES-256 or equivalent)
   - Encrypt PII in transit (TLS 1.3)
   - Use access controls (least privilege)
   - Implement audit logging

2. **Commitment Computation**
   - Canonical JSON serialization (deterministic)
   - SHA-256 hashing
   - Store mapping: address → PII
   - Store mapping: commitment → PII (for verification)

3. **Event Monitoring**
   - Monitor blockchain for "gdpr_data_erased" events
   - Automated PII deletion upon erasure event
   - Audit log of all deletions
   - Compliance report generation

4. **Data Retention**
   - Honor GDPR erasure requests within 30 days
   - Delete PII, not the commitment mapping
   - Maintain audit trail of erasure requests

5. **Access Controls**
   - Only authorized parties can access PII
   - User can access their own PII (GDPR Article 15)
   - Regulators can access with proper authorization
   - Log all access attempts

### Reference Implementation

```python
# Example off-chain PII storage service (Python)

import hashlib
import json
from cryptography.fernet import Fernet
import psycopg2

class GDPRCompliantKYCStorage:
    def __init__(self, db_conn, encryption_key):
        self.db = db_conn
        self.cipher = Fernet(encryption_key)

    def store_kyc_data(self, address: str, pii_data: dict) -> bytes:
        """Store PII and return commitment hash"""
        # Serialize to canonical JSON
        pii_json = json.dumps(pii_data, sort_keys=True)

        # Compute commitment
        commitment = hashlib.sha256(pii_json.encode()).digest()

        # Encrypt PII
        encrypted_pii = self.cipher.encrypt(pii_json.encode())

        # Store in database
        cursor = self.db.cursor()
        cursor.execute(
            "INSERT INTO kyc_pii (address, commitment, encrypted_data) "
            "VALUES (%s, %s, %s)",
            (address, commitment.hex(), encrypted_pii)
        )
        self.db.commit()

        return commitment

    def get_kyc_data(self, address: str) -> dict:
        """Retrieve PII for an address"""
        cursor = self.db.cursor()
        cursor.execute(
            "SELECT encrypted_data FROM kyc_pii WHERE address = %s",
            (address,)
        )
        row = cursor.fetchone()

        if not row:
            raise ValueError(f"No PII found for {address}")

        # Decrypt PII
        decrypted = self.cipher.decrypt(row[0])
        return json.loads(decrypted)

    def erase_kyc_data(self, address: str):
        """Delete PII (GDPR Article 17)"""
        cursor = self.db.cursor()

        # Log erasure before deleting
        cursor.execute(
            "INSERT INTO erasure_audit (address, erased_at) "
            "VALUES (%s, NOW())",
            (address,)
        )

        # Delete PII
        cursor.execute(
            "DELETE FROM kyc_pii WHERE address = %s",
            (address,)
        )

        self.db.commit()
```

## GDPR Compliance Checklist

- [x] **Article 5 (Data Minimization)**: Only essential data (commitment) stored on-chain
- [x] **Article 17 (Right to Erasure)**: PII can be deleted off-chain on request
- [x] **Article 25 (Privacy by Design)**: Architecture designed for privacy from ground up
- [x] **Article 32 (Security)**: PII encrypted, access controlled
- [x] **Article 30 (Records of Processing)**: Audit trail via blockchain events
- [x] **Article 33 (Breach Notification)**: Immutable event log for compliance

## Migration Guide

For existing deployments with PII on-chain:

1. **Stop accepting new on-chain PII immediately**
2. **Deploy updated protobuf definitions**
3. **For existing records**:
   - Extract PII to off-chain storage
   - Compute commitments
   - Store commitments on-chain (via upgrade migration)
   - Mark old fields as deprecated
4. **Monitor for erasure requests**
5. **Document migration in audit trail**

## Testing

### Test Commitment Verification

```go
func TestPIICommitmentVerification(t *testing.T) {
    // Off-chain PII
    piiData := OffChainKYCData{
        VerificationID: "TEST-001",
        Documents:      []string{"passport"},
        Jurisdiction:   "US",
        RiskScore:      "low",
    }

    // Compute commitment
    commitment, err := ComputePIICommitment(piiData)
    require.NoError(t, err)
    require.Len(t, commitment, 32)

    // Store on-chain
    msg := &types.MsgSubmitKYC{
        Address:       testAddr,
        KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
        Provider:      providerAddr,
        PiiCommitment: commitment,
    }

    // Submit to blockchain
    resp, err := msgServer.SubmitKYC(ctx, msg)
    require.NoError(t, err)
    require.True(t, resp.Success)

    // Retrieve from blockchain
    record, err := keeper.GetKYCRecord(ctx, testAddr)
    require.NoError(t, err)

    // Verify commitment matches
    require.Equal(t, commitment, record.PiiCommitment)

    // Verify we can validate PII
    recomputedCommitment, _ := ComputePIICommitment(piiData)
    require.Equal(t, commitment, recomputedCommitment)
}
```

### Test GDPR Erasure

```go
func TestGDPRErasure(t *testing.T) {
    // Submit erasure request
    msg := &types.MsgEraseGDPRData{
        Address:       testAddr,
        ErasureReason: "GDPR Article 17 request",
    }

    resp, err := msgServer.EraseGDPRData(ctx, msg)
    require.NoError(t, err)
    require.True(t, resp.Success)
    require.NotEmpty(t, resp.ErasureEventId)

    // Verify event was emitted
    events := ctx.EventManager().Events()
    found := false
    for _, event := range events {
        if event.Type == "gdpr_data_erased" {
            found = true
            // Check event attributes
            for _, attr := range event.Attributes {
                if string(attr.Key) == "address" {
                    require.Equal(t, testAddr, string(attr.Value))
                }
            }
        }
    }
    require.True(t, found, "erasure event not emitted")

    // Off-chain system would now delete PII
}
```

## Security Considerations

1. **Commitment Collision**: SHA-256 provides 256-bit security (collision probability negligible)
2. **Rainbow Tables**: Cannot reverse hash to get PII (one-way function)
3. **Provider Trust**: Users must trust KYC providers to store PII securely
4. **Key Management**: Off-chain encryption keys must be protected
5. **Event Monitoring**: Off-chain systems must reliably monitor blockchain events

## Audit Compliance

### For Auditors

The on-chain commitment provides:
- **Proof of existence**: PII existed at verification time
- **Proof of integrity**: Commitment hasn't changed (blockchain immutability)
- **Proof of erasure**: "gdpr_data_erased" event provides audit trail
- **Non-repudiation**: Provider signature on commitment submission

The off-chain storage provides:
- **Actual PII**: Available to authorized parties (user, regulators)
- **Erasure capability**: Can be deleted on GDPR request
- **Audit logs**: Who accessed what, when

## Conclusion

This architecture satisfies GDPR requirements while maintaining blockchain immutability. The key insight: **store proofs on-chain, data off-chain**.

## References

- [GDPR Article 17 (Right to Erasure)](https://gdpr-info.eu/art-17-gdpr/)
- [GDPR Article 5 (Principles)](https://gdpr-info.eu/art-5-gdpr/)
- [GDPR Article 25 (Data Protection by Design)](https://gdpr-info.eu/art-25-gdpr/)
- [Blockchain and GDPR Whitepaper](https://www.europarl.europa.eu/RegData/etudes/STUD/2019/634445/EPRS_STU(2019)634445_EN.pdf)
