# TODO: Add input validation to CosmWasm vc-issuer contract

---
status: pending
priority: p1
issue_id: "003"
tags: [code-review, security, cosmwasm, dos-prevention]
dependencies: []
---

## Problem Statement

The CosmWasm `vc-issuer` contract's `ExecuteMsg` enum defines messages without length constraints or format validation, creating DoS and storage bloat attack vectors.

**Impact:** Attackers can submit extremely large strings, consuming excessive gas and bloating chain state.

## Findings

**Location:** `/home/decri/blockchain-projects/aura/contracts/vc-issuer/src/msg.rs`

**Vulnerable Messages:**
```rust
pub enum ExecuteMsg {
    RequestVc {
        metadata: String,  // No length limit - could be megabytes
    },
    FulfillRequest {
        credential_base64: String,  // Could be massive
    },
    RevokeVc {
        reason: String,  // Unbounded string
    },
}
```

**Attack Vectors:**
1. **DoS Attack:** Submit 10MB `metadata` string - consumes excessive gas
2. **Storage Bloat:** Large strings stored on-chain permanently
3. **Gas Exhaustion:** Large strings exceed block gas limits

## Proposed Solutions

### Option 1: Add validation in execute handlers (Recommended)
**Pros:** Quick implementation, backward compatible
**Cons:** Validation at execution time (gas already consumed for deserialization)
**Effort:** Small (1-2 hours)
**Risk:** Low

```rust
const MAX_METADATA_LEN: usize = 10_000;
const MAX_CREDENTIAL_LEN: usize = 50_000;
const MAX_REASON_LEN: usize = 1_000;

fn execute_request_vc(..., metadata: String) -> Result<...> {
    if metadata.len() > MAX_METADATA_LEN {
        return Err(ContractError::MetadataTooLarge);
    }
    // ...
}
```

### Option 2: Custom deserialization with limits
**Pros:** Rejects at deserialization, before gas consumed
**Cons:** More complex, requires custom serde
**Effort:** Medium (4 hours)
**Risk:** Medium

## Recommended Action

Option 1 - Add length validation at the start of each execute handler.

## Technical Details

**Files to Modify:**
- `contracts/vc-issuer/src/contract.rs`
- `contracts/vc-issuer/src/error.rs` (add new error types)

**Suggested Limits:**
| Field | Max Length | Rationale |
|-------|------------|-----------|
| metadata | 10KB | Typical VC metadata |
| credential_base64 | 50KB | Encoded VC with signatures |
| reason | 1KB | Revocation reason text |

## Acceptance Criteria

- [ ] All string inputs have length validation
- [ ] New error types: `MetadataTooLarge`, `CredentialTooLarge`, `ReasonTooLarge`
- [ ] Unit tests for boundary conditions
- [ ] Integration test with oversized inputs

## Work Log

| Date | Action | Notes |
|------|--------|-------|
| 2025-12-08 | Identified | Security Sentinel agent review |

## Resources

- Contract source: `contracts/vc-issuer/src/contract.rs`
- Message definitions: `contracts/vc-issuer/src/msg.rs`
