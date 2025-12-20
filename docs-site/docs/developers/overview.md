---
sidebar_position: 1
---

# Developer Overview

Welcome to the Aura developer documentation. This guide will help you understand how to build applications and services on the Aura blockchain.

## Architecture

Aura is built on the Cosmos SDK, providing a modular architecture for blockchain application development. The chain consists of 27 custom modules organized into several key categories:

### Core Modules

- **Identity Module** - W3C Verifiable Credentials issuance and management
- **VC Registry** - Credential schema registry and revocation lists
- **Confidence Score** - Privacy-preserving identity confidence calculation
- **AI Assistant Network** - Decentralized oracle network for off-chain verification

### DeFi Modules

- **DEX** - Decentralized exchange with automated market makers
- **Bridge** - Cross-chain asset transfers via IBC
- **Staking** - Proof-of-Stake validator delegation
- **Governance** - On-chain voting with zero-knowledge proofs

### Compliance Modules

- **KYC** - Know Your Customer verification without storing PII
- **AML** - Anti-Money Laundering monitoring
- **Sanctions** - Regulatory compliance checks

## Development Stack

### Required Tools

- **Go 1.21+** - Primary development language
- **Cosmos SDK v0.50+** - Application framework
- **Tendermint** - BFT consensus engine
- **Protocol Buffers** - Data serialization
- **gRPC** - API communication

### Recommended Tools

- **Ignite CLI** - Blockchain scaffolding
- **Buf** - Protocol buffer management
- **CosmWasm** - Smart contract development
- **Hermes** - IBC relayer for testing

## Getting Started

### 1. Set Up Development Environment

```bash
# Clone repository
git clone https://github.com/aura-blockchain/aura.git
cd aura

# Install dependencies
make install

# Run tests
make test
```

### 2. Understand the Module System

Aura uses the Cosmos SDK module pattern:

```
chain/
├── x/                      # Custom modules
│   ├── identity/          # Identity verification
│   ├── vcregistry/        # VC management
│   ├── dex/               # Decentralized exchange
│   └── ...                # 24 more modules
├── proto/                 # Protocol buffer definitions
├── app/                   # Application setup
└── cmd/                   # CLI commands
```

### 3. Explore the Modules

Each module provides:
- **Keeper** - State management and business logic
- **Messages** - Transaction types
- **Queries** - Read-only state access
- **Events** - Transaction notifications
- **CLI Commands** - User interface

## Building Applications

### Using the REST API

Aura provides a REST API for application integration:

```javascript
// Query account balance
const response = await fetch(
  'https://api.aura.network/cosmos/bank/v1beta1/balances/aura1...'
);
const data = await response.json();
```

### Using gRPC

For high-performance applications:

```go
import (
    "google.golang.org/grpc"
    banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

conn, _ := grpc.Dial("localhost:9090", grpc.WithInsecure())
client := banktypes.NewQueryClient(conn)

res, _ := client.Balance(context.Background(), &banktypes.QueryBalanceRequest{
    Address: "aura1...",
    Denom:   "uaura",
})
```

### Using JavaScript SDK

```bash
npm install @aura/sdk
```

```javascript
import { AuraClient } from '@aura/sdk';

const client = new AuraClient('https://rpc.aura.network');
await client.connect();

// Query balance
const balance = await client.getBalance('aura1...');

// Send transaction
const tx = await client.sendTokens(
  'aura1from...',
  'aura1to...',
  [{ denom: 'uaura', amount: '1000000' }]
);
```

## Common Development Patterns

### Working with Verifiable Credentials

```go
// Issue a credential
msg := vctypes.NewMsgIssueVC(
    issuerDID,
    subjectDID,
    credentialType,
    claims,
)

// Verify a credential
credential, err := keeper.GetCredential(ctx, vcID)
if err != nil {
    return err
}

isValid := keeper.VerifyCredential(ctx, credential)
```

### Cross-Chain Transfers

```go
// Initiate IBC transfer
msg := ibctransfertypes.NewMsgTransfer(
    sourcePort,
    sourceChannel,
    token,
    sender,
    receiver,
    timeoutHeight,
    timeoutTimestamp,
)
```

### Governance Proposals

```go
// Submit proposal
content := govtypes.ContentFromProposalType(
    "Update Parameters",
    "Description",
    govtypes.ProposalTypeText,
)

msg := govtypes.NewMsgSubmitProposal(
    content,
    deposit,
    proposer,
)
```

## Testing

### Unit Tests

```bash
# Run all tests
make test

# Run specific module tests
cd x/identity && go test -v

# Run with coverage
make test-coverage
```

### Integration Tests

```bash
# Run integration tests
make test-integration

# Run e2e tests
make test-e2e
```

### Local Testnet

```bash
# Start local testnet
make testnet-start

# Run tests against testnet
make testnet-test

# Stop testnet
make testnet-stop
```

## SDK Documentation

### Official SDKs

- **Go SDK** - Native Cosmos SDK integration
- **JavaScript/TypeScript** - [@aura/sdk](https://www.npmjs.com/package/@aura/sdk)
- **Python** - [aura-py](https://pypi.org/project/aura-py)
- **Rust** - [aura-rs](https://crates.io/crates/aura)

### Community SDKs

Check the [SDK Registry](https://github.com/aura-blockchain/awesome-aura) for community-maintained SDKs.

## Resources

### Documentation

- [Module Development Guide](/docs/developers/module-development)
- [SDK Integration Guide](/docs/developers/sdk-integration)
- [API Reference](https://api-docs.aura.network)
- [Protocol Buffer Specs](https://buf.build/aura-blockchain/aura)

### Examples

- [Example DApp](https://github.com/aura-blockchain/example-dapp)
- [Credential Issuer](https://github.com/aura-blockchain/example-issuer)
- [IBC Integration](https://github.com/aura-blockchain/example-ibc)

### Developer Tools

- [Block Explorer](https://explorer.aura.network)
- [Testnet Faucet](https://faucet.aura.network)
- [RPC Endpoints](https://docs.aura.network/rpc-endpoints)
- [Developer Playground](https://play.aura.network)

## Best Practices

### Security

- Always validate user inputs
- Use proper error handling
- Implement access controls
- Follow Cosmos SDK security guidelines
- Audit smart contracts before deployment

### Performance

- Batch transactions when possible
- Use pagination for large queries
- Cache frequently accessed data
- Optimize gas usage
- Monitor transaction costs

### Testing

- Write comprehensive unit tests
- Test edge cases and error conditions
- Use integration tests for multi-module interactions
- Perform load testing before production
- Test on testnet before mainnet

## Getting Help

### Community Support

- [Discord](https://discord.gg/aura) - Developer chat
- [Forum](https://forum.aura.network) - Technical discussions
- [Stack Overflow](https://stackoverflow.com/questions/tagged/aura) - Q&A

### Contributing

- [Contributing Guide](https://github.com/aura-blockchain/aura/blob/main/CONTRIBUTING.md)
- [Code of Conduct](https://github.com/aura-blockchain/aura/blob/main/CODE_OF_CONDUCT.md)
- [Development Roadmap](https://github.com/aura-blockchain/aura/projects)

## Next Steps

Ready to build? Check out:

- [Module Development Guide](/docs/developers/module-development) - Create custom modules
- [SDK Integration Guide](/docs/developers/sdk-integration) - Integrate Aura into your app
- [API Reference](https://api-docs.aura.network) - Complete API documentation
