---
sidebar_position: 5
---

# Governance Voting

**Difficulty:** Beginner | **Time:** 10 minutes

Participate in on-chain governance by submitting proposals and voting.

## Governance Overview

Aura uses on-chain governance for:
- Protocol parameter changes
- Software upgrades
- Community spending
- Issuer registration

## Proposal Lifecycle

1. **Deposit Period** (2 days): Proposal needs minimum deposit
2. **Voting Period** (5 days): Stakers vote
3. **Execution**: If passed, changes apply automatically

## Voting

### View Active Proposals

```bash
aurad query gov proposals --status voting_period
```

### Cast Your Vote

```bash
aurad tx gov vote <proposal-id> yes \
  --from my-wallet \
  --chain-id aura-testnet-1 \
  -y
```

Vote options: `yes`, `no`, `abstain`, `no_with_veto`

### Check Proposal Status

```bash
aurad query gov proposal <proposal-id>
```

## Submit a Proposal

### 1. Create Proposal JSON

```json
{
  "title": "Enable New Credential Type",
  "description": "Add support for EducationCredential type",
  "type": "Text",
  "deposit": "10000000uaura"
}
```

### 2. Submit Proposal

```bash
aurad tx gov submit-proposal \
  --title "Enable New Credential Type" \
  --description "Add support for EducationCredential" \
  --type Text \
  --deposit 10000000uaura \
  --from my-wallet \
  -y
```

### 3. Others Add to Deposit

```bash
aurad tx gov deposit <proposal-id> 5000000uaura \
  --from other-wallet \
  -y
```

## Proposal Types

| Type | Purpose |
|------|---------|
| Text | Community signaling |
| ParameterChange | Modify chain parameters |
| SoftwareUpgrade | Coordinate upgrades |
| CommunityPoolSpend | Fund initiatives |

## Voting Power

Your voting power = your staked tokens. Validators vote on behalf of delegators by default, but you can override.

## Next Steps

- [Stake Tokens](./stake-tokens) - Get voting power
- [Module Guide: Governance](/docs/modules/governance)
