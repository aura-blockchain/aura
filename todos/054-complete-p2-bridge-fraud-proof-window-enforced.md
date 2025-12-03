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

- [x] Fraud proof window parameter added (FraudProofWindow in params.go)
- [x] Transfers held until window expires (UnlockTokens creates PendingTransfer)
- [x] Fraud proof submission handler (SubmitFraudProof message handler)
- [x] Finalization requires window expiry (FinalizeTransfer checks unlock_time)
- [x] Tests for challenge period (6 comprehensive tests passing)
- [x] EndBlock auto-finalization (ProcessExpiredPendingTransfers)

## Implementation Summary

Successfully implemented fraud proof window enforcement with the following changes:

### 1. Protobuf Updates
- Added `MsgFinalizeTransfer` and `MsgSubmitFraudProof` messages to `tx.proto`
- Both message types support the complete fraud proof workflow
- Regenerated protobuf files with `buf generate`

### 2. Message Handlers (`msg_server.go`)
- `FinalizeTransfer()`: Finalizes pending transfers after fraud proof window expires
  - Checks window expiration (unlock_time <= current_time)
  - Verifies transfer not challenged
  - Unlocks/mints tokens following checks-effects-interactions pattern
  - Distinguishes between native unlocks and wrapped mints

- `SubmitFraudProof()`: Allows fraud proof submission during window
  - Validates fraud proof window still open
  - Creates fraud proof record
  - Marks pending transfer as challenged to prevent finalization
  - Emits events for audit trail

### 3. Keeper Functions (`keeper.go`)
- `ProcessExpiredPendingTransfers()`: Called from EndBlock to auto-finalize
  - Iterates all pending transfers
  - Skips challenged transfers
  - Finalizes expired unchallenged transfers
  - Logs all operations for monitoring
  - Gas-efficient with error handling

- `FinalizeTransfer()`: Core finalization logic
  - Comprehensive security checks
  - Supports both native token unlock and wrapped token minting
  - Updates transfer status to COMPLETED
  - Records minted amounts for rate limiting

### 4. Module Integration (`module.go`)
- Updated `EndBlock()` to call `ProcessExpiredPendingTransfers()`
- Registered new message types in `RegisterInterfaces()`

### 5. Security Features
- Fraud proof window configurable (1 hour to 30 days)
- Pending transfers in escrow during window
- Challenge mechanism blocks finalization
- Auto-finalization prevents stuck transfers
- Complete audit trail via events
- Follows checks-effects-interactions pattern

### 6. Test Coverage
All 6 fraud proof window tests passing:
- `TestFraudProofWindowEnforcement` - Window timing enforcement
- `TestFraudProofChallengeBlocksFinalization` - Challenge prevention
- `TestFraudProofSubmissionAfterWindowFails` - Late challenge rejection
- `TestFraudProofWindowParameter` - Parameter validation
- `TestFraudProofWindow` - Basic workflow
- `TestFraudProofWindowSecurity` - Security scenarios

Plus existing tests for:
- Multiple pending transfers
- Auto-finalization logic
- Fraud proof submission workflow

## Files Modified

- `/proto/aura/bridge/v1beta1/tx.proto` - New message types
- `/chain/x/bridge/keeper/msg_server.go` - Message handlers
- `/chain/x/bridge/keeper/keeper.go` - Core finalization logic
- `/chain/x/bridge/module.go` - EndBlock integration
- `/chain/x/bridge/keeper/fraud_proof_window_test.go` - Comprehensive tests

## Test Results

```
=== RUN   TestFraudProofWindowEnforcement
--- PASS: TestFraudProofWindowEnforcement (0.00s)
=== RUN   TestFraudProofChallengeBlocksFinalization
--- PASS: TestFraudProofChallengeBlocksFinalization (0.00s)
=== RUN   TestFraudProofSubmissionAfterWindowFails
--- PASS: TestFraudProofSubmissionAfterWindowFails (0.00s)
=== RUN   TestFraudProofWindowParameter
--- PASS: TestFraudProofWindowParameter (0.00s)
=== RUN   TestFraudProofWindow
--- PASS: TestFraudProofWindow (0.00s)
=== RUN   TestFraudProofWindowSecurity
--- PASS: TestFraudProofWindowSecurity (0.00s)
PASS
ok      github.com/aequitas/aura/chain/x/bridge/keeper  0.388s
```

## Security Impact

**HIGH SECURITY IMPROVEMENT**

The fraud proof window provides critical protection against:
- ✅ Malicious validator collusion
- ✅ Invalid transfer attestation
- ✅ Replay attacks during finalization
- ✅ Unauthorized token minting
- ✅ Front-running attacks

Users and validators now have a configurable window (default 7 days per params.go) to detect and challenge fraudulent transfers before tokens are released.
