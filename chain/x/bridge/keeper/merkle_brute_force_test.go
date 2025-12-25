// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/bridge/keeper"
)

// TestBruteForceVerification tests the brute-force proof verification
func TestBruteForceVerification(t *testing.T) {
	k := &keeper.Keeper{}

	// Simple 2-element tree
	transactions := [][]byte{
		[]byte("tx1"),
		[]byte("tx2"),
	}

	root := keeper.ComputeMerkleRoot(transactions)
	require.NotNil(t, root)

	// Generate proof for tx1 (index 0)
	proof, err := keeper.GenerateMerkleProof(transactions, 0)
	require.NoError(t, err)

	// Extract just the hashes (no indices)
	var proofBytes []byte
	for _, hash := range proof.Proof {
		proofBytes = append(proofBytes, hash...)
	}

	// Verify using the keeper method
	txHash := sha256.Sum256(transactions[0])
	valid := k.VerifyMerkleProofBytes(root, txHash[:], proofBytes)
	require.True(t, valid, "brute force verification should succeed for 2-element tree")

	// Test with 4 elements
	transactions4 := [][]byte{
		[]byte("tx1"),
		[]byte("tx2"),
		[]byte("tx3"),
		[]byte("tx4"),
	}

	root4 := keeper.ComputeMerkleRoot(transactions4)
	proof4, err := keeper.GenerateMerkleProof(transactions4, 2)
	require.NoError(t, err)

	var proofBytes4 []byte
	for _, hash := range proof4.Proof {
		proofBytes4 = append(proofBytes4, hash...)
	}

	txHash4 := sha256.Sum256(transactions4[2])
	valid4 := k.VerifyMerkleProofBytes(root4, txHash4[:], proofBytes4)
	require.True(t, valid4, "brute force verification should succeed for 4-element tree")
}

// TestBruteForceLimit tests that brute force has a reasonable limit
func TestBruteForceLimit(t *testing.T) {
	k := &keeper.Keeper{}

	// Create a fake proof with 11 elements (2^11 = 2048 attempts)
	// This should be rejected as too expensive
	fakeRoot := make([]byte, 32)
	fakeLeaf := make([]byte, 32)
	fakeProof := make([]byte, 32*11) // 11 proof elements

	valid := k.VerifyMerkleProofBytes(fakeRoot, fakeLeaf, fakeProof)
	require.False(t, valid, "proof with >10 elements should be rejected by brute force")
}
