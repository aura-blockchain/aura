---
id: "058"
title: "VCRegistry Index Built Twice During Genesis"
status: done
priority: p2
category: data-integrity
module: vcregistry
severity: HIGH
source: data-integrity-review
resolved_commit: 22a00a0
---

# VCRegistry User Presentation Index Built Incorrectly

## Problem

During genesis import, indexes are built from both primary data AND exported indexes, resulting in duplicates.

## Affected Files

- `chain/x/vcregistry/keeper/genesis.go:85-94, 210-224`

## Impact

- Duplicate entries in indexes
- Query results inconsistent
- Cannot trust index data

## Required Fix

Only build indexes from primary data, validate against exported indexes.

## Acceptance Criteria

- [x] Index built once from primary data only
- [x] Exported index used for validation, not building
- [x] Duplicate detection
- [x] Tests for index consistency

## Resolution

Fixed in commit 22a00a0. Changes implemented:

1. **User Presentation Index (lines 87-116)**:
   - Build index from presentations only (via appendUserPresentation)
   - Validate exported index matches built index
   - Error if count or content mismatches

2. **User Attribute VC Index (lines 129-158)**:
   - Build index from attribute VCs only (via appendUserAttributeVC)
   - Validate exported index matches built index
   - Error if count or content mismatches

3. **Pending Disclosure Index (lines 184-206)**:
   - Import explicitly (not auto-built from requests)
   - Validate imported index after building
   - Added documentation explaining why this differs

4. **Comprehensive Tests Added**:
   - TestGenesisIndexNoDuplicates with 5 test cases
   - Tests duplicate prevention for presentations
   - Tests duplicate prevention for attributes
   - Tests validation detects mismatches
   - Tests validation detects count mismatches
   - Tests pending disclosure index validation

All acceptance criteria met. Index integrity guaranteed during genesis import/export.
