# AURA Testnet Monitoring Guide

Quick reference for monitoring and troubleshooting the 4-validator local testnet.

---

## Quick Start - Essential Commands

### 1. Quick Health Check (Single Command)
```bash
./scripts/testnet-monitor.sh quick
```
**Returns:** Overall health status of all 4 validators
- ✓ HEALTHY: 3+ validators producing blocks
- ⚠ DEGRADED: 1-2 validators producing blocks
- ✗ DOWN: No validators producing blocks

### 2. Live Monitoring Dashboard
```bash
./scripts/testnet-monitor.sh watch
```
**Updates every 3 seconds with:**
- Block height progression
- Peer connections
- Memory/CPU usage
- Sync status
- Network variance

**Press Ctrl+C to exit**

### 3. Watch Block Production
```bash
./scripts/testnet-monitor.sh watch-blocks validator-1
```
**Real-time block feed:**
```
▲ Block 1234 at 2025-12-03T10:15:23Z (avg: 3.2s/block)
▲ Block 1235 at 2025-12-03T10:15:26Z (avg: 3.1s/block)
```

---

## Individual Validator Commands

### Check Single Validator Status
```bash
./scripts/testnet-monitor.sh status validator-1
./scripts/testnet-monitor.sh status validator-2
./scripts/testnet-monitor.sh status validator-3
./scripts/testnet-monitor.sh status validator-4
```

**Returns:**
- Container status (running/stopped)
- Chain status (height, sync, voting power)
- Network info (peers, listening)
- Resource usage (memory, CPU)
- Access endpoints (RPC, API, gRPC)

### Tail Logs for a Validator
```bash
./scripts/testnet-monitor.sh logs validator-1
./scripts/testnet-monitor.sh logs validator-2
# etc...
```

**Press Ctrl+C to exit**

---

## Error Detection and Log Analysis

### Scan All Validators for Errors
```bash
./scripts/testnet-monitor.sh check-logs all
```

### Scan Specific Validator
```bash
./scripts/testnet-monitor.sh check-logs validator-1
```

### Scan Last N Lines
```bash
./scripts/testnet-monitor.sh check-logs validator-2 200
```

**Detects Common Patterns:**
- `ERROR`, `FATAL`, `panic`
- `failed`, `timeout`
- `connection refused`
- `consensus failure`
- `validator.*double.*sign`
- `insufficient.*voting.*power`

---

## Performance Metrics

### Measure Block Production Performance
```bash
./scripts/testnet-monitor.sh performance
```

**Samples 30 seconds of block production and reports:**
- Average block time
- Blocks per minute
- Resource usage across all validators

**Example Output:**
```
Start Height:              1234
End Height:                1244
Blocks Produced:           10
Time Period:               30s
Average Block Time:        3.0s
Blocks per Minute:         20.0

Resource Usage:
  validator-1    Memory: 256MB    CPU: 12%
  validator-2    Memory: 248MB    CPU: 11%
  validator-3    Memory: 252MB    CPU: 13%
  validator-4    Memory: 250MB    CPU: 12%
```

---

## Network Health Checks

### Check Peer Connectivity and Consensus
```bash
./scripts/testnet-monitor.sh network
```

**Checks:**
- Peer connectivity (each validator should have 3 peers)
- Consensus status (height synchronization)
- RPC endpoint availability
- API endpoint availability

**Example Output:**
```
Peer Connectivity:
✓ validator-1: 3/3 peers connected
✓ validator-2: 3/3 peers connected
✓ validator-3: 3/3 peers connected
⚠ validator-4: 2/3 peers connected

Consensus Status:
✓ validator-1: Height 1234 (in sync)
✓ validator-2: Height 1234 (in sync)
✓ validator-3: Height 1234 (in sync)
✗ validator-4: Height 1220 (14 blocks behind)

RPC Endpoint Availability:
✓ validator-1 RPC (27657): Available
✓ validator-2 RPC (27757): Available
✓ validator-3 RPC (27857): Available
✗ validator-4 RPC (27957): Unavailable
```

---

## Troubleshooting Commands

