---
id: "072"
title: "Bridge No Validator Slashing"
status: ready
priority: p2
category: security
module: bridge
severity: CRITICAL
source: bridge-security-matrix
---

# Bridge No Validator Slashing

## Problem

Malicious validators face no punishment. No economic deterrent for signing fraudulent transfers.

## Impact

- No fraud deterrent
- Validators can collude risk-free
- Bridge security relies only on honesty

## Required Fix

Implement slashing for:
- Signing fraudulent transfers
- Being offline during critical operations
- Double-signing

## Acceptance Criteria

- [ ] Slashing conditions defined
- [ ] Slashing amounts configurable
- [ ] Integration with staking/slashing module
- [ ] Evidence submission mechanism
- [ ] Tests for slashing scenarios
