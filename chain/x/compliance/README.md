# Compliance Module

The Compliance module provides comprehensive regulatory compliance features for the Aura blockchain, including KYC/AML, transaction monitoring, sanctions screening, GDPR compliance, and tax reporting.

## Features

### 1. KYC/AML Integration

- **Identity Verification**: Multi-level KYC verification (Basic, Intermediate, Advanced)
- **AML Risk Profiling**: Risk-based approach to customer due diligence
- **Enhanced Due Diligence**: Additional verification for high-risk customers
- **PEP Screening**: Politically Exposed Person identification
- **Source of Funds**: Documentation and verification
- **Ongoing Monitoring**: Continuous risk assessment

**Implementation Files:**
- `keeper/kyc_aml.go` - Core KYC/AML functionality
- Lines 1-400+ with comprehensive KYC submission, validation, and AML profiling

### 2. Transaction Monitoring

- **Real-time Monitoring**: Analyze all transactions against compliance rules
- **Velocity Limits**: Detect unusual transaction frequency
- **Structuring Detection**: Identify potential structuring behavior
- **Large Transaction Alerts**: Flag transactions exceeding thresholds
- **Pattern Recognition**: Detect suspicious patterns (round amounts, etc.)
- **Smurfing Detection**: Identify coordinated multi-account activity
- **Automated Escalation**: Auto-escalate critical alerts to SARs

**Monitoring Rules:**
- Velocity (24-hour volume limits)
- Large transactions
- Structuring (multiple txs below threshold)
- Smurfing (coordinated accounts)
- Round amounts
- Rapid fund movement
- High-risk jurisdictions

**Implementation Files:**
- `keeper/transaction_monitoring.go` - Transaction monitoring engine
- Lines 1-450+ with comprehensive rule checking and alerting

### 3. Sanctions Screening

- **OFAC SDN Screening**: Office of Foreign Assets Control Specially Designated Nationals
- **Multi-List Support**: EU Sanctions, UN Sanctions, FSE, Non-SDN
- **Automated Screening**: Real-time screening on transactions
- **Cache Management**: Efficient screening with configurable caching
- **Manual Review Workflow**: Process for reviewing potential matches
- **Block Lists**: Prevent sanctioned addresses from transacting

**Supported Lists:**
- OFAC SDN (Specially Designated Nationals)
- OFAC FSE (Foreign Sanctions Evaders)
- OFAC Non-SDN Programs
- EU Consolidated Sanctions List
- UN Security Council Sanctions
- UK HM Treasury Sanctions

**Implementation Files:**
- `keeper/sanctions.go` - Sanctions screening implementation
- Lines 1-450+ with multi-list screening and matching

### 4. GDPR Compliance

- **Consent Management**: Record and track user consents
- **Data Access Requests**: Right to access personal data
- **Right to Erasure**: Right to be forgotten (with legal exceptions)
- **Data Portability**: Export data in structured format
- **Rectification**: Correct inaccurate data
- **Data Retention**: Automated cleanup based on retention policies
- **Processing Purposes**: Track lawful basis for processing

**GDPR Rights Supported:**
- Right to access
- Right to rectification
- Right to erasure
- Right to restrict processing
- Right to data portability
- Right to object

**Implementation Files:**
- `keeper/gdpr.go` - GDPR compliance implementation
- Lines 1-500+ with complete GDPR request handling

### 5. Tax Reporting

- **US Tax Forms**: 1099-MISC, 1099-K, Form 8949
- **Capital Gains Calculation**: Long-term and short-term gains
- **Income Classification**: Categorize taxable events
- **Multi-Jurisdiction**: Support for US, EU, and other jurisdictions
- **Automated Reporting**: Generate tax forms automatically
- **CSV Export**: Export transaction data for tax preparation
- **Tax Estimates**: Calculate estimated tax liability

**Supported Forms:**
- **1099-MISC**: Miscellaneous income (staking, airdrops, etc.)
- **1099-K**: Payment card and third-party network transactions
- **Form 8949**: Sales and dispositions of capital assets
- **Schedule D**: Capital gains and losses summary

**Implementation Files:**
- `keeper/tax_reporting.go` - Tax reporting implementation
- Lines 1-550+ with comprehensive tax form generation

## Usage

### Initialize Compliance Module

```go
import (
    "github.com/aequitas/aura/chain/x/compliance"
    "github.com/aequitas/aura/chain/x/compliance/keeper"
    "github.com/aequitas/aura/chain/x/compliance/types"
)

// Create keeper with default params
params := types.DefaultParams()
keeper := keeper.NewKeeper(params)

// Create module
module := compliance.NewAppModule(keeper)
```

