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
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/testutil/mocks"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
)

// BridgeKeeper creates a bridge keeper for testing
func BridgeKeeper(t *testing.T) (*keeper.Keeper, sdk.Context) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	ps := paramtypes.NewSubspace(cdc,
		codec.NewLegacyAmino(),
		storeKey,
		storetypes.NewMemoryStoreKey("mem_bridge"),
		"BridgeParams",
	)

	// Create mock keepers
	bankKeeper := mocks.NewMockBankKeeper()
	accountKeeper := mocks.NewMockAccountKeeper()
	vcKeeper := mocks.NewMockVCRegistryKeeper()

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		&ps,
		bankKeeper,
		accountKeeper,
		vcKeeper,
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	params := types.DefaultParams()
	require.NoError(t, k.SetParams(ctx, params))

	return k, ctx
}

// BridgeKeeperWithMocks creates a bridge keeper with controllable mocks
func BridgeKeeperWithMocks(t *testing.T) (*keeper.Keeper, sdk.Context, *mocks.MockBankKeeper, *mocks.MockAccountKeeper, *mocks.MockVCRegistryKeeper) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	ps := paramtypes.NewSubspace(cdc,
		codec.NewLegacyAmino(),
		storeKey,
		storetypes.NewMemoryStoreKey("mem_bridge"),
		"BridgeParams",
	)

	// Create mock keepers
	bankKeeper := mocks.NewMockBankKeeper()
	accountKeeper := mocks.NewMockAccountKeeper()
	vcKeeper := mocks.NewMockVCRegistryKeeper()

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		&ps,
		bankKeeper,
		accountKeeper,
		vcKeeper,
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	params := types.DefaultParams()
	require.NoError(t, k.SetParams(ctx, params))

	return k, ctx, bankKeeper, accountKeeper, vcKeeper
}
