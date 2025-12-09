# TODO: Implement Identity module gRPC services

---
status: pending
priority: p2
issue_id: "012"
tags: [code-review, architecture, grpc, identity]
dependencies: []
---

## Problem Statement

The Identity module lacks gRPC services (`msg_server.go`, `query_server.go`). Users cannot query identity data via standard gRPC endpoints.

**Impact:** Cannot query DIDs via gRPC (CLI works but not programmatic access).

## Findings

**Missing Files:**
- `/home/decri/blockchain-projects/aura/chain/x/identity/keeper/msg_server.go`
- `/home/decri/blockchain-projects/aura/chain/x/identity/keeper/query_server.go`

**Proto Definitions Ready:** 20 RPC methods already defined in proto.

**Comparison:** VCRegistry has same issue (missing `query_server.go`).

## Proposed Solutions

### Option 1: Implement complete gRPC services (Recommended)
**Pros:** Full API compatibility
**Cons:** Time investment
**Effort:** Medium (4-6 hours)
**Risk:** Low

```go
// msg_server.go
type msgServer struct {
    Keeper
}

func NewMsgServer(k Keeper) types.MsgServer {
    return &msgServer{Keeper: k}
}

func (ms msgServer) CreateDID(ctx context.Context, msg *types.MsgCreateDID) (*types.MsgCreateDIDResponse, error) {
    sdkCtx := sdk.UnwrapSDKContext(ctx)
    // Implementation
}

// query_server.go
type queryServer struct {
    Keeper
}

func NewQueryServer(k Keeper) types.QueryServer {
    return &queryServer{Keeper: k}
}

func (qs queryServer) DID(ctx context.Context, req *types.QueryDIDRequest) (*types.QueryDIDResponse, error) {
    // Implementation
}
```

## Acceptance Criteria

- [ ] msg_server.go implements all Msg handlers
- [ ] query_server.go implements all Query handlers
- [ ] Services registered in module.go
- [ ] gRPC queries work via `grpcurl`
- [ ] Unit tests for all handlers

## Work Log

| Date | Action | Notes |
|------|--------|-------|
| 2025-12-08 | Identified | Architecture Strategist agent review |