### Submit KYC Verification

```go
err := keeper.SubmitKYC(
    "aura1address",
    types.KYCLevelBasic,
    "KYCProvider",
    "VERIFICATION_ID",
    []string{"passport", "utility_bill"},
    "US",
)
```

### Monitor Transaction

```go
alerts, err := keeper.MonitorTransaction(
    "transaction_hash",
    "from_address",
    "to_address",
    "amount",
    timestamp,
)

// Process alerts
for _, alert := range alerts {
    if alert.RiskLevel == types.TxRiskCritical {
        // Handle critical alert
    }
}
```

### Screen for Sanctions

```go
result, err := keeper.ScreenSanctions("address", false)

if result.Status == types.SanctionsConfirmed {
    // Block transaction
    return fmt.Errorf("address is sanctioned")
}

if result.RequiresManualReview {
    // Queue for manual review
}
```

### Record GDPR Consent

```go
err := keeper.RecordGDPRConsent(
    "address",
    "data_processing",
    true,  // consented
    "1.0", // policy version
    "ip_address",
    "user_agent",
)
```

### Generate Tax Report

```go
transactions := []*types.TaxTransaction{
    {
        TransactionHash: "tx1",
        Timestamp:       time.Now(),
        TransactionType: "trade",
        Amount:          "1000",
        CostBasis:       "800",
        FairMarketValue: "1000",
        GainLoss:        "200",
        IsIncome:        false,
    },
}

reportID, err := keeper.GenerateTaxReport(
    "address",
    "2024",
    "US",
    "1099-MISC",
    transactions,
)
```

## Configuration

### Module Parameters

```go
type ComplianceParams struct {
    // KYC/AML settings
    KYCRequired      bool
    MinimumKYCLevel  KYCLevel
    KYCExpiryDays    uint64

    // Transaction monitoring
    TransactionMonitoringEnabled bool
    VelocityLimit24h             string
    SingleTransactionLimit       string
    StructuringThresholdCount    uint32

    // Sanctions screening
    SanctionsScreeningEnabled bool
    SanctionsList             []string
    ScreeningCacheHours       uint64

    // GDPR
    GDPREnabled          bool
    DataRetentionDays    uint64
    ProcessingPurposes   []string

    // Tax reporting
    TaxReportingEnabled bool
    TaxJurisdictions    []string
    TaxYearEnd          string
}
```

### Default Configuration

```go
params := types.DefaultParams()
// Returns:
// - KYC expiry: 365 days
// - Velocity limit: 1M units/24h
// - Single tx limit: 100K units
// - Structuring threshold: 5 transactions
// - Sanctions lists: OFAC_SDN, EU_SANCTIONS, UN_SANCTIONS
// - Data retention: 2555 days (~7 years)
// - Tax jurisdictions: US, EU
```

## Integration with External Services

### KYC Provider Integration

```go
type KYCProvider interface {
    VerifyIdentity(address, documentType string, documents [][]byte) (*types.KYCRecord, error)
    GetVerificationStatus(verificationID string) (*types.KYCRecord, error)
    UpdateRiskScore(address string) (string, error)
}

// Register provider
keeper.RegisterKYCProvider("ProviderName", myKYCProvider)
```

### Sanctions Provider Integration

```go
type SanctionsProvider interface {
    ScreenAddress(address string) (*types.SanctionsScreeningResult, error)
    CheckLists(lists []string) ([]*types.SanctionsMatch, error)
}

// Register provider
keeper.RegisterSanctionsProvider("ProviderName", mySanctionsProvider)
```

### Tax Report Generator Integration

```go
type TaxReportGenerator interface {
    GenerateReport(address, taxYear, reportType string, transactions []*types.TaxTransaction) (*types.TaxReport, error)
    ExportToFile(report *types.TaxReport, format string) (string, error)
}

// Register generator
keeper.RegisterTaxReportGenerator("US", myTaxGenerator)
```

## API Reference

### KYC/AML Functions

- `SubmitKYC()` - Submit KYC verification record
- `GetKYCRecord()` - Retrieve KYC record
- `ValidateKYCLevel()` - Check KYC compliance
- `UpdateAMLRiskScore()` - Update AML risk assessment
- `GetAMLProfile()` - Retrieve AML profile
- `ReportSuspiciousActivity()` - File SAR
- `SetPEPStatus()` - Mark as Politically Exposed Person
- `RequireEnhancedDueDiligence()` - Flag for EDD

### Transaction Monitoring Functions

