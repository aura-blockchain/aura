# CRITICAL: Race Condition in Bridge Transfer ID Generation

**Status:** ready
**Priority:** P0 (MAINNET BLOCKER)
**Severity:** CRITICAL
**Category:** Data Integrity / Concurrency
**CWE:** CWE-362 (Concurrent Execution using Shared Resource with Improper Synchronization)

## Summary

Read-modify-write race condition in bridge transfer ID generation can cause duplicate transfer IDs, leading to fund loss and state corruption.

## Location

- **File:** `chain/x/bridge/keeper/keeper.go`
- **Lines:** 145-165
- **Function:** `InitiateTransfer()`

## Vulnerability Details

```go
// RACE CONDITION - Read-modify-write without atomic operation
func (k Keeper) InitiateTransfer(ctx sdk.Context, transfer types.Transfer) (uint64, error) {
    store := ctx.KVStore(k.storeKey)

    // READ: Get current counter
    nextID := k.getNextTransferID(ctx)  // Returns 100

    // WRITE: Store transfer with ID
    transfer.ID = nextID
    bz := k.cdc.MustMarshal(&transfer)
    store.Set(types.TransferKey(nextID), bz)

    // WRITE: Increment counter
    k.setNextTransferID(ctx, nextID+1)  // Store 101

    return nextID, nil
}
```

**Race Scenario:**

```
Time | Node A                      | Node B
-----|----------------------------|---------------------------
T1   | Read counter = 100         |
T2   |                            | Read counter = 100
T3   | Create transfer ID=100     |
T4   |                            | Create transfer ID=100 (DUPLICATE!)
T5   | Write counter = 101        |
T6   |                            | Write counter = 101 (overwrites!)
T7   | Commit block               |
T8   |                            | Consensus fails (different app hash)
```

## Impact

- **Duplicate Transfer IDs:** Two transfers can have same ID
- **Fund Loss:** Transfer 2 overwrites transfer 1, funds lost
- **State Corruption:** Invalid transfer state in database
- **Consensus Failure:** Nodes compute different app hashes
- **Audit Trail Corruption:** Cannot track individual transfers

## Root Cause

Blockchain state transitions are **NOT single-threaded** during:
1. Transaction validation in CheckTx (mempool)
2. Parallel transaction execution (if enabled)
3. Multiple messages in same transaction

The counter increment is not atomic.

## Required Fix

**Option 1: Use Block Height + Tx Index (Recommended)**

```go
// Deterministic, unique, no race condition
func (k Keeper) InitiateTransfer(ctx sdk.Context, transfer types.Transfer) (uint64, error) {
    // Combine block height and transaction index for unique ID
    // Format: (blockHeight << 32) | txIndex
    blockHeight := uint64(ctx.BlockHeight())
    txIndex := uint64(ctx.TxIndex())

    transferID := (blockHeight << 32) | txIndex

    // Verify uniqueness (defensive check)
    if k.HasTransfer(ctx, transferID) {
        return 0, errorsmod.Wrapf(
            types.ErrDuplicateTransfer,
            "transfer ID %d already exists (height=%d, txIndex=%d)",
            transferID, blockHeight, txIndex,
        )
    }

    transfer.ID = transferID
    k.SetTransfer(ctx, &transfer)

    return transferID, nil
}
```

**Benefits:**
- ✅ Deterministic (same input = same ID)
- ✅ No race condition
- ✅ No counter storage needed
- ✅ Can derive block height from ID: `height = id >> 32`
- ✅ Can derive tx index from ID: `txIndex = id & 0xFFFFFFFF`

**Option 2: Atomic Counter (Alternative)**

```go
func (k Keeper) InitiateTransfer(ctx sdk.Context, transfer types.Transfer) (uint64, error) {
    store := ctx.KVStore(k.storeKey)

    // Atomic read-increment-write
    counterKey := types.TransferCounterKey()

    var counter types.Counter
    bz := store.Get(counterKey)
    if bz == nil {
        counter.Value = 0
    } else {
        k.cdc.MustUnmarshal(bz, &counter)
    }

    transferID := counter.Value
    counter.Value++

    // Store updated counter BEFORE storing transfer
    store.Set(counterKey, k.cdc.MustMarshal(&counter))

    // Now store transfer
    transfer.ID = transferID
    k.SetTransfer(ctx, &transfer)

    return transferID, nil
}
```

**Issues with Option 2:**
- ❌ Still has theoretical race in parallel execution
- ❌ Requires counter storage
- ⚠️ Counter becomes contention point

**Recommendation:** Use **Option 1** (block height + tx index)

## Testing Requirements

```go
func TestTransferID_Uniqueness_ParallelExecution(t *testing.T) {
    // Simulate parallel transaction execution
    var wg sync.WaitGroup
    ids := make([]uint64, 100)
    mu := sync.Mutex{}

    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()

            // Simulate transaction execution
            id, err := keeper.InitiateTransfer(ctx, transfer)
            require.NoError(t, err)

            mu.Lock()
            ids[idx] = id
            mu.Unlock()
        }(i)
    }

    wg.Wait()

    // Verify all IDs are unique
    idSet := make(map[uint64]bool)
    for _, id := range ids {
        require.False(t, idSet[id], "Duplicate transfer ID: %d", id)
        idSet[id] = true
    }
}

func TestTransferID_Deterministic(t *testing.T) {
    // Same block height + tx index = same ID
    ctx1 := sdk.NewContext(nil, abci.Header{Height: 100}, false, nil).WithTxIndex(5)
    ctx2 := sdk.NewContext(nil, abci.Header{Height: 100}, false, nil).WithTxIndex(5)

    id1, _ := keeper.InitiateTransfer(ctx1, transfer)
    id2, _ := keeper.InitiateTransfer(ctx2, transfer)

    require.Equal(t, id1, id2, "Same context should produce same ID")
}

func TestTransferID_DecodeBlockHeight(t *testing.T) {
    ctx := sdk.NewContext(nil, abci.Header{Height: 12345}, false, nil).WithTxIndex(67)

    id, _ := keeper.InitiateTransfer(ctx, transfer)

    // Decode block height from ID
    decodedHeight := id >> 32
    decodedTxIndex := id & 0xFFFFFFFF

    require.Equal(t, uint64(12345), decodedHeight)
    require.Equal(t, uint64(67), decodedTxIndex)
}
```

## Migration Plan

If existing transfers use counter-based IDs:

1. **Don't change existing IDs** - Keep them as-is
2. **Mark cutover block height** - e.g., height 100000
3. **Dual ID scheme:**
   ```go
   if ctx.BlockHeight() < CutoverHeight {
       return k.generateLegacyID(ctx)
   }
   return k.generateDeterministicID(ctx)
   ```
4. **Document ID format change** in upgrade notes

## Acceptance Criteria

- [ ] Implement deterministic ID generation (block height + tx index)
- [ ] Remove counter-based ID generation
- [ ] Add uniqueness verification
- [ ] Comprehensive concurrency tests
- [ ] Fuzz testing for edge cases
- [ ] Document ID format
- [ ] Update all code that assumes sequential IDs

## References

- Data Integrity Review: CRITICAL #1
- [CWE-362: Concurrent Execution using Shared Resource](https://cwe.mitre.org/data/definitions/362.html)
- [Race Condition](https://en.wikipedia.org/wiki/Race_condition)
- Cosmos SDK: Transaction execution model

## Related Issues

- See also: todos/047-ready-p1-data-integrity-bridge-counter.md
- Similar pattern may exist in other keeper ID generation

---

**DO NOT DEPLOY TO MAINNET UNTIL FIXED - Fund loss risk**
