---
id: "069"
title: "Identity Attribute Access Control Missing"
status: ready
priority: p2
category: security
module: identity
severity: HIGH
source: identity-privacy-audit
---

# Identity Attribute Access Control Missing

## Problem

Identity attributes can be read by anyone. No selective disclosure or access control.

## Impact

- Privacy violation
- Cannot control who sees what
- All or nothing access

## Required Fix

Implement attribute-level access control:
- Selective disclosure
- Per-attribute permissions
- Consent-based sharing

## Acceptance Criteria

- [ ] Attribute-level access control
- [ ] Selective disclosure support
- [ ] Consent tracking
- [ ] Tests for access control
