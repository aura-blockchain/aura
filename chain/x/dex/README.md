# DEX Module

## Overview

The DEX (Decentralized Exchange) module provides comprehensive trading functionality including AMM-based liquidity pools with constant product formula (x*y=k), P2P orderbook with HTLC atomic swaps, commit-reveal scheme for front-running protection, and cross-chain trading capabilities. It supports slippage protection, batch execution, and comprehensive security features.

## Features

- **AMM Liquidity Pools**: Constant product (x*y=k) automated market maker
- **P2P Orderbook**: Peer-to-peer order matching with AURA as base currency
- **HTLC Atomic Swaps**: Hash Time-Locked Contracts for trustless trading
- **Front-Running Protection**: Commit-reveal scheme for MEV resistance
- **Batch Execution**: Group orders for fair pricing
- **Slippage Protection**: Configurable maximum slippage tolerance
- **Fee Tiers**: Multiple fee levels based on trading volume
- **LP Token Management**: Fungible liquidity provider tokens
- **Cross-Chain Support**: Integration with bridge module for multi-chain trading

## State

### AMM Liquidity Pools
- **LiquidityPool**: Pool reserves, fees, LP token supply
- **Pool ID**: Unique identifier for each trading pair
- **Reserves**: TokenA and TokenB amounts
- **LP Tokens**: Shares representing pool ownership

### P2P Orderbook
- **SwapOrder**: Buy/sell orders with AURA pairs
- **Order Types**: Buy AURA, Sell AURA
- **Order Status**: Pending, matched, executed, cancelled, expired

### HTLC Contracts
- **HashTimeLock**: Secret hash and timelock conditions
- **HTLC Status**: Open, claimed, refunded
- **Cross-Chain HTLCs**: Linked contracts across chains

### Security Features
- **Rate Limiting**: Per-user operation limits
- **Order Commitments**: Two-phase commit-reveal
- **Price Oracles**: TWAP and external price feeds
- **Circuit Breakers**: Automatic pause on anomalies

## Messages

### AMM Liquidity Pool Messages

#### MsgCreatePool
Create new liquidity pool with initial reserves.

**Example**:
```json
{
  "creator": "aura1...",
  "denom_a": "uaura",
  "denom_b": "ubtc",
  "amount_a": {"denom": "uaura", "amount": "1000000"},
  "amount_b": {"denom": "ubtc", "amount": "100000"}
}
```

**Response**:
```json
{
  "pool_id": "pool_1",
  "lp_tokens": "1000000"
}
```

#### MsgAddLiquidity
Add liquidity to existing pool, mints LP tokens.

**Example**:
```json
{
  "provider": "aura1...",
  "pool_id": "pool_1",
  "amount_a": {"denom": "uaura", "amount": "500000"},
  "amount_b": {"denom": "ubtc", "amount": "50000"}
}
```

**Response**:
```json
{
  "lp_tokens_minted": "500000",
  "pool_share_percent": "0.333333"
}
```

#### MsgRemoveLiquidity
Remove liquidity by burning LP tokens.

**Example**:
```json
{
  "provider": "aura1...",
  "pool_id": "pool_1",
  "lp_tokens": "250000"
}
```

**Response**:
```json
{
  "amount_a": {"denom": "uaura", "amount": "250000"},
  "amount_b": {"denom": "ubtc", "amount": "25000"}
}
```

#### MsgSwapExactIn
Swap exact input for minimum output with slippage protection.

**Example**:
```json
{
  "sender": "aura1...",
  "pool_id": "pool_1",
  "coin_in": {"denom": "uaura", "amount": "10000"},
  "min_amount_out": "950",
  "max_slippage_bps": 500
}
```

**Response**:
```json
{
  "amount_out": "980",
  "effective_price": "0.098",
  "price_impact_percent": "0.15"
}
```

### P2P Orderbook Messages

#### MsgCreateOrder
Create buy or sell order for AURA pairs.

**Example**:
```json
{
  "creator": "aura1...",
  "order_type": "BUY_AURA",
  "aura_amount": "1000000",
  "other_coin": "ubtc",
  "other_amount": "100000"
}
```

**Response**:
```json
{
  "order_id": "order_abc123",
  "status": "PENDING",
  "message": "order created"
}
```

#### MsgCancelOrder
Cancel pending order, refunds escrowed funds.

**Example**:
```json
{
  "creator": "aura1...",
  "order_id": "order_abc123"
}
```

**Response**:
```json
{
  "success": true
}
```

#### MsgExecuteSwap
Execute matched order through orderbook.

**Example**:
```json
{
  "initiator": "aura1...",
  "order_id": "order_abc123",
  "secret": "preimage_data"
}
```

**Response**:
```json
{
  "success": true,
  "swap_id": "swap_xyz789"
}
```

### Commit-Reveal Messages (Front-Running Protection)

#### MsgCommitOrder
Phase 1: Commit order hash without revealing details.

