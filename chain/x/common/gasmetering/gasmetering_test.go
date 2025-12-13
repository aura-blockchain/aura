package gasmetering

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func newGasContext(t *testing.T) (sdk.Context, storetypes.StoreKey) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey("gasmeter")
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		ChainID: "aura-gas",
		Time:    time.Unix(1_700_000_300, 0),
	}, false, log.NewNopLogger())

	return ctx.WithGasMeter(storetypes.NewGasMeter(1_000_000)), storeKey
}

func TestMeteredStoreChargesGas(t *testing.T) {
	ctx, storeKey := newGasContext(t)
	store := ctx.KVStore(storeKey)
	ms := NewMeteredStore(ctx.Context(), store, DefaultGasConfig())
	gm := ctx.GasMeter()

	before := gm.GasConsumed()
	ms.Set([]byte("k1"), []byte("v1"))
	require.Greater(t, gm.GasConsumed(), before)

	afterSet := gm.GasConsumed()
	val := ms.Get([]byte("k1"))
	require.Equal(t, []byte("v1"), val)
	require.Greater(t, gm.GasConsumed(), afterSet)

	afterGet := gm.GasConsumed()
	require.True(t, ms.Has([]byte("k1")))
	require.Greater(t, gm.GasConsumed(), afterGet)

	afterHas := gm.GasConsumed()
	ms.Delete([]byte("k1"))
	require.Greater(t, gm.GasConsumed(), afterHas)
}

func TestMeteredIteratorHonorsLimit(t *testing.T) {
	ctx, storeKey := newGasContext(t)
	store := ctx.KVStore(storeKey)

	store.Set([]byte("a1"), []byte("v1"))
	store.Set([]byte("a2"), []byte("v2"))
	store.Set([]byte("a3"), []byte("v3"))

	config := DefaultGasConfig()
	config.MaxIterationResults = 2
	ms := NewMeteredStore(ctx.Context(), store, config)

	iter := ms.Iterator([]byte("a"), []byte("b"))
	defer iter.Close()

	require.True(t, iter.Valid())
	iter.Next()
	require.True(t, iter.Valid())
	iter.Next()
	// Limit reached; iterator should now report invalid.
	require.False(t, iter.Valid())
}

func TestIterateWithLimit(t *testing.T) {
	ctx, storeKey := newGasContext(t)
	store := ctx.KVStore(storeKey)
	store.Set([]byte("k1"), []byte("v1"))
	store.Set([]byte("k2"), []byte("v2"))
	store.Set([]byte("k3"), []byte("v3"))

	err := IterateWithLimit(ctx.Context(), store, []byte("k"), 2, DefaultGasConfig(), func(key, value []byte) error {
		return nil
	})
	require.Error(t, err, "expected limit error when more than 2 items exist")
}

func TestConsumeGasHelpers(t *testing.T) {
	ctx, _ := newGasContext(t)
	config := DefaultGasConfig()
	gm := ctx.GasMeter()

	start := gm.GasConsumed()
	ConsumeGasForCrypto(ctx.Context(), "hash", config)
	require.Greater(t, gm.GasConsumed(), start)

	current := gm.GasConsumed()
	ConsumeGasForCrypto(ctx.Context(), "unknown-op", config)
	require.Greater(t, gm.GasConsumed(), current)

	current = gm.GasConsumed()
	ConsumeGasForValidation(ctx.Context(), "signature", config)
	require.Greater(t, gm.GasConsumed(), current)

	current = gm.GasConsumed()
	ConsumeGasForValidation(ctx.Context(), "custom", config)
	require.Greater(t, gm.GasConsumed(), current)
}
