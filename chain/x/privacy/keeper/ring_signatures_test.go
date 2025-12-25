// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/privacy/keeper"
	"github.com/aequitas/aura/chain/x/privacy/types"
)

type RingSignaturesTestSuite struct {
	suite.Suite

	keeper *keeper.Keeper
	ctx    sdk.Context
}

func TestRingSignaturesTestSuite(t *testing.T) {
	suite.Run(t, new(RingSignaturesTestSuite))
}

func (suite *RingSignaturesTestSuite) SetupTest() {
	input := keepertest.CreateTestInput(suite.T())
	suite.keeper = keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil,
		nil,
	)
	suite.ctx = input.Ctx

	// Enable ring signatures and set params
	params := types.DefaultParams()
	params.EnableRingSignatures = true
	params.MinRingSize = 3
	params.MaxRingSize = 11
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)
}

// Helper to create valid ring signature
func (suite *RingSignaturesTestSuite) createValidRingSignature(ringSize int) *keeper.RingSignature {
	publicKeys := make([][]byte, ringSize)
	for i := 0; i < ringSize; i++ {
		publicKeys[i] = []byte("public_key_" + string(rune('0'+i)))
	}

	return &keeper.RingSignature{
		PublicKeys: publicKeys,
		Signature:  []byte("valid_signature"),
		KeyImage:   []byte("unique_key_image"),
		Message:    []byte("test_message"),
	}
}

// VerifyRingSignature Tests

func (suite *RingSignaturesTestSuite) TestVerifyRingSignature_NotEnabled() {
	// Disable ring signatures
	params := suite.keeper.GetParams(suite.ctx)
	params.EnableRingSignatures = false
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	signature := suite.createValidRingSignature(5)

	valid, err := suite.keeper.VerifyRingSignature(suite.ctx, signature)

	suite.Require().Error(err)
	suite.Require().False(valid)
	suite.Require().Contains(err.Error(), "ring signatures not enabled")
}

func (suite *RingSignaturesTestSuite) TestVerifyRingSignature_RingSizeTooSmall() {
	signature := suite.createValidRingSignature(2) // Less than min 3

	valid, err := suite.keeper.VerifyRingSignature(suite.ctx, signature)

	suite.Require().Error(err)
	suite.Require().False(valid)
	suite.Require().Contains(err.Error(), "below minimum")
}

func (suite *RingSignaturesTestSuite) TestVerifyRingSignature_RingSizeTooLarge() {
	signature := suite.createValidRingSignature(12) // More than max 11

	valid, err := suite.keeper.VerifyRingSignature(suite.ctx, signature)

	suite.Require().Error(err)
	suite.Require().False(valid)
	suite.Require().Contains(err.Error(), "exceeds maximum")
}

func (suite *RingSignaturesTestSuite) TestVerifyRingSignature_KeyImageAlreadyUsed() {
	signature := suite.createValidRingSignature(5)

	// Use the signature once
	valid, err := suite.keeper.VerifyRingSignature(suite.ctx, signature)
	suite.Require().NoError(err)
	suite.Require().True(valid)

	// Try to use the same key image again
	valid, err = suite.keeper.VerifyRingSignature(suite.ctx, signature)

	suite.Require().Error(err)
	suite.Require().False(valid)
	suite.Require().Equal(types.ErrKeyImageAlreadyUsed, err)
}

func (suite *RingSignaturesTestSuite) TestVerifyRingSignature_InvalidSignature() {
	signature := suite.createValidRingSignature(5)
	signature.Signature = []byte("invalid") // Invalid signature marker

	valid, err := suite.keeper.VerifyRingSignature(suite.ctx, signature)

	suite.Require().Error(err)
	suite.Require().False(valid)
	suite.Require().Contains(err.Error(), "invalid ring signature")
}

