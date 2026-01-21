// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"context"
	"testing"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/dataregistry/ipfs"
	"github.com/aequitas/aura/chain/x/dataregistry/keeper"
	"github.com/aequitas/aura/chain/x/dataregistry/params"
	"github.com/aequitas/aura/chain/x/dataregistry/types"
	"github.com/stretchr/testify/require"
)

func TestKeeperFunctionality(t *testing.T) {
	input := keepertest.CreateTestInput(t)

	// Create params store
	paramsStore := params.NewStore(types.DefaultParams())

	// Create keeper
	k := keeper.NewKeeper(
		keepertest.WrapStoreService(input.StoreKey),
		input.Cdc,
		paramsStore,
		"aura1authority",
		keepertest.Logger(),
	)

	// Test basic keeper operations
	t.Run("GetParams", func(t *testing.T) {
		params, _ := k.GetParams(input.Ctx)
		require.NotNil(t, params)
		require.True(t, params.MaxStorageBytes > 0)
	})

	t.Run("SetParams", func(t *testing.T) {
		newParams := types.DefaultParams()
		newParams.MaxStorageBytes = 1000000
		err := k.SetParams(newParams)
		require.NoError(t, err)

		retrieved, _ := k.GetParams(input.Ctx)
		require.Equal(t, uint64(1000000), retrieved.MaxStorageBytes)
	})

	t.Run("DataItemCounter", func(t *testing.T) {
		// Get initial counter
		nextID := k.GetNextDataID(input.Ctx)
		require.Equal(t, uint64(1), nextID)

		// Set and retrieve counter
		err := k.SetNextDataID(input.Ctx, 42)
		require.NoError(t, err)

		retrieved := k.GetNextDataID(input.Ctx)
		require.Equal(t, uint64(42), retrieved)
	})

	t.Run("DataItemCRUD", func(t *testing.T) {
		// Create data item
		item := types.DataItem{
			DataId:          "test-data-1",
			OwnerAddress:    "aura1owner",
			DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:     []byte("hash123"),
			StorageLocation: "ipfs://test",
			Status:          types.DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION,
		}

		// Test SetDataItem
		err := k.SetDataItem(input.Ctx, item)
		require.NoError(t, err)

		// Test GetDataItem
		retrieved, found := k.GetDataItem(input.Ctx, "test-data-1")
		require.True(t, found)
		require.Equal(t, item.DataId, retrieved.DataId)
		require.Equal(t, item.OwnerAddress, retrieved.OwnerAddress)
		require.Equal(t, item.DataType, retrieved.DataType)

		// Test DeleteDataItem
		err = k.DeleteDataItem(input.Ctx, "test-data-1")
		require.NoError(t, err)

		_, found = k.GetDataItem(input.Ctx, "test-data-1")
		require.False(t, found)
	})

	t.Run("CheckAccess", func(t *testing.T) {
		// Create data item with private access
		item := types.DataItem{
			DataId:       "test-data-2",
			OwnerAddress: "aura1owner",
			DataType:     types.DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:  []byte("hash456"),
			Status:       types.DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION,
			AccessPolicy: &types.AccessPolicy{
				Mode: types.AccessMode_ACCESS_MODE_PRIVATE,
			},
		}

		err := k.SetDataItem(input.Ctx, item)
		require.NoError(t, err)

		// Owner should have access
		hasAccess := k.CheckAccess(input.Ctx, "test-data-2", "aura1owner")
		require.True(t, hasAccess)

		// Non-owner should not have access (private mode)
		hasAccess = k.CheckAccess(input.Ctx, "test-data-2", "aura1other")
		require.False(t, hasAccess)

		// Test public access
		item.AccessPolicy.Mode = types.AccessMode_ACCESS_MODE_PUBLIC
		err = k.SetDataItem(input.Ctx, item)
		require.NoError(t, err)

		hasAccess = k.CheckAccess(input.Ctx, "test-data-2", "aura1other")
		require.True(t, hasAccess)
	})
}

