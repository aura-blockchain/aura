---
id: "029"
title: "Governance Hardcoded Voting Power"
status: ready
priority: p1
category: security
module: governance
severity: CRITICAL
cvss: 10.0
source: governance-security-audit
---

# Governance Hardcoded Voting Power

## Problem

The `GetVotingPower` function returns hardcoded "1000" for every address. Everyone has equal voting power regardless of stake, delegation, or other factors.

## Affected Files

- `chain/x/governance/keeper/keeper.go:424-428`

## Vulnerability

```go
func (k *Keeper) GetVotingPower(ctx sdk.Context, address string) string {
    // Simplified: return "1000" for testing
    return "1000"  // EVERYONE HAS SAME POWER!
}
```

## Impact

- Sybil attacks trivial (create many wallets)
- Staking has no governance weight
- Whale votes equal to minimum stake
- Governance completely compromised

## Attack Scenario

```
1. Attacker creates 1000 wallets (cost: gas only)
2. Each wallet has voting power "1000"
3. Total attacker voting power: 1,000,000
4. Attacker controls governance with no stake
5. Passes malicious proposals, drains treasury
```

## Required Fix

```go
func (k *Keeper) GetVotingPower(ctx sdk.Context, address string) (sdkmath.Int, error) {
    totalPower := sdkmath.ZeroInt()

    addr, err := sdk.AccAddressFromBech32(address)
    if err != nil {
        return sdkmath.ZeroInt(), err
    }

    // 1. Direct staked tokens
    stakedAmount := k.stakingKeeper.GetDelegatorBonded(ctx, addr)
    totalPower = totalPower.Add(stakedAmount)

    // 2. Delegated voting power (from others who delegated to this address)
    delegatedPower := k.GetDelegatedVotingPower(ctx, address)
    totalPower = totalPower.Add(delegatedPower)

    // 3. Subtract power delegated away
    delegatedAway := k.GetPowerDelegatedAway(ctx, address)
    totalPower = totalPower.Sub(delegatedAway)

    // Ensure non-negative
    if totalPower.IsNegative() {
        totalPower = sdkmath.ZeroInt()
    }

    return totalPower, nil
}

func (k *Keeper) GetDelegatedVotingPower(ctx sdk.Context, delegatee string) sdkmath.Int {
    totalDelegated := sdkmath.ZeroInt()

    // Iterate all delegations to this address
    k.IterateDelegations(ctx, func(delegation types.VoteDelegation) bool {
        if delegation.Delegatee == delegatee {
            // Get delegator's staked tokens
            delegatorAddr, _ := sdk.AccAddressFromBech32(delegation.Delegator)
            delegatorStake := k.stakingKeeper.GetDelegatorBonded(ctx, delegatorAddr)
            totalDelegated = totalDelegated.Add(delegatorStake)
        }
        return false // continue iteration
    })

    return totalDelegated
}
```

## Acceptance Criteria

- [ ] Voting power derived from staked tokens
- [ ] Vote delegation properly calculated
- [ ] Sybil resistance achieved
- [ ] Tests for voting power calculation
- [ ] Tests for delegation scenarios
- [ ] Tests for Sybil attack prevention
