package keeper

import (
	"sync"
	"testing"
	"time"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	runtime "github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// TestKeepers holds all keeper instances for integration testing
type TestKeepers struct {
	// Add keeper fields as needed
}

// TestInput represents a standard test input for keeper tests
type TestInput struct {
	Ctx         sdk.Context
	Cdc         codec.Codec
	StoreKey    *storetypes.KVStoreKey
	MemStoreKey *storetypes.MemoryStoreKey
	DB          *dbm.MemDB
	CMS         store.CommitMultiStore
}

// CreateTestInput creates a standard test input for keeper testing
func CreateTestInput(t testing.TB) TestInput {
	t.Helper()

	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())

	storeKey := storetypes.NewKVStoreKey("test")
	memStoreKey := storetypes.NewMemoryStoreKey("mem_test")

	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	cms.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)

	require.NoError(t, cms.LoadLatestVersion())

	ctx := sdk.NewContext(cms, cmtproto.Header{
		Height: 1,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	return TestInput{
		Ctx:         ctx,
		Cdc:         cdc,
		StoreKey:    storeKey,
		MemStoreKey: memStoreKey,
		DB:          db,
		CMS:         cms,
	}
}

// CreateTestInputWithKeys creates test input with custom store keys
func CreateTestInputWithKeys(t testing.TB, keys ...string) TestInput {
	t.Helper()

	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())

	storeKeys := make([]*storetypes.KVStoreKey, len(keys))
	for i, key := range keys {
		storeKeys[i] = storetypes.NewKVStoreKey(key)
		cms.MountStoreWithDB(storeKeys[i], storetypes.StoreTypeIAVL, db)
	}

	require.NoError(t, cms.LoadLatestVersion())

	ctx := sdk.NewContext(cms, cmtproto.Header{
		Height: 1,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	return TestInput{
		Ctx:      ctx,
		Cdc:      cdc,
		StoreKey: storeKeys[0],
		DB:       db,
		CMS:      cms,
	}
}

// GenTestAddr generates a random test address
func GenTestAddr() sdk.AccAddress {
	return sdk.AccAddress([]byte("test_address_______"))
}

// GenTestAddrs generates multiple test addresses
func GenTestAddrs(count int) []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, count)
	for i := 0; i < count; i++ {
		addrs[i] = sdk.AccAddress(append([]byte("addr"), byte(i)))
	}
	return addrs
}

// MockTime returns a deterministic time for testing
func MockTime() time.Time {
	return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
}

// NewTestCodec creates a codec for testing
func NewTestCodec() codec.Codec {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	return codec.NewProtoCodec(interfaceRegistry)
}

// AdvanceBlockHeight advances the block height in the context
func AdvanceBlockHeight(ctx sdk.Context, blocks int64) sdk.Context {
	header := ctx.BlockHeader()
	header.Height += blocks
	header.Time = header.Time.Add(time.Duration(blocks) * 5 * time.Second)
	return ctx.WithBlockHeader(header)
}

// AdvanceTime advances time in the context
func AdvanceTime(ctx sdk.Context, duration time.Duration) sdk.Context {
	header := ctx.BlockHeader()
	header.Time = header.Time.Add(duration)
	return ctx.WithBlockHeader(header)
}

// CreateTestInputWithTime creates test input with a specific block time
func CreateTestInputWithTime(t testing.TB, blockTime time.Time) TestInput {
	t.Helper()

	input := CreateTestInput(t)
	input.Ctx = input.Ctx.WithBlockTime(blockTime)
	return input
}

// CreateTestInputWithKeysAndTime creates test input with custom store keys and time
func CreateTestInputWithKeysAndTime(t testing.TB, blockTime time.Time, keys ...string) TestInput {
	t.Helper()

	input := CreateTestInputWithKeys(t, keys...)
	input.Ctx = input.Ctx.WithBlockTime(blockTime)
	return input
}

// CreateTestInputWithStoreKey creates test input with a specific store key name
func CreateTestInputWithStoreKey(t testing.TB, storeKeyName string) TestInput {
	t.Helper()

	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())

	storeKey := storetypes.NewKVStoreKey(storeKeyName)
	memStoreKey := storetypes.NewMemoryStoreKey("mem_" + storeKeyName)

	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	cms.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)

	require.NoError(t, cms.LoadLatestVersion())

	ctx := sdk.NewContext(cms, cmtproto.Header{
		Height: 1,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	return TestInput{
		Ctx:         ctx,
		Cdc:         cdc,
		StoreKey:    storeKey,
		MemStoreKey: memStoreKey,
		DB:          db,
		CMS:         cms,
	}
}

// WrapStoreService wraps a store key as a KVStoreService for runtime compatibility
func WrapStoreService(storeKey *storetypes.KVStoreKey) corestore.KVStoreService {
	return runtime.NewKVStoreService(storeKey)
}

// Logger returns a no-op logger for testing
func Logger() log.Logger {
	return log.NewNopLogger()
}

// ConfigureSDK configures the Cosmos SDK with Aura-specific prefixes.
// This function uses sync.Once internally to ensure configuration happens only once.
// Safe to call multiple times from different test files.
func ConfigureSDK() {
	configureSDKOnce.Do(func() {
		cfg := sdk.GetConfig()
		cfg.SetBech32PrefixForAccount("aura", "aurapub")
		cfg.SetBech32PrefixForValidator("auravaloper", "auravaloperpub")
		cfg.SetBech32PrefixForConsensusNode("auravalcons", "auravalconspub")
		cfg.Seal()
	})
}

var configureSDKOnce = &sync.Once{}