func TestSetBankKeeper(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	paramsStore := params.NewStore(types.DefaultParams())

	k := keeper.NewKeeper(
		keepertest.WrapStoreService(input.StoreKey),
		input.Cdc,
		paramsStore,
		"aura1authority",
		keepertest.Logger(),
	)

	// Create a mock bank keeper
	mockBankKeeper := &mockBankKeeper{}

	// Set the bank keeper
	k.SetBankKeeper(mockBankKeeper)

	// Verify it was set (we can't directly access the private field,
	// but we verify by checking that no panic occurs when setting)
	require.NotNil(t, k)
}

func TestSetAuthority(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	paramsStore := params.NewStore(types.DefaultParams())

	k := keeper.NewKeeper(
		keepertest.WrapStoreService(input.StoreKey),
		input.Cdc,
		paramsStore,
		"aura1initial",
		keepertest.Logger(),
	)

	// Initial authority
	require.Equal(t, "aura1initial", k.GetAuthority())

	// Set new authority
	k.SetAuthority("aura1newauthority")

	// Verify authority was updated
	require.Equal(t, "aura1newauthority", k.GetAuthority())
}

func TestGetAuthority(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	paramsStore := params.NewStore(types.DefaultParams())

	testCases := []struct {
		name      string
		authority string
	}{
		{"empty authority", ""},
		{"valid authority", "aura1authority"},
		{"long authority", "aura1verylongauthorityaddressfortesting"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			k := keeper.NewKeeper(
				keepertest.WrapStoreService(input.StoreKey),
				input.Cdc,
				paramsStore,
				tc.authority,
				keepertest.Logger(),
			)

			got := k.GetAuthority()
			require.Equal(t, tc.authority, got)
		})
	}
}

func TestLogger(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	paramsStore := params.NewStore(types.DefaultParams())

	k := keeper.NewKeeper(
		keepertest.WrapStoreService(input.StoreKey),
		input.Cdc,
		paramsStore,
		"aura1authority",
		keepertest.Logger(),
	)

	// Get logger from keeper
	logger := k.Logger(input.Ctx)

	// Verify logger is not nil
	require.NotNil(t, logger)

	// Verify logger can be used without panicking
	logger.Info("test message")
	logger.Debug("debug message")
	logger.Error("error message")
}

func TestSetIPFSClient(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	paramsStore := params.NewStore(types.DefaultParams())

	k := keeper.NewKeeper(
		keepertest.WrapStoreService(input.StoreKey),
		input.Cdc,
		paramsStore,
		"aura1authority",
		keepertest.Logger(),
	)

	// Default mock client is set in NewKeeper
	defaultClient := k.GetIPFSClient()
	require.NotNil(t, defaultClient)

	// Create a new mock client
	newMockClient := ipfs.NewMockClient()

	// Set new IPFS client
	k.SetIPFSClient(newMockClient)

	// Verify client was updated
	retrievedClient := k.GetIPFSClient()
	require.NotNil(t, retrievedClient)
	require.Same(t, newMockClient, retrievedClient)
}

func TestGetIPFSClient(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	paramsStore := params.NewStore(types.DefaultParams())

	k := keeper.NewKeeper(
		keepertest.WrapStoreService(input.StoreKey),
		input.Cdc,
		paramsStore,
		"aura1authority",
		keepertest.Logger(),
	)

	// NewKeeper sets a default mock client
	client := k.GetIPFSClient()
	require.NotNil(t, client)

	// Verify it implements IPFSClient interface by calling methods
	ctx := context.Background()
	connected := client.IsConnected(ctx)
	require.True(t, connected)

	nodeID, err := client.GetNodeID(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, nodeID)
}

