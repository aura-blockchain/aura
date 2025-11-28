# Aura Blockchain Monitoring & Alerting System

A comprehensive, production-ready monitoring and alerting infrastructure for the Aura blockchain.

## Features

### 1. Real-time Transaction Monitoring
- Monitors all blockchain transactions in real-time
- Tracks transaction status, gas usage, and amounts
- Identifies large transactions automatically
- Provides transaction statistics and analytics

**Implementation:** `keeper/transaction_monitor.go`

### 2. Alert System for Security Threats
- Multi-severity alert system (Critical, High, Medium, Low, Info)
- Alert acknowledgment and resolution workflow
- Configurable cooldown periods for critical alerts
- Alert retention and cleanup policies

**Implementation:** `alerting/alert_manager.go`, `keeper/alerts.go`

### 3. Anomaly Detection using Machine Learning
- Statistical anomaly detection using Z-score analysis
- Transaction and network pattern anomaly detection
- Self-training model with periodic updates
- Configurable anomaly thresholds

**Implementation:** `ml/anomaly_detector.go`, `keeper/anomaly_detection.go`

### 4. Prometheus Metrics Integration
- Comprehensive metrics for all monitoring aspects
- Transaction metrics (count, gas usage, duration)
- Alert metrics (active, total, resolution time)
- Validator metrics (uptime, missed blocks)
- Network health metrics (TPS, block time, congestion)
- Economic metrics (gas price, TVL)
- Security metrics (events, threats, mitigations)

**Implementation:** `metrics/prometheus.go`

### 5. Grafana Dashboard Setup
Pre-configured dashboards for:
- Network Health Monitoring
- Security & SIEM
- Validator Performance
- Economics (TVL, Gas Prices)

**Configuration:** `../../grafana/dashboards/*.json`

### 6. Centralized Log Aggregation
- Multi-module log collection
- Log level filtering (Debug, Info, Warning, Error, Fatal)
- Trace ID support for distributed tracing
- Log search and export capabilities
- Configurable retention policies

**Implementation:** `keeper/log_aggregator.go`

### 7. Security Information and Event Management (SIEM)
- Security event recording and analysis
- Threat level assessment (1-10 scale)
- Event mitigation tracking
- Trend analysis and statistics
- Event type categorization

**Implementation:** `siem/siem_manager.go`

### 8. Blockchain Explorer Integration
- Sync status monitoring
- Sync lag detection and alerting
- Health status tracking
- API error monitoring

**Implementation:** `keeper/explorer_integration.go`

### 9. Validator Uptime Monitoring
- Real-time uptime percentage tracking
- Consecutive miss detection
- Automatic jailing detection
- Validator status monitoring (active, jailed, unbonding)

**Implementation:** `keeper/validator_monitor.go`

### 10. Network Health Dashboard
- Block time monitoring
- Transactions per second (TPS)
- Network congestion analysis
- Consensus health scoring
- Mempool size tracking
- Peer count monitoring

**Implementation:** `keeper/network_health.go`

### 11. Gas Price Tracking and Alerts
- Real-time gas price tracking
- Price history with configurable size
- Volatility scoring
- Trend direction analysis (increasing, decreasing, stable)
- Spike detection and alerting

**Implementation:** `keeper/gas_price_tracker.go`

### 12. Total Value Locked (TVL) Monitoring
- Total TVL across all modules
- Per-module TVL tracking
- 24h and 7d change calculation
- Largest pools ranking
- Significant change alerting

**Implementation:** `keeper/tvl_monitor.go`

### 13. Large Transaction Alert System
- Configurable threshold for large transactions
- Automatic flagging and alerting
- Transaction details tracking
- Integration with anomaly detection

**Implementation:** `keeper/transaction_monitor.go` (lines 29-38)

### 14. Failed Transaction Pattern Analysis
- Pattern recognition for failed transactions
- Occurrence counting
- Affected address tracking
- Configurable pattern threshold
- Automatic severity assignment

