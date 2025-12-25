# Aura Governance

## Overview

Aura uses on-chain governance for protocol upgrades and parameter changes.

## Proposal Types

| Type | Description | Deposit | Voting Period |
|------|-------------|---------|---------------|
| Text | Non-binding signals | 10,000 AURA | 7 days |
| Parameter | Change module params | 50,000 AURA | 14 days |
| Software Upgrade | Chain upgrades | 100,000 AURA | 21 days |
| Community Pool | Fund allocation | 50,000 AURA | 14 days |

## Voting Power

- 1 AURA = 1 vote (direct voting)
- Delegated stake inherits voting power
- Validators vote on behalf of delegators (unless overridden)

## Thresholds

- **Quorum**: 33.4% of staked tokens must vote
- **Pass**: >50% Yes votes (excluding Abstain)
- **Veto**: <33.4% NoWithVeto votes

## Privacy Voting

Aura supports commit-reveal voting for sensitive proposals:
1. Commit phase (7 days): Submit encrypted vote hash
2. Reveal phase (3 days): Reveal vote with salt
3. Tally: Decrypt and count votes

## Emergency Governance

For critical security issues, validator supermajority (67%) can:
- Pause affected modules
- Expedite software upgrades (48-hour voting)
- Revoke compromised authorizations

## Participation

```bash
# Submit proposal
aurad tx gov submit-proposal --type text --title "..." --description "..."

# Deposit
aurad tx gov deposit [proposal-id] 10000uaura

# Vote
aurad tx gov vote [proposal-id] yes|no|abstain|no_with_veto
```

## Resources

- [Governance Module Docs](docs/modules/governance/README.md)
- [ADR-003: Privacy-Preserving Voting](docs/architecture/adr/ADR-003-privacy-voting.md)
