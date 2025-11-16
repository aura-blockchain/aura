# Compliance Module - Quick Reference Guide

## File Locations

### Core Implementation
- **Main Keeper:** `chain/x/compliance/keeper/keeper.go`
- **KYC/AML:** `chain/x/compliance/keeper/kyc_aml.go`
- **Transaction Monitoring:** `chain/x/compliance/keeper/transaction_monitoring.go`
- **Sanctions Screening:** `chain/x/compliance/keeper/sanctions.go`
- **GDPR Compliance:** `chain/x/compliance/keeper/gdpr.go`
- **Tax Reporting:** `chain/x/compliance/keeper/tax_reporting.go`
- **Types:** `chain/x/compliance/types/types.go`
- **Tests:** `chain/x/compliance/keeper/keeper_test.go`

### Documentation
- **Privacy Policy:** `docs/compliance/PRIVACY_POLICY.md`
- **Terms of Service:** `docs/compliance/TERMS_OF_SERVICE.md`
- **Securities Law Analysis:** `docs/compliance/SECURITIES_LAW_ANALYSIS.md`
- **Implementation Summary:** `docs/compliance/IMPLEMENTATION_SUMMARY.md`
- **Module README:** `chain/x/compliance/README.md`

### Protocol Buffers
- **Proto Definitions:** `proto/aura/compliance/v1beta1/compliance.proto`

---

## Quick Start

### Initialize Module
```go
import (
    "github.com/aequitas/aura/chain/x/compliance/keeper"
    "github.com/aequitas/aura/chain/x/compliance/types"
)

params := types.DefaultParams()
keeper := keeper.NewKeeper(params)
```

### KYC Verification
```go
err := keeper.SubmitKYC(
    "aura1address",
    types.KYCLevelBasic,
    "ProviderName",
    "verification-id-123",
    []string{"passport", "utility_bill"},
    "US",
)
```

### Monitor Transaction
```go
alerts, err := keeper.MonitorTransaction(
    "tx-hash",
    "from-address",
    "to-address",
    "amount",
    time.Now(),
)
```

### Screen for Sanctions
```go
result, err := keeper.ScreenSanctions("address", false)
if result.Status == types.SanctionsConfirmed {
    // Block transaction
}
```

### Record GDPR Consent
```go
err := keeper.RecordGDPRConsent(
    "address",
    "data_processing",
    true,
    "1.0",
    "ip-address",
    "user-agent",
)
```

### Generate Tax Report
```go
reportID, err := keeper.GenerateTaxReport(
    "address",
    "2024",
    "US",
    "1099-MISC",
    transactions,
)
```

---

## Key Functions by Feature

### KYC/AML (kyc_aml.go)
- `SubmitKYC()` - Submit KYC verification
- `GetKYCRecord()` - Retrieve KYC record
- `ValidateKYCLevel()` - Check KYC compliance
- `ReportSuspiciousActivity()` - File SAR
- `SetPEPStatus()` - Mark as PEP
- `UpdateAMLRiskScore()` - Update risk assessment

### Transaction Monitoring (transaction_monitoring.go)
- `MonitorTransaction()` - Analyze transaction
- `GetTransactionAlerts()` - Get alerts
- `ReviewAlert()` - Review alert
- `AddMonitoringRule()` - Add custom rule
- `GetAlertStatistics()` - Get statistics

### Sanctions (sanctions.go)
- `ScreenSanctions()` - Screen address
- `GetSanctionsResult()` - Get result
- `ReviewSanctionsMatch()` - Review match
- `ValidateSanctionsCompliance()` - Check compliance
- `GetSanctionsStatistics()` - Get statistics

### GDPR (gdpr.go)
- `RecordGDPRConsent()` - Record consent
- `RequestGDPRData()` - Data access request
- `ProcessRightToErasure()` - Delete data
- `ProcessRectification()` - Correct data
- `CleanupExpiredData()` - Clean old data

### Tax (tax_reporting.go)
- `GenerateTaxReport()` - Generate report
- `Generate1099MISC()` - 1099-MISC form
- `Generate1099K()` - 1099-K form
- `Generate8949()` - Form 8949
- `CalculateCapitalGains()` - Calculate gains

---

## Default Parameters

```go
KYCRequired:                  false
MinimumKYCLevel:              KYCLevelBasic
KYCExpiryDays:                365
TransactionMonitoringEnabled: true
VelocityLimit24h:             "1000000"  // 1M units
SingleTransactionLimit:       "100000"   // 100K units
StructuringThresholdCount:    5
SanctionsScreeningEnabled:    true
SanctionsList:                ["OFAC_SDN", "EU_SANCTIONS", "UN_SANCTIONS"]
ScreeningCacheHours:          24
GDPREnabled:                  true
DataRetentionDays:            2555       // ~7 years
TaxReportingEnabled:          true
TaxJurisdictions:             ["US", "EU"]
```

