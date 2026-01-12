// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"context"
	"testing"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/dataregistry/keeper"
	"github.com/aequitas/aura/chain/x/dataregistry/params"
	"github.com/aequitas/aura/chain/x/dataregistry/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

// MockBankKeeper implements the BankKeeper interface for testing
type MockBankKeeper struct {
	MintCoinsCalled     bool
	SendCoinsCalled     bool
	MintCoinsError      error
	SendCoinsError      error
	MintedModuleName    string
	SentFromModule      string
	SentToAddr          sdk.AccAddress
}

func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{}
}

func (m *MockBankKeeper) MintCoins(ctx context.Context, moduleName string, amt interface{}) error {
	m.MintCoinsCalled = true
	m.MintedModuleName = moduleName
	return m.MintCoinsError
}

func (m *MockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr interface{}, amt interface{}) error {
	m.SendCoinsCalled = true
	m.SentFromModule = senderModule
	if addr, ok := recipientAddr.(sdk.AccAddress); ok {
		m.SentToAddr = addr
	}
	return m.SendCoinsError
}

// DataAdvancedTestSuite is the test suite for data_advanced.go functions
type DataAdvancedTestSuite struct {
	suite.Suite
	input  keepertest.TestInput
	keeper *keeper.Keeper
}

func (s *DataAdvancedTestSuite) SetupTest() {
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

func TestDataAdvancedTestSuite(t *testing.T) {
	suite.Run(t, new(DataAdvancedTestSuite))
}

// ===== ENCRYPTION/DECRYPTION TESTS =====

func (s *DataAdvancedTestSuite) TestEncryptData_ValidKey() {
	// Create a 32-byte encryption key
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	data := []byte("test data for encryption")
	additionalEntropy := []byte("extra-entropy")

	encrypted, nonce, err := s.keeper.EncryptData(s.input.Ctx, data, key, additionalEntropy)

	s.Require().NoError(err)
	s.Require().NotNil(encrypted)
	s.Require().NotNil(nonce)
	s.Require().Len(nonce, 12, "GCM nonce should be 12 bytes")
	s.Require().NotEqual(data, encrypted, "encrypted data should differ from plaintext")
	s.Require().Greater(len(encrypted), len(data), "encrypted data should be larger due to auth tag")
}

func (s *DataAdvancedTestSuite) TestEncryptData_InvalidKeyLength_TooShort() {
	// Create a 16-byte key (too short)
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i)
	}

	data := []byte("test data")

	encrypted, nonce, err := s.keeper.EncryptData(s.input.Ctx, data, key, nil)

	s.Require().Error(err)
	s.Require().Nil(encrypted)
	s.Require().Nil(nonce)
	s.Require().Contains(err.Error(), "encryption key must be 32 bytes")
}

func (s *DataAdvancedTestSuite) TestEncryptData_InvalidKeyLength_TooLong() {
	// Create a 64-byte key (too long)
	key := make([]byte, 64)
	for i := range key {
		key[i] = byte(i)
	}

	data := []byte("test data")

	encrypted, nonce, err := s.keeper.EncryptData(s.input.Ctx, data, key, nil)

	s.Require().Error(err)
	s.Require().Nil(encrypted)
	s.Require().Nil(nonce)
	s.Require().Contains(err.Error(), "encryption key must be 32 bytes")
}

func (s *DataAdvancedTestSuite) TestEncryptDecrypt_RoundTrip() {
	// Create a 32-byte encryption key
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 100)
	}

	originalData := []byte("This is the original plaintext data for round-trip testing!")
	additionalEntropy := []byte("unique-entropy-123")

	// Encrypt
	encrypted, nonce, err := s.keeper.EncryptData(s.input.Ctx, originalData, key, additionalEntropy)
	s.Require().NoError(err)
	s.Require().NotNil(encrypted)
	s.Require().NotNil(nonce)

	// Decrypt
	decrypted, err := s.keeper.DecryptData(encrypted, nonce, key)
	s.Require().NoError(err)
	s.Require().NotNil(decrypted)
	s.Require().Equal(originalData, decrypted, "decrypted data should match original")
}

func (s *DataAdvancedTestSuite) TestDecryptData_ValidDecryption() {
	// Create a 32-byte encryption key
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i * 2)
	}

	originalData := []byte("secret message")

	// First encrypt
	encrypted, nonce, err := s.keeper.EncryptData(s.input.Ctx, originalData, key, nil)
	s.Require().NoError(err)

	// Then decrypt
	decrypted, err := s.keeper.DecryptData(encrypted, nonce, key)
	s.Require().NoError(err)
	s.Require().Equal(originalData, decrypted)
}

func (s *DataAdvancedTestSuite) TestDecryptData_WrongKey() {
	// Create the correct encryption key
	correctKey := make([]byte, 32)
	for i := range correctKey {
		correctKey[i] = byte(i)
	}

	// Create a wrong decryption key
	wrongKey := make([]byte, 32)
	for i := range wrongKey {
		wrongKey[i] = byte(255 - i)
	}

	originalData := []byte("secret message")

	// Encrypt with correct key
	encrypted, nonce, err := s.keeper.EncryptData(s.input.Ctx, originalData, correctKey, nil)
	s.Require().NoError(err)

	// Try to decrypt with wrong key
	decrypted, err := s.keeper.DecryptData(encrypted, nonce, wrongKey)
	s.Require().Error(err)
	s.Require().Nil(decrypted)
	s.Require().Contains(err.Error(), "decryption failed")
}