**Example**:
```json
{
  "sender": "aura1...",
  "commit_hash": "sha256_of_order_plus_salt"
}
```

**Response**:
```json
{
  "commit_id": "commit_123",
  "reveal_deadline": "2025-12-09T10:05:00Z"
}
```

#### MsgRevealOrder
Phase 2: Reveal order details matching commit hash.

**Example**:
```json
{
  "sender": "aura1...",
  "commit_id": "commit_123",
  "order_type": "BUY_AURA",
  "aura_amount": "1000000",
  "other_coin": "ubtc",
  "other_amount": "100000",
  "salt": "random_salt_bytes"
}
```

**Response**:
```json
{
  "success": true,
  "order_id": "order_abc123",
  "message": "queued for batch execution"
}
```

### HTLC Atomic Swap Messages

#### MsgCreateHTLC
Create Hash Time-Locked Contract for atomic swap.

**Example**:
```json
{
  "sender": "aura1...",
  "recipient": "aura1other...",
  "amount": {"denom": "uaura", "amount": "10000"},
  "secret_hash": "sha256_of_secret",
  "timelock_duration": 3600
}
```

**Response**:
```json
{
  "htlc_id": "htlc_xyz789"
}
```

#### MsgClaimHTLC
Claim HTLC by revealing secret preimage.

**Example**:
```json
{
  "recipient": "aura1other...",
  "htlc_id": "htlc_xyz789",
  "secret": "secret_preimage"
}
```

**Response**:
```json
{
  "success": true
}
```

#### MsgRefundHTLC
Refund expired HTLC after timelock.

**Example**:
```json
{
  "sender": "aura1...",
  "htlc_id": "htlc_xyz789"
}
```

**Response**:
```json
{
  "success": true
}
```

## Queries

### QueryPool
Get liquidity pool details.

**Request**:
```bash
aurad query dex pool pool_1
```

**Response**:
```json
{
  "pool": {
    "pool_id": "pool_1",
    "denom_a": "uaura",
    "denom_b": "ubtc",
    "reserve_a": "1000000",
    "reserve_b": "100000",
    "total_lp_tokens": "1000000",
    "fee_rate": "0.003"
  }
}
```

### QueryAllPools
List all liquidity pools with pagination.

**Request**:
```bash
aurad query dex all-pools
```

### QueryGetQuote
Get swap quote without executing.

**Request**:
```bash
aurad query dex quote pool_1 uaura 10000
```

**Response**:
```json
{
  "estimated_output": "980",
  "effective_price": "0.098",
  "price_impact": "0.15",
  "fee": "30"
}
```

### QueryPoolStats
Get pool statistics and metrics.

**Request**:
```bash
aurad query dex pool-stats pool_1
```

**Response**:
```json
{
  "stats": {
    "total_volume_24h": "5000000",
    "total_fees_24h": "15000",
    "total_transactions_24h": 150,
    "tvl": "1100000"
  }
}
```

### QueryOrderbook
Get orderbook for trading pair.

**Request**:
```bash
aurad query dex orderbook aura-btc
```

**Response**:
```json
{
  "buy_orders": [...],
  "sell_orders": [...],
  "spread": "0.001"
}
```

### QueryOrder
Get specific order details.

**Request**:
```bash
aurad query dex order order_abc123
```

**Response**:
```json
{
  "order": {
    "order_id": "order_abc123",
    "user_address": "aura1...",
    "order_type": "BUY_AURA",
    "status": "PENDING",
    "aura_amount": "1000000",
    "other_coin": "ubtc",
    "other_amount": "100000"
  }
}
```

### QueryUserOrders
Get all orders for user address.

**Request**:
```bash
aurad query dex user-orders aura1...
```

### QueryMarketPrice
Get current market price for coin.

**Request**:
```bash
aurad query dex price ubtc
```

**Response**:
```json
{
  "price": "0.098",
  "source": "pool_1",
  "last_updated": "2025-12-09T10:00:00Z"
}
```

### QuerySpotPrice
Get instantaneous pool price.

**Request**:
```bash
aurad query dex spot pool_1 uaura ubtc
```

**Response**:
```json
{
  "spot_price": "0.100"
}
```

### QueryHTLC
Get HTLC contract details.

**Request**:
```bash
aurad query dex htlc htlc_xyz789
```

**Response**:
```json
{
  "htlc": {
    "htlc_id": "htlc_xyz789",
    "sender": "aura1...",
    "recipient": "aura1other...",
    "amount": "10000",
    "secret_hash": "sha256_hash",
    "timelock": "2025-12-09T11:00:00Z",
    "status": "OPEN"
  }
}
```

## Events

