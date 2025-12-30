// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package dataregistry

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

	"github.com/aequitas/aura/chain/x/dataregistry/keeper"
	"github.com/aequitas/aura/chain/x/dataregistry/params"
	"github.com/aequitas/aura/chain/x/dataregistry/types"
)

func setupModule(t *testing.T) (AppModule, *keeper.Keeper, sdk.Context, codec.JSONCodec) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	binaryCodec := codec.NewProtoCodec(interfaceRegistry)
	var cdc codec.JSONCodec = binaryCodec

	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	storeService := runtime.NewKVStoreService(storeKey)
	paramsStore := params.NewStore(types.DefaultParams())
	k := keeper.NewKeeper(storeService, binaryCodec, paramsStore, "authority", log.NewNopLogger())

	return NewAppModule(k), k, ctx, cdc
}

func TestAppModule_Name(t *testing.T) {
	module, _, _, _ := setupModule(t)
	require.Equal(t, types.ModuleName, module.Name())
}

func TestAppModuleBasic_Name(t *testing.T) {
	basic := AppModuleBasic{}
	require.Equal(t, types.ModuleName, basic.Name())
}

func TestAppModuleBasic_RegisterLegacyAminoCodec(t *testing.T) {
	basic := AppModuleBasic{}
	cdc := codec.NewLegacyAmino()

	require.NotPanics(t, func() {
		basic.RegisterLegacyAminoCodec(cdc)
	})
}

func TestAppModuleBasic_RegisterInterfaces(t *testing.T) {
	basic := AppModuleBasic{}
	registry := codectypes.NewInterfaceRegistry()

	require.NotPanics(t, func() {
		basic.RegisterInterfaces(registry)
	})
}

func TestAppModule_BeginBlock(t *testing.T) {
	module, _, _, _ := setupModule(t)

	require.NotPanics(t, func() {
		module.BeginBlock()
	})
}

func TestAppModule_EndBlock(t *testing.T) {
	module, _, _, _ := setupModule(t)

	require.NotPanics(t, func() {
		module.EndBlock()
	})
}

func TestAppModule_IsOnePerModuleType(t *testing.T) {
	module, _, _, _ := setupModule(t)

	require.NotPanics(t, func() {
		module.IsOnePerModuleType()
	})
}

func TestAppModule_IsAppModule(t *testing.T) {
	module, _, _, _ := setupModule(t)

	require.NotPanics(t, func() {
		module.IsAppModule()
	})
}

func TestNewAppModule(t *testing.T) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	storeService := runtime.NewKVStoreService(storeKey)
	paramsStore := params.NewStore(types.DefaultParams())
	k := keeper.NewKeeper(storeService, cdc, paramsStore, "authority", log.NewNopLogger())

	module := NewAppModule(k)
	require.NotNil(t, module)
	require.Equal(t, types.ModuleName, module.Name())
}