func (s *DataAdvancedTestSuite) TestDecryptData_WrongNonce() {
	// Create encryption key
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	originalData := []byte("secret message")

	// Encrypt
	encrypted, _, err := s.keeper.EncryptData(s.input.Ctx, originalData, key, nil)
	s.Require().NoError(err)

	// Create a wrong nonce
	wrongNonce := make([]byte, 12)
	for i := range wrongNonce {
		wrongNonce[i] = byte(255 - i)
	}

	// Try to decrypt with wrong nonce
	decrypted, err := s.keeper.DecryptData(encrypted, wrongNonce, key)
	s.Require().Error(err)
	s.Require().Nil(decrypted)
	s.Require().Contains(err.Error(), "decryption failed")
}

func (s *DataAdvancedTestSuite) TestDecryptData_InvalidKeyLength() {
	// Create a short key for decryption
	shortKey := make([]byte, 16)
	encrypted := []byte("some encrypted data")
	nonce := make([]byte, 12)

	decrypted, err := s.keeper.DecryptData(encrypted, nonce, shortKey)
	s.Require().Error(err)
	s.Require().Nil(decrypted)
	s.Require().Contains(err.Error(), "encryption key must be 32 bytes")
}

// ===== DATA VERSIONING TESTS =====

func (s *DataAdvancedTestSuite) TestCreateDataVersion_Success() {
	// First create a data item
	owner := "aura1owner123"
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx,
		owner,
		types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Test Document",
		"A test document for versioning",
		[]byte("original-content-hash"),
		"ipfs://Qm123456",
		false,
		nil,
		map[string]string{"version": "1"},
		nil,
		[]string{"test"},
	)
	s.Require().NoError(err)
	s.Require().NotEmpty(dataID)

	// Create a new version
	newContentHash := []byte("updated-content-hash-v2")
	newStorageLocation := "ipfs://Qm789012"
	changeLog := "Updated document content"

	version, err := s.keeper.CreateDataVersion(
		s.input.Ctx,
		dataID,
		owner,
		newContentHash,
		newStorageLocation,
		changeLog,
	)

	s.Require().NoError(err)
	s.Require().NotNil(version)
	s.Require().Equal(dataID, version.DataID)
	s.Require().Equal(uint64(1), version.VersionNum)
	s.Require().Equal(newStorageLocation, version.StorageLocation)
	s.Require().Equal(changeLog, version.ChangeLog)
	s.Require().Equal(owner, version.CreatedBy)
	s.Require().NotEmpty(version.VersionID)

	// Verify the data item was updated
	item, found := s.keeper.GetDataItem(s.input.Ctx, dataID)
	s.Require().True(found)
	s.Require().Equal(newContentHash, item.ContentHash)
	s.Require().Equal(newStorageLocation, item.StorageLocation)
}

func (s *DataAdvancedTestSuite) TestCreateDataVersion_ItemNotFound() {
	version, err := s.keeper.CreateDataVersion(
		s.input.Ctx,
		"nonexistent-data-id",
		"aura1owner",
		[]byte("hash"),
		"ipfs://test",
		"changelog",
	)

	s.Require().Error(err)
	s.Require().Nil(version)
	s.Require().ErrorIs(err, types.ErrDataItemNotFound)
}

func (s *DataAdvancedTestSuite) TestCreateDataVersion_Unauthorized() {
	// Create a data item with one owner
	owner := "aura1owner123"
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx,
		owner,
		types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Test Doc",
		"Description",
		[]byte("hash"),
		"ipfs://test",
		false, nil, nil, nil, nil,
	)
	s.Require().NoError(err)

	// Try to create version as different user
	differentOwner := "aura1different456"
	version, err := s.keeper.CreateDataVersion(
		s.input.Ctx,
		dataID,
		differentOwner,
		[]byte("new-hash"),
		"ipfs://new",
		"unauthorized update",
	)

	s.Require().Error(err)
	s.Require().Nil(version)
	s.Require().ErrorIs(err, types.ErrUnauthorized)
}

func (s *DataAdvancedTestSuite) TestGetDataVersions() {
	// Create a data item
	owner := "aura1owner"
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx,
		owner,
		types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Title",
		"Desc",
		[]byte("hash"),
		"ipfs://test",
		false, nil, nil, nil, nil,
	)
	s.Require().NoError(err)

	// Get versions (should be empty since helper storage is stub)
	versions := s.keeper.GetDataVersions(s.input.Ctx, dataID)
	s.Require().NotNil(versions)
	// Note: The stub implementation returns empty slice
	s.Require().Len(versions, 0)
}

