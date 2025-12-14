# Consensus Monitoring - Quick Start Guide

Get up and running with CometBFT consensus monitoring in 5 minutes.

## Prerequisites

- Aura testnet running (4 validators)
- Docker and docker-compose installed
- Prometheus and Grafana containers running

## Quick Setup

### 1. Verify Testnet is Running

```bash
docker ps --filter "name=aura-validator" --format "{{.Names}} - {{.Status}}"
```

Expected output:
```
aura-validator-1 - Up X minutes
aura-validator-2 - Up X minutes
aura-validator-3 - Up X minutes
aura-validator-4 - Up X minutes
```

### 2. Check Consensus Metrics

```bash
curl -s http://localhost:27660/metrics | grep "^cometbft_consensus" | head -5
```

Expected output:
```
cometbft_consensus_height{validator="validator-1"} 1234
cometbft_consensus_round{validator="validator-1"} 0
cometbft_consensus_step{validator="validator-1"} 3
...
```

### 3. Verify Prometheus is Scraping

```bash
curl -s http://localhost:9094/api/v1/query?query=cometbft_consensus_height | jq '.data.result[].metric.validator'
```

Expected output:
```
"validator-1"
"validator-2"
"validator-3"
"validator-4"
```

### 4. Access Grafana Dashboard

1. Open browser to: `http://localhost:3002`
2. Login: `admin` / `admin`
3. Navigate to: **Dashboards → CometBFT Consensus Monitoring**

You should see:
- ✅ Current height increasing
- ✅ Round = 0 (green)
- ✅ Block production rate ~6 blocks/min
- ✅ All validators showing data

### 5. Test Alert Rules

```bash
# Check that alerts are loaded
curl -s http://localhost:9094/api/v1/rules | jq '.data.groups[] | select(.name == "consensus-health") | .rules[].name'
```

Expected output:
```
"ConsensusStuck"
"HighConsensusRounds"
"ConsensusRoundStuck"
"LowPrevoteParticipation"
...
```

## Quick Tests

### Test 1: Simulate Validator Failure

```bash
# Stop one validator
docker stop aura-validator-4

# Watch consensus continue with 3/4 validators
watch "curl -s http://localhost:9094/api/v1/query?query=cometbft_consensus_height | jq -r '.data.result[].value[1]'"

# Restart validator
docker start aura-validator-4
```

**What to observe**:
- Height continues to increase (3/4 = 75% > 67% required)
- Validator-4 catches up after restart
- No critical alerts (3/4 validators still meeting quorum)

### Test 2: Check Vote Reception

```bash
curl -s "http://localhost:9094/api/v1/query?query=cometbft_consensus_prevotes_received" | \
  jq '.data.result[] | {validator: .metric.validator, round: .metric.round, prevotes: .value[1]}'
```

**What to expect**:
- Round 0: 3-4 prevotes
- Voting power reaching threshold

### Test 3: Monitor Round Duration

```bash
curl -s "http://localhost:9094/api/v1/query?query=histogram_quantile(0.95,rate(cometbft_consensus_round_duration_seconds_bucket[5m]))" | \
  jq -r '.data.result[].value[1]'
```

**What to expect**:
- < 10 seconds for normal operation
- Spikes indicate network issues

## Common Issues & Quick Fixes

### Issue: No metrics showing

**Fix**:
```bash
# Restart validator with new binary
docker restart aura-validator-1 aura-validator-2 aura-validator-3 aura-validator-4

# Wait 30 seconds
sleep 30

# Check again
curl -s http://localhost:27660/metrics | grep cometbft_consensus_height
```

### Issue: Prometheus not scraping

**Fix**:
```bash
# Check Prometheus targets
curl -s http://localhost:9094/targets | grep validator

# Restart Prometheus
docker restart aura-testnet-prometheus

# Verify scraping
sleep 15
curl -s "http://localhost:9094/api/v1/query?query=up{job=~'validator.*'}"
```

### Issue: Grafana dashboard empty

**Fix**:
```bash
# Verify dashboard exists
ls -la grafana/dashboards/consensus-monitoring.json

# Restart Grafana
docker restart aura-testnet-grafana

# Re-import dashboard if needed
# (Use Grafana UI: + → Import → Upload JSON)
```

### Issue: Alerts not firing

**Fix**:
```bash
# Reload Prometheus rules
curl -X POST http://localhost:9094/-/reload

# Check rule status
curl -s http://localhost:9094/api/v1/rules | jq '.data.groups[] | select(.name == "consensus-health")'
```

## Key Metrics to Watch

### Health Indicators

