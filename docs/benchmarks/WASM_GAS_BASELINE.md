# WASM Gas Consumption Baselines

**Last Updated:** December 3, 2024
**Test Environment:** Local development machine
**Chain Version:** Cosmos SDK 0.53.4 + CometBFT
**Contract Versions:** Optimized with `wasm-opt`

## Purpose

This document records baseline gas consumption measurements for WASM operations on the Aura blockchain. These baselines inform gas limit configuration and help detect performance regressions.

## Test Environment

### Hardware Specifications

```
CPU: [To be filled after benchmarks]
RAM: [To be filled after benchmarks]
Disk: [To be filled after benchmarks]
OS: Linux (WSL2 on Windows)
```

### Software Versions

```
Go: 1.21+
Rust: [To be filled]
CosmWasm: wasmd integration
Optimization: wasm-opt -Os
```

## Benchmark Execution

### Running Benchmarks

```bash
# From project root
./scripts/benchmark-wasm-gas.sh

# Or directly
cd chain
go test -bench=^BenchmarkWasm -benchmem -benchtime=100x \
  ./x/wasm/keeper -timeout 30m
```

### Test Contracts

| Contract | Size | Description | Path |
|----------|------|-------------|------|
| binding_tester.wasm | 157 KB | Basic binding tests | `contracts/artifacts/binding_tester.wasm` |
| vc_issuer.wasm | 230 KB | VC issuance logic | `contracts/artifacts/vc_issuer.wasm` |
| schema.wasm | 236 KB | Schema validation | `contracts/artifacts/schema.wasm` |

## Baseline Results

### Store Code Operations

**Benchmark:** `BenchmarkWasmStoreCode`

| Contract | Size (KB) | Gas Consumed | Gas/Byte | Ops/Sec | Notes |
|----------|-----------|--------------|----------|---------|-------|
| binding_tester | 157 | TBD | TBD | TBD | Small contract |
| vc_issuer | 230 | TBD | TBD | TBD | Medium contract |
| schema | 236 | TBD | TBD | TBD | Medium contract |

**Expected Range:** 8,000,000 - 18,000,000 gas for 150-250KB contracts

**Factors Affecting Cost:**
- Contract bytecode size (linear relationship)
- WASM validation complexity
- Storage write amplification
- Compression efficiency

### Instantiate Contract Operations

**Benchmark:** `BenchmarkWasmInstantiateContract`

| Init Complexity | Init Msg Size | Gas Consumed | Ops/Sec | Notes |
|----------------|---------------|--------------|---------|-------|
| Simple (admin only) | ~50 bytes | TBD | TBD | Minimal state |
| Complex (with config) | ~150 bytes | TBD | TBD | Configuration state |

**Expected Range:** 500,000 - 3,000,000 gas for typical initialization

**Factors Affecting Cost:**
- Number of initial storage writes
- Initialization logic complexity
- Constructor computations
- Event emission overhead

### Execute Contract Operations

**Benchmark:** `BenchmarkWasmExecuteContract`

| Operation Type | Msg Size | Gas Consumed | Ops/Sec | Notes |
|----------------|----------|--------------|---------|-------|
| Simple query | ~30 bytes | TBD | TBD | Read-only |
| Simple write | ~80 bytes | TBD | TBD | Single KV update |
| Batch write | ~300 bytes | TBD | TBD | 5 KV updates |
| Compute heavy | ~50 bytes | TBD | TBD | 1000 iterations |

**Expected Range:** 100,000 - 5,000,000 gas depending on operation type

**Factors Affecting Cost:**
- Number of storage reads/writes
- Computational complexity
- Cross-contract message calls
- Event emission count

### Full Lifecycle

**Benchmark:** `BenchmarkWasmFullLifecycle`

| Metric | Gas Consumed | % of Total | Notes |
|--------|--------------|------------|-------|
| Store code | TBD | TBD% | Upload bytecode |
| Instantiate | TBD | TBD% | Initialize contract |
| Execute | TBD | TBD% | Run method |
| **Total** | **TBD** | **100%** | Complete lifecycle |

**Expected Total:** 10,000,000 - 25,000,000 gas for complete lifecycle

### Reentrancy Protection Overhead

**Benchmark:** `BenchmarkWasmReentrancyProtection`

| Metric | Gas Consumed | Overhead | Notes |
|--------|--------------|----------|-------|
| Execute with protection | TBD | TBD | Call stack tracking |
| Expected overhead | ~50,000 - 100,000 | ~5-10% | Security tax |

**Analysis:** Reentrancy protection adds minimal overhead while preventing critical vulnerabilities.

### Admin Operations

**Benchmark:** `BenchmarkWasmAdminOperations`

| Operation | Gas Consumed | Ops/Sec | Notes |
|-----------|--------------|---------|-------|
| SetContractAdmin | TBD | TBD | Single storage write |
| GetContractAdmin | TBD | TBD | Single storage read |
| IsContractAdmin | TBD | TBD | Read + comparison |

**Expected Range:** 20,000 - 150,000 gas for admin operations

**Analysis:** Admin operations are lightweight storage operations with minimal computational overhead.

