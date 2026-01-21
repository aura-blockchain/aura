# Architecture

Aura is a Cosmos SDK application with custom modules and supporting services.

## Core Components

- `chain/` Cosmos SDK app and `aurad` binary
- `proto/` protobuf definitions for modules and services
- `contracts/` smart contracts and bindings
- `explorer/`, `faucet/`, `wallet/` supporting services for testnet operations
- `sdk/` client SDKs

## Configuration

- Chain configuration is generated under `~/.aura/config/` after `aurad init`
- Reference configurations live under `networks/`

## Testnet Infrastructure

The public testnet (`aura-mvp-1`) runs across two OVH servers with a validator/sentry architecture.

### Node Topology

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           AURA Testnet (aura-mvp-1)                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  aura-testnet (10.10.0.1)              services-testnet (10.10.0.4)        │
│  ┌─────────────────────────┐           ┌─────────────────────────┐         │
│  │  val1 (hermit)          │           │  val3 (hermit)          │         │
│  │  RPC: 127.0.0.1:26657   │           │  RPC: 127.0.0.1:26657   │         │
│  │  P2P: 26656             │           │  P2P: 26656             │         │
│  │  ~/.aura-mvp-val1       │           │  ~/.aura-mvp-val3       │         │
│  └──────────┬──────────────┘           └──────────┬──────────────┘         │
│             │                                     │                         │
│  ┌──────────┴──────────────┐           ┌──────────┴──────────────┐         │
│  │  val2 (hermit)          │           │  val4 (hermit)          │         │
│  │  RPC: 127.0.0.1:26757   │           │  RPC: 127.0.0.1:26757   │         │
│  │  P2P: 26756             │           │  P2P: 26756             │         │
│  │  ~/.aura-mvp-val2       │           │  ~/.aura-mvp-val4       │         │
│  └──────────┬──────────────┘           └──────────┴──────────────┘         │
│             │                                     │                         │
│             └───────────┐             ┌───────────┘                         │
│                         ▼             ▼                                     │
│  ┌──────────────────────────┐   ┌──────────────────────────┐               │
│  │  sentry1 (public)        │◄──►  sentry2 (public)        │               │
│  │  RPC: 0.0.0.0:26680      │   │  RPC: 0.0.0.0:26680      │               │
│  │  P2P: 0.0.0.0:26681      │   │  P2P: 0.0.0.0:26681      │               │
│  │  API: 0.0.0.0:1319       │   │  API: 0.0.0.0:1319       │               │
│  │  gRPC: 0.0.0.0:9092      │   │  gRPC: 0.0.0.0:9092      │               │
│  │  ~/.aura-mvp-sentry1     │   │  ~/.aura-mvp-sentry2     │               │
│  └──────────────────────────┘   └──────────────────────────┘               │
│             ▲                              ▲                                │
│             │         Public Internet      │                                │
│             └──────────────────────────────┘                                │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Validator Configuration (Hermit Mode)

Validators run in "hermit mode" for security:
- `pex = false` - No public peer discovery
- `persistent_peers` - Only connected to sentries
- `addr_book_strict = false` - Allow private IP connections
- API/gRPC disabled
- RPC bound to localhost only

### Sentry Configuration (Public-Facing)

Sentries expose public services:
- `pex = true` - Public peer discovery enabled
- `persistent_peers` - All validators + other sentries
- `private_peer_ids` - Validator node IDs (never gossiped)
- API/gRPC/RPC bound to 0.0.0.0

### Systemd Services

| Node | Service Name |
|------|--------------|
| val1 | `aurad-mvp-val1` |
| val2 | `aurad-mvp-val2` |
| val3 | `aurad-mvp-val3` |
| val4 | `aurad-mvp-val4` |
| sentry1 | `aurad-mvp-sentry1` |
| sentry2 | `aurad-mvp-sentry2` |

### Binary Management

All nodes use Cosmovisor for binary management:
```
~/.aura-mvp-{node}/cosmovisor/genesis/bin/aurad
```

## Historical Issues (Resolved)

### Non-Determinism in State Machine (Fixed 2026-01-20)

A floating-point arithmetic issue in the economics module (`chain/x/economics/keeper/vesting.go`)
was causing AppHash divergence between nodes. The `big.Float` calculations have been replaced
with deterministic integer arithmetic using `sdkmath.Int`.

**Note**: Nodes that previously diverged from genesis cannot replay from genesis block 0.
They must sync state from an existing healthy node. New nodes joining after this fix is deployed
should be able to replay from genesis normally.

See `docs/PUBLIC_TESTNET_SECURITY.md` for security controls and `scripts/public-testnet-health-check.sh` for automated endpoint checks.
