---
id: "025"
title: "WASM Contract Admin Functions Unimplemented"
status: ready
priority: p1
category: security
module: wasm
severity: HIGH
cvss: 7.8
source: security-audit-report
---

# WASM Contract Admin Functions Unimplemented

## Problem

Critical admin functions `UpdateAdmin` (line 252) and `ClearAdmin` (line 284) only emit events but don't actually change admin. Contracts can never be upgraded or administered.

## Affected Files

- `chain/x/wasm/keeper/msg_server.go:252, 284`

## Vulnerability

```go
func (ms msgServer) UpdateAdmin(goCtx context.Context, msg *types.MsgUpdateAdmin) (*types.MsgUpdateAdminResponse, error) {
    // ... validation ...

    // Note: In production, this would call wasmd keeper to update admin
    // For now, we just emit the event
    ctx.EventManager().EmitEvents(sdk.Events{
        sdk.NewEvent(types.EventTypeUpdateAdmin, ...),
    })

    return &types.MsgUpdateAdminResponse{}, nil  // DOES NOTHING
}
```

## Impact

- Contracts cannot be upgraded
- Security vulnerabilities cannot be patched
- Admin functions non-functional
- Potential for locked/bricked contracts

## Required Fix

```go
func (ms msgServer) UpdateAdmin(goCtx context.Context, msg *types.MsgUpdateAdmin) (*types.MsgUpdateAdminResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // Verify sender is current admin
    signers := msg.GetSigners()
    if len(signers) == 0 {
        return nil, status.Error(codes.Unauthenticated, "no signers")
    }

    contractInfo, err := ms.Keeper.GetContractInfo(ctx, msg.Contract)
    if err != nil {
        return nil, status.Error(codes.NotFound, "contract not found")
    }

    if contractInfo.Admin != signers[0].String() {
        return nil, status.Error(codes.PermissionDenied, "sender is not contract admin")
    }

    // ACTUALLY UPDATE THE ADMIN
    contractInfo.Admin = msg.NewAdmin
    if err := ms.Keeper.SetContractInfo(ctx, msg.Contract, contractInfo); err != nil {
        return nil, status.Error(codes.Internal, "failed to update admin")
    }

    ctx.EventManager().EmitEvents(sdk.Events{
        sdk.NewEvent(
            types.EventTypeUpdateAdmin,
            sdk.NewAttribute(types.AttributeKeyContractAddr, msg.Contract),
            sdk.NewAttribute(types.AttributeKeyNewAdmin, msg.NewAdmin),
        ),
    })

    return &types.MsgUpdateAdminResponse{}, nil
}

func (ms msgServer) ClearAdmin(goCtx context.Context, msg *types.MsgClearAdmin) (*types.MsgClearAdminResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // Verify sender is current admin
    signers := msg.GetSigners()
    if len(signers) == 0 {
        return nil, status.Error(codes.Unauthenticated, "no signers")
    }

    contractInfo, err := ms.Keeper.GetContractInfo(ctx, msg.Contract)
    if err != nil {
        return nil, status.Error(codes.NotFound, "contract not found")
    }

    if contractInfo.Admin != signers[0].String() {
        return nil, status.Error(codes.PermissionDenied, "sender is not contract admin")
    }

    // ACTUALLY CLEAR THE ADMIN
    contractInfo.Admin = ""
    if err := ms.Keeper.SetContractInfo(ctx, msg.Contract, contractInfo); err != nil {
        return nil, status.Error(codes.Internal, "failed to clear admin")
    }

    ctx.EventManager().EmitEvents(sdk.Events{
        sdk.NewEvent(
            types.EventTypeClearAdmin,
            sdk.NewAttribute(types.AttributeKeyContractAddr, msg.Contract),
        ),
    })

    return &types.MsgClearAdminResponse{}, nil
}
```

## Acceptance Criteria

- [ ] UpdateAdmin actually updates the contract admin
- [ ] ClearAdmin actually clears the contract admin
- [ ] Proper admin authorization checks
- [ ] Tests for admin update functionality
- [ ] Tests for unauthorized admin change rejection
