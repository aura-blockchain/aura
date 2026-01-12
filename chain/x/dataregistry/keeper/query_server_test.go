// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/query"
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

	t.Run("QueryDataItem_NilRequest", func(t *testing.T) {
		resp, err := queryServer.DataItem(input.Ctx, nil)
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "request cannot be nil")
	})

	t.Run("QueryDataItem_EmptyDataId", func(t *testing.T) {
		req := &drtypes.QueryDataItemRequest{
			DataId:    "",
			Requester: "aura1owner",
		}

		resp, err := queryServer.DataItem(input.Ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "data_id cannot be empty")
	})

	t.Run("QueryUserDataItems_NilRequest", func(t *testing.T) {
		resp, err := queryServer.UserDataItems(input.Ctx, nil)
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "request cannot be nil")
	})

	t.Run("QueryUserDataItems_EmptyOwnerAddress", func(t *testing.T) {
		req := &drtypes.QueryUserDataItemsRequest{
			OwnerAddress: "",
		}

		resp, err := queryServer.UserDataItems(input.Ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "owner_address cannot be empty")
	})

	t.Run("QuerySearchDataItems_NilRequest", func(t *testing.T) {
		resp, err := queryServer.SearchDataItems(input.Ctx, nil)
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "request cannot be nil")
	})

	t.Run("QueryStats_NilRequest", func(t *testing.T) {
		resp, err := queryServer.Stats(input.Ctx, nil)
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "request cannot be nil")
	})

	t.Run("QueryParams_NilRequest", func(t *testing.T) {
		resp, err := queryServer.Params(input.Ctx, nil)
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "request cannot be nil")
	})
}

