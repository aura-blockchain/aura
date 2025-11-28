# AURA Contract Registry Module

## Overview

The Contract Registry module provides a comprehensive registration, validation, and compliance system for smart contracts deployed on the AURA blockchain. All contracts must register with this module to interact with AURA's custom bindings and native modules.

## Purpose

This module serves as the central registry and security checkpoint for all smart contracts:

- **Contract Registration**: Store metadata, security policies, and compliance requirements
- **Execution Validation**: Verify contracts and users meet requirements before execution
- **Compliance Enforcement**: Enforce KYC/AML, VC, and confidence score requirements
- **Rate Limiting**: Prevent abuse through per-contract and per-user rate limits
- **Lifecycle Management**: Manage contract status (active, paused, deprecated, frozen)
- **Metrics Collection**: Track contract usage, failures, and performance

## Architecture

The Contract Registry sits between the WASM module and AURA's native modules:

```
Contract Execution → Contract Registry Validation → WASM Module → Contract Code
                  ↓
            Check Registration
            Check Compliance
            Check VC Requirements
            Check Rate Limits
            Check Security Policy
```

## Key Components

### Contract Info

Each registered contract has comprehensive metadata:

```go
type ContractInfo struct {
    Address           string                    // Contract address
    Creator           string                    // Creator address
    Admin             string                    // Admin address (can update)
    CodeID            uint64                    // WASM code ID
    Label             string                    // Human-readable label
    Metadata          ContractMetadata          // Additional metadata
    SecurityPolicy    SecurityPolicy            // Security configuration
    Compliance        ComplianceRequirements    // Compliance rules
    Status            ContractStatus            // Current status
    CreatedAt         time.Time                 // Creation timestamp
    UpdatedAt         time.Time                 // Last update timestamp
    ExecutionCount    uint64                    // Total executions
    LastExecutionTime time.Time                 // Last execution
}
```

### Contract Metadata

Identity and categorization information:

```go
type ContractMetadata struct {
    Name              string      // Contract name
    Description       string      // Description
    Version           string      // Contract version
    Author            string      // Author/organization
    License           string      // License (MIT, Apache, etc.)
    Repository        string      // Source code repository URL
    Documentation     string      // Documentation URL
    Tags              []string    // Searchable tags
    Category          string      // Contract category
    RequiredVCTypes   []string    // Required VC types for users
    MinConfidenceScore uint64     // Minimum confidence score required
}
```

### Security Policy

Security configuration and limits:

```go
type SecurityPolicy struct {
    MaxGasPerExecution     uint64              // Max gas per execution
    MaxGasPerBlock         uint64              // Max cumulative gas per block
    RateLimitPerUser       RateLimit           // Per-user rate limit
    RateLimitPerContract   RateLimit           // Per-contract rate limit
    AllowedCallers         []string            // Whitelist (empty = all)
    BlockedCallers         []string            // Blacklist
    RequireAuthentication  bool                // Require session auth
    RequireMultisig        bool                // Require multisig approval
    PauseEnabled           bool                // Can be paused
    DeprecationWarning     bool                // Show deprecation warning
}
```

### Compliance Requirements

Compliance rules for contract interaction:

```go
type ComplianceRequirements struct {
    RequireKYC             bool                 // KYC required
    MinKYCLevel            string               // Minimum KYC level
    RequireSanctionsCheck  bool                 // Screen against sanctions
    AllowedJurisdictions   []string             // Allowed jurisdictions
    BlockedJurisdictions   []string             // Blocked jurisdictions
    RequireVCs             []string             // Required VC types
    MinConfidenceScore     uint64               // Minimum CS
    RequireGDPRConsent     bool                 // GDPR consent required
    DataProcessingPurpose  string               // Data processing purpose
}
```

### Contract Status

Lifecycle states:

- **Active**: Normal operation, all features enabled
- **Paused**: Temporarily disabled (admin or emergency)
- **Deprecated**: Still functional but marked for replacement
- **Frozen**: Permanently disabled (governance or security)

## Core Operations

### Registration

```go
// Register a new contract
func (k Keeper) RegisterContract(
    ctx sdk.Context,
    info ContractInfo,
) error
```

Requirements:
- Valid contract address
- Complete metadata
- Creator must meet KYC requirements (if configured)
- Security policy within acceptable ranges
- Admin signature

### Validation Before Execution

```go
// Validate contract execution
func (k Keeper) ValidateContractExecution(
    ctx sdk.Context,
    contractAddr string,
    sender string,
    msg []byte,
) error
```

Checks performed:
1. Contract is registered
2. Contract status is Active
3. User meets KYC requirements
4. User passes sanctions screening
5. User has required VCs
6. User meets minimum confidence score
7. Rate limits not exceeded
8. User not blacklisted
9. Gas limits not exceeded

### Rate Limiting

Per-user and per-contract rate limiting with time windows:

```go
type RateLimit struct {
    MaxRequests    uint64        // Max requests
    TimeWindow     time.Duration // Time window (hourly, daily)
    CurrentCount   uint64        // Current count
    WindowStart    time.Time     // Window start time
}
```

### Metrics Collection

Tracked metrics:
- Total contracts registered
- Contracts by status
- Executions per contract
- Failures per contract (with reasons)
- Gas usage per contract
- Rate limit violations
- Average execution time

## API

### Msg Server

**MsgRegisterContract**: Register new contract
**MsgUpdateContractMetadata**: Update metadata (admin only)
**MsgUpdateSecurityPolicy**: Update security policy (admin only)
**MsgPauseContract**: Pause contract (admin or governance)
**MsgUnpauseContract**: Unpause contract (admin or governance)
**MsgDeprecateContract**: Mark as deprecated (admin or governance)

### Query Server

**QueryContractInfo**: Get contract details
**QueryContractsByCreator**: List contracts by creator
**QueryContractsByTag**: List contracts by tag
**QueryRegisteredContracts**: List all contracts (paginated)
**QueryContractMetrics**: Get contract metrics

## Security Features

- **Authorization**: Only admin or governance can modify contracts
- **Input Validation**: All parameters validated
- **Compliance Caching**: Cache compliance results (1 block TTL)
- **Rate Limit Cleanup**: Automatic cleanup of old counters
- **Audit Logging**: All operations logged for security audits

## Integration Points

### With WASM Module
- Validates all contract executions
- Enforces security policies
- Tracks execution metrics

### With VCRegistry
- Verifies user has required credentials
- Checks credential validity and revocation

### With Compliance
- Verifies KYC/AML status
- Screens against sanctions lists
- Enforces jurisdiction restrictions

### With Auth
- Checks user roles and permissions
- Validates multisig requirements
- Verifies session authentication

### With ConfidenceScore
- Checks user confidence score
- Enforces minimum score requirements

## Development Status

**Phase**: Foundation (Phase 1 - Structure Created)
**Status**: In Development

## Implementation Plan

### Phase 4: Contract Registry Module (In Progress)
- [x] Create module structure
- [ ] Define proto messages
- [ ] Implement keeper with all dependencies
- [ ] Implement core registry operations
- [ ] Implement validation logic
- [ ] Implement rate limiting
- [ ] Implement compliance enforcement
- [ ] Implement metrics collection
- [ ] Implement Msg/Query servers
- [ ] Implement module and genesis
- [ ] Wire into app
- [ ] Add comprehensive tests (100% coverage)
- [ ] Add integration tests

## Directory Structure

```
contractregistry/
├── README.md                       # This file
├── keeper/
│   ├── keeper.go                  # Main keeper
│   ├── registry.go                # Registry operations
│   ├── validation.go              # Validation logic
│   ├── rate_limiting.go           # Rate limit tracking
│   ├── compliance.go              # Compliance enforcement
│   ├── metrics.go                 # Metrics collection
│   ├── msg_server.go              # Message handlers
│   └── query_server.go            # Query handlers
├── types/
│   ├── contract_info.go           # Contract info types
│   ├── security_policy.go         # Security types
│   ├── compliance.go              # Compliance types
│   ├── genesis.go                 # Genesis state
│   ├── msgs.go                    # Message types
│   ├── queries.go                 # Query types
│   └── errors.go                  # Custom errors
└── module.go                      # Module definition
```

## Usage Examples

### Register a Contract

```go
msg := &types.MsgRegisterContract{
    Sender: creator,
    ContractAddress: contractAddr,
    Metadata: types.ContractMetadata{
        Name: "My VC-Gated DAO",
        Description: "DAO requiring verified credentials",
        RequiredVCTypes: []string{"kyc_basic", "government_id"},
        MinConfidenceScore: 50,
    },
    SecurityPolicy: types.SecurityPolicy{
        MaxGasPerExecution: 1000000,
        RateLimitPerUser: types.RateLimit{
            MaxRequests: 100,
            TimeWindow: time.Hour,
        },
    },
    Compliance: types.ComplianceRequirements{
        RequireKYC: true,
        MinKYCLevel: "basic",
        RequireSanctionsCheck: true,
    },
}
```

### Query Contract Info

```bash
aurad query contractregistry contract <contract-address>
```

## Testing Strategy

- **Unit Tests**: Test all keeper methods
- **Integration Tests**: Test cross-module interactions
- **Security Tests**: Verify all validation checks
- **Performance Tests**: Test rate limiting under load
- **Compliance Tests**: Verify enforcement of all compliance rules

## Documentation

- [Contract Registry Guide](../../../docs/operators/smart-contracts/contract-registry.md)
- [Registration Requirements](../../../docs/developers/smart-contracts/registration.md)
- [Compliance Guide](../../../docs/developers/smart-contracts/compliance.md)

## License

Copyright (c) 2025 AURA Blockchain
