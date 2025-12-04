# AURA Byzantine Fault Tolerance Test Guide

## Overview

This guide explains how to run comprehensive Byzantine Fault Tolerance (BFT) tests on the AURA blockchain's 4-validator testnet. The test suite validates that the blockchain maintains consensus with 3 out of 4 validators, halts appropriately with only 2 validators, and recovers properly when validators are restarted.

## Prerequisites

### 1. Running 4-Validator Testnet

The testnet must be initialized and running:

```bash
# Initialize testnet (one-time)
cd /home/decri/blockchain-projects/aura
./scripts/testnet-init.sh

# Start the testnet
cd /home/decri/blockchain-projects/aura
docker-compose -f docker-compose.testnet.yml up -d
```

### 2. Verify Testnet Health

Check that all 4 validators are running and synced:

```bash
# Check status of all validators
./scripts/testnet-manage.sh status

# Or check individually
./scripts/testnet-manage.sh query validator-1
./scripts/testnet-manage.sh query validator-2
./scripts/testnet-manage.sh query validator-3
./scripts/testnet-manage.sh query validator-4
```

Expected output should show all validators with increasing block heights and `catching_up: false`.

### 3. Required Tools

The BFT test script requires:
- `docker` - For container management
- `curl` - For RPC queries
- `jq` - For JSON parsing

Install if missing:
```bash
# Ubuntu/Debian
sudo apt-get install curl jq

# macOS
brew install curl jq
```

## Running the BFT Test

### Basic Execution

Run the comprehensive BFT test suite:

```bash
cd /home/decri/blockchain-projects/aura
./scripts/test-bft-comprehensive.sh
```

### With Verbose Output

For detailed real-time feedback:

```bash
./scripts/test-bft-comprehensive.sh --verbose
```

### Custom Output Directory

Save test reports to a specific directory:

```bash
./scripts/test-bft-comprehensive.sh --output-dir ./bft_results
```

### Combined Options

```bash
./scripts/test-bft-comprehensive.sh --verbose --output-dir ./bft_results
```

## Test Scenarios

### Scenario 1: Baseline Recording (Step 1)

**Purpose**: Establish initial state for comparison

**What Happens**:
1. Script waits 30 seconds for validators to stabilize
2. Records current block height on all 4 validators
3. Stores baseline for comparison with later measurements

**Success Criteria**:
- All validators have consistent block heights
- Heights are increasing (blocks being produced)
- All RPC endpoints are accessible

**Expected Output**:
```
BASELINE_VAL1=12345
BASELINE_VAL2=12345
BASELINE_VAL3=12345
BASELINE_VAL4=12345
```

### Scenario 2: 3/4 Validators Consensus Test (Steps 2-4)

**Purpose**: Verify that consensus continues with 3 active validators

**What Happens**:
1. Stops `aura-validator-3` (simulating validator offline)
2. Waits 30 seconds for network to adjust
3. Checks if remaining 3 validators continue producing blocks

**Success Criteria**:
- Chain continues producing blocks with 3/4 validators
- Block heights increase on validators 1, 2, and 4
- No consensus failures occur

**Expected Behavior**:
- BFT consensus threshold: 2/3 of validators must agree
- With 3/4 validators: 2/3 = 66.7%, so 3 > 2 validators can reach consensus
- Chain should produce 1-2 blocks per second

**Sample Output**:
```
3_OF_4_val1=12360
3_OF_4_val2=12359
3_OF_4_val4=12360
```

### Scenario 3: Validator Catch-up (Steps 5-6)

**Purpose**: Verify stopped validator can catch up via state sync

**What Happens**:
1. Records current heights of active validators
2. Restarts `aura-validator-3`
3. Monitors state sync process
4. Records final height when caught up

**Success Criteria**:
- RPC becomes available within 20 seconds
- State sync completes within 120 seconds
- Validator-3 height is within 5 blocks of other validators

**Expected Behavior**:
- State sync replays blocks from snapshot
- Catch-up usually completes in 30-60 seconds
- May see "catching_up: true" initially, then transitions to "catching_up: false"

**Sample Output**:
```
VAL3_CATCH_UP_HEIGHT=12370
VAL1_HEIGHT_AT_RESTART=12372
CATCH_UP_DIFF=2
```

### Scenario 4: 2/4 Validators Consensus Halt (Steps 7-8)

**Purpose**: Verify that consensus halts with insufficient validators

**What Happens**:
1. Stops both `aura-validator-2` and `aura-validator-3`
2. Waits 30 seconds
3. Checks if block production stops on remaining 2 validators

**Success Criteria**:
- NO blocks are produced with 2/4 validators
- Chain state is stable (no infinite loops or crashes)
- Remaining validators remain healthy

