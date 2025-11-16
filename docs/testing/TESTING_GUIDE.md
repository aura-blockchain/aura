# Aura Blockchain Testing Guide

## Overview

This guide provides comprehensive information about the testing infrastructure for the Aura blockchain project.

## Table of Contents

1. [Testing Philosophy](#testing-philosophy)
2. [Test Types](#test-types)
3. [Running Tests](#running-tests)
4. [Writing Tests](#writing-tests)
5. [Coverage Requirements](#coverage-requirements)
6. [CI/CD Integration](#cicd-integration)
7. [Best Practices](#best-practices)

## Testing Philosophy

The Aura blockchain project follows a comprehensive testing strategy:

- **Unit Tests**: Test individual functions and components in isolation
- **Integration Tests**: Test interactions between multiple modules
- **End-to-End Tests**: Test complete user workflows
- **Stress Tests**: Verify performance under high load
- **Chaos Tests**: Ensure resilience under failure conditions
- **Regression Tests**: Prevent reintroduction of bugs

### Coverage Goals

- **Minimum**: 80% code coverage across all modules
- **Target**: 90% code coverage
- **Critical Paths**: 100% coverage required

## Test Types

### 1. Unit Tests

**Location**: `chain/x/*/keeper/*_test.go`, `chain/x/*/types/*_test.go`

**Purpose**: Test individual functions and methods

**Example**:
```go
func TestKeeperSetGet(t *testing.T) {
    ctx := testutil.SetupTestContext(t)
    keeper := NewKeeper(...)

    // Test Set
    err := keeper.Set(ctx.SdkCtx, "key", "value")
    require.NoError(t, err)

    // Test Get
    value, err := keeper.Get(ctx.SdkCtx, "key")
    require.NoError(t, err)
    require.Equal(t, "value", value)
}
```

**Run**:
```bash
cd chain
go test ./x/modulename/... -v
```

### 2. Integration Tests

**Location**: `chain/testing/integration/`

**Purpose**: Test module interactions

**Example**:
```go
func (s *IntegrationTestSuite) TestModuleInteraction() {
    // Setup
    s.SetupTest()

    // Test cross-module functionality
    // 1. Create identity
    // 2. Issue VC for identity
    // 3. Verify VC is linked

    s.Require().NotNil(s.App)
}
```

**Run**:
```bash
cd chain
go test ./testing/integration/... -v -tags=integration
```

### 3. End-to-End Tests

**Location**: `chain/testing/e2e/`

**Purpose**: Test complete workflows

**Example**:
```go
func TestIdentityLifecycle(t *testing.T) {
    scenario := IdentityLifecycleScenario()
    RunScenario(t, scenario)
}
```

**Run**:
```bash
cd chain
go test ./testing/e2e/... -v -tags=e2e
```

### 4. Stress Tests

**Location**: `chain/testing/stress/`

**Purpose**: Verify performance under load

**Run**:
```bash
cd chain
go test ./testing/stress/... -v -tags=stress -timeout 30m
```

### 5. Chaos Tests

**Location**: `chain/testing/chaos/`

**Purpose**: Test resilience with failure injection

**Run**:
```bash
cd chain
go test ./testing/chaos/... -v -tags=chaos
```

### 6. Benchmarks

**Location**: `chain/testing/benchmark/`

**Purpose**: Measure performance

**Run**:
```bash
cd chain
go test ./testing/benchmark/... -bench=. -benchmem
```

## Running Tests

### Run All Tests

```bash
cd chain
go test ./... -v
```

### Run Tests for Specific Module

```bash
cd chain
go test ./x/vcregistry/... -v
```

### Run Tests with Coverage

```bash
cd chain
go test ./... -v -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Run Tests with Race Detector

```bash
cd chain
go test ./... -v -race
```

### Run Short Tests Only

```bash
cd chain
go test ./... -v -short
```

### Run Specific Test

```bash
cd chain
go test ./x/vcregistry/keeper -v -run TestKeeperIssueVC
```

### Run Tests in Parallel

```bash
cd chain
go test ./... -v -parallel 4
```

## Writing Tests

### Test Structure

```go
package keeper

import (
    "testing"

    "github.com/stretchr/testify/require"
    "github.com/aequitas/aura/chain/testing/testutil"
)

func TestMyFunction(t *testing.T) {
    // Setup
    ctx := testutil.SetupTestContext(t)

    // Execute
    result, err := MyFunction(ctx.SdkCtx)

    // Verify
    require.NoError(t, err)
    require.NotNil(t, result)
    require.Equal(t, expectedValue, result)
}
```

### Table-Driven Tests

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "valid",
            wantErr: false,
        },
        {
            name:    "invalid input",
            input:   "",
            wantErr: true,
        },
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

### Using Test Suites

```go
type KeeperTestSuite struct {
    suite.Suite
    keeper Keeper
    ctx    sdk.Context
}

func (suite *KeeperTestSuite) SetupTest() {
    // Setup before each test
    suite.ctx = testutil.SetupTestContext(suite.T()).SdkCtx
    suite.keeper = NewKeeper(...)
}

func (suite *KeeperTestSuite) TestSomething() {
    // Test implementation
    suite.Require().NotNil(suite.keeper)
}

func TestKeeperTestSuite(t *testing.T) {
    suite.Run(t, new(KeeperTestSuite))
}
```

### Mocking

```go
type MockKeeper struct {
    mock.Mock
}

func (m *MockKeeper) Get(ctx sdk.Context, key string) (string, error) {
    args := m.Called(ctx, key)
    return args.String(0), args.Error(1)
}

func TestWithMock(t *testing.T) {
    mockKeeper := new(MockKeeper)
    mockKeeper.On("Get", mock.Anything, "key").Return("value", nil)

    result, err := mockKeeper.Get(ctx, "key")
    require.NoError(t, err)
    require.Equal(t, "value", result)

    mockKeeper.AssertExpectations(t)
}
```

## Coverage Requirements

### Measuring Coverage

```bash
cd chain
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

### Coverage by Module

```bash
cd chain
for dir in x/*/; do
    echo "Coverage for $dir:"
    go test ./$dir/... -coverprofile=coverage.out
    go tool cover -func=coverage.out | grep total
done
```

### Coverage Report

```bash
cd chain
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
# Open coverage.html in browser
```

### Coverage Thresholds

The project enforces the following coverage thresholds:

- **Critical modules** (keeper, handler): 90% minimum
- **Types and utilities**: 80% minimum
- **Test utilities**: 70% minimum
- **Overall project**: 85% minimum

## CI/CD Integration

### GitHub Actions

Tests run automatically on:
- Every push to `master` or `main`
- Every pull request
- Daily scheduled runs (full test suite)

### Workflow Files

- `.github/workflows/test-suite.yml` - Main test suite
- `.github/workflows/regression-tests.yml` - Regression tests
- `.github/workflows/ci.yml` - Basic CI checks

### Local CI Simulation

```bash
# Run the same tests as CI
cd chain
go test ./... -v -race -coverprofile=coverage.out -covermode=atomic
go test ./testing/integration/... -v -tags=integration
go test ./testing/e2e/... -v -tags=e2e
```

## Best Practices

### 1. Test Naming

```go
// Good
func TestKeeperIssueVC_Success(t *testing.T) { }
func TestKeeperIssueVC_InvalidInput(t *testing.T) { }

// Bad
func TestVC(t *testing.T) { }
func Test1(t *testing.T) { }
```

### 2. Test Independence

```go
// Good - each test is independent
func TestA(t *testing.T) {
    ctx := testutil.SetupTestContext(t)
    // ...
}

func TestB(t *testing.T) {
    ctx := testutil.SetupTestContext(t)
    // ...
}

// Bad - tests depend on each other
var globalState string

func TestA(t *testing.T) {
    globalState = "value"
}

func TestB(t *testing.T) {
    require.Equal(t, "value", globalState) // Depends on TestA
}
```

### 3. Clear Assertions

```go
// Good
require.Equal(t, expectedValue, actualValue, "value should match expected")
require.True(t, condition, "condition should be true because X")

// Bad
require.True(t, expectedValue == actualValue)
```

### 4. Setup and Teardown

```go
func TestWithSetup(t *testing.T) {
    // Setup
    ctx := testutil.SetupTestContext(t)
    defer cleanup(ctx) // Teardown

    // Test body
}
```

### 5. Error Testing

```go
// Good - test specific error
func TestError(t *testing.T) {
    err := Function()
    require.Error(t, err)
    require.Contains(t, err.Error(), "expected error message")
}

// Bad - just check error exists
func TestError(t *testing.T) {
    err := Function()
    require.Error(t, err)
}
```

### 6. Test Organization

```
module/
├── keeper/
│   ├── keeper.go
│   ├── keeper_test.go
│   ├── query.go
│   └── query_test.go
├── types/
│   ├── types.go
│   └── types_test.go
└── handler/
    ├── handler.go
    └── handler_test.go
```

### 7. Use Test Helpers

```go
// Create test helpers in testutil package
func SetupTestContext(t *testing.T) *TestContext {
    t.Helper()
    // Setup code
}

func GenerateTestAddress() sdk.AccAddress {
    return sdk.AccAddress([]byte("test_address"))
}
```

### 8. Skip Long Tests in Short Mode

```go
func TestLongRunning(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping long test in short mode")
    }

    // Long test implementation
}
```

## Debugging Tests

### Verbose Output

```bash
go test ./... -v
```

### Run Single Test

```bash
go test ./x/vcregistry/keeper -v -run TestKeeperIssueVC
```

### Print Debug Info

```go
func TestDebug(t *testing.T) {
    result := Function()
    t.Logf("Result: %+v", result) // Prints only if test fails or -v flag
}
```

### Use Delve Debugger

```bash
dlv test ./x/vcregistry/keeper -- -test.run TestKeeperIssueVC
```

## Continuous Improvement

### 1. Review Coverage Reports

Regularly check coverage reports and improve low-coverage areas.

### 2. Update Baselines

Update performance baselines when intentional improvements are made.

### 3. Add Regression Tests

When fixing bugs, add tests to prevent regression.

### 4. Maintain Test Documentation

Keep this guide and test comments up to date.

### 5. Refactor Tests

Refactor tests when they become hard to maintain.

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Cosmos SDK Testing Guide](https://docs.cosmos.network/main/building-modules/testing)
- [Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)

## Support

For questions about testing:
- Check existing tests for examples
- Ask in #testing channel on Discord
- Open an issue on GitHub
- Consult the core team

---

Last Updated: January 13, 2025
