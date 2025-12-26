# Aequitas / AURA Identity Blockchain

![Build](https://github.com/decristofaroj/aura/workflows/Comprehensive%20CI%2FCD%20Pipeline/badge.svg) ![Coverage](https://codecov.io/gh/decristofaroj/aura/branch/main/graph/badge.svg) ![Quality](https://sonarcloud.io/api/project_badges/measure?project=decristofaroj_aura&metric=alert_status) ![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg) ![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go) ![Discord](https://img.shields.io/badge/Discord-Join%20Us-5865F2?logo=discord&logoColor=white&link=https://discord.gg/aura) ![Twitter](https://img.shields.io/badge/Twitter-@AuraNetwork-1DA1F2?logo=twitter&logoColor=white)

## A Zero-PII Identity Blockchain for Decentralized Trust and W3C Verifiable Credentials

Aequitas is a Layer-1 blockchain built on the Cosmos SDK that serves as a decentralized identity trust anchor. It enables issuance, verification, and management of W3C-compliant verifiable credentials while maintaining zero personally identifiable information (PII) on-chain. The network is secured by proof-of-identity (PoI) rewards and powered by a decentralized AI assistant network.

## Key Features

- **Zero-PII Architecture** - Identity verification without storing personal data on-chain
- **W3C Verifiable Credentials** - Industry-standard credential issuance and verification
- **AI Assistant Network** - Decentralized oracle nodes performing off-chain identity verification
- **Proof-of-Identity Rewards** - Economic incentives for users completing verification routines
- **High-Speed BFT Consensus** - Tendermint-based DPoS with 2-3 second block times
- **IBC Interoperability** - Permissionless token and credential transfer via Inter-Blockchain Communication
- **Governance-Controlled** - Democratic token holder voting through zero-knowledge proofs
- **Adaptive Fraud Detection** - Self-improving ML models preventing identity abuse

## Quick Start (< 5 Minutes)

### Prerequisites

- Go 1.21+ (for chain development)
- PHP 8.1+ (for wallet/helper utilities)
- Docker and Docker Compose (optional, for containerized setup)
- Node.js 18+ (for tooling)
- Python 3.9+ (for economics and analysis tools)

### Installation

```bash
# Clone the repository
git clone https://github.com/decristofaroj/aura.git
cd aura

# Install PHP dependencies (required before git hooks)
composer install

# Install Node dependencies
npm install

# Install Python tools
pip install -r requirements.txt

# Verify setup
composer run test

> **Note:** Run `composer install` once before making commits so the Husky pre-commit hook can execute the PHP checks. If Composer is missing, the hook will skip those checks until Composer is installed.
```

### Start a Local Validator Node

```bash
# Initialize the chain
./scripts/init-chain.sh

# Start the validator
./scripts/start-validator.sh

# Node will run on localhost:26657 (RPC) and localhost:1317 (REST API)
```

### Generate and Manage Wallets

```bash
# Create a new wallet
php wallet/cli.php generate-address

# Check wallet balance
php wallet/cli.php balance --address aura1xxxxx

# Export verifiable credential
php wallet/cli.php export-credential --address aura1xxxxx
```

### Join the AI Assistant Network

```bash
# Register as an AI Assistant node
./scripts/register-assistant.sh --locale en-US --stake 10000aeq

# Your node will begin verifying identity proofs
```

## Configuration

### Network Setup

```bash
# Configure for testnet
export AURA_NETWORK=testnet
export AURA_RPC_PORT=26657
export AURA_REST_PORT=1317

# Configure for mainnet
export AURA_NETWORK=mainnet
```

### Key Configuration Files

- `infra/node-config.yaml` - Validator and network parameters
- `chain/app.go` - Cosmos SDK application configuration
- `docs/config/` - Architecture and deployment specifications

### Environment Variables

```bash
AURA_NETWORK           # testnet or mainnet
AURA_RPC_PORT          # Tendermint RPC port (default: 26657)
AURA_REST_PORT         # REST API port (default: 1317)
AURA_LOG_LEVEL         # Logging level (INFO, DEBUG, ERROR)
AURA_VALIDATOR_KEY     # Path to validator private key
```

## Basic Usage Examples

### Create and Query Verifiable Credentials

```bash
# Issue a credential
curl -X POST http://localhost:1317/identity/issue-credential \
  -H "Content-Type: application/json" \
  -d '{
    "issuer": "aura1assistant...",
    "holder": "aura1user...",
    "credential_type": "isVerifiedHuman",
    "proof": "proof_hash_here"
  }'

# Query credential status
curl http://localhost:1317/identity/credential/CRED_ID

# Check revocation status
curl http://localhost:1317/identity/revocation-list
```

### Participate in Governance

```bash
# Submit a governance proposal
php tools/governance/submit-proposal.php \
  --title "Increase Validator Set" \
  --description "Expand validator set to 150 nodes" \
  --type PARAMETER_CHANGE

# Vote on proposal (as verified holder)
php tools/governance/vote.php \
  --proposal-id 1 \
  --vote YES \
  --using-credential VC:isVerifiedHuman
```

### Manage AI Assistant Nodes

```bash
# List registered assistants by locale
curl http://localhost:1317/assistants?locale=en-US

# Get assistant performance metrics
curl http://localhost:1317/assistants/aura1assistant.../metrics

# Check fraud detection feedback loop status
curl http://localhost:1317/assistants/fraud-detection-status
```

### Economics and Analysis

```bash
# Regenerate verifier fee data
python tools/aggregate_verifier_fees.py \
  --input-dir data/verifier-fee-events \
  --output docs/economics/models/verifier-fee-data.csv

# Build economics scenarios
python tools/build_economics_notebook.py

# Analyze tokenomics
python tools/economics_analysis.py --scenario baseline
```

## Architecture Overview

### System Components

```
Aequitas Blockchain
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

### Module Organization

```
aura/
├── chain/                    # Cosmos SDK application
│   ├── app.go               # App initialization
│   └── modules/
│       ├── identity/        # VC and identity logic
│       ├── assistant/       # AI assistant registry
│       └── governance/      # DAO voting
├── ai-assistant/            # AI oracle nodes
├── wallet/                  # Light client and helpers
│   └── php/                # PHP wallet utilities
├── zkp/                     # Zero-knowledge proofs
├── verifier-portal/         # Verifier UX
├── docs/                    # Documentation, RFCs, economics
│   ├── rfcs/               # Request for comments (0002-0007)
│   └── economics/          # Tokenomics models and scenarios
├── tests/                   # PHPUnit test suite
└── tools/                   # Python economics and analysis tools
```

## Testing

### PHP Tests (Wallet Utilities)

```bash
# Run all PHPUnit tests
composer run phpunit

# Run specific test class
./vendor/bin/phpunit tests/Wallet/BalanceCalculatorTest.php

# Code quality checks
composer run phpcs      # Code style
composer run phpstan    # Static analysis
```

### Go Tests (Chain)

```bash
# Run chain tests
cd chain && go test ./...

# Run with coverage
go test -cover ./...
```

### Python Tests (Tools)

```bash
# Test economics tools
python -m pytest tools/tests/

# Run linting
pylint tools/
mypy tools/
```

## Development Setup

### Prerequisites

- Go 1.21+
- PHP 8.1+ with Composer
- Node.js 18+
- Python 3.9+
- Docker (optional)

### Initial Environment

```bash
# Clone repository
git clone https://github.com/decristofaroj/aura.git
cd aura

# Install all dependencies
composer install
npm install
pip install -r requirements.txt

# Install pre-commit hooks
pre-commit install

# Verify setup
composer run test
npm run lint
```

### Working with Documentation

The project includes comprehensive RFC documentation:

```bash
# Review core RFCs
cat docs/rfcs/0002-identity-verification.md      # Core identity spec
cat docs/rfcs/0003-ai-assistant-network.md       # AI oracle design
cat docs/rfcs/0005-tokenomics.md                 # Economics model
cat docs/rfcs/0006-governance.md                 # DAO structure
cat docs/rfcs/0007-privacy-preservation.md       # Privacy design

# Economics modeling
cat docs/economics/models/economics-scenarios.ipynb
```

## Network Specifications

### Consensus & Performance

| Parameter | Value |
|-----------|-------|
| Consensus | Tendermint BFT-DPoS |
| Block Time | 2-3 seconds |
| Finality | 1 block |
| Initial Validators | 100 |
| Max Validators | 300+ (via governance) |

### Tokenomics (AURA)

| Component | Allocation | Purpose |
|-----------|-----------|---------|
| Protocol Emissions | 40% | Validator & assistant rewards |
| Proof-of-Identity Mining | 20% | User verification rewards |
| Ecosystem Fund | 20% | Development & partnerships |
| Core Team | 10% | Team vesting |
| Foundation | 10% | Operations & grants |

### Inclusion Routines (Identity Verification)

- **IR:SELFIE_01** - Liveness detection with photo ID
- **IR:GEOLOCATION** - Location verification quest
- **IR:SOCIAL_GRAPH** - Social media credential linking
- **IR:GOV_ID_01** - Government ID verification

## Documentation

- **[Technical Specification](docs/Aequitas_AURAcoin_Blockchain.md)** - Complete protocol details
- **[Project Status](PROJECT_STATUS.md)** - Current milestones and progress
- **[RFCs](docs/rfcs/)** - Technical specifications and design documents
- **[Economics Models](docs/economics/models/)** - Tokenomics and financial scenarios
- **[Operations Runbooks](docs/ops/)** - Deployment and operational guides

## Roadmap

### Phase 1: Foundation (Complete)
- Core blockchain with Cosmos SDK
- Identity verification modules
- AI assistant network framework
- W3C credential support

### Phase 2: Integration (In Progress)
- Full AI assistant deployment
- Economics model implementation
- Governance system activation
- Security audits

### Phase 3: Expansion
- IBC bridge deployment
- Mobile wallet launch
- Enterprise integrations
- zkML upgrades

### Phase 4: Scaling
- Horizontal sharding
- Privacy-focused features
- Interoperability enhancements

## Contributing

We welcome contributions to the Aequitas protocol!

**Before contributing:**
1. Read [CONTRIBUTING.md](CONTRIBUTING.md)
2. Review relevant RFCs in `docs/rfcs/`
3. Check [PROJECT_STATUS.md](PROJECT_STATUS.md) for active areas
4. Follow code style and testing requirements

**Contribution areas:**
- Chain module implementation
- AI assistant improvements
- PHP wallet utilities
- Documentation and examples
- Security and performance optimizations

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

Apache License 2.0 - See [LICENSE](LICENSE) file for complete text.

## Contact & Support

- **GitHub Issues**: [Report bugs and feature requests](https://github.com/decristofaroj/aura/issues)
- **Discussions**: [Community discussions](https://github.com/decristofaroj/aura/discussions)
- **Documentation**: [Full docs](docs/README.md)
- **Project Status**: [Current milestones](PROJECT_STATUS.md)

## Security Considerations

- Identity verification is performed off-chain by AI assistants
- No PII is ever stored on-chain
- Zero-knowledge proofs protect voter privacy in governance
- All credentials use cryptographic signatures for authenticity
- See [SECURITY.md](SECURITY.md) for vulnerability reporting

## Disclaimer

Aequitas is an experimental blockchain system under active development. The protocol, especially identity verification mechanisms, should be thoroughly tested before production use. Users should understand the technical and operational risks involved.

## References

- **GitHub Repository**: https://github.com/decristofaroj/aura
- **Cosmos SDK**: https://docs.cosmos.network
- **W3C Verifiable Credentials**: https://www.w3.org/TR/vc-data-model/
- **Inter-Blockchain Communication**: https://ibcprotocol.org/
- **Tendermint Consensus**: https://tendermint.com/

---

**Latest Update**: November 2025 | **Status**: Implementation Phase | **Version**: 1.0 (Specification)