func (s *DataAdvancedTestSuite) TestRestoreDataVersion_Success() {
	// Create a data item
	owner := "aura1owner"
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx,
		owner,
		types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Title",
		"Desc",
		[]byte("hash"),
		"ipfs://test",
		false, nil, nil, nil, nil,
	)
	s.Require().NoError(err)

	// Note: Since getDataVersions returns empty slice in the stub implementation,
	// RestoreDataVersion will fail with "version not found"
	err = s.keeper.RestoreDataVersion(s.input.Ctx, dataID, 1, owner)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "version not found")
}

func (s *DataAdvancedTestSuite) TestRestoreDataVersion_ItemNotFound() {
	err := s.keeper.RestoreDataVersion(s.input.Ctx, "nonexistent", 1, "aura1owner")
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrDataItemNotFound)
}

func (s *DataAdvancedTestSuite) TestRestoreDataVersion_Unauthorized() {
	// Create a data item with one owner
	owner := "aura1owner"
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx,
		owner,
		types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Title",
		"Desc",
		[]byte("hash"),
		"ipfs://test",
		false, nil, nil, nil, nil,
	)
	s.Require().NoError(err)

	// Try to restore as different user
	err = s.keeper.RestoreDataVersion(s.input.Ctx, dataID, 1, "aura1different")
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrUnauthorized)
}

// ===== PROVENANCE TRACKING TESTS =====

func (s *DataAdvancedTestSuite) TestRecordProvenance() {
	dataID := "test-data-id"
	action := "created"
	actor := "aura1actor"
	details := map[string]string{
		"method": "manual",
		"source": "upload",
	}

	err := s.keeper.RecordProvenance(s.input.Ctx, dataID, action, actor, details)
	s.Require().NoError(err)
}

func (s *DataAdvancedTestSuite) TestGetProvenanceTrail() {
	dataID := "test-data-id"

	// Record some provenance events
	s.keeper.RecordProvenance(s.input.Ctx, dataID, "created", "aura1creator", nil)
	s.keeper.RecordProvenance(s.input.Ctx, dataID, "updated", "aura1updater", nil)

	// Get trail (stub returns empty slice)
	trail := s.keeper.GetProvenanceTrail(s.input.Ctx, dataID)
	s.Require().NotNil(trail)
	s.Require().Len(trail, 0) // Stub implementation
}

// ===== RETENTION POLICY TESTS =====

func (s *DataAdvancedTestSuite) TestSetRetentionPolicy_Success() {
	// Create a data item with initialized metadata
	owner := "aura1owner"
	metadata := make(map[string]string)
	metadata["initial"] = "value"
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx,
		owner,
		types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Title",
		"Desc",
		[]byte("hash"),
		"ipfs://test",
		false, nil,
		metadata,
		nil, nil,
	)
	s.Require().NoError(err)

	// Set retention policy
	retentionDays := uint64(365)
	autoDelete := true
	notifyBefore := uint64(30)

	policy, err := s.keeper.SetRetentionPolicy(
		s.input.Ctx,
		dataID,
		retentionDays,
		autoDelete,
		notifyBefore,
	)

	s.Require().NoError(err)
	s.Require().NotNil(policy)
	s.Require().Equal(dataID, policy.DataID)
	s.Require().Equal(retentionDays, policy.RetentionDays)
	s.Require().Equal(autoDelete, policy.AutoDelete)
	s.Require().Equal(notifyBefore, policy.NotifyBefore)
	s.Require().NotEmpty(policy.PolicyID)

	// Verify expiration date is set correctly
	expectedExpiry := s.input.Ctx.BlockTime().AddDate(0, 0, int(retentionDays))
	s.Require().Equal(expectedExpiry, policy.ExpiresAt)

	// Verify metadata was updated
	item, found := s.keeper.GetDataItem(s.input.Ctx, dataID)
	s.Require().True(found)
	s.Require().Equal(policy.PolicyID, item.Metadata["retention_policy"])
}

func (s *DataAdvancedTestSuite) TestSetRetentionPolicy_ItemNotFound() {
	policy, err := s.keeper.SetRetentionPolicy(
		s.input.Ctx,
		"nonexistent",
		365,
		true,
		30,
	)

	s.Require().Error(err)
	s.Require().Nil(policy)
	s.Require().ErrorIs(err, types.ErrDataItemNotFound)
}

func (s *DataAdvancedTestSuite) TestProcessExpiredData() {
	// Call ProcessExpiredData (stub implementation returns 0, 0)
	deleted, notified, err := s.keeper.ProcessExpiredData(s.input.Ctx)

	s.Require().NoError(err)
	s.Require().Equal(0, deleted)
	s.Require().Equal(0, notified)
}

// ===== QUALITY SCORE TESTS =====

