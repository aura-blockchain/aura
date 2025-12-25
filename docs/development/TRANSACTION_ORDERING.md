# Transaction Ordering and Concurrency

## Overview

Aura inherits Cosmos SDK transaction ordering guarantees. Understanding these guarantees is critical for building reliable applications that handle concurrent transactions.

## Sequence-Based Ordering (Nonces)

Every account has a sequence number (nonce) that increments with each transaction.

- **Strict Ordering**: Transactions from the same account execute in sequence order (0, 1, 2, ...)
- **Gap Rejection**: Transaction with sequence N+2 is rejected if sequence N+1 hasn't executed
- **No Reordering**: Mempool cannot reorder transactions from the same account

## Same-Block Transaction Handling

Multiple transactions in the same block are deterministically ordered:

1. **Priority-based ordering**: Higher gas price → earlier execution (when mempool priority is enabled)
2. **Deterministic fallback**: When priorities are equal, lexicographic ordering by transaction hash
3. **Account sequence**: Within same account, sequence number determines order regardless of priority

## Concurrent Transaction Conflicts

### State Conflicts

When two concurrent transactions modify the same state:

- **Winner**: First transaction in block order commits successfully
- **Loser**: Second transaction may fail if state dependencies are violated (e.g., insufficient balance)

### Example: Double-Spend Prevention

```
Block N:
  Tx1: Alice sends 100 AURA to Bob (sequence: 5)
  Tx2: Alice sends 100 AURA to Carol (sequence: 6)
  Alice's balance: 100 AURA

Result: Tx1 succeeds, Tx2 fails (insufficient balance)
```

## Replay Protection

- **Sequence numbers** prevent replay attacks - same transaction cannot be resubmitted
- **Chain-id** binds transactions to specific chain (testnet vs mainnet)
- **Block height/time** optional timeout for time-sensitive transactions

## Best Practices

1. **Check sequence**: Always query current sequence before submitting transactions
2. **Handle failures**: Implement retry logic for sequence mismatches and state conflicts
3. **Use timeouts**: Set `timeout_height` for time-sensitive operations to prevent stale execution
4. **Avoid assumptions**: Don't assume transaction order unless same account with sequential nonces
5. **Idempotency**: Design operations to be safely retried (critical for cross-chain operations)

## Module-Specific Considerations

### Identity Module
- Key rotations have grace periods - concurrent operations during grace period use old key
- Change requests have status transitions - concurrent requests may conflict on status field

### Compliance Module
- KYC records use version field for optimistic concurrency control
- Rate limits are enforced per-block, not per-transaction

### VC Registry
- Daily mint limits checked at BeginBlock - all transactions in block share same daily counter
- Revocation Merkle tree updated after each revocation in deterministic order
