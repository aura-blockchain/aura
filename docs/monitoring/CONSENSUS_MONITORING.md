# CometBFT Consensus Monitoring

Complete guide to monitoring consensus health and performance in the Aura blockchain.

## Overview

The Aura blockchain now includes comprehensive CometBFT consensus monitoring that provides real-time visibility into:

- **Consensus rounds and steps** - Track consensus progress and detect stalls
- **Vote participation** - Monitor prevotes and precommits from validators
- **Voting power** - Track whether consensus has sufficient voting power to progress
- **Block production rate** - Monitor blockchain throughput
- **Validator participation** - Track individual validator performance
- **Consensus timeouts** - Detect network or performance issues

## Architecture

### Metrics Collection

Consensus metrics are collected via a custom collector that polls the CometBFT RPC `/consensus_state` endpoint:

```
┌─────────────┐
│  CometBFT   │
│  RPC :26657 │
└──────┬──────┘
       │
       │ /consensus_state (5s poll)
       ▼
┌─────────────────────────┐
│ ConsensusMetricsCollector│
│  (Go routine)           │
└──────┬──────────────────┘
       │
       │ Prometheus metrics
       ▼
┌─────────────┐
│ Prometheus  │
│   :26660    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Grafana   │
│    :3002    │
└─────────────┘
```

### Metrics Exposed

#### Height Tracking
- `cometbft_consensus_height` - Current consensus height
- `cometbft_consensus_height_rate` - Block production rate (blocks/minute)

#### Round and Step
- `cometbft_consensus_round` - Current consensus round
- `cometbft_consensus_step` - Current consensus step (0-7)
  - 0 = NewHeight
  - 1 = NewRound
  - 2 = Propose
  - 3 = Prevote
  - 4 = PrevoteWait
  - 5 = Precommit
  - 6 = PrecommitWait
  - 7 = Commit

#### Vote Metrics
- `cometbft_consensus_prevotes_received` - Number of prevotes received (by round)
- `cometbft_consensus_precommits_received` - Number of precommits received (by round)
- `cometbft_consensus_prevotes_voting_power` - Total voting power of prevotes
- `cometbft_consensus_precommits_voting_power` - Total voting power of precommits

#### Proposer Tracking
- `cometbft_consensus_is_proposer` - Whether this validator is proposer (1=yes, 0=no)
- `cometbft_consensus_proposer_index` - Index of current proposer

#### Block State
- `cometbft_consensus_has_proposal` - Block proposal received (1=yes, 0=no)
- `cometbft_consensus_has_locked_block` - Block locked (1=yes, 0=no)
- `cometbft_consensus_has_valid_block` - Valid block available (1=yes, 0=no)

#### Performance
- `cometbft_consensus_round_duration_seconds` - Histogram of round durations
- `cometbft_consensus_timeouts_total` - Counter of consensus timeouts

#### Validator Participation
- `cometbft_consensus_validator_participation` - Validator participation rate (0-1)
- `cometbft_consensus_validator_missed_blocks_total` - Counter of missed blocks

## Grafana Dashboard

### Access

Dashboard: **CometBFT Consensus Monitoring**
URL: `http://localhost:3002/d/consensus-monitoring`

Default credentials:
- Username: `admin`
- Password: `admin`

### Dashboard Panels

#### Row 1: Overview Stats
- **Current Height** - Latest block height across all validators
- **Current Round** - Current consensus round (🟢 = 0, 🟡 = 1-2, 🔴 = 3+)
- **Consensus Step** - Current step in consensus protocol
- **Block Production Rate** - Blocks per minute (expected: ~6 for 10s block time)

#### Row 2: Height and Round Trends
- **Consensus Height Over Time** - Line chart showing height progression
- **Consensus Rounds** - Track rounds over time (spikes indicate issues)

#### Row 3: Vote Reception
- **Prevotes Received by Round** - Number of prevotes received
- **Precommits Received by Round** - Number of precommits received

#### Row 4: Voting Power
- **Prevote Voting Power** - Total voting power (threshold: 2.4M = 67%)
- **Precommit Voting Power** - Total voting power (threshold: 2.4M = 67%)

#### Row 5: Performance
- **Consensus Round Duration** - p50, p95, p99 percentiles
- **Validator Participation Rate** - Percentage of blocks signed

#### Row 6: Operational Status
- **Consensus Timeouts Rate** - Rate of timeouts by type
- **Has Block Proposal** - Whether proposal received
- **Has Locked Block** - Whether validator locked on a block
- **Has Valid Block** - Whether valid block is available

