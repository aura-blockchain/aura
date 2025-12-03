---
id: "026"
title: "Biometric Authentication Bypass"
status: ready
priority: p1
category: security
module: walletsecurity
severity: HIGH
cvss: 7.5
source: security-audit-report
---

# Biometric Authentication Bypass

## Problem

The `AuthenticateBiometric` function (line 585) has a trivial check that accepts any non-empty proof as valid.

## Affected Files

- `chain/x/walletsecurity/keeper/msg_server.go:585`

## Vulnerability

```go
func (ms msgServer) AuthenticateBiometric(goCtx context.Context, msg *wspb.MsgAuthenticateBiometric) (*wspb.MsgAuthenticateBiometricResponse, error) {
    // Simplified authentication
    authenticated := len(msg.BiometricProof) > 0  // WRONG: Any bytes = authenticated

    return &wspb.MsgAuthenticateBiometricResponse{
        Authenticated: authenticated,  // Always true if proof not empty
    }
}
```

## Attack Scenario

```go
// Attacker sends any random bytes as biometric proof
msg := &MsgAuthenticateBiometric{
    WalletAddress: "victim_wallet",
    BiometricProof: []byte("literally anything"),
}
// Returns: Authenticated: true
// Attacker bypasses biometric security
```

## Impact

- Complete biometric authentication bypass
- Unauthorized wallet access
- Multi-factor authentication failure

## Required Fix

```go
func (ms msgServer) AuthenticateBiometric(goCtx context.Context, msg *wspb.MsgAuthenticateBiometric) (*wspb.MsgAuthenticateBiometricResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // Verify signer matches wallet address
    signers := msg.GetSigners()
    if len(signers) == 0 {
        return nil, status.Error(codes.Unauthenticated, "no signers")
    }

    walletAddr, err := sdk.AccAddressFromBech32(msg.WalletAddress)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "invalid wallet address")
    }

    if !signers[0].Equals(walletAddr) {
        return nil, status.Error(codes.PermissionDenied, "signer does not match wallet")
    }

    // Get registered biometric template
    biometricConfig, err := ms.Keeper.GetBiometricConfig(ctx, msg.WalletAddress)
    if err != nil {
        return nil, status.Error(codes.NotFound, "biometric not configured for this wallet")
    }

    // Validate proof structure
    if len(msg.BiometricProof) < MinBiometricProofSize {
        return nil, status.Error(codes.InvalidArgument, "biometric proof too short")
    }

    // Parse biometric proof
    proof, err := parseBiometricProof(msg.BiometricProof)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "invalid biometric proof format")
    }

    // Verify timestamp to prevent replay
    if time.Since(proof.Timestamp) > BiometricProofMaxAge {
        return nil, status.Error(codes.DeadlineExceeded, "biometric proof expired")
    }

    // Check if this proof was already used (replay protection)
    proofHash := sha256.Sum256(msg.BiometricProof)
    if ms.Keeper.IsBiometricProofUsed(ctx, msg.WalletAddress, proofHash[:]) {
        return nil, status.Error(codes.AlreadyExists, "biometric proof already used")
    }

    // ACTUALLY VERIFY THE BIOMETRIC
    // Option 1: Verify against stored template hash
    // Option 2: Use secure enclave verification
    // Option 3: Use external biometric verification service

    authenticated := ms.verifyBiometricTemplate(
        biometricConfig.TemplateHash,
        proof.BiometricData,
        biometricConfig.Threshold,
    )

    if authenticated {
        // Mark proof as used
        ms.Keeper.MarkBiometricProofUsed(ctx, msg.WalletAddress, proofHash[:])
    }

    return &wspb.MsgAuthenticateBiometricResponse{
        Authenticated: authenticated,
    }, nil
}

func (ms msgServer) verifyBiometricTemplate(templateHash []byte, proofData []byte, threshold float64) bool {
    // Hash the proof data and compare with stored template
    // In production, this would use proper biometric matching algorithms
    proofHash := sha256.Sum256(proofData)
    return bytes.Equal(templateHash, proofHash[:])
}
```

## Acceptance Criteria

- [ ] Actual biometric verification implemented
- [ ] Replay protection with proof tracking
- [ ] Timestamp validation
- [ ] Signer verification
- [ ] Tests for bypass attempts
- [ ] Tests for replay protection
