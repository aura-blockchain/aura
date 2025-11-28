package keeper

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/aequitas/aura/chain/x/dataregistry/params"
	"github.com/aequitas/aura/chain/x/dataregistry/types"
	pb "github.com/aequitas/aura/proto/aura/dataregistry/v1beta1"
)

func TestQueryServer_DataItem(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(store)
	queryServer := NewQueryServer(keeper)
	ctx := context.Background()

	// Create a test data item
	ownerAddr := "aura1test"
	dataID, err := keeper.StoreDataItem(
		ownerAddr,
		types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Test Photo",
		"A test photo",
		[]byte("hash123"),
		"QmTest",
		false,
		nil,
		nil,
		&types.AccessPolicy{Mode: types.AccessMode_ACCESS_MODE_PUBLIC},
		[]string{"test"},
	)
	if err != nil {
		t.Fatalf("Failed to store data item: %v", err)
	}

	tests := []struct {
		name       string
		req        *pb.QueryDataItemRequest
		wantErr    bool
		wantExists bool
		wantAccess bool
		wantDataID string
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name:    "empty data_id",
			req:     &pb.QueryDataItemRequest{DataId: "", Requester: "aura1other"},
			wantErr: true,
		},
		{
			name:       "non-existent data item",
			req:        &pb.QueryDataItemRequest{DataId: "data:nonexistent", Requester: "aura1other"},
			wantErr:    false,
			wantExists: false,
			wantAccess: false,
		},
		{
			name:       "public access - owner",
			req:        &pb.QueryDataItemRequest{DataId: dataID, Requester: ownerAddr},
			wantErr:    false,
			wantExists: true,
			wantAccess: true,
			wantDataID: dataID,
		},
		{
			name:       "public access - other user",
			req:        &pb.QueryDataItemRequest{DataId: dataID, Requester: "aura1other"},
			wantErr:    false,
			wantExists: true,
			wantAccess: true,
			wantDataID: dataID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.DataItem(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("DataItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if resp.Exists != tt.wantExists {
				t.Errorf("DataItem() exists = %v, want %v", resp.Exists, tt.wantExists)
			}
			if resp.HasAccess != tt.wantAccess {
				t.Errorf("DataItem() hasAccess = %v, want %v", resp.HasAccess, tt.wantAccess)
			}
			if tt.wantAccess && resp.DataItem != nil && resp.DataItem.DataId != tt.wantDataID {
				t.Errorf("DataItem() dataID = %v, want %v", resp.DataItem.DataId, tt.wantDataID)
			}
		})
	}
}

func TestQueryServer_UserDataItems(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(store)
	queryServer := NewQueryServer(keeper)
	ctx := context.Background()

	// Create test data items
	ownerAddr := "aura1test"
	_, err := keeper.StoreDataItem(
		ownerAddr,
		types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Photo 1",
		"First photo",
		[]byte("hash1"),
		"QmTest1",
		false,
		nil,
		nil,
		&types.AccessPolicy{Mode: types.AccessMode_ACCESS_MODE_PUBLIC},
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to store data item: %v", err)
	}

	_, err = keeper.StoreDataItem(
		ownerAddr,
		types.DataItemType_DATA_ITEM_TYPE_VIDEO,
		"Video 1",
		"First video",
		[]byte("hash2"),
		"QmTest2",
		false,
		nil,
		nil,
		&types.AccessPolicy{Mode: types.AccessMode_ACCESS_MODE_PUBLIC},
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to store data item: %v", err)
	}

	tests := []struct {
		name      string
		req       *pb.QueryUserDataItemsRequest
		wantErr   bool
		wantCount int
		wantTotal uint64
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name:    "empty owner_address",
			req:     &pb.QueryUserDataItemsRequest{OwnerAddress: ""},
			wantErr: true,
		},
		{
			name:      "all items",
			req:       &pb.QueryUserDataItemsRequest{OwnerAddress: ownerAddr},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "filter by type",
			req: &pb.QueryUserDataItemsRequest{
				OwnerAddress: ownerAddr,
				TypeFilter:   types.DataItemType_DATA_ITEM_TYPE_PHOTO,
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:      "non-existent user",
			req:       &pb.QueryUserDataItemsRequest{OwnerAddress: "aura1nonexistent"},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "pagination limit one",
			req: &pb.QueryUserDataItemsRequest{
				OwnerAddress: ownerAddr,
				Pagination: &query.PageRequest{
					Limit:      1,
					CountTotal: true,
				},
			},
			wantErr:   false,
			wantCount: 1,
			wantTotal: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.UserDataItems(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserDataItems() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if len(resp.DataItems) != tt.wantCount {
				t.Errorf("UserDataItems() count = %v, want %v", len(resp.DataItems), tt.wantCount)
			}
			if tt.wantTotal > 0 {
				if resp.Pagination == nil {
					t.Errorf("UserDataItems() pagination expected but got nil")
					return
				}
				if resp.Pagination.Total != tt.wantTotal {
					t.Errorf("UserDataItems() total = %v, want %v", resp.Pagination.Total, tt.wantTotal)
				}
			}
		})
	}
}

