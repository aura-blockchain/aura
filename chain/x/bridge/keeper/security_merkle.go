package keeper

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"github.com/aequitas/aura/chain/x/bridge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================================================================
// MERKLE PROOF VERIFICATION
// ============================================================================

// VerifyMerkleProof verifies a Merkle proof for cross-chain transfers
// This ensures that a transaction truly occurred on the source chain
func (k Keeper) VerifyMerkleProof(
	ctx sdk.Context,
	proof *types.MerkleProof,
	transferData []byte,
) error {
	if proof == nil {
		return fmt.Errorf("merkle proof cannot be nil")
	}

	// Verify the leaf matches the transfer data
	leafHash := sha256.Sum256(transferData)
	if !bytes.Equal(leafHash[:], proof.Leaf) {
		return fmt.Errorf("leaf hash mismatch: expected %x, got %x", proof.Leaf, leafHash[:])
	}

	// Verify the Merkle path
	currentHash := proof.Leaf
	for i, siblingHash := range proof.Proof {
		var combinedHash []byte

		// Determine order based on index
		if i < len(proof.Indices) && proof.Indices[i]%2 == 0 {
			// Current hash is on the left
			combinedHash = append(currentHash, siblingHash...)
		} else {
			// Current hash is on the right
			combinedHash = append(siblingHash, currentHash...)
		}

		// Hash the combined pair
		hash := sha256.Sum256(combinedHash)
		currentHash = hash[:]
	}

	// Verify root matches
	if !bytes.Equal(currentHash, proof.Root) {
		return fmt.Errorf("root hash mismatch: expected %x, got %x", proof.Root, currentHash)
	}

	// Verify the root against a known block header from the source chain
	storedRoot := k.GetMerkleRoot(ctx, proof.ChainId, proof.BlockHeight)
	if storedRoot == nil {
		return fmt.Errorf(
			"no stored merkle root for chain %s at height %d - cannot verify proof against source chain",
			proof.ChainId,
			proof.BlockHeight,
		)
	}

	// Verify that the computed root matches the stored root from the source chain
	if !bytes.Equal(proof.Root, storedRoot) {
		return fmt.Errorf(
			"merkle root does not match source chain: proof root %x, stored root %x",
			proof.Root,
			storedRoot,
		)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"merkle_proof_verified",
			sdk.NewAttribute("chain_id", proof.ChainId),
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", proof.BlockHeight)),
			sdk.NewAttribute("root", fmt.Sprintf("%x", proof.Root)),
		),
	)

	return nil
}

// StoreMerkleRoot stores a verified Merkle root from a source chain
// This is typically called by relayers who monitor source chains
func (k Keeper) StoreMerkleRoot(
	ctx sdk.Context,
	chainId string,
	blockHeight uint64,
	root []byte,
) error {
	store := ctx.KVStore(k.storeKey)
	key := types.MerkleRootKey(chainId, blockHeight)

	store.Set(key, root)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"merkle_root_stored",
			sdk.NewAttribute("chain_id", chainId),
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", blockHeight)),
			sdk.NewAttribute("root", fmt.Sprintf("%x", root)),
		),
	)

	return nil
}

// GetMerkleRoot retrieves a stored Merkle root
func (k Keeper) GetMerkleRoot(
	ctx sdk.Context,
	chainId string,
	blockHeight uint64,
) []byte {
	store := ctx.KVStore(k.storeKey)
	key := types.MerkleRootKey(chainId, blockHeight)

	return store.Get(key)
}

// VerifyTransferWithProof verifies a cross-chain transfer using a Merkle proof
func (k Keeper) VerifyTransferWithProof(
	ctx sdk.Context,
	transfer *types.CrossChainTransfer,
	proof *types.MerkleProof,
) error {
	// Encode transfer data for verification
	transferData := k.cdc.MustMarshal(transfer)

	// Verify the Merkle proof
	if err := k.VerifyMerkleProof(ctx, proof, transferData); err != nil {
		// Slash validator if they submitted an invalid proof
		if transfer.RelayerAddress != "" {
			k.SlashValidatorForInvalidProof(ctx, transfer.RelayerAddress, transfer.TransferId)
		}
		return fmt.Errorf("merkle proof verification failed: %w", err)
	}

	return nil
}

// ComputeMerkleRoot computes a Merkle root from a list of transfers
// This is used for testing and validation purposes
func ComputeMerkleRoot(transfers [][]byte) []byte {
	if len(transfers) == 0 {
		return nil
	}

	// Build the Merkle tree bottom-up
	currentLevel := make([][]byte, len(transfers))
	for i, transfer := range transfers {
		hash := sha256.Sum256(transfer)
		currentLevel[i] = hash[:]
	}

	// Build up the tree
	for len(currentLevel) > 1 {
		nextLevel := make([][]byte, (len(currentLevel)+1)/2)

		for i := 0; i < len(currentLevel); i += 2 {
			var combined []byte
			combined = append(combined, currentLevel[i]...)

			// If there's a pair, add it; otherwise duplicate the last node
			if i+1 < len(currentLevel) {
				combined = append(combined, currentLevel[i+1]...)
			} else {
				combined = append(combined, currentLevel[i]...)
			}

			hash := sha256.Sum256(combined)
			nextLevel[i/2] = hash[:]
		}

		currentLevel = nextLevel
	}

	return currentLevel[0]
}

// GenerateMerkleProof generates a Merkle proof for a specific leaf
// This is used for testing purposes
func GenerateMerkleProof(transfers [][]byte, leafIndex int) (*types.MerkleProof, error) {
	if leafIndex < 0 || leafIndex >= len(transfers) {
		return nil, fmt.Errorf("invalid leaf index: %d", leafIndex)
	}

	// Compute leaf hash
	leafHash := sha256.Sum256(transfers[leafIndex])

	// Build proof path
	var proofPath [][]byte
	var indices []uint64
	currentIndex := uint64(leafIndex)

	// Build the tree and extract the proof path
	currentLevel := make([][]byte, len(transfers))
	for i, transfer := range transfers {
		hash := sha256.Sum256(transfer)
		currentLevel[i] = hash[:]
	}

	for len(currentLevel) > 1 {
		nextLevel := make([][]byte, (len(currentLevel)+1)/2)

		// Find sibling for current index
		var siblingIndex uint64
		if currentIndex%2 == 0 {
			siblingIndex = currentIndex + 1
		} else {
			siblingIndex = currentIndex - 1
		}

		// Add sibling to proof if it exists
		if int(siblingIndex) < len(currentLevel) {
			proofPath = append(proofPath, currentLevel[siblingIndex])
			indices = append(indices, currentIndex)
		}

		// Build next level
		for i := 0; i < len(currentLevel); i += 2 {
			var combined []byte
			combined = append(combined, currentLevel[i]...)

			if i+1 < len(currentLevel) {
				combined = append(combined, currentLevel[i+1]...)
			} else {
				combined = append(combined, currentLevel[i]...)
			}

			hash := sha256.Sum256(combined)
			nextLevel[i/2] = hash[:]
		}

		currentLevel = nextLevel
		currentIndex = currentIndex / 2
	}

	root := currentLevel[0]

	return &types.MerkleProof{
		Root:    root,
		Leaf:    leafHash[:],
		Proof:   proofPath,
		Indices: indices,
	}, nil
}
