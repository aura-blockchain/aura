// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// TestDelegationMarshaling verifies that delegation storage properly preserves ALL fields
// This is a critical test for the data corruption bug fix
func TestDelegationMarshaling(t *testing.T) {
	ctx, keeper := setupConfKeeper(t)

	// Create a delegation with all fields populated
	originalDelegation := &ScoreDelegation{
		DelegationID:   "test-delegation-001",
		Delegator:      "aura1delegator",
		Delegate:       "aura1delegate",
		DelegatedScore: 5000,
		DelegationType: DelegationTypeValidation,
		StartHeight:    100,
		EndHeight:      200,
		Active:         true,
		Revocable:      true,
		RewardSharePct: 25,
		CreatedHeight:  100,
		LastUpdated:    100,
	}

	// Store the delegation
	err := keeper.storeDelegation(ctx, originalDelegation)
	require.NoError(t, err, "failed to store delegation")

	// Retrieve the delegation
	retrievedDelegation, found := keeper.getDelegation(ctx, "test-delegation-001")
	require.True(t, found, "delegation not found after storage")
	require.NotNil(t, retrievedDelegation, "retrieved delegation is nil")

	// Verify ALL fields are preserved (not just DelegationID!)
	require.Equal(t, originalDelegation.DelegationID, retrievedDelegation.DelegationID, "DelegationID mismatch")
	require.Equal(t, originalDelegation.Delegator, retrievedDelegation.Delegator, "Delegator mismatch")
	require.Equal(t, originalDelegation.Delegate, retrievedDelegation.Delegate, "Delegate mismatch")
	require.Equal(t, originalDelegation.DelegatedScore, retrievedDelegation.DelegatedScore, "DelegatedScore mismatch")
	require.Equal(t, originalDelegation.DelegationType, retrievedDelegation.DelegationType, "DelegationType mismatch")
	require.Equal(t, originalDelegation.StartHeight, retrievedDelegation.StartHeight, "StartHeight mismatch")
	require.Equal(t, originalDelegation.EndHeight, retrievedDelegation.EndHeight, "EndHeight mismatch")
	require.Equal(t, originalDelegation.Active, retrievedDelegation.Active, "Active mismatch")
	require.Equal(t, originalDelegation.Revocable, retrievedDelegation.Revocable, "Revocable mismatch")
	require.Equal(t, originalDelegation.RewardSharePct, retrievedDelegation.RewardSharePct, "RewardSharePct mismatch")
	require.Equal(t, originalDelegation.CreatedHeight, retrievedDelegation.CreatedHeight, "CreatedHeight mismatch")
	require.Equal(t, originalDelegation.LastUpdated, retrievedDelegation.LastUpdated, "LastUpdated mismatch")
}

// TestDelegationExpirationIndex verifies that the expiration index is properly maintained
func TestDelegationExpirationIndex(t *testing.T) {
	ctx, keeper := setupConfKeeper(t)

	// Create a delegation that expires at height 500
	delegation := &ScoreDelegation{
		DelegationID:   "expiring-delegation",
		Delegator:      "aura1delegator",
		Delegate:       "aura1delegate",
		DelegatedScore: 1000,
		DelegationType: DelegationTypeValidation,
		StartHeight:    100,
		EndHeight:      500,
		Active:         true,
		Revocable:      false,
		RewardSharePct: 10,
		CreatedHeight:  100,
		LastUpdated:    100,
	}

	// Store the delegation
	err := keeper.storeDelegation(ctx, delegation)
	require.NoError(t, err, "failed to store delegation")

	// Query delegations expiring at height 500
	expiringDelegations, err := keeper.getDelegationsExpiringAtHeight(ctx, 500)
	require.NoError(t, err, "failed to query expiring delegations")
	require.Len(t, expiringDelegations, 1, "expected 1 expiring delegation")
	require.Equal(t, "expiring-delegation", expiringDelegations[0].DelegationID, "wrong delegation in expiration index")

	// Query delegations expiring at a different height (should be empty)
	otherDelegations, err := keeper.getDelegationsExpiringAtHeight(ctx, 600)
	require.NoError(t, err, "failed to query expiring delegations")
	require.Len(t, otherDelegations, 0, "expected 0 delegations expiring at height 600")
}

// TestDelegationExpirationIndexCleanup verifies that the expiration index is cleaned up when delegation becomes inactive
func TestDelegationExpirationIndexCleanup(t *testing.T) {
	ctx, keeper := setupConfKeeper(t)

	// Create an active delegation
	delegation := &ScoreDelegation{
		DelegationID:   "cleanup-test",
		Delegator:      "aura1delegator",
		Delegate:       "aura1delegate",
		DelegatedScore: 1000,
		DelegationType: DelegationTypeValidation,
		StartHeight:    100,
		EndHeight:      500,
		Active:         true,
		Revocable:      true,
		RewardSharePct: 10,
		CreatedHeight:  100,
		LastUpdated:    100,
	}

	// Store the delegation
	err := keeper.storeDelegation(ctx, delegation)
	require.NoError(t, err, "failed to store delegation")

	// Verify it's in the expiration index
	expiringDelegations, err := keeper.getDelegationsExpiringAtHeight(ctx, 500)
	require.NoError(t, err)
	require.Len(t, expiringDelegations, 1, "delegation not found in expiration index")

	// Mark delegation as inactive and update
	delegation.Active = false
	delegation.LastUpdated = 200
	err = keeper.storeDelegation(ctx, delegation)
	require.NoError(t, err, "failed to update delegation")

	// Verify it's removed from the expiration index
	expiringDelegations, err = keeper.getDelegationsExpiringAtHeight(ctx, 500)
	require.NoError(t, err)
	require.Len(t, expiringDelegations, 0, "delegation should be removed from expiration index when inactive")
}

