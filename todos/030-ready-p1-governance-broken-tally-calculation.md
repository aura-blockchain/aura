---
id: "030"
title: "Governance Broken Tally Calculation"
status: ready
priority: p1
category: security
module: governance
severity: CRITICAL
cvss: 10.0
source: governance-security-audit
---

# Governance Broken Tally Calculation

## Problem

The `calculateTallyResult` function sets votes to "1" instead of accumulating them. Every vote just overwrites the previous count.

## Affected Files

- `chain/x/governance/keeper/keeper.go:396-422`

## Vulnerability

```go
for _, vote := range votes {
    switch vote.Option {
    case 1: // Yes
        tally.Yes = "1"       // WRONG: Sets to "1", doesn't add
    case 2: // No
        tally.No = "1"        // Same bug
    case 3: // Abstain
        tally.Abstain = "1"   // Same bug
    case 4: // NoWithVeto
        tally.NoWithVeto = "1" // Same bug
    }
}
```

## Impact

- 1000 Yes votes → tally shows Yes: "1"
- Cannot determine actual vote outcome
- Governance decisions meaningless
- Proposals pass/fail incorrectly

## Required Fix

```go
func (k *Keeper) calculateTallyResult(ctx sdk.Context, proposalId uint64) (*types.TallyResult, error) {
    votes, err := k.GetProposalVotes(ctx, proposalId)
    if err != nil {
        return nil, err
    }

    // Use proper integers for accumulation
    var (
        yesVotes      = sdkmath.ZeroInt()
        noVotes       = sdkmath.ZeroInt()
        abstainVotes  = sdkmath.ZeroInt()
        noWithVeto    = sdkmath.ZeroInt()
    )

    for _, vote := range votes {
        // Get voter's actual voting power
        voterPower, err := k.GetVotingPower(ctx, vote.Voter)
        if err != nil {
            continue // Skip invalid voters
        }

        switch vote.Option {
        case types.VoteOption_VOTE_OPTION_YES:
            yesVotes = yesVotes.Add(voterPower)
        case types.VoteOption_VOTE_OPTION_NO:
            noVotes = noVotes.Add(voterPower)
        case types.VoteOption_VOTE_OPTION_ABSTAIN:
            abstainVotes = abstainVotes.Add(voterPower)
        case types.VoteOption_VOTE_OPTION_NO_WITH_VETO:
            noWithVeto = noWithVeto.Add(voterPower)
        }
    }

    return &types.TallyResult{
        Yes:        yesVotes.String(),
        No:         noVotes.String(),
        Abstain:    abstainVotes.String(),
        NoWithVeto: noWithVeto.String(),
    }, nil
}
```

## Acceptance Criteria

- [ ] Vote counts properly accumulated
- [ ] Voting power weighted correctly
- [ ] Tests for multiple vote accumulation
- [ ] Tests for weighted vote tallying
- [ ] Integration tests for proposal outcomes
