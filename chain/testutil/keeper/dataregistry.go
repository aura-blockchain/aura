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

	drkeeper "github.com/aequitas/aura/chain/x/dataregistry/keeper"
	drparams "github.com/aequitas/aura/chain/x/dataregistry/params"
	drtypes "github.com/aequitas/aura/chain/x/dataregistry/types"
)

const (
	DataRegistryStoreKey = "dataregistry"
)

// DataRegistryKeeper creates a data registry keeper for testing
func DataRegistryKeeper(t *testing.T) (*drkeeper.Keeper, sdk.Context) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(DataRegistryStoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	logger := log.NewNopLogger()

	paramsStore := drparams.NewStore(drtypes.DefaultParams())

	k := drkeeper.NewKeeper(
		runtime.NewKVStoreService(storeKey),
		cdc,
		paramsStore,
		TestAuthority,
		logger,
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, logger)

	return k, ctx
}
