---
id: "050"
title: "Compliance Sanctions Cache Never Expires"
status: ready
priority: p2
category: security
module: compliance
severity: CRITICAL
source: compliance-audit
---

# Compliance Sanctions Cache Never Expires

## Problem

Sanctions screening cache never expires. Cached "CLEAR" status persists forever even if address is added to OFAC SDN list.

## Affected Files

- `chain/x/compliance/keeper/keeper.go:24`
- `chain/x/compliance/keeper/msg_server.go:92-106`

## Impact

- Sanctioned addresses continue transacting using stale "CLEAR" status
- OFAC compliance violation
- Regulatory enforcement risk

## Required Fix

```go
// Add cache expiry check
func (s *msgServer) ScreenSanctions(ctx sdk.Context, address string) (*types.SanctionsResult, error) {
    params := s.Keeper.GetParams(ctx)

    // Check cache with expiry
    cached, err := s.Keeper.GetSanctionsResult(ctx, address)
    if err == nil && cached != nil {
        age := ctx.BlockTime().Sub(cached.ScreenedAt.AsTime())
        maxAge := time.Duration(params.ScreeningCacheHours) * time.Hour

        if age > maxAge {
            cached = nil  // Cache expired, force refresh
        }
    }

    if cached == nil {
        // Perform fresh screening
        cached, err = s.performFreshScreening(ctx, address)
    }

    return cached, err
}
```

## Acceptance Criteria

- [ ] Cache expiry implemented
- [ ] Configurable cache duration in params
- [ ] BeginBlocker refresh of stale entries
- [ ] Tests for cache expiry