**Expected Behavior**:
- BFT threshold requires 2/3 consensus
- With 2/4 validators: 2/3 = 1.33 validators needed, but only 2 available
- However, with 2/4: 1/2 < 2/3, so consensus cannot be reached
- Chain halts gracefully at current height

**Sample Output**:
```
2_OF_4_BLOCKS_PRODUCED=0
```

**Important Note**: This is the critical test! If blocks continue to be produced with 2/4 validators, there's a BFT consensus bug.

### Scenario 5: Chain Recovery (Steps 9-10)

**Purpose**: Verify chain resumes when validators reconnect

**What Happens**:
1. Restarts `aura-validator-2` and `aura-validator-3`
2. Waits for RPC endpoints to become available
3. Allows 30 seconds for gossip/synchronization
4. Checks if block production resumes on all 4 validators

**Success Criteria**:
- All validators reconnect successfully
- RPC endpoints become available within 20 seconds
- Block production resumes
- Final heights are consistent across all validators

**Expected Behavior**:
- Network auto-discovery reconnects validators
- Peer gossip synchronizes mempool and state
- Consensus immediately resumes with 4/4 validators
- No data corruption or inconsistency

**Sample Output**:
```
RECOVERY_BLOCK_PRODUCTION=4
FINAL_HEIGHT_val1=12450
FINAL_HEIGHT_val2=12450
FINAL_HEIGHT_val3=12450
FINAL_HEIGHT_val4=12450
```

## Test Output Files

The script generates two report files:

### 1. Detailed Log Report (`.log`)

Format: `bft_test_YYYYMMDD_HHMMSS.log`

Contains:
- Timestamped events for each test phase
- Block heights at each step
- Success/failure indicators
- Diagnostic information

Example:
```
[2024-12-03 15:23:45] [INFO] Stopping aura-validator-3...
[2024-12-03 15:23:47] [SUCCESS] aura-validator-3 stopped
[2024-12-03 15:24:17] [INFO] Checking block production on remaining validators
[2024-12-03 15:24:27] [SUCCESS] Block production active on val1
```

### 2. JSON Report (`.json`)

Format: `bft_test_YYYYMMDD_HHMMSS.json`

Machine-readable format for:
- Automated analysis
- Integration with monitoring systems
- Historical tracking

Structure:
```json
{
  "test_name": "AURA BFT Comprehensive Test",
  "test_start": "2024-12-03 15:23:00",
  "test_end": "2024-12-03 15:45:00",
  "scenarios": {
    "baseline": { ... },
    "three_of_four": { ... },
    "catch_up": { ... },
    "two_of_four": { ... },
    "recovery": { ... }
  },
  "results": {
    "overall_status": "PASSED",
    "passed": 5,
    "failed": 0,
    "warnings": 0
  }
}
```

## Interpreting Results

### Successful BFT Test Outcome

A successful test exhibits these characteristics:

1. **Baseline Phase**: All 4 validators at identical or nearly identical block heights
2. **3/4 Consensus**: Chain produces blocks consistently without validator-3
3. **Catch-up Phase**: Validator-3 reaches within 5 blocks of others within 60 seconds
4. **2/4 Halt**: No blocks produced for 10+ seconds with only 2 validators
5. **Recovery**: Chain resumes and all 4 validators reach same height

### Common Issues and Troubleshooting

#### Issue: "RPC not available" on all validators

**Causes**:
- Testnet not running
- Docker daemon stopped
- Port mapping issues

**Solution**:
```bash
# Check if containers are running
docker ps | grep aura-validator

# Start testnet if stopped
cd /home/decri/blockchain-projects/aura
docker-compose -f docker-compose.testnet.yml up -d
```

#### Issue: Block heights not increasing

**Causes**:
- Validators not yet synced
- Network partition
- Consensus failure

**Solution**:
```bash
# Check individual validator logs
docker logs aura-validator-1 | tail -50

# Check consensus state
curl -s http://localhost:26657/consensus_state | jq '.'

# Wait longer for sync
sleep 60
```

#### Issue: 2/4 halt test shows continued block production

**This indicates a BFT consensus bug!**

**Investigation**:
1. Check validator logs for consensus errors:
   ```bash
   docker logs aura-validator-1 | grep -i "consensus\|prevote\|precommit"
   ```

2. Check voting power:
   ```bash
   curl -s http://localhost:26657/validators | jq '.result.validators'
   ```

3. Verify Tendermint configuration in `~/.aura/config/config.toml`

#### Issue: Validator doesn't catch up after restart

**Causes**:
- State sync disabled
- Snapshot not available
- Network issues