func (suite *RingSignaturesTestSuite) TestVerifyRingSignature_EmptySignature() {
	signature := suite.createValidRingSignature(5)
	signature.Signature = []byte{} // Empty signature

	valid, err := suite.keeper.VerifyRingSignature(suite.ctx, signature)

	suite.Require().Error(err)
	suite.Require().False(valid)
}

func (suite *RingSignaturesTestSuite) TestVerifyRingSignature_EmptyKeyImage() {
	signature := suite.createValidRingSignature(5)
	signature.KeyImage = []byte{} // Empty key image

	valid, err := suite.keeper.VerifyRingSignature(suite.ctx, signature)

	suite.Require().Error(err)
	suite.Require().False(valid)
}

func (suite *RingSignaturesTestSuite) TestVerifyRingSignature_EmptyPublicKeys() {
	signature := suite.createValidRingSignature(5)
	signature.PublicKeys = [][]byte{} // Empty public keys

	valid, err := suite.keeper.VerifyRingSignature(suite.ctx, signature)

	suite.Require().Error(err)
	suite.Require().False(valid)
}

func (suite *RingSignaturesTestSuite) TestVerifyRingSignature_Success() {
	signature := suite.createValidRingSignature(5)

	valid, err := suite.keeper.VerifyRingSignature(suite.ctx, signature)

	suite.Require().NoError(err)
	suite.Require().True(valid)

	// Verify key image was stored
	exists := suite.keeper.KeyImageExists(suite.ctx, signature.KeyImage)
	suite.Require().True(exists)
}

// KeyImageExists Tests

func TestKeyImageExists(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	keyImage := []byte("test_key_image")

	// Should not exist initially
	exists := k.KeyImageExists(input.Ctx, keyImage)
	require.False(t, exists)

	// Store it
	err := k.StoreKeyImage(input.Ctx, keyImage)
	require.NoError(t, err)

	// Should exist now
	exists = k.KeyImageExists(input.Ctx, keyImage)
	require.True(t, exists)
}

// StoreKeyImage Tests

func TestStoreKeyImage(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	keyImage := []byte("test_key_image")

	// First storage should succeed
	err := k.StoreKeyImage(input.Ctx, keyImage)
	require.NoError(t, err)

	// Second storage should fail (already exists)
	err = k.StoreKeyImage(input.Ctx, keyImage)
	require.Error(t, err)
	require.Equal(t, types.ErrKeyImageAlreadyUsed, err)
}

// GenerateRingSignature Tests

func TestGenerateRingSignature(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	message := []byte("test_message")
	publicKeys := [][]byte{
		[]byte("public_key_0"),
		[]byte("public_key_1"),
		[]byte("public_key_2"),
	}
	secretIndex := 1

	signature, err := k.GenerateRingSignature(input.Ctx, message, publicKeys, secretIndex)

	require.NoError(t, err)
	require.NotNil(t, signature)
	require.Equal(t, publicKeys, signature.PublicKeys)
	require.NotEmpty(t, signature.Signature)
	require.NotEmpty(t, signature.KeyImage)
	require.Equal(t, message, signature.Message)
}

// VerifyLinkableRingSignature Tests

func (suite *RingSignaturesTestSuite) TestVerifyLinkableRingSignature_KeyImageAlreadyUsed() {
	signature := suite.createValidRingSignature(5)

	// Store the key image first
	err := suite.keeper.StoreKeyImage(suite.ctx, signature.KeyImage)
	suite.Require().NoError(err)

	// Try to verify linkable signature
	valid, err := suite.keeper.VerifyLinkableRingSignature(suite.ctx, signature)

	suite.Require().Error(err)
	suite.Require().False(valid)
	suite.Require().Equal(types.ErrKeyImageAlreadyUsed, err)
}

func (suite *RingSignaturesTestSuite) TestVerifyLinkableRingSignature_Success() {
	signature := suite.createValidRingSignature(5)

	valid, err := suite.keeper.VerifyLinkableRingSignature(suite.ctx, signature)

	suite.Require().NoError(err)
	suite.Require().True(valid)
}

// GetRingMembers Tests

