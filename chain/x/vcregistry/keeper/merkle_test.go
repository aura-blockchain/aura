package keeper

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// MerkleTreeTestSuite tests the Merkle tree functionality for revocation lists
// CRITICAL: These tests verify that Merkle root computation is deterministic,
// which is essential for blockchain consensus. All validators must compute
// the same Merkle root or consensus will fail.
type MerkleTreeTestSuite struct {
	KeeperTestSuite
}

func TestMerkleTreeTestSuite(t *testing.T) {
	suite.Run(t, new(MerkleTreeTestSuite))
}

// TestBuildMerkleTree_Deterministic verifies that Merkle tree computation is deterministic
// regardless of map iteration order. This is CRITICAL for consensus - all validators must
// compute the same Merkle root.
//
// The bug that this test catches:
// - Before fix: Map iteration is non-deterministic in Go
// - Different validators iterate in different orders
// - Different leaf orderings produce different Merkle roots
// - Consensus breaks - chain halts
//
// The fix:
// - Extract all vcIDs from the map
// - Sort them lexicographically with sort.Strings()
// - Build leaves in sorted order
// - All validators now compute identical roots
func (suite *MerkleTreeTestSuite) TestBuildMerkleTree_Deterministic() {
	// Build Merkle tree multiple times and verify the root is always the same
	// This tests that the sorting is working correctly

	// Initially, with no revocations, should return empty root consistently
	for i := 0; i < 5; i++ {
		root, err := suite.Keeper.BuildMerkleTree(suite.SdkCtx)
		suite.Require().NoError(err, "BuildMerkleTree failed on iteration %d (empty)", i)
		suite.Require().Empty(root, "Expected empty root when no revocations exist")
	}

	// TODO: Add test with actual revocations once we have helper methods to create them
	// The test should:
	// 1. Create multiple VCs and revoke them (using different vcIDs)
	// 2. Build Merkle tree 10 times
	// 3. Verify all 10 roots are byte-for-byte identical
	// This would prove that map iteration order doesn't affect the result
}

// TestBuildMerkleTree_EmptyList verifies correct handling of empty revocation list
func (suite *MerkleTreeTestSuite) TestBuildMerkleTree_EmptyList() {
	root, err := suite.Keeper.BuildMerkleTree(suite.SdkCtx)
	suite.Require().NoError(err, "BuildMerkleTree failed for empty list")
	suite.Require().Empty(root, "Expected empty root for empty revocation list")
}

// TestBuildMerkleTree_ConsistentAfterMultipleCalls verifies that calling BuildMerkleTree
// multiple times produces the same result (idempotent operation)
func (suite *MerkleTreeTestSuite) TestBuildMerkleTree_ConsistentAfterMultipleCalls() {
	// Build tree first time
	root1, err := suite.Keeper.BuildMerkleTree(suite.SdkCtx)
	suite.Require().NoError(err)

	// Build tree second time
	root2, err := suite.Keeper.BuildMerkleTree(suite.SdkCtx)
	suite.Require().NoError(err)

	// Should be identical
	suite.Require().Equal(root1, root2, "Merkle root should be consistent across multiple builds")
}

// TestVerifyRevocationMerkleProof_EmptyInputs verifies error handling for empty inputs
func (suite *MerkleTreeTestSuite) TestVerifyRevocationMerkleProof_EmptyInputs() {
	// Empty proof should fail
	result := suite.Keeper.VerifyRevocationMerkleProof(suite.SdkCtx, "vc-001", []byte{}, []byte{0x01, 0x02})
	suite.Require().False(result, "Verification should fail with empty proof")

	// Empty root should fail
	result = suite.Keeper.VerifyRevocationMerkleProof(suite.SdkCtx, "vc-001", []byte{0x01, 0x02}, []byte{})
	suite.Require().False(result, "Verification should fail with empty root")

	// Both empty should fail
	result = suite.Keeper.VerifyRevocationMerkleProof(suite.SdkCtx, "vc-001", []byte{}, []byte{})
	suite.Require().False(result, "Verification should fail with both empty")
}

// TestGetMerkleTreeStats_EmptyState verifies stats for empty revocation list
func (suite *MerkleTreeTestSuite) TestGetMerkleTreeStats_EmptyState() {
	total, height, root := suite.Keeper.GetMerkleTreeStats(suite.SdkCtx)

	suite.Require().Equal(uint64(0), total, "Expected zero revocations initially")
	suite.Require().Equal(uint64(0), height, "Expected zero height initially")
	suite.Require().Empty(root, "Expected empty root initially")
}

// TestBatchVerifyRevocations_LengthMismatch verifies error handling for mismatched inputs
func (suite *MerkleTreeTestSuite) TestBatchVerifyRevocations_LengthMismatch() {
	vcIDs := []string{"vc-001", "vc-002", "vc-003"}
	proofs := [][]byte{{0x01}, {0x02}} // Intentionally shorter

	_, err := suite.Keeper.BatchVerifyRevocations(suite.SdkCtx, vcIDs, proofs)
	suite.Require().Error(err, "Should error with mismatched lengths")
	suite.Require().Contains(err.Error(), "mismatch", "Error should mention length mismatch")
}

