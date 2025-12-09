---
status: pending
priority: p3
issue_id: "120"
tags: [code-review, quality, logging, observability]
dependencies: ["100"]
---

# P3 MEDIUM: Logging Inconsistent and Often Missing

## Problem Statement

Logging patterns vary across modules - some use structured logging, some use fmt.Printf, and many critical operations have no logging at all.

**Why it matters:** Inconsistent logging makes debugging production issues extremely difficult and slows incident response.

## Findings

### Current Patterns

**1. Structured Logging (Good)**
```go
// Some modules use proper SDK logging
ctx.Logger().Info("swap executed",
    "pool_id", poolID,
    "amount_in", amountIn,
    "amount_out", amountOut,
)
```

**2. Printf Debugging (Bad)**
```go
// Some modules use fmt.Printf
fmt.Printf("DEBUG: processing order %s\n", orderID)
```

**3. No Logging (Critical)**
```go
// Many critical operations have no logging
func (k Keeper) ExecuteTransfer(ctx sdk.Context, ...) error {
    // No logging at all
    return k.bankKeeper.SendCoins(...)
}
```

### Missing Logging Categories

| Category | Status | Impact |
|----------|--------|--------|
| Transaction execution | Partial | Can't trace tx flow |
| Error conditions | Rare | Can't debug failures |
| State transitions | None | Can't audit changes |
| Security events | Partial | Can't detect attacks |
| Performance metrics | None | Can't identify bottlenecks |

## Proposed Solutions

### Solution A: Standardize Structured Logging (Recommended)
**Effort:** 2-3 days | **Risk:** Low

**1. Create logging helpers:**

```go
// chain/pkg/log/log.go
package log

func TxStart(ctx sdk.Context, msgType string, sender string) {
    ctx.Logger().Info("tx_start",
        "msg_type", msgType,
        "sender", sender,
        "block", ctx.BlockHeight(),
    )
}

func TxSuccess(ctx sdk.Context, msgType string, attrs ...interface{}) {
    args := append([]interface{}{"msg_type", msgType}, attrs...)
    ctx.Logger().Info("tx_success", args...)
}

func TxError(ctx sdk.Context, msgType string, err error, attrs ...interface{}) {
    args := append([]interface{}{"msg_type", msgType, "error", err.Error()}, attrs...)
    ctx.Logger().Error("tx_error", args...)
}

func SecurityEvent(ctx sdk.Context, event string, attrs ...interface{}) {
    args := append([]interface{}{"security_event", event}, attrs...)
    ctx.Logger().Warn("security", args...)
}
```

**2. Standard usage pattern:**

```go
func (ms msgServer) Swap(ctx context.Context, msg *types.MsgSwap) (*types.MsgSwapResponse, error) {
    sdkCtx := sdk.UnwrapSDKContext(ctx)

    log.TxStart(sdkCtx, "MsgSwap", msg.Sender)

    result, err := ms.k.ExecuteSwap(sdkCtx, msg)
    if err != nil {
        log.TxError(sdkCtx, "MsgSwap", err,
            "pool_id", msg.PoolId,
            "amount_in", msg.AmountIn,
        )
        return nil, err
    }

    log.TxSuccess(sdkCtx, "MsgSwap",
        "pool_id", msg.PoolId,
        "amount_in", msg.AmountIn,
        "amount_out", result.AmountOut,
    )

    return result, nil
}
```

**3. Security event logging:**

```go
// Failed signature verification
log.SecurityEvent(ctx, "invalid_signature",
    "address", address,
    "reason", "recovery_id_mismatch",
)

// Rate limit exceeded
log.SecurityEvent(ctx, "rate_limit_exceeded",
    "address", address,
    "operation", "bridge_withdrawal",
)
```

## Recommended Action

**GO WITH SOLUTION A**: Standardize logging across all modules.

## Technical Details

### Log Levels Usage

| Level | Usage |
|-------|-------|
| Debug | Detailed internal state (off in production) |
| Info | Normal operations, tx start/success |
| Warn | Security events, unusual conditions |
| Error | Failures, tx errors |

### Files to Create/Modify

- `chain/pkg/log/log.go` - New logging helpers
- All `msg_server.go` files - Add logging
- All keeper files - Add logging for important operations

## Acceptance Criteria

- [ ] Logging helpers package created
- [ ] All message handlers log start/success/error
- [ ] Security events logged consistently
- [ ] No fmt.Printf in production code
- [ ] Critical operations have trace logging
- [ ] Log format is JSON-compatible

## Work Log

| Date | Action | Result |
|------|--------|--------|
| 2025-12-08 | Code quality review identified logging gaps | P3 Medium |

## Resources

- [Cosmos SDK Logging](https://docs.cosmos.network/main/core/context#logging)
- [Structured Logging Best Practices](https://www.honeycomb.io/blog/structured-logging)
