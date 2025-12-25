// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/confidencescore/params"
	"github.com/aequitas/aura/chain/x/confidencescore/types"
)

// setupTestKeeper creates a test keeper for delegation tests
func setupTestKeeperForDelegation(t *testing.T) (*Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey("confidencescore")
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	paramsStore := params.NewStore(types.DefaultParams())

	keeper := NewKeeper(
		runtime.NewKVStoreService(storeKey),
		cdc,
		paramsStore,
		"authority",
		log.NewNopLogger(),
	)

	return keeper, ctx
}

// TestDelegationPagination tests the pagination functionality
func TestDelegationPagination(t *testing.T) {
	keeper, ctx := setupTestKeeperForDelegation(t)

	// Create 150 delegations to test pagination
	numDelegations := 150
	for i := 0; i < numDelegations; i++ {
		delegation := &ScoreDelegation{
			DelegationID:   fmt.Sprintf("del-%d", i),
			Delegator:      fmt.Sprintf("delegator-%d", i%10), // 10 unique delegators
			Delegate:       fmt.Sprintf("delegate-%d", i%5),   // 5 unique delegates
			DelegatedScore: uint64(100 + i),
			DelegationType: DelegationTypeValidation,
			StartHeight:    uint64(ctx.BlockHeight()),
			EndHeight:      uint64(ctx.BlockHeight() + 1000),
			Active:         true,
			Revocable:      true,
			RewardSharePct: 10,
			CreatedHeight:  uint64(ctx.BlockHeight()),
			LastUpdated:    uint64(ctx.BlockHeight()),
		}
		err := keeper.storeDelegation(ctx, delegation)
		require.NoError(t, err)
	}

	// Test pagination - first page
	page1, hasMore, err := keeper.getDelegationsPaginated(ctx, 0, 50)
	require.NoError(t, err)
	require.True(t, hasMore)
	require.Len(t, page1, 50)

	// Test pagination - second page
	page2, hasMore, err := keeper.getDelegationsPaginated(ctx, 50, 50)
	require.NoError(t, err)
	require.True(t, hasMore)
	require.Len(t, page2, 50)

	// Test pagination - third page
	page3, hasMore, err := keeper.getDelegationsPaginated(ctx, 100, 50)
	require.NoError(t, err)
	require.False(t, hasMore) // No more results after this
	require.Len(t, page3, 50)

	// Test pagination - beyond end
	page4, hasMore, err := keeper.getDelegationsPaginated(ctx, 200, 50)
	require.NoError(t, err)
	require.False(t, hasMore)
	require.Len(t, page4, 0)

	// Verify no duplicates across pages
	seen := make(map[string]bool)
	for _, d := range append(append(page1, page2...), page3...) {
		require.False(t, seen[d.DelegationID], "duplicate delegation found: %s", d.DelegationID)
		seen[d.DelegationID] = true
	}

	// Should have all 150 delegations
	require.Len(t, seen, numDelegations)
}