**Solution**:
```bash
# Check state sync config
docker exec aura-validator-3 cat /home/aura/.aura/config/config.toml | grep -A5 "statesync"

# Check for snapshots
curl -s http://localhost:26657/status | jq '.result.sync_info'

# Manual sync option
docker exec aura-validator-3 aurad tendermint unsafe-reset-all --home /home/aura/.aura
```

## Performance Expectations

### Block Production Rate

- **Normal**: 1-2 blocks per second per validator
- **With 3/4**: Slight slowdown possible but should maintain 0.5+ blocks/sec
- **Timeout**: Should occur within 10 seconds of detecting < 2/3 consensus

### State Sync Duration

- **Initial snapshot download**: 10-30 seconds
- **Blocks replay**: 20-60 seconds
- **Total catch-up**: 30-120 seconds maximum

### Network Recovery

- **Peer reconnection**: 5-10 seconds
- **Gossip propagation**: 5-15 seconds
- **Consensus resumption**: 10-30 seconds

## Advanced Testing

### Running Multiple Tests

Execute BFT tests sequentially:

```bash
#!/bin/bash
for i in {1..5}; do
    echo "Running BFT test iteration $i..."
    ./scripts/test-bft-comprehensive.sh --output-dir ./bft_results_run_$i
    sleep 60  # Wait between iterations
done
```

### Load Testing During BFT

Combine with transaction load to test consensus under stress:

```bash
# Terminal 1: Start transaction load
./scripts/tx-load-generator.sh --rate 10 --duration 300 &

# Terminal 2: Run BFT test
sleep 10
./scripts/test-bft-comprehensive.sh --verbose
```

### Monitoring with Prometheus/Grafana

While test runs, observe metrics:

1. Open Grafana: http://localhost:3002
2. Login: admin/aura-testnet-admin
3. View dashboards:
   - "AURA Testnet Overview"
   - "Validator Consensus State"
   - "Network P2P"

Key metrics to observe:
- `tendermint_consensus_state` - Current consensus state
- `tendermint_consensus_rounds` - Round number changes
- `tendermint_consensus_validators` - Active validator count
- `tendermint_p2p_peers_count` - Peer connections

## Security Considerations

### BFT Property Guarantees

The AURA blockchain uses the Cosmos SDK's Tendermint consensus engine, which guarantees:

1. **Safety**: With < 1/3 validators Byzantine, no conflicting commits
2. **Liveness**: With < 2/3 validators down, consensus continues
3. **Termination**: Within timeouts, consensus reaches a decision

### Test Validation

This BFT test validates:
- ✓ Safety: No data corruption with 3/4 validators
- ✓ Liveness: Chain continues with 3/4 validators
- ✓ Byzantine tolerance: Stops at 2/4 (< 2/3 threshold)
- ✓ Recovery: Resumes after all validators restart

### What This Test Does NOT Validate

- Byzantine validator behavior (double-signing, equivocation)
- Fork detection and recovery
- Time synchronization attacks
- Mempool fork handling
- Light client verification

## References

- [Tendermint Consensus Algorithm](https://tendermint.com/docs/introduction/what-is-tendermint.html)
- [Cosmos SDK Documentation](https://docs.cosmos.network/)
- [Byzantine Fault Tolerance](https://en.wikipedia.org/wiki/Byzantine_fault_tolerance)
- [BFT Consensus in Tendermint](https://docs.tendermint.com/master/introduction/)

## Support

For issues or questions:

1. Check the troubleshooting section above
2. Review validator logs: `docker logs aura-validator-N`
3. Check RPC status: `curl -s http://localhost:XXXX/status | jq '.'`
4. Review test output files in the output directory

## Appendix: Quick Reference

### Port Reference Table

| Validator | RPC Port | API Port | gRPC Port | P2P Port | Metrics Port |
|-----------|----------|----------|-----------|----------|--------------|
| validator-1 | 26657 | 1317 | 9090 | 26656 | 26660 |
| validator-2 | 26757 | 1417 | 9190 | 26756 | 26760 |
| validator-3 | 26857 | 1517 | 9290 | 26856 | 26860 |
| validator-4 | 26957 | 1617 | 9390 | 26956 | 26960 |

### Container Names

- `aura-validator-1`
- `aura-validator-2`
- `aura-validator-3`
- `aura-validator-4`

### Common Commands

```bash
# Start testnet
docker-compose -f docker-compose.testnet.yml up -d

# Stop testnet
docker-compose -f docker-compose.testnet.yml down

# View logs
docker logs -f aura-validator-1

# Check status
curl -s http://localhost:26657/status | jq '.'

# Stop a validator
docker stop aura-validator-3

# Start a validator
docker start aura-validator-3

# Run BFT test
./scripts/test-bft-comprehensive.sh --verbose
```