func (s *DataAdvancedTestSuite) TestCalculateQualityScore_Success() {
	// Create a fully populated data item for maximum quality score
	owner := "aura1owner"
	geo := &types.GeoLocation{
		Latitude:  40.7128,
		Longitude: -74.0060,
	}
	metadata := map[string]string{
		"author": "test",
		"format": "pdf",
	}

	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx,
		owner,
		types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Complete Document",
		"This is a fully described document for quality testing",
		// Create a 64-byte hash to trigger 100% consistency score
		make([]byte, 64),
		"ipfs://Qm123456",
		false,
		geo,
		metadata,
		nil,
		[]string{"test", "quality"},
	)
	s.Require().NoError(err)

	// Add verifications to improve accuracy score
	s.keeper.VerifyDataItem(s.input.Ctx, dataID, "aura1v1", types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED, 80, "", "", nil)
	s.keeper.VerifyDataItem(s.input.Ctx, dataID, "aura1v2", types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED, 85, "", "", nil)
	s.keeper.VerifyDataItem(s.input.Ctx, dataID, "aura1v3", types.VerificationLevel_VERIFICATION_LEVEL_AUTHORITY_VERIFIED, 95, "", "", nil)

	// Calculate quality score
	score, err := s.keeper.CalculateQualityScore(s.input.Ctx, dataID)

	s.Require().NoError(err)
	s.Require().NotNil(score)
	s.Require().Equal(dataID, score.DataID)

	// Completeness: 100 (all fields present: title, description, metadata, tags, geo)
	s.Require().Equal(uint64(100), score.CompletenessScore)

	// Accuracy: 100 (3+ verifications)
	s.Require().Equal(uint64(100), score.AccuracyScore)

	// Timeliness: 100 (just created)
	s.Require().Equal(uint64(100), score.TimelinessScore)

	// Consistency: 100 (64-byte hash)
	s.Require().Equal(uint64(100), score.ConsistencyScore)

	// Overall should be weighted average: (100*30 + 100*40 + 100*15 + 100*15) / 100 = 100
	s.Require().Equal(uint64(100), score.OverallScore)
}

func (s *DataAdvancedTestSuite) TestCalculateQualityScore_MinimalItem() {
	// Create a minimal data item
	owner := "aura1owner"
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx,
		owner,
		types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"", // Empty title
		"", // Empty description
		[]byte("short"), // Short hash (not 64 bytes)
		"ipfs://test",
		false,
		nil,  // No geo
		nil,  // No metadata
		nil,
		nil,  // No tags
	)
	s.Require().NoError(err)

	score, err := s.keeper.CalculateQualityScore(s.input.Ctx, dataID)

	s.Require().NoError(err)
	s.Require().NotNil(score)

	// Completeness: 0 (no title, description, metadata, tags, geo)
	s.Require().Equal(uint64(0), score.CompletenessScore)

	// Accuracy: 25 (no verifications)
	s.Require().Equal(uint64(25), score.AccuracyScore)

	// Timeliness: 100 (just created)
	s.Require().Equal(uint64(100), score.TimelinessScore)

	// Consistency: 50 (hash is not 64 bytes)
	s.Require().Equal(uint64(50), score.ConsistencyScore)
}

func (s *DataAdvancedTestSuite) TestCalculateQualityScore_NotFound() {
	score, err := s.keeper.CalculateQualityScore(s.input.Ctx, "nonexistent")

	s.Require().Error(err)
	s.Require().Nil(score)
	s.Require().ErrorIs(err, types.ErrDataItemNotFound)
}

func (s *DataAdvancedTestSuite) TestCalculateQualityScore_PartialCompleteness() {
	// Create an item with some fields filled
	owner := "aura1owner"
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx,
		owner,
		types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Title Only", // Has title
		"",           // No description
		[]byte("hash"),
		"ipfs://test",
		false,
		nil,
		map[string]string{"key": "value"}, // Has metadata
		nil,
		nil, // No tags
	)
	s.Require().NoError(err)

	score, err := s.keeper.CalculateQualityScore(s.input.Ctx, dataID)

	s.Require().NoError(err)
	s.Require().NotNil(score)

	// Completeness: 40 (title=20, metadata=20, no desc, no tags, no geo)
	s.Require().Equal(uint64(40), score.CompletenessScore)
}

func (s *DataAdvancedTestSuite) TestCalculateQualityScore_VariousVerificationLevels() {
	owner := "aura1owner"

	// Test with 1 verification
	dataID1, _ := s.keeper.StoreDataItem(
		s.input.Ctx, owner, types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Photo 1", "Desc", []byte("hash1"), "ipfs://1",
		false, nil, nil, nil, nil,
	)
	s.keeper.VerifyDataItem(s.input.Ctx, dataID1, "aura1v1", types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED, 80, "", "", nil)

	score1, err := s.keeper.CalculateQualityScore(s.input.Ctx, dataID1)
	s.Require().NoError(err)
	s.Require().Equal(uint64(50), score1.AccuracyScore) // 1 verification = 50

	// Test with 2 verifications
	dataID2, _ := s.keeper.StoreDataItem(
		s.input.Ctx, owner, types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Photo 2", "Desc", []byte("hash2"), "ipfs://2",
		false, nil, nil, nil, nil,
	)
	s.keeper.VerifyDataItem(s.input.Ctx, dataID2, "aura1v1", types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED, 80, "", "", nil)
	s.keeper.VerifyDataItem(s.input.Ctx, dataID2, "aura1v2", types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED, 85, "", "", nil)

	score2, err := s.keeper.CalculateQualityScore(s.input.Ctx, dataID2)
	s.Require().NoError(err)
	s.Require().Equal(uint64(75), score2.AccuracyScore) // 2 verifications = 75
}