**Implementation:** `keeper/failed_tx_analyzer.go`

### 15. 24/7 Security Operations Center (SOC) Framework
- Active operator tracking
- Incident categorization
- Response time monitoring
- Configurable rotation intervals
- Optional approval requirements

**Metrics:** SOC-specific Prometheus metrics in `metrics/prometheus.go` (lines 210-238)

## Installation

### Prerequisites
- Go 1.25+
- Prometheus
- Grafana
- CometBFT

### Setup

1. **Install the monitoring module:**
```bash
cd chain/x/monitoring
go mod download
```

2. **Configure Prometheus:**
```bash
cp ../../prometheus/prometheus.yml /etc/prometheus/
cp ../../prometheus/rules/monitoring-alerts.yml /etc/prometheus/rules/
systemctl restart prometheus
```

3. **Import Grafana Dashboards:**
```bash
# Import dashboards via Grafana UI or API
curl -X POST http://localhost:3000/api/dashboards/import \
  -H "Content-Type: application/json" \
  -d @../../grafana/dashboards/network-health.json
```

## Configuration

### Parameters

Configure monitoring parameters in your application:

```go
params := types.Params{
    // Transaction Monitoring
    EnableTransactionMonitoring: true,
    LargeTransactionThreshold:   1000000,

    // Alerts
    EnableAlerts:              true,
    AlertRetentionPeriod:      30 * 24 * time.Hour,
    CriticalAlertCooldown:     5 * time.Minute,

    // Anomaly Detection
    EnableAnomalyDetection:    true,
    AnomalyThreshold:          0.75,
    MLModelUpdateInterval:     24 * time.Hour,

    // Prometheus
    EnablePrometheusMetrics:   true,
    PrometheusPort:            9090,

    // Additional parameters...
}
```

### Default Values

See `types/params.go` for all default parameter values.

## Usage

### Initialize Keeper

```go
import (
    "github.com/aequitas/aura/chain/x/monitoring/keeper"
    "github.com/aequitas/aura/chain/x/monitoring/types"
)

params := types.DefaultParams()
k := keeper.NewKeeper(&params)
defer k.Close()
```

### Monitor Transactions

```go
tx := &types.TransactionMonitorData{
    TxHash:      "abc123",
    Sender:      "aura1sender",
    Receiver:    "aura1receiver",
    Amount:      1000000,
    GasUsed:     50000,
    GasPrice:    100,
    Status:      "success",
    Timestamp:   time.Now(),
    BlockHeight: 1000,
    Module:      "bank",
}

err := k.MonitorTransaction(tx)
```

### Create Alerts

```go
alert, err := k.CreateAlert(
    types.AlertTypeSecurityThreat,
    types.SeverityHigh,
    "Suspicious transaction detected",
    map[string]interface{}{
        "tx_hash": "abc123",
        "reason":  "unusual_pattern",
    },
)
```

### Track Validator Uptime

```go
err := k.UpdateValidatorUptime(
    "auravaloper1test",
    "TestValidator",
    blockHeight,
    signed, // true if validator signed the block
)
```

### Monitor Network Health

```go
health := &types.NetworkHealth{
    BlockHeight:       1000,
    BlockTime:         6.5,
    TPS:               100.5,
    NetworkCongestion: 0.3,
    ConsensusHealth:   0.95,
}

err := k.UpdateNetworkHealth(health)
```

### Log Aggregation

```go
err := k.LogEntry(
    types.LogLevelError,
    "my-module",
    "Error occurred",
    map[string]interface{}{"error": "details"},
    "trace-id",
    "span-id",
)
```

## Metrics

### Available Metrics

All metrics are prefixed with `aura_monitoring_`:

