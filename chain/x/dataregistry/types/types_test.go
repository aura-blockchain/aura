package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTypeAliases(t *testing.T) {
	// Test that type aliases are properly exported
	var dataItem *DataItem
	require.Nil(t, dataItem)

	var verification *Verification
	require.Nil(t, verification)

	var params *Params
	require.Nil(t, params)
}

func TestDataItemTypeEnums(t *testing.T) {
	// Test that enum constants exist and are distinct
	require.NotEqual(t, DataItemType_DATA_ITEM_TYPE_UNSPECIFIED, DataItemType_DATA_ITEM_TYPE_VEHICLE_REGISTRATION)
	require.NotEqual(t, DataItemType_DATA_ITEM_TYPE_VEHICLE_REGISTRATION, DataItemType_DATA_ITEM_TYPE_VEHICLE_INSURANCE)
	require.NotEqual(t, DataItemType_DATA_ITEM_TYPE_PROPERTY_DEED, DataItemType_DATA_ITEM_TYPE_LEASE_AGREEMENT)
	require.NotEqual(t, DataItemType_DATA_ITEM_TYPE_CONTRACT, DataItemType_DATA_ITEM_TYPE_RECEIPT)
	require.NotEqual(t, DataItemType_DATA_ITEM_TYPE_PHOTO, DataItemType_DATA_ITEM_TYPE_VIDEO)
	require.NotEqual(t, DataItemType_DATA_ITEM_TYPE_VIDEO, DataItemType_DATA_ITEM_TYPE_AUDIO)
	require.NotEqual(t, DataItemType_DATA_ITEM_TYPE_GOLF_SCORE, DataItemType_DATA_ITEM_TYPE_TEST_SCORE)
	require.NotEqual(t, DataItemType_DATA_ITEM_TYPE_NFT, DataItemType_DATA_ITEM_TYPE_DIGITAL_ART)
	require.NotEqual(t, DataItemType_DATA_ITEM_TYPE_VACCINATION_RECORD, DataItemType_DATA_ITEM_TYPE_MEDICAL_RECORD)
}

func TestDataItemStatusEnums(t *testing.T) {
	require.NotEqual(t, DataItemStatus_DATA_ITEM_STATUS_UNSPECIFIED, DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION)
	require.NotEqual(t, DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION, DataItemStatus_DATA_ITEM_STATUS_VERIFIED)
	require.NotEqual(t, DataItemStatus_DATA_ITEM_STATUS_VERIFIED, DataItemStatus_DATA_ITEM_STATUS_REJECTED)
}

func TestVerificationLevelEnums(t *testing.T) {
	require.NotEqual(t, VerificationLevel_VERIFICATION_LEVEL_UNSPECIFIED, VerificationLevel_VERIFICATION_LEVEL_SELF_ATTESTED)
	require.NotEqual(t, VerificationLevel_VERIFICATION_LEVEL_SELF_ATTESTED, VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED)
	require.NotEqual(t, VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED, VerificationLevel_VERIFICATION_LEVEL_AUTHORITY_VERIFIED)
}

func TestAccessModeEnums(t *testing.T) {
	require.NotEqual(t, AccessMode_ACCESS_MODE_PRIVATE, AccessMode_ACCESS_MODE_WHITELIST)
	require.NotEqual(t, AccessMode_ACCESS_MODE_WHITELIST, AccessMode_ACCESS_MODE_PUBLIC)
	require.NotEqual(t, AccessMode_ACCESS_MODE_PUBLIC, AccessMode_ACCESS_MODE_VERIFIED_USERS)
}

func TestMsgTypeAliases(t *testing.T) {
	// Test message type aliases
	var msgStore *MsgStoreDataItem
	require.Nil(t, msgStore)

	var msgUpdate *MsgUpdateDataItem
	require.Nil(t, msgUpdate)

	var msgDelete *MsgDeleteDataItem
	require.Nil(t, msgDelete)

	var msgVerify *MsgVerifyDataItem
	require.Nil(t, msgVerify)

	var msgRevoke *MsgRevokeDataItem
	require.Nil(t, msgRevoke)
}

func TestQueryTypeAliases(t *testing.T) {
	// Test query type aliases
	var queryItem *QueryDataItemRequest
	require.Nil(t, queryItem)

	var queryUser *QueryUserDataItemsRequest
	require.Nil(t, queryUser)

	var querySearch *QuerySearchDataItemsRequest
	require.Nil(t, querySearch)

	var queryVerifications *QueryDataItemVerificationsRequest
	require.Nil(t, queryVerifications)

	var queryStats *QueryStatsRequest
	require.Nil(t, queryStats)

	var queryParams *QueryParamsRequest
	require.Nil(t, queryParams)
}

func TestEventTypeAliases(t *testing.T) {
	// Test event type aliases
	var eventStored *EventDataItemStored
	require.Nil(t, eventStored)

	var eventUpdated *EventDataItemUpdated
	require.Nil(t, eventUpdated)

	var eventDeleted *EventDataItemDeleted
	require.Nil(t, eventDeleted)

	var eventVerified *EventDataItemVerified
	require.Nil(t, eventVerified)

	var eventRevoked *EventDataItemRevoked
	require.Nil(t, eventRevoked)
}

func TestGenesisTypeAlias(t *testing.T) {
	var genesis *GenesisState
	require.Nil(t, genesis)
}

func TestSpecializedDataStructures(t *testing.T) {
	// Test VehicleRegistrationData
	var vehicle *VehicleRegistrationData
	require.Nil(t, vehicle)

	// Test PhotoData
	var photo *PhotoData
	require.Nil(t, photo)

	// Test GolfScoreData
	var golf *GolfScoreData
	require.Nil(t, golf)
}

func TestDataItemStruct(t *testing.T) {
	// Create a data item with correct field names
	item := &DataItem{
		DataId:          "test-item-1",
		OwnerAddress:    "aura1owner",
		DataType:        DataItemType_DATA_ITEM_TYPE_PHOTO,
		ContentHash:     []byte("hash123"),
		StorageLocation: "ipfs://Qm...",
		Status:          DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION,
		AccessPolicy: &AccessPolicy{
			Mode: AccessMode_ACCESS_MODE_PRIVATE,
		},
	}

	// Verify field values
	require.Equal(t, "test-item-1", item.DataId)
	require.Equal(t, "aura1owner", item.OwnerAddress)
	require.Equal(t, DataItemType_DATA_ITEM_TYPE_PHOTO, item.DataType)
	require.Equal(t, DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION, item.Status)
	require.NotNil(t, item.AccessPolicy)
}

func TestVerificationStruct(t *testing.T) {
	verification := &Verification{
		VerifierAddress: "verifier-1",
		Level:           VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
		ConfidenceScore: 95,
	}

	require.NotEmpty(t, verification.VerifierAddress)
	require.Equal(t, VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED, verification.Level)
	require.Greater(t, verification.ConfidenceScore, uint64(0))
}
