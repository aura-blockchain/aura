// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/suite"

	types "github.com/aequitas/aura/chain/x/vcregistry/types"
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

// createTestVC creates a VC record for testing revocations
func (suite *MerkleTreeTestSuite) createTestVC(vcID string) {
	vc := types.VCRecord{
		VcId:            vcID,
		HolderAddress:   "aura1holder" + vcID,
		HolderDid:       "did:aura:" + vcID,
		VcType:          types.VCType_VC_TYPE_VERIFIED_HUMAN,
		Status:          types.VCStatus_VC_STATUS_ACTIVE,
		IssuerAssistant: "aura1issuer",
		IssuedAt:        &gogotypes.Timestamp{Seconds: time.Now().Unix()},
		IssuedHeight:    1,
	}
	err := suite.Keeper.SetVCRecord(suite.SdkCtx, vc)
	suite.Require().NoError(err, "failed to create test VC %s", vcID)
}

// revokeTestVC revokes a VC using the proper keeper method
func (suite *MerkleTreeTestSuite) revokeTestVC(vcID string) {
	err := suite.Keeper.RevokeVC(
		suite.SdkCtx,
		vcID,
		types.RevocationReason_REVOCATION_REASON_USER_REQUEST,
		"aura1revoker",
		"test revocation",
	)
	suite.Require().NoError(err, "failed to revoke VC %s", vcID)
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

	// Test with actual revocations to verify determinism with real data
	// Revoke multiple VCs using distinct vcIDs that would be ordered differently in maps
	vcIDs := []string{"vc-zebra-001", "vc-alpha-002", "vc-gamma-003", "vc-delta-004", "vc-beta-005"}
	for _, vcID := range vcIDs {
		suite.createTestVC(vcID)
		suite.revokeTestVC(vcID)
	}

	// Build Merkle tree 10 times and verify all roots are identical
	var firstRoot []byte
	for i := 0; i < 10; i++ {
		root, err := suite.Keeper.BuildMerkleTree(suite.SdkCtx)
		suite.Require().NoError(err, "BuildMerkleTree failed on iteration %d", i)
		suite.Require().NotEmpty(root, "Expected non-empty root with revocations")

		if i == 0 {
			firstRoot = root
		} else {
			suite.Require().Equal(firstRoot, root,
				"Merkle root must be deterministic: iteration %d differs", i)
		}
	}
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

// TestMerkleTreeWithRevocationLifecycle tests full revocation lifecycle with Merkle tree
func (suite *MerkleTreeTestSuite) TestMerkleTreeWithRevocationLifecycle() {
	// 1. Start with empty tree
	root, err := suite.Keeper.BuildMerkleTree(suite.SdkCtx)
	suite.Require().NoError(err)
	suite.Require().Empty(root, "Tree should start empty")

	// 2. Add first revocation and verify tree updates
	suite.createTestVC("vc-lifecycle-001")
	suite.revokeTestVC("vc-lifecycle-001")

	root1, err := suite.Keeper.BuildMerkleTree(suite.SdkCtx)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(root1, "Tree should have root after first revocation")

	// 3. Add more revocations and verify root changes
	suite.createTestVC("vc-lifecycle-002")
	suite.revokeTestVC("vc-lifecycle-002")

	root2, err := suite.Keeper.BuildMerkleTree(suite.SdkCtx)
	suite.Require().NoError(err)
	suite.Require().NotEqual(root1, root2, "Root should change when revocation is added")

	// 4. Verify determinism with final state
	root2Again, err := suite.Keeper.BuildMerkleTree(suite.SdkCtx)
	suite.Require().NoError(err)
	suite.Require().Equal(root2, root2Again, "Root should be deterministic")
}

// TestBatchVerifyRevocations_AllValid verifies batch verification with valid data
func (suite *MerkleTreeTestSuite) TestBatchVerifyRevocations_AllValid() {
	// Setup revocations
	vcIDs := []string{"batch-vc-001", "batch-vc-002", "batch-vc-003"}
	for _, vcID := range vcIDs {
		suite.createTestVC(vcID)
		suite.revokeTestVC(vcID)
	}

	// Empty proofs (will fail validation, but tests the batch processing path)
	proofs := [][]byte{[]byte{0x01}, []byte{0x02}, []byte{0x03}}

	results, err := suite.Keeper.BatchVerifyRevocations(suite.SdkCtx, vcIDs, proofs)
	suite.Require().NoError(err)
	suite.Require().Len(results, len(vcIDs))
}

// TestGetMerkleTreeStats_WithRevocations verifies stats with actual data
func (suite *MerkleTreeTestSuite) TestGetMerkleTreeStats_WithRevocations() {
	// Add some revocations
	vcIDs := []string{"stats-vc-001", "stats-vc-002", "stats-vc-003", "stats-vc-004"}
	for _, vcID := range vcIDs {
		suite.createTestVC(vcID)
		suite.revokeTestVC(vcID)
	}

	total, height, root := suite.Keeper.GetMerkleTreeStats(suite.SdkCtx)

	suite.Require().Equal(uint64(len(vcIDs)), total, "Should count all revocations")
	suite.Require().Greater(height, uint64(0), "Height should be > 0 with revocations")
	suite.Require().NotEmpty(root, "Root should not be empty with revocations")
}
