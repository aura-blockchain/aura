// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/bridge/keeper"
)

// TestConstructTransactionLeaf_Deterministic verifies transaction leaf hashing is deterministic
func TestConstructTransactionLeaf_Deterministic(t *testing.T) {
	k := &keeper.Keeper{}

	// Same inputs should produce same hash
	leaf1 := k.ConstructTransactionLeaf("paw", "tx_abc123", "sender1", "1000000", "upaw")
	leaf2 := k.ConstructTransactionLeaf("paw", "tx_abc123", "sender1", "1000000", "upaw")

	require.Equal(t, leaf1, leaf2, "same inputs should produce same hash")
	require.Len(t, leaf1, 32, "SHA256 hash should be 32 bytes")
}

// TestConstructTransactionLeaf_CaseNormalization verifies case normalization
func TestConstructTransactionLeaf_CaseNormalization(t *testing.T) {
	k := &keeper.Keeper{}

	// Chain and tx hash should be case-insensitive
	leaf1 := k.ConstructTransactionLeaf("paw", "tx_abc123", "sender1", "1000000", "upaw")
	leaf2 := k.ConstructTransactionLeaf("PAW", "TX_ABC123", "sender1", "1000000", "upaw")

	require.Equal(t, leaf1, leaf2, "chain and tx hash should be case-normalized")
}

// TestConstructTransactionLeaf_DifferentInputs verifies different inputs produce different hashes
func TestConstructTransactionLeaf_DifferentInputs(t *testing.T) {
	k := &keeper.Keeper{}

	baseLeaf := k.ConstructTransactionLeaf("paw", "tx123", "sender1", "1000000", "upaw")

	// Different source chain
	leaf2 := k.ConstructTransactionLeaf("xai", "tx123", "sender1", "1000000", "upaw")
	require.NotEqual(t, baseLeaf, leaf2, "different source chain should produce different hash")

	// Different tx hash
	leaf3 := k.ConstructTransactionLeaf("paw", "tx456", "sender1", "1000000", "upaw")
	require.NotEqual(t, baseLeaf, leaf3, "different tx hash should produce different hash")

	// Different sender
	leaf4 := k.ConstructTransactionLeaf("paw", "tx123", "sender2", "1000000", "upaw")
	require.NotEqual(t, baseLeaf, leaf4, "different sender should produce different hash")

	// Different amount
	leaf5 := k.ConstructTransactionLeaf("paw", "tx123", "sender1", "2000000", "upaw")
	require.NotEqual(t, baseLeaf, leaf5, "different amount should produce different hash")

	// Different denom
	leaf6 := k.ConstructTransactionLeaf("paw", "tx123", "sender1", "1000000", "uxai")
	require.NotEqual(t, baseLeaf, leaf6, "different denom should produce different hash")
}

// TestVerifyMerkleProofBytes_Valid tests valid Merkle proof verification
func TestVerifyMerkleProofBytes_Valid(t *testing.T) {
	k := &keeper.Keeper{}

	// Build a simple Merkle tree with 4 transactions
	transactions := [][]byte{
		[]byte("transaction1"),
		[]byte("transaction2"),
		[]byte("transaction3"),
		[]byte("transaction4"),
	}

	// Compute Merkle root
	root := keeper.ComputeMerkleRoot(transactions)
	require.NotNil(t, root)
	require.Len(t, root, 32)

	// Generate proof for transaction at index 1
	proof, err := keeper.GenerateMerkleProof(transactions, 1)
	require.NoError(t, err)
	require.NotNil(t, proof)

	// Convert proof to raw bytes
	var proofBytes []byte
	for _, siblingHash := range proof.Proof {
		proofBytes = append(proofBytes, siblingHash...)
	}

	// Compute the transaction leaf hash
	txHash := sha256.Sum256(transactions[1])

	// Verify the proof
	valid := k.VerifyMerkleProofBytes(root, txHash[:], proofBytes)
	require.True(t, valid, "valid Merkle proof should verify successfully")
}

// TestVerifyMerkleProofBytes_InvalidLeaf tests rejection of wrong leaf
func TestVerifyMerkleProofBytes_InvalidLeaf(t *testing.T) {
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

	var proofBytes []byte
	for _, siblingHash := range proof.Proof {
		proofBytes = append(proofBytes, siblingHash...)
	}

	// Use wrong transaction hash
	wrongTxHash := sha256.Sum256([]byte("wrong_transaction"))

	// Verification should fail
	valid := k.VerifyMerkleProofBytes(root, wrongTxHash[:], proofBytes)
	require.False(t, valid, "Merkle proof with wrong leaf should fail verification")
}

