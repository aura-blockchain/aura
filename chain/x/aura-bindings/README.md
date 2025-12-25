# AURA Custom Bindings Module

## Overview

This directory contains custom CosmWasm bindings that enable smart contracts to interact with AURA's native blockchain modules. These bindings provide a secure, controlled interface for contracts to query and execute operations across the AURA ecosystem.

## Purpose

The aura-bindings module bridges smart contracts with AURA's unique features:

- **Query Plugin**: Allows contracts to query AURA module state
- **Message Plugin**: Allows contracts to execute AURA module operations
- **Type Definitions**: Defines custom query and message types
- **Security Middleware**: Applies security controls to all binding operations

## Architecture

The bindings follow CosmWasm's custom binding architecture:

```
Smart Contract → AuraQuery/AuraMsg → Query/Message Plugin → AURA Keeper → Response
```

### Query Plugin

Enables read-only queries to AURA modules:

**VCRegistry Queries**:
- `QueryVCStatus`: Get status of a verifiable credential
- `QueryUserVCs`: List all VCs for a user (with filters)
- `QueryResolveDID`: Resolve DID to DID document and active VCs
- `QueryValidateMintEligibility`: Check if user can mint specific VC type
- `QueryCheckRevocation`: Verify VC revocation status with proof
- `QueryGetDisclosurePolicy`: Get user's attribute disclosure policy
- `QueryListPendingDisclosures`: List pending disclosure requests

**Compliance Queries**:
- `QueryKYCStatus`: Get user's KYC verification level
- `QuerySanctionsCheck`: Check if user is on sanctions lists
- `QueryComplianceVerify`: Verify user meets compliance requirements
- `QueryTransactionMonitoring`: Get transaction monitoring alerts
- `QueryGDPRStatus`: Get user's data privacy consent status

**Auth Queries**:
- `QueryHasRole`: Check if address has specific role
- `QueryCheckPermission`: Check if address has permission
- `QueryGetRoleAssignments`: List all role assignments for address
- `QueryMultisigWallet`: Get multisig wallet details
- `QuerySessionStatus`: Get session authentication status

**ConfidenceScore Queries**:
- `QueryUserScore`: Get user's confidence score and verification status
- `QueryHasCompletedIR`: Check if user completed inclusion routine
- `QueryArenaScore`: Get user's score in specific arena
- `QueryAnchorInfo`: Get anchor IR details for user
- `QueryScoreHistory`: Get historical score data

**DataRegistry Queries**:
- `QueryGetDataItem`: Retrieve data item by ID
- `QueryCheckDataAccess`: Verify access permissions for data
- `QueryListUserDataItems`: List data items owned by user
- `QueryDataItemVerifications`: Get verification records for data

**EconomicSecurity Queries**:
- `QueryCheckSpendingLimit`: Verify transaction against spending limits
- `QueryIsWhaleTransaction`: Check if transaction exceeds whale threshold
- `QueryVestingSchedule`: Get vesting schedule details
- `QueryMEVBalance`: Get MEV rewards balance
- `QueryDynamicFeeMultiplier`: Get current dynamic fee multiplier

### Message Plugin

Enables contracts to execute operations on AURA modules:

**VCRegistry Messages**:
- `MsgRequestDisclosure`: Request selective attribute disclosure from holder
- `MsgVerifyPresentation`: Verify a credential presentation
- `MsgCreatePresentation`: Create presentation from VCs

**InclusionRoutines Messages**:
- `MsgRecordIRCompletion`: Record completion of inclusion routine

**ContractRegistry Messages**:
- `MsgRegisterContract`: Register contract in registry with metadata

**Compliance Messages**:
- `MsgReportSuspiciousActivity`: Report suspicious activity for investigation

**Monitoring Messages**:
- `MsgReportContractEvent`: Report contract event for monitoring

## Security Features

- **Authorization Checks**: All messages verified against contract registry
- **Rate Limiting**: Query rate limits prevent abuse (1000 queries/block)
- **Input Validation**: All parameters validated before processing
- **Reentrancy Protection**: Guards against recursive binding calls
- **Telemetry**: All binding operations logged and metered

## Usage in Smart Contracts

### Rust Contract Example

```rust
use aura_bindings::{AuraQuery, AuraMsg, AuraQueryResponse};

// Query user's VCs
let query = AuraQuery::QueryUserVCs {
    address: user_address.to_string(),
    status_filter: Some("active".to_string()),
    type_filter: None,
};
let response: AuraQueryResponse = deps.querier.query(&query.into())?;

// Request attribute disclosure
let msg = AuraMsg::MsgRequestDisclosure {
    holder: holder_address.to_string(),
    verifier: info.sender.to_string(),
    attributes: vec!["email".to_string(), "age".to_string()],
    purpose: "KYC verification".to_string(),
};
```

## Development Status

**Phase**: Foundation (Phase 1 - Structure Created)
**Status**: In Development

## Implementation Plan

### Phase 3: Custom Bindings Module (In Progress)
- [x] Create module structure
- [ ] Define proto messages (queries and messages)
- [ ] Implement query plugin (all query types)
- [ ] Implement message plugin (all message types)
- [ ] Wire plugins to wasmd
- [ ] Add security middleware
- [ ] Add comprehensive tests
- [ ] Achieve 100% code coverage

## Directory Structure

```
aura-bindings/
├── README.md                    # This file
├── keeper/
│   ├── keeper.go               # Main keeper with dependencies
│   ├── query_plugin.go         # Query plugin implementation
│   ├── msg_plugin.go           # Message plugin implementation
│   └── security_middleware.go  # Security controls
├── types/
│   ├── query.go                # Query type definitions
│   ├── msg.go                  # Message type definitions
│   └── errors.go               # Custom errors
└── module.go                   # Module definition
```

## Testing Strategy

- **Unit Tests**: Test each query/message type independently
- **Integration Tests**: Test cross-module interactions
- **Security Tests**: Verify all security controls
- **Performance Tests**: Benchmark query/message latency
- **Fuzz Tests**: Random input testing for robustness

## Performance Considerations

- **Query Caching**: Frequently accessed data cached with short TTL
- **Batch Queries**: Support batching multiple queries in single call
- **Gas Optimization**: Efficient gas accounting for all operations
- **Rate Limiting**: Prevent query/message spam

## Documentation

- [Custom Bindings API Reference](../../../docs/developers/smart-contracts/custom-bindings.md)
- [Query Types Reference](../../../docs/developers/smart-contracts/query-reference.md)
- [Message Types Reference](../../../docs/developers/smart-contracts/message-reference.md)
- [Security Best Practices](../../../docs/developers/smart-contracts/security.md)

## Events

### EventCustomQuery
Emitted when custom query is executed.

**Attributes**: `query_type`, `contract_address`

### EventCustomMessage
Emitted when custom message is processed.

**Attributes**: `message_type`, `sender`, `contract_address`

### EventRateLimitHit
Emitted when rate limit is exceeded.

**Attributes**: `contract_address`, `limit_type`

### EventQueryStats
Emitted for query statistics.

**Attributes**: `query_count`, `avg_gas_used`

### EventMessageStats
Emitted for message execution statistics.

**Attributes**: `message_count`, `total_gas_used`

## Version Compatibility

- **CosmWasm**: 1.5+
- **wasmd**: v0.50.0+
- **AURA Chain**: Latest

## License

Copyright (c) 2025 AURA Blockchain
