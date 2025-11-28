# Verifiable Credential (VC) Registry Module - Design Document

**Module:** `vcregistry`
**Version:** v1beta1
**Author:** AURA Core Team
**Date:** 2025-11-13

---

## 1. OVERVIEW

The VC Registry module manages the lifecycle of Verifiable Credentials (VCs) on the AURA blockchain, serving as an immutable registry for W3C-compliant credentials and their status. This module does NOT store PII but maintains the cryptographic proofs and status records that enable decentralized identity verification.

### 1.1 Core Responsibilities

- VC minting based on ConfidenceScore (CS) thresholds
- VC status management (Active, Revoked, Expired, Suspended)
- DID document registry and resolution
- Revocation list management
- VC policy definitions and enforcement
- Status queries for verifier applications

### 1.2 Integration Points

- **confidencescore**: Queries CS to determine VC eligibility
- **inclusionroutines**: Checks IR prerequisites for specific VCs
- **governance**: Handles VC policy updates and emergency suspensions
- **identitychange**: Coordinates with identity change workflows

---

## 2. VC TYPES

The module supports multiple VC types, each with specific minting criteria:

### 2.1 Core Credential Types

| VC Type | Description | CS Threshold | IR Prerequisites | Expiry |
|---------|-------------|--------------|------------------|--------|
| `VC:isVerifiedHuman` | Basic verified human status | 10,000 | IR-000 (Anchor) | None |
| `VC:isAgeOver18` | Age verification (18+) | 10,000 | IR-000 + IR-601 or gov ID | 1 year |
| `VC:isAgeOver21` | Age verification (21+) | 10,000 | IR-000 + IR-601 or gov ID | 1 year |
| `VC:isResidentOf` | Geographic residency | 12,000 | IR-000 + location IRs | 6 months |
| `VC:hasBiometricAuth` | Strong biometric binding | 15,000 | IR-000 + biometric arena | None |
| `VC:hasKYCVerification` | KYC compliance | 20,000 | IR-315 or IR-316 | 1 year |
| `VC:isNotaryPublic` | Notary public status | 25,000 | IR-820 + jurisdiction | None |
| `VC:isProfessionalLicense` | Professional licensing | 20,000 | IR-206 or IR-604 | 2 years |

### 2.2 Arena Focus Credentials

| VC Type | Description | CS Threshold | Arena Focus | Expiry |
|---------|-------------|--------------|-------------|--------|
| `VC:BiometricFocus` | Biometric specialization | 15,000 | Biometric (5,000+) | None |
| `VC:SocialFocus` | Social graph specialist | 15,000 | Social (5,000+) | None |
| `VC:GeoLocationFocus` | Location specialist | 15,000 | Geo (5,000+) | None |
| `VC:HighAssuranceFocus` | High-assurance specialist | 20,000 | HighAssurance (5,000+) | None |

---

## 3. VC STATE MODEL

### 3.1 Status Enumeration

```protobuf
enum VCStatus {
  VC_STATUS_UNSPECIFIED = 0;
  VC_STATUS_ACTIVE = 1;       // Valid and usable
  VC_STATUS_REVOKED = 2;      // Permanently revoked
  VC_STATUS_EXPIRED = 3;      // Passed expiration date
  VC_STATUS_SUSPENDED = 4;    // Temporarily suspended (governance)
  VC_STATUS_PENDING = 5;      // Minting in progress
}
```

### 3.2 State Transitions

```
PENDING → ACTIVE        (automatic on successful mint)
ACTIVE → REVOKED        (user-initiated or fraud detection)
ACTIVE → EXPIRED        (automatic on expiry date)
ACTIVE → SUSPENDED      (governance decision)
SUSPENDED → ACTIVE      (governance restoration)
SUSPENDED → REVOKED     (governance permanent revocation)

Terminal states: REVOKED, EXPIRED (no further transitions)
```

---

## 4. MINTING CRITERIA

### 4.1 Universal Requirements

All VC minting requires:
1. Completed IR-000 (Anchor)
2. Minimum CS threshold met
3. No active suspensions
4. Valid verification status (not revoked)

### 4.2 Credential-Specific Logic

**VC:isVerifiedHuman**
```
IF cs_total >= 10,000 AND has_anchor == true
  THEN eligible
```

**VC:isAgeOver21**
```
IF cs_total >= 10,000 AND has_anchor == true AND
   (has_ir("IR-601") OR has_ir("IR-206") OR has_ir("IR-215"))
  THEN eligible
```

**VC:hasKYCVerification**
```
IF cs_total >= 20,000 AND has_anchor == true AND
   (has_ir("IR-315") OR has_ir("IR-316"))
  THEN eligible
```

