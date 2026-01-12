// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/dataregistry/keeper"
	"github.com/aequitas/aura/chain/x/dataregistry/params"
	"github.com/aequitas/aura/chain/x/dataregistry/types"
	"github.com/stretchr/testify/suite"
)

type DataItemTestSuite struct {
	suite.Suite
	input  keepertest.TestInput
	keeper *keeper.Keeper
}

func (s *DataItemTestSuite) SetupTest() {
	s.input = keepertest.CreateTestInput(s.T())
	paramsStore := params.NewStore(types.DefaultParams())
	s.keeper = keeper.NewKeeper(
		keepertest.WrapStoreService(s.input.StoreKey),
		s.input.Cdc,
		paramsStore,
		"aura1authority",
		keepertest.Logger(),
	)
}

func TestDataItemTestSuite(t *testing.T) {
	suite.Run(t, new(DataItemTestSuite))
}

func (s *DataItemTestSuite) TestStoreDataItem_Success() {
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx,
		"aura1owner123",
		types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Test Photo",
		"A test photo description",
		[]byte("contenthash123"),
		"ipfs://Qm123456",
		false,
		nil,
		map[string]string{"key": "value"},
		&types.AccessPolicy{Mode: types.AccessMode_ACCESS_MODE_PRIVATE},
		[]string{"test", "photo"},
	)

	s.Require().NoError(err)
	s.Require().NotEmpty(dataID)

	// Verify stored item
	item, found := s.keeper.GetDataItem(s.input.Ctx, dataID)
	s.Require().True(found)
	s.Require().Equal("aura1owner123", item.OwnerAddress)
	s.Require().Equal(types.DataItemType_DATA_ITEM_TYPE_PHOTO, item.DataType)
	s.Require().Equal("Test Photo", item.Title)
	s.Require().Equal(types.DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION, item.Status)
}

func (s *DataItemTestSuite) TestStoreDataItem_InvalidOwner() {
	_, err := s.keeper.StoreDataItem(
		s.input.Ctx,
		"", // Empty owner
		types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Test",
		"Desc",
		[]byte("hash"),
		"ipfs://test",
		false, nil, nil, nil, nil,
	)

	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrInvalidOwner)
}

func (s *DataItemTestSuite) TestStoreDataItem_InvalidContentHash() {
	_, err := s.keeper.StoreDataItem(
		s.input.Ctx,
		"aura1owner",
		types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Test",
		"Desc",
		[]byte{}, // Empty hash
		"ipfs://test",
		false, nil, nil, nil, nil,
	)

	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrInvalidContentHash)
}

func (s *DataItemTestSuite) TestStoreDataItem_InvalidStorageLocation() {
	_, err := s.keeper.StoreDataItem(
		s.input.Ctx,
		"aura1owner",
		types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Test",
		"Desc",
		[]byte("hash"),
		"", // Empty storage location
		false, nil, nil, nil, nil,
	)

	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrInvalidStorageLocation)
}

func (s *DataItemTestSuite) TestUpdateDataItem_Success() {
	// First create a data item
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx,
		"aura1owner",
		types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Original Title",
		"Original Description",
		[]byte("hash123"),
		"ipfs://test123",
		false, nil, nil, nil, nil,
	)
	s.Require().NoError(err)

	// Update it
	err = s.keeper.UpdateDataItem(
		s.input.Ctx,
		dataID,
		"aura1owner",
		"Updated Title",
		"Updated Description",
		map[string]string{"newkey": "newvalue"},
		&types.AccessPolicy{Mode: types.AccessMode_ACCESS_MODE_PUBLIC},
		[]string{"updated"},
	)
	s.Require().NoError(err)

	// Verify updates
	item, found := s.keeper.GetDataItem(s.input.Ctx, dataID)
	s.Require().True(found)
	s.Require().Equal("Updated Title", item.Title)
	s.Require().Equal("Updated Description", item.Description)
	s.Require().Equal(types.AccessMode_ACCESS_MODE_PUBLIC, item.AccessPolicy.Mode)
}

func (s *DataItemTestSuite) TestUpdateDataItem_NotFound() {
	err := s.keeper.UpdateDataItem(
		s.input.Ctx,
		"nonexistent-id",
		"aura1owner",
		"Title",
		"Desc",
		nil, nil, nil,
	)

	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrDataItemNotFound)
}

