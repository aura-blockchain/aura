# Transaction Monitoring Integration

## Overview

The compliance module provides real-time transaction monitoring through the `MonitoredBankKeeper`, which wraps the standard Cosmos SDK bank keeper to evaluate AML (Anti-Money Laundering) rules before executing coin transfers.

## Architecture

```
Transaction Request
       |
       v
MonitoredBankKeeper
       |
       +-- MonitorTransaction() -----> Evaluate Rules
       |                               |
       |                               +-- Large Transaction Check
       |                               +-- Velocity Check
       |                               +-- Structuring Detection
       |                               +-- Sanctions Screening
       |                               |
       |                               v
       |                          Generate Alerts
       |
       +-- ShouldBlockTransaction() -> Block if Critical Risk
       |
       v
Standard BankKeeper (if allowed)
       |
       v
Update AML Profiles
```

## Features

### 1. Real-Time Monitoring

All coin transfers are monitored before execution:
- `SendCoins` - peer-to-peer transfers
- `InputOutputCoins` - multi-send transactions
- `SendCoinsFromModuleToAccount` - module withdrawals
- `SendCoinsFromAccountToModule` - module deposits

### 2. Monitoring Rules

#### Large Transaction Rule
Triggers when a single transaction exceeds a threshold:
```go
TransactionMonitoringRule{
    Id:        "large_transaction",
    RuleType:  "large_transaction",
    RiskLevel: TX_RISK_HIGH,
    Parameters: {
        "threshold": "10000000", // 10M tokens
    },
}
```

#### Velocity Rule
Monitors 24-hour transaction volume:
```go
TransactionMonitoringRule{
    Id:        "velocity",
    RuleType:  "velocity",
    RiskLevel: TX_RISK_MEDIUM,
    Parameters: {
        "threshold_24h": "50000000", // 50M tokens/24h
    },
}
```

#### Structuring Rule
Detects patterns of multiple transactions to avoid thresholds:
```go
TransactionMonitoringRule{
    Id:        "structuring",
    RuleType:  "structuring",
    RiskLevel: TX_RISK_CRITICAL,
    Parameters: {
        "count_threshold": "10", // 10+ txs in period
    },
}
```

### 3. Sanctions Screening

Checks sender and recipient addresses against sanctions lists:
- OFAC SDN List
- EU Sanctions
- UN Sanctions
- Custom lists

Sanctioned addresses trigger **CRITICAL** risk alerts that block transactions.

### 4. Risk Levels

| Level | Description | Action |
|-------|-------------|--------|
| LOW | Normal activity | Allow, log |
| MEDIUM | Noteworthy pattern | Allow, alert |
| HIGH | Suspicious activity | Allow, alert for review |
| CRITICAL | Sanctioned/prohibited | **BLOCK TRANSACTION** |

### 5. AML Profiling

Each address has an AML profile tracking:
- Total transaction count
- Total volume transacted
- Risk level (auto-calculated)
- Last assessment timestamp
- PEP (Politically Exposed Person) status
- Source of funds

## Integration

### In `app/app.go`

```go
// Create base bank keeper
baseBankKeeper := bankkeeper.NewBaseKeeper(...)

// Create compliance keeper
complianceKeeper := compliancekeeper.NewKeeper(...)

// Wrap bank keeper with monitoring
monitoredBankKeeper := compliancekeeper.NewMonitoredBankKeeper(
    baseBankKeeper,
    complianceKeeper,
)

// Use monitoredBankKeeper in all other modules
stakingKeeper := stakingkeeper.NewKeeper(
    ...,
    monitoredBankKeeper, // Instead of baseBankKeeper
    ...,
)
```

### Configuration

Enable monitoring via module parameters:
```go
ComplianceParams{
    TransactionMonitoringEnabled: true,
    SingleTransactionLimit:       "10000000",
    VelocityLimit_24H:            "50000000",
    StructuringThresholdCount:    10,
    SanctionsScreeningEnabled:    true,
    SanctionsList: []string{
        "OFAC_SDN",
        "EU_SANCTIONS",
    },
}
```

## Usage Examples

### Example 1: Normal Transaction (Allowed)

```go
from := sdk.AccAddress("aura1sender...")
to := sdk.AccAddress("aura1recipient...")
amount := sdk.NewCoins(sdk.NewInt64Coin("uaura", 1000))

// Monitoring: Evaluates rules, no alerts
// Result: Transaction proceeds
err := bankKeeper.SendCoins(ctx, from, to, amount)
```

### Example 2: Large Transaction (Alerted, Allowed)

```go
amount := sdk.NewCoins(sdk.NewInt64Coin("uaura", 15000000)) // 15M

// Monitoring: Generates HIGH risk alert
// Result: Transaction proceeds, alert stored for review
err := bankKeeper.SendCoins(ctx, from, to, amount)

// Alert can be queried later
alerts, _ := complianceKeeper.GetTransactionAlerts(ctx, from.String())
```

