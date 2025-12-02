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