// TestGetDelegationsPaginatedSmallSet verifies that pagination works with small dataset
func TestGetDelegationsPaginatedSmallSet(t *testing.T) {
	ctx, keeper := setupConfKeeper(t)

	// Create multiple delegations
	for i := 0; i < 5; i++ {
		delegation := &ScoreDelegation{
			DelegationID:   sdk.AccAddress([]byte{byte(i)}).String(),
			Delegator:      "aura1delegator",
			Delegate:       "aura1delegate",
			DelegatedScore: uint64(1000 * (i + 1)),
			DelegationType: DelegationTypeValidation,
			StartHeight:    100,
			EndHeight:      0, // indefinite
			Active:         true,
			Revocable:      true,
			RewardSharePct: 10,
			CreatedHeight:  100,
			LastUpdated:    100,
		}
		err := keeper.storeDelegation(ctx, delegation)
		require.NoError(t, err)
	}

	// Use paginated method to retrieve all
	delegations, hasMore, err := keeper.getDelegationsPaginated(ctx, 0, 100)
	require.NoError(t, err)
	require.Len(t, delegations, 5, "expected 5 delegations")
	require.False(t, hasMore, "should not have more results")
}

// TestGetDelegationsPaginated verifies that pagination works correctly
func TestGetDelegationsPaginated(t *testing.T) {
	ctx, keeper := setupConfKeeper(t)

	// Create 10 delegations
	for i := 0; i < 10; i++ {
		delegation := &ScoreDelegation{
			DelegationID:   sdk.AccAddress([]byte{byte(i)}).String(),
			Delegator:      "aura1delegator",
			Delegate:       "aura1delegate",
			DelegatedScore: uint64(1000 * (i + 1)),
			DelegationType: DelegationTypeValidation,
			StartHeight:    uint64(100 + i),
			EndHeight:      0,
			Active:         true,
			Revocable:      true,
			RewardSharePct: 10,
			CreatedHeight:  uint64(100 + i),
			LastUpdated:    uint64(100 + i),
		}
		err := keeper.storeDelegation(ctx, delegation)
		require.NoError(t, err)
	}

	// Test pagination: first page
	delegations, hasMore, err := keeper.getDelegationsPaginated(ctx, 0, 5)
	require.NoError(t, err)
	require.Len(t, delegations, 5, "expected 5 delegations in first page")
	require.True(t, hasMore, "expected more results")

	// Test pagination: second page
	delegations, hasMore, err = keeper.getDelegationsPaginated(ctx, 5, 5)
	require.NoError(t, err)
	require.Len(t, delegations, 5, "expected 5 delegations in second page")
	require.False(t, hasMore, "expected no more results")

	// Test pagination: beyond available data
	delegations, hasMore, err = keeper.getDelegationsPaginated(ctx, 10, 5)
	require.NoError(t, err)
	require.Len(t, delegations, 0, "expected 0 delegations beyond available data")
	require.False(t, hasMore, "expected no more results")
}

// TestGetDelegationsByUser verifies user-specific delegation queries
func TestGetDelegationsByUser(t *testing.T) {
	ctx, keeper := setupConfKeeper(t)

	// Create delegations for multiple users
	users := []string{"aura1user1", "aura1user2", "aura1user3"}
	for i, user := range users {
		for j := 0; j < 3; j++ {
			delegation := &ScoreDelegation{
				DelegationID:   sdk.AccAddress([]byte{byte(i), byte(j)}).String(),
				Delegator:      user,
				Delegate:       "aura1delegate",
				DelegatedScore: uint64(1000 * (j + 1)),
				DelegationType: DelegationTypeValidation,
				StartHeight:    100,
				EndHeight:      0,
				Active:         j%2 == 0, // alternate active/inactive
				Revocable:      true,
				RewardSharePct: 10,
				CreatedHeight:  100,
				LastUpdated:    100,
			}
			err := keeper.storeDelegation(ctx, delegation)
			require.NoError(t, err)
		}
	}

	// Query all delegations for user1
	delegations, _, err := keeper.getDelegationsByUser(ctx, "aura1user1", false, 0, 100)
	require.NoError(t, err)
	require.Len(t, delegations, 3, "expected 3 delegations for user1")

	// Query only active delegations for user1
	activeDelegations, _, err := keeper.getDelegationsByUser(ctx, "aura1user1", true, 0, 100)
	require.NoError(t, err)
	require.Len(t, activeDelegations, 2, "expected 2 active delegations for user1")

	// Verify all returned delegations are for user1
	for _, delegation := range delegations {
		require.Equal(t, "aura1user1", delegation.Delegator, "delegation should be for user1")
	}
}
