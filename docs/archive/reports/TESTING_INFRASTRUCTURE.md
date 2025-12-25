# AURA Blockchain Testing Infrastructure

## Overview

This document describes the comprehensive testing infrastructure for the AURA blockchain project. The testing infrastructure is designed to ensure code quality, security, and reliability across all 21 custom modules.

## Table of Contents

1. [Test Structure](#test-structure)
2. [Test Types](#test-types)
3. [Test Utilities](#test-utilities)
4. [Running Tests](#running-tests)
5. [Coverage Requirements](#coverage-requirements)
6. [CI/CD Integration](#cicd-integration)
7. [Writing Tests](#writing-tests)

## Test Structure

```
chain/
├── testing/
│   ├── testutil/           # Test utilities and helpers
│   │   ├── common.go       # Common test setup
│   │   ├── fixtures.go     # Test data fixtures
│   │   ├── mocks.go        # Mock implementations
│   │   ├── generators.go   # Test data generators
│   │   ├── assertions.go   # Custom assertions
│   │   ├── invariants.go   # Invariant checking
│   │   └── test_template.go # Test templates
│   ├── integration/        # Integration tests
│   │   ├── suite.go
│   │   ├── integration_test.go
│   │   └── comprehensive_integration_test.go
│   ├── e2e/               # End-to-end tests
│   │   ├── e2e_test.go
│   │   └── scenarios.go
│   ├── stress/            # Stress tests
│   │   ├── load_test.go
│   │   └── load_test_test.go
│   ├── benchmark/         # Benchmark tests
│   │   ├── benchmark.go
│   │   └── benchmark_test.go
│   ├── fuzz/              # Fuzz tests
│   │   └── fuzz_test.go
│   └── chaos/             # Chaos engineering tests
│       ├── chaos.go
│       └── chaos_test.go
└── x/                     # Module tests
    └── [module]/
        └── keeper/
            ├── keeper_test.go
            ├── msg_server_test.go
            ├── query_server_test.go
            ├── genesis_test.go
            └── invariants_test.go
```

## Test Types

### 1. Unit Tests

Unit tests verify individual components in isolation.

**Location**: `chain/x/[module]/keeper/*_test.go`

**Coverage Requirements**: >80% per module

**Test Files**:
- `keeper_test.go` - Core keeper functionality
- `msg_server_test.go` - Message handlers
- `query_server_test.go` - Query handlers
- `genesis_test.go` - Genesis import/export
- `invariants_test.go` - Module invariants

**Example**:
```go
func TestKeeperMethod(t *testing.T) {
    ctx := testutil.SetupTestContext(t)
    keeper := &Keeper{}

    result, err := keeper.SomeMethod(ctx.SdkCtx, params)
    require.NoError(t, err)
    require.NotNil(t, result)
}
```

### 2. Integration Tests

Integration tests verify interactions between multiple modules.

**Location**: `chain/testing/integration/`

**Test Scenarios**:
- VC Registry + Identity Change
- DEX + Bridge
- Inclusion Routines + Confidence Score + Prevalidation
- Governance + Economic Security
- Validator Security + Network Security
- Compliance + Privacy
- Data Registry + IPFS + Cryptography

**Example**:
```go
func (s *IntegrationTestSuite) TestModuleInteraction() {
    // Setup
    // Execute cross-module operation
    // Verify state consistency
}
```

### 3. End-to-End Tests

E2E tests verify complete transaction workflows from submission to finalization.

**Location**: `chain/testing/e2e/`

**Scenarios**:
- Full transaction lifecycle
- Upgrade paths
- Emergency response flows
- Multi-signature workflows

### 4. Stress Tests

Stress tests verify system behavior under high load.

**Location**: `chain/testing/stress/`

**Tests**:
- High transaction volume
- Concurrent operations
- Resource exhaustion scenarios
- Rate limiting verification

### 5. Benchmark Tests

Benchmark tests measure performance characteristics.

**Location**: `chain/testing/benchmark/`

**Metrics**:
- Transaction throughput (TPS)
- Memory usage
- CPU utilization
- Latency measurements

### 6. Fuzz Tests

Fuzz tests use randomized inputs to discover edge cases and vulnerabilities.

**Location**: `chain/testing/fuzz/`

**Targets**:
- Message validation
- Address validation
- Amount validation
- Cryptographic operations
- Signature verification
- JSON parsing

**Example**:
```go
func FuzzMessageValidation(f *testing.F) {
    f.Add("did:aura:test", "data", int64(1000))
    f.Fuzz(func(t *testing.T, did string, data string, amount int64) {
        // Test should never panic
        defer func() {
            if r := recover(); r != nil {
                t.Errorf("Panic with inputs: %v", r)
            }
        }()
        // Validate inputs
    })
}
```

### 7. Security Tests

Security tests verify protection against attack vectors.

**Tests**:
- Double spend attempts
- Replay attacks
- Sybil attacks
- MEV attacks
- Permission bypasses
- Rate limit evasion

### 8. Chaos Tests

Chaos tests verify system resilience under failure conditions.

**Location**: `chain/testing/chaos/`

**Scenarios**:
- Network partitions
- Node failures
- Byzantine behavior
- Resource exhaustion

## Test Utilities

### Test Context Setup

```go
ctx := testutil.SetupTestContext(t)
// Returns TestContext with:
// - ctx.Ctx (context.Context)
// - ctx.SdkCtx (sdk.Context)
// - ctx.DB (in-memory database)
// - ctx.CMS (commit multi-store)
// - ctx.Logger
```

### Test Fixtures

```go
fixtures := testutil.NewTestFixtures()
// Provides:
// - fixtures.Addresses (test addresses)
// - fixtures.ValidatorAddrs (validator addresses)
// - fixtures.Amounts (test coin amounts)
// - fixtures.Timestamps (test timestamps)
```

### Mock Keepers

```go
bankKeeper := testutil.NewMockBankKeeper()
accountKeeper := testutil.NewMockAccountKeeper()
stakingKeeper := testutil.NewMockStakingKeeper()
```

### Random Data Generators

```go
addr := testutil.RandomAddress()
amount := testutil.RandomAmount("aura")
did := testutil.RandomDID()
ipfsHash := testutil.RandomIPFSHash()
timestamp := testutil.RandomTimestamp()
```

### Custom Assertions

```go
testutil.AssertEventEmitted(t, ctx, "event_type")
testutil.AssertBalanceEqual(t, expected, actual)
testutil.AssertCoinsEqual(t, expected, actual)
testutil.AssertStoreHasKey(t, store, key)
```

### Invariant Checking

```go
checker := testutil.NewInvariantChecker(t)
checker.RegisterInvariant(invariantFunc)
checker.CheckAll(ctx)
```

## Running Tests

### Run All Tests

```bash
cd chain
go test ./... -v
```

### Run Module Tests

```bash
# Run all tests for a specific module
go test ./x/auth/... -v

# Run specific test types
go test -run TestKeeper ./x/auth/...
go test -run TestMsgServer ./x/auth/...
go test -run TestQueryServer ./x/auth/...
```

### Run Integration Tests

```bash
go test -tags=integration ./testing/integration/... -v
```

### Run E2E Tests

```bash
go test ./testing/e2e/... -v -timeout=60m
```

### Run Stress Tests

```bash
go test ./testing/stress/... -v -timeout=30m
```

### Run Benchmark Tests

```bash
go test -bench=. -benchmem ./testing/benchmark/...
```

### Run Fuzz Tests

```bash
go test -fuzz=FuzzMessageValidation -fuzztime=5m ./testing/fuzz/...
```

### Run with Coverage

```bash
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -html=coverage.out -o coverage.html
go tool cover -func=coverage.out
```

### Run with Race Detection

```bash
go test ./... -race -v
```

## Coverage Requirements

### Module Coverage Goals

| Module | Target Coverage |
|--------|----------------|
| x/auth | >85% |
| x/bridge | >80% |
| x/compliance | >80% |
| x/confidencescore | >80% |
| x/cryptography | >85% |
| x/dataregistry | >80% |
| x/dex | >80% |
| x/economicsecurity | >80% |
| x/governance | >80% |
| All other modules | >75% |

### Coverage Analysis

```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out

# View coverage by function
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Check coverage threshold
go test ./... -coverprofile=coverage.out && \
  go tool cover -func=coverage.out | \
  grep total | \
  awk '{print $3}' | \
  sed 's/%//' | \
  awk '{if ($1 < 80) exit 1}'
```

## CI/CD Integration

### GitHub Actions Workflow

The CI pipeline runs the following test suites:

1. **Unit Tests** - All module tests (matrix across Go versions)
2. **Integration Tests** - Cross-module interaction tests
3. **Stress Tests** - Load and performance tests
4. **Fuzz Tests** - Randomized input testing
5. **E2E Tests** - Complete workflow tests
6. **Coverage** - Code coverage analysis
7. **Security** - Vulnerability scanning

### Test Matrix

- **Go Versions**: 1.24, 1.25
- **Modules**: All 21 modules tested separately
- **OS**: Ubuntu, macOS, Windows (for build)

### Coverage Reporting

- Coverage reports uploaded to Codecov
- HTML reports generated as artifacts
- Coverage threshold enforced (>80%)

## Writing Tests

### Test Naming Conventions

- Test functions: `Test[Component]_[Scenario]`
- Benchmark functions: `Benchmark[Operation]`
- Fuzz functions: `Fuzz[Target]`

### Test Structure

```go
func TestComponentName(t *testing.T) {
    // Setup
    ctx := testutil.SetupTestContext(t)
    fixtures := testutil.NewTestFixtures()

    // Test cases
    testCases := []struct {
        name     string
        input    interface{}
        expected interface{}
        wantErr  bool
    }{
        {"valid input", validInput, expectedOutput, false},
        {"invalid input", invalidInput, nil, true},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Execute
            result, err := functionUnderTest(tc.input)

            // Assert
            if tc.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
                require.Equal(t, tc.expected, result)
            }
        })
    }
}
```

### Best Practices

1. **Isolation**: Each test should be independent
2. **Determinism**: Tests should produce consistent results
3. **Coverage**: Test happy path, error cases, and edge cases
4. **Performance**: Keep unit tests fast (<100ms)
5. **Documentation**: Comment complex test scenarios
6. **Fixtures**: Use test utilities for common setup
7. **Assertions**: Use meaningful assertion messages
8. **Cleanup**: Ensure proper cleanup in teardown

### Module-Specific Tests

Each module should have:

1. **Keeper Tests**
   - Test all keeper methods
   - Test state persistence
   - Test error handling

2. **Message Server Tests**
   - Test all message handlers
   - Test input validation
   - Test authorization
   - Test state changes
   - Test event emission

3. **Query Server Tests**
   - Test all query handlers
   - Test pagination
   - Test filtering
   - Test error cases

4. **Genesis Tests**
   - Test genesis import
   - Test genesis export
   - Test round-trip consistency
   - Test validation

5. **Invariant Tests**
   - Test module invariants
   - Test cross-module invariants
   - Test state consistency

## Test Coverage Report

Generated automatically in CI/CD and available as artifact.

### Quick Commands

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run integration tests
make test-integration

# Run benchmarks
make test-bench

# Run fuzz tests
make test-fuzz
```

## Troubleshooting

### Common Issues

1. **Race conditions**: Run with `-race` flag
2. **Flaky tests**: Check for time dependencies
3. **Slow tests**: Profile with `-cpuprofile`
4. **Coverage gaps**: Use `go tool cover` to identify

### Debug Tests

```bash
# Verbose output
go test -v

# Run single test
go test -run TestSpecificTest

# Debug with Delve
dlv test -- -test.run TestSpecificTest
```

## Contributing

When adding new features:

1. Write tests first (TDD)
2. Ensure >80% coverage
3. Add integration tests for cross-module features
4. Update test documentation
5. Verify CI passes

## Maintenance

- Review and update tests with each release
- Remove obsolete tests
- Refactor common test patterns into utilities
- Monitor CI test duration
- Update coverage requirements as needed
