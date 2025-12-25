// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	irkeeper "github.com/aequitas/aura/chain/x/inclusionroutines/keeper"
	irparams "github.com/aequitas/aura/chain/x/inclusionroutines/params"
	irtypes "github.com/aequitas/aura/chain/x/inclusionroutines/types"
)

// InclusionRoutinesKeeper creates an inclusion routines keeper for testing
func InclusionRoutinesKeeper(t *testing.T) (*irkeeper.Keeper, sdk.Context) {
	t.Helper()

	// Create store key
	storeKey := storetypes.NewKVStoreKey(irtypes.StoreKey)

	// Create in-memory database and commit multi-store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	// Create codec
	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	// Create params store
	paramsStore := irparams.NewStore(irtypes.DefaultParams())

	// Create logger
	logger := log.NewNopLogger()

	// Create KVStoreService from store key (Cosmos SDK v0.50 pattern)
	storeService := runtime.NewKVStoreService(storeKey)

	// Create keeper with all required dependencies
	k := irkeeper.NewKeeper(
		storeService,
		cdc,
		paramsStore,
		"aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn", // governance address
		logger,
	)

	// Create context
	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, logger)

	return k, ctx
}
