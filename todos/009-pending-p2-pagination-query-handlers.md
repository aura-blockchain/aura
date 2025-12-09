# TODO: Add pagination to all query handlers

---
status: pending
priority: p2
issue_id: "009"
tags: [code-review, performance, scalability, critical]
dependencies: []
---

## Problem Statement

Multiple query handlers return unbounded result sets without pagination, causing DoS vulnerabilities and memory exhaustion at scale.

**Impact:** At 10,000+ records, queries exceed gRPC limits, cause OOM, and degrade node performance.

## Findings

**Affected Query Handlers:**

1. **DEX Module** (`chain/x/dex/keeper/query_server.go`):
```go
func (qs queryServer) AllPools(...) {
    pools := qs.keeper.GetAllPools(sdkCtx)  // Unbounded read
    return &dexpb.QueryAllPoolsResponse{Pools: pools}, nil
}
```

2. **Governance Module** (`chain/x/governance/keeper/query_server.go`):
```go
func (qs queryServer) Proposals(...) {
    proposals := qs.Keeper.GetAllProposals(ctx)  // Unbounded + in-memory filter
    // Filters applied AFTER loading everything
}
```

3. **NetworkSecurity Module** (`chain/x/networksecurity/keeper/query_server.go`):
```go
func (qs queryServer) AllPeers(...) {
    peers := qs.GetAllPeers(ctx)  // Unbounded read
}
```

**Performance Impact:**

| Dataset Size | Current Latency | Memory | Status |
|--------------|-----------------|--------|--------|
| 1,000 items  | ~50ms          | 500KB  | OK     |
| 10,000 items | ~500ms         | 5MB    | SLOW   |
| 100,000 items| ~5s+           | 50MB+  | FAIL   |

## Proposed Solutions

### Option 1: Implement Cosmos SDK pagination (Recommended)
**Pros:** Standard pattern, efficient
**Cons:** Requires proto changes for pagination fields
**Effort:** Medium (2-3 hours per module)
**Risk:** Low

```go
func (qs queryServer) AllPools(ctx context.Context, req *dexpb.QueryAllPoolsRequest) (*dexpb.QueryAllPoolsResponse, error) {
    sdkCtx := sdk.UnwrapSDKContext(ctx)
    store := sdkCtx.KVStore(qs.keeper.storeKey)
    poolStore := prefix.NewStore(store, types.PoolPrefix)

    pools := []*types.LiquidityPool{}
    pageRes, err := query.Paginate(poolStore, req.Pagination, func(key []byte, value []byte) error {
        var pool types.LiquidityPool
        if err := qs.keeper.cdc.Unmarshal(value, &pool); err != nil {
            return err
        }
        pools = append(pools, &pool)
        return nil
    })
    if err != nil {
        return nil, err
    }

    return &dexpb.QueryAllPoolsResponse{
        Pools:      pools,
        Pagination: pageRes,
    }, nil
}
```

## Technical Details

**Modules Requiring Pagination:**
- DEX: AllPools, AllOrders, OrdersByUser
- Governance: Proposals, Votes
- NetworkSecurity: AllPeers, AllAlerts
- Compliance: KYCRecords
- VCRegistry: AllVCs

**Proto Changes:**
Add pagination to request/response messages:
```protobuf
message QueryAllPoolsRequest {
    cosmos.base.query.v1beta1.PageRequest pagination = 1;
}

message QueryAllPoolsResponse {
    repeated LiquidityPool pools = 1;
    cosmos.base.query.v1beta1.PageResponse pagination = 2;
}
```

## Acceptance Criteria

- [ ] All "GetAll" queries support pagination
- [ ] Default page size: 100, max: 1000
- [ ] Proto files updated with pagination fields
- [ ] Performance test: 100k items returns paginated in <50ms
- [ ] Backward compatible (pagination optional)

## Work Log

| Date | Action | Notes |
|------|--------|-------|
| 2025-12-08 | Identified | Performance Oracle agent review |
