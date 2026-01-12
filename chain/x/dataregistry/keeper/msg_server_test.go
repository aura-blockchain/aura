// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

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

func TestMsgServerVerifyDataItem(t *testing.T) {
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

	t.Run("VerifyDataItem_Success", func(t *testing.T) {
		// First store a data item
		storeMsg := &types.MsgStoreDataItem{
			Creator:         "aura1owner",
			DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:     []byte("hash_verify1"),
			StorageLocation: "ipfs://verify1",
			Title:           "Photo to Verify",
		}

		storeResp, err := msgServer.StoreDataItem(input.Ctx, storeMsg)
		require.NoError(t, err)

		// Now verify it
		verifyMsg := &types.MsgVerifyDataItem{
			Verifier:           "aura1verifier",
			DataId:             storeResp.DataId,
			Level:              types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
			ConfidenceScore:    85,
			Notes:              "Verified by peer review",
			VerificationMethod: "manual_review",
			Proof:              []byte("signature_proof"),
		}

		verifyResp, err := msgServer.VerifyDataItem(input.Ctx, verifyMsg)
		require.NoError(t, err)
		require.NotNil(t, verifyResp)
		require.NotNil(t, verifyResp.VerifiedAt)

		// Verify the item was updated
		item, found := k.GetDataItem(input.Ctx, storeResp.DataId)
		require.True(t, found)
		require.Equal(t, types.DataItemStatus_DATA_ITEM_STATUS_VERIFIED, item.Status)
		require.Equal(t, types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED, item.VerificationLevel)
		require.Len(t, item.Verifications, 1)
		require.Equal(t, "aura1verifier", item.Verifications[0].VerifierAddress)
		require.Equal(t, uint64(85), item.Verifications[0].ConfidenceScore)
	})

	t.Run("VerifyDataItem_NilMessage", func(t *testing.T) {
		_, err := msgServer.VerifyDataItem(input.Ctx, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "message cannot be nil")
	})

	t.Run("VerifyDataItem_EmptyVerifier", func(t *testing.T) {
		verifyMsg := &types.MsgVerifyDataItem{
			Verifier:           "",
			DataId:             "some-data-id",
			Level:              types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
			ConfidenceScore:    80,
			VerificationMethod: "manual_review",
		}

		_, err := msgServer.VerifyDataItem(input.Ctx, verifyMsg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "verifier cannot be empty")
	})

	t.Run("VerifyDataItem_EmptyDataId", func(t *testing.T) {
		verifyMsg := &types.MsgVerifyDataItem{
			Verifier:           "aura1verifier",
			DataId:             "",
			Level:              types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
			ConfidenceScore:    80,
			VerificationMethod: "manual_review",
		}

		_, err := msgServer.VerifyDataItem(input.Ctx, verifyMsg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "data_id cannot be empty")
	})

	t.Run("VerifyDataItem_InvalidVerificationLevel", func(t *testing.T) {
		verifyMsg := &types.MsgVerifyDataItem{
			Verifier:           "aura1verifier",
			DataId:             "some-data-id",
			Level:              types.VerificationLevel_VERIFICATION_LEVEL_UNSPECIFIED,
			ConfidenceScore:    80,
			VerificationMethod: "manual_review",
		}

		_, err := msgServer.VerifyDataItem(input.Ctx, verifyMsg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid verification level")
	})

	t.Run("VerifyDataItem_ConfidenceScoreExceeds100", func(t *testing.T) {
		verifyMsg := &types.MsgVerifyDataItem{
			Verifier:           "aura1verifier",
			DataId:             "some-data-id",
			Level:              types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
			ConfidenceScore:    150,
			VerificationMethod: "manual_review",
		}

		_, err := msgServer.VerifyDataItem(input.Ctx, verifyMsg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "confidence_score must be between 0 and 100")
	})

	t.Run("VerifyDataItem_EmptyVerificationMethod", func(t *testing.T) {
		verifyMsg := &types.MsgVerifyDataItem{
			Verifier:           "aura1verifier",
			DataId:             "some-data-id",
			Level:              types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
			ConfidenceScore:    80,
			VerificationMethod: "",
		}

		_, err := msgServer.VerifyDataItem(input.Ctx, verifyMsg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "verification_method cannot be empty")
	})

	t.Run("VerifyDataItem_ItemNotFound", func(t *testing.T) {
		verifyMsg := &types.MsgVerifyDataItem{
			Verifier:           "aura1verifier",
			DataId:             "nonexistent-data-id",
			Level:              types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
			ConfidenceScore:    80,
			VerificationMethod: "manual_review",
		}

		_, err := msgServer.VerifyDataItem(input.Ctx, verifyMsg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "data item not found")
	})

	t.Run("VerifyDataItem_ItemRevoked", func(t *testing.T) {
		// Store a data item
		storeMsg := &types.MsgStoreDataItem{
			Creator:         "aura1owner",
			DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:     []byte("hash_verify_revoked"),
			StorageLocation: "ipfs://verify_revoked",
			Title:           "Photo to be Revoked",
		}

		storeResp, err := msgServer.StoreDataItem(input.Ctx, storeMsg)
		require.NoError(t, err)

		// Revoke it using keeper directly
		err = k.RevokeDataItem(input.Ctx, storeResp.DataId, "aura1authority", "test revocation")
		require.NoError(t, err)

		// Try to verify the revoked item
		verifyMsg := &types.MsgVerifyDataItem{
			Verifier:           "aura1verifier",
			DataId:             storeResp.DataId,
			Level:              types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
			ConfidenceScore:    80,
			VerificationMethod: "manual_review",
		}

		_, err = msgServer.VerifyDataItem(input.Ctx, verifyMsg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "data item is revoked")
	})

	t.Run("VerifyDataItem_AllVerificationLevels", func(t *testing.T) {
		levels := []types.VerificationLevel{
			types.VerificationLevel_VERIFICATION_LEVEL_SELF_ATTESTED,
			types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
			types.VerificationLevel_VERIFICATION_LEVEL_AI_VERIFIED,
			types.VerificationLevel_VERIFICATION_LEVEL_AUTHORITY_VERIFIED,
			types.VerificationLevel_VERIFICATION_LEVEL_BLOCKCHAIN_ANCHORED,
		}

		for i, level := range levels {
			// Store a new data item for each level
			storeMsg := &types.MsgStoreDataItem{
				Creator:         "aura1owner",
				DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
				ContentHash:     []byte("hash_level_" + string(rune(i))),
				StorageLocation: "ipfs://level_" + string(rune(i)),
				Title:           "Photo for Level Test",
			}

			storeResp, err := msgServer.StoreDataItem(input.Ctx, storeMsg)
			require.NoError(t, err)

			// Verify with the level
			verifyMsg := &types.MsgVerifyDataItem{
				Verifier:           "aura1verifier",
				DataId:             storeResp.DataId,
				Level:              level,
				ConfidenceScore:    90,
				VerificationMethod: "test_method",
			}

			_, err = msgServer.VerifyDataItem(input.Ctx, verifyMsg)
			require.NoError(t, err)
		}
	})
}

func TestMsgServerStoreDataItem_Validation(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	paramsStore := params.NewStore(types.DefaultParams())
	k := keeper.NewKeeper(
		keepertest.WrapStoreService(input.StoreKey),
		input.Cdc,
		paramsStore,
		"aura1authority",
		keepertest.Logger(),
	)
	msgServer := keeper.NewMsgServer(k)

	t.Run("NilMessage", func(t *testing.T) {
		_, err := msgServer.StoreDataItem(input.Ctx, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "message cannot be nil")
	})

	t.Run("EmptyCreator", func(t *testing.T) {
		msg := &types.MsgStoreDataItem{
			Creator:         "",
			DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:     []byte("hash"),
			StorageLocation: "ipfs://test",
			Title:           "Test",
		}
		_, err := msgServer.StoreDataItem(input.Ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "creator cannot be empty")
	})

	t.Run("UnspecifiedDataType", func(t *testing.T) {
		msg := &types.MsgStoreDataItem{
			Creator:         "aura1owner",
			DataType:        types.DataItemType_DATA_ITEM_TYPE_UNSPECIFIED,
			ContentHash:     []byte("hash"),
			StorageLocation: "ipfs://test",
			Title:           "Test",
		}
		_, err := msgServer.StoreDataItem(input.Ctx, msg)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidDataType)
	})

	t.Run("EmptyTitle", func(t *testing.T) {
		msg := &types.MsgStoreDataItem{
			Creator:         "aura1owner",
			DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:     []byte("hash"),
			StorageLocation: "ipfs://test",
			Title:           "",
		}
		_, err := msgServer.StoreDataItem(input.Ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "title cannot be empty")
	})

	t.Run("EmptyContentHash", func(t *testing.T) {
		msg := &types.MsgStoreDataItem{
			Creator:         "aura1owner",
			DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:     []byte{},
			StorageLocation: "ipfs://test",
			Title:           "Test",
		}
		_, err := msgServer.StoreDataItem(input.Ctx, msg)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidContentHash)
	})

	t.Run("EmptyStorageLocation", func(t *testing.T) {
		msg := &types.MsgStoreDataItem{
			Creator:         "aura1owner",
			DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:     []byte("hash"),
			StorageLocation: "",
			Title:           "Test",
		}
		_, err := msgServer.StoreDataItem(input.Ctx, msg)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidStorageLocation)
	})

	t.Run("WhitelistModeWithoutAllowedAddresses", func(t *testing.T) {
		msg := &types.MsgStoreDataItem{
			Creator:         "aura1owner",
			DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:     []byte("hash"),
			StorageLocation: "ipfs://test",
			Title:           "Test",
			AccessPolicy: &types.AccessPolicy{
				Mode:             types.AccessMode_ACCESS_MODE_WHITELIST,
				AllowedAddresses: []string{}, // Empty whitelist
			},
		}
		_, err := msgServer.StoreDataItem(input.Ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "whitelist mode requires at least one allowed address")
	})

	t.Run("PrivateModeWithAllowedAddresses", func(t *testing.T) {
		msg := &types.MsgStoreDataItem{
			Creator:         "aura1owner",
			DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:     []byte("hash"),
			StorageLocation: "ipfs://test",
			Title:           "Test",
			AccessPolicy: &types.AccessPolicy{
				Mode:             types.AccessMode_ACCESS_MODE_PRIVATE,
				AllowedAddresses: []string{"aura1addr"}, // Should not have addresses in private mode
			},
		}
		_, err := msgServer.StoreDataItem(input.Ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "private mode should not have allowed addresses")
	})
}

