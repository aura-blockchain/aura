# AURA DEX & Cross-Chain Bridge - Complete Specification

**Date:** November 13, 2025
**Status:** Proto Definitions Complete - Ready for Implementation

---

## 🎯 Overview

AURA now has complete proto definitions for:
1. **DEX Module** - AMM liquidity pools + P2P orderbook with HTLC atomic swaps
2. **Bridge Module** - Cross-chain bridges to PAW and XAI blockchains
3. **Altcoin Support** - 15 popular cryptocurrencies
4. **Shared Identity** - Identity verification works across all three chains

**Adapted from:** Crypto/XAI and PAW projects (your existing code)
**Total Proto Definitions:** 1,770+ lines across 7 proto files

---

## 📦 DEX Module (4 Proto Files)

### 1. Liquidity Pools (`liquidity_pool.proto`)

**AMM Implementation:**
- Constant product formula (x * y = k)
- LP token issuance and burning
- Fee collection (0.3% trading + 0.05% protocol)
- Swap statistics and price tracking

**Supported Trading Pairs (15 total):**

| Category | Pairs |
|----------|-------|
| **Stablecoins** | AURA/USDT, AURA/USDC, AURA/DAI |
| **Major Crypto** | AURA/BTC, AURA/ETH, AURA/LTC, AURA/DOGE |
| **Privacy Coins** | AURA/XMR, AURA/ZEC, AURA/DASH |
| **Bitcoin Forks** | AURA/BCH |
| **Cosmos Ecosystem** | AURA/OSMO (Osmosis), AURA/ATOM |
| **Cross-Chain** | AURA/PAW, AURA/XAI |

### 2. P2P Orderbook & HTLC (`swap.proto`)

**P2P Orderbook:**
- Buy/Sell orders on-chain
- Auto-matching engine
- Order lifecycle tracking
- Market price discovery

**HTLC Atomic Swaps:**
- Hash Time-Locked Contracts
- Secret hash locking
- Timelock refunds
- Trustless cross-chain swaps
- No counterparty risk

**Order Types:**
- BUY - Buy AURA (sell other coin)
- SELL - Sell AURA (buy other coin)

**Order Status:**
- PENDING → MATCHED → HTLC_CREATED → COMPLETED

### 3. Transaction Messages (`tx.proto`)

**AMM Messages (4):**
```protobuf
MsgCreatePool       // Create new liquidity pool
MsgAddLiquidity     // Add liquidity, receive LP tokens
MsgRemoveLiquidity  // Burn LP tokens, receive assets
MsgSwapExactIn      // Swap with slippage protection
```

**P2P Orderbook Messages (3):**
```protobuf
MsgCreateOrder   // Create buy/sell order
MsgCancelOrder   // Cancel pending order
MsgExecuteSwap   // Execute matched swap
```

**HTLC Messages (3):**
```protobuf
MsgCreateHTLC   // Create Hash Time-Locked Contract
MsgClaimHTLC    // Claim HTLC with secret
MsgRefundHTLC   // Refund expired HTLC
```

### 4. Query Service (`query.proto`)

**10 Query Endpoints:**
```
/aura/dex/v1beta1/pools                    // List all pools
/aura/dex/v1beta1/pools/{pool_id}          // Get pool details
/aura/dex/v1beta1/quote/{pool_id}/{...}    // Get swap quote
/aura/dex/v1beta1/pools/{pool_id}/stats    // Pool statistics
/aura/dex/v1beta1/orderbook/{pair}         // Orderbook for pair
/aura/dex/v1beta1/orders/{order_id}        // Get order details
/aura/dex/v1beta1/orders/user/{address}    // User's orders
/aura/dex/v1beta1/price/{coin}             // Market price
/aura/dex/v1beta1/coins                    // Supported coins
/aura/dex/v1beta1/htlc/{htlc_id}          // HTLC details
```

---

## 🌉 Bridge Module (3 Proto Files)

### 1. Cross-Chain Bridge (`bridge.proto`)

**Core Features:**
- Lock & mint mechanism for token transfers
- Multi-sig validator verification
- Wrapped tokens (paw.token, xai.coin)
- Shared identity across AURA, PAW, XAI
- Cross-chain swap routing
- Relayer performance tracking