**VC:BiometricFocus**
```
IF cs_total >= 15,000 AND has_anchor == true AND
   arena_score["Biometric"] >= 5,000
  THEN eligible
```

### 4.3 Minting Flow

```
1. User submits MsgMintVC(vc_type)
2. Module validates signer owns DID
3. Query confidencescore keeper for CS and IR completions
4. Check VC-specific prerequisites
5. Verify no existing active VC of same type (if singleton)
6. Generate VC ID (deterministic or random)
7. Create VC record with ACTIVE status
8. Update DID document with new credential
9. Emit EventVCMinted
10. Return VC ID
```

---

## 5. REVOCATION PROCESS

### 5.1 Revocation Initiators

1. **User (Holder)**: Self-revocation via `MsgRevokeVC`
2. **Governance**: Emergency revocation via `MsgAdminRevokeVC`
3. **Fraud Detection**: Automatic revocation on CS slash below threshold

### 5.2 Revocation Reasons

```protobuf
enum RevocationReason {
  REVOCATION_REASON_UNSPECIFIED = 0;
  REVOCATION_REASON_USER_REQUEST = 1;        // User-initiated
  REVOCATION_REASON_FRAUD_DETECTED = 2;      // Fraud slash
  REVOCATION_REASON_CS_BELOW_THRESHOLD = 3;  // CS dropped below minimum
  REVOCATION_REASON_IR_INVALIDATED = 4;      // Required IR was invalidated
  REVOCATION_REASON_EXPIRED = 5;             // Natural expiration
  REVOCATION_REASON_GOVERNANCE = 6;          // Governance decision
  REVOCATION_REASON_SECURITY_COMPROMISE = 7; // Security incident
}
```

### 5.3 Revocation Registry

Uses a Merkle tree-based revocation list for efficient status checking:

```
RevocationList:
  - merkle_root: bytes32
  - total_revocations: uint64
  - last_updated_height: uint64

RevocationEntry:
  - vc_id: string
  - revoked_at: timestamp
  - reason: RevocationReason
  - revoker: string (holder_did or "governance")
  - merkle_proof: bytes[]
```

Verifiers query `CheckVCStatus(vc_id)` and validate Merkle proof against current root.

---

## 6. DID METHODS

### 6.1 DID Format

```
did:aura:<network>:<unique-identifier>

Examples:
- did:aura:mainnet:abc123def456
- did:aura:testnet:xyz789ghi012
```

### 6.2 DID Document Structure

```json
{
  "@context": "https://www.w3.org/ns/did/v1",
  "id": "did:aura:mainnet:abc123def456",
  "controller": "aura1...",
  "verificationMethod": [{
    "id": "did:aura:mainnet:abc123def456#key-1",
    "type": "Ed25519VerificationKey2020",
    "controller": "did:aura:mainnet:abc123def456",
    "publicKeyMultibase": "z6Mk..."
  }],
  "authentication": ["did:aura:mainnet:abc123def456#key-1"],
  "assertionMethod": ["did:aura:mainnet:abc123def456#key-1"],
  "service": [{
    "id": "did:aura:mainnet:abc123def456#vc-service",
    "type": "VerifiableCredentialService",
    "serviceEndpoint": "https://aura.network/vc-registry"
  }],
  "credentialSubject": {
    "credentials": [
      "urn:uuid:vc:isVerifiedHuman:12345",
      "urn:uuid:vc:isAgeOver21:67890"
    ]
  }
}
```

### 6.3 DID Resolution

```
Query: ResolveDID(did)
Returns: DIDDocument + VC list + metadata

Caching: Verifiers can cache DID documents with TTL
Update: DIDDocuments update automatically on VC mint/revoke
```

---

## 7. STATUS REGISTRY

### 7.1 Query Interface

```protobuf
service Query {
  rpc GetVCStatus(GetVCStatusRequest) returns (GetVCStatusResponse);
  rpc CheckVCRevocation(CheckVCRevocationRequest) returns (CheckVCRevocationResponse);
  rpc GetRevocationList(GetRevocationListRequest) returns (GetRevocationListResponse);
  rpc ResolveDID(ResolveDIDRequest) returns (ResolveDIDResponse);
  rpc ListUserVCs(ListUserVCsRequest) returns (ListUserVCsResponse);
}
```

### 7.2 Status Check Flow

```
Verifier → Query GetVCStatus(vc_id)
           ↓
        Module checks:
          1. VC exists
          2. Current status
          3. Expiration date
          4. Revocation record
           ↓
        Return: {
          status: ACTIVE | REVOKED | EXPIRED | SUSPENDED
          valid_until: timestamp
          revoked_at: timestamp (if revoked)
          reason: RevocationReason (if revoked)
          merkle_proof: bytes (for trustless verification)
        }
```

