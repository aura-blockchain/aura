// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/bridge/keeper"
)

// TestVerifyMerkleProofBytes_WithIndices tests the new format with indices
func TestVerifyMerkleProofBytes_WithIndices(t *testing.T) {
	k := &keeper.Keeper{}

	transactions := [][]byte{
		[]byte("transaction1"),
		[]byte("transaction2"),
		[]byte("transaction3"),
		[]byte("transaction4"),
	}

	root := keeper.ComputeMerkleRoot(transactions)
	proof, err := keeper.GenerateMerkleProof(transactions, 1)
	require.NoError(t, err)

	// New format: encode indices with hashes
	var proofBytes []byte
	for i, siblingHash := range proof.Proof {
		// 1 byte for index + 32 bytes for hash
		proofBytes = append(proofBytes, byte(proof.Indices[i]))
		proofBytes = append(proofBytes, siblingHash...)
	}

	txHash := sha256.Sum256(transactions[1])
	valid := k.VerifyMerkleProofBytes(root, txHash[:], proofBytes)
	require.True(t, valid, "proof with indices should verify")
}

// TestVerifyMerkleProofBytes_BothFormats tests both formats produce same result
func TestVerifyMerkleProofBytes_BothFormats(t *testing.T) {
	k := &keeper.Keeper{}

	transactions := [][]byte{
		[]byte("tx1"),
		[]byte("tx2"),
		[]byte("tx3"),
		[]byte("tx4"),
	}

	root := keeper.ComputeMerkleRoot(transactions)

	// Test each transaction
	for idx := 0; idx < len(transactions); idx++ {
		proof, err := keeper.GenerateMerkleProof(transactions, idx)
		require.NoError(t, err)

		txHash := sha256.Sum256(transactions[idx])

		// Format 1: Hash only (brute force)
		var proofBytesHashOnly []byte
		for _, siblingHash := range proof.Proof {
			proofBytesHashOnly = append(proofBytesHashOnly, siblingHash...)
		}

		// Format 2: Index + Hash (deterministic)
		var proofBytesWithIndices []byte
		for i, siblingHash := range proof.Proof {
			proofBytesWithIndices = append(proofBytesWithIndices, byte(proof.Indices[i]))
			proofBytesWithIndices = append(proofBytesWithIndices, siblingHash...)
		}

		// Both should verify successfully
		validHashOnly := k.VerifyMerkleProofBytes(root, txHash[:], proofBytesHashOnly)
		validWithIndices := k.VerifyMerkleProofBytes(root, txHash[:], proofBytesWithIndices)

		require.True(t, validHashOnly, "hash-only format should verify for index %d", idx)
		require.True(t, validWithIndices, "index+hash format should verify for index %d", idx)
	}
}

// TestVerifyMerkleProofBytes_EmptyProof tests single-element tree
func TestVerifyMerkleProofBytes_EmptyProof(t *testing.T) {
	k := &keeper.Keeper{}

	transactions := [][]byte{
		[]byte("only_one"),
	}

	root := keeper.ComputeMerkleRoot(transactions)
	txHash := sha256.Sum256(transactions[0])

	// Empty proof for single element
	emptyProof := []byte{}

	valid := k.VerifyMerkleProofBytes(root, txHash[:], emptyProof)
	require.True(t, valid, "empty proof should verify for single-element tree")
}

// TestVerifyMerkleProofBytes_PowerOfTwo tests trees with power-of-2 elements
func TestVerifyMerkleProofBytes_PowerOfTwo(t *testing.T) {
	k := &keeper.Keeper{}

	// Test with 2, 4, 8, 16 elements
	sizes := []int{2, 4, 8, 16}

	for _, size := range sizes {
		transactions := make([][]byte, size)
		for i := 0; i < size; i++ {
			transactions[i] = []byte{byte(i)}
		}

		root := keeper.ComputeMerkleRoot(transactions)

		// Test middle element
		testIndex := size / 2
		proof, err := keeper.GenerateMerkleProof(transactions, testIndex)
		require.NoError(t, err)

		var proofBytes []byte
		for _, siblingHash := range proof.Proof {
			proofBytes = append(proofBytes, siblingHash...)
		}

		txHash := sha256.Sum256(transactions[testIndex])
		valid := k.VerifyMerkleProofBytes(root, txHash[:], proofBytes)
		require.True(t, valid, "proof should verify for tree size %d", size)
	}
}

