---
id: "064"
title: "Compliance KYC Expiry Never Checked"
status: ready
priority: p2
category: compliance
module: compliance
severity: HIGH
source: compliance-audit
---

# Compliance KYC Expiry Not Enforced

## Problem

KYC records have expires_at field but expiry is never checked or enforced. Expired KYC still considered valid.

## Impact

- Expired verifications accepted
- Re-verification not triggered
- Compliance gap

## Required Fix

Add KYC status validation that checks expiry before allowing operations.

## Acceptance Criteria

- [ ] Expiry check in ValidateKYCStatus
- [ ] BeginBlocker to process expired records
- [ ] Events for expiring KYC
- [ ] Tests for expiry enforcement
