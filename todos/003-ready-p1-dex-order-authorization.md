---
status: ready
priority: p1
issue_id: "003"
tags: [code-review, security, dex]
dependencies: []
---

# Missing Authorization in DEX Order Cancellation

## Problem Statement

Order cancellation checks `order.UserAddress != msg.Creator` but does NOT verify that `msg.Creator` is the authenticated transaction signer.

**Why it matters:** An attacker can cancel any user's orders by setting `msg.Creator` to the victim's address, enabling market manipulation and denial of service attacks.

## Findings

### Evidence
- **File:** `chain/x/dex/keeper/msg_server.go`
- **Lines:** 137-153

```go
// PARTIALLY VULNERABLE - Lines 137-153
func (ms msgServer) CancelOrder(goCtx context.Context, msg *dexpb.MsgCancelOrder) (*dexpb.MsgCancelOrderResponse, error) {
    order := ms.keeper.GetOrder(ctx, msg.OrderId)
    if order == nil {
        return nil, fmt.Errorf("order not found")
    }
    // GOOD: Checks ownership
    if order.UserAddress != msg.Creator {
        return nil, fmt.Errorf("cannot cancel order owned by another address")
    }
    // BAD: Never verifies msg.Creator is the TX signer!
    // Attacker can set msg.Creator = victim's address
}
```

### Attack Vector
1. Attacker observes victim's open order
2. Attacker submits CancelOrder with `Creator: victim_address`
3. Order cancelled without victim's consent
4. Attacker can then frontrun with own order

### Impact
- Order cancellation denial of service
- Market manipulation via forced cancellations
- Frontrunning attacks

## Proposed Solutions

### Option A: Verify Signer Matches Creator (Recommended)
**Pros:** Direct fix, matches governance fix pattern
**Cons:** None
**Effort:** Small (30 min - 1 hour)
**Risk:** Low

```go
func (ms msgServer) CancelOrder(goCtx context.Context, msg *dexpb.MsgCancelOrder) (*dexpb.MsgCancelOrderResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // Verify signer matches creator
    signers := msg.GetSigners()
    if len(signers) == 0 {
        return nil, status.Error(codes.Unauthenticated, "no signers")
    }

    creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, err.Error())
    }

    if !creatorAddr.Equals(signers[0]) {
        return nil, status.Error(codes.PermissionDenied, "creator must be transaction signer")
    }

    // ... rest of existing logic
}
```

## Recommended Action
Apply signer verification to all DEX message handlers: CreateOrder, CancelOrder, ModifyOrder, etc.

## Technical Details

### Affected Files
- `chain/x/dex/keeper/msg_server.go` (CancelOrder and likely other handlers)

### Acceptance Criteria
- [ ] CancelOrder verifies signer matches msg.Creator
- [ ] All other DEX message handlers have signer verification
- [ ] Unit tests verify unauthorized cancellation is rejected
- [ ] Integration test confirms legitimate cancellation works

## Work Log

| Date | Action | Learnings |
|------|--------|-----------|
| 2025-12-02 | Finding identified in security review | Pattern matches governance issue |

## Resources
- Related: P1-001 (Governance signer verification)
