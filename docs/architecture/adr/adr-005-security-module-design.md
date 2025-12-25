# ADR-005: Security Module Design

## Status
Accepted

## Context
Aura needs centralized security controls for rate limiting, reentrancy protection, module pausing, and security event logging.

## Decision
The security module provides cross-cutting security concerns:
1. **Rate Limiting**: Per-address and per-action rate limits
2. **Reentrancy Guards**: Prevent reentrant calls to sensitive functions
3. **Module Pausing**: Emergency pause capability for any module
4. **Security Events**: Centralized security event logging
5. **Input Validation**: Common validation helpers

Other modules depend on security module for these features rather than implementing their own.

## Consequences

### Positive
- Consistent security behavior across modules
- Single point for security policy updates
- Centralized audit trail

### Negative
- Security module is critical dependency
- Must be carefully tested

### Neutral
- All modules import security keeper
