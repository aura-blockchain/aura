# AURA BFT Test - Expected Results and Analysis

## Summary

This document describes the expected outcomes of running the AURA Byzantine Fault Tolerance comprehensive test suite on a healthy 4-validator testnet.

## Test Execution Timeline

The complete test suite takes approximately **20-30 minutes** to run:

```
Phase Duration      Component
-----  ----------  -----------------------------------------------
1      2 min       Baseline recording and prerequisites check
2      5 min       3/4 consensus test (validator-3 stopped)
3      5 min       Validator catch-up (validator-3 restarted)
4      5 min       2/4 consensus halt test (validator-2,3 stopped)
5      5 min       Chain recovery (all validators restarted)
-      30 min      Total
```

## Expected Metrics Output

### Phase 1: Baseline Recording

When the test starts with all 4 validators running and synced:

```
================================================================================
STEP: Step 1: Record Current Block Heights (Baseline)
================================================================================

[2024-12-03 15:23:45] [INFO] Waiting for validators to sync (30s)...
[2024-12-03 15:24:15] [INFO] val1: Block height = 12345
[2024-12-03 15:24:15] [INFO] val2: Block height = 12345
[2024-12-03 15:24:15] [INFO] val3: Block height = 12345
[2024-12-03 15:24:15] [INFO] val4: Block height = 12345
[2024-12-03 15:24:15] [SUCCESS] Baseline recorded

BASELINE_VAL1=12345
BASELINE_VAL2=12345
BASELINE_VAL3=12345
BASELINE_VAL4=12345
```

**Expected characteristics**:
- All 4 validators have **identical or within 1 block** of each other
- Block heights are **12000+** (depends on how long testnet has been running)
- All validators return within 1-2 seconds of query
- Status shows `catching_up: false` for all validators

### Phase 2: 3/4 Consensus Test

After stopping validator-3 and waiting 30 seconds:

```
================================================================================
STEP: Step 2-4: Test 3/4 Validators Consensus
================================================================================

[2024-12-03 15:24:20] [INFO] Stopping validator-3 (aura-validator-3)...
[2024-12-03 15:24:22] [SUCCESS] aura-validator-3 stopped
[2024-12-03 15:24:22] [INFO] Verifying validator-3 is stopped...
[2024-12-03 15:24:22] [SUCCESS] validator-3 is stopped

[2024-12-03 15:24:52] [INFO] Checking block production on remaining validators (3/4)...
[2024-12-03 15:25:02] [INFO] val1: 12345 → 12347 (2 blocks produced)
[2024-12-03 15:25:02] [SUCCESS] Block production active on val1
[2024-12-03 15:25:12] [INFO] val2: 12345 → 12347 (2 blocks produced)
[2024-12-03 15:25:12] [SUCCESS] Block production active on val2
[2024-12-03 15:25:22] [INFO] val4: 12345 → 12347 (2 blocks produced)
[2024-12-03 15:25:22] [SUCCESS] Block production active on val4

3_OF_4_val1=12360
3_OF_4_val2=12360
3_OF_4_val4=12360
```

**Expected characteristics**:
- Validators 1, 2, 4 **continue producing blocks**
- Block heights increase by **2-4 blocks per 10-second interval**
- All 3 remaining validators have **consistent block heights**
- No consensus errors or timeouts in validator logs
- **This is a PASS condition** - demonstrates 2/3 consensus works

### Phase 3: Validator Catch-up

After restarting validator-3:

```
================================================================================
STEP: Step 5-6: Test Validator Catch-up via State Sync
================================================================================

[2024-12-03 15:25:25] [INFO] Recording current heights of active validators...
[2024-12-03 15:25:25] [INFO] val1: height = 12360
[2024-12-03 15:25:25] [INFO] val2: height = 12360
[2024-12-03 15:25:25] [INFO] val4: height = 12360

[2024-12-03 15:25:30] [INFO] Restarting validator-3...
[2024-12-03 15:25:32] [SUCCESS] aura-validator-3 started
[2024-12-03 15:25:32] [INFO] Waiting for RPC to be available...
[2024-12-03 15:25:35] [SUCCESS] RPC for val3 is accessible

[2024-12-03 15:25:35] [INFO] Waiting for state sync to complete (120s timeout)...
[2024-12-03 15:25:38] [INFO] .....
[2024-12-03 15:26:25] [SUCCESS] validator-3 has synced

[2024-12-03 15:26:25] [INFO] validator-3 final height: 12360
[2024-12-03 15:26:25] [INFO] validator-1 current height: 12360
[2024-12-03 15:26:25] [SUCCESS] validator-3 has caught up (within 5 blocks)

VAL3_CATCH_UP_HEIGHT=12360
VAL1_HEIGHT_AT_RESTART=12360
CATCH_UP_DIFF=0
```

