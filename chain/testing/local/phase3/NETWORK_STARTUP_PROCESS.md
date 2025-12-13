# Phase 3.1: Network Startup Process Documentation

## Overview

This document describes the complete network startup process for the Aura 4-node testnet, including infrastructure, initialization, and verification steps.

## Architecture

The Aura testnet consists of:

- **4 Validator Nodes**: `validator-1`, `validator-2`, `validator-3`, `validator-4`
- **2 Sentry Nodes**: `sentry-1`, `sentry-2` (protect validators from direct P2P exposure)
- **1 Observer Node**: `observer-1` (non-validating full node)
- **1 Counter Node**: `counter` (specialized node for testing)
- **Monitoring Infrastructure**: Prometheus, Grafana
- **Block Explorer**: For network visualization
- **Faucet Service**: For testnet token distribution

## Container Names

Docker containers follow the naming pattern: `aura-<role>-<number>`

Active containers:
- `aura-validator-1`, `aura-validator-2`, `aura-validator-3`, `aura-validator-4`
- `aura-sentry-1`, `aura-sentry-2`
- `aura-observer-1`
- `aura-counter`
- `aura-testnet-prometheus`
- `aura-testnet-grafana`
- `aura-block-explorer`
- `aura-testnet-proxy`
- `aura-faucet-backend`, `aura-faucet-db`, `aura-faucet-redis`

## Port Mappings

### Validator 1 (Primary)
- **RPC**: http://localhost:26657 (Tendermint RPC)
- **API**: http://localhost:1317 (Cosmos REST API)
- **gRPC**: localhost:9090 (gRPC queries)
- **P2P**: localhost:26656 (Peer-to-peer communication)
- **Metrics**: http://localhost:26660 (Prometheus metrics)

### Validator 2
- **RPC**: http://localhost:26757
- **API**: http://localhost:1417
- **gRPC**: localhost:9190
- **P2P**: localhost:26756
- **Metrics**: http://localhost:26760

### Validator 3
- **RPC**: http://localhost:26857
- **API**: http://localhost:1517
- **gRPC**: localhost:9290
- **P2P**: localhost:26856
- **Metrics**: http://localhost:26860

### Validator 4
- **RPC**: http://localhost:26957
- **API**: http://localhost:1617
- **gRPC**: localhost:9390
- **P2P**: localhost:26956
- **Metrics**: http://localhost:26960

### Monitoring
- **Prometheus**: http://localhost:9091
- **Grafana**: http://localhost:3001 (admin/aura-testnet-admin)
- **Block Explorer**: http://localhost:10080

## Startup Process

### Prerequisites

1. **Docker Running**: Ensure Docker daemon is active
   ```bash
   docker ps >/dev/null 2>&1 || sudo systemctl start docker
   ```

2. **Environment Loaded**: Source the Aura environment
   ```bash
   cd /home/hudson/blockchain-projects/aura
   source env.sh
   ```

3. **Testnet Data Initialized**: If first run, initialize testnet data
   ```bash
   ./scripts/testnet-init.sh
   ```

### Launch Sequence

The `launch-testnet.sh` script performs a 6-step launch sequence:

#### Step 1: Docker Status Check
- Verifies Docker daemon is running
- Exits with error if Docker is unavailable

#### Step 2: Volume Population
- Runs `testnet-data/populate-volumes.sh`
- Copies genesis files, validator keys, and config to Docker volumes
- Ensures all 4 validators have consistent genesis state

#### Step 3: Container Startup
- Executes `./scripts/testnet-manage.sh start`
- Launches all containers via `docker-compose.testnet.yml`
- Starts validators in parallel

#### Step 4: Initialization Wait (90 seconds)
- Allows time for:
  - P2P peer discovery
  - Genesis block creation
  - Initial consensus round
  - Store initialization
  - Module bootstrapping

#### Step 5: Container Verification
- Lists all running containers
- Displays status and port mappings
- Confirms all expected containers are up

