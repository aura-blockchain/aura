# ADR-004: Privacy Module Architecture

## Status
Accepted

## Context
Aura requires privacy features for sensitive identity and financial data while maintaining regulatory compliance capabilities.

## Decision
The privacy module implements:
1. **Shielded Balances**: Encrypted balance storage with view keys
2. **Mixing Pools**: Transaction mixing for unlinkability
3. **ZK Proofs**: Zero-knowledge proofs for private transactions
4. **Selective Disclosure**: Compliance-authorized data access

Key design choices:
- View keys allow authorized parties to audit
- GDPR compliance through key management
- Migration path from v1 (removed private keys in v2)

## Consequences

### Positive
- Strong privacy for users
- Regulatory compliance possible
- Auditable by authorized parties

### Negative
- Complex cryptographic implementation
- Higher computational requirements

### Neutral
- Requires key management infrastructure
