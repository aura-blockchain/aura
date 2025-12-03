# CRITICAL: Incomplete Cryptographic Signature Verification in Bridge

**Status:** ready
**Priority:** P0 (MAINNET BLOCKER)
**Severity:** CRITICAL
**CWE:** CWE-347 (Improper Verification of Cryptographic Signature)
**CVSS Score:** 9.8

## Summary

Cross-chain signature verification in the bridge module only checks signature length (≥64 bytes) without actually verifying the cryptographic signature. This allows any attacker to link arbitrary PAW/XAI addresses using garbage data.

## Location

- **File:** `chain/x/bridge/keeper/keeper.go`
- **Lines:** 256-287, 307-338
- **Functions:** `verifyPAWSignature()`, `verifyXAISignature()`

## Vulnerability Details

```go
// Current INSECURE implementation:
func (k Keeper) verifyPAWSignature(signature []byte) bool {
    // TODO: Implement full secp256k1 signature verification
    // For now, we verify the signature is present and non-empty
    if len(signature) < 64 {
        return false
    }
    return len(signature) >= 64  // ANY 64+ bytes passes!
}
```

**Attack Scenario:**
1. Attacker generates 64 random bytes
2. Calls `LinkPAWAddress` with victim's PAW address and random signature
3. Signature "verification" passes
4. Attacker gains control of victim's cross-chain identity
5. Can execute unauthorized cross-chain operations

## Impact

- **Identity Theft:** Attackers can link any PAW/XAI address to their Aura identity
- **Unauthorized Operations:** Full control over cross-chain transfers
- **Fund Loss:** Ability to drain bridged assets
- **Reputation Damage:** Complete failure of security model

## Required Fix

Implement full secp256k1 signature verification:

```go
func (k Keeper) verifyPAWSignature(pawAddress string, message []byte, signature []byte) bool {
    // 1. Decode PAW address to extract public key hash
    pubKeyHash, err := decodePAWAddress(pawAddress)
    if err != nil {
        return false
    }

    // 2. Parse signature (R, S, V format)
    if len(signature) != 65 {
        return false
    }
    r := new(big.Int).SetBytes(signature[:32])
    s := new(big.Int).SetBytes(signature[32:64])
    v := signature[64]

    // 3. Recover public key from signature
    pubKey, err := crypto.Ecrecover(crypto.Keccak256(message), signature)
    if err != nil {
        return false
    }

    // 4. Verify public key matches PAW address
    recoveredHash := crypto.Keccak256(pubKey[1:])[12:]
    if !bytes.Equal(recoveredHash, pubKeyHash) {
        return false
    }

    // 5. Verify signature cryptographically
    return crypto.VerifySignature(pubKey, crypto.Keccak256(message), signature[:64])
}
```

## Testing Requirements

Add comprehensive security tests:

```go
func TestBridgeSignatureVerification_Security(t *testing.T) {
    tests := []struct {
        name          string
        signature     []byte
        shouldPass    bool
        attackVector  string
    }{
        {
            name:          "valid signature",
            signature:     generateValidSignature(),
            shouldPass:    true,
            attackVector:  "",
        },
        {
            name:          "random 64 bytes",
            signature:     randomBytes(64),
            shouldPass:    false,
            attackVector:  "Attacker provides random data",
        },
        {
            name:          "wrong signer",
            signature:     signWithWrongKey(),
            shouldPass:    false,
            attackVector:  "Attacker uses their own key",
        },
        {
            name:          "replayed signature",
            signature:     replayOldSignature(),
            shouldPass:    false,
            attackVector:  "Replay attack",
        },
        {
            name:          "malleated signature",
            signature:     malleateValidSignature(),
            shouldPass:    false,
            attackVector:  "Signature malleability",
        },
    }
    // ... test implementation
}
```

## Acceptance Criteria

- [ ] Implement full secp256k1 signature recovery and verification
- [ ] Verify recovered public key matches claimed address
- [ ] Add replay protection (nonces)
- [ ] Add signature expiration timestamps
- [ ] Comprehensive security test suite (20+ attack vectors)
- [ ] External cryptography audit of implementation
- [ ] Documentation of signature format and security properties

## References

- [CWE-347: Improper Verification of Cryptographic Signature](https://cwe.mitre.org/data/definitions/347.html)
- [OWASP: Cryptographic Failures](https://owasp.org/Top10/A02_2021-Cryptographic_Failures/)
- [Bitcoin: ECDSA Signature Verification](https://en.bitcoin.it/wiki/Elliptic_Curve_Digital_Signature_Algorithm)
- [Ethereum: Signature Recovery](https://github.com/ethereum/go-ethereum/blob/master/crypto/signature_cgo.go)

## Related Issues

- Security Audit Report: CRITICAL-001
- See also: todos/019-ready-p1-bridge-validator-signature-weakness.md

---

**DO NOT DEPLOY TO MAINNET UNTIL THIS IS FIXED**
