# Aura Blockchain - Grafana Metrics Integration

**Target Audience**: AI coding agents and developers

## Metrics Architecture

**Status**: ✅ Metrics fully implemented and wired
**Location**: `/aura/chain/x/monitoring/metrics/prometheus.go`
**Namespace**: `aura_monitoring_*`
**Dashboard**: Aura Network Live Metrics (Grafana Cloud)

### Metrics Endpoints

Aura exposes Prometheus metrics on **3 ports**:

1. **Port 26660** - CometBFT consensus metrics
   - Block production, consensus rounds, validator voting

2. **Port 26661** - Cosmos SDK application metrics
   - Transaction processing, module operations, state changes

3. **Port 9090** - Custom Aura monitoring module metrics (60+ custom metrics)
   - Validator uptime/performance
   - Transaction volume/gas analytics
   - Network health indicators
   - TVL tracking
   - Security event monitoring
   - Alert generation

## Available Metrics

### Core Metrics (Always Available)
```
# Transactions
aura_monitoring_total_transactions{status,module}
aura_monitoring_transaction_gas_used{module}
aura_monitoring_transaction_duration_seconds
aura_monitoring_failed_transactions_total{reason,module}

# Validators
aura_monitoring_validator_uptime_percentage{validator_address,moniker}
aura_monitoring_validator_missed_blocks_total{validator_address,moniker}
aura_monitoring_jailed_validators
aura_monitoring_active_validators

# Network Health
aura_monitoring_block_time_seconds
aura_monitoring_transactions_per_second
aura_monitoring_mempool_size
aura_monitoring_peer_count
aura_monitoring_consensus_health

# Gas/Economics
aura_monitoring_current_gas_price
aura_monitoring_average_gas_price
aura_monitoring_total_tvl
aura_monitoring_tvl_by_module{module}

# Security
aura_monitoring_security_events_total{event_type,severity}
aura_monitoring_threat_level{event_type}
aura_monitoring_mitigated_threats_total

# Alerts
aura_monitoring_alerts_total{type,severity}
aura_monitoring_active_alerts{type,severity}
```

**Full list**: See `/blockchain-projects/METRICS_REFERENCE.md`

## Exposing Metrics

### Automatic Exposure
Metrics are **automatically exposed** when the Aura node starts. No additional configuration required.

```bash
cd /home/decri/blockchain-projects/aura/chain
./aurad start --home ~/.aura
```

**Metrics immediately available at:**
- `http://localhost:26660/metrics` (CometBFT)
- `http://localhost:26661/metrics` (Cosmos SDK App)
- `http://localhost:9090/metrics` (Custom Monitoring)

### Verification

```bash
# Check CometBFT metrics
curl -s http://localhost:26660/metrics | grep tendermint

# Check Cosmos SDK metrics
curl -s http://localhost:26661/metrics | grep cosmos

# Check custom Aura monitoring metrics
curl -s http://localhost:9090/metrics | grep aura_monitoring

# Verify Prometheus is scraping
curl -s http://localhost:9091/targets | grep aura
```

## Prometheus Configuration

**Location**: `/etc/prometheus/prometheus.yml`

```yaml
scrape_configs:
  # Aura CometBFT
  - job_name: 'aura-cometbft'
    static_configs:
      - targets: ['localhost:26660']
        labels:
          blockchain: aura
          component: consensus

  # Aura Cosmos SDK App
  - job_name: 'aura-app'
    static_configs:
      - targets: ['localhost:26661']
        labels:
          blockchain: aura
          component: application

  # Aura Custom Monitoring Module
  - job_name: 'aura-monitoring'
    static_configs:
      - targets: ['localhost:9090']
        labels:
          blockchain: aura
          component: monitoring
```

**Remote Write**: Configured to send to Grafana Cloud (already set up)

## Grafana Dashboard

**Location**: Grafana Cloud - https://altrestackmon.grafana.net
**Dashboard Name**: "Aura Network Live Metrics"
**Public Access**: Enabled (share via external link)

### Accessing the Dashboard

1. **Grafana Cloud** (recommended for investors/stakeholders):
   ```
   https://altrestackmon.grafana.net/dashboards
   Click: "Aura Network Live Metrics"
   Share → Share externally → Copy external link
   ```

2. **Local Grafana** (development):
   ```
   http://localhost:3000
   Login: admin/admin
   ```

### Dashboard Panels

The dashboard shows:
- Transaction volume and throughput
- Validator performance and uptime
- Network health indicators
- Gas price analytics
- TVL metrics
- Security alerts
- Active alerts/anomalies

## Implementation Details

### Metrics Registration

Metrics are registered in `/aura/chain/x/monitoring/metrics/prometheus.go`:

```go
func NewMonitoringMetrics() *MonitoringMetrics {
    // Singleton pattern - metrics registered once
    metricsOnce.Do(func() {
        singletonMetrics = createMonitoringMetrics()
    })
    return singletonMetrics
}
```

### Metrics Updates

Metrics are updated throughout the Aura codebase:

- **Transaction metrics**: Updated in keeper methods when processing txs
- **Validator metrics**: Updated during EndBlock
- **Network metrics**: Updated in consensus/p2p layer
- **Security metrics**: Updated by security module event handlers

### No Additional Wiring Required

All metrics are:
1. ✅ Already implemented in code
2. ✅ Automatically registered on node start
3. ✅ Automatically exposed on HTTP endpoints
4. ✅ Automatically scraped by Prometheus
5. ✅ Automatically pushed to Grafana Cloud
6. ✅ Automatically displayed on dashboard

**Just start the node - metrics flow automatically.**

## Troubleshooting

### Metrics Not Showing

```bash
# 1. Verify node is running
ps aux | grep aurad

# 2. Check metrics endpoints respond
curl http://localhost:26660/metrics
curl http://localhost:26661/metrics
curl http://localhost:9090/metrics

# 3. Verify Prometheus is scraping
curl http://localhost:9091/targets | grep aura

# 4. Check Prometheus logs
sudo journalctl -u prometheus -n 50

# 5. Check node logs
tail -f ~/.aura/data/aura.log
```

### Empty Dashboard

**Cause**: Node not running
**Solution**: Start the Aura node

```bash
cd /home/decri/blockchain-projects/aura/chain
./aurad start --home ~/.aura
```

Metrics appear within 15 seconds of node start (Prometheus scrape interval).

## Adding New Metrics

To add custom metrics (AI agents can do this):

1. **Add metric to struct** in `prometheus.go`:
   ```go
   type MonitoringMetrics struct {
       // ... existing metrics
       MyNewMetric prometheus.Counter
   }
   ```

2. **Register in createMonitoringMetrics()**:
   ```go
   MyNewMetric: promauto.NewCounter(
       prometheus.CounterOpts{
           Namespace: "aura",
           Subsystem: "monitoring",
           Name:      "my_new_metric_total",
           Help:      "Description of metric",
       },
   ),
   ```

3. **Update in keeper methods**:
   ```go
   k.metrics.MyNewMetric.Inc()
   ```

4. **Add to dashboard**: Edit dashboard JSON, add new panel with query:
   ```promql
   aura_monitoring_my_new_metric_total
   ```

No Prometheus config changes needed - new metrics auto-discovered.

## Reference Documents

- **Metrics List**: `/home/decri/blockchain-projects/METRICS_REFERENCE.md`
- **Setup Status**: `/home/decri/blockchain-projects/SETUP_STATUS.md`
- **Prometheus Config**: `/etc/prometheus/prometheus.yml`
- **Dashboard JSON**: `/home/decri/blockchain-projects/dashboards/aura-network-dashboard.json`
