# Aura Testnet Validation Checklist

**Print this page for manual test tracking**
**Date**: ________________  **Tester**: ________________

---

## Pre-Validation Setup

- [ ] Testnet is running (all validators operational)
- [ ] `aurad` binary is available and correct version
- [ ] Environment variables configured:
  - [ ] `CHAIN_ID` set
  - [ ] `NODE` set to RPC endpoint
  - [ ] `VALIDATOR_HOME` set correctly
  - [ ] `VALIDATOR_KEY` exists in keyring
- [ ] Node is reachable: `curl -s $NODE/health`
- [ ] Node is synced: `aurad status --node $NODE | jq '.sync_info.catching_up'` returns `false`

---

## Category 1: Basic Chain Operations (CRITICAL)

- [ ] **Test 1**: Query chain status - blocks incrementing
- [ ] **Test 2**: Check validator set - all validators bonded
- [ ] **Test 3**: Query latest block - valid block data
- [ ] **Test 4**: Check transaction mempool - responds correctly
- [ ] **Test 5**: Query network info - peers connected
- [ ] **Test 6**: Check consensus state - height advancing
- [ ] **Test 7**: Query module accounts - all expected modules present
- [ ] **Test 8**: Check bank total supply - matches genesis

**Status**: _____ / 8 Passed | Notes: _________________________________

---

## Category 2: Account Operations (CRITICAL)

- [ ] **Test 9**: List keys in keyring - validator keys present
- [ ] **Test 10**: Create test account - new account created successfully
- [ ] **Test 11**: Query validator balance - balance > 0
- [ ] **Test 12**: Query account info - valid account data
- [ ] **Test 13**: Send simple transfer - transaction successful
- [ ] **Test 14**: Verify transfer completed - balance updated (Nice-to-Have)
- [ ] **Test 15**: Query transaction by hash - tx details retrieved (Nice-to-Have)

**Status**: _____ / 7 Passed | Notes: _________________________________

---

## Category 3A: DEX Module (CRITICAL)

- [ ] **Test 16**: Create liquidity pool - pool created
- [ ] **Test 17**: Query pool info - reserves match
- [ ] **Test 18**: Execute swap - swap successful, slippage OK
- [ ] **Test 19**: Add liquidity to pool - liquidity added (Nice-to-Have)
- [ ] **Test 20**: Remove liquidity - liquidity removed (Nice-to-Have)

**Status**: _____ / 5 Passed | Notes: _________________________________

---

## Category 3B: Bridge Module

- [ ] **Test 21**: Lock tokens for bridge transfer - tokens locked (CRITICAL)
- [ ] **Test 22**: Query bridge transfer status - status visible (Nice-to-Have)
- [ ] **Test 23**: Link cross-chain address - linking recorded (Nice-to-Have)

**Status**: _____ / 3 Passed | Notes: _________________________________

---

## Category 3C: Compliance Module

- [ ] **Test 24**: Submit KYC record - KYC submitted (CRITICAL)
- [ ] **Test 25**: Query KYC record - record retrieved (Nice-to-Have)
- [ ] **Test 26**: Screen sanctions - screening completed (Nice-to-Have)

**Status**: _____ / 3 Passed | Notes: _________________________________

---

## Category 3D: Identity Module

- [ ] **Test 27**: Create DID - DID created (CRITICAL)
- [ ] **Test 28**: Query DID document - DID retrieved (Nice-to-Have)

**Status**: _____ / 2 Passed | Notes: _________________________________

---

## Category 4: WASM Contracts

- [ ] **Test 29**: Store WASM contract code - code stored (CRITICAL)
- [ ] **Test 30**: Query stored WASM code - code info retrieved (CRITICAL)
- [ ] **Test 31**: Instantiate WASM contract - contract instantiated (Nice-to-Have)
- [ ] **Test 32**: Execute WASM contract function - execution successful (Nice-to-Have)
- [ ] **Test 33**: Query WASM contract state - state retrieved (Nice-to-Have)

**Status**: _____ / 5 Passed | Notes: _________________________________

---

## Category 5A: Governance

- [ ] **Test 34**: Submit governance proposal - proposal submitted (CRITICAL)
- [ ] **Test 35**: Add deposit to proposal - deposit added (Nice-to-Have)
- [ ] **Test 36**: Vote on proposal - vote recorded (Nice-to-Have)
- [ ] **Test 37**: Query proposal status - status retrieved (Nice-to-Have)

