---
id: "074"
title: "Compliance Module Missing Invariants"
status: ready
priority: p3
category: data-integrity
module: compliance
severity: MEDIUM
source: compliance-audit
---

# Compliance Module Missing Invariants

## Problem

No invariants defined for complex compliance state.

## Impact

- Invalid state not detected
- Silent data corruption
- Cannot verify system health

## Required Fix

Implement invariants for:
- All KYC records have valid addresses
- Sanctions results reference existing addresses
- Tax reports have valid year formats
- GDPR consents have valid types

## Acceptance Criteria

- [ ] KYC record invariants
- [ ] Sanctions result invariants
- [ ] Tax report invariants
- [ ] GDPR consent invariants
- [ ] Invariant registration
