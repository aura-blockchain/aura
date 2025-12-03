---
id: "051"
title: "Compliance No Audit Logging"
status: ready
priority: p2
category: compliance
module: compliance
severity: CRITICAL
source: compliance-audit
---

# Compliance No Audit Logging

## Problem

Events are defined but NOT EMITTED in message handlers. No audit trail for KYC submissions, SAR filings, GDPR requests.

## Affected Files

- `chain/x/compliance/keeper/msg_server.go:27-214`
- `chain/x/compliance/types/events.go`

## Impact

- No audit trail for regulatory investigations
- BSA/AML compliance gap
- Cannot prove compliance actions were taken

## Required Fix

Add event emission to all message handlers:

```go
ctx.EventManager().EmitEvent(
    sdk.NewEvent(
        types.EventTypeKYCSubmitted,
        sdk.NewAttribute(types.AttributeKeyAddress, req.Address),
        sdk.NewAttribute(types.AttributeKeyKYCLevel, req.KycLevel.String()),
        sdk.NewAttribute(types.AttributeKeyProvider, req.Provider),
        sdk.NewAttribute(types.AttributeKeyTimestamp, ctx.BlockTime().String()),
    ),
)
```

## Acceptance Criteria

- [ ] Events emitted in SubmitKYC
- [ ] Events emitted in ReportSuspiciousActivity
- [ ] Events emitted in ScreenSanctions
- [ ] Events emitted in RecordGDPRConsent
- [ ] Events emitted in RequestGDPRData
- [ ] Tests verify event emission
