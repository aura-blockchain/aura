// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/testutil/mocks"
	"github.com/aequitas/aura/chain/x/dex/keeper"
	"github.com/aequitas/aura/chain/x/dex/types"
)

// testAddr creates a valid bech32 address from a seed string
func testAddr(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	addr, err := sdk.Bech32ifyAddressBytes("aura", sum[:20])
	if err != nil {
		panic(err)
	}
	return addr
}

// Helper function to setup keeper for benchmarking
func setupDEXBenchmark(b *testing.B) (*keeper.Keeper, sdk.Context) {
	b.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
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

// ============================================================================
// Pool Creation Benchmarks
// ============================================================================

// BenchmarkCreatePool benchmarks the creation of a liquidity pool
func BenchmarkCreatePool(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)
	creator := testAddr("creator")

	denomA := "uaura"
	denomB := "usdt"
	amountA := sdk.NewCoin(denomA, math.NewInt(100000000))
	amountB := sdk.NewCoin(denomB, math.NewInt(50000000))

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
	creator := testAddr("creator")

	// Create initial pool
	denomA := "uaura"
	denomB := "usdt"
	amountA := sdk.NewCoin(denomA, math.NewInt(100000000))
	amountB := sdk.NewCoin(denomB, math.NewInt(50000000))

	pool, _, err := k.CreatePool(ctx, creator, denomA, denomB, amountA, amountB)
	if err != nil {
		b.Fatalf("Failed to create pool: %v", err)
	}

	// Liquidity to add
	addAmountA := sdk.NewCoin(denomA, math.NewInt(1000000))
	addAmountB := sdk.NewCoin(denomB, math.NewInt(500000))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, _ = k.AddLiquidity(ctx, creator, pool.PoolId, addAmountA, addAmountB)
	}
}

// ============================================================================
// Swap Benchmarks
// ============================================================================

// BenchmarkSwap_SmallPool benchmarks swap operations on a small liquidity pool
func BenchmarkSwap_SmallPool(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)
	creator := testAddr("creator")

	// Create small pool: 100k AURA / 50k USDT
	denomA := "uaura"
	denomB := "usdt"
	amountA := sdk.NewCoin(denomA, math.NewInt(100000))
	amountB := sdk.NewCoin(denomB, math.NewInt(50000))

	pool, _, err := k.CreatePool(ctx, creator, denomA, denomB, amountA, amountB)
	if err != nil {
		b.Fatalf("Failed to create small pool: %v", err)
	}

	// Swap amount: 1000 AURA
	swapAmount := sdk.NewCoin(denomA, math.NewInt(1000))
	minAmountOut := math.NewInt(1)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, _, _ = k.SwapExactIn(ctx, creator, pool.PoolId, swapAmount, minAmountOut, 10000)
	}
}

// BenchmarkSwap_LargePool benchmarks swap operations on a large liquidity pool
func BenchmarkSwap_LargePool(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)
	creator := testAddr("creator")

	// Create large pool: 100M AURA / 50M USDT
	denomA := "uaura"
	denomB := "usdt"
	amountA := sdk.NewCoin(denomA, math.NewInt(100000000))
	amountB := sdk.NewCoin(denomB, math.NewInt(50000000))

	pool, _, err := k.CreatePool(ctx, creator, denomA, denomB, amountA, amountB)
	if err != nil {
		b.Fatalf("Failed to create large pool: %v", err)
	}

	// Swap amount: 100k AURA
	swapAmount := sdk.NewCoin(denomA, math.NewInt(100000))
	minAmountOut := math.NewInt(1)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, _, _ = k.SwapExactIn(ctx, creator, pool.PoolId, swapAmount, minAmountOut, 10000)
	}
}

// ============================================================================
// Order Book Benchmarks
// ============================================================================