func TestSearchDataItemsWithGeoFilter(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	paramsStore := params.NewStore(types.DefaultParams())

	k := keeper.NewKeeper(
		keepertest.WrapStoreService(input.StoreKey),
		input.Cdc,
		paramsStore,
		"aura1authority",
		keepertest.Logger(),
	)

	// Create data items with geo locations
	// Item 1: New York City (40.7128, -74.0060)
	item1 := types.DataItem{
		DataId:       "geo-data-1",
		OwnerAddress: "aura1owner",
		DataType:     types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		ContentHash:  []byte("hash1"),
		Status:       types.DataItemStatus_DATA_ITEM_STATUS_VERIFIED,
		AccessPolicy: &types.AccessPolicy{
			Mode: types.AccessMode_ACCESS_MODE_PUBLIC,
		},
		GeoLocation: &types.GeoLocation{
			Latitude:  40.7128,
			Longitude: -74.0060,
		},
	}

	// Item 2: Los Angeles (34.0522, -118.2437) - far from NYC
	item2 := types.DataItem{
		DataId:       "geo-data-2",
		OwnerAddress: "aura1owner",
		DataType:     types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		ContentHash:  []byte("hash2"),
		Status:       types.DataItemStatus_DATA_ITEM_STATUS_VERIFIED,
		AccessPolicy: &types.AccessPolicy{
			Mode: types.AccessMode_ACCESS_MODE_PUBLIC,
		},
		GeoLocation: &types.GeoLocation{
			Latitude:  34.0522,
			Longitude: -118.2437,
		},
	}

	// Item 3: Newark, NJ (40.7357, -74.1724) - close to NYC
	item3 := types.DataItem{
		DataId:       "geo-data-3",
		OwnerAddress: "aura1owner",
		DataType:     types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		ContentHash:  []byte("hash3"),
		Status:       types.DataItemStatus_DATA_ITEM_STATUS_VERIFIED,
		AccessPolicy: &types.AccessPolicy{
			Mode: types.AccessMode_ACCESS_MODE_PUBLIC,
		},
		GeoLocation: &types.GeoLocation{
			Latitude:  40.7357,
			Longitude: -74.1724,
		},
	}

	// Store all items
	err := k.SetDataItem(input.Ctx, item1)
	require.NoError(t, err)
	err = k.SetDataItem(input.Ctx, item2)
	require.NoError(t, err)
	err = k.SetDataItem(input.Ctx, item3)
	require.NoError(t, err)

	// Search from NYC with small radius (should find NYC and Newark only)
	// Using squared distance: 0.05 radius means items within ~0.22 units of lat/lon
	searchLocation := &types.GeoLocation{
		Latitude:  40.7128,
		Longitude: -74.0060,
	}

	results := k.SearchDataItems(
		input.Ctx,
		"",  // query
		nil, // tags
		types.DataItemType_DATA_ITEM_TYPE_UNSPECIFIED, // type filter
		searchLocation,   // geo location
		0.05,             // radius (squared distance threshold)
		"aura1requester", // requester
	)

	// Should find NYC and Newark (close), but not LA (far)
	require.Len(t, results, 2)

	foundIDs := make(map[string]bool)
	for _, r := range results {
		foundIDs[r.DataId] = true
	}
	require.True(t, foundIDs["geo-data-1"], "NYC item should be found")
	require.True(t, foundIDs["geo-data-3"], "Newark item should be found")
	require.False(t, foundIDs["geo-data-2"], "LA item should not be found")

	// Search with larger radius (should find all)
	results = k.SearchDataItems(
		input.Ctx,
		"",
		nil,
		types.DataItemType_DATA_ITEM_TYPE_UNSPECIFIED,
		searchLocation,
		5000.0, // very large radius
		"aura1requester",
	)

	require.Len(t, results, 3)

	// Search with zero radius (no geo filter applied - should find all)
	results = k.SearchDataItems(
		input.Ctx,
		"",
		nil,
		types.DataItemType_DATA_ITEM_TYPE_UNSPECIFIED,
		searchLocation,
		0, // zero radius means no geo filter
		"aura1requester",
	)

	require.Len(t, results, 3)
}

// mockBankKeeper implements keeper.BankKeeper for testing
type mockBankKeeper struct{}

func (m *mockBankKeeper) MintCoins(ctx context.Context, moduleName string, amt interface{}) error {
	return nil
}

func (m *mockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr interface{}, amt interface{}) error {
	return nil
}
