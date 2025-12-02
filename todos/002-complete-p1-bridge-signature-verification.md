---
status: ready
priority: p1
issue_id: "002"
tags: [code-review, security, bridge, critical]
dependencies: []
---

# Bridge Minting Without Validator Signature Verification

## Problem Statement

The `MintTokens` and `UnlockTokens` functions accept validator signatures as parameters but NEVER cryptographically verify them. The code only checks the COUNT of signatures, not their validity.

**Why it matters:** This is a **CRITICAL** vulnerability that allows unlimited token theft from the bridge module. An attacker can submit fake signatures and drain all bridge funds.

## Findings

### Evidence
- **File:** `chain/x/bridge/keeper/msg_server.go`
- **Lines:** 92-163, 184-201

```go
// VULNERABLE CODE - Lines 184-201
func (ms msgServer) UnlockTokens(goCtx context.Context, msg *bridgepb.MsgUnlockTokens) (*bridgepb.MsgUnlockTokensResponse, error) {
    // ...
    required := ms.Keeper.GetParams(ctx).MinConfirmations
    if uint64(len(msg.ValidatorSignatures)) < required {
        return nil, status.Error(codes.FailedPrecondition, "insufficient signatures")
    }
    // NO SIGNATURE VERIFICATION - just checks count!

    // IMMEDIATE UNLOCK:
    if ms.Keeper.bankKeeper != nil {
        if err := ms.Keeper.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, sdk.NewCoins(coin)); err != nil {
            return nil, status.Error(codes.Internal, err.Error())
        }
    }
}
```

### Attack Vector
1. Attacker crafts fake validator signatures (random bytes)
2. Submits UnlockTokens with `validator_signatures = ["fake1", "fake2", "fake3"]`
3. Receives unlocked tokens without any cross-chain proof

### Impact
- **CRITICAL**: Unlimited token theft from bridge module
- Bridge insolvency
- Cross-chain attack enabling token inflation on all connected chains

## Proposed Solutions

### Option A: Implement Full Signature Verification (Required)
**Pros:** Actually secure, industry standard
**Cons:** Requires cryptographic implementation
**Effort:** Medium (4-8 hours)
**Risk:** Medium (must implement crypto correctly)

```go
func (ms msgServer) UnlockTokens(goCtx context.Context, msg *bridgepb.MsgUnlockTokens) (*bridgepb.MsgUnlockTokensResponse, error) {
    // ... existing checks ...

    // Verify each validator signature
    validators := ms.Keeper.stakingKeeper.GetBondedValidatorsByPower(ctx)
    signedPower := sdk.ZeroInt()
    totalPower := sdk.ZeroInt()

    for _, val := range validators {
        totalPower = totalPower.Add(val.GetConsensusPower(sdk.DefaultPowerReduction))
    }

    // Reconstruct message to verify
    signBytes := ms.Keeper.constructBridgeProof(transferID, msg.BurnTxHash, msg.Amount, msg.Denom)

    for _, sig := range msg.ValidatorSignatures {
        valAddr, err := sdk.ValAddressFromBech32(sig.ValidatorAddress)
        if err != nil {
            continue
        }

        validator, found := ms.Keeper.stakingKeeper.GetValidator(ctx, valAddr)
        if !found {
            continue
        }

        pubKey := validator.GetConsPubKey()
        if !pubKey.VerifySignature(signBytes, sig.Signature) {
            return nil, status.Error(codes.InvalidArgument, "invalid validator signature")
        }

        signedPower = signedPower.Add(validator.GetConsensusPower(sdk.DefaultPowerReduction))
    }

    // Require 2/3+ voting power
    requiredPower := totalPower.MulRaw(2).QuoRaw(3)
    if signedPower.LT(requiredPower) {
        return nil, status.Error(codes.FailedPrecondition, "insufficient validator power")
    }

    // ... proceed with unlock
}
```

### Option B: Disable Bridge Until Fixed
**Pros:** Prevents exploitation
**Cons:** Feature unavailable
**Effort:** Small (30 min)
**Risk:** Low

## Recommended Action
**IMMEDIATE**: Implement Option A before ANY deployment. This is a critical security vulnerability that could result in total loss of bridge funds.

## Technical Details

### Affected Files
- `chain/x/bridge/keeper/msg_server.go` (MintTokens, UnlockTokens)
- May need new file for signature verification helpers

### Components Affected
- Bridge module token minting/unlocking
- Cross-chain transfer security

### Acceptance Criteria
- [ ] All validator signatures are cryptographically verified
- [ ] Signatures must come from active validators in the validator set
- [ ] 2/3+ voting power required for unlock operations
- [ ] Invalid signatures are rejected with clear error messages
- [ ] Comprehensive tests for signature verification
- [ ] Fuzz tests for signature parsing

## Work Log

| Date | Action | Learnings |
|------|--------|-----------|
| 2025-12-02 | Finding identified in security review | CRITICAL - blocks any deployment |

## Resources
- [IBC Light Client Security](https://github.com/cosmos/ibc)
- [Tendermint Signature Verification](https://docs.tendermint.com/v0.37/spec/core/encoding.html)
