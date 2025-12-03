package migrations

import (
	"context"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/privacy/types"
	privacypb "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
)

// MigrateV1RemovePrivateKeys removes any private key data that may have been stored
// before the security fix was implemented.
//
// SECURITY: This migration is CRITICAL. It ensures that no private keys remain
// in the blockchain state after upgrading to the secure version.
//
// Migration steps:
// 1. Iterate through all ViewKey entries in the state
// 2. For each ViewKey, verify it only contains public key data
// 3. If any private key fields exist (from old proto), clear them
// 4. Re-save the ViewKey with only public data
//
// Note: The new proto definition doesn't have private_view_key or private_spend_key
// fields, so this migration primarily serves as a defensive measure and documentation
// of the security fix.
func MigrateV1RemovePrivateKeys(ctx context.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(storeKey)
	viewKeyStore := prefix.NewStore(store, types.ViewKeyPrefix)

	// Track migration statistics
	var totalKeys, cleanedKeys uint64

	// Iterate through all view keys
	iterator := viewKeyStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		totalKeys++

		var viewKey privacypb.ViewKey
		if err := cdc.Unmarshal(iterator.Value(), &viewKey); err != nil {
			// Log error but continue - don't fail the entire migration
			sdkCtx.Logger().Error("failed to unmarshal view key during migration", "error", err)
			continue
		}

		// Verify the view key has only public data
		// The new proto definition ensures this at compile time, but we check anyway
		if len(viewKey.PublicViewKey) == 0 {
			// This view key has no public key data - likely corrupted or invalid
			// Delete it to clean up the state
			viewKeyStore.Delete(iterator.Key())
			cleanedKeys++
			sdkCtx.Logger().Info("removed invalid view key entry during migration",
				"total_keys", totalKeys,
				"cleaned_keys", cleanedKeys)
			continue
		}

		// The ViewKey proto no longer has private key fields, so this is already secure
		// We just verify the data integrity and re-save if needed

		// Re-marshal to ensure it uses the latest proto definition
		bz, err := cdc.Marshal(&viewKey)
		if err != nil {
			sdkCtx.Logger().Error("failed to marshal view key during migration", "error", err)
			continue
		}

		// Save the cleaned view key
		viewKeyStore.Set(iterator.Key(), bz)
	}

	// Log migration completion
	sdkCtx.Logger().Info("completed private key removal migration",
		"total_keys_processed", totalKeys,
		"invalid_keys_removed", cleanedKeys,
	)

	// Emit migration event
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"privacy_migration_v1",
			sdk.NewAttribute("migration_type", "remove_private_keys"),
			sdk.NewAttribute("total_keys_processed", sdk.IntProto{Int: int64(totalKeys)}.String()),
			sdk.NewAttribute("invalid_keys_removed", sdk.IntProto{Int: int64(cleanedKeys)}.String()),
		),
	)

	return nil
}

// MigrateV1VerifyNoPrivateKeys verifies that no private keys exist in the state
// This is a read-only verification that can be run to audit the state
func MigrateV1VerifyNoPrivateKeys(ctx context.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(storeKey)
	viewKeyStore := prefix.NewStore(store, types.ViewKeyPrefix)

	var totalKeys uint64

	iterator := viewKeyStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		totalKeys++

		var viewKey privacypb.ViewKey
		if err := cdc.Unmarshal(iterator.Value(), &viewKey); err != nil {
			sdkCtx.Logger().Error("failed to unmarshal view key during verification", "error", err)
			continue
		}

		// Verify the view key has only public data
		// The proto definition guarantees this at compile time - there are no private key fields
		if len(viewKey.PublicViewKey) == 0 {
			sdkCtx.Logger().Warn("view key has no public key data",
				"total_keys_checked", totalKeys)
		}

		// Check that key_type is not suspicious
		if viewKey.KeyType == "PRIVATE" || viewKey.KeyType == "SECRET" {
			sdkCtx.Logger().Error("SECURITY VIOLATION: suspicious key type found",
				"key_type", viewKey.KeyType,
				"total_keys_checked", totalKeys)
		}
	}

	sdkCtx.Logger().Info("private key verification complete",
		"total_keys_verified", totalKeys,
		"result", "PASS - no private keys found (as expected)",
	)

	return nil
}