---

## Line Number Reference

### KYC/AML (kyc_aml.go)
- `SubmitKYC`: Lines 15-89
- `GetKYCRecord`: Lines 91-105
- `ValidateKYCLevel`: Lines 107-127
- `ReportSuspiciousActivity`: Lines 197-243
- `SetPEPStatus`: Lines 272-291

### Transaction Monitoring (transaction_monitoring.go)
- `MonitorTransaction`: Lines 11-72
- `checkVelocityRule`: Lines 96-142
- `checkThresholdRule`: Lines 144-168
- `checkStructuringRule`: Lines 170-225
- `GetTransactionAlerts`: Lines 286-304

### Sanctions (sanctions.go)
- `ScreenSanctions`: Lines 10-87
- `checkOFAC_SDN`: Lines 117-145
- `ReviewSanctionsMatch`: Lines 262-288
- `ValidateSanctionsCompliance`: Lines 354-377
- `GetSanctionsStatistics`: Lines 403-431

### GDPR (gdpr.go)
- `RecordGDPRConsent`: Lines 10-59
- `RequestGDPRData`: Lines 100-140
- `ProcessRightToErasure`: Lines 221-256
- `CleanupExpiredData`: Lines 390-420
- `GetGDPRComplianceStatus`: Lines 422-455

### Tax (tax_reporting.go)
- `GenerateTaxReport`: Lines 11-94
- `Generate1099MISC`: Lines 112-126
- `Generate1099K`: Lines 128-149
- `CalculateCapitalGains`: Lines 166-195
- `EstimateTaxLiability`: Lines 268-319

---

## Running Tests

```bash
cd chain/x/compliance/keeper
go test -v
```

Expected output:
```
PASS: TestNewKeeper
PASS: TestSubmitKYC
PASS: TestValidateKYCLevel
PASS: TestReportSuspiciousActivity
PASS: TestMonitorTransaction
PASS: TestScreenSanctions
PASS: TestRecordGDPRConsent
PASS: TestRequestGDPRData
PASS: TestGenerateTaxReport
PASS: TestCalculateCapitalGains
PASS: TestGetAlertStatistics
PASS: TestGetHighRiskAddresses
PASS: TestCleanupExpiredData

Total: 13 tests passing in ~2 seconds
```

---

## Error Codes

Common error scenarios:

### KYC/AML
- `"address cannot be empty"` - Missing address
- `"KYC verification required"` - No KYC record
- `"KYC verification expired"` - Expired record
- `"insufficient KYC level"` - Below minimum level

### Transaction Monitoring
- `"Transaction monitoring failed"` - Monitoring error
- Alert not generated if rules disabled or not triggered

### Sanctions
- `"sanctions screening failed"` - Screening error
- `"address is on sanctions list"` - Confirmed match
- `"potential sanctions match pending review"` - Needs review

### GDPR
- `"GDPR features not enabled"` - Module disabled
- `"required consent not found"` - Missing consent
- `"cannot erase data: retention required"` - Legal hold

### Tax
- `"tax reporting not enabled"` - Module disabled
- `"invalid jurisdiction"` - Unknown jurisdiction
- `"1099-K not required"` - Below thresholds

---

## Statistics Functions

All modules include statistics functions:

```go
// Transaction Monitoring
stats := keeper.GetAlertStatistics()
// Returns: total_alerts, reviewed_alerts, pending_alerts, alerts_by_risk

// Sanctions Screening
stats := keeper.GetSanctionsStatistics()
// Returns: total_screened, clear, matches, confirmed, pending_review

// Tax Reporting
stats := keeper.GetTaxReportingStatistics()
// Returns: total_reports, filed_reports, unfiled_reports, reports_by_year
```

---

## Integration Checklist

### Pre-Production
- [ ] Legal review of all documentation
- [ ] Select KYC provider
- [ ] Select sanctions screening provider
- [ ] Configure transaction limits
- [ ] Set data retention periods
- [ ] Update contact information in legal docs

### Production
- [ ] Enable monitoring rules
- [ ] Configure alerts and notifications
- [ ] Set up audit logging
- [ ] Enable GDPR features for EU users
- [ ] Configure tax reporting for jurisdictions

### Post-Production
- [ ] Daily sanctions list updates
- [ ] Weekly compliance reviews
- [ ] Quarterly audits
- [ ] Annual legal document reviews

---

## Support Resources

- **Full Documentation:** `chain/x/compliance/README.md`
- **Implementation Summary:** `docs/compliance/IMPLEMENTATION_SUMMARY.md`
- **Legal Docs:** `docs/compliance/`
- **Test Examples:** `chain/x/compliance/keeper/keeper_test.go`

---

**Version:** 1.0
**Last Updated:** November 13, 2025
