---
id: "066"
title: "Compliance No Rate Limiting on Expensive Operations"
status: ready
priority: p2
category: security
module: compliance
severity: CRITICAL
source: compliance-audit
---

# Compliance No Rate Limiting on Expensive Operations

## Problem

Expensive operations like sanctions screening have no rate limits. Can overwhelm external providers, no cost for queries.

## Affected Files

- `chain/x/compliance/keeper/query_server.go:47-68`

## Impact

- DoS of sanctions providers
- API quota exhaustion
- Service disruption

## Required Fix

Add rate limiting parameters and enforcement.

## Acceptance Criteria

- [ ] Per-address rate limits
- [ ] Per-provider rate limits
- [ ] Optional fees for expensive queries
- [ ] Tests for rate limit enforcement
