---
status: pending
priority: p3
issue_id: "119"
tags: [code-review, testing, performance, benchmarks]
dependencies: ["100"]
---

# P3 MEDIUM: Performance Benchmark Tests Missing

## Problem Statement

No benchmark tests exist to measure and track performance of critical operations, making it impossible to detect performance regressions.

**Why it matters:** Without benchmarks, performance degradation goes unnoticed until production issues occur.

## Findings

### Missing Benchmarks

| Operation | Priority | Current Benchmark |
|-----------|----------|-------------------|
| Swap execution | Critical | None |
| Order matching | Critical | None |
| Signature verification | Critical | None |
| ZK proof verification | High | None |
| State serialization | High | None |
| Query operations | Medium | None |

### Performance-Critical Paths

**1. DEX Swap Path**
```
User → MsgSwap → ValidateBasic → DeductFees → ExecuteSwap → UpdatePool → EmitEvent
```

**2. Batch Execution Path**
```
EndBlocker → GetQueuedOrders → Sort → Match → Execute × N → UpdateState
```

**3. Bridge Verification Path**
```
MsgDeposit → ParseSignatures → RecoverPubKeys × N → VerifyThreshold → Mint
```

## Proposed Solutions

### Solution A: Add Comprehensive Benchmarks (Recommended)
**Effort:** 2-3 days | **Risk:** Low

```go
// chain/x/dex/keeper/benchmark_test.go

func BenchmarkSwap_SmallPool(b *testing.B) {
    ctx, keeper := setupBenchmark(b)
    pool := createPool(ctx, keeper, 1000, 1000)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = keeper.ExecuteSwap(ctx, pool.ID, sdk.NewInt(10), sdk.ZeroInt())
    }
}

func BenchmarkSwap_LargePool(b *testing.B) {
    ctx, keeper := setupBenchmark(b)
    pool := createPool(ctx, keeper, 1_000_000_000, 1_000_000_000)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = keeper.ExecuteSwap(ctx, pool.ID, sdk.NewInt(1_000_000), sdk.ZeroInt())
    }
}

func BenchmarkBatchExecution_100Orders(b *testing.B) {
    ctx, keeper := setupBenchmark(b)
    createQueuedOrders(ctx, keeper, 100)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = keeper.ExecuteBatch(ctx)
    }
}

func BenchmarkBatchExecution_10000Orders(b *testing.B) {
    ctx, keeper := setupBenchmark(b)
    createQueuedOrders(ctx, keeper, 10000)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = keeper.ExecuteBatch(ctx)
    }
}

func BenchmarkSignatureVerification(b *testing.B) {
    ctx, keeper := setupBenchmark(b)
    sig, msg := createTestSignature()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = keeper.VerifySignature(ctx, msg, sig)
    }
}

func BenchmarkZKProofVerification(b *testing.B) {
    ctx, keeper := setupBenchmark(b)
    proof := generateTestProof()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = keeper.VerifyProof(ctx, "age", proof, publicInputs)
    }
}
```

### Benchmark CI Integration

```yaml
# .github/workflows/benchmark.yml
name: Performance Benchmarks

on:
  pull_request:
    paths:
      - 'chain/**'

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run benchmarks
        run: |
          cd chain
          go test -bench=. -benchmem ./... | tee benchmark.txt
      - name: Compare with baseline
        uses: benchmark-action/github-action-benchmark@v1
```

## Recommended Action

**GO WITH SOLUTION A**: Add benchmarks for all critical paths.

## Technical Details

### Files to Create

- `chain/x/dex/keeper/benchmark_test.go`
- `chain/x/bridge/keeper/benchmark_test.go`
- `chain/x/zkp/keeper/benchmark_test.go`
- `chain/x/identity/keeper/benchmark_test.go`

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkSwap -benchmem ./x/dex/...

# Compare results
go test -bench=. -count=5 ./... > new.txt
benchstat old.txt new.txt
```

## Acceptance Criteria

- [ ] Swap execution benchmark
- [ ] Batch execution benchmarks (100, 1000, 10000 orders)
- [ ] Signature verification benchmark
- [ ] ZK proof verification benchmark
- [ ] Query operation benchmarks
- [ ] Baseline numbers documented
- [ ] CI integration (optional for testnet)

## Work Log

| Date | Action | Result |
|------|--------|--------|
| 2025-12-08 | Performance analysis identified missing benchmarks | P3 Medium |

## Resources

- [Go Benchmarking](https://pkg.go.dev/testing#hdr-Benchmarks)
- [benchstat Tool](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
