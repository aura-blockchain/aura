# AURA Smart Contracts

## Overview

This directory contains reference smart contracts and templates for the AURA blockchain. These contracts demonstrate how to use AURA's custom bindings to integrate with verifiable credentials, compliance systems, and other native blockchain features.

## Purpose

This directory serves multiple purposes:

- **Reference Implementations**: Production-quality example contracts
- **Templates**: Starting points for new contract development
- **Testing**: Contracts used for integration testing
- **Documentation**: Living examples of AURA features
- **Best Practices**: Demonstrate security and design patterns

## Directory Structure

```
contracts/
├── README.md                           # This file
├── packages/
│   └── aura-bindings/                 # Rust bindings for AURA custom queries/messages
│       ├── src/
│       │   ├── lib.rs                 # Main library
│       │   ├── query.rs               # AuraQuery definitions
│       │   ├── msg.rs                 # AuraMsg definitions
│       │   └── response.rs            # Response types
│       ├── Cargo.toml
│       └── README.md
├── vc-gated-dao/                      # VC-Gated DAO template
│   ├── src/
│   ├── examples/
│   ├── schema/
│   ├── Cargo.toml
│   └── README.md
├── compliance-dex/                    # Compliance-Checked DEX template
│   ├── src/
│   ├── examples/
│   ├── schema/
│   ├── Cargo.toml
│   └── README.md
├── credential-marketplace/            # Credential Marketplace template
│   ├── src/
│   ├── examples/
│   ├── schema/
│   ├── Cargo.toml
│   └── README.md
├── identity-lending/                  # Identity-Based Lending template
│   ├── src/
│   ├── examples/
│   ├── schema/
│   ├── Cargo.toml
│   └── README.md
└── Cargo.toml                         # Workspace manifest
```

## Contract Templates

Active development is tracked in [`PORTFOLIO_PLAN.md`](./PORTFOLIO_PLAN.md). It enumerates the two flagship reference contracts we are building next:

1. **VC Issuer (`vc-issuer`)** – on-chain issuance authority that mints Aura verifiable credentials.
2. **Disclosure Verifier (`vc-disclosure-verifier`)** – manages selective-disclosure workflows and grants access tokens after proof validation.

Each contract design in the plan includes the state model, entry points, Aura keeper interactions, and testing/CI requirements. The descriptions below remain as long-term inspiration; the portfolio file reflects the authoritative roadmap.

### Current Contracts

- `binding-tester/` – smoke-test contract used by `chain/x/aura-bindings`.
- `vc-issuer/` – production-grade issuer contract under active development (see folder README for implementation status).

### 1. VC-Gated DAO

A decentralized autonomous organization that requires verified credentials for membership.

**Features**:
- Membership gated by specific VC types
- Minimum confidence score requirement
- Weighted voting based on credentials
- Proposal creation and voting
- Automatic execution of passed proposals

**Use Cases**:
- Professional associations
- Alumni networks
- Verified communities
- Governance bodies

**AURA Integrations**:
- `QueryUserVCs`: Verify member credentials
- `QueryUserScore`: Check confidence score
- `QueryCheckRevocation`: Verify credentials not revoked

### 2. Compliance-Checked DEX

A decentralized exchange with built-in KYC/AML compliance checks.

**Features**:
- KYC verification before trading
- Sanctions screening for all users
- Large transaction reporting
- Spending limit enforcement
- Whale protection mechanisms
- Liquidity pools with access control

**Use Cases**:
- Compliant token swaps
- Regulated asset trading
- Institutional DeFi
- Cross-border payments

**AURA Integrations**:
- `QueryKYCStatus`: Verify user KYC
- `QuerySanctionsCheck`: Screen for sanctions
- `QueryCheckSpendingLimit`: Enforce limits
- `QueryIsWhaleTransaction`: Detect large trades
- `MsgReportSuspiciousActivity`: Report violations

### 3. Credential Marketplace

A marketplace for buying and selling access to verified credentials.

**Features**:
- Sellers must have verified credentials
- Buyers must be KYC verified
- Escrow system for payments
- Selective attribute disclosure
- Reputation system for sellers
- Dispute resolution

**Use Cases**:
- Credential verification services
- Background check marketplace
- Professional certification validation
- Identity verification as a service

**AURA Integrations**:
- `QueryUserVCs`: Verify seller credentials
- `QueryKYCStatus`: Verify buyer compliance
- `MsgRequestDisclosure`: Request attribute disclosure
- `MsgVerifyPresentation`: Verify credential presentations

### 4. Identity-Based Lending

A lending protocol with dynamic interest rates based on user identity and reputation.

**Features**:
- Collateral ratios based on confidence score
- Interest rates based on VC types
- KYC requirements for loan amounts
- Reputation-based credit scoring
- Automated liquidations
- Default protection

**Use Cases**:
- Undercollateralized lending
- Credit scoring
- Risk-adjusted lending
- Identity-backed loans

**AURA Integrations**:
- `QueryUserScore`: Get confidence score for rates
- `QueryUserVCs`: Check for professional credentials
- `QueryKYCStatus`: Verify compliance for loan sizes
- `QueryHasCompletedIR`: Check reputation building activities

## AURA Bindings Package

The `packages/aura-bindings/` directory contains Rust types and interfaces for interacting with AURA's custom bindings.

### Installation

Add to your contract's `Cargo.toml`:

```toml
[dependencies]
aura-bindings = { path = "../packages/aura-bindings" }
```

### Usage

