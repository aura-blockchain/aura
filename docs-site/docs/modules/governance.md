---
sidebar_position: 6
---

# Governance & Operations Modules

On-chain governance, monitoring, and AI oracle network.

## governance

Enhanced governance with zero-knowledge voting.

### Key Features
- Standard Cosmos governance
- ZK-voting (anonymous votes)
- 1-person-1-vote mode
- Proposal execution

### Messages
| Message | Description |
|---------|-------------|
| `MsgSubmitProposal` | Submit governance proposal |
| `MsgVote` | Cast standard vote |
| `MsgVoteZK` | Cast anonymous ZK vote |
| `MsgDeposit` | Deposit to proposal |

### Queries
```bash
aurad query gov proposals
aurad query gov proposal <proposal-id>
aurad query gov votes <proposal-id>
aurad query gov tally <proposal-id>
```

### Proposal Types
| Type | Description |
|------|-------------|
| Text | Community signaling |
| ParameterChange | Modify chain parameters |
| SoftwareUpgrade | Coordinate upgrades |
| CommunityPoolSpend | Fund initiatives |
| RegisterIssuer | Approve credential issuer |

### ZK Voting
Anonymous voting using zero-knowledge proofs:
```bash
aurad tx gov vote-zk <proposal-id> yes \
  --proof-type snark \
  --from my-wallet
```

---

## monitoring

Chain health monitoring and alerting.

### Key Features
- Block production metrics
- Validator performance
- Transaction throughput
- Resource utilization

### Queries
```bash
aurad query monitoring chain-health
aurad query monitoring validator-metrics <val-addr>
aurad query monitoring block-stats --blocks 100
aurad query monitoring alerts
```

### Metrics Exposed
Prometheus metrics available at `:26660/metrics`:
- `aura_block_time_seconds`
- `aura_validators_active`
- `aura_tx_throughput`
- `aura_mempool_size`

---

## aiassistant

Decentralized AI oracle network for off-chain verification.

### Key Features
- Verification challenges
- AI node registration
- Fraud detection ML models
- Reward distribution

### Messages
| Message | Description |
|---------|-------------|
| `MsgRegisterAIAssistant` | Register as AI node |
| `MsgSubmitVerification` | Submit verification result |
| `MsgChallengeResult` | Challenge AI result |
| `MsgUpdateModel` | Update ML model (governance) |

### Queries
```bash
aurad query aiassistant assistants
aurad query aiassistant assistant <address>
aurad query aiassistant pending-tasks
aurad query aiassistant model-version
```

### Node Types
| Type | Description |
|------|-------------|
| Verifier | Identity verification |
| FraudDetector | Fraud detection |
| LocaleSpecific | Regional verification |
