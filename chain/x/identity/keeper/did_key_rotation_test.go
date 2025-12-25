// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/identity/types"
)

// TestRotateDIDKey tests basic key rotation functionality
func TestRotateDIDKey(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create test identity
	did := "did:aura:test123"
	owner := "aura1testowner"
	oldKey := "old-verification-method"

	now := ctx.BlockTime()
	identity := &types.IdentityRecord{
		Did:                 did,
		Address:             owner,
		Status:              types.IdentityStatusActive,
		VerificationMethods: []string{oldKey},
		CreatedAt:           now,
		UpdatedAt:           &now,
	}

	err := keeper.SetIdentityRecord(ctx, identity)
	require.NoError(t, err)

	// Rotate key
	newKey := "new-verification-method"
	reason := "Key compromised - rotating for security"

	rotation, err := keeper.RotateDIDKey(ctx, did, owner, newKey, reason)
	require.NoError(t, err)
	require.NotNil(t, rotation)

	// Verify rotation details
	require.Equal(t, did, rotation.Did)
	require.Equal(t, oldKey, rotation.OldVerificationMethod)
	require.Equal(t, newKey, rotation.NewVerificationMethod)
	require.Equal(t, owner, rotation.InitiatedBy)
	require.Equal(t, reason, rotation.Reason)
	require.Equal(t, types.DIDKeyRotationStatusPending, rotation.Status)

	// Verify grace period is set correctly
	params, err := keeper.GetParams(ctx)
	require.NoError(t, err)
	expectedGracePeriod := ctx.BlockTime().Add(time.Duration(params.Change.KeyRotationGracePeriodSeconds) * time.Second)
	require.Equal(t, expectedGracePeriod.Unix(), rotation.GracePeriodEnd.Unix())

	// Verify identity record updated
	updatedIdentity, err := keeper.GetIdentityRecord(ctx, did)
	require.NoError(t, err)
	require.Contains(t, updatedIdentity.VerificationMethods, newKey)
	require.Contains(t, updatedIdentity.VerificationMethods, oldKey) // Old key still present during grace period
}

// TestRotateDIDKey_Unauthorized tests that unauthorized users cannot rotate keys
func TestRotateDIDKey_Unauthorized(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create test identity
	did := "did:aura:test123"
	owner := "aura1testowner"
	unauthorized := "aura1unauthorized"

	now := ctx.BlockTime()
	identity := &types.IdentityRecord{
		Did:                 did,
		Address:             owner,
		Status:              types.IdentityStatusActive,
		VerificationMethods: []string{"old-key"},
		CreatedAt:           now,
		UpdatedAt:           &now,
	}

	err := keeper.SetIdentityRecord(ctx, identity)
	require.NoError(t, err)

	// Try to rotate key as unauthorized user
	_, err = keeper.RotateDIDKey(ctx, did, unauthorized, "new-key", "test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not authorized")
}

// TestRotateDIDKey_RotationInProgress tests that concurrent rotations are prevented
func TestRotateDIDKey_RotationInProgress(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create test identity
	did := "did:aura:test123"
	owner := "aura1testowner"

	now := ctx.BlockTime()
	identity := &types.IdentityRecord{
		Did:                 did,
		Address:             owner,
		Status:              types.IdentityStatusActive,
		VerificationMethods: []string{"old-key"},
		CreatedAt:           now,
		UpdatedAt:           &now,
	}

	err := keeper.SetIdentityRecord(ctx, identity)
	require.NoError(t, err)

	// First rotation
	_, err = keeper.RotateDIDKey(ctx, did, owner, "new-key-1", "first rotation")
	require.NoError(t, err)

	// Try second rotation while first is pending
	_, err = keeper.RotateDIDKey(ctx, did, owner, "new-key-2", "second rotation")
	require.Error(t, err)
	require.Contains(t, err.Error(), "rotation already in progress")
}

