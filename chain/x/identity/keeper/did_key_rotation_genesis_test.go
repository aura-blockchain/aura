package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/identity/types"
)

// TestGenesisImportExport_DIDKeyRotations tests that DID key rotations and histories
// are properly imported and exported through genesis
func TestGenesisImportExport_DIDKeyRotations(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create test identities
	did1 := "did:aura:test1"
	did2 := "did:aura:test2"
	owner := "aura1testowner"

	for _, did := range []string{did1, did2} {
		identity := &types.IdentityRecord{
			Did:                 did,
			Address:             owner,
			Status:              types.IdentityStatusActive,
			VerificationMethods: []string{"old-key"},
			CreatedAt:           timestamppb.New(ctx.BlockTime()),
			UpdatedAt:           timestamppb.New(ctx.BlockTime()),
		}
		err := keeper.SetIdentityRecord(ctx, identity)
		require.NoError(t, err)

		// Rotate keys
		_, err = keeper.RotateDIDKey(ctx, did, owner, "new-key", "test rotation")
		require.NoError(t, err)
	}

	// Export genesis
	exportedGenesis, err := keeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NotNil(t, exportedGenesis)

	// Verify exported data contains rotations and histories
	require.Len(t, exportedGenesis.DidKeyRotations, 2, "Should have 2 key rotations")
	require.Len(t, exportedGenesis.DidKeyHistories, 2, "Should have 2 key histories")

	// Verify rotation details
	for _, rotation := range exportedGenesis.DidKeyRotations {
		require.Equal(t, "old-key", rotation.OldVerificationMethod)
		require.Equal(t, "new-key", rotation.NewVerificationMethod)
		require.Equal(t, owner, rotation.InitiatedBy)
		require.Equal(t, types.DIDKeyRotationStatusPending, rotation.Status)
	}

	// Verify history details
	for _, history := range exportedGenesis.DidKeyHistories {
		require.Len(t, history.Entries, 1, "Each history should have 1 entry")
		require.Equal(t, "old-key", history.Entries[0].VerificationMethod)
		require.Equal(t, owner, history.Entries[0].RotatedBy)
	}

	// Create new keeper for import
	keeper2, ctx2 := setupKeeperForTest(t)

	// Import genesis into new keeper
	err = keeper2.InitGenesis(ctx2, exportedGenesis)
	require.NoError(t, err)

	// Verify imported rotations
	for _, did := range []string{did1, did2} {
		rotation, err := keeper2.GetDIDKeyRotation(ctx2, did)
		require.NoError(t, err)
		require.Equal(t, "old-key", rotation.OldVerificationMethod)
		require.Equal(t, "new-key", rotation.NewVerificationMethod)
		require.Equal(t, types.DIDKeyRotationStatusPending, rotation.Status)

		// Verify imported history
		history, err := keeper2.GetDIDKeyHistory(ctx2, did)
		require.NoError(t, err)
		require.Len(t, history.Entries, 1)
		require.Equal(t, "old-key", history.Entries[0].VerificationMethod)
	}

	// Verify identity records were also imported with correct verification methods
	for _, did := range []string{did1, did2} {
		record, err := keeper2.GetIdentityRecord(ctx2, did)
		require.NoError(t, err)
		require.Contains(t, record.VerificationMethods, "new-key")
		require.Contains(t, record.VerificationMethods, "old-key") // Still in grace period
	}
}

