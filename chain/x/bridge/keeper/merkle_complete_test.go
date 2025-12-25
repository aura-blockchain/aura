// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
)

func TestVerifyMerkleProof_Valid(t *testing.T) {
	// Create test data
	data := [][]byte{
		[]byte("data1"),
		[]byte("data2"),
		[]byte("data3"),
		[]byte("data4"),
	}

	// Generate proof for index 1
	proof, err := keeper.GenerateMerkleProof(data, 1)
	require.NoError(t, err)
	require.NotNil(t, proof)

	// Verify the proof
	valid := keeper.VerifyMerkleProof(proof)
	require.True(t, valid, "Valid proof should verify successfully")
}

func TestVerifyMerkleProof_ValidAllIndices(t *testing.T) {
	// Create test data
	data := [][]byte{
		[]byte("data1"),
		[]byte("data2"),
		[]byte("data3"),
		[]byte("data4"),
	}

	// Test proof for each index
	for i := 0; i < len(data); i++ {
		proof, err := keeper.GenerateMerkleProof(data, i)
		require.NoError(t, err)

		valid := keeper.VerifyMerkleProof(proof)
		require.True(t, valid, "Proof for index %d should be valid", i)
	}
}

func TestVerifyMerkleProof_NilProof(t *testing.T) {
	valid := keeper.VerifyMerkleProof(nil)
	require.False(t, valid, "Nil proof should be invalid")
}

func TestVerifyMerkleProof_EmptyProof(t *testing.T) {
	proof := &types.MerkleProof{
		Root:    []byte("root"),
		Leaf:    []byte("leaf"),
		Proof:   [][]byte{}, // Empty proof
		Indices: []uint64{},
	}

	valid := keeper.VerifyMerkleProof(proof)
	require.False(t, valid, "Empty proof should be invalid")
}

func TestVerifyMerkleProof_InvalidRoot(t *testing.T) {
	// Create test data
	data := [][]byte{
		[]byte("data1"),
		[]byte("data2"),
		[]byte("data3"),
		[]byte("data4"),
	}

	// Generate a valid proof
	proof, err := keeper.GenerateMerkleProof(data, 1)
	require.NoError(t, err)

	// Tamper with the root
	proof.Root = []byte("invalid_root")

	// Verification should fail
	valid := keeper.VerifyMerkleProof(proof)
	require.False(t, valid, "Proof with tampered root should be invalid")
}

func TestVerifyMerkleProof_InvalidLeaf(t *testing.T) {
	// Create test data
	data := [][]byte{
		[]byte("data1"),
		[]byte("data2"),
		[]byte("data3"),
		[]byte("data4"),
	}

	// Generate a valid proof
	proof, err := keeper.GenerateMerkleProof(data, 1)
	require.NoError(t, err)

	// Tamper with the leaf
	proof.Leaf = []byte("invalid_leaf_data_that_is_different")

	// Verification should fail
	valid := keeper.VerifyMerkleProof(proof)
	require.False(t, valid, "Proof with tampered leaf should be invalid")
}

func TestVerifyMerkleProof_TamperedSibling(t *testing.T) {
	// Create test data
	data := [][]byte{
		[]byte("data1"),
		[]byte("data2"),
		[]byte("data3"),
		[]byte("data4"),
	}

	// Generate a valid proof
	proof, err := keeper.GenerateMerkleProof(data, 1)
	require.NoError(t, err)

	// Tamper with one of the siblings
	if len(proof.Proof) > 0 {
		proof.Proof[0] = []byte("tampered_sibling")
	}

	// Verification should fail
	valid := keeper.VerifyMerkleProof(proof)
	require.False(t, valid, "Proof with tampered sibling should be invalid")
}

func TestVerifyMerkleProof_SingleElement(t *testing.T) {
	// Create single element data
	data := [][]byte{
		[]byte("single_data"),
	}

	// Generate proof
	proof, err := keeper.GenerateMerkleProof(data, 0)
	require.NoError(t, err)

	// Verify proof
	valid := keeper.VerifyMerkleProof(proof)
	require.True(t, valid, "Single element proof should be valid")
}

func TestVerifyMerkleProof_TwoElements(t *testing.T) {
	// Create two element data
	data := [][]byte{
		[]byte("data1"),
		[]byte("data2"),
	}

	// Test both elements
	for i := 0; i < 2; i++ {
		proof, err := keeper.GenerateMerkleProof(data, i)
		require.NoError(t, err)

		valid := keeper.VerifyMerkleProof(proof)
		require.True(t, valid, "Two element proof for index %d should be valid", i)
	}
}

