# WASM & Bridge Security Monitoring

**ROADMAP Task #16**: Monitoring probes and alerts for WASM tx failures, state load errors, and signature mismatches.

## Overview

This document describes the comprehensive monitoring system for WASM contract execution, bridge signature verification, and state integrity in the Aura blockchain. The monitoring system provides real-time visibility into critical security metrics and alerts operators to potential issues before they become critical.

## Architecture

### Components

```
┌─────────────────────────────────────────────────────────────┐
│ Aura Blockchain                                             │
│                                                             │
│  ┌────────────┐         ┌────────────┐                     │
│  │ WASM Module│────────▶│ Prometheus │                     │
│  │ (telemetry)│         │  Metrics   │                     │
│  └────────────┘         └────────────┘                     │
│                                │                            │
│  ┌────────────┐                │                            │
│  │Bridge Module│───────────────┘                            │
│  │ (telemetry)│                                             │
│  └────────────┘                                             │
└─────────────────────────────────────────────────────────────┘
                                 │
                                 ▼
                   ┌──────────────────────────┐
                   │ Prometheus Server        │
                   │ - Scrapes metrics        │
                   │ - Evaluates alert rules  │
                   │ - Stores time series     │
                   └──────────────────────────┘
                                 │
                    ┌────────────┴────────────┐
                    │                         │
                    ▼                         ▼
          ┌─────────────────┐      ┌──────────────────┐
          │ Alertmanager    │      │ Grafana          │
          │ - Routes alerts │      │ - Visualizations │
          │ - Deduplication │      │ - Dashboards     │
          │ - Notifications │      │ - Historical     │
          └─────────────────┘      └──────────────────┘
```

### Metric Categories

1. **WASM Transaction Failures**: Track failures by contract, error type, and operation
2. **State Load Errors**: Monitor database integrity and unmarshal failures
3. **Signature Verification**: Track mismatches, recovery failures, and performance

## Prometheus Metrics

### WASM Module Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `aura_wasm_tx_total` | Counter | contract_address, operation | Total WASM transactions |
| `aura_wasm_tx_failures_total` | Counter | contract_address, error_type, operation | WASM transaction failures |
| `aura_wasm_instantiation_failures_total` | Counter | code_id, error_type | Contract instantiation failures |
| `aura_wasm_circuit_breaker_state` | Gauge | state | Circuit breaker state (0=closed, 1=open, 2=half-open) |
| `aura_wasm_validation_cache_total` | Counter | - | Total validation cache lookups |
| `aura_wasm_validation_cache_hits_total` | Counter | - | Validation cache hits |
| `aura_wasm_execution_duration_seconds` | Histogram | contract_address, success | Execution duration |
| `aura_wasm_gas_used_total` | Counter | contract_address | Total gas consumed |

### Bridge Module Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `aura_bridge_signature_verifications_total` | Counter | chain, signature_type | Total signature verifications |
| `aura_bridge_signature_mismatches_total` | Counter | chain, signature_type, error_reason | Signature verification failures |
| `aura_bridge_invalid_recovery_id_total` | Counter | chain | Invalid recovery IDs |
| `aura_bridge_pubkey_recovery_failures_total` | Counter | chain, recovery_id | Public key recovery failures |
| `aura_bridge_signature_verification_duration_seconds` | Histogram | chain, success | Verification duration |

### State Integrity Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `aura_state_load_errors_total` | Counter | module, store, error_type | State load errors |
| `aura_unmarshal_errors_total` | Counter | module, proto_type | Protobuf unmarshal errors |
| `aura_state_corruption_total` | Counter | module, store | Detected state corruption |
| `aura_kvstore_iteration_errors_total` | Counter | store | KVStore iteration errors |

## Alert Rules

### Critical Alerts (Immediate Action Required)

#### WasmTxFailureRateHigh
- **Trigger**: WASM tx failure rate > 5% for 5 minutes
- **Severity**: Critical
- **Action**: Investigate contract issues, check for attacks, review error logs
- **Runbook**: Check contract code, verify gas limits, inspect state

#### StateLoadErrorsDetected
- **Trigger**: Any state load errors detected for 2 minutes
- **Severity**: Critical
- **Action**: Database integrity check, verify protobuf schemas
- **Runbook**: Urgent database inspection, check unmarshal errors

#### SignatureMismatchRateHigh
- **Trigger**: Signature mismatch rate > 1% for 5 minutes
- **Severity**: High
- **Action**: Check for attacks, verify wallet compatibility
- **Runbook**: Investigate signature format, check for replay attacks

