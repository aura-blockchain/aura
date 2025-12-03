---
id: "070"
title: "DEX Front-Running - No Commit-Reveal"
status: ready
priority: p2
category: security
module: dex
severity: HIGH
source: dex-security-audit
---

# DEX Front-Running Vulnerability

## Problem

Orders submitted in plain text. Validators/MEV bots can front-run orders.

## Impact

- Value extraction by front-runners
- Worse execution prices for users
- MEV exploitation

## Required Fix

Implement commit-reveal scheme for large orders:
1. Commit phase: Hash of order
2. Reveal phase: Actual order details
3. Batch execution

## Acceptance Criteria

- [ ] Commit-reveal mechanism
- [ ] Configurable threshold for commit-reveal
- [ ] Batch auction for committed orders
- [ ] Tests for front-running resistance
