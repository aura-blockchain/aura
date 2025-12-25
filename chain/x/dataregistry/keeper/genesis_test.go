// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"
	"time"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/dataregistry/keeper"
	"github.com/aequitas/aura/chain/x/dataregistry/params"
	"github.com/aequitas/aura/chain/x/dataregistry/types"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"
)

func TestInitGenesis(t *testing.T) {
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

	// Create genesis state with sample data
	defaultParams := types.DefaultParams()
	now, _ := gogotypes.TimestampProto(time.Now())
	genesis := types.GenesisState{
		Params: &defaultParams,
		DataItems: []*types.DataItem{
			{
				DataId:          "genesis-data-1",
				OwnerAddress:    "aura1owner",
				DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
				ContentHash:     []byte("hash1"),
				StorageLocation: "ipfs://genesis1",
				Status:          types.DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION,
				CreatedAt:       now,
				AccessPolicy: &types.AccessPolicy{
					Mode: types.AccessMode_ACCESS_MODE_PRIVATE,
				},
			},
			{
				DataId:          "genesis-data-2",
				OwnerAddress:    "aura1owner2",
				DataType:        types.DataItemType_DATA_ITEM_TYPE_VIDEO,
				ContentHash:     []byte("hash2"),
				StorageLocation: "ipfs://genesis2",
				Status:          types.DataItemStatus_DATA_ITEM_STATUS_VERIFIED,
				CreatedAt:       now,
				AccessPolicy: &types.AccessPolicy{
					Mode: types.AccessMode_ACCESS_MODE_PRIVATE,
				},
			},
		},
		NextDataId: 100,
	}

	// Initialize genesis
	err := k.InitGenesis(input.Ctx, genesis)
	require.NoError(t, err)

	// Verify params were set
	retrievedParams, _ := k.GetParams(input.Ctx)
	require.Equal(t, defaultParams.MaxStorageBytes, retrievedParams.MaxStorageBytes)

	// Verify data items were stored
	item1, found := k.GetDataItem(input.Ctx, "genesis-data-1")
	require.True(t, found)
	require.Equal(t, "aura1owner", item1.OwnerAddress)

	item2, found := k.GetDataItem(input.Ctx, "genesis-data-2")
	require.True(t, found)
	require.Equal(t, "aura1owner2", item2.OwnerAddress)

	// Verify next data ID was set
	nextID := k.GetNextDataID(input.Ctx)
	require.Equal(t, uint64(100), nextID)
}

func TestExportGenesis(t *testing.T) {
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

	// Add some data items
	item1 := types.DataItem{
		DataId:          "export-data-1",
		OwnerAddress:    "aura1owner",
		DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		ContentHash:     []byte("hash1"),
		StorageLocation: "ipfs://export1",
		Status:          types.DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION,
	}

	item2 := types.DataItem{
		DataId:          "export-data-2",
		OwnerAddress:    "aura1owner2",
		DataType:        types.DataItemType_DATA_ITEM_TYPE_VIDEO,
		ContentHash:     []byte("hash2"),
		StorageLocation: "ipfs://export2",
		Status:          types.DataItemStatus_DATA_ITEM_STATUS_VERIFIED,
	}

	require.NoError(t, k.SetDataItem(input.Ctx, item1))
	require.NoError(t, k.SetDataItem(input.Ctx, item2))

	// Set next data ID
	require.NoError(t, k.SetNextDataID(input.Ctx, 42))

	// Export genesis
	exported := k.ExportGenesis(input.Ctx)

	// Verify exported data
	require.NotNil(t, exported.Params)
	require.Len(t, exported.DataItems, 2)
	require.Equal(t, uint64(42), exported.NextDataId)

	// Verify data items are in export
	foundItem1 := false
	foundItem2 := false
	for _, item := range exported.DataItems {
		if item.DataId == "export-data-1" {
			foundItem1 = true
			require.Equal(t, "aura1owner", item.OwnerAddress)
		}
		if item.DataId == "export-data-2" {
			foundItem2 = true
			require.Equal(t, "aura1owner2", item.OwnerAddress)
		}
	}
	require.True(t, foundItem1, "export-data-1 not found in export")
	require.True(t, foundItem2, "export-data-2 not found in export")
}

func TestGenesisRoundTrip(t *testing.T) {
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

	// Create original genesis state
	defaultParams := types.DefaultParams()
	defaultParams.MaxStorageBytes = 5000000
	now, _ := gogotypes.TimestampProto(time.Now())
	original := types.GenesisState{
		Params: &defaultParams,
		DataItems: []*types.DataItem{
			{
				DataId:          "roundtrip-data-1",
				OwnerAddress:    "aura1owner",
				DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
				ContentHash:     []byte("hash1"),
				StorageLocation: "ipfs://roundtrip1",
				Status:          types.DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION,
				Title:           "Test Photo",
				Tags:            []string{"test", "roundtrip"},
				CreatedAt:       now,
				AccessPolicy: &types.AccessPolicy{
					Mode: types.AccessMode_ACCESS_MODE_PRIVATE,
				},
			},
		},
		NextDataId: 999,
	}

	// Initialize genesis
	err := k.InitGenesis(input.Ctx, original)
	require.NoError(t, err)

	// Export genesis
	exported := k.ExportGenesis(input.Ctx)

	// Compare exported with original
	require.Equal(t, original.Params.MaxStorageBytes, exported.Params.MaxStorageBytes)
	require.Equal(t, original.NextDataId, exported.NextDataId)
	require.Len(t, exported.DataItems, len(original.DataItems))

	// Verify data item details match
	require.Equal(t, original.DataItems[0].DataId, exported.DataItems[0].DataId)
	require.Equal(t, original.DataItems[0].OwnerAddress, exported.DataItems[0].OwnerAddress)
	require.Equal(t, original.DataItems[0].Title, exported.DataItems[0].Title)
}

