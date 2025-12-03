---
id: "063"
title: "Compliance AML Profiles Never Updated"
status: ready
priority: p2
category: compliance
module: compliance
severity: HIGH
source: compliance-audit
---

# Compliance AML Profiles Never Updated

## Problem

AML profiles are stored but never updated. No automatic risk reassessment, no transaction volume tracking, no PEP status updates.

## Affected Files

- `chain/x/compliance/keeper/keeper_kvstore.go:76-116`

## Impact

- Static risk profiles don't reflect current behavior
- No continuous monitoring
- AML compliance gap

## Required Fix

Update AML profiles on each transaction with:
- Transaction count increment
- Volume tracking
- Risk level reassessment
- PEP status checks

## Acceptance Criteria

- [ ] Profile updated on transactions
- [ ] Risk level recalculated
- [ ] Volume and count tracked
- [ ] Events emitted on risk changes
- [ ] Tests for profile updates
