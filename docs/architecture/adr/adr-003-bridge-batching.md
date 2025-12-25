# ADR-003: Bridge Transaction Batching

## Status
Accepted

## Context
The bridge module processes pending cross-chain transfers in BeginBlocker. Without batching, a large number of pending transfers could cause chain halts due to unbounded iteration.

## Decision
Implement `MaxPendingTransfersPerBlock` parameter (default: 100) to limit transfers processed per block:
1. ProcessExpiredPendingTransfers processes at most 100 transfers per block
2. Remaining transfers are processed in subsequent blocks
3. Parameter is governance-adjustable

Implementation: `chain/x/bridge/keeper/abci.go`

## Consequences

### Positive
- Prevents chain halts under high load
- Predictable block times
- Governance can adjust based on network capacity

### Negative
- High-volume periods may have transfer delays
- Additional state tracking for processed count

### Neutral
- Most blocks process far fewer than 100 transfers
