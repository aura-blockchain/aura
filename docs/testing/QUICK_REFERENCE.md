# Testing Quick Reference Guide

Quick reference for common testing tasks in the Aura blockchain project.

## Running Tests

### Basic Commands

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run with coverage
make test-coverage

# Run specific module
cd chain && go test ./x/vcregistry/... -v

# Run single test
cd chain && go test ./x/vcregistry/keeper -v -run TestKeeperIssueVC
```

### Advanced Testing

```bash
# Integration tests
make test-integration

# End-to-end tests
make test-e2e

# Stress tests
make test-stress

# Chaos tests
make test-chaos

# Benchmarks
make test-bench

# Race detector
make test-race
```

## Writing Tests

### Basic Test Template

```go
func TestMyFunction(t *testing.T) {
    // Setup
    ctx := testutil.SetupTestContext(t)

    // Execute
    result, err := MyFunction(ctx.SdkCtx)

    // Verify
    require.NoError(t, err)
    require.Equal(t, expected, result)
}
```

### Table-Driven Test

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid", "valid", false},
        {"invalid", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Validate(tt.input)
            if tt.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

## Coverage

```bash
# Generate coverage report
make test-coverage

# View in browser
cd chain && open coverage.html

# Check coverage threshold
cd chain && go tool cover -func=coverage.out | grep total
```

## CI/CD

### Local CI Simulation

```bash
# Run same tests as CI
make ci

# Individual checks
make fmt
make lint
make test-coverage
```

### GitHub Actions

Tests run automatically on:
- Push to main/master/develop
- Pull requests
- Daily at 2 AM UTC

## Common Test Utilities

```go
// Setup test context
ctx := testutil.SetupTestContext(t)

// Generate test address
addr := testutil.GenerateTestAddress()

// Generate multiple addresses
addrs := testutil.GenerateTestAddresses(10)

// Create test codec
codec := testutil.CreateTestCodec()

// Mock time
mockTime := testutil.MockTime()
```

## Debugging Tests

```bash
# Verbose output
go test ./... -v

# Run single test
go test ./x/module -run TestName -v

# Print values (only shows on failure or -v)
t.Logf("Value: %+v", value)

# Use delve debugger
dlv test ./x/module -- -test.run TestName
```

## Benchmarking

```bash
# Run benchmarks
make test-bench

# Run specific benchmark
cd chain && go test ./testing/benchmark -bench=BenchmarkTransactionProcessing

# With memory stats
cd chain && go test ./testing/benchmark -bench=. -benchmem

# Compare with baseline
cd chain && benchstat baseline.txt current.txt
```

## Test Tags

```bash
# Skip long tests
go test ./... -short

# Run integration tests
go test ./testing/integration/... -tags=integration

# Run E2E tests
go test ./testing/e2e/... -tags=e2e

# Run stress tests
go test ./testing/stress/... -tags=stress
```

## Best Practices

1. **Test Independence**: Each test should be independent
2. **Clear Names**: Use descriptive test names
3. **Setup/Teardown**: Use proper lifecycle management
4. **Error Messages**: Provide clear assertion messages
5. **Table Tests**: Use for multiple similar test cases
6. **Skip Long Tests**: Use `if testing.Short()` for long tests
7. **Race Detector**: Run with `-race` flag regularly
8. **Coverage**: Aim for 90%+ on critical paths

## Makefile Targets

```bash
make help              # Show all available targets
make build             # Build project
make test              # Run all tests
make test-unit         # Unit tests
make test-integration  # Integration tests
make test-e2e          # E2E tests
make test-stress       # Stress tests
make test-chaos        # Chaos tests
make test-coverage     # Coverage report
make test-race         # Race detector
make test-bench        # Benchmarks
make lint              # Run linter
make fmt               # Format code
make security-scan     # Security scan
make clean             # Clean artifacts
make ci                # CI pipeline
```

## File Locations

```
Testing Infrastructure:
├── chain/testing/coverage/      - Coverage framework
├── chain/testing/testutil/      - Test utilities
├── chain/testing/integration/   - Integration tests
├── chain/testing/e2e/           - E2E scenarios
├── chain/testing/stress/        - Load tests
├── chain/testing/chaos/         - Chaos engineering
├── chain/testing/benchmark/     - Performance benchmarks
└── docs/testing/                - Documentation
    ├── TESTING_GUIDE.md         - Complete guide
    ├── TESTNET_CONFIGURATION.md - Testnet setup
    ├── BUG_BOUNTY_PROGRAM.md    - Security program
    └── QUICK_REFERENCE.md       - This file
```

## Getting Help

- Read full guide: `docs/testing/TESTING_GUIDE.md`
- Check examples in existing tests
- Ask in #testing Discord channel
- Open GitHub issue with [testing] tag

## Checklist for New Features

- [ ] Write unit tests (90%+ coverage)
- [ ] Add integration test if cross-module
- [ ] Add E2E scenario if user-facing
- [ ] Run `make test-coverage` locally
- [ ] Run `make test-race` locally
- [ ] Update documentation if needed
- [ ] Ensure CI passes

---

Last Updated: January 13, 2025
