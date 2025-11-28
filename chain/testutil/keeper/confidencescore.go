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
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	cskeeper "github.com/aequitas/aura/chain/x/confidencescore/keeper"
	csparams "github.com/aequitas/aura/chain/x/confidencescore/params"
	cstypes "github.com/aequitas/aura/chain/x/confidencescore/types"
)

// ConfidenceScoreKeeper creates a confidence score keeper for testing
func ConfidenceScoreKeeper(t *testing.T) (*cskeeper.Keeper, sdk.Context) {
	t.Helper()

	// Create store key
	storeKey := storetypes.NewKVStoreKey(cstypes.StoreKey)

	// Create in-memory database and commit multi-store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	// Create codec
	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	// Create params store
	paramsStore := csparams.NewStore(cstypes.DefaultParams())

	// Create logger
	logger := log.NewNopLogger()

	// Create KVStoreService from store key (Cosmos SDK v0.50 pattern)
	storeService := runtime.NewKVStoreService(storeKey)

	// Create keeper with all required dependencies
	k := cskeeper.NewKeeper(
		storeService,
		cdc,
		paramsStore,
		"aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn", // test authority address
		logger,
	)

	// Create context
	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, logger)

	return k, ctx
}