#### Step 6: Health Check
- Queries `/health` endpoint on each validator
- Ports checked: 27657, 27757, 27857, 27957
- Note: Port mapping differs from management script (27xxx vs 26xxx)

#### Bonus: Consensus Status
- Runs `./scripts/testnet-monitor.sh quick`
- Shows block height across all nodes
- Confirms consensus is active

### Launch Command

```bash
./launch-testnet.sh
```

Expected output:
```
==================================================
  Aura 4-Node Testnet Launcher
==================================================

Step 1/6: Checking Docker status...
✓ Docker is running

Step 2/6: Populating testnet volumes...
✓ Volumes populated

Step 3/6: Starting 4-node testnet...
✓ Testnet started

Step 4/6: Waiting for node initialization (90 seconds)...
✓ Initialization period complete

Step 5/6: Verifying container status...
[Container listing]

Step 6/6: Health check on all nodes...
✓ Port 27657: Healthy
✓ Port 27757: Healthy
✓ Port 27857: Healthy
✓ Port 27957: Healthy

Consensus Status:
[Block heights]

==================================================
  Testnet Launch Complete!
==================================================
```

## Management Commands

### Start/Stop/Restart
```bash
./scripts/testnet-manage.sh start    # Start all validators
./scripts/testnet-manage.sh stop     # Stop all validators
./scripts/testnet-manage.sh restart  # Restart all validators
```

### Status Monitoring
```bash
./scripts/testnet-manage.sh status   # Show validator status and block heights
./scripts/testnet-manage.sh health   # Check Docker health status
./scripts/testnet-manage.sh ports    # Show all port mappings
```

### Logs
```bash
./scripts/testnet-manage.sh logs validator-1    # Follow logs for validator-1
docker logs -f aura-validator-2                 # Alternative method
```

### RPC Queries
```bash
./scripts/testnet-manage.sh query validator-1   # Query RPC status
curl -s http://localhost:26657/status | jq     # Direct RPC query
```

### Container Execution
```bash
./scripts/testnet-manage.sh exec validator-1 aurad status
./scripts/testnet-manage.sh exec validator-2 sh
```

## Verification Steps

### 1. Check All Containers Running
```bash
docker ps --format "table {{.Names}}\t{{.Status}}" | grep aura
```

Expected: All 15+ containers in "Up" status

### 2. Verify Block Production
```bash
# Check validator-1 height
curl -s http://localhost:26657/status | jq -r '.result.sync_info.latest_block_height'

# Wait 5 seconds
sleep 5

# Check again - should increment
curl -s http://localhost:26657/status | jq -r '.result.sync_info.latest_block_height'
```

Expected: Block height increases

### 3. Verify All Validators at Same Height
```bash
for port in 26657 26757 26857 26957; do
  height=$(curl -s http://localhost:$port/status | jq -r '.result.sync_info.latest_block_height')
  echo "Port $port: Height $height"
done
```

Expected: All heights within 1-2 blocks of each other

### 4. Check Validator Set
```bash
curl -s http://localhost:26657/validators | jq '.result.validators | length'
```

Expected: 4 validators

### 5. Verify P2P Connectivity
```bash
curl -s http://localhost:26657/net_info | jq '.result.n_peers'
```

Expected: 3+ peers (connected to other validators)

### 6. Check Consensus Voting
```bash
curl -s http://localhost:26657/consensus_state | jq '.result.round_state.height_vote_set[0].prevotes_bit_array'
```

Expected: Shows voting pattern (e.g., "BA{4:xxxx}")

## Store Initialization

### Verify Store Creation
```bash
./scripts/testnet-manage.sh verify-stores validator-1
```

This checks that all required KVStores are initialized:
- `auth`
- `bank`
- `staking`
- `governance`
- All 27 Aura custom modules

### AppHash Consistency
```bash
./scripts/testnet-manage.sh check-apphash validator-1
```

Verifies that:
- AppHash is recorded at genesis
- AppHash remains consistent across queries
- Store state is deterministic

## Byzantine Fault Tolerance Test

