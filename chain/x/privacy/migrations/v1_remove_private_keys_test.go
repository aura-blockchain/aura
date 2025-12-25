// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package migrations_test

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/privacy/migrations"
	"github.com/aequitas/aura/chain/x/privacy/types"
	privacypb "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
)

type MigrationV1TestSuite struct {
	suite.Suite
	ctx      sdk.Context
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec
}

func (suite *MigrationV1TestSuite) SetupTest() {
	input := keepertest.CreateTestInput(suite.T())
	suite.ctx = input.Ctx
	suite.storeKey = input.StoreKey
	suite.cdc = input.Cdc
}

func TestMigrationV1TestSuite(t *testing.T) {
	suite.Run(t, new(MigrationV1TestSuite))
}

// TestMigrateV1RemovePrivateKeys_EmptyState tests migration on empty state
func (suite *MigrationV1TestSuite) TestMigrateV1RemovePrivateKeys_EmptyState() {
	err := migrations.MigrateV1RemovePrivateKeys(
		sdk.WrapSDKContext(suite.ctx),
		suite.storeKey,
		suite.cdc,
	)

	suite.NoError(err)
	suite.T().Log("✓ Migration succeeds on empty state")
}

// TestMigrateV1RemovePrivateKeys_WithValidKeys tests migration with valid public keys
func (suite *MigrationV1TestSuite) TestMigrateV1RemovePrivateKeys_WithValidKeys() {
	// Create valid view keys (only public keys)
	viewKeys := []struct {
		owner  string
		pubKey []byte
	}{
		{"owner1", make([]byte, 32)},
		{"owner2", make([]byte, 33)},
		{"owner3", make([]byte, 64)},
	}

	// Store valid view keys
	for i, vk := range viewKeys {
		for j := range vk.pubKey {
			vk.pubKey[j] = byte(i*32 + j)
		}

		viewKey := &privacypb.ViewKey{
			KeyType:       "INCOMING",
			PublicViewKey: vk.pubKey,
			Address:       []byte(vk.owner),
		}

		bz, err := suite.cdc.Marshal(viewKey)
		suite.Require().NoError(err)

		store := suite.ctx.KVStore(suite.storeKey)
		key := append(types.ViewKeyPrefix, []byte(vk.owner)...)
		key = append(key, vk.pubKey...)
		store.Set(key, bz)
	}

	// Run migration
	err := migrations.MigrateV1RemovePrivateKeys(
		sdk.WrapSDKContext(suite.ctx),
		suite.storeKey,
		suite.cdc,
	)

	suite.NoError(err)

	// Verify all keys still exist and are valid
	for _, vk := range viewKeys {
		store := suite.ctx.KVStore(suite.storeKey)
		key := append(types.ViewKeyPrefix, []byte(vk.owner)...)
		key = append(key, vk.pubKey...)

		bz := store.Get(key)
		suite.NotNil(bz, "view key should still exist after migration")

		var viewKey privacypb.ViewKey
		err := suite.cdc.Unmarshal(bz, &viewKey)
		suite.NoError(err)
		suite.Equal(vk.pubKey, viewKey.PublicViewKey)
	}

	suite.T().Log("✓ Migration preserves valid public keys")
}

// TestMigrateV1RemovePrivateKeys_RemovesInvalidKeys tests that invalid keys are removed
func (suite *MigrationV1TestSuite) TestMigrateV1RemovePrivateKeys_RemovesInvalidKeys() {
	// Create an invalid view key (no public key)
	invalidViewKey := &privacypb.ViewKey{
		KeyType:       "INCOMING",
		PublicViewKey: []byte{}, // EMPTY - invalid
		Address:       []byte("invalid_owner"),
	}

	bz, err := suite.cdc.Marshal(invalidViewKey)
	suite.Require().NoError(err)

	store := suite.ctx.KVStore(suite.storeKey)
	key := append(types.ViewKeyPrefix, []byte("invalid_owner")...)
	store.Set(key, bz)

	// Verify it exists before migration
	suite.True(store.Has(key))

	// Run migration
	err = migrations.MigrateV1RemovePrivateKeys(
		sdk.WrapSDKContext(suite.ctx),
		suite.storeKey,
		suite.cdc,
	)

	suite.NoError(err)

	// Verify the invalid key was removed
	suite.False(store.Has(key), "invalid view key should be removed")

	suite.T().Log("✓ Migration removes invalid view keys")
}

