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

	"github.com/aequitas/aura/chain/x/validatorsecurity/keeper"
	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

const (
	ValidatorSecurityStoreKey = "validatorsecurity"
	ValidatorSecurityMemKey   = "validatorsecurity_mem"
)

// ValidatorSecurityKeeper creates a validator security keeper for testing
func ValidatorSecurityKeeper(t *testing.T) (keeper.Keeper, sdk.Context) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(ValidatorSecurityStoreKey)
	memKey := storetypes.NewMemoryStoreKey(ValidatorSecurityMemKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	authority := "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn" // governance address

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		memKey,
		authority,
		nil, // stakingKeeper
		nil, // slashingKeeper
		nil, // bankKeeper
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	params := types.DefaultParams()
	require.NoError(t, k.SetParams(ctx, params))

	return k, ctx
}
