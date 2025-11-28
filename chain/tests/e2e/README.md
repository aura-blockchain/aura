# End-to-End (E2E) Tests

This directory contains end-to-end tests for the Aura blockchain.

## Overview

E2E tests verify complete system behavior, including multi-node scenarios, cross-chain interactions, and full transaction lifecycles.

## Structure

```
e2e/
├── chain.go           - Chain and multi-chain test infrastructure
├── README.md          - This file
├── bridge_e2e_test.go - Bridge E2E tests
├── ibc_e2e_test.go    - IBC/cross-chain E2E tests
└── network_e2e_test.go - Network behavior E2E tests
```

## Usage

### Single Chain E2E Test

```go
package e2e

import (
    "testing"
    "github.com/stretchr/testify/require"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestChainStartup(t *testing.T) {
    // Create chain with 4 validators
    chain := NewChain(t, "test-chain-1", 4)

    // Start chain
    err := chain.Start()
    require.NoError(t, err)
    defer chain.Stop()

    // Test functionality
    account := chain.CreateAccount(sdk.NewCoins(
        sdk.NewCoin("uaura", sdk.NewInt(1000000)),
    ))

    // Wait for blocks
    err = chain.WaitForHeight(10)
    require.NoError(t, err)
}
```

### Multi-Chain E2E Test

```go
func TestCrossChainTransfer(t *testing.T) {
    // Create 2 chains
    suite := NewMultiChainTest(t, 2)

    // Start all chains
    err := suite.StartAll()
    require.NoError(t, err)
    defer suite.StopAll()

    // Get chains
    chainA := suite.GetChain(0)
    chainB := suite.GetChain(1)

    // Create accounts
    senderA := chainA.CreateAccount(testCoins)
    receiverB := chainB.CreateAccount(sdk.NewCoins())

    // Simulate IBC transfer
    err = suite.SimulateIBCTransfer(
        0, 1,
        senderA, receiverB,
        sdk.NewCoins(sdk.NewCoin("uaura", sdk.NewInt(100))),
    )
    require.NoError(t, err)

    // Wait for relayer
    err = suite.WaitForRelayer()
    require.NoError(t, err)
}
```

## Test Scenarios

### Single Chain Scenarios

1. **Chain Startup**: Validator initialization, genesis setup
2. **Block Production**: Continuous block creation
3. **Transaction Flow**: Submit, execute, commit
4. **State Sync**: State synchronization between nodes
5. **Upgrades**: Chain upgrade scenarios

### Multi-Chain Scenarios

1. **IBC Transfers**: Cross-chain token transfers
2. **Bridge Operations**: Bridging to external chains
3. **Relayer Behavior**: Packet relay and acknowledgments
4. **Cross-Chain Governance**: Multi-chain proposals
5. **Network Partitions**: Handling network splits

## Chain Configuration

### Validator Setup

```go
chain := NewChain(t, "aura-test", 4)

// Access validators
val := chain.GetValidator(0)
t.Log("Validator:", val.Moniker, "Power:", val.Power)
```

### Adding Full Nodes

```go
// Add non-validator nodes
node1 := chain.AddFullNode("full-node-1")
node2 := chain.AddFullNode("full-node-2")
```

### Genesis Configuration

```go
chain.Genesis.InitialSupply = map[string]int64{
    "uaura": 1000000000000,
    "uusdt": 1000000000000,
}
```

## Time and Block Progression

### Wait for Specific Height

```go
// Wait for block 100
err := chain.WaitForHeight(100)
require.NoError(t, err)
```

### Wait for Next Block

```go
// Wait for next block (useful for time-dependent logic)
err := chain.WaitForNextBlock()
require.NoError(t, err)
```

### Get Current State

```go
height := chain.GetHeight()
blockTime := chain.GetTime()
```

## Transaction Testing

### Send Transaction

```go
from := chain.CreateAccount(initialBalance)

// Create message
msg := &types.MsgSend{
    FromAddress: from.String(),
    ToAddress:   to.String(),
    Amount:      amount,
}

// Send transaction
err := chain.SendTransaction(from, msg)
require.NoError(t, err)
```

### Query State

```go
// Query account balance
data, err := chain.Query("/bank/balances/"+addr.String(), nil)
require.NoError(t, err)
```

## Running E2E Tests

```bash
# Run all E2E tests
go test ./tests/e2e/...

# Run specific E2E test
go test ./tests/e2e/ -run TestChainStartup

# Run with verbose output
go test -v ./tests/e2e/...

# Run with longer timeout (E2E tests may take longer)
go test -timeout 30m ./tests/e2e/...
```

## Best Practices

1. **Cleanup**: Always defer `chain.Stop()` or `suite.StopAll()`
2. **Timeouts**: Use appropriate timeouts for block production
3. **Isolation**: Each test should start fresh chains
4. **Logging**: Use `t.Log()` to track test progress
5. **Error Handling**: Check all errors from chain operations
6. **Real Scenarios**: Test actual user workflows end-to-end

## Performance Considerations

- E2E tests are slower than unit/integration tests
- Run E2E tests separately from faster tests
- Use `-short` flag to skip E2E tests in quick runs:

```go
func TestLongRunningE2E(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping E2E test in short mode")
    }
    // ... test code
}
```

## Troubleshooting

- **Chain Won't Start**: Check validator configuration
- **Timeouts**: Increase test timeout or block time
- **State Issues**: Ensure proper cleanup between tests
- **Multi-Chain Issues**: Verify relayer is properly simulated