func TestQueryServer_SearchDataItems(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(store)
	queryServer := NewQueryServer(keeper)
	ctx := context.Background()

	// Create test data items with tags
	ownerAddr := "aura1test"
	_, err := keeper.StoreDataItem(
		ownerAddr,
		types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Beach Photo",
		"A photo of the beach",
		[]byte("hash1"),
		"QmTest1",
		false,
		nil,
		nil,
		&types.AccessPolicy{Mode: types.AccessMode_ACCESS_MODE_PUBLIC},
		[]string{"beach", "vacation"},
	)
	if err != nil {
		t.Fatalf("Failed to store data item: %v", err)
	}

	_, err = keeper.StoreDataItem(
		ownerAddr,
		types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Mountain Photo",
		"A photo of mountains",
		[]byte("hash2"),
		"QmTest2",
		false,
		nil,
		nil,
		&types.AccessPolicy{Mode: types.AccessMode_ACCESS_MODE_PRIVATE},
		[]string{"mountain", "hiking"},
	)
	if err != nil {
		t.Fatalf("Failed to store data item: %v", err)
	}

	tests := []struct {
		name      string
		req       *pb.QuerySearchDataItemsRequest
		wantErr   bool
		wantCount int
		wantTotal uint64
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "search with tag - public access",
			req: &pb.QuerySearchDataItemsRequest{
				Tags:      []string{"beach"},
				Requester: "aura1other",
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "search with tag - no access to private",
			req: &pb.QuerySearchDataItemsRequest{
				Tags:      []string{"mountain"},
				Requester: "aura1other",
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "search with tag - owner access",
			req: &pb.QuerySearchDataItemsRequest{
				Tags:      []string{"mountain"},
				Requester: ownerAddr,
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "search by type",
			req: &pb.QuerySearchDataItemsRequest{
				TypeFilter: types.DataItemType_DATA_ITEM_TYPE_PHOTO,
				Requester:  ownerAddr,
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "pagination limit one count total",
			req: &pb.QuerySearchDataItemsRequest{
				Tags:      []string{"beach"},
				Requester: ownerAddr,
				Pagination: &query.PageRequest{
					Limit:      1,
					CountTotal: true,
				},
			},
			wantErr:   false,
			wantCount: 1,
			wantTotal: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.SearchDataItems(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchDataItems() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if len(resp.DataItems) != tt.wantCount {
				t.Errorf("SearchDataItems() count = %v, want %v", len(resp.DataItems), tt.wantCount)
			}
			if tt.wantTotal > 0 {
				if resp.Pagination == nil {
					t.Errorf("SearchDataItems() pagination expected but got nil")
					return
				}
				if resp.Pagination.Total != tt.wantTotal {
					t.Errorf("SearchDataItems() total = %v, want %v", resp.Pagination.Total, tt.wantTotal)
				}
			}
		})
	}
}

func TestQueryServer_DataItemVerifications(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(store)
	queryServer := NewQueryServer(keeper)
	ctx := context.Background()

	// Create a test data item
	ownerAddr := "aura1test"
	dataID, err := keeper.StoreDataItem(
		ownerAddr,
		types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Test Photo",
		"A test photo",
		[]byte("hash123"),
		"QmTest",
		false,
		nil,
		nil,
		&types.AccessPolicy{Mode: types.AccessMode_ACCESS_MODE_PUBLIC},
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to store data item: %v", err)
	}

	// Add a verification
	err = keeper.VerifyDataItem(
		dataID,
		"aura1verifier",
		types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
		95,
		"Looks good",
		"Manual review",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to verify data item: %v", err)
	}

	tests := []struct {
		name      string
		req       *pb.QueryDataItemVerificationsRequest
		wantErr   bool
		wantCount int
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name:    "empty data_id",
			req:     &pb.QueryDataItemVerificationsRequest{DataId: ""},
			wantErr: true,
		},
		{
			name:      "existing verifications",
			req:       &pb.QueryDataItemVerificationsRequest{DataId: dataID},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:      "non-existent data item",
			req:       &pb.QueryDataItemVerificationsRequest{DataId: "data:nonexistent"},
			wantErr:   true,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.DataItemVerifications(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("DataItemVerifications() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if len(resp.Verifications) != tt.wantCount {
				t.Errorf("DataItemVerifications() count = %v, want %v", len(resp.Verifications), tt.wantCount)
			}
		})
	}
}

func TestQueryServer_Stats(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(store)
	queryServer := NewQueryServer(keeper)
	ctx := context.Background()

	// Create test data items
	ownerAddr := "aura1test"
	dataID, err := keeper.StoreDataItem(
		ownerAddr,
		types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Photo 1",
		"First photo",
		[]byte("hash1"),
		"QmTest1",
		false,
		nil,
		nil,
		&types.AccessPolicy{Mode: types.AccessMode_ACCESS_MODE_PUBLIC},
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to store data item: %v", err)
	}

	// Verify one item
	err = keeper.VerifyDataItem(
		dataID,
		"aura1verifier",
		types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
		95,
		"Verified",
		"Manual review",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to verify data item: %v", err)
	}

	tests := []struct {
		name                   string
		req                    *pb.QueryStatsRequest
		wantErr                bool
		wantTotalDataItems     uint64
		wantTotalVerifiedItems uint64
		wantTotalVerifications uint64
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name:                   "valid request",
			req:                    &pb.QueryStatsRequest{},
			wantErr:                false,
			wantTotalDataItems:     1,
			wantTotalVerifiedItems: 1,
			wantTotalVerifications: 1,
		},
	}

	expectedStats := keeper.GetStats()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.Stats(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Stats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if resp.TotalDataItems != tt.wantTotalDataItems {
				t.Errorf("Stats() TotalDataItems = %v, want %v", resp.TotalDataItems, tt.wantTotalDataItems)
			}
			if resp.TotalVerifiedItems != tt.wantTotalVerifiedItems {
				t.Errorf("Stats() TotalVerifiedItems = %v, want %v", resp.TotalVerifiedItems, tt.wantTotalVerifiedItems)
			}
			if resp.TotalVerifications != tt.wantTotalVerifications {
				t.Errorf("Stats() TotalVerifications = %v, want %v", resp.TotalVerifications, tt.wantTotalVerifications)
			}
			if resp.TotalStorageBytes != expectedStats.TotalStorageBytes {
				t.Errorf("Stats() TotalStorageBytes = %v, want %v", resp.TotalStorageBytes, expectedStats.TotalStorageBytes)
			}
		})
	}
}