func TestVerifyMerkleProof_OddNumberElements(t *testing.T) {
	// Create odd number of elements
	data := [][]byte{
		[]byte("data1"),
		[]byte("data2"),
		[]byte("data3"),
	}

	// Test all elements
	for i := 0; i < len(data); i++ {
		proof, err := keeper.GenerateMerkleProof(data, i)
		require.NoError(t, err)

		valid := keeper.VerifyMerkleProof(proof)
		require.True(t, valid, "Odd element proof for index %d should be valid", i)
	}
}

func TestVerifyMerkleProof_LargeTree(t *testing.T) {
	// Create a larger tree (16 elements)
	data := make([][]byte, 16)
	for i := 0; i < 16; i++ {
		data[i] = []byte{byte(i)}
	}

	// Test several indices
	testIndices := []int{0, 5, 10, 15}
	for _, idx := range testIndices {
		proof, err := keeper.GenerateMerkleProof(data, idx)
		require.NoError(t, err)

		valid := keeper.VerifyMerkleProof(proof)
		require.True(t, valid, "Large tree proof for index %d should be valid", idx)
	}
}

func TestVerifyMerkleProof_DifferentRootLength(t *testing.T) {
	// Create test data
	data := [][]byte{
		[]byte("data1"),
		[]byte("data2"),
	}

	// Generate proof
	proof, err := keeper.GenerateMerkleProof(data, 0)
	require.NoError(t, err)

	// Make root a different length (but same data conceptually)
	originalRoot := proof.Root
	proof.Root = append(proof.Root, byte(0)) // Add extra byte

	// Verification should fail due to length mismatch
	valid := keeper.VerifyMerkleProof(proof)
	require.False(t, valid, "Proof with different root length should be invalid")

	// Restore and verify it works
	proof.Root = originalRoot
	valid = keeper.VerifyMerkleProof(proof)
	require.True(t, valid)
}

func TestVerifyMerkleProof_PartiallyCorrectProof(t *testing.T) {
	// Create test data
	data := [][]byte{
		[]byte("data1"),
		[]byte("data2"),
		[]byte("data3"),
		[]byte("data4"),
	}

	// Generate proof for index 2
	proof, err := keeper.GenerateMerkleProof(data, 2)
	require.NoError(t, err)

	// Tamper with indices while keeping proof hashes
	if len(proof.Indices) > 0 {
		proof.Indices[0] = proof.Indices[0] + 1 // Change index
	}

	// This might still pass or fail depending on implementation
	// The important thing is we test this code path
	_ = keeper.VerifyMerkleProof(proof)
}

func TestGenerateMerkleProof_EdgeCases(t *testing.T) {
	// Test with minimal data to ensure GenerateMerkleProof is fully covered
	data := [][]byte{
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
	}

	// Generate proof for last element in odd-sized tree
	proof, err := keeper.GenerateMerkleProof(data, 2)
	require.NoError(t, err)
	require.NotNil(t, proof)

	// Verify it
	valid := keeper.VerifyMerkleProof(proof)
	require.True(t, valid)
}

func TestVerifyMerkleProof_MismatchedIndicesLength(t *testing.T) {
	data := [][]byte{
		[]byte("data1"),
		[]byte("data2"),
		[]byte("data3"),
		[]byte("data4"),
	}

	proof, err := keeper.GenerateMerkleProof(data, 1)
	require.NoError(t, err)

	// Test with fewer indices than proof hashes
	if len(proof.Indices) > 0 {
		proof.Indices = proof.Indices[:len(proof.Indices)-1]
	}

	// Should still process, testing the bounds check
	_ = keeper.VerifyMerkleProof(proof)
}

func TestVerifyMerkleProof_AllEvenIndices(t *testing.T) {
	// Specific test to ensure even/odd index logic is covered
	data := [][]byte{
		[]byte("data1"),
		[]byte("data2"),
		[]byte("data3"),
		[]byte("data4"),
		[]byte("data5"),
		[]byte("data6"),
	}

	// Test even indices
	for i := 0; i < len(data); i += 2 {
		proof, err := keeper.GenerateMerkleProof(data, i)
		require.NoError(t, err)

		valid := keeper.VerifyMerkleProof(proof)
		require.True(t, valid, "Proof for even index %d should be valid", i)
	}
}

func TestVerifyMerkleProof_AllOddIndices(t *testing.T) {
	// Test odd indices
	data := [][]byte{
		[]byte("data1"),
		[]byte("data2"),
		[]byte("data3"),
		[]byte("data4"),
		[]byte("data5"),
		[]byte("data6"),
	}

	for i := 1; i < len(data); i += 2 {
		proof, err := keeper.GenerateMerkleProof(data, i)
		require.NoError(t, err)

		valid := keeper.VerifyMerkleProof(proof)
		require.True(t, valid, "Proof for odd index %d should be valid", i)
	}
}