- **Transactions:** `total_transactions`, `transaction_gas_used`, `large_transactions_total`
- **Alerts:** `alerts_total`, `active_alerts`, `alert_resolution_time_seconds`
- **Anomalies:** `anomaly_detections_total`, `anomaly_score`
- **Validators:** `validator_uptime_percentage`, `validator_missed_blocks_total`, `active_validators`
- **Network:** `block_time_seconds`, `transactions_per_second`, `network_congestion`
- **Economics:** `total_tvl`, `current_gas_price`, `gas_price_spikes_total`
- **Security:** `security_events_total`, `threat_level`, `mitigated_threats_total`

### Querying Metrics

```promql
# Average block time over 5 minutes
rate(aura_monitoring_block_time_seconds[5m])

# Active critical alerts
aura_monitoring_active_alerts{severity="critical"}

# Total TVL change
delta(aura_monitoring_total_tvl[24h])
```

## Alerting Rules

Prometheus alerting rules are defined in `prometheus/rules/monitoring-alerts.yml`:

- **Network Health:** High congestion, low consensus health, high block time
- **Validators:** Validator down, validator jailed
- **Security:** High anomaly rate, failed transaction patterns, security threats
- **Economics:** Gas price spikes, significant TVL changes

## Testing

Run the comprehensive test suite:

```bash
cd keeper
go test -v -cover
```

Tests include:
- Transaction monitoring
- Alert management
- Validator uptime tracking
- Network health monitoring
- Gas price tracking
- TVL monitoring
- Log aggregation
- Failed transaction patterns
- Data cleanup

## Architecture

```
monitoring/
├── types/           # Type definitions, errors, parameters
├── keeper/          # Core keeper implementation
│   ├── keeper.go                  # Main keeper
│   ├── transaction_monitor.go     # Transaction monitoring
│   ├── alerts.go                  # Alert management
│   ├── anomaly_detection.go       # Anomaly detection integration
│   ├── validator_monitor.go       # Validator monitoring
│   ├── network_health.go          # Network health
│   ├── gas_price_tracker.go       # Gas price tracking
│   ├── tvl_monitor.go             # TVL monitoring
│   ├── failed_tx_analyzer.go      # Failed TX analysis
│   ├── log_aggregator.go          # Log aggregation
│   └── explorer_integration.go    # Explorer integration
├── metrics/         # Prometheus metrics
├── ml/              # Machine learning models
├── alerting/        # Alert management
├── siem/            # SIEM implementation
├── params/          # Parameter storage
└── module.go        # Module registration
```

## Production Deployment

### Best Practices

1. **Resource Allocation:**
   - Allocate sufficient memory for log aggregation
   - Configure appropriate retention periods
   - Monitor Prometheus storage usage

2. **Security:**
   - Secure Prometheus endpoints
   - Implement authentication for Grafana
   - Encrypt sensitive alert data
   - Regularly review security events

3. **High Availability:**
   - Deploy Prometheus with remote storage
   - Set up Grafana clustering
   - Configure alert redundancy
   - Implement backup strategies

4. **Performance:**
   - Adjust scrape intervals based on load
   - Use metric relabeling for efficiency
   - Optimize query performance
   - Monitor monitoring system overhead

### Monitoring the Monitoring System

Monitor the health of the monitoring system itself:

```promql
# Prometheus health
up{job="aura-monitoring"}

# Scrape duration
scrape_duration_seconds{job="aura-monitoring"}

# Time series count
prometheus_tsdb_symbol_table_size_bytes
```

## SOC Framework

The 24/7 Security Operations Center framework includes:

1. **Operator Management:**
   - Track active operators
   - Rotation scheduling
   - Shift handoffs

2. **Incident Response:**
   - Incident categorization by severity
   - Response time tracking
   - Resolution workflows

3. **Metrics:**
   - `soc_active_operators`: Number of active SOC operators
   - `soc_incidents_total`: Total incidents by severity
   - `soc_response_time_seconds`: Incident response times

4. **Configuration:**
   ```go
   params.EnableSOC = true
   params.SOCRotationInterval = 8 * time.Hour
   params.RequireSOCApproval = false
   ```

## Support

For issues and feature requests, please open an issue in the repository.

## License

Copyright (c) 2025 Aura Network
