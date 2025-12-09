# VC Registry Module

## Overview

The Verifiable Credentials (VC) Registry module provides W3C-standard credential issuance, DID management, credential revocation with Merkle trees, presentation verification, and attribute-based selective disclosure. It integrates with confidence scoring, incident response, and arena systems for trust-based credential minting.

## Features

- **VC Issuance**: Mint verifiable credentials based on user achievements
- **DID Management**: W3C-compliant decentralized identifier documents
- **Revocation Registry**: Merkle tree-based revocation with trustless verification
- **VC Types**: KYC, Education, Employment, Reputation, Achievement, Custom
- **Policy-Based Minting**: Configurable requirements (CS threshold, IR completion, arena score)
- **Expiration Management**: Automatic expiration with renewal workflows
- **Selective Disclosure**: Zero-knowledge attribute disclosure
- **Presentation Verification**: Verify credential presentations with proof chains
- **Rate Limiting**: Prevent spam and abuse

## State

### VCRecord
- **VC ID**: Unique credential identifier
- **VC Type**: Credential category (KYC, Education, etc.)
- **Holder DID**: Owner's decentralized identifier
- **Status**: Active, Revoked, Expired, Suspended
- **Issued/Expires At**: Validity period
- **Credential Hash**: SHA-256 of full W3C credential
- **CS at Mint**: Confidence score when issued
- **Policy Version**: Minting policy used

### VCPolicy
- **CS Threshold**: Minimum confidence score required
- **Required IR IDs**: Prerequisite incident responses
- **Required Arena**: Arena completion requirement
- **Expiry Duration**: Days until expiration (0 = never)
- **Singleton**: Only one active VC per user
- **Annual Renewal**: Renewal requirement flag

### DIDDocument
- **DID**: did:aura:mainnet:...
- **Controller**: Owner address
- **Verification Methods**: Public keys for proof
- **Credential IDs**: List of associated VCs
- **Service Endpoints**: External service links

### RevocationList
- **Merkle Root**: Current revocation tree root
- **Total Revocations**: Count of revoked credentials
- **Last Updated**: Block height and timestamp

## Messages

### MsgMintVC
Mint new verifiable credential.

**Example**:
```json
{
  "holder_address": "aura1...",
  "holder_did": "did:aura:mainnet:abc123",
  "vc_type": "VC_TYPE_KYC",
  "metadata": {
    "kyc_level": "intermediate",
    "provider": "provider_name"
  }
}
```

**Response**:
```json
{
  "vc_id": "vc_xyz789",
  "issued_at": "2025-12-09T10:00:00Z",
  "expires_at": "2026-12-09T10:00:00Z",
  "credential_hash": "sha256_hash"
}
```

### MsgRevokeVC
Revoke credential (holder-initiated).

**Example**:
```json
{
  "holder_address": "aura1...",
  "vc_id": "vc_xyz789",
  "reason_text": "Credential no longer valid"
}
```

**Response**:
```json
{
  "revoked_at": "2025-12-09T10:00:00Z",
  "merkle_updated": true
}
```

### MsgAdminRevokeVC
Revoke credential (governance).

**Example**:
```json
{
  "authority": "aura1gov...",
  "vc_id": "vc_xyz789",
  "reason": "REVOCATION_FRAUD",
  "evidence": "ipfs://Qm..."
}
```

### MsgSuspendVC
Temporarily suspend credential.

**Example**:
```json
{
  "authority": "aura1gov...",
  "vc_id": "vc_xyz789",
  "reason": "Under investigation",
  "suspension_duration_days": 30
}
```

**Response**:
```json
{
  "suspended_at": "2025-12-09T10:00:00Z",
  "reactivate_at": "2026-01-08T10:00:00Z"
}
```

### MsgCreateVCPolicy
Create credential minting policy (governance).

**Example**:
```json
{
  "authority": "aura1gov...",
  "vc_type_name": "Professional Certification",
  "vc_type_enum": "VC_TYPE_EDUCATION",
  "cs_threshold": 750,
  "required_ir_ids": ["ir_education_verify"],
  "required_arena": "professional_skills",
  "required_arena_score": 80,
  "expiry_duration_days": 365,
  "singleton": true,
  "requires_annual_renewal": true,
  "metadata_uri": "ipfs://Qm..."
}
```

**Response**:
```json
{
  "policy_id": "policy_professional",
  "version": "v1.0"
}
```

