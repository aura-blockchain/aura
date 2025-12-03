---
id: "055"
title: "GDPR Consent Withdrawal Not Enforced"
status: ready
priority: p2
category: compliance
module: compliance
severity: CRITICAL
source: compliance-audit
---

# GDPR Consent Withdrawal Not Enforced

## Problem

Consent withdrawal can be recorded but there's no enforcement mechanism to stop processing or delete data when consent is withdrawn.

## Affected Files

- `chain/x/compliance/keeper/msg_server.go:142-165`

## Impact

- GDPR Article 7(3) violation
- Data processing continues after withdrawal
- Cannot demonstrate compliance

## Required Fix

When consent is withdrawn, actually enforce it by:
1. Marking data as "do not process"
2. Triggering data deletion where legally allowed
3. Emitting events for downstream systems

## Acceptance Criteria

- [ ] Consent withdrawal triggers data flagging
- [ ] Processing halted for withdrawn consent
- [ ] Data deletion triggered where allowed
- [ ] Event emission for audit
- [ ] Tests for enforcement
