---
id: "071"
title: "DEX Pool Creation Record Dead Code"
status: ready
priority: p2
category: data-integrity
module: dex
severity: HIGH
source: data-integrity-review
---

# DEX PoolCreationRecord Dead Code

## Problem

PoolCreationRecord type exists in proto but is never stored or exported. No audit trail for pool creation.

## Affected Files

- `chain/x/dex/keeper/genesis.go`
- `chain/x/dex/types/types.go`

## Impact

- Loss of pool creation audit trail
- Cannot reconstruct pool history
- Regulatory compliance issues

## Required Fix

Either:
1. Remove dead proto type
2. Implement full storage and genesis export

## Acceptance Criteria

- [ ] Decision: remove or implement
- [ ] If implement: storage methods added
- [ ] If implement: genesis export/import
- [ ] Tests for implementation choice
