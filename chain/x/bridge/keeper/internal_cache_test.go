// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/bridge/types"
)

type InternalCacheTestSuite struct {
	KeeperTestSuite
}

func TestInternalCacheTestSuite(t *testing.T) {
	suite.Run(t, new(InternalCacheTestSuite))
}

func (suite *InternalCacheTestSuite) SetupTest() {
	suite.KeeperTestSuite.SetupTest()
}

// =============================================================================
// transferCacheKey Tests (internal function)
// =============================================================================

func (suite *InternalCacheTestSuite) TestTransferCacheKey_ValidID() {
	key := transferCacheKey("transfer-123")
	suite.Equal("transfer-123", key)
}

func (suite *InternalCacheTestSuite) TestTransferCacheKey_EmptyID() {
	key := transferCacheKey("")
	suite.Equal("", key)
}

func (suite *InternalCacheTestSuite) TestTransferCacheKey_SpecialChars() {
	key := transferCacheKey("transfer-abc-123_test")
	suite.Equal("transfer-abc-123_test", key)
}

func (suite *InternalCacheTestSuite) TestTransferCacheKey_LongID() {
	longID := "transfer-" + string(make([]byte, 1000))
	key := transferCacheKey(longID)
	suite.Equal(longID, key)
}

// =============================================================================
// getTransferWithCache Tests (internal function)
// =============================================================================

func (suite *InternalCacheTestSuite) TestGetTransferWithCache_EmptyID() {
	transfer, found := suite.Keeper.getTransferWithCache(suite.SdkCtx, "")
	suite.False(found)
	suite.Nil(transfer)
}

func (suite *InternalCacheTestSuite) TestGetTransferWithCache_NotFound() {
	transfer, found := suite.Keeper.getTransferWithCache(suite.SdkCtx, "nonexistent-transfer")
	suite.False(found)
	suite.Nil(transfer)
}

func (suite *InternalCacheTestSuite) TestGetTransferWithCache_Found() {
	// First create a transfer using public API
	testTransfer := &types.CrossChainTransfer{
		TransferId:  "internal-cache-test-1",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x1234567890",
		Amount:      sdkmath.NewInt(1000),
		Denom:       "uaura",
		Status:      types.TransferStatus_PENDING,
	}
	err := suite.Keeper.setTransfer(suite.SdkCtx, testTransfer)
	suite.NoError(err)

	// Now retrieve it with cache
	retrieved, found := suite.Keeper.getTransferWithCache(suite.SdkCtx, "internal-cache-test-1")
	suite.True(found)
	suite.NotNil(retrieved)
	suite.Equal("internal-cache-test-1", retrieved.TransferId)
	suite.Equal("aura", retrieved.SourceChain)
}

func (suite *InternalCacheTestSuite) TestGetTransferWithCache_CacheHitAndMiss() {
	// Create and store a transfer
	testTransfer := &types.CrossChainTransfer{
		TransferId:  "cache-hit-miss-test",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0xabc",
		Amount:      sdkmath.NewInt(500),
		Denom:       "uaura",
		Status:      types.TransferStatus_CONFIRMED,
	}
	err := suite.Keeper.setTransfer(suite.SdkCtx, testTransfer)
	suite.NoError(err)

	// First retrieval - cache miss, loads from store
	retrieved1, found1 := suite.Keeper.getTransferWithCache(suite.SdkCtx, "cache-hit-miss-test")
	suite.True(found1)
	suite.NotNil(retrieved1)

	// Second retrieval - should be cache hit
	retrieved2, found2 := suite.Keeper.getTransferWithCache(suite.SdkCtx, "cache-hit-miss-test")
	suite.True(found2)
	suite.NotNil(retrieved2)
	suite.Equal(retrieved1.TransferId, retrieved2.TransferId)
}

// =============================================================================
// setTransferWithCache Tests (internal function)
// =============================================================================

func (suite *InternalCacheTestSuite) TestSetTransferWithCache_NilTransfer() {
	err := suite.Keeper.setTransferWithCache(suite.SdkCtx, nil)
	suite.NoError(err) // nil transfer should be handled gracefully
}

func (suite *InternalCacheTestSuite) TestSetTransferWithCache_EmptyID() {
	transfer := &types.CrossChainTransfer{
		TransferId: "", // Empty ID
	}
	err := suite.Keeper.setTransferWithCache(suite.SdkCtx, transfer)
	suite.NoError(err) // Empty ID should be handled gracefully
}