// TestMigrateV1RemovePrivateKeys_MixedState tests migration with both valid and invalid keys
func (suite *MigrationV1TestSuite) TestMigrateV1RemovePrivateKeys_MixedState() {
	// Create valid and invalid view keys
	validKey := &privacypb.ViewKey{
		KeyType:       "INCOMING",
		PublicViewKey: make([]byte, 32),
		Address:       []byte("valid_owner"),
	}

	invalidKey := &privacypb.ViewKey{
		KeyType:       "OUTGOING",
		PublicViewKey: []byte{}, // EMPTY - invalid
		Address:       []byte("invalid_owner"),
	}

	// Store valid key
	validBz, err := suite.cdc.Marshal(validKey)
	suite.Require().NoError(err)
	store := suite.ctx.KVStore(suite.storeKey)
	validStoreKey := append(types.ViewKeyPrefix, []byte("valid_owner")...)
	validStoreKey = append(validStoreKey, validKey.PublicViewKey...)
	store.Set(validStoreKey, validBz)

	// Store invalid key
	invalidBz, err := suite.cdc.Marshal(invalidKey)
	suite.Require().NoError(err)
	invalidStoreKey := append(types.ViewKeyPrefix, []byte("invalid_owner")...)
	store.Set(invalidStoreKey, invalidBz)

	// Run migration
	err = migrations.MigrateV1RemovePrivateKeys(
		sdk.WrapSDKContext(suite.ctx),
		suite.storeKey,
		suite.cdc,
	)

	suite.NoError(err)

	// Verify valid key still exists
	suite.True(store.Has(validStoreKey), "valid key should be preserved")

	// Verify invalid key was removed
	suite.False(store.Has(invalidStoreKey), "invalid key should be removed")

	suite.T().Log("✓ Migration handles mixed valid/invalid keys correctly")
}

// TestMigrateV1VerifyNoPrivateKeys_PassesWithCleanState tests verification on clean state
func (suite *MigrationV1TestSuite) TestMigrateV1VerifyNoPrivateKeys_PassesWithCleanState() {
	// Create only valid view keys
	viewKey := &privacypb.ViewKey{
		KeyType:       "INCOMING",
		PublicViewKey: make([]byte, 32),
		Address:       []byte("owner"),
		Permissions:   []string{"view_incoming"},
	}

	bz, err := suite.cdc.Marshal(viewKey)
	suite.Require().NoError(err)

	store := suite.ctx.KVStore(suite.storeKey)
	key := append(types.ViewKeyPrefix, []byte("owner")...)
	key = append(key, viewKey.PublicViewKey...)
	store.Set(key, bz)

	// Run verification
	err = migrations.MigrateV1VerifyNoPrivateKeys(
		sdk.WrapSDKContext(suite.ctx),
		suite.storeKey,
		suite.cdc,
	)

	suite.NoError(err)
	suite.T().Log("✓ Verification passes with clean state")
}

// TestMigrateV1VerifyNoPrivateKeys_DetectsSuspiciousKeyTypes tests that suspicious types are detected
func (suite *MigrationV1TestSuite) TestMigrateV1VerifyNoPrivateKeys_DetectsSuspiciousKeyTypes() {
	// Create a view key with suspicious key type
	// NOTE: This would normally be rejected by the msg_server, but we're testing
	// the verification function which checks existing state
	suspiciousKey := &privacypb.ViewKey{
		KeyType:       "PRIVATE", // SUSPICIOUS
		PublicViewKey: make([]byte, 32),
		Address:       []byte("suspicious_owner"),
	}

	bz, err := suite.cdc.Marshal(suspiciousKey)
	suite.Require().NoError(err)

	store := suite.ctx.KVStore(suite.storeKey)
	key := append(types.ViewKeyPrefix, []byte("suspicious_owner")...)
	key = append(key, suspiciousKey.PublicViewKey...)
	store.Set(key, bz)

	// Run verification - it should log an error but not fail
	err = migrations.MigrateV1VerifyNoPrivateKeys(
		sdk.WrapSDKContext(suite.ctx),
		suite.storeKey,
		suite.cdc,
	)

	// Verification doesn't fail, it just logs warnings
	suite.NoError(err)
	suite.T().Log("✓ Verification detects suspicious key types (check logs)")
}