### 7.3 Batch Queries

Support batch status checks for efficiency:

```protobuf
message BatchVCStatusRequest {
  repeated string vc_ids = 1;
}

message BatchVCStatusResponse {
  map<string, VCStatusInfo> statuses = 1;
}
```

---

## 8. VC POLICIES

### 8.1 VCPolicy Definition

```protobuf
message VCPolicy {
  string vc_type = 1;                    // e.g., "VC:isAgeOver21"
  uint64 cs_threshold = 2;               // Minimum CS
  repeated string required_ir_ids = 3;   // Required IRs
  uint64 required_arena_score = 4;       // Arena-specific requirement
  string required_arena = 5;             // Arena name
  uint64 expiry_duration_days = 6;       // 0 = no expiry
  bool singleton = 7;                    // Only one active per user
  bool requires_annual_renewal = 8;
  string metadata_uri = 9;               // IPFS/URI for policy details
  VCPolicyStatus status = 10;
}

enum VCPolicyStatus {
  VC_POLICY_STATUS_UNSPECIFIED = 0;
  VC_POLICY_STATUS_DRAFT = 1;
  VC_POLICY_STATUS_ACTIVE = 2;
  VC_POLICY_STATUS_DEPRECATED = 3;
}
```

### 8.2 Policy Management

- Created via governance proposals
- Can be updated (version increments)
- Deprecated but never deleted (historical record)
- Active policies enforce minting criteria

### 8.3 Custom VC Types

Governance can create custom VC types via proposals:

```
MsgCreateVCPolicy(
  vc_type: "VC:CustomCredential",
  cs_threshold: 15000,
  required_ir_ids: ["IR-601", "IR-605"],
  expiry_duration_days: 365
)
```

---

## 9. INTEGRATION WITH CONFIDENCE SCORE

### 9.1 Keeper Interface

```go
type ConfidenceScoreKeeper interface {
    GetUserScore(ctx sdk.Context, address string) (uint64, error)
    GetUserCompletions(ctx sdk.Context, address string) ([]IRCompletion, error)
    GetArenaScore(ctx sdk.Context, address string, arena string) (uint64, error)
    HasCompletedIR(ctx sdk.Context, address string, irID string) (bool, error)
    GetAnchorInfo(ctx sdk.Context, address string) (AnchorInfo, error)
}
```

### 9.2 Verification Checks

```go
func (k Keeper) ValidateMintEligibility(ctx sdk.Context, holderAddress string, vcType string) error {
    // Get VC policy
    policy := k.GetVCPolicy(ctx, vcType)

    // Check CS threshold
    cs := k.csKeeper.GetUserScore(ctx, holderAddress)
    if cs < policy.CsThreshold {
        return ErrInsufficientCS
    }

    // Check anchor
    anchorInfo := k.csKeeper.GetAnchorInfo(ctx, holderAddress)
    if !anchorInfo.Completed {
        return ErrAnchorRequired
    }

    // Check required IRs
    for _, irID := range policy.RequiredIRIds {
        if !k.csKeeper.HasCompletedIR(ctx, holderAddress, irID) {
            return ErrMissingRequiredIR
        }
    }

    // Check arena requirement
    if policy.RequiredArena != "" {
        arenaScore := k.csKeeper.GetArenaScore(ctx, holderAddress, policy.RequiredArena)
        if arenaScore < policy.RequiredArenaScore {
            return ErrInsufficientArenaScore
        }
    }

    return nil
}
```

### 9.3 Automatic Revocation on CS Changes

The module subscribes to CS slash events:

```go
func (k Keeper) HandleCSSlashEvent(ctx sdk.Context, event EventScoreSlashed) {
    // Get all VCs for user
    vcs := k.GetUserVCs(ctx, event.WalletAddress)

    // Check each VC's threshold
    for _, vc := range vcs {
        policy := k.GetVCPolicy(ctx, vc.VCType)
        if event.NewScore < policy.CsThreshold {
            // Auto-revoke
            k.RevokeVC(ctx, vc.VCID, REVOCATION_REASON_CS_BELOW_THRESHOLD)
        }
    }
}
```

---

## 10. DATA MODELS

### 10.1 VCRecord

```protobuf
message VCRecord {
  string vc_id = 1;                        // Unique identifier
  string vc_type = 2;                      // e.g., "VC:isVerifiedHuman"
  string holder_did = 3;                   // DID of holder
  string holder_address = 4;               // Bech32 address
  VCStatus status = 5;
  google.protobuf.Timestamp issued_at = 6;
  google.protobuf.Timestamp expires_at = 7;
  uint64 issued_height = 8;
  bytes credential_hash = 9;               // SHA256 of full credential
  bytes verifier_plugin_hash = 10;         // Hash of issuing plug-in
  string issuer_assistant = 11;            // AI assistant address
  repeated string prerequisite_ir_ids = 12;
  map<string, string> metadata = 13;       // Extensible metadata
}
```