func (s *DataAdvancedTestSuite) TestCalculateQualityScore_Timeliness() {
	owner := "aura1owner"

	// Note: The timeliness calculation uses time.Since(createdAt) which compares
	// against the real wall clock, not the context block time.
	// Since items are created with current block time (which is close to now),
	// they will always have 100% timeliness in tests.

	// Create item
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx, owner, types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Photo", "Desc", []byte("hash"), "ipfs://test",
		false, nil, nil, nil, nil,
	)
	s.Require().NoError(err)

	// Just created - should be 100 (age < 7 days since wall clock time is used)
	score, err := s.keeper.CalculateQualityScore(s.input.Ctx, dataID)
	s.Require().NoError(err)
	s.Require().Equal(uint64(100), score.TimelinessScore)

	// Note: We cannot truly test aged items without modifying the CreatedAt field
	// directly or creating items in the past. The timeliness function uses wall clock
	// time via time.Since() which cannot be mocked through context manipulation.
	// For production, the CreatedAt field determines timeliness.

	// Verify the timeliness score is calculated and reasonable
	s.Require().LessOrEqual(score.TimelinessScore, uint64(100))
	s.Require().GreaterOrEqual(score.TimelinessScore, uint64(25))
}

// ===== MINT VERIFICATION REWARD TESTS =====

func (s *DataAdvancedTestSuite) TestMintVerificationReward_Success() {
	// Ensure SDK is configured for aura prefix
	keepertest.ConfigureSDK()

	// Create mock bank keeper
	mockBank := NewMockBankKeeper()

	// Create a data item first
	owner := "aura1owner"
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx, owner, types.DataItemType_DATA_ITEM_TYPE_CERTIFICATION,
		"Certification", "Test cert", []byte("hash"), "ipfs://test",
		false, nil, nil, nil, nil,
	)
	s.Require().NoError(err)

	// Generate a valid aura address for verifier
	verifierAddr := keepertest.GenTestAddr()
	verifier := verifierAddr.String()

	// Mint verification reward
	err = s.keeper.MintVerificationReward(s.input.Ctx, verifier, dataID, mockBank)

	s.Require().NoError(err)
	s.Require().True(mockBank.MintCoinsCalled)
	s.Require().True(mockBank.SendCoinsCalled)
	s.Require().Equal(types.ModuleName, mockBank.MintedModuleName)
	s.Require().Equal(types.ModuleName, mockBank.SentFromModule)
	s.Require().Equal(verifierAddr, mockBank.SentToAddr)
}

func (s *DataAdvancedTestSuite) TestMintVerificationReward_InvalidVerifierAddress() {
	// Ensure SDK is configured for aura prefix
	keepertest.ConfigureSDK()

	mockBank := NewMockBankKeeper()

	// Create a data item
	owner := "aura1owner"
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx, owner, types.DataItemType_DATA_ITEM_TYPE_CERTIFICATION,
		"Cert", "Desc", []byte("hash"), "ipfs://test",
		false, nil, nil, nil, nil,
	)
	s.Require().NoError(err)

	// Use invalid address format
	invalidVerifier := "invalid-address"

	err = s.keeper.MintVerificationReward(s.input.Ctx, invalidVerifier, dataID, mockBank)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invalid verifier address")
	// MintCoins should still be called (before address validation)
	s.Require().True(mockBank.MintCoinsCalled)
	// SendCoins should not be called due to address validation failure
	s.Require().False(mockBank.SendCoinsCalled)
}

func (s *DataAdvancedTestSuite) TestMintVerificationReward_MintCoinsError() {
	// Ensure SDK is configured for aura prefix
	keepertest.ConfigureSDK()

	mockBank := NewMockBankKeeper()
	mockBank.MintCoinsError = types.ErrUnauthorized

	// Generate valid verifier address
	verifierAddr := keepertest.GenTestAddr()

	err := s.keeper.MintVerificationReward(s.input.Ctx, verifierAddr.String(), "test-data-id", mockBank)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "failed to mint verification reward")
	s.Require().True(mockBank.MintCoinsCalled)
	s.Require().False(mockBank.SendCoinsCalled)
}

func (s *DataAdvancedTestSuite) TestMintVerificationReward_SendCoinsError() {
	// Ensure SDK is configured for aura prefix
	keepertest.ConfigureSDK()

	mockBank := NewMockBankKeeper()
	mockBank.SendCoinsError = types.ErrUnauthorized

	// Create a data item
	owner := "aura1owner"
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx, owner, types.DataItemType_DATA_ITEM_TYPE_CERTIFICATION,
		"Cert", "Desc", []byte("hash"), "ipfs://test",
		false, nil, nil, nil, nil,
	)
	s.Require().NoError(err)

	// Generate valid verifier address
	verifierAddr := keepertest.GenTestAddr()

	err = s.keeper.MintVerificationReward(s.input.Ctx, verifierAddr.String(), dataID, mockBank)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "failed to send verification reward")
	s.Require().True(mockBank.MintCoinsCalled)
	s.Require().True(mockBank.SendCoinsCalled)
}

