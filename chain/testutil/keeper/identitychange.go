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

	idkeeper "github.com/aequitas/aura/chain/x/identitychange/keeper"
	idparams "github.com/aequitas/aura/chain/x/identitychange/params"
	idtypes "github.com/aequitas/aura/chain/x/identitychange/types"
)

const (
	IdentityChangeStoreKey = "identitychange"
	// Default governance authority for testing
	TestAuthority = "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
)

// IdentityChangeKeeper creates an identity change keeper for testing
func IdentityChangeKeeper(t *testing.T) (*idkeeper.Keeper, sdk.Context) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(IdentityChangeStoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	logger := log.NewNopLogger()

	paramsStore := idparams.NewStore(idtypes.DefaultParams())

	k := idkeeper.NewKeeper(
		runtime.NewKVStoreService(storeKey),
		cdc,
		paramsStore,
		TestAuthority,
		logger,
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, logger)

	return k, ctx
}
