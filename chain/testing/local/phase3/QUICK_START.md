# Phase 3: Quick Start Guide

## Prerequisites Check

```bash
# 1. Navigate to project
cd /home/hudson/blockchain-projects/aura

# 2. Source environment
source env.sh

# 3. Verify binary exists
ls -l chain/aurad
# Should show: -rwxr-xr-x ... chain/aurad

# 4. Verify validators initialized
ls -ld ~/.aura/validator{1..4}
# Should show 4 directories
```

## Option 1: Automated Testing (Recommended)

**Run the complete test suite automatically:**

```bash
cd chain/testing/local/phase3

# Execute full consensus test suite
./4-validator-consensus-test.sh

# Wait for completion (approximately 3-5 minutes)
# Results will be saved to: consensus_test_results_YYYYMMDD_HHMMSS.log
```

**What it does:**
1. Verifies all 4 validators are set up
2. Tests consensus with 4 validators (100% voting power) → Should produce blocks
3. Tests consensus with 3 validators (75% voting power) → Should produce blocks
4. Tests consensus with 2 validators (50% voting power) → Should HALT
5. Tests consensus with 1 validator (25% voting power) → Should HALT
6. Restarts all validators and verifies sync
7. Generates comprehensive report

**Expected output:**
```
========================================
TEST SUITE SUMMARY
========================================
Results:
  [✓] TEST 1: PASS
  [✓] TEST 2: PASS
  [✓] TEST 3: PASS
  [✓] TEST 4: PASS
  [✓] TEST 5: PASS

[✓] ALL CONSENSUS TESTS PASSED
```

## Option 2: Manual Testing

**Control validators manually for custom testing:**

### Step 1: Start Network

```bash
cd chain/testing/local/phase3

# Start all validators
./validator-control.sh start all

# Verify status
./validator-control.sh status
```

Expected output:
```
Validator 1: RUNNING
  Height: 123 | Peers: 3
Validator 2: RUNNING
  Height: 123 | Peers: 3
Validator 3: RUNNING
  Height: 123 | Peers: 3
Validator 4: RUNNING
  Height: 123 | Peers: 3

Active Validators: 4/4
Active Voting Power: 100%
[✓] Consensus: ACTIVE (can produce blocks)
```

### Step 2: Test Consensus Thresholds

**Test with 3 validators (75% - should work):**
```bash
./validator-control.sh stop 4
./validator-control.sh status
# Should show: Consensus: ACTIVE (75%)
```

**Test with 2 validators (50% - should halt):**
```bash
./validator-control.sh stop 3
./validator-control.sh status
# Should show: Consensus: HALTED (50%)
```

**Restart and verify recovery:**
```bash
./validator-control.sh start all
./validator-control.sh status
# Should show: Consensus: ACTIVE (100%)
```

### Step 3: Monitor Block Production

**Real-time monitoring:**
```bash
./validator-control.sh monitor
```

Output updates every 2 seconds:
```
V1: ● H:1234 P:3
V2: ● H:1235 P:3
V3: ● H:1235 P:3
V4: ● H:1235 P:3

Consensus: ACTIVE (100%)
```

Press `Ctrl+C` to stop monitoring.

### Step 4: Analyze Results

**Generate comprehensive analysis:**
```bash
./consensus-analyzer.sh all
```

This creates a detailed markdown report with:
- Validator status
- Voting power distribution
- Block production analysis
- Peer connectivity
- BFT consensus properties

## Option 3: Interactive Analysis

**Use the interactive analyzer:**

```bash
./consensus-analyzer.sh
```

Menu:
```
1. Analyze Validator Set
2. Analyze Consensus State
3. Analyze Block Production
4. Analyze Peer Connectivity
5. Generate Full Report
6. Run All Analyses
0. Exit
```

Select options to explore different aspects of the network.

## Common Commands Reference

### Validator Control