#### WasmCircuitBreakerOpen
- **Trigger**: Circuit breaker state = open for 1 minute
- **Severity**: Critical
- **Action**: Contract registry integration degraded, system in permissive mode
- **Runbook**: Check registry health, inspect circuit breaker logs

### High Priority Alerts

#### WasmContractFailureSpike
- **Trigger**: Specific contract > 10% failure rate for 3 minutes
- **Severity**: High
- **Action**: Contract may have critical bug or be under attack
- **Runbook**: Investigate contract, consider pausing if malicious

#### SignatureFailuresByChain
- **Trigger**: Chain-specific signature mismatch rate > 5% for 3 minutes
- **Severity**: Critical
- **Action**: Chain-specific signature issue or attack
- **Runbook**: Check chain signature implementation, verify recovery ID handling

#### MultipleSubsystemFailures
- **Trigger**: 2+ subsystems failing simultaneously for 5 minutes
- **Severity**: Critical
- **Action**: Potential coordinated attack or systemic issue
- **Runbook**: Emergency system health assessment, activate incident response

### Warning Alerts

#### WasmValidationCacheLow
- **Trigger**: Cache hit rate < 50% for 10 minutes
- **Severity**: Warning
- **Action**: Performance degradation, increased latency
- **Runbook**: Monitor cache size, check eviction issues, review TTL

#### SignatureVerificationSlow
- **Trigger**: p95 verification time > 100ms for 10 minutes
- **Severity**: Warning
- **Action**: Performance impact on transaction processing
- **Runbook**: Check CPU contention, optimize verification code

## Grafana Dashboard

### Dashboard: WASM & Bridge Security Monitoring

Location: `/home/decri/blockchain-projects/aura/grafana/dashboards/wasm-bridge-security.json`

#### Panels

1. **WASM Transaction Failure Rate** (Graph)
   - Overall failure rate
   - Per-contract failure rates
   - Alert threshold at 5%

2. **WASM Failures by Error Type** (Stacked Graph)
   - out_of_gas
   - unauthorized
   - contract_not_found
   - registry_error
   - rate_limited
   - circuit_breaker_open
   - other

3. **State Load Errors by Module** (Graph)
   - Per-module error rates
   - Store-specific errors
   - Error type classification

4. **Unmarshal Errors** (Graph)
   - Module-specific unmarshal failures
   - Proto type classification

5. **Signature Verification Mismatch Rate** (Graph)
   - Overall mismatch rate
   - Chain-specific rates (PAW, XAI, Aura)
   - Alert threshold at 1%

6. **Signature Verification by Chain** (Graph)
   - Success vs mismatch rates
   - Per-chain breakdown

7. **WASM Circuit Breaker State** (Stat)
   - Current state (Closed/Open/Half-Open)
   - Color-coded background

8. **WASM Validation Cache Hit Rate** (Stat)
   - Cache efficiency percentage
   - Color thresholds: <50% red, <70% yellow, >=70% green

9. **Invalid Recovery IDs** (Stat)
   - Per-chain invalid recovery ID rate

10. **Public Key Recovery Failures** (Stat)
    - Per-chain recovery failure rate

11. **WASM Instantiation Failures** (Graph)
    - Per-code-ID failure rates
    - Error type classification

12. **Signature Verification Duration** (Graph)
    - p95 latency per chain
    - Performance threshold at 100ms

13. **State Corruption Events** (Table)
    - Module, store, corruption count (1h)
    - Color-coded by severity

14. **KVStore Iteration Errors** (Graph)
    - Per-store error rates

15. **Multi-Subsystem Health** (Graph)
    - Combined view of WASM, state, signature, unmarshal errors
    - Detects correlated failures

### Variables

- `$datasource`: Prometheus data source
- `$contract`: Contract address filter (multi-select, all)
- `$chain`: Chain filter (multi-select, all)

### Annotations

- Security alerts from Prometheus
- Displayed as red markers on graphs
- Tooltip shows alert details

## Configuration

### Prometheus Configuration

File: `/home/decri/blockchain-projects/aura/prometheus/prometheus.yml`

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "rules/*.yml"

scrape_configs:
  - job_name: 'aura-monitoring'
    static_configs:
      - targets: ['localhost:9090']
    metrics_path: '/metrics'
    scrape_interval: 10s
```

### Alert Rules Configuration

File: `/home/decri/blockchain-projects/aura/prometheus/rules/wasm-bridge-alerts.yml`

Contains all alert rules for WASM, bridge, and state monitoring.

## Usage

### Starting the Monitoring Stack

```bash
# Start Prometheus
cd /home/decri/blockchain-projects/aura/prometheus
prometheus --config.file=prometheus.yml

