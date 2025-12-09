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
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/testutil/mocks"
	"github.com/aequitas/aura/chain/x/dex/keeper"
	"github.com/aequitas/aura/chain/x/dex/types"
)

// DexKeeper creates a dex keeper for testing
func DexKeeper(t *testing.T) (*keeper.Keeper, sdk.Context) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	// Create mock keepers
	bankKeeper := mocks.NewMockBankKeeper()
	accountKeeper := mocks.NewMockAccountKeeper()
	vcKeeper := mocks.NewMockVCRegistryKeeper()
	securityKeeper := mocks.NewMockSecurityKeeper()

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		bankKeeper,
		accountKeeper,
		vcKeeper,
		securityKeeper,
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	params := types.DefaultParams()
	require.NoError(t, k.SetParams(ctx, &params))

	return k, ctx
}

// DexKeeperWithMocks creates a dex keeper with controllable mocks
func DexKeeperWithMocks(t *testing.T) (*keeper.Keeper, sdk.Context, *mocks.MockBankKeeper, *mocks.MockAccountKeeper, *mocks.MockVCRegistryKeeper, *mocks.MockSecurityKeeper) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	// Create mock keepers
	bankKeeper := mocks.NewMockBankKeeper()
	accountKeeper := mocks.NewMockAccountKeeper()
	vcKeeper := mocks.NewMockVCRegistryKeeper()
	securityKeeper := mocks.NewMockSecurityKeeper()

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		bankKeeper,
		accountKeeper,
		vcKeeper,
		securityKeeper,
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	params := types.DefaultParams()
	require.NoError(t, k.SetParams(ctx, &params))

	return k, ctx, bankKeeper, accountKeeper, vcKeeper, securityKeeper
}