func (suite *InternalCacheTestSuite) TestSetTransferWithCache_Valid() {
	transfer := &types.CrossChainTransfer{
		TransferId:  "set-cache-internal-test",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0xdef",
		Amount:      sdkmath.NewInt(2000),
		Denom:       "uaura",
		Status:      types.TransferStatus_PENDING,
	}

	err := suite.Keeper.setTransferWithCache(suite.SdkCtx, transfer)
	suite.NoError(err)

	// Verify it can be retrieved
	retrieved, found := suite.Keeper.getTransferWithCache(suite.SdkCtx, "set-cache-internal-test")
	suite.True(found)
	suite.NotNil(retrieved)
	suite.Equal("set-cache-internal-test", retrieved.TransferId)
}

func (suite *InternalCacheTestSuite) TestSetTransferWithCache_UpdateExisting() {
	// Create initial transfer
	transfer := &types.CrossChainTransfer{
		TransferId:  "update-cache-internal-test",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x111",
		Amount:      sdkmath.NewInt(1000),
		Denom:       "uaura",
		Status:      types.TransferStatus_PENDING,
	}
	err := suite.Keeper.setTransferWithCache(suite.SdkCtx, transfer)
	suite.NoError(err)

	// Update the transfer
	transfer.Status = types.TransferStatus_CONFIRMED
	transfer.Amount = sdkmath.NewInt(2000)
	err = suite.Keeper.setTransferWithCache(suite.SdkCtx, transfer)
	suite.NoError(err)

	// Verify the update
	retrieved, found := suite.Keeper.getTransferWithCache(suite.SdkCtx, "update-cache-internal-test")
	suite.True(found)
	suite.NotNil(retrieved)
	suite.Equal(types.TransferStatus_CONFIRMED, retrieved.Status)
	suite.Equal(sdkmath.NewInt(2000), retrieved.Amount)
}

// =============================================================================
// deleteTransferWithCache Tests (internal function)
// =============================================================================

func (suite *InternalCacheTestSuite) TestDeleteTransferWithCache_EmptyID() {
	// Should not panic with empty ID
	suite.Keeper.deleteTransferWithCache(suite.SdkCtx, "")
}

func (suite *InternalCacheTestSuite) TestDeleteTransferWithCache_Nonexistent() {
	// Should not panic with nonexistent ID
	suite.Keeper.deleteTransferWithCache(suite.SdkCtx, "nonexistent-delete")
}

func (suite *InternalCacheTestSuite) TestDeleteTransferWithCache_Valid() {
	// Create a transfer
	transfer := &types.CrossChainTransfer{
		TransferId:  "delete-cache-internal-test",
		SourceChain: "aura",
		TargetChain: "ethereum",
		Sender:      sdk.AccAddress("sender____________").String(),
		Recipient:   "0x222",
		Amount:      sdkmath.NewInt(500),
		Denom:       "uaura",
		Status:      types.TransferStatus_PENDING,
	}
	err := suite.Keeper.setTransferWithCache(suite.SdkCtx, transfer)
	suite.NoError(err)

	// Verify it exists
	_, found := suite.Keeper.getTransferWithCache(suite.SdkCtx, "delete-cache-internal-test")
	suite.True(found)

	// Delete it
	suite.Keeper.deleteTransferWithCache(suite.SdkCtx, "delete-cache-internal-test")

	// Verify it's gone
	_, found = suite.Keeper.getTransferWithCache(suite.SdkCtx, "delete-cache-internal-test")
	suite.False(found)
}

// =============================================================================
// initTransferCache Tests (internal function)
// =============================================================================

func (suite *InternalCacheTestSuite) TestInitTransferCache_DefaultSize() {
	err := suite.Keeper.initTransferCache(0) // 0 should use default
	suite.NoError(err)
}

func (suite *InternalCacheTestSuite) TestInitTransferCache_NegativeSize() {
	err := suite.Keeper.initTransferCache(-1) // negative should use default
	suite.NoError(err)
}

func (suite *InternalCacheTestSuite) TestInitTransferCache_CustomSize() {
	err := suite.Keeper.initTransferCache(100)
	suite.NoError(err)
}

func (suite *InternalCacheTestSuite) TestInitTransferCache_LargeSize() {
	err := suite.Keeper.initTransferCache(10000)
	suite.NoError(err)
}

func (suite *InternalCacheTestSuite) TestInitTransferCache_Reinitialize() {
	// Initialize with one size
	err := suite.Keeper.initTransferCache(50)
	suite.NoError(err)

	// Re-initialize with different size
	err = suite.Keeper.initTransferCache(200)
	suite.NoError(err)
}
