# Phase 3: 4-Validator Consensus Testing

This directory contains comprehensive consensus testing scripts for a 4-validator Aura testnet.

## Overview

The Phase 3 test suite validates Byzantine Fault Tolerance (BFT) consensus properties with a 4-validator network where each validator has equal voting power (25%).

### BFT Consensus Requirements

- **Total Validators**: 4
- **Voting Power per Validator**: 25%
- **Consensus Threshold**: >66% (>2/3 of voting power)

### Expected Behavior

| Active Validators | Voting Power | Can Produce Blocks? | Reason |
|------------------|--------------|---------------------|--------|
| 4 | 100% | ✓ Yes | >66% threshold met |
| 3 | 75% | ✓ Yes | >66% threshold met |
| 2 | 50% | ✗ No | <66% threshold |
| 1 | 25% | ✗ No | <66% threshold |
| 0 | 0% | ✗ No | No validators |

## Prerequisites

1. **Environment Setup**
   ```bash
   cd /home/hudson/blockchain-projects/aura
   source env.sh
   ```

2. **Build Binary**
   ```bash
   cd chain
   make build
   # Binary location: /home/hudson/blockchain-projects/aura/chain/aurad
   ```

3. **4-Validator Setup**
   Run the 4-validator setup script (from Phase 3 setup agent) to initialize:
   - 4 validator nodes with separate homes
   - Equal voting power distribution
   - Proper P2P connectivity
   - Genesis configuration

## Test Scripts

### 1. 4-validator-consensus-test.sh

**Main consensus test suite** - Validates all BFT properties.

#### Tests Performed

1. **Test 1: All 4 Validators (100% voting power)**
   - Expected: Chain produces blocks
   - Validates: Full network operation

2. **Test 2: 3 Validators (75% voting power)**
   - Expected: Chain produces blocks
   - Validates: Consensus with >2/3 threshold

3. **Test 3: 2 Validators (50% voting power)**
   - Expected: Chain HALTS
   - Validates: Insufficient voting power detection

4. **Test 4: 1 Validator (25% voting power)**
   - Expected: Chain HALTS
   - Validates: Single validator cannot produce blocks

5. **Test 5: Restart and Sync**
   - Expected: All validators restart and sync
   - Validates: Recovery and synchronization

#### Usage

```bash
# Make executable
chmod +x 4-validator-consensus-test.sh

# Run full test suite
./4-validator-consensus-test.sh

# Results saved to:
# consensus_test_results_YYYYMMDD_HHMMSS.log
```

#### Output

- Colored console output (green/red for pass/fail)
- Detailed test logs with timestamps
- Block height tracking
- Voting power calculations
- Pass/fail summary

### 2. consensus-analyzer.sh

**Interactive analysis tool** - Provides detailed consensus state analysis.

#### Features

- Validator set analysis (voting power distribution)
- Consensus state monitoring
- Block production analysis
- Peer connectivity analysis
- Comprehensive report generation

#### Usage

```bash
# Make executable
chmod +x consensus-analyzer.sh

# Interactive mode (menu-driven)
./consensus-analyzer.sh

# Command-line mode
./consensus-analyzer.sh validators    # Analyze validator set
./consensus-analyzer.sh consensus     # Analyze consensus state
./consensus-analyzer.sh blocks        # Analyze block production
./consensus-analyzer.sh peers         # Analyze peer connectivity
./consensus-analyzer.sh report        # Generate full report
./consensus-analyzer.sh all           # Run all analyses

# Report saved to:
# consensus_analysis_YYYYMMDD_HHMMSS.md
```

#### Menu Options

1. Analyze Validator Set - Shows voting power distribution
2. Analyze Consensus State - Shows active validators and consensus status
3. Analyze Block Production - Monitors recent blocks and proposers
4. Analyze Peer Connectivity - Shows peer connections per validator
5. Generate Full Report - Creates comprehensive markdown report
6. Run All Analyses - Executes all analysis modules

### 3. validator-control.sh

**Validator management tool** - Simplifies validator operations.

#### Features

- Start/stop/restart individual validators or all validators
- Real-time status monitoring
- Live block production monitoring
- Log tailing

#### Usage

