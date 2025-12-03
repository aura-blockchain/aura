---
id: "062"
title: "Compliance Messages Missing GetSigners Implementation"
status: ready
priority: p2
category: security
module: compliance
severity: HIGH
source: compliance-audit
---

# Compliance Messages Missing GetSigners Implementation

## Problem

Protobuf messages don't implement GetSigners() method required for Cosmos SDK transaction verification.

## Affected Files

- `proto/aura/compliance/v1beta1/compliance.proto:285-371`

## Impact

- Cannot verify message signatures
- Transaction security compromised
- SDK validation bypassed

## Required Fix

Add cosmos.msg.v1.signer option to proto messages OR implement GetSigners() manually.

## Acceptance Criteria

- [ ] GetSigners() implemented for all message types
- [ ] ValidateBasic() implemented for all messages
- [ ] Tests for signer verification
