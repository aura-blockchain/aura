---
sidebar_position: 7
---

# Infrastructure Modules

Core utilities, registries, and smart contract support.

## dataregistry

On-chain data registration and verification.

### Key Features
- Data hash registration
- Timestamp proofs
- Data integrity verification
- Metadata storage

### Messages
| Message | Description |
|---------|-------------|
| `MsgRegisterData` | Register data hash |
| `MsgUpdateData` | Update metadata |
| `MsgVerifyData` | Verify data existence |

### Queries
```bash
aurad query dataregistry data <hash>
aurad query dataregistry data-by-owner <address>
aurad query dataregistry verify <hash>
```

---

## contractregistry

Smart contract registry and verification.

### Key Features
- Contract metadata
- Source verification
- Audit status tracking
- Version management

### Messages
| Message | Description |
|---------|-------------|
| `MsgRegisterContract` | Register deployed contract |
| `MsgVerifySource` | Submit verified source |
| `MsgUpdateAuditStatus` | Update audit status |

### Queries
```bash
aurad query contractregistry contract <address>
aurad query contractregistry contracts-by-creator <address>
aurad query contractregistry audited-contracts
```

---

## wasm

CosmWasm smart contract execution.

### Key Features
- WASM contract deployment
- Contract instantiation
- Contract execution
- Contract migration

### Messages
| Message | Description |
|---------|-------------|
| `MsgStoreCode` | Upload WASM binary |
| `MsgInstantiateContract` | Create contract instance |
| `MsgExecuteContract` | Execute contract |
| `MsgMigrateContract` | Upgrade contract |

### Queries
```bash
aurad query wasm codes
aurad query wasm code <code-id>
aurad query wasm contract <address>
aurad query wasm contract-state smart <address> <query>
```

### Development
```bash
# Build contract
cargo wasm

# Optimize
docker run --rm -v "$(pwd)":/code \
  cosmwasm/optimizer:0.15.0

# Deploy
aurad tx wasm store artifacts/mycontract.wasm \
  --from my-wallet --gas auto -y
```

---

## auth

Authentication extensions for Aura.

### Key Features
- Account types
- Fee grants
- Authz authorizations
- Vesting accounts

### Queries
```bash
aurad query auth account <address>
aurad query auth accounts
aurad query feegrant grants <grantee>
aurad query authz grants <granter> <grantee>
```

---

## common

Shared utilities across modules.

### Features
- Common type definitions
- Utility functions
- Error codes
- Constants

This module provides no direct user-facing functionality.

---

## internal

Internal module utilities (not user-facing).

### Features
- Module communication
- State access helpers
- Testing utilities

Internal use only; no public API.
