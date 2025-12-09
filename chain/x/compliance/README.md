# Compliance Module

## Overview

The Compliance module provides KYC/AML, sanctions screening, GDPR compliance, and tax reporting features. It implements privacy-preserving verification with on-chain commitments and off-chain PII storage, rate limiting for expensive operations, and comprehensive audit trails for regulatory requirements.

## Features

- **KYC Verification**: Multi-level identity verification (None, Basic, Intermediate, Advanced)
- **AML Risk Profiling**: Transaction monitoring and risk assessment
- **Sanctions Screening**: OFAC/EU/UN sanctions list checking with caching
- **Transaction Monitoring**: Rule-based alerts for suspicious activity
- **GDPR Compliance**: Data access, rectification, erasure, and portability
- **Tax Reporting**: Automated tax document generation (1099, 8949, etc.)
- **Privacy-Preserving**: PII stored off-chain, only SHA-256 commitments on-chain
- **Rate Limiting**: DoS protection for expensive screening operations
- **Audit Trail**: Complete history of KYC updates and data access requests

## State

### KYCRecord
- **Address**: User blockchain address
- **KYC Level**: None, Basic, Intermediate, Advanced
- **Provider**: KYC service provider address
- **Verified At**: Timestamp of verification
- **Expires At**: Expiration date
- **PII Commitment**: SHA-256 hash of off-chain PII data
- **Jurisdiction**: ISO 3166-1 alpha-2 country code
- **Version**: Auto-incremented on updates

### AMLProfile
- **Risk Level**: Low, Medium, High, Severe
- **Risk Factors**: Array of identified risk indicators
- **Total Volume**: Cumulative transaction volume
- **Suspicious Activities**: List of flagged transactions
- **PEP Status**: Politically Exposed Person flag

### SanctionsScreeningResult
- **Status**: Clear, Match, Confirmed, Pending Review
- **Matches**: Potential sanctions list hits
- **Screened At**: Cache timestamp
- **Requires Manual Review**: Flag for compliance team

### TransactionAlert
- **Rule ID**: Monitoring rule that triggered
- **Risk Level**: Low, Medium, High, Critical
- **Reviewed**: Manual review completed flag
- **Resolution**: escalate, dismiss, file_sar

## Messages

### MsgSubmitKYC
Submit KYC verification (provider only).

**Example**:
```json
{
  "address": "aura1...",
  "kyc_level": "KYC_LEVEL_INTERMEDIATE",
  "provider": "aura1provider...",
  "pii_commitment": "sha256_hash_of_pii_json",
  "jurisdiction": "US"
}
```

### MsgReportSuspiciousActivity
Report suspicious transaction (compliance officer).

**Example**:
```json
{
  "reporter": "aura1officer...",
  "address": "aura1suspect...",
  "transaction_hash": "tx_hash",
  "activity_type": "structuring",
  "description": "Multiple transactions just below reporting threshold",
  "indicators": ["velocity", "round_amounts", "pattern_matching"]
}
```

**Response**:
```json
{
  "activity_id": "sar_abc123"
}
```

### MsgScreenSanctions
Screen address against sanctions lists.

**Example**:
```json
{
  "address": "aura1...",
  "force_refresh": false
}
```

**Response**:
```json
{
  "status": "SANCTIONS_CLEAR",
  "requires_review": false
}
```

### MsgRecordGDPRConsent
Record user consent (GDPR Article 6).

**Example**:
```json
{
  "address": "aura1...",
  "consent_type": "data_processing",
  "consented": true,
  "consent_version": "v1.2"
}
```

### MsgRequestGDPRData
Request data access (GDPR Article 15).

**Example**:
```json
{
  "address": "aura1...",
  "request_type": "access"
}
```

**Response**:
```json
{
  "request_id": "gdpr_req_abc"
}
```

### MsgEraseGDPRData
Request data erasure (GDPR Article 17).

**Example**:
```json
{
  "address": "aura1...",
  "erasure_reason": "User requested account deletion"
}
```

**Response**:
```json
{
  "success": true,
  "erasure_event_id": "erasure_xyz"
}
```

### MsgGenerateTaxReport
Generate tax reporting documents.

**Example**:
```json
{
  "address": "aura1...",
  "tax_year": "2025",
  "jurisdiction": "US",
  "report_type": "1099-MISC",
  "file_path": "/tmp/tax_reports/"
}
```

**Response**:
```json
{
  "report_id": "tax_2025_abc",
  "file_path": "/tmp/tax_reports/aura1..._2025_1099.pdf"
}
```

## Queries

### QueryKYCRecord
```bash
aurad query compliance kyc-record aura1...
```

### QueryKYCHistory
```bash
aurad query compliance kyc-history aura1...
```

### QueryAMLProfile
```bash
aurad query compliance aml-profile aura1...
```

### QuerySanctionsScreening
```bash
aurad query compliance sanctions-screening aura1... --force-refresh
```

### QueryTransactionAlerts
```bash
aurad query compliance transaction-alerts aura1... --unreviewed-only
```

### QueryTaxReport
```bash
aurad query compliance tax-report aura1... --tax-year 2025 --jurisdiction US
```

## Events

| Event Type | Attributes | Description |
|------------|------------|-------------|
| `kyc_submitted` | `address`, `kyc_level`, `provider` | KYC verification submitted |
| `kyc_updated` | `address`, `old_level`, `new_level`, `version` | KYC level updated |
| `suspicious_activity_reported` | `activity_id`, `address`, `activity_type` | SAR filed |
| `sanctions_screening_completed` | `address`, `status`, `matches` | Sanctions check completed |
| `transaction_alert_triggered` | `alert_id`, `address`, `rule_id`, `risk_level` | Monitoring alert |
| `gdpr_consent_recorded` | `address`, `consent_type`, `consented` | Consent updated |
| `gdpr_data_requested` | `request_id`, `address`, `request_type` | Data access request |
| `gdpr_data_erased` | `erasure_event_id`, `address` | Data erasure executed |
| `tax_report_generated` | `report_id`, `address`, `tax_year` | Tax report created |

## Errors

| Code | Name | Description |
|------|------|-------------|
| 1 | `ErrKYCNotFound` | KYC record not found |
| 2 | `ErrUnauthorizedProvider` | Provider not approved |
| 3 | `ErrBlockedJurisdiction` | Jurisdiction on OFAC sanctions list |
| 4 | `ErrRateLimitExceeded` | Too many requests in time window |
| 5 | `ErrSanctionsMatch` | Address matches sanctions list |

## Integration Notes

### For Wallet Developers

1. **KYC Status Display**: Show verification level and expiration
2. **Privacy Notice**: Display GDPR consent requirements
3. **Sanctions Check**: Verify address clear before high-value transactions
4. **Tax Export**: Provide tax document download functionality
5. **Alert Notifications**: Display unresolved transaction alerts

### Security Considerations

- **PII Handling**: Store PII off-chain, never in wallet or on-chain
- **Rate Limits**: Implement client-side throttling for screening
- **Jurisdiction Checks**: Block transactions from sanctioned countries
- **Audit Logs**: Maintain complete audit trail of data access

### Best Practices

- **KYC Expiration**: Check expiration before high-value operations
- **Screening Cache**: Use cached results within validity period
- **GDPR Compliance**: Implement data portability and erasure workflows
- **Tax Automation**: Generate reports annually for users
