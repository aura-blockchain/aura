# Bridge Module - Breaking Changes

## Validator Signature Format Change (CRITICAL)

**Date:** 2025-12-25
**Severity:** HIGH - BREAKING CHANGE
**Affects:** All bridge validators

### Summary

Fixed validator signature verification order-dependency issue by requiring validators to include their address in signed messages.

### The Problem

Previously, all validators signed the same message:
```
sourceChain:burnTxHash:sender:amount:denom
```

This created ambiguity where:
- If signatures happened to match multiple validators (rare but possible)
- The first match in sorted validator address order would win
- Non-deterministic behavior could occur across nodes

### The Solution

Each validator now signs a unique message including their address:
```
sourceChain:burnTxHash:sender:amount:denom:validator:validatorAddress
```

This ensures:
- Each signature can only match ONE specific validator
- No order-dependency or ambiguity
- Deterministic consensus across all nodes

### Migration Required

**ALL validators MUST update their signing code before this change is deployed.**

#### Old Code (DO NOT USE)
```go
msg := fmt.Sprintf("%s:%s:%s:%s:%s",
    sourceChain, burnTxHash, sender, amount, denom)
msgHash := sha256.Sum256([]byte(msg))
signature := validatorKey.Sign(msgHash[:])
```

#### New Code (REQUIRED)
```go
msgBase := fmt.Sprintf("%s:%s:%s:%s:%s",
    sourceChain, burnTxHash, sender, amount, denom)
msgWithValidator := fmt.Sprintf("%s:validator:%s", msgBase, validatorAddress)
msgHash := sha256.Sum256([]byte(msgWithValidator))
signature := validatorKey.Sign(msgHash[:])
```

### Files Modified

- `chain/x/bridge/keeper/msg_server.go`
  - `verifyRawValidatorSignatures()` - Changed signature from `msgHash []byte` to `msgBase string`
  - Added validator-specific message building inside verification loop
  - Updated all comments and documentation

### Testing

All existing tests have been updated to use the new signing format. Run:
```bash
cd chain
go test ./x/bridge/keeper/...
```

### Deployment Plan

1. **Testnet First:** Deploy to testnet and verify with all validators
2. **Validator Coordination:** Ensure ALL validators have updated their signing code
3. **Mainnet Upgrade:** Coordinate a specific block height for the upgrade
4. **Monitor:** Watch for any signature verification failures post-upgrade

### Backwards Compatibility

**NONE** - This is a hard breaking change. Old signatures will NOT verify against new code.

### References

- Original Issue: Validator signature verification order-dependency
- Security Impact: Medium (consensus safety)
- PR: [To be filled]

