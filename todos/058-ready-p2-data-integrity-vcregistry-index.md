---
id: "058"
title: "VCRegistry Index Built Twice During Genesis"
status: ready
priority: p2
category: data-integrity
module: vcregistry
severity: HIGH
source: data-integrity-review
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

- [ ] Index built once from primary data only
- [ ] Exported index used for validation, not building
- [ ] Duplicate detection
- [ ] Tests for index consistency
