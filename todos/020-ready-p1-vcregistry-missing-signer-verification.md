---
id: "020"
title: "Missing Signer Verification in VCRegistry Module"
status: ready
priority: p1
category: security
module: vcregistry
severity: CRITICAL
cvss: 9.0
source: security-audit-report
---

# Missing Signer Verification in VCRegistry Module

## Problem

The `CreatePresentation` function accepts `msg.Creator` without verifying the transaction signer matches. Attackers can create verifiable credential presentations using other users' VCs.

## Affected Files

- `chain/x/vcregistry/keeper/msg_server.go:34-36`

## Vulnerability

```go
// Line 34-36: Only validates non-empty, doesn't verify signer
func (m *MsgServer) CreatePresentation(ctx context.Context, msg *vcregistrypb.MsgCreatePresentation) (*vcregistrypb.MsgCreatePresentationResponse, error) {
    if msg.Creator == "" {  // Only checks non-empty
        return nil, types.ErrInvalidHolderAddress
    }
    // NO verification that tx signer == msg.Creator
```

## Attack Scenario

```go
// Alice has valuable VCs (KYC, accreditation, etc.)
// Attacker creates presentation claiming to be Alice
msg := &MsgCreatePresentation{
    Creator: "alice_address",  // Attacker claims to be Alice
    VcIds: ["alice-kyc-vc", "alice-accreditation-vc"],
}
// Attacker gets presentation with Alice's credentials
// Uses for unauthorized access, fraud, etc.
```

## Impact

- Identity theft
- VC credential abuse
- Unauthorized access to services requiring VCs
- Reputation damage

## Required Fix

```go
func (m *MsgServer) CreatePresentation(ctx context.Context, msg *vcregistrypb.MsgCreatePresentation) (*vcregistrypb.MsgCreatePresentationResponse, error) {
    sdkCtx := sdk.UnwrapSDKContext(ctx)

    // CRITICAL: Verify transaction signer
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

    // Verify holder actually owns all requested VCs
    for _, vcId := range msg.VcIds {
        vc, err := m.keeper.GetVC(sdkCtx, vcId)
        if err != nil {
            return nil, fmt.Errorf("VC %s not found", vcId)
        }

        // Verify msg.Creator is the VC holder
        if vc.Holder != msg.Creator {
            return nil, types.ErrUnauthorized.Wrapf(
                "creator %s does not own VC %s (holder: %s)",
                msg.Creator, vcId, vc.Holder,
            )
        }
    }

    // Rest of function...
}
```

## Acceptance Criteria

- [ ] Signer verification added to CreatePresentation
- [ ] VC ownership verification added
- [ ] All other msg_server functions reviewed for same issue
- [ ] Tests for unauthorized presentation creation rejection
- [ ] Tests for VC ownership verification