### Example 3: Sanctioned Address (Blocked)

```go
sanctionedAddr := sdk.AccAddress("aura1sanctioned...")

// Monitoring: Detects sanctioned address
// Result: Transaction BLOCKED with error
err := bankKeeper.SendCoins(ctx, sanctionedAddr, to, amount)
// err: "transaction blocked by compliance: Critical risk detected: Transaction from sanctioned address"
```

### Example 4: Multiple High Risk Factors (Blocked)

```go
// Address with:
// - PEP status
// - Multiple suspicious activity reports

// Monitoring: Multiple HIGH risk alerts
// Result: Transaction BLOCKED (defense in depth)
err := bankKeeper.SendCoins(ctx, from, to, amount)
// err: "transaction blocked by compliance: Multiple high risk factors detected (2)"
```

## Alert Management

### Querying Alerts

```bash
# Get alerts for an address
aurad query compliance transaction-alerts [address]

# Get unreviewed alerts only
aurad query compliance transaction-alerts [address] --unreviewed-only
```

### Reviewing Alerts

Alerts require manual review and resolution:

```go
alert := complianceKeeper.GetTransactionAlert(ctx, alertID)
alert.Reviewed = true
alert.ReviewedAt = &currentTime
alert.Reviewer = "compliance_officer_address"
alert.Resolution = "dismissed" // or "escalate", "file_sar"
complianceKeeper.SetTransactionAlert(ctx, address, alert)
```

## Events

The monitoring system emits events for external systems:

### EventTypeTransactionAlert
```json
{
  "type": "transaction_alert",
  "attributes": [
    {"key": "alert_id", "value": "large_tx_aura1..."},
    {"key": "address", "value": "aura1..."},
    {"key": "risk_level", "value": "HIGH"},
    {"key": "rule_id", "value": "large_transaction"},
    {"key": "description", "value": "Large transaction detected: 15000000 > 10000000"}
  ]
}
```

### EventTypeComplianceViolation
```json
{
  "type": "compliance_violation",
  "attributes": [
    {"key": "from", "value": "aura1..."},
    {"key": "to", "value": "aura1..."},
    {"key": "amount", "value": "10000000uaura"},
    {"key": "reason", "value": "Critical risk detected: sanctioned address"},
    {"key": "blocked", "value": "true"}
  ]
}
```

## Security Considerations

### Defense in Depth

1. **Pre-flight Checks**: Monitoring occurs BEFORE transaction execution
2. **Multiple Rules**: All enabled rules are evaluated
3. **Critical Blocking**: CRITICAL alerts always block transactions
4. **Multiple High Risk**: 2+ HIGH risk factors block transactions
5. **Audit Trail**: All alerts persisted to state, events emitted

### Privacy (GDPR Compliance)

- No PII stored on-chain
- Only cryptographic commitments and risk scores
- Off-chain systems handle sensitive data
- Right to erasure via MsgEraseGDPRData

### Performance

- Monitoring adds minimal overhead (~1-2ms per transaction)
- Rules evaluated in parallel where possible
- Caching for sanctions screening results
- No network calls during transaction execution

## Testing

### Unit Tests

```bash
# Run transaction monitoring tests
go test ./x/compliance/keeper/... -run TestMonitorTransaction

# Run monitored bank keeper tests
go test ./x/compliance/keeper/... -run TestMonitoredBankKeeper
```

### Integration Tests

```bash
# Run full compliance integration tests
go test ./x/compliance/keeper/... -v
```

## Monitoring Dashboard (External)

For production deployments, integrate with external monitoring systems:

1. Subscribe to `EventTypeTransactionAlert` events
2. Track alert rates by risk level
3. Monitor blocked transaction rates
4. Alert on critical violations
5. Generate SAR (Suspicious Activity Reports)

## References

- [FATF Recommendations](https://www.fatf-gafi.org/recommendations.html) - AML/CFT standards
- [FinCEN](https://www.fincen.gov/) - Financial crimes enforcement
- [OFAC SDN List](https://home.treasury.gov/policy-issues/financial-sanctions/specially-designated-nationals-and-blocked-persons-list-sdn-human-readable-lists) - Sanctions screening
- [GDPR Article 32](https://gdpr-info.eu/art-32-gdpr/) - Security of processing

## Future Enhancements

- [ ] Machine learning risk scoring
- [ ] Graph analysis for transaction networks
- [ ] Integration with Chainalysis/Elliptic
- [ ] Automated SAR filing
- [ ] Risk-based transaction limits
- [ ] Time-based velocity windows (1h, 24h, 7d)
- [ ] Geographic risk scoring
- [ ] Counterparty risk assessment