**Expected characteristics**:
- RPC becomes available within **5-10 seconds**
- State sync completes within **30-90 seconds** (depends on block volume)
- Final height difference is **0-5 blocks** (acceptable window)
- `catching_up` field transitions from `true` to `false`
- **This is a PASS condition** - demonstrates state sync recovery

**Alternative acceptable outcome** (slower catch-up):
```
[2024-12-03 15:26:40] [INFO] validator-3 final height: 12365
[2024-12-03 15:26:40] [INFO] validator-1 current height: 12365
[2024-12-03 15:26:40] [SUCCESS] validator-3 has caught up (within 5 blocks)

VAL3_CATCH_UP_HEIGHT=12365
VAL1_HEIGHT_AT_RESTART=12367
CATCH_UP_DIFF=2
```

### Phase 4: 2/4 Consensus Halt

After stopping validators 2 and 3:

```
================================================================================
STEP: Step 7-8: Test 2/4 Validators (Consensus Halted)
================================================================================

[2024-12-03 15:26:50] [INFO] This scenario tests that consensus halts with 2/4 validators

[2024-12-03 15:26:50] [INFO] Stopping validator-2 (aura-validator-2)...
[2024-12-03 15:26:52] [SUCCESS] aura-validator-2 stopped
[2024-12-03 15:26:52] [INFO] Stopping validator-3 (aura-validator-3)...
[2024-12-03 15:26:54] [SUCCESS] aura-validator-3 stopped
[2024-12-03 15:26:54] [INFO] Verifying validator-2 and validator-3 are stopped...
[2024-12-03 15:26:54] [SUCCESS] Both validator-2 and validator-3 are stopped

[2024-12-03 15:27:24] [INFO] Checking block production on remaining validators (2/4)...
[2024-12-03 15:27:34] [INFO] val1: 12365 → 12365 (0 blocks produced)
[2024-12-03 15:27:34] [WARNING] No blocks produced in 10s
[2024-12-03 15:27:44] [INFO] val4: 12365 → 12365 (0 blocks produced)
[2024-12-03 15:27:44] [WARNING] No blocks produced in 10s

[2024-12-03 15:27:44] [SUCCESS] Chain has halted as expected (no blocks produced)

2_OF_4_BLOCKS_PRODUCED=0
```

**Expected characteristics**:
- Block height **does not increase** on validators 1 and 4
- No new blocks are added to the chain
- Consensus timeout occurs (typically ~30 seconds after node falls below 2/3)
- Validator logs show consensus waiting for votes
- **This is a PASS condition** - demonstrates BFT threshold enforcement

**Critical behavior to verify**:

The test checks that with only 2 out of 4 validators:
- Quorum = 2/3 of validator set = ~1.33 validators (rounds up to 2)
- With only 2/4 validators: can achieve at most 50% agreement
- 50% < 66.7% required consensus
- Therefore: **0 blocks should be produced**

If blocks ARE produced with 2/4 validators, this indicates a **critical BFT consensus bug**.

### Phase 5: Chain Recovery

After restarting all validators:

```
================================================================================
STEP: Step 9-10: Test Chain Recovery
================================================================================

[2024-12-03 15:27:50] [INFO] Restarting all validators...
[2024-12-03 15:27:52] [INFO] Starting aura-validator-2...
[2024-12-03 15:27:54] [SUCCESS] aura-validator-2 started
[2024-12-03 15:27:54] [INFO] Starting aura-validator-3...
[2024-12-03 15:27:56] [SUCCESS] aura-validator-3 started

[2024-12-03 15:27:56] [INFO] Verifying all validators are running...
[2024-12-03 15:27:56] [SUCCESS] All validators are running

[2024-12-03 15:27:56] [INFO] Waiting for validators to reconnect to RPC...
[2024-12-03 15:27:59] [SUCCESS] RPC for val2 is accessible
[2024-12-03 15:28:02] [SUCCESS] RPC for val3 is accessible

[2024-12-03 15:28:32] [INFO] Checking block production after recovery...
[2024-12-03 15:28:42] [INFO] val1: 12365 → 12367 (2 blocks produced)
[2024-12-03 15:28:42] [SUCCESS] Block production active on val1
[2024-12-03 15:28:52] [INFO] val2: 12365 → 12367 (2 blocks produced)
[2024-12-03 15:28:52] [SUCCESS] Block production active on val2
[2024-12-03 15:29:02] [INFO] val3: 12360 → 12367 (7 blocks produced from validator's perspective)
[2024-12-03 15:29:02] [SUCCESS] Block production active on val3
[2024-12-03 15:29:12] [INFO] val4: 12365 → 12367 (2 blocks produced)
[2024-12-03 15:29:12] [SUCCESS] Block production active on val4

RECOVERY_BLOCK_PRODUCTION=4
FINAL_HEIGHT_val1=12378
FINAL_HEIGHT_val2=12378
FINAL_HEIGHT_val3=12378
FINAL_HEIGHT_val4=12378
```

