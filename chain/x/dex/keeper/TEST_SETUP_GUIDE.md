# DEX Keeper Test Setup Guide

## Quick Reference for Writing DEX Tests

### 1. Basic Test Setup

```go
func TestMyFeature(t *testing.T) {
    suite := SetupKeeperTestSuite(t)
    ctx := suite.Ctx
    keeper := suite.DexKeeper
    mockBank := suite.BankKeeper

    // Your test code here
}
```

### 2. Funding Test Accounts

**CRITICAL:** Always fund accounts before operations that require balance.

```go
userAddr := keepertest.GenTestAddr()
user := userAddr.String()

// Fund the account
mockBank.SetBalance(userAddr, "uaura", math.NewInt(1_000000))
mockBank.SetBalance(userAddr, "usdt", math.NewInt(1_000000))

// Now you can create pools, trade, etc.
pool, lpTokens, err := keeper.CreatePool(ctx, user, "uaura", "usdt",
    sdk.NewCoin("uaura", math.NewInt(500000)),
    sdk.NewCoin("usdt", math.NewInt(500000)))
```

### 3. Time Management for Security Features

#### Pool Creation Cooldown (1 hour default)

```go
// First pool - no cooldown
keeper.RecordPoolCreation(ctx, creator, "pool1", "uaura", "usdt", amountA, amountB)

// Advance time for second pool
ctx = ctx.WithBlockTime(ctx.BlockTime().Add(1 * time.Hour))
keeper.RecordPoolCreation(ctx, creator, "pool2", "uaura", "usdc", amountA, amountB)
```

#### Front-Running Protection (2 blocks default)

```go
// First trade
keeper.RecordTradeBlock(ctx, userAddr, poolID)

// Advance blocks for next trade
ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 2)
// Now next trade is allowed
```

### 4. Security Parameters

Current defaults (hardcoded in `GetSecurityParams()`):

```go
MinBlockDelay:            2          // blocks between trades
MaxTradeSizePercent:      "0.20"     // 20% of pool
MaxPriceImpactPercent:    "10.00"    // 10%
LiquidityLockupSeconds:   86400      // 24 hours
PoolCreationCooldown:     3600       // 1 hour (seconds)
MaxPoolsPerCreator:       10         // max pools per address
TwapWindowBlocks:         100        // TWAP calculation window
MinPoolCreationLiquidity: "1000000000" // minimum liquidity
MinLiquidityBlocks:       5          // blocks between add/remove liquidity
MinTradeAmount:           "1000"     // minimum trade size
```

**Note:** Currently cannot override params in tests. Tests must work with defaults.

### 5. Common Test Patterns

#### Testing Pool Creation

```go
func TestPoolCreation(t *testing.T) {
    suite := SetupKeeperTestSuite(t)
    ctx := suite.Ctx
    keeper := suite.DexKeeper
    mockBank := suite.BankKeeper

    userAddr := suite.TestAccs[0]
    user := userAddr.String()

    // Fund account
    mockBank.SetBalance(userAddr, "uaura", math.NewInt(1_000000))
    mockBank.SetBalance(userAddr, "usdt", math.NewInt(1_000000))

    // Create pool
    pool, lpTokens, err := keeper.CreatePool(ctx, user, "uaura", "usdt",
        sdk.NewCoin("uaura", math.NewInt(500000)),
        sdk.NewCoin("usdt", math.NewInt(500000)))

    require.NoError(t, err)
    require.NotNil(t, pool)
    require.True(t, lpTokens.IsPositive())
}
```

#### Testing Swaps

```go
func TestSwap(t *testing.T) {
    suite := SetupKeeperTestSuite(t)
    ctx := suite.Ctx
    keeper := suite.DexKeeper
    mockBank := suite.BankKeeper

    // Create pool first (with funded creator)
    creator := suite.TestAccs[0]
    mockBank.SetBalance(creator, "uaura", math.NewInt(10_000000))
    mockBank.SetBalance(creator, "usdt", math.NewInt(10_000000))

    pool, _, err := keeper.CreatePool(ctx, creator.String(), "uaura", "usdt",
        sdk.NewCoin("uaura", math.NewInt(1_000000)),
        sdk.NewCoin("usdt", math.NewInt(1_000000)))
    require.NoError(t, err)

    // Fund trader
    trader := suite.TestAccs[1]
    mockBank.SetBalance(trader, "uaura", math.NewInt(100000))

    // Perform swap
    amountOut, err := keeper.SwapExactAmountIn(ctx, trader.String(), pool.PoolId,
        sdk.NewCoin("uaura", math.NewInt(10000)),
        "usdt", math.NewInt(9000))

    require.NoError(t, err)
    require.True(t, amountOut.IsPositive())
}
```

#### Testing Security Limits