```bash
# Make executable
chmod +x validator-control.sh

# Status commands
./validator-control.sh status         # Full status display
./validator-control.sh quick          # Compact status
./validator-control.sh monitor        # Live monitoring (Ctrl+C to stop)

# Control commands
./validator-control.sh start 1        # Start validator 1
./validator-control.sh start all      # Start all validators
./validator-control.sh stop 2         # Stop validator 2
./validator-control.sh stop all       # Stop all validators
./validator-control.sh restart 3      # Restart validator 3
./validator-control.sh restart all    # Restart all validators

# Logs
./validator-control.sh logs 1         # Tail validator 1 logs
```

#### Monitor Mode

Live monitoring displays:
```
V1: ● H:1234 P:3    # Validator 1: Running, Height 1234, 3 Peers
V2: ● H:1234 P:3    # Validator 2: Running, Height 1234, 3 Peers
V3: ● H:1234 P:3    # Validator 3: Running, Height 1234, 3 Peers
V4: ● STOPPED       # Validator 4: Not running

Consensus: ACTIVE (75%)
```

## Typical Test Workflow

### 1. Initial Setup

```bash
# Source environment
cd /home/hudson/blockchain-projects/aura
source env.sh

# Verify binary
ls -l chain/aurad

# Check validator homes exist
ls -ld ~/.aura/validator{1..4}
```

### 2. Start Network

```bash
cd chain/testing/local/phase3

# Start all validators
./validator-control.sh start all

# Verify status
./validator-control.sh status
```

### 3. Run Consensus Tests

```bash
# Run full test suite
./4-validator-consensus-test.sh

# Monitor results in real-time
# Tests will automatically stop/start validators
# and measure consensus behavior
```

### 4. Analyze Results

```bash
# Generate analysis report
./consensus-analyzer.sh report

# Or run interactive analysis
./consensus-analyzer.sh
# Select option 6 (Run All Analyses)
```

### 5. Manual Testing

```bash
# Monitor in one terminal
./validator-control.sh monitor

# Control validators in another terminal
./validator-control.sh stop 4    # Stop validator 4
# Observe: Consensus remains active (75% > 66%)

./validator-control.sh stop 3    # Stop validator 3
# Observe: Consensus HALTS (50% < 66%)

./validator-control.sh start all # Restart all
# Observe: Consensus resumes
```

## Validator Configuration

### Validator Homes
- Validator 1: `/home/hudson/.aura/validator1`
- Validator 2: `/home/hudson/.aura/validator2`
- Validator 3: `/home/hudson/.aura/validator3`
- Validator 4: `/home/hudson/.aura/validator4`

### Port Assignments

| Validator | RPC Port | P2P Port | gRPC Port |
|-----------|----------|----------|-----------|
| 1 | 26657 | 26656 | 9090 |
| 2 | 26667 | 26666 | 9092 |
| 3 | 26677 | 26676 | 9093 |
| 4 | 26687 | 26686 | 9094 |

### RPC Endpoints
- Validator 1: `http://localhost:26657`
- Validator 2: `http://localhost:26667`
- Validator 3: `http://localhost:26677`
- Validator 4: `http://localhost:26687`

## Understanding Test Results

### Successful Test Output

```
========================================
TEST 1: All 4 validators running (100% voting power)
========================================

[TEST] Active Validators: 4/4
[TEST] Active Voting Power: 100%
[TEST] Expected Result: PRODUCE_BLOCKS

[INFO] Network Status:
  Validator 1: RUNNING (height: 1234)
  Validator 2: RUNNING (height: 1234)
  Validator 3: RUNNING (height: 1234)
  Validator 4: RUNNING (height: 1234)
  Active Validators: 4/4
  Active Voting Power: 100%
  Consensus Status: CAN PRODUCE BLOCKS (>66% voting power)

[INFO] Monitoring block production for 15s...
  Start Height: 1234
  End Height: 1237
  Blocks Produced: 3 in 15s
  Block Rate: 0.20 blocks/sec
  [✓] Chain is producing blocks

========================================
TEST 1 VERIFICATION
========================================
[✓] PASS: Chain produced blocks as expected with 100% voting power
```

### Failed Test Output

```
[✗] FAIL: Chain produced blocks unexpectedly with 50% voting power
```

### Test Summary

```
========================================
TEST SUITE SUMMARY
========================================
Completed: 2025-12-13 12:34:56

Results:
  [✓] TEST 1: PASS
  [✓] TEST 2: PASS
  [✓] TEST 3: PASS
  [✓] TEST 4: PASS
  [✓] TEST 5: PASS

[INFO] Total: 5/5 tests passed

[✓] ALL CONSENSUS TESTS PASSED

Results saved to: consensus_test_results_20251213_123456.log
```

