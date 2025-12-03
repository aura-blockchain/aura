---
id: "059"
title: "Confidence Score AddScoreChange TODO Breaks Feature"
status: ready
priority: p2
category: data-integrity
module: confidencescore
severity: HIGH
source: data-integrity-review
---

# Confidence Score AddScoreChange TODO Breaks Feature

## Problem

Critical TODO in production code uses "unknown" as wallet address. All score changes stored under same key, making history queries useless.

## Affected Files

- `chain/x/confidencescore/keeper/keeper.go:227-253`

## Vulnerable Code

```go
key := fmt.Sprintf("%s%s/%d/%s",
    types.ScoreHistoryStoreKeyPrefix,
    "unknown",  // BUG: Always "unknown"!!!
    change.BlockHeight,
    change.TxHash)
```

## Impact

- Cannot query score history by address
- All user histories mixed together
- Feature completely broken

## Required Fix

Pass wallet address as parameter and use it in key.

## Acceptance Criteria

- [ ] AddScoreChange accepts wallet address parameter
- [ ] Correct key format with actual address
- [ ] History queries work per-address
- [ ] Tests for history retrieval