// TestGenesisImportExport_CompletedRotations tests genesis with completed rotations
func TestGenesisImportExport_CompletedRotations(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create test identity
	did := "did:aura:test123"
	owner := "aura1testowner"

	identity := &types.IdentityRecord{
		Did:                 did,
		Address:             owner,
		Status:              types.IdentityStatusActive,
		VerificationMethods: []string{"old-key"},
		CreatedAt:           timestamppb.New(ctx.BlockTime()),
		UpdatedAt:           timestamppb.New(ctx.BlockTime()),
	}
	err := keeper.SetIdentityRecord(ctx, identity)
	require.NoError(t, err)

	// Rotate key
	rotation, err := keeper.RotateDIDKey(ctx, did, owner, "new-key", "test rotation")
	require.NoError(t, err)

	// Advance time past grace period
	newTime := rotation.GracePeriodEnd.AsTime().Add(1 * time.Hour)
	ctx = ctx.WithBlockTime(newTime)

	// Complete rotation
	err = keeper.CompleteKeyRotation(ctx, did)
	require.NoError(t, err)

	// Export genesis
	exportedGenesis, err := keeper.ExportGenesis(ctx)
	require.NoError(t, err)

	// Verify completed rotation is exported
	require.Len(t, exportedGenesis.DidKeyRotations, 1)
	require.Equal(t, types.DIDKeyRotationStatusCompleted, exportedGenesis.DidKeyRotations[0].Status)

	// Create new keeper and import
	keeper2, ctx2 := setupKeeperForTest(t)
	err = keeper2.InitGenesis(ctx2, exportedGenesis)
	require.NoError(t, err)

	// Verify completed rotation
	rotation2, err := keeper2.GetDIDKeyRotation(ctx2, did)
	require.NoError(t, err)
	require.Equal(t, types.DIDKeyRotationStatusCompleted, rotation2.Status)

	// Verify old key is no longer in verification methods
	record2, err := keeper2.GetIdentityRecord(ctx2, did)
	require.NoError(t, err)
	require.NotContains(t, record2.VerificationMethods, "old-key")
	require.Contains(t, record2.VerificationMethods, "new-key")
}

// TestGenesisImportExport_MultipleRotations tests genesis with multiple sequential rotations
func TestGenesisImportExport_MultipleRotations(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create test identity
	did := "did:aura:test123"
	owner := "aura1testowner"

	identity := &types.IdentityRecord{
		Did:                 did,
		Address:             owner,
		Status:              types.IdentityStatusActive,
		VerificationMethods: []string{"key-1"},
		CreatedAt:           timestamppb.New(ctx.BlockTime()),
		UpdatedAt:           timestamppb.New(ctx.BlockTime()),
	}
	err := keeper.SetIdentityRecord(ctx, identity)
	require.NoError(t, err)

	// Perform multiple rotations
	keys := []string{"key-2", "key-3", "key-4"}
	for _, newKey := range keys {
		// Rotate to new key
		rotation, err := keeper.RotateDIDKey(ctx, did, owner, newKey, "rotation")
		require.NoError(t, err)

		// Advance time past grace period
		ctx = ctx.WithBlockTime(rotation.GracePeriodEnd.AsTime().Add(1 * time.Hour))

		// Complete rotation
		err = keeper.CompleteKeyRotation(ctx, did)
		require.NoError(t, err)
	}

	// Export genesis
	exportedGenesis, err := keeper.ExportGenesis(ctx)
	require.NoError(t, err)

	// Verify all rotations are exported (last one should be there)
	require.NotEmpty(t, exportedGenesis.DidKeyRotations)

	// Verify history has all keys
	require.Len(t, exportedGenesis.DidKeyHistories, 1)
	require.Len(t, exportedGenesis.DidKeyHistories[0].Entries, 3, "Should have 3 historical entries")

	// Import into new keeper
	keeper2, ctx2 := setupKeeperForTest(t)
	err = keeper2.InitGenesis(ctx2, exportedGenesis)
	require.NoError(t, err)

	// Verify history is preserved
	history, err := keeper2.GetDIDKeyHistory(ctx2, did)
	require.NoError(t, err)
	require.Len(t, history.Entries, 3)

	// Verify current verification method is the last key
	current, err := keeper2.GetCurrentVerificationMethod(ctx2, did)
	require.NoError(t, err)
	require.Equal(t, "key-4", current)
}

// TestGenesisImportExport_EmptyState tests genesis with no rotations
func TestGenesisImportExport_EmptyState(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Export empty genesis
	exportedGenesis, err := keeper.ExportGenesis(ctx)
	require.NoError(t, err)

	// Should have empty slices (may be nil or empty, both are valid)
	require.Empty(t, exportedGenesis.DidKeyRotations, "Should have no key rotations")
	require.Empty(t, exportedGenesis.DidKeyHistories, "Should have no key histories")

	// Import into new keeper should not error
	keeper2, ctx2 := setupKeeperForTest(t)
	err = keeper2.InitGenesis(ctx2, exportedGenesis)
	require.NoError(t, err)

	// Verify no rotations or histories exist after import
	rotations, err := keeper2.GetAllDIDKeyRotations(ctx2)
	require.NoError(t, err)
	require.Empty(t, rotations)

	histories, err := keeper2.GetAllDIDKeyHistories(ctx2)
	require.NoError(t, err)
	require.Empty(t, histories)
}