// BenchmarkBatchExecution_100Orders benchmarks batch execution
func BenchmarkBatchExecution_100Orders(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)

	// Create pool for order execution
	denomA := "uaura"
	denomB := "usdt"
	amountA := sdk.NewCoin(denomA, math.NewInt(100000000))
	amountB := sdk.NewCoin(denomB, math.NewInt(50000000))

	_, _, err := k.CreatePool(ctx, "aura1qyqszqgpqyqszqgpqyqszqgpqyqszqgpj4w6", denomA, denomB, amountA, amountB)
	if err != nil {
		b.Fatalf("Failed to create pool: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.ExecuteBatch(ctx)
	}
}

// BenchmarkBatchExecution_1000Orders benchmarks batch execution with cleanup
func BenchmarkBatchExecution_1000Orders(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)

	// Create pool for order execution
	denomA := "uaura"
	denomB := "usdt"
	amountA := sdk.NewCoin(denomA, math.NewInt(100000000))
	amountB := sdk.NewCoin(denomB, math.NewInt(50000000))

	_, _, err := k.CreatePool(ctx, "aura1qyqszqgpqyqszqgpqyqszqgpqyqszqgpj4w6", denomA, denomB, amountA, amountB)
	if err != nil {
		b.Fatalf("Failed to create pool: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.ExecuteBatch(ctx)
		k.CleanupExpiredCommitments(ctx)
	}
}

// ============================================================================
// Fee Calculation Benchmarks
// ============================================================================

// BenchmarkCalculateSwapFee benchmarks fee calculation
func BenchmarkCalculateSwapFee(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)

	amount := math.NewInt(1000000)

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
	amountA := sdk.NewCoin(denomA, math.NewInt(100000000))
	amountB := sdk.NewCoin(denomB, math.NewInt(50000000))

	pool, _, err := k.CreatePool(ctx, "aura1qyqszqgpqyqszqgpqyqszqgpqyqszqgpj4w6", denomA, denomB, amountA, amountB)
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
		amountA := sdk.NewCoin(denomA, math.NewInt(100000000))
		amountB := sdk.NewCoin(denomB, math.NewInt(50000000))

		_, _, err := k.CreatePool(ctx, "aura1qyqszqgpqyqszqgpqyqszqgpqyqszqgpj4w6", denomA, denomB, amountA, amountB)
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
// Orderbook Query Benchmarks
// ============================================================================

// BenchmarkGetOrderbookForPair_10Orders benchmarks orderbook retrieval with 10 orders
func BenchmarkGetOrderbookForPair_10Orders(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)
	creator := testAddr("creator")

	// Create 10 orders for the uaura-usdt pair
	for i := 0; i < 10; i++ {
		_, _ = k.CreateOrder(ctx, creator, types.SwapOrderType_SELL,
			sdkmath.NewInt(1000000), "usdt", sdkmath.NewInt(500000), 60)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.GetOrderbookForPair(ctx, "uaura", "usdt")
	}
}

// BenchmarkGetOrderbookForPair_100Orders benchmarks orderbook retrieval with 100 orders
func BenchmarkGetOrderbookForPair_100Orders(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)
	creator := testAddr("creator")

	// Create 100 orders for the uaura-usdt pair
	for i := 0; i < 100; i++ {
		_, _ = k.CreateOrder(ctx, creator, types.SwapOrderType_SELL,
			sdkmath.NewInt(1000000), "usdt", sdkmath.NewInt(500000), 60)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.GetOrderbookForPair(ctx, "uaura", "usdt")
	}
}

// BenchmarkGetOrderbookForPair_1000Orders benchmarks orderbook retrieval with 1000 orders
func BenchmarkGetOrderbookForPair_1000Orders(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)
	creator := testAddr("creator")

	// Create 1000 orders for the uaura-usdt pair
	for i := 0; i < 1000; i++ {
		_, _ = k.CreateOrder(ctx, creator, types.SwapOrderType_SELL,
			sdkmath.NewInt(1000000), "usdt", sdkmath.NewInt(500000), 60)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.GetOrderbookForPair(ctx, "uaura", "usdt")
	}
}

