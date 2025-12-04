# AURA Testnet Monitoring Runbook

**Purpose:** Comprehensive guide for monitoring the 4-validator local testnet during development and testing.

**Target Audience:** Developers, QA engineers, DevOps

**Testnet Configuration:**
- Chain ID: `aura-local-4`
- Validators: 4 nodes
- Consensus: CometBFT (requires 3/4 for consensus)
- Expected block time: 2-4 seconds

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Monitoring Tools Overview](#monitoring-tools-overview)
3. [Daily Monitoring Workflow](#daily-monitoring-workflow)
4. [Health Check Procedures](#health-check-procedures)
5. [Performance Monitoring](#performance-monitoring)
6. [Log Analysis](#log-analysis)
7. [Alerting and Error Detection](#alerting-and-error-detection)
8. [Troubleshooting Guide](#troubleshooting-guide)
9. [Incident Response](#incident-response)
10. [Best Practices](#best-practices)

---

## Quick Start

### Essential Monitoring Commands

**1. Quick Health Check (30 seconds)**
```bash
./scripts/testnet-monitor.sh quick
```
Returns: HEALTHY | DEGRADED | DOWN

**2. Live Monitoring Dashboard**
```bash
./scripts/testnet-monitor.sh watch
```
Real-time updates every 3 seconds. Press Ctrl+C to exit.

**3. Watch Block Production**
```bash
./scripts/testnet-monitor.sh watch-blocks validator-1
```
See blocks being produced in real-time.

---

## Monitoring Tools Overview

### 1. testnet-monitor.sh (Primary Tool)

**Location:** `/home/decri/blockchain-projects/aura/scripts/testnet-monitor.sh`

**Capabilities:**
- Quick health checks
- Continuous monitoring
- Performance metrics
- Log analysis
- Network health
- Diagnostics
- Automatic remediation

**Full command reference:**
```bash
./scripts/testnet-monitor.sh help
```

### 2. testnet-manage.sh (Management)

**Location:** `/home/decri/blockchain-projects/aura/scripts/testnet-manage.sh`

**Capabilities:**
- Start/stop/restart validators
- View logs
- Execute commands in containers
- BFT testing
- Port mappings

### 3. continuous-monitor.sh (Automation)

**Location:** `/home/decri/blockchain-projects/aura/scripts/continuous-monitor.sh`

**Usage:**
```bash
./scripts/continuous-monitor.sh [interval_seconds] [log_file]

# Example: Check every 60 seconds, log to monitoring.log
./scripts/continuous-monitor.sh 60 monitoring.log
```

Runs automated health checks and logs results for long-running test sessions.

### 4. Prometheus + Grafana (Visualization)

**Prometheus:** http://localhost:9094
**Grafana:** http://localhost:3002 (admin/aura-testnet-admin)

Pre-configured dashboards for:
- Validator overview
- Block production
- Network health
- Resource usage
- Alerts

---

## Daily Monitoring Workflow

### Morning Startup Checklist

**1. Start the testnet (if not already running)**
```bash
docker-compose -f docker-compose.testnet.yml up -d
```

**2. Wait for startup (30 seconds)**
```bash
sleep 30
```

**3. Quick health check**
```bash
./scripts/testnet-monitor.sh quick
```

**Expected output:**
```
✓ validator-1 - Running (Height: 1234)
✓ validator-2 - Running (Height: 1234)
✓ validator-3 - Running (Height: 1234)
✓ validator-4 - Running (Height: 1234)

Summary:
  Containers Running:      4/4
  Healthy Validators:      4/4
  Producing Blocks:        4/4

✓ Testnet is HEALTHY - Consensus operating normally
```

**4. Network health check**
```bash
./scripts/testnet-monitor.sh network
```

Verify:
- All validators have 3 peers
- Heights are synchronized (variance <5 blocks)
- RPC/API endpoints available

**5. Baseline performance metrics**
```bash
./scripts/testnet-monitor.sh performance
```

Record baseline:
- Average block time: ~3s
- Blocks per minute: ~20
- Memory usage per validator: <500MB
- CPU usage per validator: <20%

**6. Check for overnight errors**
```bash
./scripts/testnet-monitor.sh check-logs all 500
```

### During Testing

**Option 1: Continuous monitoring in separate terminal**
```bash
./scripts/testnet-monitor.sh watch
```

**Option 2: Automated monitoring with logging**
```bash
./scripts/continuous-monitor.sh 30 test-session.log &
```

**Option 3: Monitor specific validator**
```bash
./scripts/testnet-monitor.sh watch-blocks validator-1
```

### End of Day Checklist

**1. Final health check**
```bash
./scripts/testnet-monitor.sh quick
```

**2. Scan for errors**
```bash
./scripts/testnet-monitor.sh check-logs all 1000
```

**3. Review performance**
```bash
./scripts/testnet-monitor.sh performance
```

**4. Optional: Clean shutdown**
```bash
./scripts/testnet-monitor.sh shutdown
```

Or leave running:
```bash
# Testnet will continue running overnight
```

---

## Health Check Procedures

### Quick Health Check (Tier 1)

**When to run:**
- Every morning
- Before starting tests
- After deploying changes
- After suspected issues

**Command:**
```bash
./scripts/testnet-monitor.sh quick
```

**Success Criteria:**
- 3+ validators running
- 3+ validators producing blocks
- Status: HEALTHY

**If DEGRADED or DOWN:**
1. Run network health check
2. Run diagnostics
3. See troubleshooting guide

### Network Health Check (Tier 2)

**When to run:**
- DEGRADED or DOWN status
- Peer connectivity suspected
- Sync issues

**Command:**
```bash
./scripts/testnet-monitor.sh network
```

**Success Criteria:**
- Each validator has 3 peers
- Height variance <5 blocks
- All RPC/API endpoints available

**Red Flags:**
- 0 peers on any validator → Network partition
- Height variance >10 blocks → Sync issues
- RPC endpoints unavailable → Node failure

### Detailed Diagnostics (Tier 3)

**When to run:**
- Network health check fails
- Containers failing to start
- Unexplained behavior

**Command:**
```bash
./scripts/testnet-monitor.sh diagnose all
```

**Provides:**
- Docker environment info
- Container status and exit codes
- Port accessibility
- Recent error counts
- Recommendations

### Per-Validator Status (Tier 4)

**When to run:**
- Investigating specific validator
- Detailed troubleshooting

**Command:**
```bash
./scripts/testnet-monitor.sh status validator-1
```

**Provides:**
- Container status
- Chain status (height, sync, voting power)
- Network info (peers, connections)
- Resource usage
- Access endpoints

---

## Performance Monitoring

### Block Production Metrics

**Measure block production:**
```bash
./scripts/testnet-monitor.sh performance
```

**Key Metrics:**

| Metric | Normal | Warning | Critical |
|--------|--------|---------|----------|
| Average block time | 2-4s | 5-10s | >10s |
| Blocks per minute | 15-30 | 6-14 | <6 |
| Height variance | <5 | 5-10 | >10 |

**Interpretation:**

- **Block time increasing:** Resource constraints, network latency
- **Block time >10s:** Consensus issues, insufficient validators
- **Height variance high:** Sync problems, network partition

### Resource Usage Monitoring

**Check resource usage:**
```bash
./scripts/testnet-monitor.sh performance
```

Or directly:
```bash
docker stats --no-stream aura-validator-1 aura-validator-2 aura-validator-3 aura-validator-4
```

**Normal Baselines:**

| Resource | Normal | Warning | Critical |
|----------|--------|---------|----------|
| Memory per validator | <500MB | 500-1000MB | >1GB |
| CPU per validator | <20% | 20-50% | >50% |
| Disk I/O | Low | Medium | High |

**Red Flags:**

- Memory constantly increasing → Memory leak
- CPU >50% sustained → Performance bottleneck
- High disk I/O → Database issues

### Transaction Throughput (Future)

**Manual calculation:**
```bash
# Get block at height N
curl -s http://localhost:27657/block?height=1000 | jq '.result.block.data.txs | length'

# Calculate TPS over period
# TPS = (total_transactions) / (time_period_seconds)
```

---

## Log Analysis

### Automated Log Scanning

**Scan all validators for errors:**
```bash
./scripts/testnet-monitor.sh check-logs all
```

**Scan specific validator:**
```bash
./scripts/testnet-monitor.sh check-logs validator-2 200
```

**Detected patterns:**
- `ERROR`, `FATAL`, `panic`
- `failed`, `timeout`
- `connection refused`
- `consensus failure`
- `validator.*double.*sign`
- `insufficient.*voting.*power`

### Manual Log Analysis

**Tail logs in real-time:**
```bash
./scripts/testnet-monitor.sh logs validator-1
```

**Search for specific patterns:**
```bash
docker logs aura-validator-1 2>&1 | grep -i "error"
docker logs aura-validator-1 2>&1 | grep -i "consensus"
docker logs aura-validator-1 2>&1 | grep -i "peer"
```

**Export logs for analysis:**
```bash
docker logs aura-validator-1 > validator-1.log 2>&1
```

### Critical Log Patterns

**Consensus Failures:**
```
ERROR.*consensus.*failure
validator.*voting.*power
failed.*propose.*block
```

**Network Issues:**
```
connection.*refused
peer.*disconnect
dial.*timeout
```

**Byzantine Faults:**
```
double.*sign
validator.*byzantine
conflicting.*block
```

**Resource Issues:**
```
out.*of.*memory
panic.*runtime
stack.*overflow
```

---

## Alerting and Error Detection

### Prometheus Alerts

**Alert configuration:** `/docker/monitoring/prometheus/rules/aura-alerts.yml`

**Pre-configured alerts:**
- Chain halt (no blocks in 30s)
- High resource usage (>80% memory/CPU)
- Low peer count (<2 peers)
- Validator offline
- High error rate

**View active alerts:**
```bash
curl -s http://localhost:9094/api/v1/alerts | jq '.data.alerts'
```

### Manual Error Detection

**Check for critical errors:**
```bash
# Consensus failures
docker-compose -f docker-compose.testnet.yml logs --tail=100 | grep -i "consensus failure"

# Double signing
docker-compose -f docker-compose.testnet.yml logs --tail=100 | grep -i "double.*sign"

# Panics
docker-compose -f docker-compose.testnet.yml logs --tail=100 | grep -i "panic"
```

### Automated Monitoring

**Run continuous monitoring:**
```bash
./scripts/continuous-monitor.sh 60 monitoring.log
```

**Monitors:**
- Health status every 60s
- Critical error detection
- Periodic detailed checks
- Logs all results

**View live stats:**
```
Checks: 10 | Healthy: 9 | Degraded: 1 | Down: 0 | Errors: 0 | Last: 14:23:45
```

---

## Troubleshooting Guide

### Chain Not Producing Blocks

**Symptoms:**
- Block height not increasing
- Status: DOWN

**Diagnosis:**
```bash
./scripts/testnet-monitor.sh quick
```

**Common Causes:**

1. **Insufficient validators (<3 running)**
   ```bash
   # Check running count
   docker ps | grep aura-validator | wc -l

   # Restart failed validators
   ./scripts/testnet-monitor.sh restart-failed
   ```

2. **Network partition**
   ```bash
   # Check peer connectivity
   ./scripts/testnet-monitor.sh network

   # Verify Docker network
   docker network inspect aura-testnet
   ```

3. **Consensus failure**
   ```bash
   # Check logs
   ./scripts/testnet-monitor.sh check-logs all

   # Look for specific errors
   docker-compose -f docker-compose.testnet.yml logs | grep -i "consensus"
   ```

**Resolution:**

```bash
# 1. Quick check
./scripts/testnet-monitor.sh quick

# 2. Restart failed validators
./scripts/testnet-monitor.sh restart-failed

# 3. If still failing, clean restart
./scripts/testnet-monitor.sh shutdown
docker-compose -f docker-compose.testnet.yml up -d

# 4. Verify
./scripts/testnet-monitor.sh watch-blocks validator-1
```

### Validator Not Syncing

**Symptoms:**
- Height significantly behind others (>10 blocks)
- "Catching up" status

**Diagnosis:**
```bash
./scripts/testnet-monitor.sh status validator-X
```

**Common Causes:**

1. **Recently started validator**
   - Normal: Will catch up within minutes
   - Monitor: `./scripts/testnet-monitor.sh watch-blocks validator-X`

2. **Resource constraints**
   ```bash
   docker stats aura-validator-X
   ```
   - High CPU/memory usage → Resource bottleneck

3. **Network latency**
   ```bash
   ./scripts/testnet-monitor.sh network
   ```
   - Low peer count → Connectivity issues

**Resolution:**

```bash
# 1. Check if catching up
curl -s http://localhost:27X57/status | jq '.result.sync_info.catching_up'

# 2. Monitor progress
./scripts/testnet-monitor.sh watch-blocks validator-X

# 3. If stuck, restart validator
docker-compose -f docker-compose.testnet.yml restart validator-X

# 4. Verify sync
./scripts/testnet-monitor.sh status validator-X
```

### High Error Count in Logs

**Symptoms:**
- `check-logs` reports many errors
- Logs show ERROR/FATAL messages

**Diagnosis:**
```bash
./scripts/testnet-monitor.sh check-logs all
```

**Categorize errors:**

**1. Connection errors** (usually benign during startup)
```
connection refused
dial timeout
peer disconnect
```

**2. Consensus errors** (critical)
```
consensus failure
insufficient voting power
failed to propose block
```

**3. Application errors** (investigate)
```
failed to execute
invalid transaction
module error
```

**Resolution:**

```bash
# 1. Identify error pattern
./scripts/testnet-monitor.sh logs validator-1 | grep ERROR

# 2. For connection errors during startup: wait and re-check
sleep 30
./scripts/testnet-monitor.sh check-logs validator-1

# 3. For persistent errors: investigate specific issue
./scripts/testnet-monitor.sh diagnose validator-1

# 4. For consensus errors: ensure sufficient validators
./scripts/testnet-monitor.sh quick
```

### Container Failing to Start

**Symptoms:**
- Container status: exited
- `docker ps` doesn't show container

**Diagnosis:**
```bash
./scripts/testnet-monitor.sh diagnose validator-X
```

**Common Causes:**

1. **Genesis file missing**
   ```bash
   docker-compose -f docker-compose.testnet.yml run --rm validator-X \
       ls -la /home/aura/.aura/config/genesis.json
   ```

2. **Port conflict**
   ```bash
   netstat -tuln | grep 27657  # Check if port in use
   ```

3. **Corrupted database**
   ```bash
   docker logs aura-validator-X 2>&1 | grep -i "database\|corruption"
   ```

**Resolution:**

```bash
# 1. Check exit code
docker inspect --format='{{.State.ExitCode}}' aura-validator-X

# 2. View startup logs
docker logs aura-validator-X

# 3. If genesis missing, reinitialize
./scripts/testnet-manage.sh clean
./scripts/testnet-init.sh
cd testnet-data && ./populate-volumes.sh && cd ..

# 4. Start again
docker-compose -f docker-compose.testnet.yml up -d

# 5. Verify
./scripts/testnet-monitor.sh quick
```

### Network Partition

**Symptoms:**
- Validators have 0 peers
- Heights diverging significantly

**Diagnosis:**
```bash
./scripts/testnet-monitor.sh network
```

**Common Causes:**

1. **Docker network issue**
   ```bash
   docker network inspect aura-testnet
   ```

2. **Firewall blocking connections**
   ```bash
   # Check if ports accessible
   nc -z localhost 27656
   nc -z localhost 27756
   ```

3. **Persistent peers misconfigured**
   ```bash
   docker exec aura-validator-1 cat /home/aura/.aura/config/config.toml | grep persistent_peers
   ```

**Resolution:**

```bash
# 1. Restart Docker network
docker-compose -f docker-compose.testnet.yml down
docker network rm aura_aura-testnet || true
docker-compose -f docker-compose.testnet.yml up -d

# 2. Verify network
docker network inspect aura_aura-testnet

# 3. Check connectivity
./scripts/testnet-monitor.sh network

# 4. If still failing, full reset
./scripts/testnet-manage.sh clean
./scripts/testnet-init.sh
# ... (reinitialize)
```

---

## Incident Response

### Severity Levels

**P0 - Critical (Immediate Response)**
- All validators down
- Chain halted (no blocks in 60s)
- Double signing detected
- Data corruption

**P1 - High (Respond within 15 minutes)**
- Consensus degraded (<3 validators)
- Validator stuck/not syncing
- High error rate (>100 errors/min)

**P2 - Medium (Respond within 1 hour)**
- Single validator down
- Performance degraded (block time >10s)
- Resource warnings

**P3 - Low (Monitor)**
- Temporary connection issues
- Minor log warnings

### Response Procedures

#### P0 Incident: Chain Halted

**1. Immediate Assessment (0-2 minutes)**
```bash
# Quick health check
./scripts/testnet-monitor.sh quick

# Check all validators
docker ps | grep aura-validator

# Check logs for critical errors
./scripts/testnet-monitor.sh check-logs all 50
```

**2. Triage (2-5 minutes)**
```bash
# Identify root cause
./scripts/testnet-monitor.sh diagnose all

# Check for double signing
docker-compose -f docker-compose.testnet.yml logs | grep -i "double.*sign"

# Check for consensus failures
docker-compose -f docker-compose.testnet.yml logs | grep -i "consensus.*fail"
```

**3. Remediation (5-10 minutes)**

**If insufficient validators:**
```bash
./scripts/testnet-monitor.sh restart-failed
```

**If consensus failure:**
```bash
# Clean restart
./scripts/testnet-monitor.sh shutdown
docker-compose -f docker-compose.testnet.yml up -d
sleep 30
./scripts/testnet-monitor.sh quick
```

**If double signing:**
```bash
# STOP offending validator immediately
docker-compose -f docker-compose.testnet.yml stop validator-X

# Investigate root cause
./scripts/testnet-monitor.sh logs validator-X

# DO NOT restart until root cause identified
```

**4. Verification (10-15 minutes)**
```bash
# Verify block production
./scripts/testnet-monitor.sh watch-blocks validator-1

# Verify network health
./scripts/testnet-monitor.sh network

# Verify performance
./scripts/testnet-monitor.sh performance
```

**5. Post-Incident (15+ minutes)**
```bash
# Export logs for analysis
docker logs aura-validator-1 > incident-validator-1.log 2>&1
docker logs aura-validator-2 > incident-validator-2.log 2>&1
docker logs aura-validator-3 > incident-validator-3.log 2>&1
docker logs aura-validator-4 > incident-validator-4.log 2>&1

# Run full diagnostic
./scripts/testnet-monitor.sh diagnose all > incident-diagnostic.txt

# Document incident (what/when/why/how)
```

---

## Best Practices

### Pre-Testing Checklist

**Before running any tests:**

1. ✅ Quick health check: `./scripts/testnet-monitor.sh quick`
2. ✅ Network health: `./scripts/testnet-monitor.sh network`
3. ✅ Baseline performance: `./scripts/testnet-monitor.sh performance`
4. ✅ Clean logs: `./scripts/testnet-monitor.sh check-logs all`
5. ✅ Start monitoring: `./scripts/testnet-monitor.sh watch` (separate terminal)

### Continuous Monitoring

**During long-running tests:**

```bash
# Terminal 1: Run your tests
# ...

# Terminal 2: Continuous monitoring
./scripts/continuous-monitor.sh 60 test-session.log

# Terminal 3: Watch blocks (optional)
./scripts/testnet-monitor.sh watch-blocks validator-1
```

### Regular Maintenance

**Daily:**
- Morning health check
- End of day log scan

**Weekly:**
- Full diagnostic: `./scripts/testnet-monitor.sh diagnose all`
- Performance analysis: `./scripts/testnet-monitor.sh performance`
- Clean restart (if needed)

**Monthly:**
- Review Grafana dashboards
- Analyze trends
- Update alert thresholds

### Documentation

**Always document:**
- Incidents and resolutions
- Configuration changes
- Performance baselines
- Known issues

**Log format:**
```
[2025-12-03 14:23:45] Incident: Chain halted
Root Cause: Validator-3 crashed due to OOM
Resolution: Increased Docker memory limit, restarted validator-3
Duration: 5 minutes
```

---

## Quick Reference

### One-Line Health Check
```bash
./scripts/testnet-monitor.sh quick && echo "✓ HEALTHY" || echo "✗ ISSUE DETECTED"
```

### Watch Block Production
```bash
while true; do curl -s http://localhost:27657/status | jq -r '.result.sync_info.latest_block_height'; sleep 3; done
```

### Count Running Validators
```bash
docker ps | grep aura-validator | wc -l
```

### Emergency Stop All
```bash
docker-compose -f docker-compose.testnet.yml down
```

### Emergency Restart All
```bash
docker-compose -f docker-compose.testnet.yml restart
```

### Full Reset
```bash
./scripts/testnet-manage.sh clean && \
./scripts/testnet-init.sh && \
cd testnet-data && ./populate-volumes.sh && cd .. && \
docker-compose -f docker-compose.testnet.yml up -d && \
sleep 30 && \
./scripts/testnet-monitor.sh quick
```

---

## Additional Resources

**Documentation:**
- Testnet Setup: `/TESTNET_SETUP.md`
- Quick Start: `/TESTNET_QUICKSTART.md`
- Monitoring Guide: `/TESTNET_MONITORING_GUIDE.md`
- Cheat Sheet: `/MONITORING_CHEATSHEET.md`
- Docker Runbook: `/docs/runbooks/LOCAL_TESTNET_DOCKER.md`

**Scripts:**
- Monitor: `/scripts/testnet-monitor.sh`
- Manage: `/scripts/testnet-manage.sh`
- Continuous Monitor: `/scripts/continuous-monitor.sh`

**Monitoring URLs:**
- Prometheus: http://localhost:9094
- Grafana: http://localhost:3002
- Validator 1 RPC: http://localhost:27657
- Validator 2 RPC: http://localhost:27757
- Validator 3 RPC: http://localhost:27857
- Validator 4 RPC: http://localhost:27957

---

**Last Updated:** 2025-12-03
**Testnet Version:** aura-local-4
**CometBFT Consensus:** 3/4 validators required