func TestQueryServer_Params(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(store)
	queryServer := NewQueryServer(keeper)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     *pb.QueryParamsRequest
		wantErr bool
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name:    "valid request",
			req:     &pb.QueryParamsRequest{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.Params(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Params() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if resp.Params == nil {
				t.Error("Params() returned nil params")
			}
			if resp.Params.MaxDataItemsPerUser == 0 {
				t.Error("Params() returned invalid max_data_items_per_user")
			}
		})
	}
}

func TestQueryServer_PrivateAccessControl(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(store)
	queryServer := NewQueryServer(keeper)
	ctx := context.Background()

	// Create a private data item
	ownerAddr := "aura1owner"
	dataID, err := keeper.StoreDataItem(
		ownerAddr,
		types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Private Photo",
		"A private photo",
		[]byte("hash123"),
		"QmPrivate",
		false,
		nil,
		nil,
		&types.AccessPolicy{Mode: types.AccessMode_ACCESS_MODE_PRIVATE},
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to store data item: %v", err)
	}

	// Test owner access
	resp, err := queryServer.DataItem(ctx, &pb.QueryDataItemRequest{
		DataId:    dataID,
		Requester: ownerAddr,
	})
	if err != nil {
		t.Fatalf("Owner access failed: %v", err)
	}
	if !resp.HasAccess {
		t.Error("Owner should have access to private item")
	}
	if resp.DataItem == nil {
		t.Error("Owner should receive data item")
	}

	// Test non-owner access
	resp, err = queryServer.DataItem(ctx, &pb.QueryDataItemRequest{
		DataId:    dataID,
		Requester: "aura1other",
	})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if resp.HasAccess {
		t.Error("Non-owner should not have access to private item")
	}
	if resp.DataItem != nil {
		t.Error("Non-owner should not receive data item")
	}
}

func TestQueryServer_WhitelistAccessControl(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(store)
	queryServer := NewQueryServer(keeper)
	ctx := context.Background()

	// Create a whitelist data item
	ownerAddr := "aura1owner"
	allowedAddr := "aura1allowed"
	dataID, err := keeper.StoreDataItem(
		ownerAddr,
		types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Whitelist Photo",
		"A whitelist photo",
		[]byte("hash123"),
		"QmWhitelist",
		false,
		nil,
		nil,
		&types.AccessPolicy{
			Mode:             types.AccessMode_ACCESS_MODE_WHITELIST,
			AllowedAddresses: []string{allowedAddr},
		},
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to store data item: %v", err)
	}

	// Test whitelisted user access
	resp, err := queryServer.DataItem(ctx, &pb.QueryDataItemRequest{
		DataId:    dataID,
		Requester: allowedAddr,
	})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if !resp.HasAccess {
		t.Error("Whitelisted user should have access")
	}

	// Test non-whitelisted user access
	resp, err = queryServer.DataItem(ctx, &pb.QueryDataItemRequest{
		DataId:    dataID,
		Requester: "aura1other",
	})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if resp.HasAccess {
		t.Error("Non-whitelisted user should not have access")
	}
}
