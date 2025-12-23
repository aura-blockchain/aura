# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- OpenAPI/Swagger spec generation (190 endpoints, 732 definitions)
- Slice pre-allocation in 96 keeper files for performance
- Comprehensive security audit documentation
- Architecture analysis for testnet readiness

### Changed
- Performance optimizations in DEX orderbook cleanup (O(k) vs O(n))
- Compliance module uses expiration index for efficient lookups

### Fixed
- P0 Security: Multisig race condition in identity module
- P0 Security: LP token atomicity in DEX liquidity pool
- P0 Security: Bridge transfer ID collision on genesis import
- P0 GDPR: Cascade deletion for identity erasure

---

## [0.1.0-testnet] - 2025-12-15

### Added
- **Core Chain**: 27 Cosmos SDK modules production-ready
- **Identity Module**: DID management, multisig wallets, role-based access control
- **Privacy Module**: Confidential transactions, ZK proofs, mixing pools
- **Compliance Module**: KYC/AML verification, sanctions screening, GDPR support
- **DEX Module**: Order book trading, liquidity pools, HTLC atomic swaps
- **Bridge Module**: Cross-chain transfers with multi-sig validation
- **VCRegistry Module**: Verifiable Credentials management
- **AI Assistant Module**: On-chain AI integration with rate limiting

### Security
- Zero P1 security vulnerabilities
- Comprehensive invariant checks in all critical modules
- Security audit passed

### Infrastructure
- Kubernetes deployment manifests
- ArgoCD GitOps configuration
- Comprehensive test infrastructure (468 test files)
- Multi-language SDK framework (Go, JavaScript, Python)

---

## [0.1.0-determinism] - 2025-11-20

### Fixed
- Deterministic consensus using LegacyDec instead of float64
- Hardened validation in prevalidation and bridge modules
- Nil-safe decimal operations

### Changed
- Removed CPU limits, increased memory for validators
- Deterministic multipliers and parameter encoding

---

## [0.0.1-alpha] - 2025-10-01

### Added
- Initial Cosmos SDK chain scaffolding
- Basic module structure for identity, privacy, compliance
- CLI daemon (aurad) foundation
- CosmWasm smart contract support
- Pre-commit hooks for code quality

---

## Version Roadmap

| Version | Status | Target |
|---------|--------|--------|
| 0.1.0-testnet | ✅ Ready | December 2025 |
| 0.2.0-mainnet-beta | Planned | Q1 2026 |
| 1.0.0-mainnet | Target | Q2 2026 |
