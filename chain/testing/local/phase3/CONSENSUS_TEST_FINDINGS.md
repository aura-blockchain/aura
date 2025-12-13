# Phase 3.2: Consensus Test Findings

## Test Execution Date
2025-12-13 11:30 UTC

## Critical Discovery: Single Validator Setup

### Finding
The current Aura testnet is configured with **only 1 active validator**, not 4 as initially assumed.

### Evidence
1. **Validator Query Results**:
   ```bash
   curl -s http://localhost:26657/validators | jq '.result.validators | length'
   # Output: 1
   ```

2. **Voting Power Distribution**:
   ```json
   {
     "address": "EFC499A9D74A08546FA3C742B7A1EEE94565DA94",
     "voting_power": "1000000"
   }
   ```
   - Single validator has 100% voting power (1,000,000 units)

3. **Genesis Configuration**:
   - Genesis file has 0 validators in `.validators[]` array
   - Only validator-1 appears to be configured as an actual validator
   - Other containers (validator-2, validator-3, validator-4) are likely full nodes, not validators

### Architecture Clarification

#### Current Setup
- **aura-validator-1**: Active validator with 100% voting power
- **aura-validator-2, validator-3, validator-4**: Full nodes (non-validating)
- **aura-sentry-1, sentry-2**: Sentry nodes
- **aura-observer-1**: Observer node

#### Why This Matters
- Stopping validator-2, validator-3, or validator-4 does NOT affect consensus
- Only stopping validator-1 would halt the network
- The testnet cannot demonstrate Byzantine Fault Tolerance (BFT) properties
- 2/3 consensus threshold cannot be tested with a single validator

## Test Results Analysis

### Test 1: 4-Node Network Baseline ✓
- **Status**: PASSED
- **Finding**: Network produces blocks with 1 active validator
- **Blocks Produced**: Continuous production
- **Peers**: validator-1 connected to 3 peers (other full nodes)

### Test 2: 3-Node Network ✓
- **Status**: PASSED (but not BFT test)
- **Action**: Stopped validator-4 container
- **Result**: Network continued producing blocks (23 blocks in 10 seconds)
- **Explanation**: validator-4 was not a validator, just a full node
- **validator-1 still had 100% voting power**

### Test 3: 2-Node Network ✗
- **Status**: FAILED (unexpected behavior revealed configuration)
- **Action**: Stopped validator-3 and validator-4 containers
- **Expected**: Network should halt (only 50% voting power)
- **Actual**: Network continued producing blocks (22 blocks in 15 seconds)
- **Explanation**: validator-1 still active with 100% voting power
- **This revealed the single-validator setup**

## Implications

### For Current Testing
1. **Cannot test BFT properties** with single validator
2. **Cannot test consensus thresholds** (need multiple validators)
3. **Cannot test Byzantine failures** (need 4+ validators minimum)
4. **Can test**:
   - Single node operation
   - P2P full node connectivity
   - Block production
   - State synchronization
   - RPC/API functionality

### For Production Readiness
1. **Must configure 4 actual validators** for proper testnet
2. Each validator should have equal voting power (25% each)
3. Genesis file must include all 4 validators
4. Each validator needs its own genesis transaction (gentx)

## Recommendations

### Immediate Actions
1. **Reconfigure testnet with 4 validators**:
   - Generate 4 separate validator keys
   - Create 4 gentx files
   - Update genesis.json with all 4 validators
   - Distribute voting power evenly (250,000 each = 25%)

2. **Re-run consensus tests** after reconfiguration

3. **Verify BFT properties**:
   - 4/4 validators: 100% voting power → consensus ✓
   - 3/4 validators: 75% voting power → consensus ✓
   - 2/4 validators: 50% voting power → halt ✗
   - 1/4 validators: 25% voting power → halt ✗

### Alternative: Document Current Architecture
If single-validator setup is intentional for current phase:
1. Update documentation to clarify architecture
2. Rename containers to reflect roles (validator-1, fullnode-2, fullnode-3, fullnode-4)
3. Plan multi-validator testnet for Phase 4 or Phase 5
4. Create separate test scenarios for current architecture

## Network Characteristics (Current Setup)

### Consensus Model
- **Type**: Single validator (not BFT)
- **Fault Tolerance**: None (single point of failure)
- **Liveness**: Depends on validator-1 uptime

### Performance
- **Block Time**: ~2 seconds per block
- **Block Production**: Consistent and reliable
- **Latency**: Low (single validator, no voting rounds)

### Connectivity
- **Validator Peers**: 3 (connected to full nodes)
- **Network Topology**: Star topology with validator-1 at center
- **Full Nodes**: Successfully sync and relay blocks

## Script Updates Needed

### test-consensus-scenarios.sh
Current script assumes 4 validators. Needs update to:
1. Detect actual validator count from RPC
2. Adjust tests based on validator distribution
3. Add pre-flight check for multi-validator requirement
4. Provide clear error message if insufficient validators

### Proposed Pre-Flight Check
```bash
VALIDATOR_COUNT=$(curl -s http://localhost:26657/validators | jq '.result.validators | length')

if [ "$VALIDATOR_COUNT" -lt 4 ]; then
    echo "ERROR: This test requires 4 validators"
    echo "Current validator count: $VALIDATOR_COUNT"
    echo "Please reconfigure testnet with 4 validators"
    exit 1
fi
```

## Next Steps for Phase 3

### 3.2 Consensus Scenarios
- [x] Run tests (revealed single-validator setup)
- [ ] Reconfigure testnet with 4 validators
- [ ] Re-run tests with proper multi-validator setup
- [ ] Document actual BFT behavior

### 3.3 Network Chaos Testing
- Can proceed with current setup for:
  - Latency testing
  - Packet loss testing
  - Network partition testing (with full nodes)
- Limited value without multiple validators

### 3.4 Malicious Peer Handling
- Can proceed with current setup
- Test invalid data from full nodes
- Cannot test Byzantine validator behavior

## Conclusion

The consensus scenario tests successfully **revealed the actual testnet architecture**: a single-validator setup with supporting full nodes. While this prevents true BFT testing, the tests achieved their diagnostic purpose.

**For production-grade testing, reconfiguration with 4 validators is required.**

## References
- Test Script: `chain/testing/local/phase3/test-consensus-scenarios.sh`
- Results Log: `chain/testing/local/phase3/consensus_test_results.log`
- Management: `scripts/testnet-manage.sh`
- RPC Endpoint: http://localhost:26657
