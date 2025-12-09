# TODO: Fix networksecurity keeper test compilation failure

---
status: pending
priority: p1
issue_id: "001"
tags: [code-review, build-failure, test, networksecurity]
dependencies: []
---

## Problem Statement

The networksecurity keeper test fails to compile due to a reference to a non-existent field `MaxPeers` on the `ConnectionConfig` proto type.

**Impact:** Tests cannot run for the networksecurity module, blocking testnet validation.

## Findings

**Location:** `/home/decri/blockchain-projects/aura/chain/x/networksecurity/keeper/invariants_test.go:142`

**Error:**
```
x/networksecurity/keeper/invariants_test.go:142:20: params.Connection.MaxPeers undefined
(type *v1beta1.ConnectionConfig has no field or method MaxPeers)
```

**Root Cause:** The test references `params.Connection.MaxPeers` but the proto-generated `ConnectionConfig` type has these fields instead:
- `MaxInboundConnections`
- `MaxOutboundConnections`
- `MaxConnectionsPerIp`

The test was likely written against an old proto definition or a planned field that was never added.

## Proposed Solutions

### Option 1: Update test to use existing field (Recommended)
**Pros:** Quick fix, maintains test intent
**Cons:** None
**Effort:** Small (5 minutes)
**Risk:** Low

```go
// Change from:
params.Connection.MaxPeers = 0 // invalid: must be positive

// To:
params.Connection.MaxInboundConnections = 0 // invalid: must be positive
```

### Option 2: Add MaxPeers field to proto
**Pros:** May reflect original design intent
**Cons:** Proto changes require regeneration, may not be needed
**Effort:** Medium (1 hour)
**Risk:** Low

## Recommended Action

Option 1 - Update the test to use an existing field that has the same validation semantics.

## Technical Details

**Affected Files:**
- `chain/x/networksecurity/keeper/invariants_test.go`

**Proto Definition:**
- `proto/aura/networksecurity/v1beta1/networksecurity.proto`

## Acceptance Criteria

- [ ] Test compiles successfully
- [ ] `go test ./chain/x/networksecurity/keeper/...` passes
- [ ] Full test suite passes: `go test ./...`

## Work Log

| Date | Action | Notes |
|------|--------|-------|
| 2025-12-08 | Identified | Found during comprehensive code review |

## Resources

- Proto definition: `proto/aura/networksecurity/v1beta1/networksecurity.pb.go:210`
- Test file: `chain/x/networksecurity/keeper/invariants_test.go`