## Alerting Rules

### Critical Alerts

#### ConsensusStuck
**Severity**: Critical
**Condition**: No height increase in 5 minutes
**Action**:
1. Check validator logs: `docker logs aura-validator-1 -f`
2. Check network connectivity between validators
3. Verify validators are running: `docker ps | grep validator`
4. Check for resource exhaustion: `docker stats`

#### ConsensusRoundStuck
**Severity**: Critical
**Condition**: Same round for 5 minutes
**Action**:
1. Check if validators are receiving votes
2. Verify network connectivity
3. Check for Byzantine behavior in logs
4. Consider restarting stuck validator

#### LowPrecommitParticipation
**Severity**: Critical
**Condition**: < 67% precommit voting power for 2 minutes
**Action**:
1. Identify which validators are not precommitting
2. Check their logs for errors
3. Verify network connectivity
4. Check if validators are jailed

#### MissingPrecommitsFromMajority
**Severity**: Critical
**Condition**: < 3 precommits received when in PrecommitWait step
**Action**:
1. Check validator connectivity
2. Verify all validators are at same height
3. Look for clock skew issues
4. Check for network partitions

### Warning Alerts

#### HighConsensusRounds
**Severity**: Warning
**Condition**: Round > 5 for 3 minutes
**Action**:
1. Check for network latency issues
2. Verify validators are responding
3. Review voting patterns
4. Consider network optimization

#### LowPrevoteParticipation
**Severity**: Warning
**Condition**: < 67% prevote voting power for 2 minutes
**Action**:
1. Identify validators not sending prevotes
2. Check for proposal issues
3. Verify network health

#### NoBlockProposal
**Severity**: Warning
**Condition**: No proposal received in Propose step
**Action**:
1. Check if proposer is online
2. Verify network connectivity to proposer
3. Check proposer's logs

#### LowBlockProductionRate
**Severity**: Warning
**Condition**: < 0.5 blocks/minute for 5 minutes
**Action**:
1. Check consensus rounds
2. Verify network health
3. Check for performance issues

#### FrequentConsensusTimeouts
**Severity**: Warning
**Condition**: > 0.1 timeouts/second for 3 minutes
**Action**:
1. Check network latency
2. Verify validator performance
3. Review timeout types

#### LongConsensusRounds
**Severity**: Warning
**Condition**: p95 round duration > 30s for 5 minutes
**Action**:
1. Check network latency
2. Verify validator resources
3. Review voting patterns

#### ValidatorParticipationDrop
**Severity**: Warning
**Condition**: Participation < 90% for 5 minutes
**Action**:
1. Check validator uptime
2. Verify network connectivity
3. Review validator logs

## Troubleshooting

### Consensus Not Progressing

**Symptoms**: Height not increasing, stuck in same round

**Diagnosis**:
```bash
# Check current consensus state
curl http://localhost:27657/consensus_state | jq .

# Check validator connectivity
docker exec aura-validator-1 aurad status | jq .sync_info

# Check network connectivity
docker exec aura-validator-1 aurad query tendermint-validator-set
```

**Solutions**:
1. Verify 2/3+ validators are online and connected
2. Check for network partitions
3. Restart lagging validators
4. Verify clock synchronization

### High Round Numbers

**Symptoms**: Round > 5, frequent round increases

**Diagnosis**:
```bash
# Check vote reception
curl http://localhost:9094/api/v1/query?query=cometbft_consensus_prevotes_received

# Check voting power
curl http://localhost:9094/api/v1/query?query=cometbft_consensus_prevotes_voting_power

# Check for timeouts
curl http://localhost:9094/api/v1/query?query=rate(cometbft_consensus_timeouts_total[5m])
```

**Solutions**:
1. Check network latency between validators
2. Verify validators have adequate resources
3. Check for Byzantine behavior
4. Review timeout configurations

### Low Voting Power

**Symptoms**: < 67% voting power for prevotes/precommits

**Diagnosis**:
```bash
# Check validator set
docker exec aura-validator-1 aurad query staking validators --output json | jq '.validators[] | {moniker, status, tokens}'

# Check jailed validators
docker exec aura-validator-1 aurad query staking validators --output json | jq '.validators[] | select(.jailed==true)'

# Check missed blocks
curl http://localhost:9094/api/v1/query?query=cometbft_consensus_validator_missed_blocks_total
```

**Solutions**:
1. Identify offline or jailed validators
2. Unjail validators if needed
3. Verify network connectivity
4. Add more validators to increase redundancy

### Slow Block Production

