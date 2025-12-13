# Phase 3 Testing Suite - File Index

## Overview

Complete testing suite for 4-validator consensus testing on Aura blockchain.

**Purpose**: Validate Byzantine Fault Tolerance (BFT) consensus properties with multiple validators.

## Quick Navigation

- **New to Phase 3?** Start with [QUICK_START.md](QUICK_START.md)
- **Detailed documentation:** [README.md](README.md)
- **Prerequisites check:** Run `./check-prerequisites.sh`
- **Run tests:** Run `./4-validator-consensus-test.sh`

## File Structure

```
phase3/
├── INDEX.md                          # This file - navigation guide
├── README.md                         # Comprehensive documentation
├── QUICK_START.md                    # Quick start guide
├── REPORT_TEMPLATE.md                # Template for test reports
│
├── check-prerequisites.sh            # Prerequisites verification script
├── 4-validator-consensus-test.sh     # Main test suite (automated)
├── consensus-analyzer.sh             # Analysis and reporting tool
├── validator-control.sh              # Validator management tool
│
├── test-consensus-scenarios.sh       # Legacy consensus tests
└── test-network-chaos.sh             # Network chaos testing
```

## Core Files

### Documentation

| File | Purpose | When to Use |
|------|---------|-------------|
| **INDEX.md** | File navigation (this file) | Finding your way around |
| **README.md** | Complete testing guide | Understanding the full system |
| **QUICK_START.md** | Quick start instructions | Getting started quickly |
| **REPORT_TEMPLATE.md** | Test report template | Documenting test results |

### Executable Scripts

| File | Purpose | When to Use |
|------|---------|-------------|
| **check-prerequisites.sh** | Verify setup is complete | Before running tests |
| **4-validator-consensus-test.sh** | Automated test suite | Running comprehensive tests |
| **consensus-analyzer.sh** | Analysis and reporting | Analyzing consensus state |
| **validator-control.sh** | Validator management | Starting/stopping validators |

### Legacy Scripts

| File | Purpose | Status |
|------|---------|--------|
| test-consensus-scenarios.sh | Earlier consensus tests | Superseded by 4-validator-consensus-test.sh |
| test-network-chaos.sh | Network chaos testing | Supplementary testing |

## Workflow Guide

### First Time Setup

```
1. Read: QUICK_START.md
2. Run: ./check-prerequisites.sh
3. Fix any issues identified
4. Run: ./4-validator-consensus-test.sh
```

### Regular Testing

```
1. Run: ./validator-control.sh status
2. Run: ./4-validator-consensus-test.sh
3. Review: consensus_test_results_*.log
4. Generate: ./consensus-analyzer.sh report
```

### Manual Exploration

```
1. Run: ./validator-control.sh start all
2. Run: ./validator-control.sh monitor
3. Run: ./consensus-analyzer.sh (interactive)
4. Experiment with stopping/starting validators
```

## Script Details

### check-prerequisites.sh

**Purpose**: Verify all prerequisites before running tests

**Checks**:
- Environment variables set
- Binary exists and is executable
- All 4 validator homes initialized
- Port availability
- Required tools installed (curl, jq, bc)
- Chain ID consistency
- Genesis configuration
- Test scripts are executable

**Usage**:
```bash
./check-prerequisites.sh
```

**Output**: Pass/fail report with remediation steps

---

### 4-validator-consensus-test.sh

**Purpose**: Main automated test suite for consensus validation

**Tests Executed**:
1. 4 validators (100% VP) → Should produce blocks
2. 3 validators (75% VP) → Should produce blocks
3. 2 validators (50% VP) → Should HALT
4. 1 validator (25% VP) → Should HALT
5. Restart and sync → Should resume

**Usage**:
```bash
./4-validator-consensus-test.sh
```

**Duration**: ~3-5 minutes

**Output**:
- Console: Colored pass/fail output
- File: `consensus_test_results_YYYYMMDD_HHMMSS.log`

**Features**:
- Automatic validator control
- Block height tracking
- Voting power calculations
- Pass/fail verification
- Comprehensive logging

---

### consensus-analyzer.sh

**Purpose**: Interactive analysis tool for consensus state

