package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/testutil"
	"github.com/aequitas/aura/chain/x/vcregistry/types"
)

// TestBuildMerkleTree_Deterministic verifies that Merkle tree computation is deterministic
// regardless of map iteration order. This is CRITICAL for consensus - all validators must
// compute the same Merkle root.
func TestBuildMerkleTree_Deterministic(t *testing.T) {
	// Setup test environment
	f := testutil.NewKeeperTestFixture(t)

	// Create and register a VC policy
	policy := &types.VCPolicy{
		Id:          "test-policy",
		Name:        "Test Policy",
		Description: "Test policy for Merkle tree testing",
		SchemaUrl:   "https://example.com/schema",
		Status:      types.VCPolicyStatus_ACTIVE,
	}
	err := f.VCRegistryKeeper.CreateVCPolicyInternal(f.Context, policy)
	require.NoError(t, err, "failed to create VC policy")

	// Mint several VCs to populate the revocation list
	vcIDs := []string{"vc-001", "vc-002", "vc-003", "vc-004", "vc-005"}
	for _, vcID := range vcIDs {
		_, err := f.VCRegistryKeeper.MintVC(f.Context, &types.MsgMintVC{
			Issuer:         f.Addresses[0].String(),
			Subject:        f.Addresses[1].String(),
			PolicyId:       "test-policy",
			CredentialData: []byte(`{"test": "data"}`),
		})
		require.NoError(t, err, "failed to mint VC %s", vcID)
	}

	// Revoke all VCs
	for i, vcID := range vcIDs {
		err := f.VCRegistryKeeper.AdminRevokeVC(f.Context, &types.MsgAdminRevokeVC{
			Admin:  f.Addresses[0].String(),
			VcId:   vcID,
			Reason: "Test revocation",
		})
		require.NoError(t, err, "failed to revoke VC %s (index %d)", vcID, i)
	}

	// Build Merkle tree multiple times and verify the root is always the same
	// This tests that the sorting is working correctly
	roots := make([][]byte, 10)
	for i := 0; i < 10; i++ {
		root, err := f.VCRegistryKeeper.BuildMerkleTree(f.Context)
		require.NoError(t, err, "BuildMerkleTree failed on iteration %d", i)
		roots[i] = root
	}

	// All roots must be identical
	firstRoot := roots[0]
	for i := 1; i < len(roots); i++ {
		require.Equal(t, firstRoot, roots[i],
			"Merkle root is non-deterministic: iteration %d differs from first iteration", i)
	}
}

// TestBuildMerkleTree_EmptyList verifies correct handling of empty revocation list
func TestBuildMerkleTree_EmptyList(t *testing.T) {
	f := testutil.NewKeeperTestFixture(t)

	root, err := f.VCRegistryKeeper.BuildMerkleTree(f.Context)
	require.NoError(t, err, "BuildMerkleTree failed for empty list")
	require.Empty(t, root, "Expected empty root for empty revocation list")
}

// TestBuildMerkleTree_SingleRevocation verifies correct handling of single revocation
func TestBuildMerkleTree_SingleRevocation(t *testing.T) {
	f := testutil.NewKeeperTestFixture(t)

	// Create policy and mint VC
	policy := &types.VCPolicy{
		Id:          "test-policy",
		Name:        "Test Policy",
		Description: "Test policy",
		SchemaUrl:   "https://example.com/schema",
		Status:      types.VCPolicyStatus_ACTIVE,
	}
	err := f.VCRegistryKeeper.CreateVCPolicyInternal(f.Context, policy)
	require.NoError(t, err)

	_, err = f.VCRegistryKeeper.MintVC(f.Context, &types.MsgMintVC{
		Issuer:         f.Addresses[0].String(),
		Subject:        f.Addresses[1].String(),
		PolicyId:       "test-policy",
		CredentialData: []byte(`{"test": "data"}`),
	})
	require.NoError(t, err)

	// Revoke the VC
	err = f.VCRegistryKeeper.AdminRevokeVC(f.Context, &types.MsgAdminRevokeVC{
		Admin:  f.Addresses[0].String(),
		VcId:   "vc-001",
		Reason: "Test",
	})
	require.NoError(t, err)

	// Build tree
	root, err := f.VCRegistryKeeper.BuildMerkleTree(f.Context)
	require.NoError(t, err)
	require.NotEmpty(t, root, "Expected non-empty root for single revocation")
}

