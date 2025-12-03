---
id: "017"
title: "Social Recovery Guardian Verification Missing"
status: ready
priority: p1
category: security
module: walletsecurity
severity: CRITICAL
cvss: 9.5
source: security-audit-report
---

# Social Recovery Guardian Verification Missing

## Problem

The `ApproveRecovery` function (line 301) does not verify that `msg.Guardian` is actually in the guardians list for the wallet being recovered. Any address can approve a recovery request.

## Affected Files

- `chain/x/walletsecurity/keeper/msg_server.go:301`

## Attack Scenario

```go
// Attacker approves recovery for victim's wallet
msg := &MsgApproveRecovery{
    RequestId: "victim-recovery-123",
    Guardian: "attacker_address",  // Attacker not actually a guardian
}

// NO CHECK that attacker_address is in config.Guardians
// Recovery gets approved and executed
// Attacker steals wallet
```

## Impact

- Complete wallet takeover
- Bypass of social recovery security
- Fund theft
- Identity theft

## Required Fix

```go
func (ms msgServer) ApproveRecovery(goCtx context.Context, msg *wspb.MsgApproveRecovery) (*wspb.MsgApproveRecoveryResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // Verify transaction signer matches msg.Guardian
    signers := msg.GetSigners()
    if len(signers) == 0 {
        return nil, status.Error(codes.Unauthenticated, "no signers")
    }

    guardianAddr, err := sdk.AccAddressFromBech32(msg.Guardian)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "invalid guardian address")
    }

    if !signers[0].Equals(guardianAddr) {
        return nil, status.Error(codes.PermissionDenied, "signer is not the claimed guardian")
    }

    // Get recovery request
    request, err := ms.Keeper.GetRecoveryRequest(ctx, msg.RequestId)
    if err != nil {
        return nil, status.Error(codes.NotFound, "recovery request not found")
    }

    // Get recovery config
    configBytes, err := ms.Keeper.GetSocialRecoveryConfig(ctx, request.WalletId)
    if err != nil {
        return nil, status.Error(codes.NotFound, "recovery config not found")
    }

    var config wspb.SocialRecoveryConfig
    ms.Keeper.cdc.Unmarshal(configBytes, &config)

    // CRITICAL: Verify guardian is authorized
    isAuthorized := false
    for _, authorizedGuardian := range config.Guardians {
        if authorizedGuardian == msg.Guardian {
            isAuthorized = true
            break
        }
    }
    if !isAuthorized {
        return nil, status.Error(codes.PermissionDenied, "not an authorized guardian")
    }

    // Check for duplicate approval
    for _, existingApproval := range request.Approvals {
        if existingApproval == msg.Guardian {
            return nil, status.Error(codes.AlreadyExists, "guardian already approved")
        }
    }

    // Rest of function...
}
```

## Acceptance Criteria

- [ ] Guardian authorization check added
- [ ] Duplicate approval prevention added
- [ ] Signer verification added
- [ ] Tests for unauthorized guardian rejection
- [ ] Tests for duplicate approval rejection
