# AURA BFT Test - Quick Reference Card

## One-Liner Quick Start

```bash
# Start testnet, run BFT test, and analyze results
cd /home/decri/blockchain-projects/aura && \
docker-compose -f docker-compose.testnet.yml up -d && \
sleep 60 && \
./scripts/test-bft-comprehensive.sh --verbose && \
python3 scripts/analyze-bft-results.py bft_test_*.log
```

## Essential Commands

### Setup
```bash
cd /home/decri/blockchain-projects/aura
docker-compose -f docker-compose.testnet.yml up -d      # Start testnet
./scripts/testnet-manage.sh status                        # Check status
```

### Run Test
```bash
./scripts/test-bft-comprehensive.sh                       # Run test
./scripts/test-bft-comprehensive.sh --verbose             # Verbose output
./scripts/test-bft-comprehensive.sh --output-dir ./results # Custom output
```

### Analyze Results
```bash
python3 scripts/analyze-bft-results.py bft_test_*.log     # Text report
python3 scripts/analyze-bft-results.py bft_test_*.log --output-format json  # JSON
python3 scripts/analyze-bft-results.py bft_test_*.log --export-csv  # CSV export
```

### Cleanup
```bash
docker-compose -f docker-compose.testnet.yml down        # Stop testnet
docker-compose -f docker-compose.testnet.yml down -v     # Stop + remove volumes
```

## Test Phases at a Glance

| Phase | Scenario | Duration | Expected Outcome |
|-------|----------|----------|------------------|
| 1 | Baseline (4/4) | 2 min | All heights equal |
| 2 | 3/4 active | 5 min | Blocks produced (+15-30) |
| 3 | Catch-up | 5 min | Val3 within 5 blocks |
| 4 | 2/4 active | 5 min | **NO blocks (critical)** |
| 5 | Recovery (4/4) | 5 min | All synced again |

## Success Criteria Summary

```
✓ Phase 1: Baseline heights identical
✓ Phase 2: Blocks produced with 3/4 validators
✓ Phase 3: Validator-3 catches up
✓ Phase 4: NO blocks produced with 2/4 validators  ← CRITICAL
✓ Phase 5: Chain resumes with 4/4 validators

= Overall Result: PASSED ✓
```

## Key RPC Endpoints

```
validator-1: http://localhost:26657/status
validator-2: http://localhost:26757/status
validator-3: http://localhost:26857/status
validator-4: http://localhost:26957/status

Query block height:
curl -s http://localhost:26657/status | jq '.result.sync_info.latest_block_height'

Query consensus state:
curl -s http://localhost:26657/consensus_state | jq '.result.round_state'
```

## Manual Container Control

```bash
# Stop single validator
docker stop aura-validator-3

# Start single validator
docker start aura-validator-3

# View logs
docker logs -f aura-validator-1 | tail -100

# Execute command
docker exec aura-validator-1 aurad status
```

## Critical Failure Indicators

| Indicator | Meaning | Action |
|-----------|---------|--------|
| Phase 2: 0 blocks | 3/4 consensus broken | INVESTIGATE |
| Phase 3: Diff > 10 | State sync failed | CHECK LOGS |
| Phase 4: > 0 blocks | **BFT BUG** | **CRITICAL** |
| Phase 5: Heights diverge | Recovery failed | RESTART |

## Verbose Output Symbols

| Symbol | Meaning |
|--------|---------|
| `✓` | Success |
| `✗` | Failure |
| `⚠` | Warning |
| `→` | Action/Event |
| `?` | Unknown status |

## Expected Metrics (Healthy Test)

```
BASELINE_VAL1 = 12345
BASELINE_VAL2 = 12345
BASELINE_VAL3 = 12345
BASELINE_VAL4 = 12345

3_OF_4_val1 = 12360 (15 blocks in 30 sec)
3_OF_4_val2 = 12360
3_OF_4_val4 = 12360

VAL3_CATCH_UP_HEIGHT = 12360
CATCH_UP_DIFF = 0-2 blocks ✓

2_OF_4_BLOCKS_PRODUCED = 0 ✓✓✓ (CRITICAL)

RECOVERY_BLOCK_PRODUCTION = 4 ✓
FINAL_HEIGHT_val1 = 12378
FINAL_HEIGHT_val2 = 12378
FINAL_HEIGHT_val3 = 12378
FINAL_HEIGHT_val4 = 12378
```

## Timing Reference

```
Prerequisites:        30s
Baseline:            30s
3/4 consensus:      300s
Catch-up:           300s
2/4 halt:           300s
Recovery:           300s
─────────────────────
Total:             1290s ≈ 20-30 minutes
```

## File Output

```
bft_test_YYYYMMDD_HHMMSS.log    # Detailed log (timestamps + events)
bft_test_YYYYMMDD_HHMMSS.json   # Machine-readable metrics

Analyze with:
python3 scripts/analyze-bft-results.py <logfile>
```

## Monitoring During Test

