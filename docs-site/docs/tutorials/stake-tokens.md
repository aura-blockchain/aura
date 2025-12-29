---
sidebar_position: 4
---

# Stake Tokens

**Difficulty:** Beginner | **Time:** 5 minutes

Delegate AURA tokens to validators and earn staking rewards.

## Why Stake?

- **Earn Rewards**: ~12-18% APR on staked tokens
- **Secure the Network**: Validators use staked tokens for consensus
- **Governance Power**: Staked tokens give voting rights

## Steps

### 1. View Available Validators

```bash
aurad query staking validators --limit 100
```

Key metrics to consider:
- **Commission**: Fee the validator takes (5-10% typical)
- **Uptime**: Historical availability
- **Voting Power**: Don't over-concentrate on top validators

### 2. Delegate Tokens

```bash
aurad tx staking delegate <validator-address> 1000000uaura \
  --from my-wallet \
  --chain-id aura-testnet-1 \
  --gas auto \
  -y
```

Note: `1000000uaura` = 1 AURA

### 3. Check Your Delegation

```bash
aurad query staking delegations $(aurad keys show my-wallet -a)
```

### 4. Claim Rewards

```bash
aurad tx distribution withdraw-all-rewards \
  --from my-wallet \
  --chain-id aura-testnet-1 \
  -y
```

## Managing Delegations

### Redelegate (Switch Validators)

```bash
aurad tx staking redelegate \
  <src-validator> <dst-validator> 500000uaura \
  --from my-wallet \
  -y
```

No unbonding period for redelegation.

### Undelegate

```bash
aurad tx staking unbond <validator-address> 500000uaura \
  --from my-wallet \
  -y
```

**Important**: 21-day unbonding period. Tokens locked during this time.

## Best Practices

1. **Diversify**: Stake across multiple validators
2. **Research**: Check validator track record
3. **Monitor**: Watch for slashing events
4. **Compound**: Regularly claim and restake rewards

## Next Steps

- [Governance](./governance) - Vote on proposals with staked tokens
- [Validator Setup](/docs/validators/setup) - Run your own validator
