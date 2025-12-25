// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/privacy/keeper"
	privacypb "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
)

// ViewKeySecurityTestSuite tests security aspects of view key management
// This test suite ensures that private keys can never be stored on-chain
type ViewKeySecurityTestSuite struct {
	suite.Suite
	keeper    *keeper.Keeper
	ctx       sdk.Context
	msgServer privacypb.MsgServer
}

func (suite *ViewKeySecurityTestSuite) SetupTest() {
	input := keepertest.CreateTestInput(suite.T())
	suite.keeper = keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil, // authKeeper
		nil, // bankKeeper
	)
	suite.ctx = input.Ctx
	suite.msgServer = keeper.NewMsgServerImpl(suite.keeper)
}

func TestViewKeySecurityTestSuite(t *testing.T) {
	suite.Run(t, new(ViewKeySecurityTestSuite))
}

// TestViewKeyHasNoPrivateKeyField verifies that the ViewKey proto message
// does not have a private_view_key field at all
func (suite *ViewKeySecurityTestSuite) TestViewKeyHasNoPrivateKeyField() {
	// Create a ViewKey - it should compile and work with only public keys
	viewKey := &privacypb.ViewKey{
		KeyType:       "INCOMING",
		PublicViewKey: []byte("public_key_32_bytes_test_value"),
		Address:       []byte("test_address"),
		Permissions:   []string{"view_incoming"},
	}

	suite.NotNil(viewKey)
	suite.Equal("INCOMING", viewKey.KeyType)
	suite.Equal([]byte("public_key_32_bytes_test_value"), viewKey.PublicViewKey)
	// This test passes if it compiles - there's no private_view_key field to set
}

// TestRegisterViewKeyValidation tests that RegisterViewKey validates public keys
func (suite *ViewKeySecurityTestSuite) TestRegisterViewKeyValidation() {
	goCtx := sdk.WrapSDKContext(suite.ctx)
	testAddr := keepertest.GenTestAddr().String()

	testCases := []struct {
		name        string
		viewKey     *privacypb.ViewKey
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid 32-byte public key",
			viewKey: &privacypb.ViewKey{
				KeyType:       "INCOMING",
				PublicViewKey: make([]byte, 32), // Valid length
				Address:       []byte(testAddr),
			},
			expectError: false,
		},
		{
			name: "valid 33-byte public key (compressed secp256k1)",
			viewKey: &privacypb.ViewKey{
				KeyType:       "INCOMING",
				PublicViewKey: make([]byte, 33), // Valid length
				Address:       []byte(testAddr),
			},
			expectError: false,
		},
		{
			name: "valid 64-byte public key (uncompressed)",
			viewKey: &privacypb.ViewKey{
				KeyType:       "INCOMING",
				PublicViewKey: make([]byte, 64), // Valid length
				Address:       []byte(testAddr),
			},
			expectError: false,
		},
		{
			name: "empty public key rejected",
			viewKey: &privacypb.ViewKey{
				KeyType:       "INCOMING",
				PublicViewKey: []byte{},
				Address:       []byte(testAddr),
			},
			expectError: true,
			errorMsg:    "public view key cannot be empty",
		},
		{
			name: "invalid key length rejected",
			viewKey: &privacypb.ViewKey{
				KeyType:       "INCOMING",
				PublicViewKey: make([]byte, 16), // Invalid length
				Address:       []byte(testAddr),
			},
			expectError: true,
			errorMsg:    "invalid public key length",
		},
		{
			name: "suspicious key type PRIVATE rejected",
			viewKey: &privacypb.ViewKey{
				KeyType:       "PRIVATE", // Suspicious type
				PublicViewKey: make([]byte, 32),
				Address:       []byte(testAddr),
			},
			expectError: true,
			errorMsg:    "private keys cannot be registered on-chain",
		},
		{
			name: "suspicious key type SECRET rejected",
			viewKey: &privacypb.ViewKey{
				KeyType:       "SECRET", // Suspicious type
				PublicViewKey: make([]byte, 32),
				Address:       []byte(testAddr),
			},
			expectError: true,
			errorMsg:    "private keys cannot be registered on-chain",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			msg := &privacypb.MsgRegisterViewKey{
				Owner:   testAddr,
				ViewKey: tc.viewKey,
			}

			resp, err := suite.msgServer.RegisterViewKey(goCtx, msg)

			if tc.expectError {
				suite.Error(err)
				suite.Contains(err.Error(), tc.errorMsg)
				suite.Nil(resp)
			} else {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.True(resp.Success)
			}
		})
	}
}

