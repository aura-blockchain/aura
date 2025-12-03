---
id: "019"
title: "Bridge Validator Signature Verification Weakness"
status: complete
priority: p1
category: security
module: bridge
severity: CRITICAL
cvss: 9.1
source: security-audit-report
completed_date: 2025-12-03
---

# Bridge Validator Signature Verification Weakness - FIXED

## Problem

While the bridge module has added cryptographic signature verification (lines 267-296), there are still critical issues:

1. **No validator authorization check** - Any address can register as a validator
2. **Validator rotation not handled** - Active validators can change during fraud proof window
3. **No replay protection** - Same signatures can be used multiple times

## Affected Files

- `chain/x/bridge/keeper/msg_server.go:267-296`

## Vulnerability in UnlockTokens

```go
// Line 282: Verifies signatures, but doesn't check:
// 1. Are these validators authorized by governance?
// 2. Are these validators currently active?
// 3. Has this exact signature set been used before (replay)?

validCount, err := ms.verifyRawValidatorSignatures(ctx, msg.ValidatorSignatures, msgHash[:], required)
```

## Attack Scenario

```go
// Attacker compromises one validator temporarily
// Gets valid signatures for unlock transaction
// Validator is removed from active set
// Attacker replays signatures later when different validators active
// Tokens unlocked without proper authorization
```

## Impact

- Unauthorized token minting
- Bridge fund theft
- Cross-chain attack
- Loss of user funds

## Required Fix

```go
func (ms msgServer) UnlockTokens(goCtx context.Context, msg *bridgepb.MsgUnlockTokens) (*bridgepb.MsgUnlockTokensResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // ... existing validation ...

    // CRITICAL: Generate unique message including nonce to prevent replay
    nonce := ms.Keeper.getAndIncrementUnlockNonce(ctx, transferID)
    msgToSign := fmt.Sprintf("%s:%s:%s:%s:%s:%d",
        transfer.SourceChain,
        msg.BurnTxHash,
        msg.Sender,
        msg.Amount,
        msg.Denom,
        nonce, // Unique nonce
    )
    msgHash := sha256.Sum256([]byte(msgToSign))

    // Get CURRENT active validator set at this block height
    activeValidators := ms.Keeper.getActiveValidatorSet(ctx, ctx.BlockHeight())
    if len(activeValidators) == 0 {
        return nil, status.Error(codes.Internal, "no active validators")
    }

    // Verify signatures are from current active set
    validCount, validatorAddrs, err := ms.verifyValidatorSignatures(
        ctx,
        msg.ValidatorSignatures,
        msgHash[:],
        required,
        activeValidators, // Only accept signatures from current set
    )
    if err != nil {
        return nil, status.Error(codes.PermissionDenied, err.Error())
    }

    // Check if this signature set was already used
    signatureHash := sha256.Sum256([]byte(strings.Join(validatorAddrs, ",")))
    if ms.Keeper.isSignatureSetUsed(ctx, transferID, signatureHash[:]) {
        return nil, status.Error(codes.AlreadyExists, "signatures already used (replay attack prevented)")
    }

    // Store used signatures to prevent replay
    ms.Keeper.markSignatureSetUsed(ctx, transferID, signatureHash[:])

    // Rest of function...
}
```

## Acceptance Criteria

- [ ] Validator authorization check against governance-approved list
- [ ] Active validator set verification at current block height
- [ ] Replay protection with nonce and signature tracking
- [ ] Tests for replay attack prevention
- [ ] Tests for validator set changes during fraud proof window