// TestRotateDIDKey_ErasedIdentity tests that keys cannot be rotated for erased identities
func TestRotateDIDKey_ErasedIdentity(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create test identity
	did := "did:aura:test123"
	owner := "aura1testowner"

	now := ctx.BlockTime()
	identity := &types.IdentityRecord{
		Did:                 did,
		Address:             owner,
		Status:              types.IdentityStatusErased,
		Erased:              true,
		ErasedAt:            &now,
		VerificationMethods: []string{"old-key"},
		CreatedAt:           now,
		UpdatedAt:           &now,
	}

	err := keeper.SetIdentityRecord(ctx, identity)
	require.NoError(t, err)

	// Try to rotate key
	_, err = keeper.RotateDIDKey(ctx, did, owner, "new-key", "test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "erased identity")
}

// TestValidateDIDKey tests key validation during grace period
func TestValidateDIDKey(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create test identity
	did := "did:aura:test123"
	owner := "aura1testowner"
	oldKey := "old-key"
	newKey := "new-key"

	now := ctx.BlockTime()
	identity := &types.IdentityRecord{
		Did:                 did,
		Address:             owner,
		Status:              types.IdentityStatusActive,
		VerificationMethods: []string{oldKey},
		CreatedAt:           now,
		UpdatedAt:           &now,
	}

	err := keeper.SetIdentityRecord(ctx, identity)
	require.NoError(t, err)

	// Rotate key
	_, err = keeper.RotateDIDKey(ctx, did, owner, newKey, "test rotation")
	require.NoError(t, err)

	// Both keys should be valid during grace period
	err = keeper.ValidateDIDKey(ctx, did, newKey)
	require.NoError(t, err, "New key should be valid")

	err = keeper.ValidateDIDKey(ctx, did, oldKey)
	require.NoError(t, err, "Old key should be valid during grace period")

	// Invalid key should fail
	err = keeper.ValidateDIDKey(ctx, did, "invalid-key")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not valid")
}

// TestValidateDIDKey_AfterGracePeriod tests key validation after grace period expires
func TestValidateDIDKey_AfterGracePeriod(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create test identity
	did := "did:aura:test123"
	owner := "aura1testowner"
	oldKey := "old-key"
	newKey := "new-key"

	now := ctx.BlockTime()
	identity := &types.IdentityRecord{
		Did:                 did,
		Address:             owner,
		Status:              types.IdentityStatusActive,
		VerificationMethods: []string{oldKey},
		CreatedAt:           now,
		UpdatedAt:           &now,
	}

	err := keeper.SetIdentityRecord(ctx, identity)
	require.NoError(t, err)

	// Rotate key
	rotation, err := keeper.RotateDIDKey(ctx, did, owner, newKey, "test rotation")
	require.NoError(t, err)

	// Advance time past grace period
	newTime := rotation.GracePeriodEnd.Add(1 * time.Hour)
	ctx = ctx.WithBlockTime(newTime)

	// Complete rotation
	err = keeper.CompleteKeyRotation(ctx, did)
	require.NoError(t, err)

	// New key should still be valid
	err = keeper.ValidateDIDKey(ctx, did, newKey)
	require.NoError(t, err, "New key should be valid")

	// Old key should no longer be valid
	err = keeper.ValidateDIDKey(ctx, did, oldKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not valid")
}

// TestCompleteKeyRotation tests completing a key rotation
func TestCompleteKeyRotation(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create test identity
	did := "did:aura:test123"
	owner := "aura1testowner"
	oldKey := "old-key"
	newKey := "new-key"

	now := ctx.BlockTime()
	identity := &types.IdentityRecord{
		Did:                 did,
		Address:             owner,
		Status:              types.IdentityStatusActive,
		VerificationMethods: []string{oldKey},
		CreatedAt:           now,
		UpdatedAt:           &now,
	}

	err := keeper.SetIdentityRecord(ctx, identity)
	require.NoError(t, err)

	// Rotate key
	rotation, err := keeper.RotateDIDKey(ctx, did, owner, newKey, "test rotation")
	require.NoError(t, err)

	// Try to complete before grace period ends
	err = keeper.CompleteKeyRotation(ctx, did)
	require.Error(t, err)
	require.Contains(t, err.Error(), "grace period not yet ended")

	// Advance time past grace period
	newTime := rotation.GracePeriodEnd.Add(1 * time.Hour)
	ctx = ctx.WithBlockTime(newTime)

	// Complete rotation
	err = keeper.CompleteKeyRotation(ctx, did)
	require.NoError(t, err)

	// Verify rotation is completed
	updatedRotation, err := keeper.GetDIDKeyRotation(ctx, did)
	require.NoError(t, err)
	require.Equal(t, types.DIDKeyRotationStatusCompleted, updatedRotation.Status)

	// Verify old key removed from identity record
	updatedIdentity, err := keeper.GetIdentityRecord(ctx, did)
	require.NoError(t, err)
	require.NotContains(t, updatedIdentity.VerificationMethods, oldKey)
	require.Contains(t, updatedIdentity.VerificationMethods, newKey)
}

// TestDIDKeyHistory tests key history tracking
func TestDIDKeyHistory(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create test identity
	did := "did:aura:test123"
	owner := "aura1testowner"
	key1 := "key-1"

	now := ctx.BlockTime()
	identity := &types.IdentityRecord{
		Did:                 did,
		Address:             owner,
		Status:              types.IdentityStatusActive,
		VerificationMethods: []string{key1},
		CreatedAt:           now,
		UpdatedAt:           &now,
	}

	err := keeper.SetIdentityRecord(ctx, identity)
	require.NoError(t, err)

	// Rotate to key 2
	key2 := "key-2"
	_, err = keeper.RotateDIDKey(ctx, did, owner, key2, "first rotation")
	require.NoError(t, err)

	// Get history
	history, err := keeper.GetDIDKeyHistory(ctx, did)
	require.NoError(t, err)
	require.NotNil(t, history)
	require.Len(t, history.Entries, 1) // Should have one entry for key1

	// Verify history entry
	require.Equal(t, key1, history.Entries[0].VerificationMethod)
	require.Equal(t, owner, history.Entries[0].RotatedBy)
	require.Equal(t, "first rotation", history.Entries[0].RotationReason)
	require.NotNil(t, history.Entries[0].ActiveUntil)

	// Complete first rotation
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(25 * time.Hour))
	err = keeper.CompleteKeyRotation(ctx, did)
	require.NoError(t, err)

	// Rotate to key 3
	key3 := "key-3"
	_, err = keeper.RotateDIDKey(ctx, did, owner, key3, "second rotation")
	require.NoError(t, err)

	// Get updated history
	history, err = keeper.GetDIDKeyHistory(ctx, did)
	require.NoError(t, err)
	require.Len(t, history.Entries, 2) // Should have two entries now

	// Verify second entry
	require.Equal(t, key2, history.Entries[1].VerificationMethod)
	require.Equal(t, "second rotation", history.Entries[1].RotationReason)
}

// TestProcessExpiredGracePeriods tests automatic completion of expired rotations
func TestProcessExpiredGracePeriods(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create multiple identities with rotations
	did1 := "did:aura:test1"
	did2 := "did:aura:test2"
	did3 := "did:aura:test3"
	owner := "aura1testowner"

	now := ctx.BlockTime()
	for _, did := range []string{did1, did2, did3} {
		identity := &types.IdentityRecord{
			Did:                 did,
			Address:             owner,
			Status:              types.IdentityStatusActive,
			VerificationMethods: []string{"old-key"},
			CreatedAt:           now,
			UpdatedAt:           &now,
		}

		err := keeper.SetIdentityRecord(ctx, identity)
		require.NoError(t, err)

		_, err = keeper.RotateDIDKey(ctx, did, owner, "new-key", "test")
		require.NoError(t, err)
	}

	// Advance time past grace period
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(25 * time.Hour))

	// Process expired grace periods
	err := keeper.ProcessExpiredGracePeriods(ctx)
	require.NoError(t, err)

	// Verify all rotations completed
	for _, did := range []string{did1, did2, did3} {
		rotation, err := keeper.GetDIDKeyRotation(ctx, did)
		require.NoError(t, err)
		require.Equal(t, types.DIDKeyRotationStatusCompleted, rotation.Status)

		// Verify old key removed
		identity, err := keeper.GetIdentityRecord(ctx, did)
		require.NoError(t, err)
		require.NotContains(t, identity.VerificationMethods, "old-key")
		require.Contains(t, identity.VerificationMethods, "new-key")
	}
}

