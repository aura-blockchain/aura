---
id: "071"
title: "DEX Pool Creation Record Dead Code"
status: complete
priority: p2
category: data-integrity
module: dex
severity: HIGH
source: data-integrity-review
completed: 2025-12-03
---

# DEX PoolCreationRecord Dead Code

## Problem

PoolCreationRecord was fully implemented but being called twice in the pool creation path, creating duplicate records.

## Affected Files

- `chain/x/dex/keeper/liquidity_pool.go`
- `chain/x/dex/keeper/security_integration.go`
- `chain/x/dex/keeper/pool_creation_record_test.go`

## Impact

- Duplicate audit trail records when creating pools
- Redundant storage writes
- Confusion about which layer handles audit recording

## Solution Implemented

Fixed duplicate recording issue:
1. Removed `RecordPoolCreation()` call from `CreatePool()`
2. Kept recording in `SecureCreatePool()` wrapper (correct layer)
3. Updated integration test to use `SecureCreatePool()`
4. Added clear documentation about separation of concerns

## Acceptance Criteria

- [x] Decision: keep implementation, fix duplication
- [x] Storage methods verified (working correctly)
- [x] Genesis export/import verified (working correctly)
- [x] Tests updated and passing (all DEX tests pass)
