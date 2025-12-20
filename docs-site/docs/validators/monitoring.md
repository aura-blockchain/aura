---
sidebar_position: 2
---

# Monitoring

Set up comprehensive monitoring for your Aura validator node.

## Prometheus Setup

### Install Prometheus

```bash
wget https://github.com/prometheus/prometheus/releases/download/v2.48.0/prometheus-2.48.0.linux-amd64.tar.gz
tar xvfz prometheus-2.48.0.linux-amd64.tar.gz
sudo mv prometheus-2.48.0.linux-amd64 /opt/prometheus
```

### Configure Prometheus

Create `/opt/prometheus/prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'aura-validator'
    static_configs:
      - targets: ['localhost:26660']  # Tendermint metrics

  - job_name: 'node-exporter'
    static_configs:
      - targets: ['localhost:9100']
```

## Grafana Dashboards

### Install Grafana

```bash
sudo apt-get install -y software-properties-common
sudo add-apt-repository "deb https://packages.grafana.com/oss/deb stable main"
wget -q -O - https://packages.grafana.com/gpg.key | sudo apt-key add -
sudo apt-get update
sudo apt-get install grafana
sudo systemctl start grafana-server
sudo systemctl enable grafana-server
```

Access Grafana at `http://localhost:3000` (default: admin/admin)

### Import Dashboards

Import pre-built Aura validator dashboard:
- Dashboard ID: Available at [Grafana Dashboards](https://grafana.com/grafana/dashboards/)
- Or use custom dashboard from [Aura repository](https://github.com/aura-blockchain/aura/tree/main/grafana)

## Key Metrics to Monitor

### Validator Health
- Block height
- Catching up status
- Missed blocks
- Voting power
- Jailed status

### System Resources
- CPU usage
- Memory usage
- Disk usage
- Network I/O
- Disk I/O

### Network Metrics
- Peer count
- Incoming connections
- Outgoing connections
- P2P latency

## Alerting

### Configure Alerts

Example alert rules for Prometheus:

```yaml
groups:
  - name: validator_alerts
    rules:
      - alert: ValidatorDown
        expr: up{job="aura-validator"} == 0
        for: 5m
        annotations:
          summary: "Validator node is down"

      - alert: HighMissedBlocks
        expr: increase(tendermint_consensus_missed_blocks[1h]) > 10
        annotations:
          summary: "Validator missing too many blocks"

      - alert: LowDiskSpace
        expr: node_filesystem_avail_bytes / node_filesystem_size_bytes < 0.1
        annotations:
          summary: "Disk space below 10%"
```

### Notification Channels

Configure notifications via:
- Email
- Slack
- Telegram
- PagerDuty
- Discord webhooks

## Logs Monitoring

### View Logs

```bash
# Systemd logs
sudo journalctl -u aurad -f

# Filter by severity
sudo journalctl -u aurad -p err -f
```

### Log Aggregation

Use tools like:
- Loki + Promtail
- ELK Stack (Elasticsearch, Logstash, Kibana)
- Datadog

## Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Aura Monitoring Guide](https://github.com/aura-blockchain/aura/blob/main/docs/ops/MONITORING_ALERTING.md)