// TestIsKeyInGracePeriod tests checking if a key is in grace period
func TestIsKeyInGracePeriod(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create test identity
	did := "did:aura:test123"
	owner := "aura1testowner"
	oldKey := "old-key"
	newKey := "new-key"

	now := ctx.BlockTime()
	identity := &types.IdentityRecord{
		Did:                 did,
		Address:             owner,
		Status:              types.IdentityStatusActive,
		VerificationMethods: []string{oldKey},
		CreatedAt:           now,
		UpdatedAt:           &now,
	}

	err := keeper.SetIdentityRecord(ctx, identity)
	require.NoError(t, err)

	// Before rotation, old key not in grace period
	inGrace := keeper.IsKeyInGracePeriod(ctx, did, oldKey)
	require.False(t, inGrace)

	// Rotate key
	rotation, err := keeper.RotateDIDKey(ctx, did, owner, newKey, "test")
	require.NoError(t, err)

	// Old key should now be in grace period
	inGrace = keeper.IsKeyInGracePeriod(ctx, did, oldKey)
	require.True(t, inGrace)

	// New key should not be in grace period (it's the current key)
	inGrace = keeper.IsKeyInGracePeriod(ctx, did, newKey)
	require.False(t, inGrace)

	// After grace period ends
	ctx = ctx.WithBlockTime(rotation.GracePeriodEnd.Add(1 * time.Hour))
	inGrace = keeper.IsKeyInGracePeriod(ctx, did, oldKey)
	require.False(t, inGrace)
}

