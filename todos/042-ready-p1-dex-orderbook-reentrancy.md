---
id: "042"
title: "DEX Orderbook Reentrancy Vulnerability"
status: ready
priority: p1
category: security
module: dex
severity: CRITICAL
cvss: 9.0
source: dex-security-audit
---

# DEX Orderbook Reentrancy Vulnerability

## Problem

Order matching and execution lacks reentrancy protection. During order execution, state can be manipulated through callbacks.

## Affected Files

- `chain/x/dex/keeper/orderbook.go`
- `chain/x/dex/keeper/msg_server.go`

## Vulnerability

```go
func (k Keeper) ExecuteOrder(ctx sdk.Context, order *types.Order) error {
    // 1. Read state
    pool := k.GetPool(ctx, order.PoolId)

    // 2. Calculate outputs
    output := k.calculateSwap(pool, order.Amount)

    // 3. Transfer tokens (EXTERNAL CALL)
    // If token has callback (e.g., ERC-777 style), attacker can re-enter
    k.bankKeeper.SendCoins(ctx, order.Maker, pool.Address, order.Amount)
    k.bankKeeper.SendCoins(ctx, pool.Address, order.Maker, output)

    // 4. Update state (BUT attacker already re-entered during step 3)
    k.updateOrderStatus(ctx, order.Id, "filled")
}
```

## Attack Scenario

```
1. Attacker places order
2. During SendCoins callback, attacker:
   - Cancels the same order
   - Places new orders
   - Manipulates pool state
3. Original execution continues with stale state
4. Double-spend or price manipulation achieved
```

## Required Fix

```go
// Use the security module's ReentrancyGuard
func (k Keeper) ExecuteOrder(ctx sdk.Context, order *types.Order) error {
    // 1. Acquire reentrancy lock
    if err := k.securityKeeper.EnterNoReentrant(ctx, "orderbook:"+order.PoolId); err != nil {
        return fmt.Errorf("reentrancy detected: %w", err)
    }
    defer k.securityKeeper.ExitNoReentrant(ctx, "orderbook:"+order.PoolId)

    // 2. Read state
    pool := k.GetPool(ctx, order.PoolId)
    if pool == nil {
        return fmt.Errorf("pool not found")
    }

    // 3. Calculate outputs
    output := k.calculateSwap(pool, order.Amount)

    // 4. CHECKS-EFFECTS-INTERACTIONS pattern
    // First: All state changes
    k.updateOrderStatus(ctx, order.Id, "filling")

    // Update pool reserves in memory
    newReserveA := pool.ReserveA.Add(order.Amount)
    newReserveB := pool.ReserveB.Sub(output)

    // 5. Validate invariants BEFORE external calls
    if newReserveB.IsNegative() {
        return fmt.Errorf("insufficient reserves")
    }

    // Verify constant product invariant
    oldK := pool.ReserveA.Mul(pool.ReserveB)
    newK := newReserveA.Mul(newReserveB)
    if newK.LT(oldK) {
        return fmt.Errorf("would violate constant product invariant")
    }

    // 6. Commit state changes (EFFECTS before INTERACTIONS)
    pool.ReserveA = newReserveA
    pool.ReserveB = newReserveB
    k.SetPool(ctx, pool)
    k.updateOrderStatus(ctx, order.Id, "filled")

    // 7. Finally: External calls (INTERACTIONS last)
    if err := k.bankKeeper.SendCoins(ctx, order.Maker, pool.Address, order.Amount); err != nil {
        // Revert state changes
        return fmt.Errorf("transfer failed: %w", err)
    }

    if err := k.bankKeeper.SendCoins(ctx, pool.Address, order.Maker, output); err != nil {
        // Revert both transfers
        k.bankKeeper.SendCoins(ctx, pool.Address, order.Maker, order.Amount)
        return fmt.Errorf("output transfer failed: %w", err)
    }

    return nil
}
```

## Acceptance Criteria

- [ ] ReentrancyGuard used in all order execution
- [ ] Checks-Effects-Interactions pattern followed
- [ ] State changes before external calls
- [ ] Invariant checks before execution
- [ ] Tests for reentrancy attack prevention
- [ ] Tests for callback manipulation