// ===== ENCRYPTION EDGE CASES =====

func (s *DataAdvancedTestSuite) TestEncryptData_EmptyData() {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	// Empty data should still encrypt
	encrypted, nonce, err := s.keeper.EncryptData(s.input.Ctx, []byte{}, key, nil)

	s.Require().NoError(err)
	s.Require().NotNil(encrypted)
	s.Require().NotNil(nonce)
	// GCM adds authentication tag even for empty data
	s.Require().Greater(len(encrypted), 0)
}

func (s *DataAdvancedTestSuite) TestEncryptData_LargeData() {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	// Create 1MB of data
	largeData := make([]byte, 1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	encrypted, nonce, err := s.keeper.EncryptData(s.input.Ctx, largeData, key, nil)

	s.Require().NoError(err)
	s.Require().NotNil(encrypted)
	s.Require().NotNil(nonce)

	// Decrypt and verify
	decrypted, err := s.keeper.DecryptData(encrypted, nonce, key)
	s.Require().NoError(err)
	s.Require().Equal(largeData, decrypted)
}

func (s *DataAdvancedTestSuite) TestEncryptData_DifferentEntropySameData() {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	data := []byte("same data")

	// Encrypt with different entropy values
	encrypted1, nonce1, err := s.keeper.EncryptData(s.input.Ctx, data, key, []byte("entropy1"))
	s.Require().NoError(err)

	encrypted2, nonce2, err := s.keeper.EncryptData(s.input.Ctx, data, key, []byte("entropy2"))
	s.Require().NoError(err)

	// Different entropy should produce different nonces
	s.Require().NotEqual(nonce1, nonce2)

	// And different ciphertexts
	s.Require().NotEqual(encrypted1, encrypted2)

	// But both should decrypt to the same plaintext
	decrypted1, err := s.keeper.DecryptData(encrypted1, nonce1, key)
	s.Require().NoError(err)
	s.Require().Equal(data, decrypted1)

	decrypted2, err := s.keeper.DecryptData(encrypted2, nonce2, key)
	s.Require().NoError(err)
	s.Require().Equal(data, decrypted2)
}

// ===== DATA VERSION EDGE CASES =====

func (s *DataAdvancedTestSuite) TestCreateDataVersion_MultipleVersions() {
	owner := "aura1owner"
	metadata := make(map[string]string)
	metadata["ver"] = "0"
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx, owner, types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Doc", "Desc", []byte("v0-hash"), "ipfs://v0",
		false, nil, metadata, nil, nil,
	)
	s.Require().NoError(err)

	// Create multiple versions
	// Note: Since getDataVersions() is a stub returning empty slice,
	// each version will have VersionNum = 1 (len([])+1)
	for i := 1; i <= 3; i++ {
		version, err := s.keeper.CreateDataVersion(
			s.input.Ctx,
			dataID,
			owner,
			[]byte("v" + string(rune('0'+i)) + "-hash"),
			"ipfs://v" + string(rune('0'+i)),
			"Version " + string(rune('0'+i)),
		)
		s.Require().NoError(err)
		s.Require().NotNil(version)
		// Stub implementation returns empty versions array, so version num is always 1
		s.Require().Equal(uint64(1), version.VersionNum)
	}
}

// ===== RETENTION POLICY EDGE CASES =====

func (s *DataAdvancedTestSuite) TestSetRetentionPolicy_ZeroRetentionDays() {
	owner := "aura1owner"
	metadata := make(map[string]string)
	metadata["key"] = "value"
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx, owner, types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Doc", "Desc", []byte("hash"), "ipfs://test",
		false, nil, metadata, nil, nil,
	)
	s.Require().NoError(err)

	// Set retention with 0 days (immediate expiry)
	policy, err := s.keeper.SetRetentionPolicy(s.input.Ctx, dataID, 0, true, 0)

	s.Require().NoError(err)
	s.Require().NotNil(policy)
	s.Require().Equal(uint64(0), policy.RetentionDays)
	// ExpiresAt should be same as block time
	s.Require().Equal(s.input.Ctx.BlockTime(), policy.ExpiresAt)
}

func (s *DataAdvancedTestSuite) TestSetRetentionPolicy_AutoDeleteFalse() {
	owner := "aura1owner"
	metadata := make(map[string]string)
	metadata["key"] = "value"
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx, owner, types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Doc", "Desc", []byte("hash"), "ipfs://test",
		false, nil, metadata, nil, nil,
	)
	s.Require().NoError(err)

	// Set retention without auto-delete
	policy, err := s.keeper.SetRetentionPolicy(s.input.Ctx, dataID, 30, false, 7)

	s.Require().NoError(err)
	s.Require().NotNil(policy)
	s.Require().False(policy.AutoDelete)
}

// ===== INTEGRATION TESTS =====