```rust
use aura_bindings::{AuraQuery, AuraMsg, AuraQueryResponse};
use cosmwasm_std::{Deps, DepsMut, Response};

// Query user's VCs
pub fn check_credentials(deps: Deps, user: String) -> Result<bool, ContractError> {
    let query = AuraQuery::QueryUserVCs {
        address: user,
        status_filter: Some("active".to_string()),
        type_filter: Some("kyc_basic".to_string()),
    };

    let response: AuraQueryResponse = deps.querier.query(&query.into())?;

    match response {
        AuraQueryResponse::UserVCs { credentials } => {
            Ok(!credentials.is_empty())
        },
        _ => Err(ContractError::UnexpectedResponse {}),
    }
}

// Request disclosure
pub fn request_disclosure(
    deps: DepsMut,
    holder: String,
    attributes: Vec<String>,
) -> Result<Response, ContractError> {
    let msg = AuraMsg::MsgRequestDisclosure {
        holder,
        verifier: deps.api.addr_canonicalize(&env.contract.address)?.to_string(),
        attributes,
        purpose: "Contract verification".to_string(),
    };

    Ok(Response::new()
        .add_message(msg)
        .add_attribute("action", "request_disclosure"))
}
```

## Development Setup

### Prerequisites

1. **Rust Toolchain** (1.75.0+)
   ```bash
   curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
   rustup target add wasm32-unknown-unknown
   ```

2. **Cargo Extensions**
   ```bash
   cargo install cargo-generate cargo-run-script
   ```

3. **Docker** (for optimization)
   ```bash
   # Install Docker from https://www.docker.com/
   ```

4. **AURA CLI**
   ```bash
   cd ../chain && make install
   ```

### Building Contracts

```bash
# Build all contracts
make build-wasm

# Build specific contract
cd vc-gated-dao && cargo wasm

# Optimize for deployment
make optimize-wasm
```

### Testing Contracts

```bash
# Test all contracts
make test-contracts

# Test specific contract
cd vc-gated-dao && cargo test

# Test with coverage
cd vc-gated-dao && cargo tarpaulin
```

### Generating Schemas

```bash
# Generate schemas for all contracts
make schema-contracts

# Generate for specific contract
cd vc-gated-dao && cargo schema
```

## Deployment

### To Testnet

```bash
# 1. Optimize contracts
make optimize-wasm

# 2. Upload contract code
aurad tx wasm store artifacts/vc_gated_dao.wasm \
  --from my-key \
  --gas auto \
  --gas-adjustment 1.3

# 3. Instantiate contract
aurad tx wasm instantiate <code-id> \
  '{"config": {...}}' \
  --from my-key \
  --label "My VC DAO" \
  --admin <admin-address>

# 4. Register with contract registry
aurad tx contractregistry register-contract <contract-address> \
  --metadata '{"name": "My DAO", ...}' \
  --from my-key
```

### To Mainnet

Mainnet deployment requires:
1. External security audit
2. Governance proposal approval
3. Community review period
4. Formal contract registration

See [Deployment Guide](../docs/operators/smart-contracts/deployment.md)

## Security Considerations

### Best Practices

1. **Always validate inputs**: Use AURA's input validation utilities
2. **Implement reentrancy guards**: Use AURA's security middleware
3. **Check credentials**: Verify VCs before sensitive operations
4. **Enforce compliance**: Always check KYC/sanctions for financial operations
5. **Rate limit**: Use contract registry rate limiting
6. **Audit logging**: Emit events for all important operations
7. **Emergency pause**: Implement pause functionality
8. **Access control**: Use AURA's role-based access control

### Security Checklist

- [ ] External security audit completed
- [ ] All inputs validated
- [ ] Reentrancy protection implemented
- [ ] Access control properly configured
- [ ] Compliance checks enforced
- [ ] Rate limiting configured
- [ ] Emergency pause mechanism
- [ ] Comprehensive tests (95%+ coverage)
- [ ] Fuzz testing performed
- [ ] Gas optimization completed
- [ ] Documentation complete

## Testing Strategy

### Unit Tests
- Test all contract functions
- Test error cases
- Test edge cases
- Mock AURA bindings

### Integration Tests
- Test with actual AURA chain
- Test cross-module interactions
- Test compliance enforcement
- Test security controls

### Security Tests
- Test reentrancy attacks
- Test authorization bypass
- Test input validation
- Test DoS scenarios

## Documentation

- [Contract Development Guide](../docs/developers/smart-contracts/quickstart.md)
- [Custom Bindings Reference](../docs/developers/smart-contracts/custom-bindings.md)
- [Security Best Practices](../docs/developers/smart-contracts/security.md)
- [Testing Guide](../docs/developers/smart-contracts/testing.md)

## Development Status

**Phase**: Foundation (Phase 1 - Structure Created)
**Status**: In Development

## Implementation Timeline

- **Phase 1** (Current): Project structure ✓
- **Phase 2**: Core wasmd integration
- **Phase 3**: Custom bindings implementation
- **Phase 4**: Contract registry module
- **Phase 5**: Contract templates (VC DAO, Compliance DEX, etc.)
- **Phase 6**: Testing infrastructure
- **Phase 7**: Documentation
- **Phase 8**: Security hardening
- **Phase 9**: Testnet deployment
- **Phase 10**: Mainnet preparation

## Contributing

When adding new contract templates:

1. Follow the existing structure
2. Include comprehensive README
3. Add unit tests (95%+ coverage)
4. Add integration tests
5. Generate schemas
6. Document AURA integrations
7. Add security considerations
8. Include deployment examples

## License

Copyright (c) 2025 AURA Blockchain

All reference contracts are provided as examples and templates.
Use at your own risk. Always conduct security audits before mainnet deployment.