### Full Diagnostic Report
```bash
# All validators
./scripts/testnet-monitor.sh diagnose all

# Specific validator
./scripts/testnet-monitor.sh diagnose validator-2
```

**Checks:**
- Docker environment
- Container status and exit codes
- Port accessibility
- Recent error counts
- Provides recommendations

### Restart Failed Validators
```bash
./scripts/testnet-monitor.sh restart-failed
```
**Automatically:**
1. Identifies stopped validators
2. Restarts them via Docker Compose
3. Waits 10 seconds
4. Runs health check

---

## Clean Shutdown Procedure

### Graceful Shutdown
```bash
./scripts/testnet-monitor.sh shutdown
```

**Steps:**
1. Stops all validators gracefully
2. Waits for clean exit
3. Stops monitoring services (Prometheus, Grafana)
4. Removes containers (keeps data volumes)

**Data is preserved.** To restart:
```bash
docker-compose -f docker-compose.testnet.yml up -d
```

**To wipe all data:**
```bash
./scripts/testnet-manage.sh clean
```

---

## Common Error Patterns and Solutions

### Error: "consensus failure"
**Cause:** Validator not participating in consensus

**Solutions:**
1. Check if validator is running: `./scripts/testnet-monitor.sh quick`
2. Verify peer connectivity: `./scripts/testnet-monitor.sh network`
3. Ensure 3/4 validators are online
4. Check voting power: `./scripts/testnet-monitor.sh status validator-X`

### Error: "connection refused"
**Cause:** Peer connectivity issues or port conflicts

**Solutions:**
1. Check if containers are running: `docker ps`
2. Verify port mappings: `./scripts/testnet-manage.sh ports`
3. Check Docker network: `docker network inspect aura-testnet`
4. Review firewall rules (if applicable)

### Error: "insufficient voting power"
**Cause:** Not enough validators online (need 3/4)

**Solutions:**
1. Count running validators: `./scripts/testnet-monitor.sh quick`
2. Start missing validators: `./scripts/testnet-monitor.sh restart-failed`
3. Verify genesis configuration has correct voting power distribution

### Error: "double sign" / "validator.*double.*sign"
**Cause:** Validator signing multiple blocks at same height (Byzantine fault)

**Solutions:**
1. **CRITICAL:** Stop the offending validator immediately
2. Check validator key uniqueness
3. Ensure only one instance per validator is running
4. Review validator configuration

### Error: "timeout"
**Cause:** Network latency, slow queries, or resource constraints

**Solutions:**
1. Check resource usage: `./scripts/testnet-monitor.sh performance`
2. Review Docker resource limits
3. Check peer connectivity
4. Monitor block production rate

### Blocks Stop Producing
**Cause:** Consensus stalled or insufficient validators

**Troubleshooting Steps:**
```bash
# 1. Check overall health
./scripts/testnet-monitor.sh quick

# 2. Watch blocks (should see new blocks every ~3s)
./scripts/testnet-monitor.sh watch-blocks validator-1

# 3. Check peer connectivity
./scripts/testnet-monitor.sh network

# 4. Verify voting power
./scripts/testnet-monitor.sh status validator-1
./scripts/testnet-monitor.sh status validator-2
./scripts/testnet-monitor.sh status validator-3
./scripts/testnet-monitor.sh status validator-4

# 5. Restart failed validators
./scripts/testnet-monitor.sh restart-failed
```

### Container Fails to Start
**Troubleshooting Steps:**
```bash
# 1. Run diagnostics
./scripts/testnet-monitor.sh diagnose validator-X

# 2. Check logs for startup errors
./scripts/testnet-monitor.sh logs validator-X

# 3. Verify genesis file exists
docker exec -it aura-validator-X ls -la /home/aura/.aura/config/genesis.json

# 4. Check config.toml
docker exec -it aura-validator-X cat /home/aura/.aura/config/config.toml

# 5. Reinitialize if needed
./scripts/testnet-manage.sh clean
./scripts/testnet-init.sh
cd testnet-data && ./populate-volumes.sh && cd ..
docker-compose -f docker-compose.testnet.yml up -d
```