// TestDelegationUserFiltering tests filtering by user with pagination
func TestDelegationUserFiltering(t *testing.T) {
	keeper, ctx := setupTestKeeperForDelegation(t)

	targetUser := "target-user"
	otherUser := "other-user"

	// Create delegations for target user
	for i := 0; i < 30; i++ {
		delegation := &ScoreDelegation{
			DelegationID:   fmt.Sprintf("target-del-%d", i),
			Delegator:      targetUser,
			Delegate:       fmt.Sprintf("delegate-%d", i),
			DelegatedScore: 100,
			DelegationType: DelegationTypeValidation,
			StartHeight:    uint64(ctx.BlockHeight()),
			EndHeight:      uint64(ctx.BlockHeight() + 1000),
			Active:         true,
			Revocable:      true,
			RewardSharePct: 10,
			CreatedHeight:  uint64(ctx.BlockHeight()),
			LastUpdated:    uint64(ctx.BlockHeight()),
		}
		err := keeper.storeDelegation(ctx, delegation)
		require.NoError(t, err)
	}

	// Create delegations for other users
	for i := 0; i < 70; i++ {
		delegation := &ScoreDelegation{
			DelegationID:   fmt.Sprintf("other-del-%d", i),
			Delegator:      otherUser,
			Delegate:       fmt.Sprintf("delegate-%d", i),
			DelegatedScore: 100,
			DelegationType: DelegationTypeValidation,
			StartHeight:    uint64(ctx.BlockHeight()),
			EndHeight:      uint64(ctx.BlockHeight() + 1000),
			Active:         true,
			Revocable:      true,
			RewardSharePct: 10,
			CreatedHeight:  uint64(ctx.BlockHeight()),
			LastUpdated:    uint64(ctx.BlockHeight()),
		}
		err := keeper.storeDelegation(ctx, delegation)
		require.NoError(t, err)
	}

	// Query target user's delegations
	userDels, hasMore, err := keeper.getDelegationsByUser(ctx, targetUser, false, 0, 100)
	require.NoError(t, err)
	require.False(t, hasMore)
	require.Len(t, userDels, 30)

	// Verify all returned delegations are for target user
	for _, d := range userDels {
		require.True(t, d.Delegator == targetUser || d.Delegate == targetUser)
	}

	// Test pagination of user delegations
	page1, hasMore, err := keeper.getDelegationsByUser(ctx, targetUser, false, 0, 10)
	require.NoError(t, err)
	require.True(t, hasMore)
	require.Len(t, page1, 10)

	page2, hasMore, err := keeper.getDelegationsByUser(ctx, targetUser, false, 10, 10)
	require.NoError(t, err)
	require.True(t, hasMore)
	require.Len(t, page2, 10)
}

// TestExpirationIndex tests the expiration index functionality
func TestExpirationIndex(t *testing.T) {
	keeper, ctx := setupTestKeeperForDelegation(t)

	currentHeight := uint64(ctx.BlockHeight())
	expirationHeight := currentHeight + 100

	// Create delegations expiring at different heights
	for i := 0; i < 50; i++ {
		height := expirationHeight
		if i < 10 {
			height = expirationHeight - 10 // Expire earlier
		} else if i < 20 {
			height = expirationHeight + 10 // Expire later
		}
		// Rest expire at expirationHeight

		delegation := &ScoreDelegation{
			DelegationID:   fmt.Sprintf("exp-del-%d", i),
			Delegator:      fmt.Sprintf("delegator-%d", i),
			Delegate:       fmt.Sprintf("delegate-%d", i),
			DelegatedScore: 100,
			DelegationType: DelegationTypeValidation,
			StartHeight:    currentHeight,
			EndHeight:      height,
			Active:         true,
			Revocable:      false,
			RewardSharePct: 10,
			CreatedHeight:  currentHeight,
			LastUpdated:    currentHeight,
		}
		err := keeper.storeDelegation(ctx, delegation)
		require.NoError(t, err)
	}

	// Query delegations expiring at expirationHeight
	expiring, err := keeper.getDelegationsExpiringAtHeight(ctx, expirationHeight)
	require.NoError(t, err)
	require.Len(t, expiring, 30, "should find 30 delegations expiring at target height")

	// Query delegations expiring earlier
	expiring, err = keeper.getDelegationsExpiringAtHeight(ctx, expirationHeight-10)
	require.NoError(t, err)
	require.Len(t, expiring, 10)

	// Query delegations expiring later
	expiring, err = keeper.getDelegationsExpiringAtHeight(ctx, expirationHeight+10)
	require.NoError(t, err)
	require.Len(t, expiring, 10)

	// Query non-existent expiration height
	expiring, err = keeper.getDelegationsExpiringAtHeight(ctx, currentHeight+9999)
	require.NoError(t, err)
	require.Len(t, expiring, 0)
}