func TestQueryDataItemVerifications(t *testing.T) {
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

	// Setup test data item with verifications
	createdAt, _ := types.TimestampProto(input.Ctx.BlockTime())
	item := drtypes.DataItem{
		DataId:          "test-verify-1",
		OwnerAddress:    "aura1owner",
		DataType:        drtypes.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		ContentHash:     []byte("hash-verify"),
		StorageLocation: "ipfs://verify1",
		Status:          drtypes.DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION,
		Title:           "Document for Verification",
		Tags:            []string{"verify"},
		CreatedAt:       createdAt,
		AccessPolicy:    &drtypes.AccessPolicy{},
		Verifications:   []*drtypes.Verification{},
	}

	require.NoError(t, k.SetDataItem(input.Ctx, item))

	// Add verifications using keeper method
	err := k.VerifyDataItem(
		input.Ctx,
		"test-verify-1",
		"aura1verifier1",
		drtypes.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
		85,
		"Verified by peer review",
		"manual-review",
		[]byte("proof1"),
	)
	require.NoError(t, err)

	err = k.VerifyDataItem(
		input.Ctx,
		"test-verify-1",
		"aura1verifier2",
		drtypes.VerificationLevel_VERIFICATION_LEVEL_AI_VERIFIED,
		92,
		"Verified by AI analysis",
		"ai-verification",
		[]byte("proof2"),
	)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		req := &drtypes.QueryDataItemVerificationsRequest{
			DataId: "test-verify-1",
		}

		resp, err := queryServer.DataItemVerifications(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Verifications, 2)

		// Check first verification
		require.Equal(t, "aura1verifier1", resp.Verifications[0].VerifierAddress)
		require.Equal(t, drtypes.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED, resp.Verifications[0].Level)
		require.Equal(t, uint64(85), resp.Verifications[0].ConfidenceScore)
		require.Equal(t, "Verified by peer review", resp.Verifications[0].Notes)

		// Check second verification
		require.Equal(t, "aura1verifier2", resp.Verifications[1].VerifierAddress)
		require.Equal(t, drtypes.VerificationLevel_VERIFICATION_LEVEL_AI_VERIFIED, resp.Verifications[1].Level)
		require.Equal(t, uint64(92), resp.Verifications[1].ConfidenceScore)
	})

	t.Run("NilRequest", func(t *testing.T) {
		resp, err := queryServer.DataItemVerifications(input.Ctx, nil)
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "request cannot be nil")
	})

	t.Run("EmptyDataId", func(t *testing.T) {
		req := &drtypes.QueryDataItemVerificationsRequest{
			DataId: "",
		}

		resp, err := queryServer.DataItemVerifications(input.Ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "data_id cannot be empty")
	})

	t.Run("ItemNotFound", func(t *testing.T) {
		req := &drtypes.QueryDataItemVerificationsRequest{
			DataId: "nonexistent-item",
		}

		resp, err := queryServer.DataItemVerifications(input.Ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
	})

	t.Run("ItemWithNoVerifications", func(t *testing.T) {
		// Create item with no verifications
		itemNoVerify := drtypes.DataItem{
			DataId:          "test-no-verify",
			OwnerAddress:    "aura1owner",
			DataType:        drtypes.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
			ContentHash:     []byte("hash-no-verify"),
			StorageLocation: "ipfs://noverify",
			Status:          drtypes.DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION,
			Title:           "No Verification",
			Tags:            []string{},
			CreatedAt:       createdAt,
			AccessPolicy:    &drtypes.AccessPolicy{},
			Verifications:   []*drtypes.Verification{},
		}
		require.NoError(t, k.SetDataItem(input.Ctx, itemNoVerify))

		req := &drtypes.QueryDataItemVerificationsRequest{
			DataId: "test-no-verify",
		}

		resp, err := queryServer.DataItemVerifications(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Empty(t, resp.Verifications)
	})
}

func TestPaginateDataItems(t *testing.T) {
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

	// Create 10 data items for pagination testing
	createdAt, _ := types.TimestampProto(input.Ctx.BlockTime())
	for i := 0; i < 10; i++ {
		item := drtypes.DataItem{
			DataId:          fmt.Sprintf("paginate-item-%d", i),
			OwnerAddress:    "aura1paginate",
			DataType:        drtypes.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
			ContentHash:     []byte(fmt.Sprintf("hash-%d", i)),
			StorageLocation: fmt.Sprintf("ipfs://paginate%d", i),
			Status:          drtypes.DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION,
			Title:           fmt.Sprintf("Document %d", i),
			Tags:            []string{"paginate"},
			CreatedAt:       createdAt,
			AccessPolicy:    &drtypes.AccessPolicy{},
		}
		require.NoError(t, k.SetDataItem(input.Ctx, item))
	}

	t.Run("NilPageRequest_ReturnsAll", func(t *testing.T) {
		req := &drtypes.QueryUserDataItemsRequest{
			OwnerAddress: "aura1paginate",
			Pagination:   nil,
		}

		resp, err := queryServer.UserDataItems(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.DataItems, 10)
		require.Nil(t, resp.Pagination)
	})

	t.Run("Offset_Basic", func(t *testing.T) {
		req := &drtypes.QueryUserDataItemsRequest{
			OwnerAddress: "aura1paginate",
			Pagination: &query.PageRequest{
				Offset: 3,
			},
		}

		resp, err := queryServer.UserDataItems(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.DataItems, 7) // 10 - 3 = 7
	})

	t.Run("Limit_Basic", func(t *testing.T) {
		req := &drtypes.QueryUserDataItemsRequest{
			OwnerAddress: "aura1paginate",
			Pagination: &query.PageRequest{
				Limit: 5,
			},
		}

		resp, err := queryServer.UserDataItems(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.DataItems, 5)
		require.NotNil(t, resp.Pagination)
		require.NotEmpty(t, resp.Pagination.NextKey) // More items available
	})

	t.Run("OffsetAndLimit_Combined", func(t *testing.T) {
		req := &drtypes.QueryUserDataItemsRequest{
			OwnerAddress: "aura1paginate",
			Pagination: &query.PageRequest{
				Offset: 2,
				Limit:  3,
			},
		}

		resp, err := queryServer.UserDataItems(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.DataItems, 3)
		require.NotNil(t, resp.Pagination)
		require.NotEmpty(t, resp.Pagination.NextKey) // More items (5 remaining after offset+limit=5)
	})

	t.Run("Offset_ExceedsTotal", func(t *testing.T) {
		req := &drtypes.QueryUserDataItemsRequest{
			OwnerAddress: "aura1paginate",
			Pagination: &query.PageRequest{
				Offset: 15, // Greater than 10 items
			},
		}

		resp, err := queryServer.UserDataItems(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Empty(t, resp.DataItems)
	})

	t.Run("Offset_EqualsTotal", func(t *testing.T) {
		req := &drtypes.QueryUserDataItemsRequest{
			OwnerAddress: "aura1paginate",
			Pagination: &query.PageRequest{
				Offset: 10, // Exactly 10 items
			},
		}

		resp, err := queryServer.UserDataItems(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Empty(t, resp.DataItems)
	})

	t.Run("Limit_ExceedsRemaining", func(t *testing.T) {
		req := &drtypes.QueryUserDataItemsRequest{
			OwnerAddress: "aura1paginate",
			Pagination: &query.PageRequest{
				Offset: 7,
				Limit:  10, // Only 3 items remaining
			},
		}

		resp, err := queryServer.UserDataItems(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.DataItems, 3)
		require.NotNil(t, resp.Pagination)
		require.Empty(t, resp.Pagination.NextKey) // No more items
	})

	t.Run("CountTotal_Enabled", func(t *testing.T) {
		req := &drtypes.QueryUserDataItemsRequest{
			OwnerAddress: "aura1paginate",
			Pagination: &query.PageRequest{
				Limit:      3,
				CountTotal: true,
			},
		}

		resp, err := queryServer.UserDataItems(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.DataItems, 3)
		require.NotNil(t, resp.Pagination)
		require.Equal(t, uint64(10), resp.Pagination.Total)
	})

	t.Run("CountTotal_Disabled", func(t *testing.T) {
		req := &drtypes.QueryUserDataItemsRequest{
			OwnerAddress: "aura1paginate",
			Pagination: &query.PageRequest{
				Limit:      3,
				CountTotal: false,
			},
		}

		resp, err := queryServer.UserDataItems(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.DataItems, 3)
		require.NotNil(t, resp.Pagination)
		require.Equal(t, uint64(0), resp.Pagination.Total)
	})

	t.Run("KeyBasedPagination_ReturnsError", func(t *testing.T) {
		req := &drtypes.QueryUserDataItemsRequest{
			OwnerAddress: "aura1paginate",
			Pagination: &query.PageRequest{
				Key: []byte("some-key"),
			},
		}

		resp, err := queryServer.UserDataItems(input.Ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "key-based pagination is not supported")
	})

	t.Run("LastPage_NoNextKey", func(t *testing.T) {
		req := &drtypes.QueryUserDataItemsRequest{
			OwnerAddress: "aura1paginate",
			Pagination: &query.PageRequest{
				Offset: 8,
				Limit:  5, // Only 2 remaining
			},
		}

		resp, err := queryServer.UserDataItems(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.DataItems, 2)
		require.NotNil(t, resp.Pagination)
		require.Empty(t, resp.Pagination.NextKey) // Last page
	})

	t.Run("ExactPage_NoNextKey", func(t *testing.T) {
		req := &drtypes.QueryUserDataItemsRequest{
			OwnerAddress: "aura1paginate",
			Pagination: &query.PageRequest{
				Offset: 5,
				Limit:  5, // Exactly remaining items
			},
		}

		resp, err := queryServer.UserDataItems(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.DataItems, 5)
		require.NotNil(t, resp.Pagination)
		require.Empty(t, resp.Pagination.NextKey) // No more items
	})

	t.Run("ZeroLimit_ReturnsAll", func(t *testing.T) {
		req := &drtypes.QueryUserDataItemsRequest{
			OwnerAddress: "aura1paginate",
			Pagination: &query.PageRequest{
				Limit: 0, // Zero limit means all items
			},
		}

		resp, err := queryServer.UserDataItems(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.DataItems, 10)
	})
}

func TestPaginateDataItemsViaSearchDataItems(t *testing.T) {
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

	// Create items with searchable tags
	createdAt, _ := types.TimestampProto(input.Ctx.BlockTime())
	for i := 0; i < 8; i++ {
		item := drtypes.DataItem{
			DataId:          fmt.Sprintf("search-item-%d", i),
			OwnerAddress:    "aura1search",
			DataType:        drtypes.DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:     []byte(fmt.Sprintf("search-hash-%d", i)),
			StorageLocation: fmt.Sprintf("ipfs://search%d", i),
			Status:          drtypes.DataItemStatus_DATA_ITEM_STATUS_VERIFIED,
			Title:           fmt.Sprintf("Search Photo %d", i),
			Tags:            []string{"searchable", "photos"},
			CreatedAt:       createdAt,
			AccessPolicy: &drtypes.AccessPolicy{
				Mode: drtypes.AccessMode_ACCESS_MODE_PUBLIC,
			},
		}
		require.NoError(t, k.SetDataItem(input.Ctx, item))
	}

	t.Run("SearchWithPagination", func(t *testing.T) {
		req := &drtypes.QuerySearchDataItemsRequest{
			Tags:      []string{"searchable"},
			Requester: "aura1search",
			Pagination: &query.PageRequest{
				Limit:      3,
				CountTotal: true,
			},
		}

		resp, err := queryServer.SearchDataItems(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.DataItems, 3)
		require.NotNil(t, resp.Pagination)
		require.Equal(t, uint64(8), resp.Pagination.Total)
		require.NotEmpty(t, resp.Pagination.NextKey)
	})

	t.Run("SearchWithOffset", func(t *testing.T) {
		req := &drtypes.QuerySearchDataItemsRequest{
			Tags:      []string{"photos"},
			Requester: "aura1search",
			Pagination: &query.PageRequest{
				Offset: 5,
				Limit:  10,
			},
		}

		resp, err := queryServer.SearchDataItems(input.Ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.DataItems, 3) // 8 - 5 = 3
	})

	t.Run("SearchKeyBasedPagination_ReturnsError", func(t *testing.T) {
		req := &drtypes.QuerySearchDataItemsRequest{
			Tags:      []string{"searchable"},
			Requester: "aura1search",
			Pagination: &query.PageRequest{
				Key: []byte("invalid-key"),
			},
		}

		resp, err := queryServer.SearchDataItems(input.Ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "key-based pagination is not supported")
	})
}