### MsgRegisterDID
Register new DID document.

**Example**:
```json
{
  "controller": "aura1...",
  "did": "did:aura:mainnet:abc123",
  "verification_methods": [
    {
      "id": "key-1",
      "type": "Ed25519VerificationKey2020",
      "controller": "did:aura:mainnet:abc123",
      "public_key": "public_key_bytes"
    }
  ],
  "metadata_uri": "ipfs://Qm..."
}
```

**Response**:
```json
{
  "did": "did:aura:mainnet:abc123",
  "created_at": "2025-12-09T10:00:00Z"
}
```

### MsgCreatePresentation
Create verifiable presentation from VCs.

**Example**:
```json
{
  "holder": "aura1...",
  "vc_ids": ["vc_xyz789", "vc_abc456"],
  "verifier": "aura1verifier...",
  "challenge": "random_challenge_string",
  "disclosed_attributes": ["name", "birthdate"]
}
```

## Queries

### QueryGetVC
```bash
aurad query vcregistry vc vc_xyz789
```

### QueryListUserVCs
```bash
aurad query vcregistry user-vcs aura1... --status active --type kyc
```

### QueryCheckVCStatus
```bash
aurad query vcregistry vc-status vc_xyz789
```

**Response includes Merkle proof for trustless verification**

### QueryBatchVCStatus
```bash
aurad query vcregistry batch-status vc_xyz789,vc_abc456,vc_def123
```

### QueryGetVCPolicy
```bash
aurad query vcregistry policy "Professional Certification"
```

### QueryResolveDID
```bash
aurad query vcregistry did did:aura:mainnet:abc123
```

### QueryValidateMintEligibility
```bash
aurad query vcregistry mint-eligibility aura1... --vc-type education
```

**Response**:
```json
{
  "eligible": false,
  "missing_requirements": ["IR: education_verify", "Arena score: 60/80"],
  "current_cs": 720,
  "required_cs": 750
}
```

### QueryStats
```bash
aurad query vcregistry stats
```

## Events

| Event Type | Attributes | Description |
|------------|------------|-------------|
| `vc_minted` | `vc_id`, `vc_type`, `holder_did`, `expires_at` | New credential issued |
| `vc_revoked` | `vc_id`, `vc_type`, `reason`, `revoker` | Credential revoked |
| `vc_expired` | `vc_id`, `holder_address` | Credential expired |
| `vc_suspended` | `vc_id`, `reason`, `reactivate_at` | Credential suspended |
| `vc_reactivated` | `vc_id` | Suspension lifted |
| `vc_policy_created` | `vc_type_name`, `cs_threshold`, `version` | Policy created |
| `did_registered` | `did`, `controller` | DID document registered |
| `merkle_root_updated` | `old_root`, `new_root`, `total_revocations` | Revocation tree updated |

## Errors

| Code | Name | Description |
|------|------|-------------|
| 1 | `ErrVCNotFound` | Credential ID not found |
| 2 | `ErrIneligibleForMinting` | User doesn't meet policy requirements |
| 3 | `ErrVCExpired` | Credential expired |
| 4 | `ErrVCRevoked` | Credential revoked |
| 5 | `ErrDIDNotFound` | DID not registered |
| 6 | `ErrDuplicateDID` | DID already exists |
| 7 | `ErrRateLimitExceeded` | Too many minting requests |
| 8 | `ErrPolicyNotFound` | VC policy not found |

## Integration Notes

### For Wallet Developers

1. **VC Display**: Show credential type, status, expiration with visual indicators
2. **Revocation Check**: Verify credentials before presenting
3. **Renewal Reminders**: Notify users of expiring credentials
4. **Selective Disclosure**: Allow users to choose disclosed attributes
5. **DID Management**: Display DID document with verification methods

### Security Considerations

- **Revocation Verification**: Always check Merkle proof for revocation status
- **Expiration Checks**: Verify expiration before accepting credentials
- **Presentation Proofs**: Validate proof chains in presentations
- **Rate Limits**: Respect minting rate limits to prevent abuse

### Best Practices

- **Policy Requirements**: Display clear requirements before minting attempts
- **Credential Renewal**: Implement auto-renewal workflows for annual VCs
- **DID Portability**: Enable DID export/import across wallets
- **Privacy**: Use selective disclosure for sensitive credentials