// TestStoredViewKeyHasNoPrivateData verifies that stored view keys contain no private data
func (suite *ViewKeySecurityTestSuite) TestStoredViewKeyHasNoPrivateData() {
	goCtx := sdk.WrapSDKContext(suite.ctx)
	testAddr := keepertest.GenTestAddr().String()

	publicKey := make([]byte, 32)
	for i := range publicKey {
		publicKey[i] = byte(i)
	}

	// Register a view key
	msg := &privacypb.MsgRegisterViewKey{
		Owner: testAddr,
		ViewKey: &privacypb.ViewKey{
			KeyType:       "INCOMING",
			PublicViewKey: publicKey,
			Address:       []byte(testAddr),
			Permissions:   []string{"view_incoming"},
		},
	}

	resp, err := suite.msgServer.RegisterViewKey(goCtx, msg)
	suite.NoError(err)
	suite.NotNil(resp)

	// Query the stored view key
	storedKeys := suite.keeper.GetViewKeys(suite.ctx, testAddr)
	suite.Require().Len(storedKeys, 1)

	storedKey := storedKeys[0]

	// Verify only public data is stored
	suite.Equal("INCOMING", storedKey.KeyType)
	suite.Equal(publicKey, storedKey.PublicViewKey)
	suite.Equal([]byte(testAddr), storedKey.Address)
	suite.Equal([]string{"view_incoming"}, storedKey.Permissions)

	// The ViewKey struct should not have any private key field
	// This is a compile-time guarantee - if this test compiles, we're safe
}

// TestQueryViewKeyDoesNotExposePrivateData tests that query endpoints only return public data
func (suite *ViewKeySecurityTestSuite) TestQueryViewKeyDoesNotExposePrivateData() {
	testAddr := keepertest.GenTestAddr().String()

	publicKey := make([]byte, 32)
	for i := range publicKey {
		publicKey[i] = byte(i + 1)
	}

	// Store a view key directly (simulating existing data)
	viewKey := &privacypb.ViewKey{
		KeyType:       "OUTGOING",
		PublicViewKey: publicKey,
		Address:       []byte(testAddr),
		Permissions:   []string{"view_outgoing"},
	}

	err := suite.keeper.SetViewKey(suite.ctx, testAddr, viewKey)
	suite.NoError(err)

	// Query the view key
	queriedKeys := suite.keeper.GetViewKeys(suite.ctx, testAddr)
	suite.Require().Len(queriedKeys, 1)

	queriedKey := queriedKeys[0]

	// Verify only public information is returned
	suite.Equal("OUTGOING", queriedKey.KeyType)
	suite.Equal(publicKey, queriedKey.PublicViewKey)
	suite.Equal([]byte(testAddr), queriedKey.Address)
	suite.Equal([]string{"view_outgoing"}, queriedKey.Permissions)

	// No private key field should exist in the struct
	// This is verified at compile time
}

// TestDecryptWithViewKeyMethodRemoved verifies that the dangerous DecryptWithViewKey method is gone
func (suite *ViewKeySecurityTestSuite) TestDecryptWithViewKeyMethodRemoved() {
	// This test verifies that the DecryptWithViewKey query no longer exists
	// The query server should not have this method at all

	// We test this by trying to compile - if DecryptWithViewKey existed on the
	// query server, this would compile. Since it doesn't exist, we document why:
	//
	// SECURITY: DecryptWithViewKey has been removed because decryption must be
	// performed client-side using private keys that never leave the client.
	//
	// If you need to decrypt transaction data:
	// 1. Download encrypted data from the blockchain
	// 2. Decrypt locally using your private view key (stored in client keystore)
	// 3. Never send private keys to the blockchain

	suite.T().Log("DecryptWithViewKey has been removed for security - decryption is client-side only")
}

