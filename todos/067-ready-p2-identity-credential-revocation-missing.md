---
id: "067"
title: "Identity Credential Revocation Not Enforced"
status: ready
priority: p2
category: security
module: identity
severity: HIGH
source: identity-privacy-audit
---

# Identity Credential Revocation Not Enforced

## Problem

Revoked credentials can still be used. Revocation status not checked during verification.

## Impact

- Revoked credentials accepted
- Cannot invalidate compromised identities
- Security control bypass

## Required Fix

Check revocation status in all credential verification paths.

## Acceptance Criteria

- [ ] Revocation check in verification
- [ ] Revocation list indexed for fast lookup
- [ ] Batch revocation support
- [ ] Tests for revoked credential rejection
