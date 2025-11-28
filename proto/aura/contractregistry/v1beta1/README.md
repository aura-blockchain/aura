# Contract Registry Protocol Buffers

## Overview

This directory contains Protocol Buffer definitions for the AURA Contract Registry module. These proto files define the message types, query interfaces, and state structures for contract registration and management.

## Purpose

The proto definitions serve as the interface specification for:

- Contract registration messages
- Contract query interfaces
- Contract metadata structures
- Security policy definitions
- Compliance requirement specifications
- Genesis state format

## Files

### contract_registry.proto
Core types and structures:
- `ContractInfo`: Complete contract information
- `ContractMetadata`: Contract metadata and categorization
- `SecurityPolicy`: Security configuration and limits
- `ComplianceRequirements`: Compliance rules
- `ContractStatus`: Lifecycle status enum
- `RateLimit`: Rate limiting configuration

### query.proto
Query service definition:
- `QueryContractInfo`: Get contract details
- `QueryContractsByCreator`: List contracts by creator
- `QueryContractsByTag`: Search by tags
- `QueryRegisteredContracts`: List all with pagination
- `QueryContractMetrics`: Get usage metrics

### tx.proto
Message service definition:
- `MsgRegisterContract`: Register new contract
- `MsgUpdateContractMetadata`: Update metadata
- `MsgUpdateSecurityPolicy`: Update security config
- `MsgPauseContract`: Pause contract
- `MsgUnpauseContract`: Resume contract
- `MsgDeprecateContract`: Mark as deprecated

### genesis.proto
Genesis state definition:
- `GenesisState`: Initial state format
- Parameter defaults
- Contract list initialization

## Usage

### Generating Go Code

```bash
cd proto
buf generate
```

### Generating Rust Code

```bash
cd contracts
cargo schema
```

## Development Status

**Phase**: Foundation (Phase 1 - Structure Created)
**Status**: To be implemented in Phase 4

## License

Copyright (c) 2025 AURA Blockchain
