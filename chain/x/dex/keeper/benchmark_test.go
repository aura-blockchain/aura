package keeper_test

import (
	"fmt"
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
	sdkmath "cosmossdk.io/math"

	"github.com/aequitas/aura/chain/testutil/mocks"
	"github.com/aequitas/aura/chain/x/dex/keeper"
	"github.com/aequitas/aura/chain/x/dex/types"
)

// ============================================================================
// Pool Creation Benchmarks
// ============================================================================

// Helper function to setup keeper for benchmarking
func setupDEXBenchmark(b *testing.B) (*keeper.Keeper, sdk.Context) {
	b.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	if err := stateStore.LoadLatestVersion(); err != nil {
		b.Fatal(err)
	}

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
	if err := k.SetParams(ctx, &params); err != nil {
		b.Fatal(err)
	}

	return k, ctx
}

// BenchmarkCreatePool benchmarks the creation of a liquidity pool
func BenchmarkCreatePool(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)
	creator := "aura1creator"

	// Setup initial pool for reference
	denomA := "uaura"
	denomB := "usdt"
	amountA := sdk.NewCoin(denomA, sdkmath.NewInt(100000000))
	amountB := sdk.NewCoin(denomB, sdkmath.NewInt(50000000))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Use different denoms for each iteration to avoid duplicate pool errors
		testDenomA := fmt.Sprintf("%s%d", denomA, i)
		testDenomB := fmt.Sprintf("%s%d", denomB, i)
		testAmountA := sdk.NewCoin(testDenomA, amountA.Amount)
		testAmountB := sdk.NewCoin(testDenomB, amountB.Amount)

		_, _, _ = k.CreatePool(ctx, creator, testDenomA, testDenomB, testAmountA, testAmountB)
	}
}

// ============================================================================
// Liquidity Addition Benchmarks
// ============================================================================

// BenchmarkAddLiquidity benchmarks adding liquidity to an existing pool
func BenchmarkAddLiquidity(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)
	creator := "aura1creator"

	// Create initial pool
	denomA := "uaura"
	denomB := "usdt"
	amountA := sdk.NewCoin(denomA, sdkmath.NewInt(100000000))
	amountB := sdk.NewCoin(denomB, sdkmath.NewInt(50000000))

	pool, _, err := k.CreatePool(ctx, creator, denomA, denomB, amountA, amountB)
	if err != nil {
		b.Fatalf("Failed to create pool: %v", err)
	}

	// Liquidity to add
	addAmountA := sdk.NewCoin(denomA, sdkmath.NewInt(1000000))
	addAmountB := sdk.NewCoin(denomB, sdkmath.NewInt(500000))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Reset pool state for consistent benchmarking
		k.SetPool(ctx, pool)
		b.StartTimer()

		_, _, _ = k.AddLiquidity(ctx, creator, pool.PoolId, addAmountA, addAmountB)
	}
}

// ============================================================================
// Swap Benchmarks
// ============================================================================

// BenchmarkSwap_SmallPool benchmarks swap operations on a small liquidity pool
func BenchmarkSwap_SmallPool(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)
	creator := "aura1creator"

	// Create small pool: 100k AURA / 50k USDT
	denomA := "uaura"
	denomB := "usdt"
	amountA := sdk.NewCoin(denomA, sdkmath.NewInt(100000))
	amountB := sdk.NewCoin(denomB, sdkmath.NewInt(50000))

	pool, _, err := k.CreatePool(ctx, creator, denomA, denomB, amountA, amountB)
	if err != nil {
		b.Fatalf("Failed to create small pool: %v", err)
	}

	// Swap amount: 1000 AURA
	swapAmount := sdk.NewCoin(denomA, sdkmath.NewInt(1000))
	minAmountOut := sdkmath.NewInt(1)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		k.SetPool(ctx, pool)
		b.StartTimer()

		_, _, _, _ = k.Swap(ctx, creator, pool.PoolId, swapAmount, denomB, minAmountOut, 10000)
	}
}