| Event Type | Attributes | Description |
|------------|------------|-------------|
| `pool_created` | `pool_id`, `creator`, `denom_a`, `denom_b` | New pool created |
| `liquidity_added` | `pool_id`, `provider`, `lp_tokens_minted` | Liquidity added to pool |
| `liquidity_removed` | `pool_id`, `provider`, `lp_tokens_burned` | Liquidity removed from pool |
| `swap_executed` | `pool_id`, `sender`, `amount_in`, `amount_out` | AMM swap completed |
| `order_created` | `order_id`, `user`, `order_type`, `status` | Order created in orderbook |
| `order_matched` | `order_id`, `matched_order_id` | Orders matched |
| `order_cancelled` | `order_id`, `user` | Order cancelled |
| `order_expired` | `order_id` | Order expired |
| `order_committed` | `commit_id`, `reveal_deadline` | Order commitment created |
| `order_revealed` | `commit_id`, `order_id` | Order revealed |
| `htlc_created` | `htlc_id`, `sender`, `recipient`, `amount` | HTLC contract created |
| `htlc_claimed` | `htlc_id`, `recipient` | HTLC claimed with secret |
| `htlc_refunded` | `htlc_id`, `sender` | HTLC refunded after timeout |

## Errors

| Code | Name | Description |
|------|------|-------------|
| 1 | `ErrPoolNotFound` | Liquidity pool does not exist |
| 2 | `ErrInsufficientLiquidity` | Pool lacks liquidity for swap |
| 3 | `ErrSlippageExceeded` | Actual slippage exceeds tolerance |
| 4 | `ErrInvalidPoolRatio` | Pool reserves ratio invalid |
| 5 | `ErrOrderNotFound` | Order ID not found |
| 6 | `ErrOrderNotPending` | Order not in pending status |
| 7 | `ErrOrderExpired` | Order past expiration time |
| 8 | `ErrInvalidOrderType` | Order type not recognized |
| 9 | `ErrHTLCNotFound` | HTLC contract not found |
| 10 | `ErrHTLCExpired` | HTLC timelock expired |
| 11 | `ErrInvalidSecret` | Secret does not match hash |
| 12 | `ErrCommitmentNotFound` | Commit ID not found |
| 13 | `ErrRevealDeadlinePassed` | Reveal window expired |
| 14 | `ErrInvalidCommitment` | Revealed data doesn't match hash |
| 15 | `ErrInsufficientBalance` | User balance too low |

## CLI Examples

### Create liquidity pool
```bash
aurad tx dex create-pool \
  --denom-a uaura \
  --denom-b ubtc \
  --amount-a 1000000uaura \
  --amount-b 100000ubtc \
  --from mykey
```

### Add liquidity
```bash
aurad tx dex add-liquidity \
  --pool-id pool_1 \
  --amount-a 500000uaura \
  --amount-b 50000ubtc \
  --from mykey
```

### Swap with slippage protection
```bash
aurad tx dex swap-exact-in \
  --pool-id pool_1 \
  --coin-in 10000uaura \
  --min-amount-out 950 \
  --max-slippage-bps 500 \
  --from mykey
```

### Create order (commit-reveal)
```bash
# Step 1: Commit
aurad tx dex commit-order \
  --commit-hash $(echo -n "order_data_salt" | sha256sum | cut -d' ' -f1) \
  --from mykey

# Step 2: Reveal (within 5 minutes)
aurad tx dex reveal-order \
  --commit-id commit_123 \
  --order-type buy \
  --aura-amount 1000000 \
  --other-coin ubtc \
  --other-amount 100000 \
  --salt random_salt \
  --from mykey
```

### Create HTLC
```bash
aurad tx dex create-htlc \
  --recipient aura1other... \
  --amount 10000uaura \
  --secret-hash $(echo -n "mysecret" | sha256sum | cut -d' ' -f1) \
  --timelock-duration 3600 \
  --from mykey
```

## Integration Notes

### For Wallet Developers

1. **Pool Display**: Show pool reserves, LP token balance, share percentage
2. **Slippage Settings**: Allow users to configure max slippage (default 0.5%-1%)
3. **Order Management**: Display active orders with cancel buttons
4. **HTLC Tracking**: Show pending HTLCs with countdown timers
5. **Commit-Reveal UX**: Guide users through two-step process with clear timing

### Security Considerations

- **Slippage Protection**: Always set reasonable max_slippage_bps (50-500 bps typical)
- **Timelock Duration**: Use appropriate HTLC timelocks (3600s minimum recommended)
- **Secret Management**: Generate cryptographically secure random secrets for HTLCs
- **Front-Running**: Use commit-reveal for large orders to prevent MEV exploitation
- **Price Impact**: Display price impact percentage before swap confirmation

### Best Practices

- **Liquidity Provision**: Add liquidity in correct ratio to minimize loss
- **Order Expiration**: Set reasonable expiration times (24-168 hours typical)
- **Batch Execution**: Wait for batch execution when using commit-reveal
- **Fee Tiers**: Consider fee tiers when selecting pools
- **Cross-Chain**: Verify bridge status before cross-chain swaps
