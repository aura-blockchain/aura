package keeper_test

import (
	"testing"

	"github.com/cosmos/gogoproto/types"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/dataregistry/keeper"
	"github.com/aequitas/aura/chain/x/dataregistry/params"
	drtypes "github.com/aequitas/aura/chain/x/dataregistry/types"
	"github.com/stretchr/testify/require"
)

func TestQueryServerFunctionality(t *testing.T) {
	input := keepertest.CreateTestInput(t)

	// Create params store
	paramsStore := params.NewStore(drtypes.DefaultParams())

	// Create keeper
	k := keeper.NewKeeper(
		keepertest.WrapStoreService(input.StoreKey),
		input.Cdc,
		paramsStore,
		"aura1authority",
		keepertest.Logger(),
	)

	// Create query server
	queryServer := keeper.NewQueryServer(k)
	require.NotNil(t, queryServer)

	// Setup test data
	createdAt1, _ := types.TimestampProto(input.Ctx.BlockTime())
	item1 := drtypes.DataItem{
		DataId:          "test-data-1",
		OwnerAddress:    "aura1owner",
		DataType:        drtypes.DataItemType_DATA_ITEM_TYPE_PHOTO,
		ContentHash:     []byte("hash1"),
		StorageLocation: "ipfs://test1",
		Status:          drtypes.DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION,
		Title:           "Photo 1",
		Tags:            []string{"test"},
		CreatedAt:       createdAt1,
		AccessPolicy:    &drtypes.AccessPolicy{},
	}

	createdAt2, _ := types.TimestampProto(input.Ctx.BlockTime())
	item2 := drtypes.DataItem{
		DataId:          "test-data-2",
		OwnerAddress:    "aura1owner",
		DataType:        drtypes.DataItemType_DATA_ITEM_TYPE_VIDEO,
		ContentHash:     []byte("hash2"),
		StorageLocation: "ipfs://test2",
		Status:          drtypes.DataItemStatus_DATA_ITEM_STATUS_VERIFIED,
		Title:           "Video 1",
		Tags:            []string{"test", "video"},
		CreatedAt:       createdAt2,
		AccessPolicy:    &drtypes.AccessPolicy{},
	}

	require.NoError(t, k.SetDataItem(input.Ctx, item1))
	require.NoError(t, k.SetDataItem(input.Ctx, item2))

	t.Run("QueryDataItem", func(t *testing.T) {
		req := &drtypes.QueryDataItemRequest{
			DataId:    "test-data-1",
			Requester: "aura1owner", // Must match owner for PRIVATE access policy
		}

		resp, err := queryServer.DataItem(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, resp.HasAccess)
		require.NotNil(t, resp.DataItem)
		require.Equal(t, item1.DataId, resp.DataItem.DataId)
		require.Equal(t, item1.OwnerAddress, resp.DataItem.OwnerAddress)
	})

	t.Run("QueryDataItem_NotFound", func(t *testing.T) {
		req := &drtypes.QueryDataItemRequest{
			DataId:    "nonexistent",
			Requester: "aura1owner",
		}

		resp, err := queryServer.DataItem(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.False(t, resp.Exists)
		require.Nil(t, resp.DataItem)
	})

	t.Run("QueryDataItem_AccessDenied", func(t *testing.T) {
		req := &drtypes.QueryDataItemRequest{
			DataId:    "test-data-1",
			Requester: "aura1other", // Different user, private access policy
		}

		resp, err := queryServer.DataItem(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, resp.Exists)
		require.False(t, resp.HasAccess)
		require.Nil(t, resp.DataItem)
	})

	t.Run("QueryUserDataItems", func(t *testing.T) {
		req := &drtypes.QueryUserDataItemsRequest{
			OwnerAddress: "aura1owner",
		}

		resp, err := queryServer.UserDataItems(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.DataItems, 2)
	})

	t.Run("QueryUserDataItems_WithTypeFilter", func(t *testing.T) {
		req := &drtypes.QueryUserDataItemsRequest{
			OwnerAddress: "aura1owner",
			TypeFilter:   drtypes.DataItemType_DATA_ITEM_TYPE_PHOTO,
		}

		resp, err := queryServer.UserDataItems(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.DataItems, 1)
		require.Equal(t, drtypes.DataItemType_DATA_ITEM_TYPE_PHOTO, resp.DataItems[0].DataType)
	})

	t.Run("QueryUserDataItems_WithStatusFilter", func(t *testing.T) {
		req := &drtypes.QueryUserDataItemsRequest{
			OwnerAddress: "aura1owner",
			StatusFilter: drtypes.DataItemStatus_DATA_ITEM_STATUS_VERIFIED,
		}

		resp, err := queryServer.UserDataItems(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.DataItems, 1)
		require.Equal(t, drtypes.DataItemStatus_DATA_ITEM_STATUS_VERIFIED, resp.DataItems[0].Status)
	})

	t.Run("QuerySearchDataItems", func(t *testing.T) {
		req := &drtypes.QuerySearchDataItemsRequest{
			Tags:      []string{"test"},
			Requester: "aura1owner",
		}

		resp, err := queryServer.SearchDataItems(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.GreaterOrEqual(t, len(resp.DataItems), 2)
	})

	t.Run("QuerySearchDataItems_WithTypeFilter", func(t *testing.T) {
		req := &drtypes.QuerySearchDataItemsRequest{
			TypeFilter: drtypes.DataItemType_DATA_ITEM_TYPE_VIDEO,
			Requester:  "aura1owner",
		}

		resp, err := queryServer.SearchDataItems(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.DataItems, 1)
		require.Equal(t, drtypes.DataItemType_DATA_ITEM_TYPE_VIDEO, resp.DataItems[0].DataType)
	})

	t.Run("QueryStats", func(t *testing.T) {
		req := &drtypes.QueryStatsRequest{}

		resp, err := queryServer.Stats(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, uint64(2), resp.TotalDataItems)
		require.Equal(t, uint64(1), resp.TotalVerifiedItems)
	})

	t.Run("QueryParams", func(t *testing.T) {
		req := &drtypes.QueryParamsRequest{}

		resp, err := queryServer.Params(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Params)
		require.True(t, resp.Params.MaxStorageBytes > 0)
	})
}
