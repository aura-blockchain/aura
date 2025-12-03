---
id: "047"
title: "Bridge Transfer Counter Off-By-One Error"
status: ready
priority: p1
category: data-integrity
module: bridge
severity: CRITICAL
source: data-integrity-review
---

# Bridge Transfer Counter Off-By-One Error

## Problem

Transfer counter restored to MAX seen, not MAX+1. Next transfer will have DUPLICATE ID, overwriting existing transfer.

## Affected Files

- `chain/x/bridge/keeper/genesis.go:25-39`

## Vulnerability

```go
var maxTransferCounter uint64
for _, transfer := range data.Transfers {
    if seq, ok := parseTransferSequence(transfer.TransferId); ok && seq > maxTransferCounter {
        maxTransferCounter = seq
    }
}
if maxTransferCounter > 0 {
    bz := make([]byte, 8)
    binary.BigEndian.PutUint64(bz, maxTransferCounter)  // WRONG: Should be maxTransferCounter+1
    k.store(ctx).Set(types.TransferCounterKey, bz)
}
```

## Impact

- Transfer ID collision after chain restart
- Silent overwrite of pending transfers
- Lost bridge transfers and funds

## Required Fix

```go
var maxTransferCounter uint64
transferIDs := make(map[uint64]bool)

for _, transfer := range data.Transfers {
    if transfer == nil {
        continue
    }

    if seq, ok := parseTransferSequence(transfer.TransferId); ok {
        // Detect duplicates
        if transferIDs[seq] {
            return fmt.Errorf("duplicate transfer ID in genesis: %d", seq)
        }
        transferIDs[seq] = true

        if seq > maxTransferCounter {
            maxTransferCounter = seq
        }
    }

    k.setTransfer(ctx, transfer)
}

// Set counter to MAX + 1 (next available ID)
if maxTransferCounter > 0 {
    bz := make([]byte, 8)
    binary.BigEndian.PutUint64(bz, maxTransferCounter+1)  // +1 CRITICAL
    k.store(ctx).Set(types.TransferCounterKey, bz)
}
```

## Acceptance Criteria

- [ ] Counter set to max+1, not max
- [ ] Duplicate detection in genesis import
- [ ] Tests for counter restoration
- [ ] Tests for duplicate rejection