**Modules**:
1. Validator Set Analysis - Voting power distribution
2. Consensus State - Active validators and status
3. Block Production - Recent blocks and proposers
4. Peer Connectivity - P2P network topology
5. Full Report - Comprehensive markdown report

**Usage**:
```bash
# Interactive mode
./consensus-analyzer.sh

# Command-line mode
./consensus-analyzer.sh [validators|consensus|blocks|peers|report|all]
```

**Output**:
- Console: Real-time analysis
- File: `consensus_analysis_YYYYMMDD_HHMMSS.md`

**Best For**:
- Debugging consensus issues
- Understanding network state
- Generating audit reports

---

### validator-control.sh

**Purpose**: Simplified validator management and monitoring

**Commands**:
- `status` - Full status display
- `quick` - Compact status
- `monitor` - Live monitoring
- `start <1-4|all>` - Start validator(s)
- `stop <1-4|all>` - Stop validator(s)
- `restart <1-4|all>` - Restart validator(s)
- `logs <1-4>` - Tail validator logs

**Usage Examples**:
```bash
# Status
./validator-control.sh status
./validator-control.sh quick
./validator-control.sh monitor

# Control
./validator-control.sh start all
./validator-control.sh stop 4
./validator-control.sh restart 2

# Logs
./validator-control.sh logs 1
```

**Best For**:
- Daily operations
- Quick status checks
- Manual testing scenarios

---

## Common Tasks

### Task: Run Full Test Suite

```bash
cd /home/hudson/blockchain-projects/aura/chain/testing/local/phase3
./check-prerequisites.sh
./4-validator-consensus-test.sh
./consensus-analyzer.sh report
```

### Task: Quick Status Check

```bash
cd /home/hudson/blockchain-projects/aura/chain/testing/local/phase3
./validator-control.sh quick
```

### Task: Start Fresh Network

```bash
cd /home/hudson/blockchain-projects/aura/chain/testing/local/phase3
./validator-control.sh stop all
sleep 3
./validator-control.sh start all
./validator-control.sh status
```

### Task: Test Specific Scenario

```bash
# Test consensus with 3 validators
./validator-control.sh start all
./validator-control.sh stop 4
./consensus-analyzer.sh consensus

# Restore
./validator-control.sh start all
```

### Task: Generate Audit Report

```bash
./consensus-analyzer.sh all
# Review: consensus_analysis_YYYYMMDD_HHMMSS.md
```

### Task: Monitor Live

```bash
./validator-control.sh monitor
# Press Ctrl+C to stop
```

### Task: Troubleshoot Issues

```bash
# 1. Check prerequisites
./check-prerequisites.sh

# 2. Check validator status
./validator-control.sh status

# 3. Check logs
./validator-control.sh logs 1

# 4. Full analysis
./consensus-analyzer.sh all
```

## Generated Files

### Log Files

Pattern: `consensus_test_results_YYYYMMDD_HHMMSS.log`

**Content**: Full output from automated test suite
- Test execution timestamps
- Block heights
- Voting power calculations
- Pass/fail results
- Error messages

**Location**: `/home/hudson/blockchain-projects/aura/chain/testing/local/phase3/`

### Report Files

Pattern: `consensus_analysis_YYYYMMDD_HHMMSS.md`

**Content**: Comprehensive markdown report
- Executive summary
- Validator details
- Recent blocks table
- BFT consensus properties
- Network topology

**Location**: `/home/hudson/blockchain-projects/aura/chain/testing/local/phase3/`

### Validator Logs

Pattern: `validator1.log`, `validator2.log`, `validator3.log`, `validator4.log`

**Content**: Real-time validator output
- Block proposals
- Consensus rounds
- Peer connections
- State sync
- Error messages

**Location**: `/home/hudson/blockchain-projects/aura/chain/testing/local/phase3/`

## Testing Scenarios

### Scenario 1: Full Consensus (4 Validators)

**Setup**: All validators running
**Expected**: Blocks produced, 100% voting power
**Command**:
```bash
./validator-control.sh start all
./consensus-analyzer.sh blocks
```

### Scenario 2: Minimum Consensus (3 Validators)