**Status**: _____ / 4 Passed | Notes: _________________________________

---

## Category 5B: Staking Operations

- [ ] **Test 38**: Query staking pool - pool data retrieved (CRITICAL)
- [ ] **Test 39**: Query validator details - validator info correct (Nice-to-Have)
- [ ] **Test 40**: Query delegations - delegations visible (Nice-to-Have)

**Status**: _____ / 3 Passed | Notes: _________________________________

---

## Category 5C: Distribution & Rewards

- [ ] **Test 41**: Query outstanding rewards - rewards visible (Nice-to-Have)
- [ ] **Test 42**: Query community pool - pool balance shown (Nice-to-Have)

**Status**: _____ / 2 Passed | Notes: _________________________________

---

## Summary Statistics

| Metric | Count |
|--------|-------|
| Total Tests | 42 |
| Critical Tests | 24 |
| Nice-to-Have Tests | 18 |
| Tests Passed | _____ |
| Tests Failed | _____ |
| Tests Skipped | _____ |

**Pass Rate**: _____ % (_____ / _____ executed)

---

## Final Validation Criteria

### Testnet is VALIDATED if:

- [x] All 24 critical tests pass
- [ ] Chain is producing blocks consistently (no stalls > 10s)
- [ ] All validators are online and signing
- [ ] Basic operations work (transfers, queries, transactions)
- [ ] Module functionality verified (DEX, Bridge, Compliance, Identity)
- [ ] No consensus failures (no slashing, no double-signs)

### Additional Success Criteria:

- [ ] 90%+ of all tests pass
- [ ] No critical errors in logs
- [ ] Transaction processing time < 10 seconds average
- [ ] Block time consistent (~6 seconds)
- [ ] Memory usage stable (no leaks)

---

## Issues Encountered

| Test # | Issue Description | Severity | Status |
|--------|-------------------|----------|--------|
|        |                   |          |        |
|        |                   |          |        |
|        |                   |          |        |
|        |                   |          |        |
|        |                   |          |        |

---

## Performance Metrics

| Metric | Value | Expected | Status |
|--------|-------|----------|--------|
| Block Time (avg) | _____ s | ~6s | _____ |
| Tx Processing Time | _____ s | <10s | _____ |
| Peak TPS | _____ | N/A | _____ |
| Validator Uptime | _____ % | >99% | _____ |
| Peer Count | _____ | ≥ validators-1 | _____ |
| Memory Usage | _____ MB | <2GB | _____ |
| Disk Usage | _____ GB | <10GB | _____ |

---

## Next Steps (After Validation)

- [ ] Save test results to file: `~/testnet-validation-results/`
- [ ] Document baseline performance metrics
- [ ] Leave testnet running for stability test (1+ hours)
- [ ] Monitor resource usage (CPU, memory, disk)
- [ ] Proceed to stress testing (if validation passes)
- [ ] Schedule long-running test (24-48 hours)
- [ ] Plan chaos engineering tests
- [ ] Document any issues for resolution

---

## Sign-off

**Validation Completed By**: ______________________ **Date**: ______________

**Result**:
- [ ] PASS - Testnet validated, ready for next phase
- [ ] PARTIAL PASS - Most tests passed, document failures
- [ ] FAIL - Critical tests failed, requires investigation

**Approver**: ______________________ **Date**: ______________

**Comments**:

_____________________________________________________________________

_____________________________________________________________________

_____________________________________________________________________

_____________________________________________________________________

_____________________________________________________________________

---

## Quick Reference Commands

```bash
# Environment setup
export CHAIN_ID="aura-testnet-1"
export NODE="http://localhost:26657"
export VALIDATOR_HOME="$HOME/.testnets/aura-testnet/node0/aurad"
export VALIDATOR_KEY="validator0"

# Quick health check
aurad status --node $NODE | jq '.sync_info'
curl -s $NODE/health

# Run automated tests
./scripts/run-testnet-validation.sh --critical-only  # Critical only
./scripts/run-testnet-validation.sh --all           # All tests

# Check results
cat ~/testnet-validation-results/validation-*.log | tail -50
```

---

**For detailed test procedures, see**: `TESTNET_VALIDATION_SUITE.md`
**For command reference, see**: `TESTNET_QUICK_REFERENCE.md`
**For overview, see**: `TESTNET_TESTING_SUMMARY.md`
