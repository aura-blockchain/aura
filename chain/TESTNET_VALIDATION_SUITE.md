# Aura Testnet Validation Suite

**Version**: 1.0
**Last Updated**: 2025-12-03
**Purpose**: Comprehensive testing checklist and commands for validating the local testnet

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Test Execution Order](#test-execution-order)
4. [Test Categories](#test-categories)
5. [Critical vs Nice-to-Have Tests](#critical-vs-nice-to-have-tests)
6. [Estimated Execution Time](#estimated-execution-time)

---

## Overview

This document provides a systematic approach to validating the Aura testnet after startup. Tests are organized by category with specific commands, expected outcomes, and dependencies clearly documented.

**Total Tests**: 42 specific validation tests
**Estimated Total Time**: 25-35 minutes (excluding WASM contract deployment preparation)

---

## Prerequisites

Before running tests, ensure:

1. **Testnet is running**: All validator nodes are operational
2. **Binary is available**: `aurad` binary is in PATH or specify full path
3. **Test accounts exist**: Validator keys are created during testnet initialization
4. **Node is synced**: Check that blocks are being produced

**Environment Variables** (adjust as needed):
```bash
export CHAIN_ID="aura-testnet-1"
export NODE="http://localhost:26657"
export VALIDATOR_HOME="$HOME/.testnets/aura-testnet/node0/aurad"
export VALIDATOR_KEY="validator0"
```

---

## Test Execution Order

Tests must be executed in the following order due to dependencies:

```
Phase 1: Infrastructure (Tests 1-8)
  └─> Phase 2: Accounts & Transfers (Tests 9-15)
      └─> Phase 3: Module Operations (Tests 16-28)
          └─> Phase 4: WASM Contracts (Tests 29-33)
              └─> Phase 5: Advanced Features (Tests 34-42)
```

---

## Test Categories

### Category 1: Basic Chain Operations (Tests 1-8)
**Time**: 3-5 minutes
**Critical**: ✓ All tests in this category are critical

#### Test 1: Query Chain Status
```bash
aurad status --node $NODE
```
**Expected Outcome**:
- JSON output with node info
- `catching_up: false`
- `latest_block_height` is incrementing

**Dependencies**: None

---

#### Test 2: Check Validator Set
```bash
aurad query staking validators --node $NODE --output json
```
**Expected Outcome**:
- List of validators (should match number of nodes in testnet)
- Each validator has `status: BOND_STATUS_BONDED`
- Combined voting power equals total stake

**Dependencies**: None

---

#### Test 3: Query Latest Block
```bash
aurad query block --node $NODE
```
**Expected Outcome**:
- Block data with transactions
- `block.header.height` matches status height
- `block.header.chain_id` equals `aura-testnet-1`

**Dependencies**: None

---

#### Test 4: Check Transaction Mempool
```bash
curl -s $NODE/num_unconfirmed_txs | jq '.'
```
**Expected Outcome**:
- JSON response
- `result.n_txs` shows pending tx count (likely 0 initially)
- `result.total` and `result.total_bytes` present

**Dependencies**: None

---

#### Test 5: Query Network Info
```bash
curl -s $NODE/net_info | jq '.result.peers'
```
**Expected Outcome**:
- Array of connected peers
- Peer count matches expected validators - 1
- Each peer has `node_info.moniker`

**Dependencies**: None

---

#### Test 6: Check Consensus State
```bash
curl -s $NODE/consensus_state | jq '.'
```
**Expected Outcome**:
- `result.round_state.height` is incrementing
- `result.round_state.step` shows consensus progress
- `result.round_state.validators` present

**Dependencies**: None

---

#### Test 7: Query Module Accounts
```bash
aurad query auth module-accounts --node $NODE --output json
```
**Expected Outcome**:
- List includes: `staking`, `distribution`, `bonded_pool`, `not_bonded_pool`, `gov`, `dex`, `bridge`, `wasm`
- Each account has valid `base_account.address`

**Dependencies**: None

---

#### Test 8: Check Bank Total Supply
```bash
aurad query bank total --node $NODE --output json
```
**Expected Outcome**:
- JSON with `supply` array
- Includes `uaura` denomination
- Amount matches genesis allocation

**Dependencies**: None

---

### Category 2: Account Operations (Tests 9-15)
**Time**: 4-6 minutes
**Critical**: Tests 9-13 are critical

#### Test 9: List Keys in Keyring
```bash
aurad keys list --home $VALIDATOR_HOME --keyring-backend test
```
**Expected Outcome**:
- Shows validator keys (validator0, validator1, etc.)
- Each key has address starting with `aura`
- Public keys are displayed

**Dependencies**: None

---

#### Test 10: Create Test Account
```bash
aurad keys add testuser1 --home $VALIDATOR_HOME --keyring-backend test
```
**Expected Outcome**:
- New account created
- Mnemonic displayed (save for recovery)
- Address starts with `aura`

**Dependencies**: None

---

#### Test 11: Query Account Balance (Validator)
```bash
VALIDATOR_ADDR=$(aurad keys show $VALIDATOR_KEY --home $VALIDATOR_HOME --keyring-backend test --address)
aurad query bank balances $VALIDATOR_ADDR --node $NODE --output json
```
**Expected Outcome**:
- Balance shows staked + available tokens
- Should have `uaura` with amount matching genesis

**Dependencies**: Test 10

---

#### Test 12: Query Account Info
```bash
aurad query auth account $VALIDATOR_ADDR --node $NODE --output json
```
**Expected Outcome**:
- Account number and sequence displayed
- Account type is `base_account` or `validator_vesting_account`
- Public key may be null if no txs sent

**Dependencies**: Test 11

---

#### Test 13: Send Simple Transfer
```bash
TEST_ADDR=$(aurad keys show testuser1 --home $VALIDATOR_HOME --keyring-backend test --address)

aurad tx bank send $VALIDATOR_ADDR $TEST_ADDR 1000000uaura \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes \
  --broadcast-mode sync
```
**Expected Outcome**:
- Transaction hash returned
- `code: 0` (success)
- Wait 5 seconds, then verify balance changed

**Dependencies**: Tests 10, 11

---

#### Test 14: Verify Transfer Completed
```bash
sleep 5
aurad query bank balances $TEST_ADDR --node $NODE --output json
```
**Expected Outcome**:
- `testuser1` balance shows 1000000uaura
- Validator balance decreased by 1000000uaura + fees

**Dependencies**: Test 13

---

#### Test 15: Query Transaction by Hash
```bash
# Replace TX_HASH with hash from Test 13
aurad query tx <TX_HASH> --node $NODE --output json
```
**Expected Outcome**:
- Transaction details displayed
- `code: 0` confirms success
- `events` array shows transfer details

**Dependencies**: Test 13

---

### Category 3: Module-Specific Tests (Tests 16-28)

#### 3A: DEX Module (Tests 16-20)
**Time**: 5-7 minutes
**Critical**: Tests 16-18 are critical

---

#### Test 16: Create Liquidity Pool
```bash
aurad tx dex create-pool \
  uaura utest \
  1000000uaura 1000000utest \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes
```
**Expected Outcome**:
- Pool created successfully
- Returns pool_id (e.g., "pool-1")
- LP tokens minted to creator

**Dependencies**: Test 14

---

#### Test 17: Query Pool Info
```bash
aurad query dex pool pool-1 --node $NODE --output json
```
**Expected Outcome**:
- Pool reserves match creation amounts
- Pool status is active
- LP token supply displayed

**Dependencies**: Test 16

---

#### Test 18: Execute Swap
```bash
aurad tx dex swap-exact-in \
  pool-1 \
  500000uaura \
  400000 \
  500 \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes
```
**Expected Outcome**:
- Swap executes successfully
- Returns amount_out
- Pool reserves updated
- Slippage within tolerance

**Dependencies**: Test 17

---

#### Test 19: Add Liquidity to Pool
```bash
aurad tx dex add-liquidity \
  pool-1 \
  500000uaura 500000utest \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes
```
**Expected Outcome**:
- Liquidity added successfully
- LP tokens minted
- Pool share percentage calculated

**Dependencies**: Test 18

---

#### Test 20: Remove Liquidity from Pool (Nice-to-Have)
```bash
# Get LP token balance first
LP_BALANCE=$(aurad query bank balances $VALIDATOR_ADDR --node $NODE --output json | jq -r '.balances[] | select(.denom | startswith("lp-pool-1")) | .amount')

aurad tx dex remove-liquidity \
  pool-1 \
  $((LP_BALANCE / 2)) \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes
```
**Expected Outcome**:
- Liquidity removed successfully
- Tokens returned proportionally
- LP tokens burned

**Dependencies**: Test 19

---

#### 3B: Bridge Module (Tests 21-23)
**Time**: 3-4 minutes
**Critical**: Test 21 is critical

---

#### Test 21: Lock Tokens for Bridge Transfer
```bash
aurad tx bridge lock-tokens \
  paw \
  paw1recipientaddress \
  100000uaura \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes
```
**Expected Outcome**:
- Tokens locked successfully
- Transfer ID returned
- Estimated completion time provided

**Dependencies**: Test 14

---

#### Test 22: Query Bridge Transfer Status
```bash
# Replace TRANSFER_ID with value from Test 21
aurad query bridge transfer <TRANSFER_ID> --node $NODE --output json
```
**Expected Outcome**:
- Transfer status displayed (pending/completed)
- Source and target chain info
- Amount and recipient visible

**Dependencies**: Test 21

---

#### Test 23: Link Cross-Chain Address (Nice-to-Have)
```bash
# This requires signatures from other chains - placeholder test
aurad tx bridge link-address \
  $VALIDATOR_ADDR \
  paw1validatoraddress \
  "" \
  <paw_signature> \
  "" \
  $VALIDATOR_ADDR \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes
```
**Expected Outcome**:
- Address linking recorded
- Linked identity ID returned
- Status: success

**Dependencies**: Test 21

---

#### 3C: Compliance Module (Tests 24-26)
**Time**: 2-3 minutes
**Critical**: Test 24 is critical

---

#### Test 24: Submit KYC Record
```bash
# Create PII commitment (SHA-256 of test data)
PII_HASH=$(echo -n '{"name":"Test User","dob":"1990-01-01"}' | sha256sum | awk '{print $1}')

aurad tx compliance submit-kyc \
  $TEST_ADDR \
  3 \
  kyc-provider-1 \
  $PII_HASH \
  US \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes
```
**Expected Outcome**:
- KYC record submitted
- Status: success
- Record stored with commitment hash

**Dependencies**: Test 10

---

#### Test 25: Query KYC Record
```bash
aurad query compliance kyc-record $TEST_ADDR --node $NODE --output json
```
**Expected Outcome**:
- KYC record displayed
- Level: INTERMEDIATE (3)
- PII commitment matches submitted hash
- Jurisdiction: US

**Dependencies**: Test 24

---

#### Test 26: Screen Sanctions (Nice-to-Have)
```bash
aurad tx compliance screen-sanctions \
  $TEST_ADDR \
  false \
  --from $TEST_ADDR \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes
```
**Expected Outcome**:
- Sanctions screening completed
- Status: CLEAR (or PENDING_REVIEW)
- No matches expected for test address

**Dependencies**: Test 25

---

#### 3D: Identity Module (Tests 27-28)
**Time**: 2-3 minutes
**Critical**: Test 27 is critical

---

#### Test 27: Create DID (Identity)
```bash
aurad tx identity create-did \
  did:aura:testuser1 \
  '{"controller":"'$TEST_ADDR'"}' \
  --from $TEST_ADDR \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes
```
**Expected Outcome**:
- DID created successfully
- DID document stored on-chain
- Controller matches test address

**Dependencies**: Test 14

---

#### Test 28: Query DID Document (Nice-to-Have)
```bash
aurad query identity did did:aura:testuser1 --node $NODE --output json
```
**Expected Outcome**:
- DID document retrieved
- Contains controller, verification methods
- Status: active

**Dependencies**: Test 27

---

### Category 4: WASM Contract Tests (Tests 29-33)
**Time**: 5-8 minutes
**Critical**: Tests 29-30 are critical

**Note**: These tests require a compiled WASM contract. If no contract is available, these tests can be skipped for initial testnet validation.

---

#### Test 29: Store WASM Contract Code
```bash
# Assuming a vc-issuer.wasm contract exists
aurad tx wasm store ./contracts/vc-issuer.wasm \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --gas auto \
  --gas-adjustment 1.3 \
  --fees 50000uaura \
  --yes
```
**Expected Outcome**:
- Contract code stored
- Code ID returned (e.g., code_id: 1)
- Gas used within limits

**Dependencies**: Test 14, requires WASM contract file

---

#### Test 30: Query Stored WASM Code
```bash
aurad query wasm code 1 --node $NODE --output json
```
**Expected Outcome**:
- Code info displayed
- Creator address matches validator
- Data hash present

**Dependencies**: Test 29

---

#### Test 31: Instantiate WASM Contract
```bash
aurad tx wasm instantiate 1 \
  '{"verifier":"'$VALIDATOR_ADDR'","issuer_name":"Aura Test Issuer"}' \
  --from $VALIDATOR_KEY \
  --label "vc-issuer-test" \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --admin $VALIDATOR_ADDR \
  --fees 20000uaura \
  --yes
```
**Expected Outcome**:
- Contract instantiated
- Contract address returned (aura1...)
- Events emitted

**Dependencies**: Test 30

---

#### Test 32: Execute WASM Contract Function
```bash
# Example: Issue a verifiable credential
CONTRACT_ADDR=$(aurad query wasm list-contract-by-code 1 --node $NODE --output json | jq -r '.contracts[0]')

aurad tx wasm execute $CONTRACT_ADDR \
  '{"issue_credential":{"subject":"'$TEST_ADDR'","claim_type":"test_claim","claim_value":"verified"}}' \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 10000uaura \
  --yes
```
**Expected Outcome**:
- Contract execution successful
- Events show credential issued
- State updated

**Dependencies**: Test 31

---

#### Test 33: Query WASM Contract State
```bash
aurad query wasm contract-state smart $CONTRACT_ADDR \
  '{"get_credential":{"subject":"'$TEST_ADDR'"}}' \
  --node $NODE \
  --output json
```
**Expected Outcome**:
- Query returns credential data
- Subject matches test address
- Claim data present

**Dependencies**: Test 32

---

### Category 5: Advanced Features (Tests 34-42)

#### 5A: Governance (Tests 34-37)
**Time**: 3-5 minutes
**Critical**: Test 34 is critical

---

#### Test 34: Submit Governance Proposal
```bash
aurad tx governance submit-proposal \
  "Test Parameter Change" \
  "Testing governance mechanism" \
  1 \
  $VALIDATOR_ADDR \
  1000000uaura \
  false \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes
```
**Expected Outcome**:
- Proposal submitted successfully
- Proposal ID returned (e.g., 1)
- Status: deposit_period

**Dependencies**: Test 14

---

#### Test 35: Add Deposit to Proposal
```bash
aurad tx governance deposit \
  1 \
  $VALIDATOR_ADDR \
  500000uaura \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes
```
**Expected Outcome**:
- Deposit added successfully
- Total deposit increases
- May trigger voting period if threshold met

**Dependencies**: Test 34

---

#### Test 36: Vote on Proposal
```bash
aurad tx governance vote \
  1 \
  $VALIDATOR_ADDR \
  1 \
  false \
  "" \
  --from $VALIDATOR_KEY \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --home $VALIDATOR_HOME \
  --keyring-backend test \
  --fees 5000uaura \
  --yes
```
**Expected Outcome**:
- Vote recorded
- Vote option: YES (1)
- Voting power counted

**Dependencies**: Test 35

---

#### Test 37: Query Proposal Status
```bash
aurad query governance proposal 1 --node $NODE --output json
```
**Expected Outcome**:
- Proposal details displayed
- Vote tallies present
- Status reflects current phase

**Dependencies**: Test 36

---

#### 5B: Staking Operations (Tests 38-40)
**Time**: 3-4 minutes
**Critical**: Test 38 is critical

---

#### Test 38: Query Staking Pool
```bash
aurad query staking pool --node $NODE --output json
```
**Expected Outcome**:
- Bonded tokens total displayed
- Not bonded tokens shown
- Totals match genesis allocation

**Dependencies**: None

---

#### Test 39: Query Validator Details
```bash
VALIDATOR_OPER=$(aurad keys show $VALIDATOR_KEY --home $VALIDATOR_HOME --keyring-backend test --bech val --address)
aurad query staking validator $VALIDATOR_OPER --node $NODE --output json
```
**Expected Outcome**:
- Validator info displayed
- Commission rate shown
- Delegations total matches bond

**Dependencies**: None

---

#### Test 40: Query Delegations (Nice-to-Have)
```bash
aurad query staking delegations-to $VALIDATOR_OPER --node $NODE --output json
```
**Expected Outcome**:
- List of delegations to validator
- Self-delegation present
- Share amounts correct

**Dependencies**: Test 39

---

#### 5C: Distribution & Rewards (Tests 41-42)
**Time**: 2-3 minutes
**Critical**: None (Nice-to-Have)

---

#### Test 41: Query Outstanding Rewards
```bash
aurad query distribution validator-outstanding-rewards $VALIDATOR_OPER --node $NODE --output json
```
**Expected Outcome**:
- Rewards accumulated (may be 0 if just started)
- Denomination is uaura

**Dependencies**: Test 39

---

#### Test 42: Query Community Pool
```bash
aurad query distribution community-pool --node $NODE --output json
```
**Expected Outcome**:
- Community pool balance shown
- May be empty initially
- Grows as fees accumulate

**Dependencies**: None

---

## Critical vs Nice-to-Have Tests

### Critical Tests (Must Pass)
These tests validate core functionality required for testnet operation:

**Category 1 - Basic Chain Operations**: Tests 1-8 (all)
**Category 2 - Account Operations**: Tests 9-13
**Category 3A - DEX**: Tests 16-18
**Category 3B - Bridge**: Test 21
**Category 3C - Compliance**: Test 24
**Category 3D - Identity**: Test 27
**Category 4 - WASM**: Tests 29-30
**Category 5A - Governance**: Test 34
**Category 5B - Staking**: Test 38

**Total Critical Tests**: 24

### Nice-to-Have Tests
These tests validate advanced features and edge cases:

**Category 2**: Tests 14-15
**Category 3A**: Tests 19-20
**Category 3B**: Tests 22-23
**Category 3C**: Tests 25-26
**Category 3D**: Test 28
**Category 4**: Tests 31-33
**Category 5A**: Tests 35-37
**Category 5B**: Test 40
**Category 5C**: Tests 41-42

**Total Nice-to-Have Tests**: 18

---

## Estimated Execution Time

### By Category
- **Category 1** (Basic Chain): 3-5 minutes
- **Category 2** (Accounts): 4-6 minutes
- **Category 3** (Modules): 12-17 minutes
  - DEX: 5-7 minutes
  - Bridge: 3-4 minutes
  - Compliance: 2-3 minutes
  - Identity: 2-3 minutes
- **Category 4** (WASM): 5-8 minutes (if contracts available)
- **Category 5** (Advanced): 8-12 minutes
  - Governance: 3-5 minutes
  - Staking: 3-4 minutes
  - Distribution: 2-3 minutes

### Total Time Estimates
- **Critical Tests Only**: 15-20 minutes
- **All Tests**: 25-35 minutes
- **All Tests + WASM**: 30-43 minutes

### Time-Saving Tips
1. Run queries in parallel where possible (no dependencies)
2. Reduce wait times between transactions (default 5s can be reduced to 3s for testnet)
3. Use `--broadcast-mode async` for fire-and-forget txs (not recommended for critical tests)
4. Pre-create test data files to avoid typing long commands

---

## Troubleshooting

### Common Issues

**Issue**: `account sequence mismatch`
**Solution**: Query account to get current sequence, or add `--sequence <num>` flag

**Issue**: `insufficient fees`
**Solution**: Increase `--fees` amount or use `--gas auto --gas-adjustment 1.3`

**Issue**: `connection refused`
**Solution**: Verify node is running and `$NODE` variable is correct

**Issue**: `tx not found`
**Solution**: Wait longer for tx to be included in a block (use `--broadcast-mode sync` or `block`)

**Issue**: `keyring backend error`
**Solution**: Ensure `--keyring-backend test` matches the backend used during key creation

---

## Test Results Logging

To log test results, use this template:

```bash
# Create results directory
mkdir -p ~/testnet-validation-results
RESULTS_FILE=~/testnet-validation-results/validation-$(date +%Y%m%d-%H%M%S).log

# Start logging
echo "Aura Testnet Validation - $(date)" | tee $RESULTS_FILE
echo "Chain ID: $CHAIN_ID" | tee -a $RESULTS_FILE
echo "Node: $NODE" | tee -a $RESULTS_FILE
echo "---" | tee -a $RESULTS_FILE

# Run each test and append result
echo "Test 1: Query Chain Status" | tee -a $RESULTS_FILE
aurad status --node $NODE 2>&1 | tee -a $RESULTS_FILE
echo "Result: PASS/FAIL" | tee -a $RESULTS_FILE
echo "" | tee -a $RESULTS_FILE

# Continue for all tests...
```

---

## Automated Test Runner

A test automation script is available at:
```
./scripts/run-testnet-validation.sh
```

Run with:
```bash
./scripts/run-testnet-validation.sh --critical-only  # Run critical tests only
./scripts/run-testnet-validation.sh --all           # Run all tests
./scripts/run-testnet-validation.sh --category dex  # Run specific category
```

---

## Success Criteria

The testnet is considered validated when:

1. ✓ **All critical tests pass** (24/24)
2. ✓ **Chain is producing blocks consistently** (no stalls for >10s)
3. ✓ **Validators are online** (all expected validators in set)
4. ✓ **Basic operations work** (transfers, queries, tx submission)
5. ✓ **Module functionality verified** (DEX, Bridge, Compliance, Identity)
6. ✓ **No consensus failures** (no validator slashing, no double-signs)

**Partial Success**:
- Critical tests pass but nice-to-have tests fail: Proceed with caution, investigate failures
- 90%+ tests pass: Generally acceptable, document known issues

**Failure**:
- Any critical test fails: Stop validation, investigate root cause before proceeding
- Chain stops producing blocks: Critical failure, requires node restart/debug
- Consensus errors: Critical failure, review validator logs

---

## Next Steps After Validation

Once validation is complete:

1. **Document Results**: Save test results log
2. **Performance Baseline**: Record block times, transaction throughput
3. **Stress Testing**: If validation passes, proceed to load testing
4. **Long-Running Test**: Leave testnet running overnight to test stability
5. **Upgrade Testing**: Test chain upgrade procedures
6. **Chaos Engineering**: Introduce failures (validator down, network partition) and test recovery

---

## Appendix: Quick Reference Commands

```bash
# Environment Setup
export CHAIN_ID="aura-testnet-1"
export NODE="http://localhost:26657"
export VALIDATOR_HOME="$HOME/.testnets/aura-testnet/node0/aurad"
export VALIDATOR_KEY="validator0"

# Common Query Commands
aurad status --node $NODE
aurad query bank balances <address> --node $NODE
aurad query staking validators --node $NODE
aurad query tx <hash> --node $NODE

# Common Transaction Commands
aurad tx bank send <from> <to> <amount> --from <key> --chain-id $CHAIN_ID --node $NODE --fees 5000uaura --yes
aurad tx dex swap-exact-in <pool> <coin_in> <min_out> <slippage> --from <key> --chain-id $CHAIN_ID --node $NODE --fees 5000uaura --yes

# Key Management
aurad keys list --home $VALIDATOR_HOME --keyring-backend test
aurad keys show <name> --home $VALIDATOR_HOME --keyring-backend test --address

# Node Management
curl -s $NODE/status | jq '.'
curl -s $NODE/net_info | jq '.result.peers'
curl -s $NODE/health
```

---

**End of Validation Suite**
