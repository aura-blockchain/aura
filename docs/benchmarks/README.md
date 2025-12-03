# Aura Blockchain Benchmarks

This directory contains benchmark documentation and results for the Aura blockchain.

## Directory Contents

### WASM Benchmarks

- **[WASM_GAS_TUNING.md](./WASM_GAS_TUNING.md)** - Comprehensive guide for tuning WASM gas limits
  - Understanding gas consumption
  - Benchmark methodology
  - Recommended gas limits
  - App.toml configuration
  - Monitoring and adjustment procedures

- **[WASM_GAS_BASELINE.md](./WASM_GAS_BASELINE.md)** - Baseline measurements and historical data
  - Benchmark results for each WASM operation
  - Performance trends over time
  - Regression detection thresholds
  - Optimization opportunities

- **Benchmark Results** - Time-stamped benchmark output files
  - `wasm_gas_benchmark_YYYYMMDD_HHMMSS.txt` - Raw benchmark data

## Quick Start

### Running WASM Gas Benchmarks

```bash
# From project root
./scripts/benchmark-wasm-gas.sh
```

This will:
1. Check for compiled contracts
2. Run comprehensive gas benchmarks
3. Generate detailed report
4. Provide tuning recommendations

### Reviewing Results

Results are saved to `wasm_gas_benchmark_YYYYMMDD_HHMMSS.txt` with:
- Gas consumption per operation
- Operations per second
- Memory usage
- Detailed metrics by benchmark type

### Updating Gas Limits

1. Review benchmark results
2. Calculate recommended limits (measured_max * 1.25)
3. Update `app.toml` in testnet configs:
   ```bash
   vim testnet-data/validator-1/config/app.toml
   # Edit [wasm] section with new limits
   ```
4. Test on local testnet before deploying to production

## Benchmark Categories

### 1. Store Code Benchmarks

Measures gas consumption for uploading contract bytecode.

**Key Metrics:**
- Total gas consumed
- Gas per byte of bytecode
- Storage write overhead

**Variables:**
- Contract size (157KB, 230KB, 236KB test contracts)
- WASM complexity
- Optimization level

### 2. Instantiate Contract Benchmarks

Measures gas consumption for contract instantiation.

**Key Metrics:**
- Initialization gas
- Gas per init message byte
- State setup overhead

**Variables:**
- Init message complexity (simple vs. complex)
- Initial state size
- Constructor logic

### 3. Execute Contract Benchmarks

Measures gas consumption for contract execution.

**Key Metrics:**
- Execution gas by operation type
- Gas per execution message byte
- Call stack overhead

**Variables:**
- Operation type (query, write, batch, compute)
- Message complexity
- Cross-contract calls

### 4. Full Lifecycle Benchmarks

Measures complete contract lifecycle from upload to execution.

**Key Metrics:**
- Total lifecycle gas
- Gas breakdown by phase
- End-to-end performance

### 5. Security Overhead Benchmarks

Measures gas cost of security features.

**Key Metrics:**
- Reentrancy protection overhead
- Admin operation costs
- Pause/unpause overhead

## Interpreting Benchmark Output

### Example Output

```
BenchmarkWasmStoreCode/SmallContract_157KB-8
Testing Binding tester contract (157KB) - Contract size: 160768 bytes (157.00 KB)
     100  12458923 ns/op  10245678 gas/op  160768 bytes  63.71 gas/byte
```

**Reading:**
- `100` = iterations run
- `12458923 ns/op` = 12.46ms per operation (wall-clock time)
- `10245678 gas/op` = ~10.2M gas consumed (blockchain cost)
- `160768 bytes` = contract size
- `63.71 gas/byte` = efficiency metric

### Gas Limit Calculation

From the above example:

```
Measured gas: 10,245,678
Safety margin: 25%
Recommended limit: 10,245,678 * 1.25 = 12,807,097
Rounded limit: 15,000,000 (for simplicity)
```

## Monitoring Gas Usage