func TestMsgServerUpdateDataItem_Validation(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	paramsStore := params.NewStore(types.DefaultParams())
	k := keeper.NewKeeper(
		keepertest.WrapStoreService(input.StoreKey),
		input.Cdc,
		paramsStore,
		"aura1authority",
		keepertest.Logger(),
	)
	msgServer := keeper.NewMsgServer(k)

	t.Run("NilMessage", func(t *testing.T) {
		_, err := msgServer.UpdateDataItem(input.Ctx, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "message cannot be nil")
	})

	t.Run("EmptyCreator", func(t *testing.T) {
		msg := &types.MsgUpdateDataItem{
			Creator: "",
			DataId:  "test-id",
			Title:   "Updated Title",
		}
		_, err := msgServer.UpdateDataItem(input.Ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "creator cannot be empty")
	})

	t.Run("EmptyDataId", func(t *testing.T) {
		msg := &types.MsgUpdateDataItem{
			Creator: "aura1owner",
			DataId:  "",
			Title:   "Updated Title",
		}
		_, err := msgServer.UpdateDataItem(input.Ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "data_id cannot be empty")
	})

	t.Run("NoFieldsToUpdate", func(t *testing.T) {
		msg := &types.MsgUpdateDataItem{
			Creator:     "aura1owner",
			DataId:      "test-id",
			Title:       "",
			Description: "",
			Metadata:    nil,
			Tags:        nil,
		}
		_, err := msgServer.UpdateDataItem(input.Ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "at least one field must be provided for update")
	})

	t.Run("WhitelistModeWithoutAddresses", func(t *testing.T) {
		msg := &types.MsgUpdateDataItem{
			Creator: "aura1owner",
			DataId:  "test-id",
			Title:   "New Title",
			AccessPolicy: &types.AccessPolicy{
				Mode:             types.AccessMode_ACCESS_MODE_WHITELIST,
				AllowedAddresses: []string{},
			},
		}
		_, err := msgServer.UpdateDataItem(input.Ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "whitelist mode requires at least one allowed address")
	})
}

