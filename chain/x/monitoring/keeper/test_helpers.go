package keeper

import (
	"context"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewTestKeeper creates a new keeper for testing with proper store service
func NewTestKeeper(t *testing.T) (Keeper, sdk.Context) {
	// Setup store
	storeKey := storetypes.NewKVStoreKey("monitoring")

	// Create database and commit multi-store
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	if err := cms.LoadLatestVersion(); err != nil {
		t.Fatalf("failed to load latest version: %v", err)
	}

	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create context with proper store
	ctx := sdk.NewContext(cms, tmproto.Header{Height: 1}, false, log.NewNopLogger())

	// Create store service from runtime
	storeService := runtime.NewKVStoreService(storeKey)

	// Create keeper
	k := NewKeeper(
		cdc,
		storeService,
		"aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn", // authority address
	)

	return *k, ctx
}

// GetTestContext returns a test SDK context
func GetTestContext() sdk.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	ctx := sdk.NewContext(cms, tmproto.Header{}, false, log.NewNopLogger())
	return ctx
}

// storeServiceWrapper wraps storetypes.KVStoreKey to implement store.KVStoreService
type storeServiceWrapper struct {
	key storetypes.StoreKey
}

// OpenKVStore implements store.KVStoreService
func (s storeServiceWrapper) OpenKVStore(ctx context.Context) storetypes.KVStore {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.KVStore(s.key)
}

// NewStoreServiceWrapper creates a store service wrapper for testing
func NewStoreServiceWrapper(key storetypes.StoreKey) storeServiceWrapper {
	return storeServiceWrapper{key: key}
}