**Expected characteristics**:
- RPC becomes available on all validators within **10-20 seconds**
- Block production **resumes** within **30-60 seconds** of restart
- All 4 validators quickly **converge to same block height**
- Final heights are **identical or within 1 block**
- No consensus errors or fork detection triggered
- **This is a PASS condition** - demonstrates automatic recovery

## Overall Test Result Summary

### PASS Scenario (Healthy Blockchain)

```
================================================================================
TEST SUMMARY
================================================================================

✓ BASELINE_SYNC: PASS
  Height variance: 0 blocks

✓ THREE_OF_FOUR: PASS
  Blocks produced: 15

✓ CATCH_UP: PASS
  Catch-up difference: 2 blocks

✓ TWO_OF_FOUR: PASS
  Blocks produced: 0

✓ RECOVERY: PASS
  Validators producing blocks: 4/4

RESULT: PASSED - All BFT tests completed successfully
================================================================================
```

### FAIL Scenario (BFT Bug)

If blocks are produced with only 2/4 validators:

```
✗ TWO_OF_FOUR: CRITICAL
  Blocks produced: 3
  *** CRITICAL: Chain continued with 2/4 validators! ***

RESULT: FAILED - CRITICAL BFT BUG DETECTED
The chain continued producing blocks with only 2/4 validators!
This violates the 2/3 consensus threshold requirement.
```

## Metrics Reference Table

### Key Metrics Collected

| Metric | Description | Expected Value |
|--------|-------------|-----------------|
| `BASELINE_VAL1/2/3/4` | Initial block height | Same (±1) |
| `3_OF_4_val1/2/4` | Height after 30s with 3 validators | +15 to +30 |
| `VAL3_CATCH_UP_HEIGHT` | validator-3 height after restart | Within 5 of current |
| `VAL1_HEIGHT_AT_RESTART` | validator-1 height when restarted | Current height |
| `CATCH_UP_DIFF` | Difference between val3 and val1 | 0-5 blocks |
| `2_OF_4_BLOCKS_PRODUCED` | Blocks produced with 2/4 validators | **0 (critical)** |
| `RECOVERY_BLOCK_PRODUCTION` | Validators producing after restart | 4 |
| `FINAL_HEIGHT_val1/2/3/4` | Final block height | All identical |

## Block Time Expectations

AURA Testnet Configuration:
- **Block time**: ~5 seconds average
- **Timeout propose**: 3 seconds
- **Timeout prevote**: 1 second
- **Timeout precommit**: 1 second
- **Timeout commit**: 1 second

**Therefore**:
- Over 30 seconds: expect ~6 blocks produced
- Over 10 seconds: expect ~2 blocks produced
- With 3/4 validators: normal block time (5s)
- With 2/4 validators: consensus halts immediately

## Consensus State Progression

### Normal Consensus (4/4 validators)

```
Heights:    [12365, 12365, 12365, 12365]
Rounds:     [0, 0, 0, 0]
Prevotes:   [✓, ✓, ✓, ✓]  (4/4 validators)
Precommits: [✓, ✓, ✓, ✓]  (4/4 validators)
Status:     COMMIT → New block → Increment height
```

### 3/4 Consensus (validator-3 down)

```
Heights:    [12365, 12365, -, 12365]
Rounds:     [0, 0, -, 0]
Prevotes:   [✓, ✓, -, ✓]  (3/3 live validators)
Precommits: [✓, ✓, -, ✓]  (3/3 live validators)
Status:     COMMIT → New block → Increment height
```

### 2/4 Consensus (validators 2,3 down)

```
Heights:    [12365, -, -, 12365]
Rounds:     [0, -, -, 0]
Prevotes:   [✓, -, -, ✓]  (2/4 validators < 2/3 required)
Precommits: [✗, -, -, ✗]  (Cannot reach 2/3)
Status:     WAITING → Timeout → Repeat → STALLED
```

