package keeper_test

import (
	"testing"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
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
		params := k.GetParams()
		require.NotNil(t, params)
		require.True(t, params.MaxStorageBytes > 0)
	})

	t.Run("SetParams", func(t *testing.T) {
		newParams := types.DefaultParams()
		newParams.MaxStorageBytes = 1000000
		err := k.SetParams(newParams)
		require.NoError(t, err)

		retrieved := k.GetParams()
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
