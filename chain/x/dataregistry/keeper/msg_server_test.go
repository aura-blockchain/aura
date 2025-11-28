package keeper

import (
	"context"
	"testing"

	"github.com/aequitas/aura/chain/x/dataregistry/params"
	"github.com/aequitas/aura/chain/x/dataregistry/types"
	pb "github.com/aequitas/aura/proto/aura/dataregistry/v1beta1"
)

func TestMsgServer_StoreDataItem(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(store)
	msgServer := NewMsgServer(keeper)
	ctx := context.Background()

	tests := []struct {
		name    string
		msg     *pb.MsgStoreDataItem
		wantErr bool
	}{
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
		},
		{
			name: "empty creator",
			msg: &pb.MsgStoreDataItem{
				Creator:         "",
				DataType:        pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
				Title:           "Test",
				ContentHash:     []byte("hash"),
				StorageLocation: "QmTest",
			},
			wantErr: true,
		},
		{
			name: "unspecified data type",
			msg: &pb.MsgStoreDataItem{
				Creator:         "aura1test",
				DataType:        pb.DataItemType_DATA_ITEM_TYPE_UNSPECIFIED,
				Title:           "Test",
				ContentHash:     []byte("hash"),
				StorageLocation: "QmTest",
			},
			wantErr: true,
		},
		{
			name: "empty title",
			msg: &pb.MsgStoreDataItem{
				Creator:         "aura1test",
				DataType:        pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
				Title:           "",
				ContentHash:     []byte("hash"),
				StorageLocation: "QmTest",
			},
			wantErr: true,
		},
		{
			name: "empty content hash",
			msg: &pb.MsgStoreDataItem{
				Creator:         "aura1test",
				DataType:        pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
				Title:           "Test",
				ContentHash:     []byte{},
				StorageLocation: "QmTest",
			},
			wantErr: true,
		},
		{
			name: "empty storage location",
			msg: &pb.MsgStoreDataItem{
				Creator:         "aura1test",
				DataType:        pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
				Title:           "Test",
				ContentHash:     []byte("hash"),
				StorageLocation: "",
			},
			wantErr: true,
		},
		{
			name: "valid message",
			msg: &pb.MsgStoreDataItem{
				Creator:         "aura1test",
				DataType:        pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
				Title:           "Test Photo",
				Description:     "A test photo",
				ContentHash:     []byte("hash123"),
				StorageLocation: "QmTest123",
				IsEncrypted:     false,
				AccessPolicy:    &pb.AccessPolicy{Mode: pb.AccessMode_ACCESS_MODE_PUBLIC},
				Tags:            []string{"test"},
			},
			wantErr: false,
		},
		{
			name: "valid message with metadata",
			msg: &pb.MsgStoreDataItem{
				Creator:         "aura1test",
				DataType:        pb.DataItemType_DATA_ITEM_TYPE_VIDEO,
				Title:           "Test Video",
				Description:     "A test video",
				ContentHash:     []byte("videohash"),
				StorageLocation: "QmVideo",
				Metadata:        map[string]string{"duration": "60s", "format": "mp4"},
				AccessPolicy:    &pb.AccessPolicy{Mode: pb.AccessMode_ACCESS_MODE_PUBLIC},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.StoreDataItem(ctx, tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("StoreDataItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if resp.DataId == "" {
				t.Error("StoreDataItem() returned empty data_id")
			}
			if resp.CreatedAt == nil {
				t.Error("StoreDataItem() returned nil created_at")
			}

			// Verify the item was stored
			item, exists := keeper.GetDataItem(resp.DataId)
			if !exists {
				t.Error("Stored data item not found in keeper")
			}
			if item.Title != tt.msg.Title {
				t.Errorf("Stored item title = %v, want %v", item.Title, tt.msg.Title)
			}
		})
	}
}

