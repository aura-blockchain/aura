---
id: "049"
title: "Compliance No Data Encryption at Rest"
status: ready
priority: p2
category: security
module: compliance
severity: CRITICAL
source: compliance-audit
---

# Compliance No Data Encryption at Rest

## Problem

All compliance data stored in plaintext in KVStore: KYC records, AML profiles, suspicious activity reports, tax information, GDPR consent records.

## Affected Files

- `chain/x/compliance/keeper/keeper_kvstore.go:30-521`

## Impact

- Node operators can read all compliance data
- Blockchain explorers may expose sensitive information
- GDPR Article 32 violation (security of processing)

## Required Fix

Implement encryption layer for sensitive data storage.

## Acceptance Criteria

- [ ] Encryption service interface defined
- [ ] AES-256-GCM encryption implemented
- [ ] All sensitive data encrypted before storage
- [ ] Key management system
- [ ] Tests for encryption/decryption
