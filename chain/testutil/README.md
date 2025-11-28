# Test Utilities Documentation

This package provides comprehensive test utilities for the Aura blockchain, including keeper helpers, test fixtures, mock implementations, and test suite frameworks.

## Table of Contents

1. [Overview](#overview)
2. [Keeper Helpers](#keeper-helpers)
3. [Test Fixtures](#test-fixtures)
4. [Mock Keepers](#mock-keepers)
5. [Test Suites](#test-suites)
6. [Usage Examples](#usage-examples)

## Overview

The `testutil` package contains all testing infrastructure needed to write comprehensive tests for the Aura blockchain.

### Package Structure

```
testutil/
├── keeper/              # Keeper test helpers
│   ├── setup.go        # Core setup utilities
│   ├── bridge.go       # Bridge keeper helper
│   ├── dex.go          # DEX keeper helper
│   └── ...             # All 20 module helpers
├── testdata/           # Test fixtures and data
│   └── fixtures.go     # Common test data
├── mocks/              # Mock implementations
│   ├── bank_keeper.go
│   ├── account_keeper.go
│   ├── staking_keeper.go
│   ├── gov_keeper.go
│   └── vcregistry_keeper.go
└── suite/              # Test suite frameworks
    └── base.go         # Base test suite
```

## Keeper Helpers

### Setup Test App

The core setup utility creates a complete test application:

```go
import keepertest "github.com/aequitas/aura/chain/testutil/keeper"

// Create basic test app
app, ctx := keepertest.SetupTestApp(t)

// Create test app with validators
app, ctx, validators := keepertest.SetupTestAppWithValidators(t, 4)
```

### Module-Specific Keeper Helpers

Each of the 20 modules has a dedicated keeper helper:

#### Bridge Module

```go
// Simple setup
keeper, ctx := keepertest.BridgeKeeper(t)

// Setup with controllable mocks
keeper, ctx, bankKeeper, accountKeeper, vcKeeper := keepertest.BridgeKeeperWithMocks(t)
```

#### DEX Module

```go
// Simple setup
keeper, ctx := keepertest.DexKeeper(t)

// Setup with controllable mocks
keeper, ctx, bankKeeper, accountKeeper, vcKeeper := keepertest.DexKeeperWithMocks(t)
```

#### Other Modules

All modules follow the same pattern:

```go
// Auth
keeper, ctx := keepertest.AuthKeeper(t)

// Compliance
keeper, ctx := keepertest.ComplianceKeeper(t)

// Confidence Score
keeper, ctx := keepertest.ConfidenceScoreKeeper(t)

// Cryptography
keeper, ctx := keepertest.CryptographyKeeper(t)

// Data Registry
keeper, ctx := keepertest.DataRegistryKeeper(t)

// Economic Security
keeper, ctx := keepertest.EconomicSecurityKeeper(t)

// Governance
keeper, ctx := keepertest.GovernanceKeeper(t)

// Identity Change
keeper, ctx := keepertest.IdentityChangeKeeper(t)

// Incident Response
keeper, ctx := keepertest.IncidentResponseKeeper(t)

// Inclusion Routines
keeper, ctx := keepertest.InclusionRoutinesKeeper(t)

// Monitoring
keeper, ctx := keepertest.MonitoringKeeper(t)

// Network Security
keeper, ctx := keepertest.NetworkSecurityKeeper(t)

// Prevalidation
keeper, ctx := keepertest.PrevalidationKeeper(t)

// Privacy
keeper, ctx := keepertest.PrivacyKeeper(t)

// Social Recovery
keeper, ctx := keepertest.SocialRecoveryKeeper(t)

// Validator Security
keeper, ctx := keepertest.ValidatorSecurityKeeper(t)

// VC Registry
keeper, ctx := keepertest.VCRegistryKeeper(t)

// Wallet Security
keeper, ctx := keepertest.WalletSecurityKeeper(t)
```

### Helper Functions

#### Create Test Accounts

```go
// Create n accounts
accounts := keepertest.CreateTestAccounts(t, 5)

// Create accounts with balances
accounts, balances := keepertest.CreateTestAccountsWithBalances(t, 5, initialBalance)

// Generate single addresses
addr := keepertest.GenTestAddress()
valAddr := keepertest.GenTestValidatorAddress()
pubKey := keepertest.GenTestPubKey()
```

#### Create Test Validators

```go
// Create validators
validators := keepertest.CreateTestValidators(t, 4)

for _, val := range validators {
    fmt.Println("Validator:", val.OperatorAddress)
}
```

#### Create Test Coins

```go
// Single denomination
coins := keepertest.CreateTestCoins(1000, "uaura")

// Multiple denominations
amounts := map[string]int64{
    "uaura": 1000000,
    "uusdt": 500000,
}
coins := keepertest.CreateMultipleTestCoins(amounts)
```

#### Context Helpers

```go
// Create new context
ctx := keepertest.NewTestContext(t)

// Create context at specific height
ctx := keepertest.NewTestContextWithHeight(t, 100)
```

#### Encoding Config

```go
// Create encoding config for tests
encodingConfig := keepertest.MakeEncodingConfig()

cdc := encodingConfig.Codec
amino := encodingConfig.Amino
```

## Test Fixtures

The `testdata` package provides predefined test data and helper functions.

### Predefined Addresses

```go
import "github.com/aequitas/aura/chain/testutil/testdata"

// Use predefined addresses
sender := testdata.TestAddr1
receiver := testdata.TestAddr2

// Validator addresses
validator := testdata.TestValAddr1
```

### Predefined Amounts

```go
// Integer amounts
amount := testdata.TestAmount100    // 100
amount := testdata.TestAmount1000   // 1000
amount := testdata.TestAmount1M     // 1,000,000

// Decimal amounts
dec := testdata.TestDec1            // 1.0
dec := testdata.TestDecHalf         // 0.5
dec := testdata.TestDecQuarter      // 0.25
```

### Predefined Coins

```go
// Single coins
coin := testdata.TestCoinAura1000   // 1000 uaura
coin := testdata.TestCoinUSDT1000   // 1000 uusdt
coin := testdata.TestCoinStake1000  // 1000 stake

// Coin sets
coins := testdata.TestCoinsAura     // 1000 uaura
coins := testdata.TestCoinsMixed    // 1000 uaura + 1000 uusdt
coins := testdata.TestCoinsEmpty    // empty coins
```

### Predefined Timestamps

```go
// Use predefined times
time1 := testdata.TestTime1         // 2024-01-01
time2 := testdata.TestTime2         // 2024-06-01
genesis := testdata.TestTimeGenesis // Genesis time

// Durations
duration := testdata.TestDuration1Day
duration := testdata.TestDuration1Week
```

### Helper Functions

```go
// Generate random addresses
addr := testdata.GenTestAddr()
valAddr := testdata.GenTestValidatorAddr()

// Generate multiple addresses
addrs := testdata.GenTestAddrs(10)
valAddrs := testdata.GenTestValAddrs(4)

// Create custom coins
coins := testdata.MakeTestCoins("uaura", 1000)
coin := testdata.MakeTestCoin("uusdt", 500)

// Multiple denominations
denoms := []string{"uaura", "uusdt", "stake"}
amounts := []int64{1000, 500, 250}
coins := testdata.MakeTestMultiCoins(denoms, amounts)

// Time helpers
future := testdata.TimeFromNow(24 * time.Hour)
past := testdata.TimeAgo(24 * time.Hour)
```

### Constants

```go
// Chain constants
chainID := testdata.TestChainID  // "aura-testnet-1"

// Gas constants
gas := testdata.DefaultGasLimit  // 200000
maxGas := testdata.MaxGasLimit   // 10000000

// Denominations
denom := testdata.TestBondDenom  // "uaura"
denom := testdata.TestUSDTDenom  // "uusdt"

// Validation
minVal := testdata.MinValidatorCount  // 4
maxVal := testdata.MaxValidatorCount  // 100
```

## Mock Keepers

Mock keepers allow you to test modules in isolation without dependencies.

### Mock Bank Keeper

```go
import "github.com/aequitas/aura/chain/testutil/mocks"

bankKeeper := mocks.NewMockBankKeeper()

// Set balances
bankKeeper.SetBalance(addr, coins)

// Simulate errors
bankKeeper.SendCoinsError = errors.New("insufficient funds")

// Check blocked addresses
bankKeeper.BlockAddress(addr)
blocked := bankKeeper.BlockedAddr(addr)
```

Methods:
- `SendCoins(ctx, from, to, amt)`
- `GetBalance(ctx, addr, denom)`
- `GetAllBalances(ctx, addr)`
- `MintCoins(ctx, module, amt)`
- `BurnCoins(ctx, module, amt)`
- `SendCoinsFromModuleToAccount(ctx, module, addr, amt)`
- `SendCoinsFromAccountToModule(ctx, addr, module, amt)`

### Mock Account Keeper

```go
accountKeeper := mocks.NewMockAccountKeeper()

// Create account
acc := accountKeeper.NewAccountWithAddress(ctx, addr)
accountKeeper.SetAccount(ctx, acc)

// Get account
acc := accountKeeper.GetAccount(ctx, addr)

// Module accounts
modAcc := accountKeeper.GetModuleAccount(ctx, "bridge")
```

Methods:
- `GetAccount(ctx, addr)`
- `SetAccount(ctx, acc)`
- `NewAccountWithAddress(ctx, addr)`
- `NewAccount(ctx, acc)`
- `GetModuleAccount(ctx, name)`
- `GetModuleAddress(name)`

### Mock Staking Keeper

```go
stakingKeeper := mocks.NewMockStakingKeeper()

// Set validator
validator := createTestValidator()
stakingKeeper.SetValidator(ctx, validator)

// Set delegation
delegation := createTestDelegation()
stakingKeeper.SetDelegation(ctx, delegation)

// Jail validator
stakingKeeper.Jail(ctx, valAddr)

// Slash validator
stakingKeeper.Slash(ctx, valAddr, fraction)
```

Methods:
- `GetValidator(ctx, addr)`
- `SetValidator(ctx, val)`
- `GetDelegation(ctx, delAddr, valAddr)`
- `SetDelegation(ctx, del)`
- `Jail(ctx, valAddr)`
- `Unjail(ctx, valAddr)`
- `Slash(ctx, valAddr, fraction)`

### Mock Governance Keeper

```go
govKeeper := mocks.NewMockGovKeeper()

// Submit proposal
proposal, err := govKeeper.SubmitProposal(ctx, msgs, metadata, title, summary, proposer)

// Add vote
err = govKeeper.AddVote(ctx, proposalID, voter, options, metadata)

// Add deposit
err = govKeeper.AddDeposit(ctx, proposalID, depositor, amount)

// Test helpers
govKeeper.SetProposalStatus(proposalID, status)
govKeeper.DeleteProposal(proposalID)
```

Methods:
- `GetProposal(ctx, id)`
- `SubmitProposal(ctx, ...)`
- `AddVote(ctx, ...)`
- `GetVote(ctx, id, voter)`
- `AddDeposit(ctx, ...)`
- `GetDeposit(ctx, id, depositor)`

### Mock VC Registry Keeper

```go
vcKeeper := mocks.NewMockVCRegistryKeeper()

// Set credential
vcKeeper.SetCredential(ctx, id, credential)

// Set confidence score
vcKeeper.SetConfidenceScore(ctx, addr, 85)

// Set verification status
vcKeeper.SetVerified(ctx, addr, true)

// Query
cred, found := vcKeeper.GetCredential(ctx, id)
score := vcKeeper.GetConfidenceScore(ctx, addr)
verified := vcKeeper.IsVerified(ctx, addr)
```

Methods:
- `GetCredential(ctx, id)`
- `SetCredential(ctx, id, cred)`
- `GetConfidenceScore(ctx, addr)`
- `SetConfidenceScore(ctx, addr, score)`
- `IsVerified(ctx, addr)`
- `SetVerified(ctx, addr, verified)`

## Test Suites

Test suites provide structured testing frameworks using testify/suite.

### Base Test Suite

```go
import (
    "testing"
    "github.com/stretchr/testify/suite"
    testsuite "github.com/aequitas/aura/chain/testutil/suite"
)

type MyTestSuite struct {
    testsuite.BaseTestSuite
}

func (suite *MyTestSuite) TestSomething() {
    // Use suite helpers
    addr := suite.GenTestAddress()
    coins := suite.GetTestCoins(1000, "uaura")

    // Use suite assertions
    suite.RequireNoError(err)
    suite.RequireEqual(expected, actual)
}

func TestMyTestSuite(t *testing.T) {
    suite.Run(t, new(MyTestSuite))
}
```

### Suite Helper Methods

```go
// Account creation
accounts := suite.CreateTestAccounts(5)
accounts, balances := suite.CreateTestAccountsWithBalances(5, coins)

// Address generation
addr := suite.GenTestAddress()
valAddr := suite.GenTestValidatorAddress()

// Predefined addresses
addrs := suite.UseTestAddresses()
valAddrs := suite.UseTestValidatorAddresses()

// Block manipulation
suite.AdvanceBlockHeight(10)
suite.AdvanceBlockTime(3600) // seconds

// Coin creation
coins := suite.GetTestCoins(1000, "uaura")
coins := suite.GetMultipleTestCoins(map[string]int64{
    "uaura": 1000,
    "uusdt": 500,
})

// Assertions (from testify)
suite.RequireNoError(err)
suite.RequireError(err)
suite.RequireEqual(expected, actual)
suite.RequireTrue(condition)
suite.RequireFalse(condition)
suite.RequireNil(object)
suite.RequireNotNil(object)
```

### Specialized Test Suites

#### Validator Test Suite

```go
type MyValidatorTest struct {
    testsuite.ValidatorTestSuite
}

func (suite *MyValidatorTest) TestValidators() {
    // Validators are pre-loaded
    val := suite.Validators[0]
}
```

#### Module Test Suite

```go
type MyModuleTest struct {
    testsuite.ModuleTestSuite
}

func (suite *MyModuleTest) SetupTest() {
    suite.ModuleName = "bridge"
    suite.BaseTestSuite.SetupTest()
}
```

#### Integration Test Suite

```go
type MyIntegrationTest struct {
    testsuite.IntegrationTestSuite
}

func (suite *MyIntegrationTest) TestIntegration() {
    // Test accounts are pre-created
    sender := suite.TestAccounts[0]
    receiver := suite.TestAccounts[1]
}
```

## Usage Examples

### Example 1: Simple Keeper Test

```go
package keeper_test

import (
    "testing"
    "github.com/stretchr/testify/require"

    keepertest "github.com/aequitas/aura/chain/testutil/keeper"
    "github.com/aequitas/aura/chain/testutil/testdata"
)

func TestBridgeTransfer(t *testing.T) {
    keeper, ctx := keepertest.BridgeKeeper(t)

    sender := testdata.TestAddr1
    amount := testdata.TestCoinAura1000

    err := keeper.InitiateTransfer(ctx, sender, "0x123", amount, "ethereum")
    require.NoError(t, err)
}
```

### Example 2: Test with Mocks

```go
func TestDEXSwap(t *testing.T) {
    keeper, ctx, bankKeeper, accountKeeper, vcKeeper := keepertest.DexKeeperWithMocks(t)

    // Setup mock behavior
    trader := testdata.TestAddr1
    bankKeeper.SetBalance(trader, testdata.TestCoinsMixed)
    vcKeeper.SetConfidenceScore(ctx, trader, 80)

    // Test swap
    err := keeper.Swap(ctx, trader, "uaura", "uusdt", 100)
    require.NoError(t, err)

    // Verify mock interactions
    balance := bankKeeper.GetBalance(ctx, trader, "uusdt")
    require.True(t, balance.Amount.GT(sdk.ZeroInt()))
}
```

### Example 3: Test Suite

```go
type BridgeKeeperTestSuite struct {
    testsuite.BaseTestSuite
    keeper *keeper.Keeper
}

func (suite *BridgeKeeperTestSuite) SetupTest() {
    suite.BaseTestSuite.SetupTest()
    suite.keeper, suite.Ctx = keepertest.BridgeKeeper(suite.T())
}

func (suite *BridgeKeeperTestSuite) TestTransfer_Success() {
    sender := suite.UseTestAddresses()[0]
    amount := suite.GetTestCoins(1000, "uaura")

    err := suite.keeper.Transfer(suite.Ctx, sender, "0x123", amount)
    suite.RequireNoError(err)
}

func (suite *BridgeKeeperTestSuite) TestTransfer_InsufficientBalance() {
    sender := suite.UseTestAddresses()[0]
    amount := suite.GetTestCoins(1000000000, "uaura")

    err := suite.keeper.Transfer(suite.Ctx, sender, "0x123", amount)
    suite.RequireError(err)
}

func TestBridgeKeeperTestSuite(t *testing.T) {
    suite.Run(t, new(BridgeKeeperTestSuite))
}
```

### Example 4: Table-Driven Test

```go
func TestValidateAddress(t *testing.T) {
    tests := []struct {
        name    string
        address string
        wantErr bool
    }{
        {
            name:    "valid address",
            address: testdata.TestAddr1.String(),
            wantErr: false,
        },
        {
            name:    "empty address",
            address: "",
            wantErr: true,
        },
        {
            name:    "invalid format",
            address: "invalid",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateAddress(tt.address)
            if tt.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

## Best Practices

1. **Use Helpers**: Leverage keeper helpers instead of manual setup
2. **Use Fixtures**: Use predefined test data from `testdata`
3. **Use Mocks**: Mock dependencies for isolated testing
4. **Use Suites**: Use test suites for organized, related tests
5. **Table Tests**: Use table-driven tests for multiple scenarios
6. **Cleanup**: Always clean up resources in defer
7. **Descriptive Names**: Use clear, descriptive test names
8. **Test Errors**: Test both success and error cases

## See Also

- [TESTING.md](../TESTING.md) - Complete testing guide
- [tests/integration/README.md](../tests/integration/README.md) - Integration tests
- [tests/e2e/README.md](../tests/e2e/README.md) - E2E tests
