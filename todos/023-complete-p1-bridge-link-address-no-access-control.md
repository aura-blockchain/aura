---
id: "023"
title: "Bridge LinkAddress Missing Access Controls"
status: ready
priority: p1
category: security
module: bridge
severity: HIGH
cvss: 8.3
source: security-audit-report
---

# Bridge LinkAddress Missing Access Controls

## Problem

The `LinkAddress` function (line 361) has no access controls. Any address can link arbitrary addresses together, creating false shared identities.

## Affected Files

- `chain/x/bridge/keeper/msg_server.go:361`

## Vulnerability

```go
func (ms msgServer) LinkAddress(goCtx context.Context, msg *bridgepb.MsgLinkAddress) (*bridgepb.MsgLinkAddressResponse, error) {
    // NO VERIFICATION that caller owns the addresses being linked
    identity := &bridgepb.SharedIdentity{
        Address:         msg.AuraAddress,
        VerifiedAura:    msg.AuraAddress != "",
        VerifiedPaw:     msg.PawAddress != "",
        VerifiedXai:     msg.XaiAddress != "",
        LinkedAddresses: linked,
    }
    // Attacker can link victim addresses to attacker addresses
}
```

## Attack Scenario

```go
// Attacker links their address to wealthy victim's addresses
msg := &MsgLinkAddress{
    AuraAddress: "attacker_address",
    PawAddress: "wealthy_victim_paw",
    XaiAddress: "wealthy_victim_xai",
}
// Attacker now appears to control victim's cross-chain identity
// Can potentially access victim's reputation, credentials, etc.
```

## Impact

- Identity theft across chains
- Reputation hijacking
- Cross-chain fund theft
- False identity claims

## Required Fix

```go
func (ms msgServer) LinkAddress(goCtx context.Context, msg *bridgepb.MsgLinkAddress) (*bridgepb.MsgLinkAddressResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // Verify signer owns at least one of the addresses being linked
    signers := msg.GetSigners()
    if len(signers) == 0 {
        return nil, status.Error(codes.Unauthenticated, "no signers")
    }

    signerStr := signers[0].String()

    // Verify signer owns the Aura address
    if msg.AuraAddress != "" && msg.AuraAddress != signerStr {
        return nil, status.Error(codes.PermissionDenied, "signer must be the Aura address owner")
    }

    // For cross-chain addresses, require signed proof from those chains
    // Option 1: Require signature from the other chain's address
    // Option 2: Use IBC verification
    // Option 3: Multi-step linking with challenge-response

    if msg.PawAddress != "" {
        // Verify proof of PAW address ownership
        if !ms.verifyPawAddressOwnership(ctx, msg.PawAddress, msg.PawProof) {
            return nil, status.Error(codes.PermissionDenied, "cannot verify PAW address ownership")
        }
    }

    if msg.XaiAddress != "" {
        // Verify proof of XAI address ownership
        if !ms.verifyXaiAddressOwnership(ctx, msg.XaiAddress, msg.XaiProof) {
            return nil, status.Error(codes.PermissionDenied, "cannot verify XAI address ownership")
        }
    }

    // Rest of function...
}
```

## Acceptance Criteria

- [ ] Signer verification for Aura address
- [ ] Cross-chain ownership proof verification
- [ ] Tests for unauthorized linking rejection
- [ ] Tests for cross-chain proof verification