func TestMsgServer_UpdateDataItem(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(store)
	msgServer := NewMsgServer(keeper)
	ctx := context.Background()

	// Create a test data item first
	ownerAddr := "aura1test"
	dataID, err := keeper.StoreDataItem(
		ownerAddr,
		pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Original Title",
		"Original Description",
		[]byte("hash"),
		"QmTest",
		false,
		nil,
		nil,
		&pb.AccessPolicy{Mode: pb.AccessMode_ACCESS_MODE_PUBLIC},
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to store data item: %v", err)
	}

	tests := []struct {
		name    string
		msg     *pb.MsgUpdateDataItem
		wantErr bool
	}{
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
		},
		{
			name: "empty creator",
			msg: &pb.MsgUpdateDataItem{
				Creator: "",
				DataId:  dataID,
				Title:   "Updated",
			},
			wantErr: true,
		},
		{
			name: "empty data_id",
			msg: &pb.MsgUpdateDataItem{
				Creator: ownerAddr,
				DataId:  "",
				Title:   "Updated",
			},
			wantErr: true,
		},
		{
			name: "no fields to update",
			msg: &pb.MsgUpdateDataItem{
				Creator: ownerAddr,
				DataId:  dataID,
			},
			wantErr: true,
		},
		{
			name: "non-existent data item",
			msg: &pb.MsgUpdateDataItem{
				Creator: ownerAddr,
				DataId:  "data:nonexistent",
				Title:   "Updated",
			},
			wantErr: true,
		},
		{
			name: "unauthorized update",
			msg: &pb.MsgUpdateDataItem{
				Creator: "aura1other",
				DataId:  dataID,
				Title:   "Updated",
			},
			wantErr: true,
		},
		{
			name: "valid update - title",
			msg: &pb.MsgUpdateDataItem{
				Creator: ownerAddr,
				DataId:  dataID,
				Title:   "Updated Title",
			},
			wantErr: false,
		},
		{
			name: "valid update - multiple fields",
			msg: &pb.MsgUpdateDataItem{
				Creator:     ownerAddr,
				DataId:      dataID,
				Title:       "New Title",
				Description: "New Description",
				Tags:        []string{"tag1", "tag2"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.UpdateDataItem(ctx, tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateDataItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if resp.UpdatedAt == nil {
				t.Error("UpdateDataItem() returned nil updated_at")
			}

			// Verify the update
			if tt.msg.Title != "" {
				item, _ := keeper.GetDataItem(dataID)
				if item.Title != tt.msg.Title {
					t.Errorf("Updated item title = %v, want %v", item.Title, tt.msg.Title)
				}
			}
		})
	}
}

func TestMsgServer_DeleteDataItem(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(store)
	msgServer := NewMsgServer(keeper)
	ctx := context.Background()

	// Create a test data item
	ownerAddr := "aura1test"
	dataID, err := keeper.StoreDataItem(
		ownerAddr,
		pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Test Photo",
		"A test photo",
		[]byte("hash"),
		"QmTest",
		false,
		nil,
		nil,
		&pb.AccessPolicy{Mode: pb.AccessMode_ACCESS_MODE_PUBLIC},
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to store data item: %v", err)
	}

	tests := []struct {
		name    string
		msg     *pb.MsgDeleteDataItem
		wantErr bool
	}{
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
		},
		{
			name: "empty creator",
			msg: &pb.MsgDeleteDataItem{
				Creator: "",
				DataId:  dataID,
			},
			wantErr: true,
		},
		{
			name: "empty data_id",
			msg: &pb.MsgDeleteDataItem{
				Creator: ownerAddr,
				DataId:  "",
			},
			wantErr: true,
		},
		{
			name: "non-existent data item",
			msg: &pb.MsgDeleteDataItem{
				Creator: ownerAddr,
				DataId:  "data:nonexistent",
			},
			wantErr: true,
		},
		{
			name: "unauthorized delete",
			msg: &pb.MsgDeleteDataItem{
				Creator: "aura1other",
				DataId:  dataID,
			},
			wantErr: true,
		},
		{
			name: "valid delete",
			msg: &pb.MsgDeleteDataItem{
				Creator: ownerAddr,
				DataId:  dataID,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.DeleteDataItem(ctx, tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteDataItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if resp.DeletedAt == nil {
				t.Error("DeleteDataItem() returned nil deleted_at")
			}

			// Verify deletion
			_, exists := keeper.GetDataItem(tt.msg.DataId)
			if exists {
				t.Error("Data item should be deleted")
			}
		})
	}
}

func TestMsgServer_VerifyDataItem(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(store)
	msgServer := NewMsgServer(keeper)
	ctx := context.Background()

	// Create a test data item
	ownerAddr := "aura1test"
	dataID, err := keeper.StoreDataItem(
		ownerAddr,
		pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Test Photo",
		"A test photo",
		[]byte("hash"),
		"QmTest",
		false,
		nil,
		nil,
		&pb.AccessPolicy{Mode: pb.AccessMode_ACCESS_MODE_PUBLIC},
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to store data item: %v", err)
	}

	tests := []struct {
		name    string
		msg     *pb.MsgVerifyDataItem
		wantErr bool
	}{
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
		},
		{
			name: "empty verifier",
			msg: &pb.MsgVerifyDataItem{
				Verifier:           "",
				DataId:             dataID,
				Level:              pb.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
				ConfidenceScore:    95,
				VerificationMethod: "manual",
			},
			wantErr: true,
		},
		{
			name: "empty data_id",
			msg: &pb.MsgVerifyDataItem{
				Verifier:           "aura1verifier",
				DataId:             "",
				Level:              pb.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
				ConfidenceScore:    95,
				VerificationMethod: "manual",
			},
			wantErr: true,
		},
		{
			name: "unspecified verification level",
			msg: &pb.MsgVerifyDataItem{
				Verifier:           "aura1verifier",
				DataId:             dataID,
				Level:              pb.VerificationLevel_VERIFICATION_LEVEL_UNSPECIFIED,
				ConfidenceScore:    95,
				VerificationMethod: "manual",
			},
			wantErr: true,
		},
		{
			name: "invalid confidence score",
			msg: &pb.MsgVerifyDataItem{
				Verifier:           "aura1verifier",
				DataId:             dataID,
				Level:              pb.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
				ConfidenceScore:    150,
				VerificationMethod: "manual",
			},
			wantErr: true,
		},
		{
			name: "empty verification method",
			msg: &pb.MsgVerifyDataItem{
				Verifier:           "aura1verifier",
				DataId:             dataID,
				Level:              pb.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
				ConfidenceScore:    95,
				VerificationMethod: "",
			},
			wantErr: true,
		},
		{
			name: "non-existent data item",
			msg: &pb.MsgVerifyDataItem{
				Verifier:           "aura1verifier",
				DataId:             "data:nonexistent",
				Level:              pb.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
				ConfidenceScore:    95,
				VerificationMethod: "manual",
			},
			wantErr: true,
		},
		{
			name: "valid verification",
			msg: &pb.MsgVerifyDataItem{
				Verifier:           "aura1verifier",
				DataId:             dataID,
				Level:              pb.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
				ConfidenceScore:    95,
				Notes:              "Looks good",
				VerificationMethod: "manual",
			},
			wantErr: false,
		},
		{
			name: "valid AI verification",
			msg: &pb.MsgVerifyDataItem{
				Verifier:           "aura1ai",
				DataId:             dataID,
				Level:              pb.VerificationLevel_VERIFICATION_LEVEL_AI_VERIFIED,
				ConfidenceScore:    87,
				Notes:              "AI analysis complete",
				VerificationMethod: "ai_ocr",
				Proof:              []byte("proof_data"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.VerifyDataItem(ctx, tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyDataItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if resp.VerifiedAt == nil {
				t.Error("VerifyDataItem() returned nil verified_at")
			}
			if resp.VerificationReward == 0 {
				t.Error("VerifyDataItem() returned zero verification_reward")
			}

			// Verify the verification was added
			item, _ := keeper.GetDataItem(dataID)
			if len(item.Verifications) == 0 {
				t.Error("Verification was not added to data item")
			}
		})
	}
}

func TestMsgServer_RevokeDataItem(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(store)
	msgServer := NewMsgServer(keeper)
	ctx := context.Background()

	// Create a test data item
	ownerAddr := "aura1test"
	dataID, err := keeper.StoreDataItem(
		ownerAddr,
		pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Test Photo",
		"A test photo",
		[]byte("hash"),
		"QmTest",
		false,
		nil,
		nil,
		&pb.AccessPolicy{Mode: pb.AccessMode_ACCESS_MODE_PUBLIC},
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to store data item: %v", err)
	}

	tests := []struct {
		name    string
		msg     *pb.MsgRevokeDataItem
		wantErr bool
	}{
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
		},
		{
			name: "empty authority",
			msg: &pb.MsgRevokeDataItem{
				Authority: "",
				DataId:    dataID,
				Reason:    "Violation",
			},
			wantErr: true,
		},
		{
			name: "empty data_id",
			msg: &pb.MsgRevokeDataItem{
				Authority: "aura1authority",
				DataId:    "",
				Reason:    "Violation",
			},
			wantErr: true,
		},
		{
			name: "empty reason",
			msg: &pb.MsgRevokeDataItem{
				Authority: "aura1authority",
				DataId:    dataID,
				Reason:    "",
			},
			wantErr: true,
		},
		{
			name: "non-existent data item",
			msg: &pb.MsgRevokeDataItem{
				Authority: "aura1authority",
				DataId:    "data:nonexistent",
				Reason:    "Violation",
			},
			wantErr: true,
		},
		{
			name: "valid revocation",
			msg: &pb.MsgRevokeDataItem{
				Authority: "aura1authority",
				DataId:    dataID,
				Reason:    "Content policy violation",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.RevokeDataItem(ctx, tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("RevokeDataItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if resp.RevokedAt == nil {
				t.Error("RevokeDataItem() returned nil revoked_at")
			}

			// Verify the item was revoked
			item, _ := keeper.GetDataItem(dataID)
			if item.Status != pb.DataItemStatus_DATA_ITEM_STATUS_REVOKED {
				t.Error("Data item was not revoked")
			}
		})
	}
}

func TestMsgServer_AccessPolicyValidation(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(store)
	msgServer := NewMsgServer(keeper)
	ctx := context.Background()

	tests := []struct {
		name    string
		msg     *pb.MsgStoreDataItem
		wantErr bool
	}{
		{
			name: "whitelist without addresses",
			msg: &pb.MsgStoreDataItem{
				Creator:         "aura1test",
				DataType:        pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
				Title:           "Test",
				ContentHash:     []byte("hash"),
				StorageLocation: "QmTest",
				AccessPolicy: &pb.AccessPolicy{
					Mode:             pb.AccessMode_ACCESS_MODE_WHITELIST,
					AllowedAddresses: []string{},
				},
			},
			wantErr: true,
		},
		{
			name: "private with allowed addresses",
			msg: &pb.MsgStoreDataItem{
				Creator:         "aura1test",
				DataType:        pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
				Title:           "Test",
				ContentHash:     []byte("hash"),
				StorageLocation: "QmTest",
				AccessPolicy: &pb.AccessPolicy{
					Mode:             pb.AccessMode_ACCESS_MODE_PRIVATE,
					AllowedAddresses: []string{"aura1other"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid whitelist",
			msg: &pb.MsgStoreDataItem{
				Creator:         "aura1test",
				DataType:        pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
				Title:           "Test",
				ContentHash:     []byte("hash"),
				StorageLocation: "QmTest",
				AccessPolicy: &pb.AccessPolicy{
					Mode:             pb.AccessMode_ACCESS_MODE_WHITELIST,
					AllowedAddresses: []string{"aura1other", "aura1friend"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid private",
			msg: &pb.MsgStoreDataItem{
				Creator:         "aura1test",
				DataType:        pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
				Title:           "Test",
				ContentHash:     []byte("hash"),
				StorageLocation: "QmTest",
				AccessPolicy: &pb.AccessPolicy{
					Mode: pb.AccessMode_ACCESS_MODE_PRIVATE,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := msgServer.StoreDataItem(ctx, tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("StoreDataItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMsgServer_MaxDataItemsLimit(t *testing.T) {
	// Create params with low limit for testing
	customParams := types.DefaultParams()
	customParams.MaxDataItemsPerUser = 2

	store := params.NewStore(customParams)
	keeper := NewKeeper(store)
	msgServer := NewMsgServer(keeper)
	ctx := context.Background()

	ownerAddr := "aura1test"

	// Store first item
	_, err := msgServer.StoreDataItem(ctx, &pb.MsgStoreDataItem{
		Creator:         ownerAddr,
		DataType:        pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
		Title:           "Photo 1",
		ContentHash:     []byte("hash1"),
		StorageLocation: "QmTest1",
	})
	if err != nil {
		t.Fatalf("Failed to store first item: %v", err)
	}

	// Store second item
	_, err = msgServer.StoreDataItem(ctx, &pb.MsgStoreDataItem{
		Creator:         ownerAddr,
		DataType:        pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
		Title:           "Photo 2",
		ContentHash:     []byte("hash2"),
		StorageLocation: "QmTest2",
	})
	if err != nil {
		t.Fatalf("Failed to store second item: %v", err)
	}

	// Try to store third item (should fail)
	_, err = msgServer.StoreDataItem(ctx, &pb.MsgStoreDataItem{
		Creator:         ownerAddr,
		DataType:        pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
		Title:           "Photo 3",
		ContentHash:     []byte("hash3"),
		StorageLocation: "QmTest3",
	})
	if err == nil {
		t.Error("Expected error when exceeding max data items limit")
	}
}

func TestMsgServer_VerificationLevelUpgrade(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(store)
	msgServer := NewMsgServer(keeper)
	ctx := context.Background()

	// Create a test data item
	ownerAddr := "aura1test"
	dataID, err := keeper.StoreDataItem(
		ownerAddr,
		pb.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Test Photo",
		"A test photo",
		[]byte("hash"),
		"QmTest",
		false,
		nil,
		nil,
		&pb.AccessPolicy{Mode: pb.AccessMode_ACCESS_MODE_PUBLIC},
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to store data item: %v", err)
	}

	// Initial verification level should be SELF_ATTESTED
	item, _ := keeper.GetDataItem(dataID)
	if item.VerificationLevel != pb.VerificationLevel_VERIFICATION_LEVEL_SELF_ATTESTED {
		t.Errorf("Initial verification level = %v, want %v", item.VerificationLevel, pb.VerificationLevel_VERIFICATION_LEVEL_SELF_ATTESTED)
	}

	// Add peer verification
	_, err = msgServer.VerifyDataItem(ctx, &pb.MsgVerifyDataItem{
		Verifier:           "aura1verifier",
		DataId:             dataID,
		Level:              pb.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
		ConfidenceScore:    90,
		VerificationMethod: "manual",
	})
	if err != nil {
		t.Fatalf("Failed to verify: %v", err)
	}

	// Check verification level upgraded
	item, _ = keeper.GetDataItem(dataID)
	if item.VerificationLevel != pb.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED {
		t.Errorf("Verification level = %v, want %v", item.VerificationLevel, pb.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED)
	}

	// Add authority verification
	_, err = msgServer.VerifyDataItem(ctx, &pb.MsgVerifyDataItem{
		Verifier:           "aura1authority",
		DataId:             dataID,
		Level:              pb.VerificationLevel_VERIFICATION_LEVEL_AUTHORITY_VERIFIED,
		ConfidenceScore:    100,
		VerificationMethod: "official",
	})
	if err != nil {
		t.Fatalf("Failed to verify: %v", err)
	}

	// Check verification level upgraded again
	item, _ = keeper.GetDataItem(dataID)
	if item.VerificationLevel != pb.VerificationLevel_VERIFICATION_LEVEL_AUTHORITY_VERIFIED {
		t.Errorf("Verification level = %v, want %v", item.VerificationLevel, pb.VerificationLevel_VERIFICATION_LEVEL_AUTHORITY_VERIFIED)
	}
}