```bash
# Terminal 1: Run test
./scripts/test-bft-comprehensive.sh --verbose

# Terminal 2: Watch block heights
watch -n 1 'for i in 26657 26757 26857 26957; do \
  echo -n "Port $i: "; \
  curl -s http://localhost:$i/status 2>/dev/null | \
  jq -r ".result.sync_info.latest_block_height"; \
done'

# Terminal 3: Watch consensus rounds
watch -n 1 'curl -s http://localhost:26657/consensus_state 2>/dev/null | \
  jq ".result.round_state | {height, round, step}"'

# Terminal 4: View logs
docker logs -f aura-validator-1 | grep -i "consensus\|prevote\|precommit"
```

## Network Topology

```
Local Machine
├── Port 26657 ──→ aura-validator-1:26657 (RPC)
├── Port 26757 ──→ aura-validator-2:26657 (RPC)
├── Port 26857 ──→ aura-validator-3:26657 (RPC)
├── Port 26957 ──→ aura-validator-4:26657 (RPC)
├── Port 27656 ──→ aura-validator-1:26656 (P2P)
├── Port 27756 ──→ aura-validator-2:26656 (P2P)
├── Port 27856 ──→ aura-validator-3:26656 (P2P)
└── Port 27956 ──→ aura-validator-4:26656 (P2P)

Internal Docker Network: 172.26.0.0/16
├── validator-1: 172.26.0.10
├── validator-2: 172.26.0.11
├── validator-3: 172.26.0.12
└── validator-4: 172.26.0.13
```

## Consensus Math

```
Total validators: 4
Voting power per validator: 1 (equal)

Quorum required: 2/3 threshold
2/3 of 4 = 2.67 (rounds up to 3)

With 4/4 active: 4 ≥ 3 → CONSENSUS ✓
With 3/4 active: 3 ≥ 3 → CONSENSUS ✓
With 2/4 active: 2 < 3 → NO CONSENSUS ✗

The test validates these thresholds work correctly.
```

## BFT Bug Detection

**The most critical test is Phase 4 (2/4 halt).**

If blocks are produced with only 2/4 validators, this indicates:
- Consensus threshold not enforced
- Quorum calculation bug
- Tendermint/Cosmos SDK bug
- Validator voting power misconfiguration

This is a **CRITICAL** security issue and must be investigated immediately.

## Environment Variables

```bash
# Optional (defaults shown)
export VERBOSE=1                    # Verbose output
export OUTPUT_DIR=./bft_results     # Results directory
export AURA_HOME=~/.aura           # Node data directory
export CHAIN_ID=aura-local-4       # Chain identifier
```

## Useful Aliases

Add to `~/.bashrc`:

```bash
alias aura-test-bft='cd /home/decri/blockchain-projects/aura && \
  ./scripts/test-bft-comprehensive.sh --verbose'

alias aura-analyze-bft='cd /home/decri/blockchain-projects/aura && \
  python3 scripts/analyze-bft-results.py bft_test_*.log'

alias aura-start='cd /home/decri/blockchain-projects/aura && \
  docker-compose -f docker-compose.testnet.yml up -d'

alias aura-stop='cd /home/decri/blockchain-projects/aura && \
  docker-compose -f docker-compose.testnet.yml down'

alias aura-status='cd /home/decri/blockchain-projects/aura && \
  ./scripts/testnet-manage.sh status'
```

Then use:
```bash
aura-start
sleep 60
aura-test-bft
aura-analyze-bft
```

## Getting Help

```bash
# BFT test help
./scripts/test-bft-comprehensive.sh --help

# Analysis tool help
python3 scripts/analyze-bft-results.py --help

# Testnet management help
./scripts/testnet-manage.sh help

# Full test guide
cat docs/testing/BFT_TEST_GUIDE.md

# Expected results reference
cat docs/testing/BFT_TEST_EXPECTED_RESULTS.md
```

## Common Issues Checklist

```
[ ] Testnet not running?
    → docker-compose -f docker-compose.testnet.yml up -d

[ ] RPC not responding?
    → docker ps | grep aura
    → curl -s http://localhost:26657/status

[ ] Test failing at baseline?
    → Validators need more sync time
    → sleep 60 and try again

[ ] Block production not starting?
    → docker logs aura-validator-1 | tail -50

[ ] Container won't stop?
    → docker stop -t 30 aura-validator-3

[ ] Weird height divergence?
    → Network issue or forking
    → Restart the testnet
```

## Performance Tips

```bash
# Run with less verbose output
./scripts/test-bft-comprehensive.sh > /dev/null

# Run in background
nohup ./scripts/test-bft-comprehensive.sh > bft.log 2>&1 &

# Save output to specific location
./scripts/test-bft-comprehensive.sh --output-dir ~/bft_results

# Quick analysis after
tail -50 ~/bft_results/bft_test_*.log
```

## Key Metrics for Success

1. **Phase 1 Baseline**: Variance ≤ 1 block
2. **Phase 2 (3/4)**: Blocks produced ≥ 10
3. **Phase 3 Catch-up**: Difference ≤ 5 blocks
4. **Phase 4 (2/4)**: Blocks produced = **0** ← CRITICAL
5. **Phase 5 Recovery**: All validators ≥ 3

All must pass for test to be considered successful.

---

**For complete details, see:**
- BFT_TEST_GUIDE.md - Comprehensive guide
- BFT_TEST_EXPECTED_RESULTS.md - Detailed outcomes
- This file - Quick reference
