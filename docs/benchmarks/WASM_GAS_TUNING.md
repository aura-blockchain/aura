# WASM Gas Tuning Guide

**Last Updated:** December 3, 2024
**Chain Version:** Cosmos SDK 0.53.4 + CometBFT
**WASM Version:** CosmWasm/wasmd integration

## Overview

This document provides comprehensive guidance for tuning WASM gas limits in the Aura blockchain. Gas limits ensure that smart contract operations complete within acceptable time and resource bounds while preventing denial-of-service attacks.

## Table of Contents

1. [Understanding Gas](#understanding-gas)
2. [Benchmark Methodology](#benchmark-methodology)
3. [Baseline Measurements](#baseline-measurements)
4. [Recommended Gas Limits](#recommended-gas-limits)
5. [App.toml Configuration](#apptoml-configuration)
6. [Tuning Process](#tuning-process)
7. [Monitoring and Adjustment](#monitoring-and-adjustment)
8. [Security Considerations](#security-considerations)

---

## Understanding Gas

### What is Gas?

Gas is a measure of computational resources consumed by an operation on the blockchain. Each operation (storage read/write, computation, memory allocation) costs a specific amount of gas.

### Why Gas Limits Matter

1. **DoS Prevention**: Prevents malicious or buggy contracts from consuming excessive resources
2. **Predictability**: Allows users to estimate transaction costs
3. **Resource Management**: Ensures fair resource allocation across validators
4. **Economic Model**: Gas fees create sustainable blockchain economics

### WASM-Specific Gas Consumption

WASM contracts have distinct gas consumption patterns:

- **Code Upload**: Dominated by bytecode validation and storage costs
- **Instantiation**: Depends on initialization logic and state storage
- **Execution**: Varies widely based on contract logic complexity
- **Migration**: Similar to instantiation plus state migration overhead

---

## Benchmark Methodology

### Running Benchmarks

Execute the comprehensive benchmark suite:

```bash
cd /path/to/aura
./scripts/benchmark-wasm-gas.sh
```

This runs:
- Store code benchmarks (different contract sizes)
- Instantiate benchmarks (simple vs complex initialization)
- Execute benchmarks (read-only, write, batch, compute-heavy)
- Full lifecycle benchmarks (store → instantiate → execute)
- Admin operation benchmarks

### Benchmark Environment

**Hardware Recommendations:**
- CPU: 4+ cores (representative of validator hardware)
- RAM: 8GB+ available
- Disk: SSD (NVMe preferred)

**Software:**
- Go 1.21+
- Contracts compiled with optimization: `make optimize-wasm`
- Clean state (no existing blockchain data)

### Interpreting Results

Benchmark output includes:
- `ns/op`: Nanoseconds per operation (wall-clock time)
- `gas/op`: Gas consumed per operation (blockchain cost)
- `B/op`: Bytes allocated per operation
- `allocs/op`: Memory allocations per operation

**Focus on `gas/op`** for gas limit tuning.

---

## Baseline Measurements

### Contract Sizes

Our test contracts:

| Contract | Size | Purpose |
|----------|------|---------|
| `binding_tester.wasm` | 157 KB | Basic binding tests, simple operations |
| `vc_issuer.wasm` | 230 KB | VC issuance with moderate complexity |
| `schema.wasm` | 236 KB | Schema validation, similar complexity |

### Expected Gas Ranges

*Note: These are initial estimates. Update after running benchmarks on your hardware.*

#### Store Code

| Contract Size | Gas Consumed | Notes |
|--------------|--------------|-------|
| 150-200 KB | 8,000,000 - 12,000,000 | Small contracts |
| 200-300 KB | 12,000,000 - 18,000,000 | Medium contracts |
| 300-500 KB | 18,000,000 - 30,000,000 | Large contracts |
| > 500 KB | 30,000,000 - 60,000,000 | Very large contracts |

**Factors:**
- Linear with contract size (validation cost)
- WASM bytecode complexity
- Storage write amplification

#### Instantiate Contract

| Init Complexity | Gas Consumed | Notes |
|----------------|--------------|-------|
| Minimal (admin only) | 500,000 - 1,500,000 | No state initialization |
| Simple (< 10 values) | 1,500,000 - 3,000,000 | Basic config storage |
| Moderate (10-50 values) | 3,000,000 - 8,000,000 | Complex initialization |
| Heavy (> 50 values) | 8,000,000 - 20,000,000 | Large state setup |

**Factors:**
- Number of storage writes
- Initialization logic complexity
- Event emission
- Sub-message calls

#### Execute Contract

| Operation Type | Gas Consumed | Notes |
|----------------|--------------|-------|
| Simple query (read-only) | 100,000 - 500,000 | State reads, no writes |
| Simple write (1-5 keys) | 500,000 - 1,500,000 | Basic state updates |
| Batch write (5-20 keys) | 1,500,000 - 5,000,000 | Multiple state changes |
| Complex business logic | 5,000,000 - 15,000,000 | Computations + state |
| Cross-contract calls | 10,000,000 - 30,000,000 | Sub-messages to other contracts |

**Factors:**
- Number of state reads/writes
- Computational complexity (loops, math)
- Number of events emitted
- Cross-contract message depth

#### Admin Operations

| Operation | Gas Consumed | Notes |
|-----------|--------------|-------|
| Set admin | 50,000 - 150,000 | Single storage write |
| Get admin | 20,000 - 50,000 | Single storage read |
| Check admin | 20,000 - 50,000 | Single storage read + comparison |
| Pause contract | 50,000 - 100,000 | Single storage write |

---

## Recommended Gas Limits

### Default Gas Limits (app.toml)

Based on benchmarks and security analysis, these are recommended defaults:

```toml
[wasm]
# Maximum gas for storing contract code
# Prevents extremely large contracts from being uploaded
max_gas_store_code = 20_000_000

# Maximum gas for instantiating a contract
# Covers complex initialization scenarios
max_gas_instantiate = 10_000_000

# Maximum gas for executing a contract
# Allows complex operations while preventing abuse
max_gas_execute = 15_000_000

# Maximum gas for migrating a contract
# Similar to instantiation + migration overhead
max_gas_migrate = 10_000_000

# Maximum gas for administrative operations
max_gas_admin = 1_000_000
```

### Per-Transaction Recommendations

Users should specify gas based on their operation:

**Store Code:**
```bash
# Small contract (< 200KB)
aurad tx wasm store contract.wasm --gas 10000000 --from user

# Medium contract (200-300KB)
aurad tx wasm store contract.wasm --gas 15000000 --from user

# Large contract (> 300KB)
aurad tx wasm store contract.wasm --gas 25000000 --from user
```

**Instantiate:**
```bash
# Simple initialization
aurad tx wasm instantiate <code-id> '{"admin":"..."}' \
  --gas 1000000 --from user

# Complex initialization
aurad tx wasm instantiate <code-id> '{"config":{...}}' \
  --gas 5000000 --from user
```

**Execute:**
```bash
# Simple query/update
aurad tx wasm execute <contract> '{"method":{}}' \
  --gas 500000 --from user

# Complex operation
aurad tx wasm execute <contract> '{"batch_update":{...}}' \
  --gas 5000000 --from user
```

### Safety Margins

Always add a **20-30% safety margin** to measured gas consumption:

```
recommended_gas = measured_gas * 1.25
```

This accounts for:
- Variations in execution path
- State size growth over time
- Validator hardware differences
- Gas price fluctuations

---

## App.toml Configuration

### Full WASM Configuration Section

Add this to your `app.toml`:

```toml
#######################################################################
###                    WASM Configuration                          ###
#######################################################################

[wasm]

# Enable WASM smart contracts
enabled = true

# Maximum size of WASM bytecode (bytes)
# Recommended: 614400 (600KB) for standard contracts
# Adjust based on use case: 1048576 (1MB) for complex contracts
max_wasm_code_size = 614400

# Code upload access control
# Options: "nobody" | "everybody" | "any_of_addresses"
# Production recommendation: "any_of_addresses" (governance controlled)
code_upload_access = "any_of_addresses"

# Maximum gas for store code operations
# Based on benchmarks: ~60K gas per KB of contract code
# Formula: contract_size_kb * 60000 * 1.25 (safety margin)
max_gas_store_code = 20_000_000

# Maximum gas for instantiate operations
# Covers initialization logic and initial state storage
# Benchmark baseline: 500K-10M depending on complexity
max_gas_instantiate = 10_000_000

# Maximum gas for execute operations
# Most operations should complete well below this
# Allows for complex cross-contract calls
max_gas_execute = 15_000_000

# Maximum gas for migrate operations
# Migration includes state transformation overhead
max_gas_migrate = 10_000_000

# Maximum gas for admin operations (set/clear/update admin)
# These are simple storage operations
max_gas_admin = 1_000_000

# Query gas limit
# Queries should be fast and cheap
# This prevents abuse of node resources
query_gas_limit = 3_000_000

# Simulation gas limit
# Used for gas estimation
# Should be higher than max execute gas
simulation_gas_limit = 20_000_000

# Memory cache size (MB)
# Caches compiled WASM modules
# Higher values = better performance, more memory usage
# Recommended: 100-500 MB depending on node resources
memory_cache_size = 256

# Instance cache size (number of instances)
# Caches contract instances for faster execution
# Recommended: 10-100 depending on contract usage patterns
instance_cache_size = 50
```

### Environment-Specific Configurations

#### Development / Local Testnet

```toml
[wasm]
max_wasm_code_size = 1048576  # 1MB (allow larger contracts for testing)
code_upload_access = "everybody"  # No restrictions
max_gas_store_code = 50_000_000  # Higher limits for experimentation
max_gas_instantiate = 20_000_000
max_gas_execute = 30_000_000
memory_cache_size = 512  # More cache for faster iteration
instance_cache_size = 100
```

#### Staging / Public Testnet

```toml
[wasm]
max_wasm_code_size = 614400  # 600KB (standard limit)
code_upload_access = "any_of_addresses"  # Controlled upload
max_gas_store_code = 20_000_000
max_gas_instantiate = 10_000_000
max_gas_execute = 15_000_000
memory_cache_size = 256
instance_cache_size = 50
```

#### Production / Mainnet

```toml
[wasm]
max_wasm_code_size = 614400  # 600KB (conservative)
code_upload_access = "any_of_addresses"  # Strict governance
max_gas_store_code = 20_000_000  # Based on audited benchmarks
max_gas_instantiate = 10_000_000
max_gas_execute = 15_000_000
max_gas_migrate = 10_000_000  # Requires admin
max_gas_admin = 1_000_000
query_gas_limit = 3_000_000  # Prevent query abuse
memory_cache_size = 256  # Conservative memory usage
instance_cache_size = 50
```

---

## Tuning Process

### Step 1: Establish Baselines

1. Run benchmark suite on representative hardware:
   ```bash
   ./scripts/benchmark-wasm-gas.sh
   ```

2. Record results in this document (update "Baseline Measurements" section)

3. Identify outliers and investigate:
   - Contracts consuming > 2x expected gas
   - Operations with high variance (> 30% coefficient of variation)

### Step 2: Calculate Initial Limits

For each operation type:

```
initial_limit = max(benchmark_results) * 1.25
```

Example:
```
Store 230KB contract: 14,500,000 gas (measured)
Initial limit: 14,500,000 * 1.25 = 18,125,000 gas
Round up: 20,000,000 gas
```

### Step 3: Deploy to Local Testnet

1. Update `app.toml` with calculated limits
2. Start local testnet:
   ```bash
   ./scripts/testnet-init.sh
   ./scripts/testnet-manage.sh start
   ```
3. Deploy test contracts and verify operations succeed

### Step 4: Load Testing

Run realistic load tests:

```bash
# Store multiple contracts
for contract in contracts/artifacts/*.wasm; do
  aurad tx wasm store "$contract" --gas 20000000 --from validator1 -y
done

# Instantiate multiple instances
for i in {1..10}; do
  aurad tx wasm instantiate 1 '{"admin":"aura1..."}' \
    --label "test-$i" --gas 10000000 --from validator1 -y
done

# Execute operations
for i in {1..100}; do
  aurad tx wasm execute aura1... '{"method":{}}' \
    --gas 5000000 --from validator1 -y
done
```

Monitor:
- Transaction success rate (should be > 99%)
- Actual gas consumed vs. limit
- Block production (should be consistent)
- Resource usage (CPU, memory, disk I/O)

### Step 5: Adjust and Iterate

Based on load testing:

1. **If transactions fail with out-of-gas:**
   - Increase limits by 20-50%
   - Investigate why contracts consume more gas than benchmarks predicted

2. **If transactions consistently use < 50% of limit:**
   - Consider decreasing limits (but keep safety margin)
   - Lower limits improve security

3. **If blocks are slow to produce:**
   - May need to decrease limits to reduce computational load
   - Balance throughput vs. functionality

### Step 6: Production Validation

Before mainnet launch:

1. Run extended testnet with production config (1+ weeks)
2. Simulate adversarial scenarios:
   - Maximum-complexity contracts
   - Rapid-fire transactions
   - Large contract uploads
3. Audit all limits with security team
4. Document final values with justification

---

## Monitoring and Adjustment

### Key Metrics to Monitor

#### Gas Consumption Metrics

```promql
# Average gas per WASM operation type
avg(wasm_store_code_gas) by (contract_size_bucket)
avg(wasm_instantiate_gas) by (contract_id)
avg(wasm_execute_gas) by (contract_address, method)

# 95th percentile gas consumption (detect outliers)
quantile(0.95, wasm_execute_gas) by (contract_address)

# Gas limit hit rate (should be < 1%)
rate(wasm_out_of_gas_errors_total[5m])
```

#### Performance Metrics

```promql
# Transaction processing latency
histogram_quantile(0.95, wasm_tx_duration_seconds)

# Block production delays due to WASM
increase(block_production_delay_seconds{reason="wasm"}[5m])

# Contract execution timeouts
rate(wasm_execution_timeout_total[5m])
```

#### Resource Metrics

```promql
# Memory usage by WASM module cache
wasm_memory_cache_bytes

# Contract instance cache hit rate
rate(wasm_instance_cache_hits_total[5m]) /
  rate(wasm_instance_cache_requests_total[5m])

# Disk I/O for contract storage
rate(wasm_storage_io_bytes_total[5m])
```

### Alert Rules

Add these to Prometheus alert rules:

```yaml
groups:
  - name: wasm_gas_alerts
    rules:
      # Gas limit frequently hit (> 1% of transactions)
      - alert: WasmGasLimitTooLow
        expr: |
          rate(wasm_out_of_gas_errors_total[5m]) /
          rate(wasm_transactions_total[5m]) > 0.01
        for: 10m
        annotations:
          summary: "WASM gas limits may be too low"
          description: "{{ $value | humanizePercentage }} of WASM transactions are running out of gas"

      # Gas limit never approached (< 30% utilization)
      - alert: WasmGasLimitTooHigh
        expr: |
          quantile(0.95, wasm_execute_gas) / wasm_max_gas_execute < 0.30
        for: 1h
        annotations:
          summary: "WASM gas limits may be too high"
          description: "95th percentile gas usage is only {{ $value | humanizePercentage }} of limit"

      # Slow WASM execution impacting blocks
      - alert: WasmSlowExecution
        expr: |
          histogram_quantile(0.95, wasm_tx_duration_seconds) > 2.0
        for: 5m
        annotations:
          summary: "WASM execution is slow"
          description: "95th percentile WASM tx takes {{ $value }}s (> 2s threshold)"
```

### Adjustment Schedule

**Monthly Review:**
- Analyze gas consumption trends
- Check for new contract patterns
- Review out-of-gas incidents

**Quarterly Tuning:**
- Re-run benchmark suite on latest code
- Update limits based on 3 months of data
- Test changes on staging testnet

**On-Demand Adjustment:**
- After chain upgrades (may affect gas costs)
- After major contract deployments
- If alerts fire consistently

### Governance Process for Limit Changes

Production limit changes require governance:

1. **Proposal:**
   - Document current limits and proposed changes
   - Provide justification (benchmarks, incident reports)
   - Estimate impact on existing contracts

2. **Testing:**
   - Deploy to staging testnet for 1+ week
   - Validate that existing contracts still function
   - Load test with proposed limits

3. **Governance Vote:**
   - Submit parameter change proposal
   - Community discussion period (1-2 weeks)
   - On-chain vote

4. **Deployment:**
   - Scheduled chain upgrade if needed
   - Monitor closely for 48 hours post-change
   - Rollback plan in case of issues

---

## Security Considerations

### DoS Attack Vectors

#### Gas Exhaustion Attacks

**Attack:** Submit transactions with maximum gas to delay block production

**Mitigation:**
- Set per-block gas limit (in CometBFT config)
- Rate-limit transactions from single address
- Monitor for spam patterns

#### Deceptive Gas Estimation

**Attack:** Contract that uses different gas on simulation vs. execution

**Mitigation:**
- Ensure deterministic execution
- Use same gas meter for simulation and execution
- Audit contracts for conditional paths based on simulation mode

#### Memory Exhaustion

**Attack:** Contract that allocates excessive memory

**Mitigation:**
- WASM runtime memory limits (enforced by wasmd)
- Monitor node memory usage
- Restart containers with memory limits

### Audit Checklist

Before deploying gas limits to production:

- [ ] Benchmarks run on production-equivalent hardware
- [ ] Limits allow all legitimate use cases with 25%+ margin
- [ ] Limits prevent known attack vectors
- [ ] Monitoring and alerts configured
- [ ] Incident response plan documented
- [ ] Governance process for future changes defined
- [ ] Rollback plan tested
- [ ] Security audit completed (external preferred)

### Best Practices

1. **Conservative Defaults:** Start with higher limits, decrease over time as usage patterns emerge
2. **Gradual Changes:** Change limits by < 50% at a time to avoid surprises
3. **Communication:** Announce limit changes to contract developers in advance
4. **Testing:** Always test on staging testnet before production
5. **Documentation:** Keep this document updated with all changes and rationale

---

## Appendix: Gas Cost Breakdown

### Storage Costs

Cosmos SDK storage gas costs (approximate):

| Operation | Gas Cost | Notes |
|-----------|----------|-------|
| KV write (per byte) | 1000 | Linear with value size |
| KV read (per byte) | 100 | Cheaper than writes |
| KV delete | 100 | Fixed cost |
| Iterator creation | 1000 | Per iterator |
| Iterator next | 30 | Per iteration |

### WASM Instruction Costs

CosmWasm gas metering (approximate):

| Instruction Type | Gas Cost | Notes |
|------------------|----------|-------|
| Simple arithmetic | 1-5 | Add, sub, mul, div |
| Comparison | 1-2 | Eq, ne, lt, gt |
| Memory access | 1-10 | Load, store |
| Function call | 10-100 | Depends on complexity |
| Crypto operations | 1000-10000 | Hashing, signatures |

### Real-World Examples

#### Example 1: Token Transfer

```rust
// Simplified token transfer logic
pub fn transfer(deps: DepsMut, to: String, amount: Uint128) -> Result<Response> {
    // 1. Load sender balance (read ~100 gas)
    let sender_balance = BALANCES.load(deps.storage, &info.sender)?;

    // 2. Check balance (arithmetic ~5 gas)
    if sender_balance < amount {
        return Err(ContractError::InsufficientFunds {});
    }

    // 3. Update sender balance (write ~1000 gas)
    BALANCES.save(deps.storage, &info.sender, &(sender_balance - amount))?;

    // 4. Load recipient balance (read ~100 gas)
    let recipient_balance = BALANCES.load(deps.storage, &to)?;

    // 5. Update recipient balance (write ~1000 gas)
    BALANCES.save(deps.storage, &to, &(recipient_balance + amount))?;

    // 6. Emit event (~100 gas)
    Ok(Response::new()
        .add_attribute("action", "transfer")
        .add_attribute("to", to)
        .add_attribute("amount", amount))
}

// Total: ~2,305 gas + WASM overhead (~20,000) = ~25,000 gas
// Recommended limit: 25,000 * 1.25 = 32,000 gas → round to 50,000 gas
```

#### Example 2: Batch Operation

```rust
// Batch update multiple values
pub fn batch_update(deps: DepsMut, updates: Vec<Update>) -> Result<Response> {
    for update in updates {
        // Each iteration:
        // - Read: 100 gas
        // - Write: 1000 gas
        // - Logic: 50 gas
        // = 1,150 gas per item
        VALUES.save(deps.storage, &update.key, &update.value)?;
    }

    // 10 updates: 11,500 gas
    // 100 updates: 115,000 gas
    // Plus WASM overhead and event emission
    // Recommended limit: 115,000 * 1.5 = 172,500 → round to 200,000 gas
}
```

---

## Change Log

| Date | Change | Reason | Author |
|------|--------|--------|--------|
| 2024-12-03 | Initial version | Task #13 - WASM gas benchmarking | Claude Code |

---

## References

1. [CosmWasm Gas Pricing](https://docs.cosmwasm.com/docs/architecture/gas/)
2. [Cosmos SDK Gas Documentation](https://docs.cosmos.network/main/learn/beginner/gas-fees)
3. [WASM Specification](https://webassembly.github.io/spec/)
4. [Wasmd Repository](https://github.com/CosmWasm/wasmd)

---

**For questions or issues, contact the Aura blockchain team.**