Built-in BFT test verifies consensus with 3/4 validators:

```bash
./scripts/testnet-manage.sh bft-test
```

Process:
1. Stops `validator-4` (leaving 3/4 active)
2. Waits 10 seconds for network adjustment
3. Checks if chain continues producing blocks
4. Restarts `validator-4`

Expected result: Chain continues with 3/4 validators (75% voting power > 2/3 threshold)

## Restart Consistency Test

Verifies AppHash consistency across full restart:

```bash
./scripts/testnet-manage.sh test-restart
```

Process:
1. Records baseline AppHash for all 4 validators
2. Stops all validators
3. Restarts all validators
4. Verifies AppHash matches baseline

Expected: All validators have identical AppHash before and after restart

## Monitoring Access

### Prometheus
```bash
curl http://localhost:9091/metrics | grep tendermint_consensus_height
```

### Grafana Dashboards
- URL: http://localhost:3001
- Credentials: admin / aura-testnet-admin
- Dashboards: Node health, consensus metrics, transaction throughput

### Block Explorer
- URL: http://localhost:10080
- Real-time block visualization
- Transaction search
- Account balances

## Common Issues

### Issue: Containers fail to start
**Symptoms**: `docker-compose up -d` fails
**Solution**:
```bash
./scripts/testnet-manage.sh clean
./scripts/testnet-init.sh
./launch-testnet.sh
```

### Issue: Nodes at different heights
**Symptoms**: Block heights diverge by >10 blocks
**Solution**:
```bash
# Check for network errors
docker logs aura-validator-1 | grep -i error

# Restart lagging validator
docker restart aura-validator-2
```

### Issue: Consensus halt
**Symptoms**: Block height stops increasing
**Solution**:
```bash
# Check validator status
./scripts/testnet-monitor.sh diagnose

# Check voting power
curl -s http://localhost:26657/validators | jq '.result.validators[] | {address, voting_power}'

# Verify >2/3 voting power online
```

### Issue: Port conflicts
**Symptoms**: "address already in use" error
**Solution**:
```bash
# Find conflicting process
sudo lsof -i :26657

# Kill process or change port mapping in docker-compose.testnet.yml
```

## Network Topology

```
                    ┌─────────────┐
                    │  Sentry-1   │
                    └──────┬──────┘
                           │
         ┌─────────────────┼─────────────────┐
         │                 │                 │
    ┌────▼────┐      ┌─────▼────┐      ┌────▼────┐
    │ Valid-1 │◄────►│ Valid-2  │◄────►│ Valid-3 │
    └────┬────┘      └─────┬────┘      └────┬────┘
         │                 │                 │
         └─────────────────┼─────────────────┘
                           │
                      ┌────▼────┐
                      │ Valid-4 │
                      └────┬────┘
                           │
                    ┌──────▼──────┐
                    │  Sentry-2   │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │  Observer   │
                    └─────────────┘
```

- **Validators**: Form full mesh for consensus
- **Sentries**: Shield validators from direct internet exposure
- **Observer**: Non-validating full node for RPC queries

## Genesis Configuration

Genesis file: `testnet-data/validator-1/config/genesis.json`

Key parameters:
- **Chain ID**: `aura-testnet-1`
- **Block Time**: ~6 seconds
- **Max Validators**: 100
- **Unbonding Time**: 21 days
- **Slashing**: Enabled
  - Double-sign: 5% slash
  - Downtime: 0.01% slash

## Next Steps

After successful startup:
1. Proceed to **Phase 3.2**: Consensus scenarios (4-node, 3-node, 2-node)
2. Execute **Phase 3.3**: Network chaos testing (latency, packet loss)
3. Implement **Phase 3.4**: Malicious peer handling tests

## References

- Management Script: `scripts/testnet-manage.sh`
- Monitoring Script: `scripts/testnet-monitor.sh`
- Docker Compose: `docker-compose.testnet.yml`
- Chaos Testing: `~/blockchain-projects/scripts/chaos-*.sh`
