---
status: pending
priority: p1
issue_id: "102"
tags: [code-review, security, bridge, cryptography, critical]
dependencies: ["100"]
---

# P1 CRITICAL: Bridge Validator Signature Verification Bypass Risk

## Problem Statement

The bridge validator signature verification tries all recovery IDs (0-7) without cryptographic commitment, enabling potential signature malleability attacks and DoS amplification.

**Why it matters:** Cross-chain bridges are high-value targets. Weak signature verification can lead to unauthorized fund transfers or bridge freezing attacks.

## Findings

### Vulnerable Code

**File:** `/home/decri/blockchain-projects/aura/chain/x/bridge/keeper/keeper.go` (lines 400-456)

```go
// Tries all recovery IDs without rate limiting
for recID := byte(0); recID <= 7; recID++ {
    tryRecID := (recoveryID + recID) % 8
    pubKey, err := k.recoverPubKeyFromSignature(msgHash[:], sigBytes, tryRecID)
    if err != nil {
        continue
    }
    derivedAddress := k.derivePawAddressFromPubKey(pubKey)
    if derivedAddress == pawAddress {
        recoveredPubKey = pubKey
        break
    }
}
```

### Vulnerabilities

1. **No cryptographic commitment to recovery ID** - Attacker could provide signature that validates against wrong recovery ID
2. **Potential signature malleability** - Multiple valid signatures for same message
3. **DoS amplification** - 8x verification attempts per signature
4. **No replay protection** - Same signature could be reused

### Security Impact
- **CVSS Score:** 8.0 (High)
- **Attack Complexity:** Medium
- **Funds at Risk:** All bridged assets

## Proposed Solutions

### Solution A: Require Recovery ID in Signed Message (Recommended)
**Effort:** 2-3 days | **Risk:** Low

```go
type SignedMessage struct {
    Message    string
    RecoveryID byte
}

func (k Keeper) VerifyBridgeSignature(ctx sdk.Context, msg, sig []byte, claimedRecoveryID byte) bool {
    // Only verify the claimed recovery ID
    pubKey, err := k.recoverPubKeyFromSignature(msgHash[:], sigBytes, claimedRecoveryID)
    if err != nil {
        return false
    }

    // Replay protection
    signatureHash := sha256.Sum256(sig)
    if k.isSignatureUsed(ctx, signatureHash) {
        return false
    }
    k.markSignatureUsed(ctx, signatureHash, ctx.BlockHeight())

    // Rate limiting
    if err := k.checkSignatureRateLimit(ctx, address); err != nil {
        return false
    }

    return true
}
```

**Pros:**
- Eliminates malleability risk
- Adds replay protection
- Reduces DoS surface

**Cons:**
- Breaking change for existing signatures
- Requires wallet/relayer updates

### Solution B: Signature Nonce System
**Effort:** 1 week | **Risk:** Medium

Add monotonically increasing nonces to all bridge messages.

## Recommended Action

**GO WITH SOLUTION A**: Require recovery ID commitment and add replay protection.

## Technical Details

### Affected Files
- `chain/x/bridge/keeper/keeper.go`
- `chain/x/bridge/keeper/msg_server.go`
- `chain/x/bridge/types/signature.go` (new file)

### Database/State Changes
- New KV store for used signatures: `SignatureUsed/{hash}` -> `blockHeight`

## Acceptance Criteria

- [ ] Signature verification uses only claimed recovery ID
- [ ] Replay protection prevents signature reuse
- [ ] Rate limiting prevents DoS amplification
- [ ] Unit tests for malleability resistance
- [ ] Integration tests for cross-chain signature flow

## Work Log

| Date | Action | Result |
|------|--------|--------|
| 2025-12-08 | Security audit identified vulnerability | P1 Critical |

## Resources

- [Ethereum Signature Malleability](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-2)
- [Bridge Security Best Practices](https://blog.chainsafe.io/bridge-security-best-practices/)