func (s *DataAdvancedTestSuite) TestIntegration_FullDataLifecycle() {
	// Ensure SDK is configured
	keepertest.ConfigureSDK()

	owner := "aura1lifecycleowner"
	mockBank := NewMockBankKeeper()

	// Create exactly 64-byte content hash for 100% consistency score
	contentHash := make([]byte, 64)
	for i := range contentHash {
		contentHash[i] = byte(i % 256)
	}

	metadata := make(map[string]string)
	metadata["issuer"] = "AWS"
	metadata["level"] = "Associate"

	// 1. Create data item
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx, owner, types.DataItemType_DATA_ITEM_TYPE_CERTIFICATION,
		"Professional Certification",
		"AWS Solutions Architect certification",
		contentHash,
		"ipfs://QmCertification123",
		false,
		&types.GeoLocation{Latitude: 37.7749, Longitude: -122.4194},
		metadata,
		nil,
		[]string{"aws", "certification", "cloud"},
	)
	s.Require().NoError(err)

	// 2. Record provenance
	err = s.keeper.RecordProvenance(s.input.Ctx, dataID, "created", owner, map[string]string{
		"source": "manual_upload",
	})
	s.Require().NoError(err)

	// 3. Set retention policy
	policy, err := s.keeper.SetRetentionPolicy(s.input.Ctx, dataID, 1095, false, 90) // 3 years
	s.Require().NoError(err)
	s.Require().NotNil(policy)

	// 4. Add verifications
	verifier1 := keepertest.GenTestAddr()
	verifier2 := keepertest.GenTestAddr()
	verifier3 := keepertest.GenTestAddr()

	s.keeper.VerifyDataItem(s.input.Ctx, dataID, verifier1.String(), types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED, 85, "", "", nil)
	s.keeper.VerifyDataItem(s.input.Ctx, dataID, verifier2.String(), types.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED, 90, "", "", nil)
	s.keeper.VerifyDataItem(s.input.Ctx, dataID, verifier3.String(), types.VerificationLevel_VERIFICATION_LEVEL_AUTHORITY_VERIFIED, 95, "", "", nil)

	// 5. Mint rewards for verifications
	err = s.keeper.MintVerificationReward(s.input.Ctx, verifier1.String(), dataID, mockBank)
	s.Require().NoError(err)

	// 6. Calculate quality score
	score, err := s.keeper.CalculateQualityScore(s.input.Ctx, dataID)
	s.Require().NoError(err)
	s.Require().NotNil(score)
	s.Require().Equal(uint64(100), score.CompletenessScore)
	s.Require().Equal(uint64(100), score.AccuracyScore)
	s.Require().Equal(uint64(100), score.TimelinessScore)
	s.Require().Equal(uint64(100), score.ConsistencyScore)
	s.Require().Equal(uint64(100), score.OverallScore)

	// 7. Create a new version
	version, err := s.keeper.CreateDataVersion(
		s.input.Ctx,
		dataID,
		owner,
		[]byte("updated-cert-hash-v2"),
		"ipfs://QmCertificationV2",
		"Renewed certification",
	)
	s.Require().NoError(err)
	s.Require().NotNil(version)

	// 8. Verify final state
	item, found := s.keeper.GetDataItem(s.input.Ctx, dataID)
	s.Require().True(found)
	s.Require().Equal(types.DataItemStatus_DATA_ITEM_STATUS_VERIFIED, item.Status)
	s.Require().Equal(types.VerificationLevel_VERIFICATION_LEVEL_AUTHORITY_VERIFIED, item.VerificationLevel)
	s.Require().Len(item.Verifications, 3)
}

// ===== ADDITIONAL EDGE CASE TESTS =====

func (s *DataAdvancedTestSuite) TestCalculateQualityScore_NilCreatedAt() {
	// Create an item
	owner := "aura1owner"
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx, owner, types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		"Photo", "Desc", []byte("hash"), "ipfs://test",
		false, nil, nil, nil, nil,
	)
	s.Require().NoError(err)

	// Get and manually modify item to have nil CreatedAt
	item, found := s.keeper.GetDataItem(s.input.Ctx, dataID)
	s.Require().True(found)
	item.CreatedAt = nil
	s.keeper.SetDataItem(s.input.Ctx, item)

	// Calculate score - should handle nil CreatedAt gracefully
	score, err := s.keeper.CalculateQualityScore(s.input.Ctx, dataID)
	s.Require().NoError(err)
	s.Require().NotNil(score)
	// Timeliness should be 25 for nil CreatedAt
	s.Require().Equal(uint64(25), score.TimelinessScore)
}

func (s *DataAdvancedTestSuite) TestEncryptData_NilEntropy() {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	data := []byte("test data")

	// Encrypt with nil entropy
	encrypted, nonce, err := s.keeper.EncryptData(s.input.Ctx, data, key, nil)
	s.Require().NoError(err)
	s.Require().NotNil(encrypted)
	s.Require().NotNil(nonce)

	// Decrypt should work
	decrypted, err := s.keeper.DecryptData(encrypted, nonce, key)
	s.Require().NoError(err)
	s.Require().Equal(data, decrypted)
}