func (s *DataItemTestSuite) TestUpdateDataItem_Unauthorized() {
	// Create item with one owner
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx,
		"aura1owner1",
		types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Title",
		"Desc",
		[]byte("hash"),
		"ipfs://test",
		false, nil, nil, nil, nil,
	)
	s.Require().NoError(err)

	// Try to update as different user
	err = s.keeper.UpdateDataItem(
		s.input.Ctx,
		dataID,
		"aura1owner2", // Different owner
		"New Title",
		"",
		nil, nil, nil,
	)

	s.Require().Error(err)
}

func (s *DataItemTestSuite) TestUpdateDataItem_Revoked() {
	// Create and revoke item
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx,
		"aura1owner",
		types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Title",
		"Desc",
		[]byte("hash"),
		"ipfs://test",
		false, nil, nil, nil, nil,
	)
	s.Require().NoError(err)

	err = s.keeper.RevokeDataItem(s.input.Ctx, dataID, "aura1authority", "test")
	s.Require().NoError(err)

	// Try to update revoked item
	err = s.keeper.UpdateDataItem(
		s.input.Ctx,
		dataID,
		"aura1owner",
		"New Title",
		"",
		nil, nil, nil,
	)

	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrDataItemRevoked)
}

func (s *DataItemTestSuite) TestVerifyDataItem_Success() {
	// Create item
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx,
		"aura1owner",
		types.DataItemType_DATA_ITEM_TYPE_CERTIFICATION,
		"Credential",
		"Test credential",
		[]byte("hash"),
		"ipfs://test",
		false, nil, nil, nil, nil,
	)
	s.Require().NoError(err)

	// Verify it
	err = s.keeper.VerifyDataItem(
		s.input.Ctx,
		dataID,
		"aura1verifier",
		types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
		85,
		"Verified successfully",
		"manual_review",
		[]byte("proof"),
	)
	s.Require().NoError(err)

	// Check verification was added
	item, found := s.keeper.GetDataItem(s.input.Ctx, dataID)
	s.Require().True(found)
	s.Require().Len(item.Verifications, 1)
	s.Require().Equal("aura1verifier", item.Verifications[0].VerifierAddress)
	s.Require().Equal(types.DataItemStatus_DATA_ITEM_STATUS_VERIFIED, item.Status)
}

func (s *DataItemTestSuite) TestVerifyDataItem_NotFound() {
	err := s.keeper.VerifyDataItem(
		s.input.Ctx,
		"nonexistent",
		"aura1verifier",
		types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
		80,
		"",
		"",
		nil,
	)

	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrDataItemNotFound)
}

func (s *DataItemTestSuite) TestVerifyDataItem_Revoked() {
	// Create and revoke
	dataID, _ := s.keeper.StoreDataItem(
		s.input.Ctx,
		"aura1owner",
		types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Title",
		"Desc",
		[]byte("hash"),
		"ipfs://test",
		false, nil, nil, nil, nil,
	)
	s.keeper.RevokeDataItem(s.input.Ctx, dataID, "aura1authority", "test")

	// Try to verify
	err := s.keeper.VerifyDataItem(
		s.input.Ctx,
		dataID,
		"aura1verifier",
		types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
		80,
		"",
		"",
		nil,
	)

	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrDataItemRevoked)
}

func (s *DataItemTestSuite) TestRevokeDataItem_Success() {
	dataID, _ := s.keeper.StoreDataItem(
		s.input.Ctx,
		"aura1owner",
		types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Title",
		"Desc",
		[]byte("hash"),
		"ipfs://test",
		false, nil, nil, nil, nil,
	)

	err := s.keeper.RevokeDataItem(s.input.Ctx, dataID, "aura1authority", "policy violation")
	s.Require().NoError(err)

	item, found := s.keeper.GetDataItem(s.input.Ctx, dataID)
	s.Require().True(found)
	s.Require().Equal(types.DataItemStatus_DATA_ITEM_STATUS_REVOKED, item.Status)
}

func (s *DataItemTestSuite) TestRevokeDataItem_NotFound() {
	err := s.keeper.RevokeDataItem(s.input.Ctx, "nonexistent", "aura1authority", "test")
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrDataItemNotFound)
}

func (s *DataItemTestSuite) TestRevokeDataItem_AlreadyRevoked() {
	dataID, _ := s.keeper.StoreDataItem(
		s.input.Ctx,
		"aura1owner",
		types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Title",
		"Desc",
		[]byte("hash"),
		"ipfs://test",
		false, nil, nil, nil, nil,
	)

	// Revoke once
	s.keeper.RevokeDataItem(s.input.Ctx, dataID, "aura1authority", "first")

	// Try to revoke again
	err := s.keeper.RevokeDataItem(s.input.Ctx, dataID, "aura1authority", "second")
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrDataItemRevoked)
}