### Prometheus Metrics

```promql
# Average gas by operation
avg(wasm_store_code_gas) by (contract_size_bucket)
avg(wasm_instantiate_gas) by (contract_id)
avg(wasm_execute_gas) by (contract_address, method)

# 95th percentile (detect outliers)
quantile(0.95, wasm_execute_gas) by (contract_address)

# Out-of-gas rate (should be < 1%)
rate(wasm_out_of_gas_errors_total[5m])
```

### Grafana Dashboards

Import dashboards from `/grafana/dashboards/`:
- `wasm-performance.json` - WASM operation performance
- `wasm-gas-consumption.json` - Gas usage by operation type
- `wasm-errors.json` - Error rates and failures

## Best Practices

### 1. Regular Benchmarking

- **Weekly:** Quick smoke tests (5-10 operations)
- **Monthly:** Full benchmark suite
- **After Chain Upgrade:** Mandatory comprehensive benchmarks

### 2. Environment Consistency

Run benchmarks on hardware similar to production validators:
- Same CPU architecture (x86_64)
- Similar CPU cores (4+)
- Similar memory (8GB+)
- SSD storage

### 3. Data Recording

Always record:
- Hardware specifications
- Software versions (Go, Rust, chain)
- Contract versions and optimization flags
- Timestamp and git commit hash

### 4. Regression Detection

Alert on:
- Gas consumption increase > 10% week-over-week
- Any operation exceeding documented max by > 20%
- New contracts consuming > 2x expected gas

### 5. Documentation Updates

After each benchmark run:
1. Update `WASM_GAS_BASELINE.md` with results
2. Document any anomalies in notes section
3. Update gas limits if needed
4. Commit changes with benchmark timestamp

## Troubleshooting

### Benchmarks Fail to Run

**Error:** `Contract file not found`

**Solution:**
```bash
cd contracts
make optimize-wasm
```

**Error:** `WASM keeper not configured`

**Solution:** Check test setup in `gas_benchmark_test.go` - may need to mock wasmd keeper properly

### Inconsistent Results

**Symptoms:** High variance between runs (> 30%)

**Causes:**
- CPU thermal throttling
- Background processes
- Disk I/O contention
- Swap usage

**Solutions:**
- Close other applications
- Run multiple iterations (`-benchtime=1000x`)
- Check system load during benchmarks
- Use dedicated benchmark machine

### Results Don't Match Production

**Symptoms:** Benchmarks show lower gas than production usage

**Causes:**
- Different contract versions
- Different state sizes
- Different execution paths
- Cold start overhead not measured

**Solutions:**
- Use production contract versions
- Test with realistic state sizes
- Profile actual transaction traces
- Include warmup iterations

## Contributing

### Adding New Benchmarks

1. Add benchmark function to `gas_benchmark_test.go`
2. Follow naming convention: `BenchmarkWasm<Operation>`
3. Use standard metrics: `gas/op`, `bytes`, `gas/byte`
4. Document in this README

### Updating Baseline Data

1. Run benchmarks: `./scripts/benchmark-wasm-gas.sh`
2. Update tables in `WASM_GAS_BASELINE.md`
3. Add entry to change log
4. Create PR with benchmark results file

### Reporting Issues

If you encounter:
- Unexpectedly high gas consumption
- Performance regressions
- Benchmark failures

Open an issue with:
- Benchmark output
- Hardware specs
- Software versions
- Steps to reproduce

## References

- [CosmWasm Documentation](https://docs.cosmwasm.com/)
- [Cosmos SDK Gas Docs](https://docs.cosmos.network/main/learn/beginner/gas-fees)
- [Go Benchmark Guide](https://pkg.go.dev/testing#hdr-Benchmarks)
- [WASM Specification](https://webassembly.github.io/spec/)

## License

Part of the Aura blockchain project. See main repository LICENSE.

---

**Last Updated:** December 3, 2024

**Maintainers:** Aura blockchain development team
