# AURA Testnet Monitoring - One-Page Cheat Sheet

## Essential Commands (Copy/Paste Ready)

```bash
# Quick health check (1 command)
./scripts/testnet-monitor.sh quick

# Live monitoring dashboard
./scripts/testnet-monitor.sh watch

# Watch blocks being produced
./scripts/testnet-monitor.sh watch-blocks validator-1

# Check all logs for errors
./scripts/testnet-monitor.sh check-logs all

# Network health (peers, sync, RPC)
./scripts/testnet-monitor.sh network

# Performance metrics (block time, TPS)
./scripts/testnet-monitor.sh performance

# Diagnose all validators
./scripts/testnet-monitor.sh diagnose all

# Restart failed validators
./scripts/testnet-monitor.sh restart-failed

# Clean shutdown
./scripts/testnet-monitor.sh shutdown
```

## Individual Validator Commands

```bash
# Detailed status
./scripts/testnet-monitor.sh status validator-1
./scripts/testnet-monitor.sh status validator-2
./scripts/testnet-monitor.sh status validator-3
./scripts/testnet-monitor.sh status validator-4

# Tail logs
./scripts/testnet-monitor.sh logs validator-1

# Check specific validator logs
./scripts/testnet-monitor.sh check-logs validator-2 200
```

## Manual RPC Queries

```bash
# Block height
curl -s http://localhost:27657/status | jq '.result.sync_info.latest_block_height'
curl -s http://localhost:27757/status | jq '.result.sync_info.latest_block_height'
curl -s http://localhost:27857/status | jq '.result.sync_info.latest_block_height'
curl -s http://localhost:27957/status | jq '.result.sync_info.latest_block_height'

# Peer count
curl -s http://localhost:27657/net_info | jq '.result.n_peers'

# Catching up?
curl -s http://localhost:27657/status | jq '.result.sync_info.catching_up'

# Voting power
curl -s http://localhost:27657/status | jq '.result.validator_info.voting_power'
```

## Docker Quick Commands

```bash
# Container status
docker ps | grep aura-validator

# View logs
docker logs -f aura-validator-1

# Resource usage
docker stats --no-stream aura-validator-1

# Health status
docker inspect --format='{{.State.Health.Status}}' aura-validator-1

# Restart validator
docker-compose -f docker-compose.testnet.yml restart validator-2

# Execute in container
docker exec -it aura-validator-1 aurad status
```

## Endpoints

| Service | Endpoint |
|---------|----------|
| Validator 1 RPC | http://localhost:27657 |
| Validator 2 RPC | http://localhost:27757 |
| Validator 3 RPC | http://localhost:27857 |
| Validator 4 RPC | http://localhost:27957 |
| Prometheus | http://localhost:9094 |
| Grafana | http://localhost:3002 (admin/aura-testnet-admin) |

## Troubleshooting Flowchart

```
Chain not producing blocks?
  ↓
./scripts/testnet-monitor.sh quick
  ↓
<3 validators running?
  ↓ YES
./scripts/testnet-monitor.sh restart-failed
  ↓
./scripts/testnet-monitor.sh watch
  ↓ Still broken?
./scripts/testnet-monitor.sh diagnose all
  ↓
./scripts/testnet-monitor.sh check-logs all
  ↓ Still broken?
Clean restart:
  ./scripts/testnet-monitor.sh shutdown
  docker-compose -f docker-compose.testnet.yml up -d
  ./scripts/testnet-monitor.sh quick
```

## Common Errors & Fixes

| Error | Fix |
|-------|-----|
| "consensus failure" | Check validators: `./scripts/testnet-monitor.sh quick` |
| "insufficient voting power" | Need 3/4 validators: `./scripts/testnet-monitor.sh restart-failed` |
| "connection refused" | Check network: `./scripts/testnet-monitor.sh network` |
| "double sign" | **CRITICAL** Stop offending validator immediately |
| "timeout" | Check resources: `./scripts/testnet-monitor.sh performance` |
| Blocks stopped | Restart failed validators: `./scripts/testnet-monitor.sh restart-failed` |

## Health Status Indicators

| Status | Meaning | Action |
|--------|---------|--------|
| ✓ HEALTHY | 3-4 validators producing | None needed |
| ⚠ DEGRADED | 1-2 validators producing | Investigate soon |
| ✗ DOWN | 0 validators producing | **Immediate action** |

## Performance Baselines

| Metric | Normal | Warning | Critical |
|--------|--------|---------|----------|
| Block time | 2-4s | 5-10s | >10s |
| Peer count | 3 | 2 | <2 |
| Height variance | <5 blocks | 5-10 blocks | >10 blocks |
| Validators running | 4 | 3 | <3 |

## Pre-Flight Checklist

Before running tests:
```bash
# 1. Health check
./scripts/testnet-monitor.sh quick

# 2. Verify sync
./scripts/testnet-monitor.sh network

# 3. Check baseline performance
./scripts/testnet-monitor.sh performance

# 4. Scan for errors
./scripts/testnet-monitor.sh check-logs all
```

## During Operation

Run in separate terminal:
```bash
./scripts/testnet-monitor.sh watch
```

Or:
```bash
./scripts/testnet-monitor.sh watch-blocks validator-1
```

## Full Reset Procedure

```bash
# 1. Clean shutdown
./scripts/testnet-monitor.sh shutdown

# 2. Wipe data
./scripts/testnet-manage.sh clean

# 3. Reinitialize
./scripts/testnet-init.sh

# 4. Populate volumes
cd testnet-data && ./populate-volumes.sh && cd ..

# 5. Start
docker-compose -f docker-compose.testnet.yml up -d

# 6. Verify
./scripts/testnet-monitor.sh quick
```

## Log File Locations (Inside Containers)

```
/home/aura/.aura/config/config.toml
/home/aura/.aura/config/app.toml
/home/aura/.aura/config/genesis.json
/home/aura/.aura/data/
```

## Emergency Commands

```bash
# Stop everything immediately
docker-compose -f docker-compose.testnet.yml down

# Kill all aura containers
docker ps | grep aura | awk '{print $1}' | xargs docker kill

# View all container logs
docker-compose -f docker-compose.testnet.yml logs --tail=50

# Rebuild containers
docker-compose -f docker-compose.testnet.yml build --no-cache
```

## Help

```bash
# Full help
./scripts/testnet-monitor.sh help

# Testnet management
./scripts/testnet-manage.sh help

# Documentation
less TESTNET_MONITORING_GUIDE.md
```
