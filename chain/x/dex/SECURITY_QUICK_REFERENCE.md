# DEX Security Features - Quick Reference

## Files and Line Numbers

| Feature | File | Lines | Key Functions |
|---------|------|-------|--------------|
| **Front-Running Protection** | `keeper/security.go` | 19-62 | `CheckFrontRunningProtection()`, `RecordTradeBlock()` |
| **TWAP Oracle** | `keeper/security.go` | 64-171 | `RecordTWAPObservation()`, `GetTWAPPrice()` |
| **Flash Loan Protection** | `keeper/security.go` | 173-220 | `CheckFlashLoanProtection()`, `RecordLiquidityBlock()` |
| **MEV Mitigation** | `keeper/security.go` | 222-252 | `CheckMEVProtection()`, `GetSwapsInCurrentBlock()` |
| **Slippage Limits** | `keeper/security.go` | 254-271 | `CheckPoolSlippageLimit()` |
| **Trade Size Caps** | `keeper/security.go` | 273-303 | `CheckMaxTradeSize()` |
| **Price Impact Rejection** | `keeper/security.go` | 305-322 | `CheckPriceImpactThreshold()` |
| **Liquidity Lock-up** | `keeper/security.go` | 324-408 | `CreateLiquidityLock()`, `CheckLiquidityLock()` |
| **Order Manipulation** | `keeper/security.go` | 410-471 | `DetectOrderManipulation()`, `FlagOrderManipulation()` |
| **Wash Trading Detection** | `keeper/security.go` | 473-533 | `DetectWashTrading()`, `IncrementWashTradeDetection()` |
| **Dust Attack Prevention** | `keeper/security.go` | 535-550 | `CheckDustAttack()` |
| **Pool Creation Limits** | `keeper/security.go` | 552-634 | `CheckPoolCreationLimits()`, `RecordPoolCreation()` |
| **Circuit Breaker** | `keeper/security.go` | 636-705 | `ActivateCircuitBreaker()`, `IsCircuitBreakerActive()` |

## Integration Wrappers

| Wrapper | File | Lines | Use Case |
|---------|------|-------|----------|
| `SecureSwapExactIn()` | `keeper/security_integration.go` | 11-74 | Production swap with all checks |
| `SecureCreatePool()` | `keeper/security_integration.go` | 76-118 | Production pool creation |
| `SecureAddLiquidity()` | `keeper/security_integration.go` | 120-158 | Production liquidity addition |
| `SecureRemoveLiquidity()` | `keeper/security_integration.go` | 160-194 | Production liquidity removal |

## Storage Keys

| Prefix | Purpose | Key Function |
|--------|---------|--------------|
| `0x10` | Trade block tracking | `TradeBlockKey(address, poolID)` |
| `0x11` | TWAP observations | `TWAPKey(poolID, blockHeight)` |
| `0x12` | Liquidity operation blocks | `LiquidityBlockKey(provider, poolID)` |
| `0x13` | Liquidity locks | `LiquidityLockKey(provider, poolID)` |
| `0x14` | Trade history | `TradeHistoryKey(address)` |
| `0x15` | Pool creation records | `PoolCreationKey(creator)` |
| `0x16` | Circuit breaker state | `CircuitBreakerKey()` |
| `0x17` | Wash trade detection | `WashTradeKey(address, poolID)` |
| `0x18` | Order manipulation | `OrderManipulationKey(address, poolID)` |

## Error Codes

| Code | Error | Description |
|------|-------|-------------|
| 20 | `ErrFrontRunningDetected` | Trade too soon after previous |
| 21 | `ErrFlashLoanDetected` | Rapid liquidity operation |
| 22 | `ErrMEVDetected` | Too many swaps per block |
| 23 | `ErrWashTradingDetected` | Suspicious trading pattern |
| 24 | `ErrOrderManipulation` | Orderbook manipulation detected |
| 25 | `ErrDustAttack` | Trade amount too small |
| 26 | `ErrPoolCreationCooldown` | Pool created too recently |
| 27 | `ErrMaxPoolsExceeded` | Too many pools created |
| 28 | `ErrCircuitBreakerActive` | Trading is paused |

## Default Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `min_block_delay` | 2 blocks | Minimum blocks between trades |
| `max_trade_size_percent` | 20% | Maximum trade as % of pool |
| `max_price_impact_percent` | 10% | Maximum price impact allowed |
| `liquidity_lockup_seconds` | 86400 (24h) | LP token lock period |
| `pool_creation_cooldown` | 3600 (1h) | Cooldown between pool creations |
| `max_pools_per_creator` | 10 | Maximum pools per address |
| `twap_window_blocks` | 100 | TWAP calculation window |
| `min_pool_creation_liquidity` | 1000 tokens | Minimum initial liquidity |
| `min_liquidity_blocks` | 5 | Blocks between add/remove |
| `wash_trade_min_interval` | 60s | Minimum time between trades |
| `min_trade_amount` | 1 token | Minimum trade size |
| `max_order_variance` | 50% | Maximum order size variance |
| `max_swaps_per_block` | 5 | Maximum swaps per block |

## Test Coverage

