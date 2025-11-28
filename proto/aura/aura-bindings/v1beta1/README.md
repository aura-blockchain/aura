# AURA Custom Bindings Protocol Buffers

## Overview

This directory contains Protocol Buffer definitions for AURA's custom CosmWasm bindings. These proto files define the query and message interfaces that smart contracts use to interact with AURA's native modules.

## Purpose

The proto definitions specify:

- Custom query message types (AuraQuery)
- Custom message types (AuraMsg)
- Response message structures
- Integration with all AURA modules

## Files

### query.proto
Query message definitions:
- `AuraQuery`: Union type for all queries
- Query variants for each module:
  - VCRegistry queries (VC status, user VCs, DID resolution, etc.)
  - Compliance queries (KYC, sanctions, GDPR)
  - Auth queries (roles, permissions, multisig)
  - ConfidenceScore queries (scores, IR completions)
  - DataRegistry queries (data items, access checks)
  - EconomicSecurity queries (limits, whale detection)
- Response types for each query

### msg.proto
Message definitions:
- `AuraMsg`: Union type for all messages
- Message variants for each module:
  - VCRegistry messages (disclosure requests, presentations)
  - InclusionRoutines messages (IR completions)
  - ContractRegistry messages (contract registration)
  - Compliance messages (suspicious activity reports)
  - Monitoring messages (contract events)
- Response types for each message

### types.proto
Common type definitions:
- Shared structures used across queries and messages
- Error code definitions
- Pagination types
- Filter types

## Message Flow

```
Contract → AuraQuery/AuraMsg → wasmd → Custom Plugin → AURA Keeper → Response
```

## Usage

### In Go (Backend)

```go
import pb "github.com/aequitas/aura/proto/aura/aura-bindings/v1beta1"

// Implement query handler
func (k Keeper) HandleAuraQuery(ctx sdk.Context, query pb.AuraQuery) (pb.AuraQueryResponse, error) {
    switch q := query.Query.(type) {
    case *pb.AuraQuery_QueryUserVcs:
        return k.handleQueryUserVCs(ctx, q.QueryUserVcs)
    // ... other cases
    }
}
```

### In Rust (Contracts)

```rust
use aura_bindings::{AuraQuery, AuraQueryResponse};

// Query from contract
let query = AuraQuery::QueryUserVCs {
    address: user.to_string(),
    status_filter: Some("active".to_string()),
    type_filter: None,
};
let response: AuraQueryResponse = deps.querier.query(&query.into())?;
```

## Schema Organization

### Query Categories

1. **VCRegistry Queries** (7 query types)
   - Credential verification and management
   - DID resolution
   - Disclosure management

2. **Compliance Queries** (5 query types)
   - KYC/AML verification
   - Sanctions screening
   - GDPR compliance

3. **Auth Queries** (5 query types)
   - Role-based access control
   - Permission checking
   - Multisig and session management

4. **ConfidenceScore Queries** (5 query types)
   - Score calculation and history
   - IR completion tracking
   - Arena scores

5. **DataRegistry Queries** (4 query types)
   - Data item management
   - Access control verification
   - Verification records

6. **EconomicSecurity Queries** (5 query types)
   - Spending limits
   - Whale detection
   - MEV rewards
   - Vesting schedules

### Message Categories

1. **VCRegistry Messages** (3 message types)
   - Disclosure requests
   - Presentation verification
   - Presentation creation

2. **InclusionRoutines Messages** (1 message type)
   - IR completion recording

3. **ContractRegistry Messages** (1 message type)
   - Contract registration

4. **Compliance Messages** (1 message type)
   - Suspicious activity reporting

5. **Monitoring Messages** (1 message type)
   - Contract event reporting

## Code Generation

### Generate Go Code
```bash
cd proto
buf generate
```

### Generate Rust Code
```bash
cd contracts/packages/aura-bindings
cargo build
```

### Generate Documentation
```bash
cd proto
buf generate --template buf.gen.yaml
```

## Versioning

The proto definitions follow semantic versioning:
- **v1beta1**: Initial beta version
- Future versions will be v1, v2, etc.

Breaking changes require new version directories.

## Development Status

**Phase**: Foundation (Phase 1 - Structure Created)
**Status**: To be implemented in Phase 3

## Implementation Timeline

- Phase 3 (Weeks 3-4): Define all proto messages
- Phase 3 (Week 4): Generate Go and Rust code
- Phase 3 (Week 5): Implement query handlers
- Phase 3 (Week 6): Implement message handlers

## Testing

Proto definitions will be tested through:
- Schema validation
- Round-trip serialization tests
- Integration tests with real contracts
- Fuzz testing with random payloads

## Documentation

- [Custom Bindings Guide](../../../../../docs/developers/smart-contracts/custom-bindings.md)
- [Query Reference](../../../../../docs/developers/smart-contracts/query-reference.md)
- [Message Reference](../../../../../docs/developers/smart-contracts/message-reference.md)

## License

Copyright (c) 2025 AURA Blockchain