func TestDefaultGenesis(t *testing.T) {
	// Test that default genesis is valid
	require.NotPanics(t, func() {
		genesis := types.DefaultGenesisState()
		require.NotNil(t, genesis)
		require.NotNil(t, genesis.Params)
		require.NotNil(t, genesis.DataItems)
		require.Equal(t, uint64(1), genesis.NextDataId)
	})
}

func TestInitGenesis_WithVerifications(t *testing.T) {
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

	// Create genesis state with verified data
	defaultParams := types.DefaultParams()
	now, _ := gogotypes.TimestampProto(time.Now())
	genesis := types.GenesisState{
		Params: &defaultParams,
		DataItems: []*types.DataItem{
			{
				DataId:          "verified-data-1",
				OwnerAddress:    "aura1owner",
				DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
				ContentHash:     []byte("hash1"),
				StorageLocation: "ipfs://verified1",
				Status:          types.DataItemStatus_DATA_ITEM_STATUS_VERIFIED,
				CreatedAt:       now,
				AccessPolicy: &types.AccessPolicy{
					Mode: types.AccessMode_ACCESS_MODE_PRIVATE,
				},
				Verifications: []*types.Verification{
					{
						VerifierAddress:    "aura1verifier1",
						VerificationMethod: "manual",
						ConfidenceScore:    95,
						Level:              types.VerificationLevel_VERIFICATION_LEVEL_AUTHORITY_VERIFIED,
						VerifiedAt:         now,
					},
					{
						VerifierAddress:    "aura1verifier2",
						VerificationMethod: "automated",
						ConfidenceScore:    85,
						Level:              types.VerificationLevel_VERIFICATION_LEVEL_AI_VERIFIED,
						VerifiedAt:         now,
					},
				},
			},
		},
		NextDataId: 1,
	}

	// Initialize genesis
	err := k.InitGenesis(input.Ctx, genesis)
	require.NoError(t, err)

	// Verify data item with verifications was stored correctly
	item, found := k.GetDataItem(input.Ctx, "verified-data-1")
	require.True(t, found)
	require.Equal(t, types.DataItemStatus_DATA_ITEM_STATUS_VERIFIED, item.Status)
	require.Len(t, item.Verifications, 2)
	require.Equal(t, "aura1verifier1", item.Verifications[0].VerifierAddress)
	require.Equal(t, "aura1verifier2", item.Verifications[1].VerifierAddress)
}

func TestInitGenesis_WithTypeSpecificData(t *testing.T) {
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

	// Create genesis state with different data types
	defaultParams := types.DefaultParams()
	now, _ := gogotypes.TimestampProto(time.Now())
	genesis := types.GenesisState{
		Params: &defaultParams,
		DataItems: []*types.DataItem{
			{
				DataId:          "photo-1",
				OwnerAddress:    "aura1owner",
				DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
				ContentHash:     []byte("photo_hash"),
				StorageLocation: "ipfs://photo",
				Status:          types.DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION,
				CreatedAt:       now,
				AccessPolicy: &types.AccessPolicy{
					Mode: types.AccessMode_ACCESS_MODE_PRIVATE,
				},
			},
			{
				DataId:          "video-1",
				OwnerAddress:    "aura1owner",
				DataType:        types.DataItemType_DATA_ITEM_TYPE_VIDEO,
				ContentHash:     []byte("video_hash"),
				StorageLocation: "ipfs://video",
				Status:          types.DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION,
				CreatedAt:       now,
				AccessPolicy: &types.AccessPolicy{
					Mode: types.AccessMode_ACCESS_MODE_PRIVATE,
				},
			},
			{
				DataId:          "audio-1",
				OwnerAddress:    "aura1owner",
				DataType:        types.DataItemType_DATA_ITEM_TYPE_AUDIO,
				ContentHash:     []byte("audio_hash"),
				StorageLocation: "ipfs://audio",
				Status:          types.DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION,
				CreatedAt:       now,
				AccessPolicy: &types.AccessPolicy{
					Mode: types.AccessMode_ACCESS_MODE_PRIVATE,
				},
			},
		},
		NextDataId: 1,
	}

	// Initialize genesis
	err := k.InitGenesis(input.Ctx, genesis)
	require.NoError(t, err)

	// Verify each type was stored correctly
	photo, found := k.GetDataItem(input.Ctx, "photo-1")
	require.True(t, found)
	require.Equal(t, types.DataItemType_DATA_ITEM_TYPE_PHOTO, photo.DataType)

	video, found := k.GetDataItem(input.Ctx, "video-1")
	require.True(t, found)
	require.Equal(t, types.DataItemType_DATA_ITEM_TYPE_VIDEO, video.DataType)

	audio, found := k.GetDataItem(input.Ctx, "audio-1")
	require.True(t, found)
	require.Equal(t, types.DataItemType_DATA_ITEM_TYPE_AUDIO, audio.DataType)
}
