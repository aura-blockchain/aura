package app

import (
	"testing"

	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/codec"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/aiassistant"
	"github.com/aequitas/aura/chain/x/bridge"
	bridgekeeper "github.com/aequitas/aura/chain/x/bridge/keeper"
	bridgetypes "github.com/aequitas/aura/chain/x/bridge/types"
	"github.com/aequitas/aura/chain/x/compliance"
	"github.com/aequitas/aura/chain/x/confidencescore"
	"github.com/aequitas/aura/chain/x/cryptography"
	"github.com/aequitas/aura/chain/x/dataregistry"
	"github.com/aequitas/aura/chain/x/dex"
	"github.com/aequitas/aura/chain/x/governance"
	"github.com/aequitas/aura/chain/x/identitychange"
	idkeeper "github.com/aequitas/aura/chain/x/identitychange/keeper"
	"github.com/aequitas/aura/chain/x/inclusionroutines"
	"github.com/aequitas/aura/chain/x/validatorsecurity"
	"github.com/aequitas/aura/chain/x/vcregistry"
	"github.com/aequitas/aura/chain/x/walletsecurity"
	"github.com/stretchr/testify/require"
)

func TestRegisterGRPCServices(t *testing.T) {
	k := idkeeper.NewKeeper(nil)
	encoding := MakeEncodingConfig()
	manager := NewModuleManager(
		encoding,
		[]aiassistant.AppModule{},
		[]identitychange.AppModule{identitychange.NewAppModule(k)},
		[]inclusionroutines.AppModule{},
		[]confidencescore.AppModule{},
		[]vcregistry.AppModule{},
		[]dataregistry.AppModule{},
		[]compliance.AppModule{},
		[]walletsecurity.AppModule{},
		[]validatorsecurity.AppModule{},
		[]cryptography.AppModule{},
		[]governance.AppModule{},
		[]dex.AppModule{},
		[]bridge.AppModule{},
	)

	server := grpc.NewServer()
	defer server.Stop()
	manager.RegisterGRPCServices(server)
	info := server.GetServiceInfo()

	require.True(t, serviceRegistered(info, "aura.identitychange.v1beta1.Msg"), "Msg service not registered")
	require.True(t, serviceRegistered(info, "aura.identitychange.v1beta1.Query"), "Query service not registered")
}

func serviceRegistered(info map[string]grpc.ServiceInfo, name string) bool {
	_, ok := info[name]
	return ok
}

func TestModuleManagerBridgeGenesis(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	legacyAmino := codec.NewLegacyAmino()
	ps := paramtypes.NewSubspace(input.Cdc, legacyAmino, input.StoreKey, input.MemStoreKey, "bridge").WithKeyTable(bridgetypes.ParamKeyTable())
	bridgeKeeper := bridgekeeper.NewKeeper(input.Cdc, input.StoreKey, &ps, nil, nil, nil)
	encoding := MakeEncodingConfig()
	manager := NewModuleManager(
		encoding,
		[]aiassistant.AppModule{},
		[]identitychange.AppModule{},
		[]inclusionroutines.AppModule{},
		[]confidencescore.AppModule{},
		[]vcregistry.AppModule{},
		[]dataregistry.AppModule{},
		[]compliance.AppModule{},
		[]walletsecurity.AppModule{},
		[]validatorsecurity.AppModule{},
		[]cryptography.AppModule{},
		[]governance.AppModule{},
		[]dex.AppModule{},
		[]bridge.AppModule{bridge.NewAppModule(bridgeKeeper)},
	)

	genesis := *bridgetypes.DefaultGenesis()
	require.NoError(t, manager.InitBridgeGenesis(input.Ctx, genesis))

	exported := manager.ExportBridgeGenesis(input.Ctx)
	require.Len(t, exported, 1)
	require.Equal(t, genesis.Params, exported[0].Params)
}
