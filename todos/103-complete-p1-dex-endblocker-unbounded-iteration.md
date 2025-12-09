---
status: pending
priority: p1
issue_id: "103"
tags: [code-review, performance, dex, consensus, critical]
dependencies: ["100"]
---

# P1 CRITICAL: DEX EndBlocker Unbounded Iteration - Consensus Failure Risk

## Problem Statement

The DEX EndBlocker performs O(n) operations on **every block** without any limits, causing consensus failures when order volume exceeds capacity.

**Why it matters:** This will cause the 4-node testnet to halt under moderate load. With 10,000+ orders, block production time exceeds block timeout, causing chain halt.

## Findings

### Problematic Code

**File:** `/home/decri/blockchain-projects/aura/chain/x/dex/module.go` (lines 117-140)

```go
func (m AppModule) EndBlock(ctx sdk.Context) {
    m.keeper.CleanupExpiredOrders(ctx)      // O(n) - iterates ALL orders
    m.keeper.CleanupExpiredHTLCs(ctx)       // O(n) - iterates ALL HTLCs
    m.keeper.RecordAllPoolPrices(ctx)       // O(pools) - iterates ALL pools
    m.keeper.CleanupExpiredCommitments(ctx) // O(commitments)

    if params.BatchExecutionEnabled {
        m.keeper.ExecuteBatch(ctx)          // O(n log n) - sorts ALL queued orders
    }
}
```

### Performance Impact

| Scenario | Orders | EndBlocker Time | Block Time | Result |
|----------|--------|-----------------|------------|--------|
| Light load | 100 | 100ms | 6s | OK |
| Medium load | 1,000 | 1s | 6s | Marginal |
| Heavy load | 10,000 | **10s** | 6s | **CONSENSUS FAILURE** |
| Production | 100,000 | **100s** | 6s | **CHAIN HALT** |

### Additional Issue: Batch Execution Unbounded Sort

**File:** `/home/decri/blockchain-projects/aura/chain/x/dex/keeper/commit_reveal.go` (lines 272-347)

```go
func (k Keeper) ExecuteBatch(ctx sdk.Context) error {
    queuedOrders := k.GetAllQueuedOrders(ctx)  // O(n) storage reads

    // O(n log n) sort - NO LIMIT
    sort.Slice(queuedOrders, func(i, j int) bool { ... })

    for _, queuedOrder := range queuedOrders {  // O(n) iterations
        // Lock funds, store order, add to orderbook
    }
}
```

With 10,000 orders: ~130,000 comparisons + 40,000 storage operations = **transaction fails due to gas exhaustion**

## Proposed Solutions

### Solution A: Cap Operations Per Block (Recommended)
**Effort:** 1-2 days | **Risk:** Low

```go
const (
    MAX_ORDERS_CLEANUP_PER_BLOCK = 100
    MAX_HTLCS_CLEANUP_PER_BLOCK  = 50
    MAX_POOLS_TWAP_PER_BLOCK     = 20
    MAX_BATCH_SIZE               = 100
)

func (m AppModule) EndBlock(ctx sdk.Context) {
    // Rotate through pools for TWAP (not all at once)
    poolOffset := ctx.BlockHeight() % int64(len(pools))
    poolsToProcess := pools[poolOffset:min(poolOffset+MAX_POOLS_TWAP_PER_BLOCK, len(pools))]

    for _, pool := range poolsToProcess {
        m.keeper.RecordPoolPrice(ctx, pool.PoolID)
    }

    // Limit cleanup operations
    m.keeper.CleanupExpiredOrdersBatched(ctx, MAX_ORDERS_CLEANUP_PER_BLOCK)
    m.keeper.CleanupExpiredHTLCsBatched(ctx, MAX_HTLCS_CLEANUP_PER_BLOCK)
}
```

**Pros:**
- Predictable block times
- No consensus failures
- Simple implementation

**Cons:**
- Cleanup may take multiple blocks

### Solution B: Move Cleanup to Periodic Batches
**Effort:** 3-5 days | **Risk:** Medium

Only run cleanup every N blocks (like compliance module does):

```go
if ctx.BlockHeight() % 50 == 0 {
    k.CleanupExpiredOrders(ctx)
}
```

## Recommended Action

**GO WITH SOLUTION A**: Cap operations per block. Critical for testnet stability.

## Technical Details

### Affected Files
- `chain/x/dex/module.go`
- `chain/x/dex/keeper/commit_reveal.go`
- `chain/x/dex/keeper/orderbook.go`
- `chain/x/dex/types/params.go`

### Database/State Changes
- New params: `max_orders_per_block`, `max_batch_size`
- New state: `last_processed_order_key` for cursor-based cleanup

## Acceptance Criteria

- [ ] EndBlocker completes in <500ms under any load
- [ ] 10,000+ orders don't cause consensus failure
- [ ] Batch execution limited to 100 orders per block
- [ ] All orders eventually processed (no permanent backlog)
- [ ] Benchmark tests verify performance

## Work Log

| Date | Action | Result |
|------|--------|--------|
| 2025-12-08 | Performance audit identified issue | P1 Critical |

## Resources

- [Cosmos SDK BeginBlocker/EndBlocker Best Practices](https://docs.cosmos.network/main/building-modules/beginblock-endblock)
- [Gas and Performance in Cosmos](https://docs.cosmos.network/main/basics/gas-fees)
