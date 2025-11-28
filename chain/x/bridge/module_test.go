package bridge_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

func TestAppModuleBasic_Name(t *testing.T) {
	module := bridge.AppModuleBasic{}
	name := module.Name()

	require.Equal(t, types.ModuleName, name)
	require.Equal(t, "bridge", name)
}

func TestAppModuleBasic_RegisterServices(t *testing.T) {
	module := bridge.AppModuleBasic{}

	// RegisterServices should not panic with nil (it's a no-op)
	require.NotPanics(t, func() {
		module.RegisterServices(nil)
	})
}

func TestAppModuleBasic_Genesis(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	basic := bridge.AppModuleBasic{}

	bz := basic.DefaultGenesis(input.Cdc)
	require.NotEmpty(t, bz)

	var enc client.TxEncodingConfig
	require.NoError(t, basic.ValidateGenesis(input.Cdc, enc, bz))

	// Invalid JSON should fail
	require.Error(t, basic.ValidateGenesis(input.Cdc, enc, []byte("invalid")))
}

func TestNewAppModule(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil,
		nil,
		nil,
		nil,
	)

	module := bridge.NewAppModule(k)
	require.NotNil(t, module)
}

func TestAppModule_Name(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil,
		nil,
		nil,
		nil,
	)

	module := bridge.NewAppModule(k)
	name := module.Name()

	require.Equal(t, types.ModuleName, name)
	require.Equal(t, "bridge", name)
}

func TestAppModule_RegisterServices_Nil(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil,
		nil,
		nil,
		nil,
	)

	module := bridge.NewAppModule(k)

	// RegisterServices should panic with nil config
	require.Panics(t, func() {
		module.RegisterServices(nil)
	})
}

// mockModuleServices is a mock implementation of ModuleServices
type mockModuleServices struct {
	msgServer   bridgepb.MsgServer
	queryServer bridgepb.QueryServer
}

func (m *mockModuleServices) RegisterMsgServer(srv bridgepb.MsgServer) {
	m.msgServer = srv
}

func (m *mockModuleServices) RegisterQueryServer(srv bridgepb.QueryServer) {
	m.queryServer = srv
}

func TestAppModule_RegisterServices_Valid(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil,
		nil,
		nil,
		nil,
	)

	module := bridge.NewAppModule(k)
	mockServices := &mockModuleServices{}

	require.NotPanics(t, func() {
		module.RegisterServices(mockServices)
	})
	require.NotNil(t, mockServices.msgServer)
	require.NotNil(t, mockServices.queryServer)
}

func TestAppModule_BeginBlock(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil,
		nil,
		nil,
		nil,
	)

	module := bridge.NewAppModule(k)

	// BeginBlock should not panic
	require.NotPanics(t, func() {
		module.BeginBlock()
	})
}

func TestAppModule_EndBlock(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil,
		nil,
		nil,
		nil,
	)

	module := bridge.NewAppModule(k)

	// EndBlock should not panic
	require.NotPanics(t, func() {
		module.EndBlock()
	})
}

func TestAppModule_InitGenesis(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	legacyAmino := codec.NewLegacyAmino()
	ps := paramtypes.NewSubspace(input.Cdc, legacyAmino, input.StoreKey, input.MemStoreKey, "bridge").WithKeyTable(types.ParamKeyTable())
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, &ps, nil, nil, nil)

	module := bridge.NewAppModule(k)
	require.NoError(t, module.InitGenesis(input.Ctx, *types.DefaultGenesis()))
	genesisState := types.GenesisState{Params: types.DefaultGenesis().Params}
	require.NoError(t, module.InitGenesis(input.Ctx, genesisState))
}

func TestAppModule_ExportGenesis(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	legacyAmino := codec.NewLegacyAmino()
	ps := paramtypes.NewSubspace(input.Cdc, legacyAmino, input.StoreKey, input.MemStoreKey, "bridge").WithKeyTable(types.ParamKeyTable())
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, &ps, nil, nil, nil)
	module := bridge.NewAppModule(k)
	state := module.ExportGenesis(input.Ctx)
	require.True(t, state.Params != nil)
	require.Equal(t, types.DefaultGenesis().Params, state.Params)
}

func TestModuleBasic_RegisterServices_Coverage(t *testing.T) {
	// Create multiple instances to ensure all paths are covered
	basic1 := bridge.AppModuleBasic{}
	basic2 := bridge.AppModuleBasic{}

	name1 := basic1.Name()
	name2 := basic2.Name()

	require.Equal(t, name1, name2)
	require.Equal(t, "bridge", name1)

	// Call RegisterServices multiple times
	basic1.RegisterServices(nil)
	basic2.RegisterServices(nil)
}

func TestAppModule_MultipleOperations(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil,
		nil,
		nil,
		nil,
	)

	module := bridge.NewAppModule(k)

	require.NotPanics(t, func() {
		_ = module.Name()
		module.BeginBlock()
		module.EndBlock()
		_ = module.InitGenesis(input.Ctx, *types.DefaultGenesis())
		_ = module.ExportGenesis(input.Ctx)
	})
}