**Setup**: 3 validators running (75% VP)
**Expected**: Blocks produced, consensus maintained
**Command**:
```bash
./validator-control.sh start all
./validator-control.sh stop 4
./consensus-analyzer.sh consensus
```

### Scenario 3: Below Threshold (2 Validators)

**Setup**: 2 validators running (50% VP)
**Expected**: Chain halts, no blocks produced
**Command**:
```bash
./validator-control.sh start all
./validator-control.sh stop 3
./validator-control.sh stop 4
./consensus-analyzer.sh blocks
```

### Scenario 4: Recovery from Halt

**Setup**: Start from halted state, restart validators
**Expected**: Chain resumes, validators sync
**Command**:
```bash
# Start from 2 validators (halted)
./validator-control.sh start all
./consensus-analyzer.sh consensus
```

### Scenario 5: Rolling Restart

**Setup**: Restart validators one at a time
**Expected**: Consensus maintained throughout
**Command**:
```bash
for i in 1 2 3 4; do
  ./validator-control.sh restart $i
  sleep 10
  ./validator-control.sh status
done
```

## Integration Points

### With Phase 2 Tests

Phase 3 builds on Phase 2:
- Phase 2: Genesis workflow, configuration parameters
- Phase 3: Consensus validation with multiple validators

### With Phase 4 (Future)

Phase 3 prepares for Phase 4:
- Phase 3: Local multi-validator consensus
- Phase 4: IBC connections, cross-chain testing

## Dependencies

### System Requirements
- Ubuntu Linux
- Go 1.24+
- curl, jq, bc, netstat

### Project Requirements
- Aura binary built (`chain/aurad`)
- 4 validator homes initialized
- Ports available: 26657, 26667, 26677, 26687 (RPC)
- Ports available: 26656, 26666, 26676, 26686 (P2P)
- Ports available: 9090, 9092, 9093, 9094 (gRPC)

### Environment
- Working directory: `/home/hudson/blockchain-projects/aura`
- Environment sourced: `source env.sh`
- GOCACHE set correctly
- LOG_DIR set correctly

## Troubleshooting Index

| Problem | Solution | Reference |
|---------|----------|-----------|
| Binary not found | Run `cd chain && make build` | README.md |
| Validator homes missing | Run 4-validator setup script | README.md |
| Ports in use | Kill old processes: `pkill -f aurad` | README.md |
| Validators not connecting | Check persistent_peers config | README.md |
| Chain not producing blocks | Ensure ≥3 validators running | QUICK_START.md |
| Tests failing | Run prerequisites check | check-prerequisites.sh |
| Logs not showing | Check validator started successfully | validator-control.sh |

## Best Practices

1. **Always check prerequisites first**
   ```bash
   ./check-prerequisites.sh
   ```

2. **Use automated tests for validation**
   ```bash
   ./4-validator-consensus-test.sh
   ```

3. **Generate reports for audit trail**
   ```bash
   ./consensus-analyzer.sh report
   ```

4. **Monitor in real-time during manual tests**
   ```bash
   ./validator-control.sh monitor
   ```

5. **Save all logs and reports**
   - Keep timestamped logs
   - Archive test results
   - Document anomalies

## Support and Resources

### Documentation
- This index: `INDEX.md`
- Full guide: `README.md`
- Quick start: `QUICK_START.md`
- Report template: `REPORT_TEMPLATE.md`

### Scripts
- Prerequisites: `./check-prerequisites.sh`
- Testing: `./4-validator-consensus-test.sh`
- Analysis: `./consensus-analyzer.sh`
- Control: `./validator-control.sh`

### Help Commands
```bash
./validator-control.sh --help
./consensus-analyzer.sh
./check-prerequisites.sh
```

## Version History

| Date | Version | Changes |
|------|---------|---------|
| 2025-12-13 | 1.0 | Initial Phase 3 testing suite |

## Contributing

When adding new tests or tools:
1. Update this INDEX.md
2. Update README.md if needed
3. Follow existing script patterns
4. Add help/usage information
5. Use consistent naming conventions
6. Document all prerequisites

## License

Part of the Aura blockchain project.

---

**Last Updated**: 2025-12-13
**Maintained By**: Aura Development Team
**Contact**: See main project README
