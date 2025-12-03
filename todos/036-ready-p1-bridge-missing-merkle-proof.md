---
id: "036"
title: "Bridge Missing Merkle Proof Verification"
status: ready
priority: p1
category: security
module: bridge
severity: CRITICAL
cvss: 9.5
source: bridge-security-matrix
---

# Bridge Missing Merkle Proof Verification

## Problem

No cryptographic proof verification that the source chain event actually occurred. Validators can attest to fake deposits.

## Affected Files

- `chain/x/bridge/keeper/msg_server.go`
- `chain/x/bridge/types/tx.proto`

## Current State

```go
// Current: Just checks validator signatures
// No proof that the burn transaction actually happened on source chain
func (ms msgServer) UnlockTokens(...) {
    // Verify validator signatures (can be colluded)
    // NO: Merkle proof of source transaction
    // NO: Light client verification
    // NO: State proof from source chain
}
```

## Impact

- Validators can fabricate deposits
- Colluding validators can drain bridge
- No cryptographic guarantee of source events

## Required Fix

```protobuf
// Update MsgUnlockTokens
message MsgUnlockTokens {
    string creator = 1;
    string source_chain = 2;
    string burn_tx_hash = 3;
    string sender = 4;
    string recipient = 5;
    string amount = 6;
    string denom = 7;
    repeated ValidatorSignature validator_signatures = 8;

    // ADD: Merkle proof of source transaction inclusion
    bytes merkle_proof = 9;
    bytes merkle_root = 10;
    uint64 source_block_height = 11;
    bytes source_block_hash = 12;
}
```

```go
func (ms msgServer) UnlockTokens(goCtx context.Context, msg *bridgepb.MsgUnlockTokens) (*bridgepb.MsgUnlockTokensResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // 1. Verify the source block is finalized (via light client or oracle)
    sourceBlockVerified := ms.Keeper.VerifySourceBlock(
        ctx,
        msg.SourceChain,
        msg.SourceBlockHeight,
        msg.SourceBlockHash,
    )
    if !sourceBlockVerified {
        return nil, status.Error(codes.FailedPrecondition, "source block not verified")
    }

    // 2. Verify Merkle proof that transaction is in the block
    txLeaf := ms.constructTransactionLeaf(msg)
    proofValid := merkle.VerifyProof(
        msg.MerkleRoot,
        txLeaf,
        msg.MerkleProof,
    )
    if !proofValid {
        return nil, status.Error(codes.InvalidArgument, "invalid merkle proof")
    }

    // 3. Verify Merkle root matches the verified block's transaction root
    expectedRoot := ms.Keeper.GetSourceBlockTxRoot(ctx, msg.SourceChain, msg.SourceBlockHeight)
    if !bytes.Equal(msg.MerkleRoot, expectedRoot) {
        return nil, status.Error(codes.InvalidArgument, "merkle root mismatch")
    }

    // 4. Then verify validator attestations as additional security
    // ...
}

func (k Keeper) VerifySourceBlock(ctx sdk.Context, chain string, height uint64, hash []byte) bool {
    // Option 1: Use IBC light client
    // Option 2: Use decentralized oracle network
    // Option 3: Use threshold attestation with fraud proofs

    // Get light client state for source chain
    clientState := k.ibcKeeper.GetClientState(ctx, chain)
    if clientState == nil {
        return false
    }

    // Verify block hash at height
    return clientState.VerifyBlockAtHeight(height, hash)
}
```

## Acceptance Criteria

- [ ] Merkle proof verification implemented
- [ ] Source block verification (light client or oracle)
- [ ] Transaction leaf construction matches source chain format
- [ ] Tests for invalid Merkle proof rejection
- [ ] Tests for wrong Merkle root rejection
- [ ] Integration with IBC light client (if available)
