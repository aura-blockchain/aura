# TODO: Optimize BeginBlocker unbounded iterations

---
status: pending
priority: p2
issue_id: "010"
tags: [code-review, performance, scalability]
dependencies: []
---

## Problem Statement

BeginBlocker performs full dataset scans on EVERY block, creating O(n) block processing time that won't scale.

**Impact:** At 10,000+ peers, BeginBlock consumes 20%+ of block time. Blocks 1000+ TPS target.

## Findings

**NetworkSecurity Module** (`chain/x/networksecurity/abci.go`):
```go
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
    // ITERATES ALL PEERS EVERY BLOCK
    peers := k.GetAllPeers(ctx)  // O(n) storage read
    for _, peer := range peers {  // O(n) iteration
        k.UpdatePeerUptime(ctx, peer.PeerId)  // O(1) write per peer
    }
}
```

**Compliance Module** (`chain/x/compliance/keeper/begin_blocker.go`):
```go
func (k *Keeper) BeginBlocker(ctx sdk.Context) {
    // Iterate ALL KYC records EVERY block
    k.IterateKYCRecords(ctx, func(record types.KYCRecord) bool {
        if currentTime.After(record.ExpiresAt.AsTime()) {
            // emit event
        }
        return false
    })
}
```

**Performance Projection:**

| Peer Count | BeginBlock Time | Daily Overhead |
|------------|-----------------|----------------|
| 100        | ~10ms          | ~3 minutes     |
| 1,000      | ~100ms         | ~29 minutes    |
| 10,000     | ~1s            | ~5 hours       |

## Proposed Solutions

### Option 1: Index-based sparse updates (Recommended)
**Pros:** O(1) amortized per block
**Cons:** Requires index management
**Effort:** Medium (1 week)
**Risk:** Low

```go
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
    currentHeight := ctx.BlockHeight()

    // Only update peers that need uptime updates (sparse index)
    peersToUpdate := k.GetPeersRequiringUpdate(ctx, currentHeight)
    for _, peerID := range peersToUpdate {
        k.UpdatePeerUptime(ctx, peerID)
    }

    // Process only expired records (indexed by expiry time)
    expiredRecords := k.GetRecordsExpiringAtHeight(ctx, currentHeight)
    for _, record := range expiredRecords {
        k.ProcessExpiry(ctx, record)
    }
}
```

### Option 2: Batch updates every N blocks
**Pros:** Simple implementation
**Cons:** Less responsive
**Effort:** Small (1 day)
**Risk:** Low

```go
if ctx.BlockHeight() % 100 == 0 {
    k.BatchUpdateAllPeers(ctx)
}
```

## Acceptance Criteria

- [ ] BeginBlock time < 50ms with 10,000 peers
- [ ] Expiry index for time-based lookups
- [ ] Peer update index for sparse updates
- [ ] Benchmark tests verify performance

## Work Log

| Date | Action | Notes |
|------|--------|-------|
| 2025-12-08 | Identified | Performance Oracle agent review |
