# AURA Monitoring & Alerting Guide

**Version:** 1.0
**Last Updated:** 2025-11-25
**Target Audience:** DevOps, SRE, Node Operators

---

## Table of Contents

1. [Overview](#overview)
2. [Monitoring Stack](#monitoring-stack)
3. [Prometheus Setup](#prometheus-setup)
4. [Grafana Dashboards](#grafana-dashboards)
5. [Alert Rules](#alert-rules)
6. [Log Aggregation](#log-aggregation)
7. [Health Checks](#health-checks)
8. [Performance Metrics](#performance-metrics)
9. [SLA Targets](#sla-targets)

---

## Overview

AURA provides built-in monitoring capabilities with Prometheus metrics, pre-configured alert rules, and Grafana dashboards.

**Key Components:**
- **Prometheus**: Metrics collection and storage
- **Grafana**: Visualization and dashboards
- **Alertmanager**: Alert routing and notification
- **ELK/Loki**: Log aggregation (optional)

---

## Monitoring Stack

### Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│ AURA Nodes   │────▶│  Prometheus  │────▶│   Grafana    │
│ (Exporters)  │     │   (Metrics)  │     │ (Dashboards) │
└──────────────┘     └──────┬───────┘     └──────────────┘
                            │
                     ┌──────▼───────┐
                     │ Alertmanager │
                     │(Notifications)│
                     └──────────────┘
```

### Metrics Exposed

**Tendermint Metrics** (Port 26660):
- Consensus metrics
- Mempool metrics
- P2P metrics
- Block height, time

**AURA Application Metrics** (Port 1317):
- Custom module metrics
- Transaction metrics
- Security events
- Business metrics

---

## Prometheus Setup

### Installation

```bash
# Download Prometheus
wget https://github.com/prometheus/prometheus/releases/download/v2.45.0/prometheus-2.45.0.linux-amd64.tar.gz
tar xvfz prometheus-2.45.0.linux-amd64.tar.gz
cd prometheus-2.45.0.linux-amd64

# Install to /opt
sudo mv prometheus promtool /usr/local/bin/
sudo mkdir -p /etc/prometheus /var/lib/prometheus
sudo mv consoles console_libraries /etc/prometheus/
```

### Configuration

```yaml
# /etc/prometheus/prometheus.yml

global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: 'aura-mainnet'
    environment: 'production'

# Load alert rules
rule_files:
  - "/etc/prometheus/rules/*.yml"

# Alertmanager configuration
alerting:
  alertmanagers:
    - static_configs:
        - targets:
            - localhost:9093

# Scrape configurations
scrape_configs:
  # AURA validator node
  - job_name: 'aura-validator'
    static_configs:
      - targets: ['10.0.1.1:26660']  # Tendermint metrics
        labels:
          instance: 'validator-1'
          node_type: 'validator'

      - targets: ['10.0.1.1:1317']   # AURA app metrics (if exposed)
        labels:
          instance: 'validator-1'
          node_type: 'validator'

  # AURA sentry nodes
  - job_name: 'aura-sentry'
    static_configs:
      - targets:
          - 'sentry1.aura.network:26660'
          - 'sentry2.aura.network:26660'
          - 'sentry3.aura.network:26660'
        labels:
          node_type: 'sentry'

  # Full nodes
  - job_name: 'aura-fullnode'
    static_configs:
      - targets:
          - 'fullnode1:26660'
          - 'fullnode2:26660'
        labels:
          node_type: 'fullnode'

  # Node exporter (system metrics)
  - job_name: 'node'
    static_configs:
      - targets:
          - 'localhost:9100'
          - 'sentry1:9100'
          - 'sentry2:9100'

  # Process exporter (aurad process)
  - job_name: 'process'
    static_configs:
      - targets:
          - 'localhost:9256'
```

### Alert Rules

Copy existing alert rules from repository:

```bash
# Copy AURA alert rules
sudo cp /home/decri/blockchain-projects/aura/prometheus/rules/monitoring-alerts.yml \
  /etc/prometheus/rules/

# Verify rules
promtool check rules /etc/prometheus/rules/*.yml

# Reload Prometheus
sudo systemctl reload prometheus
```

### Systemd Service

```bash
sudo tee /etc/systemd/system/prometheus.service > /dev/null <<EOF
[Unit]
Description=Prometheus
Wants=network-online.target
After=network-online.target

[Service]
User=prometheus
Group=prometheus
Type=simple
ExecStart=/usr/local/bin/prometheus \\
  --config.file=/etc/prometheus/prometheus.yml \\
  --storage.tsdb.path=/var/lib/prometheus/ \\
  --web.console.templates=/etc/prometheus/consoles \\
  --web.console.libraries=/etc/prometheus/console_libraries \\
  --web.listen-address=0.0.0.0:9090 \\
  --storage.tsdb.retention.time=90d

Restart=always

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable prometheus
sudo systemctl start prometheus
```

---

## Grafana Dashboards

### Installation

```bash
# Install Grafana
sudo apt-get install -y software-properties-common
wget -q -O - https://packages.grafana.com/gpg.key | sudo apt-key add -
echo "deb https://packages.grafana.com/oss/deb stable main" | sudo tee /etc/apt/sources.list.d/grafana.list
sudo apt-get update
sudo apt-get install -y grafana

# Start Grafana
sudo systemctl start grafana-server
sudo systemctl enable grafana-server

# Access at http://localhost:3000 (admin/admin)
```

### Configure Data Source

```bash
# Add Prometheus as data source via UI or API
curl -X POST http://admin:admin@localhost:3000/api/datasources \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Prometheus",
    "type": "prometheus",
    "url": "http://localhost:9090",
    "access": "proxy",
    "isDefault": true
  }'
```

### Import AURA Dashboards

```bash
# Import pre-built dashboard
# Located at: /home/decri/blockchain-projects/aura/grafana/dashboards/security-monitoring.json

# Via UI:
# 1. Go to Dashboards → Import
# 2. Upload JSON file or paste JSON content
# 3. Select Prometheus data source
# 4. Import

# Via API:
curl -X POST http://admin:admin@localhost:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @/home/decri/blockchain-projects/aura/grafana/dashboards/security-monitoring.json
```

### Key Dashboards

**1. Validator Dashboard**
- Uptime percentage
- Missed blocks counter
- Signing percentage trend
- Voting power
- Commission earnings
- Delegator count

**2. Node Health Dashboard**
- Block height
- Sync status
- Peer count
- Memory usage
- CPU utilization
- Disk I/O

**3. Security Monitoring Dashboard** (Existing)
- Active alerts by severity
- Security events by type
- Anomaly detections
- Failed transactions
- Threat level indicators

**4. Network Dashboard**
- Total validators
- Network uptime
- Transaction throughput
- Gas price trends
- Mempool size

---

## Alert Rules

AURA includes comprehensive alert rules at:
`/home/decri/blockchain-projects/aura/prometheus/rules/monitoring-alerts.yml`

### Alert Categories

**1. Network Health Alerts**

```yaml
# High network congestion
- alert: HighNetworkCongestion
  expr: aura_monitoring_network_congestion > 0.8
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Network congestion critically high"

# Low consensus health
- alert: LowConsensusHealth
  expr: aura_monitoring_consensus_health < 0.7
  for: 3m
  labels:
    severity: high
  annotations:
    summary: "Consensus health degraded"

# High block time
- alert: HighBlockTime
  expr: aura_monitoring_block_time_seconds > 10
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Block time elevated"
```

**2. Validator Alerts**

```yaml
# Validator not signing
- alert: ValidatorDown
  expr: aura_monitoring_validator_uptime_percentage < 95
  for: 10m
  labels:
    severity: critical
  annotations:
    summary: "Validator experiencing downtime"

# Validator jailed
- alert: ValidatorJailed
  expr: aura_monitoring_jailed_validators > 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "Validator(s) jailed"
```

**3. Security Alerts**

```yaml
# High anomaly rate
- alert: HighAnomalyRate
  expr: rate(aura_monitoring_anomaly_detections_total[5m]) > 0.1
  for: 5m
  labels:
    severity: high
  annotations:
    summary: "Elevated anomaly detection rate"

# Security threat detected
- alert: SecurityEventThreat
  expr: aura_monitoring_threat_level > 7
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "High-severity security event"
```

**4. System Alerts**

```yaml
# High mempool size
- alert: HighMemPoolSize
  expr: aura_monitoring_mempool_size > 10000
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Mempool congested"

# Sync lag
- alert: ExplorerSyncLag
  expr: aura_monitoring_explorer_sync_lag_blocks > 100
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Node sync lagging"
```

### Alertmanager Configuration

```yaml
# /etc/prometheus/alertmanager.yml

global:
  resolve_timeout: 5m
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_from: 'alerts@aura.network'
  smtp_auth_username: 'alerts@aura.network'
  smtp_auth_password: 'your-password'

# Route tree
route:
  group_by: ['alertname', 'cluster', 'severity']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
  receiver: 'team-email'

  routes:
    # Critical alerts → PagerDuty + Email + Slack
    - match:
        severity: critical
      receiver: 'pagerduty'
      continue: true

    - match:
        severity: critical
      receiver: 'team-email'
      continue: true

    - match:
        severity: critical
      receiver: 'slack-critical'

    # High severity → Email + Slack
    - match:
        severity: high
      receiver: 'team-email'
      continue: true

    - match:
        severity: high
      receiver: 'slack-high'

    # Warning → Slack only
    - match:
        severity: warning
      receiver: 'slack-warnings'

# Receivers
receivers:
  - name: 'team-email'
    email_configs:
      - to: 'ops@yourcompany.com'
        headers:
          Subject: '{{ .GroupLabels.severity | toUpper }} - {{ .GroupLabels.alertname }}'

  - name: 'pagerduty'
    pagerduty_configs:
      - service_key: 'YOUR_PAGERDUTY_KEY'
        description: '{{ .CommonAnnotations.summary }}'

  - name: 'slack-critical'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK'
        channel: '#alerts-critical'
        title: '🚨 CRITICAL: {{ .GroupLabels.alertname }}'
        text: '{{ .CommonAnnotations.summary }}'
        send_resolved: true

  - name: 'slack-high'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK'
        channel: '#alerts-high'
        title: '⚠️  HIGH: {{ .GroupLabels.alertname }}'
        text: '{{ .CommonAnnotations.summary }}'

  - name: 'slack-warnings'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK'
        channel: '#alerts-warnings'
        title: 'ℹ️ Warning: {{ .GroupLabels.alertname }}'
        text: '{{ .CommonAnnotations.summary }}'

# Inhibition rules (silence related alerts)
inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'instance']
```

---

## Log Aggregation

### ELK Stack (Elasticsearch, Logstash, Kibana)

**Filebeat Configuration:**

```yaml
# /etc/filebeat/filebeat.yml

filebeat.inputs:
  # AURA systemd logs
  - type: journald
    id: aurad-logs
    include_matches:
      - _SYSTEMD_UNIT=aurad.service

  # System logs
  - type: log
    enabled: true
    paths:
      - /var/log/auth.log
      - /var/log/syslog

# Elasticsearch output
output.elasticsearch:
  hosts: ["https://elasticsearch.yourcompany.com:9200"]
  index: "aura-logs-%{+yyyy.MM.dd}"
  username: "elastic"
  password: "${ES_PASSWORD}"

# Kibana dashboards
setup.kibana:
  host: "https://kibana.yourcompany.com:5601"

# Processors
processors:
  - add_host_metadata: ~
  - add_cloud_metadata: ~
  - add_fields:
      target: ''
      fields:
        environment: production
        cluster: aura-mainnet
```

### Loki (Alternative to ELK)

```yaml
# /etc/promtail/config.yml

server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  - job_name: journal
    journal:
      max_age: 12h
      labels:
        job: systemd-journal
    relabel_configs:
      - source_labels: ['__journal__systemd_unit']
        target_label: 'unit'

      - source_labels: ['__journal__hostname']
        target_label: 'hostname'

  - job_name: aurad
    static_configs:
      - targets:
          - localhost
        labels:
          job: aurad
          __path__: /var/log/aurad/*.log
```

---

## Health Checks

### Endpoint Health Checks

```bash
#!/bin/bash
# health-check.sh

# RPC health
RPC_STATUS=$(curl -s http://localhost:26657/health | jq -r .result)
if [ "$RPC_STATUS" != "{}" ]; then
  echo "ERROR: RPC unhealthy"
  exit 1
fi

# API health
API_STATUS=$(curl -s http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info | jq -r .default_node_info.network)
if [ "$API_STATUS" != "aura-mainnet-1" ]; then
  echo "ERROR: API unhealthy"
  exit 1
fi

# Sync status
CATCHING_UP=$(curl -s http://localhost:26657/status | jq -r .result.sync_info.catching_up)
if [ "$CATCHING_UP" = "true" ]; then
  echo "WARNING: Node catching up"
  exit 2
fi

# Peer count
PEERS=$(curl -s http://localhost:26657/net_info | jq -r '.result.n_peers | tonumber')
if [ "$PEERS" -lt 3 ]; then
  echo "ERROR: Low peer count: $PEERS"
  exit 1
fi

echo "OK: All health checks passed"
exit 0
```

### Automated Health Monitoring

```bash
# Add to crontab
*/5 * * * * /usr/local/bin/health-check.sh || echo "Health check failed" | mail -s "AURA Health Alert" ops@yourcompany.com
```

---

## Performance Metrics

### Key Performance Indicators

**Blockchain Metrics:**
- Block height progression rate
- Block time average (target: 6s)
- Transaction throughput (TPS)
- Mempool size

**Validator Metrics:**
- Uptime percentage (target: 99.9%+)
- Signing percentage (target: 100%)
- Missed blocks (target: 0)
- Voting power percentage

**System Metrics:**
- CPU utilization (target: <70%)
- Memory usage (target: <80%)
- Disk I/O (IOPS, latency)
- Network bandwidth

**Application Metrics:**
- RPC request rate
- API request rate
- Error rate (target: <0.1%)
- Response time (p95, p99)

### Performance Queries

```promql
# Block time average (5m)
rate(cometbft_consensus_block_interval_seconds_sum[5m]) / rate(cometbft_consensus_block_interval_seconds_count[5m])

# Transaction throughput
rate(cometbft_consensus_total_txs[5m])

# Validator uptime %
(1 - (aura_validatorsecurity_missed_blocks / cometbft_consensus_height)) * 100

# Memory usage %
(process_resident_memory_bytes / node_memory_MemTotal_bytes) * 100

# API p95 response time
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

---

## SLA Targets

### Production SLA

**Validator SLA:**
- Uptime: 99.9% (8.76 hours downtime/year)
- Signing: 99.9% of blocks
- Response Time: <100ms (RPC queries)
- Incident Response: <5 minutes (critical)

**Full Node SLA:**
- Uptime: 99.5% (43.8 hours downtime/year)
- Response Time: <200ms (RPC queries)
- Sync Lag: <100 blocks
- Incident Response: <15 minutes (critical)

### Monitoring SLA Compliance

```promql
# Uptime SLA (30 days)
avg_over_time(up{job="aura-validator"}[30d]) * 100

# Query performance SLA
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[30d])) < 0.1
```

---

**Document Status**: Production Ready
**Review Cycle**: Quarterly
**Next Review**: 2026-02-25

**Related Documentation:**
- Existing Alert Rules: `/home/decri/blockchain-projects/aura/prometheus/rules/monitoring-alerts.yml`
- Existing Dashboards: `/home/decri/blockchain-projects/aura/grafana/dashboards/security-monitoring.json`
- Production Deployment Guide: `PRODUCTION_DEPLOYMENT_GUIDE.md`
- Troubleshooting Guide: `TROUBLESHOOTING.md`