func (s *DataAdvancedTestSuite) TestDecryptData_TamperedCiphertext() {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	data := []byte("secret message")

	// Encrypt
	encrypted, nonce, err := s.keeper.EncryptData(s.input.Ctx, data, key, nil)
	s.Require().NoError(err)

	// Tamper with ciphertext
	if len(encrypted) > 0 {
		encrypted[0] ^= 0xFF
	}

	// Try to decrypt - should fail due to authentication
	decrypted, err := s.keeper.DecryptData(encrypted, nonce, key)
	s.Require().Error(err)
	s.Require().Nil(decrypted)
	s.Require().Contains(err.Error(), "decryption failed")
}

func (s *DataAdvancedTestSuite) TestRecordProvenance_NilDetails() {
	// Record provenance with nil details
	err := s.keeper.RecordProvenance(s.input.Ctx, "data-id", "action", "actor", nil)
	s.Require().NoError(err)
}

func (s *DataAdvancedTestSuite) TestRecordProvenance_EmptyDetails() {
	// Record provenance with empty details
	err := s.keeper.RecordProvenance(s.input.Ctx, "data-id", "action", "actor", map[string]string{})
	s.Require().NoError(err)
}

func (s *DataAdvancedTestSuite) TestSetRetentionPolicy_LargeRetentionDays() {
	owner := "aura1owner"
	metadata := make(map[string]string)
	metadata["key"] = "value"
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx, owner, types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Doc", "Desc", []byte("hash"), "ipfs://test",
		false, nil, metadata, nil, nil,
	)
	s.Require().NoError(err)

	// Set retention with large retention days (10 years)
	policy, err := s.keeper.SetRetentionPolicy(s.input.Ctx, dataID, 3650, false, 365)

	s.Require().NoError(err)
	s.Require().NotNil(policy)
	s.Require().Equal(uint64(3650), policy.RetentionDays)
	s.Require().Equal(uint64(365), policy.NotifyBefore)
}

func (s *DataAdvancedTestSuite) TestRestoreDataVersion_VersionNotFound() {
	// Create a data item
	owner := "aura1owner"
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx, owner, types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Title", "Desc", []byte("hash"), "ipfs://test",
		false, nil, nil, nil, nil,
	)
	s.Require().NoError(err)

	// Try to restore a version that doesn't exist
	// The stub returns empty versions, so any version number will be not found
	err = s.keeper.RestoreDataVersion(s.input.Ctx, dataID, 999, owner)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "version not found")
}

func (s *DataAdvancedTestSuite) TestCalculateQualityScore_EmptyKey() {
	// Create an item with empty key in metadata
	owner := "aura1owner"
	metadata := make(map[string]string)
	metadata[""] = "empty key"

	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx, owner, types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Title", "", // Empty description
		[]byte("hash"), "ipfs://test",
		false, nil, metadata, nil, nil,
	)
	s.Require().NoError(err)

	score, err := s.keeper.CalculateQualityScore(s.input.Ctx, dataID)
	s.Require().NoError(err)
	s.Require().NotNil(score)
	// Completeness: 40 (title=20, metadata=20)
	s.Require().Equal(uint64(40), score.CompletenessScore)
}

func (s *DataAdvancedTestSuite) TestMintVerificationReward_EmptyAddress() {
	keepertest.ConfigureSDK()
	mockBank := NewMockBankKeeper()

	// Create a data item
	owner := "aura1owner"
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx, owner, types.DataItemType_DATA_ITEM_TYPE_CERTIFICATION,
		"Cert", "Desc", []byte("hash"), "ipfs://test",
		false, nil, nil, nil, nil,
	)
	s.Require().NoError(err)

	// Empty address should fail
	err = s.keeper.MintVerificationReward(s.input.Ctx, "", dataID, mockBank)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invalid verifier address")
}

func (s *DataAdvancedTestSuite) TestEncryptData_DeterministicNonce() {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	data := []byte("test data")
	entropy := []byte("same-entropy")

	// Encrypt twice with same context and entropy
	encrypted1, nonce1, err := s.keeper.EncryptData(s.input.Ctx, data, key, entropy)
	s.Require().NoError(err)

	encrypted2, nonce2, err := s.keeper.EncryptData(s.input.Ctx, data, key, entropy)
	s.Require().NoError(err)

	// With deterministic RNG based on block context and same entropy,
	// nonces should be the same
	s.Require().Equal(nonce1, nonce2)
	s.Require().Equal(encrypted1, encrypted2)
}

func (s *DataAdvancedTestSuite) TestCreateDataVersion_LongChangeLog() {
	owner := "aura1owner"
	dataID, err := s.keeper.StoreDataItem(
		s.input.Ctx, owner, types.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF,
		"Doc", "Desc", []byte("original-hash"), "ipfs://original",
		false, nil, map[string]string{}, nil, nil,
	)
	s.Require().NoError(err)

	// Create a long changelog
	longChangeLog := ""
	for i := 0; i < 100; i++ {
		longChangeLog += "This is a detailed changelog entry. "
	}

	version, err := s.keeper.CreateDataVersion(
		s.input.Ctx,
		dataID,
		owner,
		[]byte("new-hash"),
		"ipfs://new",
		longChangeLog,
	)

	s.Require().NoError(err)
	s.Require().NotNil(version)
	s.Require().Equal(longChangeLog, version.ChangeLog)
}
