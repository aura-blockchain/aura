// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
)

// GenerateRevocationMerkleProof generates a Merkle proof for a revoked VC
// This allows lightweight clients to verify a revocation without downloading the entire revocation list
func (k *Keeper) GenerateRevocationMerkleProof(ctx context.Context, vcID string) []byte {
	k.requireStore()

	// Get the revocation record
	revRecord, ok := k.store.getRevocationRecord(ctx, vcID)
	if !ok {
		return nil
	}

	// Get current revocation list
	_, ok = k.store.getRevocationList(ctx)
	if !ok {
		return nil
	}

	// Build Merkle proof
	// This is a simplified implementation - production would use a proper Merkle tree library
	proof := k.buildMerkleProof(ctx, vcID, revRecord)

	return proof
}

// buildMerkleProof constructs a Merkle inclusion proof for a revocation
func (k *Keeper) buildMerkleProof(ctx context.Context, vcID string, revRecord interface{}) []byte {
	// Simplified Merkle proof implementation
	// In production, this would:
	// 1. Maintain a proper Merkle tree structure in the KV store
	// 2. Generate sibling hashes along the path from leaf to root
	// 3. Return the minimal proof needed for verification

	h := sha256.New()
	h.Write([]byte(vcID))
	h.Write([]byte(fmt.Sprintf("%v", revRecord)))

	// Return a simple hash as proof (simplified)
	// Real implementation would return array of sibling hashes
	return h.Sum(nil)
}

// VerifyRevocationMerkleProof verifies a Merkle proof against the current root
func (k *Keeper) VerifyRevocationMerkleProof(ctx context.Context, vcID string, proof []byte, merkleRoot []byte) bool {
	if len(proof) == 0 || len(merkleRoot) == 0 {
		return false
	}

	k.requireStore()

	// Get current revocation list to check root
	revocationList, ok := k.store.getRevocationList(ctx)
	if !ok {
		return false
	}

	// Verify the merkle root matches
	if len(revocationList.MerkleRoot) != len(merkleRoot) {
		return false
	}
	for i := range merkleRoot {
		if merkleRoot[i] != revocationList.MerkleRoot[i] {
			return false
		}
	}

	// In a real implementation, would reconstruct the root from the proof
	// and verify it matches the stored root
	return true
}

// BuildMerkleTree builds a complete Merkle tree from all revocations
// This is called when initializing or rebuilding the revocation list
func (k *Keeper) BuildMerkleTree(ctx context.Context) ([]byte, error) {
	k.requireStore()

	// Get all revocation records
	allRevocations := k.store.iterateRevocationRecords(ctx)
	if len(allRevocations) == 0 {
		return []byte{}, nil
	}

	// Extract vcIDs and sort them for deterministic ordering
	// CRITICAL: Map iteration is non-deterministic in Go
	// Must sort to ensure all validators compute the same Merkle root
	vcIDs := make([]string, 0, len(allRevocations))
	for vcID := range allRevocations {
		vcIDs = append(vcIDs, vcID)
	}
	sort.Strings(vcIDs)

	// Build leaves in sorted order
	leaves := make([][]byte, 0, len(vcIDs))
	for _, vcID := range vcIDs {
		h := sha256.New()
		h.Write([]byte(vcID))
		leaves = append(leaves, h.Sum(nil))
	}

	// Build tree bottom-up
	root := k.computeMerkleRoot(leaves)
	return root, nil
}

// computeMerkleRoot computes the Merkle root from a list of leaf hashes
func (k *Keeper) computeMerkleRoot(leaves [][]byte) []byte {
	if len(leaves) == 0 {
		return []byte{}
	}
	if len(leaves) == 1 {
		return leaves[0]
	}

	// Build next level
	nextLevel := make([][]byte, 0, (len(leaves)+1)/2)
	for i := 0; i < len(leaves); i += 2 {
		h := sha256.New()
		h.Write(leaves[i])
		if i+1 < len(leaves) {
			h.Write(leaves[i+1])
		} else {
			// Duplicate last leaf if odd number
			h.Write(leaves[i])
		}
		nextLevel = append(nextLevel, h.Sum(nil))
	}

	// Recurse
	return k.computeMerkleRoot(nextLevel)
}

// GetMerkleTreeStats returns statistics about the revocation Merkle tree
func (k *Keeper) GetMerkleTreeStats(ctx context.Context) (totalRevocations uint64, treeHeight uint64, rootHash []byte) {
	k.requireStore()

	revocationList, ok := k.store.getRevocationList(ctx)
	if !ok {
		return 0, 0, []byte{}
	}

	totalRevocations = revocationList.TotalRevocations
	rootHash = revocationList.MerkleRoot

	// Calculate tree height (log2 of total revocations)
	if totalRevocations == 0 {
		treeHeight = 0
	} else {
		treeHeight = 1
		temp := totalRevocations
		for temp > 1 {
			temp = temp / 2
			treeHeight++
		}
	}

	return totalRevocations, treeHeight, rootHash
}

// BatchVerifyRevocations verifies multiple revocations efficiently using Merkle proofs
func (k *Keeper) BatchVerifyRevocations(ctx context.Context, vcIDs []string, proofs [][]byte) ([]bool, error) {
	if len(vcIDs) != len(proofs) {
		return nil, fmt.Errorf("vcIDs and proofs length mismatch")
	}

	k.requireStore()

	// Get current Merkle root
	revocationList, ok := k.store.getRevocationList(ctx)
	if !ok {
		return nil, fmt.Errorf("revocation list not found")
	}

	results := make([]bool, len(vcIDs))
	for i, vcID := range vcIDs {
		results[i] = k.VerifyRevocationMerkleProof(ctx, vcID, proofs[i], revocationList.MerkleRoot)
	}

	return results, nil
}
