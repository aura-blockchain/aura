# Aura Blockchain Testing Guide

Comprehensive guide to testing the Aura blockchain, covering all test types, frameworks, and best practices.

## Table of Contents

1. [Overview](#overview)
2. [Test Infrastructure](#test-infrastructure)
3. [Unit Testing](#unit-testing)
4. [Integration Testing](#integration-testing)
5. [End-to-End Testing](#end-to-end-testing)
6. [Test Utilities](#test-utilities)
7. [Best Practices](#best-practices)
8. [Running Tests](#running-tests)
9. [Coverage](#coverage)
10. [CI/CD Integration](#cicd-integration)

## Overview

The Aura blockchain has a comprehensive testing infrastructure supporting three levels of testing:

- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test module interactions and cross-module flows
- **E2E Tests**: Test complete system behavior with multiple nodes

### Test Philosophy

1. **Comprehensive Coverage**: All modules have test coverage
2. **Fast Feedback**: Unit tests run quickly for rapid development
3. **Real Scenarios**: Integration and E2E tests simulate real usage
4. **Maintainable**: Tests are easy to understand and maintain
5. **Reliable**: Tests are deterministic and reproducible

## Test Infrastructure

### Directory Structure

```
chain/
├── testutil/                  # Test utilities and helpers
│   ├── keeper/               # Keeper test helpers
│   │   ├── setup.go         # Main setup utilities
│   │   ├── bridge.go        # Bridge keeper helper
│   │   ├── dex.go           # DEX keeper helper
│   │   └── ... (all 20 modules)
│   ├── testdata/            # Test fixtures and data
│   │   └── fixtures.go      # Common test data
│   ├── mocks/               # Mock implementations
│   │   ├── bank_keeper.go   # Mock bank keeper
│   │   ├── account_keeper.go
│   │   ├── staking_keeper.go
│   │   ├── gov_keeper.go
│   │   └── vcregistry_keeper.go
│   └── suite/               # Test suite frameworks
│       └── base.go          # Base test suite
├── tests/
│   ├── integration/         # Integration tests
│   │   ├── setup.go
│   │   └── README.md
│   └── e2e/                 # End-to-end tests
│       ├── chain.go
│       └── README.md
└── x/                       # Module tests
    └── {module}/
        └── keeper/
            └── *_test.go    # Module-specific tests
```

## Unit Testing

Unit tests verify individual keeper methods, msg handlers, and business logic in isolation.

### Creating a Unit Test

```go
package keeper_test

import (
    "testing"
    "github.com/stretchr/testify/require"

    keepertest "github.com/aequitas/aura/chain/testutil/keeper"
    "github.com/aequitas/aura/chain/x/bridge/types"
)

func TestBridgeTransfer(t *testing.T) {
    keeper, ctx := keepertest.BridgeKeeper(t)

    // Test bridge transfer logic
    msg := types.MsgBridgeTransfer{
        Sender:      testAddr,
        Receiver:    "0x123...",
        Amount:      sdk.NewCoins(sdk.NewCoin("uaura", sdk.NewInt(1000))),
        TargetChain: "ethereum",
    }

    err := keeper.BridgeTransfer(ctx, msg)
    require.NoError(t, err)

    // Verify state changes
    transfer, found := keeper.GetTransfer(ctx, "transfer-id")
    require.True(t, found)
    require.Equal(t, msg.Amount, transfer.Amount)
}
```

### Using Test Suites

```go
package keeper_test

import (
    "testing"
    "github.com/stretchr/testify/suite"

    keepertest "github.com/aequitas/aura/chain/testutil/keeper"
    testsuite "github.com/aequitas/aura/chain/testutil/suite"
)

type BridgeKeeperTestSuite struct {
    testsuite.BaseTestSuite
    keeper *keeper.Keeper
}

func (suite *BridgeKeeperTestSuite) SetupTest() {
    suite.BaseTestSuite.SetupTest()
    suite.keeper, suite.Ctx = keepertest.BridgeKeeper(suite.T())
}

func (suite *BridgeKeeperTestSuite) TestBridgeTransfer() {
    // Test implementation
}

func TestBridgeKeeperSuite(t *testing.T) {
    suite.Run(t, new(BridgeKeeperTestSuite))
}
```

### Module-Specific Keeper Helpers

Each module has a dedicated keeper helper in `testutil/keeper/`:

```go
// Setup bridge keeper with mocks
keeper, ctx := keepertest.BridgeKeeper(t)

// Setup with controllable mocks
keeper, ctx, bankKeeper, accountKeeper, vcKeeper := keepertest.BridgeKeeperWithMocks(t)
```

Available helpers:
- `AuthKeeper(t)`
- `BridgeKeeper(t)` / `BridgeKeeperWithMocks(t)`
- `ComplianceKeeper(t)`
- `ConfidenceScoreKeeper(t)`
- `CryptographyKeeper(t)`
- `DataRegistryKeeper(t)`
- `DexKeeper(t)` / `DexKeeperWithMocks(t)`
- `EconomicSecurityKeeper(t)`
- `GovernanceKeeper(t)`
- `IdentityChangeKeeper(t)`
- `IncidentResponseKeeper(t)`
- `InclusionRoutinesKeeper(t)`
- `MonitoringKeeper(t)`
- `NetworkSecurityKeeper(t)`
- `PrevalidationKeeper(t)`
- `PrivacyKeeper(t)`
- `SocialRecoveryKeeper(t)`
- `ValidatorSecurityKeeper(t)`
- `VCRegistryKeeper(t)`
- `WalletSecurityKeeper(t)`

## Integration Testing

Integration tests verify that multiple modules work correctly together.

### Creating an Integration Test

```go
package integration

import (
    "testing"
    "github.com/stretchr/testify/require"

    "github.com/aequitas/aura/chain/tests/integration"
)

func TestDEXAndBankIntegration(t *testing.T) {
    suite := integration.SetupIntegrationTest(t)
    defer suite.Cleanup()

    // Get test accounts
    creator := suite.GetAccount(0)
    trader := suite.GetAccount(1)

    // Test cross-module functionality
    err := suite.CreatePoolAndSwap(
        creator,
        "uaura", "uusdt",
        1000000, 1000000,
        100000,
    )
    require.NoError(t, err)

    // Verify state across modules
    // ... check bank balances, pool state, etc.
}
```

### Integration Test with Validators

```go
func TestGovernanceFlow(t *testing.T) {
    suite := integration.SetupIntegrationTestWithValidators(t, 4)
    defer suite.Cleanup()

    proposer := suite.GetAccount(0)
    voters := suite.Accounts[:3]

    // Test proposal and voting
    err := suite.SubmitProposalAndVote(
        proposer,
        voters,
        "Parameter Change",
        "Update DEX fee parameter",
    )
    require.NoError(t, err)
}
```

### Cross-Module Test Scenarios

- **DEX + Bank**: Pool creation, swaps, liquidity provision
- **Bridge + Bank + VCRegistry**: Cross-chain transfers with identity
- **Governance + All**: Parameter updates across all modules
- **Security Modules**: Complete incident response flows
- **Identity Flow**: VC issuance → confidence score → verification

See [tests/integration/README.md](tests/integration/README.md) for more details.

## End-to-End Testing

E2E tests verify complete system behavior with realistic multi-node scenarios.

### Single Chain E2E Test

```go
package e2e

import (
    "testing"
    "github.com/stretchr/testify/require"

    "github.com/aequitas/aura/chain/tests/e2e"
)

func TestChainBootstrap(t *testing.T) {
    // Create chain with 4 validators
    chain := e2e.NewChain(t, "aura-testnet-1", 4)

    err := chain.Start()
    require.NoError(t, err)
    defer chain.Stop()

    // Wait for blocks to be produced
    err = chain.WaitForHeight(10)
    require.NoError(t, err)

    height := chain.GetHeight()
    require.Equal(t, int64(10), height)
}
```

### Multi-Chain E2E Test

```go
func TestCrossChainBridge(t *testing.T) {
    // Create 2 chains
    suite := e2e.NewMultiChainTest(t, 2)

    err := suite.StartAll()
    require.NoError(t, err)
    defer suite.StopAll()

    chainA := suite.GetChain(0)
    chainB := suite.GetChain(1)

    sender := chainA.CreateAccount(initialBalance)
    receiver := chainB.CreateAccount(sdk.NewCoins())

    // Simulate cross-chain transfer
    err = suite.SimulateIBCTransfer(0, 1, sender, receiver, amount)
    require.NoError(t, err)

    err = suite.WaitForRelayer()
    require.NoError(t, err)

    // Verify transfer completed
    // ...
}
```

See [tests/e2e/README.md](tests/e2e/README.md) for more details.

## Test Utilities

### SetupTestApp

Creates a complete test application:

```go
app, ctx := keeper.SetupTestApp(t)
```

### Test Accounts

```go
// Create accounts
accounts := keeper.CreateTestAccounts(t, 5)

// Create with balances
accounts, balances := keeper.CreateTestAccountsWithBalances(t, 5, initialBalance)

// Generate random address
addr := keeper.GenTestAddress()
valAddr := keeper.GenTestValidatorAddress()
```

### Test Fixtures

```go
import "github.com/aequitas/aura/chain/testutil/testdata"

// Use predefined addresses
addr := testdata.TestAddr1

// Use predefined amounts
amount := testdata.TestAmount1000

// Use predefined coins
coins := testdata.TestCoinsMixed  // 1000 uaura + 1000 uusdt

// Generate test data
addrs := testdata.GenTestAddrs(10)
```

### Mock Keepers

```go
import "github.com/aequitas/aura/chain/testutil/mocks"

// Create mock keepers
bankKeeper := mocks.NewMockBankKeeper()
accountKeeper := mocks.NewMockAccountKeeper()
stakingKeeper := mocks.NewMockStakingKeeper()
govKeeper := mocks.NewMockGovKeeper()
vcKeeper := mocks.NewMockVCRegistryKeeper()

// Set up mock behavior
bankKeeper.SetBalance(addr, coins)
accountKeeper.SetAccount(ctx, account)
vcKeeper.SetConfidenceScore(ctx, addr, 80)
```

## Best Practices

### 1. Test Naming

```go
// Good: Descriptive test names
func TestBridgeTransfer_ValidTransfer_Success(t *testing.T)
func TestDEXSwap_InsufficientLiquidity_ReturnsError(t *testing.T)

// Bad: Vague test names
func TestTransfer(t *testing.T)
func TestError(t *testing.T)
```

### 2. Test Organization

```go
func TestFeature(t *testing.T) {
    // Arrange: Setup test data
    keeper, ctx := keepertest.BridgeKeeper(t)
    sender := testdata.TestAddr1

    // Act: Execute the functionality
    err := keeper.Transfer(ctx, sender, amount)

    // Assert: Verify results
    require.NoError(t, err)
}
```

### 3. Table-Driven Tests

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid input", "valid", false},
        {"empty input", "", true},
        {"invalid format", "###", true},
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

### 4. Cleanup

```go
func TestFeature(t *testing.T) {
    suite := integration.SetupIntegrationTest(t)
    defer suite.Cleanup()  // Always cleanup

    // Test code...
}
```

### 5. Error Testing

```go
// Test both success and failure cases
func TestTransfer_Success(t *testing.T) { ... }
func TestTransfer_InsufficientBalance_Error(t *testing.T) { ... }
func TestTransfer_InvalidAddress_Error(t *testing.T) { ... }
```

## Running Tests

### Run All Tests

```bash
# All tests
go test ./...

# Specific module
go test ./x/bridge/...

# Specific package
go test ./x/bridge/keeper
```

### Run by Type

```bash
# Unit tests only (fast)
go test -short ./...

# Integration tests
go test ./tests/integration/...

# E2E tests
go test ./tests/e2e/...
```

### Verbose Output

```bash
# See test names and output
go test -v ./x/bridge/keeper

# See all logs
go test -v -count=1 ./...
```

### Run Specific Tests

```bash
# Run specific test
go test ./x/bridge/keeper -run TestBridgeTransfer

# Run tests matching pattern
go test ./... -run Bridge

# Run test suite
go test ./x/bridge/keeper -run TestBridgeKeeperSuite
```

### Parallel Testing

```bash
# Run tests in parallel
go test -parallel 4 ./...

# Disable parallel (for debugging)
go test -parallel 1 ./...
```

## Coverage

### Generate Coverage

```bash
# Generate coverage report
go test -cover ./...

# Detailed coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Coverage by package
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### Coverage Targets

- **Unit Tests**: Target 80%+ coverage
- **Integration Tests**: Cover cross-module flows
- **E2E Tests**: Cover critical user journeys

### View Coverage

```bash
# Open HTML report
go tool cover -html=coverage.out

# Terminal view
go tool cover -func=coverage.out | grep total
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Unit Tests
        run: go test -short -race -coverprofile=coverage.out ./...

      - name: Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out

  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4

      - name: Integration Tests
        run: go test ./tests/integration/...

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4

      - name: E2E Tests
        run: go test -timeout 30m ./tests/e2e/...
```

### Test Strategy

1. **PR Checks**: Run unit tests on every PR
2. **Merge Checks**: Run integration tests before merge
3. **Nightly**: Run full E2E test suite
4. **Release**: Full test suite + performance tests

## Troubleshooting

### Common Issues

**Tests are flaky**
- Ensure deterministic test data
- Avoid time-dependent logic or use mocks
- Check for race conditions

**Tests are slow**
- Use `-short` flag to skip slow tests
- Run tests in parallel
- Mock external dependencies

**Coverage is low**
- Write tests for error cases
- Test edge cases
- Cover all exported functions

### Debug Tests

```bash
# Run with verbose logging
go test -v ./x/bridge/keeper

# Run single test with logs
go test -v -run TestBridgeTransfer ./x/bridge/keeper

# Check for race conditions
go test -race ./...

# Debug with delve
dlv test ./x/bridge/keeper -- -test.run TestBridgeTransfer
```

## Additional Resources

- [testutil/README.md](testutil/README.md) - Test utilities documentation
- [tests/integration/README.md](tests/integration/README.md) - Integration testing guide
- [tests/e2e/README.md](tests/e2e/README.md) - E2E testing guide
- [Cosmos SDK Testing](https://docs.cosmos.network/main/building-modules/testing) - Upstream docs

## Contributing

When adding new functionality:

1. Write unit tests for keeper methods
2. Write integration tests for cross-module flows
3. Add E2E tests for critical user journeys
4. Aim for 80%+ test coverage
5. Document test scenarios in comments

For questions or issues, see the main [README.md](README.md) or open an issue.