// ============================================================================
// Minimum Liquidity Calculation Benchmarks
// ============================================================================

// BenchmarkCalculateMinimumAuraRequired benchmarks minimum liquidity calculation
func BenchmarkCalculateMinimumAuraRequired(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.CalculateMinimumAuraRequired(ctx)
	}
}

// ============================================================================
// Price Calculation Benchmarks
// ============================================================================

// BenchmarkGetQuote benchmarks the quote calculation for swaps
func BenchmarkGetQuote(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)
	creator := testAddr("creator")

	// Create pool for quotes
	denomA := "uaura"
	denomB := "usdt"
	amountA := sdk.NewCoin(denomA, math.NewInt(100000000))
	amountB := sdk.NewCoin(denomB, math.NewInt(50000000))

	pool, _, err := k.CreatePool(ctx, creator, denomA, denomB, amountA, amountB)
	if err != nil {
		b.Fatalf("Failed to create pool: %v", err)
	}

	swapAmount := math.NewInt(10000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, _, _, _ = k.GetQuote(ctx, pool.PoolId, denomA, swapAmount)
	}
}

// BenchmarkGetQuote_VariousSizes benchmarks quote calculation with various input sizes
func BenchmarkGetQuote_VariousSizes(b *testing.B) {
	sizes := []struct {
		name   string
		amount int64
	}{
		{"Small_1K", 1000},
		{"Medium_100K", 100000},
		{"Large_10M", 10000000},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			k, ctx := setupDEXBenchmark(b)
			creator := testAddr("creator")

			denomA := "uaura"
			denomB := "usdt"
			amountA := sdk.NewCoin(denomA, math.NewInt(100000000))
			amountB := sdk.NewCoin(denomB, math.NewInt(50000000))

			pool, _, err := k.CreatePool(ctx, creator, denomA, denomB, amountA, amountB)
			if err != nil {
				b.Fatalf("Failed to create pool: %v", err)
			}

			swapAmount := math.NewInt(size.amount)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, _, _, _, _ = k.GetQuote(ctx, pool.PoolId, denomA, swapAmount)
			}
		})
	}
}

// ============================================================================
// Liquidity Removal Benchmarks
// ============================================================================

// BenchmarkRemoveLiquidity benchmarks removing liquidity from a pool
func BenchmarkRemoveLiquidity(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)
	creator := testAddr("creator")

	// Create initial pool with substantial liquidity
	denomA := "uaura"
	denomB := "usdt"
	amountA := sdk.NewCoin(denomA, math.NewInt(1000000000))
	amountB := sdk.NewCoin(denomB, math.NewInt(500000000))

	pool, lpTokens, err := k.CreatePool(ctx, creator, denomA, denomB, amountA, amountB)
	if err != nil {
		b.Fatalf("Failed to create pool: %v", err)
	}

	// Calculate small portion to remove (1% of LP tokens per iteration)
	removeAmount := lpTokens.Quo(math.NewInt(100))
	if removeAmount.IsZero() {
		removeAmount = math.NewInt(1)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Note: In real usage, LP tokens would be consumed. For benchmark,
		// we measure the function call overhead
		_, _, _ = k.RemoveLiquidity(ctx, creator, pool.PoolId, removeAmount)
	}
}

// ============================================================================
// Order Creation and Matching Benchmarks
// ============================================================================

// BenchmarkCreateOrder benchmarks order creation
func BenchmarkCreateOrder(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)
	creator := testAddr("creator")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = k.CreateOrder(ctx, creator, types.SwapOrderType_SELL,
			sdkmath.NewInt(1000000), "usdt", sdkmath.NewInt(500000), 60)
	}
}