// BenchmarkSwap_LargePool benchmarks swap operations on a large liquidity pool
func BenchmarkSwap_LargePool(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)
	creator := "aura1creator"

	// Create large pool: 100M AURA / 50M USDT
	denomA := "uaura"
	denomB := "usdt"
	amountA := sdk.NewCoin(denomA, sdkmath.NewInt(100000000))
	amountB := sdk.NewCoin(denomB, sdkmath.NewInt(50000000))

	pool, _, err := k.CreatePool(ctx, creator, denomA, denomB, amountA, amountB)
	if err != nil {
		b.Fatalf("Failed to create large pool: %v", err)
	}

	// Swap amount: 100k AURA
	swapAmount := sdk.NewCoin(denomA, sdkmath.NewInt(100000))
	minAmountOut := sdkmath.NewInt(1)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		k.SetPool(ctx, pool.PoolId)
		b.StartTimer()

		_, _, _, _ = k.Swap(ctx, creator, pool.PoolId, swapAmount, denomB, minAmountOut, 10000)
	}
}

// ============================================================================
// Order Book Benchmarks
// ============================================================================

// BenchmarkBatchExecution_100Orders benchmarks batch execution of 100 orders
func BenchmarkBatchExecution_100Orders(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)

	// Create pool for order execution
	denomA := "uaura"
	denomB := "usdt"
	amountA := sdk.NewCoin(denomA, sdkmath.NewInt(100000000))
	amountB := sdk.NewCoin(denomB, sdkmath.NewInt(50000000))

	pool, _, err := k.CreatePool(ctx, "aura1creator", denomA, denomB, amountA, amountB)
	if err != nil {
		b.Fatalf("Failed to create pool: %v", err)
	}

	// Create 100 limit orders
	orders := make([]*types.LimitOrder, 100)
	for i := 0; i < 100; i++ {
		orders[i] = &types.LimitOrder{
			OrderId:   fmt.Sprintf("order-%d", i),
			Owner:     fmt.Sprintf("aura1owner%d", i),
			PoolId:    pool.PoolId,
			Side:      types.OrderSideBuy,
			Price:     "0.5",
			Quantity:  sdkmath.NewInt(1000),
			Filled:    sdkmath.ZeroInt(),
			Status:    types.OrderStatusOpen,
			CreatedAt: ctx.BlockHeight(),
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Store orders in keeper
		for _, order := range orders {
			k.SetOrder(ctx, order)
		}
		b.StartTimer()

		// Execute batch (this will process all matching orders)
		_ = k.ExecuteBatch(ctx, pool.PoolId, 100)
	}
}

// BenchmarkBatchExecution_1000Orders benchmarks batch execution of 1000 orders
func BenchmarkBatchExecution_1000Orders(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)

	// Create pool for order execution
	denomA := "uaura"
	denomB := "usdt"
	amountA := sdk.NewCoin(denomA, sdkmath.NewInt(100000000))
	amountB := sdk.NewCoin(denomB, sdkmath.NewInt(50000000))

	pool, _, err := k.CreatePool(ctx, "aura1creator", denomA, denomB, amountA, amountB)
	if err != nil {
		b.Fatalf("Failed to create pool: %v", err)
	}

	// Create 1000 limit orders
	orders := make([]*types.LimitOrder, 1000)
	for i := 0; i < 1000; i++ {
		orders[i] = &types.LimitOrder{
			OrderId:   fmt.Sprintf("order-%d", i),
			Owner:     fmt.Sprintf("aura1owner%d", i),
			PoolId:    pool.PoolId,
			Side:      types.OrderSideBuy,
			Price:     "0.5",
			Quantity:  sdkmath.NewInt(1000),
			Filled:    sdkmath.ZeroInt(),
			Status:    types.OrderStatusOpen,
			CreatedAt: ctx.BlockHeight(),
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Store orders in keeper
		for _, order := range orders {
			k.SetOrder(ctx, order)
		}
		b.StartTimer()

		// Execute batch (this will process up to 1000 orders)
		_ = k.ExecuteBatch(ctx, pool.PoolId, 1000)
	}
}