```go
func TestPoolCreationLimit(t *testing.T) {
    suite := SetupKeeperTestSuite(t)
    ctx := suite.Ctx
    keeper := suite.DexKeeper

    creator := "aura1creator1"

    // Create up to limit (default: 10)
    for i := 1; i <= 10; i++ {
        // Advance time for cooldown
        if i > 1 {
            ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Hour + time.Second))
        }

        err := keeper.CheckPoolCreationLimit(ctx, creator)
        require.NoError(t, err)

        keeper.RecordPoolCreation(ctx, creator, fmt.Sprintf("pool%d", i),
            "uaura", "usdt", amount, amount)
    }

    // 11th pool should fail
    ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Hour + time.Second))
    err := keeper.CheckPoolCreationLimit(ctx, creator)
    require.Error(t, err)
    require.ErrorIs(t, err, types.ErrPoolCreationLimitExceeded)
}
```

### 6. Mock Bank Keeper Behavior

The `MockBankKeeper` has specific behavior:

1. **Balance Tracking:** Only tracks addresses explicitly added via `SetBalance` or `SendCoinsFromModuleToAccount`
2. **Balance Checks:** `SendCoinsFromAccountToModule` checks and deducts balance for tracked addresses
3. **Insufficient Balance:** Returns error for unfunded addresses trying to send non-zero amounts
4. **Module Operations:** `SendCoinsFromModuleToAccount` and `MintCoins` don't require pre-funding

```go
// This will FAIL (insufficient balance)
err := keeper.CreatePool(ctx, unfundedUser, "uaura", "usdt", ...)

// This will SUCCEED
mockBank.SetBalance(userAddr, "uaura", math.NewInt(1000000))
err := keeper.CreatePool(ctx, user, "uaura", "usdt", ...)
```

### 7. Common Pitfalls

❌ **Don't:** Forget to fund accounts before operations
```go
// BAD - will fail with insufficient balance
pool, _, err := keeper.CreatePool(ctx, user, "uaura", "usdt", coins...)
```

✅ **Do:** Always fund first
```go
// GOOD
mockBank.SetBalance(userAddr, "uaura", math.NewInt(1_000000))
pool, _, err := keeper.CreatePool(ctx, user, "uaura", "usdt", coins...)
```

❌ **Don't:** Forget time advancement between security-restricted operations
```go
// BAD - will fail cooldown
keeper.RecordPoolCreation(ctx, creator, "pool1", ...)
keeper.RecordPoolCreation(ctx, creator, "pool2", ...) // FAILS
```

✅ **Do:** Advance time appropriately
```go
// GOOD
keeper.RecordPoolCreation(ctx, creator, "pool1", ...)
ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Hour))
keeper.RecordPoolCreation(ctx, creator, "pool2", ...) // SUCCESS
```

❌ **Don't:** Assume you can override security params
```go
// BAD - params can't be set in tests yet
params := types.DefaultSecurityParams()
params.MaxPoolsPerCreator = 3  // This won't affect keeper behavior
```

✅ **Do:** Work with default params
```go
// GOOD - test with default limit of 10
for i := 1; i <= 10; i++ {
    // Create pools within default limit
}
```

### 8. Test File Organization

```
x/dex/keeper/
├── keeper_test_suite.go           # Base test suite setup
├── test_helpers_test.go           # Helper functions (SetupKeeperTestSuite)
├── keeper_comprehensive_test.go   # setupTestKeeper, MockBankKeeper
├── pool_creation_record_test.go   # Pool creation tracking tests
├── security_test.go                # Security feature tests
├── query_server_test.go           # Query handler tests
└── ...
```

### 9. Running Tests

```bash
# All DEX tests
go test ./x/dex/keeper/...

# Specific test
go test -v ./x/dex/keeper/... -run TestPoolCreationLimit_Enforcement

# With coverage
go test -cover ./x/dex/keeper/...

# Verbose with count (disable cache)
go test -v -count=1 ./x/dex/keeper/...
```

### 10. Debug Tips

**Test failing with "insufficient balance":**
- Check if account is funded with `mockBank.SetBalance`
- Verify correct address type (sdk.AccAddress vs string)
- Ensure denomination matches ("uaura" not "aura")

**Test failing with "cooldown" or "limit exceeded":**
- Check if time is advanced between operations
- Verify working with default params (can't override)
- Ensure context is updated: `ctx = ctx.WithBlockTime(...)`

**Test failing with "pool not found":**
- Create pool before querying/trading
- Verify pool ID matches expected format
- Check if pool creator had sufficient balance

---

## Need More Help?

See:
- `DEX_TEST_FIXES_SUMMARY.md` for recent fixes
- `MOCK_BANK_KEEPER_EXPLANATION.md` for mock bank details
- `security.go` for security parameter details
- `types/security_types.go` for parameter defaults
