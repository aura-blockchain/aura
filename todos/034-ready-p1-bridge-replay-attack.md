---
id: "034"
title: "Bridge Replay Attack Vulnerability"
status: ready
priority: p1
category: security
module: bridge
severity: CRITICAL
cvss: 10.0
source: bridge-security-matrix
---

# Bridge Replay Attack Vulnerability

## Problem

Source transaction hashes can be reused to unlock unlimited tokens. No tracking of processed source transactions.

## Affected Files

- `chain/x/bridge/keeper/msg_server.go`
- `chain/x/bridge/keeper/keeper.go`

## Vulnerability

When `UnlockTokens` is called, the source transaction hash is not tracked as "already processed". An attacker can:
1. Make a legitimate deposit on source chain
2. Call `UnlockTokens` with valid signatures
3. Call `UnlockTokens` again with SAME signatures and hash
4. Repeat indefinitely - unlimited token minting

## Impact

- **Max Loss: UNLIMITED**
- Complete bridge drain
- Token supply inflation
- Cross-chain arbitrage exploit

## Required Fix

```go
// Add storage key for processed source hashes
var ProcessedSourceHashesKey = []byte("processed_source_hashes")

func (k Keeper) IsSourceHashProcessed(ctx sdk.Context, sourceChain, sourceHash string) bool {
    store := ctx.KVStore(k.storeKey)
    key := append(ProcessedSourceHashesKey, []byte(sourceChain+":"+sourceHash)...)
    return store.Has(key)
}

func (k Keeper) MarkSourceHashProcessed(ctx sdk.Context, sourceChain, sourceHash string) {
    store := ctx.KVStore(k.storeKey)
    key := append(ProcessedSourceHashesKey, []byte(sourceChain+":"+sourceHash)...)
    store.Set(key, []byte{1})
}

func (ms msgServer) UnlockTokens(goCtx context.Context, msg *bridgepb.MsgUnlockTokens) (*bridgepb.MsgUnlockTokensResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // CRITICAL: Check if this source hash was already processed
    if ms.Keeper.IsSourceHashProcessed(ctx, msg.SourceChain, msg.BurnTxHash) {
        return nil, status.Error(codes.AlreadyExists, "source transaction already processed (replay attack prevented)")
    }

    // ... existing validation ...

    // CRITICAL: Mark as processed BEFORE minting (prevent reentrancy)
    ms.Keeper.MarkSourceHashProcessed(ctx, msg.SourceChain, msg.BurnTxHash)

    // ... mint tokens ...

    return &bridgepb.MsgUnlockTokensResponse{}, nil
}
```

## Acceptance Criteria

- [ ] Processed source hash tracking implemented
- [ ] Check happens BEFORE any state changes
- [ ] Mark happens BEFORE minting (reentrancy safe)
- [ ] Genesis export/import of processed hashes
- [ ] Tests for replay attack prevention
- [ ] Tests for cross-chain replay prevention
