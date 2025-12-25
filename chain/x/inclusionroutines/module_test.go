package inclusionroutines

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	runtime "github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmod "github.com/cosmos/cosmos-sdk/types/module"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/aequitas/aura/chain/x/inclusionroutines/keeper"
	"github.com/aequitas/aura/chain/x/inclusionroutines/params"
	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
)

// mockInvariantRegistry implements sdk.InvariantRegistry for testing
type mockInvariantRegistry struct {
	routes []string
}

func (m *mockInvariantRegistry) RegisterRoute(moduleName, route string, inv sdk.Invariant) {
	m.routes = append(m.routes, moduleName+"/"+route)
}

func setupModule(t *testing.T) (AppModule, *keeper.Keeper, sdk.Context, codec.Codec) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	protoCodec := codec.NewProtoCodec(interfaceRegistry)
	paramsStore := params.NewStore(types.DefaultParams())
	k := keeper.NewKeeper(
		runtime.NewKVStoreService(storeKey),
		protoCodec,
		paramsStore,
		"authority",
		log.NewNopLogger(),
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())
	module := NewAppModule(k)
	return module, k, ctx, protoCodec
}

func TestRegisterServices(t *testing.T) {
	module, _, _, protoCodec := setupModule(t)

	msgServer := grpc.NewServer()
	queryServer := grpc.NewServer()
	configurator := sdkmod.NewConfigurator(protoCodec, msgServer, queryServer)

	module.RegisterServices(configurator)
	require.NoError(t, configurator.Error())
}

func TestAppModuleBasic_Name(t *testing.T) {
	basic := AppModuleBasic{}
	require.Equal(t, types.ModuleName, basic.Name())
}

func TestAppModule_Name(t *testing.T) {
	module, _, _, _ := setupModule(t)
	require.Equal(t, types.ModuleName, module.Name())
}

func TestAppModuleBasic_RegisterInterfaces(t *testing.T) {
	basic := AppModuleBasic{}
	registry := codectypes.NewInterfaceRegistry()
	// Should not panic
	basic.RegisterInterfaces(registry)
}

func TestAppModuleBasic_RegisterLegacyAminoCodec(t *testing.T) {
	basic := AppModuleBasic{}
	cdc := codec.NewLegacyAmino()
	// Should not panic
	basic.RegisterLegacyAminoCodec(cdc)
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

	// Get default genesis
	defaultGen := basic.DefaultGenesis(cdc)

	// Validate should pass
	err := basic.ValidateGenesis(cdc, nil, defaultGen)
	require.NoError(t, err)
}

func TestAppModuleBasic_ValidateGenesis_Invalid(t *testing.T) {
	basic := AppModuleBasic{}
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Invalid JSON should fail
	err := basic.ValidateGenesis(cdc, nil, []byte("invalid json"))
	require.Error(t, err)
}

func TestAppModuleBasic_ValidateGenesis_Empty(t *testing.T) {
	basic := AppModuleBasic{}
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Empty/whitespace should pass (returns nil early)
	err := basic.ValidateGenesis(cdc, nil, []byte("   "))
	require.NoError(t, err)
}

func TestAppModuleBasic_GetTxCmd(t *testing.T) {
	basic := AppModuleBasic{}
	// Returns nil by design
	require.Nil(t, basic.GetTxCmd())
}

func TestAppModuleBasic_GetQueryCmd(t *testing.T) {
	basic := AppModuleBasic{}
	// Returns nil by design
	require.Nil(t, basic.GetQueryCmd())
}

func TestAppModule_InitGenesis(t *testing.T) {
	module, _, ctx, cdc := setupModule(t)

	// Get default genesis
	basic := AppModuleBasic{}
	genesis := basic.DefaultGenesis(cdc)

	// InitGenesis should succeed - no panic
	require.NotPanics(t, func() {
		module.InitGenesis(ctx, cdc, genesis)
	})
}

func TestAppModule_InitGenesis_Empty(t *testing.T) {
	module, _, ctx, cdc := setupModule(t)

	// Empty genesis should use defaults
	require.NotPanics(t, func() {
		module.InitGenesis(ctx, cdc, []byte("   "))
	})
}

func TestAppModule_ExportGenesis(t *testing.T) {
	module, _, ctx, cdc := setupModule(t)

	// Initialize first
	basic := AppModuleBasic{}
	genesis := basic.DefaultGenesis(cdc)
	module.InitGenesis(ctx, cdc, genesis)

	// Export should succeed
	exported := module.ExportGenesis(ctx, cdc)
	require.NotNil(t, exported)
}

func TestAppModule_ConsensusVersion(t *testing.T) {
	module, _, _, _ := setupModule(t)
	require.Equal(t, uint64(1), module.ConsensusVersion())
}

func TestAppModule_BeginBlock(t *testing.T) {
	module, _, ctx, _ := setupModule(t)
	// Should not error
	err := module.BeginBlock(ctx)
	require.NoError(t, err)
}

func TestAppModule_EndBlock(t *testing.T) {
	module, _, ctx, _ := setupModule(t)
	// Should not error
	err := module.EndBlock(ctx)
	require.NoError(t, err)
}

func TestAppModule_RegisterInvariants(t *testing.T) {
	module, _, _, _ := setupModule(t)
	registry := &mockInvariantRegistry{}
	// Should not panic
	module.RegisterInvariants(registry)
}

func TestAppModule_IsAppModule(t *testing.T) {
	module, _, _, _ := setupModule(t)
	// Should not panic - used for interface compliance
	module.IsAppModule()
}

func TestAppModule_IsOnePerModuleType(t *testing.T) {
	module, _, _, _ := setupModule(t)
	// Should not panic - used for depinject
	module.IsOnePerModuleType()
}
