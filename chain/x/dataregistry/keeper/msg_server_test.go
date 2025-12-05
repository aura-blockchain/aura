package keeper_test

import (
	"testing"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/dataregistry/keeper"
	"github.com/aequitas/aura/chain/x/dataregistry/params"
	"github.com/aequitas/aura/chain/x/dataregistry/types"
	"github.com/stretchr/testify/require"
)

func TestMsgServerFunctionality(t *testing.T) {
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

	// Create message server
	msgServer := keeper.NewMsgServer(k)
	require.NotNil(t, msgServer)

	t.Run("StoreDataItem", func(t *testing.T) {
		msg := &types.MsgStoreDataItem{
			Creator:         "aura1owner",
			DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:     []byte("hash123"),
			StorageLocation: "ipfs://test",
			Title:           "Test Photo",
			Description:     "A test photo",
			Tags:            []string{"test", "photo"},
			AccessPolicy: &types.AccessPolicy{
				Mode: types.AccessMode_ACCESS_MODE_PUBLIC,
			},
		}

		resp, err := msgServer.StoreDataItem(input.Ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp.DataId)

		// Verify data was stored
		item, found := k.GetDataItem(input.Ctx, resp.DataId)
		require.True(t, found)
		require.Equal(t, msg.Creator, item.OwnerAddress)
		require.Equal(t, msg.ContentHash, item.ContentHash)
	})

	t.Run("UpdateDataItem", func(t *testing.T) {
		// First store data
		storeMsg := &types.MsgStoreDataItem{
			Creator:         "aura1owner",
			DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:     []byte("hash456"),
			StorageLocation: "ipfs://test2",
			Title:           "Original Title",
		}

		storeResp, err := msgServer.StoreDataItem(input.Ctx, storeMsg)
		require.NoError(t, err)

		// Now update it
		updateMsg := &types.MsgUpdateDataItem{
			Creator:     "aura1owner",
			DataId:      storeResp.DataId,
			Title:       "Updated Title",
			Description: "Updated Description",
			Tags:        []string{"updated"},
		}

		updateResp, err := msgServer.UpdateDataItem(input.Ctx, updateMsg)
		require.NoError(t, err)
		require.NotNil(t, updateResp)

		// Verify update
		item, found := k.GetDataItem(input.Ctx, storeResp.DataId)
		require.True(t, found)
		require.Equal(t, "Updated Title", item.Title)
		require.Equal(t, "Updated Description", item.Description)
	})

	t.Run("UpdateDataItem_Unauthorized", func(t *testing.T) {
		// Store data as one owner
		storeMsg := &types.MsgStoreDataItem{
			Creator:         "aura1owner",
			DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:     []byte("hash789"),
			Title:           "Original Photo",
			StorageLocation: "ipfs://original",
		}

		storeResp, err := msgServer.StoreDataItem(input.Ctx, storeMsg)
		require.NoError(t, err)

		// Try to update as different owner
		updateMsg := &types.MsgUpdateDataItem{
			Creator: "aura1attacker",
			DataId:  storeResp.DataId,
			Title:   "Hacked Title",
		}

		_, err = msgServer.UpdateDataItem(input.Ctx, updateMsg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("DeleteDataItem", func(t *testing.T) {
		// Store data
		storeMsg := &types.MsgStoreDataItem{
			Creator:         "aura1owner",
			DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:     []byte("hash_delete"),
			Title:           "Photo to Delete",
			StorageLocation: "ipfs://delete",
		}

		storeResp, err := msgServer.StoreDataItem(input.Ctx, storeMsg)
		require.NoError(t, err)

		// Delete it
		deleteMsg := &types.MsgDeleteDataItem{
			Creator: "aura1owner",
			DataId:  storeResp.DataId,
		}

		deleteResp, err := msgServer.DeleteDataItem(input.Ctx, deleteMsg)
		require.NoError(t, err)
		require.NotNil(t, deleteResp)

		// Verify deletion
		_, found := k.GetDataItem(input.Ctx, storeResp.DataId)
		require.False(t, found)
	})

	t.Run("DeleteDataItem_Unauthorized", func(t *testing.T) {
		// Store data
		storeMsg := &types.MsgStoreDataItem{
			Creator:         "aura1owner",
			DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:     []byte("hash_delete2"),
			Title:           "Protected Photo",
			StorageLocation: "ipfs://protected",
		}

		storeResp, err := msgServer.StoreDataItem(input.Ctx, storeMsg)
		require.NoError(t, err)

		// Try to delete as different owner
		deleteMsg := &types.MsgDeleteDataItem{
			Creator: "aura1attacker",
			DataId:  storeResp.DataId,
		}

		_, err = msgServer.DeleteDataItem(input.Ctx, deleteMsg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unauthorized")
	})
}
