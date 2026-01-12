# AURA MVP Modules

The MVP release includes 12 essential modules for credential verification on the AURA blockchain.

## Module Overview

| Module | Purpose | Status |
|--------|---------|--------|
| auth | Account authentication and management | SDK Standard |
| bank | Token transfers and balances | SDK Standard |
| staking | Validator staking and delegation | SDK Standard |
| slashing | Validator misbehavior penalties | SDK Standard |
| distribution | Block rewards distribution | SDK Standard |
| governance | On-chain governance proposals | AURA Custom |
| identity | DID-based identity management | AURA Core |
| vcregistry | Verifiable Credential registry | AURA Core |
| dataregistry | Data item storage and verification | AURA Core |
| compliance | Regulatory compliance checks | AURA Core |
| prevalidation | Transaction pre-validation | AURA Core |
| wasm | CosmWasm smart contracts | SDK Extension |

## Module Dependencies

```
                    ┌──────────┐
                    │   auth   │
                    └────┬─────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
    ┌────▼────┐    ┌────▼────┐    ┌─────▼─────┐
    │  bank   │    │ staking │    │ slashing  │
    └────┬────┘    └────┬────┘    └───────────┘
         │              │
         │         ┌────▼────────┐
         │         │distribution │
         │         └─────────────┘
         │
    ┌────▼────────────────────────────────┐
    │           AURA Modules              │
    │  ┌──────────┬──────────┬──────────┐ │
    │  │ identity │vcregistry│dataregistry│
    │  └────┬─────┴────┬─────┴────┬─────┘ │
    │       │          │          │       │
    │  ┌────▼──────────▼──────────▼────┐  │
    │  │         compliance            │  │
    │  └────┬──────────────────────────┘  │
    │       │                             │
    │  ┌────▼─────────┐                   │
    │  │prevalidation │                   │
    │  └──────────────┘                   │
    └─────────────────────────────────────┘
```

## AURA Core Modules

### identity

Manages decentralized identities (DIDs) on-chain.

**CLI Commands:**
```bash
# Register a new identity
aurad tx identity register [did] --from <key>

# Query identity
aurad query identity did [did]

# Update identity
aurad tx identity update [did] [document-json] --from <key>
```

### vcregistry

Stores and verifies Verifiable Credentials.

**CLI Commands:**
```bash
# Register credential schema
aurad tx vcregistry register-schema [schema-json] --from <key>

# Issue credential
aurad tx vcregistry issue [credential-json] --from <key>

# Verify credential
aurad query vcregistry verify [credential-id]

# Revoke credential
aurad tx vcregistry revoke [credential-id] --from <key>
```

### dataregistry

Stores data items with access control and verification.

**CLI Commands:**
```bash
# Store data item
aurad tx dataregistry store [data-id] [content-hash] --from <key>

# Query data item
aurad query dataregistry item [data-id]

# Update data item
aurad tx dataregistry update [data-id] [new-content-hash] --from <key>

# Delete data item
aurad tx dataregistry delete [data-id] --from <key>
```

**Access Modes:**
- `PRIVATE` - Only owner can access
- `PUBLIC` - Anyone can access
- `WHITELIST` - Specific addresses can access
- `VERIFIED_USERS` - Only verified users can access

### compliance

Checks transactions against compliance rules.

**CLI Commands:**
```bash
# Query compliance rules
aurad query compliance rules

# Check address compliance
aurad query compliance check [address]
```

### governance

Handles on-chain governance for AURA-specific proposals.

**CLI Commands:**
```bash
# Submit proposal
aurad tx governance submit-proposal [proposal-json] --from <key>

# Vote on proposal
aurad tx governance vote [proposal-id] [yes|no|abstain] --from <key>

# Query proposal
aurad query governance proposal [proposal-id]
```

### prevalidation

Pre-validates transactions before execution.

Runs automatically on all transactions. No direct CLI interaction required.

## Deferred Modules

The following modules are excluded from MVP and will be added in future releases:

- **security** - Advanced security monitoring
- **walletsecurity** - Wallet security features
- **validatorsecurity** - Validator security
- **networksecurity** - Network-level security
- **dex** - Decentralized exchange
- **bridge** - Cross-chain bridge
- **aiassistant** - AI-powered assistance
- **confidencescore** - Credential confidence scoring
- **incidentresponse** - Security incident handling
- **monitoring** - Advanced monitoring

## Building MVP

```bash
# Build MVP binary
make build-mvp

# Run MVP tests
make test-mvp

# Generate MVP genesis
./scripts/generate-mvp-genesis.sh aura-mvp-1

# Build release artifacts
./scripts/build-mvp-release.sh v1.0.0-mvp
```

## Testing

Run module-specific tests:

```bash
# All MVP module tests
make test-mvp

# Individual modules
go test ./x/identity/... -v
go test ./x/vcregistry/... -v
go test ./x/dataregistry/... -v
go test ./x/compliance/... -v
go test ./x/governance/... -v
go test ./x/prevalidation/... -v
```

## Upgrade Path

MVP is designed to support seamless upgrade to full release:

1. **State Migration**: MVP state is fully compatible with full release
2. **Module Activation**: Deferred modules can be enabled via governance
3. **Genesis Export**: `aurad export` produces valid genesis for upgrades

To upgrade from MVP to full release:
1. Export state: `aurad export --height <height> > genesis-export.json`
2. Migrate genesis: `aurad migrate genesis-export.json --target v2.0.0`
3. Restart with full binary