**Data Structures:**
```protobuf
CrossChainTransfer  // Track transfers between chains
ChainConfig         // Config for each connected chain
BridgeValidator     // Authorized relayers
WrappedToken        // Tokens from PAW/XAI on AURA
SharedIdentity      // Identity verification across chains
CrossChainSwap      // Multi-chain swap orchestration
RelayerStats        // Relayer performance metrics
```

**Transfer Lifecycle:**
```
PENDING → CONFIRMED → RELAYED → COMPLETED
         ↓
       FAILED → REFUNDED
```

### 2. Transaction Messages (`tx.proto`)

**Bridge Messages (7):**
```protobuf
MsgLockTokens       // Lock AURA tokens → transfer to PAW/XAI
MsgMintTokens       // Mint wrapped tokens (validator-only)
MsgUnlockTokens     // Unlock after burn proof from target chain
MsgBurnTokens       // Burn wrapped tokens → unlock on source
MsgLinkAddress      // Link addresses for shared identity
MsgCrossChainSwap   // Execute swap across chains
MsgRelayTransfer    // Relay transfer (relayer-only)
```

**Example Flow - AURA to PAW:**
```
1. User: MsgLockTokens (lock 1000 AURA for PAW transfer)
2. Validators: Confirm on AURA
3. Relayer: MsgMintTokens on PAW chain (mint wrapped AURA)
4. User receives wAURA on PAW
```

**Example Flow - PAW to AURA:**
```
1. User: Burn wAURA on PAW
2. Validators: Sign burn proof
3. User: MsgUnlockTokens on AURA (with validator signatures)
4. User receives original AURA
```

### 3. Query Service (`query.proto`)

**11 Query Endpoints:**
```
/aura/bridge/v1beta1/transfers                // All transfers
/aura/bridge/v1beta1/transfers/{transfer_id}  // Transfer status
/aura/bridge/v1beta1/transfers/user/{address} // User's transfers
/aura/bridge/v1beta1/chains                   // Connected chains
/aura/bridge/v1beta1/chains/{chain_id}        // Chain config
/aura/bridge/v1beta1/wrapped                  // All wrapped tokens
/aura/bridge/v1beta1/wrapped/{denom}          // Wrapped token info
/aura/bridge/v1beta1/identity/{address}       // Shared identity
/aura/bridge/v1beta1/swaps/{swap_id}          // Cross-chain swap
/aura/bridge/v1beta1/stats                    // Bridge statistics
/aura/bridge/v1beta1/validators               // Bridge validators
/aura/bridge/v1beta1/relayers/{address}       // Relayer stats
```

---

## 🪙 Supported Altcoins (15 Total)

| # | Symbol | Name | Type | Support Method |
|---|--------|------|------|---------------|
| 1 | USDT | Tether USD | Stablecoin | IBC via Osmosis |
| 2 | USDC | USD Coin | Stablecoin | IBC via Osmosis |
| 3 | DAI | Dai | Stablecoin | IBC via Osmosis |
| 4 | BTC | Bitcoin | Crypto | Wrapped (multi-sig) |
| 5 | ETH | Ethereum | Crypto | Wrapped or IBC |
| 6 | LTC | Litecoin | Crypto | Wrapped |
| 7 | DOGE | Dogecoin | Crypto | Wrapped |
| 8 | XMR | Monero | Privacy | Wrapped |
| 9 | BCH | Bitcoin Cash | Crypto | Wrapped |
| 10 | ZEC | Zcash | Privacy | Wrapped |
| 11 | DASH | Dash | Crypto | Wrapped |
| 12 | OSMO | Osmosis | Cosmos | IBC (native) |
| 13 | ATOM | Cosmos Hub | Cosmos | IBC (native) |
| 14 | PAW | PAW Chain | Custom | Bridge |
| 15 | XAI | XAI Chain | Custom | Bridge |

---

## 🔗 Super-Compatibility Features

### 1. Shared Identity Verification

**Cross-Chain Recognition:**
- Verify identity once on AURA (complete IRs)
- Identity recognized on PAW automatically
- Identity recognized on XAI automatically
- Users can link multiple addresses across chains

**Example:**
```protobuf
SharedIdentity {
  address: "aura1abc..."
  verified_aura: true
  verified_paw: true
  verified_xai: true
  aura_ir_score: 100
  linked_addresses: {
    "paw": "paw1xyz...",
    "xai": "xai1def..."
  }
  reputation_score: 950
}
```