func TestMsgServerDeleteDataItem_Validation(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	paramsStore := params.NewStore(types.DefaultParams())
	k := keeper.NewKeeper(
		keepertest.WrapStoreService(input.StoreKey),
		input.Cdc,
		paramsStore,
		"aura1authority",
		keepertest.Logger(),
	)
	msgServer := keeper.NewMsgServer(k)

	t.Run("NilMessage", func(t *testing.T) {
		_, err := msgServer.DeleteDataItem(input.Ctx, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "message cannot be nil")
	})

	t.Run("EmptyCreator", func(t *testing.T) {
		msg := &types.MsgDeleteDataItem{
			Creator: "",
			DataId:  "test-id",
		}
		_, err := msgServer.DeleteDataItem(input.Ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "creator cannot be empty")
	})

	t.Run("EmptyDataId", func(t *testing.T) {
		msg := &types.MsgDeleteDataItem{
			Creator: "aura1owner",
			DataId:  "",
		}
		_, err := msgServer.DeleteDataItem(input.Ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "data_id cannot be empty")
	})

	t.Run("ItemNotFound", func(t *testing.T) {
		msg := &types.MsgDeleteDataItem{
			Creator: "aura1owner",
			DataId:  "nonexistent-id",
		}
		_, err := msgServer.DeleteDataItem(input.Ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "data item not found")
	})
}