// ============================================================================
// TWAP Calculation Benchmarks
// ============================================================================

// BenchmarkGetTWAPPrice benchmarks TWAP price calculation
func BenchmarkGetTWAPPrice(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)

	// Create pool
	denomA := "uaura"
	denomB := "usdt"
	amountA := sdk.NewCoin(denomA, sdkmath.NewInt(100000000))
	amountB := sdk.NewCoin(denomB, sdkmath.NewInt(50000000))

	pool, _, err := k.CreatePool(ctx, "aura1creator", denomA, denomB, amountA, amountB)
	if err != nil {
		b.Fatalf("Failed to create pool: %v", err)
	}

	// Add some price observations
	for i := 0; i < 10; i++ {
		k.RecordPrice(ctx, pool.PoolId, sdkmath.LegacyNewDecWithPrec(5, 1), ctx.BlockHeight()+int64(i))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, _ = k.GetTWAPPriceWithCount(ctx, pool.PoolId, 10)
	}
}

// ============================================================================
// Fee Calculation Benchmarks
// ============================================================================

// BenchmarkCalculateSwapFee benchmarks fee calculation
func BenchmarkCalculateSwapFee(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)

	amount := sdkmath.NewInt(1000000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = k.CalculateSwapFee(ctx, amount)
	}
}

// BenchmarkCalculateEffectiveFee benchmarks IR boost fee calculation
func BenchmarkCalculateEffectiveFee(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)

	address := "aura1user"
	baseFee := sdkmath.LegacyNewDecWithPrec(3, 3) // 0.003 (0.3%)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.CalculateEffectiveFee(ctx, address, baseFee)
	}
}

// ============================================================================
// Pool Query Benchmarks
// ============================================================================

// BenchmarkGetPool benchmarks single pool retrieval
func BenchmarkGetPool(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)

	// Create pool
	denomA := "uaura"
	denomB := "usdt"
	amountA := sdk.NewCoin(denomA, sdkmath.NewInt(100000000))
	amountB := sdk.NewCoin(denomB, sdkmath.NewInt(50000000))

	pool, _, err := k.CreatePool(ctx, "aura1creator", denomA, denomB, amountA, amountB)
	if err != nil {
		b.Fatalf("Failed to create pool: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.GetPool(ctx, pool.PoolId)
	}
}

// BenchmarkGetAllPools benchmarks retrieval of all pools
func BenchmarkGetAllPools(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)

	// Create 100 pools
	for i := 0; i < 100; i++ {
		denomA := fmt.Sprintf("token%da", i)
		denomB := fmt.Sprintf("token%db", i)
		amountA := sdk.NewCoin(denomA, sdkmath.NewInt(100000000))
		amountB := sdk.NewCoin(denomB, sdkmath.NewInt(50000000))

		_, _, err := k.CreatePool(ctx, "aura1creator", denomA, denomB, amountA, amountB)
		if err != nil {
			b.Fatalf("Failed to create pool %d: %v", i, err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.GetAllPools(ctx)
	}
}

// ============================================================================
// Price Impact Calculation Benchmarks
// ============================================================================

// BenchmarkCalculatePriceImpact benchmarks price impact calculation
func BenchmarkCalculatePriceImpact(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)

	// Create pool
	denomA := "uaura"
	denomB := "usdt"
	amountA := sdk.NewCoin(denomA, sdkmath.NewInt(100000000))
	amountB := sdk.NewCoin(denomB, sdkmath.NewInt(50000000))

	pool, _, err := k.CreatePool(ctx, "aura1creator", denomA, denomB, amountA, amountB)
	if err != nil {
		b.Fatalf("Failed to create pool: %v", err)
	}

	swapAmountIn := sdkmath.NewInt(100000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.CalculatePriceImpact(ctx, pool, swapAmountIn, denomA)
	}
}
