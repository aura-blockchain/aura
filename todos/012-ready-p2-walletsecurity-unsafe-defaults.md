---
status: ready
priority: p2
issue_id: "012"
tags: [code-review, security, genesis]
dependencies: []
---

# WalletSecurity Genesis Has Unsafe Default Parameters

## Problem Statement

The `DefaultGenesisState()` creates genesis with security-critical parameters set to unsafe defaults that could enable attacks.

**Why it matters:** A chain initialized with these defaults has unlimited spending, no phishing protection, and minimal dust attack protection.

## Findings

### Evidence
- **File:** `chain/x/walletsecurity/types/genesis.go`
- **Lines:** 12-52

```go
BiometricEnabled:             false,  // Biometric auth disabled by default
RequireDomainVerification:    false,  // Domain verification disabled
DefaultDailyLimit:            "0",    // Zero spending limit = unlimited!
MinDustAmount:                "1",    // Minimal dust protection
```

### Impact
- **Unlimited spending** (DailyLimit = "0" means no limit)
- **No phishing protection** (RequireDomainVerification = false)
- **Minimal dust attack protection**

## Proposed Solutions

### Option A: Set Secure Defaults (Recommended)
**Pros:** Secure by default
**Cons:** May need adjustment for specific deployments
**Effort:** Small (30 min)
**Risk:** Low

```go
DefaultDailyLimit:            "1000000000", // 1000 tokens (reasonable default)
RequireDomainVerification:    true,         // Security by default
BiometricEnabled:             true,         // Enable if platform supports
MinDustAmount:                "1000",       // Higher threshold
```

## Recommended Action
Change defaults to secure values. Document that operators can adjust for their needs.

## Technical Details

### Affected Files
- `chain/x/walletsecurity/types/genesis.go:12-52`

### Acceptance Criteria
- [ ] DefaultDailyLimit is a reasonable positive value
- [ ] RequireDomainVerification defaults to true
- [ ] MinDustAmount provides meaningful protection
- [ ] Documentation explains default choices

## Work Log

| Date | Action | Learnings |
|------|--------|-----------|
| 2025-12-02 | Finding identified in data integrity review | Security-critical defaults |
