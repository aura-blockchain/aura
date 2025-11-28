# WASM + Contract Registry Integration Tests

Comprehensive integration test suite validating the complete integration between the WASM module and Contract Registry module in the AURA blockchain.

## Overview

This test suite provides production-grade end-to-end testing for:
- Contract upload, instantiation, and execution flow
- Security enforcement (VCs, KYC, compliance)
- Rate limiting and DoS protection
- Contract lifecycle management
- Performance benchmarking
- Security attack simulations

## Test Files

### 1. `wasm_registry_test.go`
Main integration test suite with 15+ comprehensive test scenarios.

**Test Categories:**
- **Basic Flow Tests** (Tests 1-5): Core contract operations
- **Security Enforcement Tests** (Tests 6-10): Validation and access control
- **Advanced Tests** (Tests 11-15): Complex scenarios and edge cases

**Coverage:**
- Contract upload → instantiate → execute → query
- Auto-registration in contract registry
- Metrics tracking and verification
- Security policy enforcement
- Lifecycle management (pause, unpause, deprecate)
- Concurrent execution safety
- Large-scale performance (1000+ calls)
- Graceful degradation

### 2. `wasm_test_utils.go`
Reusable test utilities and helper functions.

**Key Components:**
- `WASMTestContext`: Complete test environment setup
- `TestUser`: User account with configurable credentials
- Mock keepers for VC, Compliance, and Confidence Score modules
- Contract setup helpers
- Execution wrappers
- Assertion utilities

**Helper Functions:**
| Function | Purpose |
|----------|---------|
| `SetupTestAppWithWasm()` | Initialize test app with all modules |
| `CreateAuthorizedUploader()` | Create user authorized to upload contracts |
| `CreateUserWithVC()` | Create user with specific VC type |
| `CreateUserWithKYCLevel()` | Create user with specific KYC level |
| `CreateUserWithConfidenceScore()` | Create user with specific confidence score |
| `SetupCompleteContract()` | Deploy and register contract in one step |
| `ExecuteAsUser()` | Execute contract as specific user |
| `AssertMetricsEqual()` | Verify metrics accuracy |
| `AssertExecutionBlocked()` | Verify security enforcement |

### 3. `wasm_bench_test.go`
Performance benchmarks measuring execution speed and overhead.

**Benchmarks:**

| Benchmark | Purpose | Baseline Target |
|-----------|---------|-----------------|
| `BenchmarkContractExecution` | Raw execution speed | <1ms per op |
| `BenchmarkWithValidation` | Overhead of validation | <2ms per op |
| `BenchmarkContractRegistration` | Registration performance | <500μs per op |
| `BenchmarkConcurrentExecution` | Parallel execution (1/2/4/8/16 threads) | Linear scaling |
| `BenchmarkMetricsUpdate` | Metrics overhead | <100μs per op |
| `BenchmarkRateLimitCheck` | Rate limit checking | <50μs per op |
| `BenchmarkValidationOnly` | Pure validation overhead | <500μs per op |
| `BenchmarkFullStackExecution` | Complete stack | <3ms per op |
| `BenchmarkMemoryUsage` | Memory allocation patterns | <10KB per op |

**Comparative Benchmarks:**
- `BenchmarkCompareValidationLevels`: Compare NoValidation vs VCOnly vs KYCOnly vs FullValidation
- `BenchmarkScalability`: Test with 10/100/1000 registered contracts

### 4. `wasm_security_test.go`
Security attack simulations and edge case testing.

**Attack Simulations:**

| Test | Attack Vector | Protection Mechanism |
|------|---------------|----------------------|
| `TestDoSViaRapidExecution` | Rapid execution attempts | Rate limiting |
| `TestDoSViaLargePayload` | 10MB+ contract upload | Size limits |
| `TestDoSViaGasExhaustion` | Excessive gas requests | Gas limits |
| `TestVCBypassAttempt` | Missing/wrong VCs | VC enforcement |
| `TestVCRevocationDuringExecution` | VC revoked mid-execution | Real-time validation |
| `TestPausedContractExecutionAttempts` | Execute paused contract | Status checks |
| `TestUnauthorizedUploadAttempt` | Upload without authorization | Authorization checks |
| `TestUnauthorizedPauseAttempt` | Pause by non-admin | Admin verification |
| `TestMaliciousContractRegistration` | Malformed registration data | Input validation |
| `TestContractLimitEnforcement` | Exceed creator contract limit | Limit enforcement |
| `TestConcurrentPauseUnpause` | Race conditions | Thread safety |
| `TestRateLimitWindowReset` | Rate limit timing attacks | Window management |

