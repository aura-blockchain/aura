# ADR-002: Store Key Prefix Conventions

## Status
Accepted

## Context
Cosmos SDK modules use KVStore with byte prefixes to organize data. Consistent prefix conventions prevent collisions and improve code maintainability.

## Decision
All modules follow these prefix conventions:
1. Single-byte prefixes for primary data types (0x01, 0x02, etc.)
2. Intentional aliases documented with comments when prefixes are reused
3. Dead code prefixes (unused) are acceptable if clearly marked
4. Each module's `types/keys.go` defines all prefixes

Verified patterns:
- contractregistry: ContractMetadataKeyPrefix is dead code (never used)
- prevalidation: PreValidatedTxPrefix is intentional alias
- privacy: MixingPoolKeyPrefix is intentional alias with comment
- validatorsecurity: SentryNodeKey is intentional alias with comment
- walletsecurity: SessionConfigPrefix is backward compat alias

## Consequences

### Positive
- No accidental data collisions
- Clear documentation of prefix purposes
- Easy to audit store usage

### Negative
- Some unused prefixes remain in code

### Neutral
- Aliases require comments explaining intent