### 10.2 RevocationRecord

```protobuf
message RevocationRecord {
  string vc_id = 1;
  google.protobuf.Timestamp revoked_at = 2;
  uint64 revoked_height = 3;
  RevocationReason reason = 4;
  string revoker = 5;                      // holder_did or "governance"
  string evidence = 6;                     // IPFS hash or metadata
  bytes merkle_proof = 7;
}
```

### 10.3 DIDDocument (on-chain simplified)

```protobuf
message DIDDocument {
  string did = 1;
  string controller = 2;                   // Bech32 address
  repeated VerificationMethod verification_methods = 3;
  repeated string credential_ids = 4;      // List of active VCs
  google.protobuf.Timestamp created = 5;
  google.protobuf.Timestamp updated = 6;
  string metadata_uri = 7;                 // Full doc on IPFS
}

message VerificationMethod {
  string id = 1;
  string type = 2;
  string controller = 3;
  bytes public_key = 4;
}
```

---

## 11. SECURITY CONSIDERATIONS

### 11.1 Access Control

- Only holder can revoke own VCs (except governance override)
- Only governance can create/update VC policies
- Only AI assistants can mint VCs (after eligibility validation)
- Rate limiting on minting to prevent spam

### 11.2 Sybil Resistance

- VC minting requires verified CS (Sybil-resistant by design)
- Each VC type checks specific IR completions
- Arena focus credentials prevent credential farming

### 11.3 Privacy Protection

- No PII stored on-chain
- VCs contain only type and status information
- Full credential data stored client-side or on IPFS
- Merkle proofs allow trustless verification without revealing full revocation list

### 11.4 Replay Protection

- Each VC has unique ID
- Minting transactions signed by holder
- Status changes tracked with block height and timestamp
- Revoked VCs cannot be reactivated

---

## 12. EVENTS

```protobuf
message EventVCMinted {
  string vc_id = 1;
  string vc_type = 2;
  string holder_did = 3;
  string holder_address = 4;
  google.protobuf.Timestamp issued_at = 5;
  uint64 block_height = 6;
}

message EventVCRevoked {
  string vc_id = 1;
  string vc_type = 2;
  RevocationReason reason = 3;
  string revoker = 4;
  google.protobuf.Timestamp revoked_at = 5;
}

message EventVCExpired {
  string vc_id = 1;
  string vc_type = 2;
  google.protobuf.Timestamp expired_at = 3;
}

message EventVCPolicyCreated {
  string vc_type = 1;
  uint64 cs_threshold = 2;
  uint64 block_height = 3;
}

message EventDIDCreated {
  string did = 1;
  string controller = 2;
  uint64 block_height = 3;
}
```

---

## 13. PARAMETERS

```yaml
params:
  # Minting limits
  max_vcs_per_user: 50
  max_mint_per_day: 5
  max_mint_per_hour: 2

  # Default expirations (overridden by policy)
  default_vc_expiry_days: 365

  # Revocation
  revocation_merkle_update_frequency: 10  # blocks

  # DID
  did_prefix: "did:aura"
  did_network: "mainnet"

  # Fees
  mint_fee: "0.1 AURA"
  revoke_fee: "0 AURA"  # Free
  policy_creation_deposit: "10000 AURA"
```

---

## 14. FUTURE ENHANCEMENTS

### 14.1 Planned Features

1. **Delegated Credentials**: Allow users to delegate VCs to other addresses
2. **VC Transfers**: Transfer ownership of certain VC types
3. **Renewal Automation**: Auto-renew expiring VCs if criteria still met
4. **Batch Minting**: Mint multiple VCs in single transaction
5. **VC Schemas**: On-chain JSON-LD schema registry
6. **Verifier Whitelist**: Policy-based trusted verifier lists

### 14.2 Research Areas

1. **zkSNARK VCs**: Full ZK credential support
2. **Selective Disclosure**: Fine-grained attribute disclosure
3. **Cross-Chain VCs**: IBC VC transfer and validation
4. **Credential Composition**: Combine multiple VCs into composite proofs

---

## 15. REFERENCES

- W3C Verifiable Credentials Data Model 1.0: https://www.w3.org/TR/vc-data-model/
- W3C Decentralized Identifiers (DIDs) v1.0: https://www.w3.org/TR/did-core/
- AURA Technical Specification v8.0
- Cosmos SDK Module Documentation
- BLS12-381 for ZK proofs
- Merkle Tree revocation registry standards

---

**END OF DESIGN DOCUMENT**