| Metric | Healthy Range | Alert Threshold |
|--------|---------------|-----------------|
| `cometbft_consensus_height_rate` | 5-7 blocks/min | < 0.5 |
| `cometbft_consensus_round` | 0-1 | > 5 |
| `cometbft_consensus_prevotes_received` | 3-4 | < 3 |
| `cometbft_consensus_precommits_received` | 3-4 | < 3 |
| `cometbft_consensus_prevotes_voting_power` | > 2,400,000 | < 2,400,000 |
| `cometbft_consensus_precommits_voting_power` | > 2,400,000 | < 2,400,000 |

### Performance Indicators

| Metric | Good | Warning | Critical |
|--------|------|---------|----------|
| Round duration (p95) | < 10s | 10-30s | > 30s |
| Timeouts/sec | 0 | < 0.1 | > 0.1 |
| Validator participation | > 99% | 90-99% | < 90% |

## Dashboard Navigation

### Panel 1: Overview (Top Row)
- **Current Height**: Should be incrementing every ~10s
- **Current Round**: Should usually be 0 (green)
- **Consensus Step**: Watch transition: 1→2→3→5→7
- **Block Rate**: Target ~6 blocks/min

### Panel 2: Trends (Middle)
- **Height Chart**: Should show steady increase
- **Rounds Chart**: Mostly flat at 0, spikes indicate issues

### Panel 3: Votes (Bottom)
- **Prevotes/Precommits**: Should show 3-4 validators voting
- **Voting Power**: Should exceed 2,400,000 (67% threshold)

## Next Steps

1. ✅ **Set up Alertmanager** - Get notified of issues
   - See: [CONSENSUS_MONITORING.md - Alertmanager Integration](./CONSENSUS_MONITORING.md#alertmanager-integration)

2. ✅ **Configure notifications** - Slack, email, PagerDuty
   - See: [CONSENSUS_MONITORING.md - Notifications](./CONSENSUS_MONITORING.md#slack-notifications)

3. ✅ **Baseline your network** - Understand normal behavior
   - Run for 24h, note typical ranges
   - Document baseline metrics

4. ✅ **Test failure scenarios** - Practice incident response
   - See: [MULTI_NODE_TESTING_PROCEDURES.md](./MULTI_NODE_TESTING_PROCEDURES.md)

5. ✅ **Set up log aggregation** - Correlate metrics with logs
   - Consider ELK stack or Loki

## Useful Commands

```bash
# Quick health check
curl -s http://localhost:27660/metrics | grep -E "cometbft_consensus_(height|round|step)" | grep -v "#"

# Watch height in real-time
watch -n 2 'curl -s http://localhost:27660/metrics | grep "cometbft_consensus_height{" | grep -v "#"'

# Check all validators
for port in 27660 27760 27860 27960; do
  echo "=== Port $port ==="
  curl -s http://localhost:$port/metrics | grep "cometbft_consensus_height{"
done

# Query Prometheus
curl -s "http://localhost:9094/api/v1/query?query=cometbft_consensus_height" | jq .

# Check alerts
curl -s http://localhost:9094/api/v1/alerts | jq '.data.alerts[] | select(.state == "firing")'

# Export metrics for analysis
curl -s http://localhost:27660/metrics > metrics_snapshot.txt
```

## Troubleshooting Checklist

- [ ] Validators running: `docker ps | grep validator`
- [ ] Metrics exposed: `curl http://localhost:27660/metrics | grep consensus`
- [ ] Prometheus scraping: `curl http://localhost:9094/targets`
- [ ] Prometheus storing: `curl http://localhost:9094/api/v1/query?query=up`
- [ ] Grafana running: `curl http://localhost:3002/api/health`
- [ ] Dashboard exists: `ls grafana/dashboards/consensus-monitoring.json`
- [ ] Alerts loaded: `curl http://localhost:9094/api/v1/rules`

## Getting Help

If issues persist:

1. Check container logs:
   ```bash
   docker logs aura-validator-1 --tail 100
   docker logs aura-testnet-prometheus --tail 100
   docker logs aura-testnet-grafana --tail 100
   ```

2. Verify network connectivity:
   ```bash
   docker exec aura-validator-1 ping -c 3 aura-validator-2
   ```

3. Check Prometheus configuration:
   ```bash
   docker exec aura-testnet-prometheus cat /etc/prometheus/prometheus.yml
   ```

4. Consult full documentation:
   - [CONSENSUS_MONITORING.md](./CONSENSUS_MONITORING.md)
   - [MULTI_NODE_TESTING_PROCEDURES.md](./MULTI_NODE_TESTING_PROCEDURES.md)
