# Governance

AURA uses Cosmos SDK governance for on-chain decision making and community contributions for off-chain development.

## Roles

### Core Maintainers
- Review and merge pull requests
- Coordinate releases and roadmap
- Triage issues and security reports

### Contributors
- Submit PRs following [CONTRIBUTING.md](CONTRIBUTING.md)
- Participate in discussions and issue triage
- Report bugs and suggest features

### Validators
- Vote on governance proposals
- Secure network through staking
- Represent delegator interests

## On-Chain Governance

| Parameter | Value |
|-----------|-------|
| Min Deposit | 512 AURA |
| Deposit Period | 14 days |
| Voting Period | 14 days |
| Quorum | 40% |
| Pass Threshold | 50% |
| Veto Threshold | 33.4% |

## CLI Commands

```bash
# Submit a proposal
aurad tx governance submit-proposal "Title" "Description" text --from <key> --initial-deposit 1000uaura

# Deposit to a proposal
aurad tx governance deposit <proposal-id> 1000uaura --from <key>

# Vote on a proposal (yes/no/abstain/no_with_veto)
aurad tx governance vote <proposal-id> yes --from <key>

# Query governance params
aurad query governance params

# Query active proposals
aurad query governance proposals

# Query specific proposal
aurad query governance proposal <id>
```

## Becoming a Maintainer

1. Demonstrate sustained quality contributions over 3+ months
2. Show deep understanding of AURA codebase and Cosmos SDK
3. Receive nomination from an existing maintainer
4. Gain consensus approval from maintainer team
