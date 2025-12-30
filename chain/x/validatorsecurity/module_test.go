// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package validatorsecurity

import (
	"encoding/json"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/validatorsecurity/keeper"
	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

func setupModule(t *testing.T) (AppModule, keeper.Keeper, sdk.Context, codec.JSONCodec) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)
	db := dbm.NewMemDB()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	binaryCodec := codec.NewProtoCodec(interfaceRegistry)
	var cdc codec.JSONCodec = binaryCodec

	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	k := keeper.NewKeeper(binaryCodec, storeKey, memKey, "authority", nil, nil, nil)

	return NewAppModule(binaryCodec, k), k, ctx, cdc
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

func TestAppModuleBasic_DefaultGenesis(t *testing.T) {
	basic := AppModuleBasic{}
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	genesis := basic.DefaultGenesis(cdc)
	require.NotNil(t, genesis)
	require.NotEmpty(t, genesis)
}

func TestAppModuleBasic_ValidateGenesis(t *testing.T) {
	basic := AppModuleBasic{}
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	tests := []struct {
		name    string
		genesis json.RawMessage
		wantErr bool
	}{
		{
			name:    "valid default genesis",
			genesis: basic.DefaultGenesis(cdc),
			wantErr: false,
		},
		{
			name:    "empty genesis",
			genesis: json.RawMessage(""),
			wantErr: true, // validatorsecurity module requires valid genesis JSON
		},
		{
			name:    "invalid json",
			genesis: json.RawMessage("{invalid}"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := basic.ValidateGenesis(cdc, nil, tt.genesis)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAppModule_ConsensusVersion(t *testing.T) {
	module, _, _, _ := setupModule(t)
	version := module.ConsensusVersion()
	require.Greater(t, version, uint64(0))
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
	memKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	k := keeper.NewKeeper(cdc, storeKey, memKey, "authority", nil, nil, nil)

	module := NewAppModule(cdc, k)
	require.NotNil(t, module)
	require.Equal(t, types.ModuleName, module.Name())
}
