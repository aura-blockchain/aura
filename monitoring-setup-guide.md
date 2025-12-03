# Monitoring Setup Guide - Task #16

## Overview

This guide covers the monitoring infrastructure added in ROADMAP task #16 for tracking:
- **WASM transaction failures** (by contract, error type)
- **State load errors** (by module, store)
- **Signature verification mismatches** (by chain, error reason)

## Quick Start

### 1. Verify Installation

```bash
cd /home/decri/blockchain-projects/aura
bash scripts/verify-monitoring.sh
```

Expected output:
```
1. Alert rules: ✓
2. Grafana dashboard: ✓
3. WASM telemetry: ✓
4. Bridge telemetry: ✓
5. Documentation: ✓

Alert rules defined: 21
Bridge keeper: ✓
WASM keeper: ✓
```

### 2. Start Prometheus

```bash
cd /home/decri/blockchain-projects/aura/prometheus
prometheus --config.file=prometheus.yml
```

Access Prometheus at: http://localhost:9090

### 3. Start Grafana

```bash
cd /home/decri/blockchain-projects/aura/grafana
grafana-server --config=grafana.ini
```

Access Grafana at: http://localhost:3000

### 4. Import Dashboard

1. Navigate to Grafana UI (http://localhost:3000)
2. Go to Dashboards → Import
3. Upload: `/home/decri/blockchain-projects/aura/grafana/dashboards/wasm-bridge-security.json`

## Key Metrics

### WASM Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|----------------|
| `aura_wasm_tx_failures_total` | WASM transaction failures by contract/error | >5% failure rate |
| `aura_wasm_circuit_breaker_state` | Circuit breaker state (0/1/2) | State = 1 (open) |
| `aura_wasm_validation_cache_hits_total` | Cache efficiency | <50% hit rate |
| `aura_wasm_instantiation_failures_total` | Contract instantiation failures | >5 in 10min |

### Bridge Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|----------------|
| `aura_bridge_signature_mismatches_total` | Signature verification failures | >1% mismatch rate |
| `aura_bridge_invalid_recovery_id_total` | Invalid recovery IDs | >0.1/sec |
| `aura_bridge_pubkey_recovery_failures_total` | Public key recovery failures | >2% failure rate |
| `aura_bridge_signature_verification_duration_seconds` | Verification latency (p95) | >100ms |

### State Integrity Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|----------------|
| `aura_state_load_errors_total` | State load errors by module | >0 errors |
| `aura_unmarshal_errors_total` | Protobuf unmarshal failures | >0 errors |
| `aura_state_corruption_total` | Detected corruption events | >0 events |
| `aura_kvstore_iteration_errors_total` | Iterator errors | >0.01/sec |

## Critical Alerts

### WasmTxFailureRateHigh
**Trigger**: WASM failure rate > 5% for 5 minutes
**Action**: Check contract logs, review error types, investigate for attacks

### StateLoadErrorsDetected
**Trigger**: Any state load errors for 2 minutes
**Action**: URGENT - Check database integrity, verify protobuf schemas

### SignatureMismatchRateHigh
**Trigger**: Signature mismatch rate > 1% for 5 minutes
**Action**: Investigate for attacks, verify wallet compatibility

### WasmCircuitBreakerOpen
**Trigger**: Circuit breaker open for 1 minute
**Action**: Check contract registry, system in degraded/permissive mode

## Query Examples

### Check WASM failure rate
```promql
rate(aura_wasm_tx_failures_total[5m]) /
clamp_min(rate(aura_wasm_tx_total[5m]), 0.0001)
```

### Check signature mismatches by chain
```promql
rate(aura_bridge_signature_mismatches_total[5m])
```

### Check state errors by module
```promql
rate(aura_state_load_errors_total[5m])
```

### Circuit breaker status
```promql
aura_wasm_circuit_breaker_state
```

## Files Created

### Prometheus
- `/prometheus/rules/wasm-bridge-alerts.yml` - Alert rules (21 alerts)

### Grafana
- `/grafana/dashboards/wasm-bridge-security.json` - Dashboard with 15 panels

### Code
- `/chain/x/wasm/keeper/telemetry.go` - WASM module metrics
- `/chain/x/bridge/keeper/telemetry.go` - Bridge module metrics
- `/chain/x/bridge/keeper/keeper.go` - Instrumented with telemetry calls

### Documentation
- `/docs/monitoring/wasm-bridge-security-monitoring.md` - Comprehensive guide
- `/scripts/verify-monitoring.sh` - Verification script
- `/monitoring-setup-guide.md` - This file

## Testing

### Verify metrics are exposed
```bash
curl http://localhost:9090/metrics | grep aura_wasm
curl http://localhost:9090/metrics | grep aura_bridge
```

### Check active alerts
```bash
curl http://localhost:9090/api/v1/alerts
```

### Query specific metric
```bash
curl 'http://localhost:9090/api/v1/query?query=aura_wasm_tx_total'
```

## Troubleshooting

### Metrics not appearing
1. Verify Prometheus is scraping: http://localhost:9090/targets
2. Check aurad is exposing metrics on port 9090
3. Verify telemetry code is being executed (check logs)

### Alerts not firing
1. Check alert rules loaded: http://localhost:9090/rules
2. Verify evaluation interval in prometheus.yml
3. Check `for` duration in alert rules

### Dashboard not loading
1. Verify Prometheus datasource in Grafana
2. Check dashboard JSON is valid
3. Verify metric names match code

## Security

- Restrict metrics endpoint to internal network in production
- Use authentication for Prometheus/Grafana
- Avoid exposing sensitive data in labels
- Regular security audits of alert thresholds

## Maintenance

- **Daily**: Review active alerts
- **Weekly**: Analyze metric trends, tune thresholds
- **Monthly**: Review retention, optimize queries
- **Quarterly**: Update runbooks, refine dashboards

## Next Steps

1. Start monitoring infrastructure (Prometheus + Grafana)
2. Import Grafana dashboard
3. Run Aura node and generate test traffic
4. Verify metrics are being collected
5. Test alert firing by triggering conditions
6. Review and tune alert thresholds based on baseline

## Support

For detailed documentation, see:
- `/docs/monitoring/wasm-bridge-security-monitoring.md`

For issues or questions, check:
- Prometheus logs: `journalctl -u prometheus`
- Grafana logs: `journalctl -u grafana`
- Application logs: `journalctl -u aurad`