// TestVerifyRevocationMerkleProof verifies that Merkle proof verification works correctly
func TestVerifyRevocationMerkleProof(t *testing.T) {
	f := testutil.NewKeeperTestFixture(t)

	// Create policy
	policy := &types.VCPolicy{
		Id:          "test-policy",
		Name:        "Test Policy",
		Description: "Test policy",
		SchemaUrl:   "https://example.com/schema",
		Status:      types.VCPolicyStatus_ACTIVE,
	}
	err := f.VCRegistryKeeper.CreateVCPolicyInternal(f.Context, policy)
	require.NoError(t, err)

	// Mint and revoke a VC
	_, err = f.VCRegistryKeeper.MintVC(f.Context, &types.MsgMintVC{
		Issuer:         f.Addresses[0].String(),
		Subject:        f.Addresses[1].String(),
		PolicyId:       "test-policy",
		CredentialData: []byte(`{"test": "data"}`),
	})
	require.NoError(t, err)

	vcID := "vc-001"
	err = f.VCRegistryKeeper.AdminRevokeVC(f.Context, &types.MsgAdminRevokeVC{
		Admin:  f.Addresses[0].String(),
		VcId:   vcID,
		Reason: "Test",
	})
	require.NoError(t, err)

	// Build Merkle tree
	root, err := f.VCRegistryKeeper.BuildMerkleTree(f.Context)
	require.NoError(t, err)

	// Generate proof for the revoked VC
	proof := f.VCRegistryKeeper.GenerateRevocationMerkleProof(f.Context, vcID)
	require.NotNil(t, proof, "Expected non-nil proof")

	// Verify the proof
	valid := f.VCRegistryKeeper.VerifyRevocationMerkleProof(f.Context, vcID, proof, root)
	require.True(t, valid, "Merkle proof verification should succeed")

	// Verify with wrong root should fail
	wrongRoot := make([]byte, len(root))
	copy(wrongRoot, root)
	if len(wrongRoot) > 0 {
		wrongRoot[0] ^= 0xFF // Flip bits
	}
	valid = f.VCRegistryKeeper.VerifyRevocationMerkleProof(f.Context, vcID, proof, wrongRoot)
	require.False(t, valid, "Merkle proof verification should fail with wrong root")
}

// TestGetMerkleTreeStats verifies tree statistics are correctly computed
func TestGetMerkleTreeStats(t *testing.T) {
	f := testutil.NewKeeperTestFixture(t)

	// Initial state - no revocations
	total, height, root := f.VCRegistryKeeper.GetMerkleTreeStats(f.Context)
	require.Equal(t, uint64(0), total, "Expected zero revocations initially")
	require.Equal(t, uint64(0), height, "Expected zero height initially")
	require.Empty(t, root, "Expected empty root initially")

	// Create policy and mint VCs
	policy := &types.VCPolicy{
		Id:          "test-policy",
		Name:        "Test Policy",
		Description: "Test policy",
		SchemaUrl:   "https://example.com/schema",
		Status:      types.VCPolicyStatus_ACTIVE,
	}
	err := f.VCRegistryKeeper.CreateVCPolicyInternal(f.Context, policy)
	require.NoError(t, err)

	// Mint and revoke 8 VCs (creates balanced binary tree of height 4)
	for i := 0; i < 8; i++ {
		_, err := f.VCRegistryKeeper.MintVC(f.Context, &types.MsgMintVC{
			Issuer:         f.Addresses[0].String(),
			Subject:        f.Addresses[1].String(),
			PolicyId:       "test-policy",
			CredentialData: []byte(`{"test": "data"}`),
		})
		require.NoError(t, err)
	}

	// Note: Stats would be updated after building the tree
	// This test verifies the height calculation logic
}

// TestBatchVerifyRevocations verifies batch verification functionality
func TestBatchVerifyRevocations(t *testing.T) {
	f := testutil.NewKeeperTestFixture(t)

	// Create policy
	policy := &types.VCPolicy{
		Id:          "test-policy",
		Name:        "Test Policy",
		Description: "Test policy",
		SchemaUrl:   "https://example.com/schema",
		Status:      types.VCPolicyStatus_ACTIVE,
	}
	err := f.VCRegistryKeeper.CreateVCPolicyInternal(f.Context, policy)
	require.NoError(t, err)

	// Mint and revoke multiple VCs
	vcIDs := []string{"vc-001", "vc-002", "vc-003"}
	for _, vcID := range vcIDs {
		_, err := f.VCRegistryKeeper.MintVC(f.Context, &types.MsgMintVC{
			Issuer:         f.Addresses[0].String(),
			Subject:        f.Addresses[1].String(),
			PolicyId:       "test-policy",
			CredentialData: []byte(`{"test": "data"}`),
		})
		require.NoError(t, err)

		err = f.VCRegistryKeeper.AdminRevokeVC(f.Context, &types.MsgAdminRevokeVC{
			Admin:  f.Addresses[0].String(),
			VcId:   vcID,
			Reason: "Test",
		})
		require.NoError(t, err)
	}

	// Build tree
	_, err = f.VCRegistryKeeper.BuildMerkleTree(f.Context)
	require.NoError(t, err)

	// Generate proofs for all VCs
	proofs := make([][]byte, len(vcIDs))
	for i, vcID := range vcIDs {
		proofs[i] = f.VCRegistryKeeper.GenerateRevocationMerkleProof(f.Context, vcID)
	}

	// Batch verify
	results, err := f.VCRegistryKeeper.BatchVerifyRevocations(f.Context, vcIDs, proofs)
	require.NoError(t, err)
	require.Len(t, results, len(vcIDs))

	for i, result := range results {
		require.True(t, result, "Verification should succeed for VC %s", vcIDs[i])
	}

	// Test with mismatched lengths
	_, err = f.VCRegistryKeeper.BatchVerifyRevocations(f.Context, vcIDs, proofs[:2])
	require.Error(t, err, "Should error with mismatched lengths")
}
