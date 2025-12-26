# AURA Testnet Faucet - Monitoring and Alerting Guide

This document describes the monitoring, logging, and alerting setup for the AURA Testnet Faucet.

## Table of Contents

1. [Overview](#overview)
2. [Health Checks](#health-checks)
3. [Logging Configuration](#logging-configuration)
4. [Metrics Collection](#metrics-collection)
5. [Alerting Rules](#alerting-rules)
6. [Dashboard Setup](#dashboard-setup)
7. [Log Analysis](#log-analysis)
8. [Performance Monitoring](#performance-monitoring)

## Overview

### Monitoring Stack
- **Application**: Built-in health checks and structured logging
- **Metrics**: Prometheus (optional)
- **Visualization**: Grafana (optional)
- **Logging**: JSON structured logs
- **Alerting**: AlertManager + custom webhooks

### Key Metrics
- Request rate (requests/minute)
- Success/failure rates
- Response times (p50, p95, p99)
- Faucet balance
- Database connection pool
- Redis memory usage
- Active connections
- Error rates

## Health Checks

### Built-in Health Endpoint

The faucet provides a `/api/v1/health` endpoint that checks:
- API server status
- Database connectivity
- Redis connectivity
- Blockchain node status

**Example Response**:
```json
{
  "status": "healthy",
  "network": "aura-testnet-1",
  "height": "123456"
}
```

### Health Check Script

```bash
#!/bin/bash
# health_check.sh

ENDPOINT="http://localhost:8080/api/v1/health"
RESPONSE=$(curl -s -w "%{http_code}" $ENDPOINT)
HTTP_CODE="${RESPONSE: -3}"

if [ "$HTTP_CODE" -eq 200 ]; then
    echo "✓ Faucet is healthy"
    exit 0
else
    echo "✗ Faucet is unhealthy (HTTP $HTTP_CODE)"
    exit 1
fi
```

### Docker Health Checks

Already configured in `docker-compose.yml`:
```yaml
healthcheck:
  test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/api/v1/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 40s
```

## Logging Configuration

### Log Levels

Configure via `LOG_LEVEL` environment variable:
- `debug`: Detailed information for debugging
- `info`: General informational messages (default)
- `warn`: Warning messages
- `error`: Error messages

### Log Format

All logs are structured JSON:
```json
{
  "level": "info",
  "msg": "Token request received",
  "address": "aura1...",
  "ip": "192.168.1.1",
  "time": "2025-11-20T12:00:00Z"
}
```

### Important Log Entries

#### Successful Request
```json
{
  "level": "info",
  "msg": "Tokens sent successfully",
  "tx_hash": "ABCD1234...",
  "recipient": "aura1...",
  "amount": 100000000,
  "time": "2025-11-20T12:00:00Z"
}
```

#### Failed Request
```json
{
  "level": "error",
  "msg": "Failed to send tokens",
  "error": "insufficient funds",
  "recipient": "aura1...",
  "time": "2025-11-20T12:00:00Z"
}
```

#### Rate Limited
```json
{
  "level": "warn",
  "msg": "Rate limit exceeded",
  "ip": "192.168.1.1",
  "type": "ip_based",
  "time": "2025-11-20T12:00:00Z"
}
```

### Log Aggregation

#### Using Docker Logs

```bash
# View all logs
docker-compose logs -f

# View specific service
docker-compose logs -f faucet-backend

# Filter by level
docker-compose logs faucet-backend | grep "level\":\"error"

# Follow last 100 lines
docker-compose logs -f --tail=100 faucet-backend
```

#### Using Promtail + Loki

**docker-compose.loki.yml**:
```yaml
version: '3.8'

services:
  loki:
    image: grafana/loki:latest
    ports:
      - "3100:3100"
    volumes:
      - ./loki-config.yaml:/etc/loki/config.yaml
      - loki_data:/loki
    command: -config.file=/etc/loki/config.yaml
    networks:
      - faucet-network

  promtail:
    image: grafana/promtail:latest
    volumes:
      - ./promtail-config.yaml:/etc/promtail/config.yaml
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
      - /var/run/docker.sock:/var/run/docker.sock
    command: -config.file=/etc/promtail/config.yaml
    networks:
      - faucet-network

volumes:
  loki_data:

networks:
  faucet-network:
    external: true
```

**loki-config.yaml**:
```yaml
auth_enabled: false

server:
  http_listen_port: 3100

ingester:
  lifecycler:
    ring:
      kvstore:
        store: inmemory
      replication_factor: 1
  chunk_idle_period: 5m
  chunk_retain_period: 30s

schema_config:
  configs:
    - from: 2025-01-01
      store: boltdb
      object_store: filesystem
      schema: v11
      index:
        prefix: index_
        period: 24h

storage_config:
  boltdb:
    directory: /loki/index
  filesystem:
    directory: /loki/chunks

limits_config:
  enforce_metric_name: false
  reject_old_samples: true
  reject_old_samples_max_age: 168h
```

**promtail-config.yaml**:
```yaml
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  - job_name: docker
    docker_sd_configs:
      - host: unix:///var/run/docker.sock
        refresh_interval: 5s
        filters:
          - name: label
            values: ["com.docker.compose.project=aura-faucet"]
    relabel_configs:
      - source_labels: ['__meta_docker_container_name']
        regex: '/(.*)'
        target_label: 'container'
      - source_labels: ['__meta_docker_container_log_stream']
        target_label: 'stream'
```

## Metrics Collection

### Application Metrics

Key metrics to track:

1. **Request Metrics**
   - Total requests
   - Successful requests
   - Failed requests
   - Rate limited requests

2. **Performance Metrics**
   - Response time (avg, p50, p95, p99)
   - Database query time
   - Transaction broadcast time

3. **Business Metrics**
   - Total distributed amount
   - Unique recipients
   - Faucet balance
   - Requests per hour/day

4. **System Metrics**
   - CPU usage
   - Memory usage
   - Disk I/O
   - Network I/O

### Prometheus Configuration

**prometheus.yml**:
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'faucet'
    static_configs:
      - targets: ['faucet-backend:8080']
    metrics_path: '/api/v1/faucet/stats'

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']

  - job_name: 'node'
    static_configs:
      - targets: ['node-exporter:9100']
```

**docker-compose.monitoring.yml**:
```yaml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    networks:
      - faucet-network

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./grafana/datasources:/etc/grafana/provisioning/datasources
    networks:
      - faucet-network

  postgres-exporter:
    image: prometheuscommunity/postgres-exporter:latest
    environment:
      DATA_SOURCE_NAME: "postgresql://faucet:password@postgres:5432/faucet?sslmode=disable"
    networks:
      - faucet-network

  redis-exporter:
    image: oliver006/redis_exporter:latest
    environment:
      REDIS_ADDR: "redis:6379"
    networks:
      - faucet-network

  node-exporter:
    image: prom/node-exporter:latest
    command:
      - '--path.rootfs=/host'
    volumes:
      - '/:/host:ro,rslave'
    networks:
      - faucet-network

volumes:
  prometheus_data:
  grafana_data:

networks:
  faucet-network:
    external: true
```

## Alerting Rules

### AlertManager Configuration

**alertmanager.yml**:
```yaml
global:
  resolve_timeout: 5m

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
  receiver: 'team-notifications'

receivers:
  - name: 'team-notifications'
    email_configs:
      - to: 'ops@aura-chain.com'
        from: 'alertmanager@aura-chain.com'
        smarthost: 'smtp.gmail.com:587'
        auth_username: 'alertmanager@aura-chain.com'
        auth_password: 'password'

    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK'
        channel: '#faucet-alerts'
        title: 'AURA Faucet Alert'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'

    webhook_configs:
      - url: 'http://your-webhook-endpoint/alerts'
```

### Alert Rules

**alerts.yml**:
```yaml
groups:
  - name: faucet_alerts
    interval: 30s
    rules:
      # Service is down
      - alert: FaucetDown
        expr: up{job="faucet"} == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Faucet service is down"
          description: "Faucet has been down for more than 2 minutes"

      # High error rate
      - alert: HighErrorRate
        expr: rate(faucet_requests_failed[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          description: "Error rate is above 10% for 5 minutes"

      # Low faucet balance
      - alert: LowFaucetBalance
        expr: faucet_balance < 10000000000
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Faucet balance is low"
          description: "Faucet balance is below 10,000 AURA"

      # Critical faucet balance
      - alert: CriticalFaucetBalance
        expr: faucet_balance < 1000000000
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Faucet balance is critically low"
          description: "Faucet balance is below 1,000 AURA"

      # High request rate
      - alert: HighRequestRate
        expr: rate(faucet_requests_total[1m]) > 10
        for: 5m
        labels:
          severity: info
        annotations:
          summary: "High request rate"
          description: "Request rate is above 10 per minute"

      # Database connection issues
      - alert: DatabaseConnectionIssues
        expr: postgresql_up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Database connection failed"
          description: "Cannot connect to PostgreSQL database"

      # Redis connection issues
      - alert: RedisConnectionIssues
        expr: redis_up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Redis connection failed"
          description: "Cannot connect to Redis"

      # High memory usage
      - alert: HighMemoryUsage
        expr: (node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes > 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: "Memory usage is above 90%"

      # High CPU usage
      - alert: HighCPUUsage
        expr: 100 - (avg by (instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High CPU usage"
          description: "CPU usage is above 80%"

      # Disk space low
      - alert: DiskSpaceLow
        expr: (node_filesystem_avail_bytes / node_filesystem_size_bytes) * 100 < 20
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Low disk space"
          description: "Disk space is below 20%"
```

### Custom Alert Script

```bash
#!/bin/bash
# alert_script.sh - Custom alerting script

FAUCET_URL="http://localhost:8080/api/v1"
SLACK_WEBHOOK="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
EMAIL="ops@aura-chain.com"

# Check faucet balance
check_balance() {
    BALANCE=$(curl -s "$FAUCET_URL/faucet/info" | jq -r '.balance')
    MIN_BALANCE=1000000000 # 1000 AURA

    if [ "$BALANCE" -lt "$MIN_BALANCE" ]; then
        send_alert "critical" "Low Faucet Balance" "Balance is ${BALANCE} uaura"
    fi
}

# Check error rate
check_errors() {
    STATS=$(curl -s "$FAUCET_URL/faucet/stats")
    FAILED=$(echo "$STATS" | jq -r '.failed_requests')
    TOTAL=$(echo "$STATS" | jq -r '.total_requests')

    if [ "$TOTAL" -gt 0 ]; then
        ERROR_RATE=$(echo "scale=2; $FAILED / $TOTAL * 100" | bc)
        if [ $(echo "$ERROR_RATE > 10" | bc) -eq 1 ]; then
            send_alert "warning" "High Error Rate" "Error rate is ${ERROR_RATE}%"
        fi
    fi
}

# Send alert
send_alert() {
    SEVERITY=$1
    TITLE=$2
    MESSAGE=$3

    # Send to Slack
    curl -X POST -H 'Content-type: application/json' \
        --data "{\"text\":\"[$SEVERITY] $TITLE: $MESSAGE\"}" \
        "$SLACK_WEBHOOK"

    # Send email (requires mailutils)
    echo "$MESSAGE" | mail -s "[$SEVERITY] $TITLE" "$EMAIL"

    # Log alert
    echo "$(date): [$SEVERITY] $TITLE - $MESSAGE" >> /var/log/faucet-alerts.log
}

# Run checks
check_balance
check_errors
```

**Cron job**:
```bash
# Run every 5 minutes
*/5 * * * * /path/to/alert_script.sh
```

## Dashboard Setup

### Grafana Dashboard

**faucet-dashboard.json** (simplified):
```json
{
  "dashboard": {
    "title": "AURA Faucet Dashboard",
    "panels": [
      {
        "title": "Request Rate",
        "targets": [
          {
            "expr": "rate(faucet_requests_total[5m])"
          }
        ]
      },
      {
        "title": "Success Rate",
        "targets": [
          {
            "expr": "rate(faucet_requests_successful[5m]) / rate(faucet_requests_total[5m])"
          }
        ]
      },
      {
        "title": "Faucet Balance",
        "targets": [
          {
            "expr": "faucet_balance"
          }
        ]
      },
      {
        "title": "Response Time (p95)",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))"
          }
        ]
      }
    ]
  }
}
```

### Key Dashboard Panels

1. **Overview Panel**
   - Current status (up/down)
   - Total requests (24h)
   - Success rate (24h)
   - Faucet balance

2. **Request Metrics**
   - Request rate (over time)
   - Success/failure breakdown
   - Rate limited requests

3. **Performance**
   - Response time (p50, p95, p99)
   - Database query time
   - API latency

4. **System Health**
   - CPU usage
   - Memory usage
   - Disk usage
   - Network I/O

5. **Business Metrics**
   - Total distributed (over time)
   - Unique recipients
   - Geographic distribution
   - Popular request times

## Log Analysis

### Common Log Queries

```bash
# Count errors in last hour
docker-compose logs --since 1h faucet-backend | grep "level\":\"error" | wc -l

# Top requesting IPs
docker-compose logs faucet-backend | grep "Token request received" | jq -r '.ip' | sort | uniq -c | sort -rn | head -10

# Failed transactions
docker-compose logs faucet-backend | grep "Failed to send tokens" | jq '.error' | sort | uniq -c

# Rate limited requests by IP
docker-compose logs faucet-backend | grep "Rate limit exceeded" | jq -r '.ip' | sort | uniq -c | sort -rn
```

### Log Analysis Tools

#### Using jq
```bash
# Parse JSON logs
docker-compose logs faucet-backend | grep "^{" | jq -r 'select(.level=="error") | "\(.time) \(.msg) \(.error)"'

# Statistics
docker-compose logs --since 24h faucet-backend | grep "Tokens sent successfully" | jq '.amount' | awk '{sum+=$1} END {print "Total distributed:", sum/1000000, "AURA"}'
```

#### Using LogCLI (Loki)
```bash
# Query logs from Loki
logcli query '{container="faucet-backend"}' --limit=100 --since=1h

# Filter by level
logcli query '{container="faucet-backend"} |= "level\":\"error"'

# Aggregate stats
logcli stats '{container="faucet-backend"}'
```

## Performance Monitoring

### Key Performance Indicators

1. **Response Time**: < 500ms (p95)
2. **Availability**: > 99.9%
3. **Error Rate**: < 1%
4. **Database Query Time**: < 100ms (p95)
5. **Request Throughput**: > 10 req/sec

### Performance Monitoring Script

```bash
#!/bin/bash
# performance_monitor.sh

ENDPOINT="http://localhost:8080/api/v1"

echo "=== AURA Faucet Performance Report ==="
echo "Date: $(date)"
echo

# Health check
echo "Health Check:"
curl -s "$ENDPOINT/health" | jq '.'
echo

# Statistics
echo "Statistics:"
curl -s "$ENDPOINT/faucet/stats" | jq '.'
echo

# Recent transactions
echo "Recent Transactions (last 5):"
curl -s "$ENDPOINT/faucet/recent" | jq '.transactions[:5]'
echo

# Database connections
echo "Database Connections:"
docker-compose exec postgres psql -U faucet -d faucet -c "SELECT count(*) FROM pg_stat_activity;" -t
echo

# Redis memory
echo "Redis Memory Usage:"
docker-compose exec redis redis-cli INFO memory | grep used_memory_human
echo

# Container stats
echo "Container Resource Usage:"
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}"
```

---

**Document Version**: 1.0.0
**Last Updated**: 2025-11-20