## Running Tests

### Run All Integration Tests
```bash
cd /home/decri/blockchain-projects/aura/chain/testing/integration
go test -v -run TestWASMRegistryIntegration
```

### Run Security Tests
```bash
go test -v -run TestWASMSecurity
```

### Run Specific Test
```bash
go test -v -run TestWASMRegistryIntegration/TestUploadContractCode
```

### Run All Tests with Coverage
```bash
go test -v -coverprofile=coverage_wasm_integration.out
go tool cover -html=coverage_wasm_integration.out -o coverage_wasm_integration.html
```

### Run Benchmarks
```bash
# All benchmarks
go test -bench=. -benchmem -benchtime=10s

# Specific benchmark
go test -bench=BenchmarkContractExecution -benchmem -benchtime=10s

# CPU profiling
go test -bench=BenchmarkFullStackExecution -cpuprofile=cpu.prof
go tool pprof cpu.prof

# Memory profiling
go test -bench=BenchmarkMemoryUsage -memprofile=mem.prof
go tool pprof mem.prof
```

### Run with Race Detection
```bash
go test -race -v
```

## Expected Performance Benchmarks

Based on production requirements:

### Execution Performance
- **Contract Execution (baseline)**: <1ms per operation
- **With Full Validation**: <3ms per operation
- **Validation Overhead**: <40% increase from baseline

### Throughput
- **Sequential**: >300 ops/sec
- **Concurrent (8 threads)**: >2000 ops/sec

### Memory
- **Per Execution**: <10KB allocations
- **Registry Overhead**: <5KB per contract

### Scalability
- **1000 contracts**: <10% degradation
- **10,000 contracts**: <25% degradation

## Test Scenarios Coverage

### ✅ Basic Operations
- [x] Upload contract code
- [x] Instantiate contract
- [x] Execute contract
- [x] Query contract
- [x] Verify metrics

### ✅ Security Enforcement
- [x] VC requirement enforcement
- [x] KYC level enforcement
- [x] Confidence score enforcement
- [x] Blacklist/whitelist enforcement
- [x] Rate limiting
- [x] Gas limits
- [x] Contract size limits

### ✅ Access Control
- [x] Authorized uploaders only
- [x] Admin-only pause/unpause
- [x] Admin-only metadata updates
- [x] Governance authority verification

### ✅ Lifecycle Management
- [x] Contract registration
- [x] Pause contract
- [x] Unpause contract
- [x] Deprecate contract
- [x] Status transitions

### ✅ Performance
- [x] Concurrent execution
- [x] Large-scale execution (1000+ calls)
- [x] Memory efficiency
- [x] Scalability (1000+ contracts)

### ✅ Security Attacks
- [x] DoS via rapid execution
- [x] DoS via large payloads
- [x] DoS via gas exhaustion
- [x] VC bypass attempts
- [x] Unauthorized operations
- [x] Race conditions
- [x] Time-based attacks

## Adding New Tests

### 1. Add to Integration Test Suite

```go
// In wasm_registry_test.go
func (s *WASMRegistryTestSuite) TestYourNewFeature() {
    ctx := s.GetContext()

    // Setup
    uploader := s.CreateAuthorizedUploader()
    contractAddr := s.SetupCompleteContract(uploader, nil)

    // Test your feature
    result, err := s.YourFunction(ctx, contractAddr)
    s.Require().NoError(err)
    s.Require().NotNil(result)

    // Verify state
    s.AssertMetricsEqual(contractAddr.String(), 1, 1, 0)

    s.T().Log("✓ Your test - PASSED")
}
```

### 2. Add Helper Function

```go
// In wasm_test_utils.go
func (wtc *WASMTestContext) YourHelperFunction(params) {
    // Implementation
}
```

### 3. Add Benchmark

```go
// In wasm_bench_test.go
func BenchmarkYourFeature(b *testing.B) {
    ctx := setupBenchmarkContext(b)
    defer cleanupBenchmarkContext(b, ctx)

    // Setup

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        // Benchmark code
    }

    b.StopTimer()
    reportBenchmarkStats(b, "Your Feature", b.N)
}
```

### 4. Add Security Test

```go
// In wasm_security_test.go
func (s *WASMSecurityTestSuite) TestYourSecurityScenario() {
    ctx := s.GetContext()

    // Setup attack scenario

    // Attempt attack
    err := s.AttemptAttack()

    // Verify blocked
    s.Require().Error(err)
    s.Require().Contains(err.Error(), "expected error")

    s.T().Log("✓ Your security test - PASSED")
}
```