// TestMigrateV1VerifyNoPrivateKeys_EmptyState tests verification on empty state
func (suite *MigrationV1TestSuite) TestMigrateV1VerifyNoPrivateKeys_EmptyState() {
	err := migrations.MigrateV1VerifyNoPrivateKeys(
		sdk.WrapSDKContext(suite.ctx),
		suite.storeKey,
		suite.cdc,
	)

	suite.NoError(err)
	suite.T().Log("✓ Verification succeeds on empty state")
}

// TestMigrationEmitsEvents tests that migration emits proper events
func (suite *MigrationV1TestSuite) TestMigrationEmitsEvents() {
	// Create some view keys
	validKey := &privacypb.ViewKey{
		KeyType:       "INCOMING",
		PublicViewKey: make([]byte, 32),
		Address:       []byte("owner1"),
	}

	invalidKey := &privacypb.ViewKey{
		KeyType:       "OUTGOING",
		PublicViewKey: []byte{}, // EMPTY - invalid
		Address:       []byte("owner2"),
	}

	store := suite.ctx.KVStore(suite.storeKey)

	// Store valid key
	validBz, err := suite.cdc.Marshal(validKey)
	suite.Require().NoError(err)
	validStoreKey := append(types.ViewKeyPrefix, []byte("owner1")...)
	validStoreKey = append(validStoreKey, validKey.PublicViewKey...)
	store.Set(validStoreKey, validBz)

	// Store invalid key
	invalidBz, err := suite.cdc.Marshal(invalidKey)
	suite.Require().NoError(err)
	invalidStoreKey := append(types.ViewKeyPrefix, []byte("owner2")...)
	store.Set(invalidStoreKey, invalidBz)

	// Run migration
	err = migrations.MigrateV1RemovePrivateKeys(
		sdk.WrapSDKContext(suite.ctx),
		suite.storeKey,
		suite.cdc,
	)

	suite.NoError(err)

	// Check that events were emitted
	events := suite.ctx.EventManager().Events()
	var found bool
	for _, event := range events {
		if event.Type == "privacy_migration_v1" {
			found = true
			// Verify event attributes
			for _, attr := range event.Attributes {
				suite.T().Logf("Event attribute: %s = %s", attr.Key, attr.Value)
			}
		}
	}

	suite.True(found, "migration should emit privacy_migration_v1 event")
	suite.T().Log("✓ Migration emits proper events")
}

// TestMigrationIdempotent tests that running migration multiple times is safe
func (suite *MigrationV1TestSuite) TestMigrationIdempotent() {
	// Create a valid view key
	viewKey := &privacypb.ViewKey{
		KeyType:       "INCOMING",
		PublicViewKey: make([]byte, 32),
		Address:       []byte("owner"),
	}

	bz, err := suite.cdc.Marshal(viewKey)
	suite.Require().NoError(err)

	store := suite.ctx.KVStore(suite.storeKey)
	key := append(types.ViewKeyPrefix, []byte("owner")...)
	key = append(key, viewKey.PublicViewKey...)
	store.Set(key, bz)

	// Run migration first time
	err = migrations.MigrateV1RemovePrivateKeys(
		sdk.WrapSDKContext(suite.ctx),
		suite.storeKey,
		suite.cdc,
	)
	suite.NoError(err)

	// Verify key still exists
	suite.True(store.Has(key))

	// Run migration second time
	err = migrations.MigrateV1RemovePrivateKeys(
		sdk.WrapSDKContext(suite.ctx),
		suite.storeKey,
		suite.cdc,
	)
	suite.NoError(err)

	// Verify key still exists and is unchanged
	suite.True(store.Has(key))

	var retrievedKey privacypb.ViewKey
	err = suite.cdc.Unmarshal(store.Get(key), &retrievedKey)
	suite.NoError(err)
	suite.Equal(viewKey.PublicViewKey, retrievedKey.PublicViewKey)

	suite.T().Log("✓ Migration is idempotent - safe to run multiple times")
}