- `MonitorTransaction()` - Analyze transaction for alerts
- `GetTransactionAlerts()` - Retrieve alerts for address
- `ReviewAlert()` - Mark alert as reviewed
- `AddMonitoringRule()` - Add custom rule
- `UpdateMonitoringRule()` - Modify existing rule
- `GetAlertStatistics()` - Get monitoring statistics

### Sanctions Screening Functions

- `ScreenSanctions()` - Screen address against sanctions lists
- `GetSanctionsResult()` - Retrieve screening result
- `ReviewSanctionsMatch()` - Manual review of match
- `BlockSanctionedAddress()` - Block confirmed sanctioned address
- `ValidateSanctionsCompliance()` - Check sanctions compliance
- `GetSanctionsStatistics()` - Get screening statistics

### GDPR Functions

- `RecordGDPRConsent()` - Record user consent
- `GetGDPRConsent()` - Retrieve consent record
- `ValidateGDPRConsent()` - Validate required consents
- `RequestGDPRData()` - Create data access request
- `ProcessRightToErasure()` - Handle deletion request
- `ProcessRectification()` - Handle correction request
- `CleanupExpiredData()` - Remove old data

### Tax Reporting Functions

- `GenerateTaxReport()` - Generate tax report
- `GetTaxReport()` - Retrieve tax report
- `MarkTaxReportFiled()` - Mark report as filed
- `Generate1099MISC()` - Generate 1099-MISC form
- `Generate1099K()` - Generate 1099-K form
- `Generate8949()` - Generate Form 8949
- `CalculateCapitalGains()` - Calculate gains/losses
- `EstimateTaxLiability()` - Estimate tax owed
- `ExportTaxReportCSV()` - Export to CSV

## Legal Documentation

The module includes comprehensive legal documentation:

### Privacy Policy
**File:** `docs/compliance/PRIVACY_POLICY.md`
- Complete GDPR-compliant privacy policy
- Data collection and usage disclosure
- User rights under GDPR and CCPA
- Data retention policies
- International data transfers
- Contact information for privacy inquiries

### Terms of Service
**File:** `docs/compliance/TERMS_OF_SERVICE.md`
- Comprehensive user agreement
- Prohibited activities
- KYC/AML requirements
- Transaction monitoring disclosure
- Tax reporting obligations
- Dispute resolution procedures
- Securities law disclosures

### Securities Law Analysis
**File:** `docs/compliance/SECURITIES_LAW_ANALYSIS.md`
- Howey Test analysis
- Token classification assessment
- Regulatory framework review
- Compliance recommendations
- Registration alternatives
- International considerations
- Risk mitigation strategies

## Testing

Comprehensive test suite included:

**File:** `keeper/keeper_test.go`
- KYC/AML testing
- Transaction monitoring tests
- Sanctions screening tests
- GDPR compliance tests
- Tax reporting tests
- Statistics and cleanup tests

Run tests:
```bash
cd chain/x/compliance/keeper
go test -v
```

## Compliance Checklist

### KYC/AML Compliance
- ✅ Customer identification program
- ✅ Risk-based customer due diligence
- ✅ Enhanced due diligence for high-risk customers
- ✅ PEP screening
- ✅ Suspicious activity monitoring
- ✅ SAR filing capability
- ✅ Record retention (5+ years)

### Sanctions Compliance
- ✅ OFAC SDN screening
- ✅ Multi-list sanctions screening
- ✅ Real-time transaction screening
- ✅ Manual review workflow
- ✅ Blocking of sanctioned addresses
- ✅ Audit trail maintenance

### Data Privacy
- ✅ GDPR compliance
- ✅ CCPA compliance
- ✅ Consent management
- ✅ Data access requests
- ✅ Right to erasure
- ✅ Data portability
- ✅ Privacy policy

### Tax Compliance
- ✅ 1099 form generation
- ✅ Capital gains calculation
- ✅ Income classification
- ✅ Multi-jurisdiction support
- ✅ Tax record retention
- ✅ IRS reporting capability

## Regulatory References

- **FinCEN**: Bank Secrecy Act / Anti-Money Laundering
- **OFAC**: Office of Foreign Assets Control
- **SEC**: Securities and Exchange Commission
- **IRS**: Internal Revenue Service
- **GDPR**: General Data Protection Regulation (EU)
- **CCPA**: California Consumer Privacy Act

## Support and Contact

For compliance-related questions:
- Compliance documentation in `docs/compliance/`
- Test examples in `keeper/keeper_test.go`
- Integration examples in this README

## License

[License Information]

## Version

Current Version: 1.0.0
Last Updated: [Date]