func (s *DataItemTestSuite) TestGetDataItemVerifications_Success() {
	// Create and verify item
	dataID, _ := s.keeper.StoreDataItem(
		s.input.Ctx,
		"aura1owner",
		types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Title",
		"Desc",
		[]byte("hash"),
		"ipfs://test",
		false, nil, nil, nil, nil,
	)

	s.keeper.VerifyDataItem(s.input.Ctx, dataID, "aura1v1", types.VerificationLevel_VERIFICATION_LEVEL_SELF_ATTESTED, 50, "note1", "method1", nil)
	s.keeper.VerifyDataItem(s.input.Ctx, dataID, "aura1v2", types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED, 80, "note2", "method2", nil)

	verifications, err := s.keeper.GetDataItemVerifications(s.input.Ctx, dataID)
	s.Require().NoError(err)
	s.Require().Len(verifications, 2)
}

func (s *DataItemTestSuite) TestGetDataItemVerifications_NotFound() {
	_, err := s.keeper.GetDataItemVerifications(s.input.Ctx, "nonexistent")
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrDataItemNotFound)
}

func (s *DataItemTestSuite) TestVerificationLevelUpgrade() {
	dataID, _ := s.keeper.StoreDataItem(
		s.input.Ctx,
		"aura1owner",
		types.DataItemType_DATA_ITEM_TYPE_CERTIFICATION,
		"Title",
		"Desc",
		[]byte("hash"),
		"ipfs://test",
		false, nil, nil, nil, nil,
	)

	// Initial level should be self-attested
	item, _ := s.keeper.GetDataItem(s.input.Ctx, dataID)
	s.Require().Equal(types.VerificationLevel_VERIFICATION_LEVEL_SELF_ATTESTED, item.VerificationLevel)

	// Add peer verification
	s.keeper.VerifyDataItem(s.input.Ctx, dataID, "aura1v1", types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED, 80, "", "", nil)

	item, _ = s.keeper.GetDataItem(s.input.Ctx, dataID)
	s.Require().Equal(types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED, item.VerificationLevel)

	// Add institution verification (higher level)
	s.keeper.VerifyDataItem(s.input.Ctx, dataID, "aura1v2", types.VerificationLevel_VERIFICATION_LEVEL_AUTHORITY_VERIFIED, 95, "", "", nil)

	item, _ = s.keeper.GetDataItem(s.input.Ctx, dataID)
	s.Require().Equal(types.VerificationLevel_VERIFICATION_LEVEL_AUTHORITY_VERIFIED, item.VerificationLevel)
}

func (s *DataItemTestSuite) TestMultipleVerifications() {
	dataID, _ := s.keeper.StoreDataItem(
		s.input.Ctx,
		"aura1owner",
		types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Title",
		"Desc",
		[]byte("hash"),
		"ipfs://test",
		false, nil, nil, nil, nil,
	)

	// Add multiple verifications
	for i := 0; i < 5; i++ {
		s.keeper.VerifyDataItem(
			s.input.Ctx,
			dataID,
			"aura1verifier"+string(rune('0'+i)),
			types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
			uint64(70+i*5),
			"verification note",
			"method",
			nil,
		)
	}

	verifications, err := s.keeper.GetDataItemVerifications(s.input.Ctx, dataID)
	s.Require().NoError(err)
	s.Require().Len(verifications, 5)
}

func (s *DataItemTestSuite) TestDataItemWithGeoLocation() {
	geo := &types.GeoLocation{
		Latitude:  40.7128,
		Longitude: -74.0060,
	}

	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx,
		"aura1owner",
		types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"NYC Photo",
		"Photo taken in NYC",
		[]byte("hash"),
		"ipfs://test",
		false,
		geo,
		nil, nil, nil,
	)
	s.Require().NoError(err)

	item, found := s.keeper.GetDataItem(s.input.Ctx, dataID)
	s.Require().True(found)
	s.Require().NotNil(item.GeoLocation)
	s.Require().Equal(40.7128, item.GeoLocation.Latitude)
	s.Require().Equal(-74.0060, item.GeoLocation.Longitude)
}

func (s *DataItemTestSuite) TestDataItemEncrypted() {
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx,
		"aura1owner",
		types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Secret Doc",
		"Encrypted document",
		[]byte("encrypted_hash"),
		"ipfs://encrypted_cid",
		true, // Encrypted
		nil, nil, nil, nil,
	)
	s.Require().NoError(err)

	item, found := s.keeper.GetDataItem(s.input.Ctx, dataID)
	s.Require().True(found)
	s.Require().True(item.IsEncrypted)
}
