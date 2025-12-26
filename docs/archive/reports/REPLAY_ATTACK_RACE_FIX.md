# Bridge Replay Attack Race Condition Fix

## Problem
Multiple transactions with the same burn hash submitted in the same block could all pass the replay check before any were marked as processed, creating a race window.

## Root Cause
Non-atomic check-and-set pattern in UnlockTokens handler:
- Line 394: Check if processed (non-atomic read)
- Lines 399-618: Expensive processing (validator checks, mint limits, Merkle proofs)
- Line 623: Mark as processed (too late - race window exists)

## Solution
Implemented atomic check-and-set pattern using dual-state markers:

### 1. New Key Prefix (chain/x/bridge/types/keys.go)
```go
ProcessingSourceHashPrefix = []byte{0x26}
ProcessingSourceHashKey(sourceChain, sourceHash string) []byte
```

### 2. New Keeper Methods (chain/x/bridge/keeper/keeper.go)
```go
TryMarkSourceHashProcessing(ctx, sourceChain, sourceHash) bool
  - Atomically checks processed + processing states
  - Sets processing marker if safe
  - Returns true only if safe to proceed

FinalizeSourceHashProcessing(ctx, sourceChain, sourceHash)
  - Sets permanent processed marker
  - Removes temporary processing marker
```

### 3. Updated Handler (chain/x/bridge/keeper/msg_server.go)
```go
BEFORE:
  if IsSourceHashProcessed() { reject }
  ... expensive processing ...
  MarkSourceHashProcessed()

AFTER:
  if !TryMarkSourceHashProcessing() { reject }  // Atomic!
  ... expensive processing ...
  FinalizeSourceHashProcessing()
```

## How It Works
1. **First tx**: TryMarkSourceHashProcessing sets processing marker → proceeds
2. **Concurrent tx**: TryMarkSourceHashProcessing sees processing marker → rejects
3. **First tx completes**: FinalizeSourceHashProcessing sets permanent marker
4. **Future txs**: TryMarkSourceHashProcessing sees processed marker → rejects

## Security Properties
- **Atomic**: Check and set in single store operation
- **No race window**: Processing state set before validation
- **Prevents replay**: Both states block duplicates
- **Permanent**: Final marker never removed
- **Audit trail**: Events for both states

## Files Modified
- chain/x/bridge/types/keys.go (new prefix + helper)
- chain/x/bridge/keeper/keeper.go (atomic methods)
- chain/x/bridge/keeper/msg_server.go (updated handler)

## Verification
```bash
cd chain && go build ./x/bridge/...
# ✓ Compiles successfully
```

## Testing Needed
1. Submit 2+ txs with same burn hash in same block
2. Verify only first tx succeeds, others rejected
3. Verify final state has only one processed marker
