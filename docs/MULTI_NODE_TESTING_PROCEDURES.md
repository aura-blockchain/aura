# Multi-Node Testnet Testing Procedures

**Project:** Aura Blockchain
**Environment:** 4-Validator Local Testnet (Docker Compose)
**Last Updated:** December 10, 2025

---

## Table of Contents

1. [Environment Setup](#1-environment-setup)
2. [Consensus Testing](#2-consensus-testing)
3. [Transaction Testing](#3-transaction-testing)
4. [Network Partition Testing](#4-network-partition-testing)
5. [Validator Operations Testing](#5-validator-operations-testing)
6. [Governance Testing](#6-governance-testing)
7. [Module-Specific Testing](#7-module-specific-testing)
8. [Stress Testing](#8-stress-testing)
9. [Recovery Testing](#9-recovery-testing)
10. [Monitoring and Observability](#10-monitoring-and-observability)

---

## 1. Environment Setup

### 1.1 Testnet Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Docker Network: aura-testnet                │
├─────────────┬─────────────┬─────────────┬─────────────┬────────┤
│ validator-1 │ validator-2 │ validator-3 │ validator-4 │ Monitor│
│ RPC: 27657  │ RPC: 27757  │ RPC: 27857  │ RPC: 27957  │        │
│ API: 2317   │ API: 2417   │ API: 2517   │ API: 2617   │Prom:9094
│ P2P: 27656  │ P2P: 27756  │ P2P: 27856  │ P2P: 27956  │Graf:3002
│ gRPC: 10090 │ gRPC: 10190 │ gRPC: 10290 │ gRPC: 10390 │        │
└─────────────┴─────────────┴─────────────┴─────────────┴────────┘
```

### 1.2 Starting the Testnet

```bash
# Navigate to project root
cd /home/decri/blockchain-projects/aura

# Ensure Docker is running
sudo service docker start

# Build the latest image (after code changes)
docker build -t aurad:latest -f Dockerfile .

# Initialize testnet (if not already done)
bash scripts/testnet-init.sh

# Populate Docker volumes
cd testnet-data && bash populate-volumes.sh && cd ..

# Fix permissions on volumes
for i in 1 2 3 4; do
    docker run --rm -v aura_validator-$i-data:/data alpine sh -c "chown -R 1000:1000 /data && chmod -R 755 /data"
done

# Start all validators
docker compose -f docker-compose.testnet.yml up -d

# Verify all containers are running
docker ps --format "table {{.Names}}\t{{.Status}}"
```

### 1.3 Useful Commands

```bash
# Check node status
curl -s http://localhost:27657/status | jq '.result.sync_info'

# Get current block height
curl -s http://localhost:27657/status | jq -r '.result.sync_info.latest_block_height'

# View validator logs
docker logs aura-validator-1 -f --tail 100

# View all validator logs simultaneously
docker compose -f docker-compose.testnet.yml logs -f

# Stop testnet
docker compose -f docker-compose.testnet.yml down

# Reset testnet (clean slate)
docker compose -f docker-compose.testnet.yml down
docker volume ls | grep aura | awk '{print $2}' | xargs docker volume rm
bash scripts/testnet-init.sh
```

### 1.4 Test Accounts

The testnet is initialized with the following accounts (keys stored in each validator's keyring):

| Account | Purpose | Validator |
|---------|---------|-----------|
| validator1 | Validator 1 operator | validator-1 |
| validator2 | Validator 2 operator | validator-2 |
| validator3 | Validator 3 operator | validator-3 |
| validator4 | Validator 4 operator | validator-4 |
| faucet | Test token distribution | validator-1 |

```bash
# List keys on validator-1
docker exec aura-validator-1 aurad keys list --keyring-backend test

# Get faucet address
docker exec aura-validator-1 aurad keys show faucet -a --keyring-backend test
```

---

## 2. Consensus Testing

### 2.1 Basic Block Production Test

**Objective:** Verify all validators are producing blocks and reaching consensus.

```bash
# Test: Verify block production over 1 minute
echo "Starting block production test..."
START_HEIGHT=$(curl -s http://localhost:27657/status | jq -r '.result.sync_info.latest_block_height')
sleep 60
END_HEIGHT=$(curl -s http://localhost:27657/status | jq -r '.result.sync_info.latest_block_height')
BLOCKS_PRODUCED=$((END_HEIGHT - START_HEIGHT))
echo "Blocks produced in 60s: $BLOCKS_PRODUCED"
echo "Expected: ~20 blocks (3s block time)"

# Verify: Should produce 18-22 blocks per minute
if [ $BLOCKS_PRODUCED -ge 15 ] && [ $BLOCKS_PRODUCED -le 25 ]; then
    echo "✅ PASS: Block production within expected range"
else
    echo "❌ FAIL: Block production outside expected range"
fi
```

### 2.2 App Hash Consistency Test

**Objective:** Verify all validators compute identical app hashes (determinism check).

```bash
# Test: Compare app hashes across all validators for same block
BLOCK_HEIGHT=$(curl -s http://localhost:27657/status | jq -r '.result.sync_info.latest_block_height')

HASH1=$(curl -s http://localhost:27657/block?height=$BLOCK_HEIGHT | jq -r '.result.block.header.app_hash')
HASH2=$(curl -s http://localhost:27757/block?height=$BLOCK_HEIGHT | jq -r '.result.block.header.app_hash')
HASH3=$(curl -s http://localhost:27857/block?height=$BLOCK_HEIGHT | jq -r '.result.block.header.app_hash')
HASH4=$(curl -s http://localhost:27957/block?height=$BLOCK_HEIGHT | jq -r '.result.block.header.app_hash')

echo "Block $BLOCK_HEIGHT app hashes:"
echo "  Validator 1: $HASH1"
echo "  Validator 2: $HASH2"
echo "  Validator 3: $HASH3"
echo "  Validator 4: $HASH4"

if [ "$HASH1" = "$HASH2" ] && [ "$HASH2" = "$HASH3" ] && [ "$HASH3" = "$HASH4" ]; then
    echo "✅ PASS: All validators have identical app hash"
else
    echo "❌ FAIL: App hash mismatch detected - CONSENSUS BUG!"
fi
```

### 2.3 Validator Signing Test

**Objective:** Verify all validators are signing blocks.

```bash
# Test: Check commit signatures for recent block
BLOCK_HEIGHT=$(curl -s http://localhost:27657/status | jq -r '.result.sync_info.latest_block_height')

echo "Checking signatures for block $BLOCK_HEIGHT..."
curl -s http://localhost:27657/block?height=$BLOCK_HEIGHT | jq '.result.block.last_commit.signatures[] | {validator_address: .validator_address, signature: (.signature != null)}'

# Count non-null signatures
SIG_COUNT=$(curl -s http://localhost:27657/block?height=$BLOCK_HEIGHT | jq '[.result.block.last_commit.signatures[] | select(.signature != null)] | length')
echo "Validators signing: $SIG_COUNT/4"

if [ "$SIG_COUNT" -ge 3 ]; then
    echo "✅ PASS: Sufficient validators signing (2/3+ threshold met)"
else
    echo "❌ FAIL: Insufficient validator signatures"
fi
```

### 2.4 Consensus State Test

**Objective:** Verify consensus is not stuck in multiple rounds.

```bash
# Test: Check consensus state for stuck rounds
curl -s http://localhost:27657/dump_consensus_state | jq '.result.round_state | {height: .height, round: .round, step: .step}'

# If round > 0 consistently, there may be a consensus issue
ROUND=$(curl -s http://localhost:27657/dump_consensus_state | jq -r '.result.round_state.round')
if [ "$ROUND" = "0" ]; then
    echo "✅ PASS: Consensus proceeding normally (round 0)"
else
    echo "⚠️  WARNING: Consensus at round $ROUND (may indicate slow finalization)"
fi
```

---

## 3. Transaction Testing

### 3.1 Basic Token Transfer Test

**Objective:** Verify transactions are processed correctly across the network.

```bash
# Get addresses
FAUCET=$(docker exec aura-validator-1 aurad keys show faucet -a --keyring-backend test)
RECIPIENT=$(docker exec aura-validator-2 aurad keys show validator2 -a --keyring-backend test)

# Check initial balances
echo "Initial balances:"
docker exec aura-validator-1 aurad query bank balances $FAUCET --node http://localhost:26657
docker exec aura-validator-1 aurad query bank balances $RECIPIENT --node http://localhost:26657

# Send tokens
echo "Sending 1000000uaura from faucet to recipient..."
docker exec aura-validator-1 aurad tx bank send $FAUCET $RECIPIENT 1000000uaura \
    --chain-id aura-local-4 \
    --keyring-backend test \
    --fees 1000uaura \
    --yes \
    --node http://localhost:26657

# Wait for transaction to be included
sleep 5

# Check final balances
echo "Final balances:"
docker exec aura-validator-1 aurad query bank balances $FAUCET --node http://localhost:26657
docker exec aura-validator-1 aurad query bank balances $RECIPIENT --node http://localhost:26657

# Verify on different node (cross-node consistency)
echo "Verifying on validator-3..."
docker exec aura-validator-3 aurad query bank balances $RECIPIENT --node http://localhost:26657
```

### 3.2 Transaction Propagation Test

**Objective:** Verify transactions submitted to one node propagate to all nodes.

```bash
# Submit transaction to validator-2
FAUCET=$(docker exec aura-validator-1 aurad keys show faucet -a --keyring-backend test)
RECIPIENT=$(docker exec aura-validator-3 aurad keys show validator3 -a --keyring-backend test)

# Submit to validator-2's RPC
docker exec aura-validator-1 aurad tx bank send $FAUCET $RECIPIENT 500000uaura \
    --chain-id aura-local-4 \
    --keyring-backend test \
    --fees 1000uaura \
    --yes \
    --node http://aura-validator-2:26657

sleep 5

# Query from validator-4 (different node)
echo "Querying transaction from validator-4..."
docker exec aura-validator-4 aurad query bank balances $RECIPIENT --node http://localhost:26657
```

### 3.3 Transaction Mempool Test

**Objective:** Verify mempool behavior under load.

```bash
# Check mempool status on each node
for port in 27657 27757 27857 27957; do
    echo "Mempool on port $port:"
    curl -s http://localhost:$port/unconfirmed_txs | jq '{n_txs: .result.n_txs, total_bytes: .result.total_bytes}'
done
```

### 3.4 Failed Transaction Test

**Objective:** Verify invalid transactions are rejected properly.

```bash
# Test: Send more tokens than available (should fail)
FAUCET=$(docker exec aura-validator-1 aurad keys show faucet -a --keyring-backend test)
RECIPIENT=$(docker exec aura-validator-2 aurad keys show validator2 -a --keyring-backend test)

echo "Attempting to send more tokens than available..."
docker exec aura-validator-1 aurad tx bank send $FAUCET $RECIPIENT 999999999999999uaura \
    --chain-id aura-local-4 \
    --keyring-backend test \
    --fees 1000uaura \
    --yes \
    --node http://localhost:26657 2>&1

# Should see insufficient funds error
```

---

## 4. Network Partition Testing

### 4.1 Single Validator Isolation Test

**Objective:** Verify network continues when 1 of 4 validators goes offline.

```bash
# Stop validator-4
echo "Stopping validator-4..."
docker stop aura-validator-4

# Wait and check if network continues
sleep 10
echo "Checking block production with 3 validators..."

START_HEIGHT=$(curl -s http://localhost:27657/status | jq -r '.result.sync_info.latest_block_height')
sleep 30
END_HEIGHT=$(curl -s http://localhost:27657/status | jq -r '.result.sync_info.latest_block_height')

if [ $END_HEIGHT -gt $START_HEIGHT ]; then
    echo "✅ PASS: Network continues with 3/4 validators"
else
    echo "❌ FAIL: Network halted with 3/4 validators"
fi

# Restart validator-4
echo "Restarting validator-4..."
docker start aura-validator-4
sleep 10

# Verify it catches up
V4_HEIGHT=$(curl -s http://localhost:27957/status | jq -r '.result.sync_info.latest_block_height')
V1_HEIGHT=$(curl -s http://localhost:27657/status | jq -r '.result.sync_info.latest_block_height')
echo "Validator-4 height: $V4_HEIGHT, Validator-1 height: $V1_HEIGHT"
```

### 4.2 Two Validator Isolation Test

**Objective:** Verify network halts when 2 of 4 validators go offline (below 2/3 threshold).

```bash
# Stop validators 3 and 4
echo "Stopping validators 3 and 4..."
docker stop aura-validator-3 aura-validator-4

sleep 10
echo "Checking if network halts with only 2 validators..."

START_HEIGHT=$(curl -s http://localhost:27657/status | jq -r '.result.sync_info.latest_block_height')
sleep 30
END_HEIGHT=$(curl -s http://localhost:27657/status | jq -r '.result.sync_info.latest_block_height')

if [ $END_HEIGHT -eq $START_HEIGHT ]; then
    echo "✅ PASS: Network correctly halted (below 2/3 threshold)"
else
    echo "⚠️  WARNING: Network continued with only 2/4 validators"
fi

# Restart validators
echo "Restarting validators 3 and 4..."
docker start aura-validator-3 aura-validator-4
sleep 15

# Verify network resumes
END_HEIGHT2=$(curl -s http://localhost:27657/status | jq -r '.result.sync_info.latest_block_height')
if [ $END_HEIGHT2 -gt $END_HEIGHT ]; then
    echo "✅ PASS: Network resumed after validators rejoined"
else
    echo "❌ FAIL: Network did not resume"
fi
```

### 4.3 Network Latency Simulation Test

**Objective:** Verify consensus handles network delays.

```bash
# Add latency to validator-2 container
echo "Adding 500ms latency to validator-2..."
docker exec aura-validator-2 tc qdisc add dev eth0 root netem delay 500ms 2>/dev/null || \
docker exec aura-validator-2 apt-get update && docker exec aura-validator-2 apt-get install -y iproute2 && \
docker exec aura-validator-2 tc qdisc add dev eth0 root netem delay 500ms

# Check if consensus still works
sleep 30
HEIGHT=$(curl -s http://localhost:27657/status | jq -r '.result.sync_info.latest_block_height')
echo "Current height: $HEIGHT"

# Remove latency
docker exec aura-validator-2 tc qdisc del dev eth0 root 2>/dev/null
```

---

## 5. Validator Operations Testing

### 5.1 Validator Status Test

**Objective:** Verify all validators are bonded and active.

```bash
# Query validators
docker exec aura-validator-1 aurad query staking validators --node http://localhost:26657 -o json | \
    jq '.validators[] | {moniker: .description.moniker, status: .status, tokens: .tokens, jailed: .jailed}'
```

### 5.2 Validator Jailing Test

**Objective:** Test validator jailing and unjailing flow.

```bash
# Note: This requires the validator to miss blocks, which can be simulated by stopping it

# Stop validator-4 for extended period (will get jailed for downtime)
echo "Stopping validator-4 to trigger jailing..."
docker stop aura-validator-4

# Wait for jailing (typically after missing ~100 blocks)
echo "Waiting for validator-4 to be jailed (this may take several minutes)..."
sleep 300  # 5 minutes

# Check if jailed
docker exec aura-validator-1 aurad query staking validators --node http://localhost:26657 -o json | \
    jq '.validators[] | select(.description.moniker == "validator-4") | {jailed: .jailed, status: .status}'

# Restart and unjail
docker start aura-validator-4
sleep 30

# Unjail (must be done by validator operator)
docker exec aura-validator-4 aurad tx slashing unjail \
    --from validator4 \
    --chain-id aura-local-4 \
    --keyring-backend test \
    --fees 1000uaura \
    --yes \
    --node http://localhost:26657
```

### 5.3 Delegation Test

**Objective:** Test token delegation to validators.

```bash
FAUCET=$(docker exec aura-validator-1 aurad keys show faucet -a --keyring-backend test)
VALIDATOR_ADDR=$(docker exec aura-validator-1 aurad query staking validators --node http://localhost:26657 -o json | \
    jq -r '.validators[0].operator_address')

echo "Delegating 10000000uaura to $VALIDATOR_ADDR..."
docker exec aura-validator-1 aurad tx staking delegate $VALIDATOR_ADDR 10000000uaura \
    --from faucet \
    --chain-id aura-local-4 \
    --keyring-backend test \
    --fees 1000uaura \
    --yes \
    --node http://localhost:26657

sleep 5

# Query delegation
docker exec aura-validator-1 aurad query staking delegations $FAUCET --node http://localhost:26657
```

---

## 6. Governance Testing

### 6.1 Proposal Submission Test

**Objective:** Test governance proposal creation.

```bash
FAUCET=$(docker exec aura-validator-1 aurad keys show faucet -a --keyring-backend test)

# Create a text proposal
docker exec aura-validator-1 aurad tx gov submit-proposal \
    --title "Test Proposal" \
    --description "This is a test proposal for the 4-validator testnet" \
    --type Text \
    --deposit 10000000uaura \
    --from faucet \
    --chain-id aura-local-4 \
    --keyring-backend test \
    --fees 1000uaura \
    --yes \
    --node http://localhost:26657

sleep 5

# Query proposals
docker exec aura-validator-1 aurad query gov proposals --node http://localhost:26657
```

### 6.2 Voting Test

**Objective:** Test voting from multiple validators.

```bash
PROPOSAL_ID=1  # Adjust based on actual proposal ID

# Vote from each validator
for i in 1 2 3 4; do
    echo "Voting YES from validator$i..."
    docker exec aura-validator-$i aurad tx gov vote $PROPOSAL_ID yes \
        --from validator$i \
        --chain-id aura-local-4 \
        --keyring-backend test \
        --fees 1000uaura \
        --yes \
        --node http://localhost:26657
    sleep 2
done

# Check vote tally
docker exec aura-validator-1 aurad query gov tally $PROPOSAL_ID --node http://localhost:26657
```

---

## 7. Module-Specific Testing

### 7.1 DEX Module Test

**Objective:** Test DEX liquidity pool operations.

```bash
# Create a liquidity pool (if DEX module is enabled)
# Query existing pools
docker exec aura-validator-1 aurad query dex pools --node http://localhost:26657

# Add liquidity (adjust parameters based on actual DEX implementation)
# docker exec aura-validator-1 aurad tx dex add-liquidity ...
```

### 7.2 Identity Module Test

**Objective:** Test DID operations.

```bash
# Query DIDs
docker exec aura-validator-1 aurad query identity dids --node http://localhost:26657

# Create a DID (adjust based on actual module implementation)
# docker exec aura-validator-1 aurad tx identity create-did ...
```

### 7.3 Bridge Module Test

**Objective:** Test cross-chain bridge operations.

```bash
# Query bridge status
docker exec aura-validator-1 aurad query bridge params --node http://localhost:26657

# Query pending transfers
docker exec aura-validator-1 aurad query bridge pending-transfers --node http://localhost:26657
```

### 7.4 WASM Module Test

**Objective:** Test smart contract deployment and execution.

```bash
# Check WASM params
docker exec aura-validator-1 aurad query wasm params --node http://localhost:26657

# List deployed contracts
docker exec aura-validator-1 aurad query wasm list-code --node http://localhost:26657

# Deploy a test contract (requires .wasm file)
# docker exec aura-validator-1 aurad tx wasm store /path/to/contract.wasm ...
```

---

## 8. Stress Testing

### 8.1 Transaction Flood Test

**Objective:** Test network under high transaction load.

```bash
#!/bin/bash
# stress_test.sh - Run from validator-1 container

FAUCET=$(aurad keys show faucet -a --keyring-backend test)
RECIPIENT=$(aurad keys show validator2 -a --keyring-backend test)

echo "Starting stress test: 100 transactions..."
for i in $(seq 1 100); do
    aurad tx bank send $FAUCET $RECIPIENT 1000uaura \
        --chain-id aura-local-4 \
        --keyring-backend test \
        --fees 1000uaura \
        --yes \
        --node http://localhost:26657 \
        --sequence $i \
        --broadcast-mode async &
done

wait
echo "Stress test complete. Check mempool and block inclusion."
```

### 8.2 Block Size Limit Test

**Objective:** Test behavior when blocks approach maximum size.

```bash
# Monitor block sizes
for i in $(seq 1 10); do
    HEIGHT=$(curl -s http://localhost:27657/status | jq -r '.result.sync_info.latest_block_height')
    SIZE=$(curl -s http://localhost:27657/block?height=$HEIGHT | jq '.result.block | (. | tostring | length)')
    echo "Block $HEIGHT size: $SIZE bytes"
    sleep 3
done
```

### 8.3 Long-Running Stability Test

**Objective:** Verify network stability over extended period.

```bash
#!/bin/bash
# stability_test.sh - Run for 1 hour

echo "Starting 1-hour stability test..."
START_TIME=$(date +%s)
END_TIME=$((START_TIME + 3600))
BLOCK_ERRORS=0
LAST_HEIGHT=0

while [ $(date +%s) -lt $END_TIME ]; do
    HEIGHT=$(curl -s http://localhost:27657/status | jq -r '.result.sync_info.latest_block_height' 2>/dev/null)

    if [ -z "$HEIGHT" ] || [ "$HEIGHT" = "null" ]; then
        echo "$(date): ERROR - Could not get block height"
        BLOCK_ERRORS=$((BLOCK_ERRORS + 1))
    elif [ "$HEIGHT" -le "$LAST_HEIGHT" ] && [ "$LAST_HEIGHT" -ne 0 ]; then
        echo "$(date): WARNING - Block height not increasing ($HEIGHT)"
        BLOCK_ERRORS=$((BLOCK_ERRORS + 1))
    else
        echo "$(date): OK - Height $HEIGHT"
        LAST_HEIGHT=$HEIGHT
    fi

    sleep 30
done

echo "Stability test complete. Errors: $BLOCK_ERRORS"
```

---

## 9. Recovery Testing

### 9.1 State Sync Test

**Objective:** Test new node catching up via state sync.

```bash
# This requires a 5th node configured for state sync
# Configuration in config.toml:
# [statesync]
# enable = true
# rpc_servers = "http://validator-1:26657,http://validator-2:26657"
# trust_height = <recent_height>
# trust_hash = "<block_hash_at_trust_height>"
```

### 9.2 Database Corruption Recovery Test

**Objective:** Test recovery from corrupted database.

```bash
# Stop validator-4
docker stop aura-validator-4

# Corrupt the database (CAUTION: test environment only)
docker run --rm -v aura_validator-4-data:/data alpine sh -c "rm -rf /data/data/application.db"

# Restart - should recover via sync
docker start aura-validator-4

# Monitor catch-up
watch -n 5 'curl -s http://localhost:27957/status | jq ".result.sync_info"'
```

### 9.3 Genesis Export/Import Test

**Objective:** Test chain export and import.

```bash
# Export genesis at current height
docker exec aura-validator-1 aurad export --node http://localhost:26657 > exported_genesis.json

# Validate exported genesis
docker exec aura-validator-1 aurad validate-genesis /path/to/exported_genesis.json
```

---

## 10. Monitoring and Observability

### 10.1 Prometheus Metrics Test

**Objective:** Verify metrics collection is working.

```bash
# Access Prometheus
curl -s http://localhost:9094/api/v1/targets | jq '.data.activeTargets[] | {instance: .labels.instance, health: .health}'

# Query specific metrics
curl -s 'http://localhost:9094/api/v1/query?query=cometbft_consensus_height' | jq '.data.result'
curl -s 'http://localhost:9094/api/v1/query?query=cometbft_consensus_rounds' | jq '.data.result'
curl -s 'http://localhost:9094/api/v1/query?query=cometbft_consensus_validators' | jq '.data.result'
```

### 10.2 Grafana Dashboard Test

**Objective:** Verify Grafana dashboards are loading.

```bash
# Access Grafana at http://localhost:3002
# Default credentials: admin/admin

# Check datasource health via API
curl -s http://admin:admin@localhost:3002/api/datasources | jq '.[].name'
```

### 10.3 Log Analysis Test

**Objective:** Check for errors and warnings in logs.

```bash
# Check for errors across all validators
for i in 1 2 3 4; do
    echo "=== Validator $i Errors ==="
    docker logs aura-validator-$i 2>&1 | grep -i "error\|panic\|fatal" | tail -10
done

# Check for consensus warnings
docker logs aura-validator-1 2>&1 | grep -i "round\|timeout\|prevote\|precommit" | tail -20
```

### 10.4 Resource Usage Test

**Objective:** Monitor container resource usage.

```bash
# Check CPU and memory usage
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}"

# Check disk usage
for i in 1 2 3 4; do
    SIZE=$(docker run --rm -v aura_validator-$i-data:/data alpine du -sh /data | cut -f1)
    echo "Validator $i data size: $SIZE"
done
```

---

## Test Checklist

Use this checklist to track testing progress:

### Consensus Tests
- [ ] Basic block production (60 seconds)
- [ ] App hash consistency across all validators
- [ ] Validator signing verification
- [ ] Consensus state (no stuck rounds)

### Transaction Tests
- [ ] Basic token transfer
- [ ] Transaction propagation across nodes
- [ ] Mempool status verification
- [ ] Failed transaction handling

### Network Partition Tests
- [ ] Single validator isolation (3/4 online)
- [ ] Two validator isolation (2/4 online - should halt)
- [ ] Recovery after partition heals

### Validator Operations Tests
- [ ] Validator status query
- [ ] Delegation flow
- [ ] Jailing/unjailing flow (optional - takes time)

### Governance Tests
- [ ] Proposal submission
- [ ] Multi-validator voting
- [ ] Vote tally verification

### Module Tests
- [ ] DEX module queries
- [ ] Identity module queries
- [ ] Bridge module queries
- [ ] WASM module queries

### Stress Tests
- [ ] Transaction flood (100 txs)
- [ ] Block size monitoring
- [ ] Long-running stability (optional - 1 hour)

### Recovery Tests
- [ ] Single validator restart and catchup
- [ ] Genesis export validation

### Monitoring Tests
- [ ] Prometheus targets healthy
- [ ] Grafana accessible
- [ ] Log analysis (no critical errors)
- [ ] Resource usage within limits

---

## Troubleshooting

### Common Issues

**1. Validators not connecting to each other**
```bash
# Check P2P connectivity
docker exec aura-validator-1 aurad status | jq '.NodeInfo.listen_addr'
docker exec aura-validator-1 curl http://aura-validator-2:26657/net_info
```

**2. Consensus stuck at round > 0**
```bash
# Check for non-determinism
# Compare app hashes at stuck height across validators
```

**3. Transaction not being included**
```bash
# Check mempool
curl http://localhost:27657/unconfirmed_txs

# Check if tx was rejected
docker logs aura-validator-1 2>&1 | grep "rejected"
```

**4. Validator falling behind**
```bash
# Check sync status
curl http://localhost:27957/status | jq '.result.sync_info.catching_up'

# Force fast sync
docker restart aura-validator-4
```

---

## Automated Test Script

Save this as `run_all_tests.sh` for automated execution:

```bash
#!/bin/bash
# run_all_tests.sh - Comprehensive testnet validation

set -e

echo "========================================"
echo "  Aura 4-Validator Testnet Test Suite  "
echo "========================================"

PASSED=0
FAILED=0

run_test() {
    local name=$1
    local cmd=$2
    echo -n "Testing: $name... "
    if eval "$cmd" > /dev/null 2>&1; then
        echo "✅ PASS"
        PASSED=$((PASSED + 1))
    else
        echo "❌ FAIL"
        FAILED=$((FAILED + 1))
    fi
}

# Consensus tests
run_test "Block production" '[ $(curl -s http://localhost:27657/status | jq -r ".result.sync_info.latest_block_height") -gt 0 ]'
run_test "Validator-1 online" 'curl -s http://localhost:27657/status | jq -e ".result"'
run_test "Validator-2 online" 'curl -s http://localhost:27757/status | jq -e ".result"'
run_test "Validator-3 online" 'curl -s http://localhost:27857/status | jq -e ".result"'
run_test "Validator-4 online" 'curl -s http://localhost:27957/status | jq -e ".result"'

# App hash consistency
HEIGHT=$(curl -s http://localhost:27657/status | jq -r '.result.sync_info.latest_block_height')
HASH1=$(curl -s http://localhost:27657/block?height=$HEIGHT | jq -r '.result.block.header.app_hash')
HASH2=$(curl -s http://localhost:27757/block?height=$HEIGHT | jq -r '.result.block.header.app_hash')
run_test "App hash consistency" '[ "$HASH1" = "$HASH2" ]'

# Monitoring
run_test "Prometheus healthy" 'curl -s http://localhost:9094/-/healthy'
run_test "Grafana healthy" 'curl -s http://localhost:3002/api/health | jq -e ".database"'

echo "========================================"
echo "Results: $PASSED passed, $FAILED failed"
echo "========================================"

exit $FAILED
```

---

*Document created: December 10, 2025*
*For use with Aura 4-validator local testnet*
