---
id: "054"
title: "Bridge Fraud Proof Window Enforced"
status: complete
priority: p2
category: security
module: bridge
severity: HIGH
source: bridge-security-matrix
completed: 2025-12-03
---

# Bridge Fraud Proof Window Not Enforced

## Problem

Transfer finalization has no fraud proof window. Malicious transfers finalized immediately without time for challenges.

## Affected Files

- `chain/x/bridge/keeper/msg_server.go`
- `chain/x/bridge/types/params.go`

## Impact

- No time to detect and challenge fraudulent transfers
- Fraud proofs useless without challenge period
- Cannot recover from validator collusion

## Required Fix

```go
// Add to params
FraudProofWindow time.Duration  // e.g., 1 hour

func (ms msgServer) UnlockTokens(...) {
    // Create pending transfer
    transfer := &types.PendingTransfer{
        // ...
        UnlockTime: ctx.BlockTime().Add(params.FraudProofWindow),
    }
    k.SetPendingTransfer(ctx, transfer)
    // Don't mint yet - wait for fraud proof window
}

// Separate finalization function
func (ms msgServer) FinalizeTransfer(ctx sdk.Context, transferId string) error {
    transfer := k.GetPendingTransfer(ctx, transferId)

    // Check fraud proof window expired
    if ctx.BlockTime().Before(transfer.UnlockTime) {
        return fmt.Errorf("fraud proof window not expired")
    }

    // Check no fraud proof submitted
    if k.HasFraudProof(ctx, transferId) {
        return fmt.Errorf("transfer challenged with fraud proof")
    }

    // Now safe to mint
    k.MintTokens(ctx, transfer)
}
```

## Acceptance Criteria

- [ ] Fraud proof window parameter added
- [ ] Transfers held until window expires
- [ ] Fraud proof submission handler
- [ ] Finalization requires window expiry
- [ ] Tests for challenge period