## Troubleshooting

### Common Issues

#### 1. Test Fails with "not authorized"
**Solution**: Ensure user is created with `CreateAuthorizedUploader()` for upload operations.

```go
// ❌ Wrong
user := s.CreateUserWithoutVC()
codeID := s.UploadTestContract(user) // Fails

// ✅ Correct
uploader := s.CreateAuthorizedUploader()
codeID := s.UploadTestContract(uploader)
```

#### 2. Validation Fails with "missing VC"
**Solution**: Create user with required VCs matching contract metadata.

```go
metadata := &crtypes.ContractMetadata{
    RequiresVc: true,
    RequiredVcTypes: []string{"KYCVerification"},
}
contractAddr := s.SetupCompleteContract(uploader, metadata)

// Create user with required VC
user := s.CreateUserWithVC("KYCVerification")
```

#### 3. Rate Limit Tests Fail
**Solution**: Ensure rate limit policy is configured and incremented correctly.

```go
policy := &crtypes.SecurityPolicy{
    RateLimitPerUser: 5,
}
contractAddr := s.SetupCompleteContractWithPolicy(uploader, nil, policy)

// Must increment BEFORE checking next time
s.RegistryKeeper.IncrementRateLimit(ctx, contractAddr.String(), user.GetAddress().String())
```

#### 4. Benchmarks Show High Variance
**Solution**: Increase benchmark time and disable CPU throttling.

```bash
# Longer benchmark time
go test -bench=. -benchtime=30s

# Run with performance governor
sudo cpupower frequency-set -g performance
```

#### 5. Memory Benchmark Inaccurate
**Solution**: Force GC before benchmark and use larger sample size.

```bash
go test -bench=BenchmarkMemoryUsage -benchtime=100000x -benchmem
```

## CI/CD Integration

### GitHub Actions Workflow

```yaml
name: WASM Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run Integration Tests
        run: |
          cd chain/testing/integration
          go test -v -race -coverprofile=coverage.out

      - name: Run Security Tests
        run: |
          cd chain/testing/integration
          go test -v -run TestWASMSecurity

      - name: Run Benchmarks
        run: |
          cd chain/testing/integration
          go test -bench=. -benchmem -benchtime=5s

      - name: Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./chain/testing/integration/coverage.out
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running WASM integration tests..."
cd chain/testing/integration
go test -v -short || exit 1

echo "Running quick benchmarks..."
go test -bench=BenchmarkContractExecution -benchtime=1s || exit 1

echo "All tests passed!"
```

## Production Readiness Checklist

- [x] All 15+ integration test scenarios passing
- [x] Security tests covering all attack vectors
- [x] Performance benchmarks meeting targets
- [x] Concurrent execution safety verified
- [x] Race condition testing with `-race` flag
- [x] Memory leak testing
- [x] Edge case coverage
- [x] Error handling verification
- [x] Metrics accuracy validation
- [x] Documentation complete

## Test Coverage Goals

- **Line Coverage**: >80%
- **Branch Coverage**: >75%
- **Critical Path Coverage**: 100%

### Generate Coverage Report

```bash
# Generate coverage
go test -coverprofile=coverage.out

# View summary
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# View in browser
open coverage.html
```

## Performance Regression Testing

### Baseline Establishment

```bash
# Capture baseline
go test -bench=. -benchmem > baseline.txt

# After changes, compare
go test -bench=. -benchmem > current.txt
benchcmp baseline.txt current.txt
```

### Automated Regression Detection

```bash
# Install benchstat
go install golang.org/x/perf/cmd/benchstat@latest

# Compare
benchstat baseline.txt current.txt
```

**Acceptable Thresholds:**
- Execution time: ±10%
- Memory allocation: ±15%
- Throughput: -10% to +20%

## Contributing

When adding new features to WASM or Contract Registry:

1. **Add integration tests** in `wasm_registry_test.go`
2. **Add benchmarks** in `wasm_bench_test.go`
3. **Add security tests** if security-related
4. **Update this documentation**
5. **Verify all tests pass**: `go test -v -race`
6. **Verify benchmarks**: `go test -bench=.`
7. **Update coverage goals** if needed

## Support

For questions or issues with tests:
- Check troubleshooting section above
- Review test code for examples
- Check existing tests for similar scenarios
- File an issue with test output

## Versioning

- **Test Suite Version**: 1.0.0
- **Last Updated**: 2025-11-25
- **Compatible with**: AURA Chain v1.x
- **Requires**: Go 1.21+