### 2. Unified Liquidity

**Cross-Chain Liquidity Access:**
- AURA DEX pools accessible from PAW wallets
- PAW liquidity accessible from AURA
- XAI tokens tradable on AURA DEX
- Arbitrage opportunities across chains

### 3. Cross-Chain Swaps

**Multi-Hop Swap Routing:**
```
Example: Swap AURA → XAI tokens
Route: AURA → OSMO → XAI
1. Swap AURA for OSMO on AURA DEX
2. Bridge OSMO to Osmosis
3. Swap OSMO for XAI on Osmosis
4. Bridge XAI back to AURA
```

**Automatic routing finds best path and price.**

---

## 💰 Fee Structure

### DEX Module Fees

**AMM Pools:**
- Trading Fee: 0.3% → Liquidity Providers
- Protocol Fee: 0.05% → AURA Treasury
- **Total: 0.35% per swap**

**P2P Orderbook:**
- No trading fees
- Only gas costs
- HTLC gas for atomic swaps

### Bridge Module Fees

**Cross-Chain Transfers:**
- Bridge Fee: 0.1% of transfer amount
- Relayer Fee: 0.05% of transfer amount
- Gas on both chains (paid separately)
- **Total: ~0.15% + gas**

**Wrapped Tokens:**
- Minting Fee: 0.1%
- Burning Fee: 0.1%
- No ongoing fees for holding

---

## 📊 Proto Files Summary

| Module | File | Lines | Purpose |
|--------|------|-------|---------|
| DEX | liquidity_pool.proto | 185 | AMM pool definitions |
| DEX | swap.proto | 175 | P2P orderbook & HTLC |
| DEX | tx.proto | 210 | Transaction messages |
| DEX | query.proto | 220 | Query service |
| Bridge | bridge.proto | 280 | Cross-chain structures |
| Bridge | tx.proto | 250 | Bridge transactions |
| Bridge | query.proto | 450 | Bridge queries |
| **Total** | **7 files** | **1,770 lines** | **Complete specification** |

---

## 🚀 Implementation Roadmap

### Phase 1: DEX Core (Week 1-2)
- ✅ Proto definitions (DONE)
- ⏳ Generate proto code with buf
- ⏳ Implement keeper (liquidity_pool.go)
- ⏳ Implement keeper (orderbook.go)
- ⏳ Implement keeper (htlc.go)
- ⏳ Unit tests for keepers

### Phase 2: Bridge Core (Week 2-3)
- ✅ Proto definitions (DONE)
- ⏳ Generate proto code
- ⏳ Implement keeper (bridge.go)
- ⏳ Implement keeper (validators.go)
- ⏳ Implement keeper (wrapped_tokens.go)
- ⏳ Unit tests for keepers

### Phase 3: CLI Commands (Week 3)
- ⏳ DEX CLI commands (10 tx + 10 query)
- ⏳ Bridge CLI commands (7 tx + 11 query)
- ⏳ Integration tests

### Phase 4: Integration (Week 4)
- ⏳ Update app.go with new modules
- ⏳ Register message handlers
- ⏳ IBC module setup for Osmosis
- ⏳ Relayer service implementation
- ⏳ End-to-end testing

### Phase 5: Wallets & UI (Week 5-6)
- ⏳ Adapt browser wallet from PAW
- ⏳ Adapt desktop wallet (Electron)
- ⏳ Mobile wallet bridge
- ⏳ Trading UI integration

---

## 🎯 Key Advantages Over Source Projects

| Feature | Crypto/PAW | AURA | Improvement |
|---------|-----------|------|-------------|
| Trading Pairs | 11 | 15 | +36% more pairs |
| Altcoin Support | Limited | 15 coins | Comprehensive |
| Cross-Chain | None | PAW + XAI | Native integration |
| Shared Identity | No | Yes | Unified verification |
| IBC Support | No | Yes (Osmosis) | Cosmos ecosystem |
| Atomic Swaps | Yes | Yes (improved) | Better UX |
| AMM Pools | Yes | Yes | Same quality |
| On-Chain Data | Python DB | Cosmos SDK | Production-grade |
| Governance | No | Yes | Decentralized control |

---

## 📝 Example Use Cases

