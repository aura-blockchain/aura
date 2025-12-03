---
id: "032"
title: "Governance No Quorum/Threshold Enforcement"
status: ready
priority: p1
category: security
module: governance
severity: CRITICAL
cvss: 9.0
source: governance-security-audit
---

# Governance No Quorum/Threshold Enforcement

## Problem

Proposals pass/fail without checking quorum (minimum participation) or threshold (minimum yes votes). A single vote can pass any proposal.

## Affected Files

- `chain/x/governance/keeper/keeper.go:370-394`

## Vulnerability

```go
func (k *Keeper) processProposalOutcome(ctx sdk.Context, proposal *types.Proposal) error {
    tally, err := k.calculateTallyResult(ctx, proposal.Id)
    if err != nil {
        return err
    }

    // WRONG: Just compares yes vs no, no quorum check
    yesVotes, _ := strconv.ParseUint(tally.Yes, 10, 64)
    noVotes, _ := strconv.ParseUint(tally.No, 10, 64)

    if yesVotes > noVotes {
        proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_PASSED
    } else {
        proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_REJECTED
    }
    // No check: Did enough people vote?
    // No check: Did yes votes exceed threshold?
    // No check: Did veto exceed veto threshold?
}
```

## Impact

- Proposals pass with 1 yes vote, 0 no votes
- Minority can control governance
- No participation requirements
- Illegitimate governance decisions

## Required Fix

```go
func (k *Keeper) processProposalOutcome(ctx sdk.Context, proposal *types.Proposal) error {
    tally, err := k.calculateTallyResult(ctx, proposal.Id)
    if err != nil {
        return err
    }

    params := k.GetParams(ctx)

    // Parse tally results
    yesVotes, _ := sdkmath.NewIntFromString(tally.Yes)
    noVotes, _ := sdkmath.NewIntFromString(tally.No)
    abstainVotes, _ := sdkmath.NewIntFromString(tally.Abstain)
    noWithVetoVotes, _ := sdkmath.NewIntFromString(tally.NoWithVeto)

    totalVotes := yesVotes.Add(noVotes).Add(abstainVotes).Add(noWithVetoVotes)

    // Get total voting power in system
    totalBondedTokens := k.stakingKeeper.TotalBondedTokens(ctx)

    // 1. CHECK QUORUM
    // Quorum = minimum participation required (e.g., 33% of bonded tokens must vote)
    quorumDec, _ := sdkmath.LegacyNewDecFromStr(params.Quorum) // e.g., "0.334"
    requiredQuorum := quorumDec.MulInt(totalBondedTokens).TruncateInt()

    if totalVotes.LT(requiredQuorum) {
        proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_REJECTED
        proposal.FinalTallyResult = tally
        proposal.RejectionReason = fmt.Sprintf(
            "quorum not reached: %s votes < %s required",
            totalVotes.String(), requiredQuorum.String(),
        )
        return k.SetProposal(ctx, *proposal)
    }

    // 2. CHECK VETO THRESHOLD
    // If > 33% vote NoWithVeto, proposal fails regardless of yes votes
    vetoThresholdDec, _ := sdkmath.LegacyNewDecFromStr(params.VetoThreshold) // e.g., "0.334"
    vetoThreshold := vetoThresholdDec.MulInt(totalVotes).TruncateInt()

    if noWithVetoVotes.GT(vetoThreshold) {
        proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_REJECTED
        proposal.FinalTallyResult = tally
        proposal.RejectionReason = fmt.Sprintf(
            "vetoed: %s veto votes > %s threshold",
            noWithVetoVotes.String(), vetoThreshold.String(),
        )
        // Burn deposits for vetoed proposals
        k.DeleteDeposits(ctx, proposal.Id)
        return k.SetProposal(ctx, *proposal)
    }

    // 3. CHECK PASS THRESHOLD
    // Yes votes must exceed threshold of non-abstaining votes
    votesExcludingAbstain := yesVotes.Add(noVotes).Add(noWithVetoVotes)
    if votesExcludingAbstain.IsZero() {
        proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_REJECTED
        proposal.FinalTallyResult = tally
        proposal.RejectionReason = "all votes were abstain"
        return k.SetProposal(ctx, *proposal)
    }

    thresholdDec, _ := sdkmath.LegacyNewDecFromStr(params.Threshold) // e.g., "0.5"
    passThreshold := thresholdDec.MulInt(votesExcludingAbstain).TruncateInt()

    if yesVotes.GT(passThreshold) {
        proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_PASSED
    } else {
        proposal.Status = types.ProposalStatus_PROPOSAL_STATUS_REJECTED
        proposal.RejectionReason = fmt.Sprintf(
            "threshold not met: %s yes votes <= %s required",
            yesVotes.String(), passThreshold.String(),
        )
    }

    proposal.FinalTallyResult = tally

    // Refund deposits for passed/rejected (non-vetoed) proposals
    k.RefundDeposits(ctx, proposal.Id)

    return k.SetProposal(ctx, *proposal)
}
```

## Acceptance Criteria

- [ ] Quorum check implemented
- [ ] Pass threshold check implemented
- [ ] Veto threshold check implemented
- [ ] Deposit handling (refund vs burn)
- [ ] Tests for quorum failure
- [ ] Tests for threshold failure
- [ ] Tests for veto rejection
