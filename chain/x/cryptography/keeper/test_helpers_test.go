package keeper_test

import (
	"context"
	"testing"
	"time"

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

	"github.com/aequitas/aura/chain/x/cryptography/keeper"
	"github.com/aequitas/aura/chain/x/cryptography/types"
)

func setupKeeper(t *testing.T) (keeper.Keeper, context.Context) {
	storeKey := storetypes.NewKVStoreKey(types.ModuleName)

	// Create database and commit multi-store
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, cms.LoadLatestVersion())

	// Create context with proper store
	header := cmtproto.Header{
		Height: 1,
		Time:   time.Now(),
	}
	ctx := sdk.NewContext(cms, header, false, log.NewNopLogger())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		log.NewNopLogger(),
		"authority",
	)

	// Initialize default params
	params := types.DefaultParams()
	err := k.SetParams(ctx, &params)
	require.NoError(t, err)

	return *k, ctx
}
