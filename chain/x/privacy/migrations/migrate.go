// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package migrations

import (
	"cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/privacy/keeper"
)

// Migrator handles in-place state migrations for the privacy module
type Migrator struct {
	keeper   *keeper.Keeper
	storeKey types.StoreKey
	cdc      codec.BinaryCodec
}

// NewMigrator creates a new Migrator instance
func NewMigrator(k *keeper.Keeper, storeKey types.StoreKey, cdc codec.BinaryCodec) Migrator {
	return Migrator{
		keeper:   k,
		storeKey: storeKey,
		cdc:      cdc,
	}
}

// Migrate1to2 migrates from consensus version 1 to version 2
//
// This migration removes any private key data that may have been stored
// before the security fix was implemented.
//
// SECURITY: This migration is CRITICAL. It ensures that no private keys remain
// in the blockchain state after upgrading to the secure version.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	// Execute the private key removal migration
	return MigrateV1RemovePrivateKeys(ctx, m.storeKey, m.cdc)
}