// TestVerifyMerkleProofBytes_InvalidRoot tests rejection of wrong root
func TestVerifyMerkleProofBytes_InvalidRoot(t *testing.T) {
	k := &keeper.Keeper{}

	transactions := [][]byte{
		[]byte("transaction1"),
		[]byte("transaction2"),
		[]byte("transaction3"),
		[]byte("transaction4"),
	}

	keeper.ComputeMerkleRoot(transactions) // Correct root (not used)
	proof, err := keeper.GenerateMerkleProof(transactions, 1)
	require.NoError(t, err)

	var proofBytes []byte
	for _, siblingHash := range proof.Proof {
		proofBytes = append(proofBytes, siblingHash...)
	}

	txHash := sha256.Sum256(transactions[1])

	// Use fake root
	fakeRoot := sha256.Sum256([]byte("fake_merkle_root"))

	// Verification should fail
	valid := k.VerifyMerkleProofBytes(fakeRoot[:], txHash[:], proofBytes)
	require.False(t, valid, "Merkle proof with wrong root should fail verification")
}

// TestVerifyMerkleProofBytes_InvalidProofFormat tests rejection of malformed proof
func TestVerifyMerkleProofBytes_InvalidProofFormat(t *testing.T) {
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

	var proofBytes []byte
	for _, siblingHash := range proof.Proof {
		proofBytes = append(proofBytes, siblingHash...)
	}

	// Corrupt the proof by adding one extra byte (not a multiple of 32)
	corruptedProof := append(proofBytes, byte(0xFF))

	txHash := sha256.Sum256(transactions[1])

	// Verification should fail due to invalid format
	valid := k.VerifyMerkleProofBytes(root, txHash[:], corruptedProof)
	require.False(t, valid, "Merkle proof with invalid format should fail verification")
}

// TestVerifyMerkleProofBytes_EmptyInputs tests rejection of empty inputs
func TestVerifyMerkleProofBytes_EmptyInputs(t *testing.T) {
	k := &keeper.Keeper{}

	validRoot := sha256.Sum256([]byte("root"))
	validLeaf := sha256.Sum256([]byte("leaf"))
	validProof := make([]byte, 32)

	// Empty root
	valid := k.VerifyMerkleProofBytes(nil, validLeaf[:], validProof)
	require.False(t, valid, "empty root should fail verification")

	// Empty leaf
	valid = k.VerifyMerkleProofBytes(validRoot[:], nil, validProof)
	require.False(t, valid, "empty leaf should fail verification")

	// Empty root and leaf
	valid = k.VerifyMerkleProofBytes(nil, nil, validProof)
	require.False(t, valid, "empty root and leaf should fail verification")
}

// TestVerifyMerkleProofBytes_LargeTree tests proof verification with larger tree
func TestVerifyMerkleProofBytes_LargeTree(t *testing.T) {
	k := &keeper.Keeper{}

	// Build tree with 16 transactions
	transactions := make([][]byte, 16)
	for i := 0; i < 16; i++ {
		transactions[i] = []byte{byte(i)}
	}

	root := keeper.ComputeMerkleRoot(transactions)
	require.NotNil(t, root)

	// Test proof for transaction in the middle (index 7)
	proof, err := keeper.GenerateMerkleProof(transactions, 7)
	require.NoError(t, err)

	var proofBytes []byte
	for _, siblingHash := range proof.Proof {
		proofBytes = append(proofBytes, siblingHash...)
	}

	txHash := sha256.Sum256(transactions[7])

	valid := k.VerifyMerkleProofBytes(root, txHash[:], proofBytes)
	require.True(t, valid, "proof for transaction in larger tree should verify")
}

// TestVerifyMerkleProofBytes_SingleTransaction tests single-element tree
func TestVerifyMerkleProofBytes_SingleTransaction(t *testing.T) {
	k := &keeper.Keeper{}

	// Single transaction tree
	transactions := [][]byte{
		[]byte("only_transaction"),
	}

	root := keeper.ComputeMerkleRoot(transactions)
	proof, err := keeper.GenerateMerkleProof(transactions, 0)
	require.NoError(t, err)

	var proofBytes []byte
	for _, siblingHash := range proof.Proof {
		proofBytes = append(proofBytes, siblingHash...)
	}

	txHash := sha256.Sum256(transactions[0])

	valid := k.VerifyMerkleProofBytes(root, txHash[:], proofBytes)
	require.True(t, valid, "proof for single transaction should verify")
}
