# TODO: Implement DID signature verification

---
status: pending
priority: p1
issue_id: "007"
tags: [code-review, security, identity, critical]
dependencies: []
---

## Problem Statement

The DID key rotation function has a critical TODO for signature verification. Without this, identity theft is possible.

**Impact:** Users can rotate other users' DID keys without authorization.

## Findings

**Location:** `/home/decri/blockchain-projects/aura/chain/x/identity/keeper/did_key_rotation.go:525`

**Current Code:**
```go
// TODO: Implement actual signature verification
```

This is in a **production security function** for identity management.

## Proposed Solutions

### Option 1: Implement secp256k1 signature verification (Recommended)
**Pros:** Standard, well-understood
**Cons:** None
**Effort:** Medium (4-6 hours)
**Risk:** Low

```go
func (k Keeper) VerifyDIDSignature(ctx sdk.Context, did string, msg []byte, sig []byte) error {
    // Get DID document
    didDoc, found := k.GetDIDDocument(ctx, did)
    if !found {
        return ErrDIDNotFound
    }

    // Get verification method
    pubKey := didDoc.VerificationMethod[0].PublicKeyMultibase

    // Verify signature
    if !crypto.VerifySignature(pubKey, msg, sig) {
        return ErrInvalidSignature
    }

    return nil
}
```

## Acceptance Criteria

- [ ] Signature verification implemented
- [ ] Invalid signatures rejected
- [ ] Key rotation requires valid signature from current key
- [ ] Unit tests for all signature scenarios

## Work Log

| Date | Action | Notes |
|------|--------|-------|
| 2025-12-08 | Identified | Security Sentinel and Pattern Analysis agents |