func TestMsgServerRevokeDataItem(t *testing.T) {
	input := keepertest.CreateTestInput(t)

	// Create params with authorized verifier
	customParams := types.DefaultParams()
	customParams.AuthorizedVerifiers = []string{"aura1authorized_verifier"}
	paramsStore := params.NewStore(customParams)

	// Create keeper with specific authority
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

	t.Run("RevokeDataItem_SuccessWithAuthority", func(t *testing.T) {
		// Store a data item
		storeMsg := &types.MsgStoreDataItem{
			Creator:         "aura1owner",
			DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:     []byte("hash_revoke1"),
			StorageLocation: "ipfs://revoke1",
			Title:           "Photo to Revoke",
		}

		storeResp, err := msgServer.StoreDataItem(input.Ctx, storeMsg)
		require.NoError(t, err)

		// Revoke it with the keeper authority
		revokeMsg := &types.MsgRevokeDataItem{
			Authority: "aura1authority",
			DataId:    storeResp.DataId,
			Reason:    "Content violates policy",
		}

		revokeResp, err := msgServer.RevokeDataItem(input.Ctx, revokeMsg)
		require.NoError(t, err)
		require.NotNil(t, revokeResp)
		require.NotNil(t, revokeResp.RevokedAt)

		// Verify the item was revoked
		item, found := k.GetDataItem(input.Ctx, storeResp.DataId)
		require.True(t, found)
		require.Equal(t, types.DataItemStatus_DATA_ITEM_STATUS_REVOKED, item.Status)
	})

	t.Run("RevokeDataItem_SuccessWithAuthorizedVerifier", func(t *testing.T) {
		// Store a data item
		storeMsg := &types.MsgStoreDataItem{
			Creator:         "aura1owner",
			DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:     []byte("hash_revoke2"),
			StorageLocation: "ipfs://revoke2",
			Title:           "Photo to Revoke 2",
		}

		storeResp, err := msgServer.StoreDataItem(input.Ctx, storeMsg)
		require.NoError(t, err)

		// Revoke it with an authorized verifier from params
		revokeMsg := &types.MsgRevokeDataItem{
			Authority: "aura1authorized_verifier",
			DataId:    storeResp.DataId,
			Reason:    "Content violates policy",
		}

		revokeResp, err := msgServer.RevokeDataItem(input.Ctx, revokeMsg)
		require.NoError(t, err)
		require.NotNil(t, revokeResp)
		require.NotNil(t, revokeResp.RevokedAt)

		// Verify the item was revoked
		item, found := k.GetDataItem(input.Ctx, storeResp.DataId)
		require.True(t, found)
		require.Equal(t, types.DataItemStatus_DATA_ITEM_STATUS_REVOKED, item.Status)
	})

	t.Run("RevokeDataItem_Unauthorized", func(t *testing.T) {
		// Store a data item
		storeMsg := &types.MsgStoreDataItem{
			Creator:         "aura1owner",
			DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:     []byte("hash_revoke3"),
			StorageLocation: "ipfs://revoke3",
			Title:           "Photo to Revoke 3",
		}

		storeResp, err := msgServer.StoreDataItem(input.Ctx, storeMsg)
		require.NoError(t, err)

		// Try to revoke with unauthorized address
		revokeMsg := &types.MsgRevokeDataItem{
			Authority: "aura1attacker",
			DataId:    storeResp.DataId,
			Reason:    "Malicious revocation",
		}

		_, err = msgServer.RevokeDataItem(input.Ctx, revokeMsg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "is not authorized to revoke data items")
	})

	t.Run("RevokeDataItem_NilMessage", func(t *testing.T) {
		_, err := msgServer.RevokeDataItem(input.Ctx, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "message cannot be nil")
	})

	t.Run("RevokeDataItem_EmptyAuthority", func(t *testing.T) {
		revokeMsg := &types.MsgRevokeDataItem{
			Authority: "",
			DataId:    "some-data-id",
			Reason:    "Some reason",
		}

		_, err := msgServer.RevokeDataItem(input.Ctx, revokeMsg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "authority cannot be empty")
	})

	t.Run("RevokeDataItem_EmptyDataId", func(t *testing.T) {
		revokeMsg := &types.MsgRevokeDataItem{
			Authority: "aura1authority",
			DataId:    "",
			Reason:    "Some reason",
		}

		_, err := msgServer.RevokeDataItem(input.Ctx, revokeMsg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "data_id cannot be empty")
	})

	t.Run("RevokeDataItem_EmptyReason", func(t *testing.T) {
		revokeMsg := &types.MsgRevokeDataItem{
			Authority: "aura1authority",
			DataId:    "some-data-id",
			Reason:    "",
		}

		_, err := msgServer.RevokeDataItem(input.Ctx, revokeMsg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "reason cannot be empty")
	})

	t.Run("RevokeDataItem_ItemNotFound", func(t *testing.T) {
		revokeMsg := &types.MsgRevokeDataItem{
			Authority: "aura1authority",
			DataId:    "nonexistent-data-id",
			Reason:    "Content violates policy",
		}

		_, err := msgServer.RevokeDataItem(input.Ctx, revokeMsg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "data item not found")
	})

	t.Run("RevokeDataItem_AlreadyRevoked", func(t *testing.T) {
		// Store a data item
		storeMsg := &types.MsgStoreDataItem{
			Creator:         "aura1owner",
			DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
			ContentHash:     []byte("hash_revoke_twice"),
			StorageLocation: "ipfs://revoke_twice",
			Title:           "Photo to Revoke Twice",
		}

		storeResp, err := msgServer.StoreDataItem(input.Ctx, storeMsg)
		require.NoError(t, err)

		// Revoke it once
		revokeMsg := &types.MsgRevokeDataItem{
			Authority: "aura1authority",
			DataId:    storeResp.DataId,
			Reason:    "First revocation",
		}

		_, err = msgServer.RevokeDataItem(input.Ctx, revokeMsg)
		require.NoError(t, err)

		// Try to revoke again
		revokeMsg2 := &types.MsgRevokeDataItem{
			Authority: "aura1authority",
			DataId:    storeResp.DataId,
			Reason:    "Second revocation",
		}

		_, err = msgServer.RevokeDataItem(input.Ctx, revokeMsg2)
		require.Error(t, err)
		require.Contains(t, err.Error(), "data item is revoked")
	})
}
