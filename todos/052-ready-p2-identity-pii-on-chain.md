---
id: "052"
title: "Identity Module Stores PII On-Chain"
status: ready
priority: p2
category: compliance
module: identity
severity: CRITICAL
source: identity-privacy-audit
---

# Identity Module Stores PII On-Chain

## Problem

Identity records store personal information directly on the immutable blockchain, including names, emails, and biometric hashes.

## Affected Files

- `chain/x/identity/types/identity.proto`
- `chain/x/identity/keeper/keeper.go`

## Impact

- GDPR Right to Erasure violation
- Privacy breach risk
- Cannot delete user data

## Required Fix

Move PII to off-chain encrypted storage, store only commitments on-chain.

## Acceptance Criteria

- [ ] PII moved off-chain
- [ ] On-chain only stores hashes/commitments
- [ ] Erasure endpoint implemented
- [ ] Migration for existing data