// TestGetCurrentVerificationMethod tests getting the current verification method
func TestGetCurrentVerificationMethod(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create test identity
	did := "did:aura:test123"
	owner := "aura1testowner"
	key1 := "key-1"

	now := ctx.BlockTime()
	identity := &types.IdentityRecord{
		Did:                 did,
		Address:             owner,
		Status:              types.IdentityStatusActive,
		VerificationMethods: []string{key1},
		CreatedAt:           now,
		UpdatedAt:           &now,
	}

	err := keeper.SetIdentityRecord(ctx, identity)
	require.NoError(t, err)

	// Get current method
	current, err := keeper.GetCurrentVerificationMethod(ctx, did)
	require.NoError(t, err)
	require.Equal(t, key1, current)

	// Rotate key
	key2 := "key-2"
	_, err = keeper.RotateDIDKey(ctx, did, owner, key2, "test")
	require.NoError(t, err)

	// Current method should now be key2
	current, err = keeper.GetCurrentVerificationMethod(ctx, did)
	require.NoError(t, err)
	require.Equal(t, key2, current)
}

// TestMultipleRotations tests multiple consecutive rotations
func TestMultipleRotations(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create test identity
	did := "did:aura:test123"
	owner := "aura1testowner"
	key1 := "key-1"

	now := ctx.BlockTime()
	identity := &types.IdentityRecord{
		Did:                 did,
		Address:             owner,
		Status:              types.IdentityStatusActive,
		VerificationMethods: []string{key1},
		CreatedAt:           now,
		UpdatedAt:           &now,
	}

	err := keeper.SetIdentityRecord(ctx, identity)
	require.NoError(t, err)

	keys := []string{"key-2", "key-3", "key-4"}

	for i, newKey := range keys {
		// Rotate to new key
		_, err := keeper.RotateDIDKey(ctx, did, owner, newKey, "rotation "+string(rune(i+1)))
		require.NoError(t, err)

		// Advance time past grace period
		ctx = ctx.WithBlockTime(ctx.BlockTime().Add(25 * time.Hour))

		// Complete rotation
		err = keeper.CompleteKeyRotation(ctx, did)
		require.NoError(t, err)

		// Verify current key
		current, err := keeper.GetCurrentVerificationMethod(ctx, did)
		require.NoError(t, err)
		require.Equal(t, newKey, current)
	}

	// Verify history has all rotations
	history, err := keeper.GetDIDKeyHistory(ctx, did)
	require.NoError(t, err)
	require.Len(t, history.Entries, 3) // key-1, key-2, key-3 all rotated
}
