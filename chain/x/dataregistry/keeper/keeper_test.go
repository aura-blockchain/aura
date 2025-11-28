package keeper

import (
	"testing"

	"github.com/aequitas/aura/chain/x/dataregistry/params"
	"github.com/aequitas/aura/chain/x/dataregistry/types"
)

func TestNewKeeper(t *testing.T) {
	paramsStore := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(paramsStore)

	if keeper == nil {
		t.Fatal("keeper should not be nil")
	}

	// Test params
	p := keeper.GetParams()
	if p.MaxDataItemsPerUser == 0 {
		t.Error("MaxDataItemsPerUser should not be 0")
	}
}

func TestStoreDataItem(t *testing.T) {
	paramsStore := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(paramsStore)

	// Store a data item
	dataID, err := keeper.StoreDataItem(
		"aura1test",
		types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Test Photo",
		"A test photo",
		[]byte("hash123"),
		"ipfs://QmTest",
		false,
		nil,
		map[string]string{"key": "value"},
		&types.AccessPolicy{Mode: types.AccessMode_ACCESS_MODE_PUBLIC},
		[]string{"test", "photo"},
	)

	if err != nil {
		t.Fatalf("StoreDataItem failed: %v", err)
	}

	if dataID == "" {
		t.Fatal("dataID should not be empty")
	}

	// Retrieve the item
	item, ok := keeper.GetDataItem(dataID)
	if !ok {
		t.Fatal("data item should exist")
	}

	if item.Title != "Test Photo" {
		t.Errorf("expected title 'Test Photo', got '%s'", item.Title)
	}

	if item.OwnerAddress != "aura1test" {
		t.Errorf("expected owner 'aura1test', got '%s'", item.OwnerAddress)
	}
}

func TestUpdateDataItem(t *testing.T) {
	paramsStore := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(paramsStore)

	// Store a data item
	dataID, err := keeper.StoreDataItem(
		"aura1test",
		types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Original Title",
		"Original Description",
		[]byte("hash123"),
		"ipfs://QmTest",
		false,
		nil,
		nil,
		&types.AccessPolicy{Mode: types.AccessMode_ACCESS_MODE_PRIVATE},
		nil,
	)

	if err != nil {
		t.Fatalf("StoreDataItem failed: %v", err)
	}

	// Update the item
	err = keeper.UpdateDataItem(
		dataID,
		"aura1test",
		"Updated Title",
		"Updated Description",
		map[string]string{"updated": "true"},
		&types.AccessPolicy{Mode: types.AccessMode_ACCESS_MODE_PUBLIC},
		[]string{"updated"},
	)

	if err != nil {
		t.Fatalf("UpdateDataItem failed: %v", err)
	}

	// Verify the update
	item, ok := keeper.GetDataItem(dataID)
	if !ok {
		t.Fatal("data item should exist")
	}

	if item.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", item.Title)
	}

	if item.AccessPolicy.Mode != types.AccessMode_ACCESS_MODE_PUBLIC {
		t.Error("access mode should be public")
	}
}

func TestVerifyDataItem(t *testing.T) {
	paramsStore := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(paramsStore)

	// Store a data item
	dataID, err := keeper.StoreDataItem(
		"aura1test",
		types.DataItemType_DATA_ITEM_TYPE_GOLF_SCORE,
		"Golf Score",
		"Pebble Beach Round",
		[]byte("hash123"),
		"ipfs://QmTest",
		false,
		nil,
		nil,
		&types.AccessPolicy{Mode: types.AccessMode_ACCESS_MODE_PUBLIC},
		nil,
	)

	if err != nil {
		t.Fatalf("StoreDataItem failed: %v", err)
	}

	// Verify the item
	err = keeper.VerifyDataItem(
		dataID,
		"aura1verifier",
		types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
		85,
		"Verified by playing partner",
		"Manual review",
		nil,
	)

	if err != nil {
		t.Fatalf("VerifyDataItem failed: %v", err)
	}

	// Check verification
	item, ok := keeper.GetDataItem(dataID)
	if !ok {
		t.Fatal("data item should exist")
	}

	if item.Status != types.DataItemStatus_DATA_ITEM_STATUS_VERIFIED {
		t.Error("status should be verified")
	}

	if len(item.Verifications) != 1 {
		t.Errorf("expected 1 verification, got %d", len(item.Verifications))
	}
}

func TestCheckAccess(t *testing.T) {
	paramsStore := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(paramsStore)

	// Store a private data item
	dataID, err := keeper.StoreDataItem(
		"aura1owner",
		types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Private Photo",
		"A private photo",
		[]byte("hash123"),
		"ipfs://QmTest",
		false,
		nil,
		nil,
		&types.AccessPolicy{Mode: types.AccessMode_ACCESS_MODE_PRIVATE},
		nil,
	)

	if err != nil {
		t.Fatalf("StoreDataItem failed: %v", err)
	}

	// Owner should have access
	if !keeper.CheckAccess(dataID, "aura1owner") {
		t.Error("owner should have access")
	}

	// Other user should not have access
	if keeper.CheckAccess(dataID, "aura1other") {
		t.Error("other user should not have access to private item")
	}

	// Store a public data item
	publicDataID, err := keeper.StoreDataItem(
		"aura1owner",
		types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Public Photo",
		"A public photo",
		[]byte("hash456"),
		"ipfs://QmTest2",
		false,
		nil,
		nil,
		&types.AccessPolicy{Mode: types.AccessMode_ACCESS_MODE_PUBLIC},
		nil,
	)

	if err != nil {
		t.Fatalf("StoreDataItem failed: %v", err)
	}

	// Anyone should have access to public item
	if !keeper.CheckAccess(publicDataID, "aura1anyone") {
		t.Error("anyone should have access to public item")
	}
}