// TestVerifyMerkleProofBytes_NonPowerOfTwo tests trees with non-power-of-2 elements
func TestVerifyMerkleProofBytes_NonPowerOfTwo(t *testing.T) {
	k := &keeper.Keeper{}

	// Test with 3, 5, 7, 9 elements
	sizes := []int{3, 5, 7, 9}

	for _, size := range sizes {
		transactions := make([][]byte, size)
		for i := 0; i < size; i++ {
			transactions[i] = []byte{byte(i)}
		}

		root := keeper.ComputeMerkleRoot(transactions)

		// Test last element (often the odd one out)
		testIndex := size - 1
		proof, err := keeper.GenerateMerkleProof(transactions, testIndex)
		require.NoError(t, err)

		var proofBytes []byte
		for _, siblingHash := range proof.Proof {
			proofBytes = append(proofBytes, siblingHash...)
		}

		txHash := sha256.Sum256(transactions[testIndex])
		valid := k.VerifyMerkleProofBytes(root, txHash[:], proofBytes)
		require.True(t, valid, "proof should verify for tree size %d (non-power-of-2)", size)
	}
}

// TestVerifyMerkleProofBytes_AllPositions tests every position in a tree
func TestVerifyMerkleProofBytes_AllPositions(t *testing.T) {
	k := &keeper.Keeper{}

	transactions := [][]byte{
		[]byte("tx0"),
		[]byte("tx1"),
		[]byte("tx2"),
		[]byte("tx3"),
		[]byte("tx4"),
		[]byte("tx5"),
		[]byte("tx6"),
		[]byte("tx7"),
	}

	root := keeper.ComputeMerkleRoot(transactions)

	// Verify proof for every position
	for idx := 0; idx < len(transactions); idx++ {
		proof, err := keeper.GenerateMerkleProof(transactions, idx)
		require.NoError(t, err, "proof generation failed for index %d", idx)

		var proofBytes []byte
		for _, siblingHash := range proof.Proof {
			proofBytes = append(proofBytes, siblingHash...)
		}

		txHash := sha256.Sum256(transactions[idx])
		valid := k.VerifyMerkleProofBytes(root, txHash[:], proofBytes)
		require.True(t, valid, "proof should verify for position %d", idx)
	}
}

// TestVerifyMerkleProofBytes_CrossValidation tests that wrong leaf fails
func TestVerifyMerkleProofBytes_CrossValidation(t *testing.T) {
	k := &keeper.Keeper{}

	transactions := [][]byte{
		[]byte("tx0"),
		[]byte("tx1"),
		[]byte("tx2"),
		[]byte("tx3"),
	}

	root := keeper.ComputeMerkleRoot(transactions)

	// Generate proof for tx1
	proof1, err := keeper.GenerateMerkleProof(transactions, 1)
	require.NoError(t, err)

	var proofBytes1 []byte
	for _, siblingHash := range proof1.Proof {
		proofBytes1 = append(proofBytes1, siblingHash...)
	}

	// Try to verify tx2 with tx1's proof (should fail)
	txHash2 := sha256.Sum256(transactions[2])
	valid := k.VerifyMerkleProofBytes(root, txHash2[:], proofBytes1)
	require.False(t, valid, "proof for tx1 should not verify tx2")

	// Verify tx1 with tx1's proof (should succeed)
	txHash1 := sha256.Sum256(transactions[1])
	valid = k.VerifyMerkleProofBytes(root, txHash1[:], proofBytes1)
	require.True(t, valid, "proof for tx1 should verify tx1")
}