---

## Validator Endpoints

### Validator 1
- **RPC:** http://localhost:27657
- **API:** http://localhost:2317
- **gRPC:** localhost:10090
- **Metrics:** http://localhost:27660

### Validator 2
- **RPC:** http://localhost:27757
- **API:** http://localhost:2417
- **gRPC:** localhost:10190
- **Metrics:** http://localhost:27760

### Validator 3
- **RPC:** http://localhost:27857
- **API:** http://localhost:2517
- **gRPC:** localhost:10290
- **Metrics:** http://localhost:27860

### Validator 4
- **RPC:** http://localhost:27957
- **API:** http://localhost:2617
- **gRPC:** localhost:10390
- **Metrics:** http://localhost:27960

### Monitoring Services
- **Prometheus:** http://localhost:9094
- **Grafana:** http://localhost:3002 (admin/aura-testnet-admin)

---

## Manual Health Check Commands

### Check Block Height via RPC
```bash
curl -s http://localhost:27657/status | jq '.result.sync_info.latest_block_height'
```

### Check Peer Count
```bash
curl -s http://localhost:27657/net_info | jq '.result.n_peers'
```

### Check if Catching Up
```bash
curl -s http://localhost:27657/status | jq '.result.sync_info.catching_up'
```

### Check Voting Power
```bash
curl -s http://localhost:27657/status | jq '.result.validator_info.voting_power'
```

### Check Container Health
```bash
docker inspect --format='{{.State.Health.Status}}' aura-validator-1
```

### Check Container Resource Usage
```bash
docker stats --no-stream aura-validator-1
```

### Get Container Logs (Last 100 Lines)
```bash
docker logs --tail 100 aura-validator-1
```

---

## Direct Docker Commands

### Start All Validators
```bash
docker-compose -f docker-compose.testnet.yml up -d
```

### Stop All Validators
```bash
docker-compose -f docker-compose.testnet.yml down
```

### Restart Specific Validator
```bash
docker-compose -f docker-compose.testnet.yml restart validator-2
```

### View Real-Time Logs
```bash
docker-compose -f docker-compose.testnet.yml logs -f validator-1
```

### Execute Command in Container
```bash
docker exec -it aura-validator-1 aurad status
docker exec -it aura-validator-1 aurad query bank balances <address>
```

---

## Monitoring Best Practices

### Pre-Flight Checks (Before Starting Tests)
```bash
# 1. Quick health check
./scripts/testnet-monitor.sh quick

# 2. Verify all validators are in sync
./scripts/testnet-monitor.sh network

# 3. Check performance baseline
./scripts/testnet-monitor.sh performance

# 4. Scan for recent errors
./scripts/testnet-monitor.sh check-logs all
```

### During Operation
```bash
# Run continuous monitoring in a separate terminal
./scripts/testnet-monitor.sh watch

# Or watch block production
./scripts/testnet-monitor.sh watch-blocks validator-1
```

### Post-Test Analysis
```bash
# Check for errors during test
./scripts/testnet-monitor.sh check-logs all

# Review performance metrics
./scripts/testnet-monitor.sh performance

# Verify network health
./scripts/testnet-monitor.sh network
```

### Regular Maintenance
```bash
# Daily: Quick health check
./scripts/testnet-monitor.sh quick

# Weekly: Full diagnostic
./scripts/testnet-monitor.sh diagnose all

# As needed: Performance analysis
./scripts/testnet-monitor.sh performance
```

---

## Alerting Thresholds

### Critical (Immediate Action Required)
- **0 validators producing blocks** → Chain halted
- **Double sign detected** → Byzantine fault
- **All RPC endpoints down** → Network failure

### Warning (Investigate Soon)
- **1-2 validators producing blocks** → Degraded consensus
- **Height variance >10 blocks** → Sync issues
- **Peer count <2 for any validator** → Connectivity issues
- **Block time >10 seconds** → Performance degradation

### Info (Monitor)
- **3 validators producing blocks** → Normal (4th can be down)
- **Height variance <5 blocks** → Normal sync variance
- **Block time 2-4 seconds** → Normal performance

