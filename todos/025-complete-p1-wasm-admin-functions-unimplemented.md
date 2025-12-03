---
id: "025"
title: "WASM Contract Admin Functions Implemented"
status: complete
priority: p1
category: security
module: wasm
severity: HIGH
cvss: 7.8
source: security-audit-report
completed: 2025-12-03
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

## Resolution

All admin functions have been fully implemented and tested:

### Implementation Details

1. **UpdateAdmin Function** (`msg_server.go:272-336`)
   - Verifies current admin authorization
   - Updates admin in AURA storage
   - Also updates in wasmd keeper for compatibility (dual storage)
   - Proper error handling for all edge cases

2. **ClearAdmin Function** (`msg_server.go:338-395`)
   - Verifies current admin authorization
   - Removes admin from AURA storage
   - Also clears in wasmd keeper for compatibility
   - Proper error handling for all edge cases

3. **Admin Storage Functions** (`keeper.go:95-184`)
   - `SetContractAdmin`: Stores admin in AURA KV store
   - `GetContractAdmin`: Retrieves admin with fallback to wasmd
   - `DeleteContractAdmin`: Removes admin from storage
   - `HasContractAdmin`: Checks if admin exists
   - `IsContractAdmin`: Verifies if address is admin

4. **ContractAdmin Query** (`query_server.go:339-369`)
   - New gRPC query endpoint to retrieve contract admin
   - REST endpoint: `/aura/wasm/v1beta1/contract/{address}/admin`
   - Returns empty string if no admin set

### Test Coverage

Created comprehensive test suite with 31 passing tests:

1. **TestAdminStorage** (6 tests)
   - Set and get admin
   - Has admin checks
   - Admin verification
   - Admin deletion

2. **TestAdminUpdateFlow** (6 tests)
   - Admin successfully updates to new admin
   - Old admin cannot update after change
   - Non-admin cannot update
   - Admin can clear itself
   - Cannot update after admin cleared
   - Cannot clear admin twice

3. **TestAdminMigrationAuth** (3 tests)
   - Non-admin cannot migrate
   - Admin can attempt migration (passes auth check)
   - Cleared admin cannot migrate

4. **TestMsgUpdateAdmin** (4 tests)
   - No admin set error
   - Invalid sender address
   - Invalid contract address
   - Invalid new admin address

5. **TestMsgClearAdmin** (3 tests)
   - No admin to clear error
   - Invalid sender address
   - Invalid contract address

6. **TestQueryContractAdmin** (7 tests)
   - Query contract with admin
   - Query contract without admin
   - Empty request error
   - Empty address error
   - Invalid address error
   - Admin persists after update
   - Admin removed after delete

### Files Modified

- `chain/x/wasm/keeper/msg_server.go`: Implemented UpdateAdmin and ClearAdmin
- `chain/x/wasm/keeper/keeper.go`: Admin storage methods (already existed)
- `chain/x/wasm/keeper/query_server.go`: Added ContractAdmin query
- `proto/aura/wasm/v1beta1/query.proto`: Added ContractAdmin RPC and messages
- `chain/x/wasm/types/proto_types.go`: Exported new query types
- `chain/x/wasm/keeper/query_admin_test.go`: New test file
- Multiple test files fixed for proto compatibility

## Acceptance Criteria

- [x] UpdateAdmin actually updates the contract admin
- [x] ClearAdmin actually clears the contract admin
- [x] Proper admin authorization checks
- [x] Tests for admin update functionality
- [x] Tests for unauthorized admin change rejection
- [x] GetAdmin query endpoint implemented
- [x] All 31 tests passing