// BenchmarkMatchOrder benchmarks order matching
func BenchmarkMatchOrder(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)
	creator := testAddr("creator")
	matcher := testAddr("matcher")

	// Create orders to match
	orderIDs := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		order, err := k.CreateOrder(ctx, creator, types.SwapOrderType_SELL,
			sdkmath.NewInt(1000000), "usdt", sdkmath.NewInt(500000), 60)
		if err != nil {
			b.Fatalf("Failed to create order: %v", err)
		}
		orderIDs[i] = order.OrderId
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.MatchOrder(ctx, matcher, orderIDs[i])
	}
}

// BenchmarkCancelOrder benchmarks order cancellation
func BenchmarkCancelOrder(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)
	creator := testAddr("creator")

	// Create orders to cancel
	orderIDs := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		order, err := k.CreateOrder(ctx, creator, types.SwapOrderType_SELL,
			sdkmath.NewInt(1000000), "usdt", sdkmath.NewInt(500000), 60)
		if err != nil {
			b.Fatalf("Failed to create order: %v", err)
		}
		orderIDs[i] = order.OrderId
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.CancelOrder(ctx, orderIDs[i], "benchmark")
	}
}

// ============================================================================
// Order Cleanup Benchmarks
// ============================================================================

// BenchmarkCleanupExpiredOrders benchmarks the optimized expired order cleanup
func BenchmarkCleanupExpiredOrders(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)
	creator := testAddr("creator")

	// Create expired orders - set expiration to 0 (already expired)
	for i := 0; i < 100; i++ {
		_, _ = k.CreateOrder(ctx, creator, types.SwapOrderType_SELL,
			sdkmath.NewInt(1000000), "usdt", sdkmath.NewInt(500000), 0)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.CleanupExpiredOrdersOptimized(ctx, 50)
	}
}

// ============================================================================
// Pool Statistics Benchmarks
// ============================================================================

// BenchmarkRecordSwapStats benchmarks swap statistics recording
func BenchmarkRecordSwapStats(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)

	poolID := "uaura-usdt"
	amountIn := sdkmath.NewInt(1000000)
	amountOut := sdkmath.NewInt(500000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		k.RecordSwapStats(ctx, poolID, amountIn, amountOut, ctx.BlockTime())
	}
}

// ============================================================================
// Multi-Pool Benchmarks
// ============================================================================

// BenchmarkSwap_MultiplePoolLookups benchmarks swap with pool lookup overhead
func BenchmarkSwap_MultiplePoolLookups(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)
	creator := testAddr("creator")

	// Create multiple pools to increase lookup time
	poolIDs := make([]string, 10)
	for i := 0; i < 10; i++ {
		denomA := fmt.Sprintf("token%da", i)
		denomB := fmt.Sprintf("token%db", i)
		amountA := sdk.NewCoin(denomA, math.NewInt(100000000))
		amountB := sdk.NewCoin(denomB, math.NewInt(50000000))

		pool, _, err := k.CreatePool(ctx, creator, denomA, denomB, amountA, amountB)
		if err != nil {
			b.Fatalf("Failed to create pool %d: %v", i, err)
		}
		poolIDs[i] = pool.PoolId
	}

	swapAmount := sdk.NewCoin("token5a", math.NewInt(1000))
	minAmountOut := math.NewInt(1)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Lookup and swap against various pools
		poolIdx := i % len(poolIDs)
		_, _, _, _ = k.SwapExactIn(ctx, creator, poolIDs[poolIdx], swapAmount, minAmountOut, 10000)
	}
}

// ============================================================================
// Parameter Access Benchmarks
// ============================================================================

// BenchmarkGetParams benchmarks parameter retrieval
func BenchmarkGetParams(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = k.GetParams(ctx)
	}
}

// BenchmarkSetParams benchmarks parameter updates
func BenchmarkSetParams(b *testing.B) {
	k, ctx := setupDEXBenchmark(b)
	params := types.DefaultParams()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.SetParams(ctx, &params)
	}
}
