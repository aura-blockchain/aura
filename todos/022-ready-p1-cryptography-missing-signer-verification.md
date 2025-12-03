---
id: "022"
title: "Missing Signer Verification in Cryptography Module"
status: ready
priority: p1
category: security
module: cryptography
severity: HIGH
cvss: 8.5
source: security-audit-report
---

# Missing Signer Verification in Cryptography Module

## Problem

Cryptography module functions lack signer verification, allowing attackers to manipulate key rotation schedules, ZK proof circuits, and quantum-resistant keys for any user.

## Affected Files

- `chain/x/cryptography/keeper/msg_server.go`

## Affected Functions

1. `CreateKeyRotationSchedule()` - Line 25
2. `RotateKey()` - Line 40
3. `RegisterZKProofCircuit()` - Line 105
4. `SubmitZKProof()` - Line 120
5. `RegisterSecureEnclave()` - Line 138
6. `GenerateQuantumResistantKey()` - Line 153
7. `AddCertificatePin()` - Line 174

## Attack Scenario

```go
// Attacker rotates victim's encryption keys
msg := &MsgRotateKey{
    Creator: "victim_address",
    KeyId: "victim-encryption-key",
    NewPublicKey: attacker_controlled_key,
}
// Victim's encrypted data now accessible to attacker
```

## Impact

- Cryptographic key compromise
- Encrypted data exposure
- ZK proof manipulation
- Certificate pinning bypass

## Required Fix

Add signer verification to all functions and verify ownership of keys/circuits being modified:

```go
func (ms msgServer) RotateKey(goCtx context.Context, msg *types.MsgRotateKey) (*types.MsgRotateKeyResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // Verify signer matches creator
    signers := msg.GetSigners()
    if len(signers) == 0 {
        return nil, status.Error(codes.Unauthenticated, "no signers")
    }

    creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "invalid creator address")
    }

    if !signers[0].Equals(creatorAddr) {
        return nil, status.Error(codes.PermissionDenied, "signer does not match creator")
    }

    // Verify key ownership
    existingKey, err := ms.Keeper.GetKey(ctx, msg.KeyId)
    if err != nil {
        return nil, status.Error(codes.NotFound, "key not found")
    }

    if existingKey.Owner != msg.Creator {
        return nil, status.Error(codes.PermissionDenied, "not key owner")
    }

    // Rest of function...
}
```

## Acceptance Criteria

- [ ] Signer verification on all 7 functions
- [ ] Key/circuit/enclave ownership verification
- [ ] Tests for unauthorized key rotation rejection
- [ ] Tests for unauthorized ZK circuit manipulation rejection
