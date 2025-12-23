package compliance

import (
	"encoding/json"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/compliance/keeper"
	compliancetypes "github.com/aequitas/aura/chain/x/compliance/types"
)

func setupModule(t *testing.T) (AppModule, *keeper.Keeper, sdk.Context, codec.JSONCodec) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey("compliance")
	db := dbm.NewMemDB()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	binaryCodec := codec.NewProtoCodec(interfaceRegistry)
	var cdc codec.JSONCodec = binaryCodec
	k := keeper.NewKeeper(binaryCodec, storeKey)

	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(k.StoreKey(), storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	return NewAppModule(k), k, ctx, cdc
}

func defaultGenesisState(t *testing.T, cdc codec.JSONCodec) *compliancetypes.GenesisState {
	t.Helper()
	var gen compliancetypes.GenesisState
	require.NoError(t, cdc.UnmarshalJSON(AppModuleBasic{}.DefaultGenesis(cdc), &gen))
	return &gen
}

func TestAppModule_Name(t *testing.T) {
	module, _, _, _ := setupModule(t)

	require.Equal(t, compliancetypes.ModuleName, module.Name())
}

func TestAppModule_GenesisLifecycle(t *testing.T) {
	module, _, ctx, cdc := setupModule(t)

	genesis := defaultGenesisState(t, cdc)
	require.NoError(t, compliancetypes.ValidateGenesis(genesis))

	module.InitGenesis(ctx, cdc, cdc.MustMarshalJSON(genesis))
	output := module.ExportGenesis(ctx, cdc)

	var exported compliancetypes.GenesisState
	require.NoError(t, cdc.UnmarshalJSON(output, &exported))
	require.Equal(t, genesis.Params, exported.Params)
}

func TestAppModuleBasic_Name(t *testing.T) {
	basic := AppModuleBasic{}
	require.Equal(t, compliancetypes.ModuleName, basic.Name())
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

	var gen compliancetypes.GenesisState
	require.NoError(t, cdc.UnmarshalJSON(genesis, &gen))
	require.NoError(t, compliancetypes.ValidateGenesis(&gen))
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
	require.Equal(t, compliancetypes.ModuleName, cmd.Use)
}

func TestAppModuleBasic_GetQueryCmd(t *testing.T) {
	basic := AppModuleBasic{}
	cmd := basic.GetQueryCmd()
	require.NotNil(t, cmd)
	require.Equal(t, compliancetypes.ModuleName, cmd.Use)
}

func TestAppModule_RegisterServices(t *testing.T) {
	module, _, _, _ := setupModule(t)

	// This is a smoke test - RegisterServices doesn't return anything
	// but we can ensure it doesn't panic
	require.NotPanics(t, func() {
		// We can't actually call RegisterServices without a full configurator,
		// but we can verify the module exists
		require.NotNil(t, module.keeper)
	})
}

func TestAppModule_ConsensusVersion(t *testing.T) {
	module, _, _, _ := setupModule(t)
	version := module.ConsensusVersion()
	require.Greater(t, version, uint64(0))
}

func TestAppModule_BeginBlock(t *testing.T) {
	module, _, ctx, _ := setupModule(t)

	// BeginBlock should not panic
	require.NotPanics(t, func() {
		module.BeginBlock(ctx)
	})
}

func TestAppModule_EndBlock(t *testing.T) {
	module, _, ctx, _ := setupModule(t)

	// EndBlock should not panic
	require.NotPanics(t, func() {
		module.EndBlock(ctx)
	})
}

func TestAppModule_IsOnePerModuleType(t *testing.T) {
	module, _, _, _ := setupModule(t)

	// Just verify the method exists and doesn't panic
	require.NotPanics(t, func() {
		module.IsOnePerModuleType()
	})
}

func TestAppModule_IsAppModule(t *testing.T) {
	module, _, _, _ := setupModule(t)

	// Just verify the method exists and doesn't panic
	require.NotPanics(t, func() {
		module.IsAppModule()
	})
}

func TestAppModule_RegisterInvariants(t *testing.T) {
	module, _, _, _ := setupModule(t)

	// RegisterInvariants should not panic
	require.NotPanics(t, func() {
		module.RegisterInvariants(nil)
	})
}