## Gas Limit Recommendations

Based on benchmarks and safety margins:

### Conservative Limits (Production)

```toml
[wasm]
max_gas_store_code = 20_000_000      # Supports up to ~330KB contracts
max_gas_instantiate = 10_000_000     # Complex initialization
max_gas_execute = 15_000_000         # Cross-contract calls
max_gas_migrate = 10_000_000         # State migration
max_gas_admin = 1_000_000            # Admin operations
query_gas_limit = 3_000_000          # Query protection
```

**Safety Margin:** 25-30% above measured maximums

### Aggressive Limits (Development)

```toml
[wasm]
max_gas_store_code = 50_000_000      # Allow experimentation
max_gas_instantiate = 20_000_000     # Complex testing
max_gas_execute = 30_000_000         # Stress testing
max_gas_migrate = 20_000_000         # Migration testing
max_gas_admin = 5_000_000            # Development overhead
query_gas_limit = 10_000_000         # Flexible queries
```

**Safety Margin:** 50-100% above measured maximums for development flexibility

## Performance Trends

### Gas Consumption by Contract Size

```
Expected linear relationship:
gas_cost ≈ base_overhead + (contract_size_bytes * gas_per_byte)

Store Code:
  base_overhead ≈ 2,000,000 gas
  gas_per_byte ≈ 60 gas/byte

Example:
  200KB contract: 2,000,000 + (200 * 1024 * 60) = 14,288,000 gas
```

### Gas Consumption by Operation Complexity

```
Instantiate:
  minimal: 500,000 gas (admin only)
  simple: 1,500,000 gas (< 10 storage writes)
  moderate: 3,000,000 gas (10-50 storage writes)
  complex: 8,000,000 gas (> 50 storage writes)

Execute:
  read-only: 100,000 - 500,000 gas
  single write: 500,000 - 1,500,000 gas
  batch write: 1,500,000 - 5,000,000 gas
  cross-contract: 10,000,000+ gas
```

## Regression Detection

### Monitoring Thresholds

Set alerts if gas consumption increases by more than:

- **10%** for same operation over 1 week
- **25%** for same operation over 1 month
- **50%** for any single operation

### Investigation Triggers

Investigate if:

1. Any operation exceeds documented maximum by > 20%
2. 95th percentile gas usage increases > 15% week-over-week
3. Out-of-gas errors exceed 1% of transactions
4. New contract pattern consumes > 2x expected gas

### Benchmark Update Schedule

- **Weekly:** Quick smoke test (5 operations)
- **Monthly:** Full benchmark suite
- **Quarterly:** Full suite with new contracts
- **After Chain Upgrade:** Mandatory full suite

## Historical Data

### December 2024 - Initial Baseline

**Status:** Benchmarks to be run

**Action Items:**
- [ ] Run `./scripts/benchmark-wasm-gas.sh`
- [ ] Fill in TBD values in tables above
- [ ] Analyze results and update gas limits if needed
- [ ] Validate limits on local testnet
- [ ] Document any anomalies or outliers

### [Future Month] - Update

*Template for future updates:*

**Changes:**
- List any contract updates
- List any chain upgrades
- List any WASM runtime changes

**Results:**
- Comparative analysis vs. previous baseline
- % change in gas consumption
- New limits (if any)

**Notes:**
- Any performance improvements
- Any regressions and root causes
- Recommendations

## Notes and Observations

### Optimization Opportunities

*To be filled after benchmarks:*

1. **Contract Optimization:**
   - Contracts consuming > 2x expected gas
   - Optimization techniques applied
   - Results after optimization

2. **Runtime Optimization:**
   - WASM runtime configuration changes
   - Cache tuning results
   - Memory management improvements

3. **Storage Optimization:**
   - KV store access patterns
   - Batch operations vs. individual
   - Iterator efficiency

### Known Issues

*To be documented:*

1. **Issue:** [Description]
   - **Impact:** [Gas overhead %]
   - **Workaround:** [Temporary solution]
   - **Fix:** [Planned resolution]

### Best Practices Identified

*To be compiled from benchmark analysis:*

1. **Contract Design:**
   - Minimize storage writes in constructors
   - Use batch operations where possible
   - Cache frequently accessed values

2. **Gas Estimation:**
   - Add 25-30% safety margin to benchmarks
   - Test with maximum expected data sizes
   - Consider worst-case execution paths

3. **Performance:**
   - Reuse contract instances when possible
   - Minimize cross-contract calls
   - Use efficient data structures

## References

1. [WASM Gas Tuning Guide](./WASM_GAS_TUNING.md)
2. [CosmWasm Gas Metering](https://docs.cosmwasm.com/docs/architecture/gas/)
3. [Benchmark Scripts](../../scripts/benchmark-wasm-gas.sh)
4. [Test Implementation](../../chain/x/wasm/keeper/gas_benchmark_test.go)

## Change Log

| Date | Changes | Author |
|------|---------|--------|
| 2024-12-03 | Initial baseline template created | Claude Code |
| [TBD] | First benchmark results | [TBD] |

---

**Next Update Due:** After first benchmark run

**Responsible:** Aura blockchain development team