## Troubleshooting

### Validators Won't Start

```bash
# Check if ports are already in use
netstat -tlnp | grep -E '26657|26667|26677|26687'

# Kill existing processes
pkill -f "aurad start"

# Retry
./validator-control.sh start all
```

### Validators Not Connecting

```bash
# Check P2P configuration in each validator home
cat ~/.aura/validator1/config/config.toml | grep persistent_peers

# Verify all validators are on same chain ID
grep chain_id ~/.aura/validator*/config/genesis.json

# Check network connectivity
curl http://localhost:26657/net_info | jq
```

### Chain Not Producing Blocks

```bash
# Verify consensus threshold
./consensus-analyzer.sh consensus

# Should show >66% active voting power
# If <66%, start more validators:
./validator-control.sh start all
```

### Logs Not Showing

```bash
# Check log file exists
ls -l /home/hudson/blockchain-projects/aura/chain/testing/local/phase3/validator*.log

# If missing, validators may not have started
./validator-control.sh status

# Restart and check logs immediately
./validator-control.sh restart 1
./validator-control.sh logs 1
```

## Advanced Testing Scenarios

### 1. Network Partition Simulation

```bash
# Simulate network split (2 validators each side)
./validator-control.sh stop 3
./validator-control.sh stop 4

# Both sides have 50% voting power - both should HALT
./consensus-analyzer.sh consensus

# Resolve partition
./validator-control.sh start all
```

### 2. Rolling Restart

```bash
# Restart validators one at a time
for i in 1 2 3 4; do
    ./validator-control.sh restart $i
    sleep 10
done

# Verify all validators synced
./consensus-analyzer.sh validators
```

### 3. Validator Failure Recovery

```bash
# Kill validator ungracefully
pkill -9 -f "validator2"

# Wait and observe consensus maintained
./validator-control.sh monitor

# Restart failed validator
./validator-control.sh start 2

# Verify sync
./consensus-analyzer.sh blocks
```

## Results and Reports

### Log Files

All test executions generate timestamped logs:
- `consensus_test_results_YYYYMMDD_HHMMSS.log` - Test suite results
- `consensus_analysis_YYYYMMDD_HHMMSS.md` - Analysis reports
- `validator1.log` to `validator4.log` - Individual validator logs

### Report Contents

Analysis reports include:
1. Executive Summary
   - Active validator count
   - Voting power percentage
   - Consensus status
   - Current block height

2. Validator Details
   - Status (running/stopped)
   - Block height
   - Peer connections
   - Validator set membership

3. Recent Blocks
   - Block heights
   - Proposer addresses
   - Timestamps

4. BFT Consensus Properties
   - Voting power distribution
   - Consensus thresholds
   - Expected behavior matrix

## Integration with Other Tests

These Phase 3 tests complement:
- **Phase 1**: Single-node testing
- **Phase 2**: Multi-node setup and genesis workflow
- **Phase 4**: (Future) IBC and cross-chain testing

## References

- [Tendermint BFT Consensus](https://docs.tendermint.com/master/introduction/what-is-tendermint.html#consensus-overview)
- [Cosmos SDK Validators](https://docs.cosmos.network/main/modules/staking)
- Aura Project Documentation: `/home/hudson/blockchain-projects/aura/README.md`

## Success Criteria

Tests are considered successful when:
1. All 5 consensus tests pass
2. Chain produces blocks with ≥3 validators (≥75% voting power)
3. Chain halts with ≤2 validators (≤50% voting power)
4. Validators can restart and resync after failures
5. All validators maintain synchronized block heights (within 2 blocks)

## Next Steps

After successful Phase 3 testing:
1. Document findings in comprehensive report
2. Capture any anomalies or unexpected behavior
3. Generate performance metrics (block times, sync times)
4. Prepare for Phase 4 (IBC and cross-chain testing)
5. Archive test results for audit trail

## Support

For issues or questions:
- Check logs: `./validator-control.sh logs <1-4>`
- Run diagnostics: `./consensus-analyzer.sh all`
- Review validator config: `cat ~/.aura/validator1/config/config.toml`
- Check genesis: `cat ~/.aura/validator1/config/genesis.json`
