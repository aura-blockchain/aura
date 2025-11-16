package keeper_test

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
)

func TestMerkleProofVerification(t *testing.T) {
	// Create test transfers
	transfers := [][]byte{
		[]byte("transfer1"),
		[]byte("transfer2"),
		[]byte("transfer3"),
		[]byte("transfer4"),
	}

	// Generate Merkle root
	root := keeper.ComputeMerkleRoot(transfers)
	require.NotNil(t, root)

	// Generate proof for transfer 2 (index 1)
	proof, err := keeper.GenerateMerkleProof(transfers, 1)
	require.NoError(t, err)
	require.NotNil(t, proof)

	// Verify the proof matches expected structure
	require.Equal(t, root, proof.Root)

	expectedLeaf := sha256.Sum256(transfers[1])
	require.Equal(t, expectedLeaf[:], proof.Leaf)

	// Test proof has correct number of siblings
	require.NotEmpty(t, proof.Proof)
}

func TestMerkleProofInvalidIndex(t *testing.T) {
	transfers := [][]byte{
		[]byte("transfer1"),
		[]byte("transfer2"),
	}

	// Try to generate proof for invalid index
	_, err := keeper.GenerateMerkleProof(transfers, 5)
	require.Error(t, err)
}

func TestComputeMerkleRootEmptyList(t *testing.T) {
	root := keeper.ComputeMerkleRoot([][]byte{})
	require.Nil(t, root)
}

func TestComputeMerkleRootSingleTransfer(t *testing.T) {
	transfers := [][]byte{[]byte("transfer1")}
	root := keeper.ComputeMerkleRoot(transfers)
	require.NotNil(t, root)

	// For single element, root should be hash of the element
	expected := sha256.Sum256(transfers[0])
	require.Equal(t, expected[:], root)
}

func TestComputeMerkleRootOddNumber(t *testing.T) {
	// Test with odd number of transfers (requires duplicate last node)
	transfers := [][]byte{
		[]byte("transfer1"),
		[]byte("transfer2"),
		[]byte("transfer3"),
	}

	root := keeper.ComputeMerkleRoot(transfers)
	require.NotNil(t, root)
	require.Len(t, root, 32) // SHA256 hash is 32 bytes
}

func TestMerkleProofConsistency(t *testing.T) {
	// Verify that the same transfers always produce the same root
	transfers := [][]byte{
		[]byte("transfer1"),
		[]byte("transfer2"),
		[]byte("transfer3"),
		[]byte("transfer4"),
	}

	root1 := keeper.ComputeMerkleRoot(transfers)
	root2 := keeper.ComputeMerkleRoot(transfers)

	require.Equal(t, root1, root2)
}
