package determinism

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

func newDeterminismContext(t *testing.T, height int64, blockTime time.Time) sdk.Context {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey("determinism")
	memKey := storetypes.NewMemoryStoreKey("determinism-mem")

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	header := cmtproto.Header{
		ChainID: "aura-determinism",
		Height:  height,
		Time:    blockTime,
		AppHash: []byte("app-hash"),
	}

	return sdk.NewContext(stateStore, header, false, log.NewNopLogger())
}

func TestDeterministicRNGConsistency(t *testing.T) {
	blockTime := time.Unix(1_700_000_000, 0).UTC()
	ctx := newDeterminismContext(t, 42, blockTime)
	wrapped := sdk.WrapSDKContext(ctx)

	rng1 := NewDeterministicRNG(wrapped, []byte("entropy"))
	rng2 := NewDeterministicRNG(wrapped, []byte("entropy"))

	require.Equal(t, rng1.Bytes(32), rng2.Bytes(32), "same ctx + entropy must produce identical bytes")
	require.Equal(t, rng1.Uint64(), rng2.Uint64())
	require.Equal(t, rng1.Int63n(10), rng2.Int63n(10))

	// Changing block height should change the seed.
	ctx2 := newDeterminismContext(t, 43, blockTime)
	rngDifferent := NewDeterministicRNG(sdk.WrapSDKContext(ctx2), []byte("entropy"))
	require.NotEqual(t, rng1.Bytes(16), rngDifferent.Bytes(16))
}

func TestDeterministicRNGIntnBounds(t *testing.T) {
	ctx := sdk.WrapSDKContext(newDeterminismContext(t, 99, time.Unix(1_700_000_100, 0).UTC()))
	rng := NewDeterministicRNG(ctx)

	for i := 0; i < 100; i++ {
		n := rng.Intn(10)
		require.GreaterOrEqual(t, n, 0)
		require.Less(t, n, 10)
	}

	require.Panics(t, func() { rng.Intn(0) })
	require.Panics(t, func() { rng.Int63n(0) })
}

func TestDeterministicShuffleSlice(t *testing.T) {
	ctx := sdk.WrapSDKContext(newDeterminismContext(t, 1, time.Unix(1_700_000_200, 0).UTC()))

	original := []int{0, 1, 2, 3, 4}
	first := append([]int(nil), original...)
	second := append([]int(nil), original...)

	DeterministicShuffleSlice(ctx, first, []byte("entropy-a"))
	DeterministicShuffleSlice(ctx, second, []byte("entropy-a"))
	require.Equal(t, first, second, "shuffle must be deterministic with same entropy")
	require.NotEqual(t, original, first, "shuffle should change order")

	third := append([]int(nil), original...)
	DeterministicShuffleSlice(ctx, third, []byte("entropy-b"))
	require.NotEqual(t, first, third, "different entropy should change deterministic order")
}

func TestDeterministicTimeHelpers(t *testing.T) {
	blockTime := time.Date(2025, 1, 2, 15, 4, 5, 0, time.UTC)
	ctx := newDeterminismContext(t, 7, blockTime)
	wrapped := sdk.WrapSDKContext(ctx)

	require.Equal(t, blockTime, GetBlockTime(wrapped))
	require.Equal(t, int64(7), GetBlockHeight(wrapped))
	require.Equal(t, blockTime.Unix(), GetBlockTimestamp(wrapped))

	require.Equal(t, time.Hour, TimeSince(wrapped, blockTime.Add(-1*time.Hour)))
	require.Equal(t, 2*time.Hour, TimeUntil(wrapped, blockTime.Add(2*time.Hour)))

	require.Equal(t, blockTime.Format(time.RFC3339), FormatBlockTime(wrapped, time.RFC3339))

	dayStart := time.Date(blockTime.Year(), blockTime.Month(), blockTime.Day(), 0, 0, 0, 0, time.UTC).Unix()
	require.Equal(t, dayStart, GetDayTimestamp(wrapped))
	require.Equal(t, blockTime.Truncate(time.Hour).Unix(), GetHourTimestamp(wrapped))

	require.False(t, IsExpired(wrapped, blockTime.Add(time.Hour)))
	require.True(t, IsExpired(wrapped, blockTime.Add(-time.Second)))
	require.True(t, IsNotYetValid(wrapped, blockTime.Add(time.Second)))
	require.False(t, IsNotYetValid(wrapped, blockTime.Add(-time.Second)))

	require.True(t, IsInWindow(wrapped, blockTime.Add(-time.Minute), blockTime.Add(time.Minute)))
	require.False(t, IsInWindow(wrapped, blockTime.Add(time.Minute), blockTime.Add(2*time.Minute)))

	require.NoError(t, ValidateTimeWindow(blockTime, blockTime.Add(time.Minute)))
	require.Error(t, ValidateTimeWindow(blockTime, blockTime.Add(-time.Minute)))

	tr, err := NewTimeRange(blockTime, blockTime.Add(2*time.Minute))
	require.NoError(t, err)
	require.True(t, tr.Contains(blockTime.Add(time.Minute)))
	require.False(t, tr.Contains(blockTime.Add(3*time.Minute)))
	require.Equal(t, 2*time.Minute, tr.Duration())
	require.True(t, tr.IsActive(wrapped))

	timer := NewDeterministicTimer(wrapped, time.Hour)
	require.False(t, timer.IsExpired(wrapped))
	require.Equal(t, time.Hour, timer.RemainingTime(wrapped))
	require.Equal(t, time.Duration(0), timer.ElapsedTime(wrapped))
	require.InDelta(t, 0.0, timer.Progress(wrapped), 1e-9)

	// Advance block time and re-evaluate.
	advanced := ctx.WithBlockTime(blockTime.Add(90 * time.Minute))
	advancedCtx := sdk.WrapSDKContext(advanced)
	require.True(t, timer.IsExpired(advancedCtx))
	require.Equal(t, time.Duration(0), timer.RemainingTime(advancedCtx))
	require.Equal(t, 90*time.Minute, timer.ElapsedTime(advancedCtx))
	require.InDelta(t, 1.0, timer.Progress(advancedCtx), 1e-9)
}