```bash
# Start
./validator-control.sh start all       # Start all
./validator-control.sh start 1         # Start validator 1

# Stop
./validator-control.sh stop all        # Stop all
./validator-control.sh stop 2          # Stop validator 2

# Restart
./validator-control.sh restart all     # Restart all
./validator-control.sh restart 3       # Restart validator 3

# Status
./validator-control.sh status          # Full status
./validator-control.sh quick           # Compact status
./validator-control.sh monitor         # Live monitoring

# Logs
./validator-control.sh logs 1          # Tail validator 1 logs
```

### Analysis

```bash
# Quick checks
./consensus-analyzer.sh validators     # Validator set analysis
./consensus-analyzer.sh consensus      # Consensus state
./consensus-analyzer.sh blocks         # Block production
./consensus-analyzer.sh peers          # Peer connectivity

# Full report
./consensus-analyzer.sh report         # Generate markdown report
./consensus-analyzer.sh all            # Run all analyses
```

### Testing

```bash
# Full test suite
./4-validator-consensus-test.sh        # Run all consensus tests
```

## Direct RPC Access

If you want to query validators directly:

```bash
# Validator 1
curl http://localhost:26657/status | jq
curl http://localhost:26657/net_info | jq
curl http://localhost:26657/validators | jq

# Validator 2
curl http://localhost:26667/status | jq

# Validator 3
curl http://localhost:26677/status | jq

# Validator 4
curl http://localhost:26687/status | jq
```

## Troubleshooting Quick Fixes

### "Binary not found"
```bash
cd /home/hudson/blockchain-projects/aura/chain
make build
```

### "Validator home not found"
```bash
# Run the 4-validator setup script first
# (Should be provided by Phase 3 setup agent)
```

### "Port already in use"
```bash
# Kill existing processes
pkill -f "aurad start"

# Restart
./validator-control.sh start all
```

### "Validators not connecting"
```bash
# Check peer configuration
cat ~/.aura/validator1/config/config.toml | grep persistent_peers

# Restart all validators
./validator-control.sh restart all
```

### "Chain not producing blocks"
```bash
# Check consensus threshold
./validator-control.sh status

# Ensure >66% voting power (3+ validators)
./validator-control.sh start all
```

## Test Execution Checklist

- [ ] Environment sourced (`source env.sh`)
- [ ] Binary built (`chain/aurad` exists)
- [ ] 4 validator homes initialized (`~/.aura/validator1-4`)
- [ ] All scripts executable (`chmod +x *.sh`)
- [ ] No port conflicts (kill old processes if needed)
- [ ] Run automated test suite (`./4-validator-consensus-test.sh`)
- [ ] Verify all tests pass (5/5 PASS)
- [ ] Generate analysis report (`./consensus-analyzer.sh report`)
- [ ] Review results logs
- [ ] Document any anomalies

## Success Criteria

✓ Test 1 passes: 4 validators produce blocks (100% voting power)
✓ Test 2 passes: 3 validators produce blocks (75% voting power)
✓ Test 3 passes: 2 validators halt chain (50% voting power)
✓ Test 4 passes: 1 validator halts chain (25% voting power)
✓ Test 5 passes: Validators restart and sync successfully
✓ All validators synchronized (within 2 blocks)
✓ No errors in validator logs

## Next Steps After Success

1. Review generated reports:
   - `consensus_test_results_YYYYMMDD_HHMMSS.log`
   - `consensus_analysis_YYYYMMDD_HHMMSS.md`

2. Archive results for audit trail

3. Document any findings in comprehensive report

4. Prepare for Phase 4 (IBC/cross-chain testing)

## Getting Help

**View detailed documentation:**
```bash
cat README.md
```

**Check script help:**
```bash
./validator-control.sh --help
```

**View recent logs:**
```bash
./validator-control.sh logs 1
```

**Interactive analysis:**
```bash
./consensus-analyzer.sh
```
