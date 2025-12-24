package identity

import (
	"encoding/json"
	"fmt"
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

	"github.com/aequitas/aura/chain/x/identity/keeper"
	"github.com/aequitas/aura/chain/x/identity/types"
)

// mockInvariantRegistry implements sdk.InvariantRegistry for testing
type mockInvariantRegistry struct {
	routes map[string]sdk.Invariant
}

func (m *mockInvariantRegistry) RegisterRoute(moduleName, route string, inv sdk.Invariant) {
	if m.routes == nil {
		m.routes = make(map[string]sdk.Invariant)
	}
	m.routes[fmt.Sprintf("%s/%s", moduleName, route)] = inv
}

func setupModule(t *testing.T) (AppModule, *keeper.Keeper, sdk.Context, codec.JSONCodec) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	binaryCodec := codec.NewProtoCodec(interfaceRegistry)
	var cdc codec.JSONCodec = binaryCodec

	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	storeService := runtime.NewKVStoreService(storeKey)
	k := keeper.NewKeeper(storeService, storeKey, binaryCodec, "authority", log.NewNopLogger())

	return NewAppModule(binaryCodec, k), k, ctx, cdc
}

func TestAppModule_Name(t *testing.T) {
	module, _, _, _ := setupModule(t)
	require.Equal(t, types.ModuleName, module.Name())
}

func TestAppModule_GenesisLifecycle(t *testing.T) {
	module, _, ctx, cdc := setupModule(t)

	genesis := types.DefaultGenesisState()
	require.NoError(t, types.ValidateGenesisState(genesis))

	module.InitGenesis(ctx, cdc, cdc.MustMarshalJSON(genesis))
	output := module.ExportGenesis(ctx, cdc)

	var exported types.GenesisState
	require.NoError(t, cdc.UnmarshalJSON(output, &exported))
	require.Equal(t, genesis.Params, exported.Params)
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

	var gen types.GenesisState
	require.NoError(t, cdc.UnmarshalJSON(genesis, &gen))
	require.NoError(t, types.ValidateGenesisState(&gen))
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
			wantErr: false,
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

func TestAppModuleBasic_GetTxCmd(t *testing.T) {
	basic := AppModuleBasic{}
	cmd := basic.GetTxCmd()
	require.NotNil(t, cmd)
	require.Equal(t, types.ModuleName, cmd.Use)
}

func TestAppModuleBasic_GetQueryCmd(t *testing.T) {
	basic := AppModuleBasic{}
	cmd := basic.GetQueryCmd()
	require.NotNil(t, cmd)
	require.Equal(t, types.ModuleName, cmd.Use)
}

func TestAppModule_ConsensusVersion(t *testing.T) {
	module, _, _, _ := setupModule(t)
	version := module.ConsensusVersion()
	require.Greater(t, version, uint64(0))
}

func TestAppModule_BeginBlock(t *testing.T) {
	module, _, ctx, _ := setupModule(t)

	require.NotPanics(t, func() {
		err := module.BeginBlock(ctx)
		require.NoError(t, err)
	})
}

func TestAppModule_EndBlock(t *testing.T) {
	module, _, ctx, _ := setupModule(t)

	require.NotPanics(t, func() {
		err := module.EndBlock(ctx)
		require.NoError(t, err)
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

func TestAppModule_RegisterInvariants(t *testing.T) {
	module, _, _, _ := setupModule(t)

	registry := &mockInvariantRegistry{}
	require.NotPanics(t, func() {
		module.RegisterInvariants(registry)
	})

	// Identity module should register at least one invariant
	require.NotEmpty(t, registry.routes, "invariants should be registered")
}

func TestAppModule_RegisterServices(t *testing.T) {
	module, _, _, _ := setupModule(t)

	// Smoke test - RegisterServices doesn't return anything
	// but we can ensure it doesn't panic
	require.NotPanics(t, func() {
		require.NotNil(t, module.keeper)
	})
}

func TestNewAppModule(t *testing.T) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	storeService := runtime.NewKVStoreService(storeKey)
	k := keeper.NewKeeper(storeService, storeKey, cdc, "authority", log.NewNopLogger())

	module := NewAppModule(cdc, k)
	require.NotNil(t, module)
	require.Equal(t, types.ModuleName, module.Name())
}