**Symptoms**: Block production rate < expected

**Diagnosis**:
```bash
# Check current rate
curl http://localhost:9094/api/v1/query?query=cometbft_consensus_height_rate

# Check round duration
curl http://localhost:9094/api/v1/query?query=histogram_quantile(0.95,rate(cometbft_consensus_round_duration_seconds_bucket[5m]))

# Check timeouts
curl http://localhost:9094/api/v1/query?query=cometbft_consensus_timeouts_total
```

**Solutions**:
1. Optimize network connectivity
2. Increase validator resources
3. Tune timeout configurations
4. Check for performance bottlenecks

## Configuration

### Metrics Collection Interval

The consensus metrics collector polls every **5 seconds** by default. To adjust:

Edit `chain/cmd/aurad/cmd/start.go`:
```go
go consensusCollector.Start(consensusCtx, 5*time.Second)
//                                         ↑ Change this
```

### Prometheus Scrape Interval

Default: **15 seconds**

Edit `prometheus/prometheus-testnet.yml`:
```yaml
global:
  scrape_interval: 15s  # Change this
```

### Alert Evaluation

Default: **15 seconds**

Edit `prometheus/prometheus-testnet.yml`:
```yaml
global:
  evaluation_interval: 15s  # Change this
```

## Metrics Retention

### Prometheus
- Default retention: **7 days**
- Location: Docker volume `prometheus-testnet-data`
- Configuration: `docker-compose.testnet.yml`

```yaml
command:
  - '--storage.tsdb.retention.time=7d'  # Change here
```

### Grafana
- Dashboards: `/home/hudson/blockchain-projects/aura/grafana/dashboards/`
- Data: Docker volume `grafana-testnet-data`

## Advanced Queries

### Average Block Time
```promql
60 / rate(cometbft_consensus_height[5m])
```

### Consensus Efficiency (% time in round 0)
```promql
(cometbft_consensus_round == 0) * 100
```

### Validator Uptime (last hour)
```promql
(1 - (rate(cometbft_consensus_validator_missed_blocks_total[1h]) / rate(cometbft_consensus_height[1h]))) * 100
```

### Network Consensus Health Score
```promql
min(
  (cometbft_consensus_height_rate > 5) * 0.3 +
  (cometbft_consensus_round < 2) * 0.3 +
  (cometbft_consensus_validator_participation > 0.9) * 0.4
)
```

## Integration with Monitoring Stack

### Alertmanager Integration

Edit `docker-compose.testnet.yml` to add Alertmanager:

```yaml
alertmanager:
  image: prom/alertmanager:latest
  ports:
    - "9093:9093"
  volumes:
    - ./alertmanager/config.yml:/etc/alertmanager/config.yml
```

### Slack Notifications

Configure `alertmanager/config.yml`:

```yaml
receivers:
  - name: 'slack'
    slack_configs:
      - api_url: 'YOUR_SLACK_WEBHOOK_URL'
        channel: '#aura-alerts'
        title: 'Aura Consensus Alert'
```

### Email Notifications

```yaml
receivers:
  - name: 'email'
    email_configs:
      - to: 'ops@example.com'
        from: 'alertmanager@aura.network'
        smarthost: 'smtp.gmail.com:587'
        auth_username: 'alertmanager@aura.network'
        auth_password: 'YOUR_PASSWORD'
```

## Best Practices

1. **Monitor Continuously**: Keep Grafana dashboard open during operations
2. **Set Up Alerts**: Configure Alertmanager for critical alerts
3. **Regular Reviews**: Review consensus metrics weekly
4. **Baseline Performance**: Establish normal ranges for your network
5. **Document Incidents**: Track consensus issues and resolutions
6. **Test Failure Scenarios**: Regularly test validator failures
7. **Optimize Network**: Use consensus metrics to optimize network topology
8. **Capacity Planning**: Use metrics for validator capacity planning

## References

- [CometBFT Consensus Spec](https://github.com/cometbft/cometbft/blob/main/spec/consensus/consensus.md)
- [Prometheus Query Documentation](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Grafana Dashboard Best Practices](https://grafana.com/docs/grafana/latest/best-practices/)
- [Aura Multi-Node Testing](./MULTI_NODE_TESTING_PROCEDURES.md)

## Support

For issues or questions:
1. Check validator logs: `docker logs aura-validator-1`
2. Review Prometheus targets: `http://localhost:9094/targets`
3. Check alert rules: `http://localhost:9094/rules`
4. Verify metrics: `curl http://localhost:27660/metrics | grep consensus`
