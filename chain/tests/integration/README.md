# Integration Tests

This directory contains integration tests for the Aura blockchain.

## Overview

Integration tests verify that multiple modules work correctly together. These tests simulate real-world scenarios where different modules interact.

## Structure

```
integration/
├── setup.go           - Integration test suite setup
├── bridge_test.go     - Bridge module integration tests
├── dex_test.go        - DEX module integration tests
├── governance_test.go - Governance integration tests
└── identity_test.go   - Identity/VC integration tests
```

## Usage

### Basic Integration Test

```go
package integration

import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestDEXAndBankIntegration(t *testing.T) {
    suite := SetupIntegrationTest(t)
    defer suite.Cleanup()

    // Get test accounts
    creator := suite.GetAccount(0)
    trader := suite.GetAccount(1)

    // Test DEX pool creation with bank interactions
    err := suite.CreatePoolAndSwap(
        creator,
        "uaura", "uusdt",
        1000000, 1000000,
        100000,
    )
    require.NoError(t, err)
}
```

### Integration Test with Validators

```go
func TestGovernanceWithValidators(t *testing.T) {
    suite := SetupIntegrationTestWithValidators(t, 4)
    defer suite.Cleanup()

    proposer := suite.GetAccount(0)

    // Test proposal submission and voting
    err := suite.SubmitProposalAndVote(
        proposer,
        suite.Accounts[:3],
        "Test Proposal",
        "Test Description",
    )
    require.NoError(t, err)
}
```

## Test Scenarios

### Cross-Module Scenarios

1. **DEX + Bank**: Pool creation, swaps, liquidity
2. **Bridge + Bank**: Cross-chain transfers
3. **Governance + All Modules**: Parameter changes
4. **VCRegistry + ConfidenceScore**: Identity verification
5. **Security Modules**: Incident response flows

### Time-Based Testing

Use `AdvanceTime` and `AdvanceBlockHeight` for testing time-dependent logic:

```go
suite.AdvanceTime(24 * time.Hour)
suite.AdvanceBlockHeight(1000)
```

## Best Practices

1. **Cleanup**: Always call `suite.Cleanup()` in defer
2. **Isolation**: Each test should be independent
3. **Real Flows**: Test actual user workflows
4. **Error Cases**: Test failure scenarios
5. **State Verification**: Verify state changes across modules

## Running Tests

```bash
# Run all integration tests
go test ./tests/integration/...

# Run specific integration test
go test ./tests/integration/ -run TestDEXAndBank

# Run with verbose output
go test -v ./tests/integration/...

# Run with coverage
go test -cover ./tests/integration/...
```

## Adding New Integration Tests

1. Create a new test file: `{module}_test.go`
2. Import the integration package
3. Use `SetupIntegrationTest` or `SetupIntegrationTestWithValidators`
4. Write test scenarios
5. Add cleanup

## Troubleshooting

- **Test Timeouts**: Increase timeout with `-timeout` flag
- **State Issues**: Ensure proper cleanup between tests
- **Dependencies**: Verify all keeper mocks are properly set up