| Test | File | Lines | Coverage |
|------|------|-------|----------|
| Front-Running | `keeper/security_test.go` | 11-35 | ✅ |
| TWAP Oracle | `keeper/security_test.go` | 37-66 | ✅ |
| Flash Loan | `keeper/security_test.go` | 68-92 | ✅ |
| MEV Protection | `keeper/security_test.go` | 94-113 | ✅ |
| Slippage Limits | `keeper/security_test.go` | 115-130 | ✅ |
| Trade Size Caps | `keeper/security_test.go` | 132-154 | ✅ |
| Price Impact | `keeper/security_test.go` | 156-169 | ✅ |
| Liquidity Lockup | `keeper/security_test.go` | 171-193 | ✅ |
| Order Manipulation | `keeper/security_test.go` | 195-209 | ✅ |
| Wash Trading | `keeper/security_test.go` | 211-240 | ✅ |
| Dust Prevention | `keeper/security_test.go` | 242-255 | ✅ |
| Pool Creation | `keeper/security_test.go` | 257-293 | ✅ |
| Circuit Breaker | `keeper/security_test.go` | 295-326 | ✅ |

## Usage Examples

### Secure Swap
```go
amountOut, price, impact, err := keeper.SecureSwapExactIn(
    ctx,
    "aura1sender...",
    "uaura-usdt",
    sdk.NewCoin("uaura", sdk.NewInt(1000)),
    sdk.NewInt(190),  // min output
    500,              // max slippage (5%)
)
```

### Secure Pool Creation
```go
pool, lpTokens, err := keeper.SecureCreatePool(
    ctx,
    "aura1creator...",
    "uaura",
    "usdt",
    sdk.NewCoin("uaura", sdk.NewInt(10000)),
    sdk.NewCoin("usdt", sdk.NewInt(2000)),
)
```

### Activate Circuit Breaker
```go
keeper.ActivateCircuitBreaker(
    ctx,
    "governance",
    "Emergency maintenance",
    []string{"uaura-usdt"}, // specific pools or empty for all
)
```

### Check TWAP Price
```go
twap, err := keeper.GetTWAPPrice(
    ctx,
    "uaura-usdt",
    100, // window in blocks
)
```

## Monitoring Queries

### Check Circuit Breaker Status
```bash
aurad query dex circuit-breaker
```

### Get TWAP Price
```bash
aurad query dex twap uaura-usdt --window 100
```

### Check Liquidity Lock
```bash
aurad query dex liquidity-lock [provider] [pool-id]
```

### View Trade History
```bash
aurad query dex trade-history [address]
```

### Check Wash Trading Flags
```bash
aurad query dex wash-trade-detection [address] [pool-id]
```

## Common Commands

### Run Security Tests
```bash
cd chain/x/dex
go test -v ./keeper -run Security
```

### Run Specific Test
```bash
go test -v ./keeper -run TestFrontRunningProtection
```

### Generate Coverage
```bash
go test -v ./keeper -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Build Module
```bash
cd chain
make build
```

### Generate Proto Files
```bash
cd proto
buf generate
```

## Security Checklist

Before deploying a swap:
- [ ] Check circuit breaker status
- [ ] Verify trade size within limits
- [ ] Confirm price impact acceptable
- [ ] Check recent trade history
- [ ] Validate minimum delay satisfied

Before creating a pool:
- [ ] Verify sufficient initial liquidity
- [ ] Check creation cooldown
- [ ] Confirm pool count under limit
- [ ] Validate token denoms

Before removing liquidity:
- [ ] Check liquidity lock expiration
- [ ] Verify flash loan protection
- [ ] Confirm circuit breaker inactive

## Important Notes

⚠️ **Production Use**: Always use `Secure*` wrapper functions, not direct keeper calls

⚠️ **Parameter Changes**: All security parameters are governance-adjustable

⚠️ **Emergency Pause**: Circuit breaker can be activated by governance

⚠️ **TWAP Oracle**: Updates automatically on every pool state change

⚠️ **Liquidity Locks**: 24-hour default lock prevents rug pulls

⚠️ **Wash Trading**: 5 suspicious trades trigger automatic flagging

## Quick Diagnostics

If a trade fails, check in order:
1. Circuit breaker active? (`IsCircuitBreakerActive()`)
2. Too soon after last trade? (`CheckFrontRunningProtection()`)
3. Too many trades this block? (`CheckMEVProtection()`)
4. Wash trading detected? (`DetectWashTrading()`)
5. Amount too small? (`CheckDustAttack()`)
6. Trade too large? (`CheckMaxTradeSize()`)
7. Price impact too high? (`CheckPriceImpactThreshold()`)

If pool creation fails:
1. Insufficient liquidity? (< 1000 tokens)
2. Too soon after last pool? (< 1 hour)
3. Too many pools? (≥ 10)

If liquidity removal fails:
1. Circuit breaker active?
2. Too soon after add? (< 5 blocks)
3. Liquidity still locked? (< 24 hours)

## File Locations

```
chain/x/dex/
├── keeper/
│   ├── security.go                 # Core security implementation
│   ├── security_integration.go     # Production wrappers
│   └── security_test.go           # Comprehensive tests
├── types/
│   ├── security_types.go          # Type helpers
│   └── security_keys.go           # Storage keys
└── SECURITY_FEATURES.md           # Full documentation

proto/aura/dex/v1beta1/
└── security.proto                 # Proto definitions
```

## Additional Resources

- **Full Documentation**: `chain/x/dex/SECURITY_FEATURES.md`
- **Implementation Summary**: `DEX_SECURITY_IMPLEMENTATION_SUMMARY.md`
- **Tests**: `chain/x/dex/keeper/security_test.go`
- **Proto Definitions**: `proto/aura/dex/v1beta1/security.proto`

---

**Last Updated**: 2025-11-13
**Version**: 1.0.0