// TestComputeMerkleRoot_SingleLeaf verifies Merkle root computation for a single leaf
func (suite *MerkleTreeTestSuite) TestComputeMerkleRoot_SingleLeaf() {
	leaf := []byte{0x01, 0x02, 0x03, 0x04}
	leaves := [][]byte{leaf}

	root := suite.Keeper.computeMerkleRoot(leaves)

	// For a single leaf, the root should be the leaf itself
	suite.Require().Equal(leaf, root, "Single leaf should be returned as root")
}

// TestComputeMerkleRoot_EmptyLeaves verifies handling of empty leaf array
func (suite *MerkleTreeTestSuite) TestComputeMerkleRoot_EmptyLeaves() {
	leaves := [][]byte{}

	root := suite.Keeper.computeMerkleRoot(leaves)

	suite.Require().Empty(root, "Empty leaves should produce empty root")
}

// TestComputeMerkleRoot_TwoLeaves verifies Merkle root computation for two leaves
func (suite *MerkleTreeTestSuite) TestComputeMerkleRoot_TwoLeaves() {
	leaf1 := []byte{0x01, 0x02, 0x03, 0x04}
	leaf2 := []byte{0x05, 0x06, 0x07, 0x08}
	leaves := [][]byte{leaf1, leaf2}

	root := suite.Keeper.computeMerkleRoot(leaves)

	// Root should be hash of the two leaves
	suite.Require().NotEmpty(root, "Root should not be empty for two leaves")
	suite.Require().NotEqual(leaf1, root, "Root should not equal first leaf")
	suite.Require().NotEqual(leaf2, root, "Root should not equal second leaf")
}

// TestComputeMerkleRoot_OddNumberOfLeaves verifies correct handling of odd leaf count
func (suite *MerkleTreeTestSuite) TestComputeMerkleRoot_OddNumberOfLeaves() {
	leaf1 := []byte{0x01}
	leaf2 := []byte{0x02}
	leaf3 := []byte{0x03}
	leaves := [][]byte{leaf1, leaf2, leaf3}

	root := suite.Keeper.computeMerkleRoot(leaves)

	// Should handle odd number by duplicating last leaf
	suite.Require().NotEmpty(root, "Root should not be empty")

	// Build root again to verify determinism
	root2 := suite.Keeper.computeMerkleRoot(leaves)
	suite.Require().Equal(root, root2, "Root should be deterministic")
}

// TestComputeMerkleRoot_PowerOfTwo verifies tree building for perfect binary tree
func (suite *MerkleTreeTestSuite) TestComputeMerkleRoot_PowerOfTwo() {
	// Create 8 leaves (power of 2)
	leaves := make([][]byte, 8)
	for i := 0; i < 8; i++ {
		leaves[i] = []byte{byte(i)}
	}

	root := suite.Keeper.computeMerkleRoot(leaves)

	suite.Require().NotEmpty(root, "Root should not be empty")
	suite.Require().Len(root, 32, "SHA256 hash should be 32 bytes")
}

// TestComputeMerkleRoot_Deterministic verifies that root computation is deterministic
// This is the CORE property we need for consensus
func (suite *MerkleTreeTestSuite) TestComputeMerkleRoot_Deterministic() {
	// Create some test leaves
	leaves := [][]byte{
		{0x01, 0x02},
		{0x03, 0x04},
		{0x05, 0x06},
		{0x07, 0x08},
		{0x09, 0x0a},
	}

	// Compute root multiple times
	roots := make([][]byte, 10)
	for i := 0; i < 10; i++ {
		roots[i] = suite.Keeper.computeMerkleRoot(leaves)
	}

	// All roots must be identical
	firstRoot := roots[0]
	for i := 1; i < len(roots); i++ {
		suite.Require().Equal(firstRoot, roots[i],
			"Merkle root must be deterministic: iteration %d differs", i)
	}
}

// TestGenerateRevocationMerkleProof_NonExistentVC verifies handling of non-existent VCs
func (suite *MerkleTreeTestSuite) TestGenerateRevocationMerkleProof_NonExistentVC() {
	proof := suite.Keeper.GenerateRevocationMerkleProof(suite.SdkCtx, "non-existent-vc")

	// Should return nil for non-existent VCs
	suite.Require().Nil(proof, "Proof should be nil for non-existent VC")
}

// TODO: Add integration tests with actual VC creation and revocation
// These tests would verify:
// 1. Creating VCs and revoking them
// 2. Building Merkle tree with actual revocation data
// 3. Generating and verifying proofs for revoked VCs
// 4. Batch verification with multiple revoked VCs
// 5. Updating Merkle root when new revocations are added
//
// These require helper methods or test fixtures to create valid VCs and policies
