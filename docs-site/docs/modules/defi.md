---
sidebar_position: 4
---

# DeFi Modules

Decentralized exchange, cross-chain bridge, and economics.

## dex

Decentralized exchange with AMM and order book.

### Key Features
- Automated Market Maker (AMM) pools
- Limit orders
- Multi-hop routing
- LP token rewards

### Messages
| Message | Description |
|---------|-------------|
| `MsgSwap` | Execute token swap |
| `MsgAddLiquidity` | Add to liquidity pool |
| `MsgRemoveLiquidity` | Withdraw from pool |
| `MsgPlaceOrder` | Place limit order |
| `MsgCancelOrder` | Cancel limit order |
| `MsgCreatePool` | Create new trading pair |

### Queries
```bash
aurad query dex pools
aurad query dex pool <pool-id>
aurad query dex quote --input <amount> --output-denom <denom>
aurad query dex orders-by-address <address>
aurad query dex liquidity-positions <address>
```

### Pool Types
- **Standard AMM**: 50/50 constant product
- **Stableswap**: Optimized for pegged assets
- **Weighted**: Custom token ratios

---

## bridge

Cross-chain asset transfers via IBC and custom bridges.

### Key Features
- IBC token transfers
- Multi-signature bridge operators
- Fraud proofs
- Rate limiting

### Messages
| Message | Description |
|---------|-------------|
| `MsgInitiateTransfer` | Start cross-chain transfer |
| `MsgCompleteTransfer` | Finalize incoming transfer |
| `MsgRegisterBridgeOperator` | Register as bridge operator |
| `MsgSubmitFraudProof` | Challenge invalid transfer |

### Queries
```bash
aurad query bridge pending-transfers
aurad query bridge transfer <transfer-id>
aurad query bridge operators
aurad query bridge supported-chains
```

### Supported Chains
- Cosmos Hub (ATOM)
- Osmosis (OSMO)
- Ethereum (via bridge operators)

---

## economics

Tokenomics, fee distribution, and economic parameters.

### Key Features
- Fee distribution to validators
- Staking rewards calculation
- Inflation schedule
- Community pool management

### Queries
```bash
aurad query economics inflation-rate
aurad query economics annual-provisions
aurad query economics distribution-params
aurad query economics community-pool
```

### Fee Distribution
| Recipient | Share |
|-----------|-------|
| Validators | 70% |
| Community Pool | 20% |
| AI Assistants | 10% |