// TestProcessExpiredDelegationsScalability tests that processing is O(k) not O(n)
func TestProcessExpiredDelegationsScalability(t *testing.T) {
	keeper, ctx := setupTestKeeperForDelegation(t)

	// Create user records for delegators and delegates
	for i := 0; i < 20; i++ {
		delegatorRecord := keeper.GetOrCreateUserRecord(ctx, fmt.Sprintf("delegator-%d", i))
		delegatorRecord.TotalScore = 0 // Will be updated when delegations expire
		require.NoError(t, keeper.SetUserRecord(ctx, delegatorRecord))

		delegateRecord := keeper.GetOrCreateUserRecord(ctx, fmt.Sprintf("delegate-%d", i))
		delegateRecord.TotalScore = 100 * 10 // Has delegated score
		require.NoError(t, keeper.SetUserRecord(ctx, delegateRecord))
	}

	currentHeight := uint64(ctx.BlockHeight())
	expirationHeight := currentHeight + 100

	// Create 1000 delegations, only 5 expiring at target height
	for i := 0; i < 1000; i++ {
		height := expirationHeight + 1000 // Most expire much later
		if i < 5 {
			height = expirationHeight // Only 5 expire at target height
		}

		delegation := &ScoreDelegation{
			DelegationID:   fmt.Sprintf("scale-del-%d", i),
			Delegator:      fmt.Sprintf("delegator-%d", i%20),
			Delegate:       fmt.Sprintf("delegate-%d", i%20),
			DelegatedScore: 100,
			DelegationType: DelegationTypeValidation,
			StartHeight:    currentHeight,
			EndHeight:      height,
			Active:         true,
			Revocable:      false,
			RewardSharePct: 10,
			CreatedHeight:  currentHeight,
			LastUpdated:    currentHeight,
		}
		err := keeper.storeDelegation(ctx, delegation)
		require.NoError(t, err)
	}

	// Move to expiration height
	ctx = ctx.WithBlockHeight(int64(expirationHeight))

	// Process expirations - should only process 5, not iterate all 1000
	processed, err := keeper.ProcessExpiredDelegations(ctx)
	require.NoError(t, err)
	require.Equal(t, 5, processed, "should only process 5 expiring delegations")

	// Verify the 5 delegations are now inactive
	for i := 0; i < 5; i++ {
		delegation, ok := keeper.getDelegation(ctx, fmt.Sprintf("scale-del-%d", i))
		require.True(t, ok)
		require.False(t, delegation.Active, "delegation should be inactive after expiration")
	}

	// Verify others are still active
	delegation, ok := keeper.getDelegation(ctx, "scale-del-10")
	require.True(t, ok)
	require.True(t, delegation.Active, "non-expiring delegation should still be active")
}

// TestExpirationIndexCleanup tests that expiration index is cleaned when delegation becomes inactive
func TestExpirationIndexCleanup(t *testing.T) {
	keeper, ctx := setupTestKeeperForDelegation(t)

	currentHeight := uint64(ctx.BlockHeight())
	expirationHeight := currentHeight + 100

	// Create a delegation with expiration
	delegation := &ScoreDelegation{
		DelegationID:   "cleanup-del",
		Delegator:      "delegator",
		Delegate:       "delegate",
		DelegatedScore: 100,
		DelegationType: DelegationTypeValidation,
		StartHeight:    currentHeight,
		EndHeight:      expirationHeight,
		Active:         true,
		Revocable:      false,
		RewardSharePct: 10,
		CreatedHeight:  currentHeight,
		LastUpdated:    currentHeight,
	}
	err := keeper.storeDelegation(ctx, delegation)
	require.NoError(t, err)

	// Verify expiration index exists
	expiring, err := keeper.getDelegationsExpiringAtHeight(ctx, expirationHeight)
	require.NoError(t, err)
	require.Len(t, expiring, 1)

	// Mark delegation as inactive
	delegation.Active = false
	err = keeper.storeDelegation(ctx, delegation)
	require.NoError(t, err)

	// Verify expiration index is cleaned up
	expiring, err = keeper.getDelegationsExpiringAtHeight(ctx, expirationHeight)
	require.NoError(t, err)
	require.Len(t, expiring, 0, "expiration index should be cleaned when delegation becomes inactive")
}
