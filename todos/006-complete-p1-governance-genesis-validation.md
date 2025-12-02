---
status: ready
priority: p1
issue_id: "006"
tags: [code-review, security, governance, genesis]
dependencies: []
---

# Governance Genesis Validation Missing Range Checks

## Problem Statement

`ValidateGenesis` checks that governance parameters are non-empty, but doesn't validate:
- Quorum/Threshold/VetoThreshold are valid percentages (0-100% or 0.0-1.0)
- Thresholds are logically consistent (e.g., veto < threshold)
- Periods are reasonable (not 0 seconds, not 1000 years)
- MinDeposit is a valid coin amount (positive, correct denom)

**Why it matters:** Invalid genesis parameters could deadlock governance, enable spam proposals, or allow instant malicious proposals to pass.

## Findings

### Evidence
- **File:** `chain/x/governance/types/genesis.go`
- **Lines:** 18-36

### Economic Attack Scenario
```json
{
  "params": {
    "quorum": "101",          // > 100% = impossible to reach
    "threshold": "0.1",       // 10% approval needed
    "veto_threshold": "0.9",  // 90% veto needed (backwards!)
    "min_deposit": "-1000",   // Negative = free proposals!
    "voting_period": "1ns"    // 1 nanosecond = instant pass
  }
}
```

Chain initialized with this genesis:
- **Governance deadlocked** (quorum impossible)
- **Spam proposals** (negative/zero deposit)
- **Instant malicious proposals** (1ns voting period)

### Impact
- Complete governance deadlock
- Economic attacks via spam
- Malicious proposals pass instantly
- Chain may require hard fork to fix

## Proposed Solutions

### Option A: Comprehensive Parameter Validation (Required)
**Pros:** Prevents all known attack vectors
**Cons:** More validation code
**Effort:** Medium (2-3 hours)
**Risk:** Low

```go
func ValidateGenesis(g *GenesisState) error {
    if g == nil {
        return fmt.Errorf("governance genesis cannot be nil")
    }
    if g.Params == nil {
        return fmt.Errorf("governance params cannot be nil")
    }

    // Parse and validate thresholds
    quorum, err := sdk.NewDecFromStr(g.Params.Quorum)
    if err != nil || quorum.LT(sdk.ZeroDec()) || quorum.GT(sdk.OneDec()) {
        return fmt.Errorf("invalid quorum: must be 0.0-1.0, got %s", g.Params.Quorum)
    }

    threshold, err := sdk.NewDecFromStr(g.Params.Threshold)
    if err != nil || threshold.LT(sdk.ZeroDec()) || threshold.GT(sdk.OneDec()) {
        return fmt.Errorf("invalid threshold: must be 0.0-1.0, got %s", g.Params.Threshold)
    }

    vetoThreshold, err := sdk.NewDecFromStr(g.Params.VetoThreshold)
    if err != nil || vetoThreshold.LT(sdk.ZeroDec()) || vetoThreshold.GT(sdk.OneDec()) {
        return fmt.Errorf("invalid veto threshold: must be 0.0-1.0, got %s", g.Params.VetoThreshold)
    }

    // Logical consistency
    if vetoThreshold.GTE(threshold) {
        return fmt.Errorf("veto_threshold (%s) must be < threshold (%s)", vetoThreshold, threshold)
    }

    // Validate deposit
    minDeposit, err := sdk.ParseCoinsNormalized(g.Params.MinDeposit)
    if err != nil {
        return fmt.Errorf("invalid min_deposit: %w", err)
    }
    if !minDeposit.IsAllPositive() {
        return fmt.Errorf("min_deposit must be positive, got %s", minDeposit)
    }

    // Validate periods (min 1 minute, max 1 year)
    if g.Params.MaxDepositPeriod.Seconds < 60 {
        return fmt.Errorf("max_deposit_period must be >= 1 minute")
    }
    if g.Params.VotingPeriod.Seconds < 60 {
        return fmt.Errorf("voting_period must be >= 1 minute")
    }
    if g.Params.VotingPeriod.Seconds > 365*24*3600 {
        return fmt.Errorf("voting_period must be <= 1 year")
    }

    return nil
}
```

## Recommended Action
Implement comprehensive validation before any genesis file is used.

## Technical Details

### Affected Files
- `chain/x/governance/types/genesis.go`

### Acceptance Criteria
- [ ] All threshold values validated as 0.0-1.0
- [ ] Logical consistency enforced (veto < threshold)
- [ ] MinDeposit validated as positive coin amount
- [ ] Voting/deposit periods have reasonable bounds
- [ ] Unit tests for each validation rule
- [ ] Invalid genesis files rejected with clear errors

## Work Log

| Date | Action | Learnings |
|------|--------|-----------|
| 2025-12-02 | Finding identified in data integrity review | Economic security critical |

## Resources
- [Cosmos SDK Governance](https://docs.cosmos.network/main/build/modules/gov)
