---
id: "057"
title: "Compliance No Jurisdiction-Based Access Control"
status: ready
priority: p2
category: compliance
module: compliance
severity: CRITICAL
source: compliance-audit
---

# Compliance No Jurisdiction-Based Access Control

## Problem

KYC records store jurisdiction but don't enforce jurisdiction-based restrictions. Users from OFAC-sanctioned countries can bypass controls.

## Affected Files

- `chain/x/compliance/keeper/msg_server.go:27-53`

## Impact

- Sanctioned country users can use platform
- OFAC compliance violation
- Regulatory enforcement risk

## Required Fix

```go
var BlockedJurisdictions = map[string]bool{
    "KP": true,  // North Korea
    "IR": true,  // Iran
    "SY": true,  // Syria
    "CU": true,  // Cuba
}

func (s *msgServer) SubmitKYC(...) {
    if BlockedJurisdictions[req.Jurisdiction] {
        return nil, status.Errorf(codes.PermissionDenied,
            "jurisdiction %s is blocked", req.Jurisdiction)
    }
    // ...
}
```

## Acceptance Criteria

- [ ] Blocked jurisdictions list in params
- [ ] Validation on KYC submission
- [ ] Validation on transactions
- [ ] Governance to update list
- [ ] Tests for blocked jurisdiction rejection