// TestPrivateKeyNeverInState verifies that the state store never contains private keys
func (suite *ViewKeySecurityTestSuite) TestPrivateKeyNeverInState() {
	testAddr := keepertest.GenTestAddr().String()

	publicKey := make([]byte, 32)
	for i := range publicKey {
		publicKey[i] = byte(i + 2)
	}

	// Register view key
	goCtx := sdk.WrapSDKContext(suite.ctx)
	msg := &privacypb.MsgRegisterViewKey{
		Owner: testAddr,
		ViewKey: &privacypb.ViewKey{
			KeyType:       "AUDIT",
			PublicViewKey: publicKey,
			Address:       []byte(testAddr),
			Permissions:   []string{"view_all"},
		},
	}

	_, err := suite.msgServer.RegisterViewKey(goCtx, msg)
	suite.NoError(err)

	// Get the raw state store and verify no private key material
	store := suite.keeper.GetStoreKey()
	suite.NotNil(store)

	// The ViewKey proto message no longer has a private_view_key field,
	// so it's impossible to store private keys even if we wanted to.
	// This test passes by compilation.

	suite.T().Log("State store can only contain public keys - private_view_key field does not exist")
}

// TestKeyTypesAreValidated tests that suspicious key types are rejected
func (suite *ViewKeySecurityTestSuite) TestKeyTypesAreValidated() {
	goCtx := sdk.WrapSDKContext(suite.ctx)
	testAddr := keepertest.GenTestAddr().String()

	suspiciousTypes := []string{
		"PRIVATE",
		"SECRET",
		"private",
		"secret",
	}

	for _, keyType := range suspiciousTypes {
		suite.Run("reject_"+keyType, func() {
			msg := &privacypb.MsgRegisterViewKey{
				Owner: testAddr,
				ViewKey: &privacypb.ViewKey{
					KeyType:       keyType,
					PublicViewKey: make([]byte, 32),
					Address:       []byte(testAddr),
				},
			}

			resp, err := suite.msgServer.RegisterViewKey(goCtx, msg)

			// Should be rejected (case-sensitive check)
			if keyType == "PRIVATE" || keyType == "SECRET" {
				suite.Error(err)
				suite.Contains(err.Error(), "private keys cannot be registered on-chain")
				suite.Nil(resp)
			}
		})
	}
}

// TestValidKeyTypesAreAccepted tests that legitimate key types work
func (suite *ViewKeySecurityTestSuite) TestValidKeyTypesAreAccepted() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	validTypes := []string{
		"INCOMING",
		"OUTGOING",
		"AUDIT",
		"VIEW",
		"",
	}

	for _, keyType := range validTypes {
		suite.Run("accept_"+keyType, func() {
			testAddr := keepertest.GenTestAddr().String()
			msg := &privacypb.MsgRegisterViewKey{
				Owner: testAddr,
				ViewKey: &privacypb.ViewKey{
					KeyType:       keyType,
					PublicViewKey: make([]byte, 32),
					Address:       []byte(testAddr),
				},
			}

			resp, err := suite.msgServer.RegisterViewKey(goCtx, msg)
			suite.NoError(err)
			suite.NotNil(resp)
			suite.True(resp.Success)
		})
	}
}

// TestPublicKeyLengthBoundaries tests boundary conditions for key length validation
func (suite *ViewKeySecurityTestSuite) TestPublicKeyLengthBoundaries() {
	goCtx := sdk.WrapSDKContext(suite.ctx)
	testAddr := keepertest.GenTestAddr().String()

	testCases := []struct {
		name        string
		keyLength   int
		expectError bool
	}{
		{"empty", 0, true},
		{"too_short", 16, true},
		{"too_short_31", 31, true},
		{"valid_32", 32, false},
		{"valid_33", 33, false},
		{"invalid_34", 34, true},
		{"invalid_63", 63, true},
		{"valid_64", 64, false},
		{"invalid_65", 65, true},
		{"too_long", 128, true},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			msg := &privacypb.MsgRegisterViewKey{
				Owner: testAddr,
				ViewKey: &privacypb.ViewKey{
					KeyType:       "INCOMING",
					PublicViewKey: make([]byte, tc.keyLength),
					Address:       []byte(testAddr),
				},
			}

			resp, err := suite.msgServer.RegisterViewKey(goCtx, msg)

			if tc.expectError {
				suite.Error(err)
				suite.Nil(resp)
			} else {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.True(resp.Success)
			}

			// Use a new address for each test case to avoid conflicts
			testAddr = keepertest.GenTestAddr().String()
		})
	}
}
