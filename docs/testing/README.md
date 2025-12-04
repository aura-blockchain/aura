# AURA Testing Documentation

This directory contains comprehensive testing guides and tools for the AURA blockchain.

## Byzantine Fault Tolerance (BFT) Testing

The BFT test suite validates that the AURA blockchain correctly implements the Tendermint consensus algorithm's fault tolerance properties.

### Quick Start

```bash
cd /home/decri/blockchain-projects/aura

# 1. Start the testnet
docker-compose -f docker-compose.testnet.yml up -d

# 2. Wait for sync
sleep 60

# 3. Run the BFT test
./scripts/test-bft-comprehensive.sh --verbose

# 4. Analyze results
python3 scripts/analyze-bft-results.py bft_test_*.log
```

### BFT Testing Files

#### Main Test Script
- **`/scripts/test-bft-comprehensive.sh`** - Complete BFT test suite
  - Tests 3/4 consensus (should work)
  - Tests 2/4 consensus (should halt)
  - Tests validator catch-up after restart
  - Tests chain recovery

#### Result Analysis Tool
- **`/scripts/analyze-bft-results.py`** - Analyzes test outputs
  - Extracts metrics from log files
  - Generates text/JSON/CSV reports
  - Validates BFT properties

#### Documentation

1. **BFT_TEST_GUIDE.md** - Complete testing guide
   - Prerequisites and setup
   - Detailed test scenario descriptions
   - Troubleshooting guide
   - Performance expectations
   - Advanced testing techniques

2. **BFT_TEST_EXPECTED_RESULTS.md** - Expected outcomes
   - Detailed results for each test phase
   - Metrics reference table
   - Consensus state progression
   - Analysis examples
   - Critical failure indicators

3. **BFT_QUICK_REFERENCE.md** - Quick reference card
   - One-liner quick start
   - Essential commands
   - Port mappings
   - Common issues
   - Useful aliases

4. **README.md** - This file

### Test Scenarios

The BFT test validates the blockchain's behavior under different validator availability scenarios:

#### Phase 1: Baseline Recording
- **Active Validators**: 4/4
- **Expected**: All validators at same height
- **Duration**: 2 minutes

#### Phase 2: 3/4 Consensus Test
- **Active Validators**: 3/4 (validator-3 stopped)
- **Expected**: Chain continues producing blocks
- **Duration**: 5 minutes
- **BFT Threshold**: 2/3 of 4 = 2.67 (rounds to 3), so 3 ≥ 3 ✓

#### Phase 3: Validator Catch-up
- **Active Validators**: 4/4 (validator-3 restarted)
- **Expected**: validator-3 syncs to current height
- **Duration**: 5 minutes
- **Mechanism**: State sync (fast block replay)

#### Phase 4: 2/4 Consensus Halt
- **Active Validators**: 2/4 (validators 2 and 3 stopped)
- **Expected**: Chain halts - NO blocks produced
- **Duration**: 5 minutes
- **BFT Threshold**: 2/3 of 4 = 2.67, so 2 < 2.67 ✗
- **CRITICAL**: If blocks ARE produced here, it's a BFT bug!

#### Phase 5: Chain Recovery
- **Active Validators**: 4/4 (all restarted)
- **Expected**: Chain resumes, all validators sync
- **Duration**: 5 minutes

### Success Criteria

```
✓ Phase 1: Baseline heights identical
✓ Phase 2: Blocks produced with 3/4 validators
✓ Phase 3: Validator-3 catches up
✓ Phase 4: NO blocks produced with 2/4 validators  [CRITICAL]
✓ Phase 5: Chain resumes with 4/4 validators
```

### Critical Test: Phase 4 (2/4 Halt)

This is the most important test. If blocks are produced with only 2/4 validators, this indicates a **critical BFT consensus bug**.

Expected: **0 blocks produced**
Failure: **> 0 blocks produced** = CRITICAL BUG

### Test Output Files

The test generates two report files:

1. **`bft_test_YYYYMMDD_HHMMSS.log`**
   - Timestamped events
   - Block heights at each step
   - Diagnostic information
   - Human-readable format

2. **`bft_test_YYYYMMDD_HHMMSS.json`**
   - Machine-readable metrics
   - Structured results
   - Suitable for parsing/integration

### Analyzing Results

```bash
# Generate text report
python3 scripts/analyze-bft-results.py bft_test_20241203_152345.log

# Generate JSON report
python3 scripts/analyze-bft-results.py bft_test_20241203_152345.log --output-format json

# Export metrics as CSV
python3 scripts/analyze-bft-results.py bft_test_20241203_152345.log --export-csv

# Save to file
python3 scripts/analyze-bft-results.py bft_test_20241203_152345.log --output analysis.txt
```

### Monitoring During Test

Run these commands in separate terminals:

```bash
# Terminal 1: Run the test
./scripts/test-bft-comprehensive.sh --verbose

# Terminal 2: Watch block heights
watch -n 1 'for port in 26657 26757 26857 26957; do
  echo -n "Port $port: "
  curl -s http://localhost:$port/status 2>/dev/null | \
    jq -r ".result.sync_info.latest_block_height"
done'

# Terminal 3: Watch consensus rounds
watch -n 1 'curl -s http://localhost:26657/consensus_state 2>/dev/null | \
  jq ".result.round_state | {height, round, step}"'

# Terminal 4: View logs
docker logs -f aura-validator-1 | grep -i "consensus\|prevote\|precommit"
```