---

## Integration with Grafana

The testnet includes pre-configured Grafana dashboards at http://localhost:3002

**Dashboards:**
- **Validator Overview:** All 4 validators on one screen
- **Block Production:** Block height, block time, missed blocks
- **Network Health:** Peer connections, consensus participation
- **Resource Usage:** CPU, memory, disk for each validator
- **Alerts:** Pre-configured alerts for critical conditions

**Login:** admin / aura-testnet-admin

**To customize alerts:**
1. Edit `/docker/monitoring/prometheus/rules/aura-alerts.yml`
2. Reload Prometheus: `docker exec aura-testnet-prometheus kill -HUP 1`

---

## Advanced Monitoring

### Prometheus Metrics Queries

**Block height:**
```
tendermint_consensus_height
```

**Peer count:**
```
tendermint_p2p_peers
```

**Memory usage:**
```
process_resident_memory_bytes
```

**Block interval:**
```
rate(tendermint_consensus_height[1m])
```

### Export Metrics for Analysis
```bash
# Export Prometheus data
curl -s http://localhost:9094/api/v1/query?query=tendermint_consensus_height > heights.json

# Parse with jq
cat heights.json | jq '.data.result'
```

---

## Quick Reference Card

```
┌─────────────────────────────────────────────────────────────────────┐
│ AURA TESTNET MONITORING - QUICK REFERENCE                          │
├─────────────────────────────────────────────────────────────────────┤
│ ESSENTIAL COMMANDS                                                  │
│   Health check:        ./scripts/testnet-monitor.sh quick          │
│   Live monitor:        ./scripts/testnet-monitor.sh watch          │
│   Watch blocks:        ./scripts/testnet-monitor.sh watch-blocks   │
│   Check logs:          ./scripts/testnet-monitor.sh check-logs all │
│   Network health:      ./scripts/testnet-monitor.sh network        │
│   Performance:         ./scripts/testnet-monitor.sh performance    │
├─────────────────────────────────────────────────────────────────────┤
│ TROUBLESHOOTING                                                     │
│   Diagnose:            ./scripts/testnet-monitor.sh diagnose all   │
│   Restart failed:      ./scripts/testnet-monitor.sh restart-failed │
│   View logs:           ./scripts/testnet-monitor.sh logs val-1     │
│   Clean shutdown:      ./scripts/testnet-monitor.sh shutdown       │
├─────────────────────────────────────────────────────────────────────┤
│ ENDPOINTS                                                           │
│   Validator 1 RPC:     http://localhost:27657                      │
│   Validator 2 RPC:     http://localhost:27757                      │
│   Validator 3 RPC:     http://localhost:27857                      │
│   Validator 4 RPC:     http://localhost:27957                      │
│   Prometheus:          http://localhost:9094                       │
│   Grafana:             http://localhost:3002                       │
├─────────────────────────────────────────────────────────────────────┤
│ REQUIREMENTS                                                        │
│   Consensus:           3/4 validators required                     │
│   Normal block time:   2-4 seconds                                 │
│   Expected peers:      3 per validator                             │
├─────────────────────────────────────────────────────────────────────┤
│ COMMON FIXES                                                        │
│   Chain halted:        ./scripts/testnet-monitor.sh restart-failed │
│   Sync issues:         ./scripts/testnet-monitor.sh network        │
│   High errors:         ./scripts/testnet-monitor.sh check-logs all │
│   Full reset:          ./scripts/testnet-manage.sh clean && init   │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Support

For detailed information, see:
- **Full testnet setup:** `/TESTNET_SETUP.md`
- **Quick start guide:** `/TESTNET_QUICKSTART.md`
- **Docker runbook:** `/docs/runbooks/LOCAL_TESTNET_DOCKER.md`
- **Production roadmap:** `/ROADMAP_PRODUCTION.md`

**Report issues:** Create a GitHub issue with diagnostic output from:
```bash
./scripts/testnet-monitor.sh diagnose all > diagnostic-report.txt
```
