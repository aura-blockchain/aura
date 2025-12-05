package keeper

import (
	"sync"
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

	"github.com/aequitas/aura/chain/x/economics/types"
)

const (
	testStoreKey = "economics"
)

var (
	// Ensure SDK config is only set up once across all tests
	setupSDKConfigOnce sync.Once
)

// setupKeeperForTest creates a keeper and context for testing
// This is an internal helper to avoid import cycles
func setupKeeperForTest(t *testing.T) (*Keeper, sdk.Context) {
	t.Helper()

	// Configure SDK with proper bech32 prefix for address validation
	// Use sync.Once to ensure this only happens once across all tests
	setupSDKConfigOnce.Do(func() {
		config := sdk.GetConfig()
		config.SetBech32PrefixForAccount("aura", "aurapub")
		config.SetBech32PrefixForValidator("auravaloper", "auravaloper pub")
		config.SetBech32PrefixForConsensusNode("auravalcons", "auravalconspub")
		config.Seal()
	})

	storeKey := storetypes.NewKVStoreKey(testStoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	logger := log.NewNopLogger()
	authority := "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr" // test governance address

	k := NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		authority,
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, logger)

	// Initialize with default params
	defaultParams := types.DefaultParams()
	require.NoError(t, k.SetParams(ctx, defaultParams))

	return k, ctx
}