### Port Reference

| Validator | RPC | REST API | gRPC | P2P | Metrics |
|-----------|-----|----------|------|-----|---------|
| validator-1 | 26657 | 1317 | 9090 | 26656 | 26660 |
| validator-2 | 26757 | 1417 | 9190 | 26756 | 26760 |
| validator-3 | 26857 | 1517 | 9290 | 26856 | 26860 |
| validator-4 | 26957 | 1617 | 9390 | 26956 | 26960 |

(Note: Above are internal port numbers. Local ports are offset, e.g., 27656 for validator-1 P2P)

### Common Issues

#### Test fails at prerequisites check
- Testnet not running: `docker-compose -f docker-compose.testnet.yml up -d`
- Check Docker: `docker ps | grep aura`

#### RPC not responding
- Check if containers are running: `docker ps`
- Test RPC: `curl -s http://localhost:26657/status`
- Check logs: `docker logs aura-validator-1`

#### Baseline heights inconsistent
- Validators need more time to sync
- Wait longer: `sleep 120` and rerun test
- Or restart testnet: `docker-compose -f docker-compose.testnet.yml restart`

#### Phase 4 (2/4) shows blocks produced
- **CRITICAL BFT BUG** - Investigate immediately!
- Check validator logs: `docker logs aura-validator-1 | grep -i consensus`
- Check voting power: `curl -s http://localhost:26657/validators | jq '.result.validators'`

### Performance Expectations

```
Block time:             ~5 seconds
Over 30 seconds:        ~6 blocks produced
Over 10 seconds:        ~2 blocks produced
3/4 validators:         Normal block time
2/4 validators:         Immediate halt (0 blocks)
```

### Test Duration

```
Prerequisites:          30 seconds
Baseline:              30 seconds
3/4 consensus:       5 minutes
Catch-up:            5 minutes
2/4 halt:            5 minutes
Recovery:            5 minutes
─────────────────────
Total:            20-30 minutes
```

## Running BFT Tests Multiple Times

For regression testing or continuous monitoring:

```bash
#!/bin/bash
for i in {1..10}; do
  echo "Run $i/10..."
  ./scripts/test-bft-comprehensive.sh --output-dir ./run_$i

  # Analyze
  python3 scripts/analyze-bft-results.py ./run_$i/bft_test_*.log

  # Wait between runs
  sleep 60
done
```

## Integration with CI/CD

To integrate into automated testing pipelines:

```bash
#!/bin/bash
set -e

# Start testnet
docker-compose -f docker-compose.testnet.yml up -d
sleep 60

# Run test
./scripts/test-bft-comprehensive.sh

# Analyze and check for failures
python3 scripts/analyze-bft-results.py bft_test_*.log > analysis.txt

# Fail if critical issues found
if grep -q "2_OF_4_BLOCKS_PRODUCED=0" analysis.txt; then
  echo "PASS: 2/4 halt test passed"
else
  echo "FAIL: 2/4 halt test did not produce expected results"
  exit 1
fi
```

## Security Properties Validated

This BFT test validates:

1. **Safety**: With < 1/3 validators Byzantine, no conflicting commits
   - Tested via: No consensus failures with 3/4 validators
   - Critical: 2/4 validators cannot create blocks

2. **Liveness**: With < 2/3 validators down, consensus continues
   - Tested via: 3/4 scenario continues producing blocks
   - Fails: 2/4 scenario halts (as expected)

3. **Byzantine Tolerance**: Chain tolerates up to 1/3 malicious validators
   - Tested via: Recovery and catch-up scenarios

## What This Test Does NOT Validate

- Byzantine validator behavior (double-signing)
- Fork detection and recovery
- Time synchronization attacks
- Mempool fork handling
- Light client verification

These require additional testing or formal verification.

## References

- [Tendermint Consensus](https://docs.tendermint.com/master/introduction/)
- [Cosmos SDK](https://docs.cosmos.network/)
- [Byzantine Fault Tolerance](https://en.wikipedia.org/wiki/Byzantine_fault_tolerance)
- [AURA Chain Repository](https://github.com/aura-nw/aura)

## Support

For questions or issues:

1. Check the appropriate documentation file above
2. Review validator logs: `docker logs aura-validator-N`
3. Test RPC connectivity: `curl -s http://localhost:XXXX/status | jq '.'`
4. Read AURA project documentation

## Files in This Directory

```
.
├── README.md                          # This file
├── BFT_TEST_GUIDE.md                 # Comprehensive testing guide
├── BFT_TEST_EXPECTED_RESULTS.md      # Detailed expected outcomes
└── BFT_QUICK_REFERENCE.md            # Quick reference card
```

## Scripts in Parent Directory

```
/scripts/
├── test-bft-comprehensive.sh         # Main BFT test script
├── analyze-bft-results.py            # Result analysis tool
├── testnet-manage.sh                 # Testnet management utility
└── testnet-init.sh                   # Testnet initialization
```

---

**Last Updated**: December 3, 2024
**Status**: Production Ready
