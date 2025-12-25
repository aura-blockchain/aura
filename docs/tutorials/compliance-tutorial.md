# Compliance Operations Tutorial

## Overview
KYC/AML compliance and regulatory operations.

## Register as KYC Provider

```bash
aurad tx compliance register-provider \
  --name "Acme KYC" \
  --jurisdiction US \
  --license-number ABC123 \
  --from provider
```

## Submit KYC Verification

```bash
aurad tx compliance verify-kyc \
  --subject aura1user... \
  --level enhanced \
  --expiry 2026-01-01 \
  --from provider
```

## Query KYC Status

```bash
aurad query compliance kyc-status aura1user...
```

## AML Check

```bash
aurad tx compliance aml-check \
  --address aura1user... \
  --transaction-amount 10000uaura \
  --from compliance-officer
```

## Sanctions Screening

```bash
aurad query compliance sanctions-check aura1user...
```

## GDPR Operations

### Request Data Export
```bash
aurad tx compliance gdpr-export \
  --subject aura1user... \
  --from user
```

### Request Data Deletion
```bash
aurad tx compliance gdpr-delete \
  --subject aura1user... \
  --from user
```

## Jurisdiction Rules

```bash
aurad query compliance jurisdiction-rules US
aurad query compliance allowed-operations aura1user...
```