func (suite *RingSignaturesTestSuite) TestGetRingMembers_SizeTooSmall() {
	members, err := suite.keeper.GetRingMembers(suite.ctx, 2) // Less than min 3

	suite.Require().Error(err)
	suite.Require().Nil(members)
	suite.Require().Contains(err.Error(), "invalid ring size")
}

func (suite *RingSignaturesTestSuite) TestGetRingMembers_SizeTooLarge() {
	members, err := suite.keeper.GetRingMembers(suite.ctx, 12) // More than max 11

	suite.Require().Error(err)
	suite.Require().Nil(members)
	suite.Require().Contains(err.Error(), "invalid ring size")
}

func (suite *RingSignaturesTestSuite) TestGetRingMembers_Success() {
	ringSize := 5

	members, err := suite.keeper.GetRingMembers(suite.ctx, ringSize)

	suite.Require().NoError(err)
	suite.Require().NotNil(members)
	suite.Require().Len(members, ringSize)
}

// AddRingMember Tests

func TestAddRingMember(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	publicKey := []byte("test_public_key")

	err := k.AddRingMember(input.Ctx, publicKey)
	require.NoError(t, err)
}

// RemoveRingMember Tests

func TestRemoveRingMember(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	publicKey := []byte("test_public_key")

	// Add first
	err := k.AddRingMember(input.Ctx, publicKey)
	require.NoError(t, err)

	// Remove
	err = k.RemoveRingMember(input.Ctx, publicKey)
	require.NoError(t, err)
}

// Standalone integration tests

func TestRingSignatureFlow_Complete(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	// Setup params
	params := types.DefaultParams()
	params.EnableRingSignatures = true
	params.MinRingSize = 3
	params.MaxRingSize = 11
	err := k.SetParams(input.Ctx, params)
	require.NoError(t, err)

	// Generate signature
	message := []byte("test_message")
	publicKeys := [][]byte{
		[]byte("public_key_0"),
		[]byte("public_key_1"),
		[]byte("public_key_2"),
	}

	signature, err := k.GenerateRingSignature(input.Ctx, message, publicKeys, 1)
	require.NoError(t, err)

	// Verify signature
	valid, err := k.VerifyRingSignature(input.Ctx, signature)
	require.NoError(t, err)
	require.True(t, valid)

	// Try to verify again (should fail due to key image reuse)
	valid, err = k.VerifyRingSignature(input.Ctx, signature)
	require.Error(t, err)
	require.False(t, valid)
	require.Equal(t, types.ErrKeyImageAlreadyUsed, err)
}

func TestRingSignature_DifferentRingSizes(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	params := types.DefaultParams()
	params.EnableRingSignatures = true
	params.MinRingSize = 3
	params.MaxRingSize = 11
	err := k.SetParams(input.Ctx, params)
	require.NoError(t, err)

	testCases := []struct {
		ringSize  int
		shouldErr bool
	}{
		{2, true},   // Below min
		{3, false},  // Min
		{5, false},  // Middle
		{11, false}, // Max
		{12, true},  // Above max
	}

	for _, tc := range testCases {
		t.Run(string(rune('0'+tc.ringSize)), func(t *testing.T) {
			publicKeys := make([][]byte, tc.ringSize)
			for i := 0; i < tc.ringSize; i++ {
				publicKeys[i] = []byte("public_key_" + string(rune('0'+i)) + "_size_" + string(rune('0'+tc.ringSize)))
			}

			signature := &keeper.RingSignature{
				PublicKeys: publicKeys,
				Signature:  []byte("valid_signature"),
				KeyImage:   []byte("key_image_size_" + string(rune('0'+tc.ringSize))),
				Message:    []byte("message"),
			}

			valid, err := k.VerifyRingSignature(input.Ctx, signature)

			if tc.shouldErr {
				require.Error(t, err)
				require.False(t, valid)
			} else {
				require.NoError(t, err)
				require.True(t, valid)
			}
		})
	}
}
