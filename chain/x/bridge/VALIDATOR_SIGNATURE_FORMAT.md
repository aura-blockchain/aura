# Bridge Validator Signature Format

## Security Fix: Validator Address in Signed Messages

**Breaking Change:** This is a BREAKING CHANGE for bridge validators. Validators must update their signing logic to include their validator address in the signed message.

## Problem

Previously, all validators signed the same message format:
```
sourceChain:burnTxHash:sender:amount:denom
```

This created an order-dependency ambiguity where:
- If two validators' signatures happened to match (or public keys overlapped)
- The first match in sorted validator address order would win
- This creates non-deterministic behavior and potential security issues

## Solution

Each validator now signs a message that includes their own validator address:
```
sourceChain:burnTxHash:sender:amount:denom:validator:validatorAddr
```

This ensures:
- Each signature can only match ONE specific validator
- No order-dependency or ambiguity
- More deterministic and secure signature verification

## For Bridge Validators

### Old Signing Code (DO NOT USE)
```go
msgToSign := fmt.Sprintf("%s:%s:%s:%s:%s",
    sourceChain,
    burnTxHash,
    sender,
    amount,
    denom,
)
msgHash := sha256.Sum256([]byte(msgToSign))
signature := validatorKey.Sign(msgHash[:])
```

### New Signing Code (REQUIRED)
```go
msgBase := fmt.Sprintf("%s:%s:%s:%s:%s",
    sourceChain,
    burnTxHash,
    sender,
    amount,
    denom,
)
// CRITICAL: Include your validator address in the message
msgWithValidator := fmt.Sprintf("%s:validator:%s", msgBase, validatorAddress)
msgHash := sha256.Sum256([]byte(msgWithValidator))
signature := validatorKey.Sign(msgHash[:])
```

## Verification Process

The chain will now:
1. For each provided signature
2. Try matching against each active validator
3. Build validator-specific message: `msgBase:validator:validatorAddr`
4. Verify signature against that validator's public key
5. Count only unique validator matches

## Migration Notes

- This change is NOT backwards compatible
- All validators must update their signing code simultaneously
- Recommend coordinated upgrade or governance proposal
- Test thoroughly on testnet before mainnet deployment

## Files Modified

- `chain/x/bridge/keeper/msg_server.go`: Lines 42-169, 455-501
  - Updated `verifyRawValidatorSignatures()` to accept `msgBase` instead of `msgHash`
  - Added validator address to signed message format
  - Updated documentation and comments

## Security Impact

**Severity:** Medium
**Type:** Determinism / Consensus Safety

This fix prevents potential consensus failures where different nodes might accept different validator signatures for the same transaction if public keys overlap or signatures coincidentally match.
