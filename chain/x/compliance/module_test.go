package compliance

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
