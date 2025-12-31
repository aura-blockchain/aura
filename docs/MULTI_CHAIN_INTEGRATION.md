# Multi-Chain Integration Guide

## Overview

Aura is part of a three-chain ecosystem alongside PAW and XAI, enabling:
- **IBC transfers** between Aura and PAW (Cosmos SDK chains)
- **Unified wallet** experience across all three chains
- **Cross-chain liquidity** via PAW DEX

## Chain Architecture

| Chain | Type | Coin Type | Prefix | Features |
|-------|------|-----------|--------|----------|
| Aura | Cosmos SDK | 118 | `aura` | IBC, CosmWasm |
| PAW | Cosmos SDK | 118 | `paw` | IBC, DEX |
| XAI | EVM-compatible | 22593 | `xai` | AI Trading |

## Unified Wallet

The shared wallet library (`wallet/shared/`) provides:
- Single mnemonic for all chains
- Consistent address derivation
- IBC transfer support
- Keplr chain configurations

### Usage

```typescript
import { MultiChainWallet } from '@aura-ecosystem/multi-chain-wallet';

// Create new wallet
const wallet = await MultiChainWallet.create();

// Get addresses for all chains
const accounts = await wallet.getMainnetAccounts();
// Returns: [{ aura1..., paw1..., xai1... }]

// Linked addresses (same public key for Aura and PAW)
const linked = await wallet.getLinkedCosmosAddresses();
```

### Address Linking

Because Aura and PAW share coin type 118:
- Same mnemonic → Same public key → Linked addresses
- `aura1abc...` and `paw1abc...` represent the same identity
- Useful for airdrop eligibility across chains

## IBC Integration

### Hermes Relayer Configuration

The relayer config at `config/hermes/config.toml` includes:
- Aura chain (aura-local-4)
- PAW chain (paw-testnet-1)

### IBC Channels

| Source | Destination | Channel |
|--------|-------------|---------|
| Aura | PAW | channel-0 |
| PAW | Aura | channel-0 |

### IBC Transfer Example

```typescript
const msg = wallet.buildIBCTransferMsg({
  sourceChain: 'aura-mainnet-1',
  destChain: 'paw-mainnet-1',
  sender: auraAddress,
  receiver: pawAddress,
  amount: { denom: 'uaura', amount: '1000000' },
  sourceChannel: 'channel-0',
});
```

## Keplr Integration

```typescript
import { toKeplrChainInfo, AURA_CONFIG } from '@aura-ecosystem/multi-chain-wallet';

// Register Aura with Keplr
await window.keplr.experimentalSuggestChain(toKeplrChainInfo(AURA_CONFIG));
```

## Network Ports

| Service | Aura | PAW | XAI |
|---------|------|-----|-----|
| RPC | 10657 | 11657 | 12657 |
| REST | 10317 | 11317 | 12317 |
| gRPC | 10090 | 11090 | 12090 |

## Testing

```bash
cd wallet/shared
npm test
```

All 36 tests verify:
- Wallet creation and import
- Multi-chain address derivation
- Address linking between Aura/PAW
- IBC transfer message building
- Keplr configuration generation