### Use Case 1: Liquidity Provider on AURA
```bash
# Create AURA/USDT pool with initial liquidity
aurad tx dex create-pool uaura usdt \
  1000000uaura 200000usdt \
  --from alice

# Add more liquidity later
aurad tx dex add-liquidity pool-1 \
  500000uaura 100000usdt \
  --from bob

# Remove liquidity
aurad tx dex remove-liquidity pool-1 50000lp \
  --from alice
```

### Use Case 2: Swap AURA for USDT
```bash
# Get quote first
aurad query dex quote pool-1 uaura 1000000

# Execute swap with slippage protection
aurad tx dex swap pool-1 \
  1000000uaura \
  --min-out 195000 \
  --max-slippage 500 \
  --from alice
```

### Use Case 3: P2P Atomic Swap
```bash
# Alice creates sell order
aurad tx dex create-order sell \
  1000uaura btc 0.0001 \
  --from alice

# Bob's order auto-matches
aurad tx dex create-order buy \
  1000uaura btc 0.0001 \
  --from bob

# HTLC automatically created
# Both parties complete swap
```

### Use Case 4: Transfer AURA to PAW Chain
```bash
# Lock AURA tokens
aurad tx bridge lock-tokens paw \
  paw1xyz... 1000000uaura \
  --from alice

# Check transfer status
aurad query bridge transfer <transfer-id>

# Alice receives wAURA on PAW chain
# (relayers handle automatically)
```

### Use Case 5: Link Identity Across Chains
```bash
# Link AURA, PAW, and XAI addresses
aurad tx bridge link-address \
  aura1abc... paw1xyz... xai1def... \
  --paw-sig <sig> \
  --xai-sig <sig> \
  --from alice

# Now verified on all three chains!
aurad query bridge identity aura1abc...
```

### Use Case 6: Cross-Chain Swap (AURA → XAI)
```bash
# Swap AURA for XAI tokens in one command
aurad tx bridge cross-chain-swap \
  aura xai 1000000uaura xai.token \
  --min-out 500000 \
  --from alice

# Automatic routing: AURA → OSMO → XAI
```

---

## 🎉 Status Summary

**Proto Definitions:** ✅ 100% Complete (1,770 lines)
**Keeper Implementation:** ⏳ 0% (next step)
**CLI Commands:** ⏳ 0%
**Integration:** ⏳ 0%
**Testing:** ⏳ 0%
**Wallets/UI:** ⏳ 0%

**Overall Progress:** ~20% Complete

---

## 📚 Documentation to Create

1. **DEX User Guide** - How to provide liquidity, swap, trade
2. **Bridge User Guide** - How to transfer between chains
3. **Developer Guide** - How to integrate with DEX/Bridge
4. **Relayer Setup Guide** - How to run a bridge relayer
5. **Validator Guide** - How to become a bridge validator
6. **Trading Strategies** - Arbitrage, liquidity provision tips
7. **Security Audit Report** - Third-party security review

---

## 🔐 Security Considerations

**DEX Module:**
- Constant product formula prevents price manipulation
- Slippage protection on all swaps
- LP tokens prevent rug pulls
- Protocol fees governance-controlled

**Bridge Module:**
- Multi-sig validator verification (2/3+ required)
- Timelock on cross-chain transfers
- Burn/mint mechanism prevents double-spending
- Relayer slashing for malicious behavior
- Rate limits on transfers

**HTLC Atomic Swaps:**
- Cryptographic hash locking
- Timelock refunds prevent fund loss
- No trusted third party needed
- Secret reveal mechanism trustless

---

## 🌟 Next Immediate Actions

1. **Generate Proto Code:**
   ```bash
   cd proto
   buf generate
   ```

2. **Implement DEX Keeper:**
   - Start with `chain/x/dex/keeper/liquidity_pool.go`
   - Port constant product formula from Python
   - Add comprehensive unit tests

3. **Implement Bridge Keeper:**
   - Start with `chain/x/bridge/keeper/keeper.go`
   - Implement lock/mint/burn/unlock logic
   - Add validator signature verification

4. **Create CLI Commands:**
   - DEX commands first (higher priority)
   - Bridge commands second
   - Test all commands

5. **Integration:**
   - Update app.go
   - Register modules
   - Run integration tests

---

**Document Status:** Complete Specification
**Last Updated:** 2025-11-13
**Ready for:** Implementation Phase

🚀 **AURA is now ready to become a complete DeFi ecosystem with cross-chain capabilities!**
