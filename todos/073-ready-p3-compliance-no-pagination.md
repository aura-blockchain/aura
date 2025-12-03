---
id: "073"
title: "Compliance Queries No Pagination"
status: ready
priority: p3
category: performance
module: compliance
severity: MEDIUM
source: compliance-audit
---

# Compliance Queries No Pagination

## Problem

GetAll* queries return entire dataset, risking DoS on large datasets.

## Impact

- Memory exhaustion on large queries
- DoS vulnerability
- Poor performance

## Required Fix

Add pagination using Cosmos SDK PageRequest/PageResponse.

## Acceptance Criteria

- [ ] Pagination on all GetAll queries
- [ ] PageRequest parameter support
- [ ] PageResponse in results
- [ ] Tests for large dataset pagination
