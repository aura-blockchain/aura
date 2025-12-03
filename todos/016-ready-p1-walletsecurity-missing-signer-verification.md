---
id: "016"
title: "Missing Signer Verification in Walletsecurity Module"
status: ready
priority: p1
category: security
module: walletsecurity
severity: CRITICAL
cvss: 9.8
source: security-audit-report
---

# Missing Signer Verification in Walletsecurity Module

## Problem

Multiple functions in the walletsecurity module accept `msg.Creator`, `msg.Sender`, or `msg.Signer` without verifying that the transaction signer matches the claimed address. This allows an attacker to impersonate any user.

## Affected Files

- `chain/x/walletsecurity/keeper/msg_server.go`

## Affected Functions (11 total)

1. `RegisterHardwareWallet()` - Line 31
2. `CreateMultiSigWallet()` - Line 79
3. `SignMultiSigTransaction()` - Line 136 (CRITICAL: Signature verification missing)
4. `ConfigureSocialRecovery()` - Line 197
5. `InitiateRecovery()` - Line 242
6. `ApproveRecovery()` - Line 301 (Guardian verification missing)
7. `ExecuteRecovery()` - Line 365
8. `SimulateTransaction()` - Line 417
9. `VerifyDomain()` - Line 433
10. `SetSpendingLimit()` - Line 456
11. `ConfigureSession()` - Line 480

## Attack Scenario

```go
// Attacker creates transaction claiming to be victim
msg := &MsgSignMultiSigTransaction{
    TxId: "victim-tx-123",
    Signer: "victim_address",  // Attacker claims to be victim
    Signature: attacker_signature,  // Attacker's signature
}

// NO VERIFICATION that transaction signer == msg.Signer
// Attacker successfully signs victim's multi-sig transaction
```

## Impact

- Complete authentication bypass
- Unauthorized multi-sig transaction signing
- Wallet configuration tampering
- Social recovery manipulation
- Fund theft

## Required Fix

Add signer verification at the start of EVERY message handler:

```go
func (ms msgServer) SignMultiSigTransaction(goCtx context.Context, msg *wspb.MsgSignMultiSigTransaction) (*wspb.MsgSignMultiSigTransactionResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // CRITICAL: Verify signer matches claimed address
    signers := msg.GetSigners()
    if len(signers) == 0 {
        return nil, status.Error(codes.Unauthenticated, "no signers")
    }

    claimedAddr, err := sdk.AccAddressFromBech32(msg.Signer)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "invalid signer address")
    }

    if !signers[0].Equals(claimedAddr) {
        return nil, status.Error(codes.PermissionDenied, "signer mismatch")
    }

    // Rest of function...
}
```

## Acceptance Criteria

- [ ] All 11 functions have signer verification
- [ ] Tests added for each function verifying rejection of mismatched signers
- [ ] No function accepts unverified msg.Creator/Sender/Signer
