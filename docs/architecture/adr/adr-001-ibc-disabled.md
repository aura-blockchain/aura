# ADR-001: IBC Disabled by Design

## Status
Accepted

## Context
Aura is designed as a privacy-focused identity and compliance blockchain. IBC (Inter-Blockchain Communication) introduces complexity and potential privacy leakage vectors that conflict with Aura's core mission.

## Decision
IBC modules are implemented but disabled by default. All IBC handlers return `ErrIBCNotEnabled` errors. This allows:
1. Future IBC enablement without code changes
2. Testing IBC compatibility without production exposure
3. Clear separation of cross-chain concerns

## Consequences

### Positive
- Reduced attack surface for privacy-sensitive operations
- Simpler security audit scope
- No cross-chain state synchronization complexity
- Clear upgrade path when IBC is needed

### Negative
- No native cross-chain interoperability initially
- Bridge module handles external chain interaction instead

### Neutral
- IBC code remains in codebase for future use
- Three modules (bridge, identity, compliance) have IBC stubs
