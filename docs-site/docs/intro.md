---
sidebar_position: 1
---

# What is Aura?

Aura is a Layer-1 blockchain built on the Cosmos SDK that serves as a decentralized identity trust anchor. It enables issuance, verification, and management of W3C-compliant verifiable credentials while maintaining zero personally identifiable information (PII) on-chain.

## Key Features

- **Zero-PII Architecture** - Identity verification without storing personal data on-chain
- **W3C Verifiable Credentials** - Industry-standard credential issuance and verification
- **AI Assistant Network** - Decentralized oracle nodes performing off-chain identity verification
- **Proof-of-Identity Rewards** - Economic incentives for users completing verification routines
- **High-Speed BFT Consensus** - Tendermint-based DPoS with 2-3 second block times
- **IBC Interoperability** - Permissionless token and credential transfer via Inter-Blockchain Communication
- **Governance-Controlled** - Democratic token holder voting through zero-knowledge proofs
- **Adaptive Fraud Detection** - Self-improving ML models preventing identity abuse

## Architecture Overview

```
Aura Blockchain
├── Consensus Layer (Tendermint BFT-DPoS)
│   ├── Byzantine Fault Tolerance
│   ├── Delegated Proof-of-Stake
│   └── 2-3 Second Block Times
├── Identity Module
│   ├── IdentityManager Contract
│   ├── VC Schema Registry
│   ├── Revocation List
│   └── Confidence Score Calculation
├── AI Assistant Network
│   ├── Decentralized Verifiers
│   ├── Locale-Specific Nodes
│   ├── Fraud Detection System
│   └── ML Model Updates
├── Tokenomics (AURA)
│   ├── Validator Rewards
│   ├── AI Assistant Rewards
│   ├── Proof-of-Identity Mining
│   └── Fee Distribution
└── Governance (1-Person, 1-Vote via ZKP)
    ├── Proposal Management
    ├── Voting System
    └── On-Chain Upgrades
```

## Use Cases

### Decentralized Identity

Aura provides a privacy-preserving identity infrastructure where users can prove attributes about themselves without revealing unnecessary personal information. Credentials are cryptographically signed and can be verified on-chain.

### Cross-Chain Identity

Through IBC (Inter-Blockchain Communication), Aura credentials can be used across multiple blockchain networks, enabling portable digital identity across the Cosmos ecosystem and beyond.

### Sybil-Resistant Applications

Applications can leverage Aura's proof-of-identity system to prevent Sybil attacks and ensure one-person-one-vote governance systems.

### Compliance and KYC

Organizations can verify user compliance without collecting or storing sensitive personal data, meeting regulatory requirements while preserving user privacy.

## Network Specifications

| Parameter | Value |
|-----------|-------|
| Consensus | Tendermint BFT-DPoS |
| Block Time | 2-3 seconds |
| Finality | 1 block |
| Initial Validators | 100 |
| Max Validators | 300+ (via governance) |
| Native Token | AURA |

## Getting Started

Ready to dive in? Check out these resources:

- [Installation Guide](/docs/getting-started/installation) - Set up your development environment
- [Quick Start](/docs/getting-started/quick-start) - Run your first node in minutes
- [Developer Overview](/docs/developers/overview) - Start building on Aura

## Community

Join our growing community:

- [Discord](https://discord.gg/aura) - Get real-time support
- [Forum](https://forum.aura.network) - Participate in discussions
- [GitHub](https://github.com/aura-blockchain/aura) - Contribute to development
- [Twitter](https://twitter.com/AuraNetwork) - Stay updated