# Start Grafana
cd /home/decri/blockchain-projects/aura/grafana
grafana-server --config=grafana.ini

# Import dashboard
# Navigate to Grafana UI → Dashboards → Import
# Upload: grafana/dashboards/wasm-bridge-security.json
```

### Accessing Dashboards

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000
- **Dashboard**: "WASM & Bridge Security Monitoring"

### Querying Metrics

#### Check WASM failure rate
```promql
rate(aura_wasm_tx_failures_total[5m]) /
clamp_min(rate(aura_wasm_tx_total[5m]), 0.0001)
```

#### Check signature mismatch rate for PAW chain
```promql
rate(aura_bridge_signature_mismatches_total{chain="paw"}[5m]) /
clamp_min(rate(aura_bridge_signature_verifications_total{chain="paw"}[5m]), 0.0001)
```

#### Check state load errors
```promql
rate(aura_state_load_errors_total[5m])
```

#### Circuit breaker status
```promql
aura_wasm_circuit_breaker_state
```

## Testing

### Verifying Metrics Collection

```bash
# Check if metrics are being exposed
curl http://localhost:9090/metrics | grep aura_wasm

# Check specific metric
curl http://localhost:9090/metrics | grep aura_bridge_signature_verifications_total

# Query Prometheus API
curl 'http://localhost:9090/api/v1/query?query=aura_wasm_tx_total'
```

### Triggering Test Alerts

```bash
# Trigger WASM failure (example - requires actual contract)
# Execute failing WASM transaction

# Trigger signature mismatch (example)
# Submit invalid signature to bridge link_address

# Trigger state load error (example)
# Corrupt database entry (testing environment only!)
```

### Verifying Alerting

```bash
# Check active alerts in Prometheus
curl http://localhost:9090/api/v1/alerts

# Check alert rules
curl http://localhost:9090/api/v1/rules
```

## Security Considerations

### Metric Exposure

- Metrics endpoint should be restricted to internal network
- Use authentication for Prometheus and Grafana in production
- Avoid exposing sensitive data in metric labels

### Alert Fatigue

- Tune alert thresholds based on baseline traffic
- Use appropriate `for` durations to avoid flapping
- Group related alerts to reduce noise

### Performance Impact

- Metrics collection has minimal overhead (<1% CPU)
- Use sampling for high-frequency operations
- Batch metric updates where possible

## Troubleshooting

### Metrics Not Appearing

1. Verify Prometheus is scraping the target
2. Check that telemetry code is being executed
3. Inspect Prometheus logs for scrape errors
4. Verify metric names and labels

### Alerts Not Firing

1. Check alert rule syntax in Prometheus
2. Verify evaluation interval
3. Check `for` duration
4. Inspect Prometheus rule evaluation logs

### Dashboard Not Loading

1. Verify Prometheus data source in Grafana
2. Check dashboard JSON syntax
3. Verify metric queries
4. Check Grafana logs

### High Cardinality Issues

If metrics have too many unique label combinations:

1. Reduce label cardinality (e.g., aggregate contract addresses)
2. Increase metric retention settings
3. Use recording rules to pre-aggregate
4. Consider metric sampling

## Maintenance

### Regular Tasks

- **Daily**: Review active alerts, check dashboard for anomalies
- **Weekly**: Analyze metric trends, tune alert thresholds
- **Monthly**: Review metric retention, optimize queries
- **Quarterly**: Update alert runbooks, refine dashboards

### Metric Retention

Default Prometheus retention: 15 days

To increase retention:
```bash
prometheus --config.file=prometheus.yml --storage.tsdb.retention.time=30d
```

### Backup

Backup Prometheus data:
```bash
tar -czf prometheus-backup-$(date +%Y%m%d).tar.gz /prometheus/data
```

Backup Grafana dashboards:
```bash
# Export dashboards via API
curl -X GET http://localhost:3000/api/dashboards/db/wasm-bridge-security
```

## References

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Cosmos SDK Telemetry](https://docs.cosmos.network/main/core/telemetry)
- [OpenTelemetry](https://opentelemetry.io/)

## Changelog

- **2024-12-03**: Initial implementation of WASM, bridge, and state monitoring (Task #16)
  - Added Prometheus metrics for WASM tx failures
  - Implemented signature verification mismatch tracking
  - Created state load error monitoring
  - Built comprehensive Grafana dashboard
  - Configured alert rules with appropriate thresholds