## Analysis Using the Result Analyzer

After running the test, analyze results:

```bash
# Generate text report
python3 scripts/analyze-bft-results.py bft_test_20241203_152345.log

# Generate JSON report
python3 scripts/analyze-bft-results.py bft_test_20241203_152345.log --output-format json

# Export metrics as CSV
python3 scripts/analyze-bft-results.py bft_test_20241203_152345.log --export-csv

# Save detailed report
python3 scripts/analyze-bft-results.py bft_test_20241203_152345.log --output analysis_report.txt
```

## Performance Benchmarks

### Expected Durations

| Phase | Target | Acceptable Range | Notes |
|-------|--------|------------------|-------|
| Prerequisites check | 10s | 5-15s | RPC queries |
| Baseline recording | 30s | 25-35s | Sync wait |
| 3/4 consensus test | 5m | 4-6m | Includes wait times |
| Catch-up test | 5m | 4-8m | State sync dependent |
| 2/4 halt test | 5m | 4-6m | Fixed wait times |
| Recovery test | 5m | 4-6m | Restart + sync |
| **Total** | **20-30m** | **18-35m** | Full suite |

### Block Production Rates

| Scenario | Expected Rate | Min Acceptable |
|----------|---------------|-----------------|
| 4/4 validators | 0.2 blocks/sec | 0.15 blocks/sec |
| 3/4 validators | 0.2 blocks/sec | 0.15 blocks/sec |
| 2/4 validators | 0 blocks | 0 blocks (critical) |

## Common Test Patterns

### Quick Test (5 minutes)

Run only 3/4 consensus test:
```bash
# Manually run Phase 2-4 only
docker stop aura-validator-3
sleep 30
# Check block production
curl -s http://localhost:26657/status | jq '.result.sync_info.latest_block_height'
docker start aura-validator-3
```

### Stress Test (Extended)

Run multiple iterations with load:
```bash
for i in {1..5}; do
  ./scripts/test-bft-comprehensive.sh --output-dir ./run_$i
  sleep 60  # Cool down
done
```

### Continuous Monitoring

Monitor consensus state during test:
```bash
# Terminal 1: Run test
./scripts/test-bft-comprehensive.sh --verbose

# Terminal 2: Watch consensus state
watch -n 1 'curl -s http://localhost:26657/consensus_state | jq .result.round_state'
```

## Validation Checklist

After test completes, verify:

- [ ] All 4 phases completed without error
- [ ] Baseline heights are consistent (±1 block)
- [ ] 3/4 scenario produced blocks (> 0)
- [ ] 2/4 scenario produced NO blocks (= 0)
- [ ] Validator-3 caught up after restart
- [ ] Final heights are consistent (±1 block)
- [ ] No BFT consensus bugs detected
- [ ] Test duration within expected range
- [ ] Report files generated successfully
- [ ] No data corruption or state inconsistency

## Troubleshooting Failed Tests

### Symptom: "RPC not available"

**Cause**: Testnet not running or port mapping wrong

**Fix**:
```bash
docker ps | grep aura
docker-compose -f docker-compose.testnet.yml up -d
```

### Symptom: Inconsistent baseline heights (>5 block variance)

**Cause**: Validators not in sync before test

**Fix**:
```bash
# Wait longer for sync
sleep 60
# Or restart validators
docker-compose -f docker-compose.testnet.yml restart
```

### Symptom: 3/4 test shows 0 blocks produced

**Cause**: Chain may have halted or validators not communicating

**Fix**:
```bash
# Check validator logs
docker logs aura-validator-1 | tail -50

# Check network connectivity
docker exec aura-validator-1 netstat -an | grep 26656

# Restart testnet
docker-compose -f docker-compose.testnet.yml restart
```

### Symptom: 2/4 halt test shows blocks still produced

**This is a CRITICAL BFT BUG - investigate immediately!**

**Investigation**:
```bash
# Check validator voting power
curl -s http://localhost:26657/validators | jq '.result.validators[].voting_power'

# Check consensus state
curl -s http://localhost:26657/consensus_state | jq '.result.round_state'

# Review logs
docker logs aura-validator-1 | grep -i "consensus\|voting"
```

## References

- [Tendermint BFT Consensus](https://docs.tendermint.com/master/introduction/)
- [Cosmos SDK Consensus](https://docs.cosmos.network/main/intro/why-cosmos.html)
- [Byzantine Fault Tolerance](https://en.wikipedia.org/wiki/Byzantine_fault_tolerance)
